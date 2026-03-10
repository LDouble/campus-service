package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"campus-service/internal/middleware"
	"campus-service/internal/pkg/response"
	"campus-service/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	User         *UserDTO    `json:"user"`
	IsNewUser    bool        `json:"is_new_user"`
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID        uint   `json:"id"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Phone     string `json:"phone"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
	Avatar   string `json:"avatar" binding:"omitempty,url,max=256"`
	Phone    string `json:"phone" binding:"omitempty,len=11"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Login 微信登录
// @Summary 微信登录
// @Description 通过微信小程序code换取token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=LoginResponse}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.authService.WeChatLogin(context.Background(), req.Code)
	if err != nil {
		response.Error(c, 401, response.CodeWeChatError, err.Error())
		return
	}

	response.Success(c, &LoginResponse{
		AccessToken:  result.Token.AccessToken,
		RefreshToken: result.Token.RefreshToken,
		ExpiresIn:    result.Token.ExpiresIn,
		TokenType:    result.Token.TokenType,
		User:         toUserDTO(result.User),
		IsNewUser:    result.IsNewUser,
	})
}

// GetProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的个人信息
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response{data=UserDTO}
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	user, err := h.authService.GetProfile(context.Background(), userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, toUserDTO(user))
}

// UpdateProfile 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前登录用户的个人信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "更新请求"
// @Success 200 {object} response.Response{data=UserDTO}
// @Router /api/v1/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 处理验证错误
		if errs, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, e := range errs {
				errors[e.Field()] = getErrorMsg(e)
			}
			response.ValidationError(c, errors)
			return
		}
		response.BadRequest(c, "参数错误")
		return
	}

	user, err := h.authService.UpdateProfile(context.Background(), userID, req.Nickname, req.Avatar, req.Phone)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toUserDTO(user))
}

// RefreshToken 刷新token
// @Summary 刷新令牌
// @Description 使用refresh_token获取新的access_token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新请求"
// @Success 200 {object} response.Response{data=RefreshTokenResponse}
// @Router /api/v1/auth/token/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	tokenPair, err := h.authService.RefreshToken(context.Background(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, response.CodeInvalidToken, err.Error())
		return
	}

	response.Success(c, &RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	})
}

// toUserDTO 转换用户模型到DTO
func toUserDTO(user interface{}) *UserDTO {
	switch u := user.(type) {
	case *struct {
		ID        uint
		Nickname  string
		Avatar    string
		Phone     string
		Status    int
		CreatedAt string
	}:
		return &UserDTO{
			ID:        u.ID,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Phone:     u.Phone,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
		}
	default:
		return nil
	}
}

// getErrorMsg 获取验证错误消息
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "此字段为必填项"
	case "max":
		return "长度不能超过" + fe.Param()
	case "min":
		return "长度不能少于" + fe.Param()
	case "len":
		return "长度必须为" + fe.Param()
	case "url":
		return "必须是有效的URL"
	case "email":
		return "必须是有效的邮箱地址"
	default:
		return "参数验证失败"
	}
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/token/refresh", h.RefreshToken)
		auth.GET("/profile", middleware.Auth(), h.GetProfile)
		auth.PUT("/profile", middleware.Auth(), h.UpdateProfile)
	}
}

// 初始化验证器
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证规则
		_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
			phone := fl.Field().String()
			return len(phone) == 0 || len(phone) == 11
		})
	}
}