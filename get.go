package oc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GetResponse SearchResponse

// GetRequest is the request payload for a call to client.Get().
type GetRequest struct {
	UUIDs      []string
	Properties string
	Filters    string
	Deleted    bool
}

func nonEmptyCSVParam(v url.Values, name string, values []string) {
	if len(values) == 0 {
		return
	}

	v.Set(name, strings.Join(values, ","))
}

func nonEmptyParam(v url.Values, name, value string) {
	if len(value) == 0 {
		return
	}

	v.Set(name, value)
}

// Get is used to retrieve properties for a given set of uuids. It requires a
// backend of OC 3.0+, because it uses a new endpoint `/get`. This endpoint
// uses what is called 'real-time get' in solr. This basically means retrieving
// document in a way that is cheaper than an ordinary search but you are still
// assured you get what is in the index.
func (c *Client) Get(ctx context.Context, req GetRequest) (*GetResponse, error) {
	if len(req.UUIDs) == 0 {
		return nil, errors.New("missing UUIDs to get")
	}

	v := url.Values{}

	nonEmptyCSVParam(v, "uuid", req.UUIDs)
	nonEmptyParam(v, "filters", req.Filters)
	nonEmptyParam(v, "properties", req.Properties)

	if req.Deleted {
		v.Set("deleted", "true")
	}

	res, err := c.fetch(ctx, "get", v, fetchWithAcceptJSON())
	if err != nil {
		return nil, err
	}

	defer safeClose(c.logger, "get response", res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	var getres GetResponse

	err = json.NewDecoder(res.Body).Decode(&getres)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &getres, nil
}
