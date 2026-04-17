package handler

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AIStreamProxyHandler proxies streaming email generation to the AI service (SSE).
// Browsers should call this route so auth stays on the gateway; the AI service URL stays server-side.
type AIStreamProxyHandler struct {
	baseURL string
	client  *http.Client
}

func NewAIStreamProxyHandler(aiServiceURL string) *AIStreamProxyHandler {
	base := strings.TrimRight(strings.TrimSpace(aiServiceURL), "/")
	return &AIStreamProxyHandler{
		baseURL: base,
		client: &http.Client{
			// No deadline on full body read — streaming responses can be long-lived.
			Timeout: 0,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 15 * time.Minute,
			},
		},
	}
}

// ProxyGenerateEmailStream forwards POST /api/ai/generate-email/stream to the AI service SSE endpoint.
func (h *AIStreamProxyHandler) ProxyGenerateEmailStream(c *fiber.Ctx) error {
	target := h.baseURL + "/ai/generate-email/stream"
	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, target, bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	ct := c.Get("Content-Type")
	if ct == "" {
		ct = "application/json"
	}
	req.Header.Set("Content-Type", ct)

	resp, err := h.client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to reach AI service"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		c.Status(resp.StatusCode)
		return c.Send(b)
	}

	if v := resp.Header.Get("Content-Type"); v != "" {
		c.Set("Content-Type", v)
	} else {
		c.Set("Content-Type", "text/event-stream")
	}
	if v := resp.Header.Get("Cache-Control"); v != "" {
		c.Set("Cache-Control", v)
	}
	if v := resp.Header.Get("X-Accel-Buffering"); v != "" {
		c.Set("X-Accel-Buffering", v)
	}

	return c.SendStream(resp.Body)
}
