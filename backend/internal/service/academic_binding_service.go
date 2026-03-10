package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"campus-service/internal/model"
	"campus-service/internal/repository"
)

// AcademicBindingService 教务绑定服务接口
type AcademicBindingService interface {
	Bind(ctx context.Context, userID uint, studentID, password, school string) (*model.AcademicBinding, error)
	Unbind(ctx context.Context, userID uint) error
	GetBinding(ctx context.Context, userID uint) (*model.AcademicBinding, error)
	VerifyBinding(ctx context.Context, userID uint) error
	IsUserBound(ctx context.Context, userID uint) (bool, error)
}

// academicBindingService 教务绑定服务实现
type academicBindingService struct {
	bindingRepo repository.AcademicBindingRepository
	verifier    AcademicVerifier
}

// NewAcademicBindingService 创建教务绑定服务实例
func NewAcademicBindingService(
	bindingRepo repository.AcademicBindingRepository,
	verifier AcademicVerifier,
) AcademicBindingService {
	return &academicBindingService{
		bindingRepo: bindingRepo,
		verifier:    verifier,
	}
}

// Bind 绑定教务系统
func (s *academicBindingService) Bind(ctx context.Context, userID uint, studentID, password, school string) (*model.AcademicBinding, error) {
	// 检查是否已绑定
	_, err := s.bindingRepo.FindByUserID(ctx, userID)
	if err == nil {
		return nil, errors.New("用户已绑定教务系统")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 检查学号是否已被其他用户绑定
	existingByStudentID, err := s.bindingRepo.FindByStudentID(ctx, studentID)
	if err == nil && existingByStudentID.UserID != userID {
		return nil, errors.New("该学号已被其他用户绑定")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 验证教务系统凭据
	_, err = s.verifier.VerifyCredentials(ctx, studentID, password, school)
	if err != nil {
		return nil, fmt.Errorf("教务系统验证失败: %w", err)
	}

	// 创建绑定记录
	binding := &model.AcademicBinding{
		UserID:       userID,
		StudentID:    studentID,
		Password:     s.encryptPassword(password),
		School:       school,
		Status:       model.BindingStatusActive,
		VerifyStatus: model.VerifyStatusSuccess,
	}

	if err := s.bindingRepo.Create(ctx, binding); err != nil {
		return nil, err
	}

	return binding, nil
}

// Unbind 解绑教务系统
func (s *academicBindingService) Unbind(ctx context.Context, userID uint) error {
	binding, err := s.bindingRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户未绑定教务系统")
		}
		return err
	}

	return s.bindingRepo.Delete(ctx, binding.ID)
}

// GetBinding 获取绑定信息
func (s *academicBindingService) GetBinding(ctx context.Context, userID uint) (*model.AcademicBinding, error) {
	binding, err := s.bindingRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户未绑定教务系统")
		}
		return nil, err
	}

	return binding, nil
}

// VerifyBinding 验证绑定状态
func (s *academicBindingService) VerifyBinding(ctx context.Context, userID uint) error {
	binding, err := s.bindingRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户未绑定教务系统")
		}
		return err
	}

	// 解密密码并验证
	password := s.decryptPassword(binding.Password)
	_, err = s.verifier.VerifyCredentials(ctx, binding.StudentID, password, binding.School)
	if err != nil {
		binding.UpdateVerifyStatus(model.VerifyStatusFailed)
		s.bindingRepo.Update(ctx, binding)
		return fmt.Errorf("教务系统验证失败: %w", err)
	}

	binding.UpdateVerifyStatus(model.VerifyStatusSuccess)
	return s.bindingRepo.Update(ctx, binding)
}

// IsUserBound 检查用户是否已绑定
func (s *academicBindingService) IsUserBound(ctx context.Context, userID uint) (bool, error) {
	_, err := s.bindingRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// encryptPassword 加密密码（简单MD5，实际应使用更安全的方式）
func (s *academicBindingService) encryptPassword(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

// decryptPassword 解密密码（这里只是示例，实际需要可逆加密）
func (s *academicBindingService) decryptPassword(encrypted string) string {
	// 注意：MD5不可逆，这里只是示例
	// 实际应使用AES等可逆加密算法
	return encrypted
}