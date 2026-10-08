//go:build darwin || linux

package runtime

import (
	"os"
	"syscall"
)

// hardLinked reports whether path shares its inode with other directory
// entries (nlink > 1) — a rename would break the link relationship
// (round-4 edge 7c).
func hardLinked(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	return ok && st.Nlink > 1
}

// preserveOwnership chowns the temp to the original's uid/gid when they
// differ (round-4 edge 6b): mode+xattr copy alone re-owns the file to the
// process and changes who can access it, while the pre-repair
// os.WriteFile preserved ownership. The temp is addressed through its HELD
// descriptor (sync-audit-7 F13) — fd-based, immune to the name swap. A
// failure — this process may not re-own the temp, e.g. a non-root chown
// to a foreign user — is an ERROR and the caller aborts the replace (the
// F8 posture).
func preserveOwnership(tmp *os.File, original string) error {
	oinfo, err := os.Stat(original)
	if err != nil {
		return err
	}
	tinfo, err := tmp.Stat()
	if err != nil {
		return err
	}
	ost, ok1 := oinfo.Sys().(*syscall.Stat_t)
	tst, ok2 := tinfo.Sys().(*syscall.Stat_t)
	if !ok1 || !ok2 {
		return nil // no ownership data on this platform
	}
	if ost.Uid == tst.Uid && ost.Gid == tst.Gid {
		return nil
	}
	return tmp.Chown(int(ost.Uid), int(ost.Gid))
}
