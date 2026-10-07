//go:build darwin

package runtime

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAppendProgressRecordPreservesACL (round-3 repair 3, F6) — the atomic
// rename-replace must not drop the original progress.md's access-control
// entries: a rename swaps the directory entry, so the replacement carries
// the temp file's ACL (none) while the pre-repair os.WriteFile preserved
// the original's. After the fix the replaced file carries the original's
// ACL. darwin-only: chmod +a is a macOS ACL verb.
func TestAppendProgressRecordPreservesACL(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("chmod", "+a", "group:_guest deny read", path).CombinedOutput(); err != nil {
		t.Skipf("cannot set an ACL here: %v (%s)", err, out)
	}
	acl := func() string {
		t.Helper()
		out, err := exec.Command("ls", "-le", path).Output()
		if err != nil {
			t.Fatalf("ls -le: %v", err)
		}
		return string(out)
	}
	before := acl()
	if !strings.Contains(before, "deny read") {
		t.Skipf("the ACL is not visible through ls -le here:\n%s", before)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	after := acl()
	if !strings.Contains(after, "deny read") {
		t.Fatalf("the original's ACL did not survive the atomic replace:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestAppendProgressRecordAclExactlyOriginal (consolidated item 2) — the
// temp is created inside the SPEC directory, so it inherits the PARENT's
// inherited ACL entries at creation; cp -p does not remove them when the
// original itself has no ACL. The replaced file must carry EXACTLY the
// original's ACL — no inherited over-grant — and the record lands.
func TestAppendProgressRecordAclExactlyOriginal(t *testing.T) {
	specDir := t.TempDir()
	// The original predates the directory ACL: it carries none.
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("chmod", "+a", "group:_guest allow read,file_inherit", specDir).CombinedOutput(); err != nil {
		t.Skipf("cannot set an inherited ACL on the directory: %v (%s)", err, out)
	}
	if before := aclOf(t, path); strings.Contains(before, "group:_guest") {
		t.Skipf("the original unexpectedly carries ACL entries:\n%s", before)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	after := aclOf(t, path)
	if strings.Contains(after, "group:_guest") {
		t.Fatalf("the replaced file GAINED the parent's inherited ACL entry the original lacked:\n%s", after)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "- new record") {
		t.Fatalf("the record did not land:\n%s", raw)
	}
}

// aclOf lists the file's ACL entries (ls -le) for exact-match assertions.
func aclOf(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("/bin/ls", "-le", path).Output()
	if err != nil {
		t.Fatalf("ls -le: %v", err)
	}
	return string(out)
}
