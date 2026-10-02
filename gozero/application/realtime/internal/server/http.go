package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	"github.com/gcc798/microservice-kit/application/realtime/internal/hub"
	"github.com/gcc798/microservice-kit/application/realtime/internal/svc"
	"github.com/gcc798/microservice-kit/common/auth"
	"github.com/gorilla/websocket"
)

func NewHTTP(ctx *svc.ServiceContext) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	health := func(w http.ResponseWriter, r *http.Request) {
		if err := ctx.Redis.Ping(r.Context()).Err(); err != nil || !ctx.Relay.Ready() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	}
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /health/ready", health)
	mux.HandleFunc("GET /health/startup", health)
	mux.HandleFunc("GET /realtime/websocket", func(w http.ResponseWriter, r *http.Request) { serveWebSocket(ctx, w, r) })
	return &http.Server{Addr: fmt.Sprintf("%s:%d", ctx.Config.HTTP.Host, ctx.Config.HTTP.Port), Handler: mux, ReadHeaderTimeout: 10 * time.Second}
}

func serveWebSocket(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	token := auth.TokenFromRequest(r, ctx.Config.Security.TokenHeader)
	claims, err := ctx.IamRpc.ValidateAccessToken(r.Context(), &iamservice.ValidateAccessTokenReq{Token: token})
	if err != nil {
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
	conn, err := (&websocket.Upgrader{CheckOrigin: sameOrigin}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := hub.NewClient(conn)
	ctx.Hub.Add(claims.UserId, client)
	defer ctx.Hub.Remove(claims.UserId, client)
	conn.SetReadLimit(1 << 20)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if client.Write(websocket.PingMessage, nil) != nil {
					_ = client.Close()
					return
				}
			}
		}
	}()
	for client.ReadMessage() == nil {
	}
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	return err == nil && strings.EqualFold(parsed.Host, r.Host)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
