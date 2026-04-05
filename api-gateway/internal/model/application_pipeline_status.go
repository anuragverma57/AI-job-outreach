package model

import "errors"

// Canonical application pipeline statuses (LOV). Keep string values in sync with the frontend.
var ApplicationPipelineStatuses = []string{
	"draft",
	"applied",
	"replied",
	"interview",
	"offer",
	"rejected",
	"ghosted",
}

var applicationPipelineStatusSet map[string]struct{}

func init() {
	applicationPipelineStatusSet = make(map[string]struct{}, len(ApplicationPipelineStatuses))
	for _, s := range ApplicationPipelineStatuses {
		applicationPipelineStatusSet[s] = struct{}{}
	}
}

var (
	ErrApplicationStatusRequired        = errors.New("status is required")
	ErrInvalidApplicationPipelineStatus = errors.New("invalid status")
)

// IsValidApplicationPipelineStatus reports whether s is one of the allowed LOV values.
func IsValidApplicationPipelineStatus(s string) bool {
	_, ok := applicationPipelineStatusSet[s]
	return ok
}

// ValidateApplicationPipelineStatus checks a non-empty LOV value for PATCH /applications/:id/status.
func ValidateApplicationPipelineStatus(s string) error {
	if s == "" {
		return ErrApplicationStatusRequired
	}
	if !IsValidApplicationPipelineStatus(s) {
		return ErrInvalidApplicationPipelineStatus
	}
	return nil
}
