package oc_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/navigacontentlab/oc-client-go"
)

func TestRetrieveUploadVersion(t *testing.T) {
	testCases := []struct {
		description string
		resFn       func(http.ResponseWriter, *http.Request)
		want        int64
	}{
		{
			"without version",
			func(w http.ResponseWriter, r *http.Request) {},
			0,
		}, // version defaults to 0
		{
			"with version",
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-OpenContent-object-version", "42")
			},
			42,
		},
	}
	for _, tc := range testCases {
		ts := httptest.NewServer(http.HandlerFunc(tc.resFn))

		client, err := oc.New(oc.Options{
			BaseURL: ts.URL,
			Auth:    oc.BasicAuth("testuser", "testpass"),
		})
		if err != nil {
			t.Errorf("%v", err)
		}

		res, err := client.Upload(context.Background(), oc.UploadRequest{})
		if err != nil {
			t.Errorf("%v", err)
		}

		if res.Version != tc.want {
			t.Errorf("expected %d, got %d", tc.want, res.Version)
		}
	}
}

// Will not actually be run since there is no output
// because we cannot run this multiple times to the same
// oc.
func ExampleClient_Upload() {
	client, err := oc.New(oc.Options{
		BaseURL: os.Getenv("OC_BASEURL"),
		Auth: oc.BasicAuth(
			os.Getenv("OC_USERNAME"),
			os.Getenv("OC_PASSWORD"),
		),
	})
	if err != nil {
		log.Fatal(err)
	}

	reader, err := os.Open("sample.jpeg")
	if err != nil {
		log.Fatal(err)
	}

	metadataReader, err := os.Open("sample-image-metadata.xml")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	uploadReq := oc.UploadRequest{
		UUID:   uuid.New().String(),
		Source: "mycustomuploader",
		Files: oc.FileSet{
			"file": oc.File{

				Name:     "sample.jpeg",
				Reader:   reader,
				Mimetype: "image/jpeg",
			},
			"metadata": oc.File{

				Name:     "sample-image.metadata.xml",
				Reader:   metadataReader,
				Mimetype: "application/vnd.iptc.g2.newsitem+xml.picture",
			},
		},
	}

	_, err = client.Upload(ctx, uploadReq)

	if err != nil {
		fmt.Println(err)
	}
}
