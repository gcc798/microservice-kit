package modules

import (
	"context"
	"errors"

	"github.com/gcc798/microservice-kit/internal/platform/captcha"
	"github.com/gcc798/microservice-kit/internal/platform/thirdparty/wechat"
)

var ErrDisabled = errors.New("module is disabled")

const (
	WeChatName  = "wechat"
	SMSName     = "sms"
	EmailName   = "email"
	CaptchaName = "captcha"
)

type WeChat interface {
	Module
	Code2Session(ctx context.Context, wxCode string) (*wechat.Code2SessionResponse, error)
}

type SMS interface {
	Module
	SendVerificationCode(ctx context.Context, phonenumber string) (string, error)
	SendSMS(ctx context.Context, phone, code, template string) error
}

type Email interface {
	Module
	SendEmail(ctx context.Context, address, code, template string) error
}

type Captcha interface {
	Module
	Generate(ctx context.Context, captchaType captcha.CaptchaType, params any) (*captcha.CaptchaData, error)
	Verify(ctx context.Context, captchaType captcha.CaptchaType, params any) error
	GetEnabledTypes(ctx context.Context) ([]captcha.CaptchaType, error)
	ImageEnabled(ctx context.Context) (bool, error)
}
