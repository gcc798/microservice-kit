package requestmeta

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type clientInfo struct {
	ip        string
	userAgent string
}

type contextKey struct{}

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := clientInfo{ip: clientIP(r), userAgent: r.UserAgent()}
		next(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, info)))
	}
}

func ClientInfo(ctx context.Context) (ip, userAgent string) {
	info, _ := ctx.Value(contextKey{}).(clientInfo)
	return info.ip, info.userAgent
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
