// Package hygiene implements the .moai hygiene engine specified by
// SPEC-MOAI-HYGIENE-001: a size-based audit-log rotator and a finished-
// session state GC.
//
// Fail-closed contract (the package godoc is the normative summary):
//
//   - report mode (the shipped default) never mutates a candidate file,
//     never rotates, and never creates a lockfile; it appends at most one
//     summary row per unit to .moai/logs/hygiene-audit.jsonl.
//   - apply mode is an explicit per-surface opt-in (--apply on the CLI
//     invocation, or workflow.hygiene.mode on the auto path).
//   - a deletion happens only for a DEAD session whose age is datable from
//     a timestamp recorded in the candidate's own content and meets the
//     minimum-age floor; file mtime is never a deletion datum.
//   - every unmeasurable liveness signal is unmeasured, and unmeasured
//     never feeds a DEAD verdict (INDETERMINATE ⇒ keep with a reason).
//   - symlinked components strictly below the resolved .moai root are
//     refused; actions execute through a directory-fd-anchored root handle.
//
// @MX:NOTE: [AUTO] the named residual loss windows are documented in
// SPEC-MOAI-HYGIENE-001 §C and next to the code that owns them; do not
// silently assume them away.
package hygiene

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Mode is the hygiene execution mode. Report is the shipped default; apply
// is the only mutating mode.
type Mode string

const (
	// ModeReport reports decisions without mutating anything.
	ModeReport Mode = "report"
	// ModeApply performs rotations and deletions.
	ModeApply Mode = "apply"
)

// ParseMode maps a configuration string onto a Mode. An unrecognizable
// string falls back to report — the non-mutating default (D30).
func ParseMode(s string) Mode {
	if Mode(s) == ModeApply {
		return ModeApply
	}
	return ModeReport
}

// Outcome names the per-sink / per-candidate outcomes recorded in audit
// rows. The closed set is asserted by the audit-row tests.
type Outcome string

const (
	OutcomeRotated            Outcome = "rotated"
	OutcomeDeleted            Outcome = "deleted"
	OutcomeSummary            Outcome = "summary"
	OutcomeSkippedUnderThresh Outcome = "skipped-under-threshold"
	OutcomeSkippedStale       Outcome = "skipped-stale"
	OutcomeSkippedAbsent      Outcome = "skipped-absent"
	OutcomeSkippedLocked      Outcome = "skipped-locked"
	OutcomeSkippedPlatform    Outcome = "skipped-platform"
	OutcomeSkippedReportMode  Outcome = "skipped-report-mode"
	OutcomeKept               Outcome = "kept"
	OutcomeAlreadyGone        Outcome = "already-gone"
	OutcomeRejudgedKeep       Outcome = "rejudged-keep"
	OutcomeSymlinkRefused     Outcome = "symlink-refused"
	OutcomeLockClassExcluded  Outcome = "lock-class-excluded"
	OutcomeError              Outcome = "error"
)

// testRoots registers the temporary directory roots a test-driven run may
// operate on (REQ-HYG-015, D22). Production entry points consult it only
// when testing.Testing() is true.
var testRoots = struct {
	sync.Mutex
	byBase map[string]string // resolved root -> resolved base
}{byBase: map[string]string{}}

// registerTestRoot allows the units' entry points to operate under root
// during tests. Called by every test that constructs an engine.
func registerTestRoot(root string) {
	resolved := resolvePath(root)
	testRoots.Lock()
	defer testRoots.Unlock()
	testRoots.byBase[resolved] = resolved
}

// resolvePath canonicalizes a path, falling back to the input when the
// resolution fails (a missing path is validated on its cleaned form).
func resolvePath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// validateRoot refuses, under a test binary, any root outside the
// registered temporary directories — the runtime half of REQ-HYG-015's
// test-isolation guarantee. Outside a test binary it is a no-op.
func validateRoot(root string) error {
	if !testing.Testing() {
		return nil
	}
	resolved := resolvePath(root)
	testRoots.Lock()
	defer testRoots.Unlock()
	for base := range testRoots.byBase {
		if resolved == base || strings.HasPrefix(resolved, base+string(os.PathSeparator)) {
			return nil
		}
	}
	return fmt.Errorf("hygiene: root %q is outside every registered test root (REQ-HYG-015)", root)
}
