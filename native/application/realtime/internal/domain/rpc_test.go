package domain

import (
	"context"
	"testing"

	realtimev1 "github.com/gcc798/microservice-kit/internal/api/realtime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPublishToUsersRejectsInvalidRequests(t *testing.T) {
	server := NewServer(nil)
	for name, request := range map[string]*realtimev1.PublishToUsersRequest{
		"empty users":  {Type: "notice"},
		"empty type":   {UserIds: []int64{1}},
		"invalid id":   {UserIds: []int64{0}, Type: "notice"},
		"invalid json": {UserIds: []int64{1}, Type: "notice", DataJson: []byte("{")},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := server.PublishToUsers(context.Background(), request)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("status.Code(%v) = %s, want %s", err, status.Code(err), codes.InvalidArgument)
			}
		})
	}
}
