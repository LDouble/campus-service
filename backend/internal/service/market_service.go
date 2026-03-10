package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"campus-service/internal/model"
	"campus-service/internal/pkg/logger"
	"campus-service/internal/repository"
)

// MarketService 二手交易服务接口
type MarketService interface {
	// 商品管理
	PublishItem(ctx context.Context, userID uint, req *PublishItemRequest) (*model.MarketItem, error)
	GetItemList(ctx context.Context, query *model.MarketItemListQuery) (*ItemListResponse, error)
	GetItemDetail(ctx context.Context, id, userID uint) (*ItemDetailResponse, error)
	UpdateItem(ctx context.Context, userID, itemID uint, req *UpdateItemRequest) (*model.MarketItem, error)
	DeleteItem(ctx context.Context, userID, itemID uint) error
	MarkAsSold(ctx context.Context, userID, itemID uint) error
	OffShelf(ctx context.Context, userID, itemID uint) error
	OnShelf(ctx context.Context, userID, itemID uint) error

	// 收藏管理
	AddFavorite(ctx context.Context, itemID, userID uint) error
	RemoveFavorite(ctx context.Context, itemID, userID uint) error
	GetMyFavorites(ctx context.Context, userID uint, page, pageSize int) (*FavoriteListResponse, error)

	// 我的商品
	GetMyItems(ctx context.Context, userID uint, page, pageSize int) (*ItemListResponse, error)
}

// PublishItemRequest 发布商品请求
type PublishItemRequest struct {
	Title         string   `json:"title" binding:"required,max=100"`
	Description   string   `json:"description" binding:"max=1000"`
	Category      string   `json:"category" binding:"required,oneof=electronics books clothes sports daily others"`
	Price         float64  `json:"price" binding:"required,min=0"`
	OriginalPrice  float64 `json:"original_price" binding:"omitempty,min=0"`
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
	OriginalPrice  float64 `json:"original_price" binding:"omitempty,min=0"`
	Condition     string   `json:"condition" binding:"omitempty,oneof=new like_new good fair"`
	Images        []string `json:"images" binding:"omitempty,max=9"`
	Location      string   `json:"location" binding:"omitempty,max=200"`
	Contact       string   `json:"contact" binding:"omitempty,max=50"`
}

// ItemListResponse 商品列表响应
type ItemListResponse struct {
	List     []model.MarketItem `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// ItemDetailResponse 商品详情响应
type ItemDetailResponse struct {
	model.MarketItem
	IsFavorited bool `json:"is_favorited"`
}

// FavoriteListResponse 收藏列表响应
type FavoriteListResponse struct {
	List     []model.MarketFavorite `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// marketService 二手交易服务实现
type marketService struct {
	marketRepo repository.MarketRepository
	userRepo   repository.UserRepository
}

// NewMarketService 创建二手交易服务实例
func NewMarketService(marketRepo repository.MarketRepository, userRepo repository.UserRepository) MarketService {
	return &marketService{
		marketRepo: marketRepo,
		userRepo:   userRepo,
	}
}

// PublishItem 发布商品
func (s *marketService) PublishItem(ctx context.Context, userID uint, req *PublishItemRequest) (*model.MarketItem, error) {
	// 序列化图片
	var imagesJSON string
	if len(req.Images) > 0 {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return nil, errors.New("图片数据格式错误")
		}
		imagesJSON = string(imagesBytes)
	}

	// 设置默认成色
	condition := req.Condition
	if condition == "" {
		condition = model.ConditionGood
	}

	// 创建商品
	item := &model.MarketItem{
		UserID:        userID,
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Condition:     condition,
		Images:        imagesJSON,
		Location:      req.Location,
		Contact:       req.Contact,
		Status:        model.ItemStatusOnSale,
	}

	if err := s.marketRepo.Create(ctx, item); err != nil {
		logger.Error("创建商品失败", zap.Error(err))
		return nil, errors.New("创建商品失败")
	}

	logger.Info("商品发布成功", zap.Uint("item_id", item.ID), zap.Uint("user_id", userID))
	return item, nil
}

