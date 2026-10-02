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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCreateLogic {
	return &UserCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserCreateLogic) UserCreate(in *pb.UserCreateReq) (*pb.Ack, error) {
	if in.UserName == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	count, err := gorm.G[model.SUser](l.svcCtx.DB).Where("user_name = ?", in.UserName).Count(l.ctx, "id")
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("用户名已存在")
	}
	if in.Phonenumber != "" {
		count, err = gorm.G[model.SUser](l.svcCtx.DB).Where("phonenumber = ?", in.Phonenumber).Count(l.ctx, "id")
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("手机号已存在")
		}
	}
	if in.Email != "" {
		count, err = gorm.G[model.SUser](l.svcCtx.DB).Where("email = ?", in.Email).Count(l.ctx, "id")
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("邮箱已存在")
		}
	}
	password := ""
	if in.Password != "" {
		bs, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		password = string(bs)
	}
	now := sql.NullTime{Time: time.Now(), Valid: true}
	err = gorm.G[model.SUser](l.svcCtx.DB).Create(l.ctx, &model.SUser{
		UserName: in.UserName, NickName: nullableString(in.NickName), Password: nullableString(password),
		UserType: int64(in.UserType), Email: nullableString(in.Email), Phonenumber: nullableString(in.Phonenumber),
		Sex: int64(in.Sex), Avatar: nullableString(in.Avatar), Status: int64(in.Status), Remark: nullableString(in.Remark),
		CreateBy: nullableInt64(in.CreateBy), UpdateBy: nullableInt64(in.UpdateBy), CreatedTime: now, UpdatedTime: now,
	})
	if err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
