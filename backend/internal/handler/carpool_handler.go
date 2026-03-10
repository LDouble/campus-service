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

// CarpoolHandler 拼车处理器
type CarpoolHandler struct {
	carpoolService service.CarpoolService
}

// NewCarpoolHandler 创建拼车处理器实例
func NewCarpoolHandler(carpoolService service.CarpoolService) *CarpoolHandler {
	return &CarpoolHandler{
		carpoolService: carpoolService,
	}
}

// CarpoolDTO 拼车行程数据传输对象
type CarpoolDTO struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"user_id"`
	Origin         string     `json:"origin"`
	Destination    string     `json:"destination"`
	DepartureTime  *time.Time `json:"departure_time"`
	Seats          int        `json:"seats"`
	AvailableSeats int        `json:"available_seats"`
	Price          float64    `json:"price"`
	VehicleType    string     `json:"vehicle_type"`
	VehicleTypeName string    `json:"vehicle_type_name"`
	VehicleNumber  string     `json:"vehicle_number"`
	Contact        string     `json:"contact"`
	Remark         string     `json:"remark"`
	Status         string     `json:"status"`
	StatusName     string     `json:"status_name"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	User           *UserBrief `json:"user,omitempty"`
}

// CarpoolDetailDTO 拼车行程详情DTO
type CarpoolDetailDTO struct {
	CarpoolDTO
	Bookings []BookingDTO `json:"bookings,omitempty"`
}

// BookingDTO 预约数据传输对象
type BookingDTO struct {
	ID         uint       `json:"id"`
	CarpoolID  uint       `json:"carpool_id"`
	UserID     uint       `json:"user_id"`
	Seats      int        `json:"seats"`
	Status     string     `json:"status"`
	StatusName string     `json:"status_name"`
	Remark     string     `json:"remark"`
	CreatedAt  time.Time  `json:"created_at"`
	User       *UserBrief `json:"user,omitempty"`
	Carpool    *CarpoolDTO `json:"carpool,omitempty"`
}

// PublishCarpoolRequest 发布拼车行程请求
type PublishCarpoolRequest struct {
	Origin        string     `json:"origin" binding:"required,max=100"`
	Destination   string     `json:"destination" binding:"required,max=100"`
	DepartureTime *time.Time `json:"departure_time" binding:"required"`
	Seats         int        `json:"seats" binding:"required,min=1,max=10"`
	Price         float64    `json:"price" binding:"min=0"`
	VehicleType   string     `json:"vehicle_type" binding:"omitempty,oneof=car taxi bus"`
	VehicleNumber string     `json:"vehicle_number" binding:"max=20"`
	Contact       string     `json:"contact" binding:"required,max=50"`
	Remark        string     `json:"remark" binding:"max=200"`
}

// UpdateCarpoolRequest 更新拼车行程请求
type UpdateCarpoolRequest struct {
	Origin        string     `json:"origin" binding:"omitempty,max=100"`
	Destination   string     `json:"destination" binding:"omitempty,max=100"`
	DepartureTime *time.Time `json:"departure_time"`
	Seats         int        `json:"seats" binding:"omitempty,min=1,max=10"`
	Price         float64    `json:"price" binding:"omitempty,min=0"`
	VehicleNumber string     `json:"vehicle_number" binding:"omitempty,max=20"`
	Contact       string     `json:"contact" binding:"omitempty,max=50"`
	Remark        string     `json:"remark" binding:"omitempty,max=200"`
}

// BookCarpoolRequest 预约拼车请求
type BookCarpoolRequest struct {
	Seats  int    `json:"seats" binding:"required,min=1"`
	Remark string `json:"remark" binding:"max=200"`
}

// CancelCarpoolRequest 取消行程请求
type CancelCarpoolRequest struct {
	Reason string `json:"reason" binding:"max=200"`
}

