package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"campus-service/internal/config"
	"campus-service/internal/model"
	"campus-service/internal/pkg/logger"
	"campus-service/internal/router"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 初始化配置
	if err := config.Init(configPath); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(config.Cfg.Server.Mode); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := model.InitDB(); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer model.CloseDB()

	// 自动迁移
	if err := model.AutoMigrate(); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 初始化Redis
	if err := model.InitRedis(); err != nil {
		logger.Fatal("初始化Redis失败", zap.Error(err))
	}
	defer model.CloseRedis()

	// 设置路由
	r := router.Setup()

	// 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  config.Cfg.Server.ReadTimeout,
		WriteTimeout: config.Cfg.Server.WriteTimeout,
	}

	// 优雅关闭
	go func() {
		logger.Info("服务器启动", zap.Int("port", config.Cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	// 给5秒时间处理未完成的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器关闭失败", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}