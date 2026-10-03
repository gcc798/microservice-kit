package middleware

import (
	"net/http"
	"strings"
)

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposeHeaders  []string
}

func CORS(cfg CORSConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" && originAllowed(origin, cfg.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			} else if origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			SetCORSHeaders(w.Header(), cfg)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	}
}

func SetCORSHeaders(header http.Header, cfg CORSConfig) {
	allowedMethods := strings.Join(defaultStrings(cfg.AllowedMethods, []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
	}), ", ")
	allowedHeaders := strings.Join(defaultStrings(cfg.AllowedHeaders, []string{
		"Authorization",
		"Content-Type",
		"clientId",
		"clientid",
		"token",
		"Token",
		"X-Requested-With",
		"Origin",
		"X-CSRF-Token",
		"AccessToken",
		"Range",
	}), ", ")
	exposeHeaders := strings.Join(defaultStrings(cfg.ExposeHeaders, []string{
		"Content-Length",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Headers",
		"Content-Disposition",
	}), ", ")

	header.Set("Access-Control-Allow-Methods", allowedMethods)
	header.Set("Access-Control-Allow-Headers", allowedHeaders)
	header.Set("Access-Control-Allow-Credentials", "true")
	header.Set("Access-Control-Expose-Headers", exposeHeaders)
	header.Set("Access-Control-Max-Age", "86400")
}

func originAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		item = strings.TrimSpace(item)
		if item == "*" || item == origin {
			return true
		}
	}
	return false
}

func defaultStrings(values, defaults []string) []string {
	if len(values) == 0 {
		return defaults
	}
	return values
}
