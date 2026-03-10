package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"campus-service/internal/model"
	"campus-service/internal/pkg/logger"
	"campus-service/internal/repository"
)

// LostFoundService 失物招领服务接口
type LostFoundService interface {
	// 失物招领管理
	Publish(ctx context.Context, userID uint, req *PublishLostFoundRequest) (*model.LostFound, error)
	GetList(ctx context.Context, query *model.LostFoundListQuery) (*LostFoundListResponse, error)
	GetDetail(ctx context.Context, id, userID uint) (*LostFoundDetailResponse, error)
	Update(ctx context.Context, userID, lfID uint, req *UpdateLostFoundRequest) (*model.LostFound, error)
	Delete(ctx context.Context, userID, lfID uint) error
	Close(ctx context.Context, userID, lfID uint) error
	Resolve(ctx context.Context, userID, lfID uint) error

	// 认领管理
	SubmitClaim(ctx context.Context, lfID, userID uint, req *SubmitClaimRequest) (*model.LostFoundClaim, error)
	GetClaims(ctx context.Context, lfID uint, page, pageSize int) (*ClaimListResponse, error)
	GetMyClaims(ctx context.Context, userID uint, page, pageSize int) (*ClaimListResponse, error)
	ApproveClaim(ctx context.Context, claimID, userID uint) error
	RejectClaim(ctx context.Context, claimID, userID uint) error

	// 我的发布
	GetMyPublish(ctx context.Context, userID uint, page, pageSize int) (*LostFoundListResponse, error)
}

// PublishLostFoundRequest 发布失物招领请求
type PublishLostFoundRequest struct {
	Type        string     `json:"type" binding:"required,oneof=lost found"`
	Title       string     `json:"title" binding:"required,max=100"`
	Description string     `json:"description" binding:"max=1000"`
	Category    string     `json:"category" binding:"required,oneof=electronics cards books clothes bags others"`
	Images      []string   `json:"images" binding:"max=9"`
	Location    string     `json:"location" binding:"max=200"`
	LostTime    *time.Time `json:"lost_time"`
	Contact     string     `json:"contact" binding:"required,max=50"`
}

// UpdateLostFoundRequest 更新失物招领请求
type UpdateLostFoundRequest struct {
	Title       string     `json:"title" binding:"omitempty,max=100"`
	Description string     `json:"description" binding:"omitempty,max=1000"`
	Category    string     `json:"category" binding:"omitempty,oneof=electronics cards books clothes bags others"`
	Images      []string   `json:"images" binding:"omitempty,max=9"`
	Location    string     `json:"location" binding:"omitempty,max=200"`
	LostTime    *time.Time `json:"lost_time"`
	Contact     string     `json:"contact" binding:"omitempty,max=50"`
}

// SubmitClaimRequest 提交认领请求
type SubmitClaimRequest struct {
	Message string   `json:"message" binding:"required,max=500"`
	Proof   []string `json:"proof" binding:"max=9"`
}

// LostFoundListResponse 失物招领列表响应
type LostFoundListResponse struct {
	List     []model.LostFound `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// LostFoundDetailResponse 失物招领详情响应
type LostFoundDetailResponse struct {
	model.LostFound
	HasClaimed bool `json:"has_claimed"`
}

// ClaimListResponse 认领列表响应
type ClaimListResponse struct {
	List     []model.LostFoundClaim `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// lostFoundService 失物招领服务实现
type lostFoundService struct {
	lfRepo   repository.LostFoundRepository
	userRepo repository.UserRepository
}

// NewLostFoundService 创建失物招领服务实例
func NewLostFoundService(lfRepo repository.LostFoundRepository, userRepo repository.UserRepository) LostFoundService {
	return &lostFoundService{
		lfRepo:   lfRepo,
		userRepo: userRepo,
	}
}

// Publish 发布失物招领
func (s *lostFoundService) Publish(ctx context.Context, userID uint, req *PublishLostFoundRequest) (*model.LostFound, error) {
	// 序列化图片
	var imagesJSON string
	if len(req.Images) > 0 {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return nil, errors.New("图片数据格式错误")
		}
		imagesJSON = string(imagesBytes)
	}

	// 创建记录
	lf := &model.LostFound{
		UserID:      userID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Images:      imagesJSON,
		Location:    req.Location,
		LostTime:    req.LostTime,
		Contact:     req.Contact,
		Status:      model.LFStatusOpen,
	}

	if err := s.lfRepo.Create(ctx, lf); err != nil {
		logger.Error("创建失物招领失败", zap.Error(err))
		return nil, errors.New("创建失物招领失败")
	}

	logger.Info("失物招领发布成功", zap.Uint("id", lf.ID), zap.Uint("user_id", userID))
	return lf, nil
}

