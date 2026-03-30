package router

import (
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/client"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/config"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/handler"
	appMiddleware "github.com/anuragverma/ai-job-outreach/api-gateway/internal/middleware"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/queue"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(app *fiber.App, db *pgxpool.Pool, rq *queue.RedisQueue, cfg *config.Config) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization",
		AllowCredentials: true,
	}))
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(requestid.New())

	// Clients
	aiClient := client.NewAIClient(cfg.AIServiceURL)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	resumeRepo := repository.NewResumeRepository(db)
	appRepo := repository.NewApplicationRepository(db)
	emailRepo := repository.NewEmailRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	resumeService := service.NewResumeService(resumeRepo, cfg.UploadDir, aiClient)
	appService := service.NewApplicationService(appRepo, resumeRepo)
	emailService := service.NewEmailService(emailRepo, appRepo, resumeRepo, aiClient, rq)

	// Handlers
	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService, cfg)
	resumeHandler := handler.NewResumeHandler(resumeService)
	appHandler := handler.NewApplicationHandler(appService)
	emailHandler := handler.NewEmailHandler(emailService)

	// Public routes
	app.Get("/health", healthHandler.Check)

	auth := app.Group("/api/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	// Protected routes
	api := app.Group("/api", appMiddleware.AuthRequired(authService))
	api.Get("/me", authHandler.Me)

	resumes := api.Group("/resumes")
	resumes.Post("/", resumeHandler.Upload)
	resumes.Get("/", resumeHandler.List)
	resumes.Delete("/:id", resumeHandler.Delete)

	applications := api.Group("/applications")
	applications.Post("/", appHandler.Create)
	applications.Get("/", appHandler.List)
	applications.Get("/:id", appHandler.GetByID)
	applications.Put("/:id", appHandler.Update)
	applications.Delete("/:id", appHandler.Delete)
	applications.Post("/:id/generate-email", emailHandler.GenerateEmail)
	applications.Post("/:id/regenerate-email", emailHandler.RegenerateEmail)
	applications.Get("/:id/email", emailHandler.GetByApplication)

	emails := api.Group("/emails")
	emails.Get("/", emailHandler.ListByStatus)
	emails.Put("/:id", emailHandler.Update)
	emails.Post("/:id/schedule", emailHandler.Schedule)
	emails.Delete("/:id/schedule", emailHandler.CancelSchedule)
	emails.Put("/:id/schedule", emailHandler.Reschedule)
}
