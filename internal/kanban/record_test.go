package kanban

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// helper: a fully-populated record used by the round-trip assertions.
func fullRecord() *Record {
	rung := RungPrimary
	return &Record{
		SessionID:       "session-abc123",
		SpecID:          "SPEC-PLACEHOLDER",
		Backend:         BackendClaude,
		EnteredAt:       "2026-01-02T15:04:05Z",
		DeepScanDir:     ".moai/reports/security-deepscan-20260102-150405",
		VerifyRung:      &rung,
		VerifyReentries: 2,
	}
}

func TestWriteThenReadRoundTripsEveryField(t *testing.T) {
	root := t.TempDir()
	want := fullRecord()

	if err := Write(root, want); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	got, err := Read(root, want.SessionID)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if got.SessionID != want.SessionID {
		t.Errorf("SessionID = %q, want %q", got.SessionID, want.SessionID)
	}
	if got.SpecID != want.SpecID {
		t.Errorf("SpecID = %q, want %q", got.SpecID, want.SpecID)
	}
	if got.Backend != want.Backend {
		t.Errorf("Backend = %q, want %q", got.Backend, want.Backend)
	}
	if got.EnteredAt != want.EnteredAt {
		t.Errorf("EnteredAt = %q, want %q", got.EnteredAt, want.EnteredAt)
	}
	if got.DeepScanDir != want.DeepScanDir {
		t.Errorf("DeepScanDir = %q, want %q", got.DeepScanDir, want.DeepScanDir)
	}
	if got.VerifyReentries != want.VerifyReentries {
		t.Errorf("VerifyReentries = %d, want %d", got.VerifyReentries, want.VerifyReentries)
	}
	if got.VerifyRung == nil {
		t.Fatalf("VerifyRung = nil, want %q", RungPrimary)
	}
	if *got.VerifyRung != RungPrimary {
		t.Errorf("VerifyRung = %q, want %q", *got.VerifyRung, RungPrimary)
	}
}

func TestRecordPathIsSessionKeyedUnderStateTodo(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".moai", "state", "todo", "session-abc123.json")

	if got := RecordPath(root, "session-abc123"); got != want {
		t.Errorf("RecordPath = %q, want %q", got, want)
	}
}

// The on-disk key set is the cross-actor contract: the orchestrator fills
// deepscan_dir / verify_rung / verify_reentries after the launcher writes the
// first three, so a renamed key silently breaks a reader this package cannot see.
func TestWrittenJSONCarriesTheDocumentedKeys(t *testing.T) {
	root := t.TempDir()
	rec := fullRecord()
	if err := Write(root, rec); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	raw, err := os.ReadFile(RecordPath(root, rec.SessionID))
	if err != nil {
		t.Fatalf("reading written record: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("written record is not valid JSON: %v", err)
	}

	for _, key := range []string{
		"session_id", "spec_id", "backend", "entered_at",
		"deepscan_dir", "verify_rung", "verify_reentries",
	} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("written record is missing key %q; got keys %v", key, decoded)
		}
	}
}

func TestVerifyRungRoundTripsEveryRung(t *testing.T) {
	for _, rung := range []Rung{RungPrimary, RungFallback, RungDegraded} {
		t.Run(string(rung), func(t *testing.T) {
			root := t.TempDir()
			r := rung
			rec := fullRecord()
			rec.VerifyRung = &r

			if err := Write(root, rec); err != nil {
				t.Fatalf("Write returned error: %v", err)
			}
			got, err := Read(root, rec.SessionID)
			if err != nil {
				t.Fatalf("Read returned error: %v", err)
			}
			if got.VerifyRung == nil {
				t.Fatalf("VerifyRung = nil, want %q", rung)
			}
			if *got.VerifyRung != rung {
				t.Errorf("VerifyRung = %q, want %q", *got.VerifyRung, rung)
			}
		})
	}
}

