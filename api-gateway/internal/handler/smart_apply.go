package handler

import (
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
)

type SmartApplyHandler struct {
	smartApplyService *service.SmartApplyService
}

func NewSmartApplyHandler(smartApplyService *service.SmartApplyService) *SmartApplyHandler {
	return &SmartApplyHandler{smartApplyService: smartApplyService}
}

func (h *SmartApplyHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req model.SmartApplyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	resp, err := h.smartApplyService.CreateDraft(c.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrNoResumesAvailable), errors.Is(err, service.ErrResumeParsedTextEmpty):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrSmartApplyInsufficient):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
