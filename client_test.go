package oc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	oc "github.com/navigacontentlab/oc-client-go/v2"
)

func TestAuthenticationMethod(t *testing.T) {
	var ctxKey struct{}

	fail := make(chan error, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		defer close(fail)

		want := "Bearer a-test-token"
		got := r.Header.Get("Authorization")

		if got != want {
			fail <- fmt.Errorf(
				"unexpected authorization token value: want %q got %q",
				want, got,
			)
		}
	}))
	defer ts.Close()

	client, err := oc.New(oc.Options{
		BaseURL: ts.URL,
		Auth: func(req *http.Request) {
			val := req.Context().Value(&ctxKey)

			if token, ok := val.(string); ok {
				req.Header.Add("Authorization", "Bearer "+token)
			}
		},
	})
	if err != nil {
		t.Fatalf("failed to create OC client: %v", err)
	}

	bearerToken := "a-test-token"

	// Create a context that has auth info
	ctx := context.WithValue(context.Background(), &ctxKey, bearerToken)

	_, err = client.GetObject(ctx, "e019fc86-301f-4307-8ff5-9814f6982bb0", 1)
	if err != nil {
		t.Fatalf("failed to make test request: %v", err)
	}

	// Check if the test server didn't get the session token.
	if err := <-fail; err != nil {
		t.Error(err)
	}
}
