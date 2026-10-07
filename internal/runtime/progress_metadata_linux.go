//go:build linux

package runtime

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// seedFileMetadata makes the temp file carry the original's metadata —
// the mode (a plain stat/chmod, applied FIRST: on Linux a chmod rewrites
// the POSIX ACL's mask, so the ACL must land after it) and then the
// extended attributes through the xattr family, POSIX ACLs included
// (Linux stores them as the system.posix_acl_access attribute) — one
// Go-native mechanism with no external tool (round-4 class closure). A
// failure is an error and the caller aborts the replace — there is no
// mode-only fallback.
func seedFileMetadata(tmp, original string) error {
	if info, err := os.Stat(original); err != nil {
		return err
	} else if err := os.Chmod(tmp, info.Mode().Perm()); err != nil {
		return err
	}
	names, err := listXattrs(original)
	if err != nil {
		return err
	}
	for _, name := range names {
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
	// Ownership rides the contract (round-4 edge 6b): mode+xattr copy alone
	// would re-own the file to the process. A chown this process cannot
	// make is an error and the replace aborts.
	return preserveOwnership(tmp, original)
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
