package oc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type SuggestType int

const (
	Facet SuggestType = iota
	Ngram
)

func (s SuggestType) String() string {
	switch s {
	case Ngram:
		return "ngram"
	case Facet:
		return "facet"
	}

	return "facet"
}

type SuggestRequest struct {
	IndexFields          []SuggestIndexField
	Limit                int
	Query                string
	IncompleteWordInText string
	Type                 SuggestType
	Timezone             string
}

type SuggestIndexField struct {
	Name           string
	IncompleteWord string
}

type Term struct {
	Name      string `json:"name"`
	Frequency int    `json:"frequency"`
}

type SuggestField struct {
	Name  string `json:"name"`
	Terms []Term `json:"terms"`
}

type SuggestResponse struct {
	Fields []SuggestField `json:"facetFields"`
}

func (sr *SuggestRequest) QueryValues() (url.Values, error) {
	q := url.Values{}

	for _, indexField := range sr.IndexFields {
		q.Add("field", indexField.Name)

		if sr.Type == Ngram && indexField.IncompleteWord == "" {
			return nil, errors.New("ngram suggest must use incompleteWord")
		}

		if indexField.IncompleteWord != "" {
			q.Add("incompleteWord", indexField.IncompleteWord)
		}
	}

	if sr.IncompleteWordInText != "" {
		q.Set("incompleteWordInText", sr.IncompleteWordInText)
	}

	if sr.Limit > 0 {
		q.Add("limit", strconv.Itoa(sr.Limit))
	}

	if sr.Query != "" {
		q.Add("q", sr.Query)
	}

	if sr.Timezone != "" {
		q.Add("timezone", sr.Timezone)
	}

	q.Set("type", sr.Type.String())

	return q, nil
}

// Suggest performs a suggest against Open Content.
func (c *Client) Suggest(ctx context.Context, req SuggestRequest) (*SuggestResponse, error) {
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

	res, err := c.fetch(ctx, "suggest", queryValues, opts...)
	if err != nil {
		return nil, err
	}

	if c.metrics != nil {
		c.metrics.incStatusCode(ctx, "suggest", res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	var resp SuggestResponse

	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
