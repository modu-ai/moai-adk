package constitution

import (
	"testing"
	"time"
)

// TestFrozenGuard_Check verifies the behavior of FrozenGuard.
func TestFrozenGuard_Check(t *testing.T) {
	guard := NewFrozenGuard()

	tests := []struct {
		name        string
		proposal    *AmendmentProposal
		currentZone Zone
		wantErr     error
	}{
		{
			name: "Evolvable zone passes",
			proposal: &AmendmentProposal{
				RuleID:   "CONST-V3R2-008",
				Before:   "Old clause",
				After:    "New clause",
				Evidence: "",
			},
			currentZone: ZoneEvolvable,
			wantErr:     nil,
		},
		{
			name: "Frozen zone requires Evidence",
			proposal: &AmendmentProposal{
				RuleID:   "CONST-V3R2-001",
				Before:   "Old clause",
				After:    "New clause",
				Evidence: "",
			},
			currentZone: ZoneFrozen,
			wantErr: &ErrFrozenAmendment{
				RuleID: "CONST-V3R2-001",
				Reason: "Frozen zone rule modification requires Evidence. Explain the Frozen→Evolvable demotion reason.",
			},
		},
		{
			name: "Frozen zone + Evidence passes (assuming demotion)",
			proposal: &AmendmentProposal{
				RuleID:   "CONST-V3R2-001",
				Before:   "Old clause",
				After:    "New clause",
				Evidence: "This rule is no longer valid. Constraint relaxation needed due to pattern change.",
			},
			currentZone: ZoneFrozen,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := guard.Check(tt.proposal, tt.currentZone)
			if tt.wantErr == nil && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("expected error %v, got nil", tt.wantErr)
			}
			if tt.wantErr != nil && err != nil {
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("error mismatch:\nwant: %v\ngot:  %v", tt.wantErr, err)
				}
			}
		})
	}
}

// TestAmendmentLog_Validate tests AmendmentLog validation.
func TestAmendmentLog_Validate(t *testing.T) {
	tests := []struct {
		name    string
		log     AmendmentLog
		wantErr bool
	}{
		{
			name: "Valid log",
			log: AmendmentLog{
				ID:            "LEARN-20260428-001",
				RuleID:        "CONST-V3R2-008",
				ZoneBefore:    ZoneEvolvable,
				ZoneAfter:     ZoneEvolvable,
				ClauseBefore:  "Old clause",
				ClauseAfter:   "New clause",
				CanaryVerdict: "passed",
				ApprovedBy:    "human",
				ApprovedAt:    time.Now(),
				RolledBack:    false,
			},
			wantErr: false,
		},
		{
			name: "Empty ID",
			log: AmendmentLog{
				RuleID:     "CONST-V3R2-008",
				ClauseBefore: "Old",
				ClauseAfter:  "New",
				ApprovedBy:   "human",
				ApprovedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Empty RuleID",
			log: AmendmentLog{
				ID:          "LEARN-20260428-001",
				ClauseBefore: "Old",
				ClauseAfter:  "New",
				ApprovedBy:   "human",
				ApprovedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Empty Clause",
			log: AmendmentLog{
				ID:         "LEARN-20260428-001",
				RuleID:     "CONST-V3R2-008",
				ApprovedBy: "human",
				ApprovedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Empty ApprovedBy",
			log: AmendmentLog{
				ID:           "LEARN-20260428-001",
				RuleID:       "CONST-V3R2-008",
				ClauseBefore: "Old",
				ClauseAfter:  "New",
				ApprovedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Empty ApprovedAt",
			log: AmendmentLog{
				ID:           "LEARN-20260428-001",
				RuleID:       "CONST-V3R2-008",
				ClauseBefore: "Old",
				ClauseAfter:  "New",
				ApprovedBy:   "human",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.log.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateLogID tests log ID generation.
func TestGenerateLogID(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		now      time.Time
		lastLogs []AmendmentLog
		want     string
	}{
		{
			name:     "First log",
			now:      now,
			lastLogs: []AmendmentLog{},
			want:     "LEARN-20260428-001",
		},
		{
			name: "Second log on same date",
			now: now,
			lastLogs: []AmendmentLog{
				{ID: "LEARN-20260428-001"},
			},
			want: "LEARN-20260428-002",
		},
		{
			name: "After sequence 009",
			now: now,
			lastLogs: []AmendmentLog{
				{ID: "LEARN-20260428-009"},
			},
			want: "LEARN-20260428-010",
		},
		{
			name: "First log on different date",
			now: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
			lastLogs: []AmendmentLog{
				{ID: "LEARN-20260428-009"},
			},
			want: "LEARN-20260429-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateLogID(tt.now, tt.lastLogs)
			if got != tt.want {
				t.Errorf("GenerateLogID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestErrFrozenAmendment_Error tests the error message.
func TestErrFrozenAmendment_Error(t *testing.T) {
	err := &ErrFrozenAmendment{
		RuleID: "CONST-V3R2-001",
		Reason: "No evidence",
	}
	want := "Frozen zone amendment rejected: CONST-V3R2-001 - No evidence"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

// TestErrCanaryRejected_Error tests the error message.
func TestErrCanaryRejected_Error(t *testing.T) {
	err := &ErrCanaryRejected{
		RuleID:        "CONST-V3R2-008",
		ScoreDrop:     0.15,
		Threshold:     0.10,
		AffectedSpecs: []string{"SPEC-001", "SPEC-002"},
	}
	got := err.Error()
	expected := "Canary rejected: CONST-V3R2-008 score drop 0.15 > threshold 0.10"
	if !contains(got, expected) {
		t.Errorf("Error() = %v, want contain %v", got, expected)
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
