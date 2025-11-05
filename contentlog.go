package oc

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

type contentlogEvents struct {
	Events []ContentlogEvent `json:"events"`
}

type ContentlogEvent struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	EventType string    `json:"eventType"`
	Created   time.Time `json:"created"`
	Content   struct {
		UUID        string      `json:"uuid"`
		Version     int         `json:"version"`
		Created     time.Time   `json:"created"`
		Source      interface{} `json:"source"`
		ContentType string      `json:"contentType"`
		Batch       bool        `json:"batch"`
	} `json:"content"`
}

func (c *Client) Contentlog(ctx context.Context, event int) ([]ContentlogEvent, error) {
	q := url.Values{}
	q.Set("event", strconv.Itoa(event))

	var events contentlogEvents

	_, err := c.GetJSON(ctx, "contentlog", q, &events)
	if err != nil {
		return nil, err
	}

	return events.Events, nil
}
