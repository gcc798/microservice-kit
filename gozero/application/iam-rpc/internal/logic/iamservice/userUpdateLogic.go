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

type UserUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateLogic {
	return &UserUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserUpdateLogic) UserUpdate(in *pb.UserUpdateReq) (*pb.Ack, error) {
	if in.Id <= 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}
	if _, err := getUserByID(l.ctx, l.svcCtx, in.Id); err != nil {
		return nil, err
	}
	var count int64
	var err error
	if in.UserName != "" {
		count, err = gorm.G[model.SUser](l.svcCtx.DB).Where("id <> ? AND user_name = ?", in.Id, in.UserName).Count(l.ctx, "id")
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("用户名已被占用")
		}
	}
	if in.Phonenumber != "" {
		count, err = gorm.G[model.SUser](l.svcCtx.DB).Where("id <> ? AND phonenumber = ?", in.Id, in.Phonenumber).Count(l.ctx, "id")
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("手机号已被占用")
		}
	}
	if in.Email != "" {
		count, err = gorm.G[model.SUser](l.svcCtx.DB).Where("id <> ? AND email = ?", in.Id, in.Email).Count(l.ctx, "id")
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("邮箱已被占用")
		}
	}
	err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.SUser{}).Where("id = ?", in.Id).Updates(map[string]any{
		"user_name": in.UserName, "nick_name": nullableString(in.NickName), "user_type": in.UserType,
		"email": nullableString(in.Email), "phonenumber": nullableString(in.Phonenumber), "sex": in.Sex,
		"avatar": nullableString(in.Avatar), "status": in.Status, "remark": nullableString(in.Remark),
		"update_by": nullableInt64(in.UpdateBy), "updated_time": time.Now(),
	}).Error
	if err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
