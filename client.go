package oc

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-log/log"
)

// Options controls the behaviour of the Open Content client.
type Options struct {
	BaseURL    string
	HTTPClient *http.Client
	Auth       AuthenticationMethod
	Logger     log.Logger
	Metrics    *Metrics
}

// AuthenticationMethod is a function that adds authentication
// information to a request.
type AuthenticationMethod func(req *http.Request)

// BasicAuth adds a basic auth authorisation header to the outgoing
// requests.
func BasicAuth(username, password string) AuthenticationMethod {
	return func(req *http.Request) {
		req.SetBasicAuth(username, password)
	}
}

// BasicAuth adds a bearer token authorisation header to the outgoing
// requests.
func BearerAuth(token string) AuthenticationMethod {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// Client is an Open Content client.
type Client struct {
	baseURL    *url.URL
	auth       AuthenticationMethod
	httpClient *http.Client
	logger     log.Logger
	metrics    *Metrics
}

// New creates a new Open Content client.
func New(opt Options) (*Client, error) {
	base := opt.BaseURL
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := opt.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	logger := opt.Logger
	if logger == nil {
		logger = log.DefaultLogger
	}

	return &Client{
		baseURL:    baseURL,
		auth:       opt.Auth,
		httpClient: client,
		logger:     logger,
		metrics:    opt.Metrics,
	}, nil
}

func (c *Client) url(resource string, q url.Values) string {
	return c.baseURL.ResolveReference(&url.URL{
		Path:     resource,
		RawQuery: q.Encode(),
	}).String()
}

type fetchOption func(req *http.Request, info *requestInfo)

type requestInfo struct {
	mainResource string
}

func fetchWithAccept(accept string) fetchOption {
	return func(req *http.Request, info *requestInfo) {
		req.Header.Set("Accept", accept)
	}
}

func fetchWithAcceptJSON() fetchOption {
	return fetchWithAccept("application/json")
}

func fetchWithNoneMatch(etag string) fetchOption {
	return func(req *http.Request, info *requestInfo) {
		req.Header.Set("If-None-Match", etag)
	}
}

func fetchWithResourceName(name string) fetchOption {
	return func(req *http.Request, info *requestInfo) {
		info.mainResource = name
	}
}

func fetchWithMethod(method string) fetchOption {
	return func(req *http.Request, info *requestInfo) {
		req.Method = method
	}
}

func (c *Client) fetch(
	ctx context.Context,
	resource string, q url.Values, opts ...fetchOption,
) (*http.Response, error) {
	start := time.Now()

	reqURL := c.url(resource, q)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	var info requestInfo

	for i := range opts {
		opts[i](req, &info)
	}

	if info.mainResource == "" {
		seg := strings.Split(resource, "/")
		info.mainResource = seg[0]
	}

	if info.mainResource == "" {
		info.mainResource = "/"
	}

	resp, err := c.doRequest(ctx, req)

	if c.metrics != nil {
		duration := time.Since(start)

		c.metrics.addDuration(ctx, info.mainResource, float64(duration.Milliseconds()))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return resp, newResponseError(resp)
	}

	return resp, nil
}

func (c *Client) getJSON(
	ctx context.Context, resource string, q url.Values, result interface{}, opts ...fetchOption,
) (http.Header, error) {
	opts = append(opts, fetchWithAcceptJSON())

	resp, err := c.fetch(ctx, resource, q, opts...)
	if err != nil {
		return nil, err
	}

	defer safeClose(c.logger, resource+" response", resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, newResponseError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decode json response: %w",
			err)
	}

	return resp.Header, nil
}

func (c *Client) getXML(
	ctx context.Context, resource string, q url.Values,
	result interface{}, opts ...fetchOption,
) error {
	opts = append(opts, fetchWithAccept("text/xml"))

	resp, err := c.fetch(ctx, resource, q, opts...)
	if err != nil {
		return err
	}

	defer safeClose(c.logger, resource+" response", resp.Body)

	if resp.StatusCode != http.StatusOK {
		return newResponseError(resp)
	}

	err = xml.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return fmt.Errorf(
			"failed to decode xml response: %w",
			err)
	}

	return nil
}

func safePath(segments ...string) string {
	var path string

	for i, s := range segments {
		if i > 0 {
			path += "/"
		}

		path += url.PathEscape(s)
	}

	return path
}

func discardAndClose(rc io.ReadCloser) error {
	_, err := io.Copy(ioutil.Discard, rc)
	if err != nil {
		return fmt.Errorf(
			"failed to drain reader: %w",
			err)
	}

	if err := rc.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}

	return nil
}

func (c *Client) doRequest(
	ctx context.Context, req *http.Request,
) (*http.Response, error) {
	req = req.WithContext(ctx)

	if c.auth != nil {
		c.auth(req)
	}

	return c.httpClient.Do(req) //nolint:wrapcheck
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	res, err := c.fetch(ctx, "infoandstats/version", nil)
	if err != nil {
		return "", err
	}

	defer safeClose(c.logger, "version body", res.Body)

	if res.StatusCode != http.StatusOK {
		return "", newResponseError(res)
	}

	versionData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response data: %w", err)
	}

	return string(versionData), nil
}
