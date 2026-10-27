//go:build !windows

package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/project"
)

// The two preserved-line rows below make a write fail through file permissions,
// which has no equivalent on Windows; they register into the shared subtest list
// from this unix-only file.
func init() {
	preservedPlatformSubtests = append(preservedPlatformSubtests,
		preservedSubtest{"p17_init_manifest_retrack", checklistPreservedRetrackSkip},
		preservedSubtest{"p23_init_resume_warnings", checklistPreservedResumeWarnings},
	)
}

// afterExecuteExecutor wraps the real init executor and runs fn once the
// deploy returned, so the project is deployed and its manifest recorded.
type afterExecuteExecutor struct {
	inner initPhaseExecutor
	fn    func()
}

func (e *afterExecuteExecutor) SetReporter(r project.ProgressReporter) { e.inner.SetReporter(r) }

func (e *afterExecuteExecutor) Execute(ctx context.Context, opts project.InitOptions) (*project.InitResult, error) {
	res, err := e.inner.Execute(ctx, opts)
	if err == nil && e.fn != nil {
		e.fn()
	}
	return res, err
}

// afterExecute makes the real executor call fn after a successful deploy.
func afterExecute(t *testing.T, fn func()) {
	t.Helper()
	orig := newInitPhaseExecutorFn
	newInitPhaseExecutorFn = func(d project.Detector, m project.MethodologyDetector, v project.ProjectValidator, i project.Initializer) initPhaseExecutor {
		return &afterExecuteExecutor{inner: orig(d, m, v, i), fn: fn}
	}
	t.Cleanup(func() { newInitPhaseExecutorFn = orig })
}

// requireNotRoot fails (never skips) when the permission-based injection cannot
// bite: root reads and writes through mode bits.
func requireNotRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Fatal("running as root: a chmod-based failure injection cannot bite, so this row would assert nothing")
	}
}

// checklistPreservedRetrackSkip is p17: one unreadable template-managed section
// file is skipped by the retrack, the skip line appears once under the manifest
// item, the item ends with the warning glyph, the run still succeeds, and the
// summary warning count is unchanged (a preserved line is not a collector warning).
func checklistPreservedRetrackSkip(t *testing.T) {
	requireNotRoot(t)
	_, projectDir := checklistFallbackEnv(t)
	forceChecklistOn(t)
	checklistRichClock(t)
	victim := filepath.Join(projectDir, ".moai", "config", "sections", "report.yaml")
	afterExecute(t, func() {
		if _, err := os.Stat(victim); err != nil {
			t.Errorf("the injection target %s was not deployed: %v", victim, err)
			return
		}
		if err := os.Chmod(victim, 0o000); err != nil {
			t.Errorf("chmod %s: %v", victim, err)
		}
	})
	t.Cleanup(func() { _ = os.Chmod(victim, 0o644) })

	cmd, out, errBuf := newChecklistInitCmd(t, nil)
	if err := runInit(cmd, nil); err != nil {
		t.Fatalf("init: %v\n%s", err, errBuf.String())
	}
	requirePreservedOnce(t, out.String(), errBuf.String(), "manifest retrack skip")
	requireItemGlyph(t, errBuf.String(), "verify.manifest", "!", "✓")
	cErr := checklistCollapse(errBuf.String())
	if !strings.Contains(cErr, "· 0 warnings ·") {
		t.Errorf("a preserved row must not be added to the warning count:\n%s", cErr)
	}
}

// checklistPreservedResumeWarnings is p23: the four resume-path warning
// fragments, each produced by a real failure, each exactly once across both
// writers, with no summary drawn.
func checklistPreservedResumeWarnings(t *testing.T) {
	requireNotRoot(t)
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "resume-proj")
	resumeSeedRun(t, projectDir, "claude")
	forceChecklistOn(t)
	checklistRichClock(t)
	t.Setenv("NO_COLOR", "1")

	// Both hook installs fail: .git is a regular file, so .git/hooks cannot be made.
	clWriteFile(t, filepath.Join(projectDir, ".git"), "gitdir: nowhere\n", 0o644)

	moaiDir := filepath.Join(projectDir, ".moai")
	orig := applyAutonomyTierBundleFn
	applyAutonomyTierBundleFn = func(a, b, c, d string) error {
		err := orig(a, b, c, d)
		// After the autonomy step (which must still succeed): a malformed global
		// settings file for the cleanup that follows, and a manifest directory
		// that refuses the retrack's atomic save.
		home, homeErr := userHomeDirFn()
		if homeErr != nil {
			t.Errorf("home: %v", homeErr)
			return err
		}
		clWriteFile(t, filepath.Join(home, ".claude", "settings.json"), "{not json", 0o644)
		if chErr := os.Chmod(moaiDir, 0o555); chErr != nil {
			t.Errorf("chmod %s: %v", moaiDir, chErr)
		}
		return err
	}
	t.Cleanup(func() {
		applyAutonomyTierBundleFn = orig
		_ = os.Chmod(moaiDir, 0o755)
	})

	cmd := newInitTestCmd()
	cmd.Flags().Bool("verbose", false, "")
	if err := cmd.Flags().Set("llm", "claude"); err != nil {
		t.Fatal(err)
	}
	// The test command defaults --no-hooks to true; the hook installs must run.
	if err := cmd.Flags().Set("no-hooks", "false"); err != nil {
		t.Fatal(err)
	}
	out, errBuf := &strings.Builder{}, &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("the resume must return nil even when its warnings fire: %v\n%s", err, errBuf.String())
	}
	for _, frag := range []string{
		"Warning: global settings env update failed",
		"Pre-push hook installation failed",
		"Pre-commit hook installation failed",
		"Warning: manifest retrack (config sections) failed",
	} {
		requirePreservedOnce(t, out.String(), errBuf.String(), frag)
	}
	if strings.Contains(checklistCollapse(errBuf.String()), "Done ·") {
		t.Errorf("a resumed run draws no summary:\n%s", checklistCollapse(errBuf.String()))
	}
}
