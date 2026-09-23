package bootstrap

import (
	"context"

	serviceconfig "github.com/gcc798/microservice-kit/application/resource/internal/config"
	"github.com/gcc798/microservice-kit/application/resource/internal/workers"
	logging "github.com/gcc798/microservice-kit/internal/logger"
)

type workerDeps struct {
	cleanup *workers.ExpiredAttachmentCleanup
}

func newWorkers(cfg *serviceconfig.Config, resources *resourceResources, domain domainDeps, log logging.Logger) (workerDeps, error) {
	if !cfg.Workers.ExpiredAttachmentCleanup.Enabled {
		return workerDeps{}, nil
	}
	cleanup, err := workers.NewExpiredAttachmentCleanup(cfg.Workers.ExpiredAttachmentCleanup, domain.attachments, resources.Redis, log)
	if err != nil {
		return workerDeps{}, err
	}
	return workerDeps{cleanup: cleanup}, nil
}

func (w workerDeps) start(ctx context.Context) {
	if w.cleanup != nil {
		w.cleanup.Start(ctx)
	}
}

func (w workerDeps) stop(ctx context.Context) error {
	if w.cleanup == nil {
		return nil
	}
	return w.cleanup.Stop(ctx)
}
