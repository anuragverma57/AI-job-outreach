package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/model"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidFileType = errors.New("only PDF files are allowed")
	ErrFileTooLarge    = errors.New("file size exceeds 5MB limit")
	ErrResumeNotOwned  = errors.New("resume does not belong to this user")
)

const maxFileSize = 5 * 1024 * 1024 // 5MB

type ResumeService struct {
	resumeRepo *repository.ResumeRepository
	uploadDir  string
}

func NewResumeService(resumeRepo *repository.ResumeRepository, uploadDir string) *ResumeService {
	return &ResumeService{
		resumeRepo: resumeRepo,
		uploadDir:  uploadDir,
	}
}

func (s *ResumeService) Upload(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (*model.Resume, error) {
	if fileHeader.Size > maxFileSize {
		return nil, ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".pdf" {
		return nil, ErrInvalidFileType
	}

	userDir := filepath.Join(s.uploadDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Use UUID filename to prevent collisions and path traversal
	storedName := uuid.New().String() + ".pdf"
	destPath := filepath.Join(userDir, storedName)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	resume, err := s.resumeRepo.Create(ctx, userID, fileHeader.Filename, destPath)
	if err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to save resume record: %w", err)
	}

	return resume, nil
}

func (s *ResumeService) List(ctx context.Context, userID string) ([]model.Resume, error) {
	return s.resumeRepo.ListByUser(ctx, userID)
}

func (s *ResumeService) Delete(ctx context.Context, userID, resumeID string) error {
	resume, err := s.resumeRepo.FindByID(ctx, resumeID)
	if err != nil {
		return err
	}

	if resume.UserID != userID {
		return ErrResumeNotOwned
	}

	if err := s.resumeRepo.Delete(ctx, resumeID); err != nil {
		return err
	}

	// Clean up the file — best effort, don't fail if file is already gone
	os.Remove(resume.FilePath)

	return nil
}
