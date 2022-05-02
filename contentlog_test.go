package oc_test

import (
	"context"
	"log"
	"testing"
)

func TestClient_Contentlog(t *testing.T) {
	client := clientFromEnvironment(t)

	entries, err := client.Contentlog(context.Background(), 0)
	if err != nil {
		log.Fatal(err)
	}

	if len(entries) == 0 {
		t.Error("did not get any log entries")
	}
}
