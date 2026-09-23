package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Jwt 管理 JSON Web Token 的签发和校验。
type Jwt struct {
	secret        []byte
	defaultExpire time.Duration
}

// Claims 保存令牌中的用户身份和标准声明。
type Claims struct {
	UserId     int64  `json:"userId"`     // 用户 ID。
	UserName   string `json:"userName"`   // 用户名。
	ClientId   string `json:"clientId"`   // 客户端 ID。
	DeviceType string `json:"deviceType"` // 设备类型。
	jwt.RegisteredClaims
}

// New 创建组件实例。
func New(secret string, expireSeconds int64) *Jwt {
	exp := time.Duration(expireSeconds) * time.Second
	if expireSeconds <= 0 {
		exp = 2 * time.Hour
	}
	return &Jwt{secret: []byte(secret), defaultExpire: exp}
}

// GenerateToken 签发访问令牌并返回令牌有效期。
func (s *Jwt) GenerateToken(userId int64, userName, clientId, deviceType string, expireSeconds ...int64) (string, int64, error) {
	expire := s.defaultExpire
	if len(expireSeconds) > 0 && expireSeconds[0] > 0 {
		expire = time.Duration(expireSeconds[0]) * time.Second
	}
	now := time.Now()
	claims := Claims{
		UserId:     userId,
		UserName:   userName,
		ClientId:   clientId,
		DeviceType: deviceType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "microservice-kit",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, err
	}
	return tokenStr, int64(expire.Seconds()), nil
}

// ValidateToken 校验令牌签名和标准声明。
func (s *Jwt) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer("microservice-kit"),
	)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// DefaultExpireSeconds 返回默认令牌有效期，单位为秒。
func (s *Jwt) DefaultExpireSeconds() int64 { return int64(s.defaultExpire.Seconds()) }
