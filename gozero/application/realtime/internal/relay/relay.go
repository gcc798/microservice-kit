package relay

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"github.com/gcc798/microservice-kit/application/realtime/internal/hub"
	"github.com/redis/go-redis/v9"
)

const Channel = "microservice-kit:realtime:deliver:v1"

type Message struct {
	UserIDs []int64         `json:"userIds"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
}

type Relay struct {
	redis *redis.Client
	hub   *hub.Hub
	ready atomic.Bool
}

func New(client *redis.Client, connections *hub.Hub) *Relay {
	return &Relay{redis: client, hub: connections}
}

func (r *Relay) Ready() bool { return r.ready.Load() }

func (r *Relay) Publish(ctx context.Context, message Message) error {
	if !r.Ready() {
		return errors.New("redis subscription is not ready")
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.redis.Publish(ctx, Channel, payload).Err()
}

func (r *Relay) Run(ctx context.Context) {
	for ctx.Err() == nil {
		pubsub := r.redis.Subscribe(ctx, Channel)
		if _, err := pubsub.Receive(ctx); err != nil {
			_ = pubsub.Close()
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}
		r.ready.Store(true)
		for {
			message, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				break
			}
			var event Message
			if json.Unmarshal([]byte(message.Payload), &event) != nil {
				continue
			}
			payload, err := json.Marshal(struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}{Type: event.Type, Data: event.Data})
			if err != nil {
				continue
			}
			for _, userID := range event.UserIDs {
				_ = r.hub.Send(userID, payload)
			}
		}
		r.ready.Store(false)
		_ = pubsub.Close()
	}
}
