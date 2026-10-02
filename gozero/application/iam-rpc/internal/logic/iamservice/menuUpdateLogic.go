package iamservicelogic

import (
	"context"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type MenuUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMenuUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuUpdateLogic {
	return &MenuUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MenuUpdateLogic) MenuUpdate(in *pb.MenuReq) (*pb.Ack, error) {
	if in.Id <= 0 {
		return nil, fmt.Errorf("菜单ID不能为空")
	}
	if in.MenuName == "" {
		return nil, fmt.Errorf("菜单名称不能为空")
	}
	if _, err := getMenuByID(l.ctx, l.svcCtx, in.Id); err != nil {
		return nil, err
	}
	if in.ParentId == in.Id {
		return nil, fmt.Errorf("不能将自己设置为父菜单")
	}
	if in.ParentId > 0 {
		parent, err := getMenuByID(l.ctx, l.svcCtx, in.ParentId)
		if err != nil {
			return nil, fmt.Errorf("父菜单不存在")
		}
		if parent.MenuType == 0 && in.MenuType == 2 {
			return nil, fmt.Errorf("目录下不能直接创建按钮")
		}
		if parent.MenuType == 1 && in.MenuType != 2 {
			return nil, fmt.Errorf("菜单下只能创建按钮")
		}
	}
	count, err := gorm.G[model.SMenu](l.svcCtx.DB).Where("id <> ? AND parent_id = ? AND menu_name = ?", in.Id, in.ParentId, in.MenuName).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("同级菜单名称已存在")
	}
	err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.SMenu{}).Where("id = ?", in.Id).Updates(map[string]any{
		"menu_name": in.MenuName, "parent_id": in.ParentId, "sort": in.Sort,
		"path": nullableString(in.Path), "component": nullableString(in.Component), "query": nullableString(in.Query),
		"is_frame": in.IsFrame, "is_cache": in.IsCache, "menu_type": in.MenuType, "visible": in.Visible,
		"status": in.Status, "perms": nullableString(in.Perms), "icon": nullableString(in.Icon),
		"remark": nullableString(in.Remark), "update_by": nullableInt64(in.UpdateBy), "updated_time": time.Now(),
	}).Error
	if err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
