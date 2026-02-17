package workflow

import (
	"errors"
	"testing"
)

func TestValidateSpecID(t *testing.T) {
	tests := []struct {
		name    string
		specID  string
		wantErr bool
	}{
		{"valid simple", "SPEC-ISSUE-1", false},
		{"valid multi-digit", "SPEC-ISSUE-123", false},
		{"valid large number", "SPEC-ISSUE-99999", false},
		{"empty", "", true},
		{"missing prefix", "ISSUE-123", true},
		{"wrong prefix", "SPEC-123", true},
		{"lowercase", "spec-issue-123", true},
		{"trailing text", "SPEC-ISSUE-123abc", true},
		{"leading space", " SPEC-ISSUE-123", true},
		{"trailing space", "SPEC-ISSUE-123 ", true},
		{"no number", "SPEC-ISSUE-", true},
		{"negative number", "SPEC-ISSUE--1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpecID(tt.specID)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ValidateSpecID(%q) = nil, want error", tt.specID)
				}
				if !errors.Is(err, ErrInvalidSPECID) {
					t.Errorf("error = %v, want ErrInvalidSPECID", err)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateSpecID(%q) = %v, want nil", tt.specID, err)
				}
			}
		})
	}
}
