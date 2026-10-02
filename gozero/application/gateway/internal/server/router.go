package server

import "strings"

type permissionRule struct {
	method   string
	path     string
	resource string
	action   string
}

var permissionRules = []permissionRule{
	{"GET", "/api/v1/api-permission/tree", "api_permission.read", "read"},
	{"GET", "/api/v1/api-permission", "api_permission.read", "read"},
	{"POST", "/api/v1/api-permission", "api_permission.create", "write"},
	{"PUT", "/api/v1/api-permission/:id", "api_permission.update", "write"},
	{"DELETE", "/api/v1/api-permission/:id", "api_permission.delete", "write"},
	{"GET", "/api/v1/role/:roleId/api-permissions", "api_permission.assign", "write"},
	{"POST", "/api/v1/role/:roleId/api-permissions", "api_permission.assign", "write"},
	{"GET", "/api/v1/user/:id/api-permissions", "api_permission.assign", "write"},
	{"POST", "/api/v1/user/:id/api-permissions", "api_permission.assign", "write"},

	{"POST", "/api/v1/user/import", "user.create", "write"},
	{"POST", "/api/v1/user/page", "user.read", "read"},
	{"DELETE", "/api/v1/user/batch", "user.delete", "write"},
	{"PUT", "/api/v1/user/:id/password", "user.update", "write"},
	{"POST", "/api/v1/user", "user.create", "write"},
	{"PUT", "/api/v1/user/:id", "user.update", "write"},
	{"GET", "/api/v1/user/:id", "user.read", "read"},
	{"DELETE", "/api/v1/user/:id", "user.delete", "write"},

	{"POST", "/api/v1/role/page", "role.read", "read"},
	{"POST", "/api/v1/role/assign", "role.assign", "write"},
	{"DELETE", "/api/v1/role/remove", "role.assign", "write"},
	{"GET", "/api/v1/role/user", "role.read", "read"},
	{"GET", "/api/v1/role/:roleId/users", "role.read", "read"},
	{"POST", "/api/v1/role/:roleId/users", "role.assign", "write"},
	{"DELETE", "/api/v1/role/:roleId/users", "role.assign", "write"},
	{"GET", "/api/v1/role/:roleId/menus", "role.read", "read"},
	{"POST", "/api/v1/role/:roleId/menus", "role.update", "write"},
	{"POST", "/api/v1/role", "role.create", "write"},
	{"PUT", "/api/v1/role/:roleId", "role.update", "write"},
	{"GET", "/api/v1/role/:roleId", "role.read", "read"},
	{"DELETE", "/api/v1/role/:roleId", "role.delete", "write"},

	{"GET", "/api/v1/menu/tree", "menu.read", "read"},
	{"GET", "/api/v1/menu", "menu.read", "read"},
	{"GET", "/api/v1/menu/:id", "menu.read", "read"},
	{"POST", "/api/v1/menu", "menu.create", "write"},
	{"PUT", "/api/v1/menu/:id", "menu.update", "write"},
	{"DELETE", "/api/v1/menu/:id", "menu.delete", "write"},

	{"POST", "/api/v1/org/page", "org.read", "read"},
	{"GET", "/api/v1/org/tree", "org.read", "read"},
	{"DELETE", "/api/v1/org/batch", "org.delete", "write"},
	{"POST", "/api/v1/org", "org.create", "write"},
	{"PUT", "/api/v1/org/:id", "org.update", "write"},
	{"GET", "/api/v1/org/:id", "org.read", "read"},
	{"DELETE", "/api/v1/org/:id", "org.delete", "write"},

	{"POST", "/api/v1/dict/page", "dict.read", "read"},
	{"GET", "/api/v1/dict/type", "dict.read", "read"},
	{"GET", "/api/v1/dict/label", "dict.read", "read"},
	{"DELETE", "/api/v1/dict/batch", "dict.delete", "write"},
	{"POST", "/api/v1/dict", "dict.create", "write"},
	{"PUT", "/api/v1/dict/:id", "dict.update", "write"},
	{"GET", "/api/v1/dict/:id", "dict.read", "read"},
	{"DELETE", "/api/v1/dict/:id", "dict.delete", "write"},

	{"POST", "/api/v1/config/page", "config.read", "read"},
	{"GET", "/api/v1/config/code", "config.read", "read"},
	{"GET", "/api/v1/config/data", "config.read", "read"},
	{"DELETE", "/api/v1/config/batch", "config.delete", "write"},
	{"POST", "/api/v1/config", "config.create", "write"},
	{"PUT", "/api/v1/config/:id", "config.update", "write"},
	{"GET", "/api/v1/config/:id", "config.read", "read"},
	{"DELETE", "/api/v1/config/:id", "config.delete", "write"},

	{"POST", "/api/v1/loginLog/page", "login_log.read", "read"},
	{"DELETE", "/api/v1/loginLog/batch", "login_log.delete", "write"},
	{"POST", "/api/v1/loginLog/clean", "login_log.delete", "write"},
	{"POST", "/api/v1/loginLog", "login_log.create", "write"},
	{"PUT", "/api/v1/loginLog/:id", "login_log.update", "write"},
	{"GET", "/api/v1/loginLog/:id", "login_log.read", "read"},
	{"DELETE", "/api/v1/loginLog/:id", "login_log.delete", "write"},

	{"POST", "/api/v1/operLog/page", "oper_log.read", "read"},
	{"DELETE", "/api/v1/operLog/batch", "oper_log.delete", "write"},
	{"POST", "/api/v1/operLog/clean", "oper_log.delete", "write"},
	{"POST", "/api/v1/operLog", "oper_log.create", "write"},
	{"PUT", "/api/v1/operLog/:id", "oper_log.update", "write"},
	{"GET", "/api/v1/operLog/:id", "oper_log.read", "read"},
	{"DELETE", "/api/v1/operLog/:id", "oper_log.delete", "write"},

	{"POST", "/api/v1/attachment/upload-file", "attachment.upload", "write"},
	{"POST", "/api/v1/attachment/:attachmentId/bind", "attachment.bind", "write"},
	{"GET", "/api/v1/attachment/business", "attachment.read", "read"},
	{"POST", "/api/v1/attachment/page", "attachment.read", "read"},
	{"GET", "/api/v1/attachment/:attachmentId/download", "attachment.download", "write"},
	{"GET", "/api/v1/attachment/:attachmentId/url", "attachment.read", "read"},
	{"GET", "/api/v1/attachment/:attachmentId", "attachment.read", "read"},
	{"DELETE", "/api/v1/attachment/:attachmentId", "attachment.delete", "write"},
}

