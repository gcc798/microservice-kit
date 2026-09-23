package config

import (
	"fmt"
	"os"
	"strings"

	shared "github.com/gcc798/microservice-kit/internal/config"
	"github.com/spf13/viper"
)

type Config struct {
	AppDir   string                 `mapstructure:"-"`
	Server   shared.Server          `mapstructure:"server"`
	GRPC     shared.Server          `mapstructure:"grpc"`
	Registry shared.Registry        `mapstructure:"registry"`
	Service  shared.ServiceEndpoint `mapstructure:"service"`
	Database shared.Database        `mapstructure:"database"`
	Redis    shared.Redis           `mapstructure:"redis"`
	JWT      shared.JWT             `mapstructure:"jwt"`
	Auth     shared.Auth            `mapstructure:"auth"`
	CORS     shared.CORS            `mapstructure:"cors"`
}

func Load(configDir string) (*Config, *viper.Viper, error) {
	cfg := new(Config)
	v, appDir, err := shared.LoadInto(configDir, shared.ServiceIAM, cfg)
	if err != nil {
		return nil, nil, err
	}
	cfg.AppDir = appDir
	cfg.Service.ID = strings.TrimSpace(cfg.Service.ID)
	if cfg.Service.ID == "" {
		host, err := os.Hostname()
		if err != nil {
			return nil, nil, fmt.Errorf("resolve service instance ID: %w", err)
		}
		cfg.Service.ID = string(shared.ServiceIAM) + "-" + host
	}
	for _, validate := range []func() error{
		func() error { return cfg.Server.Validate("server") },
		func() error { return cfg.GRPC.Validate("grpc") },
		cfg.Registry.Validate, cfg.Service.Validate, cfg.Database.Validate, cfg.Redis.Validate,
		cfg.JWT.Validate, cfg.Auth.Validate, func() error { return cfg.CORS.Validate(shared.CurrentEnv()) },
	} {
		if err := validate(); err != nil {
			return nil, nil, err
		}
	}
	return cfg, v, nil
}
