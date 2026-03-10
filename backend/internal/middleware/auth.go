package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"campus-service/internal/pkg/jwt"
	"campus-service/internal/pkg/response"
)

const (
	// ContextKeyUserID 用户ID上下文键
	ContextKeyUserID = "user_id"
	// ContextKeyOpenID OpenID上下文键
	ContextKeyOpenID = "openid"
)

// Auth JWT认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, response.CodeUnauthorized, "请先登录")
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, response.CodeInvalidToken, "无效的认证格式")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析token
		claims, err := jwt.ParseAccessToken(tokenString)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				response.Unauthorized(c, response.CodeTokenExpired, "令牌已过期，请刷新")
			} else {
				response.Unauthorized(c, response.CodeInvalidToken, "无效的令牌")
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyOpenID, claims.OpenID)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		return userID.(uint)
	}
	return 0
}

// GetOpenID 从上下文获取OpenID
func GetOpenID(c *gin.Context) string {
	if openID, exists := c.Get(ContextKeyOpenID); exists {
		return openID.(string)
	}
	return ""
}