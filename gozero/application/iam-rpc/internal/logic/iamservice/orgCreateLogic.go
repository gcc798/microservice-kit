package iamservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type OrgCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrgCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgCreateLogic {
	return &OrgCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrgCreateLogic) OrgCreate(in *pb.OrgCreateReq) (*pb.Ack, error) {
	if in.OrgName == "" {
		return nil, errors.New("组织名称不能为空")
	}
	if in.OrgCode != "" {
		count, err := gorm.G[model.SOrg](l.svcCtx.DB).Where("org_code = ?", in.OrgCode).Count(l.ctx, "id")
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("组织编码已存在")
		}
	}
	ancestors, err := buildOrgAncestors(l.ctx, l.svcCtx, in.ParentId)
	if err != nil {
		return nil, err
	}
	orgType := in.OrgType
	if orgType == "" {
		orgType = "company"
	}
	now := sql.NullTime{Time: time.Now(), Valid: true}
	if err := gorm.G[model.SOrg](l.svcCtx.DB).Create(l.ctx, &model.SOrg{
		ParentId: in.ParentId, Ancestors: nullableString(ancestors), OrgName: in.OrgName,
		OrgCode: nullableString(in.OrgCode), OrgType: orgType, Leader: nullableString(in.Leader),
		Phone: nullableString(in.Phone), Email: nullableString(in.Email), Status: int64(in.Status),
		Sort: in.Sort, Remark: nullableString(in.Remark), CreatedTime: now, UpdatedTime: now,
	}); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
