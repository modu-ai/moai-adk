//go:build darwin

package runtime

import (
	"fmt"
	"os"
	"os/exec"
)

// seedFileMetadata copies the original's mode, access-control entries, and
// extended attributes onto the temp file via cp -p. Route (ii) disposition
// (decision-index Q2, leader ruling (a) — OVERRIDABLE by an operator
// ruling, SPEC-PROGRESS-RECORD-IO-001): the exec-based seeder is KEPT —
// the kernel surface structurally blocks explicit-ACL copy —
// clonefileat/SYS_COPYFILE/setattrlist measured (clonefileat copies xattrs
// but not ACLs; the SYS_COPYFILE trap returns EINVAL; the ACL bit is
// excluded from ATTR_CMN_SETMASK) — and cp -p is the ACL-guaranteeing
// seeder on darwin. A failure is an error and the caller aborts the
// replace — there is no mode-only fallback (sync-audit-4 F8).
//
// Live residuals under route (ii), named at their measured harm class:
//   - fd-verify→rename: the pre/post Lstat+fdMatchesName checks below
//     bound, but cannot close, a name-swap→symlink race between the final
//     re-check and cp's own open — victim-overwrite stays possible inside
//     that microsecond window (darwin has no fd-anchored ACL copy;
//     copyfile(3) is userspace).
//   - delete-denied rename (decision-index Q5): a write-allowed/
//     delete-denied ACL on the target makes the caller's atomic rename
//     fail EPERM where the pre-repair in-place write succeeded, so no
//     record lands; the unwritable-dir/hardlink in-place fallback is the
//     noted future fix shape.
//
// A kauth_filesec reimplementation remains the named future-fix direction
// (follow-up-SPEC candidate): the M1 probe (card t1598, ad9ba32b2)
// measured the inert kauth-filesec xattr name — a nil-error write leaves
// the enforced ACL unchanged on APFS — and the com.apple.system.*
// namespace EPERM-gated for non-root.
func seedFileMetadata(tmp *os.File, tmpPath, original string) error {
	// The temp was created in the original's directory, so os.CreateTemp
	// handed it the PARENT's inherited ACL entries at birth. Strip them
	// FIRST (chmod -N — the temp is still empty), then cp -p copies the
	// original's data, mode, and ACL exactly: inherited entries cannot
	// survive, and an explicit original ACL is applied after the strip, so
	// the replaced file's ACL is EXACTLY the original's (sync-audit-5
	// item 2 — the inverse face of F6).
	//
	// F13 shrink (gate round-46 item 1 — as far as darwin allows): the
	// seeder re-validates the temp IMMEDIATELY before its path-based steps
	// and rejects a swapped-in SYMLINK outright — chmod/cp must never
	// write through an attacker-planted link — then re-checks after them.
	// The residual between the re-check and cp's own open is a
	// microsecond race that darwin's platform constraints cannot close
	// without cgo (copyfile(3) is userspace; no fd-anchored ACL copy
	// exists) — dispositioned LIVE under route (ii) (decision-index Q2,
	// ruling (a)): victim-overwrite stays possible in this window, and
	// the fd-verify→rename residual is re-documented in the header
	// comment above.
	if info, lerr := os.Lstat(tmpPath); lerr != nil {
		return lerr
	} else if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("darwin seeder: temp %s was swapped for a symlink", tmpPath)
	} else if !fdMatchesName(tmp, tmpPath) {
		return fmt.Errorf("darwin seeder: temp %s no longer matches the held fd", tmpPath)
	}
	if out, err := exec.Command("chmod", "-N", tmpPath).CombinedOutput(); err != nil {
		return fmt.Errorf("chmod -N %s: %v (%s)", tmpPath, err, out)
	}
	if out, err := exec.Command("cp", "-p", original, tmpPath).CombinedOutput(); err != nil {
		return fmt.Errorf("cp -p %s %s: %v (%s)", original, tmpPath, err, out)
	}
	if info, lerr := os.Lstat(tmpPath); lerr != nil {
		return lerr
	} else if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("darwin seeder: temp %s was swapped for a symlink during the copy", tmpPath)
	} else if !fdMatchesName(tmp, tmpPath) {
		return fmt.Errorf("darwin seeder: temp %s no longer matches the held fd after the copy", tmpPath)
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
