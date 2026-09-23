package runtimeconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gcc798/microservice-kit/internal/platform/redislock"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	cacheKeyPrefix = "microservice-kit:config:"
	lockKeyPrefix  = "microservice-kit:lock:config:"
)

// Source 在 Redis 没有缓存时加载持久化配置。
type Source interface {
	// Load 按配置编码读取持久化配置数据。
	Load(ctx context.Context, code string) ([]byte, error)
}

type gormSource struct{ db *gorm.DB }

// configRow 是共享运行时配置所需的最小持久化投影。
// 完整领域模型由 SYS 服务维护，共享基础设施不依赖该模型。
type configRow struct {
	Data []byte `gorm:"column:data"` // 配置 JSON 数据。
}

// TableName 返回运行时配置表名。
func (configRow) TableName() string { return "s_config" }

// NewGormSource 创建基于 s_config 表的数据源。
func NewGormSource(db *gorm.DB) Source { return &gormSource{db: db} }

// Load 从 s_config 表读取指定配置。
func (s *gormSource) Load(ctx context.Context, code string) ([]byte, error) {
	var config configRow
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&config).Error; err != nil {
		return nil, err
	}
	return append([]byte(nil), config.Data...), nil
}

// Store 是带数据库回源能力的共享 Redis 运行时配置存储。
type Store struct {
	redis  *redis.Client     // Redis 客户端。
	source Source            // 缓存未命中时的数据源。
	locker *redislock.Locker // 配置编码级分布式锁。
}

// NewStore 创建运行时配置存储。
func NewStore(redisClient *redis.Client, source Source, locker *redislock.Locker) *Store {
	return &Store{redis: redisClient, source: source, locker: locker}
}

// Get 读取并解码一个配置编码对应的配置。
func (s *Store) Get(ctx context.Context, code string, target any) error {
	raw, err := s.GetRaw(ctx, code)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode runtime configuration %q: %w", code, err)
	}
	return nil
}

// GetRaw 读取 Redis 数据；缓存未命中时在分布式锁内回源 s_config。
func (s *Store) GetRaw(ctx context.Context, code string) ([]byte, error) {
	if code == "" {
		return nil, errors.New("configuration code is empty")
	}
	raw, err := s.redis.Get(ctx, CacheKey(code)).Bytes()
	if err == nil {
		if validateErr := Validate(code, raw); validateErr == nil {
			return raw, nil
		}
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("read runtime configuration %q from Redis: %w", code, err)
	}
	var loaded []byte
	if err := s.WithCodeLock(ctx, code, func() error {
		cached, err := s.redis.Get(ctx, CacheKey(code)).Bytes()
		if err == nil {
			if validateErr := Validate(code, cached); validateErr == nil {
				loaded = cached
				return nil
			}
			if err := s.redis.Del(ctx, CacheKey(code)).Err(); err != nil {
				return fmt.Errorf("delete invalid runtime configuration %q from Redis: %w", code, err)
			}
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			return fmt.Errorf("recheck runtime configuration %q in Redis: %w", code, err)
		}
		persistent, err := s.source.Load(ctx, code)
		if err != nil {
			return fmt.Errorf("load runtime configuration %q from s_config: %w", code, err)
		}
		if err := Validate(code, persistent); err != nil {
			return fmt.Errorf("validate runtime configuration %q: %w", code, err)
		}
		if err := s.redis.Set(ctx, CacheKey(code), persistent, 0).Err(); err != nil {
			return fmt.Errorf("cache runtime configuration %q: %w", code, err)
		}
		loaded = persistent
		return nil
	}); err != nil {
		return nil, err
	}
	return loaded, nil
}

// WithCodeLock 串行化指定配置编码的缓存读取和修改。
func (s *Store) WithCodeLock(ctx context.Context, code string, fn func() error) error {
	return s.locker.WithLock(ctx, LockKey(code), fn)
}

// DeleteCache 删除共享运行时配置缓存；调用方应先持有配置编码锁。
func (s *Store) DeleteCache(ctx context.Context, code string) error {
	if err := s.redis.Del(ctx, CacheKey(code)).Err(); err != nil {
		return fmt.Errorf("delete runtime configuration cache %q: %w", code, err)
	}
	return nil
}

// SetCache 写入共享运行时配置缓存；调用方应先持有配置编码锁。
func (s *Store) SetCache(ctx context.Context, code string, data []byte) error {
	if err := Validate(code, data); err != nil {
		return fmt.Errorf("validate runtime configuration %q: %w", code, err)
	}
	if err := s.redis.Set(ctx, CacheKey(code), data, 0).Err(); err != nil {
		return fmt.Errorf("write runtime configuration cache %q: %w", code, err)
	}
	return nil
}

// CacheKey 返回配置编码对应的 Redis 缓存键。
func CacheKey(code string) string { return cacheKeyPrefix + code }

// LockKey 返回配置编码对应的 Redis 分布式锁键。
func LockKey(code string) string { return lockKeyPrefix + code }