// GetList 获取失物招领列表
func (s *lostFoundService) GetList(ctx context.Context, query *model.LostFoundListQuery) (*LostFoundListResponse, error) {
	query.SetDefaults()

	items, total, err := s.lfRepo.List(ctx, query)
	if err != nil {
		logger.Error("获取失物招领列表失败", zap.Error(err))
		return nil, errors.New("获取失物招领列表失败")
	}

	return &LostFoundListResponse{
		List:     items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetDetail 获取失物招领详情
func (s *lostFoundService) GetDetail(ctx context.Context, id, userID uint) (*LostFoundDetailResponse, error) {
	lf, err := s.lfRepo.FindByIDWithUser(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("记录不存在")
		}
		logger.Error("获取失物招领详情失败", zap.Error(err))
		return nil, errors.New("获取失物招领详情失败")
	}

	// 增加浏览次数
	go func() {
		_ = s.lfRepo.IncrementViewCount(context.Background(), id)
	}()

	// 检查是否已认领
	hasClaimed := false
	if userID > 0 {
		hasClaimed, _ = s.lfRepo.HasPendingClaim(ctx, id, userID)
	}

	return &LostFoundDetailResponse{
		LostFound:  *lf,
		HasClaimed: hasClaimed,
	}, nil
}

// Update 更新失物招领
func (s *lostFoundService) Update(ctx context.Context, userID, lfID uint, req *UpdateLostFoundRequest) (*model.LostFound, error) {
	lf, err := s.lfRepo.FindByID(ctx, lfID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("记录不存在")
		}
		return nil, errors.New("获取记录失败")
	}

	// 验证权限
	if !lf.IsOwner(userID) {
		return nil, errors.New("无权修改此记录")
	}

	// 只有进行中状态可以修改
	if !lf.CanEdit() {
		return nil, errors.New("记录状态不允许修改")
	}

	// 更新字段
	if req.Title != "" {
		lf.Title = req.Title
	}
	if req.Description != "" {
		lf.Description = req.Description
	}
	if req.Category != "" {
		lf.Category = req.Category
	}
	if len(req.Images) > 0 {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return nil, errors.New("图片数据格式错误")
		}
		lf.Images = string(imagesBytes)
	}
	if req.Location != "" {
		lf.Location = req.Location
	}
	if req.LostTime != nil {
		lf.LostTime = req.LostTime
	}
	if req.Contact != "" {
		lf.Contact = req.Contact
	}

	lf.UpdatedAt = time.Now()

	if err := s.lfRepo.Update(ctx, lf); err != nil {
		logger.Error("更新失物招领失败", zap.Error(err))
		return nil, errors.New("更新失物招领失败")
	}

	return lf, nil
}

// Delete 删除失物招领
func (s *lostFoundService) Delete(ctx context.Context, userID, lfID uint) error {
	lf, err := s.lfRepo.FindByID(ctx, lfID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("记录不存在")
		}
		return errors.New("获取记录失败")
	}

	// 验证权限
	if !lf.IsOwner(userID) {
		return errors.New("无权删除此记录")
	}

	// 已解决的记录不能删除
	if !lf.CanDelete() {
		return errors.New("已解决的记录无法删除")
	}

	// 硬删除
	if err := s.lfRepo.DB().Delete(lf).Error; err != nil {
		logger.Error("删除失物招领失败", zap.Error(err))
		return errors.New("删除失物招领失败")
	}

	return nil
}

