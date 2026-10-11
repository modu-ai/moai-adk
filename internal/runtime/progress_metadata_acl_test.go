package runtime

import (
	"encoding/binary"
	"os"
	"testing"
)

// TestMinimalAclBlobUsesPosixOtherTag (sync-audit-5 F10) — the minimal
// ACL's third entry is ACL_OTHER (0x20), NOT ACL_MASK (0x10): the buggy
// constant made the kernel reject every constructed minimal ACL with
// EINVAL, killing the §G durable carrier for every progress.md without
// extended ACLs on Linux. The blob decodes to exactly USER_OBJ(1),
// GROUP_OBJ(4), OTHER(0x20) with the mode's permission thirds — MASK(16)
// is correctly absent from a minimal ACL.
func TestMinimalAclBlobUsesPosixOtherTag(t *testing.T) {
	mode := os.FileMode(0o640)
	b := minimalAclBlob(mode)
	if len(b) != 4+8*3 {
		t.Fatalf("minimal acl blob %d bytes, want 28 (header + 3 entries)", len(b))
	}
	if v := binary.LittleEndian.Uint32(b[0:4]); v != 2 {
		t.Fatalf("acl version %d, want 2", v)
	}
	wantTags := []uint16{1, 4, 0x20} // USER_OBJ, GROUP_OBJ, OTHER
	wantPerms := []uint16{6, 4, 0}   // rw-, r--, --- for 0640
	for i, want := range wantTags {
		off := 4 + 8*i
		if got := binary.LittleEndian.Uint16(b[off:]); got != want {
			t.Fatalf("entry %d tag %d (0x%x), want %d (0x%x) — ACL_OTHER is 0x20, not ACL_MASK 0x10", i, got, got, want, want)
		}
		if got := binary.LittleEndian.Uint16(b[off+2:]); got != wantPerms[i] {
			t.Fatalf("entry %d perm %o, want %o", i, got, wantPerms[i])
		}
		if id := binary.LittleEndian.Uint32(b[off+4:]); id != 0xFFFFFFFF {
			t.Fatalf("entry %d id %d, want 0xFFFFFFFF (ACL_UNDEFINED_ID)", i, id)
		}
	}
	if aclOtherObj != 0x20 {
		t.Fatalf("aclOtherObj constant is 0x%x, want 0x20 (POSIX ACL_OTHER)", aclOtherObj)
	}
	if aclMaskObj != 0x10 {
		t.Fatalf("aclMaskObj constant is 0x%x, want 0x10 (POSIX ACL_MASK — valid only in extended ACLs)", aclMaskObj)
	}
}
