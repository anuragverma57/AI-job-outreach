package main

import (
	"fmt"
	"log"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/config"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/database"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/router"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	app := fiber.New(fiber.Config{
		AppName: "API Gateway",
	})

	router.Setup(app, db, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("API Gateway starting on %s", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
