package captcha

import (
	"context"
	"time"
)

// CaptchaType 验证码类型
type CaptchaType string

const (
	// CaptchaTypeImage 定义业务常量。
	CaptchaTypeImage CaptchaType = "image" // 图形验证码
	// CaptchaTypeSMS 定义业务常量。
	CaptchaTypeSMS CaptchaType = "sms" // 短信验证码
	// CaptchaTypeEmail 定义业务常量。
	CaptchaTypeEmail CaptchaType = "email" // 邮箱验证码
)

// CaptchaData 验证码数据
type CaptchaData struct {
	ID       string      `json:"id"`       // 验证码ID
	Type     CaptchaType `json:"type"`     // 验证码类型
	Data     interface{} `json:"data"`     // 验证码数据
	ExpireAt time.Time   `json:"expireAt"` // 过期时间
}

// CaptchaProvider 定义验证码提供者能力。
type CaptchaProvider interface {
	// GetType 获取验证码类型
	GetType() CaptchaType

	// Generate 生成验证码
	Generate(ctx context.Context, params interface{}) (*CaptchaData, error)

	// Verify 验证验证码
	Verify(ctx context.Context, params interface{}) error

	// IsEnabled 是否启用
	IsEnabled() bool
}
