package cli

// Regression tests for the repair after the audit of db0a69698a (card issue-1948), through the real `moai af` verbs,
// a real ledger and real git. Values a write hook refuses are assembled at run time (afD1Words, afD1Value).

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
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

// afRepairACLEntries is the ACL entries `ls -lde` prints under the permission line of path (none when it has no ACL).
func afRepairACLEntries(t *testing.T, path string) []string {
	t.Helper()
	out, err := exec.Command("ls", "-lde", path).Output()
	if err != nil {
		t.Fatalf("ls -lde %s: %v", path, err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")[1:]
}

// N22: the ledger lives below a directory with an inheritable ACL (TMPDIR points inside it, as a project under a
// shared directory would). The log of a failed gate may hold unmasked output, so the file and its directory must end
// without the ACL the 0600 mode does not remove.
func TestAFLand_TheLogFileEndsWithoutAnInheritedACL(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS inheritable ACLs")
	}
	afSkipWithoutSh(t)
	parent, err := os.MkdirTemp("", "moai-af-acl-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(parent) })
	acl := "everyone allow read,execute,list,search,readattr,readextattr,readsecurity,file_inherit,directory_inherit"
	if out, cerr := exec.Command("chmod", "+a", acl, parent).CombinedOutput(); cerr != nil {
		t.Fatalf("chmod +a: %v %s", cerr, out)
	}
	t.Setenv("TMPDIR", parent) // before the first t.TempDir(), which makes its root below TMPDIR
	db, repo, _, _, _ := afQueuedCard(t)
	if got := afRepairACLEntries(t, filepath.Dir(db)); len(got) == 0 {
		t.Fatal("the fixture is wrong: the ledger directory inherited no ACL")
	}
	_, reason, _ := afRepairLand(t, db, repo, "printf 'hello from the gate\\n'; exit 1")
	m := afD1LogPath.FindStringSubmatch(reason)
	if m == nil {
		t.Fatalf("the reason names no log file: %.300q", reason)
	}
	for _, p := range []string{filepath.Dir(m[1]), m[1]} {
		if got := afRepairACLEntries(t, p); len(got) != 0 {
			t.Errorf("%s has an ACL: %q", p, got)
		}
	}
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
