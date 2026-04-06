package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetApplicationStatusCounts returns per-status counts for the user's applications.
func (r *AnalyticsRepository) GetApplicationStatusCounts(ctx context.Context, userID string) (map[string]int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT status, COUNT(*)::bigint FROM applications WHERE user_id = $1 GROUP BY status`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var status string
		var n int64
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[status] = int(n)
	}
	return out, rows.Err()
}

// GetEmailStatusCounts joins emails to applications so only the user's rows are counted.
func (r *AnalyticsRepository) GetEmailStatusCounts(ctx context.Context, userID string) (map[string]int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT e.status, COUNT(*)::bigint
		 FROM emails e
		 INNER JOIN applications a ON a.id = e.application_id
		 WHERE a.user_id = $1
		 GROUP BY e.status`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var status string
		var n int64
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[status] = int(n)
	}
	return out, rows.Err()
}
