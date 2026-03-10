package repository

import (
	"context"

	"gorm.io/gorm"

	"campus-service/internal/model"
)

// MarketRepository 二手交易仓储接口
type MarketRepository interface {
	// 商品相关
	Create(ctx context.Context, item *model.MarketItem) error
	Update(ctx context.Context, item *model.MarketItem) error
	FindByID(ctx context.Context, id uint) (*model.MarketItem, error)
	FindByIDWithUser(ctx context.Context, id uint) (*model.MarketItem, error)
	List(ctx context.Context, query *model.MarketItemListQuery) ([]model.MarketItem, int64, error)
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.MarketItem, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	IncrementViewCount(ctx context.Context, id uint) error

	// 收藏相关
	CreateFavorite(ctx context.Context, favorite *model.MarketFavorite) error
	DeleteFavorite(ctx context.Context, itemID, userID uint) error
	FindFavorite(ctx context.Context, itemID, userID uint) (*model.MarketFavorite, error)
	FindFavoritesByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.MarketFavorite, int64, error)
	IsFavorited(ctx context.Context, itemID, userID uint) (bool, error)

	// 事务支持
	WithTx(tx *gorm.DB) MarketRepository
	DB() *gorm.DB
}

// marketRepository 二手交易仓储实现
type marketRepository struct {
	db *gorm.DB
}

// NewMarketRepository 创建二手交易仓储实例
func NewMarketRepository(db *gorm.DB) MarketRepository {
	return &marketRepository{db: db}
}

// Create 创建商品
func (r *marketRepository) Create(ctx context.Context, item *model.MarketItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update 更新商品
func (r *marketRepository) Update(ctx context.Context, item *model.MarketItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// FindByID 根据ID查找商品
func (r *marketRepository) FindByID(ctx context.Context, id uint) (*model.MarketItem, error) {
	var item model.MarketItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// FindByIDWithUser 根据ID查找商品（包含用户信息）
func (r *marketRepository) FindByIDWithUser(ctx context.Context, id uint) (*model.MarketItem, error) {
	var item model.MarketItem
	err := r.db.WithContext(ctx).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// List 分页查询商品列表
func (r *marketRepository) List(ctx context.Context, query *model.MarketItemListQuery) ([]model.MarketItem, int64, error) {
	var items []model.MarketItem
	var total int64

	db := r.db.WithContext(ctx).Model(&model.MarketItem{})

	// 筛选条件
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	} else {
		// 默认只显示在售商品
		db = db.Where("status = ?", model.ItemStatusOnSale)
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("title LIKE ? OR description LIKE ?", keyword, keyword)
	}
	if query.MinPrice > 0 {
		db = db.Where("price >= ?", query.MinPrice)
	}
	if query.MaxPrice > 0 {
		db = db.Where("price <= ?", query.MaxPrice)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC"
	switch query.SortBy {
	case "price":
		orderBy = "price ASC"
	case "price_desc":
		orderBy = "price DESC"
	case "time":
		orderBy = "created_at ASC"
	case "time_desc":
		orderBy = "created_at DESC"
	}

	// 分页查询
	err := db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	}).
		Order(orderBy).
		Offset(query.Offset()).
		Limit(query.PageSize).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListByUserID 查询用户发布的商品
func (r *marketRepository) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.MarketItem, int64, error) {
	var items []model.MarketItem
	var total int64

	db := r.db.WithContext(ctx).Model(&model.MarketItem{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// UpdateStatus 更新商品状态
func (r *marketRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.MarketItem{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// IncrementViewCount 增加浏览次数
func (r *marketRepository) IncrementViewCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&model.MarketItem{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// CreateFavorite 创建收藏
func (r *marketRepository) CreateFavorite(ctx context.Context, favorite *model.MarketFavorite) error {
	return r.db.WithContext(ctx).Create(favorite).Error
}

// DeleteFavorite 删除收藏
func (r *marketRepository) DeleteFavorite(ctx context.Context, itemID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("item_id = ? AND user_id = ?", itemID, userID).
		Delete(&model.MarketFavorite{}).Error
}

// FindFavorite 查找收藏记录
func (r *marketRepository) FindFavorite(ctx context.Context, itemID, userID uint) (*model.MarketFavorite, error) {
	var favorite model.MarketFavorite
	err := r.db.WithContext(ctx).
		Where("item_id = ? AND user_id = ?", itemID, userID).
		First(&favorite).Error
	if err != nil {
		return nil, err
	}
	return &favorite, nil
}

// FindFavoritesByUserID 查询用户收藏的商品
func (r *marketRepository) FindFavoritesByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.MarketFavorite, int64, error) {
	var favorites []model.MarketFavorite
	var total int64

	db := r.db.WithContext(ctx).Model(&model.MarketFavorite{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Preload("Item").
		Preload("Item.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&favorites).Error

	if err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}

// IsFavorited 检查是否已收藏
func (r *marketRepository) IsFavorited(ctx context.Context, itemID, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.MarketFavorite{}).
		Where("item_id = ? AND user_id = ?", itemID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// WithTx 返回使用事务的仓储实例
func (r *marketRepository) WithTx(tx *gorm.DB) MarketRepository {
	return &marketRepository{db: tx}
}

// DB 返回底层数据库连接（用于事务）
func (r *marketRepository) DB() *gorm.DB {
	return r.db
}