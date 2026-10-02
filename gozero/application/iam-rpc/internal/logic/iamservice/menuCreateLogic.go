package iamservicelogic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type MenuCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMenuCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenuCreateLogic {
	return &MenuCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MenuCreateLogic) MenuCreate(in *pb.MenuReq) (*pb.Ack, error) {
	if in.MenuName == "" {
		return nil, fmt.Errorf("菜单名称不能为空")
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
	count, err := gorm.G[model.SMenu](l.svcCtx.DB).Where("parent_id = ? AND menu_name = ?", in.ParentId, in.MenuName).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("同级菜单名称已存在")
	}
	now := sql.NullTime{Time: time.Now(), Valid: true}
	err = gorm.G[model.SMenu](l.svcCtx.DB).Create(l.ctx, &model.SMenu{
		MenuName: in.MenuName, ParentId: in.ParentId, Sort: in.Sort,
		Path: nullableString(in.Path), Component: nullableString(in.Component), Query: nullableString(in.Query),
		IsFrame: in.IsFrame, IsCache: in.IsCache, MenuType: in.MenuType, Visible: in.Visible, Status: in.Status,
		Perms: nullableString(in.Perms), Icon: nullableString(in.Icon), Remark: nullableString(in.Remark),
		CreateBy: nullableInt64(in.CreateBy), UpdateBy: nullableInt64(in.UpdateBy), CreatedTime: now, UpdatedTime: now,
	})
	if err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
