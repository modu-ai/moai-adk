package cli

// Regression tests for the repair after the audit of db0a69698a (card issue-1948), through the real `moai af` verbs,
// a real ledger and real git. Values a write hook refuses are assembled at run time (afD1Words, afD1Value).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// afRepairLand runs `af land` with the gate on the standard queued card and returns the verb's JSON line.
func afRepairLand(t *testing.T, db, repo, gate string) (outcome, reason, stdout string) {
	t.Helper()
	out, _, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", gate)
	if err != nil {
		t.Fatalf("the verb failed: %v (stdout %.200q)", err, out)
	}
	var res struct{ Outcome, Reason string }
	if jerr := json.Unmarshal([]byte(out), &res); jerr != nil {
		t.Fatalf("stdout %.200q is not the verb's JSON line", out)
	}
	return res.Outcome, res.Reason, out
}

// N23: gate-logs is a symbolic link. The verb must not follow it: no log in the target, its mode unchanged, and the
// failure is recorded with a reason that says logging failed.
func TestAFLand_ALogDirectoryThatIsALinkIsRefusedAndItsTargetIsLeftAlone(t *testing.T) {
	afSkipWithoutSh(t)
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges on native Windows")
	}
	db, repo, _, st, _ := afQueuedCard(t)
	target := filepath.Join(t.TempDir(), "elsewhere")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(filepath.Dir(db), "gate-logs")); err != nil {
		t.Fatal(err)
	}
	outcome, reason, _ := afRepairLand(t, db, repo, "echo hello; exit 1")
	if outcome != "failed" || !strings.Contains(reason, "logging failed") {
		t.Errorf("outcome %q reason %.200q: want a failure that says logging failed", outcome, reason)
	}
	fi, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o755 {
		t.Errorf("the link's target has mode %v, want it untouched at 0755", fi.Mode().Perm())
	}
	if entries, _ := os.ReadDir(target); len(entries) != 0 {
		t.Errorf("the link's target holds %v, want nothing", entries)
	}
	if rows, err := st.Queue(context.Background()); err != nil || len(rows) != 1 || rows[0].FailReason == "" {
		t.Errorf("the failure was not recorded: %+v %v", rows, err)
	}
}
