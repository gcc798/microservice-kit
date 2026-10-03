package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

func Schedule(expression string, lockTTL time.Duration) (cron.Schedule, error) {
	if lockTTL <= 0 {
		return nil, fmt.Errorf("worker lock TTL must be positive")
	}
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expression)
	if err != nil {
		return nil, err
	}
	first := schedule.Next(time.Now())
	if lockTTL >= schedule.Next(first).Sub(first) {
		return nil, fmt.Errorf("worker lock TTL must be shorter than the schedule interval")
	}
	return schedule, nil
}

func Claim(ctx context.Context, client *redis.Client, key string, ttl time.Duration) (bool, error) {
	return client.SetNX(ctx, key, "claimed", ttl).Result()
}
