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
// users read+write (the permissive shape F9 protects against): default
// user_obj/group_obj/other entries plus a mask.
func defaultAclBlob(perm uint16) []byte {
	const (
		aclVersion2 = 2
		tagDefUser  = 0x8002
		tagDefGroup = 0x8008
		tagDefMask  = 0x8010
		tagDefOther = 0x8020
		undefinedID = 0xFFFFFFFF
	)
	type entry struct {
		tag, perm uint16
		id        uint32
	}
	entries := []entry{
		{tagDefUser, 6, undefinedID},
		{tagDefGroup, 6, undefinedID},
		{tagDefMask, 6, undefinedID},
		{tagDefOther, perm, undefinedID},
	}
	b := make([]byte, 4+8*len(entries))
	binary.LittleEndian.PutUint32(b[0:4], aclVersion2)
	for i, e := range entries {
		off := 4 + 8*i
		binary.LittleEndian.PutUint16(b[off:], e.tag)
		binary.LittleEndian.PutUint16(b[off+2:], e.perm)
		binary.LittleEndian.PutUint32(b[off+4:], e.id)
	}
	return b
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
	// Permissive default ACL on the directory: every new file inherits
	// group+other rw access.
	if err := unix.Setxattr(specDir, "system.posix_acl_default", defaultAclBlob(6), 0); err != nil {
		t.Skipf("filesystem does not support default ACLs here: %v", err)
	}
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	// Mode 0640: other users get NOTHING.
	if err := os.WriteFile(path, []byte(pre), 0o640); err != nil {
		t.Fatal(err)
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
	blob := buf[:n]
	if n != 4+8*3 {
		t.Fatalf("replaced file's access ACL has %d bytes, want 28 (the minimal 3-entry ACL — inheritance was overwritten)", n)
	}
	otherPerm := binary.LittleEndian.Uint16(blob[4+8*2+2:])
	if otherPerm != 0 {
		t.Fatalf("other entry perm %o, want 0 (mode 0640) — inherited permission leaked", otherPerm)
	}
}