// GetItemList 获取商品列表
func (s *marketService) GetItemList(ctx context.Context, query *model.MarketItemListQuery) (*ItemListResponse, error) {
	query.SetDefaults()

	items, total, err := s.marketRepo.List(ctx, query)
	if err != nil {
		logger.Error("获取商品列表失败", zap.Error(err))
		return nil, errors.New("获取商品列表失败")
	}

	return &ItemListResponse{
		List:     items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetItemDetail 获取商品详情
func (s *marketService) GetItemDetail(ctx context.Context, id, userID uint) (*ItemDetailResponse, error) {
	item, err := s.marketRepo.FindByIDWithUser(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品不存在")
		}
		logger.Error("获取商品详情失败", zap.Error(err))
		return nil, errors.New("获取商品详情失败")
	}

	// 增加浏览次数
	go func() {
		_ = s.marketRepo.IncrementViewCount(context.Background(), id)
	}()

	// 检查是否已收藏
	isFavorited := false
	if userID > 0 {
		isFavorited, _ = s.marketRepo.IsFavorited(ctx, id, userID)
	}

	return &ItemDetailResponse{
		MarketItem:  *item,
		IsFavorited: isFavorited,
	}, nil
}

// UpdateItem 更新商品
func (s *marketService) UpdateItem(ctx context.Context, userID, itemID uint, req *UpdateItemRequest) (*model.MarketItem, error) {
	item, err := s.marketRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品不存在")
		}
		return nil, errors.New("获取商品失败")
	}

	// 验证权限
	if !item.IsOwner(userID) {
		return nil, errors.New("无权修改此商品")
	}

	// 只有在售状态可以修改
	if !item.CanEdit() {
		return nil, errors.New("商品状态不允许修改")
	}

	// 更新字段
	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.Price >= 0 {
		item.Price = req.Price
	}
	if req.OriginalPrice >= 0 {
		item.OriginalPrice = req.OriginalPrice
	}
	if req.Condition != "" {
		item.Condition = req.Condition
	}
	if len(req.Images) > 0 {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return nil, errors.New("图片数据格式错误")
		}
		item.Images = string(imagesBytes)
	}
	if req.Location != "" {
		item.Location = req.Location
	}
	if req.Contact != "" {
		item.Contact = req.Contact
	}

	item.UpdatedAt = time.Now()

	if err := s.marketRepo.Update(ctx, item); err != nil {
		logger.Error("更新商品失败", zap.Error(err))
		return nil, errors.New("更新商品失败")
	}

	return item, nil
}

// DeleteItem 删除商品
func (s *marketService) DeleteItem(ctx context.Context, userID, itemID uint) error {
	item, err := s.marketRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品不存在")
		}
		return errors.New("获取商品失败")
	}

	// 验证权限
	if !item.IsOwner(userID) {
		return errors.New("无权删除此商品")
	}

	// 已售出商品不能删除
	if !item.CanDelete() {
		return errors.New("已售出商品无法删除")
	}

	// 硬删除
	if err := s.marketRepo.DB().Delete(item).Error; err != nil {
		logger.Error("删除商品失败", zap.Error(err))
		return errors.New("删除商品失败")
	}

	return nil
}

// MarkAsSold 标记为已售出
func (s *marketService) MarkAsSold(ctx context.Context, userID, itemID uint) error {
	item, err := s.marketRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品不存在")
		}
		return errors.New("获取商品失败")
	}

	// 验证权限
	if !item.IsOwner(userID) {
		return errors.New("无权操作此商品")
	}

	// 只有在售状态可以标记
	if item.Status != model.ItemStatusOnSale {
		return errors.New("商品状态不允许此操作")
	}

	if err := s.marketRepo.UpdateStatus(ctx, itemID, model.ItemStatusSold); err != nil {
		logger.Error("标记已售失败", zap.Error(err))
		return errors.New("标记已售失败")
	}

	return nil
}

