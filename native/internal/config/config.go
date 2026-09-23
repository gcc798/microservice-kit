package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/viper"
)

// AppEnvVar 是运行环境变量名。
const AppEnvVar = "MS_K_APP_ENV"

// Service 标识一个可独立部署的服务。
type Service string

const (
	ServiceGateway     Service = "gateway"
	ServiceIAM         Service = "iam"
	ServiceSystem      Service = "sys"
	ServiceResource    Service = "resource"
	ServiceRealtime    Service = "realtime"
	ServiceUserManager Service = "usermgr"
)

// Server 描述 HTTP 或 gRPC 服务的监听配置。
type Server struct {
	Port        int    `mapstructure:"port"`        // 监听端口。
	TLSCertFile string `mapstructure:"tlsCertFile"` // TLS 证书路径。
	TLSKeyFile  string `mapstructure:"tlsKeyFile"`  // TLS 私钥路径。
}

// Registry 描述服务注册中心连接配置。
type Registry struct {
	Driver    string `mapstructure:"driver"`    // 注册中心类型。
	Address   string `mapstructure:"address"`   // 注册中心地址。
	Prefix    string `mapstructure:"prefix"`    // 服务键前缀。
	Namespace string `mapstructure:"namespace"` // 命名空间。
	Group     string `mapstructure:"group"`     // 服务分组。
	Username  string `mapstructure:"username"`  // 认证用户名。
	Password  string `mapstructure:"password"`  // 认证密码。
}

// ServiceEndpoint 描述服务实例的标识和通告地址。
type ServiceEndpoint struct {
	ID            string `mapstructure:"id"`            // 实例唯一标识。
	AdvertiseHost string `mapstructure:"advertiseHost"` // 对外通告地址。
}

// Gateway 描述网关专属配置。
type Gateway struct {
	RateLimitPerMinute int `mapstructure:"rateLimitPerMinute"` // 每分钟请求上限。
}

// Database 描述数据库连接池配置。
type Database struct {
	DSN                    string `mapstructure:"dsn"`                    // 数据库连接串。
	MaxOpenConns           int    `mapstructure:"maxOpenConns"`           // 最大打开连接数。
	MaxIdleConns           int    `mapstructure:"maxIdleConns"`           // 最大空闲连接数。
	ConnMaxLifetimeMinutes int    `mapstructure:"connMaxLifetimeMinutes"` // 连接最大存活时间，单位为分钟。
	SlowThreshold          int    `mapstructure:"slowThreshold"`          // 慢查询阈值，单位为毫秒。
}

// Redis 描述 Redis 连接配置。
type Redis struct {
	Addr     string `mapstructure:"addr"`     // Redis 地址。
	Password string `mapstructure:"password"` // Redis 密码。
	DB       int    `mapstructure:"db"`       // Redis 逻辑库编号。
}

// JWT 描述 JSON Web Token 配置。
type JWT struct {
	Secret string `mapstructure:"secret"` // 签名密钥。
	Expire int64  `mapstructure:"expire"` // 默认有效期，单位为秒。
}

// Auth 描述 HTTP 身份认证配置。
type Auth struct {
	TokenHeader     string `mapstructure:"tokenHeader"`     // 令牌请求头名称。
	AllowConcurrent bool   `mapstructure:"allowConcurrent"` // 是否允许同一用户并发登录。
}

// CORS 描述跨域资源共享配置。
type CORS struct {
	Enabled bool `mapstructure:"enabled"` // 是否启用跨域中间件。
}

// WebSocket 描述 WebSocket 连接配置。
type WebSocket struct {
	Enabled             bool `mapstructure:"enabled"`             // 是否启用 WebSocket。
	TimeoutEnabled      bool `mapstructure:"timeoutEnabled"`      // 是否启用读写超时。
	ReadTimeoutSeconds  int  `mapstructure:"readTimeoutSeconds"`  // 读取超时时间，单位为秒。
	WriteTimeoutSeconds int  `mapstructure:"writeTimeoutSeconds"` // 写入超时时间，单位为秒。
	HeartbeatEnabled    bool `mapstructure:"heartbeatEnabled"`    // 是否启用心跳。
	MaxReadTimeouts     int  `mapstructure:"maxReadTimeouts"`     // 最大连续读取超时次数。
}

