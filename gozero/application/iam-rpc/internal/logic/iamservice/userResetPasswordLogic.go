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
)

type UserResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserResetPasswordLogic {
	return &UserResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserResetPasswordLogic) UserResetPassword(in *pb.UserPasswordReq) (*pb.Ack, error) {
	if in.Id <= 0 || in.NewPassword == "" {
		return nil, fmt.Errorf("参数不能为空")
	}
	if _, err := getUserByID(l.ctx, l.svcCtx, in.Id); err != nil {
		return nil, err
	}
	password, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SUser{}).Where("id = ?", in.Id).Updates(map[string]any{
		"password": string(password), "updated_time": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
