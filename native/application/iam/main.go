package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gcc798/microservice-kit/application/iam/internal/bootstrap"
	serviceconfig "github.com/gcc798/microservice-kit/application/iam/internal/config"
	sharedconfig "github.com/gcc798/microservice-kit/internal/config"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/telemetry"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, _, err := serviceconfig.Load("application/iam")
	if err != nil {
		return err
	}
	log, err := logging.NewLogger(sharedconfig.CurrentEnv(), cfg.AppDir)
	if err != nil {
		return err
	}
	shutdownTelemetry, err := telemetry.Init(ctx, string(sharedconfig.ServiceIAM), cfg.Service.ID, sharedconfig.CurrentEnv())
	if err != nil {
		return err
	}
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = errors.Join(err, shutdownTelemetry(shutdown))
	}()

	app, err := bootstrap.New(cfg, log)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, app.Close()) }()
	return app.Run(ctx)
}
