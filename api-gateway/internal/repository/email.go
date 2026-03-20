package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailNotFound = errors.New("email not found")

type EmailRepository struct {
	db *pgxpool.Pool
}

func NewEmailRepository(db *pgxpool.Pool) *EmailRepository {
	return &EmailRepository{db: db}
}

// CreateOrReplace deletes any existing draft email for the application and creates a new one
func (r *EmailRepository) CreateOrReplace(ctx context.Context, email *model.Email) (*model.Email, error) {
	result := &model.Email{}

	keyPointsJSON, _ := json.Marshal(email.KeyPoints)

	// Remove existing draft for this application
	_, _ = r.db.Exec(ctx, `DELETE FROM emails WHERE application_id = $1 AND status = 'draft'`, email.ApplicationID)

	err := r.db.QueryRow(ctx,
		`INSERT INTO emails (application_id, subject, body, status, match_score, key_points, reasoning)
		 VALUES ($1, $2, $3, 'draft', $4, $5, $6)
		 RETURNING id, application_id, subject, body, status, match_score, key_points, reasoning, scheduled_at, sent_at, retry_count, created_at, updated_at`,
		email.ApplicationID, email.Subject, email.Body, email.MatchScore, keyPointsJSON, email.Reasoning,
	).Scan(&result.ID, &result.ApplicationID, &result.Subject, &result.Body, &result.Status,
		&result.MatchScore, &result.KeyPoints, &result.Reasoning, &result.ScheduledAt, &result.SentAt,
		&result.RetryCount, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *EmailRepository) FindByApplicationID(ctx context.Context, applicationID string) (*model.Email, error) {
	result := &model.Email{}

	err := r.db.QueryRow(ctx,
		`SELECT id, application_id, subject, body, status, match_score, key_points, reasoning, scheduled_at, sent_at, retry_count, created_at, updated_at
		 FROM emails WHERE application_id = $1
		 ORDER BY created_at DESC LIMIT 1`,
		applicationID,
	).Scan(&result.ID, &result.ApplicationID, &result.Subject, &result.Body, &result.Status,
		&result.MatchScore, &result.KeyPoints, &result.Reasoning, &result.ScheduledAt, &result.SentAt,
		&result.RetryCount, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, err
	}
	return result, nil
}

func (r *EmailRepository) FindByID(ctx context.Context, id string) (*model.Email, error) {
	result := &model.Email{}

	err := r.db.QueryRow(ctx,
		`SELECT id, application_id, subject, body, status, match_score, key_points, reasoning, scheduled_at, sent_at, retry_count, created_at, updated_at
		 FROM emails WHERE id = $1`,
		id,
	).Scan(&result.ID, &result.ApplicationID, &result.Subject, &result.Body, &result.Status,
		&result.MatchScore, &result.KeyPoints, &result.Reasoning, &result.ScheduledAt, &result.SentAt,
		&result.RetryCount, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, err
	}
	return result, nil
}

func (r *EmailRepository) Update(ctx context.Context, id, subject, body string) (*model.Email, error) {
	result := &model.Email{}

	err := r.db.QueryRow(ctx,
		`UPDATE emails SET subject = $2, body = $3, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, application_id, subject, body, status, match_score, key_points, reasoning, scheduled_at, sent_at, retry_count, created_at, updated_at`,
		id, subject, body,
	).Scan(&result.ID, &result.ApplicationID, &result.Subject, &result.Body, &result.Status,
		&result.MatchScore, &result.KeyPoints, &result.Reasoning, &result.ScheduledAt, &result.SentAt,
		&result.RetryCount, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, err
	}
	return result, nil
}
