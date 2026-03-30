package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/client"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/queue"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
)

var (
	ErrResumeParsedTextEmpty = errors.New("resume has no parsed text; upload a resume first")
	ErrEmailNotOwned         = errors.New("email does not belong to this user")
	ErrEmailNotSchedulable   = errors.New("only draft emails can be scheduled")
	ErrScheduleInPast        = errors.New("send_at must be in the future")
	ErrEmailNotScheduled     = errors.New("email is not currently scheduled")
	ErrMaxDelayExceeded      = errors.New("maximum schedule delay is 30 days")
)

const maxScheduleDelay = 30 * 24 * time.Hour

type EmailService struct {
	emailRepo  *repository.EmailRepository
	appRepo    *repository.ApplicationRepository
	resumeRepo *repository.ResumeRepository
	aiClient   *client.AIClient
	queue      *queue.RedisQueue
}

func NewEmailService(
	emailRepo *repository.EmailRepository,
	appRepo *repository.ApplicationRepository,
	resumeRepo *repository.ResumeRepository,
	aiClient *client.AIClient,
	q *queue.RedisQueue,
) *EmailService {
	return &EmailService{
		emailRepo:  emailRepo,
		appRepo:    appRepo,
		resumeRepo: resumeRepo,
		aiClient:   aiClient,
		queue:      q,
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

func (s *EmailService) ScheduleEmail(ctx context.Context, userID, emailID string, req model.ScheduleEmailRequest) (*model.Email, error) {
	email, err := s.emailRepo.FindByID(ctx, emailID)
	if err != nil {
		return nil, err
	}

	if err := s.verifyEmailOwnership(ctx, email, userID); err != nil {
		return nil, err
	}

	if email.Status != "draft" {
		return nil, ErrEmailNotSchedulable
	}

	sendAt, err := resolveSendAt(req)
	if err != nil {
		return nil, err
	}

	if err := s.queue.Enqueue(ctx, email.ID, sendAt); err != nil {
		return nil, fmt.Errorf("failed to enqueue email: %w", err)
	}

	updated, err := s.emailRepo.UpdateStatus(ctx, email.ID, "scheduled", &sendAt)
	if err != nil {
		_, _ = s.queue.Cancel(ctx, email.ID)
		return nil, fmt.Errorf("failed to update email status: %w", err)
	}

	return updated, nil
}

func (s *EmailService) CancelSchedule(ctx context.Context, userID, emailID string) (*model.Email, error) {
	email, err := s.emailRepo.FindByID(ctx, emailID)
	if err != nil {
		return nil, err
	}

	if err := s.verifyEmailOwnership(ctx, email, userID); err != nil {
		return nil, err
	}

	if email.Status != "scheduled" {
		return nil, ErrEmailNotScheduled
	}

	_, _ = s.queue.Cancel(ctx, email.ID)

	return s.emailRepo.UpdateStatus(ctx, email.ID, "draft", nil)
}

func (s *EmailService) RescheduleEmail(ctx context.Context, userID, emailID string, req model.ScheduleEmailRequest) (*model.Email, error) {
	email, err := s.emailRepo.FindByID(ctx, emailID)
	if err != nil {
		return nil, err
	}

	if err := s.verifyEmailOwnership(ctx, email, userID); err != nil {
		return nil, err
	}

	if email.Status != "scheduled" && email.Status != "draft" {
		return nil, ErrEmailNotSchedulable
	}

	sendAt, err := resolveSendAt(req)
	if err != nil {
		return nil, err
	}

	// Remove old schedule if any, then enqueue the new one
	_, _ = s.queue.Cancel(ctx, email.ID)

	if err := s.queue.Enqueue(ctx, email.ID, sendAt); err != nil {
		return nil, fmt.Errorf("failed to enqueue email: %w", err)
	}

	updated, err := s.emailRepo.UpdateStatus(ctx, email.ID, "scheduled", &sendAt)
	if err != nil {
		_, _ = s.queue.Cancel(ctx, email.ID)
		return nil, err
	}

	return updated, nil
}

func (s *EmailService) ListByStatus(ctx context.Context, userID, status string) ([]model.Email, error) {
	return s.emailRepo.ListByUserAndStatus(ctx, userID, status)
}

func (s *EmailService) verifyEmailOwnership(ctx context.Context, email *model.Email, userID string) error {
	app, err := s.appRepo.FindByID(ctx, email.ApplicationID)
	if err != nil {
		return err
	}
	if app.UserID != userID {
		return ErrEmailNotOwned
	}
	return nil
}

// resolveSendAt converts either send_at or delay_seconds into a concrete time, with validation
func resolveSendAt(req model.ScheduleEmailRequest) (time.Time, error) {
	now := time.Now()

	if req.SendAt != nil {
		sendAt := req.SendAt.UTC()
		// Allow 30 seconds of clock skew
		if sendAt.Before(now.Add(-30 * time.Second)) {
			return time.Time{}, ErrScheduleInPast
		}
		if sendAt.After(now.Add(maxScheduleDelay)) {
			return time.Time{}, ErrMaxDelayExceeded
		}
		return sendAt, nil
	}

	if req.DelaySeconds != nil && *req.DelaySeconds > 0 {
		delay := time.Duration(*req.DelaySeconds) * time.Second
		if delay > maxScheduleDelay {
			return time.Time{}, ErrMaxDelayExceeded
		}
		return now.Add(delay), nil
	}

	return time.Time{}, fmt.Errorf("%w: provide send_at or delay_seconds", ErrInvalidInput)
}
