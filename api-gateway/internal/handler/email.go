package handler

import (
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
)

type EmailHandler struct {
	emailService *service.EmailService
}

func NewEmailHandler(emailService *service.EmailService) *EmailHandler {
	return &EmailHandler{emailService: emailService}
}

func (h *EmailHandler) GenerateEmail(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	appID := c.Params("id")

	var req model.GenerateEmailRequest
	if err := c.BodyParser(&req); err != nil {
		req.Tone = "formal"
	}
	if req.Tone == "" {
		req.Tone = "formal"
	}

	email, err := h.emailService.GenerateEmail(c.Context(), userID, appID, req.Tone)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"email": email})
}

func (h *EmailHandler) RegenerateEmail(c *fiber.Ctx) error {
	return h.GenerateEmail(c)
}

func (h *EmailHandler) GetByApplication(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	appID := c.Params("id")

	email, err := h.emailService.GetByApplicationID(c.Context(), userID, appID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"email": email})
}

func (h *EmailHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	emailID := c.Params("id")

	var req model.UpdateEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	email, err := h.emailService.UpdateEmail(c.Context(), userID, emailID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"email": email})
}

func (h *EmailHandler) Schedule(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	emailID := c.Params("id")

	var req model.ScheduleEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	email, err := h.emailService.ScheduleEmail(c.Context(), userID, emailID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"email": email})
}

func (h *EmailHandler) CancelSchedule(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	emailID := c.Params("id")

	email, err := h.emailService.CancelSchedule(c.Context(), userID, emailID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"email": email})
}

func (h *EmailHandler) Reschedule(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	emailID := c.Params("id")

	var req model.ScheduleEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	email, err := h.emailService.RescheduleEmail(c.Context(), userID, emailID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{"email": email})
}

func (h *EmailHandler) ListByStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	status := c.Query("status", "scheduled")

	emails, err := h.emailService.ListByStatus(c.Context(), userID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch emails"})
	}

	return c.JSON(fiber.Map{"emails": emails})
}

func (h *EmailHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrApplicationNotOwned):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you don't have access to this application"})
	case errors.Is(err, service.ErrEmailNotOwned):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you don't have access to this email"})
	case errors.Is(err, service.ErrResumeParsedTextEmpty):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrEmailNotSchedulable):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrEmailNotScheduled):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrScheduleInPast):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrMaxDelayExceeded):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repository.ErrApplicationNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
	case errors.Is(err, repository.ErrEmailNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no email found"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
