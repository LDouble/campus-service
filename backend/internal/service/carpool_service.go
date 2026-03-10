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

// CarpoolService 拼车服务接口
type CarpoolService interface {
	// 行程管理
	PublishCarpool(ctx context.Context, userID uint, req *PublishCarpoolRequest) (*model.Carpool, error)
	GetCarpoolList(ctx context.Context, query *model.CarpoolListQuery) (*CarpoolListResponse, error)
	GetCarpoolDetail(ctx context.Context, id uint) (*CarpoolDetailResponse, error)
	UpdateCarpool(ctx context.Context, userID, carpoolID uint, req *UpdateCarpoolRequest) (*model.Carpool, error)
	DeleteCarpool(ctx context.Context, userID, carpoolID uint) error

	// 行程操作
	BookCarpool(ctx context.Context, carpoolID, userID uint, seats int, remark string) (*model.CarpoolBooking, error)
	CancelBooking(ctx context.Context, bookingID, userID uint) error
	ConfirmBooking(ctx context.Context, bookingID, userID uint) error
	DepartCarpool(ctx context.Context, carpoolID, userID uint) error
	CancelCarpool(ctx context.Context, carpoolID, userID uint, reason string) error

	// 我的行程
	GetMyCarpools(ctx context.Context, userID uint, page, pageSize int) (*CarpoolListResponse, error)
	GetMyBookings(ctx context.Context, userID uint, page, pageSize int) (*BookingListResponse, error)
}

// PublishCarpoolRequest 发布拼车行程请求
type PublishCarpoolRequest struct {
	Origin        string     `json:"origin" binding:"required,max=100"`
	Destination   string     `json:"destination" binding:"required,max=100"`
	DepartureTime *time.Time `json:"departure_time" binding:"required"`
	Seats         int        `json:"seats" binding:"required,min=1,max=10"`
	Price         float64    `json:"price" binding:"min=0"`
	VehicleType   string     `json:"vehicle_type" binding:"omitempty,oneof=car taxi bus"`
	VehicleNumber string     `json:"vehicle_number" binding:"max=20"`
	Contact       string     `json:"contact" binding:"required,max=50"`
	Remark        string     `json:"remark" binding:"max=200"`
}

// UpdateCarpoolRequest 更新拼车行程请求
type UpdateCarpoolRequest struct {
	Origin        string     `json:"origin" binding:"omitempty,max=100"`
	Destination   string     `json:"destination" binding:"omitempty,max=100"`
	DepartureTime *time.Time `json:"departure_time"`
	Seats         int        `json:"seats" binding:"omitempty,min=1,max=10"`
	Price         float64    `json:"price" binding:"omitempty,min=0"`
	VehicleNumber string     `json:"vehicle_number" binding:"omitempty,max=20"`
	Contact       string     `json:"contact" binding:"omitempty,max=50"`
	Remark        string     `json:"remark" binding:"omitempty,max=200"`
}

