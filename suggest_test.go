package oc_test

import (
	"testing"

	"github.com/navigacontentlab/oc-client-go"
)

func TestSuggestResponse_Unmarshal(t *testing.T) {
	var resp oc.SuggestResponse

	tests := []struct {
		Term  string
		Count int
	}{
		{"Article", 3704712},
		{"Image", 3420565},
		{"Assignment", 203104},
		{"Planning", 190488},
		{"Concept", 114278},
		{"List", 19357},
		{"Package", 12853},
		{"Template", 897},
	}

	loadTestData(t, "suggestresponse.json", &resp)

	if len(resp.Fields) != 1 {
		t.Errorf("unexpected fields size, expected %d, got %d", 1, len(resp.Fields))
	}

	if resp.Fields[0].Name != "contenttype" {
		t.Error("field name it not contenttype")
	}

	for i, term := range resp.Fields[0].Terms {
		if term.Name != tests[i].Term {
			t.Errorf("unexpected term name for index %d, expected %q, got %q", i, tests[i].Term, term.Name)
		} else if term.Frequency != tests[i].Count {
			t.Errorf("unexpected term frequency for index %d, expected %d, got %d", i, tests[i].Count, term.Frequency)
		}
	}
}
