// Package cli — regression tests for the project-dir resolution both Stop-hook
// review gates use to find `.moai/config/sections/workflow.yaml`.
//
// Claude Code sends `cwd` in the Stop hook payload; `project_dir` is a
// legacy/internal field (internal/hook/types.go marks it deprecated in favour of
// CWD). resolveProjectDirFromInput documented a "ProjectDir first, then Cwd"
// order but only ever implemented the first arm, so a real payload resolved to
// "" and both gate readers fail-CLOSED to false regardless of configuration —
// a third way for the gates to be unreachable. These tests pin the documented
// order, plus the CLAUDE_PROJECT_DIR fallback that keeps the Go layer agreeing
// with the shell wrappers' own resolution.
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestResolveProjectDirFromInput_Order(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "/env/project")

	for _, tc := range []struct {
		name  string
		input *hook.HookInput
		want  string
	}{
		{"nil input falls back to env", nil, "/env/project"},
		{"project_dir wins", &hook.HookInput{ProjectDir: "/explicit", CWD: "/cwd"}, "/explicit"},
		{"cwd used when project_dir absent", &hook.HookInput{CWD: "/cwd"}, "/cwd"},
		{"env used when both absent", &hook.HookInput{}, "/env/project"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveProjectDirFromInput(tc.input); got != tc.want {
				t.Errorf("resolveProjectDirFromInput = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestResolveProjectDirFromInput_NoEnvYieldsEmpty pins the final fail-CLOSED
// arm: with no payload hint and no env, resolution yields "" and the gate
// readers return false (the gate stays inert rather than reading a stray dir).
func TestResolveProjectDirFromInput_NoEnvYieldsEmpty(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	if got := resolveProjectDirFromInput(&hook.HookInput{}); got != "" {
		t.Errorf("resolveProjectDirFromInput = %q, want empty", got)
	}
	if readCodexReviewGateEnabled("") || readMultiReviewGateEnabled("") {
		t.Error("unresolved project dir must read both gates as disabled")
	}
}
