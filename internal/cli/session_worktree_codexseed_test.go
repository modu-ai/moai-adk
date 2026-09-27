package cli

// SPEC-HANDOFF-NEUTRAL-001 M1.3 — worktree .codex/hooks.json seeding tests
// (AC-HN-005/006/007, design D2.5 option A + the launcher-entry backfill).
//
// The seeding reuses codexwiring's hooks.json generation (RenderHooks): a
// freshly materialized tree receives the MoAI-owned table, an existing file is
// left byte-identical (user-owned entries preserved), and a seeding failure
// never blocks tree creation or launcher entry (fail-open).

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedTestSeams installs the materializer seams over a temp project root with
// the REAL seeding implementation active, plus neutral M7 git-config stubs so
// no global git config is touched.
func seedTestSeams(t *testing.T, root string, addOverride func(destDir string) error) {
	t.Helper()
	swapSessionWorktreeSeams(t, swSeams{
		short:     func() string { return "deadbeef" },
		commonDir: func() (string, error) { return filepath.Join(root, ".git"), nil },
		add: func(destDir, _, _ string) (string, error) {
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return "", err
			}
			if addOverride != nil {
				if err := addOverride(destDir); err != nil {
					return "", err
				}
			}
			abs, err := filepath.Abs(destDir)
			if err != nil {
				return destDir, nil
			}
			return abs, nil
		},
		configSet: func(string, string, string) error { return nil },
		seedHooks: seedCodexHooksReal,
	})
	swapM7Seams(t, m7Seams{
		safeDirAdd: func(string) error { return nil },
		globalGet:  func(string) string { return "" },
		gitVersion: func() gitVersionInfo { return gitVersionInfo{Major: 2, Minor: 50, Patch: 0} },
	})
}

// TestNew_SeedsCodexHooksJson is AC-HN-005: materializing a worktree seeds
// .codex/hooks.json carrying the MoAI-owned handlers (generation logic reused
// from internal/codexwiring, not duplicated here).
func TestNew_SeedsCodexHooksJson(t *testing.T) {
	root := t.TempDir()
	seedTestSeams(t, root, nil)

	var out bytes.Buffer
	wtPath, err := materializeSessionWorktree("WT-deadbeef-seed", &out)
	if err != nil {
		t.Fatalf("materializeSessionWorktree: %v (out: %s)", err, out.String())
	}

	hooksPath := filepath.Join(wtPath, ".codex", "hooks.json")
	data, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("new worktree must carry .codex/hooks.json: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("seeded hooks.json must be valid JSON: %v (%s)", err, data)
	}
	if _, ok := doc["description"]; !ok {
		t.Error("seeded hooks.json must carry the MoAI ownership description")
	}
	hooks, ok := doc["hooks"].(map[string]any)
	if !ok || len(hooks) == 0 {
		t.Fatalf("seeded hooks.json must carry the MoAI event table; got %v", doc["hooks"])
	}
	if !strings.Contains(string(data), `"moai hook `) {
		t.Errorf("seeded hooks.json must contain moai hook commands; got %s", data)
	}
}

// admittedEntrySeed applies the launcher-entry backfill at its post-admission
// position (audit F0): the entry flows call seedAdmittedWorktreeHooks only
// after the concurrent-writer admission succeeds — resolveWorktreeL2Path is
// validation-only. The admission order itself is the guard test's
// (TestWorktreeLaunchRejectsConcurrentWriter) assertion; this helper covers the
// backfill behavior.
func admittedEntrySeed(t *testing.T, args []string, warn *bytes.Buffer) {
	t.Helper()
	seedAdmittedWorktreeHooks(args, warn)
}

// TestEnterWorktree_SeedsMissingCodexHooks is AC-HN-006: entering an existing
// tree through the launcher resolution path backfills a missing hooks.json,
// and an existing file stays byte-identical (user-owned entries preserved).
func TestEnterWorktree_SeedsMissingCodexHooks(t *testing.T) {
	root := t.TempDir()
	origFind := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	t.Cleanup(func() { findProjectRootFn = origFind })

	tree := filepath.Join(root, ".claude", "worktrees", "seed-entry")
	if err := os.MkdirAll(tree, 0o755); err != nil {
		t.Fatalf("mkdir tree: %v", err)
	}
	hooksPath := filepath.Join(tree, ".codex", "hooks.json")

	// Short-name entry into a tree without wiring: seeded.
	var warn bytes.Buffer
	admittedEntrySeed(t, []string{"--worktree", "seed-entry"}, &warn)
	if _, err := os.Stat(hooksPath); err != nil {
		t.Fatalf("launcher entry must backfill a missing .codex/hooks.json: %v", err)
	}

	// Existing user-owned file: byte-identical after another entry.
	userDoc := `{"description":"user desc","hooks":{"UserEvent":[{"hooks":[{"type":"command","command":"user-tool"}]}]}}`
	if err := os.WriteFile(hooksPath, []byte(userDoc), 0o600); err != nil {
		t.Fatalf("write user doc: %v", err)
	}
	admittedEntrySeed(t, []string{"--worktree", "seed-entry"}, &warn)
	after, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(after) != userDoc {
		t.Errorf("existing hooks.json must stay byte-identical; got %s", after)
	}

	// Absolute-path entry (L1 prefix): a fresh tree is seeded too.
	tree2 := filepath.Join(root, ".claude", "worktrees", "seed-entry-abs")
	if err := os.MkdirAll(tree2, 0o755); err != nil {
		t.Fatalf("mkdir tree2: %v", err)
	}
	admittedEntrySeed(t, []string{"--worktree", tree2}, &warn)
	if _, err := os.Stat(filepath.Join(tree2, ".codex", "hooks.json")); err != nil {
		t.Fatalf("absolute-path entry must seed a fresh tree: %v", err)
	}
}

// TestNew_SeedFailureFailOpen is AC-HN-007: a forced seeding failure (path
// pollution — .codex occupied by a regular file) neither blocks worktree
// creation nor stays silent; a stderr diagnostic names the failure.
func TestNew_SeedFailureFailOpen(t *testing.T) {
	root := t.TempDir()
	seedTestSeams(t, root, func(destDir string) error {
		// Path pollution: after the tree directory exists, occupy the .codex
		// path with a regular file so the seed's directory creation fails.
		return os.WriteFile(filepath.Join(destDir, ".codex"), []byte("occupied"), 0o644)
	})

	var out bytes.Buffer
	wtPath, err := materializeSessionWorktree("WT-deadbeef-seedfail", &out)
	if err != nil {
		t.Fatalf("seeding failure must not block worktree creation: %v (out: %s)", err, out.String())
	}
	if wtPath == "" {
		t.Fatal("materialization must still return the worktree path")
	}
	if !strings.Contains(out.String(), "codex hooks") {
		t.Errorf("seeding failure must emit a stderr diagnostic; got %q", out.String())
	}
	if _, statErr := os.Stat(filepath.Join(wtPath, ".codex", "hooks.json")); statErr == nil {
		t.Error("polluted path must not have produced a hooks.json")
	}
}
