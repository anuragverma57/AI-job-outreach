package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/client"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
)

var (
	ErrResumeParsedTextEmpty = errors.New("resume has no parsed text; upload a resume first")
	ErrEmailNotOwned         = errors.New("email does not belong to this user")
)

type EmailService struct {
	emailRepo *repository.EmailRepository
	appRepo   *repository.ApplicationRepository
	resumeRepo *repository.ResumeRepository
	aiClient  *client.AIClient
}

func NewEmailService(
	emailRepo *repository.EmailRepository,
	appRepo *repository.ApplicationRepository,
	resumeRepo *repository.ResumeRepository,
	aiClient *client.AIClient,
) *EmailService {
	return &EmailService{
		emailRepo:  emailRepo,
		appRepo:    appRepo,
		resumeRepo: resumeRepo,
		aiClient:   aiClient,
	}
}

func (s *EmailService) GenerateEmail(ctx context.Context, userID, applicationID, tone string) (*model.Email, error) {
	app, err := s.appRepo.FindByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrApplicationNotOwned
	}

	if app.ResumeID == nil || *app.ResumeID == "" {
		return nil, fmt.Errorf("application has no linked resume")
	}

	resume, err := s.resumeRepo.FindByID(ctx, *app.ResumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resume: %w", err)
	}

	if resume.ParsedText == nil || *resume.ParsedText == "" {
		return nil, ErrResumeParsedTextEmpty
	}

	if tone == "" {
		tone = "formal"
	}

	aiReq := &client.GenerateEmailRequest{
		ResumeText:     *resume.ParsedText,
		JobDescription: app.JobDescription,
		CompanyName:    app.CompanyName,
		Role:           app.Role,
		RecruiterEmail: app.RecruiterEmail,
		JobLink:        app.JobLink,
		Tone:           tone,
	}

	aiResp, err := s.aiClient.GenerateEmail(aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI email generation failed: %w", err)
	}

	keyPointsJSON, _ := json.Marshal(aiResp.KeyPoints)

	email := &model.Email{
		ApplicationID: applicationID,
		Subject:       aiResp.Subject,
		Body:          aiResp.Body,
		MatchScore:    &aiResp.MatchScore,
		KeyPoints:     keyPointsJSON,
		Reasoning:     &aiResp.Reasoning,
	}

	return s.emailRepo.CreateOrReplace(ctx, email)
}

func (s *EmailService) GetByApplicationID(ctx context.Context, userID, applicationID string) (*model.Email, error) {
	app, err := s.appRepo.FindByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrApplicationNotOwned
	}

	return s.emailRepo.FindByApplicationID(ctx, applicationID)
}

func (s *EmailService) UpdateEmail(ctx context.Context, userID, emailID string, req model.UpdateEmailRequest) (*model.Email, error) {
	email, err := s.emailRepo.FindByID(ctx, emailID)
	if err != nil {
		return nil, err
	}

	// Verify ownership through the application
	app, err := s.appRepo.FindByID(ctx, email.ApplicationID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrEmailNotOwned
	}

	if req.Subject == "" || req.Body == "" {
		return nil, fmt.Errorf("%w: subject and body are required", ErrInvalidInput)
	}

	return s.emailRepo.Update(ctx, emailID, req.Subject, req.Body)
}
