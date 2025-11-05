package oc

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// HealthRequest is a request to the health check endpoint.
type HealthRequest struct {
	// SkipIndexer tells OC to omit the indexer health check.
	SkipIndexer bool
	// SkipSolr tells OC to omit the Solr health check.
	SkipSolr bool
	// SkipStorage tells OC to omit the storage health check.
	SkipStorage bool
}

// Health is a OpenContent health check response.
type Health struct {
	Indexer             bool  `json:"indexer"`
	Solr                bool  `json:"index"`
	Database            bool  `json:"database"`
	Storage             bool  `json:"filesystem"`
	FreeSystemDiskSpace int64 `json:"freeSystemDiskSpace"`
	MaximumMemory       int   `json:"maximumMemory"`
	CurrentMemory       int   `json:"currentMemory"`
	ActiveConfiguration struct {
		Checksum     string    `json:"checksum"`
		LastModified time.Time `json:"lastModified"`
	} `json:"activeConfiguration"`
	TempConfiguration struct {
		Checksum     string    `json:"checksum"`
		LastModified time.Time `json:"lastModified"`
	} `json:"tempConfiguration"`
}

// Health performs a health check request against OC.
func (c *Client) Health(ctx context.Context, req HealthRequest) (Health, error) {
	q := url.Values{}

	q.Set("indexer", strconv.FormatBool(!req.SkipIndexer))
	q.Set("solr", strconv.FormatBool(!req.SkipSolr))
	q.Set("storage", strconv.FormatBool(!req.SkipStorage))

	var health Health

	_, err := c.GetJSON(ctx, "health", q, &health)
	if err != nil {
		return health, err
	}

	return health, nil
}
