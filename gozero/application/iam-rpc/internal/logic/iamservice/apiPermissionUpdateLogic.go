package iamservicelogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApiPermissionUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApiPermissionUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiPermissionUpdateLogic {
	return &ApiPermissionUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApiPermissionUpdateLogic) ApiPermissionUpdate(in *pb.ApiPermissionSaveReq) (*pb.Ack, error) {
	if in.Id <= 0 {
		return nil, fmt.Errorf("权限ID不能为空")
	}
	if _, err := getApiPermissionByID(l.ctx, l.svcCtx, in.Id); err != nil {
		return nil, err
	}
	in.Method = strings.ToUpper(in.Method)
	in.Action = normalizeApiPermissionAction(in.Code, in.Action)
	if err := validateApiPermission(l.ctx, l.svcCtx, in); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SApiPermission{}).Where("id = ?", in.Id).Updates(map[string]any{
		"parent_id": in.ParentId, "module": in.Module, "code": in.Code, "name": in.Name,
		"node_type": in.NodeType, "action": in.Action, "method": nullableString(in.Method),
		"path": nullableString(in.Path), "sort": in.Sort, "status": in.Status,
		"remark": nullableString(in.Remark), "update_by": nullableInt64(in.UserId), "updated_time": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
