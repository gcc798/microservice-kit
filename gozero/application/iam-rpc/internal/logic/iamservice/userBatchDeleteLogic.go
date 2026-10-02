package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserBatchDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserBatchDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserBatchDeleteLogic {
	return &UserBatchDeleteLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UserBatchDeleteLogic) UserBatchDelete(in *pb.BatchIdsReq) (*pb.Ack, error) {
	if len(in.Ids) == 0 {
		return &pb.Ack{Msg: "ok"}, nil
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&model.SUser{}, in.Ids).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
