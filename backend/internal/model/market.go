package model

import (
	"time"

	"gorm.io/gorm"
)

// MarketItem 二手商品
type MarketItem struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"index;not null"`
	Title       string     `json:"title" gorm:"size:100;not null"`
	Description string     `json:"description" gorm:"type:text"`
	Category    string     `json:"category" gorm:"size:20;not null;index"` // electronics, books, clothes, sports, others
	Price       float64    `json:"price" gorm:"not null"`
	OriginalPrice float64  `json:"original_price" gorm:"default:0"` // 原价（可选）
	Condition   string     `json:"condition" gorm:"size:20;default:'good'"` // new, like_new, good, fair
	Images      string     `json:"images" gorm:"type:text"` // JSON数组存储图片URL
	Location    string     `json:"location" gorm:"size:200"` // 交易地点
	Contact     string     `json:"contact" gorm:"size:50"` // 联系方式
	Status      string     `json:"status" gorm:"size:20;default:'on_sale';index"` // on_sale, sold, off_shelf
	ViewCount   int        `json:"view_count" gorm:"default:0"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	User        User       `json:"user" gorm:"foreignKey:UserID"`
	Favorites   []MarketFavorite `json:"favorites,omitempty" gorm:"foreignKey:ItemID"`
}

// TableName 指定表名
func (MarketItem) TableName() string {
	return "market_items"
}

// MarketFavorite 商品收藏
type MarketFavorite struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ItemID    uint      `json:"item_id" gorm:"index;not null"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`

	Item      MarketItem `json:"item" gorm:"foreignKey:ItemID"`
	User      User       `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (MarketFavorite) TableName() string {
	return "market_favorites"
}

// MarketCategory 商品分类常量
const (
	CategoryElectronics = "electronics" // 电子产品
	CategoryBooks       = "books"       // 书籍
	CategoryClothes     = "clothes"     // 服饰
	CategorySports      = "sports"      // 运动用品
	CategoryDaily       = "daily"       // 日用品
	CategoryOthers      = "others"      // 其他
)

// MarketItemCondition 商品成色常量
const (
	ConditionNew     = "new"      // 全新
	ConditionLikeNew = "like_new" // 几乎全新
	ConditionGood    = "good"     // 良好
	ConditionFair    = "fair"     // 一般
)

// MarketItemStatus 商品状态常量
const (
	ItemStatusOnSale   = "on_sale"   // 在售中
	ItemStatusSold     = "sold"      // 已售出
	ItemStatusOffShelf = "off_shelf" // 已下架
)

// ValidCategories 有效的商品分类
var ValidCategories = []string{
	CategoryElectronics,
	CategoryBooks,
	CategoryClothes,
	CategorySports,
	CategoryDaily,
	CategoryOthers,
}

// ValidConditions 有效的商品成色
var ValidConditions = []string{
	ConditionNew,
	ConditionLikeNew,
	ConditionGood,
	ConditionFair,
}

// ValidItemStatuses 有效的商品状态
var ValidItemStatuses = []string{
	ItemStatusOnSale,
	ItemStatusSold,
	ItemStatusOffShelf,
}

// IsValidCategory 检查分类是否有效
func (m *MarketItem) IsValidCategory() bool {
	for _, c := range ValidCategories {
		if m.Category == c {
			return true
		}
	}
	return false
}

// IsValidCondition 检查成色是否有效
func (m *MarketItem) IsValidCondition() bool {
	for _, c := range ValidConditions {
		if m.Condition == c {
			return true
		}
	}
	return false
}

// IsValidStatus 检查状态是否有效
func (m *MarketItem) IsValidStatus() bool {
	for _, s := range ValidItemStatuses {
		if m.Status == s {
			return true
		}
	}
	return false
}

// CanEdit 检查是否可以编辑
func (m *MarketItem) CanEdit() bool {
	return m.Status == ItemStatusOnSale
}

// CanDelete 检查是否可以删除
func (m *MarketItem) CanDelete() bool {
	return m.Status != ItemStatusSold
}

// IsOwner 检查是否是商品发布者
func (m *MarketItem) IsOwner(userID uint) bool {
	return m.UserID == userID
}

// BeforeCreate GORM钩子 - 创建前验证
func (m *MarketItem) BeforeCreate(tx *gorm.DB) error {
	if !m.IsValidCategory() {
		m.Category = CategoryOthers
	}
	if !m.IsValidCondition() {
		m.Condition = ConditionGood
	}
	if !m.IsValidStatus() {
		m.Status = ItemStatusOnSale
	}
	return nil
}

// MarketItemListQuery 商品列表查询参数
type MarketItemListQuery struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Category  string `form:"category" binding:"omitempty,oneof=electronics books clothes sports daily others ''"`
	Status    string `form:"status" binding:"omitempty,oneof=on_sale sold off_shelf ''"`
	Keyword   string `form:"keyword" binding:"omitempty,max=50"`
	MinPrice  float64 `form:"min_price" binding:"omitempty,min=0"`
	MaxPrice  float64 `form:"max_price" binding:"omitempty,min=0"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=price price_desc time time_desc ''"`
}

// SetDefaults 设置默认值
func (q *MarketItemListQuery) SetDefaults() {
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
func (q *MarketItemListQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// CategoryNames 分类名称映射
var CategoryNames = map[string]string{
	CategoryElectronics: "电子产品",
	CategoryBooks:       "书籍",
	CategoryClothes:     "服饰",
	CategorySports:      "运动用品",
	CategoryDaily:       "日用品",
	CategoryOthers:      "其他",
}

// ConditionNames 成色名称映射
var ConditionNames = map[string]string{
	ConditionNew:     "全新",
	ConditionLikeNew: "几乎全新",
	ConditionGood:    "良好",
	ConditionFair:    "一般",
}

// StatusNames 状态名称映射
var StatusNames = map[string]string{
	ItemStatusOnSale:   "在售中",
	ItemStatusSold:     "已售出",
	ItemStatusOffShelf: "已下架",
}

// GetCategoryName 获取分类名称
func (m *MarketItem) GetCategoryName() string {
	if name, ok := CategoryNames[m.Category]; ok {
		return name
	}
	return "未知分类"
}

// GetConditionName 获取成色名称
func (m *MarketItem) GetConditionName() string {
	if name, ok := ConditionNames[m.Condition]; ok {
		return name
	}
	return "未知成色"
}

// GetStatusName 获取状态名称
func (m *MarketItem) GetStatusName() string {
	if name, ok := StatusNames[m.Status]; ok {
		return name
	}
	return "未知状态"
}