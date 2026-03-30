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

type GenerateEmailRequest struct {
	Tone string `json:"tone"`
}

type UpdateEmailRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type ScheduleEmailRequest struct {
	SendAt       *time.Time `json:"send_at,omitempty"`
	DelaySeconds *int64     `json:"delay_seconds,omitempty"`
}
