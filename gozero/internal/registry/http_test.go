package registry

import (
	"net/http"
	"testing"

	"github.com/zeromicro/go-zero/rest"
)

func TestDecodeHTTPInstances(t *testing.T) {
	values := []string{`{"service":"iam-api","endpoint":"http://127.0.0.1:9011","routes":[{"method":"GET","path":"/api/v1/user/:id"}]}`}
	instances, err := Decode(values)
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 1 || instances[0].Service != "iam-api" || instances[0].Routes[0].Path != "/api/v1/user/:id" {
		t.Fatalf("Decode() = %#v", instances)
	}
}

func TestPublishHTTPRouteSelection(t *testing.T) {
	routes := []rest.Route{
		{Method: http.MethodGet, Path: "/health"},
		{Method: http.MethodGet, Path: "/api/v1/user/:id"},
	}
	registered := Routes(routes)
	if len(registered) != 1 || registered[0].Path != "/api/v1/user/:id" {
		t.Fatalf("registered routes = %#v", registered)
	}
}
