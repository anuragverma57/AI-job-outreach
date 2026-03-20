package handler

import (
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ApplicationHandler struct {
	appService *service.ApplicationService
}

func NewApplicationHandler(appService *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

func (h *ApplicationHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req model.CreateApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	app, err := h.appService.Create(c.Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"application": app})
}

func (h *ApplicationHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	apps, err := h.appService.List(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch applications"})
	}

	return c.JSON(fiber.Map{"applications": apps})
}

func (h *ApplicationHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	appID := c.Params("id")

	app, err := h.appService.GetByID(c.Context(), userID, appID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"application": app})
}

func (h *ApplicationHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	appID := c.Params("id")

	var req model.UpdateApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	app, err := h.appService.Update(c.Context(), userID, appID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"application": app})
}

func (h *ApplicationHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	appID := c.Params("id")

	if err := h.appService.Delete(c.Context(), userID, appID); err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"message": "application deleted"})
}

func (h *ApplicationHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrApplicationNotOwned):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you don't have access to this application"})
	case errors.Is(err, repository.ErrApplicationNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
	case errors.Is(err, repository.ErrResumeNotFound):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "resume not found"})
	case errors.Is(err, service.ErrResumeNotOwned):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "resume does not belong to you"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
