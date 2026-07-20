package cli

// SPEC-CLIFIX-CONCURRENCY-001 M1 — Reproduction tests (RED phase).
//
// Two reproduction tests written BEFORE the M1 fix (TDD RED-GREEN-REFACTOR):
//
//   1. TestSettingsLocalConcurrentWrites — N-goroutine concurrent writers hit
//      two production functions (ensureSettingsLocalJSON + syncPermissionModeTo-
//      SettingsLocal) that currently call writeSettingsMap (plain os.WriteFile,
//      no lock, no temp+rename). The un-serialized read-modify-write loses one
//      writer's keys or produces a truncated file. FAILS pre-M1; PASSES post-M1
//      (both functions re-routed through the locked+atomic mutateSettingsLocal
//      seam).
//
//   2. TestRemoveGLMEnvComplete — removeGLMEnv's delete set is missing
//      CLAUDE_CODE_AUTO_COMPACT_WINDOW (config.EnvClaudeCodeAutoCompactWindow),
//      so a 1M auto-compact window injected by moai glm/cg persists into
//      subsequent moai cc sessions. FAILS pre-M1 (key left behind); PASSES
//      post-M1 (key added to the delete set).
//
// Per SPEC §D.5, each test MUST demonstrably fail against the pre-fix commit.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// =========================================================================
// AC-CONC-001-001 — Concurrent writers must not lose updates or truncate.
// =========================================================================

// TestSettingsLocalConcurrentWrites verifies that concurrent read-modify-write
// operations on settings.local.json through the production launch-path helpers
// do not lose updates or produce truncated output.
//
// Pre-M1: ensureSettingsLocalJSON and syncPermissionModeToSettingsLocal both
// call writeSettingsMap (plain os.WriteFile, no lock). Concurrent RMW loses the
// last-loser's top-level keys — this test FAILS.
//
// Post-M1: both helpers route through mutateSettingsLocal (flock + temp+rename),
// serializing the RMW so every key survives — this test PASSES.
func TestSettingsLocalConcurrentWrites(t *testing.T) {
	tmp := t.TempDir()
	settingsPath := filepath.Join(tmp, "settings.local.json")
	// Seed with a base key so the file starts non-empty (mirrors real usage
	// where settings.local.json already exists when a second session races).
	mustWriteFile(t, settingsPath, `{"base":"seed"}`)

	const N = 100
	var wg sync.WaitGroup
	wg.Add(N * 2)
	start := make(chan struct{})
	var firstErr sync.Map // keyed by goroutine index; stores error

	// Writer family A: ensureSettingsLocalJSON sets teammateMode="tmux".
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			if err := ensureSettingsLocalJSON(settingsPath); err != nil {
				firstErr.Store(i, err)
			}
		}(i)
	}
	// Writer family B: syncPermissionModeToSettingsLocal sets
	// permissions.defaultMode="bypassPermissions".
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			if err := syncPermissionModeToSettingsLocal(settingsPath, "bypassPermissions"); err != nil {
				firstErr.Store(i+N, err)
			}
		}(i)
	}
	close(start) // release all goroutines simultaneously to maximize interleaving
	wg.Wait()

	firstErr.Range(func(_, val any) bool {
		t.Fatalf("concurrent writer returned error (truncated read mid-write): %v", val)
		return false
	})

	// Final file must be valid JSON (no truncation) and contain BOTH writers'
	// top-level keys (no lost update).
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read final settings.local.json: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("final settings.local.json is truncated/invalid JSON (concurrent write corruption): %v\nraw (%d bytes): %q", err, len(data), string(data))
	}

	if tm, _ := result["teammateMode"].(string); tm != "tmux" {
		t.Errorf("teammateMode lost or wrong after concurrent writes: got %v, want %q (lost update — RMW not serialized)", result["teammateMode"], "tmux")
	}
	perms, _ := result["permissions"].(map[string]any)
	if perms == nil || perms["defaultMode"] != "bypassPermissions" {
		t.Errorf("permissions.defaultMode lost after concurrent writes: got %v (lost update — RMW not serialized)", perms)
	}
}

// =========================================================================
// AC-CONC-001-002 — removeGLMEnv must clear CLAUDE_CODE_AUTO_COMPACT_WINDOW.
// =========================================================================

// TestRemoveGLMEnvComplete verifies that removeGLMEnv removes every GLM-era env
// key including CLAUDE_CODE_AUTO_COMPACT_WINDOW (config.EnvClaudeCodeAutoCom-
// pactWindow), so a 1M auto-compact window does not persist into subsequent
// moai cc sessions.
//
// Pre-M1: the delete set at launcher.go:268-281 omits the auto-compact key, so
// it survives removeGLMEnv — this test FAILS.
//
// Post-M1: the key is added to the delete set via config.EnvClaudeCodeAutoCom-
// pactWindow — this test PASSES.
func TestRemoveGLMEnvComplete(t *testing.T) {
	tmp := t.TempDir()
	settingsPath := filepath.Join(tmp, "settings.local.json")
	// Fixture mirrors what moai glm/cg injects, including the 1M auto-compact
	// window key that REQ-CONC-001-002 targets.
	original := `{
  "teammateMode": "tmux",
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "glm-key",
    "ANTHROPIC_BASE_URL": "https://open.z.ai/api",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-4.6",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.5-air",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4-flash",
    "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1",
    "API_TIMEOUT_MS": "3000000",
    "CLAUDE_CODE_AUTO_COMPACT_WINDOW": "1000000",
    "MOAI_STATUSLINE_CONTEXT_SIZE": "1000000"
  }
}`
	mustWriteFile(t, settingsPath, original)

	if err := removeGLMEnv(settingsPath); err != nil {
		t.Fatalf("removeGLMEnv: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read after removeGLMEnv: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse after removeGLMEnv: %v\nraw: %s", err, data)
	}

	env, _ := m["env"].(map[string]any)
	if _, stillPresent := env[config.EnvClaudeCodeAutoCompactWindow]; stillPresent {
		t.Errorf("CLAUDE_CODE_AUTO_COMPACT_WINDOW was NOT removed by removeGLMEnv — a 1M auto-compact window would persist into moai cc sessions (leftover key defect)")
	}
}
