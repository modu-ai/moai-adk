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
	"time"
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

// Engine is the combined hygiene pass: the rotator and the GC run in one
// invocation and fail independently (REQ-HYG-012) — a rotator error never
// prevents the GC run and vice versa; each outcome is returned to its
// caller for logging.
type Engine struct {
	LogDir           string
	MoaiRoot         string
	RegistryPath     string
	TranscriptRoots  []string
	MaxBytes         int64
	KeptRotations    int
	MinAge           time.Duration
	TranscriptWindow time.Duration
	HeartbeatWindow  time.Duration
}

// Run executes both units in the given mode. The rotator's and the GC's
// errors are independent: errRotator and errGC are non-nil only for their
// own unit, and a failure of one never blocks the other (REQ-HYG-012).
func (e *Engine) Run(mode Mode) (rotatorRows []AuditRow, gcReport *GCReport, errRotator error, errGC error) {
	r := &Rotator{LogDir: e.LogDir, MaxBytes: e.MaxBytes, KeptRotations: e.KeptRotations}
	rotatorRows, errRotator = r.Run(mode)

	g := &GC{
		MoaiRoot:         e.MoaiRoot,
		RegistryPath:     e.RegistryPath,
		TranscriptRoots:  e.TranscriptRoots,
		MinAge:           e.MinAge,
		TranscriptWindow: e.TranscriptWindow,
		HeartbeatWindow:  e.HeartbeatWindow,
	}
	gcReport, errGC = g.Run(mode)
	return rotatorRows, gcReport, errRotator, errGC
}

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

// resolvePath canonicalizes a path through its nearest existing ancestor:
// the ancestor is symlink-resolved (the macOS /var → /private/var shape)
// and the non-existing remainder rejoined, so a root registered before its
// fixture tree exists and the same root validated after creation resolve
// identically.
func resolvePath(path string) string {
	clean := filepath.Clean(path)
	ancestor := clean
	var remainder []string
	for {
		if _, err := os.Stat(ancestor); err == nil {
			break
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			break // reached the filesystem root
		}
		remainder = append([]string{filepath.Base(ancestor)}, remainder...)
		ancestor = parent
	}
	if resolved, err := filepath.EvalSymlinks(ancestor); err == nil {
		return filepath.Join(append([]string{resolved}, remainder...)...)
	}
	return clean
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
