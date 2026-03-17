package handler

import (
	"errors"
	"time"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/config"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	user, tokens, err := h.authService.Register(c.Context(), req)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	h.setAuthCookies(c, tokens)

	return c.Status(fiber.StatusCreated).JSON(model.AuthResponse{User: *user})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	userAgent := c.Get("User-Agent")
	ipAddress := c.IP()

	user, tokens, err := h.authService.Login(c.Context(), req, userAgent, ipAddress)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	h.setAuthCookies(c, tokens)

	return c.JSON(model.AuthResponse{User: *user})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	rawRefreshToken := c.Cookies("refresh_token")
	if rawRefreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "refresh token not found",
		})
	}

	userAgent := c.Get("User-Agent")
	ipAddress := c.IP()

	user, tokens, err := h.authService.Refresh(c.Context(), rawRefreshToken, userAgent, ipAddress)
	if err != nil {
		h.clearAuthCookies(c)
		return h.handleAuthError(c, err)
	}

	h.setAuthCookies(c, tokens)

	return c.JSON(model.AuthResponse{User: *user})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	rawRefreshToken := c.Cookies("refresh_token")

	_ = h.authService.Logout(c.Context(), rawRefreshToken)

	h.clearAuthCookies(c)

	return c.JSON(fiber.Map{"message": "logged out"})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(model.AuthResponse{User: *user})
}

// --- cookie helpers ---

func (h *AuthHandler) setAuthCookies(c *fiber.Ctx, tokens *service.TokenPair) {
	sameSite := h.sameSiteMode()

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   int(h.cfg.JWT.AccessTokenExpiry.Seconds()),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: sameSite,
		Domain:   h.cfg.Cookie.Domain,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/api/auth",
		MaxAge:   int(h.cfg.JWT.RefreshTokenExpiry.Seconds()),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: sameSite,
		Domain:   h.cfg.Cookie.Domain,
	})
}

func (h *AuthHandler) clearAuthCookies(c *fiber.Ctx) {
	sameSite := h.sameSiteMode()

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: sameSite,
		Domain:   h.cfg.Cookie.Domain,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: sameSite,
		Domain:   h.cfg.Cookie.Domain,
	})
}

func (h *AuthHandler) sameSiteMode() string {
	switch h.cfg.Cookie.SameSite {
	case "Strict":
		return "Strict"
	case "None":
		return "None"
	default:
		return "Lax"
	}
}

func (h *AuthHandler) handleAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	case errors.Is(err, service.ErrTokenExpired):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token has expired"})
	case errors.Is(err, service.ErrTokenRevoked):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "session revoked, please login again"})
	case errors.Is(err, repository.ErrEmailExists):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
