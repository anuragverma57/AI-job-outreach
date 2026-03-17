package router

import (
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/config"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/handler"
	appMiddleware "github.com/anuragverma/ai-job-outreach/api-gateway/internal/middleware"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(app *fiber.App, db *pgxpool.Pool, cfg *config.Config) {
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(requestid.New())

	// Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)

	// Handlers
	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService, cfg)

	// Public routes
	app.Get("/health", healthHandler.Check)

	auth := app.Group("/api/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	// Protected routes (everything under /api beyond /api/auth)
	api := app.Group("/api", appMiddleware.AuthRequired(authService))
	api.Get("/me", authHandler.Me)
}
