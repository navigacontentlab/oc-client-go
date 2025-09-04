package oc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ObjectResponse struct {
	ETag        string
	ContentType string
	Body        io.ReadCloser
	Version     int64
}

func (c *Client) GetObject(ctx context.Context, uuid string, version int64) (*ObjectResponse, error) {
	q := url.Values{}

	if version != 0 {
		q.Set("version", strconv.FormatInt(version, 10))
	}

	res, err := c.fetch(ctx, joinPath("objects", uuid), q)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	v, err := objectVersionFromHeader(res.Header, versionOptional)
	if err != nil {
		c.logger.Log(err.Error())
	}

	return &ObjectResponse{
		ETag:        res.Header.Get("ETag"),
		ContentType: res.Header.Get("Content-Type"),
		Body:        res.Body,
		Version:     v,
	}, nil
}

const (
	versionRequired = true
	versionOptional = false
)

func objectVersionFromHeader(header http.Header, required bool) (int64, error) {
	versionHeader := header.Get("X-Opencontent-Object-Version")
	if versionHeader == "" {
		if required {
			return 0, errors.New("missing X-Opencontent-Object-Version in response")
		}

		return 0, nil
	}

	v, err := strconv.ParseInt(versionHeader, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid X-Opencontent-Object-Version %q: %w",
			versionHeader, err,
		)
	}

	return v, nil
}

type UndeleteOptions struct {
	// Unit is passed to OC as a X-Imid-Unit HTTP header is passed
	Unit string
}

func (c *Client) Undelete(ctx context.Context, uuid string, options *UndeleteOptions) error {
	reqURL := c.url(joinPath("objects", uuid, "undelete"), nil)

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	if options != nil && options.Unit != "" {
		req.Header.Set("X-Imid-Unit", options.Unit)
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return newResponseError(resp)
	}

	return discardAndClose(resp.Body)
}

type ConsistencyCheckResponse struct {
	Database ConsistencyStatus `json:"database"`
	Index    ConsistencyStatus `json:"index"`
	Storage  ConsistencyStatus `json:"storage"`
}

type ConsistencyStatus struct {
	Deleted  bool   `json:"deleted"`
	Checksum string `json:"checksum"`
	Version  int    `json:"version"`
}

