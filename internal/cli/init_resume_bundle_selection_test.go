package cli

// init_resume_bundle_selection_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// codex round-2 finding 1: the bundle selection a failing init run requested must
// survive into its resume.
//
// The failing run requests --bundles and is interrupted in the post-executor window
// (card t1619). The ensure runs before the executor now and installs the requested
// selection first, so the init marker is the record that carries the selection into a
// rerun with no --bundles of its own. The finding is re-aimed at that record.

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

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
		injectInitWindowFailure(t)

		if err := initRunWithBundles(t, projectDir, "both", "frontend"); err == nil || !strings.Contains(err.Error(), "initialization interrupted after the executor") {
			t.Fatalf("run 1 did not fail at the post-executor window (the injection did not bite): %v", err)
		}
		home := userHomeDirOrEmpty()
		if got := initPendingBundles(projectDir); !slices.Equal(got, []string{"frontend"}) {
			t.Fatalf("precondition: the interrupted run's init marker records bundles %v, want [frontend]", got)
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
		injectInitWindowFailure(t)

		if err := initRunWithBundles(t, projectDir, "both", "frontend"); err == nil || !strings.Contains(err.Error(), "initialization interrupted after the executor") {
			t.Fatalf("run 1 did not fail at the post-executor window (the injection did not bite): %v", err)
		}
		home := userHomeDirOrEmpty()
		if err := initRunWithBundles(t, projectDir, "", "devops"); err != nil {
			t.Fatalf("run 2 (the rerun with an explicit --bundles devops) failed: %v", err)
		}
		if got := manifestBundles(t, home); !slices.Contains(got, "devops") {
			t.Errorf("an explicit --bundles devops on the rerun must install devops: the manifest records bundles %v", got)
		}
	})
}
