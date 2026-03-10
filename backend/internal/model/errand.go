package model

import (
	"time"

	"gorm.io/gorm"
)

// Errand 跑腿任务
type Errand struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"index;not null"`
	Title       string     `json:"title" gorm:"size:100;not null"`
	Type        string     `json:"type" gorm:"size:20;not null;index"` // express, food, print, other
	Description string     `json:"description" gorm:"type:text"`
	Reward      float64    `json:"reward" gorm:"default:0;not null"`
	Location    string     `json:"location" gorm:"size:200"`
	Deadline    *time.Time `json:"deadline"`
	Status      string     `json:"status" gorm:"size:20;default:'pending';index"` // pending, in_progress, completed, cancelled
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	User   User         `json:"user" gorm:"foreignKey:UserID"`
	Orders []ErrandOrder `json:"orders,omitempty" gorm:"foreignKey:ErrandID"`
}

// TableName 指定表名
func (Errand) TableName() string {
	return "errands"
}

// ErrandOrder 跑腿订单
type ErrandOrder struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ErrandID  uint      `json:"errand_id" gorm:"index;not null"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Status    string    `json:"status" gorm:"size:20;default:'pending';index"` // pending, accepted, cancelled
	Remark    string    `json:"remark" gorm:"size:200"`
	CreatedAt time.Time `json:"created_at"`

	Errand Errand `json:"errand" gorm:"foreignKey:ErrandID"`
	User   User   `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (ErrandOrder) TableName() string {
	return "errand_orders"
}

// ErrandType 跑腿类型常量
const (
	ErrandTypeExpress = "express" // 快递代取
	ErrandTypeFood    = "food"    // 外卖代拿
	ErrandTypePrint   = "print"   // 打印复印
	ErrandTypeOther   = "other"   // 其他
)

// ErrandStatus 跑腿状态常量
const (
	ErrandStatusPending    = "pending"     // 待接单
	ErrandStatusInProgress = "in_progress" // 进行中
	ErrandStatusCompleted  = "completed"   // 已完成
	ErrandStatusCancelled  = "cancelled"   // 已取消
)

// ErrandOrderStatus 订单状态常量
const (
	OrderStatusPending   = "pending"   // 待处理
	OrderStatusAccepted  = "accepted"  // 已接受
	OrderStatusCompleted = "completed" // 已完成
	OrderStatusCancelled = "cancelled" // 已取消
)

// ValidErrandTypes 有效的跑腿类型
var ValidErrandTypes = []string{
	ErrandTypeExpress,
	ErrandTypeFood,
	ErrandTypePrint,
	ErrandTypeOther,
}

// ValidErrandStatuses 有效的跑腿状态
var ValidErrandStatuses = []string{
	ErrandStatusPending,
	ErrandStatusInProgress,
	ErrandStatusCompleted,
	ErrandStatusCancelled,
}

// ValidOrderStatuses 有效的订单状态
var ValidOrderStatuses = []string{
	OrderStatusPending,
	OrderStatusAccepted,
	OrderStatusCompleted,
	OrderStatusCancelled,
}

// IsValidType 检查类型是否有效
func (e *Errand) IsValidType() bool {
	for _, t := range ValidErrandTypes {
		if e.Type == t {
			return true
		}
	}
	return false
}

// IsValidStatus 检查状态是否有效
func (e *Errand) IsValidStatus() bool {
	for _, s := range ValidErrandStatuses {
		if e.Status == s {
			return true
		}
	}
	return false
}

// CanAccept 检查是否可以接单
func (e *Errand) CanAccept() bool {
	return e.Status == ErrandStatusPending
}

// CanCancel 检查是否可以取消
func (e *Errand) CanCancel() bool {
	return e.Status == ErrandStatusPending || e.Status == ErrandStatusInProgress
}

// CanComplete 检查是否可以完成
func (e *Errand) CanComplete() bool {
	return e.Status == ErrandStatusInProgress
}

// IsOwner 检查是否是任务发布者
func (e *Errand) IsOwner(userID uint) bool {
	return e.UserID == userID
}

// BeforeCreate GORM钩子 - 创建前验证
func (e *Errand) BeforeCreate(tx *gorm.DB) error {
	if !e.IsValidType() {
		e.Type = ErrandTypeOther
	}
	if !e.IsValidStatus() {
		e.Status = ErrandStatusPending
	}
	return nil
}

// ErrandListQuery 跑腿列表查询参数
type ErrandListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=express food print other ''"`
	Status   string `form:"status" binding:"omitempty,oneof=pending in_progress completed cancelled ''"`
	Keyword  string `form:"keyword" binding:"omitempty,max=50"`
}

// SetDefaults 设置默认值
func (q *ErrandListQuery) SetDefaults() {
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
func (q *ErrandListQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}