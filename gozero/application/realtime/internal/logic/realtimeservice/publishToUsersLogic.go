package realtimeservicelogic

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gcc798/microservice-kit/application/realtime/internal/relay"
	"github.com/gcc798/microservice-kit/application/realtime/internal/svc"
	"github.com/gcc798/microservice-kit/application/realtime/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxPublishUsers = 1000

type PublishToUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishToUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishToUsersLogic {
	return &PublishToUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PublishToUsersLogic) PublishToUsers(in *pb.PublishToUsersReq) (*pb.PublishToUsersResp, error) {
	userIDs, data, err := validatePublish(in)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.Relay.Publish(l.ctx, relay.Message{UserIDs: userIDs, Type: strings.TrimSpace(in.Type), Data: data}); err != nil {
		return nil, status.Errorf(codes.Unavailable, "publish realtime message: %v", err)
	}
	return &pb.PublishToUsersResp{}, nil
}

func validatePublish(in *pb.PublishToUsersReq) ([]int64, json.RawMessage, error) {
	if in == nil || len(in.UserIds) == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "user_ids must not be empty")
	}
	if len(in.UserIds) > maxPublishUsers {
		return nil, nil, status.Errorf(codes.InvalidArgument, "user_ids cannot contain more than %d users", maxPublishUsers)
	}
	if strings.TrimSpace(in.Type) == "" {
		return nil, nil, status.Error(codes.InvalidArgument, "type must not be empty")
	}
	data := in.DataJson
	if len(data) == 0 {
		data = []byte("null")
	}
	if !json.Valid(data) {
		return nil, nil, status.Error(codes.InvalidArgument, "data_json must be valid JSON")
	}
	seen := make(map[int64]struct{}, len(in.UserIds))
	userIDs := make([]int64, 0, len(in.UserIds))
	for _, userID := range in.UserIds {
		if userID < 1 {
			return nil, nil, status.Error(codes.InvalidArgument, "user_ids must contain positive IDs")
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}
	return userIDs, json.RawMessage(data), nil
}
