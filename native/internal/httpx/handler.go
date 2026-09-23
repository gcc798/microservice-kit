package httpx

import "github.com/labstack/echo/v5"

// Handler 将直接写入响应的控制器适配为 Echo 处理器契约。
func Handler(handler func(*echo.Context)) echo.HandlerFunc {
	return func(c *echo.Context) error {
		handler(c)
		return nil
	}
}
