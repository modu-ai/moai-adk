//go:build darwin

package runtime

import (
	"fmt"
	"os"
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
func seedFileMetadata(tmp *os.File, tmpPath, original string) error {
	// The temp was created in the original's directory, so os.CreateTemp
	// handed it the PARENT's inherited ACL entries at birth. Strip them
	// FIRST (chmod -N — the temp is still empty), then cp -p copies the
	// original's data, mode, and ACL exactly: inherited entries cannot
	// survive, and an explicit original ACL is applied after the strip, so
	// the replaced file's ACL is EXACTLY the original's (sync-audit-5
	// item 2 — the inverse face of F6).
	//
	// F13 note: cp -p / chmod are PATH-based — the caller verifies the
	// held fd against the name immediately before and after this call, and
	// the CONTENT write is fully fd-based; the residual between those
	// verifications and cp's internal writes is the flagged darwin
	// conflict (see the round report).
	if out, err := exec.Command("chmod", "-N", tmpPath).CombinedOutput(); err != nil {
		return fmt.Errorf("chmod -N %s: %v (%s)", tmpPath, err, out)
	}
	if out, err := exec.Command("cp", "-p", original, tmpPath).CombinedOutput(); err != nil {
		return fmt.Errorf("cp -p %s %s: %v (%s)", original, tmpPath, err, out)
	}
	// Ownership rides the contract too (round-4 edge 6b): cp -p usually
	// chowns as it copies, and preserveOwnership no-ops when it already
	// matches — a mismatch it cannot fix is an error and the replace
	// aborts. Ownership is applied through the held fd.
	if err := preserveOwnership(tmp, original); err != nil {
		return err
	}
	return nil
}
