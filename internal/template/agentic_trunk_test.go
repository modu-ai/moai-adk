package template

// agentic_trunk_test.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M4 (card
// t1831). AC-AGT-015: the github-flow sweep guard stays armed only for
// github-flow — an agentic-trunk root does not arm it (REQ-AGT-016). The
// dedicated test is the flip judge: the existing TestGitHubFlowSweepGuard
// prints a PASS in both states, so it cannot separate arming from not.

import (
	"os"
	"path/filepath"
	"testing"
)

// sweepGuardWriteRoot writes a git-strategy.yaml naming the given workflow
// under a synthetic root (an empty workflow writes no file).
func sweepGuardWriteRoot(t *testing.T, workflow string) string {
	t.Helper()
	dir := t.TempDir()
	if workflow == "" {
		return dir
	}
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n    mode: manual\n    manual:\n        workflow: " + workflow + "\n"
	if err := os.WriteFile(filepath.Join(sections, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestAgenticTrunkSweepGuardNotArmed is AC-AGT-015: fed an agentic-trunk
// root, a git-flow root and a github-flow root, with MOAI_SWEEP_GUARD_ARMED
// cleared, sweepGuardArmed answers true for the github-flow root only. A
// mutant that arms the guard for agentic-trunk turns the first subtest red.
func TestAgenticTrunkSweepGuardNotArmed(t *testing.T) {
	t.Setenv(sweepArmEnv, "") // the env override must not leak into the verdict

	cases := []struct {
		name     string
		workflow string
		want     bool
	}{
		{name: "agentic-trunk", workflow: "agentic-trunk", want: false},
		{name: "git-flow", workflow: "git-flow", want: false},
		{name: "github-flow", workflow: "github-flow", want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, why := sweepGuardArmed(sweepGuardWriteRoot(t, tc.workflow))
			if got != tc.want {
				t.Errorf("sweepGuardArmed(%s) = %v (%s), want %v", tc.name, got, why, tc.want)
			}
		})
	}

	t.Run("no config does not arm", func(t *testing.T) {
		if got, why := sweepGuardArmed(sweepGuardWriteRoot(t, "")); got {
			t.Errorf("sweepGuardArmed(empty root) = %v (%s), want false", got, why)
		}
	})
}
