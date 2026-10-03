package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPStatusUsesEnvelopeCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	HTTPStatus(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":404,"msg":"不存在"}`))
	})(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
