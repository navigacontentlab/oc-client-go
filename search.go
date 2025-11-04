package oc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type BoundaryType int

const (
	Inclusive BoundaryType = iota
	Exclusive
)

func (ts BoundaryType) String() string {
	switch ts {
	case Inclusive:
		return "inclusive"
	case Exclusive:
		return "exclusive"
	default:
		return "unknown"
	}
}

type DateBoundary struct {
	Type BoundaryType
	Date *time.Time
}

type DateRange struct {
	Start DateBoundary
	End   DateBoundary
}

type SearchRequest struct {
	IfNoneMatch string
	Start       int
	Limit       int
	Property    string
	Properties  string
	Filters     string
	Query       string
	FilterQuery string
	Sort        []SearchSort
	ContentType string // OC type, i.e. Article
	Deleted     bool
	Facets      SearchFacets
	Highlight   []string
	Created     DateRange
	Updated     DateRange
}

type SearchFacets struct {
	Fields   []string
	Limit    int
	MinCount int
}

type SearchSort struct {
	IndexField string
	Descending bool
}

type SearchResponse struct {
	Hits      Hits                  `json:"hits"`
	Facet     FacetFields           `json:"facet"`
	Stats     Stats                 `json:"stats"`
	Highlight map[string]Properties `json:"highlight"`
}

type Hits struct {
	TotalHits    int   `json:"totalHits"`
	Items        []Hit `json:"hits"`
	IncludedHits int   `json:"includedHits"`
}

type Hit struct {
	ID         string     `json:"id"`
	Version    int        `json:"version"`
	Properties Properties `json:"properties"`
}

func (h *Hit) UnmarshalJSON(data []byte) error {
	var inner hit

	if err := json.Unmarshal(data, &inner); err != nil {
		return err //nolint:wrapcheck
	}

	h.ID = inner.ID

	if len(inner.Versions) == 0 {
		return nil
	}

	h.Version = inner.Versions[0].ID
	h.Properties = inner.Versions[0].Properties

	return nil
}

func (h Hit) MarshalJSON() ([]byte, error) {
	inner := hit{
		ID: h.ID,
		Versions: []version{{
			ID:         h.Version,
			Properties: h.Properties,
		}},
	}

	return json.Marshal(&inner) //nolint:wrapcheck
}

type hit struct {
	ID       string    `json:"id"`
	Versions []version `json:"versions"`
}

type version struct {
	ID         int        `json:"id"`
	Properties Properties `json:"properties"`
}

type Properties map[string][]interface{}

func (p Properties) Get(property string) (string, bool) {
	if p == nil {
		return "", false
	}

	values, ok := p[property]
	if !ok || len(values) == 0 {
		return "", false
	}

	value, ok := values[0].(string)
	if !ok {
		return "", false
	}

	return value, true
}

func (p Properties) GetValues(property string) ([]string, bool) {
	if p == nil {
		return nil, false
	}

	values, ok := p[property]
	if !ok {
		return nil, false
	}

	v := make([]string, len(values))

	for i := range values {
		s, ok := values[i].(string)
		if !ok {
			continue
		}

		v[i] = s
	}

	return v, true
}

func (p Properties) Relationships(property string) ([]Properties, bool) {
	if p == nil {
		return nil, false
	}

	values, ok := p[property]
	if !ok {
		return nil, false
	}

	v := make([]Properties, len(values))

	for i := range values {
		m, ok := values[i].(map[string]interface{})
		if !ok {
			continue
		}

		relProps := make(Properties)

		for k := range m {
			relValues, ok := m[k].([]interface{})
			if !ok {
				continue
			}

			relProps[k] = relValues
		}

		v[i] = relProps
	}

	return v, true
}

type FacetFields struct {
	Fields []FacetField `json:"fields"`
}

type FacetField struct {
	Year        int         `json:"year"`
	Month       int         `json:"month"`
	Day         int         `json:"day"`
	FacetField  string      `json:"facetField"`
	Frequencies []Frequency `json:"frequencies"`
}

type Frequency struct {
	Term      string `json:"term"`
	Frequency int    `json:"frequency"`
}

type Stats struct {
	Duration int         `json:"duration"`
	Hits     interface{} `json:"hits"`
}

type Highlights map[string][]string

func (sr *SearchRequest) QueryValues() (url.Values, error) {
	q := url.Values{}

	if sr.ContentType != "" {
		q.Add("contenttype", sr.ContentType)
	}

	q.Add("start", strconv.Itoa(sr.Start))

	if sr.Limit == 0 {
		sr.Limit = 15
	}

	q.Add("limit", strconv.Itoa(sr.Limit))

	if sr.Property != "" {
		q.Add("property", sr.Property)
	}

	if sr.Properties != "" {
		q.Add("properties", sr.Properties)
	}

	if sr.Filters != "" {
		q.Add("filters", sr.Filters)
	}

	if sr.Query != "" {
		q.Add("q", sr.Query)
	}

	if sr.FilterQuery != "" {
		q.Add("fq", sr.FilterQuery)
	}

	for _, sort := range sr.Sort {
		q.Add("sort.indexfield", sort.IndexField)
		q.Set(
			"sort."+sort.IndexField+".ascending",
			strconv.FormatBool(!sort.Descending),
		)
	}

	// Handle created date range with inclusive/exclusive support
	if sr.Created.Start.Date != nil {
		paramName := "created.start"
		if sr.Created.Start.Type == Inclusive {
			paramName += "inclusive"
		} else {
			paramName += "exclusive"
		}

		q.Add(paramName, sr.Created.Start.Date.Format("2006-01-02T15:04:05Z"))
	}

	if sr.Created.End.Date != nil {
		paramName := "created.end"
		if sr.Created.End.Type == Inclusive {
			paramName += "inclusive"
		} else {
			paramName += "exclusive"
		}

		q.Add(paramName, sr.Created.End.Date.Format("2006-01-02T15:04:05Z"))
	}

	// Handle updated date range with inclusive/exclusive support
	if sr.Updated.Start.Date != nil {
		paramName := "updated.start"
		if sr.Updated.Start.Type == Inclusive {
			paramName += "inclusive"
		} else {
			paramName += "exclusive"
		}

		q.Add(paramName, sr.Updated.Start.Date.Format("2006-01-02T15:04:05Z"))
	}

	if sr.Updated.End.Date != nil {
		paramName := "updated.end"
		if sr.Updated.End.Type == Inclusive {
			paramName += "inclusive"
		} else {
			paramName += "exclusive"
		}

		q.Add(paramName, sr.Updated.End.Date.Format("2006-01-02T15:04:05Z"))
	}

	if len(sr.Facets.Fields) > 0 {
		q.Set("facet.indexfield", strings.Join(sr.Facets.Fields, "\n"))

		if sr.Facets.Limit > 0 {
			q.Set("facet.limit", strconv.Itoa(sr.Facets.Limit))
		}

		if sr.Facets.MinCount > 0 {
			q.Set("facet.mincount", strconv.Itoa(sr.Facets.MinCount))
		}
	}

	if len(sr.Highlight) > 0 {
		q.Set("highlight.indexfield", strings.Join(sr.Highlight, "\n"))
	}

	return q, nil
}

// Search performs a search against Open Content.
func (c *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
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

	if req.IfNoneMatch != "" {
		opts = append(opts, fetchWithNoneMatch(req.IfNoneMatch))
	}

	res, err := c.fetch(ctx, "search", queryValues, opts...)
	if err != nil {
		return nil, err
	}

	if c.metrics != nil {
		c.metrics.incStatusCode(ctx, "search", res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, newResponseError(res)
	}

	var resp SearchResponse

	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
