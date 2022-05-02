package oc_test

import (
	"context"
	"os"
	"testing"
)

func TestClient_CheckExists(t *testing.T) {
	client := clientFromEnvironment(t)

	uuid := os.Getenv("TEST_DOC_UUID")
	if uuid == "" {
		t.Skip("need a TEST_DOC_UUID to test CheckExists()")
	}

	info, err := client.CheckExists(context.Background(), uuid)
	if err != nil {
		t.Fatal(err)
	}

	if !info.Exists {
		t.Fatal("expected object to exist")
	}

	if info.ContentLength == 0 {
		t.Error("expected the object to have a length")
	}

	if info.Version == 0 {
		t.Error("expected the object to have a version")
	}

	if len(info.ContentType) == 0 {
		t.Error("expected the object to have a content type")
	}

	if len(info.ETag) == 0 {
		t.Error("expected the object to have an Etag")
	}
}
