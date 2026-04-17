package model

import (
	"encoding/json"
	"time"
)

type Email struct {
	ID            string          `json:"id"`
	ApplicationID string          `json:"application_id"`
	Subject       string          `json:"subject"`
	Body          string          `json:"body"`
	Status        string          `json:"status"`
	MatchScore    *float64        `json:"match_score,omitempty"`
	KeyPoints     json.RawMessage `json:"key_points,omitempty"`
	Reasoning     *string         `json:"reasoning,omitempty"`
	ScheduledAt   *time.Time      `json:"scheduled_at,omitempty"`
	SentAt        *time.Time      `json:"sent_at,omitempty"`
	RetryCount    int             `json:"retry_count"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// GenerateEmailDraft persists a client-generated email (e.g. after SSE streaming) without calling the LLM again.
type GenerateEmailDraft struct {
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	MatchScore  float64  `json:"match_score"`
	KeyPoints   []string `json:"key_points"`
	Reasoning   string   `json:"reasoning"`
}

type GenerateEmailRequest struct {
	Tone  string               `json:"tone"`
	Draft *GenerateEmailDraft  `json:"draft,omitempty"`
}

type UpdateEmailRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type ScheduleEmailRequest struct {
	SendAt       *time.Time `json:"send_at,omitempty"`
	DelaySeconds *int64     `json:"delay_seconds,omitempty"`
}
