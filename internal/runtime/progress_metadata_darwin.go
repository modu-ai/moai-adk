//go:build darwin

package runtime

import (
	"fmt"
	"os/exec"
)

// seedFileMetadata copies the original's mode, access-control entries, and
// extended attributes onto the temp file via cp -p. This is the
// LEADER-ACCEPTED darwin exception to the round-4 class closure (ruling
// i): the kernel surface structurally blocks explicit-ACL copy —
// clonefileat/SYS_COPYFILE/setattrlist measured (clonefileat copies xattrs
// but not ACLs; the SYS_COPYFILE trap returns EINVAL; the ACL bit is
// excluded from ATTR_CMN_SETMASK) — and cp -p is the ACL-guaranteeing
// seeder on darwin; a kauth_filesec reimplementation is a follow-up-SPEC
// candidate. A failure is an error and the caller aborts the replace —
// there is no mode-only fallback (sync-audit-4 F8).
func seedFileMetadata(tmp, original string) error {
	if out, err := exec.Command("cp", "-p", original, tmp).CombinedOutput(); err != nil {
		return fmt.Errorf("cp -p %s %s: %v (%s)", original, tmp, err, out)
	}
	// Ownership rides the contract too (round-4 edge 6b): cp -p usually
	// chowns as it copies, and preserveOwnership no-ops when it already
	// matches — a mismatch it cannot fix is an error and the replace
	// aborts.
	if err := preserveOwnership(tmp, original); err != nil {
		return err
	}
	return nil
}
