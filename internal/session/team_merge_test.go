package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMergeTeamCheckpoints_HappyPath은 여러 에이전트의 checkpoint를 병합합니다.
// SPEC-V3R2-RT-004 AC-08: team-mode merge.
func TestMergeTeamCheckpoints_HappyPath(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	specID := "SPEC-TEAM-001"
	phase := PhaseRun

	// 에이전트별 checkpoint 파일 작성
	agentNames := []string{"agent-a", "agent-b"}
	for i, agent := range agentNames {
		state := PhaseState{
			Phase:  phase,
			SPECID: specID,
			Checkpoint: &RunCheckpoint{
				SPECID:        specID,
				Status:        "pass",
				Harness:       "standard",
				TestsTotal:    50,
				TestsPassed:   50 - i,
				FilesModified: i + 1,
			},
			UpdatedAt: time.Now(),
			Provenance: ProvenanceTag{
				Source: "session",
				Origin: fmt.Sprintf("agent-%s", agent),
				Loaded: time.Now(),
			},
		}

		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			t.Fatalf("marshal 실패: %v", err)
		}

		// 에이전트별 checkpoint 파일 경로: checkpoint-{phase}-{specID}-{agent}.json
		agentPath := filepath.Join(tempDir, fmt.Sprintf("checkpoint-%s-%s-%s.json", phase, specID, agent))
		if err := os.WriteFile(agentPath, data, 0644); err != nil {
			t.Fatalf("에이전트 checkpoint 쓰기 실패: %v", err)
		}
	}

	// 병합
	merged, err := store.MergeTeamCheckpoints(specID, phase, agentNames)
	if err != nil {
		t.Fatalf("MergeTeamCheckpoints() 실패: %v", err)
	}

	// 병합 결과 검증
	if merged.Phase != phase {
		t.Errorf("Phase = %v, want %v", merged.Phase, phase)
	}
	if merged.SPECID != specID {
		t.Errorf("SPECID = %v, want %v", merged.SPECID, specID)
	}
	if merged.Provenance.Source != "session" {
		t.Errorf("Provenance.Source = %v, want 'session'", merged.Provenance.Source)
	}

	rc, ok := merged.Checkpoint.(*RunCheckpoint)
	if !ok {
		t.Fatal("Checkpoint은 *RunCheckpoint이어야 함")
	}
	// 합산 검증: TestsTotal = 50+50=100, TestsPassed = 50+49=99, FilesModified = 1+2=3
	if rc.TestsTotal != 100 {
		t.Errorf("TestsTotal = %d, want 100", rc.TestsTotal)
	}
	if rc.TestsPassed != 99 {
		t.Errorf("TestsPassed = %d, want 99", rc.TestsPassed)
	}
	if rc.FilesModified != 3 {
		t.Errorf("FilesModified = %d, want 3", rc.FilesModified)
	}
}

// TestMergeTeamCheckpoints_BlockerBubble은 에이전트 중 하나에 미해결 blocker가 있으면
// 에러를 반환합니다 (REQ-051 bubble-mode).
func TestMergeTeamCheckpoints_BlockerBubble(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	specID := "SPEC-TEAM-BLOCKER"
	phase := PhaseRun

	// 정상 에이전트
	normalState := PhaseState{
		Phase:  phase,
		SPECID: specID,
		Checkpoint: &RunCheckpoint{
			SPECID:  specID,
			Status:  "pass",
			Harness: "standard",
		},
		UpdatedAt:  time.Now(),
		Provenance: ProvenanceTag{Source: "session", Origin: "agent-ok", Loaded: time.Now()},
	}
	data, _ := json.MarshalIndent(normalState, "", "  ")
	normalPath := filepath.Join(tempDir, fmt.Sprintf("checkpoint-%s-%s-agent-ok.json", phase, specID))
	if err := os.WriteFile(normalPath, data, 0644); err != nil {
		t.Fatalf("정상 에이전트 checkpoint 쓰기 실패: %v", err)
	}

	// blocker 있는 에이전트
	blockedState := PhaseState{
		Phase:  phase,
		SPECID: specID,
		BlockerRpt: &BlockerReport{
			Kind:            "error",
			Message:         "test failed",
			RequestedAction: "fix_test",
			Provenance:      ProvenanceTag{Source: "session", Origin: "agent-blocked", Loaded: time.Now()},
			Resolved:        false,
			Timestamp:       time.Now(),
		},
		UpdatedAt:  time.Now(),
		Provenance: ProvenanceTag{Source: "session", Origin: "agent-blocked", Loaded: time.Now()},
	}
	blockedData, _ := json.MarshalIndent(blockedState, "", "  ")
	blockedPath := filepath.Join(tempDir, fmt.Sprintf("checkpoint-%s-%s-agent-blocked.json", phase, specID))
	if err := os.WriteFile(blockedPath, blockedData, 0644); err != nil {
		t.Fatalf("blocker 에이전트 checkpoint 쓰기 실패: %v", err)
	}

	// 병합 시 blocker가 bubble-up 되어야 함
	_, err := store.MergeTeamCheckpoints(specID, phase, []string{"agent-ok", "agent-blocked"})
	if err == nil {
		t.Error("미해결 blocker가 있는 에이전트 병합 시 에러를 반환해야 함")
		return
	}
	if err != ErrBlockerOutstanding {
		t.Errorf("에러 = %v, want ErrBlockerOutstanding", err)
	}
}

// TestMergeTeamCheckpoints_MissingFile은 파일이 없는 에이전트가 있으면
// 에러를 반환합니다.
func TestMergeTeamCheckpoints_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	store := NewFileSessionStore(tempDir, 3600*time.Second)

	_, err := store.MergeTeamCheckpoints("SPEC-001", PhaseRun, []string{"nonexistent-agent"})
	if err == nil {
		t.Error("없는 에이전트 파일로 병합 시 에러를 반환해야 함")
	}
}

// TestProvenanceRoundTrip은 Provenance 필드가 JSON 마샬/언마샬 후에도 보존됩니다.
// SPEC-V3R2-RT-004 AC-07: Provenance round-trip.
func TestProvenanceRoundTrip(t *testing.T) {
	original := PhaseState{
		Phase:  PhaseRun,
		SPECID: "SPEC-PROV-001",
		Checkpoint: &RunCheckpoint{
			SPECID:  "SPEC-PROV-001",
			Status:  "pass",
			Harness: "thorough",
		},
		UpdatedAt: time.Now().Truncate(time.Second),
		Provenance: ProvenanceTag{
			Source: "user",
			Origin: "/path/to/config.yaml",
			Loaded: time.Now().Truncate(time.Second),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() 실패: %v", err)
	}

	var roundTripped PhaseState
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("Unmarshal() 실패: %v", err)
	}

	if roundTripped.Provenance.Source != original.Provenance.Source {
		t.Errorf("Provenance.Source = %q, want %q", roundTripped.Provenance.Source, original.Provenance.Source)
	}
	if roundTripped.Provenance.Origin != original.Provenance.Origin {
		t.Errorf("Provenance.Origin = %q, want %q", roundTripped.Provenance.Origin, original.Provenance.Origin)
	}
	if !roundTripped.Provenance.Loaded.Equal(original.Provenance.Loaded) {
		t.Errorf("Provenance.Loaded = %v, want %v", roundTripped.Provenance.Loaded, original.Provenance.Loaded)
	}
}
