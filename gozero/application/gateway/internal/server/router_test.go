package server

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	registry "github.com/gcc798/microservice-kit/internal/registry"
)

func TestBuildRouteTable(t *testing.T) {
	instances := []registry.HTTPInstance{
		{Service: "iam-api", Endpoint: "http://127.0.0.1:9011", Routes: []registry.HTTPRoute{{Method: "GET", Path: "/api/v1/user/:id"}}},
		{Service: "iam-api", Endpoint: "http://127.0.0.2:9011", Routes: []registry.HTTPRoute{{Method: "GET", Path: "/api/v1/user/:userId"}}},
		{Service: "sys-api", Endpoint: "http://127.0.0.1:9012", Routes: []registry.HTTPRoute{{Method: "GET", Path: "/api/v1/config/code"}}},
	}
	table, err := buildRouteTable(instances)
	if err != nil {
		t.Fatal(err)
	}
	if target := table.match("GET", "/api/v1/config/code"); target == nil || target.service != "sys-api" {
		t.Fatalf("static route target = %#v", target)
	}
	if target := table.match("GET", "/api/v1/user/42"); target == nil || target.service != "iam-api" || len(target.endpoints) != 2 {
		t.Fatalf("parameter route target = %#v", target)
	}
	instances = append(instances, registry.HTTPInstance{Service: "other", Endpoint: "http://127.0.0.1:9999", Routes: []registry.HTTPRoute{{Method: "GET", Path: "/api/v1/user/:name"}}})
	if _, err := buildRouteTable(instances); err == nil {
		t.Fatal("route ownership conflict must fail")
	}
}

func TestEveryBusinessRouteHasPermissionPolicy(t *testing.T) {
	data, err := os.ReadFile("../openapi/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	parameter := regexp.MustCompile(`\{[^/]+\}`)
	for path, operations := range document.Paths {
		if !strings.HasPrefix(path, "/api/v1/") {
			continue
		}
		route := parameter.ReplaceAllString(path, ":id")
		for method := range operations {
			method = strings.ToUpper(method)
			if requiredPermission(method, route) == nil && !isPermissionExempt(method, route) {
				t.Errorf("%s %s has no permission policy", method, path)
			}
		}
	}
}

func TestPublicPaths(t *testing.T) {
	for path, want := range map[string]bool{
		"/login": true, "/auth/refresh": true, "/resource/sms/code": true, "/captcha/image": true,
		"/logout": false, "/auth/anything": false, "/api/v1/user": false, "/authentic": false,
	} {
		if got := isPublic(path); got != want {
			t.Fatalf("isPublic(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestRequiredPermission(t *testing.T) {
	for _, test := range []struct {
		method     string
		path       string
		want       string
		wantAction string
	}{
		{"POST", "/api/v1/user", "user.create", "write"},
		{"POST", "/api/v1/user/page", "user.read", "read"},
		{"GET", "/api/v1/role/42/api-permissions", "api_permission.assign", "write"},
		{"POST", "/api/v1/role/42/menus", "role.update", "write"},
		{"GET", "/api/v1/attachment/business", "attachment.read", "read"},
		{"GET", "/api/v1/attachment/42/download", "attachment.download", "write"},
		{"POST", "/api/v1/user/password/change", "", ""},
		{"GET", "/api/v1/menu/user/tree", "", ""},
	} {
		got := requiredPermission(test.method, test.path)
		if test.want == "" && got != nil {
			t.Errorf("requiredPermission(%q, %q) = %q, want none", test.method, test.path, got.resource)
		}
		if test.want != "" && (got == nil || got.resource != test.want || got.action != test.wantAction) {
			t.Errorf("requiredPermission(%q, %q) = %#v, want %q/%q", test.method, test.path, got, test.want, test.wantAction)
		}
	}
}

func TestPermissionExempt(t *testing.T) {
	if !isPermissionExempt("POST", "/api/v1/user/password/change") || !isPermissionExempt("GET", "/api/v1/menu/user/tree") {
		t.Fatal("authenticated self-service routes must remain permission-exempt")
	}
	if isPermissionExempt("GET", "/api/v1/user/42") {
		t.Fatal("business routes must fail closed when no permission rule exists")
	}
}
