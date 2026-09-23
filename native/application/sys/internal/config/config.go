package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	shared "github.com/gcc798/microservice-kit/internal/config"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type Workers struct {
	LogCleanup LogCleanup `mapstructure:"logCleanup"`
}

type LogCleanup struct {
	Enabled        bool   `mapstructure:"enabled"`
	Cron           string `mapstructure:"cron"`
	RetentionDays  int32  `mapstructure:"retentionDays"`
	LockTTLMinutes int    `mapstructure:"lockTTLMinutes"`
}

type Config struct {
	AppDir   string                 `mapstructure:"-"`
	Server   shared.Server          `mapstructure:"server"`
	GRPC     shared.Server          `mapstructure:"grpc"`
	Registry shared.Registry        `mapstructure:"registry"`
	Service  shared.ServiceEndpoint `mapstructure:"service"`
	Database shared.Database        `mapstructure:"database"`
	Redis    shared.Redis           `mapstructure:"redis"`
	Auth     shared.Auth            `mapstructure:"auth"`
	CORS     shared.CORS            `mapstructure:"cors"`
	Workers  Workers                `mapstructure:"workers"`
}

func Load(configDir string) (*Config, *viper.Viper, error) {
	cfg := new(Config)
	v, appDir, err := shared.LoadInto(configDir, shared.ServiceSystem, cfg)
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
		cfg.Service.ID = string(shared.ServiceSystem) + "-" + host
	}
	for _, validate := range []func() error{
		func() error { return cfg.Server.Validate("server") },
		func() error { return cfg.GRPC.Validate("grpc") },
		cfg.Registry.Validate, cfg.Service.Validate, cfg.Database.Validate, cfg.Redis.Validate,
		cfg.Auth.Validate, func() error { return cfg.CORS.Validate(shared.CurrentEnv()) },
	} {
		if err := validate(); err != nil {
			return nil, nil, err
		}
	}
	if err := validateLogCleanup(cfg.Workers.LogCleanup); err != nil {
		return nil, nil, err
	}
	return cfg, v, nil
}

func validateLogCleanup(cfg LogCleanup) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.RetentionDays <= 0 {
		return fmt.Errorf("workers.logCleanup.retentionDays must be positive")
	}
	if cfg.LockTTLMinutes <= 0 {
		return fmt.Errorf("workers.logCleanup.lockTTLMinutes must be positive")
	}
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cfg.Cron)
	if err != nil {
		return fmt.Errorf("workers.logCleanup.cron is invalid: %w", err)
	}
	first := schedule.Next(time.Now())
	if time.Duration(cfg.LockTTLMinutes)*time.Minute >= schedule.Next(first).Sub(first) {
		return fmt.Errorf("workers.logCleanup.lockTTLMinutes must be shorter than the cron interval")
	}
	return nil
}
