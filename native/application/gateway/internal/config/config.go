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
	Registry shared.Registry        `mapstructure:"registry"`
	Service  shared.ServiceEndpoint `mapstructure:"service"`
	Gateway  shared.Gateway         `mapstructure:"gateway"`
	CORS     shared.CORS            `mapstructure:"cors"`
}

func Load(configDir string) (*Config, *viper.Viper, error) {
	cfg := new(Config)
	v, appDir, err := shared.LoadInto(configDir, shared.ServiceGateway, cfg)
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
		cfg.Service.ID = string(shared.ServiceGateway) + "-" + host
	}
	if err := cfg.Server.Validate("server"); err != nil {
		return nil, nil, err
	}
	if (cfg.Server.TLSCertFile == "") != (cfg.Server.TLSKeyFile == "") {
		return nil, nil, fmt.Errorf("both server.tlsCertFile and server.tlsKeyFile are required for TLS")
	}
	for _, validate := range []func() error{cfg.Registry.Validate, cfg.Service.Validate, cfg.Gateway.Validate, func() error { return cfg.CORS.Validate(shared.CurrentEnv()) }} {
		if err := validate(); err != nil {
			return nil, nil, err
		}
	}
	return cfg, v, nil
}
