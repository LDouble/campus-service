package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"campus-service/internal/model"
	"campus-service/internal/pkg/logger"
	"campus-service/internal/repository"
)

// ErrandService 跑腿服务接口
type ErrandService interface {
	// 任务管理
	PublishErrand(ctx context.Context, userID uint, req *PublishErrandRequest) (*model.Errand, error)
	GetErrandList(ctx context.Context, query *model.ErrandListQuery) (*ErrandListResponse, error)
	GetErrandDetail(ctx context.Context, id uint) (*ErrandDetailResponse, error)
	UpdateErrand(ctx context.Context, userID, errandID uint, req *UpdateErrandRequest) (*model.Errand, error)
	DeleteErrand(ctx context.Context, userID, errandID uint) error

	// 任务操作
	AcceptErrand(ctx context.Context, errandID, userID uint, remark string) (*model.ErrandOrder, error)
	CancelErrand(ctx context.Context, errandID, userID uint, reason string) error
	CompleteErrand(ctx context.Context, errandID, userID uint) error

	// 我的任务
	GetMyErrands(ctx context.Context, userID uint, page, pageSize int) (*ErrandListResponse, error)
	GetMyOrders(ctx context.Context, userID uint, page, pageSize int) (*OrderListResponse, error)
}

// PublishErrandRequest 发布跑腿任务请求
type PublishErrandRequest struct {
	Title       string     `json:"title" binding:"required,max=100"`
	Type        string     `json:"type" binding:"required,oneof=express food print other"`
	Description string     `json:"description" binding:"max=500"`
	Reward      float64    `json:"reward" binding:"min=0"`
	Location    string     `json:"location" binding:"max=200"`
	Deadline    *time.Time `json:"deadline"`
}

// UpdateErrandRequest 更新跑腿任务请求
type UpdateErrandRequest struct {
	Title       string     `json:"title" binding:"omitempty,max=100"`
	Description string     `json:"description" binding:"omitempty,max=500"`
	Reward      float64    `json:"reward" binding:"omitempty,min=0"`
	Location    string     `json:"location" binding:"omitempty,max=200"`
	Deadline    *time.Time `json:"deadline"`
}

