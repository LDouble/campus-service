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

// MarketHandler 二手交易处理器
type MarketHandler struct {
	marketService service.MarketService
}

// NewMarketHandler 创建二手交易处理器实例
func NewMarketHandler(marketService service.MarketService) *MarketHandler {
	return &MarketHandler{
		marketService: marketService,
	}
}

// MarketItemDTO 商品数据传输对象
type MarketItemDTO struct {
	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Category      string     `json:"category"`
	CategoryName  string     `json:"category_name"`
	Price         float64    `json:"price"`
	OriginalPrice float64    `json:"original_price"`
	Condition     string     `json:"condition"`
	ConditionName string     `json:"condition_name"`
	Images        []string   `json:"images"`
	Location      string     `json:"location"`
	Contact       string     `json:"contact"`
	Status        string     `json:"status"`
	StatusName    string     `json:"status_name"`
	ViewCount     int        `json:"view_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	User          *UserBrief `json:"user,omitempty"`
}

// MarketItemDetailDTO 商品详情DTO
type MarketItemDetailDTO struct {
	MarketItemDTO
	IsFavorited bool `json:"is_favorited"`
}

// FavoriteDTO 收藏数据传输对象
type FavoriteDTO struct {
	ID        uint          `json:"id"`
	ItemID    uint          `json:"item_id"`
	UserID    uint          `json:"user_id"`
	CreatedAt time.Time     `json:"created_at"`
	Item      *MarketItemDTO `json:"item,omitempty"`
}

// PublishItemRequest 发布商品请求
type PublishItemRequest struct {
	Title         string   `json:"title" binding:"required,max=100"`
	Description   string   `json:"description" binding:"max=1000"`
	Category      string   `json:"category" binding:"required,oneof=electronics books clothes sports daily others"`
	Price         float64  `json:"price" binding:"required,min=0"`
	OriginalPrice float64  `json:"original_price" binding:"omitempty,min=0"`
	Condition     string   `json:"condition" binding:"omitempty,oneof=new like_new good fair"`
	Images        []string `json:"images" binding:"max=9"`
	Location      string   `json:"location" binding:"max=200"`
	Contact       string   `json:"contact" binding:"required,max=50"`
}

// UpdateItemRequest 更新商品请求
type UpdateItemRequest struct {
	Title         string   `json:"title" binding:"omitempty,max=100"`
	Description   string   `json:"description" binding:"omitempty,max=1000"`
	Category      string   `json:"category" binding:"omitempty,oneof=electronics books clothes sports daily others"`
	Price         float64  `json:"price" binding:"omitempty,min=0"`
	OriginalPrice float64  `json:"original_price" binding:"omitempty,min=0"`
	Condition     string   `json:"condition" binding:"omitempty,oneof=new like_new good fair"`
	Images        []string `json:"images" binding:"omitempty,max=9"`
	Location      string   `json:"location" binding:"omitempty,max=200"`
	Contact       string   `json:"contact" binding:"omitempty,max=50"`
}

// ItemListResponse 商品列表响应
type ItemListResponse struct {
	List     []MarketItemDTO `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// FavoriteListResponse 收藏列表响应
type FavoriteListResponse struct {
	List     []FavoriteDTO `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// Publish 发布商品
// @Summary 发布商品
// @Description 发布一个二手商品
// @Tags 二手交易
// @Accept json
// @Produce json
// @Param request body PublishItemRequest true "发布请求"
// @Success 200 {object} response.Response{data=MarketItemDTO}
// @Router /api/v1/market/items [post]
func (h *MarketHandler) Publish(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	var req PublishItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleMarketBindError(c, err)
		return
	}

	item, err := h.marketService.PublishItem(c.Request.Context(), userID, &service.PublishItemRequest{
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Condition:     req.Condition,
		Images:        req.Images,
		Location:      req.Location,
		Contact:       req.Contact,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Created(c, toItemDTO(*item))
}

// List 获取商品列表
// @Summary 获取商品列表
// @Description 分页获取二手商品列表
// @Tags 二手交易
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param category query string false "分类" Enums(electronics, books, clothes, sports, daily, others)
// @Param keyword query string false "关键词"
// @Param min_price query number false "最低价格"
// @Param max_price query number false "最高价格"
// @Param sort_by query string false "排序" Enums(price, price_desc, time, time_desc)
// @Success 200 {object} response.Response{data=ItemListResponse}
// @Router /api/v1/market/items [get]
func (h *MarketHandler) List(c *gin.Context) {
	var query model.MarketItemListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		handleMarketBindError(c, err)
		return
	}

	result, err := h.marketService.GetItemList(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]MarketItemDTO, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, toItemDTO(item))
	}

	response.Success(c, &ItemListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// Detail 获取商品详情
// @Summary 获取商品详情
// @Description 根据ID获取商品详情
// @Tags 二手交易
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} response.Response{data=MarketItemDetailDTO}
// @Router /api/v1/market/items/{id} [get]
func (h *MarketHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	userID := middleware.GetUserID(c)

	result, err := h.marketService.GetItemDetail(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err.Error() == "商品不存在" {
			response.NotFound(c, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	dto := toItemDTO(result.MarketItem)
	detailDTO := MarketItemDetailDTO{
		MarketItemDTO: dto,
		IsFavorited:   result.IsFavorited,
	}

	response.Success(c, detailDTO)
}

// Update 更新商品
// @Summary 更新商品
// @Description 更新商品信息
// @Tags 二手交易
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Param request body UpdateItemRequest true "更新请求"
// @Success 200 {object} response.Response{data=MarketItemDTO}
// @Router /api/v1/market/items/{id} [put]
func (h *MarketHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleMarketBindError(c, err)
		return
	}

	item, err := h.marketService.UpdateItem(c.Request.Context(), userID, uint(id), &service.UpdateItemRequest{
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Condition:     req.Condition,
		Images:        req.Images,
		Location:      req.Location,
		Contact:       req.Contact,
	})

	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, toItemDTO(*item))
}

// Delete 删除商品
// @Summary 删除商品
// @Description 删除商品
// @Tags 二手交易
// @Param id path int true "商品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/market/items/{id} [delete]
func (h *MarketHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.marketService.DeleteItem(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// MarkSold 标记已售
// @Summary 标记商品已售
// @Description 标记商品为已售出
// @Tags 二手交易
// @Param id path int true "商品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/market/items/{id}/sold [post]
func (h *MarketHandler) MarkSold(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.marketService.MarkAsSold(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// OffShelf 下架商品
// @Summary 下架商品
// @Description 下架商品
// @Tags 二手交易
// @Param id path int true "商品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/market/items/{id}/off-shelf [post]
func (h *MarketHandler) OffShelf(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.marketService.OffShelf(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// OnShelf 重新上架
// @Summary 重新上架商品
// @Description 重新上架已下架的商品
// @Tags 二手交易
// @Param id path int true "商品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/market/items/{id}/on-shelf [post]
func (h *MarketHandler) OnShelf(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.marketService.OnShelf(c.Request.Context(), userID, uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// AddFavorite 添加收藏
// @Summary 收藏商品
// @Description 收藏一个商品
// @Tags 二手交易
// @Param id path int true "商品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/market/items/{id}/favorite [post]
func (h *MarketHandler) AddFavorite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.marketService.AddFavorite(c.Request.Context(), uint(id), userID); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// RemoveFavorite 取消收藏
// @Summary 取消收藏
// @Description 取消收藏商品
// @Tags 二手交易
// @Param id path int true "商品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/market/items/{id}/favorite [delete]
func (h *MarketHandler) RemoveFavorite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.marketService.RemoveFavorite(c.Request.Context(), uint(id), userID); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// MyItems 获取我的商品
// @Summary 获取我发布的商品
// @Description 获取当前用户发布的商品列表
// @Tags 二手交易
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=ItemListResponse}
// @Router /api/v1/market/my/items [get]
func (h *MarketHandler) MyItems(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.marketService.GetMyItems(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]MarketItemDTO, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, toItemDTO(item))
	}

	response.Success(c, &ItemListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// MyFavorites 获取我的收藏
// @Summary 获取我的收藏
// @Description 获取当前用户收藏的商品列表
// @Tags 二手交易
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=FavoriteListResponse}
// @Router /api/v1/market/my/favorites [get]
func (h *MarketHandler) MyFavorites(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, response.CodeUnauthorized, "未登录")
		return
	}

	page, pageSize := getPageParams(c)

	result, err := h.marketService.GetMyFavorites(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	list := make([]FavoriteDTO, 0, len(result.List))
	for _, fav := range result.List {
		dto := FavoriteDTO{
			ID:        fav.ID,
			ItemID:    fav.ItemID,
			UserID:    fav.UserID,
			CreatedAt: fav.CreatedAt,
		}
		if fav.Item.ID != 0 {
			itemDTO := toItemDTO(fav.Item)
			dto.Item = &itemDTO
		}
		list = append(list, dto)
	}

	response.Success(c, &FavoriteListResponse{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// RegisterRoutes 注册路由
func (h *MarketHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	items := r.Group("/market/items")
	{
		// 公开路由
		items.GET("", h.List)
		items.GET("/:id", h.Detail)

		// 需要认证的路由
		authGroup := items.Group("")
		authGroup.Use(authMiddleware)
		{
			authGroup.POST("", h.Publish)
			authGroup.PUT("/:id", h.Update)
			authGroup.DELETE("/:id", h.Delete)
			authGroup.POST("/:id/sold", h.MarkSold)
			authGroup.POST("/:id/off-shelf", h.OffShelf)
			authGroup.POST("/:id/on-shelf", h.OnShelf)
			authGroup.POST("/:id/favorite", h.AddFavorite)
			authGroup.DELETE("/:id/favorite", h.RemoveFavorite)
		}
	}

	// 我的商品和收藏
	my := r.Group("/market/my")
	my.Use(authMiddleware)
	{
		my.GET("/items", h.MyItems)
		my.GET("/favorites", h.MyFavorites)
	}
}

// 辅助函数

// toItemDTO 转换商品到DTO
func toItemDTO(m model.MarketItem) MarketItemDTO {
	dto := MarketItemDTO{
		ID:            m.ID,
		UserID:        m.UserID,
		Title:         m.Title,
		Description:   m.Description,
		Category:      m.Category,
		CategoryName:  m.GetCategoryName(),
		Price:         m.Price,
		OriginalPrice: m.OriginalPrice,
		Condition:     m.Condition,
		ConditionName: m.GetConditionName(),
		Location:      m.Location,
		Contact:       m.Contact,
		Status:        m.Status,
		StatusName:    m.GetStatusName(),
		ViewCount:     m.ViewCount,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
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

// handleMarketBindError 处理绑定错误
func handleMarketBindError(c *gin.Context, err error) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range errs {
			errors[e.Field()] = getMarketErrorMsg(e)
		}
		response.ValidationError(c, errors)
		return
	}
	response.BadRequest(c, "参数错误: "+err.Error())
}

// getMarketErrorMsg 获取验证错误消息
func getMarketErrorMsg(fe validator.FieldError) string {
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
		_ = v.RegisterValidation("category", func(fl validator.FieldLevel) bool {
			c := fl.Field().String()
			return service.ValidateCategory(c)
		})
		_ = v.RegisterValidation("condition", func(fl validator.FieldLevel) bool {
			c := fl.Field().String()
			return service.ValidateCondition(c)
		})
	}
}