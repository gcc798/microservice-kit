package iamservicelogic

import (
	"context"
	"errors"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type OrgDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrgDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgDeleteLogic {
	return &OrgDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrgDeleteLogic) OrgDelete(in *pb.IdReq) (*pb.Ack, error) {
	if _, err := getOrgByID(l.ctx, l.svcCtx, in.Id); err != nil {
		return nil, err
	}
	childCount, err := gorm.G[model.SOrg](l.svcCtx.DB).Where("parent_id = ?", in.Id).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if childCount > 0 {
		return nil, errors.New("存在子组织，无法删除")
	}
	userCount, err := gorm.G[model.SUser](l.svcCtx.DB).Where("org_id = ?", in.Id).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if userCount > 0 {
		return nil, errors.New("组织下存在用户，无法删除")
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&model.SOrg{}, in.Id).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
