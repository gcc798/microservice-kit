package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	operationBatchSize = 100
	operationQueueSize = 1000
	operationBodyLimit = 64 << 10
)

type operationWriter struct {
	client   sysservice.SysService
	logs     chan *sysservice.OperLogReq
	cancel   context.CancelFunc
	done     chan struct{}
	stopOnce sync.Once
}

func newOperationWriter(client sysservice.SysService) *operationWriter {
	ctx, cancel := context.WithCancel(context.Background())
	w := &operationWriter{client: client, logs: make(chan *sysservice.OperLogReq, operationQueueSize), cancel: cancel, done: make(chan struct{})}
	go w.run(ctx)
	return w
}

func (w *operationWriter) Write(entry *sysservice.OperLogReq) {
	if entry == nil {
		return
	}
	select {
	case w.logs <- entry:
	default:
		logx.Error("operation log queue is full")
	}
}

func (w *operationWriter) Stop() {
	w.stopOnce.Do(func() {
		w.cancel()
		<-w.done
	})
}

func (w *operationWriter) run(ctx context.Context) {
	defer close(w.done)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	batch := make([]*sysservice.OperLogReq, 0, operationBatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		callCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := w.client.OperLogBatchCreate(callCtx, &sysservice.OperLogBatchReq{Logs: batch})
		cancel()
		if err != nil {
			logx.Errorf("write operation logs: %v", err)
		}
		batch = batch[:0]
	}
	for {
		select {
		case entry := <-w.logs:
			batch = append(batch, entry)
			if len(batch) == operationBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			for {
				select {
				case entry := <-w.logs:
					batch = append(batch, entry)
				default:
					flush()
					return
				}
			}
		}
	}
}

type auditResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *auditResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *auditResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *auditResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *auditResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func newOperationLog(r *http.Request, body []byte, status int, started time.Time, operatorName, deviceType string) *sysservice.OperLogReq {
	if shouldSkipOperationLog(r.URL.Path) {
		return nil
	}
	if operatorName == "" {
		operatorName = "-"
	}
	if deviceType == "" {
		deviceType = "unknown"
	}
	entry := &sysservice.OperLogReq{
		Title: titleFromPath(r.URL.Path), BusinessType: businessType(r.Method, r.URL.Path),
		Method: r.Method + " " + r.URL.Path, RequestMethod: r.Method, DeviceType: deviceType,
		OperName: operatorName, OperUrl: r.URL.Path, OperIp: requestIP(r), OperParam: operationParams(r, body),
		Status: "0", CostTime: time.Since(started).Milliseconds(), UserAgent: r.UserAgent(), OperTimeUnixMilli: started.UnixMilli(),
	}
	if status >= http.StatusBadRequest {
		entry.Status = "1"
		entry.ErrorMsg = http.StatusText(status)
	}
	return entry
}

func readAuditBody(r *http.Request) []byte {
	if r.Body == nil || r.Method == http.MethodGet || r.Method == http.MethodDelete || strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, operationBodyLimit+1))
	if err != nil {
		return nil
	}
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), r.Body))
	if len(body) > operationBodyLimit {
		return nil
	}
	return body
}

func operationParams(r *http.Request, body []byte) string {
	if len(body) == 0 {
		return truncate(sanitizeQuery(r.URL.Query()).Encode(), 2000)
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "[请求体不是有效 JSON]"
	}
	redact(value)
	encoded, err := json.Marshal(value)
	if err != nil {
		return "[请求体脱敏失败]"
	}
	return truncate(string(encoded), 2000)
}

func redact(value any) {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if sensitiveField(key) {
				current[key] = "***"
			} else {
				redact(child)
			}
		}
	case []any:
		for _, child := range current {
			redact(child)
		}
	}
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func shouldSkipOperationLog(path string) bool {
	return path == "/health" || strings.HasPrefix(path, "/health/") || strings.HasPrefix(path, "/api/v1/operLog")
}

func titleFromPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(path, "/api/v1"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "未知操作"
	}
	titles := map[string]string{"user": "用户", "role": "角色", "menu": "菜单", "org": "组织", "dict": "字典", "config": "配置", "loginLog": "登录日志", "operLog": "操作日志", "attachment": "附件"}
	if title := titles[parts[0]]; title != "" {
		return title
	}
	return parts[0]
}

func businessType(method, path string) string {
	switch {
	case strings.Contains(path, "/export"):
		return "EXPORT"
	case strings.Contains(path, "/import"):
		return "IMPORT"
	case strings.Contains(path, "/grant"), strings.Contains(path, "/permission"):
		return "GRANT"
	case strings.Contains(path, "/clean"):
		return "CLEAN"
	case strings.Contains(path, "/page"), strings.Contains(path, "/list"):
		return "QUERY"
	}
	switch method {
	case http.MethodGet:
		return "QUERY"
	case http.MethodPost:
		return "CREATE"
	case http.MethodPut, http.MethodPatch:
		return "UPDATE"
	case http.MethodDelete:
		return "DELETE"
	default:
		return "OTHER"
	}
}

func sanitizeQuery(values url.Values) url.Values {
	clean := make(url.Values, len(values))
	for key, items := range values {
		if sensitiveField(key) {
			clean[key] = []string{"***"}
		} else {
			clean[key] = append([]string(nil), items...)
		}
	}
	return clean
}

func sensitiveField(field string) bool {
	normalized := strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(field))
	switch normalized {
	case "password", "oldpassword", "newpassword", "accesstoken", "refreshtoken", "token", "authorization", "secret", "clientsecret", "accesskey", "accesskeyid", "accesskeysecret", "secretkey", "secretaccesskey", "code", "smscode":
		return true
	default:
		return false
	}
}

func requestIP(r *http.Request) string {
	if value := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); value != "" {
		return value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Real-IP")); value != "" {
		return value
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
