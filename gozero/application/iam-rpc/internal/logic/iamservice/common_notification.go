package iamservicelogic

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/internal/runtimeconfig"
	"gopkg.in/gomail.v2"
)

func sendSMS(ctx context.Context, svcCtx *svc.ServiceContext, phone, code, template string) error {
	var config runtimeconfig.SMS
	if err := loadRuntimeConfig(ctx, svcCtx, runtimeconfig.CodeSMS, &config); err != nil {
		return err
	}
	if !config.Enabled {
		return fmt.Errorf("短信服务未启用")
	}
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("初始化短信服务失败: %w", err)
	}
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phone
	request.SignName = config.SignName
	request.TemplateCode = template
	request.TemplateParam = fmt.Sprintf(`{"code":"%s"}`, code)
	response, err := client.SendSms(request)
	if err != nil {
		return fmt.Errorf("发送短信失败: %w", err)
	}
	if response.Code != "OK" {
		return fmt.Errorf("发送短信失败: %s", response.Message)
	}
	return nil
}

func sendEmail(ctx context.Context, svcCtx *svc.ServiceContext, address, code, template string) error {
	var config runtimeconfig.Email
	if err := loadRuntimeConfig(ctx, svcCtx, runtimeconfig.CodeEmail, &config); err != nil {
		return err
	}
	if !config.Enabled {
		return fmt.Errorf("邮件服务未启用")
	}
	message := gomail.NewMessage()
	message.SetHeader("From", config.From)
	message.SetHeader("To", address)
	message.SetHeader("Subject", "验证码")
	message.SetBody("text/html", fmt.Sprintf(template, code))
	if err := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password).DialAndSend(message); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}
