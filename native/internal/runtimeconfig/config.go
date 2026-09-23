// Package runtimeconfig 定义基于数据库的模块运行时配置契约。
package runtimeconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	CodeWeChat  = "integration.wechat"
	CodeSMS     = "integration.sms"
	CodeEmail   = "integration.email"
	CodeCaptcha = "auth.captcha"
)

// WeChatConfig 描述微信能力配置。
type WeChatConfig struct {
	Enabled    bool   `json:"enabled"`    // 是否启用。
	AppID      string `json:"appId"`      // 应用标识。
	Secret     string `json:"secret"`     // 应用密钥。
	TemplateID string `json:"templateId"` // 模板标识。
}

// SMSConfig 描述短信能力配置。
type SMSConfig struct {
	Enabled         bool   `json:"enabled"`         // 是否启用。
	AccessKeyID     string `json:"accessKeyId"`     // 访问密钥标识。
	AccessKeySecret string `json:"accessKeySecret"` // 访问密钥。
	SignName        string `json:"signName"`        // 短信签名。
	TemplateCode    string `json:"templateCode"`    // 短信模板编码。
}

// EmailConfig 描述邮件能力配置。
type EmailConfig struct {
	Enabled  bool   `json:"enabled"`  // 是否启用。
	Host     string `json:"host"`     // 邮件服务器地址。
	Port     int    `json:"port"`     // 邮件服务器端口。
	Username string `json:"username"` // 登录用户名。
	Password string `json:"password"` // 登录密码。
	From     string `json:"from"`     // 发件人地址。
}

// CaptchaConfig 描述全部验证码提供方配置。
type CaptchaConfig struct {
	Image ImageCaptchaConfig `json:"image"`
	SMS   SMSCaptchaConfig   `json:"sms"`
	Email EmailCaptchaConfig `json:"email"`
}

type ImageCaptchaConfig struct {
	Enabled bool `json:"enabled"` // 是否启用。
	Length  int  `json:"length"`  // 验证码长度。
	Width   int  `json:"width"`   // 图片宽度。
	Height  int  `json:"height"`  // 图片高度。
	Expire  int  `json:"expire"`  // 有效期，单位为秒。
}

type SMSCaptchaConfig struct {
	Enabled  bool   `json:"enabled"`  // 是否启用。
	Length   int    `json:"length"`   // 验证码长度。
	Expire   int    `json:"expire"`   // 有效期，单位为秒。
	Template string `json:"template"` // 短信模板。
	Provider string `json:"provider"` // 提供方名称。
}

type EmailCaptchaConfig struct {
	Enabled  bool   `json:"enabled"`  // 是否启用。
	Length   int    `json:"length"`   // 验证码长度。
	Expire   int    `json:"expire"`   // 有效期，单位为秒。
	Template string `json:"template"` // 邮件模板。
}

// Validate 校验已知运行时模块配置编码对应的 JSON 数据。
func Validate(code string, data []byte) error {
	if len(data) == 0 || !json.Valid(data) {
		return errors.New("configuration data must be valid JSON")
	}
	switch code {
	case CodeWeChat:
		var cfg WeChatConfig
		if err := decodeStrict(data, &cfg); err != nil {
			return err
		}
		if cfg.Enabled && (strings.TrimSpace(cfg.AppID) == "" || strings.TrimSpace(cfg.Secret) == "") {
			return errors.New("wechat appId and secret are required when enabled")
		}
	case CodeSMS:
		var cfg SMSConfig
		if err := decodeStrict(data, &cfg); err != nil {
			return err
		}
		if cfg.Enabled && (cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.SignName == "" || cfg.TemplateCode == "") {
			return errors.New("SMS credentials, signName and templateCode are required when enabled")
		}
	case CodeEmail:
		var cfg EmailConfig
		if err := decodeStrict(data, &cfg); err != nil {
			return err
		}
		if cfg.Enabled && (cfg.Host == "" || cfg.Port <= 0 || cfg.Username == "" || cfg.Password == "" || cfg.From == "") {
			return errors.New("email host, port, username, password and from are required when enabled")
		}
	case CodeCaptcha:
		var cfg CaptchaConfig
		if err := decodeStrict(data, &cfg); err != nil {
			return err
		}
		if cfg.Image.Enabled && (cfg.Image.Length <= 0 || cfg.Image.Width <= 0 || cfg.Image.Height <= 0 || cfg.Image.Expire <= 0) {
			return errors.New("enabled image captcha requires positive length, width, height and expire")
		}
		if cfg.SMS.Enabled && (cfg.SMS.Length <= 0 || cfg.SMS.Expire <= 0 || cfg.SMS.Template == "" || cfg.SMS.Provider == "") {
			return errors.New("enabled SMS captcha requires positive length and expire, template and provider")
		}
		if cfg.Email.Enabled && (cfg.Email.Length <= 0 || cfg.Email.Expire <= 0 || cfg.Email.Template == "") {
			return errors.New("enabled email captcha requires positive length and expire and template")
		}
	}
	return nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode runtime configuration: %w", err)
	}
	return nil
}
