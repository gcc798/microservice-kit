package config

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type testConfig struct {
	Server   Server          `mapstructure:"server"`
	GRPC     Server          `mapstructure:"grpc"`
	Registry Registry        `mapstructure:"registry"`
	Service  ServiceEndpoint `mapstructure:"service"`
	Database Database        `mapstructure:"database"`
	Redis    Redis           `mapstructure:"redis"`
	JWT      JWT             `mapstructure:"jwt"`
}

func TestServiceConfigsContainOnlyOwnedSections(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve config test path")
	}
	nativeDir := filepath.Join(filepath.Dir(sourceFile), "..", "..")
	services := []struct {
		name     string
		service  Service
		sections []string
	}{
		{name: "gateway", service: ServiceGateway, sections: []string{"cors", "gateway", "registry", "server", "service"}},
		{name: "iam", service: ServiceIAM, sections: []string{"auth", "cors", "database", "grpc", "jwt", "redis", "registry", "server", "service"}},
		{name: "sys", service: ServiceSystem, sections: []string{"auth", "cors", "database", "grpc", "redis", "registry", "server", "service", "workers"}},
		{name: "resource", service: ServiceResource, sections: []string{"auth", "cors", "database", "grpc", "redis", "registry", "server", "service", "storage", "workers"}},
		{name: "realtime", service: ServiceRealtime, sections: []string{"auth", "cors", "grpc", "redis", "registry", "server", "service", "websocket"}},
	}
	for _, tt := range services {
		t.Run(tt.name, func(t *testing.T) {
			for _, profile := range []string{"dev", "prod"} {
				t.Run(profile, func(t *testing.T) {
					examplePath := filepath.Join(nativeDir, "application", tt.name, "conf.example.yaml")
					data, err := os.ReadFile(examplePath)
					if err != nil {
						t.Fatal(err)
					}
					var document map[string]any
					if err := yaml.Unmarshal(data, &document); err != nil {
						t.Fatal(err)
					}
					sections := make([]string, 0, len(document))
					for section := range document {
						sections = append(sections, section)
					}
					slices.Sort(sections)
					if !slices.Equal(sections, tt.sections) {
						t.Fatalf("%s sections = %v, want %v", examplePath, sections, tt.sections)
					}
					dir := t.TempDir()
					if err := os.WriteFile(filepath.Join(dir, "conf."+profile+".yaml"), data, 0o600); err != nil {
						t.Fatal(err)
					}
					t.Setenv(AppEnvVar, profile)
					var cfg testConfig
					_, _, err = LoadInto(dir, tt.service, &cfg)
					if err != nil {
						t.Fatalf("Load(%s/%s) error = %v", tt.name, profile, err)
					}
					if tt.service != ServiceGateway && tt.service != ServiceRealtime && cfg.Database.DSN == "" {
						t.Fatalf("%s config is missing database.dsn", tt.name)
					}
					if tt.service == ServiceIAM && (cfg.Redis.Addr == "" || cfg.JWT.Secret == "") {
						t.Fatal("IAM config contains an empty required value")
					}
					if tt.service == ServiceResource && cfg.Redis.Addr == "" {
						t.Fatal("Resource config contains an empty required Redis value")
					}
				})
			}
		})
	}
}

func TestLoadEnvironmentOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "conf.dev.yaml")
	content := []byte(`
server: { port: 9010 }
grpc: { port: 9110 }
registry: { driver: consul, address: "http://consul:8500", prefix: "/microservice-kit/services" }
service: { id: "", advertiseHost: "127.0.0.1" }
database: { dsn: yaml-dsn, maxOpenConns: 100, maxIdleConns: 10, connMaxLifetimeMinutes: 60, slowThreshold: 500 }
redis: { addr: "yaml-redis:6379", password: "", db: 0 }
jwt: { secret: yaml-secret-that-is-long-enough-for-tests, expire: 7200 }
auth: { tokenHeader: Authorization, allowConcurrent: false }
cors: { enabled: false }
websocket: { enabled: true, timeoutEnabled: true, readTimeoutSeconds: 60, writeTimeoutSeconds: 10, heartbeatEnabled: true, maxReadTimeouts: 3 }
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(AppEnvVar, "dev")
	t.Setenv("MS_K_SERVER_PORT", "8080")
	t.Setenv("MS_K_SERVICE_ID", "iam-explicit")
	t.Setenv("MS_K_DATABASE_DSN", "environment-dsn")
	t.Setenv("MS_K_REDIS_ADDR", "environment-redis:6379")
	t.Setenv("MS_K_JWT_SECRET", "environment-secret-with-at-least-32-characters")

	var cfg testConfig
	_, _, err := LoadInto(dir, ServiceIAM, &cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Service.ID != "iam-explicit" {
		t.Fatalf("Service.ID = %q, want explicit environment value", cfg.Service.ID)
	}
	if cfg.Database.DSN != "environment-dsn" {
		t.Fatalf("Database.DSN = %q", cfg.Database.DSN)
	}
	if cfg.Redis.Addr != "environment-redis:6379" {
		t.Fatalf("Redis.Addr = %q", cfg.Redis.Addr)
	}
	if cfg.JWT.Secret != "environment-secret-with-at-least-32-characters" {
		t.Fatalf("JWT.Secret was not overridden by the environment")
	}
}

func TestAtomicConfigValidation(t *testing.T) {
	t.Parallel()

	if err := (Server{Port: 9009}).Validate("server"); err != nil {
		t.Fatal(err)
	}
	if err := (Registry{Driver: "consul", Address: "http://consul:8500"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Database{DSN: "postgres-dsn", MaxOpenConns: 100, MaxIdleConns: 10, ConnMaxLifetimeMinutes: 60, SlowThreshold: 500}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (CORS{Enabled: true}).Validate("prod"); err == nil {
		t.Fatal("prod config with CORS enabled was accepted")
	}
	if err := (JWT{Secret: "short", Expire: 7200}).Validate(); err == nil {
		t.Fatal("short JWT secret was accepted")
	}
}

func TestCurrentEnvMustBeExplicit(t *testing.T) {
	t.Setenv(AppEnvVar, "")
	if got := CurrentEnv(); got != "" {
		t.Fatalf("CurrentEnv() = %q, want empty", got)
	}
	var cfg testConfig
	if _, _, err := LoadInto(t.TempDir(), ServiceIAM, &cfg); err == nil {
		t.Fatal("Load() accepted a missing MS_K_APP_ENV")
	}
}

func TestLoadRejectsImplicitConfiguration(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "conf.dev.yaml"), []byte("server:\n  port: 9009\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(AppEnvVar, "dev")
	var cfg testConfig
	_, _, err := LoadInto(dir, ServiceGateway, &cfg)
	if err == nil || !strings.Contains(err.Error(), "must be explicitly configured") {
		t.Fatalf("Load() error = %v, want explicit configuration error", err)
	}
}

func TestLoadRejectsUnknownProfile(t *testing.T) {
	t.Setenv(AppEnvVar, "staging")
	var cfg testConfig
	if _, _, err := LoadInto(t.TempDir(), ServiceIAM, &cfg); err == nil {
		t.Fatal("Load() accepted unsupported profile")
	}
}

func TestEnvironmentName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"database.maxOpenConns": "MS_K_DATABASE_MAX_OPEN_CONNS",
		"storage.useSSL":        "MS_K_STORAGE_USE_SSL",
	}
	for key, want := range tests {
		if got := environmentName(key); got != want {
			t.Errorf("environmentName(%q) = %q, want %q", key, got, want)
		}
	}
}
