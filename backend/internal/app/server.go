package app

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"backend/config"
	"backend/internal/controller"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *sql.DB
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


	userRepo    := repository.NewUserRepository(db)
	orderRepo   := repository.NewOrderRepository(db, logger)
	courierRepo := repository.NewCourierRepository(db, logger)


	userSvc    := service.NewUserService(userRepo)
	orderSvc   := service.NewOrderService(orderRepo, courierRepo) 
	courierSvc := service.NewCourierService(courierRepo, logger)


	userCtrl    := controller.NewUserController(userSvc, cfg.JWTSecret)
	orderCtrl   := controller.NewOrderController(orderSvc)
	courierCtrl := controller.NewCourierController(courierSvc)

	registerUserRoutes(router, userCtrl)
	registerOrderRoutes(router, orderCtrl)
	registerCourierRoutes(router, courierCtrl)


	httpSrv := &http.Server{
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
		srv:    httpSrv,
	}
}


func (s *Server) Start() error               { return s.srv.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }



func registerUserRoutes(r *gin.Engine, uc *controller.UserController) {
	r.POST("/register", uc.Register)
	r.POST("/login", uc.Login)
}

func registerOrderRoutes(r *gin.Engine, oc *controller.OrderController) {
	orders := r.Group("/orders")
	{
		orders.POST("", oc.CreateOrder)
		orders.GET("", oc.GetOrders)
		orders.GET("/:id", oc.GetOrder)
		orders.PUT("/:id", oc.UpdateOrder)
		orders.DELETE("/:id", oc.DeleteOrder)
	}
}

func registerCourierRoutes(r *gin.Engine, cc *controller.CourierController) {
	couriers := r.Group("/couriers")
	{
		couriers.GET("/nearest", cc.FindNearestCouriers)
		couriers.GET("/:id", cc.GetCourier)
		couriers.PUT("/:id/status", cc.UpdateStatus)
		couriers.PUT("/:id/location", cc.UpdateLocation)
	}
}
