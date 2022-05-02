package oc_test

import (
	"context"
	"testing"
)

func TestClient_Changelog(t *testing.T) {
	client := clientFromEnvironment(t)

	feed, err := client.Changelog(context.Background(), 0, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(feed.Entries) != 1 {
		t.Errorf(
			"unexpected number of changelog entries returned: expected 1 got %d",
			len(feed.Entries),
		)
	}
}
