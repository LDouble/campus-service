package handler

import (
	"encoding/json"
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

// LostFoundHandler 失物招领处理器
type LostFoundHandler struct {
	lfService service.LostFoundService
}

// NewLostFoundHandler 创建失物招领处理器实例
func NewLostFoundHandler(lfService service.LostFoundService) *LostFoundHandler {
	return &LostFoundHandler{
		lfService: lfService,
	}
}

// LostFoundDTO 失物招领数据传输对象
type LostFoundDTO struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id"`
	Type         string     `json:"type"`
	TypeName     string     `json:"type_name"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Category     string     `json:"category"`
	CategoryName string     `json:"category_name"`
	Images       []string   `json:"images"`
	Location     string     `json:"location"`
	LostTime     *time.Time `json:"lost_time"`
	Contact      string     `json:"contact"`
	Status       string     `json:"status"`
	StatusName   string     `json:"status_name"`
	ViewCount    int        `json:"view_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	User         *UserBrief `json:"user,omitempty"`
}

// LostFoundDetailDTO 失物招领详情DTO
type LostFoundDetailDTO struct {
	LostFoundDTO
	HasClaimed bool `json:"has_claimed"`
}

// ClaimDTO 认领数据传输对象
type ClaimDTO struct {
	ID         uint          `json:"id"`
	LostFoundID uint         `json:"lost_found_id"`
	UserID     uint          `json:"user_id"`
	Message    string        `json:"message"`
	Proof      []string      `json:"proof"`
	Status     string        `json:"status"`
	StatusName string        `json:"status_name"`
	CreatedAt  time.Time     `json:"created_at"`
	User       *UserBrief    `json:"user,omitempty"`
	LostFound  *LostFoundDTO `json:"lost_found,omitempty"`
}

// PublishLostFoundRequest 发布失物招领请求
type PublishLostFoundRequest struct {
	Type        string     `json:"type" binding:"required,oneof=lost found"`
	Title       string     `json:"title" binding:"required,max=100"`
	Description string     `json:"description" binding:"max=1000"`
	Category    string     `json:"category" binding:"required,oneof=electronics cards books clothes bags others"`
	Images      []string   `json:"images" binding:"max=9"`
	Location    string     `json:"location" binding:"max=200"`
	LostTime    *time.Time `json:"lost_time"`
	Contact     string     `json:"contact" binding:"required,max=50"`
}

// UpdateLostFoundRequest 更新失物招领请求
type UpdateLostFoundRequest struct {
	Title       string     `json:"title" binding:"omitempty,max=100"`
	Description string     `json:"description" binding:"omitempty,max=1000"`
	Category    string     `json:"category" binding:"omitempty,oneof=electronics cards books clothes bags others"`
	Images      []string   `json:"images" binding:"omitempty,max=9"`
	Location    string     `json:"location" binding:"omitempty,max=200"`
	LostTime    *time.Time `json:"lost_time"`
	Contact     string     `json:"contact" binding:"omitempty,max=50"`
}

// SubmitClaimRequest 提交认领请求
type SubmitClaimRequest struct {
	Message string   `json:"message" binding:"required,max=500"`
	Proof   []string `json:"proof" binding:"max=9"`
}

