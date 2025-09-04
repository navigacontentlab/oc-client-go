package oc_test

import (
	"context"
	"os"
	"testing"

	oc "github.com/navigacontentlab/oc-client-go/v2"
)

func TestClient_Get(t *testing.T) {
	client := clientFromEnvironment(t)
	requireOCVersion(t, client, ">= 3.0.0-M1")

	uuid := os.Getenv("TEST_DOC_UUID")
	if uuid == "" {
		t.Skip("need a TEST_DOC_UUID to test Get()")
	}

	entries, err := client.Get(context.Background(), oc.GetRequest{
		UUIDs: []string{uuid},
	})
	if err != nil {
		t.Fatal(err)
	}

	if entries.Hits.TotalHits != 1 {
		t.Errorf("expected 1 hit to be returned, got %d", entries.Hits.TotalHits)
	}
}
