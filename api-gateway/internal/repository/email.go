package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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

func (r *EmailRepository) UpdateStatus(ctx context.Context, id, status string, scheduledAt *time.Time) (*model.Email, error) {
	result := &model.Email{}

	err := r.db.QueryRow(ctx,
		`UPDATE emails SET status = $2, scheduled_at = $3, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, application_id, subject, body, status, match_score, key_points, reasoning, scheduled_at, sent_at, retry_count, created_at, updated_at`,
		id, status, scheduledAt,
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

func (r *EmailRepository) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE emails SET status = 'sent', sent_at = $2, updated_at = NOW() WHERE id = $1`,
		id, sentAt,
	)
	return err
}

func (r *EmailRepository) MarkFailed(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE emails SET status = 'failed', updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

func (r *EmailRepository) IncrementRetry(ctx context.Context, id string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`UPDATE emails SET retry_count = retry_count + 1, updated_at = NOW()
		 WHERE id = $1 RETURNING retry_count`,
		id,
	).Scan(&count)
	return count, err
}

func (r *EmailRepository) ListByUserAndStatus(ctx context.Context, userID, status string) ([]model.Email, error) {
	rows, err := r.db.Query(ctx,
		`SELECT e.id, e.application_id, e.subject, e.body, e.status, e.match_score, e.key_points, e.reasoning,
		        e.scheduled_at, e.sent_at, e.retry_count, e.created_at, e.updated_at
		 FROM emails e
		 JOIN applications a ON e.application_id = a.id
		 WHERE a.user_id = $1 AND e.status = $2
		 ORDER BY e.scheduled_at ASC`,
		userID, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []model.Email
	for rows.Next() {
		var email model.Email
		if err := rows.Scan(&email.ID, &email.ApplicationID, &email.Subject, &email.Body, &email.Status,
			&email.MatchScore, &email.KeyPoints, &email.Reasoning, &email.ScheduledAt, &email.SentAt,
			&email.RetryCount, &email.CreatedAt, &email.UpdatedAt); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	if emails == nil {
		emails = []model.Email{}
	}
	return emails, rows.Err()
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
