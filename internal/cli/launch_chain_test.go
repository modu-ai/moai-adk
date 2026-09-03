package cli

// Tests for the launch-path chain population (SPEC-CHAIN-CORE-001
// REQ-CHAIN-005 Path A, card t242): the launcher appends a node-enter event
// at a worktree launch boundary and sets MOAI_CHAIN_NODE_ID on the child
// environment.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
)

// overrideProjectRoot points findProjectRootFn at a temp directory for the
// test's duration (the cc_test.go pattern) and pins CLAUDE_PROJECT_DIR to the
// same tree so resolveChainStore — which resolves via findStateDir — writes
// the ledger there too, never to the real project.
func overrideProjectRoot(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { findProjectRootFn = origFn })
	t.Setenv(config.EnvClaudeProjectDir, tmpDir)
	return tmpDir
}

// reopenLedger reopens the chain ledger under the given project root and
// returns all nodes built from it.
func reopenLedger(t *testing.T, root string) []chain.WorktreeNode {
	t.Helper()
	store, err := chain.NewStore(filepath.Join(root, ".moai", "state", "chain", ChainEventsFile))
	if err != nil {
		t.Fatalf("reopen ledger: %v", err)
	}
	return store.BuildNodes()
}

// envValue returns the value for key in env, or "" when absent.
func envValue(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

func TestInjectChainNodeForLaunchCreatesRootNode(t *testing.T) {
	root := overrideProjectRoot(t)

	env := injectChainNodeForLaunch(
		[]string{"--worktree", "feat-login"}, []string{"PATH=/bin"}, nil)

	nodeID := envValue(env, config.EnvChainNodeID)
	if nodeID == "" {
		t.Fatalf("expected %s set on the child env, env=%v", config.EnvChainNodeID, env)
	}

	nodes := reopenLedger(t, root)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node in ledger, got %d", len(nodes))
	}
	n := nodes[0]
	if n.NodeID != nodeID {
		t.Errorf("ledger node %q != env node %q", n.NodeID, nodeID)
	}
	// A parentless launch records depth 1 (parent depth 0 + 1) — the same
	// behavior TestCreateNodeAtSpawnDepth0 asserts in internal/chain. Note:
	// design.md §3 describes this case as a "depth-0 root node"; the code,
	// its test, and the ledger disagree with that prose. Recorded in the
	// t242 verdict as a separate documentation mismatch.
	if n.Depth != 1 {
		t.Errorf("root node depth = %d, want 1 (first node)", n.Depth)
	}
	if n.ParentNodeID != "" {
		t.Errorf("root node parent = %q, want empty", n.ParentNodeID)
	}
	wantPath := filepath.Join(root, ".claude", "worktrees", "feat-login")
	if n.WorktreePath != wantPath {
		t.Errorf("worktree_path = %q, want %q", n.WorktreePath, wantPath)
	}
}

func TestInjectChainNodeForLaunchNestedDepth(t *testing.T) {
	root := overrideProjectRoot(t)

	// First launch: depth-0 root (no parent in the environment).
	t.Setenv(config.EnvChainNodeID, "")
	parentEnv := injectChainNodeForLaunch(
		[]string{"--worktree", "wt-parent"}, nil, nil)
	parentID := envValue(parentEnv, config.EnvChainNodeID)
	if parentID == "" {
		t.Fatal("parent launch produced no node ID")
	}

	// Nested launch from a context carrying the parent's node ID.
	t.Setenv(config.EnvChainNodeID, parentID)
	childEnv := injectChainNodeForLaunch(
		[]string{"--worktree", "wt-child"}, nil, nil)
	childID := envValue(childEnv, config.EnvChainNodeID)
	if childID == "" {
		t.Fatal("child launch produced no node ID")
	}
	if childID == parentID {
		t.Fatal("child reused the parent node ID")
	}

	nodes := reopenLedger(t, root)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes in ledger, got %d", len(nodes))
	}
	for _, n := range nodes {
		if n.NodeID != childID {
			continue
		}
		if n.Depth != 2 {
			t.Errorf("child depth = %d, want 2 (parent at 1)", n.Depth)
		}
		if n.ParentNodeID != parentID {
			t.Errorf("child parent = %q, want %q", n.ParentNodeID, parentID)
		}
		last := len(n.OriginChain) - 1
		if last < 1 || n.OriginChain[last] != childID || n.OriginChain[last-1] != parentID {
			t.Errorf("child origin_chain = %v, want [..., parent, child]", n.OriginChain)
		}
	}
}

func TestInjectChainNodeForLaunchNoWorktreeTarget(t *testing.T) {
	root := overrideProjectRoot(t)

	cases := []struct {
		name string
		args []string
	}{
		{"no flag", []string{"--model", "opus"}},
		{"bare --worktree", []string{"--worktree"}},
		{"bare --worktree before flag", []string{"--worktree", "--model", "opus"}},
		{"bare -w", []string{"-w"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := injectChainNodeForLaunch(tc.args, []string{"PATH=/bin"}, nil)
			if envValue(env, config.EnvChainNodeID) != "" {
				t.Errorf("no worktree target: env was modified: %v", env)
			}
		})
	}

	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "chain", ChainEventsFile)); !os.IsNotExist(err) {
		t.Errorf("no-target launch wrote a ledger, stat err=%v", err)
	}
}

func TestInjectChainNodeForLaunchAbsoluteTarget(t *testing.T) {
	root := overrideProjectRoot(t)

	abs := filepath.Join(t.TempDir(), "l2-tree")
	env := injectChainNodeForLaunch([]string{"--worktree", abs}, nil, nil)
	nodeID := envValue(env, config.EnvChainNodeID)
	if nodeID == "" {
		t.Fatal("absolute-target launch produced no node ID")
	}

	nodes := reopenLedger(t, root)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node in ledger, got %d", len(nodes))
	}
	if nodes[0].WorktreePath != abs {
		t.Errorf("worktree_path = %q, want the absolute target %q", nodes[0].WorktreePath, abs)
	}
}

func TestInjectChainNodeForLaunchFailOpen(t *testing.T) {
	root := overrideProjectRoot(t)

	// Make the chain state directory path un-creatable: a file where the
	// directory would go. The launch must proceed (env unchanged) with a
	// warning.
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "state", "chain"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var warn bytes.Buffer
	env := injectChainNodeForLaunch(
		[]string{"--worktree", "feat-login"}, []string{"PATH=/bin"}, &warn)

	if envValue(env, config.EnvChainNodeID) != "" {
		t.Errorf("fail-open violated: env was modified: %v", env)
	}
	if !strings.Contains(warn.String(), "chain node not recorded") {
		t.Errorf("expected a fail-open warning, got %q", warn.String())
	}
}

func TestExtractWorktreeLaunchTarget(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"two-token form", []string{"claude", "--worktree", "feat-x"}, "feat-x"},
		{"equals form", []string{"claude", "--worktree=feat-x"}, "feat-x"},
		{"first wins", []string{"--worktree", "a", "--worktree", "b"}, "a"},
		{"bare flag", []string{"--worktree"}, ""},
		{"no flag", []string{"--model", "opus"}, ""},
		{"after pass-through marker", []string{"--", "--worktree", "x"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractWorktreeLaunchTarget(tc.args); got != tc.want {
				t.Errorf("extractWorktreeLaunchTarget(%v) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}
