// Package redislock 提供支持上下文取消的 Redis 分布式锁。
package redislock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	randv2 "math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"
)

const releaseScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
end
return 0
`

var ErrNotOwner = errors.New("redis lock is no longer owned by this holder")

// Locker 基于一个 Redis 客户端创建相互独立的分布式锁。
type Locker struct {
	client   *redis.Client // Redis 客户端。
	ttl      time.Duration // 锁的过期时间。
	retryMin time.Duration // 重试间隔下限。
	retryMax time.Duration // 重试间隔上限。
}

// Option 定制 Locker 的行为。
type Option func(*Locker)

// WithTTL 设置锁的过期时间。
func WithTTL(ttl time.Duration) Option {
	return func(l *Locker) { l.ttl = ttl }
}

// WithRetryRange 设置获取锁时的随机重试间隔范围。
func WithRetryRange(minimum, maximum time.Duration) Option {
	return func(l *Locker) {
		l.retryMin = minimum
		l.retryMax = maximum
	}
}

// New 创建 Redis 分布式锁管理器。
func New(client *redis.Client, options ...Option) *Locker {
	l := &Locker{
		client:   client,
		ttl:      10 * time.Second,
		retryMin: 20 * time.Millisecond,
		retryMax: 80 * time.Millisecond,
	}
	for _, option := range options {
		option(l)
	}
	if l.retryMax < l.retryMin {
		l.retryMax = l.retryMin
	}
	return l
}

// Lock 表示一次已经获取的锁。
type Lock struct {
	client *redis.Client // Redis 客户端。
	key    string        // 锁键。
	token  string        // 当前持有者令牌。
}

// Acquire 持续尝试获取锁，直到成功或上下文被取消。
func (l *Locker) Acquire(ctx context.Context, key string) (*Lock, error) {
	if l == nil || l.client == nil {
		return nil, errors.New("redis lock client is nil")
	}
	if key == "" {
		return nil, errors.New("redis lock key is empty")
	}
	if l.ttl <= 0 {
		return nil, errors.New("redis lock ttl must be positive")
	}
	token, err := newToken()
	if err != nil {
		return nil, fmt.Errorf("create redis lock token: %w", err)
	}
	for {
		acquired, err := l.client.SetNX(ctx, key, token, l.ttl).Result()
		if err != nil {
			return nil, fmt.Errorf("acquire redis lock %q: %w", key, err)
		}
		if acquired {
			return &Lock{client: l.client, key: key, token: token}, nil
		}
		timer := time.NewTimer(l.retryDelay())
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, fmt.Errorf("acquire redis lock %q: %w", key, ctx.Err())
		case <-timer.C:
		}
	}
}

// TryClaim 尝试一次性占用一个有时间边界的执行窗口。
// 占用成功后不会提前释放，避免其他实例在任务完成后重复执行同一任务。
func (l *Locker) TryClaim(ctx context.Context, key string) (bool, error) {
	if l == nil || l.client == nil {
		return false, errors.New("redis lock client is nil")
	}
	if key == "" {
		return false, errors.New("redis lock key is empty")
	}
	if l.ttl <= 0 {
		return false, errors.New("redis lock ttl must be positive")
	}
	token, err := newToken()
	if err != nil {
		return false, fmt.Errorf("create redis lock token: %w", err)
	}
	claimed, err := l.client.SetNX(ctx, key, token, l.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("claim redis execution window %q: %w", key, err)
	}
	return claimed, nil
}

// Release 仅当随机令牌仍属于当前持有者时释放锁。
func (l *Lock) Release(ctx context.Context) error {
	if l == nil || l.client == nil {
		return errors.New("redis lock is nil")
	}
	result, err := l.client.Eval(ctx, releaseScript, []string{l.key}, l.token).Int64()
	if err != nil {
		return fmt.Errorf("release redis lock %q: %w", l.key, err)
	}
	if result == 0 {
		return ErrNotOwner
	}
	return nil
}

// WithLock 获取指定锁，执行函数后安全释放锁。
func (l *Locker) WithLock(ctx context.Context, key string, fn func() error) (err error) {
	lock, err := l.Acquire(ctx, key)
	if err != nil {
		return err
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		releaseErr := lock.Release(releaseCtx)
		if releaseErr != nil && !errors.Is(releaseErr, ErrNotOwner) {
			err = errors.Join(err, releaseErr)
		}
	}()
	return fn()
}

func (l *Locker) retryDelay() time.Duration {
	if l.retryMax <= l.retryMin {
		return l.retryMin
	}
	return l.retryMin + time.Duration(randv2.Int64N(int64(l.retryMax-l.retryMin)+1))
}

func newToken() (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(data[:]), nil
}
