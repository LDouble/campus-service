package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`              // 业务状态码
	Message string      `json:"message"`           // 提示信息
	Data    interface{} `json:"data,omitempty"`    // 响应数据
	Error   *ErrorInfo  `json:"error,omitempty"`   // 错误详情
}

// ErrorInfo 错误详情
type ErrorInfo struct {
	Type    string `json:"type,omitempty"`    // 错误类型
	Detail  string `json:"detail,omitempty"`  // 错误详情
	Request string `json:"request,omitempty"` // 请求ID
}

// 业务状态码定义
const (
	CodeSuccess         = 0
	CodeBadRequest      = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeConflict        = 409
	CodeInternalError   = 500
	CodeServiceUnavailable = 503
	
	// 业务错误码 (1000+)
	CodeInvalidToken    = 1001
	CodeTokenExpired    = 1002
	CodeUserNotFound    = 1003
	CodeUserDisabled    = 1004
	CodeWeChatError     = 1005
	CodeInvalidParam    = 1006
	CodeBindingRequired = 1010 // 需要教务绑定
)

// 状态码消息映射
var codeMessages = map[int]string{
	CodeSuccess:         "success",
	CodeBadRequest:      "请求参数错误",
	CodeUnauthorized:    "未授权访问",
	CodeForbidden:      "禁止访问",
	CodeNotFound:        "资源不存在",
	CodeConflict:       "资源冲突",
	CodeInternalError:  "服务器内部错误",
	CodeServiceUnavailable: "服务暂不可用",
	CodeInvalidToken:   "无效的令牌",
	CodeTokenExpired:   "令牌已过期",
	CodeUserNotFound:   "用户不存在",
	CodeUserDisabled:   "用户已被禁用",
	CodeWeChatError:    "微信登录失败",
	CodeInvalidParam:   "参数验证失败",
	CodeBindingRequired: "请先绑定教务系统",
}

// GetMessage 获取状态码对应的消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: GetMessage(CodeSuccess),
		Data:    data,
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeSuccess,
		Message: "创建成功",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code int, message string) {
	if message == "" {
		message = GetMessage(code)
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Error: &ErrorInfo{
			Type: "business_error",
		},
	})
}

// BadRequest 参数错误响应
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, code int, message string) {
	if code == 0 {
		code = CodeUnauthorized
	}
	Error(c, http.StatusUnauthorized, code, message)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

// InternalError 服务器错误响应
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeInternalError, message)
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeInvalidParam,
		Message: GetMessage(CodeInvalidParam),
		Data:    errors,
		Error: &ErrorInfo{
			Type: "validation_error",
		},
	})
}