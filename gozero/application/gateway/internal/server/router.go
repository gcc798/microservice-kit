package server

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"

	registry "github.com/gcc798/microservice-kit/internal/registry"
)

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

type routeTarget struct {
	method    string
	path      string
	service   string
	endpoints []*url.URL
	next      *atomic.Uint64
}

type routeTable struct {
	targets  []*routeTarget
	services map[string]struct{}
}

func buildRouteTable(instances []registry.HTTPInstance) (*routeTable, error) {
	byRoute := make(map[string]*routeTarget)
	services := make(map[string]struct{})
	selectors := make(map[string]*atomic.Uint64)
	for _, instance := range instances {
		endpoint, err := url.Parse(instance.Endpoint)
		if err != nil {
			return nil, err
		}
		services[instance.Service] = struct{}{}
		for _, route := range instance.Routes {
			key := strings.ToUpper(route.Method) + " " + canonicalPath(route.Path)
			target, ok := byRoute[key]
			if !ok {
				selector := selectors[instance.Service]
				if selector == nil {
					selector = &atomic.Uint64{}
					selectors[instance.Service] = selector
				}
				byRoute[key] = &routeTarget{method: strings.ToUpper(route.Method), path: route.Path, service: instance.Service, endpoints: []*url.URL{endpoint}, next: selector}
				continue
			}
			if target.service != instance.Service {
				return nil, fmt.Errorf("HTTP route %s belongs to both %s and %s", key, target.service, instance.Service)
			}
			if !hasEndpoint(target.endpoints, endpoint.String()) {
				target.endpoints = append(target.endpoints, endpoint)
			}
		}
	}

	targets := make([]*routeTarget, 0, len(byRoute))
	for _, target := range byRoute {
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool {
		return routeLess(targets[i].path, targets[j].path)
	})
	return &routeTable{targets: targets, services: services}, nil
}

func (t *routeTable) match(method, path string) *routeTarget {
	if t == nil {
		return nil
	}
	for _, target := range t.targets {
		if target.method == method && matchRoute(target.path, path) {
			return target
		}
	}
	return nil
}

func (t *routeTable) ready() error {
	if t == nil {
		return errors.New("HTTP routes have not been discovered")
	}
	for _, service := range []string{"iam-api", "sys-api", "resource-api", "realtime"} {
		if _, ok := t.services[service]; !ok {
			return fmt.Errorf("HTTP service %s has not been discovered", service)
		}
	}
	return nil
}

func canonicalPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = ":"
		}
	}
	return strings.Join(parts, "/")
}

func routeLess(left, right string) bool {
	leftParts := strings.Split(strings.TrimPrefix(left, "/"), "/")
	rightParts := strings.Split(strings.TrimPrefix(right, "/"), "/")
	for i := 0; i < min(len(leftParts), len(rightParts)); i++ {
		leftRank, rightRank := routeSegmentRank(leftParts[i]), routeSegmentRank(rightParts[i])
		if leftRank != rightRank {
			return leftRank > rightRank
		}
	}
	if len(leftParts) != len(rightParts) {
		return len(leftParts) > len(rightParts)
	}
	return left < right
}

func routeSegmentRank(segment string) int {
	if segment == "*" {
		return 0
	}
	if strings.HasPrefix(segment, ":") {
		return 1
	}
	return 2
}

func hasEndpoint(endpoints []*url.URL, endpoint string) bool {
	for _, existing := range endpoints {
		if existing.String() == endpoint {
			return true
		}
	}
	return false
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
	for i := range patternParts {
		if patternParts[i] == "*" {
			return true
		}
		if i >= len(pathParts) {
			return false
		}
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
	return len(patternParts) == len(pathParts)
}
