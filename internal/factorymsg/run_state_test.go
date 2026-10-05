package factorymsg

// run_state_test.go — the shared tri-state accessor (SPEC-STALE-RUN-LABEL-001
// REQ-SRL-003): active / measured not-active / unavailable must stay apart,
// because the hook prescription gate acts on the first two and fails open on
// the third. Companion to the gate-level TestPrescriptionGateUnavailableFailsOpen
// in internal/hook, which is the behavior-level RED/GREEN evidence.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func seedRunStatus(t *testing.T, root, run, status string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close factory state: %v", err)
		}
	}()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: run, LeadSessionID: "lead", Backend: "test", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE runs SET status=? WHERE run_id=?`, status, run); err != nil {
		t.Fatal(err)
	}
}

func TestPrescriptionGateProbeRunStateTriState(t *testing.T) {
	t.Run("state renders for logs and notices", func(t *testing.T) {
		for state, want := range map[RunState]string{
			RunStateActive:      "active",
			RunStateNotActive:   "not-active",
			RunStateUnavailable: "unavailable",
		} {
			if got := state.String(); got != want {
				t.Errorf("RunState(%d).String() = %q, want %q", state, got, want)
			}
		}
	})
	t.Run("active run measures active", func(t *testing.T) {
		root := t.TempDir()
		seedRunStatus(t, root, "run-a", "active")
		state, detail, err := ProbeRunState(context.Background(), root, "run-a")
		if err != nil || state != RunStateActive || detail != "active" {
			t.Fatalf("ProbeRunState = (%v, %q, %v), want (RunStateActive, \"active\", nil)", state, detail, err)
		}
	})
	t.Run("retired run measures not-active with the raw status", func(t *testing.T) {
		root := t.TempDir()
		seedRunStatus(t, root, "run-r", "retired")
		state, detail, err := ProbeRunState(context.Background(), root, "run-r")
		if err != nil || state != RunStateNotActive || detail != "retired" {
			t.Fatalf("ProbeRunState = (%v, %q, %v), want (RunStateNotActive, \"retired\", nil)", state, detail, err)
		}
	})
	t.Run("unknown run measures not-active absent", func(t *testing.T) {
		root := t.TempDir()
		seedRunStatus(t, root, "run-other", "active")
		state, detail, err := ProbeRunState(context.Background(), root, "run-ghost")
		if err != nil || state != RunStateNotActive || detail != "absent" {
			t.Fatalf("ProbeRunState = (%v, %q, %v), want (RunStateNotActive, \"absent\", nil)", state, detail, err)
		}
	})
	t.Run("absent factory DB measures not-active", func(t *testing.T) {
		state, detail, err := ProbeRunState(context.Background(), t.TempDir(), "run-x")
		if err != nil || state != RunStateNotActive || detail != "absent" {
			t.Fatalf("ProbeRunState = (%v, %q, %v), want measured not-active (never a factory project)", state, detail, err)
		}
	})
	t.Run("corrupt factory DB measures unavailable", func(t *testing.T) {
		root := t.TempDir()
		// The garbage DB must land on the path the accessor actually reads
		// (homestate.FactoryDBPath), not on a guessed spelling of it.
		path, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("not a database"), 0o600); err != nil {
			t.Fatal(err)
		}
		state, _, err := ProbeRunState(context.Background(), root, "run-x")
		if state != RunStateUnavailable || err == nil {
			t.Fatalf("ProbeRunState = (%v, %v), want (RunStateUnavailable, non-nil) — measurement failure is not a verdict", state, err)
		}
	})
	t.Run("malformed run id measures not-active", func(t *testing.T) {
		root := t.TempDir()
		seedRunStatus(t, root, "run-ok", "active")
		state, detail, err := ProbeRunState(context.Background(), root, "bad id!")
		if err != nil || state != RunStateNotActive || detail != "absent" {
			t.Fatalf("ProbeRunState = (%v, %q, %v), want measured not-active for a malformed id", state, detail, err)
		}
	})
}

func TestPrescriptionGateActiveRunExists(t *testing.T) {
	t.Run("active run present", func(t *testing.T) {
		root := t.TempDir()
		seedRunStatus(t, root, "run-live", "active")
		seedRunStatus(t, root, "run-dead", "retired")
		active, err := ActiveRunExists(context.Background(), root)
		if err != nil || !active {
			t.Fatalf("ActiveRunExists = (%v, %v), want (true, nil)", active, err)
		}
	})
	t.Run("only retired runs", func(t *testing.T) {
		root := t.TempDir()
		seedRunStatus(t, root, "run-dead", "retired")
		active, err := ActiveRunExists(context.Background(), root)
		if err != nil || active {
			t.Fatalf("ActiveRunExists = (%v, %v), want (false, nil)", active, err)
		}
	})
	t.Run("absent DB measures false", func(t *testing.T) {
		active, err := ActiveRunExists(context.Background(), t.TempDir())
		if err != nil || active {
			t.Fatalf("ActiveRunExists = (%v, %v), want (false, nil)", active, err)
		}
	})
}

// SPEC-FACTORY-STALE-RUN-HEAL-001 REQ-SRH-001/-006: the read-only listing of
// active runs the notice-line table is computed from.
func TestActiveRunIDsAt(t *testing.T) {
	t.Run("lists only active runs, sorted", func(t *testing.T) {
		root := t.TempDir()
		seedRunStatus(t, root, "run-b", "active")
		seedRunStatus(t, root, "run-a", "active")
		seedRunStatus(t, root, "run-r", "retired")
		dbPath, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		got, err := ActiveRunIDsAt(context.Background(), dbPath)
		if err != nil {
			t.Fatalf("ActiveRunIDsAt: %v", err)
		}
		if want := []string{"run-a", "run-b"}; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("ActiveRunIDsAt = %q, want %q", got, want)
		}
	})
	t.Run("a project without a factory database lists nothing and creates nothing", func(t *testing.T) {
		root := t.TempDir()
		dbPath, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		got, err := ActiveRunIDsAt(context.Background(), dbPath)
		if err != nil || len(got) != 0 {
			t.Fatalf("ActiveRunIDsAt on an absent database = (%q, %v), want (empty, nil)", got, err)
		}
		if _, statErr := os.Stat(dbPath); !os.IsNotExist(statErr) {
			t.Errorf("the read-only listing created %s", dbPath)
		}
	})
	t.Run("an unreadable database is an error, not an empty answer", func(t *testing.T) {
		root := t.TempDir()
		dbPath, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dbPath, []byte("definitely not a sqlite database"), 0o600); err != nil {
			t.Fatal(err)
		}
		if got, err := ActiveRunIDsAt(context.Background(), dbPath); err == nil {
			t.Fatalf("ActiveRunIDsAt on a corrupt database = (%q, nil), want an error", got)
		}
	})
}

func TestValidRunID(t *testing.T) {
	for id, want := range map[string]bool{"runX": true, "tlwgk9": true, "run-1.a:b": true, "": false, "bad id!": false, "../x": false, "-lead": false} {
		if got := ValidRunID(id); got != want {
			t.Errorf("ValidRunID(%q) = %v, want %v", id, got, want)
		}
	}
}
