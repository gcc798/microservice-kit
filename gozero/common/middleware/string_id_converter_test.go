package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStringIDConverter(t *testing.T) {
	body := []byte(`{"id":"9007199254740993","nested":{"roleIds":["9007199254740994",7]},"clientId":"web-admin","count":9007199254740995}`)
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	StringIDConverter(func(_ http.ResponseWriter, request *http.Request) {
		got, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"clientId":"web-admin","count":9007199254740995,"id":9007199254740993,"nested":{"roleIds":[9007199254740994,7]}}`
		if string(got) != want {
			t.Fatalf("converted body = %s, want %s", got, want)
		}
	})(httptest.NewRecorder(), request)
}
