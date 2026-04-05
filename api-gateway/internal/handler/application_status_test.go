package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/service"
	"github.com/gofiber/fiber/v2"
)

type stubApplicationService struct {
	updateStatusFn func(ctx context.Context, userID, appID string, req model.UpdateApplicationStatusRequest) (*model.Application, error)
}

func (s *stubApplicationService) Create(context.Context, string, model.CreateApplicationRequest) (*model.Application, error) {
	panic("unexpected call to Create")
}

func (s *stubApplicationService) List(context.Context, string) ([]model.Application, error) {
	panic("unexpected call to List")
}

func (s *stubApplicationService) GetByID(context.Context, string, string) (*model.Application, error) {
	panic("unexpected call to GetByID")
}

func (s *stubApplicationService) Update(context.Context, string, string, model.UpdateApplicationRequest) (*model.Application, error) {
	panic("unexpected call to Update")
}

func (s *stubApplicationService) Delete(context.Context, string, string) error {
	panic("unexpected call to Delete")
}

func (s *stubApplicationService) UpdateStatus(ctx context.Context, userID, appID string, req model.UpdateApplicationStatusRequest) (*model.Application, error) {
	if s.updateStatusFn == nil {
		panic("updateStatusFn not set")
	}
	return s.updateStatusFn(ctx, userID, appID, req)
}

func TestApplicationHandler_UpdateStatus(t *testing.T) {
	updated := &model.Application{
		ID: "app-1", UserID: "user-1", CompanyName: "Co", Role: "Eng",
		Status: "interview", CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC(),
	}

	stub := &stubApplicationService{
		updateStatusFn: func(ctx context.Context, userID, appID string, req model.UpdateApplicationStatusRequest) (*model.Application, error) {
			if userID != "user-1" || appID != "app-1" || req.Status != "interview" {
				t.Fatalf("unexpected args: user=%q app=%q status=%q", userID, appID, req.Status)
			}
			return updated, nil
		},
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", "user-1")
		return c.Next()
	})
	h := &ApplicationHandler{appService: stub}
	app.Patch("/applications/:id/status", h.UpdateStatus)

	req := httptest.NewRequest("PATCH", "/applications/app-1/status", strings.NewReader(`{"status":"interview"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Application model.Application `json:"application"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("json: %v body=%s", err, string(body))
	}
	if out.Application.ID != updated.ID || out.Application.Status != "interview" {
		t.Fatalf("response application: %+v", out.Application)
	}
}

func TestApplicationHandler_UpdateStatus_invalidBody(t *testing.T) {
	stub := &stubApplicationService{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", "user-1")
		return c.Next()
	})
	h := &ApplicationHandler{appService: stub}
	app.Patch("/applications/:id/status", h.UpdateStatus)

	req := httptest.NewRequest("PATCH", "/applications/app-1/status", strings.NewReader(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestApplicationHandler_UpdateStatus_errorMapping(t *testing.T) {
	tests := []struct {
		name       string
		svcErr     error
		wantStatus int
	}{
		{"invalid input", service.ErrInvalidInput, fiber.StatusBadRequest},
		{"not owned", service.ErrApplicationNotOwned, fiber.StatusForbidden},
		{"not found", repository.ErrApplicationNotFound, fiber.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errVal := tc.svcErr
			stub := &stubApplicationService{
				updateStatusFn: func(context.Context, string, string, model.UpdateApplicationStatusRequest) (*model.Application, error) {
					return nil, errVal
				},
			}
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("userID", "user-1")
				return c.Next()
			})
			h := &ApplicationHandler{appService: stub}
			app.Patch("/applications/:id/status", h.UpdateStatus)

			req := httptest.NewRequest("PATCH", "/applications/app-1/status", strings.NewReader(`{"status":"replied"}`))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("want status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
		})
	}
}
