package database

import (
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/semconv/v1.39.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DSN                    string
	MaxIdleConns           int
	MaxOpenConns           int
	ConnMaxLifetimeMinutes int
}

func Open(cfg Config) (*gorm.DB, error) {
	pool, err := otelsql.Open("pgx", cfg.DSN, otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL))
	if err != nil {
		return nil, fmt.Errorf("instrument postgres: %w", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{})
	if err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Use(&IDGenPlugin{}); err != nil {
		return nil, fmt.Errorf("register id generator: %w", err)
	}

	pool, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres connection pool: %w", err)
	}
	pool.SetMaxIdleConns(cfg.MaxIdleConns)
	pool.SetMaxOpenConns(cfg.MaxOpenConns)
	pool.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)
	return db, nil
}

func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	pool, err := db.DB()
	if err != nil {
		return fmt.Errorf("get postgres connection pool: %w", err)
	}
	return pool.Close()
}