// LostFoundListResponse 失物招领列表响应
type LostFoundListResponse struct {
	List     []LostFoundDTO `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ClaimListResponse 认领列表响应
type ClaimListResponse struct {
	List     []ClaimDTO `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// Publish 发布失物招领
// @Summary 发布失物招领
// @Description 发布寻物启事或失物招领
// @Tags 失物招领
// @Accept json
// @Produce json
// @Param request body PublishLostFoundRequest true "发布请求"
// @Success 200 {object} response.Response{data=LostFoundDTO}
// @Router /api/v1/lost-found [post]
func (h *LostFoundHandler) Publish(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	var req PublishLostFoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleLFBindError(c, err)
		return
	}

	lf, err := h.lfService.Publish(c.Request.Context(), userID, &service.PublishLostFoundRequest{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Images:      req.Images,
		Location:    req.Location,
		LostTime:    req.LostTime,
		Contact:     req.Contact,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Created(c, toLostFoundDTO(*lf))
}

// List 获取失物招领列表
// @Summary 获取失物招领列表
// @Description 分页获取失物招领列表
// @Tags 失物招领
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "类型" Enums(lost, found)
// @Param category query string false "分类" Enums(electronics, cards, books, clothes, bags, others)
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=LostFoundListResponse}
// @Router /api/v1/lost-found [get]
func (h *LostFoundHandler) List(c *gin.Context) {
	var query model.LostFoundListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		handleLFBindError(c, err)
		return
	}

	result, err := h.lfService.GetList(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]LostFoundDTO, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, toLostFoundDTO(item))
	}

	response.Success(c, &LostFoundListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// Detail 获取失物招领详情
// @Summary 获取失物招领详情
// @Description 根据ID获取失物招领详情
// @Tags 失物招领
// @Produce json
// @Param id path int true "记录ID"
// @Success 200 {object} response.Response{data=LostFoundDetailDTO}
// @Router /api/v1/lost-found/{id} [get]
func (h *LostFoundHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	userID := middleware.GetUserID(c)

	result, err := h.lfService.GetDetail(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err.Error() == "记录不存在" {
			response.NotFound(c, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	dto := toLostFoundDTO(result.LostFound)
	detailDTO := LostFoundDetailDTO{
		LostFoundDTO: dto,
		HasClaimed:   result.HasClaimed,
	}

	response.Success(c, detailDTO)
}

// Update 更新失物招领
// @Summary 更新失物招领
// @Description 更新失物招领信息
// @Tags 失物招领
// @Accept json
// @Produce json
// @Param id path int true "记录ID"
// @Param request body UpdateLostFoundRequest true "更新请求"
// @Success 200 {object} response.Response{data=LostFoundDTO}
// @Router /api/v1/lost-found/{id} [put]
func (h *LostFoundHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	var req UpdateLostFoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleLFBindError(c, err)
		return
	}

	lf, err := h.lfService.Update(c.Request.Context(), userID, uint(id), &service.UpdateLostFoundRequest{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Images:      req.Images,
		Location:    req.Location,
		LostTime:    req.LostTime,
		Contact:     req.Contact,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toLostFoundDTO(*lf))
}

// Delete 删除失物招领
// @Summary 删除失物招领
// @Description 删除失物招领记录
// @Tags 失物招领
// @Param id path int true "记录ID"
// @Success 200 {object} response.Response
// @Router /api/v1/lost-found/{id} [delete]
func (h *LostFoundHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	if err := h.lfService.Delete(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Close 关闭失物招领
// @Summary 关闭失物招领
// @Description 关闭失物招领记录
// @Tags 失物招领
// @Param id path int true "记录ID"
// @Success 200 {object} response.Response
// @Router /api/v1/lost-found/{id}/close [post]
func (h *LostFoundHandler) Close(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	if err := h.lfService.Close(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Resolve 标记已解决
// @Summary 标记已解决
// @Description 标记失物招领为已解决
// @Tags 失物招领
// @Param id path int true "记录ID"
// @Success 200 {object} response.Response
// @Router /api/v1/lost-found/{id}/resolve [post]
func (h *LostFoundHandler) Resolve(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	if err := h.lfService.Resolve(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// SubmitClaim 提交认领
// @Summary 提交认领申请
// @Description 提交失物招领认领申请
// @Tags 失物招领
// @Accept json
// @Param id path int true "记录ID"
// @Param request body SubmitClaimRequest true "认领请求"
// @Success 200 {object} response.Response{data=ClaimDTO}
// @Router /api/v1/lost-found/{id}/claim [post]
func (h *LostFoundHandler) SubmitClaim(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	var req SubmitClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleLFBindError(c, err)
		return
	}

	claim, err := h.lfService.SubmitClaim(c.Request.Context(), uint(id), userID, &service.SubmitClaimRequest{
		Message: req.Message,
		Proof:   req.Proof,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Created(c, toClaimDTO(*claim))
}

// GetClaims 获取认领列表
// @Summary 获取认领列表
// @Description 获取失物招领的认领申请列表
// @Tags 失物招领
// @Param id path int true "记录ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=ClaimListResponse}
// @Router /api/v1/lost-found/{id}/claims [get]
func (h *LostFoundHandler) GetClaims(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.lfService.GetClaims(c.Request.Context(), uint(id), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]ClaimDTO, 0, len(result.List))
	for _, claim := range result.List {
		list = append(list, toClaimDTO(claim))
	}

	response.Success(c, &ClaimListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// ApproveClaim 通过认领
// @Summary 通过认领申请
// @Description 通过失物招领认领申请
// @Tags 失物招领
// @Param claim_id path int true "认领ID"
// @Success 200 {object} response.Response
// @Router /api/v1/lost-found/claims/{claim_id}/approve [post]
func (h *LostFoundHandler) ApproveClaim(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	claimID, err := strconv.ParseUint(c.Param("claim_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的认领ID")
		return
	}

	if err := h.lfService.ApproveClaim(c.Request.Context(), uint(claimID), userID); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// RejectClaim 拒绝认领
// @Summary 拒绝认领申请
// @Description 拒绝失物招领认领申请
// @Tags 失物招领
// @Param claim_id path int true "认领ID"
// @Success 200 {object} response.Response
// @Router /api/v1/lost-found/claims/{claim_id}/reject [post]
func (h *LostFoundHandler) RejectClaim(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	claimID, err := strconv.ParseUint(c.Param("claim_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的认领ID")
		return
	}

	if err := h.lfService.RejectClaim(c.Request.Context(), uint(claimID), userID); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// MyPublish 获取我的发布
// @Summary 获取我发布的失物招领
// @Description 获取当前用户发布的失物招领列表
// @Tags 失物招领
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=LostFoundListResponse}
// @Router /api/v1/lost-found/my/published [get]
func (h *LostFoundHandler) MyPublish(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.lfService.GetMyPublish(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]LostFoundDTO, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, toLostFoundDTO(item))
	}

	response.Success(c, &LostFoundListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// MyClaims 获取我的认领
// @Summary 获取我的认领申请
// @Description 获取当前用户提交的认领申请列表
// @Tags 失物招领
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=ClaimListResponse}
// @Router /api/v1/lost-found/my/claims [get]
func (h *LostFoundHandler) MyClaims(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.lfService.GetMyClaims(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]ClaimDTO, 0, len(result.List))
	for _, claim := range result.List {
		list = append(list, toClaimDTO(claim))
	}

	response.Success(c, &ClaimListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// RegisterRoutes 注册路由
func (h *LostFoundHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	lf := r.Group("/lost-found")
	{
		// 公开路由
		lf.GET("", h.List)
		lf.GET("/:id", h.Detail)
		lf.GET("/:id/claims", h.GetClaims)

		// 需要认证的路由
		authGroup := lf.Group("")
		authGroup.Use(authMiddleware)
		{
			authGroup.POST("", h.Publish)
			authGroup.PUT("/:id", h.Update)
			authGroup.DELETE("/:id", h.Delete)
			authGroup.POST("/:id/close", h.Close)
			authGroup.POST("/:id/resolve", h.Resolve)
			authGroup.POST("/:id/claim", h.SubmitClaim)
		}

		// 认领审核路由
		claims := r.Group("/lost-found/claims")
		claims.Use(authMiddleware)
		{
			claims.POST("/:claim_id/approve", h.ApproveClaim)
			claims.POST("/:claim_id/reject", h.RejectClaim)
		}

		// 我的发布和认领
		my := r.Group("/lost-found/my")
		my.Use(authMiddleware)
		{
			my.GET("/published", h.MyPublish)
			my.GET("/claims", h.MyClaims)
		}
	}
}

// 辅助函数

// toLostFoundDTO 转换失物招领到DTO
func toLostFoundDTO(m model.LostFound) LostFoundDTO {
	dto := LostFoundDTO{
		ID:           m.ID,
		UserID:       m.UserID,
		Type:         m.Type,
		TypeName:     m.GetTypeName(),
		Title:        m.Title,
		Description:  m.Description,
		Category:     m.Category,
		CategoryName: m.GetCategoryName(),
		Location:     m.Location,
		LostTime:     m.LostTime,
		Contact:      m.Contact,
		Status:       m.Status,
		StatusName:   m.GetStatusName(),
		ViewCount:    m.ViewCount,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}

	// 解析图片
	if m.Images != "" {
		var images []string
		if err := json.Unmarshal([]byte(m.Images), &images); err == nil {
			dto.Images = images
		}
	}

	if m.User.ID != 0 {
		dto.User = &UserBrief{
			ID:       m.User.ID,
			Nickname: m.User.Nickname,
			Avatar:   m.User.Avatar,
		}
	}

	return dto
}

// toClaimDTO 转换认领到DTO
func toClaimDTO(c model.LostFoundClaim) ClaimDTO {
	dto := ClaimDTO{
		ID:          c.ID,
		LostFoundID: c.LostFoundID,
		UserID:      c.UserID,
		Message:     c.Message,
		Status:      c.Status,
		StatusName:  c.GetClaimStatusName(),
		CreatedAt:   c.CreatedAt,
	}

	// 解析证明材料
	if c.Proof != "" {
		var proof []string
		if err := json.Unmarshal([]byte(c.Proof), &proof); err == nil {
			dto.Proof = proof
		}
	}

	if c.User.ID != 0 {
		dto.User = &UserBrief{
			ID:       c.User.ID,
			Nickname: c.User.Nickname,
			Avatar:   c.User.Avatar,
		}
	}

	if c.LostFound.ID != 0 {
		lfDTO := toLostFoundDTO(c.LostFound)
		dto.LostFound = &lfDTO
	}

	return dto
}

// handleLFBindError 处理绑定错误
func handleLFBindError(c *gin.Context, err error) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range errs {
			errors[e.Field()] = getLFErrorMsg(e)
		}
		response.ValidationError(c, errors)
		return
	}
	response.BadRequest(c, "参数错误: "+err.Error())
}

// getLFErrorMsg 获取验证错误消息
func getLFErrorMsg(fe validator.FieldError) string {
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
		_ = v.RegisterValidation("lf_type", func(fl validator.FieldLevel) bool {
			t := fl.Field().String()
			return service.ValidateLostFoundType(t)
		})
		_ = v.RegisterValidation("lf_category", func(fl validator.FieldLevel) bool {
			c := fl.Field().String()
			return service.ValidateLostFoundCategory(c)
		})
	}
}