package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
	commonauth "github.com/gcc798/microservice-kit/common/auth"
	"github.com/gcc798/microservice-kit/common/requestmeta"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
)

type authClient struct {
	ClientId      string
	ClientKey     string
	DeviceType    string
	Timeout       int64
	ActiveTimeout int64
}

type tokenSession struct {
	UserID        int64  `json:"userId"`
	UserName      string `json:"userName"`
	OrgID         int64  `json:"orgId"`
	ClientID      string `json:"clientId"`
	ClientKey     string `json:"clientKey"`
	DeviceType    string `json:"deviceType"`
	Timeout       int64  `json:"timeout"`
	ActiveTimeout int64  `json:"activeTimeout"`
	AccessHash    string `json:"accessHash"`
}

type loginUser struct {
	Id          int64
	UserName    string
	NickName    string
	Email       string
	Phonenumber string
	Avatar      string
	UserType    int64
	OrgID       int64
	Roles       []string
	Permissions []string
	OpenID      string
	UnionID     string
}

func loginWithRPC(ctx context.Context, svcCtx *svc.ServiceContext, req *types.LoginReq) (*loginUser, *authClient, error) {
	loginIP, userAgent := requestmeta.ClientInfo(ctx)
	code := req.Code
	if code == "" && req.SmsCode != "" {
		code = req.SmsCode
	}
	resp, err := svcCtx.IamRpcClient.AuthLogin(ctx, &iamservice.AuthLoginReq{
		ClientKey:   req.ClientId,
		GrantType:   req.GrantType,
		Username:    req.Username,
		Password:    req.Password,
		Code:        code,
		Phonenumber: req.Phonenumber,
		Email:       req.Email,
		WxCode:      req.WxCode,
		Uuid:        req.Uuid,
		LoginIp:     loginIP,
		UserAgent:   userAgent,
	})
	if err != nil {
		return nil, nil, err
	}
	if resp.UserInfo == nil {
		return nil, nil, fmt.Errorf("登录失败")
	}
	user := &loginUser{
		Id:          resp.UserInfo.UserId,
		UserName:    resp.UserInfo.Username,
		NickName:    resp.UserInfo.Nickname,
		Email:       resp.UserInfo.Email,
		Phonenumber: resp.UserInfo.Phonenumber,
		Avatar:      resp.UserInfo.Avatar,
		UserType:    int64(resp.UserInfo.UserType),
		OrgID:       resp.UserInfo.OrgId,
		OpenID:      resp.UserInfo.OpenId,
		UnionID:     resp.UserInfo.UnionId,
	}
	if err := enrichLoginUserAuthContext(ctx, svcCtx, user); err != nil {
		return nil, nil, err
	}
	return user, &authClient{
		ClientId:      resp.ClientId,
		ClientKey:     resp.ClientKey,
		DeviceType:    resp.DeviceType,
		Timeout:       resp.Timeout,
		ActiveTimeout: resp.ActiveTimeout,
	}, nil
}

func buildLoginResponse(ctx context.Context, svcCtx *svc.ServiceContext, user *loginUser, client *authClient) (*types.CommonResp, error) {
	accessToken, accessExpiresIn, err := generateAccessToken(user, client, svcCtx.Config.Jwt.Secret)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败")
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("生成Token失败")
	}
	if err := storeTokenSession(ctx, svcCtx, user, client, accessToken, refreshToken, !svcCtx.Config.Auth.AllowConcurrent); err != nil {
		return nil, fmt.Errorf("生成Token失败")
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: map[string]interface{}{
		"access_token":       accessToken,
		"refresh_token":      refreshToken,
		"expires_in":         accessExpiresIn,
		"refresh_expires_in": client.Timeout,
		"client_id":          client.ClientId,
		"open_id":            user.OpenID,
		"user_info": map[string]interface{}{
			"userId":      user.Id,
			"username":    user.UserName,
			"nickname":    user.NickName,
			"phonenumber": user.Phonenumber,
			"email":       user.Email,
			"avatar":      user.Avatar,
			"userType":    user.UserType,
			"orgId":       user.OrgID,
			"roles":       user.Roles,
			"permissions": user.Permissions,
			"openId":      user.OpenID,
			"unionId":     user.UnionID,
		},
	}}, nil
}

func refreshLoginToken(ctx context.Context, svcCtx *svc.ServiceContext, refreshToken string, clientId string) (*types.CommonResp, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("RefreshToken 无效或已过期")
	}
	refreshKey := commonauth.RefreshTokenKey(refreshToken)
	encoded, err := svcCtx.Redis.GetDel(ctx, refreshKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("RefreshToken 无效或已过期")
		}
		return nil, fmt.Errorf("RefreshToken 无效或已过期")
	}
	var session tokenSession
	if json.Unmarshal([]byte(encoded), &session) != nil {
		return nil, fmt.Errorf("RefreshToken 会话数据无效")
	}
	// refresh flow uses cached client metadata; clientId must match the stored token owner
	if session.ClientKey != "" && session.ClientKey != clientId {
		return nil, fmt.Errorf("客户端不匹配")
	}
	_ = svcCtx.Redis.Del(ctx, commonauth.AccessTokenHashKey(session.AccessHash)).Err()
	_ = svcCtx.Redis.SRem(ctx, commonauth.UserSessionsKey(session.UserID, session.ClientID), refreshKey).Err()
	userResp, err := svcCtx.IamRpcClient.UserProfile(ctx, &iamservice.IdReq{Id: session.UserID})
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	client := &authClient{ClientId: session.ClientID, ClientKey: clientId, DeviceType: session.DeviceType, Timeout: session.Timeout, ActiveTimeout: session.ActiveTimeout}
	user := &loginUser{Id: userResp.UserId, UserName: userResp.UserName, NickName: userResp.NickName, Email: userResp.Email, Phonenumber: userResp.Phonenumber, Avatar: userResp.Avatar, UserType: int64(userResp.UserType), OrgID: userResp.OrgId, OpenID: userResp.OpenId, UnionID: userResp.UnionId}
	if err := enrichLoginUserAuthContext(ctx, svcCtx, user); err != nil {
		return nil, fmt.Errorf("用户权限上下文获取失败")
	}
	accessToken, accessExpiresIn, err := generateAccessToken(user, client, svcCtx.Config.Jwt.Secret)
	if err != nil {
		return nil, fmt.Errorf("生成新Token失败")
	}
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("生成新Token失败")
	}
	if err := storeTokenSession(ctx, svcCtx, user, client, accessToken, newRefreshToken, false); err != nil {
		return nil, fmt.Errorf("更新RefreshToken失败")
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: map[string]interface{}{"access_token": accessToken, "refresh_token": newRefreshToken, "expires_in": accessExpiresIn, "refresh_expires_in": client.Timeout}}, nil
}

