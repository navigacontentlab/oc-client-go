package oc_test

import (
	"context"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	oc "github.com/navigacontentlab/oc-client-go/v2"
)

func TestClient_Health(t *testing.T) {
	client := clientFromEnvironment(t)

	res, err := client.Health(context.Background(), oc.HealthRequest{})
	if err != nil {
		log.Fatal(err)
	}

	want := oc.Health{
		Indexer:  true,
		Solr:     true,
		Database: true,
		Storage:  true,
	}

	got := oc.Health{
		Indexer:  res.Indexer,
		Solr:     res.Solr,
		Database: res.Database,
		Storage:  res.Storage,
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Health() mismatch (-want +got):\n%s", diff)
	}
}