// Close 关闭失物招领
func (s *lostFoundService) Close(ctx context.Context, userID, lfID uint) error {
	lf, err := s.lfRepo.FindByID(ctx, lfID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("记录不存在")
		}
		return errors.New("获取记录失败")
	}

	// 验证权限
	if !lf.IsOwner(userID) {
		return errors.New("无权操作此记录")
	}

	// 只有进行中状态可以关闭
	if lf.Status != model.LFStatusOpen {
		return errors.New("记录状态不允许此操作")
	}

	if err := s.lfRepo.UpdateStatus(ctx, lfID, model.LFStatusClosed); err != nil {
		logger.Error("关闭失物招领失败", zap.Error(err))
		return errors.New("关闭失物招领失败")
	}

	return nil
}

// Resolve 标记为已解决
func (s *lostFoundService) Resolve(ctx context.Context, userID, lfID uint) error {
	lf, err := s.lfRepo.FindByID(ctx, lfID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("记录不存在")
		}
		return errors.New("获取记录失败")
	}

	// 验证权限
	if !lf.IsOwner(userID) {
		return errors.New("无权操作此记录")
	}

	// 只有进行中或已关闭状态可以标记为已解决
	if lf.Status == model.LFStatusResolved {
		return errors.New("记录已解决")
	}

	if err := s.lfRepo.UpdateStatus(ctx, lfID, model.LFStatusResolved); err != nil {
		logger.Error("标记已解决失败", zap.Error(err))
		return errors.New("标记已解决失败")
	}

	return nil
}

// SubmitClaim 提交认领
func (s *lostFoundService) SubmitClaim(ctx context.Context, lfID, userID uint, req *SubmitClaimRequest) (*model.LostFoundClaim, error) {
	// 检查失物招领是否存在
	lf, err := s.lfRepo.FindByID(ctx, lfID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("记录不存在")
		}
		return nil, errors.New("获取记录失败")
	}

	// 只有进行中状态可以认领
	if lf.Status != model.LFStatusOpen {
		return nil, errors.New("该记录已关闭认领")
	}

	// 不能认领自己发布的
	if lf.IsOwner(userID) {
		return nil, errors.New("不能认领自己发布的记录")
	}

	// 检查是否已有待审核的认领
	hasClaimed, err := s.lfRepo.HasPendingClaim(ctx, lfID, userID)
	if err != nil {
		return nil, errors.New("检查认领状态失败")
	}
	if hasClaimed {
		return nil, errors.New("您已提交过认领申请，请等待审核")
	}

	// 序列化证明材料
	var proofJSON string
	if len(req.Proof) > 0 {
		proofBytes, err := json.Marshal(req.Proof)
		if err != nil {
			return nil, errors.New("证明材料格式错误")
		}
		proofJSON = string(proofBytes)
	}

	// 创建认领记录
	claim := &model.LostFoundClaim{
		LostFoundID: lfID,
		UserID:      userID,
		Message:     req.Message,
		Proof:       proofJSON,
		Status:      model.ClaimStatusPending,
	}

	if err := s.lfRepo.CreateClaim(ctx, claim); err != nil {
		logger.Error("提交认领失败", zap.Error(err))
		return nil, errors.New("提交认领失败")
	}

	logger.Info("认领申请提交成功", zap.Uint("claim_id", claim.ID), zap.Uint("lf_id", lfID), zap.Uint("user_id", userID))
	return claim, nil
}

