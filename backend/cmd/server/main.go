package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"family-tree-backend/internal/auth"
	"family-tree-backend/internal/config"
	"family-tree-backend/internal/database"
	"family-tree-backend/internal/handler"
	"family-tree-backend/internal/logger"
	"family-tree-backend/internal/middleware"
	"family-tree-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Database
	pool, err := database.NewPool(ctx, cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, "migrations", log); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}

	// Services
	jwtService := auth.NewJWTService(cfg.JWTSecret)
	userService := service.NewUserService(pool, jwtService, cfg.TelegramBotToken, log)
	treeService := service.NewTreeService(pool)
	personService := service.NewPersonService(pool)
	relationService := service.NewRelationService(pool)

	// Handlers
	authHandler := handler.NewAuthHandler(userService)
	treeHandler := handler.NewTreeHandler(treeService, personService, relationService)
	personHandler := handler.NewPersonHandler(personService, treeService)
	relationHandler := handler.NewRelationHandler(relationService, treeService)

	// Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// CORS for development
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public routes
	r.POST("/api/auth/telegram", authHandler.Telegram)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.JWTMiddleware(jwtService))

	protected.GET("/trees", treeHandler.List)
	protected.POST("/trees", treeHandler.Create)
	protected.GET("/trees/:id", treeHandler.Get)
	protected.DELETE("/trees/:id", treeHandler.Delete)

	protected.POST("/persons", personHandler.Create)
	protected.PATCH("/persons/:id", personHandler.Update)
	protected.DELETE("/persons/:id", personHandler.Delete)

	protected.POST("/relations", relationHandler.Create)
	protected.DELETE("/relations/:id", relationHandler.Delete)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.ServerPort
	}
	if port == "" {
		port = "8080"
	}

	// Server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("port", port).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")
	ctx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	fmt.Println("server stopped")
}
