package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/config"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/database"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/queue"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/repository"
	"github.com/anuragverma/ai-job-outreach/api-gateway/internal/sender"
)

const (
	pollInterval = 5 * time.Second
	maxRetries   = 3
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("worker: failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("worker: database connected")

	rq, err := queue.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		log.Fatalf("worker: failed to connect to redis: %v", err)
	}
	defer rq.Close()
	log.Println("worker: redis connected")

	emailRepo := repository.NewEmailRepository(db)
	appRepo := repository.NewApplicationRepository(db)
	mailer := sender.NewSMTPSender(cfg.SMTP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("worker: polling every %s (max retries: %d)", pollInterval, maxRetries)

	for {
		select {
		case <-sigCh:
			log.Println("worker: shutting down...")
			return
		default:
		}

		emailID, err := rq.ClaimDue(ctx)
		if err != nil {
			log.Printf("worker: redis claim error: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		if emailID == "" {
			time.Sleep(pollInterval)
			continue
		}

		log.Printf("worker: processing email %s", emailID)
		processEmail(ctx, emailID, emailRepo, appRepo, mailer, rq)
	}
}

func processEmail(
	ctx context.Context,
	emailID string,
	emailRepo *repository.EmailRepository,
	appRepo *repository.ApplicationRepository,
	mailer *sender.SMTPSender,
	rq *queue.RedisQueue,
) {
	email, err := emailRepo.FindByID(ctx, emailID)
	if err != nil {
		log.Printf("worker: email %s not found in DB, skipping: %v", emailID, err)
		return
	}

	if email.Status != "scheduled" {
		log.Printf("worker: email %s status is '%s', not 'scheduled' — skipping", emailID, email.Status)
		return
	}

	app, err := appRepo.FindByID(ctx, email.ApplicationID)
	if err != nil {
		log.Printf("worker: application %s not found for email %s: %v", email.ApplicationID, emailID, err)
		handleRetry(ctx, emailID, emailRepo, rq)
		return
	}

	if app.RecruiterEmail == "" {
		log.Printf("worker: no recruiter_email on application %s, marking failed", app.ID)
		_ = emailRepo.MarkFailed(ctx, emailID)
		return
	}

	err = mailer.Send(app.RecruiterEmail, email.Subject, email.Body)
	if err != nil {
		log.Printf("worker: SMTP send failed for email %s: %v", emailID, err)
		handleRetry(ctx, emailID, emailRepo, rq)
		return
	}

	now := time.Now()
	if err := emailRepo.MarkSent(ctx, emailID, now); err != nil {
		log.Printf("worker: failed to mark email %s as sent: %v", emailID, err)
	}

	_ = appRepo.UpdateStatus(ctx, app.ID, "applied")

	log.Printf("worker: email %s sent to %s, application %s -> applied", emailID, app.RecruiterEmail, app.ID)
}

func handleRetry(ctx context.Context, emailID string, emailRepo *repository.EmailRepository, rq *queue.RedisQueue) {
	retryCount, err := emailRepo.IncrementRetry(ctx, emailID)
	if err != nil {
		log.Printf("worker: failed to increment retry for %s: %v", emailID, err)
		return
	}

	if retryCount >= maxRetries {
		log.Printf("worker: email %s exhausted %d retries, marking failed", emailID, maxRetries)
		_ = emailRepo.MarkFailed(ctx, emailID)
		return
	}

	// Exponential backoff: 30s, 90s, 270s ...
	backoff := time.Duration(30*(1<<(retryCount-1))) * time.Second
	nextTry := time.Now().Add(backoff)

	log.Printf("worker: re-enqueuing email %s (retry %d/%d) at %s", emailID, retryCount, maxRetries, nextTry.Format(time.RFC3339))
	_ = rq.Enqueue(ctx, emailID, nextTry)
}
