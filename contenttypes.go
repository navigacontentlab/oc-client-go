package oc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ContentTypesRequest struct {
	Temporary bool
}

type ContentTypesResponse struct {
	ContentTypes []struct {
		Name       string `json:"name"`
		Properties []struct {
			Name           string `json:"name"`
			Type           string `json:"type"`
			MultiValued    bool   `json:"multiValued"`
			Searchable     bool   `json:"searchable"`
			ReadOnly       bool   `json:"readOnly"`
			Description    string `json:"description"`
			Suggest        bool   `json:"suggest"`
			IndexFieldType string `json:"indexFieldType"`
		} `json:"properties"`
	} `json:"contentTypes"`
}

func (cr *ContentTypesRequest) QueryValues() (url.Values, error) {
	q := url.Values{}

	if cr.Temporary {
		q.Add("temporary", strconv.FormatBool(cr.Temporary))
	}

	return q, nil
}

// ContentTypes gets the schema (all content types and properties) from Open Content.
func (c *Client) ContentTypes(ctx context.Context, req ContentTypesRequest) (*ContentTypesResponse, error) {
	var (
		queryValues url.Values
		err         error
	)

	opts := []fetchOption{
		fetchWithAcceptJSON(),
	}

	queryValues, err = req.QueryValues()
	if err != nil {
		return nil, err
	}

	res, err := c.fetch(ctx, "contenttypes", queryValues, opts...)
	if err != nil {
		return nil, err
	}

	if c.metrics != nil {
		c.metrics.incStatusCode(ctx, "contenttypes", res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	var resp ContentTypesResponse

	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
