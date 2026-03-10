package middleware

import (
	"context"

	"github.com/gin-gonic/gin"

	"campus-service/internal/pkg/response"
	"campus-service/internal/service"
)

// BindingCheck 教务绑定检查中间件
func BindingCheck(bindingService service.AcademicBindingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, response.CodeUnauthorized, "未登录")
			c.Abort()
			return
		}

		// 检查用户是否已绑定教务系统
		isBound, err := bindingService.IsUserBound(context.Background(), userID)
		if err != nil {
			response.Error(c, 500, response.CodeInternalError, "检查绑定状态失败")
			c.Abort()
			return
		}

		if !isBound {
			response.Error(c, 403, response.CodeBindingRequired, "请先绑定教务系统")
			c.Abort()
			return
		}

		c.Next()
	}
}