package model

import "time"

type Application struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ResumeID       *string   `json:"resume_id,omitempty"`
	CompanyName    string    `json:"company_name"`
	Role           string    `json:"role"`
	RecruiterEmail string    `json:"recruiter_email,omitempty"`
	JobDescription string    `json:"job_description,omitempty"`
	JobLink        string    `json:"job_link,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateApplicationRequest struct {
	ResumeID       string `json:"resume_id"`
	CompanyName    string `json:"company_name"`
	Role           string `json:"role"`
	RecruiterEmail string `json:"recruiter_email"`
	JobDescription string `json:"job_description"`
	JobLink        string `json:"job_link"`
}

type UpdateApplicationRequest struct {
	CompanyName    *string `json:"company_name,omitempty"`
	Role           *string `json:"role,omitempty"`
	RecruiterEmail *string `json:"recruiter_email,omitempty"`
	JobDescription *string `json:"job_description,omitempty"`
	JobLink        *string `json:"job_link,omitempty"`
}

// UpdateApplicationStatusRequest is the body for PATCH /api/applications/:id/status.
type UpdateApplicationStatusRequest struct {
	Status string `json:"status"`
}
