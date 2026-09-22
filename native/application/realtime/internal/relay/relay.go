package relay

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	logging "github.com/gcc798/microservice-kit/internal/logger"
	websocketx "github.com/gcc798/microservice-kit/internal/platform/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const Channel = "microservice-kit:realtime:deliver:v1"

type Message struct {
	UserIDs []int64         `json:"userIds"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
}

type Relay struct {
	redis  *redis.Client
	hub    *websocketx.Hub
	logger logging.Logger
	ready  atomic.Bool
}

func New(client *redis.Client, hub *websocketx.Hub, logger logging.Logger) *Relay {
	return &Relay{redis: client, hub: hub, logger: logger}
}

func (r *Relay) Start(ctx context.Context) {
	go r.consume(ctx)
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
	receivers, err := r.redis.Publish(ctx, Channel, payload).Result()
	if err != nil {
		return err
	}
	if receivers == 0 {
		return errors.New("redis channel has no subscribers")
	}
	return nil
}

func (r *Relay) consume(ctx context.Context) {
	for {
		pubsub := r.redis.Subscribe(ctx, Channel)
		if _, err := pubsub.Receive(ctx); err != nil {
			_ = pubsub.Close()
			if ctx.Err() == nil {
				r.logger.Error("realtime redis subscription failed", zap.Error(err))
				if !waitRetry(ctx) {
					return
				}
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
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				r.logger.Error("invalid realtime redis message", zap.Error(err))
				continue
			}
			for _, userID := range event.UserIDs {
				if err := r.hub.SendToUser(userID, event.Type, event.Data); err != nil {
					r.logger.Error("realtime message delivery failed", zap.Int64("userId", userID), zap.Error(err))
				}
			}
		}
		r.ready.Store(false)
		_ = pubsub.Close()
		if ctx.Err() != nil {
			return
		}
		r.logger.Warn("realtime redis subscription stopped")
		if !waitRetry(ctx) {
			return
		}
	}
}

func waitRetry(ctx context.Context) bool {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func WaitReady(ctx context.Context, relay *Relay) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if relay.Ready() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
