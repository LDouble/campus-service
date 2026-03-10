package router

import (
	"github.com/gin-gonic/gin"

	"campus-service/internal/config"
	"campus-service/internal/handler"
	"campus-service/internal/middleware"
	"campus-service/internal/model"
	"campus-service/internal/repository"
	"campus-service/internal/service"
)

// Setup 设置路由
func Setup() *gin.Engine {
	// 设置运行模式
	gin.SetMode(config.Cfg.Server.Mode)

	// 创建路由引擎
	r := gin.New()

	// 注册全局中间件
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康检查
	healthHandler := handler.NewHealthHandler()
	healthHandler.RegisterRoutes(r)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证路由
		userRepo := repository.NewUserRepository(model.DB)
		authService := service.NewAuthService(userRepo)
		authHandler := handler.NewAuthHandler(authService)
		authHandler.RegisterRoutes(v1)

		// 跑腿服务路由
		errandRepo := repository.NewErrandRepository(model.DB)
		errandService := service.NewErrandService(errandRepo, userRepo)
		errandHandler := handler.NewErrandHandler(errandService)
		errandHandler.RegisterRoutes(v1, middleware.Auth())

		// 拼车服务路由
		carpoolRepo := repository.NewCarpoolRepository(model.DB)
		carpoolService := service.NewCarpoolService(carpoolRepo, userRepo)
		carpoolHandler := handler.NewCarpoolHandler(carpoolService)
		carpoolHandler.RegisterRoutes(v1, middleware.Auth())

		// 二手交易路由
		marketRepo := repository.NewMarketRepository(model.DB)
		marketService := service.NewMarketService(marketRepo, userRepo)
		marketHandler := handler.NewMarketHandler(marketService)
		marketHandler.RegisterRoutes(v1, middleware.Auth())

		// 失物招领路由
		lfRepo := repository.NewLostFoundRepository(model.DB)
		lfService := service.NewLostFoundService(lfRepo, userRepo)
		lfHandler := handler.NewLostFoundHandler(lfService)
		lfHandler.RegisterRoutes(v1, middleware.Auth())
	}

	// 404处理
	r.NoRoute(handler.NotFoundHandler)
	r.NoMethod(handler.MethodNotAllowedHandler)

	return r
}