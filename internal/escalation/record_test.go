package escalation_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// sampleRecord returns an open contract-kind ownership-move record for card.
func sampleRecord(card string) escalation.Record {
	fp := escalation.Fingerprint(escalation.ClassOwnershipMove, "internal/bar/x.go", "internal/foo/**")
	return escalation.Record{
		SchemaVersion: escalation.RecordSchemaVersion,
		Card:          card,
		Spec:          "SPEC-A-001",
		Kind:          escalation.KindContract,
		Class:         escalation.ClassOwnershipMove,
		Fingerprint:   fp,
		ContractRef:   "contract.yaml:12",
		EscalateOn:    "ownership-move",
		Status:        escalation.StatusOpen,
		Occurrences:   1,
		HeadSHA:       escalationtest.HeadSHA,
		DetectedAt:    "2026-01-02T03:04:05Z",
		UpdatedAt:     "2026-01-02T03:04:05Z",
		NotObserved:   []string{},
		Observation:   "Write internal/bar/x.go",
		Options:       []string{"Widen ownership.write", "Revert the write"},
	}
}

// writeAt places a record's encoding at its path (the writer proper lands
// in a later milestone; this criterion's M1 half is the path and marking).
func writeAt(t *testing.T, path string, r escalation.Record) {
	t.Helper()
	data, err := r.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// AC-AE-021 path half (REQ-AE-018): a record lands at
// <worktree root>/.moai/reports/<card>/escalation/<class>-<fingerprint>.md,
// the queue store is untouched, and the card is needs-decision exactly while a
// contract or operational record is open.
func TestRecordPathAndQueueUntouched(t *testing.T) {
	t.Setenv(config.EnvHome, t.TempDir())
	w := escalationtest.NewWorktree(t, "t9001")

	queueDir := kanban.StateDirForRoot(w.Root)
	queueFile := filepath.Join(queueDir, "backlog.db")
	if err := os.MkdirAll(queueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	queueBytes := []byte("queue store fixture bytes\n")
	if err := os.WriteFile(queueFile, queueBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	r := sampleRecord("t9001")
	if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(r.Fingerprint) {
		t.Fatalf("fingerprint %q is not 16 lowercase hex characters", r.Fingerprint)
	}
	path := escalation.RecordPath(w.Root, "t9001", r.Class, r.Fingerprint, 0)
	want := filepath.Join(w.Root, ".moai", "reports", "t9001", "escalation", "ownership-move-"+r.Fingerprint+".md")
	if path != want {
		t.Fatalf("RecordPath = %q, want %q", path, want)
	}
	if got := escalation.RecordPath(w.Root, "t9001", r.Class, r.Fingerprint, 2); got != filepath.Join(filepath.Dir(want), "ownership-move-"+r.Fingerprint+"-2.md") {
		t.Errorf("re-trip path = %q", got)
	}

	nd, err := escalation.NeedsDecision(w.Root, "t9001")
	if err != nil || nd {
		t.Fatalf("NeedsDecision before any record = %v, %v; want false", nd, err)
	}

	writeAt(t, path, r)
	nd, err = escalation.NeedsDecision(w.Root, "t9001")
	if err != nil || !nd {
		t.Fatalf("NeedsDecision with an open contract record = %v, %v; want true", nd, err)
	}

	// Another card's open record never marks this card.
	other := sampleRecord("t9002")
	writeAt(t, escalation.RecordPath(w.Root, "t9002", other.Class, other.Fingerprint, 0), other)

	// Resolving it clears the marking.
	r.Status = escalation.StatusResolved
	r.Decider = "human"
	writeAt(t, path, r)
	nd, err = escalation.NeedsDecision(w.Root, "t9001")
	if err != nil || nd {
		t.Fatalf("NeedsDecision after resolve = %v, %v; want false", nd, err)
	}

	// An open operational record marks the card again.
	op := sampleRecord("t9001")
	op.Kind, op.Class, op.EscalateOn = escalation.KindOperational, escalation.ClassBudgetExceeded, ""
	op.Fingerprint = escalation.Fingerprint(escalation.ClassBudgetExceeded, "operations")
	op.ContractRef = "config:workflow.autonomy.escalation.budget_default.operations"
	opPath := escalation.RecordPath(w.Root, "t9001", op.Class, op.Fingerprint, 0)
	writeAt(t, opPath, op)
	nd, err = escalation.NeedsDecision(w.Root, "t9001")
	if err != nil || !nd {
		t.Fatalf("NeedsDecision with an open operational record = %v, %v; want true", nd, err)
	}
	if err := os.Remove(opPath); err != nil {
		t.Fatal(err)
	}

	// A revoke record never counts, even if it read as open.
	rv := sampleRecord("t9001")
	rv.Kind, rv.Class, rv.EscalateOn, rv.ContractRef = escalation.KindRevoke, escalation.ClassRevokeOperator, "", ""
	rv.Fingerprint = escalation.Fingerprint(escalation.ClassRevokeOperator, "operator")
	writeAt(t, escalation.RecordPath(w.Root, "t9001", rv.Class, rv.Fingerprint, 0), rv)
	nd, err = escalation.NeedsDecision(w.Root, "t9001")
	if err != nil || nd {
		t.Fatalf("NeedsDecision with only a revoke record open = %v, %v; want false", nd, err)
	}

	// The record round-trips through its frontmatter.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	back, err := escalation.ParseRecord(data)
	if err != nil {
		t.Fatalf("ParseRecord: %v", err)
	}
	if back.Fingerprint != r.Fingerprint || back.Status != escalation.StatusResolved || back.Decider != "human" || back.Class != r.Class {
		t.Errorf("round-trip = %+v", back)
	}
	for _, section := range []string{"## Observation", "## Options", "## Not observed"} {
		if !bytes.Contains(data, []byte(section)) {
			t.Errorf("record body lacks %q", section)
		}
	}

	got, err := os.ReadFile(queueFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, queueBytes) {
		t.Error("queue store file bytes changed")
	}
}
