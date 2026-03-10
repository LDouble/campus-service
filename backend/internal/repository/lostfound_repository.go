package repository

import (
	"context"

	"gorm.io/gorm"

	"campus-service/internal/model"
)

// LostFoundRepository 失物招领仓储接口
type LostFoundRepository interface {
	// 失物招领相关
	Create(ctx context.Context, lf *model.LostFound) error
	Update(ctx context.Context, lf *model.LostFound) error
	FindByID(ctx context.Context, id uint) (*model.LostFound, error)
	FindByIDWithUser(ctx context.Context, id uint) (*model.LostFound, error)
	List(ctx context.Context, query *model.LostFoundListQuery) ([]model.LostFound, int64, error)
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.LostFound, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	IncrementViewCount(ctx context.Context, id uint) error

	// 认领相关
	CreateClaim(ctx context.Context, claim *model.LostFoundClaim) error
	FindClaimByID(ctx context.Context, id uint) (*model.LostFoundClaim, error)
	FindClaimsByLostFoundID(ctx context.Context, lostFoundID uint, page, pageSize int) ([]model.LostFoundClaim, int64, error)
	FindClaimsByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.LostFoundClaim, int64, error)
	UpdateClaimStatus(ctx context.Context, id uint, status string) error
	HasPendingClaim(ctx context.Context, lostFoundID, userID uint) (bool, error)

	// 事务支持
	WithTx(tx *gorm.DB) LostFoundRepository
	DB() *gorm.DB
}

// lostFoundRepository 失物招领仓储实现
type lostFoundRepository struct {
	db *gorm.DB
}

// NewLostFoundRepository 创建失物招领仓储实例
func NewLostFoundRepository(db *gorm.DB) LostFoundRepository {
	return &lostFoundRepository{db: db}
}

// Create 创建失物招领记录
func (r *lostFoundRepository) Create(ctx context.Context, lf *model.LostFound) error {
	return r.db.WithContext(ctx).Create(lf).Error
}

// Update 更新失物招领记录
func (r *lostFoundRepository) Update(ctx context.Context, lf *model.LostFound) error {
	return r.db.WithContext(ctx).Save(lf).Error
}

// FindByID 根据ID查找失物招领记录
func (r *lostFoundRepository) FindByID(ctx context.Context, id uint) (*model.LostFound, error) {
	var lf model.LostFound
	err := r.db.WithContext(ctx).First(&lf, id).Error
	if err != nil {
		return nil, err
	}
	return &lf, nil
}

// FindByIDWithUser 根据ID查找失物招领记录（包含用户信息）
func (r *lostFoundRepository) FindByIDWithUser(ctx context.Context, id uint) (*model.LostFound, error) {
	var lf model.LostFound
	err := r.db.WithContext(ctx).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		First(&lf, id).Error
	if err != nil {
		return nil, err
	}
	return &lf, nil
}

// List 分页查询失物招领列表
func (r *lostFoundRepository) List(ctx context.Context, query *model.LostFoundListQuery) ([]model.LostFound, int64, error) {
	var items []model.LostFound
	var total int64

	db := r.db.WithContext(ctx).Model(&model.LostFound{})

	// 筛选条件
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	} else {
		// 默认只显示进行中的
		db = db.Where("status = ?", model.LFStatusOpen)
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("title LIKE ? OR description LIKE ?", keyword, keyword)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC"
	switch query.SortBy {
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

// ListByUserID 查询用户发布的失物招领
func (r *lostFoundRepository) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.LostFound, int64, error) {
	var items []model.LostFound
	var total int64

	db := r.db.WithContext(ctx).Model(&model.LostFound{}).Where("user_id = ?", userID)

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

// UpdateStatus 更新失物招领状态
func (r *lostFoundRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.LostFound{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// IncrementViewCount 增加浏览次数
func (r *lostFoundRepository) IncrementViewCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&model.LostFound{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// CreateClaim 创建认领记录
func (r *lostFoundRepository) CreateClaim(ctx context.Context, claim *model.LostFoundClaim) error {
	return r.db.WithContext(ctx).Create(claim).Error
}

// FindClaimByID 根据ID查找认领记录
func (r *lostFoundRepository) FindClaimByID(ctx context.Context, id uint) (*model.LostFoundClaim, error) {
	var claim model.LostFoundClaim
	err := r.db.WithContext(ctx).
		Preload("LostFound").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		First(&claim, id).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

// FindClaimsByLostFoundID 查询失物招领的认领记录
func (r *lostFoundRepository) FindClaimsByLostFoundID(ctx context.Context, lostFoundID uint, page, pageSize int) ([]model.LostFoundClaim, int64, error) {
	var claims []model.LostFoundClaim
	var total int64

	db := r.db.WithContext(ctx).Model(&model.LostFoundClaim{}).Where("lost_found_id = ?", lostFoundID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&claims).Error

	if err != nil {
		return nil, 0, err
	}

	return claims, total, nil
}

// FindClaimsByUserID 查询用户的认领记录
func (r *lostFoundRepository) FindClaimsByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.LostFoundClaim, int64, error) {
	var claims []model.LostFoundClaim
	var total int64

	db := r.db.WithContext(ctx).Model(&model.LostFoundClaim{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Preload("LostFound").
		Preload("LostFound.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&claims).Error

	if err != nil {
		return nil, 0, err
	}

	return claims, total, nil
}

// UpdateClaimStatus 更新认领状态
func (r *lostFoundRepository) UpdateClaimStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.LostFoundClaim{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// HasPendingClaim 检查是否已有待审核的认领
func (r *lostFoundRepository) HasPendingClaim(ctx context.Context, lostFoundID, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.LostFoundClaim{}).
		Where("lost_found_id = ? AND user_id = ? AND status = ?", lostFoundID, userID, model.ClaimStatusPending).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// WithTx 返回使用事务的仓储实例
func (r *lostFoundRepository) WithTx(tx *gorm.DB) LostFoundRepository {
	return &lostFoundRepository{db: tx}
}

// DB 返回底层数据库连接（用于事务）
func (r *lostFoundRepository) DB() *gorm.DB {
	return r.db
}