package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"

	"campus-service/internal/config"
	"campus-service/internal/model"
	"campus-service/internal/pkg/jwt"
	redisClient "campus-service/internal/model"
	"campus-service/internal/repository"
)

// WeChatLoginResponse 微信登录响应
type WeChatLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// LoginResult 登录结果
type LoginResult struct {
	Token     *jwt.TokenPair `json:"token"`
	User      *model.User    `json:"user"`
	IsNewUser bool           `json:"is_new_user"`
}

// AuthService 认证服务接口
type AuthService interface {
	WeChatLogin(ctx context.Context, code string) (*LoginResult, error)
	GetProfile(ctx context.Context, userID uint) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uint, nickname, avatar, phone string) (*model.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
}

// authService 认证服务实现
type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// WeChatLogin 微信登录
func (s *authService) WeChatLogin(ctx context.Context, code string) (*LoginResult, error) {
	// 调用微信API获取openid
	wxResp, err := s.getWeChatOpenID(code)
	if err != nil {
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	if wxResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s", wxResp.ErrMsg)
	}

	// 查找或创建用户
	user, isNew, err := s.findOrCreateUser(ctx, wxResp.OpenID, wxResp.UnionID)
	if err != nil {
		return nil, err
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("用户已被禁用")
	}

	// 生成token
	tokenPair, err := jwt.GenerateTokenPair(user.ID, user.OpenID)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 存储token到Redis
	if err := s.storeTokens(ctx, user.ID, tokenPair); err != nil {
		return nil, fmt.Errorf("存储令牌失败: %w", err)
	}

	return &LoginResult{
		Token:     tokenPair,
		User:      user,
		IsNewUser: isNew,
	}, nil
}

// GetProfile 获取用户信息
func (s *authService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return user, nil
}

// UpdateProfile 更新用户信息
func (s *authService) UpdateProfile(ctx context.Context, userID uint, nickname, avatar, phone string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	user.UpdateProfile(nickname, avatar, phone)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// RefreshToken 刷新token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	// 解析refresh token
	claims, err := jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 验证refresh token是否在Redis中
	storedToken, err := redisClient.GetRefreshToken(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("无效的刷新令牌")
	}

	if storedToken != refreshToken {
		return nil, errors.New("刷新令牌不匹配")
	}

	// 检查用户状态
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if !user.IsActive() {
		return nil, errors.New("用户已被禁用")
	}

	// 生成新的token对
	tokenPair, err := jwt.GenerateTokenPair(user.ID, user.OpenID)
	if err != nil {
		return nil, err
	}

	// 存储新token
	if err := s.storeTokens(ctx, user.ID, tokenPair); err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// getWeChatOpenID 调用微信API获取openid
func (s *authService) getWeChatOpenID(code string) (*WeChatLoginResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		config.Cfg.WeChat.AppID,
		config.Cfg.WeChat.AppSecret,
		code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wxResp WeChatLoginResponse
	if err := json.Unmarshal(body, &wxResp); err != nil {
		return nil, err
	}

	return &wxResp, nil
}

// findOrCreateUser 查找或创建用户
func (s *authService) findOrCreateUser(ctx context.Context, openID, unionID string) (*model.User, bool, error) {
	// 尝试查找用户
	user, err := s.userRepo.FindByOpenID(ctx, openID)
	if err == nil {
		return user, false, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	// 创建新用户
	user = &model.User{
		OpenID:  openID,
		UnionID: unionID,
		Status:  model.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, false, err
	}

	return user, true, nil
}

// storeTokens 存储token到Redis
func (s *authService) storeTokens(ctx context.Context, userID uint, tokenPair *jwt.TokenPair) error {
	// 存储access token
	if err := redisClient.SetAccessToken(ctx, userID, tokenPair.AccessToken, config.Cfg.JWT.AccessTokenExpiry); err != nil {
		return err
	}

	// 存储refresh token
	if err := redisClient.SetRefreshToken(ctx, userID, tokenPair.RefreshToken, config.Cfg.JWT.RefreshTokenExpiry); err != nil {
		return err
	}

	return nil
}