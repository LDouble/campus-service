package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"campus-service/internal/config"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotValidYet = errors.New("token not valid yet")
)

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	OpenID   string `json:"openid"`
	TokenType string `json:"token_type"` // access or refresh
	jwt.RegisteredClaims
}

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access_token 过期时间(秒)
	TokenType    string `json:"token_type"` // Bearer
}

// GenerateTokenPair 生成令牌对
func GenerateTokenPair(userID uint, openID string) (*TokenPair, error) {
	now := time.Now()
	
	// 生成 access token
	accessExpiry := now.Add(config.Cfg.JWT.AccessTokenExpiry)
	accessClaims := Claims{
		UserID:    userID,
		OpenID:    openID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    config.Cfg.JWT.Issuer,
		},
	}
	
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(config.Cfg.JWT.AccessTokenSecret))
	if err != nil {
		return nil, err
	}

	// 生成 refresh token
	refreshExpiry := now.Add(config.Cfg.JWT.RefreshTokenExpiry)
	refreshClaims := Claims{
		UserID:    userID,
		OpenID:    openID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    config.Cfg.JWT.Issuer,
		},
	}
	
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(config.Cfg.JWT.RefreshTokenSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(config.Cfg.JWT.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ParseAccessToken 解析 access token
func ParseAccessToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.Cfg.JWT.AccessTokenSecret)
}

// ParseRefreshToken 解析 refresh token
func ParseRefreshToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.Cfg.JWT.RefreshTokenSecret)
}

// parseToken 解析令牌
func parseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}