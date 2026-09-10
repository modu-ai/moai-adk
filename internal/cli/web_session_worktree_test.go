package cli

// SPEC-SESSION-WORKTREE-001 M3 — `moai web` auto-entry + advisory suppression
// tests.
//
// runWeb blocks at web.Run until SIGINT, so these tests substitute webRunFn
// (the package-level seam) with a no-op, swap the session-worktree git seams
// via swapSessionWorktreeSeams, force the feature ON via MOAI_SESSION_WORKTREE=1
// (REQ-SW-003), and swap findProjectRootFn so project-root resolution never
// touches the real filesystem. stdout is captured through cmd.OutOrStdout() so
// the emitWorktreeAdvisory output (REQ-SW-013) is directly observable.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/web"
)

// advisoryMarker is the substring emitWorktreeAdvisory always prints (both the
// auto-create and the recommendation phrasings contain it — see AC-WBG-009).
const advisoryMarker = "main-checkout-branch-guard.md"

// runWebHarness builds an isolated newWebCmd wired with the swapped seams and
// returns the stdout/stderr buffers. The caller restores nothing — cleanup is
// registered inside. The returned projectRoot is the temp dir the swapped
// findProjectRootFn returns.
func runWebHarness(t *testing.T, seams swSeams, opts ...func(*runWebHarnessCfg)) (*bytes.Buffer, *bytes.Buffer, string) {
	t.Helper()

	cfg := runWebHarnessCfg{
		materializePath: t.TempDir(),
	}
	for _, o := range opts {
		o(&cfg)
	}

	swapSessionWorktreeSeams(t, seams)

	// findProjectRootFn → temp dir (never touches the real filesystem).
	origFind := findProjectRootFn
	findProjectRootFn = func() (string, error) {
		return cfg.projectRoot, nil
	}
	t.Cleanup(func() { findProjectRootFn = origFind })

	// webRunFn → no-op (web.Run blocks; we never want to start a real server).
	origRun := webRunFn
	webRunFn = func(ctx context.Context, _ web.Config) error { return nil }
	t.Cleanup(func() { webRunFn = origRun })

	// Port reclamation: ensurePortFree returns nil when the port is free.
	origCheck := checkPortInUse
	checkPortInUse = func(int) bool { return false }
	t.Cleanup(func() { checkPortInUse = origCheck })

	// cwd save/restore — runWeb os.Chdir's into the worktree on success, which
	// is process-global and would leak into sibling tests.
	origCwd, cwdErr := os.Getwd()
	t.Cleanup(func() {
		if cwdErr == nil {
			_ = os.Chdir(origCwd)
		}
	})

	cmd := newWebCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{})

	t.Cleanup(func() {
		// Reset the package-level flag vars between tests (newWebCmd binds to
		// these; leftover state from a prior test would change defaults).
		webPort = 3041
		webNoOpen = false
		webNoReuse = false
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("runWeb Execute returned error: %v (stderr=%s)", err, errOut.String())
	}
	return out, errOut, cfg.projectRoot
}

type runWebHarnessCfg struct {
	materializePath string
	projectRoot     string
}

// withRootDir sets the temp dir findProjectRootFn returns.
func withRootDir(p string) func(*runWebHarnessCfg) {
	return func(c *runWebHarnessCfg) { c.projectRoot = p }
}

// materializeSeams returns swSeams configured for a SUCCESSFUL worktree
// materialization (inWt=false, add returns the given path, configSet succeeds).
// The add seam also mkdir's the dest so runWeb's subsequent os.Chdir succeeds —
// wtMaterialized only flips true when the process actually enters the worktree.
func materializeSeams(dest string) swSeams {
	return swSeams{
		inWt:      func() bool { return false },
		short:     func() string { return "abcdef12" },
		commonDir: func() (string, error) { return filepath.Join(dest, ".git"), nil },
		add: func(d, _, _ string) (string, error) {
			if err := os.MkdirAll(d, 0o755); err != nil {
				return d, err
			}
			return d, nil
		},
		configSet: func(_, _, _ string) error { return nil },
	}
}

// failBackSeams returns swSeams where materialization fails (REQ-SW-004 path).
func failBackSeams() swSeams {
	return swSeams{
		inWt:      func() bool { return false },
		short:     func() string { return "abcdef12" },
		commonDir: func() (string, error) { return "", errFailBackSentinel },
		add:       func(_, _, _ string) (string, error) { return "", errFailBackSentinel },
		configSet: func(_, _, _ string) error { return nil },
	}
}

var errFailBackSentinel = &strErr{"simulated non-git dir (REQ-SW-004 fail-back trigger)"}

type strErr struct{ s string }

func (e *strErr) Error() string { return e.s }

// TestRunWeb_AutoEntryOn_SuppressesAdvisory — REQ-SW-013 / AC-SW-013 positive:
// feature ON + materialization succeeds → emitWorktreeAdvisory is SUPPRESSED.
func TestRunWeb_AutoEntryOn_SuppressesAdvisory(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	dest := t.TempDir()
	out, _, _ := runWebHarness(t, materializeSeams(dest), withRootDir(dest))

	if strings.Contains(out.String(), advisoryMarker) {
		t.Errorf("REQ-SW-013: advisory should be SUPPRESSED on materialization, but stdout contains it:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "MoAI Web Console starting") {
		t.Errorf("expected the console-start line in stdout, got:\n%s", out.String())
	}
}

// TestRunWeb_AutoEntryOff_AdvisoryFires — REQ-SW-001 / AC-SW-001 negative
// control: feature OFF → advisory fires exactly as today (byte-identical).
func TestRunWeb_AutoEntryOff_AdvisoryFires(t *testing.T) {
	// MOAI_SESSION_WORKTREE deliberately UNSET.
	root := t.TempDir()
	out, _, _ := runWebHarness(t, materializeSeams(root), withRootDir(root))

	if !strings.Contains(out.String(), advisoryMarker) {
		t.Errorf("REQ-SW-001: advisory MUST fire when feature is OFF, but stdout lacks it:\n%s", out.String())
	}
}

// TestRunWeb_AutoEntryOn_FailBack_AdvisoryFires — REQ-SW-004 / AC-SW-013
// negative control: feature ON + materialization fell back → advisory STILL
// fires (the hazard was NOT avoided by construction).
func TestRunWeb_AutoEntryOn_FailBack_AdvisoryFires(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	root := t.TempDir()
	out, _, _ := runWebHarness(t, failBackSeams(), withRootDir(root))

	if !strings.Contains(out.String(), advisoryMarker) {
		t.Errorf("REQ-SW-004: advisory MUST fire on fail-back (hazard not avoided), but stdout lacks it:\n%s", out.String())
	}
}
