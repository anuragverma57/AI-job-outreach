package router

import (
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(app *fiber.App, db *pgxpool.Pool) {
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(requestid.New())

	healthHandler := handler.NewHealthHandler(db)

	app.Get("/health", healthHandler.Check)
}
