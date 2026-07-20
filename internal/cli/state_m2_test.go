package cli

// SPEC-CLI-TUX-V3-005 M2 characterization tests for the state.go Printer migration.
//
// These tests pin the EXACT stdout bytes of the Data() output paths (JSON and the
// composed human block) to the golden master captured from the pre-migration code
// (runStateDump/runShowBlocker/printPhaseStateHuman emitted via bare fmt.Print*).
// Byte-identity holds because Printer.Data in ModePlain performs fmt.Fprintln(w, v)
// — identical to the former fmt.Println(v) on the same writer.
//
// The three human status messages ("No checkpoint...", "No blockers found",
// "No outstanding blockers found") were re-routed from stdout to stderr by the
// migration (Printer.Info writes to stderr). Those subtests assert the message is
// now on stderr and ABSENT from stdout (the load-bearing channel change), without
// over-coupling to Printer.Info's status glyph.
//
// Non-parallel: t.Chdir is process-global.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/session"
)

// mkdirAll creates a directory with default perms (test-setup helper).
func mkdirAll(path string) error { return os.MkdirAll(path, 0755) }

// m2FixedTime is the deterministic timestamp embedded in every golden fixture.
var m2FixedTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// m2SetupState creates a temp .moai/state, chdirs the test there (so findStateDir
// resolves it), and writes the golden checkpoint. The on-disk state is identical
// to the one used to capture the pre-migration golden master.
func m2SetupState(t *testing.T) {
	t.Helper()
	base := t.TempDir()
	stateDir := filepath.Join(base, ".moai", "state")
	store := session.NewFileSessionStore(stateDir, 3600*time.Second)
	ckptState := session.PhaseState{
		Phase:  session.PhaseRun,
		SPECID: "SPEC-GOLDEN-001",
		Checkpoint: &session.RunCheckpoint{
			SPECID:        "SPEC-GOLDEN-001",
			Status:        "pass",
			Harness:       "standard",
			TestsTotal:    10,
			TestsPassed:   10,
			FilesModified: 3,
		},
		BlockerRpt: &session.BlockerReport{
			Kind:      "missing_input",
			Resolved:  true,
			Timestamp: m2FixedTime,
			Phase:     session.PhaseRun,
			SPECID:    "SPEC-GOLDEN-001",
		},
		UpdatedAt:  m2FixedTime,
		Provenance: session.ProvenanceTag{Source: "session", Origin: "cli", Loaded: m2FixedTime},
	}
	if err := store.Checkpoint(ckptState); err != nil {
		t.Fatalf("Checkpoint: %v", err)
	}
	t.Chdir(base)
}

// m2Printer returns a deterministic ModePlain Printer writing to the two buffers.
func m2Printer(stdout, stderr *bytes.Buffer) printer.Printer {
	return printer.New(printer.WithWriters(stdout, stderr), printer.WithMode(printer.ModePlain))
}

// wantJSON is the exact stdout of `state dump run SPEC-GOLDEN-001 --format json`
// (golden master). Top-level fields are alphabetical because PhaseState.MarshalJSON
// round-trips the checkpoint through map[string]any.
const wantJSON = `{
  "checkpoint": {
    "files_modified": 3,
    "harness": "standard",
    "spec_id": "SPEC-GOLDEN-001",
    "status": "pass",
    "tests_passed": 10,
    "tests_total": 10
  },
  "phase": "run",
  "spec_id": "SPEC-GOLDEN-001",
  "blocker_report": {
    "kind": "missing_input",
    "message": "",
    "requested_action": "",
    "provenance": {
      "source": "",
      "origin": "",
      "loaded": "0001-01-01T00:00:00Z"
    },
    "resolved": true,
    "timestamp": "2026-01-01T00:00:00Z",
    "phase": "run",
    "spec_id": "SPEC-GOLDEN-001"
  },
  "updated_at": "2026-01-01T00:00:00Z",
  "provenance": {
    "source": "session",
    "origin": "cli",
    "loaded": "2026-01-01T00:00:00Z"
  }
}
`

// wantHuman is the exact stdout of the composed human-format block (golden master).
// The checkpoint sub-block marshals the concrete RunCheckpoint struct (declaration
// order), indented with a 2-space prefix via MarshalIndent(_, "  ", "  ").
const wantHuman = `Phase:     run
SPEC ID:   SPEC-GOLDEN-001
Updated:   2026-01-01T00:00:00Z
Provenance: source=session origin=cli
Blocker:   kind=missing_input resolved=true
Checkpoint:
  {
    "spec_id": "SPEC-GOLDEN-001",
    "status": "pass",
    "harness": "standard",
    "tests_total": 10,
    "tests_passed": 10,
    "files_modified": 3
  }
`

func TestStateM2_JSON_StdoutByteIdentical(t *testing.T) {
	m2SetupState(t)
	var stdout, stderr bytes.Buffer
	p := m2Printer(&stdout, &stderr)
	if err := runStateDump(p, "run", "SPEC-GOLDEN-001", "json", true); err != nil {
		t.Fatalf("runStateDump json: %v", err)
	}
	if got := stdout.String(); got != wantJSON {
		t.Errorf("JSON stdout mismatch\n--- want (%d bytes) ---\n%q\n--- got (%d bytes) ---\n%q", len(wantJSON), wantJSON, len(got), got)
	}
	if stderr.Len() != 0 {
		t.Errorf("JSON path wrote %d bytes to stderr (expected none): %q", stderr.Len(), stderr.String())
	}
}

