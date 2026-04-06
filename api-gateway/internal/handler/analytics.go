package handler

import (
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) Summary(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	summary, err := h.analyticsService.GetSummary(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch analytics summary",
		})
	}

	return c.JSON(fiber.Map{"summary": summary})
}
