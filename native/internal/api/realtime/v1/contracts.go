package realtimev1

import (
	"context"

	"github.com/gcc798/microservice-kit/internal/transport"
)

const ServiceName = "realtime"

type API interface {
	PublishToUsers(context.Context, *PublishToUsersRequest) error
}

type Remote struct{ pool *transport.ClientPool }

func NewRemote(pool *transport.ClientPool) *Remote { return &Remote{pool: pool} }

func (r *Remote) PublishToUsers(ctx context.Context, request *PublishToUsersRequest) error {
	conn, err := r.pool.Conn(ctx, ServiceName)
	if err != nil {
		return err
	}
	_, err = NewRealtimeServiceClient(conn).PublishToUsers(ctx, request)
	return err
}
