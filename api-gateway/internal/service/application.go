package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
)

var ErrApplicationNotOwned = errors.New("application does not belong to this user")

type ApplicationService struct {
	appRepo    *repository.ApplicationRepository
	resumeRepo *repository.ResumeRepository
}

func NewApplicationService(appRepo *repository.ApplicationRepository, resumeRepo *repository.ResumeRepository) *ApplicationService {
	return &ApplicationService{appRepo: appRepo, resumeRepo: resumeRepo}
}

func (s *ApplicationService) Create(ctx context.Context, userID string, req model.CreateApplicationRequest) (*model.Application, error) {
	if req.CompanyName == "" || req.Role == "" {
		return nil, fmt.Errorf("%w: company_name and role are required", ErrInvalidInput)
	}

	// Verify resume belongs to user if provided
	if req.ResumeID != "" {
		resume, err := s.resumeRepo.FindByID(ctx, req.ResumeID)
		if err != nil {
			return nil, fmt.Errorf("resume not found: %w", err)
		}
		if resume.UserID != userID {
			return nil, ErrResumeNotOwned
		}
	}

	app := &model.Application{
		UserID:         userID,
		CompanyName:    req.CompanyName,
		Role:           req.Role,
		RecruiterEmail: req.RecruiterEmail,
		JobDescription: req.JobDescription,
		JobLink:        req.JobLink,
	}

	if req.ResumeID != "" {
		app.ResumeID = &req.ResumeID
	}

	return s.appRepo.Create(ctx, app)
}

func (s *ApplicationService) List(ctx context.Context, userID string) ([]model.Application, error) {
	return s.appRepo.ListByUser(ctx, userID)
}

func (s *ApplicationService) GetByID(ctx context.Context, userID, appID string) (*model.Application, error) {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrApplicationNotOwned
	}
	return app, nil
}

func (s *ApplicationService) Update(ctx context.Context, userID, appID string, req model.UpdateApplicationRequest) (*model.Application, error) {
	existing, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID {
		return nil, ErrApplicationNotOwned
	}
	return s.appRepo.Update(ctx, appID, &req)
}

func (s *ApplicationService) Delete(ctx context.Context, userID, appID string) error {
	existing, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return ErrApplicationNotOwned
	}
	return s.appRepo.Delete(ctx, appID)
}

// UpdateStatus sets pipeline status (LOV) for a user-owned application (Option A: any LOV from any state).
func (s *ApplicationService) UpdateStatus(ctx context.Context, userID, appID string, req model.UpdateApplicationStatusRequest) (*model.Application, error) {
	if err := model.ValidateApplicationPipelineStatus(req.Status); err != nil {
		switch {
		case errors.Is(err, model.ErrApplicationStatusRequired):
			return nil, fmt.Errorf("%w: status is required", ErrInvalidInput)
		case errors.Is(err, model.ErrInvalidApplicationPipelineStatus):
			return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
		default:
			return nil, err
		}
	}

	existing, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID {
		return nil, ErrApplicationNotOwned
	}

	return s.appRepo.UpdateStatus(ctx, appID, req.Status)
}
