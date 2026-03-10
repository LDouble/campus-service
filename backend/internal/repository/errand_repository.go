package repository

import (
	"context"

	"gorm.io/gorm"

	"campus-service/internal/model"
)

// ErrandRepository 跑腿仓储接口
type ErrandRepository interface {
	// Errand 相关
	Create(ctx context.Context, errand *model.Errand) error
	Update(ctx context.Context, errand *model.Errand) error
	FindByID(ctx context.Context, id uint) (*model.Errand, error)
	FindByIDWithUser(ctx context.Context, id uint) (*model.Errand, error)
	List(ctx context.Context, query *model.ErrandListQuery) ([]model.Errand, int64, error)
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.Errand, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error

	// ErrandOrder 相关
	CreateOrder(ctx context.Context, order *model.ErrandOrder) error
	UpdateOrder(ctx context.Context, order *model.ErrandOrder) error
	FindOrderByID(ctx context.Context, id uint) (*model.ErrandOrder, error)
	FindOrderByErrandID(ctx context.Context, errandID uint) (*model.ErrandOrder, error)
	FindAcceptedOrderByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.ErrandOrder, int64, error)
	FindPendingOrderByErrandID(ctx context.Context, errandID uint) (*model.ErrandOrder, error)
	DeletePendingOrdersByErrandID(ctx context.Context, errandID uint) error

	// 事务支持
	WithTx(tx *gorm.DB) ErrandRepository
	DB() *gorm.DB
}

// errandRepository 跑腿仓储实现
type errandRepository struct {
	db *gorm.DB
}

// NewErrandRepository 创建跑腿仓储实例
func NewErrandRepository(db *gorm.DB) ErrandRepository {
	return &errandRepository{db: db}
}

// Create 创建跑腿任务
func (r *errandRepository) Create(ctx context.Context, errand *model.Errand) error {
	return r.db.WithContext(ctx).Create(errand).Error
}

// Update 更新跑腿任务
func (r *errandRepository) Update(ctx context.Context, errand *model.Errand) error {
	return r.db.WithContext(ctx).Save(errand).Error
}

// FindByID 根据ID查找跑腿任务
func (r *errandRepository) FindByID(ctx context.Context, id uint) (*model.Errand, error) {
	var errand model.Errand
	err := r.db.WithContext(ctx).First(&errand, id).Error
	if err != nil {
		return nil, err
	}
	return &errand, nil
}

// FindByIDWithUser 根据ID查找跑腿任务（包含用户信息）
func (r *errandRepository) FindByIDWithUser(ctx context.Context, id uint) (*model.Errand, error) {
	var errand model.Errand
	err := r.db.WithContext(ctx).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		First(&errand, id).Error
	if err != nil {
		return nil, err
	}
	return &errand, nil
}

// List 分页查询跑腿任务列表
func (r *errandRepository) List(ctx context.Context, query *model.ErrandListQuery) ([]model.Errand, int64, error) {
	var errands []model.Errand
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Errand{})

	// 筛选条件
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("title LIKE ? OR description LIKE ?", keyword, keyword)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	}).
		Order("created_at DESC").
		Offset(query.Offset()).
		Limit(query.PageSize).
		Find(&errands).Error

	if err != nil {
		return nil, 0, err
	}

	return errands, total, nil
}

// ListByUserID 查询用户发布的跑腿任务
func (r *errandRepository) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.Errand, int64, error) {
	var errands []model.Errand
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Errand{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&errands).Error

	if err != nil {
		return nil, 0, err
	}

	return errands, total, nil
}

// UpdateStatus 更新跑腿任务状态
func (r *errandRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.Errand{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// CreateOrder 创建跑腿订单
func (r *errandRepository) CreateOrder(ctx context.Context, order *model.ErrandOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// UpdateOrder 更新跑腿订单
func (r *errandRepository) UpdateOrder(ctx context.Context, order *model.ErrandOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// FindOrderByID 根据ID查找跑腿订单
func (r *errandRepository) FindOrderByID(ctx context.Context, id uint) (*model.ErrandOrder, error) {
	var order model.ErrandOrder
	err := r.db.WithContext(ctx).
		Preload("Errand").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindOrderByErrandID 根据跑腿任务ID查找订单
func (r *errandRepository) FindOrderByErrandID(ctx context.Context, errandID uint) (*model.ErrandOrder, error) {
	var order model.ErrandOrder
	err := r.db.WithContext(ctx).
		Where("errand_id = ?", errandID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindAcceptedOrderByUserID 查询用户接受的跑腿订单
func (r *errandRepository) FindAcceptedOrderByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.ErrandOrder, int64, error) {
	var orders []model.ErrandOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&model.ErrandOrder{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Preload("Errand").
		Preload("Errand.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// FindPendingOrderByErrandID 查找跑腿任务的待处理订单
func (r *errandRepository) FindPendingOrderByErrandID(ctx context.Context, errandID uint) (*model.ErrandOrder, error) {
	var order model.ErrandOrder
	err := r.db.WithContext(ctx).
		Where("errand_id = ? AND status = ?", errandID, model.OrderStatusPending).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// DeletePendingOrdersByErrandID 删除跑腿任务的所有待处理订单
func (r *errandRepository) DeletePendingOrdersByErrandID(ctx context.Context, errandID uint) error {
	return r.db.WithContext(ctx).
		Where("errand_id = ? AND status = ?", errandID, model.OrderStatusPending).
		Delete(&model.ErrandOrder{}).Error
}

// WithTx 返回使用事务的仓储实例
func (r *errandRepository) WithTx(tx *gorm.DB) ErrandRepository {
	return &errandRepository{db: tx}
}

// DB 返回底层数据库连接（用于事务）
func (r *errandRepository) DB() *gorm.DB {
	return r.db
}