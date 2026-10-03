package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/config"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	sharedworker "github.com/gcc798/microservice-kit/internal/worker"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const logCleanupLockKey = "worker:sys:log-cleanup"

type LogCleanup struct {
	cron          *cron.Cron
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	retentionDays int
	lockTTL       time.Duration
}

func NewLogCleanup(ctx context.Context, cfg config.LogCleanupConf, svcCtx *svc.ServiceContext) (*LogCleanup, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.RetentionDays <= 0 {
		return nil, fmt.Errorf("workers.logCleanup.retentionDays must be positive")
	}
	lockTTL := time.Duration(cfg.LockTTLMinutes) * time.Minute
	schedule, err := sharedworker.Schedule(cfg.Cron, lockTTL)
	if err != nil {
		return nil, fmt.Errorf("schedule SYS log cleanup: %w", err)
	}
	worker := &LogCleanup{cron: cron.New(), ctx: ctx, svcCtx: svcCtx, retentionDays: cfg.RetentionDays, lockTTL: lockTTL}
	worker.cron.Schedule(schedule, cron.FuncJob(worker.run))
	return worker, nil
}

func (w *LogCleanup) Start() {
	if w != nil {
		w.cron.Start()
	}
}

func (w *LogCleanup) Stop() {
	if w != nil {
		<-w.cron.Stop().Done()
	}
}

func (w *LogCleanup) run() {
	claimed, err := sharedworker.Claim(w.ctx, w.svcCtx.Redis, logCleanupLockKey, w.lockTTL)
	if err != nil || !claimed {
		if err != nil {
			logx.Errorf("claim SYS log cleanup: %v", err)
		}
		return
	}
	cutoff := time.Now().AddDate(0, 0, -w.retentionDays)
	loginLogs, err := gorm.G[model.SLoginLog](w.svcCtx.DB).Where("login_time < ?", cutoff).Delete(w.ctx)
	if err != nil {
		logx.Errorf("clean login logs: %v", err)
		return
	}
	operationLogs, err := gorm.G[model.SOperLog](w.svcCtx.DB).Where("oper_time < ?", cutoff).Delete(w.ctx)
	if err != nil {
		logx.Errorf("clean operation logs: %v", err)
		return
	}
	logx.Infof("cleaned SYS logs: login=%d operation=%d", loginLogs, operationLogs)
}
