package runtime

import (
	"encoding/binary"
	"os"
)

// POSIX ACL tag constants (linux acl.h; the same vocabulary the
// system.posix_acl_access / system.posix_acl_default xattrs carry).
const (
	aclUserObj  uint16 = 0x01 // 1
	aclUser     uint16 = 0x02 // 2
	aclGroupObj uint16 = 0x04 // 4
	aclGroup    uint16 = 0x08 // 8
	aclMaskObj  uint16 = 0x10 // 16
	aclOtherObj uint16 = 0x20 // 32

	aclVersion2  uint32 = 2
	aclUndefined uint32 = 0xFFFFFFFF
)

// minimalAclBlob encodes the MINIMAL POSIX access ACL for the mode —
// user_obj/group_obj/other entries derived from the permission bits, no
// mask entry. This is the representation every mode-only file carries;
// the linux seeder writes it to OVERWRITE whatever default-ACL entries
// the temp inherited from its parent directory (sync-audit-5 F9).
func minimalAclBlob(mode os.FileMode) []byte {
	m := uint32(mode & 0o777)
	type entry struct {
		tag, perm uint16
	}
	entries := []entry{
		{aclUserObj, uint16(m >> 6)},
		{aclGroupObj, uint16(m>>3) & 0x7},
		{aclOtherObj, uint16(m) & 0x7},
	}
	b := make([]byte, 4+8*len(entries))
	binary.LittleEndian.PutUint32(b[0:4], aclVersion2)
	for i, e := range entries {
		off := 4 + 8*i
		binary.LittleEndian.PutUint16(b[off:], e.tag)
		binary.LittleEndian.PutUint16(b[off+2:], e.perm)
		binary.LittleEndian.PutUint32(b[off+4:], aclUndefined)
	}
	return b
}
