package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPStatus keeps the response envelope and HTTP status consistent. goctl's
// generated handlers always use OkJsonCtx for a successful logic call, while
// logic responses may carry an explicit 4xx/5xx code.
func HTTPStatus(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writer := &statusWriter{ResponseWriter: w}
		next(writer, r)
		writer.commit(writer.status)
	}
}

type statusWriter struct {
	http.ResponseWriter
	status    int
	committed bool
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *statusWriter) Write(body []byte) (int, error) {
	status := w.status
	if status == 0 {
		status = http.StatusOK
	}
	if status == http.StatusOK && strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		var envelope struct {
			Code int `json:"code"`
		}
		if json.Unmarshal(body, &envelope) == nil && envelope.Code >= 400 && envelope.Code <= 599 {
			status = envelope.Code
		}
	}
	w.commit(status)
	return w.ResponseWriter.Write(body)
}

func (w *statusWriter) commit(status int) {
	if w.committed {
		return
	}
	if status == 0 {
		status = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(status)
	w.committed = true
}
