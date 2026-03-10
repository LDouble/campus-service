package model

import (
	"time"

	"gorm.io/gorm"
)

// LostFound 失物招领
type LostFound struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"index;not null"`
	Type        string     `json:"type" gorm:"size:20;not null;index"` // lost, found
	Title       string     `json:"title" gorm:"size:100;not null"`
	Description string     `json:"description" gorm:"type:text"`
	Category    string     `json:"category" gorm:"size:20;not null;index"` // electronics, cards, books, clothes, bags, others
	Images      string     `json:"images" gorm:"type:text"` // JSON数组存储图片URL
	Location    string     `json:"location" gorm:"size:200"` // 丢失/拾取地点
	LostTime    *time.Time `json:"lost_time" gorm:"type:datetime"` // 丢失时间
	Contact     string     `json:"contact" gorm:"size:50"` // 联系方式
	Status      string     `json:"status" gorm:"size:20;default:'open';index"` // open, closed, resolved
	ViewCount   int        `json:"view_count" gorm:"default:0"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	User        User       `json:"user" gorm:"foreignKey:UserID"`
	Claims      []LostFoundClaim `json:"claims,omitempty" gorm:"foreignKey:LostFoundID"`
}

// LostFoundClaim 认领记录
type LostFoundClaim struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	LostFoundID uint      `json:"lost_found_id" gorm:"index;not null"`
	UserID      uint      `json:"user_id" gorm:"index;not null"`
	Message     string     `json:"message" gorm:"type:text"` // 认领说明
	Proof      string     `json:"proof" gorm:"type:text"` // 证明材料（图片URL）
	Status      string    `json:"status" gorm:"size:20;default:'pending';index"` // pending, approved, rejected
	CreatedAt   time.Time `json:"created_at"`

	LostFound   LostFound `json:"lost_found" gorm:"foreignKey:LostFoundID"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (LostFound) TableName() string {
	return "lost_founds"
}

// TableName 指定表名
func (LostFoundClaim) TableName() string {
	return "lost_found_claims"
}

// LostFoundType 类型常量
const (
	TypeLost  = "lost"  // 寻物启事
	TypeFound = "found" // 失物招领
)

// LostFoundCategory 分类常量
const (
	CategoryLFElectronics = "electronics" // 电子产品
	CategoryLFCards       = "cards"       // 证件卡片
	CategoryLFBooks       = "books"       // 书籍文具
	CategoryLFClothes     = "clothes"     // 服饰配件
	CategoryLFBags        = "bags"        // 包袋箱包
	CategoryLFOthers      = "others"      // 其他
)

// LostFoundStatus 状态常量
const (
	LFStatusOpen     = "open"     // 进行中
	LFStatusClosed   = "closed"   // 已关闭
	LFStatusResolved = "resolved" // 已解决
)

// ClaimStatus 认领状态常量
const (
	ClaimStatusPending  = "pending"  // 待审核
	ClaimStatusApproved = "approved" // 已通过
	ClaimStatusRejected = "rejected" // 已拒绝
)

// ValidLostFoundTypes 有效的类型
var ValidLostFoundTypes = []string{
	TypeLost,
	TypeFound,
}

// ValidLostFoundCategories 有效的分类
var ValidLostFoundCategories = []string{
	CategoryLFElectronics,
	CategoryLFCards,
	CategoryLFBooks,
	CategoryLFClothes,
	CategoryLFBags,
	CategoryLFOthers,
}

// ValidLostFoundStatuses 有效的状态
var ValidLostFoundStatuses = []string{
	LFStatusOpen,
	LFStatusClosed,
	LFStatusResolved,
}

// ValidClaimStatuses 有效的认领状态
var ValidClaimStatuses = []string{
	ClaimStatusPending,
	ClaimStatusApproved,
	ClaimStatusRejected,
}

// IsValidType 检查类型是否有效
func (lf *LostFound) IsValidType() bool {
	for _, t := range ValidLostFoundTypes {
		if lf.Type == t {
			return true
		}
	}
	return false
}

// IsValidCategory 检查分类是否有效
func (lf *LostFound) IsValidCategory() bool {
	for _, c := range ValidLostFoundCategories {
		if lf.Category == c {
			return true
		}
	}
	return false
}

// IsValidStatus 检查状态是否有效
func (lf *LostFound) IsValidStatus() bool {
	for _, s := range ValidLostFoundStatuses {
		if lf.Status == s {
			return true
		}
	}
	return false
}

// CanEdit 检查是否可以编辑
func (lf *LostFound) CanEdit() bool {
	return lf.Status == LFStatusOpen
}

// CanDelete 检查是否可以删除
func (lf *LostFound) CanDelete() bool {
	return lf.Status != LFStatusResolved
}

// IsOwner 检查是否是发布者
func (lf *LostFound) IsOwner(userID uint) bool {
	return lf.UserID == userID
}

// BeforeCreate GORM钩子 - 创建前验证
func (lf *LostFound) BeforeCreate(tx *gorm.DB) error {
	if !lf.IsValidType() {
		lf.Type = TypeLost
	}
	if !lf.IsValidCategory() {
		lf.Category = CategoryLFOthers
	}
	if !lf.IsValidStatus() {
		lf.Status = LFStatusOpen
	}
	return nil
}

// LostFoundListQuery 列表查询参数
type LostFoundListQuery struct {
	Page     int        `form:"page" binding:"omitempty,min=1"`
	PageSize int        `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type     string     `form:"type" binding:"omitempty,oneof=lost found ''"`
	Category string     `form:"category" binding:"omitempty,oneof=electronics cards books clothes bags others ''"`
	Status   string     `form:"status" binding:"omitempty,oneof=open closed resolved ''"`
	Keyword  string     `form:"keyword" binding:"omitempty,max=50"`
	SortBy   string     `form:"sort_by" binding:"omitempty,oneof=time time_desc ''"`
}

