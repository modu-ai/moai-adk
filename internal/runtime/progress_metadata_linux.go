//go:build linux

package runtime

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// posixAclAccess is the xattr name Linux stores a file's ACCESS ACL in.
const posixAclAccess = "system.posix_acl_access"

// seedFileMetadata makes the temp carry the original's metadata — the
// mode, the extended attributes, the POSIX ACL (extended, else the minimal
// ACL built from the mode — F9), and ownership — ALL through the temp's
// HELD descriptor (Fchmod/Fsetxattr/Fchown): fd-based, immune to the
// name swap (sync-audit-7 F13). The SOURCE side is read by path (the
// resolved original — documented). A failure is an error and the caller
// aborts the replace — there is no mode-only fallback; an ACL/xattr
// UNSUPPORTED filesystem keeps the chmod'd perms and ownership and lands
// the record (gate round-46 item 6d).
func seedFileMetadata(tmp *os.File, tmpPath, original string) error {
	info, err := os.Stat(original)
	if err != nil {
		return err
	}
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	names, err := listXattrs(original)
	if err != nil {
		return err
	}
	hasAcl := false
	fd := int(tmp.Fd())
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
		if err := unix.Fsetxattr(fd, name, val, 0); err != nil {
			return fmt.Errorf("set xattr %s: %w", name, err)
		}
	}
	if !hasAcl {
		if err := unix.Fsetxattr(fd, posixAclAccess, minimalAclBlob(info.Mode().Perm()), 0); err != nil {
			if aclUnsupported(err) {
				// An ACL-less filesystem (Docker tmpfs and friends) has no
				// inherited ACL surface to overwrite — the chmod'd perms
				// stand and the record lands (gate round-46 item 6d).
				return preserveOwnership(tmp, original)
			}
			return fmt.Errorf("write minimal access acl: %w", err)
		}
	}
	// Ownership rides the contract (round-4 edge 6b): mode+xattr copy alone
	// would re-own the file to the process. A chown this process cannot
	// make is an error and the replace aborts.
	return preserveOwnership(tmp, original)
}

// aclUnsupported reports whether err is the filesystem's "ACLs not
// supported here" answer (the ENOTSUP/EOPNOTSUPP family, wrapped or not):
// on such a filesystem the F9 overwrite has no surface, so the seeder
// keeps the chmod'd perms and lands the record instead of aborting (gate
// round-46 item 6d).
func aclUnsupported(err error) bool {
	return errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.EOPNOTSUPP)
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
