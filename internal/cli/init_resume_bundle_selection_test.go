package cli

// init_resume_bundle_selection_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// codex round-2 finding 1: the bundle selection a failing init run requested must
// survive into its resume.
//
// The failing run requests --bundles and dies at the user-asset ensure before any
// journal exists (a lock refusal). Neither the user manifest nor the journal records
// the selection, so a rerun with no --bundles of its own resumed the ensure with an
// empty selection: the resume completed, the requested bundle was never installed,
// and the checkpoint was cleared.

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// stubFailingFirstEnsure makes the first user-asset ensure fail the way a lock
// refusal does, before any journal is written. Later ensures run the real install.
func stubFailingFirstEnsure(t *testing.T) {
	t.Helper()
	calls := 0
	stubEnsure(t, func(homeDir string, selection []string, out io.Writer) error {
		calls++
		if calls == 1 {
			return errors.New("simulated early ensure failure: user lock acquisition refused")
		}
		return ensureUserAssetsLocked(homeDir, selection, out)
	})
}

// initRunWithBundles runs init on projectDir with the given harness and bundle
// selections. An empty value names no flag, which is the resume path under test.
func initRunWithBundles(t *testing.T, projectDir, llm, bundles string) error {
	t.Helper()
	cmd := newInitTestCmd()
	if llm != "" {
		if err := cmd.Flags().Set("llm", llm); err != nil {
			t.Fatal(err)
		}
	}
	if bundles != "" {
		if err := cmd.Flags().Set("bundles", bundles); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	return runInit(cmd, []string{projectDir})
}

// manifestBundles returns the bundle list the user manifest records under home.
func manifestBundles(t *testing.T, home string) []string {
	t.Helper()
	m, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		t.Fatalf("load user manifest: %v", err)
	}
	return m.Bundles
}

// TestInitResumeRestoresBundleSelectionAfterLockFailure reproduces finding 1. The
// first run requests frontend and fails at the user-asset ensure before any journal
// is written. The rerun names no --bundles, so the resumed ensure must install the
// bundle the failing run requested. The control reruns with an explicit devops, which
// must install devops and not the recorded frontend.
func TestInitResumeRestoresBundleSelectionAfterLockFailure(t *testing.T) {
	t.Run("rerun_without_flag_installs_the_requested_bundle", func(t *testing.T) {
		isolatedInitHome(t)
		projectDir := filepath.Join(t.TempDir(), "bundle-resume-proj")
		stubFailingFirstEnsure(t)

		if err := initRunWithBundles(t, projectDir, "both", "frontend"); err == nil || !strings.Contains(err.Error(), "user-asset install failed") {
			t.Fatalf("run 1 did not fail at the user-asset ensure (the injection did not bite): %v", err)
		}
		home := userHomeDirOrEmpty()
		if _, err := os.Stat(userassets.JournalPath(home)); !os.IsNotExist(err) {
			t.Fatalf("precondition: a pending-install journal exists after an early ensure failure (stat err = %v)", err)
		}
		if got := manifestBundles(t, home); len(got) != 0 {
			t.Fatalf("precondition: the failed run recorded bundles %v before its ensure ran", got)
		}

		if err := initRunWithBundles(t, projectDir, "", ""); err != nil {
			t.Fatalf("run 2 (the rerun without --bundles) failed: %v", err)
		}
		if got := manifestBundles(t, home); !slices.Contains(got, "frontend") {
			t.Errorf("the resumed run did not install the bundle the failing run requested: the manifest records bundles %v, want frontend among them", got)
		}
	})

	t.Run("control_explicit_bundle_on_rerun_wins_over_the_recorded_one", func(t *testing.T) {
		isolatedInitHome(t)
		projectDir := filepath.Join(t.TempDir(), "bundle-explicit-proj")
		stubFailingFirstEnsure(t)

		if err := initRunWithBundles(t, projectDir, "both", "frontend"); err == nil || !strings.Contains(err.Error(), "user-asset install failed") {
			t.Fatalf("run 1 did not fail at the user-asset ensure (the injection did not bite): %v", err)
		}
		home := userHomeDirOrEmpty()
		if err := initRunWithBundles(t, projectDir, "", "devops"); err != nil {
			t.Fatalf("run 2 (the rerun with an explicit --bundles devops) failed: %v", err)
		}
		if got := manifestBundles(t, home); !slices.Contains(got, "devops") || slices.Contains(got, "frontend") {
			t.Errorf("an explicit --bundles devops on the rerun must install devops and not the recorded frontend: the manifest records bundles %v", got)
		}
	})
}