// ConsistencyCheck performs a check against the database, storage,
// and index.
func (c *Client) ConsistencyCheck(ctx context.Context, uuid string) (*ConsistencyStatus, error) {
	var status ConsistencyStatus

	_, err := c.getJSON(
		ctx,
		"objects/"+uuid+"/consistency-check",
		nil, &status,
	)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

type DeleteOptions struct {
	// IfMatch causes the object to only be deleted if its ETag
	// matches the provided value.
	IfMatch string
	// Unit is passed to OC as a X-Imid-Unit HTTP header is passed
	Unit string
}

// Delete deletes an object.
func (c *Client) Delete(ctx context.Context, uuid string, options *DeleteOptions) error {
	reqURL := c.url(joinPath("objects", uuid), nil)

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	if options != nil && options.Unit != "" {
		req.Header.Set("X-Imid-Unit", options.Unit)
	}

	if options != nil && options.IfMatch != "" {
		req.Header.Set("If-Match", options.IfMatch)
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return newResponseError(resp)
	}

	return discardAndClose(resp.Body)
}

type PurgeOptions struct {
	// IfMatch causes the object to only be purged if its ETag
	// matches the provided value.
	IfMatch string
}

// Purge purges an object.
func (c *Client) Purge(ctx context.Context, uuid string, options *PurgeOptions) error {
	reqURL := c.url(joinPath("objects", uuid, "purge"), nil)

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	if options != nil && options.IfMatch != "" {
		req.Header.Set("If-Match", options.IfMatch)
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return newResponseError(resp)
	}

	return discardAndClose(resp.Body)
}

type FileResponse struct {
	ETag        string
	ContentType string
	Version     int64
	Body        io.ReadCloser
}

func (c *Client) GetFile(ctx context.Context, uuid string, filename string, version int64) (*FileResponse, error) {
	q := url.Values{}

	if version != 0 {
		q.Set("version", strconv.FormatInt(version, 10))
	}

	res, err := c.fetch(ctx, joinPath("objects", uuid, "files", filename), q)
	if c.metrics != nil && res != nil {
		c.metrics.incStatusCode(ctx, "objects", res.StatusCode)
	}

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	v, err := objectVersionFromHeader(res.Header, versionOptional)
	if err != nil {
		c.logger.Logf("failed to get version for file response: %v", err)
	}

	return &FileResponse{
		ETag:        res.Header.Get("ETag"),
		ContentType: res.Header.Get("Content-Type"),
		Body:        res.Body,
		Version:     v,
	}, nil
}

type FileList struct {
	Version   int64
	Primary   ObjectFile   `json:"primary"`
	Metadata  []ObjectFile `json:"metadata"`
	Preview   ObjectFile   `json:"preview"`
	Thumb     ObjectFile   `json:"thumb"`
	Created   *time.Time   `json:"created"`
	Updated   *time.Time   `json:"updated"`
	EventType string       `json:"eventType"`
}

type ObjectFile struct {
	Name     string `json:"filename"`
	Mimetype string `json:"mimetype"`
}

func (c *Client) ListFiles(ctx context.Context, uuid string, version int64) (*FileList, error) {
	q := url.Values{}

	if version != 0 {
		q.Set("version", strconv.FormatInt(version, 10))
	}

	var list FileList

	headers, err := c.getJSON(
		ctx, joinPath("objects", uuid, "files"), q, &list,
		fetchWithResourceName("objects/files"),
	)
	if err != nil {
		return nil, err
	}

	v, err := objectVersionFromHeader(headers, versionRequired)
	if err != nil {
		c.logger.Logf("failed to get version for files response: %v", err)
	}

	list.Version = v

	return &list, nil
}

type ReplaceMetadataRequest struct {
	UUID        string
	Filename    string
	ContentType string
	Body        io.Reader
	IfMatch     string
	Batch       bool
	Unit        string
}

func (c *Client) ReplaceMetadataFile(ctx context.Context, req ReplaceMetadataRequest) error {
	q := url.Values{}

	if req.Batch {
		q.Set("batch", "true")
	}

	reqURL := c.url(joinPath(
		"objects", req.UUID,
		"files", "metadata", req.Filename,
	), q)

	r, err := http.NewRequest("PUT", reqURL, req.Body)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	if req.Unit != "" {
		r.Header.Set("X-Imid-Unit", req.Unit)
	}

	r.Header.Set("Content-Type", req.ContentType)

	if req.IfMatch != "" {
		r.Header.Set("If-Match", req.IfMatch)
	}

	resp, err := c.doRequest(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return newResponseError(resp)
	}

	return discardAndClose(resp.Body)
}

type DeleteMetadataRequest struct {
	// UUID of the document to delete a metadata file for.
	UUID string
	// Filename of the metadata file to delete.
	Filename string
	// IfMatch only deletes the metadata file if the main object
	// still has a matching ETag. Optional.
	IfMatch string
	// Unit is used when passing a X-Imid-Unit header to OC
	Unit string
}

func (c *Client) DeleteMetadataFile(ctx context.Context, req DeleteMetadataRequest) error {
	reqURL := c.url(joinPath(
		"objects", req.UUID,
		"files", "metadata", req.Filename,
	), nil)

	r, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	if req.Unit != "" {
		r.Header.Set("X-Imid-Unit", req.Unit)
	}

	if req.IfMatch != "" {
		r.Header.Set("If-Match", req.IfMatch)
	}

	resp, err := c.doRequest(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return newResponseError(resp)
	}

	return discardAndClose(resp.Body)
}

func (c *Client) GetMetadataFile(ctx context.Context, uuid string, version int) (*FileResponse, error) {
	q := url.Values{}

	if version != 0 {
		q.Set("version", strconv.Itoa(version))
	}

	res, err := c.fetch(
		ctx, joinPath("objects", uuid, "files", "metadata"),
		q, fetchWithResourceName("metadata"))
	if c.metrics != nil && res != nil {
		c.metrics.incStatusCode(ctx, "metadata", res.StatusCode)
	}

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	return &FileResponse{
		ETag:        res.Header.Get("ETag"),
		ContentType: res.Header.Get("Content-Type"),
		Body:        res.Body,
	}, nil
}

func (c *Client) Properties(ctx context.Context, uuid string, properties PropertyList) (*PropertyResult, error) {
	q := url.Values{}

	list, err := properties.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("bad property list: %w", err)
	}

	q.Set("properties", string(list))

	var res PropertyResult

	_, err = c.getJSON(
		ctx, joinPath("objects", uuid, "properties"),
		q, &res, fetchWithResourceName("objects/properties"))
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) PropertiesVersion(
	ctx context.Context,
	uuid string, version int64, properties PropertyList,
) (*PropertyResult, error) {
	q := url.Values{}

	list, err := properties.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("bad property list: %w", err)
	}

	q.Set("properties", string(list))
	q.Set("version", fmt.Sprintf("%d", version))

	var res PropertyResult

	_, err = c.getJSON(
		ctx, joinPath("objects", uuid, "properties"),
		q, &res, fetchWithResourceName("objects/properties"))
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) Head(ctx context.Context, uuid string) (int64, error) {
	var q url.Values

	resp, err := c.fetch(
		ctx, joinPath("objects", uuid), q,
		fetchWithMethod(http.MethodHead),
	)
	if err != nil {
		return 0, err
	}

	err = resp.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("failed to close response body: %w", err)
	}

	versionHeader := resp.Header.Get("X-Opencontent-Object-Version")

	v, err := strconv.ParseInt(versionHeader, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid X-Opencontent-Object-Version %q returned for %q",
			versionHeader, resp.Request.URL.String(),
		)
	}

	return v, nil
}

type ExistsResponse struct {
	Exists        bool
	ETag          string
	ContentType   string
	ContentLength int
	Version       int64
}

// CheckExists does a HEAD request against an object and returns
// information about the object.
func (c *Client) CheckExists(ctx context.Context, uuid string) (*ExistsResponse, error) {
	res, err := c.fetch(
		ctx, joinPath("objects", uuid), nil,
		fetchWithMethod(http.MethodHead),
	)
	if err != nil {
		return nil, err
	}

	defer safeClose(c.logger, "exists check body", res.Body)

	if res.StatusCode == http.StatusNotFound {
		return &ExistsResponse{}, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	result := ExistsResponse{
		Exists:      true,
		ContentType: res.Header.Get("Content-Type"),
		ETag:        res.Header.Get("Etag"),
	}

	lengthString := res.Header.Get("Content-Length")

	contentLength, err := strconv.Atoi(lengthString)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid content length %q from server: %w",
			lengthString, err)
	}

	result.ContentLength = contentLength

	v, err := objectVersionFromHeader(res.Header, versionRequired)
	if err != nil {
		c.logger.Logf("failed to get version for exists response: %v", err)
	}

	result.Version = v

	return &result, nil
}
