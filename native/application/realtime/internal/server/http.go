package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	iamv1 "github.com/gcc798/microservice-kit/internal/api/iam/v1"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	websocketx "github.com/gcc798/microservice-kit/internal/platform/websocket"
	"github.com/labstack/echo/v5"
	goredis "github.com/redis/go-redis/v9"
)

const WebSocketPath = "/realtime/websocket"

type HTTPOptions struct {
	Port                int
	CORS                bool
	TokenHeader         string
	TimeoutEnabled      bool
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	HeartbeatEnabled    bool
	MaxReadTimeouts     int
}

func NewHTTP(opts HTTPOptions, security iamv1.API, client *goredis.Client, relay ReadyRelay, hub *websocketx.Hub, log logging.Logger) *http.Server {
	e := echo.New()
	e.GET("/health", healthHandler(client, relay))
	e.GET("/health/ready", healthHandler(client, relay))
	e.GET("/health/live", func(c *echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "alive"}) })
	e.GET("/health/startup", healthHandler(client, relay))

	wsHandler := websocketx.NewHandler(hub, log, opts.CORS)
	if opts.TimeoutEnabled {
		wsHandler.ConfigureTimeouts(time.Duration(opts.ReadTimeoutSeconds)*time.Second, time.Duration(opts.WriteTimeoutSeconds)*time.Second)
	}
	if opts.HeartbeatEnabled {
		wsHandler.RegisterHeartbeatMessageBuilder(opts.MaxReadTimeouts, func(*websocketx.Client) any { return map[string]string{"type": "ping"} })
	}
	e.GET(WebSocketPath, authenticatedWebSocketHandler(security, opts.TokenHeader, wsHandler))

	return &http.Server{Addr: ":" + strconv.Itoa(opts.Port), Handler: e, ReadHeaderTimeout: 10 * time.Second, MaxHeaderBytes: 1 << 20}
}

type ReadyRelay interface{ Ready() bool }

func authenticatedWebSocketHandler(security iamv1.API, tokenHeader string, handler *websocketx.Handler) echo.HandlerFunc {
	return func(c *echo.Context) error {
		token := c.Request().Header.Get(tokenHeader)
		if token == "" {
			token = c.QueryParam(tokenHeader)
		}
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "未登录"})
		}
		claims, err := security.ValidateAccessToken(c.Request().Context(), token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "未登录或登录已失效"})
		}
		clientID := c.Request().Header.Get("clientid")
		if clientID == "" {
			clientID = c.QueryParam("clientid")
		}
		if clientID != "" && clientID != claims.ClientId {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "客户端ID与Token不匹配"})
		}
		c.Set("userId", claims.UserId)
		handler.ServeWs(c)
		return nil
	}
}

func healthHandler(client *goredis.Client, relay ReadyRelay) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if err := client.Ping(c.Request().Context()).Err(); err != nil || !relay.Ready() {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	}
}
