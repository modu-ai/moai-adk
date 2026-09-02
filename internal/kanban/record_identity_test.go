package kanban

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// preChangeRecord mirrors the Record schema exactly as it stood before the
// lane number and card identifier were added, so a fixture can be produced by
// MARSHALLING the pre-change struct rather than hand-authored. A hand-indented
// fixture would fail a correct implementation for reasons the requirement does
// not care about (acceptance.md AC-KRS-007(a)).
type preChangeRecord struct {
	SessionID       string `json:"session_id"`
	SpecID          string `json:"spec_id"`
	Role            string `json:"role,omitempty"`
	Backend         string `json:"backend"`
	EnteredAt       string `json:"entered_at"`
	DeepScanDir     string `json:"deepscan_dir"`
	VerifyRung      *Rung  `json:"verify_rung,omitempty"`
	VerifyReentries int    `json:"verify_reentries"`
}

// AC-KRS-004: the lane number is data BESIDE the role, never routed through
// WithRole — which admits only the known role set and would drop `lane-3`
// silently.
func TestLaneNumberIsRecordedDistinctlyFromRole(t *testing.T) {
	t.Parallel()

	lane := NewRecord("sess-lane", "", BackendClaude).WithRole(RoleLane).WithLane(3)
	if lane.Lane != 3 {
		t.Fatalf("lane number = %d, want 3", lane.Lane)
	}
	if lane.Role != RoleLane {
		t.Fatalf("role = %q, want %q", lane.Role, RoleLane)
	}

	lead := NewRecord("sess-lead", "", BackendClaude).WithRole(RoleLead)
	if lead.Lane != 0 {
		t.Fatalf("lead lane number = %d, want 0 (not a lane)", lead.Lane)
	}
}

// A non-positive lane number is not a lane: lanes number from 1, so the zero
// value stays the "not a lane" signal rather than being overwritten by junk.
func TestWithLaneIgnoresNonPositiveNumbers(t *testing.T) {
	t.Parallel()

	for _, n := range []int{0, -1} {
		rec := NewRecord("sess", "", BackendClaude).WithLane(n)
		if rec.Lane != 0 {
			t.Fatalf("WithLane(%d) stored %d, want 0", n, rec.Lane)
		}
	}
}

// AC-KRS-005: the card identifier is a field distinct from the SPEC id, and it
// round-trips through the writer.
func TestCardIdentifierRoundTripsDistinctFromSpecID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rec := NewRecord("sess-card", "", BackendClaude).WithRole(RoleLane).WithLane(2).WithCard("t207")
	if err := Write(root, rec); err != nil {
		t.Fatalf("Write: %v", err)
	}

	back, err := Read(root, "sess-card")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if back.CardID != "t207" {
		t.Fatalf("card id = %q, want %q", back.CardID, "t207")
	}
	if back.SpecID != "" {
		t.Fatalf("spec id = %q, want empty — the card id must not populate it", back.SpecID)
	}
	if back.Lane != 2 {
		t.Fatalf("lane = %d, want 2", back.Lane)
	}
}

// AC-KRS-007(b): a non-lane record carrying no card identifier encodes neither
// new key. This is what keeps such a record byte-compatible with a pre-change
// reader instead of gaining two zero-valued keys it never had.
func TestNewFieldsAreOmittedWhenEmpty(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rec := NewRecord("sess-plain", "SPEC-X", BackendGLM).WithRole(RoleLead)
	if err := Write(root, rec); err != nil {
		t.Fatalf("Write: %v", err)
	}

	raw, err := os.ReadFile(RecordPath(root, "sess-plain"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	for _, key := range []string{`"lane"`, `"card_id"`} {
		if strings.Contains(string(raw), key) {
			t.Fatalf("encoded record carries %s though it is empty:\n%s", key, raw)
		}
	}
}

// AC-KRS-007(a): a record produced by the PRE-CHANGE writer still reads, the
// new fields report as not recorded, and rewriting it leaves the pre-existing
// keys byte-identical to their input form.
func TestPreChangeRecordReadsAndRewritesByteIdentically(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	old := preChangeRecord{
		SessionID:       "sess-legacy",
		SpecID:          "SPEC-OLD",
		Role:            RoleLane,
		Backend:         BackendClaude,
		EnteredAt:       "2026-08-23T17:47:22Z",
		DeepScanDir:     "",
		VerifyReentries: 0,
	}
	encoded, err := json.MarshalIndent(old, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	encoded = append(encoded, '\n')

	path := RecordPath(root, old.SessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	back, err := Read(root, old.SessionID)
	if err != nil {
		t.Fatalf("Read of a pre-change record failed: %v", err)
	}
	if back.Lane != 0 || back.CardID != "" {
		t.Fatalf("pre-change record reported lane=%d card=%q, want not recorded", back.Lane, back.CardID)
	}
	if back.Role != RoleLane || back.SpecID != "SPEC-OLD" || back.Backend != BackendClaude {
		t.Fatalf("pre-change keys did not survive the read: %+v", back)
	}

	if err := Write(root, back); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	rewritten, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after rewrite: %v", err)
	}
	if !bytes.Equal(encoded, rewritten) {
		t.Fatalf("rewrite altered the pre-existing keys.\nbefore:\n%s\nafter:\n%s", encoded, rewritten)
	}
}
