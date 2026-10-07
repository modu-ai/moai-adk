//go:build linux

package runtime

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

// defaultAclBlob builds a POSIX default-ACL xattr value granting other
// users read+write (the permissive shape F9 protects against). The tags
// are the POSIX vocabulary — USER_OBJ 1, GROUP_OBJ 4, MASK 0x10, OTHER
// 0x20 (the default ACL is selected by the xattr NAME, not by the tag
// values; invalid tags make Setxattr fail with EINVAL and the fixture
// would silently skip on ACL-capable filesystems — gate round-39 item 2).
func defaultAclBlob(perm uint16) []byte {
	type entry struct {
		tag, perm uint16
		id        uint32
	}
	entries := []entry{
		{aclUserObj, 6, aclUndefined},
		{aclGroupObj, 4, aclUndefined},
		{aclMaskObj, perm, aclUndefined},
		{aclOtherObj, perm, aclUndefined},
	}
	b := make([]byte, 4+8*len(entries))
	binary.LittleEndian.PutUint32(b[0:4], uint32(aclVersion2))
	for i, e := range entries {
		off := 4 + 8*i
		binary.LittleEndian.PutUint16(b[off:], e.tag)
		binary.LittleEndian.PutUint16(b[off+2:], e.perm)
		binary.LittleEndian.PutUint32(b[off+4:], e.id)
	}
	return b
}

// decodeAclEntries parses a system.posix_acl_access/default blob into its
// (tag, perm) pairs — version checked, ids ignored.
func decodeAclEntries(t *testing.T, blob []byte) (tags, perms []uint16) {
	t.Helper()
	if len(blob) < 4 {
		t.Fatalf("acl blob too short: %d bytes", len(blob))
	}
	if v := binary.LittleEndian.Uint32(blob[0:4]); v != uint32(aclVersion2) {
		t.Fatalf("acl version %d, want 2", v)
	}
	n := (len(blob) - 4) / 8
	for i := 0; i < n; i++ {
		off := 4 + 8*i
		tags = append(tags, binary.LittleEndian.Uint16(blob[off:]))
		perms = append(perms, binary.LittleEndian.Uint16(blob[off+2:]))
	}
	return tags, perms
}

// TestAppendProgressRecordOverwritesInheritedDefaultAcl (sync-audit-5 F9)
// — CreateTemp inherits the parent directory's default ACL, so a replaced
// progress.md would carry the DIRECTORY's permissive access ACL instead of
// the original's (probe: UID 65534 denied before, allowed after). The
// seeder must ALWAYS write the original's effective access ACL — the
// minimal ACL from the mode when no extended ACL exists — fully
// overwriting inheritance. linux-only: POSIX default ACLs are a Linux
// mechanism; the decisive run is CI linux.
func TestAppendProgressRecordOverwritesInheritedDefaultAcl(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission checks — probe meaningless")
	}
	specDir := t.TempDir()
	// Permissive default ACL on the directory: every new file inherits a
	// group/mask/other shape granting rw beyond the mode.
	if err := unix.Setxattr(specDir, "system.posix_acl_default", defaultAclBlob(6), 0); err != nil {
		t.Skipf("filesystem does not support default ACLs here: %v", err)
	}
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	// Mode 0640: other users get NOTHING.
	if err := os.WriteFile(path, []byte(pre), 0o640); err != nil {
		t.Fatal(err)
	}
	// Precondition (not a skip target on ACL-capable filesystems): the
	// original INHERITED the permissive access ACL — it carries the mask
	// entry (0x10) and the permissive other (6), which is exactly the
	// over-grant F9 removes.
	buf := make([]byte, 128)
	n, err := unix.Getxattr(path, posixAclAccess, buf)
	if err != nil {
		t.Skipf("the filesystem did not apply the default ACL to the new file: %v", err)
	}
	inheritedTags, inheritedPerms := decodeAclEntries(t, buf[:n])
	hasMask := false
	for _, tag := range inheritedTags {
		if tag == uint16(aclMaskObj) {
			hasMask = true
		}
	}
	if !hasMask || !containsPerm(inheritedPerms, 6) {
		t.Skipf("the inherited ACL does not carry the permissive shape (tags %v perms %v)", inheritedTags, inheritedPerms)
	}

	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	n, err = unix.Getxattr(path, posixAclAccess, buf)
	if err != nil {
		// No access ACL on the result == the minimal-from-mode state the
		// kernel reports as absent when the file carries only mode bits —
		// also acceptable: it is equivalent to the original.
		t.Logf("no explicit acl_access on the result (minimal state): %v", err)
		return
	}
	tags, perms := decodeAclEntries(t, buf[:n])
	wantTags := []uint16{uint16(aclUserObj), uint16(aclGroupObj), uint16(aclOtherObj)}
	wantPerms := []uint16{6, 4, 0} // mode 0640 → rw-, r--, ---
	if len(tags) != len(wantTags) {
		t.Fatalf("replaced file's access ACL has %d entries (%v), want the 3-entry minimal %v — inherited entries leaked", len(tags), tags, wantTags)
	}
	for i := range wantTags {
		if tags[i] != wantTags[i] || perms[i] != wantPerms[i] {
			t.Fatalf("replaced entry %d is tag %d perm %o, want tag %d perm %o — inheritance was not fully overwritten", i, tags[i], perms[i], wantTags[i], wantPerms[i])
		}
	}
}

func containsPerm(perms []uint16, want uint16) bool {
	for _, p := range perms {
		if p == want {
			return true
		}
	}
	return false
}