// ErrandListResponse 跑腿任务列表响应
type ErrandListResponse struct {
	List     []model.Errand `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ErrandDetailResponse 跑腿任务详情响应
type ErrandDetailResponse struct {
	model.Errand
	Order *model.ErrandOrder `json:"order,omitempty"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List     []model.ErrandOrder `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// errandService 跑腿服务实现
type errandService struct {
	errandRepo repository.ErrandRepository
	userRepo   repository.UserRepository
}

// NewErrandService 创建跑腿服务实例
func NewErrandService(errandRepo repository.ErrandRepository, userRepo repository.UserRepository) ErrandService {
	return &errandService{
		errandRepo: errandRepo,
		userRepo:   userRepo,
	}
}

// PublishErrand 发布跑腿任务
func (s *errandService) PublishErrand(ctx context.Context, userID uint, req *PublishErrandRequest) (*model.Errand, error) {
	// 验证截止时间
	if req.Deadline != nil && req.Deadline.Before(time.Now()) {
		return nil, errors.New("截止时间不能早于当前时间")
	}

	// 创建跑腿任务
	errand := &model.Errand{
		UserID:      userID,
		Title:       req.Title,
		Type:        req.Type,
		Description: req.Description,
		Reward:      req.Reward,
		Location:    req.Location,
		Deadline:    req.Deadline,
		Status:      model.ErrandStatusPending,
	}

	if err := s.errandRepo.Create(ctx, errand); err != nil {
		logger.Error("创建跑腿任务失败", zap.Error(err))
		return nil, errors.New("创建跑腿任务失败")
	}

	logger.Info("跑腿任务发布成功", zap.Uint("errand_id", errand.ID), zap.Uint("user_id", userID))
	return errand, nil
}

// GetErrandList 获取跑腿任务列表
func (s *errandService) GetErrandList(ctx context.Context, query *model.ErrandListQuery) (*ErrandListResponse, error) {
	query.SetDefaults()

	errands, total, err := s.errandRepo.List(ctx, query)
	if err != nil {
		logger.Error("获取跑腿任务列表失败", zap.Error(err))
		return nil, errors.New("获取跑腿任务列表失败")
	}

	return &ErrandListResponse{
		List:     errands,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetErrandDetail 获取跑腿任务详情
func (s *errandService) GetErrandDetail(ctx context.Context, id uint) (*ErrandDetailResponse, error) {
	errand, err := s.errandRepo.FindByIDWithUser(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("跑腿任务不存在")
		}
		logger.Error("获取跑腿任务详情失败", zap.Error(err))
		return nil, errors.New("获取跑腿任务详情失败")
	}

	// 如果任务已被接单，获取订单信息
	var order *model.ErrandOrder
	if errand.Status == model.ErrandStatusInProgress || errand.Status == model.ErrandStatusCompleted {
		order, _ = s.errandRepo.FindOrderByErrandID(ctx, id)
	}

	return &ErrandDetailResponse{
		Errand: *errand,
		Order:  order,
	}, nil
}

// UpdateErrand 更新跑腿任务
func (s *errandService) UpdateErrand(ctx context.Context, userID, errandID uint, req *UpdateErrandRequest) (*model.Errand, error) {
	errand, err := s.errandRepo.FindByID(ctx, errandID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("跑腿任务不存在")
		}
		return nil, errors.New("获取跑腿任务失败")
	}

	// 验证权限
	if !errand.IsOwner(userID) {
		return nil, errors.New("无权修改此任务")
	}

	// 只有待接单状态可以修改
	if errand.Status != model.ErrandStatusPending {
		return nil, errors.New("任务状态不允许修改")
	}

	// 更新字段
	if req.Title != "" {
		errand.Title = req.Title
	}
	if req.Description != "" {
		errand.Description = req.Description
	}
	if req.Reward > 0 {
		errand.Reward = req.Reward
	}
	if req.Location != "" {
		errand.Location = req.Location
	}
	if req.Deadline != nil {
		if req.Deadline.Before(time.Now()) {
			return nil, errors.New("截止时间不能早于当前时间")
		}
		errand.Deadline = req.Deadline
	}

	if err := s.errandRepo.Update(ctx, errand); err != nil {
		logger.Error("更新跑腿任务失败", zap.Error(err))
		return nil, errors.New("更新跑腿任务失败")
	}

	return errand, nil
}

// DeleteErrand 删除跑腿任务
func (s *errandService) DeleteErrand(ctx context.Context, userID, errandID uint) error {
	errand, err := s.errandRepo.FindByID(ctx, errandID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("跑腿任务不存在")
		}
		return errors.New("获取跑腿任务失败")
	}

	// 验证权限
	if !errand.IsOwner(userID) {
		return errors.New("无权删除此任务")
	}

	// 只有待接单状态可以删除
	if errand.Status != model.ErrandStatusPending {
		return errors.New("任务状态不允许删除")
	}

	// 软删除
	if err := s.errandRepo.UpdateStatus(ctx, errandID, model.ErrandStatusCancelled); err != nil {
		logger.Error("删除跑腿任务失败", zap.Error(err))
		return errors.New("删除跑腿任务失败")
	}

	return nil
}

// AcceptErrand 接受跑腿任务
func (s *errandService) AcceptErrand(ctx context.Context, errandID, userID uint, remark string) (*model.ErrandOrder, error) {
	// 获取跑腿任务
	errand, err := s.errandRepo.FindByID(ctx, errandID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("跑腿任务不存在")
		}
		return nil, errors.New("获取跑腿任务失败")
	}

	// 检查是否可以接单
	if !errand.CanAccept() {
		return nil, errors.New("该任务当前无法接单")
	}

	// 不能接自己的任务
	if errand.IsOwner(userID) {
		return nil, errors.New("不能接自己发布的任务")
	}

	// 创建订单
	order := &model.ErrandOrder{
		ErrandID: errandID,
		UserID:   userID,
		Status:   model.OrderStatusAccepted,
		Remark:   remark,
	}

	// 使用事务
	txErr := s.errandRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 创建订单
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 更新任务状态
		if err := tx.Model(&model.Errand{}).Where("id = ?", errandID).
			Update("status", model.ErrandStatusInProgress).Error; err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		logger.Error("接单失败", zap.Error(txErr))
		return nil, errors.New("接单失败，请稍后重试")
	}

	logger.Info("接单成功", zap.Uint("errand_id", errandID), zap.Uint("user_id", userID))
	return order, nil
}

// CancelErrand 取消跑腿任务
func (s *errandService) CancelErrand(ctx context.Context, errandID, userID uint, reason string) error {
	errand, err := s.errandRepo.FindByID(ctx, errandID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("跑腿任务不存在")
		}
		return errors.New("获取跑腿任务失败")
	}

	// 检查是否可以取消
	if !errand.CanCancel() {
		return errors.New("该任务当前无法取消")
	}

	// 验证权限：发布者或接单者可以取消
	if errand.IsOwner(userID) {
		// 发布者取消
		return s.cancelByOwner(ctx, errand)
	}

	// 检查是否是接单者
	order, err := s.errandRepo.FindOrderByErrandID(ctx, errandID)
	if err != nil {
		return errors.New("您不是该任务的参与者")
	}

	if order.UserID != userID {
		return errors.New("无权取消此任务")
	}

	// 接单者取消
	return s.cancelByRunner(ctx, errand, order)
}

// cancelByOwner 发布者取消任务
func (s *errandService) cancelByOwner(ctx context.Context, errand *model.Errand) error {
	txErr := s.errandRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 更新任务状态
		if err := tx.Model(&model.Errand{}).Where("id = ?", errand.ID).
			Update("status", model.ErrandStatusCancelled).Error; err != nil {
			return err
		}

		// 如果有接单者，取消订单
		if errand.Status == model.ErrandStatusInProgress {
			if err := tx.Model(&model.ErrandOrder{}).
				Where("errand_id = ? AND status = ?", errand.ID, model.OrderStatusAccepted).
				Update("status", model.OrderStatusCancelled).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if txErr != nil {
		logger.Error("取消任务失败", zap.Error(txErr))
		return errors.New("取消任务失败")
	}

	logger.Info("任务已取消", zap.Uint("errand_id", errand.ID))
	return nil
}

// cancelByRunner 接单者取消任务
func (s *errandService) cancelByRunner(ctx context.Context, errand *model.Errand, order *model.ErrandOrder) error {
	txErr := s.errandRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 取消订单
		if err := tx.Model(order).Update("status", model.OrderStatusCancelled).Error; err != nil {
			return err
		}

		// 恢复任务状态为待接单
		if err := tx.Model(&model.Errand{}).Where("id = ?", errand.ID).
			Update("status", model.ErrandStatusPending).Error; err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		logger.Error("取消接单失败", zap.Error(txErr))
		return errors.New("取消接单失败")
	}

	logger.Info("接单已取消", zap.Uint("errand_id", errand.ID), zap.Uint("order_id", order.ID))
	return nil
}

// CompleteErrand 完成跑腿任务
func (s *errandService) CompleteErrand(ctx context.Context, errandID, userID uint) error {
	errand, err := s.errandRepo.FindByID(ctx, errandID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("跑腿任务不存在")
		}
		return errors.New("获取跑腿任务失败")
	}

	// 检查是否可以完成
	if !errand.CanComplete() {
		return errors.New("该任务当前无法完成")
	}

	// 只有发布者可以确认完成
	if !errand.IsOwner(userID) {
		return errors.New("只有发布者可以确认完成")
	}

	// 获取订单
	order, err := s.errandRepo.FindOrderByErrandID(ctx, errandID)
	if err != nil {
		return errors.New("获取订单信息失败")
	}

	// 使用事务更新状态
	txErr := s.errandRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 更新任务状态
		if err := tx.Model(&model.Errand{}).Where("id = ?", errandID).
			Update("status", model.ErrandStatusCompleted).Error; err != nil {
			return err
		}

		// 更新订单状态
		now := time.Now()
		if err := tx.Model(&model.ErrandOrder{}).Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"status":       model.OrderStatusCompleted,
				"completed_at": &now,
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		logger.Error("完成任务失败", zap.Error(txErr))
		return errors.New("完成任务失败")
	}

	logger.Info("任务已完成", zap.Uint("errand_id", errandID))
	return nil
}

// GetMyErrands 获取我发布的跑腿任务
func (s *errandService) GetMyErrands(ctx context.Context, userID uint, page, pageSize int) (*ErrandListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	errands, total, err := s.errandRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取我的跑腿任务失败", zap.Error(err))
		return nil, errors.New("获取我的跑腿任务失败")
	}

	return &ErrandListResponse{
		List:     errands,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetMyOrders 获取我接的跑腿订单
func (s *errandService) GetMyOrders(ctx context.Context, userID uint, page, pageSize int) (*OrderListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	orders, total, err := s.errandRepo.FindAcceptedOrderByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取我的订单失败", zap.Error(err))
		return nil, errors.New("获取我的订单失败")
	}

	return &OrderListResponse{
		List:     orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ValidateErrandType 验证跑腿类型
func ValidateErrandType(t string) bool {
	for _, valid := range model.ValidErrandTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// FormatErrandType 格式化跑腿类型显示名称
func FormatErrandType(t string) string {
	switch t {
	case model.ErrandTypeExpress:
		return "快递代取"
	case model.ErrandTypeFood:
		return "外卖代拿"
	case model.ErrandTypePrint:
		return "打印复印"
	case model.ErrandTypeOther:
		return "其他"
	default:
		return "未知类型"
	}
}

// FormatErrandStatus 格式化跑腿状态显示名称
func FormatErrandStatus(s string) string {
	switch s {
	case model.ErrandStatusPending:
		return "待接单"
	case model.ErrandStatusInProgress:
		return "进行中"
	case model.ErrandStatusCompleted:
		return "已完成"
	case model.ErrandStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}