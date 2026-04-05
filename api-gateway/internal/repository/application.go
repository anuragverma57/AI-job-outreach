package repository

import (
	"context"
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrApplicationNotFound = errors.New("application not found")

type ApplicationRepository struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(ctx context.Context, app *model.Application) (*model.Application, error) {
	result := &model.Application{}

	err := r.db.QueryRow(ctx,
		`INSERT INTO applications (user_id, resume_id, company_name, role, recruiter_email, job_description, job_link)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, user_id, resume_id, company_name, role, recruiter_email, job_description, job_link, status, created_at, updated_at`,
		app.UserID, app.ResumeID, app.CompanyName, app.Role, app.RecruiterEmail, app.JobDescription, app.JobLink,
	).Scan(&result.ID, &result.UserID, &result.ResumeID, &result.CompanyName, &result.Role,
		&result.RecruiterEmail, &result.JobDescription, &result.JobLink, &result.Status, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *ApplicationRepository) ListByUser(ctx context.Context, userID string) ([]model.Application, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, resume_id, company_name, role, recruiter_email, job_description, job_link, status, created_at, updated_at
		 FROM applications WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		var app model.Application
		if err := rows.Scan(&app.ID, &app.UserID, &app.ResumeID, &app.CompanyName, &app.Role,
			&app.RecruiterEmail, &app.JobDescription, &app.JobLink, &app.Status, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	if apps == nil {
		apps = []model.Application{}
	}
	return apps, rows.Err()
}

func (r *ApplicationRepository) FindByID(ctx context.Context, id string) (*model.Application, error) {
	app := &model.Application{}

	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, resume_id, company_name, role, recruiter_email, job_description, job_link, status, created_at, updated_at
		 FROM applications WHERE id = $1`,
		id,
	).Scan(&app.ID, &app.UserID, &app.ResumeID, &app.CompanyName, &app.Role,
		&app.RecruiterEmail, &app.JobDescription, &app.JobLink, &app.Status, &app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	return app, nil
}

func (r *ApplicationRepository) Update(ctx context.Context, id string, req *model.UpdateApplicationRequest) (*model.Application, error) {
	app := &model.Application{}

	err := r.db.QueryRow(ctx,
		`UPDATE applications SET
			company_name = COALESCE($2, company_name),
			role = COALESCE($3, role),
			recruiter_email = COALESCE($4, recruiter_email),
			job_description = COALESCE($5, job_description),
			job_link = COALESCE($6, job_link),
			updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, user_id, resume_id, company_name, role, recruiter_email, job_description, job_link, status, created_at, updated_at`,
		id, req.CompanyName, req.Role, req.RecruiterEmail, req.JobDescription, req.JobLink,
	).Scan(&app.ID, &app.UserID, &app.ResumeID, &app.CompanyName, &app.Role,
		&app.RecruiterEmail, &app.JobDescription, &app.JobLink, &app.Status, &app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	return app, nil
}

func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id, status string) (*model.Application, error) {
	app := &model.Application{}
	err := r.db.QueryRow(ctx,
		`UPDATE applications SET status = $2, updated_at = NOW() WHERE id = $1
		 RETURNING id, user_id, resume_id, company_name, role, recruiter_email, job_description, job_link, status, created_at, updated_at`,
		id, status,
	).Scan(&app.ID, &app.UserID, &app.ResumeID, &app.CompanyName, &app.Role,
		&app.RecruiterEmail, &app.JobDescription, &app.JobLink, &app.Status, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	return app, nil
}

func (r *ApplicationRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM applications WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrApplicationNotFound
	}
	return nil
}
