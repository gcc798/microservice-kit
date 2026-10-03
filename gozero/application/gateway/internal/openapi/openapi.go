package openapi

import (
	_ "embed"
	"net/http"
)

//go:embed swagger.json
var document []byte

func Document(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(document)
}
