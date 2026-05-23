// path_resolve_test.go — unit tests for the cwd-leak helper introduced by
// SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 M3.
//
// resolveQualityProjectDir is the package-quality sibling of the parent
// hook package's resolveProjectRootFromEnv. It prefers cfg.ProjectDir
// over CLAUDE_PROJECT_DIR env var over os.Getwd() fallback. No .moai/
// existence guard — quality-gate operations are contractually invoked
// from project roots and may need to operate in subdirectories during
// pre-commit hook execution.
package quality

import (
	"testing"
)

// TestResolveQualityProjectDir_PrefersCfgProjectDir verifies that when
// cfg.ProjectDir is non-empty, the helper returns it verbatim — no env-var
// or os.Getwd() fallback is consulted.
func TestResolveQualityProjectDir_PrefersCfgProjectDir(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "/tmp/from-env")
	cfg := GateConfig{ProjectDir: "/tmp/from-cfg"}
	got := resolveQualityProjectDir(cfg, "test")
	if got != "/tmp/from-cfg" {
		t.Fatalf("expected /tmp/from-cfg, got %q", got)
	}
}

// TestResolveQualityProjectDir_FallsBackToEnv verifies that when
// cfg.ProjectDir is empty, the helper falls back to CLAUDE_PROJECT_DIR.
func TestResolveQualityProjectDir_FallsBackToEnv(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "/tmp/from-env")
	cfg := GateConfig{} // empty ProjectDir
	got := resolveQualityProjectDir(cfg, "test")
	if got != "/tmp/from-env" {
		t.Fatalf("expected /tmp/from-env, got %q", got)
	}
}

// TestResolveQualityProjectDir_FallsBackToGetwd verifies that when both
// cfg.ProjectDir and CLAUDE_PROJECT_DIR are empty, the helper falls back
// to os.Getwd().
func TestResolveQualityProjectDir_FallsBackToGetwd(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	cfg := GateConfig{}
	got := resolveQualityProjectDir(cfg, "test")
	if got == "" {
		t.Fatal("expected non-empty cwd fallback, got empty string")
	}
}
