package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// t459 R1: the curator-machine ARMING TRANSITION.
//
// t280 (SPEC-INBOX-DRAIN-GAP-001) covers "marker present -> stand down" and
// "marker absent -> rotate" as two independent states. It does NOT cover the
// transition between them on one machine: a curator install whose
// .moai/state/lsel/ disappears. acceptance §B records that transition as an
// unguarded one; this test measures whether it is actually reachable.
func TestT459_MarkerRemovalArmsTheCap(t *testing.T) {
	root := t.TempDir() // /tmp-class isolation; never this project's tree
	inbox := LessonsInboxPath(root)
	if err := os.MkdirAll(filepath.Dir(inbox), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, ".moai", "state", "lsel")
	if err := os.MkdirAll(marker, 0o755); err != nil {
		t.Fatal(err)
	}
	// A curator machine also carries the companion drain offset.
	offset := filepath.Join(marker, "drain-offset.json")
	if err := os.WriteFile(offset, []byte(`{"offset":4200}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// Fill the inbox past the cap with UNREAD stubs.
	line := strings.Repeat("x", 250) + "\n"
	f, err := os.OpenFile(inbox, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	for written := 0; written < config.DefaultInboxMaxBytes+1024; written += len(line) {
		if _, err := f.WriteString(line); err != nil {
			t.Fatal(err)
		}
	}
	_ = f.Close()
	before, _ := os.Stat(inbox)
	t.Logf("STATE A (curator, marker present): inbox=%d bytes, cap=%d",
		before.Size(), config.DefaultInboxMaxBytes)

	// STATE A: marker present -> stand down.
	appendLessonsInboxStub(root, "test_fail:pkg:", "A", "test:pkg")
	if _, err := os.Stat(inbox + ".1"); err == nil {
		t.Fatalf("STATE A: rotated while the curator marker was present (stand-down broken)")
	}
	afterA, _ := os.Stat(inbox)
	t.Logf("STATE A result: no .1 archive, inbox=%d bytes -> stand-down HELD", afterA.Size())

	// TRANSITION: the marker disappears (state dir wiped / fresh worktree /
	// broken session-start wiring). Nothing else about the machine changes.
	if err := os.RemoveAll(marker); err != nil {
		t.Fatal(err)
	}

	// STATE B: same machine, same inbox, one more append.
	appendLessonsInboxStub(root, "test_fail:pkg:", "B", "test:pkg")
	gen1, err := os.Stat(inbox + ".1")
	if err != nil {
		t.Fatalf("STATE B: cap did NOT arm after marker removal — risk NOT reproduced: %v", err)
	}
	afterB, _ := os.Stat(inbox)
	t.Logf("STATE B result: cap ARMED — %d bytes moved to .1, live inbox now %d bytes",
		gen1.Size(), afterB.Size())

	// The harm: the drain reads only the live path (drain.sh --inbox <live>),
	// so every stub now in .1 is unreachable to it.
	liveLines := countLinesT459(t, inbox)
	archivedLines := countLinesT459(t, inbox+".1")
	t.Logf("HARM: drain-reachable stubs=%d, archived-unreachable stubs=%d",
		liveLines, archivedLines)
	if archivedLines == 0 {
		t.Fatalf("expected archived stubs, got 0")
	}
}

func countLinesT459(t *testing.T, p string) int {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		return 0
	}
	return strings.Count(string(b), "\n")
}
