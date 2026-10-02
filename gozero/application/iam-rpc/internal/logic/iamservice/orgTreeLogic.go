package iamservicelogic

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrgTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrgTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgTreeLogic {
	return &OrgTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrgTreeLogic) OrgTree(in *pb.Empty) (*pb.OrgTreeResp, error) {
	var rows []orgRow
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SOrg{}).Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.OrgTreeResp{Records: buildOrgTree(rows, 0)}, nil
}
