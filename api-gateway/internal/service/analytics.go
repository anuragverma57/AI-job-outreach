package service

import (
	"context"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsService(repo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// GetSummary loads aggregates and returns the analytics summary with LOV zero-filled and derived rates.
func (s *AnalyticsService) GetSummary(ctx context.Context, userID string) (*model.AnalyticsSummary, error) {
	appByStatus, err := s.repo.GetApplicationStatusCounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	emailByStatus, err := s.repo.GetEmailStatusCounts(ctx, userID)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, v := range appByStatus {
		total += v
	}

	byStatus := make(map[string]int, len(model.ApplicationPipelineStatuses))
	for _, key := range model.ApplicationPipelineStatuses {
		byStatus[key] = appByStatus[key]
	}

	emails := model.AnalyticsEmailStats{
		Sent:      emailByStatus["sent"],
		Scheduled: emailByStatus["scheduled"],
		Failed:    emailByStatus["failed"],
		Draft:     emailByStatus["draft"],
	}

	// --- Rates (MVP) ---
	//
	// Denominator "past_draft": applications not in draft — i.e. opportunities the user has moved
	// beyond the initial draft (includes applied, replied, etc.). We use max(1, past_draft) to avoid
	// divide-by-zero when the user has no applications or only drafts.
	//
	// reply_rate: among non-draft applications, what fraction show a positive reply signal
	// (status is replied, interview, or offer). Formula:
	//   (replied + interview + offer) / max(1, total_applications - draft)
	//
	// interview_rate: among non-draft applications, what fraction reached at least interview:
	//   interview / max(1, total_applications - draft)
	draftCount := byStatus["draft"]
	pastDraft := total - draftCount
	if pastDraft < 0 {
		pastDraft = 0
	}
	denom := pastDraft
	if denom < 1 {
		denom = 1
	}

	repliedOrBetter := byStatus["replied"] + byStatus["interview"] + byStatus["offer"]
	replyRate := float64(repliedOrBetter) / float64(denom)
	interviewRate := float64(byStatus["interview"]) / float64(denom)

	return &model.AnalyticsSummary{
		TotalApplications:    total,
		ApplicationsByStatus: byStatus,
		Emails:               emails,
		Rates: model.AnalyticsRates{
			ReplyRate:     replyRate,
			InterviewRate: interviewRate,
		},
	}, nil
}