// CarpoolListResponse 拼车行程列表响应
type CarpoolListResponse struct {
	List     []model.Carpool `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// CarpoolDetailResponse 拼车行程详情响应
type CarpoolDetailResponse struct {
	model.Carpool
	Bookings []model.CarpoolBooking `json:"bookings,omitempty"`
}

// BookingListResponse 预约列表响应
type BookingListResponse struct {
	List     []model.CarpoolBooking `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// carpoolService 拼车服务实现
type carpoolService struct {
	carpoolRepo repository.CarpoolRepository
	userRepo    repository.UserRepository
}

// NewCarpoolService 创建拼车服务实例
func NewCarpoolService(carpoolRepo repository.CarpoolRepository, userRepo repository.UserRepository) CarpoolService {
	return &carpoolService{
		carpoolRepo: carpoolRepo,
		userRepo:    userRepo,
	}
}

// PublishCarpool 发布拼车行程
func (s *carpoolService) PublishCarpool(ctx context.Context, userID uint, req *PublishCarpoolRequest) (*model.Carpool, error) {
	// 验证出发时间
	if req.DepartureTime != nil && req.DepartureTime.Before(time.Now()) {
		return nil, errors.New("出发时间不能早于当前时间")
	}

	// 设置默认车辆类型
	vehicleType := req.VehicleType
	if vehicleType == "" {
		vehicleType = model.VehicleTypeCar
	}

	// 创建拼车行程
	carpool := &model.Carpool{
		UserID:         userID,
		Origin:         req.Origin,
		Destination:    req.Destination,
		DepartureTime:  req.DepartureTime,
		Seats:          req.Seats,
		AvailableSeats: req.Seats,
		Price:          req.Price,
		VehicleType:    vehicleType,
		VehicleNumber:  req.VehicleNumber,
		Contact:        req.Contact,
		Remark:         req.Remark,
		Status:         model.CarpoolStatusPending,
	}

	if err := s.carpoolRepo.Create(ctx, carpool); err != nil {
		logger.Error("创建拼车行程失败", zap.Error(err))
		return nil, errors.New("创建拼车行程失败")
	}

	logger.Info("拼车行程发布成功", zap.Uint("carpool_id", carpool.ID), zap.Uint("user_id", userID))
	return carpool, nil
}

// GetCarpoolList 获取拼车行程列表
func (s *carpoolService) GetCarpoolList(ctx context.Context, query *model.CarpoolListQuery) (*CarpoolListResponse, error) {
	query.SetDefaults()

	carpools, total, err := s.carpoolRepo.List(ctx, query)
	if err != nil {
		logger.Error("获取拼车行程列表失败", zap.Error(err))
		return nil, errors.New("获取拼车行程列表失败")
	}

	return &CarpoolListResponse{
		List:     carpools,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetCarpoolDetail 获取拼车行程详情
func (s *carpoolService) GetCarpoolDetail(ctx context.Context, id uint) (*CarpoolDetailResponse, error) {
	carpool, err := s.carpoolRepo.FindByIDWithUser(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("拼车行程不存在")
		}
		logger.Error("获取拼车行程详情失败", zap.Error(err))
		return nil, errors.New("获取拼车行程详情失败")
	}

	// 获取已确认的预约
	bookings, _ := s.carpoolRepo.FindBookingsByCarpoolID(ctx, id)

	return &CarpoolDetailResponse{
		Carpool:  *carpool,
		Bookings: bookings,
	}, nil
}

// UpdateCarpool 更新拼车行程
func (s *carpoolService) UpdateCarpool(ctx context.Context, userID, carpoolID uint, req *UpdateCarpoolRequest) (*model.Carpool, error) {
	carpool, err := s.carpoolRepo.FindByID(ctx, carpoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("拼车行程不存在")
		}
		return nil, errors.New("获取拼车行程失败")
	}

	// 验证权限
	if !carpool.IsOwner(userID) {
		return nil, errors.New("无权修改此行程")
	}

	// 只有待出发状态可以修改
	if carpool.Status != model.CarpoolStatusPending {
		return nil, errors.New("行程状态不允许修改")
	}

	// 更新字段
	if req.Origin != "" {
		carpool.Origin = req.Origin
	}
	if req.Destination != "" {
		carpool.Destination = req.Destination
	}
	if req.DepartureTime != nil {
		if req.DepartureTime.Before(time.Now()) {
			return nil, errors.New("出发时间不能早于当前时间")
		}
		carpool.DepartureTime = req.DepartureTime
	}
	if req.Seats > 0 {
		// 检查座位数是否足够
		usedSeats := carpool.Seats - carpool.AvailableSeats
		if req.Seats < usedSeats {
			return nil, errors.New("座位数不能少于已预约数量")
		}
		carpool.Seats = req.Seats
		carpool.AvailableSeats = req.Seats - usedSeats
	}
	if req.Price >= 0 {
		carpool.Price = req.Price
	}
	if req.VehicleNumber != "" {
		carpool.VehicleNumber = req.VehicleNumber
	}
	if req.Contact != "" {
		carpool.Contact = req.Contact
	}
	if req.Remark != "" {
		carpool.Remark = req.Remark
	}

	if err := s.carpoolRepo.Update(ctx, carpool); err != nil {
		logger.Error("更新拼车行程失败", zap.Error(err))
		return nil, errors.New("更新拼车行程失败")
	}

	return carpool, nil
}

// DeleteCarpool 删除拼车行程
func (s *carpoolService) DeleteCarpool(ctx context.Context, userID, carpoolID uint) error {
	carpool, err := s.carpoolRepo.FindByID(ctx, carpoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("拼车行程不存在")
		}
		return errors.New("获取拼车行程失败")
	}

	// 验证权限
	if !carpool.IsOwner(userID) {
		return errors.New("无权删除此行程")
	}

	// 只有待出发状态可以删除
	if carpool.Status != model.CarpoolStatusPending {
		return errors.New("行程状态不允许删除")
	}

	// 检查是否有已确认的预约
	bookings, err := s.carpoolRepo.FindBookingsByCarpoolID(ctx, carpoolID)
	if err == nil {
		for _, b := range bookings {
			if b.Status == model.BookingStatusConfirmed {
				return errors.New("存在已确认的预约，无法删除")
			}
		}
	}

	// 软删除（改为取消状态）
	if err := s.carpoolRepo.UpdateStatus(ctx, carpoolID, model.CarpoolStatusCancelled); err != nil {
		logger.Error("删除拼车行程失败", zap.Error(err))
		return errors.New("删除拼车行程失败")
	}

	return nil
}

