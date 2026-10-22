package homestate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), coverage for the changed functions. The per-function
// measurement (decision d-20261010T085601Z-991b) found OpenFactory and OpenFactoryBounded at 76.9%,
// and planTransition at 84.0%, below the 85% target. Each test below takes one failure branch that
// the profile reported as uncovered; none of them changes the behaviour under test.

// The record database cannot be opened: the file at the factory database path is not a database,
// so the schema statement fails and both open functions return the error.
func TestOpenFactoryRefusesAnUnopenableRecord(t *testing.T) {
	withRetainedFactoryFixture(t, func(fixtureRoot string) {
		var root string
		if fixtureRoot == "" {
			root = factorySandbox(t)
		} else {
			home := filepath.Join(fixtureRoot, "home")
			t.Setenv("HOME", home)
			t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
			t.Setenv("MOAI_CLAUDE_BIN", filepath.Join(home, "bin", "claude"))
			root = filepath.Join(fixtureRoot, "sandbox-project")
			if err := os.MkdirAll(root, 0o700); err != nil {
				t.Fatalf("create sandbox project root: %v", err)
			}
		}
		if err := EnsureProjectLayout(root); err != nil {
			t.Fatalf("EnsureProjectLayout: %v", err)
		}
		path, err := FactoryDBPath(root)
		if err != nil {
			t.Fatalf("FactoryDBPath: %v", err)
		}
		const corrupt = "this is not a sqlite database, not even close\n"
		if err := os.WriteFile(path, []byte(corrupt), 0o600); err != nil {
			t.Fatalf("write the corrupt record: %v", err)
		}
		if _, err := OpenFactory(root); err == nil {
			t.Fatal("OpenFactory on a corrupt record: err = nil, want the open to fail")
		}
		if _, err := OpenFactoryBounded(root, time.Second); err == nil {
			t.Fatal("OpenFactoryBounded on a corrupt record: err = nil, want the open to fail")
		}
		if got, err := os.ReadFile(path); err != nil || string(got) != corrupt {
			t.Fatalf("corrupt record changed after refusal: bytes=%q err=%v", got, err)
		}
	})
}

// The project layout cannot be created: the moai home sits under a regular file, so the layout step
// fails before the record is opened, and both open functions return the error.
func TestOpenFactoryRefusesAnUnlayoutableHome(t *testing.T) {
	root := factorySandbox(t)
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("a regular file, not a directory\n"), 0o600); err != nil {
		t.Fatalf("write the blocking file: %v", err)
	}
	t.Setenv("MOAI_HOME", filepath.Join(blocker, "moai-home"))
	if _, err := OpenFactory(root); err == nil {
		t.Fatal("OpenFactory with an unlayoutable home: err = nil, want the layout error")
	}
	if _, err := OpenFactoryBounded(root, time.Second); err == nil {
		t.Fatal("OpenFactoryBounded with an unlayoutable home: err = nil, want the layout error")
	}
}

// A lease acquisition whose worker heartbeat write is refused fails as a whole: the trigger refuses
// the UPDATE on workers, so the transition returns the write error and commits nothing, and the card
// stays assigned at its version.
func TestLeaseAcquisitionRefusesAFailedHeartbeatWrite(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")
	frPlace(t, db, Card{RunID: frRun, CardID: "acq", State: CardAssigned, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir})
	if _, err := db.DB.Exec(`CREATE TRIGGER refuse_worker_heartbeat BEFORE UPDATE ON workers BEGIN SELECT RAISE(ABORT, 'worker heartbeat refused'); END`); err != nil {
		t.Fatalf("create the refusing trigger: %v", err)
	}
	_, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "acq", To: CardLeased, ExpectedVersion: 1, Actor: "worker-1", Now: frNow})
	if err == nil || !strings.Contains(err.Error(), "worker heartbeat refused") {
		t.Fatalf("lease acquisition with a refused heartbeat write: err = %v, want the write refusal", err)
	}
	got, err := db.LoadCard(ctx, frRun, "acq")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != CardAssigned || got.Version != 1 || got.LeaseHolder != "" {
		t.Fatalf("card after the refused acquisition = %s v%d holder=%q, want assigned at v1 with no holder", got.State, got.Version, got.LeaseHolder)
	}
}

// The worker roster cannot be read during a lease acquisition: the table is gone, so the registered-
// owner check fails and the transition returns the read error before any write.
func TestLeaseAcquisitionRefusesAnUnreadableRoster(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")
	frPlace(t, db, Card{RunID: frRun, CardID: "acq", State: CardAssigned, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir})
	if _, err := db.DB.Exec(`DROP TABLE workers`); err != nil {
		t.Fatalf("drop the roster: %v", err)
	}
	if _, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "acq", To: CardLeased, ExpectedVersion: 1, Actor: "worker-1", Now: frNow}); err == nil {
		t.Fatal("lease acquisition with an unreadable roster: err = nil, want the read error")
	}
	got, err := db.LoadCard(ctx, frRun, "acq")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != CardAssigned || got.Version != 1 {
		t.Fatalf("card after the unreadable roster = %s v%d, want assigned at v1", got.State, got.Version)
	}
}
