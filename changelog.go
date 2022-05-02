package oc

import (
	"context"
	"encoding/xml"
	"net/url"
	"strconv"
)

type ChangelogEntry struct {
	Text    string `xml:",chardata"`
	ID      string `xml:"id"`
	Title   string `xml:"title"`
	Updated string `xml:"updated"`
}

type Feed struct {
	XMLName      xml.Name         `xml:"feed"`
	Text         string           `xml:",chardata"`
	TotalChanges string           `xml:"totalChanges,attr"`
	Xmlns        string           `xml:"xmlns,attr"`
	Entries      []ChangelogEntry `xml:"entry"`
	Title        string           `xml:"title"`
}

func (c *Client) Changelog(ctx context.Context, start int, limit int) (*Feed, error) {
	q := url.Values{}
	q.Set("start", strconv.Itoa(start))
	q.Set("limit", strconv.Itoa(limit))

	var feed Feed

	err := c.getXML(ctx, "changelog", q, &feed)
	if err != nil {
		return nil, err
	}

	return &feed, nil
}
