//go:build linux

package runtime

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// posixAclAccess is the xattr name Linux stores a file's ACCESS ACL in.
const posixAclAccess = "system.posix_acl_access"

// seedFileMetadata makes the temp file carry the original's metadata —
// the mode (a plain stat/chmod, applied FIRST: on Linux a chmod rewrites
// the POSIX ACL's mask, so the ACL must land after it), then the extended
// attributes through the xattr family, POSIX ACLs included (Linux stores
// them as the system.posix_acl_access attribute) — one Go-native
// mechanism with no external tool (round-4 class closure). A failure is
// an error and the caller aborts the replace — there is no mode-only
// fallback.
//
// F9: the temp is created in the original's directory, and os.CreateTemp
// INHERITS that directory's default ACL as the temp's access ACL. An
// original carrying only the minimal ACL (mode bits — which listxattr
// does NOT enumerate) must never keep the inherited one: the seeder
// therefore ALWAYS writes the original's effective access ACL — the
// copied extended ACL when one exists, otherwise the minimal ACL
// constructed from the mode — fully overwriting inheritance, so the
// replaced file is never wider than the original.
func seedFileMetadata(tmp, original string) error {
	info, err := os.Stat(original)
	if err != nil {
		return err
	}
	if err := os.Chmod(tmp, info.Mode().Perm()); err != nil {
		return err
	}
	names, err := listXattrs(original)
	if err != nil {
		return err
	}
	hasAcl := false
	for _, name := range names {
		if name == posixAclAccess {
			hasAcl = true
		}
		sz, err := unix.Getxattr(original, name, nil)
		if err != nil {
			return fmt.Errorf("read xattr %s: %w", name, err)
		}
		val := make([]byte, sz)
		if _, err := unix.Getxattr(original, name, val); err != nil {
			return fmt.Errorf("read xattr %s: %w", name, err)
		}
		if err := unix.Setxattr(tmp, name, val, 0); err != nil {
			return fmt.Errorf("set xattr %s: %w", name, err)
		}
	}
	if !hasAcl {
		if err := setMinimalAcl(tmp, info.Mode().Perm()); err != nil {
			return err
		}
	}
	// Ownership rides the contract (round-4 edge 6b): mode+xattr copy alone
	// would re-own the file to the process. A chown this process cannot
	// make is an error and the replace aborts.
	return preserveOwnership(tmp, original)
}

// setMinimalAcl writes the minimal POSIX access ACL — user_obj/group_obj/
// other entries derived from the mode — as system.posix_acl_access,
// REPLACING whatever inherited ACL the temp carries. The binary form is
// the kernel's: uint32 version (2), then per entry {uint16 tag, uint16
// perm, uint32 id}, little-endian; 0xFFFFFFFF is ACL_UNDEFINED_ID.
func setMinimalAcl(path string, mode os.FileMode) error {
	const (
		aclVersion2 = 2
		aclUserObj  = 1
		aclGroupObj = 4
		aclOtherObj = 16
		undefinedID = 0xFFFFFFFF
	)
	m := uint32(mode & 0o777)
	type entry struct {
		tag, perm uint16
		id        uint32
	}
	entries := []entry{
		{aclUserObj, uint16(m >> 6), undefinedID},
		{aclGroupObj, uint16(m>>3) & 0x7, undefinedID},
		{aclOtherObj, uint16(m) & 0x7, undefinedID},
	}
	b := make([]byte, 4+8*len(entries))
	binary.LittleEndian.PutUint32(b[0:4], aclVersion2)
	for i, e := range entries {
		off := 4 + 8*i
		binary.LittleEndian.PutUint16(b[off:], e.tag)
		binary.LittleEndian.PutUint16(b[off+2:], e.perm)
		binary.LittleEndian.PutUint32(b[off+4:], e.id)
	}
	if err := unix.Setxattr(path, posixAclAccess, b, 0); err != nil {
		return fmt.Errorf("write minimal access acl: %w", err)
	}
	return nil
}

// listXattrs returns the NUL-separated attribute name list of path, sizing
// the buffer through the ERANGE retry the xattr family specifies.
func listXattrs(path string) ([]string, error) {
	buf := make([]byte, 256)
	for {
		n, err := unix.Listxattr(path, buf)
		if err == unix.ERANGE {
			buf = make([]byte, len(buf)*2)
			continue
		}
		if err != nil {
			return nil, err
		}
		var names []string
		for _, name := range strings.Split(string(buf[:n]), "\x00") {
			if name != "" {
				names = append(names, name)
			}
		}
		return names, nil
	}
}
