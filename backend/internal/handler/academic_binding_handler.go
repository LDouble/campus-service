package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"campus-service/internal/middleware"
	"campus-service/internal/model"
	"campus-service/internal/pkg/response"
	"campus-service/internal/service"
)

// AcademicBindingHandler 教务绑定处理器
type AcademicBindingHandler struct {
	bindingService service.AcademicBindingService
}

// NewAcademicBindingHandler 创建教务绑定处理器实例
func NewAcademicBindingHandler(bindingService service.AcademicBindingService) *AcademicBindingHandler {
	return &AcademicBindingHandler{
		bindingService: bindingService,
	}
}

// BindRequest 绑定请求
type BindRequest struct {
	StudentID string `json:"student_id" binding:"required,max=32"`
	Password  string `json:"password" binding:"required,max=64"`
	School    string `json:"school" binding:"required,max=64"`
}

// BindingDTO 绑定数据传输对象
type BindingDTO struct {
	ID           uint   `json:"id"`
	StudentID    string `json:"student_id"`
	School       string `json:"school"`
	Status       int    `json:"status"`
	VerifyStatus int    `json:"verify_status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Bind 绑定教务系统
// @Summary 绑定教务系统
// @Description 绑定用户的教务系统账号
// @Tags 教务绑定
// @Accept json
// @Produce json
// @Param request body BindRequest true "绑定请求"
// @Success 200 {object} response.Response{data=BindingDTO}
// @Router /api/v1/academic/bind [post]
func (h *AcademicBindingHandler) Bind(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	var req BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

	binding, err := h.bindingService.Bind(context.Background(), userID, req.StudentID, req.Password, req.School)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toBindingDTO(binding))
}

// Unbind 解绑教务系统
// @Summary 解绑教务系统
// @Description 解绑用户的教务系统账号
// @Tags 教务绑定
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/academic/unbind [delete]
func (h *AcademicBindingHandler) Unbind(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	err := h.bindingService.Unbind(context.Background(), userID)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetBinding 获取绑定信息
// @Summary 获取绑定信息
// @Description 获取用户的教务系统绑定信息
// @Tags 教务绑定
// @Produce json
// @Success 200 {object} response.Response{data=BindingDTO}
// @Router /api/v1/academic/binding [get]
func (h *AcademicBindingHandler) GetBinding(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	binding, err := h.bindingService.GetBinding(context.Background(), userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, toBindingDTO(binding))
}

// VerifyBinding 验证绑定
// @Summary 验证绑定
// @Description 验证用户的教务系统绑定状态
// @Tags 教务绑定
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/academic/verify [post]
func (h *AcademicBindingHandler) VerifyBinding(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	err := h.bindingService.VerifyBinding(context.Background(), userID)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// toBindingDTO 转换绑定模型到DTO
func toBindingDTO(binding *model.AcademicBinding) *BindingDTO {
	return &BindingDTO{
		ID:           binding.ID,
		StudentID:    binding.StudentID,
		School:       binding.School,
		Status:       binding.Status,
		VerifyStatus: binding.VerifyStatus,
		CreatedAt:    binding.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    binding.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// RegisterRoutes 注册路由
func (h *AcademicBindingHandler) RegisterRoutes(r *gin.RouterGroup) {
	academic := r.Group("/academic")
	academic.Use(middleware.Auth())
	{
		academic.POST("/bind", h.Bind)
		academic.DELETE("/unbind", h.Unbind)
		academic.GET("/binding", h.GetBinding)
		academic.POST("/verify", h.VerifyBinding)
	}
}