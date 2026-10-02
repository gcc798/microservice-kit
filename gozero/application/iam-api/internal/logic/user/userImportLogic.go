// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/logic/commonutil"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserImportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserImportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserImportLogic {
	return &UserImportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserImportLogic) UserImport(req *types.UserImportReq) (resp *types.CommonResp, err error) {
	userID := commonutil.UserIDFromContext(l.ctx)
	users := make([]*iamservice.UserCreateReq, 0, len(req.Users))
	for _, user := range req.Users {
		users = append(users, &iamservice.UserCreateReq{
			UserName:    user.UserName,
			NickName:    user.NickName,
			Password:    user.Password,
			UserType:    int32(user.UserType),
			Email:       user.Email,
			Phonenumber: user.Phonenumber,
			Sex:         int32(user.Sex),
			Avatar:      user.Avatar,
			Status:      int32(user.Status),
			Remark:      user.Remark,
			CreateBy:    userID,
			UpdateBy:    userID,
		})
	}
	if _, err := l.svcCtx.IamRpcClient.UserImport(l.ctx, &iamservice.UserImportReq{Users: users}); err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: "ok"}, nil
}