// design.md §7: deepscan_dir / verify_rung / verify_reentries are written
// independently on a best-effort record, so a record carrying no rung at all is
// reachable. The schema must distinguish that from a rung recorded as empty,
// because the downstream suppression check is an allow-list over a *recorded*
// PRIMARY or FALLBACK — collapsing the two would rebuild the deny-list hazard.
func TestAbsentRungIsDistinguishableFromEmptyRung(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		root := t.TempDir()
		rec := fullRecord()
		rec.VerifyRung = nil

		if err := Write(root, rec); err != nil {
			t.Fatalf("Write returned error: %v", err)
		}
		got, err := Read(root, rec.SessionID)
		if err != nil {
			t.Fatalf("Read returned error: %v", err)
		}
		if got.VerifyRung != nil {
			t.Errorf("VerifyRung = %q, want nil (never recorded)", *got.VerifyRung)
		}
	})

	t.Run("recorded empty", func(t *testing.T) {
		root := t.TempDir()
		empty := Rung("")
		rec := fullRecord()
		rec.VerifyRung = &empty

		if err := Write(root, rec); err != nil {
			t.Fatalf("Write returned error: %v", err)
		}
		got, err := Read(root, rec.SessionID)
		if err != nil {
			t.Fatalf("Read returned error: %v", err)
		}
		if got.VerifyRung == nil {
			t.Fatal("VerifyRung = nil, want a non-nil pointer to the empty rung")
		}
		if *got.VerifyRung != "" {
			t.Errorf("VerifyRung = %q, want the empty rung", *got.VerifyRung)
		}
	})
}

func TestReadReturnsErrorWhenRecordIsAbsent(t *testing.T) {
	root := t.TempDir()

	if _, err := Read(root, "no-such-session"); err == nil {
		t.Fatal("Read on an absent record returned nil error, want an error")
	}
}

func TestReadReturnsErrorOnMalformedJSON(t *testing.T) {
	root := t.TempDir()
	path := RecordPath(root, "session-abc123")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("preparing fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("writing malformed fixture: %v", err)
	}

	if _, err := Read(root, "session-abc123"); err == nil {
		t.Fatal("Read on malformed JSON returned nil error, want an error")
	}
}

func TestWriteRejectsUnusableSessionIDs(t *testing.T) {
	cases := map[string]string{
		"empty":          "",
		"path separator": "a/b",
		"parent escape":  "..",
	}
	for name, sessionID := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			rec := fullRecord()
			rec.SessionID = sessionID

			if err := Write(root, rec); err == nil {
				t.Fatalf("Write with session id %q returned nil error, want an error", sessionID)
			}
		})
	}
}

// A record write must never block a launch. WriteBestEffort has no error return
// precisely so a launch path structurally cannot gate on it; this asserts it
// survives the failure Write reports.
func TestWriteBestEffortFailsOpenOnUnwritableStateDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory permissions do not deny writes on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses directory permissions")
	}

	root := t.TempDir()
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("preparing state directory: %v", err)
	}
	if err := os.Chmod(stateDir, 0o500); err != nil {
		t.Fatalf("making state directory unwritable: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(stateDir, 0o755) })

	rec := fullRecord()

	if err := Write(root, rec); err == nil {
		t.Error("Write into an unwritable state directory returned nil error, want an error")
	}

	// The fail-open guarantee: no panic, no value a caller could gate the launch on.
	WriteBestEffort(root, rec)
}

func TestNewRecordStampsAnRFC3339EnteredAt(t *testing.T) {
	rec := NewRecord("session-abc123", "SPEC-PLACEHOLDER", BackendGLM)

	if rec.SessionID != "session-abc123" {
		t.Errorf("SessionID = %q, want %q", rec.SessionID, "session-abc123")
	}
	if rec.SpecID != "SPEC-PLACEHOLDER" {
		t.Errorf("SpecID = %q, want %q", rec.SpecID, "SPEC-PLACEHOLDER")
	}
	if rec.Backend != BackendGLM {
		t.Errorf("Backend = %q, want %q", rec.Backend, BackendGLM)
	}
	if _, err := time.Parse(time.RFC3339, rec.EnteredAt); err != nil {
		t.Errorf("EnteredAt = %q, which does not parse as RFC3339: %v", rec.EnteredAt, err)
	}
	if rec.VerifyRung != nil {
		t.Errorf("VerifyRung = %q, want nil on a freshly-entered record", *rec.VerifyRung)
	}
	if rec.VerifyReentries != 0 {
		t.Errorf("VerifyReentries = %d, want 0", rec.VerifyReentries)
	}
}

// A chain with no SPEC identifier heads at plan-phase, so an empty spec_id is a
// legitimate state rather than a missing field.
func TestRecordWithoutSpecIDRoundTrips(t *testing.T) {
	root := t.TempDir()
	rec := NewRecord("session-abc123", "", BackendClaude)

	if err := Write(root, rec); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	got, err := Read(root, rec.SessionID)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if got.SpecID != "" {
		t.Errorf("SpecID = %q, want the empty identifier", got.SpecID)
	}
}
