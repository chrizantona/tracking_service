package app

import (
	"context"
	"database/sql" // Возвращаем стандартный импорт
	"net/http"
	"time"
	"backend/config"
	"backend/internal/controller"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	// Удаляем импорт sqlx
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *sql.DB 
	router *gin.Engine
	srv    *http.Server
}

func NewServer(cfg *config.Config, logger *zap.Logger, db *sql.DB) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(middleware.ZapLogger(logger))
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	registerUserRoutes(router, cfg, db)
	registerOrderRoutes(router, cfg, db, logger)
	registerCourierRoutes(router, cfg, db, logger)

	srv := &http.Server{
		Addr:           ":" + cfg.ServerPort,
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
		router: router,
		srv:    srv,
	}
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func registerUserRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB) {
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService, cfg.JWTSecret)

	router.POST("/register", userController.Register)
	router.POST("/login", userController.Login)
}

func registerOrderRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB, logger *zap.Logger) {
	orderRepo := repository.NewOrderRepository(db, logger)
	orderService := service.NewOrderService(orderRepo)
	orderController := controller.NewOrderController(orderService)

	router.POST("/orders", orderController.CreateOrder)
	router.GET("/orders", orderController.GetOrders)
	router.GET("/orders/:id", orderController.GetOrder)
	router.PUT("/orders/:id", orderController.UpdateOrder)
	router.DELETE("/orders/:id", orderController.DeleteOrder)
}

func registerCourierRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB, logger *zap.Logger) {
	courierRepo := repository.NewCourierRepository(db, logger) 
	courierService := service.NewCourierService(courierRepo, logger)
	courierController := controller.NewCourierController(courierService)

	courierRoutes := router.Group("/couriers")
    {
        courierRoutes.GET("/:id", courierController.GetCourier)
        courierRoutes.PUT("/:id/status", courierController.UpdateStatus)
        courierRoutes.PUT("/:id/location", courierController.UpdateLocation)
        courierRoutes.GET("/nearest", courierController.FindNearestCouriers)
    }
}