package workers

import (
	"context"
	"fmt"
	"time"

	serviceconfig "github.com/gcc798/microservice-kit/application/sys/internal/config"
	sys "github.com/gcc798/microservice-kit/application/sys/internal/domain"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/platform/redislock"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const logCleanupLockKey = "worker:sys:log-cleanup"

type LogCleanup struct {
	cron          *cron.Cron
	ctx           context.Context
	api           sys.LogCleanupService
	locker        *redislock.Locker
	logger        logging.Logger
	retentionDays int32
}

func NewLogCleanup(cfg serviceconfig.LogCleanup, api sys.LogCleanupService, redisClient *redis.Client, logger logging.Logger) (*LogCleanup, error) {
	worker := &LogCleanup{
		cron: cron.New(cron.WithSeconds()), api: api, logger: logger, retentionDays: cfg.RetentionDays,
		locker: redislock.New(redisClient, redislock.WithTTL(time.Duration(cfg.LockTTLMinutes)*time.Minute)),
	}
	if _, err := worker.cron.AddFunc(cfg.Cron, worker.run); err != nil {
		return nil, fmt.Errorf("schedule SYS log cleanup: %w", err)
	}
	return worker, nil
}

func (w *LogCleanup) Start(ctx context.Context) {
	w.ctx = ctx
	w.cron.Start()
}

func (w *LogCleanup) Stop(ctx context.Context) error {
	select {
	case <-w.cron.Stop().Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *LogCleanup) run() {
	claimed, err := w.locker.TryClaim(w.ctx, logCleanupLockKey)
	if err != nil {
		w.logger.Error("claim SYS log cleanup failed", zap.Error(err))
		return
	}
	if !claimed {
		return
	}
	loginLogs, operationLogs, err := w.api.Clean(w.ctx, int(w.retentionDays))
	if err != nil {
		w.logger.Error("clean SYS logs failed", zap.Error(err))
		return
	}
	w.logger.Info("cleaned SYS logs", zap.Int64("loginLogs", loginLogs), zap.Int64("operationLogs", operationLogs))
}
