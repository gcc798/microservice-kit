package iamservicelogic

import (
	"context"
	"errors"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"
	"github.com/gcc798/microservice-kit/common/auth"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ValidateAccessTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateAccessTokenLogic {
	return &ValidateAccessTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateAccessTokenLogic) ValidateAccessToken(in *pb.ValidateAccessTokenReq) (*pb.ValidateAccessTokenResp, error) {
	if in == nil {
		return nil, status.Error(codes.Unauthenticated, "missing access token")
	}
	claims, err := auth.ParseAccessToken(in.Token, l.svcCtx.Config.Jwt.Secret)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid access token")
	}
	active, err := l.svcCtx.Redis.Exists(l.ctx, auth.AccessTokenKey(in.Token)).Result()
	if err != nil {
		return nil, status.Error(codes.Unavailable, "failed to validate session")
	}
	if active != 1 {
		return nil, status.Error(codes.Unauthenticated, "access token is revoked")
	}
	if _, err := gorm.G[model.SUser](l.svcCtx.DB).Select("id").Where("id = ? AND status = 0", claims.UserID).First(l.ctx); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Unauthenticated, "user is unavailable")
		}
		return nil, status.Error(codes.Unavailable, "failed to validate user")
	}
	return &pb.ValidateAccessTokenResp{
		UserId: claims.UserID, UserName: claims.UserName, ClientId: claims.ClientID,
		DeviceType: claims.DeviceType, OrgId: claims.OrgID, Roles: claims.Roles, Permissions: claims.Permissions,
	}, nil
}
