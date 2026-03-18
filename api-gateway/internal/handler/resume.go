package handler

import (
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ResumeHandler struct {
	resumeService *service.ResumeService
}

func NewResumeHandler(resumeService *service.ResumeService) *ResumeHandler {
	return &ResumeHandler{resumeService: resumeService}
}

func (h *ResumeHandler) Upload(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file is required (form field: 'file')",
		})
	}

	resume, err := h.resumeService.Upload(c.Context(), userID, file)
	if err != nil {
		return h.handleResumeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"resume": resume,
	})
}

func (h *ResumeHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	resumes, err := h.resumeService.List(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch resumes",
		})
	}

	return c.JSON(fiber.Map{
		"resumes": resumes,
	})
}

func (h *ResumeHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	resumeID := c.Params("id")

	if resumeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "resume id is required",
		})
	}

	if err := h.resumeService.Delete(c.Context(), userID, resumeID); err != nil {
		return h.handleResumeError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "resume deleted",
	})
}

func (h *ResumeHandler) handleResumeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidFileType):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrFileTooLarge):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrResumeNotOwned):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you don't have access to this resume"})
	case errors.Is(err, repository.ErrResumeNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "resume not found"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
