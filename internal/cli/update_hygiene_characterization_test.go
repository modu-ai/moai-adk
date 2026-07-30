package cli

// This file is the M1 PRESERVE safety net for SPEC-CLIFIX-HYGIENE-001.
//
// It pins observable behavior of the residual over-ceiling portion of update.go
// that M5 may later decompose. Per plan.md §F M1 + acceptance.md AC-HYG-001-001,
// the M5 decomposition MUST keep this suite byte-identical (PASS unchanged).
//
// Scope (per plan.md §F M1): the dry-run plan path AND the settings sync/merge
// path. Two surfaces are characterized:
//
//  1. cleanLegacyHooks — the settings sync/merge path's pure-function core.
//     The exact legacy-pattern set + the "return true iff modified" contract +
//     the empty-group/empty-hookType pruning semantics are pinned as goldens.
//  2. runAgencyMigrationAdapter — the dry-run plan path's no-source branch.
//     The "ErrMigrateNoSource → return nil" suppression is the load-bearing
//     contract: a project with no .agency source MUST NOT fail update.
//
// Characterization tests capture CURRENT behavior, not desired behavior. If a
// golden changes after decomposition, that is a regression to investigate, not
// a test to update. (Per moai-ref-testing-pyramid: "characterization tests
// capture current behavior".)

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// --- Settings sync/merge path: cleanLegacyHooks characterization ---

// legacyHookPatternsGolden is the exact set of legacy hook command substrings
// cleanLegacyHooks removes. Observed on undecomposed update.go (M1 baseline).
// M5 decomposition MUST preserve this set verbatim — the strings are the
// moai-handle wrapper filenames + the deprecated Python hook filenames.
var legacyHookPatternsGolden = []string{
	"handle-session-end.sh",
	"handle-session-start.sh",
	"handle-stop.sh",
	"handle-pre-tool.sh",
	"handle-post-tool.sh",
	"handle-agent-hook.sh",
	"handle-compact.sh",
	"post_tool__code_formatter.py",
	"post_tool__linter.py",
	"post_tool__ast_grep_scan.py",
}

// TestCleanLegacyHooks_NoHooksSectionCharacterization pins that a settings map
// with no "hooks" key short-circuits with (false, no mutation). This is the
// common-case fast path ensureGlobalSettingsEnv depends on.
func TestCleanLegacyHooks_NoHooksSectionCharacterization(t *testing.T) {
	settings := map[string]any{"env": map[string]any{"PATH": "/x"}}
	changed := cleanLegacyHooks(settings)
	if changed {
		t.Errorf("cleanLegacyHooks on map without hooks key = true, want false")
	}
	if _, stillThere := settings["hooks"]; stillThere {
		t.Errorf("cleanLegacyHooks injected a hooks key — must not mutate absent keys")
	}
}

// TestCleanLegacyHooks_RemovesEveryLegacyPatternCharacterization pins that EVERY
// golden legacy pattern is removed from a hook command, and that the function
// reports changed=true. This is the per-pattern coverage the settings sync path
// relies on — dropping any one pattern would silently leave a stale hook.
func TestCleanLegacyHooks_RemovesEveryLegacyPatternCharacterization(t *testing.T) {
	for _, pattern := range legacyHookPatternsGolden {
		t.Run(pattern, func(t *testing.T) {
			settings := map[string]any{
				"hooks": map[string]any{
					"SessionStart": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/" + pattern + "\"",
								},
							},
						},
					},
				},
			}
			changed := cleanLegacyHooks(settings)
			if !changed {
				t.Errorf("pattern %q: cleanLegacyHooks = false, want true (pattern must trigger removal)", pattern)
			}
			// The whole SessionStart group becomes empty → hookType pruned → "hooks" pruned.
			if _, stillThere := settings["hooks"]; stillThere {
				t.Errorf("pattern %q: expected top-level hooks key pruned, still present: %v",
					pattern, settings["hooks"])
			}
		})
	}
}

// TestCleanLegacyHooks_PreservesNonLegacyHookCharacterization pins that a hook
// command that does NOT contain any legacy pattern survives unchanged. The
// settings sync path must not over-delete user-authored or unrelated hooks.
func TestCleanLegacyHooks_PreservesNonLegacyHookCharacterization(t *testing.T) {
	userCmd := "/usr/local/bin/my-custom-hook.sh"
	settings := map[string]any{
		"hooks": map[string]any{
			"Stop": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"command": userCmd},
					},
				},
			},
		},
	}
	changed := cleanLegacyHooks(settings)
	if changed {
		t.Errorf("cleanLegacyHooks on non-legacy hook = true, want false (must not report changes when nothing removed)")
	}
	hooksMap, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("hooks key missing or wrong type after clean: %v", settings["hooks"])
	}
	stopList, ok := hooksMap["Stop"].([]any)
	if !ok || len(stopList) != 1 {
		t.Fatalf("Stop hook list shape changed: %v", hooksMap["Stop"])
	}
}