// BookCarpool 预约拼车
func (s *carpoolService) BookCarpool(ctx context.Context, carpoolID, userID uint, seats int, remark string) (*model.CarpoolBooking, error) {
	// 获取拼车行程
	carpool, err := s.carpoolRepo.FindByID(ctx, carpoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("拼车行程不存在")
		}
		return nil, errors.New("获取拼车行程失败")
	}

	// 检查是否可以预约
	if !carpool.CanBook() {
		return nil, errors.New("该行程当前无法预约")
	}

	// 不能预约自己的行程
	if carpool.IsOwner(userID) {
		return nil, errors.New("不能预约自己发布的行程")
	}

	// 检查座位数
	if !carpool.HasAvailableSeats(seats) {
		return nil, errors.New("座位数不足")
	}

	// 检查是否已经预约过
	_, err = s.carpoolRepo.FindConfirmedBookingByUser(ctx, carpoolID, userID)
	if err == nil {
		return nil, errors.New("您已经预约过该行程")
	}

	// 创建预约
	booking := &model.CarpoolBooking{
		CarpoolID: carpoolID,
		UserID:    userID,
		Seats:     seats,
		Status:    model.BookingStatusConfirmed, // 直接确认
		Remark:    remark,
	}

	// 使用事务
	txErr := s.carpoolRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 创建预约
		if err := tx.Create(booking).Error; err != nil {
			return err
		}

		// 更新可用座位数
		newAvailableSeats := carpool.AvailableSeats - seats
		if err := tx.Model(&model.Carpool{}).Where("id = ?", carpoolID).
			Update("available_seats", newAvailableSeats).Error; err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		logger.Error("预约失败", zap.Error(txErr))
		return nil, errors.New("预约失败，请稍后重试")
	}

	logger.Info("预约成功", zap.Uint("carpool_id", carpoolID), zap.Uint("user_id", userID))
	return booking, nil
}

// CancelBooking 取消预约
func (s *carpoolService) CancelBooking(ctx context.Context, bookingID, userID uint) error {
	booking, err := s.carpoolRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("预约不存在")
		}
		return errors.New("获取预约失败")
	}

	// 验证权限
	if booking.UserID != userID {
		return errors.New("无权取消此预约")
	}

	// 检查状态
	if booking.Status != model.BookingStatusConfirmed && booking.Status != model.BookingStatusPending {
		return errors.New("该预约已无法取消")
	}

	// 获取行程
	carpool, err := s.carpoolRepo.FindByID(ctx, booking.CarpoolID)
	if err != nil {
		return errors.New("获取行程失败")
	}

	// 行程已出发则无法取消
	if carpool.Status == model.CarpoolStatusDeparted {
		return errors.New("行程已出发，无法取消预约")
	}

	// 使用事务
	txErr := s.carpoolRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 更新预约状态
		if err := tx.Model(booking).Update("status", model.BookingStatusCancelled).Error; err != nil {
			return err
		}

		// 恢复座位数
		newAvailableSeats := carpool.AvailableSeats + booking.Seats
		if err := tx.Model(&model.Carpool{}).Where("id = ?", carpool.ID).
			Update("available_seats", newAvailableSeats).Error; err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		logger.Error("取消预约失败", zap.Error(txErr))
		return errors.New("取消预约失败")
	}

	logger.Info("预约已取消", zap.Uint("booking_id", bookingID))
	return nil
}

// ConfirmBooking 确认预约（车主确认）
func (s *carpoolService) ConfirmBooking(ctx context.Context, bookingID, userID uint) error {
	booking, err := s.carpoolRepo.FindBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("预约不存在")
		}
		return errors.New("获取预约失败")
	}

	// 获取行程
	carpool, err := s.carpoolRepo.FindByID(ctx, booking.CarpoolID)
	if err != nil {
		return errors.New("获取行程失败")
	}

	// 验证权限（只有车主可以确认）
	if !carpool.IsOwner(userID) {
		return errors.New("无权确认此预约")
	}

	// 检查状态
	if booking.Status != model.BookingStatusPending {
		return errors.New("该预约无法确认")
	}

	// 更新状态
	booking.Status = model.BookingStatusConfirmed
	if err := s.carpoolRepo.UpdateBooking(ctx, booking); err != nil {
		logger.Error("确认预约失败", zap.Error(err))
		return errors.New("确认预约失败")
	}

	logger.Info("预约已确认", zap.Uint("booking_id", bookingID))
	return nil
}

