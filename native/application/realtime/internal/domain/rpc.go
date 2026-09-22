package domain

import (
	"context"
	"encoding/json"

	"github.com/gcc798/microservice-kit/application/realtime/internal/relay"
	realtimev1 "github.com/gcc798/microservice-kit/internal/api/realtime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxPublishUsers = 1000

type Server struct {
	realtimev1.UnimplementedRealtimeServiceServer
	relay *relay.Relay
}

func NewServer(relay *relay.Relay) *Server { return &Server{relay: relay} }

func (s *Server) PublishToUsers(ctx context.Context, request *realtimev1.PublishToUsersRequest) (*realtimev1.PublishToUsersResponse, error) {
	if request == nil || len(request.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids must not be empty")
	}
	if len(request.UserIds) > maxPublishUsers {
		return nil, status.Errorf(codes.InvalidArgument, "user_ids cannot contain more than %d users", maxPublishUsers)
	}
	if request.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "type must not be empty")
	}
	data := request.DataJson
	if len(data) == 0 {
		data = []byte("null")
	}
	if !json.Valid(data) {
		return nil, status.Error(codes.InvalidArgument, "data_json must be valid JSON")
	}
	userIDs := make([]int64, 0, len(request.UserIds))
	seen := make(map[int64]struct{}, len(request.UserIds))
	for _, userID := range request.UserIds {
		if userID < 1 {
			return nil, status.Error(codes.InvalidArgument, "user_ids must contain positive IDs")
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}
	if len(userIDs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids must not be empty")
	}
	if err := s.relay.Publish(ctx, relay.Message{UserIDs: userIDs, Type: request.Type, Data: json.RawMessage(data)}); err != nil {
		return nil, status.Errorf(codes.Unavailable, "publish realtime message: %v", err)
	}
	return &realtimev1.PublishToUsersResponse{}, nil
}

var _ realtimev1.RealtimeServiceServer = (*Server)(nil)
