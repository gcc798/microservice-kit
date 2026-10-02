package iamservicelogic

import (
	"context"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserChangePasswordLogic {
	return &UserChangePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserChangePasswordLogic) UserChangePassword(in *pb.UserChangePasswordReq) (*pb.Ack, error) {
	if in.UserId <= 0 || in.OldPassword == "" || in.NewPassword == "" {
		return nil, fmt.Errorf("参数不能为空")
	}
	user, err := gorm.G[model.SUser](l.svcCtx.DB).Select("password").Where("id = ?", in.UserId).First(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(nullString(user.Password)), []byte(in.OldPassword)); err != nil {
		return nil, fmt.Errorf("旧密码错误")
	}
	password, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SUser{}).Where("id = ?", in.UserId).Updates(map[string]any{
		"password": string(password), "updated_time": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
