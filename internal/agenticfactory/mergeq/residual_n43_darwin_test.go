package mergeq

// N44 rest, the mutant N43-leftover-uncounted (audits 46dbcce0d4 and a36ea8c368, residual classification): the retention pass
// counts the bytes of a stale temp file it could NOT remove ("a leftover that stays still holds its bytes"), and the mutant that
// stopped counting them survived the whole suite, because no test makes a removal fail. On macOS a file flagged uchg cannot be
// removed, which makes one.

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestResidualN44_AStaleLeftoverThatCannotBeRemovedStillCountsTowardTheTotalCap(t *testing.T) {
	dir := t.TempDir() + "/gate-logs"
	l := &FailureLog{Dir: dir, TotalCap: 2 << 10}
	older, err := l.Write("c0", "s0", "gate", []byte(strings.Repeat(tempTestLine, 30))) // 510 bytes
	if err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(older, old, old); err != nil {
		t.Fatal(err)
	}
	stuck := leftoverTemp(t, dir, "stuck", 3000, 1000*time.Hour)
	if out, err := exec.Command("chflags", "uchg", stuck).CombinedOutput(); err != nil {
		t.Skipf("chflags uchg: %v %s", err, out)
	}
	t.Cleanup(func() { _ = exec.Command("chflags", "nouchg", stuck).Run() })
	// the logs alone (1020 bytes) fit the cap of 2048; with the leftover that stays (4020) they do not
	newer, err := l.Write("c1", "s1", "gate", []byte(strings.Repeat(tempTestLine, 30)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stuck); err != nil {
		t.Fatalf("setup: the leftover was removed although it is flagged immutable: %v", err)
	}
	if _, err := os.Stat(older); err == nil {
		t.Error("the older log stayed although the bytes of the leftover that cannot be removed put the directory over its cap")
	}
	if _, err := os.Stat(newer); err != nil {
		t.Errorf("the log just written is gone: %v", err)
	}
}