// LoadInto 加载并校验服务所需配置，再解码到服务私有配置类型。
func LoadInto(configDir string, service Service, target any) (*viper.Viper, string, error) {
	profile := CurrentEnv()
	if profile != "dev" && profile != "prod" {
		return nil, "", fmt.Errorf("%s must be dev or prod", AppEnvVar)
	}
	if service != ServiceGateway && service != ServiceIAM && service != ServiceSystem && service != ServiceResource && service != ServiceRealtime && service != ServiceUserManager {
		return nil, "", fmt.Errorf("unknown config service %q", service)
	}
	v := viper.New()
	v.SetEnvPrefix("MS_K")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := bindEnvironment(v); err != nil {
		return nil, "", err
	}
	configFileName := fmt.Sprintf("conf.%s.yaml", profile)
	foundPath, err := ResolveFilePath(configDir, configFileName)
	if err != nil {
		return nil, "", err
	}
	v.SetConfigFile(foundPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, "", fmt.Errorf("read config from %s: %w", foundPath, err)
	}
	if err := requireExplicitConfiguration(v, service); err != nil {
		return nil, "", err
	}
	if err := v.Unmarshal(target); err != nil {
		return nil, "", fmt.Errorf("decode config: %w", err)
	}
	return v, filepath.Dir(foundPath), nil
}

func requireExplicitConfiguration(v *viper.Viper, service Service) error {
	database := []string{"database.dsn", "database.maxOpenConns", "database.maxIdleConns", "database.connMaxLifetimeMinutes", "database.slowThreshold"}
	registry := []string{"registry.driver"}
	switch v.GetString("registry.driver") {
	case "consul", "nacos":
		registry = append(registry, "registry.address")
	case "etcd":
		registry = append(registry, "registry.address", "registry.prefix")
	}
	server := []string{"server.port"}
	grpc := []string{"grpc.port", "service.id", "service.advertiseHost"}
	auth := []string{"auth.tokenHeader", "cors.enabled"}

	var keys []string
	switch service {
	case ServiceGateway:
		keys = append(server, "server.tlsCertFile", "server.tlsKeyFile", "gateway.rateLimitPerMinute", "cors.enabled")
		keys = append(keys, registry...)
		keys = append(keys, "service.id", "service.advertiseHost")
	case ServiceIAM:
		keys = append(keys, server...)
		keys = append(keys, grpc...)
		keys = append(keys, registry...)
		keys = append(keys, database...)
		keys = append(keys, "redis.addr", "redis.password", "redis.db", "jwt.secret", "jwt.expire")
		keys = append(keys, auth...)
		keys = append(keys, "auth.allowConcurrent")
	case ServiceRealtime:
		keys = append(keys, server...)
		keys = append(keys, grpc...)
		keys = append(keys, registry...)
		keys = append(keys, "redis.addr", "redis.password", "redis.db")
		keys = append(keys, auth...)
		keys = append(keys, "websocket.timeoutEnabled", "websocket.readTimeoutSeconds", "websocket.writeTimeoutSeconds", "websocket.heartbeatEnabled", "websocket.maxReadTimeouts")
	case ServiceSystem:
		keys = append(keys, server...)
		keys = append(keys, grpc...)
		keys = append(keys, registry...)
		keys = append(keys, database...)
		keys = append(keys, "redis.addr", "redis.password", "redis.db")
		keys = append(keys, auth...)
	case ServiceResource:
		keys = append(keys, server...)
		keys = append(keys, grpc...)
		keys = append(keys, registry...)
		keys = append(keys, database...)
		keys = append(keys, "redis.addr", "redis.password", "redis.db")
		keys = append(keys, "storage.endpoint", "storage.accessKey", "storage.secretKey", "storage.region", "storage.bucket", "storage.useSSL")
		keys = append(keys, auth...)
	case ServiceUserManager:
		keys = []string{"database.dsn"}
	}
	for _, key := range keys {
		_, environmentSet := os.LookupEnv(environmentName(key))
		if !v.InConfig(key) && !environmentSet {
			return fmt.Errorf("%s must be explicitly configured", key)
		}
	}
	return nil
}

