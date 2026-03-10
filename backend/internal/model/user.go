package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OpenID    string         `gorm:"uniqueIndex;size:64;not null" json:"openid"`
	UnionID   string         `gorm:"size:64" json:"unionid"`
	Nickname  string         `gorm:"size:64" json:"nickname"`
	Avatar    string         `gorm:"size:256" json:"avatar"`
	Phone     string         `gorm:"size:20" json:"phone"`
	Status    int            `gorm:"default:1;index" json:"status"` // 1:正常 0:禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserStatus 用户状态常量
const (
	UserStatusDisabled = 0
	UserStatusActive   = 1
)

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// UpdateProfile 更新用户资料
func (u *User) UpdateProfile(nickname, avatar, phone string) {
	if nickname != "" {
		u.Nickname = nickname
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	if phone != "" {
		u.Phone = phone
	}
}