// CarpoolListResponse 拼车行程列表响应
type CarpoolListResponse struct {
	List     []CarpoolDTO `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// BookingListResponse 预约列表响应
type BookingListResponse struct {
	List     []BookingDTO `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// Publish 发布拼车行程
// @Summary 发布拼车行程
// @Description 发布一个新的拼车行程
// @Tags 拼车
// @Accept json
// @Produce json
// @Param request body PublishCarpoolRequest true "发布请求"
// @Success 200 {object} response.Response{data=CarpoolDTO}
// @Router /api/v1/carpools [post]
func (h *CarpoolHandler) Publish(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	var req PublishCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleCarpoolBindError(c, err)
		return
	}

	// 验证出发时间
	if req.DepartureTime != nil && req.DepartureTime.Before(time.Now()) {
		response.BadRequest(c, "出发时间不能早于当前时间")
		return
	}

	carpool, err := h.carpoolService.PublishCarpool(c.Request.Context(), userID, &service.PublishCarpoolRequest{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureTime: req.DepartureTime,
		Seats:         req.Seats,
		Price:         req.Price,
		VehicleType:   req.VehicleType,
		VehicleNumber: req.VehicleNumber,
		Contact:       req.Contact,
		Remark:        req.Remark,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Created(c, toCarpoolDTO(*carpool))
}

// List 获取拼车行程列表
// @Summary 获取拼车行程列表
// @Description 分页获取拼车行程列表
// @Tags 拼车
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param origin query string false "出发地"
// @Param destination query string false "目的地"
// @Param status query string false "行程状态" Enums(pending, departed, cancelled)
// @Param date query string false "出发日期(2006-01-02)"
// @Success 200 {object} response.Response{data=CarpoolListResponse}
// @Router /api/v1/carpools [get]
func (h *CarpoolHandler) List(c *gin.Context) {
	var query model.CarpoolListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		handleCarpoolBindError(c, err)
		return
	}

	result, err := h.carpoolService.GetCarpoolList(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]CarpoolDTO, 0, len(result.List))
	for _, cp := range result.List {
		list = append(list, toCarpoolDTO(cp))
	}

	response.Success(c, &CarpoolListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// Detail 获取拼车行程详情
// @Summary 获取拼车行程详情
// @Description 根据ID获取拼车行程详情
// @Tags 拼车
// @Produce json
// @Param id path int true "行程ID"
// @Success 200 {object} response.Response{data=CarpoolDetailDTO}
// @Router /api/v1/carpools/{id} [get]
func (h *CarpoolHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的行程ID")
		return
	}

	result, err := h.carpoolService.GetCarpoolDetail(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "拼车行程不存在" {
			response.NotFound(c, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	dto := toCarpoolDTO(result.Carpool)
	detailDTO := CarpoolDetailDTO{
		CarpoolDTO: dto,
	}

	if len(result.Bookings) > 0 {
		bookings := make([]BookingDTO, 0, len(result.Bookings))
		for _, b := range result.Bookings {
			bookings = append(bookings, toBookingDTO(b))
		}
		detailDTO.Bookings = bookings
	}

	response.Success(c, detailDTO)
}

// Update 更新拼车行程
// @Summary 更新拼车行程
// @Description 更新拼车行程信息
// @Tags 拼车
// @Accept json
// @Produce json
// @Param id path int true "行程ID"
// @Param request body UpdateCarpoolRequest true "更新请求"
// @Success 200 {object} response.Response{data=CarpoolDTO}
// @Router /api/v1/carpools/{id} [put]
func (h *CarpoolHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的行程ID")
		return
	}

	var req UpdateCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleCarpoolBindError(c, err)
		return
	}

	carpool, err := h.carpoolService.UpdateCarpool(c.Request.Context(), userID, uint(id), &service.UpdateCarpoolRequest{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureTime: req.DepartureTime,
		Seats:         req.Seats,
		Price:         req.Price,
		VehicleNumber: req.VehicleNumber,
		Contact:       req.Contact,
		Remark:        req.Remark,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toCarpoolDTO(*carpool))
}

// Delete 删除拼车行程
// @Summary 删除拼车行程
// @Description 删除拼车行程（软删除）
// @Tags 拼车
// @Param id path int true "行程ID"
// @Success 200 {object} response.Response
// @Router /api/v1/carpools/{id} [delete]
func (h *CarpoolHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的行程ID")
		return
	}

	if err := h.carpoolService.DeleteCarpool(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Book 预约拼车
// @Summary 预约拼车
// @Description 预约一个拼车行程
// @Tags 拼车
// @Accept json
// @Produce json
// @Param id path int true "行程ID"
// @Param request body BookCarpoolRequest true "预约请求"
// @Success 200 {object} response.Response{data=BookingDTO}
// @Router /api/v1/carpools/{id}/book [post]
func (h *CarpoolHandler) Book(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的行程ID")
		return
	}

	var req BookCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Seats = 1 // 默认预约1个座位
	}

	booking, err := h.carpoolService.BookCarpool(c.Request.Context(), uint(id), userID, req.Seats, req.Remark)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toBookingDTO(*booking))
}

// CancelBooking 取消预约
// @Summary 取消预约
// @Description 取消拼车预约
// @Tags 拼车
// @Param booking_id path int true "预约ID"
// @Success 200 {object} response.Response
// @Router /api/v1/carpool-bookings/{booking_id}/cancel [post]
func (h *CarpoolHandler) CancelBooking(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	bookingID, err := strconv.ParseUint(c.Param("booking_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的预约ID")
		return
	}

	if err := h.carpoolService.CancelBooking(c.Request.Context(), uint(bookingID), userID); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Depart 确认出发
// @Summary 确认出发
// @Description 确认拼车行程出发
// @Tags 拼车
// @Param id path int true "行程ID"
// @Success 200 {object} response.Response
// @Router /api/v1/carpools/{id}/depart [post]
func (h *CarpoolHandler) Depart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的行程ID")
		return
	}

	if err := h.carpoolService.DepartCarpool(c.Request.Context(), uint(id), userID); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Cancel 取消行程
// @Summary 取消拼车行程
// @Description 取消拼车行程
// @Tags 拼车
// @Accept json
// @Param id path int true "行程ID"
// @Param request body CancelCarpoolRequest true "取消请求"
// @Success 200 {object} response.Response
// @Router /api/v1/carpools/{id}/cancel [post]
func (h *CarpoolHandler) Cancel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的行程ID")
		return
	}

	var req CancelCarpoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = ""
	}

	if err := h.carpoolService.CancelCarpool(c.Request.Context(), uint(id), userID, req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// MyPublished 获取我发布的行程
// @Summary 获取我发布的行程
// @Description 获取当前用户发布的拼车行程列表
// @Tags 拼车
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=CarpoolListResponse}
// @Router /api/v1/carpools/my/published [get]
func (h *CarpoolHandler) MyPublished(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.carpoolService.GetMyCarpools(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]CarpoolDTO, 0, len(result.List))
	for _, cp := range result.List {
		list = append(list, toCarpoolDTO(cp))
	}

	response.Success(c, &CarpoolListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// MyBookings 获取我的预约
// @Summary 获取我的预约
// @Description 获取当前用户的拼车预约列表
// @Tags 拼车
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=BookingListResponse}
// @Router /api/v1/carpools/my/bookings [get]
func (h *CarpoolHandler) MyBookings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.carpoolService.GetMyBookings(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]BookingDTO, 0, len(result.List))
	for _, b := range result.List {
		list = append(list, toBookingDTO(b))
	}

	response.Success(c, &BookingListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// RegisterRoutes 注册路由
func (h *CarpoolHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	carpools := r.Group("/carpools")
	{
		// 公开路由
		carpools.GET("", h.List)
		carpools.GET("/:id", h.Detail)

		// 需要认证的路由
		authGroup := carpools.Group("")
		authGroup.Use(authMiddleware)
		{
			authGroup.POST("", h.Publish)
			authGroup.PUT("/:id", h.Update)
			authGroup.DELETE("/:id", h.Delete)
			authGroup.POST("/:id/book", h.Book)
			authGroup.POST("/:id/depart", h.Depart)
			authGroup.POST("/:id/cancel", h.Cancel)
			authGroup.GET("/my/published", h.MyPublished)
			authGroup.GET("/my/bookings", h.MyBookings)
		}
	}

	// 预约相关路由
	bookings := r.Group("/carpool-bookings")
	bookings.Use(authMiddleware)
	{
		bookings.POST("/:booking_id/cancel", h.CancelBooking)
	}
}

// 辅助函数

// toCarpoolDTO 转换拼车行程到DTO
func toCarpoolDTO(c model.Carpool) CarpoolDTO {
	dto := CarpoolDTO{
		ID:              c.ID,
		UserID:          c.UserID,
		Origin:          c.Origin,
		Destination:     c.Destination,
		DepartureTime:   c.DepartureTime,
		Seats:           c.Seats,
		AvailableSeats:  c.AvailableSeats,
		Price:           c.Price,
		VehicleType:     c.VehicleType,
		VehicleTypeName: service.FormatVehicleType(c.VehicleType),
		VehicleNumber:   c.VehicleNumber,
		Contact:         c.Contact,
		Remark:          c.Remark,
		Status:          c.Status,
		StatusName:      service.FormatCarpoolStatus(c.Status),
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}

	if c.User.ID != 0 {
		dto.User = &UserBrief{
			ID:       c.User.ID,
			Nickname: c.User.Nickname,
			Avatar:   c.User.Avatar,
		}
	}

	return dto
}

// toBookingDTO 转换预约到DTO
func toBookingDTO(b model.CarpoolBooking) BookingDTO {
	dto := BookingDTO{
		ID:         b.ID,
		CarpoolID:  b.CarpoolID,
		UserID:     b.UserID,
		Seats:      b.Seats,
		Status:     b.Status,
		StatusName: service.FormatBookingStatus(b.Status),
		Remark:     b.Remark,
		CreatedAt:  b.CreatedAt,
	}

	if b.User.ID != 0 {
		dto.User = &UserBrief{
			ID:       b.User.ID,
			Nickname: b.User.Nickname,
			Avatar:   b.User.Avatar,
			Phone:    b.User.Phone,
		}
	}

	if b.Carpool.ID != 0 {
		carpoolDTO := toCarpoolDTO(b.Carpool)
		dto.Carpool = &carpoolDTO
	}

	return dto
}

// handleCarpoolBindError 处理绑定错误
func handleCarpoolBindError(c *gin.Context, err error) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range errs {
			errors[e.Field()] = getCarpoolErrorMsg(e)
		}
		response.ValidationError(c, errors)
		return
	}
	response.BadRequest(c, "参数错误: "+err.Error())
}

// getCarpoolErrorMsg 获取验证错误消息
func getCarpoolErrorMsg(fe validator.FieldError) string {
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
		_ = v.RegisterValidation("vehicle_type", func(fl validator.FieldLevel) bool {
			t := fl.Field().String()
			return service.ValidateVehicleType(t)
		})
	}
}