type Targets struct {
	IAM      string
	SYS      string
	Resource string
	Realtime string
}

func (t Targets) targetForPath(path string) string {
	if path == "/realtime/websocket" {
		return t.Realtime
	}
	if hasPathPrefix(path, "/api/v1/attachment") {
		return t.Resource
	}
	for _, prefix := range []string{"/api/v1/dict", "/api/v1/config", "/api/v1/loginLog", "/api/v1/operLog"} {
		if hasPathPrefix(path, prefix) {
			return t.SYS
		}
	}
	for _, prefix := range []string{
		"/login", "/logout", "/auth", "/captcha", "/resource/sms/code",
		"/api/v1/user", "/api/v1/role", "/api/v1/menu", "/api/v1/api-permission", "/api/v1/org",
	} {
		if hasPathPrefix(path, prefix) {
			return t.IAM
		}
	}
	return ""
}

func hasPathPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func requiredPermission(method, path string) *permissionRule {
	for i := range permissionRules {
		rule := &permissionRules[i]
		if rule.method == method && matchRoute(rule.path, path) {
			return rule
		}
	}
	return nil
}

func isPermissionExempt(method, path string) bool {
	return method == "POST" && path == "/api/v1/user/password/change" ||
		method == "GET" && path == "/api/v1/menu/user/tree"
}

func matchRoute(pattern, requestPath string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			if pathParts[i] == "" {
				return false
			}
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}