// OffShelf 下架商品
func (s *marketService) OffShelf(ctx context.Context, userID, itemID uint) error {
	item, err := s.marketRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品不存在")
		}
		return errors.New("获取商品失败")
	}

	// 验证权限
	if !item.IsOwner(userID) {
		return errors.New("无权操作此商品")
	}

	// 只有在售状态可以下架
	if item.Status != model.ItemStatusOnSale {
		return errors.New("商品状态不允许此操作")
	}

	if err := s.marketRepo.UpdateStatus(ctx, itemID, model.ItemStatusOffShelf); err != nil {
		logger.Error("下架商品失败", zap.Error(err))
		return errors.New("下架商品失败")
	}

	return nil
}

// OnShelf 重新上架
func (s *marketService) OnShelf(ctx context.Context, userID, itemID uint) error {
	item, err := s.marketRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品不存在")
		}
		return errors.New("获取商品失败")
	}

	// 验证权限
	if !item.IsOwner(userID) {
		return errors.New("无权操作此商品")
	}

	// 只有已下架状态可以重新上架
	if item.Status != model.ItemStatusOffShelf {
		return errors.New("商品状态不允许此操作")
	}

	if err := s.marketRepo.UpdateStatus(ctx, itemID, model.ItemStatusOnSale); err != nil {
		logger.Error("上架商品失败", zap.Error(err))
		return errors.New("上架商品失败")
	}

	return nil
}

// AddFavorite 添加收藏
func (s *marketService) AddFavorite(ctx context.Context, itemID, userID uint) error {
	// 检查商品是否存在
	item, err := s.marketRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品不存在")
		}
		return errors.New("获取商品失败")
	}

	// 不能收藏自己的商品
	if item.IsOwner(userID) {
		return errors.New("不能收藏自己的商品")
	}

	// 检查是否已收藏
	isFavorited, err := s.marketRepo.IsFavorited(ctx, itemID, userID)
	if err != nil {
		return errors.New("检查收藏状态失败")
	}
	if isFavorited {
		return errors.New("已收藏该商品")
	}

	// 创建收藏
	favorite := &model.MarketFavorite{
		ItemID: itemID,
		UserID: userID,
	}

	if err := s.marketRepo.CreateFavorite(ctx, favorite); err != nil {
		logger.Error("添加收藏失败", zap.Error(err))
		return errors.New("添加收藏失败")
	}

	return nil
}

// RemoveFavorite 取消收藏
func (s *marketService) RemoveFavorite(ctx context.Context, itemID, userID uint) error {
	if err := s.marketRepo.DeleteFavorite(ctx, itemID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("未收藏该商品")
		}
		logger.Error("取消收藏失败", zap.Error(err))
		return errors.New("取消收藏失败")
	}

	return nil
}

// GetMyFavorites 获取我的收藏
func (s *marketService) GetMyFavorites(ctx context.Context, userID uint, page, pageSize int) (*FavoriteListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	favorites, total, err := s.marketRepo.FindFavoritesByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取收藏列表失败", zap.Error(err))
		return nil, errors.New("获取收藏列表失败")
	}

	return &FavoriteListResponse{
		List:     favorites,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetMyItems 获取我发布的商品
func (s *marketService) GetMyItems(ctx context.Context, userID uint, page, pageSize int) (*ItemListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	items, total, err := s.marketRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取我的商品失败", zap.Error(err))
		return nil, errors.New("获取我的商品失败")
	}

	return &ItemListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ValidateCategory 验证分类
func ValidateCategory(c string) bool {
	for _, valid := range model.ValidCategories {
		if c == valid {
			return true
		}
	}
	return false
}

// ValidateCondition 验证成色
func ValidateCondition(c string) bool {
	for _, valid := range model.ValidConditions {
		if c == valid {
			return true
		}
	}
	return false
}

// FormatCategory 格式化分类名称
func FormatCategory(c string) string {
	if name, ok := model.CategoryNames[c]; ok {
		return name
	}
	return "未知分类"
}

// FormatCondition 格式化成色名称
func FormatCondition(c string) string {
	if name, ok := model.ConditionNames[c]; ok {
		return name
	}
	return "未知成色"
}

// FormatItemStatus 格式化状态名称
func FormatItemStatus(s string) string {
	if name, ok := model.StatusNames[s]; ok {
		return name
	}
	return "未知状态"
}