package oc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	oc "github.com/navigacontentlab/oc-client-go/v2"
)

func ExampleClient_Search() {
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

	ctx := context.Background()

	searchReq := oc.SearchRequest{
		Start:       0,
		Limit:       15,
		Properties:  "uuid,Headline,created,updated",
		ContentType: "Article",
		Sort: []oc.SearchSort{
			{IndexField: "updated"},
		},
	}

	_, err = client.Search(ctx, searchReq)
	if err != nil {
		fmt.Println(err)
	}
}

// Test unmarshalling and helper functions, this should be
// complemented with some fuzz testing as well.
func TestSearchResponse__Unmarshal(t *testing.T) {
	var resp oc.SearchResponse

	loadTestData(t, "searchresponse.json", &resp)

	scooterID := "bd121fb8-addb-4121-8455-008229bf1dae"
	scooterMayhem := mustHitItem(t, resp.Hits, scooterID)

	expectHeadline := "Scooters cause chaos and mayhem"

	gotHeadline, ok := scooterMayhem.Properties.Get("Headline")
	if !ok {
		t.Error("missing Headline property")
	}

	if gotHeadline != expectHeadline {
		t.Errorf("unexpected Headline expected %q, got %q", expectHeadline, gotHeadline)
	}

	rels, ok := scooterMayhem.Properties.Relationships("ConceptRelations")
	if !ok {
		t.Error("missing ConceptRelations relationships")
	}

	if len(rels) == 0 {
		t.Errorf("no ConceptRelations relationships for %s", scooterID)
	}

	expectCID := "af7adf45-4506-4e4a-8841-b3434cef462f"
	expectCHeadline := "Ronnie J. Willis"

	var gotCHeadline *string

	for i := range rels {
		if id, _ := rels[i].Get("uuid"); id != expectCID {
			continue
		}

		h, ok := rels[i].Get("Headline")
		if !ok {
			t.Errorf("missing Headline property for content relation %s", expectCID)
			break
		}

		gotCHeadline = &h
	}

	if gotCHeadline == nil {
		t.Errorf("could not find Headline for content relation %s", expectCID)
	} else if *gotCHeadline != expectCHeadline {
		t.Errorf("unexpected Headline for content relation %s expected %q, got %q", expectCID, expectCHeadline, *gotCHeadline)
	}

	expectHighlight := "<em>Scooters</em> cause chaos and mayhem"

	highlight, ok := resp.Highlight[scooterID]
	if !ok {
		t.Errorf("missing highlights for %s", scooterID)
	}

	highlightHeadline, ok := highlight.Get("Headline")
	if !ok {
		t.Errorf("missing highlight Headline for %s", scooterID)
	}

	if highlightHeadline != expectHighlight {
		t.Errorf("unexpected highlight for %s expected %q, got %q", scooterID, expectHighlight, highlightHeadline)
	}
}

// Test support for multiple property values.
func TestSearchResponse__Unmarshal__Multi(t *testing.T) {
	var resp oc.SearchResponse

	loadTestData(t, "searchresponse-2.json", &resp)

	hit := mustHitItem(t, resp.Hits, "49478a1c-ab79-5cef-bb78-4f659e287387")

	uuids, ok := hit.Properties.GetValues("ArticleMetaImageUuids")
	if !ok {
		t.Error("missing ArticleMetaImageUuids")
		return
	}

	expect := []string{
		"24db3152-7479-58ed-9d87-c0b364ee68e9",
		"586e7978-0520-5353-b397-c410246c2a41",
	}

	if len(uuids) != len(expect) {
		t.Errorf("missing unexpected number of ArticleMetaImageUuids, want %d, got %d",
			len(expect), len(uuids))

		return
	}

	for i := range expect {
		if uuids[i] != expect[i] {
			t.Errorf("missing unexpected ArticleMetaImageUuids[%d], want %q, got %q",
				i, expect[i], uuids[i])
		}
	}
}

func mustHitItem(t *testing.T, hits oc.Hits, uuid string) *oc.Hit {
	t.Helper()

	var h *oc.Hit

	for i := range hits.Items {
		if hits.Items[i].ID == uuid {
			h = &hits.Items[i]
		}
	}

	if h == nil {
		t.Fatalf("failed to find hit with ID %s", uuid)
	}

	return h
}

func loadTestData(t *testing.T, name string, v interface{}) {
	t.Helper()

	path := "test/" + name

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		t.Fatalf("failed to read test data from %q: %v", path, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("failed to unmarshal test data into %T: %v", v, err)
	}
}
