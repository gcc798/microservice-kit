package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/gcc798/microservice-kit/common/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	targets Targets
	proxies map[string]*httputil.ReverseProxy
	client  *http.Client
	iam     iamservice.IamService
	header  string
}

func New(targets Targets, iam iamservice.IamService, tokenHeader string) (*Server, error) {
	proxies := make(map[string]*httputil.ReverseProxy, 4)
	for _, rawURL := range []string{targets.IAM, targets.SYS, targets.Resource, targets.Realtime} {
		target, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		proxies[rawURL] = httputil.NewSingleHostReverseProxy(target)
	}
	return &Server{targets: targets, proxies: proxies, client: &http.Client{Timeout: 2 * time.Second}, iam: iam, header: tokenHeader}, nil
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.serveHTTP)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health/live":
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
		return
	case "/health", "/health/ready", "/health/startup":
		if err := s.ready(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready", "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
		return
	}
	target := s.targets.targetForPath(r.URL.Path)
	if target == "" {
		http.NotFound(w, r)
		return
	}
	if !isPublic(r.URL.Path) {
		claims, err := s.iam.ValidateAccessToken(r.Context(), &iamservice.ValidateAccessTokenReq{Token: auth.TokenFromRequest(r, s.header)})
		if err != nil {
			if status.Code(err) != codes.Unauthenticated {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "认证服务暂不可用"})
				return
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "未登录或登录已失效"})
			return
		}
		clientID := strings.TrimSpace(r.Header.Get("clientid"))
		if clientID == "" {
			clientID = strings.TrimSpace(r.URL.Query().Get("clientid"))
		}
		if clientID != "" && claims.ClientId != "" && clientID != claims.ClientId {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "客户端ID与Token不匹配"})
			return
		}
		permission := requiredPermission(r.Method, r.URL.Path)
		if permission == nil && hasPathPrefix(r.URL.Path, "/api/v1") && !isPermissionExempt(r.Method, r.URL.Path) {
			http.NotFound(w, r)
			return
		}
		if permission != nil {
			result, err := s.iam.CheckPermission(r.Context(), &iamservice.CheckPermissionReq{
				UserId: claims.UserId, Resource: permission.resource, Action: permission.action,
			})
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "权限服务暂不可用"})
				return
			}
			if !result.Allowed {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权限访问"})
				return
			}
		}
		r.Header.Set("X-Microservice-Kit-User-ID", strconv.FormatInt(claims.UserId, 10))
	}

	s.proxies[target].ServeHTTP(w, r)
}

func isPublic(path string) bool {
	for _, prefix := range []string{"/login", "/logout", "/auth", "/captcha", "/resource/sms/code"} {
		if hasPathPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func (s *Server) ready(ctx context.Context) error {
	for _, target := range []string{s.targets.IAM, s.targets.SYS, s.targets.Resource, s.targets.Realtime} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target+"/health/ready", nil)
		if err != nil {
			return err
		}
		resp, err := s.client.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return &statusError{target: target, status: resp.StatusCode}
		}
	}
	return nil
}

type statusError struct {
	target string
	status int
}

func (e *statusError) Error() string {
	return e.target + " returned " + strconv.Itoa(e.status)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
