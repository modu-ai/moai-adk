package constitution

import (
	"testing"
	"time"
)

// TestFrozenGuard_Check는 FrozenGuard의 동작을 검증한다.
func TestFrozenGuard_Check(t *testing.T) {
	guard := NewFrozenGuard()

	tests := []struct {
		name        string
		proposal    *AmendmentProposal
		currentZone Zone
		wantErr     error
	}{
		{
			name: "Evolvable zone은 통과",
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
			name: "Frozen zone은 Evidence 필수",
			proposal: &AmendmentProposal{
				RuleID:   "CONST-V3R2-001",
				Before:   "Old clause",
				After:    "New clause",
				Evidence: "",
			},
			currentZone: ZoneFrozen,
			wantErr: &ErrFrozenAmendment{
				RuleID: "CONST-V3R2-001",
				Reason: "Frozen zone rule 수정에는 Evidence(증거)가 필수이다. Frozen→Evolvable demotion 사유를 설명하라.",
			},
		},
		{
			name: "Frozen zone + Evidence는 통과 (demotion 가정)",
			proposal: &AmendmentProposal{
				RuleID:   "CONST-V3R2-001",
				Before:   "Old clause",
				After:    "New clause",
				Evidence: "이 rule은 더 이상 유효하지 않음. 패턴 변경으로 인해 제약 완화 필요.",
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

// TestAmendmentLog_Validate는 AmendmentLog 검증을 테스트한다.
func TestAmendmentLog_Validate(t *testing.T) {
	tests := []struct {
		name    string
		log     AmendmentLog
		wantErr bool
	}{
		{
			name: "유효한 log",
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
			name: "ID 비어있음",
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
			name: "RuleID 비어있음",
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
			name: "Clause 비어있음",
			log: AmendmentLog{
				ID:         "LEARN-20260428-001",
				RuleID:     "CONST-V3R2-008",
				ApprovedBy: "human",
				ApprovedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "ApprovedBy 비어있음",
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
			name: "ApprovedAt 비어있음",
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

// TestGenerateLogID는 로그 ID 생성을 테스트한다.
func TestGenerateLogID(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		now      time.Time
		lastLogs []AmendmentLog
		want     string
	}{
		{
			name:     "첫 번째 로그",
			now:      now,
			lastLogs: []AmendmentLog{},
			want:     "LEARN-20260428-001",
		},
		{
			name: "같은 날짜의 두 번째 로그",
			now: now,
			lastLogs: []AmendmentLog{
				{ID: "LEARN-20260428-001"},
			},
			want: "LEARN-20260428-002",
		},
		{
			name: "시퀀스 009 이후",
			now: now,
			lastLogs: []AmendmentLog{
				{ID: "LEARN-20260428-009"},
			},
			want: "LEARN-20260428-010",
		},
		{
			name: "다른 날짜의 첫 번째 로그",
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

// TestErrFrozenAmendment_Error는 에러 메시지를 테스트한다.
func TestErrFrozenAmendment_Error(t *testing.T) {
	err := &ErrFrozenAmendment{
		RuleID: "CONST-V3R2-001",
		Reason: "Evidence 없음",
	}
	want := "Frozen zone amendment 거부: CONST-V3R2-001 - Evidence 없음"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

// TestErrCanaryRejected_Error는 에러 메시지를 테스트한다.
func TestErrCanaryRejected_Error(t *testing.T) {
	err := &ErrCanaryRejected{
		RuleID:        "CONST-V3R2-008",
		ScoreDrop:     0.15,
		Threshold:     0.10,
		AffectedSpecs: []string{"SPEC-001", "SPEC-002"},
	}
	got := err.Error()
	expected := "Canary rejected: CONST-V3R2-008 점수 하락 0.15 > 임계값 0.10"
	if !contains(got, expected) {
		t.Errorf("Error() = %v, want contain %v", got, expected)
	}
}

// contains는 문자열 포함 여부를 확인한다.
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