func TestStateM2_Human_StdoutByteIdentical(t *testing.T) {
	m2SetupState(t)
	var stdout, stderr bytes.Buffer
	p := m2Printer(&stdout, &stderr)
	if err := runStateDump(p, "run", "SPEC-GOLDEN-001", "human", true); err != nil {
		t.Fatalf("runStateDump human: %v", err)
	}
	if got := stdout.String(); got != wantHuman {
		t.Errorf("Human stdout mismatch\n--- want (%d bytes) ---\n%q\n--- got (%d bytes) ---\n%q", len(wantHuman), wantHuman, len(got), got)
	}
	if stderr.Len() != 0 {
		t.Errorf("Human path wrote %d bytes to stderr (expected none): %q", stderr.Len(), stderr.String())
	}
}

// TestStateM2_NoCheckpoint_StatusToStderr verifies the documented M2 channel
// re-routing: the "No checkpoint found" status moved stdout → stderr.
func TestStateM2_NoCheckpoint_StatusToStderr(t *testing.T) {
	base := t.TempDir()
	if err := mkdirAll(filepath.Join(base, ".moai", "state")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)
	var stdout, stderr bytes.Buffer
	p := m2Printer(&stdout, &stderr)
	if err := runStateDump(p, "run", "SPEC-NOEXIST-999", "json", false); err != nil {
		t.Fatalf("runStateDump noexist: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("status leaked to stdout (expected empty): %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "No checkpoint found for phase run, SPEC SPEC-NOEXIST-999") {
		t.Errorf("status message not on stderr; got stderr=%q", stderr.String())
	}
}

func TestStateM2_ShowBlocker_JSON_StdoutByteIdentical(t *testing.T) {
	base := t.TempDir()
	stateDir := filepath.Join(base, ".moai", "state")
	if err := mkdirAll(stateDir); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)
	store := session.NewFileSessionStore(stateDir, 3600*time.Second)
	blocker := session.BlockerReport{
		Kind:            "missing_input",
		Message:         "SPEC ID required",
		RequestedAction: "provide_spec_id",
		Provenance:      session.ProvenanceTag{Source: "session", Origin: "cli", Loaded: m2FixedTime},
		Resolved:        false,
		Timestamp:       m2FixedTime,
		Phase:           session.PhaseRun,
		SPECID:          "SPEC-BLOCK-GOLDEN",
	}
	if err := store.RecordBlocker(blocker); err != nil {
		t.Fatalf("RecordBlocker: %v", err)
	}
	const wantBlockerJSON = `{
  "kind": "missing_input",
  "message": "SPEC ID required",
  "requested_action": "provide_spec_id",
  "provenance": {
    "source": "session",
    "origin": "cli",
    "loaded": "2026-01-01T00:00:00Z"
  },
  "resolved": false,
  "timestamp": "2026-01-01T00:00:00Z",
  "phase": "run",
  "spec_id": "SPEC-BLOCK-GOLDEN"
}
`
	var stdout, stderr bytes.Buffer
	p := m2Printer(&stdout, &stderr)
	if err := runShowBlocker(p); err != nil {
		t.Fatalf("runShowBlocker: %v", err)
	}
	if got := stdout.String(); got != wantBlockerJSON {
		t.Errorf("blocker JSON stdout mismatch\n--- want ---\n%q\n--- got ---\n%q", wantBlockerJSON, got)
	}
}

func TestStateM2_NoBlockers_StatusToStderr(t *testing.T) {
	base := t.TempDir()
	if err := mkdirAll(filepath.Join(base, ".moai", "state")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)
	var stdout, stderr bytes.Buffer
	p := m2Printer(&stdout, &stderr)
	if err := runShowBlocker(p); err != nil {
		t.Fatalf("runShowBlocker empty: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("status leaked to stdout (expected empty): %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "No blockers found") {
		t.Errorf("status not on stderr; got=%q", stderr.String())
	}
}

func TestStateM2_NoOutstanding_StatusToStderr(t *testing.T) {
	base := t.TempDir()
	stateDir := filepath.Join(base, ".moai", "state")
	if err := mkdirAll(stateDir); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)
	store := session.NewFileSessionStore(stateDir, 3600*time.Second)
	resolved := session.BlockerReport{
		Kind:      "missing_input",
		Resolved:  true,
		Timestamp: m2FixedTime,
		Phase:     session.PhaseRun,
		SPECID:    "SPEC-RESOLVED",
	}
	if err := store.RecordBlocker(resolved); err != nil {
		t.Fatalf("RecordBlocker resolved: %v", err)
	}
	var stdout, stderr bytes.Buffer
	p := m2Printer(&stdout, &stderr)
	if err := runShowBlocker(p); err != nil {
		t.Fatalf("runShowBlocker resolved-only: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("status leaked to stdout (expected empty): %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "No outstanding blockers found") {
		t.Errorf("status not on stderr; got=%q", stderr.String())
	}
}
