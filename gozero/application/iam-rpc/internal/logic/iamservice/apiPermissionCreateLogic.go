package iamservicelogic

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ApiPermissionCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApiPermissionCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiPermissionCreateLogic {
	return &ApiPermissionCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApiPermissionCreateLogic) ApiPermissionCreate(in *pb.ApiPermissionSaveReq) (*pb.ApiPermission, error) {
	in.Method = strings.ToUpper(in.Method)
	in.Action = normalizeApiPermissionAction(in.Code, in.Action)
	if err := validateApiPermission(l.ctx, l.svcCtx, in); err != nil {
		return nil, err
	}
	now := sql.NullTime{Time: time.Now(), Valid: true}
	permission := model.SApiPermission{
		ParentId: in.ParentId, Module: in.Module, Code: in.Code, Name: in.Name,
		NodeType: int64(in.NodeType), Action: in.Action, Method: nullableString(in.Method), Path: nullableString(in.Path),
		Sort: in.Sort, Status: int64(in.Status), Remark: nullableString(in.Remark),
		CreateBy: nullableInt64(in.UserId), UpdateBy: nullableInt64(in.UserId), CreatedTime: now, UpdatedTime: now,
	}
	if err := gorm.G[model.SApiPermission](l.svcCtx.DB).Create(l.ctx, &permission); err != nil {
		return nil, err
	}
	row, err := getApiPermissionByID(l.ctx, l.svcCtx, permission.Id)
	if err != nil {
		return nil, err
	}
	return toApiPermissionPB(*row), nil
}
