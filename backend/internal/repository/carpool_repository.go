package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"campus-service/internal/model"
)

// CarpoolRepository 拼车仓储接口
type CarpoolRepository interface {
	// Carpool 相关
	Create(ctx context.Context, carpool *model.Carpool) error
	Update(ctx context.Context, carpool *model.Carpool) error
	FindByID(ctx context.Context, id uint) (*model.Carpool, error)
	FindByIDWithUser(ctx context.Context, id uint) (*model.Carpool, error)
	List(ctx context.Context, query *model.CarpoolListQuery) ([]model.Carpool, int64, error)
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.Carpool, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	UpdateAvailableSeats(ctx context.Context, id uint, seats int) error

	// CarpoolBooking 相关
	CreateBooking(ctx context.Context, booking *model.CarpoolBooking) error
	UpdateBooking(ctx context.Context, booking *model.CarpoolBooking) error
	FindBookingByID(ctx context.Context, id uint) (*model.CarpoolBooking, error)
	FindBookingsByCarpoolID(ctx context.Context, carpoolID uint) ([]model.CarpoolBooking, error)
	FindBookingsByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.CarpoolBooking, int64, error)
	FindConfirmedBookingByUser(ctx context.Context, carpoolID, userID uint) (*model.CarpoolBooking, error)

	// 事务支持
	WithTx(tx *gorm.DB) CarpoolRepository
	DB() *gorm.DB
}

// carpoolRepository 拼车仓储实现
type carpoolRepository struct {
	db *gorm.DB
}

// NewCarpoolRepository 创建拼车仓储实例
func NewCarpoolRepository(db *gorm.DB) CarpoolRepository {
	return &carpoolRepository{db: db}
}

// Create 创建拼车行程
func (r *carpoolRepository) Create(ctx context.Context, carpool *model.Carpool) error {
	return r.db.WithContext(ctx).Create(carpool).Error
}

// Update 更新拼车行程
func (r *carpoolRepository) Update(ctx context.Context, carpool *model.Carpool) error {
	return r.db.WithContext(ctx).Save(carpool).Error
}

// FindByID 根据ID查找拼车行程
func (r *carpoolRepository) FindByID(ctx context.Context, id uint) (*model.Carpool, error) {
	var carpool model.Carpool
	err := r.db.WithContext(ctx).First(&carpool, id).Error
	if err != nil {
		return nil, err
	}
	return &carpool, nil
}

// FindByIDWithUser 根据ID查找拼车行程（包含用户信息）
func (r *carpoolRepository) FindByIDWithUser(ctx context.Context, id uint) (*model.Carpool, error) {
	var carpool model.Carpool
	err := r.db.WithContext(ctx).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		Preload("Bookings", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", model.BookingStatusConfirmed).
				Preload("User", "id, nickname, avatar")
		}).
		First(&carpool, id).Error
	if err != nil {
		return nil, err
	}
	return &carpool, nil
}

// List 分页查询拼车行程列表
func (r *carpoolRepository) List(ctx context.Context, query *model.CarpoolListQuery) ([]model.Carpool, int64, error) {
	var carpools []model.Carpool
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Carpool{})

	// 筛选条件
	if query.Origin != "" {
		origin := "%" + query.Origin + "%"
		db = db.Where("origin LIKE ?", origin)
	}
	if query.Destination != "" {
		destination := "%" + query.Destination + "%"
		db = db.Where("destination LIKE ?", destination)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Date != "" {
		// 筛选某天的行程
		date, err := time.Parse("2006-01-02", query.Date)
		if err == nil {
			nextDay := date.AddDate(0, 0, 1)
			db = db.Where("departure_time >= ? AND departure_time < ?", date, nextDay)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	}).
		Order("departure_time ASC").
		Offset(query.Offset()).
		Limit(query.PageSize).
		Find(&carpools).Error

	if err != nil {
		return nil, 0, err
	}

	return carpools, total, nil
}

// ListByUserID 查询用户发布的拼车行程
func (r *carpoolRepository) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.Carpool, int64, error) {
	var carpools []model.Carpool
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Carpool{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Order("departure_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&carpools).Error

	if err != nil {
		return nil, 0, err
	}

	return carpools, total, nil
}

// UpdateStatus 更新拼车行程状态
func (r *carpoolRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.Carpool{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateAvailableSeats 更新可用座位数
func (r *carpoolRepository) UpdateAvailableSeats(ctx context.Context, id uint, seats int) error {
	return r.db.WithContext(ctx).
		Model(&model.Carpool{}).
		Where("id = ?", id).
		Update("available_seats", seats).Error
}

// CreateBooking 创建拼车预约
func (r *carpoolRepository) CreateBooking(ctx context.Context, booking *model.CarpoolBooking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

// UpdateBooking 更新拼车预约
func (r *carpoolRepository) UpdateBooking(ctx context.Context, booking *model.CarpoolBooking) error {
	return r.db.WithContext(ctx).Save(booking).Error
}

// FindBookingByID 根据ID查找拼车预约
func (r *carpoolRepository) FindBookingByID(ctx context.Context, id uint) (*model.CarpoolBooking, error) {
	var booking model.CarpoolBooking
	err := r.db.WithContext(ctx).
		Preload("Carpool").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		First(&booking, id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// FindBookingsByCarpoolID 根据拼车行程ID查找预约
func (r *carpoolRepository) FindBookingsByCarpoolID(ctx context.Context, carpoolID uint) ([]model.CarpoolBooking, error) {
	var bookings []model.CarpoolBooking
	err := r.db.WithContext(ctx).
		Where("carpool_id = ?", carpoolID).
		Preload("User", "id, nickname, avatar").
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

// FindBookingsByUserID 查询用户的拼车预约
func (r *carpoolRepository) FindBookingsByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.CarpoolBooking, int64, error) {
	var bookings []model.CarpoolBooking
	var total int64

	db := r.db.WithContext(ctx).Model(&model.CarpoolBooking{}).Where("user_id = ?", userID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Preload("Carpool").
		Preload("Carpool.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar, phone")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&bookings).Error

	if err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

// FindConfirmedBookingByUser 查找用户在某个行程的已确认预约
func (r *carpoolRepository) FindConfirmedBookingByUser(ctx context.Context, carpoolID, userID uint) (*model.CarpoolBooking, error) {
	var booking model.CarpoolBooking
	err := r.db.WithContext(ctx).
		Where("carpool_id = ? AND user_id = ? AND status = ?", carpoolID, userID, model.BookingStatusConfirmed).
		First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// WithTx 返回使用事务的仓储实例
func (r *carpoolRepository) WithTx(tx *gorm.DB) CarpoolRepository {
	return &carpoolRepository{db: tx}
}

// DB 返回底层数据库连接（用于事务）
func (r *carpoolRepository) DB() *gorm.DB {
	return r.db
}