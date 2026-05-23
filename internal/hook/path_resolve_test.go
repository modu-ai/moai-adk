// path_resolve_test.go — unit tests for cwd-leak helpers introduced by
// SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 M1.
//
// These tests cover resolveProjectRootFromEnv and resolveProjectRootFromInputOrEnv,
// the siblings of resolveProjectRoot (post_tool_metrics.go) that omit the
// .moai/ existence guard. The .moai/-guarded resolveProjectRoot is exercised
// elsewhere; here we exercise the env→cwd and input→env→cwd fallback paths.
package hook

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestResolveProjectRootFromEnv_PrefersEnvVar verifies that when
// CLAUDE_PROJECT_DIR is set, the helper returns it verbatim — no os.Getwd()
// fallback is consulted.
func TestResolveProjectRootFromEnv_PrefersEnvVar(t *testing.T) {
	t.Setenv(config.EnvClaudeProjectDir, "/tmp/from-env")
	got := resolveProjectRootFromEnv("test")
	if got != "/tmp/from-env" {
		t.Fatalf("expected /tmp/from-env, got %q", got)
	}
}

// TestResolveProjectRootFromEnv_FallsBackToGetwd verifies that when
// CLAUDE_PROJECT_DIR is unset, the helper falls back to os.Getwd().
func TestResolveProjectRootFromEnv_FallsBackToGetwd(t *testing.T) {
	t.Setenv(config.EnvClaudeProjectDir, "")
	got := resolveProjectRootFromEnv("test")
	if got == "" {
		t.Fatal("expected non-empty cwd fallback, got empty string")
	}
}

// TestResolveProjectRootFromInputOrEnv_PrefersInputCWD verifies that when
// input.CWD is non-empty, it wins over the env var.
func TestResolveProjectRootFromInputOrEnv_PrefersInputCWD(t *testing.T) {
	t.Setenv(config.EnvClaudeProjectDir, "/tmp/from-env")
	input := &HookInput{CWD: "/tmp/from-input"}
	got := resolveProjectRootFromInputOrEnv(input, "test")
	if got != "/tmp/from-input" {
		t.Fatalf("expected /tmp/from-input, got %q", got)
	}
}

// TestResolveProjectRootFromInputOrEnv_FallsBackToEnv verifies that when
// input is nil or input.CWD is empty, the helper falls back to the env-var path.
func TestResolveProjectRootFromInputOrEnv_FallsBackToEnv(t *testing.T) {
	t.Setenv(config.EnvClaudeProjectDir, "/tmp/from-env")
	got := resolveProjectRootFromInputOrEnv(nil, "test")
	if got != "/tmp/from-env" {
		t.Fatalf("nil input should fall back to env, got %q", got)
	}
	got = resolveProjectRootFromInputOrEnv(&HookInput{}, "test")
	if got != "/tmp/from-env" {
		t.Fatalf("empty CWD should fall back to env, got %q", got)
	}
}

// TestResolveProjectRootFromInputOrEnv_FallsBackToGetwd verifies that when
// both input.CWD and the env var are empty, the helper falls back to os.Getwd().
func TestResolveProjectRootFromInputOrEnv_FallsBackToGetwd(t *testing.T) {
	t.Setenv(config.EnvClaudeProjectDir, "")
	got := resolveProjectRootFromInputOrEnv(nil, "test")
	if got == "" {
		t.Fatal("expected non-empty cwd fallback, got empty string")
	}
}
