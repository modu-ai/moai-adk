package session

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPhaseStateMarshalUnmarshal(t *testing.T) {
	original := PhaseState{
		Phase:  PhasePlan,
		SPECID: "SPEC-001",
		Checkpoint: &PlanCheckpoint{
			SPECID:       "SPEC-001",
			Status:       "approved",
			ResearchPath: "/research/SPEC-001",
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "session",
			Origin: "cli",
			Loaded: time.Now(),
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	// Unmarshal
	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	// Verify fields
	if unmarshaled.Phase != original.Phase {
		t.Errorf("Phase = %v, want %v", unmarshaled.Phase, original.Phase)
	}
	if unmarshaled.SPECID != original.SPECID {
		t.Errorf("SPECID = %v, want %v", unmarshaled.SPECID, original.SPECID)
	}

	// Verify checkpoint type and content
	pc, ok := unmarshaled.Checkpoint.(*PlanCheckpoint)
	if !ok {
		t.Fatal("Checkpoint is not *PlanCheckpoint")
	}
	if pc.Status != "approved" {
		t.Errorf("Checkpoint.Status = %v, want approved", pc.Status)
	}
	if pc.ResearchPath != "/research/SPEC-001" {
		t.Errorf("Checkpoint.ResearchPath = %v, want /research/SPEC-001", pc.ResearchPath)
	}
}

func TestPhaseStateMarshalUnmarshalRunCheckpoint(t *testing.T) {
	original := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-002",
		Checkpoint: &RunCheckpoint{
			SPECID:        "SPEC-002",
			Status:        "pass",
			TestsTotal:    100,
			TestsPassed:   95,
			FilesModified: 12,
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "project",
			Origin: "/path/to/config",
			Loaded: time.Now(),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	rc, ok := unmarshaled.Checkpoint.(*RunCheckpoint)
	if !ok {
		t.Fatal("Checkpoint is not *RunCheckpoint")
	}
	if rc.TestsTotal != 100 {
		t.Errorf("TestsTotal = %v, want 100", rc.TestsTotal)
	}
	if rc.TestsPassed != 95 {
		t.Errorf("TestsPassed = %v, want 95", rc.TestsPassed)
	}
}

func TestPhaseStateMarshalUnmarshalSyncCheckpoint(t *testing.T) {
	original := PhaseState{
		Phase:  PhaseSync,
		SPECID: "SPEC-003",
		Checkpoint: &SyncCheckpoint{
			SPECID:     "SPEC-003",
			PRNumber:   123,
			PRURL:      "https://github.com/modu-ai/moai-adk/pull/123",
			DocsSynced: true,
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "user",
			Origin: "manual_input",
			Loaded: time.Now(),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	sc, ok := unmarshaled.Checkpoint.(*SyncCheckpoint)
	if !ok {
		t.Fatal("Checkpoint is not *SyncCheckpoint")
	}
	if sc.PRNumber != 123 {
		t.Errorf("PRNumber = %v, want 123", sc.PRNumber)
	}
	if !sc.DocsSynced {
		t.Errorf("DocsSynced = %v, want true", sc.DocsSynced)
	}
}

func TestPhaseStateWithBlocker(t *testing.T) {
	original := PhaseState{
		Phase:  PhaseReview,
		SPECID: "SPEC-004",
		BlockerRpt: &BlockerReport{
			Kind:            "quality_gate",
			Message:         "Lint errors found",
			RequestedAction: "fix_lint",
			Resolved:        false,
		},
		UpdatedAt: time.Now(),
		Provenance: ProvenanceTag{
			Source: "hook",
			Origin: "pre-commit",
			Loaded: time.Now(),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled PhaseState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if unmarshaled.BlockerRpt == nil {
		t.Fatal("BlockerRpt should not be nil")
	}
	if unmarshaled.BlockerRpt.Kind != "quality_gate" {
		t.Errorf("BlockerRpt.Kind = %v, want quality_gate", unmarshaled.BlockerRpt.Kind)
	}
}

func TestProvenanceTag(t *testing.T) {
	tag := ProvenanceTag{
		Source: "project",
		Origin: "/path/to/config",
		Loaded: time.Now(),
	}

	data, err := json.Marshal(tag)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var unmarshaled ProvenanceTag
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if unmarshaled.Source != tag.Source {
		t.Errorf("Source = %v, want %v", unmarshaled.Source, tag.Source)
	}
	if unmarshaled.Origin != tag.Origin {
		t.Errorf("Origin = %v, want %v", unmarshaled.Origin, tag.Origin)
	}
}
