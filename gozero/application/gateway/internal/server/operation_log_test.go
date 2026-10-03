package server

import (
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSanitizeQuery(t *testing.T) {
	got := sanitizeQuery(url.Values{"token": {"secret"}, "page": {"1"}})
	if got.Get("token") != "***" || got.Get("page") != "1" {
		t.Fatalf("sanitized query = %v", got)
	}
}

func TestOperationParamsRedactsNestedJSONAndRestoresBody(t *testing.T) {
	request := httptest.NewRequest("POST", "/login", strings.NewReader(`{"userName":"admin","password":"secret","nested":{"accessKeySecret":"key"}}`))
	body := readAuditBody(request)
	if got := operationParams(request, body); got != `{"nested":{"accessKeySecret":"***"},"password":"***","userName":"admin"}` {
		t.Fatalf("sanitized body = %s", got)
	}
	restored, err := io.ReadAll(request.Body)
	if err != nil || string(restored) != string(body) {
		t.Fatalf("restored body = %q, err = %v", restored, err)
	}
}
