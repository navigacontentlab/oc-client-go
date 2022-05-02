package oc

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

type eventlogEvents struct {
	Events []EventlogEvent `json:"events"`
}

type EventlogEvent struct {
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

func (c *Client) Eventlog(ctx context.Context, event int) ([]EventlogEvent, error) {
	q := url.Values{}
	q.Set("event", strconv.Itoa(event))

	var events eventlogEvents

	_, err := c.getJSON(ctx, "eventlog", q, &events)
	if err != nil {
		return nil, err
	}

	return events.Events, nil
}
