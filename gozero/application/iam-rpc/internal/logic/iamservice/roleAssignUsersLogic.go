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
	"gorm.io/gorm/clause"
)

type RoleAssignUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleAssignUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleAssignUsersLogic {
	return &RoleAssignUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleAssignUsersLogic) RoleAssignUsers(in *pb.RoleUsersReq) (*pb.Ack, error) {
	if in.RoleId <= 0 {
		return nil, fmt.Errorf("角色ID不能为空")
	}
	if _, err := getRoleByID(l.ctx, l.svcCtx, in.RoleId); err != nil {
		return nil, err
	}
	userIDs := uniqueInt64Values(in.UserIds)
	if len(userIDs) == 0 {
		return &pb.Ack{Msg: "ok"}, nil
	}
	userCount, err := gorm.G[model.SUser](l.svcCtx.DB).Where("id IN ?", userIDs).Count(l.ctx, "id")
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if userCount != int64(len(userIDs)) {
		return nil, fmt.Errorf("部分用户不存在")
	}
	now := sql.NullTime{Time: time.Now(), Valid: true}
	rows := make([]model.MUserRole, 0, len(userIDs))
	for _, userID := range userIDs {
		rows = append(rows, model.MUserRole{
			UserId: userID, RoleId: in.RoleId, CreateBy: nullableInt64(in.OperatorId),
			UpdateBy: nullableInt64(in.OperatorId), CreatedTime: now, UpdatedTime: now,
		})
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
