package server

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gcc798/microservice-kit/application/gateway/internal/openapi"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
	"github.com/gcc798/microservice-kit/common/auth"
	registry "github.com/gcc798/microservice-kit/internal/registry"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
)

type Server struct {
	routes     atomic.Pointer[routeTable]
	subscriber *discov.Subscriber
	iam        iamservice.IamService
	header     string
	audit      *operationWriter
}

func New(config registry.HTTPConfig, iam iamservice.IamService, system sysservice.SysService, tokenHeader string) (*Server, error) {
	subscriber, err := registry.Subscribe(config)
	if err != nil {
		return nil, err
	}
	s := &Server{subscriber: subscriber, iam: iam, header: tokenHeader, audit: newOperationWriter(system)}
	subscriber.AddListener(s.reload)
	s.reload()
	return s, nil
}

func (s *Server) Close() {
	s.audit.Stop()
	s.subscriber.Close()
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.serveHTTP)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/swagger/doc.json" {
		openapi.Document(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/swagger/") {
		httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")).ServeHTTP(w, r)
		return
	}
	switch r.URL.Path {
	case "/health/live":
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
		return
	case "/health", "/health/ready", "/health/startup":
		if err := s.ready(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready", "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
		return
	}
	table := s.routes.Load()
	target := table.match(r.Method, r.URL.Path)
	if target == nil || len(target.endpoints) == 0 {
		http.NotFound(w, r)
		return
	}
	started := time.Now()
	auditBody := readAuditBody(r)
	auditWriter := &auditResponseWriter{ResponseWriter: w}
	w = auditWriter
	var operatorName, deviceType string
	defer func() {
		s.audit.Write(newOperationLog(r, auditBody, auditWriter.Status(), started, operatorName, deviceType))
	}()
	if !isPublic(r.URL.Path) {
		claims, err := s.iam.ValidateAccessToken(r.Context(), &iamservice.ValidateAccessTokenReq{Token: auth.TokenFromRequest(r, s.header)})
		if err != nil {
			writeError(w, http.StatusUnauthorized, "未登录或登录已失效")
			return
		}
		clientID := strings.TrimSpace(r.Header.Get("clientid"))
		if clientID == "" {
			clientID = strings.TrimSpace(r.URL.Query().Get("clientid"))
		}
		if clientID != "" && claims.ClientId != "" && clientID != claims.ClientId {
			writeError(w, http.StatusUnauthorized, "客户端ID与Token不匹配")
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
				writeError(w, http.StatusInternalServerError, "权限检查失败")
				return
			}
			if !result.Allowed {
				writeError(w, http.StatusForbidden, "无权限访问")
				return
			}
		}
		r.Header.Set("X-Microservice-Kit-User-ID", strconv.FormatInt(claims.UserId, 10))
		operatorName = strconv.FormatInt(claims.UserId, 10) + "-" + claims.UserName
		deviceType = claims.DeviceType
	}

	endpoint := target.endpoints[(target.next.Add(1)-1)%uint64(len(target.endpoints))]
	proxy := httputil.NewSingleHostReverseProxy(endpoint)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		logx.Errorf("proxy %s to %s: %v", r.URL.Path, endpoint, err)
		writeError(w, http.StatusBadGateway, "服务调用失败")
	}
	proxy.ServeHTTP(w, r)
}

func isPublic(path string) bool {
	return path == "/login" || path == "/auth/refresh" || path == "/resource/sms/code" || strings.HasPrefix(path, "/captcha/")
}

func (s *Server) ready() error {
	return s.routes.Load().ready()
}

func (s *Server) reload() {
	instances, err := registry.Decode(s.subscriber.Values())
	if err != nil {
		logx.Errorf("decode HTTP routes from etcd: %v", err)
		return
	}
	table, err := buildRouteTable(instances)
	if err != nil {
		logx.Errorf("build HTTP routes from etcd: %v", err)
		return
	}
	s.routes.Store(table)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"code": status, "msg": message})
}
