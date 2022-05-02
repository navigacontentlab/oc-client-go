package oc_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/navigacontentlab/oc-client-go"
)

// TestAPI is an example of how an application or library that uses OC
// Client would go about mocking it. In short: create an interface for
// the API surface that you're actually using.
func TestAPI(t *testing.T) {
	mockOC := &mockOCClient{}

	ctx := context.Background()

	got, err := CalculateMegaHash(ctx, mockOC, "11111111-e819-423c-861c-026e7f6fc412")
	if err != nil {
		t.Error("mocked exists check shouldn't fail")
	}

	want := "4fef5814581f0c7466618d8879112231bc6d6d134673c029ef1a4ac6799b2db5"
	if got != want {
		t.Errorf("invalid mega hash, got %q, wanted %q", got, want)
	}
}

type mockOCClient struct {
}

func (m *mockOCClient) CheckExists(_ context.Context, _ string) (*oc.ExistsResponse, error) {
	return &oc.ExistsResponse{
		Exists:  true,
		ETag:    "this-is-a-hash-from-OC",
		Version: 43,
	}, nil
}

type ExistsChecker interface {
	CheckExists(ctx context.Context, uuid string) (*oc.ExistsResponse, error)
}

func CalculateMegaHash(ctx context.Context, client ExistsChecker, uuid string) (string, error) {
	res, err := client.CheckExists(ctx, uuid)
	if err != nil {
		return "", fmt.Errorf(
			"failed to check if document exists: %w", err)
	}

	if !res.Exists {
		return "", errors.New("you cannot do anything to that which does not exist")
	}

	hash := sha256.New()

	_, err = fmt.Fprintf(hash, "%s+%d+%s", uuid, res.Version, res.ETag)
	if err != nil {
		return "", fmt.Errorf("failed to hash document info: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// TestAPI__MockServer goes a step further and mocks the actual OC
// server. This is probably more useful when testing the client
// itself, but could also be used for testing application code.
func TestAPI__MockServer(t *testing.T) {
	docA := "31749523-e819-423c-861c-026e7f6fc412"
	docB := "11111111-e819-423c-861c-026e7f6fc412"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if user != "user" || pass != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/objects/"+docA {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Add("content-type", "application/xml;charset=UTF-8")
		w.Header().Add("content-length", "681")
		w.Header().Add("etag", "ca9bb8c2dc8df7e62654bbb2dc037527ca9bb8c2dc8df7e62654bbb2dc037527")
		w.Header().Add("x-opencontent-object-version", "1")
		w.WriteHeader(http.StatusOK)
	}))

	defer ts.Close()

	client, err := oc.New(oc.Options{
		BaseURL:    ts.URL,
		Auth:       oc.BasicAuth("user", "password"),
		HTTPClient: ts.Client(),
	})
	if err != nil {
		t.Fatalf("failed to create OC client: %v", err)
	}

	got, err := client.CheckExists(context.Background(), docA)
	if err != nil {
		t.Fatalf("failed to check if document A exists: %v", err)
	}

	want := oc.ExistsResponse{
		Exists:        true,
		ETag:          "ca9bb8c2dc8df7e62654bbb2dc037527ca9bb8c2dc8df7e62654bbb2dc037527",
		ContentLength: 681,
		Version:       1,
		ContentType:   "application/xml;charset=UTF-8",
	}

	if diff := cmp.Diff(&want, got); diff != "" {
		t.Errorf("CheckExists() mismatch (-want +got):\n%s", diff)
	}

	res, err := client.CheckExists(context.Background(), docB)
	if err != nil {
		t.Fatalf("failed to check if document B exists: %v", err)
	}

	if res.Exists {
		t.Fatal("expected document A to not exist")
	}
}
