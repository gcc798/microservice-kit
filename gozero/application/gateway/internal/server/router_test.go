package server

import "testing"

func TestTargetForPath(t *testing.T) {
	targets := Targets{IAM: "iam", SYS: "sys", Resource: "resource", Realtime: "realtime"}
	for path, want := range map[string]string{
		"/login":                    "iam",
		"/api/v1/user/1":            "iam",
		"/api/v1/config/page":       "sys",
		"/api/v1/attachment/1":      "resource",
		"/realtime/websocket":       "realtime",
		"/api/v1/configuration":     "",
		"/realtime/websocket/extra": "",
	} {
		if got := targets.targetForPath(path); got != want {
			t.Fatalf("targetForPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestPublicPaths(t *testing.T) {
	for path, want := range map[string]bool{"/login": true, "/captcha/image": true, "/api/v1/user": false, "/authentic": false} {
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