// GetClaims 获取认领列表
func (s *lostFoundService) GetClaims(ctx context.Context, lfID uint, page, pageSize int) (*ClaimListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	claims, total, err := s.lfRepo.FindClaimsByLostFoundID(ctx, lfID, page, pageSize)
	if err != nil {
		logger.Error("获取认领列表失败", zap.Error(err))
		return nil, errors.New("获取认领列表失败")
	}

	return &ClaimListResponse{
		List:     claims,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetMyClaims 获取我的认领
func (s *lostFoundService) GetMyClaims(ctx context.Context, userID uint, page, pageSize int) (*ClaimListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	claims, total, err := s.lfRepo.FindClaimsByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取我的认领失败", zap.Error(err))
		return nil, errors.New("获取我的认领失败")
	}

	return &ClaimListResponse{
		List:     claims,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ApproveClaim 通过认领
func (s *lostFoundService) ApproveClaim(ctx context.Context, claimID, userID uint) error {
	claim, err := s.lfRepo.FindClaimByID(ctx, claimID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("认领记录不存在")
		}
		return errors.New("获取认领记录失败")
	}

	// 验证权限 - 只有发布者可以审核
	if !claim.LostFound.IsOwner(userID) {
		return errors.New("无权审核此认领")
	}

	// 只有待审核状态可以审核
	if claim.Status != model.ClaimStatusPending {
		return errors.New("该认领已处理")
	}

	// 使用事务
	tx := s.lfRepo.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新认领状态
	if err := tx.Model(&model.LostFoundClaim{}).Where("id = ?", claimID).Update("status", model.ClaimStatusApproved).Error; err != nil {
		tx.Rollback()
		logger.Error("更新认领状态失败", zap.Error(err))
		return errors.New("更新认领状态失败")
	}

	// 将失物招领标记为已解决
	if err := tx.Model(&model.LostFound{}).Where("id = ?", claim.LostFoundID).Update("status", model.LFStatusResolved).Error; err != nil {
		tx.Rollback()
		logger.Error("更新失物招领状态失败", zap.Error(err))
		return errors.New("更新失物招领状态失败")
	}

	// 拒绝其他待审核的认领
	if err := tx.Model(&model.LostFoundClaim{}).
		Where("lost_found_id = ? AND id != ? AND status = ?", claim.LostFoundID, claimID, model.ClaimStatusPending).
		Update("status", model.ClaimStatusRejected).Error; err != nil {
		tx.Rollback()
		logger.Error("拒绝其他认领失败", zap.Error(err))
		return errors.New("处理其他认领失败")
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败", zap.Error(err))
		return errors.New("操作失败")
	}

	logger.Info("认领审核通过", zap.Uint("claim_id", claimID), zap.Uint("lf_id", claim.LostFoundID))
	return nil
}

// RejectClaim 拒绝认领
func (s *lostFoundService) RejectClaim(ctx context.Context, claimID, userID uint) error {
	claim, err := s.lfRepo.FindClaimByID(ctx, claimID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("认领记录不存在")
		}
		return errors.New("获取认领记录失败")
	}

	// 验证权限 - 只有发布者可以审核
	if !claim.LostFound.IsOwner(userID) {
		return errors.New("无权审核此认领")
	}

	// 只有待审核状态可以审核
	if claim.Status != model.ClaimStatusPending {
		return errors.New("该认领已处理")
	}

	if err := s.lfRepo.UpdateClaimStatus(ctx, claimID, model.ClaimStatusRejected); err != nil {
		logger.Error("拒绝认领失败", zap.Error(err))
		return errors.New("拒绝认领失败")
	}

	return nil
}

// GetMyPublish 获取我发布的失物招领
func (s *lostFoundService) GetMyPublish(ctx context.Context, userID uint, page, pageSize int) (*LostFoundListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	items, total, err := s.lfRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取我的发布失败", zap.Error(err))
		return nil, errors.New("获取我的发布失败")
	}

	return &LostFoundListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ValidateLostFoundType 验证类型
func ValidateLostFoundType(t string) bool {
	for _, valid := range model.ValidLostFoundTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// ValidateLostFoundCategory 验证分类
func ValidateLostFoundCategory(c string) bool {
	for _, valid := range model.ValidLostFoundCategories {
		if c == valid {
			return true
		}
	}
	return false
}

// FormatLostFoundType 格式化类型名称
func FormatLostFoundType(t string) string {
	if name, ok := model.LostFoundTypeNames[t]; ok {
		return name
	}
	return "未知类型"
}

// FormatLostFoundCategory 格式化分类名称
func FormatLostFoundCategory(c string) string {
	if name, ok := model.LostFoundCategoryNames[c]; ok {
		return name
	}
	return "未知分类"
}

// FormatLostFoundStatus 格式化状态名称
func FormatLostFoundStatus(s string) string {
	if name, ok := model.LostFoundStatusNames[s]; ok {
		return name
	}
	return "未知状态"
}

// FormatClaimStatus 格式化认领状态名称
func FormatClaimStatus(s string) string {
	if name, ok := model.ClaimStatusNames[s]; ok {
		return name
	}
	return "未知状态"
}