func invalidateByToken(ctx context.Context, svcCtx *svc.ServiceContext, token string) {
	if token == "" {
		return
	}
	accessKey := commonauth.AccessTokenKey(token)
	refreshKey, err := svcCtx.Redis.Get(ctx, accessKey).Result()
	if err != nil {
		return
	}
	encoded, _ := svcCtx.Redis.Get(ctx, refreshKey).Result()
	var session tokenSession
	_ = json.Unmarshal([]byte(encoded), &session)
	pipe := svcCtx.Redis.TxPipeline()
	pipe.Del(ctx, accessKey, refreshKey)
	pipe.SRem(ctx, commonauth.UserSessionsKey(session.UserID, session.ClientID), refreshKey)
	_, _ = pipe.Exec(ctx)
}

func generateAccessToken(user *loginUser, client *authClient, secret string) (string, int64, error) {
	expireSeconds := client.ActiveTimeout
	if expireSeconds <= 0 {
		expireSeconds = 1800
	}
	expireAt := time.Now().Add(time.Duration(expireSeconds) * time.Second)
	claims := commonauth.AccessClaims{
		UserID:      user.Id,
		UserName:    user.UserName,
		ClientID:    client.ClientId,
		DeviceType:  client.DeviceType,
		OrgID:       user.OrgID,
		Roles:       user.Roles,
		Permissions: user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "MS_K-gozero",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return tokenString, expireSeconds, nil
}

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func storeTokenSession(ctx context.Context, svcCtx *svc.ServiceContext, user *loginUser, client *authClient, accessToken, refreshToken string, revokeExisting bool) error {
	if revokeExisting {
		if err := revokeUserSessions(ctx, svcCtx, user.Id, client.ClientId); err != nil {
			return err
		}
	}
	accessTTL := client.ActiveTimeout
	if accessTTL <= 0 {
		accessTTL = 1800
	}
	if client.Timeout <= 0 {
		return fmt.Errorf("RefreshToken 过期时间必须大于0")
	}
	session := tokenSession{UserID: user.Id, UserName: user.UserName, OrgID: user.OrgID, ClientID: client.ClientId, ClientKey: client.ClientKey, DeviceType: client.DeviceType, Timeout: client.Timeout, ActiveTimeout: client.ActiveTimeout, AccessHash: commonauth.TokenHash(accessToken)}
	encoded, err := json.Marshal(session)
	if err != nil {
		return err
	}
	refreshKey := commonauth.RefreshTokenKey(refreshToken)
	setKey := commonauth.UserSessionsKey(user.Id, client.ClientId)
	pipe := svcCtx.Redis.TxPipeline()
	pipe.Set(ctx, commonauth.AccessTokenKey(accessToken), refreshKey, time.Duration(accessTTL)*time.Second)
	pipe.Set(ctx, refreshKey, encoded, time.Duration(client.Timeout)*time.Second)
	pipe.SAdd(ctx, setKey, refreshKey)
	pipe.Expire(ctx, setKey, time.Duration(client.Timeout)*time.Second)
	_, err = pipe.Exec(ctx)
	return err
}

func revokeUserSessions(ctx context.Context, svcCtx *svc.ServiceContext, userID int64, clientID string) error {
	setKey := commonauth.UserSessionsKey(userID, clientID)
	refreshKeys, err := svcCtx.Redis.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}
	pipe := svcCtx.Redis.TxPipeline()
	for _, refreshKey := range refreshKeys {
		encoded, err := svcCtx.Redis.Get(ctx, refreshKey).Bytes()
		if err == nil {
			var session tokenSession
			if json.Unmarshal(encoded, &session) == nil {
				pipe.Del(ctx, commonauth.AccessTokenHashKey(session.AccessHash))
			}
		}
		pipe.Del(ctx, refreshKey)
	}
	pipe.Del(ctx, setKey)
	_, err = pipe.Exec(ctx)
	return err
}

func enrichLoginUserAuthContext(ctx context.Context, svcCtx *svc.ServiceContext, user *loginUser) error {
	authCtx, err := svcCtx.IamRpcClient.UserAuthContext(ctx, &iamservice.UserAuthContextReq{UserId: user.Id})
	if err != nil {
		return fmt.Errorf("获取用户权限上下文失败: %w", err)
	}
	user.OrgID = authCtx.OrgId
	user.Roles = authCtx.Roles
	user.Permissions = authCtx.Permissions
	return nil
}
