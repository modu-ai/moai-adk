package codexwiring

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func plantLock(t *testing.T, root string, body []byte) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(WiringLockRelPath))
	writeFile(t, path, string(body))
	return path
}

func lockBody(t *testing.T, pid int, start string) []byte {
	t.Helper()
	raw, err := json.Marshal(WiringLockPayload{PID: pid, ProcessStart: start})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// TestWiringLockOwnership covers the wiring lock's liveness rules
// (design §A.2): a live owner holds it, a dead owner's or a reused PID's lock
// is cleared, an unreadable lock is held until its grace passes, and a taken
// lock is released on release().
func TestWiringLockOwnership(t *testing.T) {
	self := homestate.CurrentProcessFingerprint()
	if self == "" {
		t.Fatal("own process identity not observable")
	}
	cases := []struct {
		name  string
		body  func(t *testing.T) []byte
		age   time.Duration
		taken bool
	}{
		{"live_owner_holds", func(t *testing.T) []byte { return lockBody(t, os.Getpid(), self) }, 0, false},
		{"dead_owner_cleared", func(t *testing.T) []byte { return lockBody(t, 999999999, "gone") }, 0, true},
		{"reused_pid_cleared", func(t *testing.T) []byte { return lockBody(t, os.Getpid(), "another-start") }, 0, true},
		{"malformed_fresh_held", func(t *testing.T) []byte { return []byte("{") }, 0, false},
		{"malformed_old_cleared", func(t *testing.T) []byte { return []byte("") }, 2 * malformedLockGrace, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := plantLock(t, root, tc.body(t))
			if tc.age > 0 {
				old := time.Now().Add(-tc.age)
				if err := os.Chtimes(path, old, old); err != nil {
					t.Fatal(err)
				}
			}
			release, err := acquireWiringLock(root)
			if tc.taken {
				if err != nil {
					t.Fatalf("lock not taken: %v", err)
				}
				release()
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("release left the lock: %v", err)
				}
				return
			}
			if !errors.Is(err, ErrWiringLockHeld) {
				t.Fatalf("err=%v want ErrWiringLockHeld", err)
			}
		})
	}
}

// TestWiringLockHeldChangesNothing covers the lock-held outcome of a pass
// (REQ-DHR-002): no wiring file is created or changed.
func TestWiringLockHeldChangesNothing(t *testing.T) {
	root := t.TempDir()
	plantLock(t, root, lockBody(t, os.Getpid(), homestate.CurrentProcessFingerprint()))
	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); !errors.Is(err, ErrWiringLockHeld) {
		t.Fatalf("err=%v want ErrWiringLockHeld", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex")); !os.IsNotExist(err) {
		t.Fatalf("a held lock still let the pass write: %v", err)
	}
}

// TestInspectJournalReadOnly covers the doctor view (REQ-DHR-004): staged
// entries and unreferenced temps are reported, the temp a staged entry names
// is not an orphan, and nothing is written, renamed, or deleted.
func TestInspectJournalReadOnly(t *testing.T) {
	root := t.TempDir()
	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatal(err)
	}
	staged := JournalEntry{ID: "x", Op: OpWrite, Path: HooksRelPath, PreHash: "a", PostHash: "b", Temp: tempPrefix + "named", State: JournalStaged}
	if err := saveJournal(root, []JournalEntry{staged, {ID: "y", Path: ConfigRelPath, State: JournalComplete}}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, ".codex", staged.Temp), "staged")
	writeFile(t, filepath.Join(root, ".codex", tempPrefix+"orphan"), "orphan")
	before := snapshotTree(t, root)
	r := InspectJournal(root)
	if after := snapshotTree(t, root); !sameMap(before, after) {
		t.Fatalf("InspectJournal changed the tree")
	}
	if r.Err != nil || len(r.Incomplete) != 1 || r.Incomplete[0].ID != "x" {
		t.Fatalf("incomplete=%+v err=%v", r.Incomplete, r.Err)
	}
	if len(r.OrphanTemps) != 1 || r.OrphanTemps[0] != ".codex/"+tempPrefix+"orphan" {
		t.Fatalf("orphans=%v", r.OrphanTemps)
	}
	if r.Command != RecoverCommand {
		t.Fatalf("command=%q", r.Command)
	}

	writeFile(t, filepath.Join(root, filepath.FromSlash(JournalRelPath)), "{not json")
	if r := InspectJournal(root); r.Err == nil {
		t.Fatal("an unreadable journal must be reported")
	}
}