// DepartCarpool 出发（完成行程）
func (s *carpoolService) DepartCarpool(ctx context.Context, carpoolID, userID uint) error {
	carpool, err := s.carpoolRepo.FindByID(ctx, carpoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("拼车行程不存在")
		}
		return errors.New("获取拼车行程失败")
	}

	// 验证权限
	if !carpool.IsOwner(userID) {
		return errors.New("只有发布者可以确认出发")
	}

	// 检查状态
	if !carpool.CanDepart() {
		return errors.New("该行程当前无法出发")
	}

	// 更新状态
	if err := s.carpoolRepo.UpdateStatus(ctx, carpoolID, model.CarpoolStatusDeparted); err != nil {
		logger.Error("确认出发失败", zap.Error(err))
		return errors.New("确认出发失败")
	}

	logger.Info("行程已出发", zap.Uint("carpool_id", carpoolID))
	return nil
}

// CancelCarpool 取消行程
func (s *carpoolService) CancelCarpool(ctx context.Context, carpoolID, userID uint, reason string) error {
	carpool, err := s.carpoolRepo.FindByID(ctx, carpoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("拼车行程不存在")
		}
		return errors.New("获取拼车行程失败")
	}

	// 验证权限
	if !carpool.IsOwner(userID) {
		return errors.New("只有发布者可以取消行程")
	}

	// 检查状态
	if !carpool.CanCancel() {
		return errors.New("该行程当前无法取消")
	}

	// 使用事务
	txErr := s.carpoolRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 更新行程状态
		if err := tx.Model(&model.Carpool{}).Where("id = ?", carpoolID).
			Update("status", model.CarpoolStatusCancelled).Error; err != nil {
			return err
		}

		// 取消所有预约
		if err := tx.Model(&model.CarpoolBooking{}).
			Where("carpool_id = ? AND status != ?", carpoolID, model.BookingStatusCancelled).
			Update("status", model.BookingStatusCancelled).Error; err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		logger.Error("取消行程失败", zap.Error(txErr))
		return errors.New("取消行程失败")
	}

	logger.Info("行程已取消", zap.Uint("carpool_id", carpoolID))
	return nil
}

// GetMyCarpools 获取我发布的拼车行程
func (s *carpoolService) GetMyCarpools(ctx context.Context, userID uint, page, pageSize int) (*CarpoolListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	carpools, total, err := s.carpoolRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取我的拼车行程失败", zap.Error(err))
		return nil, errors.New("获取我的拼车行程失败")
	}

	return &CarpoolListResponse{
		List:     carpools,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetMyBookings 获取我的拼车预约
func (s *carpoolService) GetMyBookings(ctx context.Context, userID uint, page, pageSize int) (*BookingListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	bookings, total, err := s.carpoolRepo.FindBookingsByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("获取我的预约失败", zap.Error(err))
		return nil, errors.New("获取我的预约失败")
	}

	return &BookingListResponse{
		List:     bookings,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ValidateVehicleType 验证车辆类型
func ValidateVehicleType(t string) bool {
	for _, valid := range model.ValidVehicleTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// FormatVehicleType 格式化车辆类型显示名称
func FormatVehicleType(t string) string {
	switch t {
	case model.VehicleTypeCar:
		return "私家车"
	case model.VehicleTypeTaxi:
		return "出租车"
	case model.VehicleTypeBus:
		return "班车"
	default:
		return "未知类型"
	}
}

// FormatCarpoolStatus 格式化行程状态显示名称
func FormatCarpoolStatus(s string) string {
	switch s {
	case model.CarpoolStatusPending:
		return "待出发"
	case model.CarpoolStatusDeparted:
		return "已出发"
	case model.CarpoolStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// FormatBookingStatus 格式化预约状态显示名称
func FormatBookingStatus(s string) string {
	switch s {
	case model.BookingStatusPending:
		return "待确认"
	case model.BookingStatusConfirmed:
		return "已确认"
	case model.BookingStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}