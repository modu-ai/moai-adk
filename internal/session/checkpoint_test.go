package session

import (
	"testing"
)

func TestCheckpointPhaseName(t *testing.T) {
	tests := []struct {
		name       string
		checkpoint Checkpoint
		expected   Phase
	}{
		{
			name:       "PlanCheckpoint",
			checkpoint: &PlanCheckpoint{},
			expected:   PhasePlan,
		},
		{
			name:       "RunCheckpoint",
			checkpoint: &RunCheckpoint{},
			expected:   PhaseRun,
		},
		{
			name:       "SyncCheckpoint",
			checkpoint: &SyncCheckpoint{},
			expected:   PhaseSync,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checkpoint.PhaseName(); got != tt.expected {
				t.Errorf("Checkpoint.PhaseName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPlanCheckpointFields(t *testing.T) {
	pc := &PlanCheckpoint{
		SPECID:       "SPEC-001",
		Status:       "approved",
		ResearchPath: "/research/SPEC-001",
	}

	if pc.SPECID != "SPEC-001" {
		t.Errorf("SPECID = %v, want SPEC-001", pc.SPECID)
	}
	if pc.Status != "approved" {
		t.Errorf("Status = %v, want approved", pc.Status)
	}
	if pc.ResearchPath != "/research/SPEC-001" {
		t.Errorf("ResearchPath = %v, want /research/SPEC-001", pc.ResearchPath)
	}
}

func TestRunCheckpointFields(t *testing.T) {
	rc := &RunCheckpoint{
		SPECID:        "SPEC-001",
		Status:        "pass",
		TestsTotal:    100,
		TestsPassed:   95,
		FilesModified: 12,
	}

	if rc.SPECID != "SPEC-001" {
		t.Errorf("SPECID = %v, want SPEC-001", rc.SPECID)
	}
	if rc.Status != "pass" {
		t.Errorf("Status = %v, want pass", rc.Status)
	}
	if rc.TestsTotal != 100 {
		t.Errorf("TestsTotal = %v, want 100", rc.TestsTotal)
	}
	if rc.TestsPassed != 95 {
		t.Errorf("TestsPassed = %v, want 95", rc.TestsPassed)
	}
	if rc.FilesModified != 12 {
		t.Errorf("FilesModified = %v, want 12", rc.FilesModified)
	}
}

func TestSyncCheckpointFields(t *testing.T) {
	sc := &SyncCheckpoint{
		SPECID:     "SPEC-001",
		PRNumber:   123,
		PRURL:      "https://github.com/modu-ai/moai-adk/pull/123",
		DocsSynced: true,
	}

	if sc.SPECID != "SPEC-001" {
		t.Errorf("SPECID = %v, want SPEC-001", sc.SPECID)
	}
	if sc.PRNumber != 123 {
		t.Errorf("PRNumber = %v, want 123", sc.PRNumber)
	}
	if sc.PRURL != "https://github.com/modu-ai/moai-adk/pull/123" {
		t.Errorf("PRURL = %v, want https://github.com/modu-ai/moai-adk/pull/123", sc.PRURL)
	}
	if !sc.DocsSynced {
		t.Errorf("DocsSynced = %v, want true", sc.DocsSynced)
	}
}

// SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — checkpoint.go Validate gap closure.
// Plan/Run/Sync Validate 함수의 모든 error 경로 (missing SPECID + invalid status +
// invalid harness) 를 명시적으로 exercise한다.

func TestPlanCheckpointValidate_MissingSPECID(t *testing.T) {
	pc := &PlanCheckpoint{SPECID: "", Status: "approved"}
	err := pc.Validate()
	if err == nil {
		t.Fatal("Validate() with empty SPECID should error")
	}
}

func TestPlanCheckpointValidate_InvalidStatus(t *testing.T) {
	pc := &PlanCheckpoint{SPECID: "SPEC-001", Status: "invalid-state"}
	err := pc.Validate()
	if err == nil {
		t.Fatal("Validate() with invalid status should error")
	}
}

func TestPlanCheckpointValidate_AllValidStatuses(t *testing.T) {
	for _, status := range []string{"approved", "draft", "rejected"} {
		status := status
		t.Run(status, func(t *testing.T) {
			pc := &PlanCheckpoint{SPECID: "SPEC-001", Status: status}
			if err := pc.Validate(); err != nil {
				t.Errorf("Validate(status=%s) should succeed, got: %v", status, err)
			}
		})
	}
}

func TestRunCheckpointValidate_MissingSPECID(t *testing.T) {
	rc := &RunCheckpoint{SPECID: "", Status: "pass", Harness: "standard"}
	err := rc.Validate()
	if err == nil {
		t.Fatal("Validate() with empty SPECID should error")
	}
}

func TestRunCheckpointValidate_InvalidStatus(t *testing.T) {
	rc := &RunCheckpoint{SPECID: "SPEC-001", Status: "bogus", Harness: "standard"}
	err := rc.Validate()
	if err == nil {
		t.Fatal("Validate() with invalid status should error")
	}
}

func TestRunCheckpointValidate_EmptyHarness(t *testing.T) {
	rc := &RunCheckpoint{SPECID: "SPEC-001", Status: "pass", Harness: ""}
	err := rc.Validate()
	if err == nil {
		t.Fatal("Validate() with empty harness should error")
	}
}

func TestRunCheckpointValidate_InvalidHarness(t *testing.T) {
	rc := &RunCheckpoint{SPECID: "SPEC-001", Status: "pass", Harness: "ultra"}
	err := rc.Validate()
	if err == nil {
		t.Fatal("Validate() with invalid harness should error")
	}
}

func TestRunCheckpointValidate_AllValidStatuses(t *testing.T) {
	for _, status := range []string{"pass", "fail", "partial"} {
		status := status
		t.Run(status, func(t *testing.T) {
			rc := &RunCheckpoint{SPECID: "SPEC-001", Status: status, Harness: "standard"}
			if err := rc.Validate(); err != nil {
				t.Errorf("Validate(status=%s) should succeed, got: %v", status, err)
			}
		})
	}
}

func TestRunCheckpointValidate_AllValidHarnesses(t *testing.T) {
	for _, harness := range []string{"minimal", "standard", "thorough"} {
		harness := harness
		t.Run(harness, func(t *testing.T) {
			rc := &RunCheckpoint{SPECID: "SPEC-001", Status: "pass", Harness: harness}
			if err := rc.Validate(); err != nil {
				t.Errorf("Validate(harness=%s) should succeed, got: %v", harness, err)
			}
		})
	}
}

func TestSyncCheckpointValidate_MissingSPECID(t *testing.T) {
	sc := &SyncCheckpoint{SPECID: ""}
	err := sc.Validate()
	if err == nil {
		t.Fatal("Validate() with empty SPECID should error")
	}
}

func TestSyncCheckpointValidate_Valid(t *testing.T) {
	sc := &SyncCheckpoint{SPECID: "SPEC-001", PRNumber: 1}
	if err := sc.Validate(); err != nil {
		t.Errorf("Validate(valid) should succeed, got: %v", err)
	}
}
