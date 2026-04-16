package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/client"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
)

var (
	ErrNoResumesAvailable     = errors.New("upload a resume first")
	ErrSmartApplyInsufficient = errors.New("smart apply extraction missing required fields")
)

const maxSmartApplyRawTextLength = 30000

type SmartApplyService struct {
	appRepo    *repository.ApplicationRepository
	emailRepo  *repository.EmailRepository
	resumeRepo *repository.ResumeRepository
	aiClient   *client.AIClient
}

func NewSmartApplyService(
	appRepo *repository.ApplicationRepository,
	emailRepo *repository.EmailRepository,
	resumeRepo *repository.ResumeRepository,
	aiClient *client.AIClient,
) *SmartApplyService {
	return &SmartApplyService{
		appRepo:    appRepo,
		emailRepo:  emailRepo,
		resumeRepo: resumeRepo,
		aiClient:   aiClient,
	}
}

func (s *SmartApplyService) CreateDraft(ctx context.Context, userID string, req model.SmartApplyRequest) (*model.SmartApplyResponse, error) {
	rawText := strings.TrimSpace(req.RawText)
	if rawText == "" {
		return nil, fmt.Errorf("%w: raw_text is required", ErrInvalidInput)
	}
	if len(rawText) > maxSmartApplyRawTextLength {
		return nil, fmt.Errorf("%w: raw_text exceeds %d characters", ErrInvalidInput, maxSmartApplyRawTextLength)
	}

	resumes, err := s.resumeRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(resumes) == 0 {
		return nil, ErrNoResumesAvailable
	}

	candidates := make([]client.SmartApplyResumeCandidate, 0, len(resumes))
	resumeByID := make(map[string]model.Resume, len(resumes))
	for _, r := range resumes {
		resumeByID[r.ID] = r
		if r.ParsedText == nil || strings.TrimSpace(*r.ParsedText) == "" {
			continue
		}
		candidates = append(candidates, client.SmartApplyResumeCandidate{
			ResumeID:   r.ID,
			ParsedText: strings.TrimSpace(*r.ParsedText),
		})
	}
	if len(candidates) == 0 {
		return nil, ErrResumeParsedTextEmpty
	}

	extractResp, err := s.aiClient.SmartApplyExtractAndMatch(&client.SmartApplyExtractRequest{
		RawText: rawText,
		Resumes: candidates,
	})
	if err != nil {
		return nil, fmt.Errorf("smart apply extraction failed: %w", err)
	}

	company := strings.TrimSpace(extractResp.CompanyName)
	role := strings.TrimSpace(extractResp.Role)
	jobDescription := strings.TrimSpace(extractResp.JobDescription)
	if company == "" || role == "" || jobDescription == "" {
		return nil, ErrSmartApplyInsufficient
	}

	selectedResumeID := strings.TrimSpace(extractResp.SelectedResumeID)
	selected, ok := resumeByID[selectedResumeID]
	if !ok || selected.ParsedText == nil || strings.TrimSpace(*selected.ParsedText) == "" {
		selectedResumeID = candidates[0].ResumeID
		selected = resumeByID[selectedResumeID]
	}

	var recruiterEmail string
	if extractResp.RecruiterEmail != nil {
		recruiterEmail = strings.TrimSpace(*extractResp.RecruiterEmail)
	}
	var jobLink string
	if extractResp.JobLink != nil {
		jobLink = strings.TrimSpace(*extractResp.JobLink)
	}

	app, err := s.appRepo.Create(ctx, &model.Application{
		UserID:         userID,
		ResumeID:       &selectedResumeID,
		CompanyName:    company,
		Role:           role,
		RecruiterEmail: recruiterEmail,
		JobDescription: jobDescription,
		JobLink:        jobLink,
	})
	if err != nil {
		return nil, err
	}

	emailAI, err := s.aiClient.GenerateEmail(&client.GenerateEmailRequest{
		ResumeText:     strings.TrimSpace(*selected.ParsedText),
		JobDescription: app.JobDescription,
		CompanyName:    app.CompanyName,
		Role:           app.Role,
		RecruiterEmail: app.RecruiterEmail,
		JobLink:        app.JobLink,
		Tone:           "formal",
	})
	if err != nil {
		return nil, fmt.Errorf("smart apply email generation failed: %w", err)
	}

	email := &model.Email{
		ApplicationID: app.ID,
		Subject:       strings.TrimSpace(emailAI.Subject),
		Body:          strings.TrimSpace(emailAI.Body),
		MatchScore:    &emailAI.MatchScore,
		Reasoning:     &emailAI.Reasoning,
	}
	createdEmail, err := s.emailRepo.CreateOrReplace(ctx, email)
	if err != nil {
		return nil, err
	}

	conf := strings.TrimSpace(extractResp.ExtractionConfidence)
	if conf == "" {
		conf = "medium"
	}

	return &model.SmartApplyResponse{
		Application: app,
		Email:       createdEmail,
		Meta: model.SmartApplyMeta{
			ExtractionConfidence: conf,
		},
	}, nil
}
