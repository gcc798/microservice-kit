package bootstrap

import (
	"context"

	serviceconfig "github.com/gcc798/microservice-kit/application/sys/internal/config"
	"github.com/gcc798/microservice-kit/application/sys/internal/workers"
	"github.com/gcc798/microservice-kit/internal/logger"
)

type workerDeps struct {
	logCleanup *workers.LogCleanup
}

func newWorkers(cfg *serviceconfig.Config, resources *sysResources, domain domainDeps, log logger.Logger) (workerDeps, error) {
	if !cfg.Workers.LogCleanup.Enabled {
		return workerDeps{}, nil
	}
	cleanup, err := workers.NewLogCleanup(cfg.Workers.LogCleanup, domain.logCleanup, resources.Redis, log)
	if err != nil {
		return workerDeps{}, err
	}
	return workerDeps{logCleanup: cleanup}, nil
}

func (w workerDeps) start(ctx context.Context) {
	if w.logCleanup != nil {
		w.logCleanup.Start(ctx)
	}
}

func (w workerDeps) stop(ctx context.Context) error {
	if w.logCleanup == nil {
		return nil
	}
	return w.logCleanup.Stop(ctx)
}
