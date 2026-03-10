package jwt

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClaims 从上下文获取JWT Claims，如果不存在则解析token
func GetClaims(c *gin.Context) (*Claims, error) {
	if v, exists := c.Get("jwt_claims"); exists {
		if claims, ok := v.(*Claims); ok {
			return claims, nil
		}
	}

	// 回退：从header解析
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, ErrInvalidToken
	}

	return ParseAccessToken(parts[1])
}