// TestCleanLegacyHooks_MixedLegacyAndUserHooksCharacterization pins the
// partial-removal contract: when a hook group mixes one legacy and one
// user-authored entry, only the legacy entry is removed and the group survives
// with the user entry intact. This is the settings sync path's surgical-delete
// guarantee.
func TestCleanLegacyHooks_MixedLegacyAndUserHooksCharacterization(t *testing.T) {
	settings := map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-post-tool.sh\""},
						map[string]any{"command": "/usr/local/bin/user-lint.sh"},
					},
				},
			},
		},
	}
	changed := cleanLegacyHooks(settings)
	if !changed {
		t.Errorf("cleanLegacyHooks on mixed group = false, want true (legacy entry must be removed)")
	}
	hooksMap := settings["hooks"].(map[string]any)
	postList := hooksMap["PostToolUse"].([]any)
	group := postList[0].(map[string]any)
	remaining := group["hooks"].([]any)
	if len(remaining) != 1 {
		t.Fatalf("expected 1 user hook remaining, got %d: %v", len(remaining), remaining)
	}
	if cmd, _ := remaining[0].(map[string]any)["command"].(string); cmd != "/usr/local/bin/user-lint.sh" {
		t.Errorf("surviving hook = %q, want the user-authored /usr/local/bin/user-lint.sh", cmd)
	}
}

// --- Dry-run plan path: runAgencyMigrationAdapter no-source characterization ---

// TestRunAgencyMigrationAdapter_NoSourceIsNilCharacterization pins the
// dry-run plan path's no-source contract: a project with NO .agency source
// directory returns nil (clean skip), NOT an error. This is the load-bearing
// suppression (ErrMigrateNoSource → return nil) that lets `moai update` proceed
// on already-migrated or never-agency projects.
//
// M5 may decompose runAgencyMigrationAdapter out of update.go; the contract
// this test pins MUST survive the move.
func TestRunAgencyMigrationAdapter_NoSourceIsNilCharacterization(t *testing.T) {
	root := t.TempDir() // no .agency, no target — clean project
	var out bytes.Buffer
	err := runAgencyMigrationAdapter(root, true /* dryRun */, false /* force */, &out)
	if err != nil {
		t.Errorf("runAgencyMigrationAdapter on project with no .agency source = %v, want nil (ErrMigrateNoSource must suppress to nil)", err)
	}
	// No source → nothing to plan → no diagnostic output expected.
	if out.Len() != 0 {
		t.Errorf("runAgencyMigrationAdapter on no-source project wrote output %q, want empty (nothing to plan)", out.String())
	}
}

// TestRunAgencyMigrationAdapter_DryRunFlagPropagatedCharacterization pins that
// the dryRun boolean is plumbed into the runner struct. With a present-but-empty
// source and dryRun=true, no filesystem mutation occurs. This is the dry-run
// plan path's "show, don't touch" guarantee that AC-HYG-001-001 references.
func TestRunAgencyMigrationAdapter_DryRunFlagPropagatedCharacterization(t *testing.T) {
	root := t.TempDir()
	// Create a minimal .agency source so the runner reaches the plan stage
	// rather than short-circuiting on ErrMigrateNoSource.
	agencyDir := filepath.Join(root, ".agency", "core")
	if err := os.MkdirAll(agencyDir, 0o755); err != nil {
		t.Fatalf("seed .agency: %v", err)
	}

	var out bytes.Buffer
	pre, _ := filepath.Glob(filepath.Join(root, ".moai") + "/*")
	err := runAgencyMigrationAdapter(root, true /* dryRun */, false /* force */, &out)
	if err != nil {
		t.Errorf("runAgencyMigrationAdapter dryRun=true returned err = %v, want nil (dry-run must not error on empty source)", err)
	}
	post, _ := filepath.Glob(filepath.Join(root, ".moai") + "/*")
	// Dry-run must not create the .moai migration target.
	if len(pre) != len(post) {
		t.Errorf("dryRun=true mutated .moai under %s: pre=%v post=%v (dry-run must be read-only)", root, pre, post)
	}
}

