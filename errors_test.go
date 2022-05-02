package oc //nolint:testpackage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOCResponseError__SurfaceErrorBody(t *testing.T) {
	rec := httptest.NewRecorder()

	ocMsg := "Object could not be uploaded or deleted: A file without metadata can't be imported"

	rec.WriteHeader(http.StatusBadRequest)
	_, _ = rec.WriteString(ocMsg)

	err := newResponseError(rec.Result()) //nolint:bodyclose

	if !strings.Contains(err.Error(), ocMsg) {
		t.Errorf("expected the error message to contain the string %q, got %q",
			ocMsg, err.Error(),
		)
	}
}

func TestOCResponseError__IgnoreGarbage(t *testing.T) {
	rec := httptest.NewRecorder()

	ocMsg := "Sane \ttext"

	rec.WriteHeader(http.StatusExpectationFailed)
	_, _ = rec.Body.WriteString(ocMsg)
	_ = rec.Body.WriteByte(0)
	_, _ = rec.Body.WriteString("Post garbage")

	err := newResponseError(rec.Result()) //nolint:bodyclose

	if !strings.HasSuffix(err.Error(), ocMsg) {
		t.Errorf("expected the error message to end with the the string %q, got %q",
			ocMsg, err.Error(),
		)
	}

	if !strings.Contains(err.Error(), ocMsg) {
		t.Errorf("expected the error message to contain the string %q, got %q",
			ocMsg, err.Error(),
		)
	}
}
