package sysservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogPageLogic {
	return &LoginLogPageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogPageLogic) LoginLogPage(in *pb.LoginLogPageReq) (*pb.LoginLogPageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SLoginLog{})
	if in.UserName != "" {
		query = query.Where("user_name LIKE ?", "%"+in.UserName+"%")
	}
	if in.Ipaddr != "" {
		query = query.Where("ipaddr LIKE ?", "%"+in.Ipaddr+"%")
	}
	if in.Status == 0 || in.Status == 1 {
		query = query.Where("status = ?", in.Status)
	}
	if in.StartTime != "" {
		query = query.Where("login_time >= ?", in.StartTime)
	}
	if in.EndTime != "" {
		query = query.Where("login_time <= ?", in.EndTime)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.SLoginLog
	if err := query.Order("login_time DESC NULLS LAST, id DESC").Limit(int(pageSize)).Offset(int((pageNum - 1) * pageSize)).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.LoginLogPageResp{Records: toLoginLogList(rows), Page: toPageInfo(total, pageNum, pageSize)}, nil
}
