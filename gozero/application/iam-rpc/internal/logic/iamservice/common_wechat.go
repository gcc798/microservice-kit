package iamservicelogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"gorm.io/gorm"
)

const (
	miniProgramUserType     int32 = 1
	defaultMiniProgramOrgID int64 = 1880159541355577346
)

type wechatCode2SessionResp struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func wechatCode2Session(appId, secret, wxCode string) (*wechatCode2SessionResp, error) {
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(appId), url.QueryEscape(secret), url.QueryEscape(wxCode))
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("调用微信接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取微信响应失败: %w", err)
	}
	var result wechatCode2SessionResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信接口错误: %s", result.ErrMsg)
	}
	if result.OpenId == "" {
		return nil, fmt.Errorf("获取微信OpenID失败")
	}
	return &result, nil
}

func authenticateXcx(ctx context.Context, svcCtx *svc.ServiceContext, phonenumber, code, wxCode string) (*userAuthRow, error) {
	if phonenumber == "" || code == "" || wxCode == "" {
		return nil, fmt.Errorf("手机号、验证码和微信code不能为空")
	}
	if !svcCtx.Config.Wechat.Enabled {
		return nil, fmt.Errorf("微信小程序登录未启用")
	}
	wxResp, err := wechatCode2Session(svcCtx.Config.Wechat.AppId, svcCtx.Config.Wechat.Secret, wxCode)
	if err != nil {
		return nil, err
	}

	var row userAuthRow
	err = svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("phonenumber = ?", phonenumber).Take(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		user := model.SUser{
			UserName: phonenumber, NickName: nullableString(phonenumber), UserType: int64(miniProgramUserType),
			OrgId: defaultMiniProgramOrgID, Phonenumber: nullableString(phonenumber), Status: 0, Sex: 2,
			OpenId: nullableString(wxResp.OpenId), UnionId: nullableString(wxResp.UnionId),
			LoginDate:   sql.NullInt64{Int64: now.Unix(), Valid: true},
			CreatedTime: sql.NullTime{Time: now, Valid: true}, UpdatedTime: sql.NullTime{Time: now, Valid: true},
		}
		if err := gorm.G[model.SUser](svcCtx.DB).Create(ctx, &user); err != nil {
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}
		row = userAuthRow{Id: user.Id, OrgId: sql.NullInt64{Int64: user.OrgId, Valid: true}, UserName: user.UserName,
			NickName: user.NickName, UserType: user.UserType, Email: user.Email, Phonenumber: user.Phonenumber,
			Avatar: user.Avatar, Password: user.Password, Status: user.Status}
	} else if err := svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("id = ?", row.Id).Updates(map[string]any{
		"open_id": nullableString(wxResp.OpenId), "union_id": nullableString(wxResp.UnionId),
		"login_date": time.Now().Unix(), "updated_time": time.Now(),
	}).Error; err != nil {
		return nil, fmt.Errorf("更新微信用户信息失败: %w", err)
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	return &row, nil
}

func authenticateWechat(ctx context.Context, svcCtx *svc.ServiceContext, wxCode string) (*userAuthRow, error) {
	if wxCode == "" {
		return nil, fmt.Errorf("微信code不能为空")
	}
	if !svcCtx.Config.Wechat.Enabled {
		return nil, fmt.Errorf("微信小程序登录未启用")
	}
	wxResp, err := wechatCode2Session(svcCtx.Config.Wechat.AppId, svcCtx.Config.Wechat.Secret, wxCode)
	if err != nil {
		return nil, err
	}

	var row userAuthRow
	err = svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("open_id = ?", wxResp.OpenId).Take(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		userName := "wx_" + wxResp.OpenId[:min(16, len(wxResp.OpenId))]
		now := time.Now()
		user := model.SUser{
			UserName: userName, NickName: nullableString("微信用户"), UserType: int64(miniProgramUserType),
			OrgId: defaultMiniProgramOrgID, Status: 0, Sex: 2, OpenId: nullableString(wxResp.OpenId),
			UnionId: nullableString(wxResp.UnionId), LoginDate: sql.NullInt64{Int64: now.Unix(), Valid: true},
			CreatedTime: sql.NullTime{Time: now, Valid: true}, UpdatedTime: sql.NullTime{Time: now, Valid: true},
		}
		if err := gorm.G[model.SUser](svcCtx.DB).Create(ctx, &user); err != nil {
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}
		row = userAuthRow{Id: user.Id, OrgId: sql.NullInt64{Int64: user.OrgId, Valid: true}, UserName: user.UserName,
			NickName: user.NickName, UserType: user.UserType, Email: user.Email, Phonenumber: user.Phonenumber,
			Avatar: user.Avatar, Password: user.Password, Status: user.Status}
	} else if wxResp.UnionId != "" {
		_ = svcCtx.DB.WithContext(ctx).Model(&model.SUser{}).Where("id = ?", row.Id).Updates(map[string]any{
			"union_id": wxResp.UnionId, "login_date": time.Now().Unix(), "updated_time": time.Now(),
		}).Error
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	return &row, nil
}
