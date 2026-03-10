package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"campus-service/internal/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器实例
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Health 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response{data=HealthResponse}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, &HealthResponse{
		Status:  "ok",
		Message: "服务运行正常",
	})
}

// Ready 就绪检查
// @Summary 就绪检查
// @Description 检查服务是否就绪
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response{data=HealthResponse}
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// TODO: 检查数据库和Redis连接
	response.Success(c, &HealthResponse{
		Status:  "ready",
		Message: "服务就绪",
	})
}

// RegisterRoutes 注册路由
func (h *HealthHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", h.Health)
	r.GET("/ready", h.Ready)
}

// NotFoundHandler 404处理器
func NotFoundHandler(c *gin.Context) {
	response.Error(c, http.StatusNotFound, response.CodeNotFound, "请求的资源不存在")
}

// MethodNotAllowedHandler 方法不允许处理器
func MethodNotAllowedHandler(c *gin.Context) {
	response.Error(c, http.StatusMethodNotAllowed, response.CodeBadRequest, "不支持的请求方法")
}