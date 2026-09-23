package database

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/gcc798/microservice-kit/internal/config"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Open(cfg config.Database, log logging.Logger) (*gorm.DB, error) {
	slowThreshold := time.Second
	if cfg.SlowThreshold > 0 {
		slowThreshold = time.Duration(cfg.SlowThreshold) * time.Millisecond
	}
	sqlDB, err := otelsql.Open("pgx", cfg.DSN, otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL))
	if err != nil {
		return nil, fmt.Errorf("instrument database connection: %w", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: gormlogger.New(stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags), gormlogger.Config{SlowThreshold: slowThreshold, LogLevel: gormlogger.Warn, IgnoreRecordNotFoundError: true, Colorful: true})})
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	pool, err := db.DB()
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	pool.SetMaxIdleConns(cfg.MaxIdleConns)
	pool.SetMaxOpenConns(cfg.MaxOpenConns)
	pool.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)
	if err := db.Use(&IDGenPlugin{}); err != nil {
		log.Warn("failed to register ID generation plugin", zap.Error(err))
	}
	return db, nil
}

func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
