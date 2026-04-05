package service

import (
	"context"
	"errors"
	"testing"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
)

func TestApplicationService_UpdateStatus_validationOnly(t *testing.T) {
	s := NewApplicationService(nil, nil)

	_, err := s.UpdateStatus(context.Background(), "user-1", "00000000-0000-0000-0000-000000000001", model.UpdateApplicationStatusRequest{Status: ""})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty status: got %v; want ErrInvalidInput", err)
	}

	_, err = s.UpdateStatus(context.Background(), "user-1", "00000000-0000-0000-0000-000000000001", model.UpdateApplicationStatusRequest{Status: "not-a-lov"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid LOV: got %v; want ErrInvalidInput", err)
	}
}
