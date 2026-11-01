package mergeq

// N22 (repair after the audit of db0a69698a): on macOS a directory with an inheritable ACL gives every file made in it
// an ACL of its own, and a 0600 mode does not remove that ACL ("-rw-------" with "group:everyone inherited allow
// read"). The failure log may hold unmasked output, so its file and its directory must end without an ACL.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repairACLEntries is the ACL entries ls prints under the permission line of path (none when there is no ACL).
func repairACLEntries(t *testing.T, path string) []string {
	t.Helper()
	out, err := exec.Command("ls", "-lde", path).Output()
	if err != nil {
		t.Fatalf("ls -lde %s: %v", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines[1:]
}

// repairACLParent is a directory whose ACL is inherited by everything made below it.
func repairACLParent(t *testing.T) string {
	t.Helper()
	parent, err := os.MkdirTemp("", "moai-af-acl-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(parent) })
	acl := "everyone allow read,execute,list,search,readattr,readextattr,readsecurity,file_inherit,directory_inherit"
	if out, err := exec.Command("chmod", "+a", acl, parent).CombinedOutput(); err != nil {
		t.Fatalf("chmod +a: %v %s", err, out)
	}
	return parent
}

func TestFailureLog_TheLogAndItsDirectoryEndWithoutAnInheritedACL(t *testing.T) {
	parent := repairACLParent(t)
	ledgerDir := filepath.Join(parent, "af")
	if err := os.Mkdir(ledgerDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := repairACLEntries(t, ledgerDir); len(got) == 0 {
		t.Fatal("the fixture is wrong: a directory made below the ACL parent inherited no ACL")
	}
	dir := filepath.Join(ledgerDir, "gate-logs")
	path, err := (&FailureLog{Dir: dir}).Write("t1", "c", "gate", []byte("output\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{dir, path} {
		if got := repairACLEntries(t, p); len(got) != 0 {
			t.Errorf("%s has an ACL: %q", p, got)
		}
	}
	// the directory existed with an inherited ACL before the first write: the write clears it too
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := repairACLEntries(t, dir); len(got) == 0 {
		t.Fatal("the fixture is wrong: the pre-made log directory inherited no ACL")
	}
	path, err = (&FailureLog{Dir: dir}).Write("t1", "c", "gate", []byte("output\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{dir, path} {
		if got := repairACLEntries(t, p); len(got) != 0 {
			t.Errorf("a directory that was there before: %s has an ACL: %q", p, got)
		}
	}
}