// SetDefaults 设置默认值
func (q *LostFoundListQuery) SetDefaults() {
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
func (q *LostFoundListQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// LostFoundTypeNames 类型名称映射
var LostFoundTypeNames = map[string]string{
	TypeLost:  "寻物启事",
	TypeFound: "失物招领",
}

// LostFoundCategoryNames 分类名称映射
var LostFoundCategoryNames = map[string]string{
	CategoryLFElectronics: "电子产品",
	CategoryLFCards:       "证件卡片",
	CategoryLFBooks:       "书籍文具",
	CategoryLFClothes:     "服饰配件",
	CategoryLFBags:        "包袋箱包",
	CategoryLFOthers:      "其他",
}

// LostFoundStatusNames 状态名称映射
var LostFoundStatusNames = map[string]string{
	LFStatusOpen:     "进行中",
	LFStatusClosed:   "已关闭",
	LFStatusResolved: "已解决",
}

// ClaimStatusNames 认领状态名称映射
var ClaimStatusNames = map[string]string{
	ClaimStatusPending:  "待审核",
	ClaimStatusApproved: "已通过",
	ClaimStatusRejected: "已拒绝",
}

// GetTypeName 获取类型名称
func (lf *LostFound) GetTypeName() string {
	if name, ok := LostFoundTypeNames[lf.Type]; ok {
		return name
	}
	return "未知类型"
}

// GetCategoryName 获取分类名称
func (lf *LostFound) GetCategoryName() string {
	if name, ok := LostFoundCategoryNames[lf.Category]; ok {
		return name
	}
	return "未知分类"
}

// GetStatusName 获取状态名称
func (lf *LostFound) GetStatusName() string {
	if name, ok := LostFoundStatusNames[lf.Status]; ok {
		return name
	}
	return "未知状态"
}

// GetClaimStatusName 获取认领状态名称
func (c *LostFoundClaim) GetClaimStatusName() string {
	if name, ok := ClaimStatusNames[c.Status]; ok {
		return name
	}
	return "未知状态"
}