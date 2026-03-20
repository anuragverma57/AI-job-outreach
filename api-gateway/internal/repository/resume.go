package repository

import (
	"context"
	"errors"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrResumeNotFound = errors.New("resume not found")

type ResumeRepository struct {
	db *pgxpool.Pool
}

func NewResumeRepository(db *pgxpool.Pool) *ResumeRepository {
	return &ResumeRepository{db: db}
}

func (r *ResumeRepository) Create(ctx context.Context, userID, fileName, filePath string) (*model.Resume, error) {
	resume := &model.Resume{}

	err := r.db.QueryRow(ctx,
		`INSERT INTO resumes (user_id, file_name, file_path)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, file_name, file_path, parsed_text, created_at`,
		userID, fileName, filePath,
	).Scan(&resume.ID, &resume.UserID, &resume.FileName, &resume.FilePath, &resume.ParsedText, &resume.CreatedAt)

	if err != nil {
		return nil, err
	}

	return resume, nil
}

func (r *ResumeRepository) ListByUser(ctx context.Context, userID string) ([]model.Resume, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, file_name, file_path, parsed_text, created_at
		 FROM resumes WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []model.Resume
	for rows.Next() {
		var resume model.Resume
		if err := rows.Scan(&resume.ID, &resume.UserID, &resume.FileName, &resume.FilePath, &resume.ParsedText, &resume.CreatedAt); err != nil {
			return nil, err
		}
		resumes = append(resumes, resume)
	}

	if resumes == nil {
		resumes = []model.Resume{}
	}

	return resumes, rows.Err()
}

func (r *ResumeRepository) FindByID(ctx context.Context, id string) (*model.Resume, error) {
	resume := &model.Resume{}

	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, file_name, file_path, parsed_text, created_at
		 FROM resumes WHERE id = $1`,
		id,
	).Scan(&resume.ID, &resume.UserID, &resume.FileName, &resume.FilePath, &resume.ParsedText, &resume.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResumeNotFound
		}
		return nil, err
	}

	return resume, nil
}

func (r *ResumeRepository) UpdateParsedText(ctx context.Context, id string, parsedText string) error {
	result, err := r.db.Exec(ctx, `UPDATE resumes SET parsed_text = $2 WHERE id = $1`, id, parsedText)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrResumeNotFound
	}
	return nil
}

func (r *ResumeRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM resumes WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrResumeNotFound
	}

	return nil
}
