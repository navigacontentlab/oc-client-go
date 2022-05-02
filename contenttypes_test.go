package oc_test

import (
	"testing"

	"github.com/navigacontentlab/oc-client-go"
)

func TestContentTypesResponse_Unmarshal(t *testing.T) {
	var resp oc.ContentTypesResponse

	tests := []struct {
		ContentType     string
		PropertiesCount int
	}{
		{"Article", 27},
		{"Assignment", 18},
		{"Concept", 23},
		{"Event", 3},
		{"Graphic", 4},
		{"Image", 20},
	}

	loadTestData(t, "contenttypesresponse.json", &resp)

	if len(tests) != len(resp.ContentTypes) {
		t.Errorf("unexpected number of contenttypes, expected %q, got %q", len(tests), len(resp.ContentTypes))
	}

	for i, contentType := range resp.ContentTypes {
		if contentType.Name != tests[i].ContentType {
			t.Errorf("unexpected term name for index %d, expected %q, got %q", i, tests[i].ContentType, contentType.Name)
		} else if len(contentType.Properties) != tests[i].PropertiesCount {
			t.Errorf("unexpected term frequency for index %d, expected %d, got %d", i,
				tests[i].PropertiesCount, len(contentType.Properties))
		}
	}
}
