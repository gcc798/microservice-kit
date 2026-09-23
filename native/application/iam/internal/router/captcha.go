package router

import "github.com/gcc798/microservice-kit/internal/httpx"

// 注册验证码相关路由。
func registerCaptchaRoutes(r *httpx.Router, ctx *RouterContext) {
	// 公开路由（无需认证）
	r.GET("/resource/sms/code", ctx.Captcha.ResourceSMSCode)

	captcha := r.Group("/captcha")
	{
		captcha.GET("/image", ctx.Captcha.GenerateImageCaptcha)    // 生成图形验证码
		captcha.POST("/sms", ctx.Captcha.SendSMSCaptcha)           // 发送短信验证码
		captcha.POST("/email", ctx.Captcha.SendEmailCaptcha)       // 发送邮箱验证码
		captcha.GET("/enabled-types", ctx.Captcha.GetEnabledTypes) // 获取启用的验证码类型
	}
}