func bindEnvironment(v *viper.Viper) error {
	keys := []string{
		"server.port", "server.tlsCertFile", "server.tlsKeyFile",
		"grpc.port",
		"registry.driver", "registry.address", "registry.prefix", "registry.namespace", "registry.group", "registry.username", "registry.password",
		"service.id", "service.advertiseHost",
		"gateway.rateLimitPerMinute",
		"database.dsn", "database.maxOpenConns", "database.maxIdleConns", "database.connMaxLifetimeMinutes", "database.slowThreshold",
		"redis.addr", "redis.password", "redis.db",
		"jwt.secret", "jwt.expire",
		"auth.tokenHeader", "auth.allowConcurrent",
		"cors.enabled",
		"storage.endpoint", "storage.accessKey", "storage.secretKey", "storage.region", "storage.bucket", "storage.useSSL",
		"websocket.enabled", "websocket.timeoutEnabled", "websocket.readTimeoutSeconds", "websocket.writeTimeoutSeconds", "websocket.heartbeatEnabled", "websocket.maxReadTimeouts",
	}
	for _, key := range keys {
		if err := v.BindEnv(key, environmentName(key)); err != nil {
			return fmt.Errorf("bind environment for %s: %w", key, err)
		}
	}
	return nil
}

func environmentName(key string) string {
	runes := []rune(key)
	var name strings.Builder
	name.WriteString("MS_K_")
	for i, current := range runes {
		switch {
		case current == '.':
			name.WriteByte('_')
		case unicode.IsUpper(current):
			previousIsLower := i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1]))
			startsWord := i > 0 && unicode.IsUpper(runes[i-1]) && i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if previousIsLower || startsWord {
				name.WriteByte('_')
			}
			name.WriteRune(current)
		default:
			name.WriteRune(unicode.ToUpper(current))
		}
	}
	return name.String()
}

func (s Server) Validate(name string) error {
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("%s.port must be between 1 and 65535", name)
	}
	return nil
}

func (r Registry) Validate() error {
	switch r.Driver {
	case "consul", "nacos":
		if r.Address == "" {
			return fmt.Errorf("registry.address is required for %s", r.Driver)
		}
	case "etcd":
		if r.Address == "" || r.Prefix == "" {
			return fmt.Errorf("registry.address and registry.prefix are required for etcd")
		}
	case "inprocess":
	default:
		return fmt.Errorf("unsupported registry driver %q", r.Driver)
	}
	return nil
}

func (s ServiceEndpoint) Validate() error {
	if s.AdvertiseHost == "" {
		return fmt.Errorf("service.advertiseHost is required")
	}
	return nil
}

func (d Database) Validate() error {
	if d.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if d.MaxOpenConns < 1 || d.MaxIdleConns < 0 || d.MaxIdleConns > d.MaxOpenConns {
		return fmt.Errorf("database connection limits are invalid")
	}
	if d.ConnMaxLifetimeMinutes < 1 || d.SlowThreshold < 1 {
		return fmt.Errorf("database lifetimes and thresholds must be positive")
	}
	return nil
}

func (r Redis) Validate() error {
	if r.Addr == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if r.DB < 0 {
		return fmt.Errorf("redis.db cannot be negative")
	}
	return nil
}

func (j JWT) Validate() error {
	if len(j.Secret) < 32 {
		return fmt.Errorf("jwt.secret must contain at least 32 characters")
	}
	if j.Expire < 1 {
		return fmt.Errorf("jwt.expire must be positive")
	}
	return nil
}

func (a Auth) Validate() error {
	if a.TokenHeader == "" {
		return fmt.Errorf("auth.tokenHeader is required")
	}
	return nil
}

func (c CORS) Validate(profile string) error {
	if profile == "prod" && c.Enabled {
		return fmt.Errorf("cors must be disabled in prod")
	}
	return nil
}

func (w WebSocket) Validate() error {
	if w.ReadTimeoutSeconds < 1 || w.WriteTimeoutSeconds < 1 || w.MaxReadTimeouts < 1 {
		return fmt.Errorf("websocket timeouts and retry limit must be positive")
	}
	return nil
}

func (g Gateway) Validate() error {
	if g.RateLimitPerMinute < 0 {
		return fmt.Errorf("gateway.rateLimitPerMinute cannot be negative")
	}
	return nil
}

func CurrentEnv() string {
	return os.Getenv(AppEnvVar)
}

func ResolveFilePath(configDir, fileName string) (string, error) {
	if strings.TrimSpace(configDir) == "" {
		return "", errors.New("config directory is required")
	}
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}

	var candidates []string
	if filepath.IsAbs(configDir) {
		candidates = []string{filepath.Join(configDir, fileName)}
	} else {
		candidates = []string{
			filepath.Join(workDir, configDir, fileName),
			filepath.Join(filepath.Dir(executable), fileName),
		}
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("config file not found: tried %v", candidates)
}
