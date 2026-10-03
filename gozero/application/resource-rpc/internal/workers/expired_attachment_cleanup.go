package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/config"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/svc"
	sharedworker "github.com/gcc798/microservice-kit/internal/worker"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

const expiredAttachmentCleanupLockKey = "worker:resource:expired-attachment-cleanup"

type ExpiredAttachmentCleanup struct {
	cron    *cron.Cron
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	lockTTL time.Duration
}

func NewExpiredAttachmentCleanup(ctx context.Context, cfg config.ExpiredAttachmentCleanupConf, svcCtx *svc.ServiceContext) (*ExpiredAttachmentCleanup, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	lockTTL := time.Duration(cfg.LockTTLMinutes) * time.Minute
	schedule, err := sharedworker.Schedule(cfg.Cron, lockTTL)
	if err != nil {
		return nil, fmt.Errorf("schedule expired attachment cleanup: %w", err)
	}
	worker := &ExpiredAttachmentCleanup{cron: cron.New(), ctx: ctx, svcCtx: svcCtx, lockTTL: lockTTL}
	worker.cron.Schedule(schedule, cron.FuncJob(worker.run))
	return worker, nil
}

func (w *ExpiredAttachmentCleanup) Start() {
	if w != nil {
		w.cron.Start()
	}
}

func (w *ExpiredAttachmentCleanup) Stop() {
	if w != nil {
		<-w.cron.Stop().Done()
	}
}

func (w *ExpiredAttachmentCleanup) run() {
	claimed, err := sharedworker.Claim(w.ctx, w.svcCtx.Redis, expiredAttachmentCleanupLockKey, w.lockTTL)
	if err != nil || !claimed {
		if err != nil {
			logx.Errorf("claim expired attachment cleanup: %v", err)
		}
		return
	}
	var attachments []model.BizAttachment
	if err := w.svcCtx.DB.WithContext(w.ctx).Where("expire_time IS NOT NULL AND expire_time < ? AND status = 0", time.Now()).Find(&attachments).Error; err != nil {
		logx.Errorf("find expired attachments: %v", err)
		return
	}
	var cleaned, failed int
	for _, attachment := range attachments {
		if err := w.svcCtx.Storage.Delete(w.ctx, attachment.FileKey); err != nil {
			failed++
			logx.Errorf("delete expired attachment %d: %v", attachment.Id, err)
			continue
		}
		if err := w.svcCtx.DB.WithContext(w.ctx).Model(&model.BizAttachment{}).Where("id = ?", attachment.Id).Update("status", 1).Error; err != nil {
			failed++
			logx.Errorf("mark expired attachment %d deleted: %v", attachment.Id, err)
			continue
		}
		cleaned++
	}
	logx.Infof("cleaned expired attachments: cleaned=%d failed=%d", cleaned, failed)
}
