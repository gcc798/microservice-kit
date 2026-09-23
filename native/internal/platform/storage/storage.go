// Package storage 提供由启动配置决定的 S3 兼容对象存储接入。
package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Storage 统一存储接口。
type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	// GetURL 返回访问 URL，expires 为 0 表示永久 URL。
	GetURL(ctx context.Context, key string, expires time.Duration) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	List(ctx context.Context, prefix string) ([]string, error)
	GetInfo(ctx context.Context, key string) (*FileInfo, error)
}

// FileInfo 文件元信息。
type FileInfo struct {
	Key          string    // 对象键。
	Size         int64     // 对象大小，单位为字节。
	LastModified time.Time // 最后修改时间。
	ContentType  string    // 内容类型。
	ETag         string    // 对象实体标签。
}

// Config S3 兼容存储的启动配置，同时作为 YAML 配置结构，
// 避免在配置包中重复定义一遍字段。
type Config struct {
	Endpoint  string `mapstructure:"endpoint"`  // 存储服务地址。
	AccessKey string `mapstructure:"accessKey"` // 访问密钥标识。
	SecretKey string `mapstructure:"secretKey"` // 访问密钥。
	Bucket    string `mapstructure:"bucket"`    // 默认存储桶。
	Region    string `mapstructure:"region"`    // 存储区域。
	UseSSL    bool   `mapstructure:"useSSL"`    // 是否使用 HTTPS。
}

// Validate 校验存储配置在启动阶段是否完整。
func (c Config) Validate() error {
	if c.Endpoint == "" || c.AccessKey == "" || c.SecretKey == "" || c.Bucket == "" {
		return fmt.Errorf("storage endpoint, credentials and bucket are required")
	}
	return nil
}

// New 按配置创建 Storage 实例。
func New(cfg Config) (Storage, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return NewS3Storage(cfg)
}
