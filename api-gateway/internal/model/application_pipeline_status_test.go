package model

import (
	"errors"
	"testing"
)

func TestApplicationPipelineStatuses_containsAllLOV(t *testing.T) {
	want := map[string]struct{}{
		"draft":     {},
		"applied":   {},
		"replied":   {},
		"interview": {},
		"offer":     {},
		"rejected":  {},
		"ghosted":   {},
	}
	got := make(map[string]struct{})
	for _, s := range ApplicationPipelineStatuses {
		got[s] = struct{}{}
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d statuses, got %d: %v", len(want), len(got), ApplicationPipelineStatuses)
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("missing LOV value %q", k)
		}
	}
}

func TestValidateApplicationPipelineStatus(t *testing.T) {
	for _, s := range ApplicationPipelineStatuses {
		if err := ValidateApplicationPipelineStatus(s); err != nil {
			t.Errorf("ValidateApplicationPipelineStatus(%q) = %v; want nil", s, err)
		}
		if !IsValidApplicationPipelineStatus(s) {
			t.Errorf("IsValidApplicationPipelineStatus(%q) = false", s)
		}
	}

	tests := []struct {
		in      string
		wantErr error
	}{
		{"", ErrApplicationStatusRequired},
		{"pending", ErrInvalidApplicationPipelineStatus},
		{"Draft", ErrInvalidApplicationPipelineStatus},
		{" applied", ErrInvalidApplicationPipelineStatus},
		{"replied ", ErrInvalidApplicationPipelineStatus},
	}
	for _, tc := range tests {
		err := ValidateApplicationPipelineStatus(tc.in)
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("ValidateApplicationPipelineStatus(%q) error = %v; want %v", tc.in, err, tc.wantErr)
		}
	}
}
