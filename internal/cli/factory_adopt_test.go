package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// adoptFixture seeds one picked card (t1) with a run-phase progress record
// and one evidence file — the AC-FLA-004 Given.
func adoptFixture(t *testing.T) (string, string, string) {
	t.Helper()
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked)
	spec := "SPEC-ADOPT-EXAMPLE-001"
	if err := store.Mutate(func(r *kanban.BacklogRecord) error {
		r.Items[0].SpecID = &spec
		return nil
	}); err != nil {
		t.Fatalf("set spec id: %v", err)
	}
	specDir := filepath.Join(root, ".moai", "specs", spec)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	progress := "# progress — run-phase body marker\n\n## §E.2 Run-phase Evidence\n\nmilestone work recorded\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progress), 0o644); err != nil {
		t.Fatal(err)
	}
	evDir := filepath.Join(root, ".moai", "reports", "t1")
	if err := os.MkdirAll(evDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evDir, "evidence.md"), []byte("evidence bytes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, spec, progress
}

func sha256OfFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// AC-FLA-004: adopting a stalled owner's picked card reads progress.md and
// the evidence FIRST and reports the recorded phase — the briefing output IS
// the recorded read.
func TestAdoptBriefsFromRecordedProgressAndEvidenceBeforeWork(t *testing.T) {
	_, spec, progress := adoptFixture(t)
	laneEnv(t)
	out, _, err := runFactory(t, "handoff", "adopt", "--card", "t1")
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if !strings.Contains(out, progress) {
		t.Fatalf("adopt output must carry the progress.md content, got %q", out)
	}
	if !strings.Contains(out, spec) {
		t.Fatalf("adopt output must name the SPEC, got %q", out)
	}
	if !strings.Contains(out, "recorded phase: run") {
		t.Fatalf("adopt output must derive the recorded phase from the markers, got %q", out)
	}
	if !strings.Contains(out, "evidence.md") {
		t.Fatalf("adopt output must list the recorded evidence, got %q", out)
	}
}

// AC-FLA-004 precondition: adoption is only for a card in picked state.
func TestAdoptRefusesCardNotPicked(t *testing.T) {
	_, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued)
	laneEnv(t)
	_, _, err := runFactory(t, "handoff", "adopt", "--card", "t1")
	if err == nil || !strings.Contains(err.Error(), "not picked") {
		t.Fatalf("error = %v, want the picked-state refusal", err)
	}
}

// A card with no recorded progress and no evidence is a fresh pickup, not a
// resumption — adopt refuses rather than laundering a silent restart.
func TestAdoptRefusesWhenNothingRecorded(t *testing.T) {
	_, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked)
	laneEnv(t)
	_, _, err := runFactory(t, "handoff", "adopt", "--card", "t1")
	if err == nil || !strings.Contains(err.Error(), "nothing recorded") {
		t.Fatalf("error = %v, want the nothing-recorded refusal", err)
	}
}

// AC-FLA-005: the previous owner's progress.md and evidence stay
// byte-unchanged, and the resumption record is appended alongside them.
func TestAdoptAppendsResumptionRecordWithoutTouchingPriorRecords(t *testing.T) {
	root, _, _ := adoptFixture(t)
	laneEnv(t)
	progressPath := filepath.Join(root, ".moai", "specs", "SPEC-ADOPT-EXAMPLE-001", "progress.md")
	evidencePath := filepath.Join(root, ".moai", "reports", "t1", "evidence.md")
	progressBefore := sha256OfFile(t, progressPath)
	evidenceBefore := sha256OfFile(t, evidencePath)

	if _, _, err := runFactory(t, "handoff", "adopt", "--card", "t1"); err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if got := sha256OfFile(t, progressPath); got != progressBefore {
		t.Fatalf("progress.md changed during adopt (%s -> %s)", progressBefore, got)
	}
	if got := sha256OfFile(t, evidencePath); got != evidenceBefore {
		t.Fatalf("evidence.md changed during adopt (%s -> %s)", evidenceBefore, got)
	}
	resumePath := filepath.Join(root, ".moai", "reports", "t1", "resumption.jsonl")
	data, err := os.ReadFile(resumePath)
	if err != nil {
		t.Fatalf("read resumption record: %v", err)
	}
	var entry struct {
		Lane          string `json:"lane"`
		RecordedPhase string `json:"recorded_phase"`
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("parse resumption record %q: %v", data, err)
	}
	if entry.Lane != "lane-1" || entry.RecordedPhase != "run" {
		t.Fatalf("resumption entry = %+v, want lane-1/run", entry)
	}
}

// The recorded-phase derivation reads the sync close marker, not just run
// markers — a card already at sync resumes at sync, never back at plan.
func TestAdoptDerivesSyncPhaseFromCloseMarker(t *testing.T) {
	root, spec, _ := adoptFixture(t)
	progressPath := filepath.Join(root, ".moai", "specs", spec, "progress.md")
	if err := os.WriteFile(progressPath, []byte("## §E.2 Run-phase Evidence\n\n## §E.4 Sync-phase Audit-Ready Signal\n\nsync_commit_sha: \"abc123def\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	laneEnv(t)
	out, _, err := runFactory(t, "handoff", "adopt", "--card", "t1")
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if !strings.Contains(out, "recorded phase: sync") {
		t.Fatalf("adopt output = %q, want recorded phase sync", out)
	}
}
