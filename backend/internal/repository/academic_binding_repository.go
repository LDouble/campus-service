package repository

import (
	"context"

	"gorm.io/gorm"

	"campus-service/internal/model"
)

// AcademicBindingRepository 教务绑定仓储接口
type AcademicBindingRepository interface {
	Create(ctx context.Context, binding *model.AcademicBinding) error
	Update(ctx context.Context, binding *model.AcademicBinding) error
	FindByID(ctx context.Context, id uint) (*model.AcademicBinding, error)
	FindByUserID(ctx context.Context, userID uint) (*model.AcademicBinding, error)
	FindByStudentID(ctx context.Context, studentID string) (*model.AcademicBinding, error)
	Delete(ctx context.Context, id uint) error
}

// academicBindingRepository 教务绑定仓储实现
type academicBindingRepository struct {
	db *gorm.DB
}

// NewAcademicBindingRepository 创建教务绑定仓储实例
func NewAcademicBindingRepository(db *gorm.DB) AcademicBindingRepository {
	return &academicBindingRepository{db: db}
}

// Create 创建教务绑定
func (r *academicBindingRepository) Create(ctx context.Context, binding *model.AcademicBinding) error {
	return r.db.WithContext(ctx).Create(binding).Error
}

// Update 更新教务绑定
func (r *academicBindingRepository) Update(ctx context.Context, binding *model.AcademicBinding) error {
	return r.db.WithContext(ctx).Save(binding).Error
}

// FindByID 根据ID查找教务绑定
func (r *academicBindingRepository) FindByID(ctx context.Context, id uint) (*model.AcademicBinding, error) {
	var binding model.AcademicBinding
	err := r.db.WithContext(ctx).Preload("User").First(&binding, id).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// FindByUserID 根据用户ID查找教务绑定
func (r *academicBindingRepository) FindByUserID(ctx context.Context, userID uint) (*model.AcademicBinding, error) {
	var binding model.AcademicBinding
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// FindByStudentID 根据学号查找教务绑定
func (r *academicBindingRepository) FindByStudentID(ctx context.Context, studentID string) (*model.AcademicBinding, error) {
	var binding model.AcademicBinding
	err := r.db.WithContext(ctx).Preload("User").Where("student_id = ?", studentID).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// Delete 删除教务绑定（软删除）
func (r *academicBindingRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.AcademicBinding{}, id).Error
}