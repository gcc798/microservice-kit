package workers

import (
	"context"
	"fmt"
	"time"

	serviceconfig "github.com/gcc798/microservice-kit/application/resource/internal/config"
	resource "github.com/gcc798/microservice-kit/application/resource/internal/domain"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/platform/redislock"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const expiredAttachmentCleanupLockKey = "worker:resource:expired-attachment-cleanup"

type ExpiredAttachmentCleanup struct {
	cron        *cron.Cron
	ctx         context.Context
	attachments resource.AttachmentService
	locker      *redislock.Locker
	logger      logging.Logger
}

func NewExpiredAttachmentCleanup(cfg serviceconfig.ExpiredAttachmentCleanup, attachments resource.AttachmentService, redisClient *redis.Client, logger logging.Logger) (*ExpiredAttachmentCleanup, error) {
	worker := &ExpiredAttachmentCleanup{
		cron: cron.New(cron.WithSeconds()), attachments: attachments, logger: logger,
		locker: redislock.New(redisClient, redislock.WithTTL(time.Duration(cfg.LockTTLMinutes)*time.Minute)),
	}
	if _, err := worker.cron.AddFunc(cfg.Cron, worker.run); err != nil {
		return nil, fmt.Errorf("schedule expired attachment cleanup: %w", err)
	}
	return worker, nil
}

func (w *ExpiredAttachmentCleanup) Start(ctx context.Context) {
	w.ctx = ctx
	w.cron.Start()
}

func (w *ExpiredAttachmentCleanup) Stop(ctx context.Context) error {
	select {
	case <-w.cron.Stop().Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *ExpiredAttachmentCleanup) run() {
	claimed, err := w.locker.TryClaim(w.ctx, expiredAttachmentCleanupLockKey)
	if err != nil {
		w.logger.Error("claim expired attachment cleanup failed", zap.Error(err))
		return
	}
	if !claimed {
		return
	}
	cleaned, failed, err := w.attachments.CleanExpired(w.ctx)
	if err != nil {
		w.logger.Error("clean expired attachments failed", zap.Error(err))
		return
	}
	w.logger.Info("cleaned expired attachments", zap.Int64("cleaned", cleaned), zap.Int64("failed", failed))
}
