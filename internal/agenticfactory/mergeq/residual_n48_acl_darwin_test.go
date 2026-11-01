package mergeq

// N48 fourth bullet (audit 46dbcce0d4, residual classification): a log written before the ACL repair (N22) still carries the
// inherited ACL (group:everyone allow read), and a new write clears the ACL of the directory and of the new file only. The
// failure log may hold unmasked gate output, so a file in the log directory that is still readable by everyone is the case
// N22 meant to end. The expectation: after a write the files of the directory have no ACL.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResidualN48_ALogWrittenBeforeTheACLRepairEndsWithoutAnACLAfterTheNextWrite(t *testing.T) {
	parent := repairACLParent(t)
	ledgerDir := filepath.Join(parent, "af")
	if err := os.Mkdir(ledgerDir, 0o700); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(ledgerDir, "gate-logs")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(dir, failureLogName("t0", "oldcandidate", "gate"))
	if err := os.WriteFile(old, []byte("written by a build before the repair\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("chmod", "+a", "everyone allow read", old).CombinedOutput(); err != nil {
		t.Fatalf("chmod +a: %v %s", err, out)
	}
	if got := repairACLEntries(t, old); len(got) == 0 {
		t.Fatal("the fixture is wrong: the old log has no ACL")
	}
	if _, err := (&FailureLog{Dir: dir}).Write("t1", "c", "gate", []byte("output\n")); err != nil {
		t.Fatal(err)
	}
	got := repairACLEntries(t, old)
	t.Logf("ACL entries of the old log after the next write: %q", got)
	if len(got) != 0 {
		t.Errorf("the log written before the repair still has an ACL after the next write: %q", got)
	}
}
