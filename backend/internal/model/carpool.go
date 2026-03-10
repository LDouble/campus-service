package model

import (
	"time"

	"gorm.io/gorm"
)

// Carpool 拼车行程
type Carpool struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	UserID        uint       `json:"user_id" gorm:"index;not null"`
	Origin        string     `json:"origin" gorm:"size:100;not null"`           // 出发地
	Destination   string     `json:"destination" gorm:"size:100;not null"`     // 目的地
	DepartureTime *time.Time `json:"departure_time" gorm:"index;not null"`      // 出发时间
	Seats         int        `json:"seats" gorm:"default:1;not null"`          // 总座位数
	AvailableSeats int       `json:"available_seats" gorm:"default:1;not null"` // 可用座位数
	Price         float64    `json:"price" gorm:"default:0;not null"`          // 费用（每人）
	VehicleType   string     `json:"vehicle_type" gorm:"size:20;default:'car'"` // 车辆类型: car, taxi, bus
	VehicleNumber string     `json:"vehicle_number" gorm:"size:20"`            // 车牌号（可选）
	Contact       string     `json:"contact" gorm:"size:50"`                   // 联系方式
	Remark        string     `json:"remark" gorm:"size:200"`                   // 备注
	Status        string     `json:"status" gorm:"size:20;default:'pending';index"` // pending, departed, cancelled
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	User    User           `json:"user" gorm:"foreignKey:UserID"`
	Bookings []CarpoolBooking `json:"bookings,omitempty" gorm:"foreignKey:CarpoolID"`
}

// TableName 指定表名
func (Carpool) TableName() string {
	return "carpools"
}

// CarpoolBooking 拼车预约
type CarpoolBooking struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CarpoolID uint      `json:"carpool_id" gorm:"index;not null"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Seats     int       `json:"seats" gorm:"default:1;not null"` // 预约座位数
	Status    string    `json:"status" gorm:"size:20;default:'pending';index"` // pending, confirmed, cancelled
	Remark    string    `json:"remark" gorm:"size:200"`
	CreatedAt time.Time `json:"created_at"`

	Carpool Carpool `json:"carpool" gorm:"foreignKey:CarpoolID"`
	User    User    `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (CarpoolBooking) TableName() string {
	return "carpool_bookings"
}

// CarpoolVehicleType 车辆类型常量
const (
	VehicleTypeCar  = "car"  // 私家车
	VehicleTypeTaxi = "taxi" // 出租车
	VehicleTypeBus  = "bus"  // 班车
)

// CarpoolStatus 行程状态常量
const (
	CarpoolStatusPending  = "pending"  // 待出发
	CarpoolStatusDeparted = "departed" // 已出发
	CarpoolStatusCancelled = "cancelled" // 已取消
)

// CarpoolBookingStatus 预约状态常量
const (
	BookingStatusPending   = "pending"   // 待确认
	BookingStatusConfirmed = "confirmed" // 已确认
	BookingStatusCancelled = "cancelled" // 已取消
)

// ValidVehicleTypes 有效的车辆类型
var ValidVehicleTypes = []string{
	VehicleTypeCar,
	VehicleTypeTaxi,
	VehicleTypeBus,
}

// ValidCarpoolStatuses 有效的行程状态
var ValidCarpoolStatuses = []string{
	CarpoolStatusPending,
	CarpoolStatusDeparted,
	CarpoolStatusCancelled,
}

// ValidBookingStatuses 有效的预约状态
var ValidBookingStatuses = []string{
	BookingStatusPending,
	BookingStatusConfirmed,
	BookingStatusCancelled,
}

// IsValidVehicleType 检查车辆类型是否有效
func (c *Carpool) IsValidVehicleType() bool {
	for _, t := range ValidVehicleTypes {
		if c.VehicleType == t {
			return true
		}
	}
	return false
}

// IsValidStatus 检查状态是否有效
func (c *Carpool) IsValidStatus() bool {
	for _, s := range ValidCarpoolStatuses {
		if c.Status == s {
			return true
		}
	}
	return false
}

// CanBook 检查是否可以预约
func (c *Carpool) CanBook() bool {
	return c.Status == CarpoolStatusPending && c.AvailableSeats > 0
}

// CanCancel 检查是否可以取消
func (c *Carpool) CanCancel() bool {
	return c.Status == CarpoolStatusPending
}

// CanDepart 检查是否可以出发
func (c *Carpool) CanDepart() bool {
	return c.Status == CarpoolStatusPending
}

// IsOwner 检查是否是行程发布者
func (c *Carpool) IsOwner(userID uint) bool {
	return c.UserID == userID
}

// HasAvailableSeats 检查是否有足够座位
func (c *Carpool) HasAvailableSeats(seats int) bool {
	return c.AvailableSeats >= seats
}

// BeforeCreate GORM钩子 - 创建前验证
func (c *Carpool) BeforeCreate(tx *gorm.DB) error {
	if !c.IsValidVehicleType() {
		c.VehicleType = VehicleTypeCar
	}
	if !c.IsValidStatus() {
		c.Status = CarpoolStatusPending
	}
	if c.Seats <= 0 {
		c.Seats = 1
	}
	if c.AvailableSeats <= 0 {
		c.AvailableSeats = c.Seats
	}
	return nil
}

// CarpoolListQuery 拼车列表查询参数
type CarpoolListQuery struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Origin      string `form:"origin" binding:"omitempty,max=50"`
	Destination string `form:"destination" binding:"omitempty,max=50"`
	Status      string `form:"status" binding:"omitempty,oneof=pending departed cancelled ''"`
	Date        string `form:"date" binding:"omitempty,datetime=2006-01-02"` // 筛选某天
}

// SetDefaults 设置默认值
func (q *CarpoolListQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
}

// Offset 计算偏移量
func (q *CarpoolListQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}