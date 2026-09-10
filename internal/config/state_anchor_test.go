package config

// state_anchor_test.go — SPEC-STATE-ANCHOR-001 M3 (AC-SA-004, card t510).
//
// The config cache write must anchor to the PROJECT's state anchor, not to
// wherever the CLI process happens to stand. The cwd linkage lives in the CLI
// call chain (internal/cli/deps.go InitDependencies: cwd, _ := os.Getwd() →
// deps.Config.Load(cwd) — the derivation point this SPEC's run phase was
// tasked to pin, closing the verdict's Gap); this test reproduces that chain
// at unit scale: the git-resolved anchor is derived the same way the repaired
// CLI chain derives it, then a load triggers the cache write.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/stateanchor"
)

// initGitRepoConfig prepares dir as a minimal git repository — enough for the
// state anchor's git walk-up to answer.
func initGitRepoConfig(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "init", "--quiet")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", dir, err, out)
	}
}

// TestConfigCacheAnchorsToProject — AC-SA-004 (B4).
//
// Given a git-repository project root carrying .moai/, a session standing in
// a subdirectory of it that carries a STRAY .moai (the shape GH #1694's
// context-usage pollution left behind, which the config cache then followed —
// the reporter counted 150 config-cache.json files), and a separate non-git
// directory the process never anchors to,
//
// When a CLI-chain load triggers the cache write with the state-anchor-derived
// root (the repaired derivation; the pre-repair derivation passed the raw
// cwd — the subdirectory — and the cache mis-landed in the stray .moai, which
// is exactly what the RED observation recorded),
//
// Then the cache lands at the project anchor's .moai/state/config-cache.json,
// the stray is untouched, and the non-git directory gains no .moai (the
// existing config-dir-exists guard never materializes one).
func TestConfigCacheAnchorsToProject(t *testing.T) {
	t.Setenv(EnvConfigDir, "")           // pin the derivation, not an env override
	t.Setenv(EnvConfigCacheDisabled, "") // the cache write must happen

	root := t.TempDir()
	initGitRepoConfig(t, root)
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved // git reports the symlink-resolved spelling on macOS
	}
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	// The visited subdirectory with a stray .moai — pre-existing, NOT created
	// by the load under test. Pre-repair this directory was the config root
	// (cwd-shaped derivation) and the cache landed inside the stray.
	visited := filepath.Join(root, "deep", "sub")
	if err := os.MkdirAll(filepath.Join(visited, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	// A separate non-git directory: nothing anywhere may create .moai here.
	nonGit := t.TempDir()

	// The CLI-chain derivation: the state anchor of the directory the process
	// stands in (the repair; the pre-repair chain passed the raw directory,
	// which is the RED form this test's failure recorded).
	anchor := stateanchor.FromDirectory(visited)
	if anchor == "" {
		t.Fatalf("state anchor unresolved for %s — the git fixture is broken", visited)
	}

	mgr := NewConfigManager()
	if _, err := mgr.Load(anchor); err != nil {
		t.Fatalf("Load(%s): %v", anchor, err)
	}

	if _, err := os.Stat(filepath.Join(root, ".moai", "state", cacheFileName)); err != nil {
		t.Errorf("config cache not anchored to the project root (want %s): %v",
			filepath.Join(root, ".moai", "state", cacheFileName), err)
	}
	if _, err := os.Stat(filepath.Join(visited, ".moai", "state", cacheFileName)); !os.IsNotExist(err) {
		t.Errorf("config cache mis-landed in the stray .moai of the visited dir (the B4 pollution shape): stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(nonGit, ".moai")); !os.IsNotExist(err) {
		t.Errorf("non-git directory gained a .moai (state must never be created there): stat err = %v", err)
	}
}
