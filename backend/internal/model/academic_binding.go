package model

import (
	"time"

	"gorm.io/gorm"
)

// AcademicBinding 教务绑定模型
type AcademicBinding struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	StudentID    string         `gorm:"size:32;not null" json:"student_id"`
	Password     string         `gorm:"size:256;not null" json:"-"` // 加密存储，不返回给前端
	School       string         `gorm:"size:64;not null" json:"school"`
	Status       int            `gorm:"default:1;index" json:"status"`       // 1:正常 0:禁用
	VerifyStatus int            `gorm:"default:0;index" json:"verify_status"` // 0:未验证 1:已验证 2:验证失败
	LastVerifyAt *time.Time     `json:"last_verify_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (AcademicBinding) TableName() string {
	return "academic_bindings"
}

// 绑定状态常量
const (
	BindingStatusDisabled = 0
	BindingStatusActive   = 1
)

// 验证状态常量
const (
	VerifyStatusPending = 0
	VerifyStatusSuccess = 1
	VerifyStatusFailed  = 2
)

// IsActive 检查绑定是否激活
func (ab *AcademicBinding) IsActive() bool {
	return ab.Status == BindingStatusActive
}

// IsVerified 检查是否已验证
func (ab *AcademicBinding) IsVerified() bool {
	return ab.VerifyStatus == VerifyStatusSuccess
}

// UpdateVerifyStatus 更新验证状态
func (ab *AcademicBinding) UpdateVerifyStatus(status int) {
	ab.VerifyStatus = status
	now := time.Now()
	ab.LastVerifyAt = &now
}