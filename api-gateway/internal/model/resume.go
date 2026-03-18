package model

import "time"

type Resume struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"-"`
	ParsedText *string   `json:"parsed_text,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
