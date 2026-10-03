package runtimeconfig

import (
	"encoding/json"
	"errors"
	"strings"
)

const (
	CodeWeChat  = "integration.wechat"
	CodeSMS     = "integration.sms"
	CodeEmail   = "integration.email"
	CodeCaptcha = "auth.captcha"
)

type WeChat struct {
	Enabled bool   `json:"enabled"`
	AppID   string `json:"appId"`
	Secret  string `json:"secret"`
}

type SMS struct {
	Enabled         bool   `json:"enabled"`
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	SignName        string `json:"signName"`
	TemplateCode    string `json:"templateCode"`
}

type Email struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

type Captcha struct {
	Image struct {
		Enabled bool `json:"enabled"`
		Length  int  `json:"length"`
		Width   int  `json:"width"`
		Height  int  `json:"height"`
		Expire  int  `json:"expire"`
	} `json:"image"`
	SMS struct {
		Enabled  bool   `json:"enabled"`
		Length   int    `json:"length"`
		Expire   int    `json:"expire"`
		Template string `json:"template"`
		Provider string `json:"provider"`
	} `json:"sms"`
	Email struct {
		Enabled  bool   `json:"enabled"`
		Length   int    `json:"length"`
		Expire   int    `json:"expire"`
		Template string `json:"template"`
	} `json:"email"`
}

func Validate(code, data string) error {
	switch code {
	case CodeWeChat:
		var value WeChat
		if err := decode(data, &value); err != nil {
			return err
		}
		if value.Enabled && (strings.TrimSpace(value.AppID) == "" || strings.TrimSpace(value.Secret) == "") {
			return errors.New("启用微信时 appId 和 secret 不能为空")
		}
	case CodeCaptcha:
		var value Captcha
		if err := decode(data, &value); err != nil {
			return err
		}
		if value.Image.Enabled && (value.Image.Length <= 0 || value.Image.Width <= 0 || value.Image.Height <= 0 || value.Image.Expire <= 0) {
			return errors.New("启用图形验证码时 length、width、height 和 expire 必须大于0")
		}
		if value.SMS.Enabled && (value.SMS.Length <= 0 || value.SMS.Expire <= 0 || value.SMS.Template == "" || value.SMS.Provider == "") {
			return errors.New("启用短信验证码时 length、expire、template 和 provider 不能为空")
		}
		if value.Email.Enabled && (value.Email.Length <= 0 || value.Email.Expire <= 0 || value.Email.Template == "") {
			return errors.New("启用邮箱验证码时 length、expire 和 template 不能为空")
		}
	case CodeSMS:
		var value SMS
		if err := decode(data, &value); err != nil {
			return err
		}
		if value.Enabled && (value.AccessKeyID == "" || value.AccessKeySecret == "" || value.SignName == "" || value.TemplateCode == "") {
			return errors.New("启用短信时访问密钥、签名和模板编码不能为空")
		}
	case CodeEmail:
		var value Email
		if err := decode(data, &value); err != nil {
			return err
		}
		if value.Enabled && (value.Host == "" || value.Port <= 0 || value.Username == "" || value.Password == "" || value.From == "") {
			return errors.New("启用邮箱时服务器、端口、账号、密码和发件人不能为空")
		}
	default:
		if !json.Valid([]byte(data)) {
			return errors.New("配置数据必须是合法 JSON")
		}
	}
	return nil
}

func decode(data string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
