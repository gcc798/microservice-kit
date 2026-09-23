package controller

import (
	"fmt"
	"strings"

	"github.com/gcc798/microservice-kit/application/iam/internal/domain/model"
	"github.com/gcc798/microservice-kit/internal/platform/jwt"
	"github.com/labstack/echo/v5"
)

// BaseController 定义业务数据结构。
type BaseController struct {
	tokens      *jwt.Jwt
	tokenHeader string
}

// NewBaseController 创建组件实例。
func NewBaseController(tokens *jwt.Jwt, tokenHeader string) *BaseController {
	return &BaseController{tokens: tokens, tokenHeader: tokenHeader}
}

// GetUserId 获取当前用户ID
func (b *BaseController) GetUserId(c *echo.Context) (int64, error) {
	if userId := c.Get("userId"); userId != nil {
		if id, ok := userId.(int64); ok {
			return id, nil
		}
	}
	return 0, fmt.Errorf("未登录或用户ID不存在")
}

// GetUserName 获取当前用户名
func (b *BaseController) GetUserName(c *echo.Context) (string, error) {
	if userName := c.Get("userName"); userName != nil {
		if name, ok := userName.(string); ok {
			return name, nil
		}
	}
	return "", fmt.Errorf("未登录或用户名不存在")
}

// GetClientId 获取客户端ID
func (b *BaseController) GetClientId(c *echo.Context) (string, error) {
	if clientId := c.Get("clientId"); clientId != nil {
		if id, ok := clientId.(string); ok {
			return id, nil
		}
	}
	return "", fmt.Errorf("客户端ID不存在")
}

// GetDeviceType 获取设备类型
func (b *BaseController) GetDeviceType(c *echo.Context) (string, error) {
	if deviceType := c.Get("deviceType"); deviceType != nil {
		if dt, ok := deviceType.(string); ok {
			return dt, nil
		}
	}
	return "", fmt.Errorf("设备类型不存在")
}

// CurrentUser 从JWT token解析当前用户信息（不查询数据库）
func (b *BaseController) CurrentUser(c *echo.Context) (*model.User, error) {
	// 先尝试从 context 中获取（middleware 可能已经解析并设置）
	if userId := c.Get("userId"); userId != nil {
		if userName := c.Get("userName"); userName != nil {
			return &model.User{
				ID:       userId.(int64),
				UserName: userName.(string),
			}, nil
		}
	}

	// 如果 context 中没有，从 token 中解析
	token := c.Request().Header.Get(b.tokenHeader)
	token = strings.TrimPrefix(token, "Bearer ")
	claims, err := b.tokens.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:       claims.UserId,
		UserName: claims.UserName,
	}, nil
}
