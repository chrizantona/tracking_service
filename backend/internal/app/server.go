package app

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"backend/config"
	"backend/internal/controller"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
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
