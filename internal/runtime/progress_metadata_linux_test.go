//go:build linux

package runtime

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

// TestAclUnsupportedFamily (gate round-46 item 6d) — the ENOTSUP /
// EOPNOTSUPP family is the filesystem's "ACLs not supported here": the
// seeder treats it as skip-the-ACL-write (perms/ownership stand, record
// lands), never as a metadata-copy failure to abort on. Wrapped errors
// must classify the same way.
func TestAclUnsupportedFamily(t *testing.T) {
	if !aclUnsupported(unix.ENOTSUP) {
		t.Fatal("ENOTSUP is not classified as ACL-unsupported")
	}
	if !aclUnsupported(unix.EOPNOTSUPP) {
		t.Fatal("EOPNOTSUPP is not classified as ACL-unsupported")
	}
	wrapped := fmt.Errorf("write minimal access acl: %w", unix.EOPNOTSUPP)
	if !aclUnsupported(wrapped) {
		t.Fatal("a wrapped EOPNOTSUPP is not classified as ACL-unsupported")
	}
	if aclUnsupported(unix.EINVAL) {
		t.Fatal("EINVAL classified as ACL-unsupported — a real blob failure must abort (F10)")
	}
}

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
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	// The ORIGINAL is created FIRST — clean, mode-only, no inherited
	// entries (gate round-46 item 2: the dir default ACL is set AFTER, so
	// the regression exercises the TEMP's inheritance during the replace,
	// not the original's own creation-time inheritance).
	if err := os.WriteFile(path, []byte(pre), 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := unix.Getxattr(path, posixAclAccess, nil); err == nil {
		t.Skipf("the original unexpectedly carries an access ACL")
	}
	// Permissive default ACL on the directory: every NEW file (the temp
	// the replace creates) inherits a group/mask/other shape granting rw
	// beyond the mode.
	if err := unix.Setxattr(specDir, "system.posix_acl_default", defaultAclBlob(6), 0); err != nil {
		t.Skipf("filesystem does not support default ACLs here: %v", err)
	}

	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 128)
	n, err := unix.Getxattr(path, posixAclAccess, buf)
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
