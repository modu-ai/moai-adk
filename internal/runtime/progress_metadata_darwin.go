//go:build darwin

package runtime

import (
	"fmt"
	"os/exec"
)

// seedFileMetadata copies the original's mode, access-control entries, and
// extended attributes onto the temp file via cp -p. This is the DELIBERATE
// darwin exception to the round-4 class closure, pending leader
// adjudication: the Go-native darwin kernel surface was measured closed —
// clonefileat(2) copies xattrs but NOT explicit ACLs (probe: the
// com.apple.provenance xattr survived, the `group:_guest deny read` entry
// did not), the raw SYS_COPYFILE trap returns EINVAL (the libc copyfile is
// userspace on modern macOS), and setattrlist cannot write the ACL —
// ATTR_CMN_EXTENDED_SECURITY (0x00400000) is excluded from
// ATTR_CMN_SETMASK (0x51C7FF00). The only remaining Go-native route is
// reimplementing the kauth_filesec wire format that no syscall accepts,
// which is its own SPEC. A failure is an error and the caller aborts the
// replace — there is no mode-only fallback (sync-audit-4 F8).
func seedFileMetadata(tmp, original string) error {
	if out, err := exec.Command("cp", "-p", original, tmp).CombinedOutput(); err != nil {
		return fmt.Errorf("cp -p %s %s: %v (%s)", original, tmp, err, out)
	}
	return nil
}
