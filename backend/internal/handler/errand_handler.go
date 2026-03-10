package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"campus-service/internal/middleware"
	"campus-service/internal/model"
	"campus-service/internal/pkg/response"
	"campus-service/internal/service"
)

// ErrandHandler 跑腿处理器
type ErrandHandler struct {
	errandService service.ErrandService
}

// NewErrandHandler 创建跑腿处理器实例
func NewErrandHandler(errandService service.ErrandService) *ErrandHandler {
	return &ErrandHandler{
		errandService: errandService,
	}
}

// ErrandDTO 跑腿任务数据传输对象
type ErrandDTO struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	TypeName    string     `json:"type_name"`
	Description string     `json:"description"`
	Reward      float64    `json:"reward"`
	Location    string     `json:"location"`
	Deadline    *time.Time `json:"deadline"`
	Status      string     `json:"status"`
	StatusName  string     `json:"status_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	User        *UserBrief `json:"user,omitempty"`
}

// UserBrief 用户简要信息
type UserBrief struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone,omitempty"`
}

// ErrandDetailDTO 跑腿任务详情DTO
type ErrandDetailDTO struct {
	ErrandDTO
	Order *OrderDTO `json:"order,omitempty"`
}

// OrderDTO 订单数据传输对象
type OrderDTO struct {
	ID        uint      `json:"id"`
	ErrandID  uint      `json:"errand_id"`
	UserID    uint      `json:"user_id"`
	Status    string    `json:"status"`
	StatusName string   `json:"status_name"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	User      *UserBrief `json:"user,omitempty"`
}

// PublishErrandRequest 发布跑腿任务请求
type PublishErrandRequest struct {
	Title       string     `json:"title" binding:"required,max=100"`
	Type        string     `json:"type" binding:"required,oneof=express food print other"`
	Description string     `json:"description" binding:"max=500"`
	Reward      float64    `json:"reward" binding:"min=0"`
	Location    string     `json:"location" binding:"max=200"`
	Deadline    *time.Time `json:"deadline"`
}

// UpdateErrandRequest 更新跑腿任务请求
type UpdateErrandRequest struct {
	Title       string     `json:"title" binding:"omitempty,max=100"`
	Description string     `json:"description" binding:"omitempty,max=500"`
	Reward      float64    `json:"reward" binding:"omitempty,min=0"`
	Location    string     `json:"location" binding:"omitempty,max=200"`
	Deadline    *time.Time `json:"deadline"`
}

// AcceptErrandRequest 接单请求
type AcceptErrandRequest struct {
	Remark string `json:"remark" binding:"max=200"`
}

// CancelErrandRequest 取消任务请求
type CancelErrandRequest struct {
	Reason string `json:"reason" binding:"max=200"`
}

// ErrandListResponse 跑腿任务列表响应
type ErrandListResponse struct {
	List     []ErrandDTO `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List     []OrderDTO `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Publish 发布跑腿任务
// @Summary 发布跑腿任务
// @Description 发布一个新的跑腿任务
// @Tags 跑腿
// @Accept json
// @Produce json
// @Param request body PublishErrandRequest true "发布请求"
// @Success 200 {object} response.Response{data=ErrandDTO}
// @Router /api/v1/errands [post]
func (h *ErrandHandler) Publish(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	var req PublishErrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	// 验证截止时间
	if req.Deadline != nil && req.Deadline.Before(time.Now()) {
		response.BadRequest(c, "截止时间不能早于当前时间")
		return
	}

	errand, err := h.errandService.PublishErrand(c.Request.Context(), userID, &service.PublishErrandRequest{
		Title:       req.Title,
		Type:        req.Type,
		Description: req.Description,
		Reward:      req.Reward,
		Location:    req.Location,
		Deadline:    req.Deadline,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Created(c, toErrandDTO(*errand))
}

// List 获取跑腿任务列表
// @Summary 获取跑腿任务列表
// @Description 分页获取跑腿任务列表
// @Tags 跑腿
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "任务类型" Enums(express, food, print, other)
// @Param status query string false "任务状态" Enums(pending, in_progress, completed, cancelled)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} response.Response{data=ErrandListResponse}
// @Router /api/v1/errands [get]
func (h *ErrandHandler) List(c *gin.Context) {
	var query model.ErrandListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		handleBindError(c, err)
		return
	}

	result, err := h.errandService.GetErrandList(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]ErrandDTO, 0, len(result.List))
	for _, e := range result.List {
		list = append(list, toErrandDTO(e))
	}

	response.Success(c, &ErrandListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// Detail 获取跑腿任务详情
// @Summary 获取跑腿任务详情
// @Description 根据ID获取跑腿任务详情
// @Tags 跑腿
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=ErrandDetailDTO}
// @Router /api/v1/errands/{id} [get]
func (h *ErrandHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	result, err := h.errandService.GetErrandDetail(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "跑腿任务不存在" {
			response.NotFound(c, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	dto := toErrandDTO(result.Errand)
	detailDTO := ErrandDetailDTO{
		ErrandDTO: dto,
	}

	if result.Order != nil {
		orderDTO := toOrderDTO(*result.Order)
		detailDTO.Order = &orderDTO
	}

	response.Success(c, detailDTO)
}

// Update 更新跑腿任务
// @Summary 更新跑腿任务
// @Description 更新跑腿任务信息
// @Tags 跑腿
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param request body UpdateErrandRequest true "更新请求"
// @Success 200 {object} response.Response{data=ErrandDTO}
// @Router /api/v1/errands/{id} [put]
func (h *ErrandHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req UpdateErrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	errand, err := h.errandService.UpdateErrand(c.Request.Context(), userID, uint(id), &service.UpdateErrandRequest{
		Title:       req.Title,
		Description: req.Description,
		Reward:      req.Reward,
		Location:    req.Location,
		Deadline:    req.Deadline,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toErrandDTO(*errand))
}

// Delete 删除跑腿任务
// @Summary 删除跑腿任务
// @Description 删除跑腿任务（软删除）
// @Tags 跑腿
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response
// @Router /api/v1/errands/{id} [delete]
func (h *ErrandHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.errandService.DeleteErrand(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Accept 接单
// @Summary 接受跑腿任务
// @Description 接受一个跑腿任务
// @Tags 跑腿
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param request body AcceptErrandRequest true "接单请求"
// @Success 200 {object} response.Response{data=OrderDTO}
// @Router /api/v1/errands/{id}/accept [post]
func (h *ErrandHandler) Accept(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req AcceptErrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// remark 可选，忽略错误
		req.Remark = ""
	}

	order, err := h.errandService.AcceptErrand(c.Request.Context(), uint(id), userID, req.Remark)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toOrderDTO(*order))
}

// Cancel 取消任务
// @Summary 取消跑腿任务
// @Description 取消跑腿任务
// @Tags 跑腿
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param request body CancelErrandRequest true "取消请求"
// @Success 200 {object} response.Response
// @Router /api/v1/errands/{id}/cancel [post]
func (h *ErrandHandler) Cancel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req CancelErrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = ""
	}

	if err := h.errandService.CancelErrand(c.Request.Context(), uint(id), userID, req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Complete 完成任务
// @Summary 完成跑腿任务
// @Description 确认完成跑腿任务
// @Tags 跑腿
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response
// @Router /api/v1/errands/{id}/complete [post]
func (h *ErrandHandler) Complete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.errandService.CompleteErrand(c.Request.Context(), uint(id), userID); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// MyPublished 获取我发布的任务
// @Summary 获取我发布的任务
// @Description 获取当前用户发布的跑腿任务列表
// @Tags 跑腿
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=ErrandListResponse}
// @Router /api/v1/errands/my/published [get]
func (h *ErrandHandler) MyPublished(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.errandService.GetMyErrands(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]ErrandDTO, 0, len(result.List))
	for _, e := range result.List {
		list = append(list, toErrandDTO(e))
	}

	response.Success(c, &ErrandListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// MyAccepted 获取我接的任务
// @Summary 获取我接的任务
// @Description 获取当前用户接受的跑腿订单列表
// @Tags 跑腿
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=OrderListResponse}
// @Router /api/v1/errands/my/accepted [get]
func (h *ErrandHandler) MyAccepted(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.errandService.GetMyOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]OrderDTO, 0, len(result.List))
	for _, o := range result.List {
		list = append(list, toOrderDTO(o))
	}

	response.Success(c, &OrderListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// RegisterRoutes 注册路由
func (h *ErrandHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	errands := r.Group("/errands")
	{
		// 公开路由
		errands.GET("", h.List)
		errands.GET("/:id", h.Detail)

		// 需要认证的路由
		authGroup := errands.Group("")
		authGroup.Use(authMiddleware)
		{
			authGroup.POST("", h.Publish)
			authGroup.PUT("/:id", h.Update)
			authGroup.DELETE("/:id", h.Delete)
			authGroup.POST("/:id/accept", h.Accept)
			authGroup.POST("/:id/cancel", h.Cancel)
			authGroup.POST("/:id/complete", h.Complete)
			authGroup.GET("/my/published", h.MyPublished)
			authGroup.GET("/my/accepted", h.MyAccepted)
		}
	}
}

// 辅助函数

// toErrandDTO 转换跑腿任务到DTO
func toErrandDTO(e model.Errand) ErrandDTO {
	dto := ErrandDTO{
		ID:          e.ID,
		UserID:      e.UserID,
		Title:       e.Title,
		Type:        e.Type,
		TypeName:    service.FormatErrandType(e.Type),
		Description: e.Description,
		Reward:      e.Reward,
		Location:    e.Location,
		Deadline:    e.Deadline,
		Status:      e.Status,
		StatusName:  service.FormatErrandStatus(e.Status),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	if e.User.ID != 0 {
		dto.User = &UserBrief{
			ID:       e.User.ID,
			Nickname: e.User.Nickname,
			Avatar:   e.User.Avatar,
		}
	}

	return dto
}

// toOrderDTO 转换订单到DTO
func toOrderDTO(o model.ErrandOrder) OrderDTO {
	dto := OrderDTO{
		ID:         o.ID,
		ErrandID:   o.ErrandID,
		UserID:     o.UserID,
		Status:     o.Status,
		StatusName: formatOrderStatus(o.Status),
		Remark:     o.Remark,
		CreatedAt:  o.CreatedAt,
	}

	if o.User.ID != 0 {
		dto.User = &UserBrief{
			ID:       o.User.ID,
			Nickname: o.User.Nickname,
			Avatar:   o.User.Avatar,
			Phone:    o.User.Phone,
		}
	}

	return dto
}

// formatOrderStatus 格式化订单状态
func formatOrderStatus(s string) string {
	switch s {
	case model.OrderStatusPending:
		return "待处理"
	case model.OrderStatusAccepted:
		return "已接受"
	case model.OrderStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// getPageParams 获取分页参数
func getPageParams(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = 10

	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}

	if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	return
}

// handleBindError 处理绑定错误
func handleBindError(c *gin.Context, err error) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range errs {
			errors[e.Field()] = getErrandErrorMsg(e)
		}
		response.ValidationError(c, errors)
		return
	}
	response.BadRequest(c, "参数错误: "+err.Error())
}

// getErrandErrorMsg 获取验证错误消息
func getErrandErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "此字段为必填项"
	case "max":
		return "长度不能超过" + fe.Param()
	case "min":
		return "不能小于" + fe.Param()
	case "oneof":
		return "必须是以下值之一: " + fe.Param()
	default:
		return "参数验证失败"
	}
}

// 初始化验证器
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证规则
		_ = v.RegisterValidation("errand_type", func(fl validator.FieldLevel) bool {
			t := fl.Field().String()
			return service.ValidateErrandType(t)
		})
	}
}