package model

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"campus-service/internal/config"
	"campus-service/internal/pkg/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	var err error

	// 配置日志级别
	var gormLog gormlogger.Interface
	if config.Cfg.Server.Mode == "release" {
		gormLog = gormlogger.Default.LogMode(gormlogger.Silent)
	} else {
		gormLog = gormlogger.Default.LogMode(gormlogger.Info)
	}

	// 连接数据库
	DB, err = gorm.Open(mysql.Open(config.Cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层sql.DB并配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully")

	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	err := DB.AutoMigrate(
		&User{},
		&Errand{},
		&ErrandOrder{},
		&Carpool{},
		&CarpoolBooking{},
		&MarketItem{},
		&MarketFavorite{},
		&LostFound{},
		&LostFoundClaim{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	logger.Info("Database migration completed")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}