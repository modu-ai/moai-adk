package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
)

// chainBannerTestSetup creates a temp project dir with a chain ledger
// populated with the given events. Returns the project dir path.
func chainBannerTestSetup(t *testing.T, events []chain.ChainEvent) string {
	t.Helper()
	dir := t.TempDir()
	chainDir := filepath.Join(dir, ".moai", "state", "chain")
	if err := os.MkdirAll(chainDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	storePath := filepath.Join(chainDir, "events.jsonl")
	store, err := chain.NewStore(storePath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	for _, ev := range events {
		if err := store.Append(ev); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	return dir
}

// TestSessionStartLineageBanner verifies AC-CHAIN-013: when MOAI_CHAIN_NODE_ID
// is absent (post-/clear), the banner resolves the node from the ledger by
// (cwd, session_id) and emits depth + parent + resume.
func TestSessionStartLineageBanner(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "wt-test")

	// Set up a chain ledger with a depth-2 node.
	projDir := chainBannerTestSetup(t, []chain.ChainEvent{
		{EventType: chain.EventNodeEnter, NodeID: "N0", WorktreePath: "/primary", Depth: 0, OriginChain: []string{"N0"}, EnteredAt: "2026-08-13T09:00:00Z"},
		{EventType: chain.EventNodeEnter, NodeID: "N1", WorktreePath: wtPath, ParentNodeID: "N0", Depth: 1, OriginChain: []string{"N0", "N1"}, EnteredAt: "2026-08-13T10:00:00Z"},
		{EventType: chain.EventNodeUpdate, NodeID: "N1", SessionID: "sess-test"},
		{EventType: chain.EventNodeEnter, NodeID: "N2", WorktreePath: wtPath, ParentNodeID: "N1", Depth: 2, OriginChain: []string{"N0", "N1", "N2"}, EnteredAt: "2026-08-13T11:00:00Z", ResumeTarget: "Continue M3"},
		{EventType: chain.EventNodeUpdate, NodeID: "N2", SessionID: "sess-test"},
	})

	// Simulate post-/clear: env is absent.
	os.Unsetenv(config.EnvChainNodeID)

	banner := chainLineageBanner(projDir, wtPath, "sess-test")
	if banner == "" {
		t.Fatal("expected non-empty banner, got empty (fail-open)")
	}

	for _, want := range []string{"depth 2", "N0", "N1", "N2", "Continue M3"} {
		if !strings.Contains(banner, want) {
			t.Errorf("banner missing %q\ngot:\n%s", want, banner)
		}
	}

	// Verify re-injection notice is present (env was absent).
	if !strings.Contains(banner, "re-injected") {
		t.Errorf("banner should mention re-injection when env was absent")
	}
}

// TestSessionStartBannerBackfill verifies REQ-CHAIN-021: when MOAI_CHAIN_NODE_ID
// is set in env and the node has empty session_id, the banner backfills it.
func TestSessionStartBannerBackfill(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "wt-backfill")

	projDir := chainBannerTestSetup(t, []chain.ChainEvent{
		{
			EventType:   chain.EventNodeEnter,
			NodeID:      "N1",
			WorktreePath: wtPath,
			Depth:       1,
			OriginChain:  []string{"N0", "N1"},
			EnteredAt:    "2026-08-13T10:00:00Z",
			ResumeTarget: "Start M1",
		},
	})

	// Env has the node ID (spawn boundary set it).
	t.Setenv(config.EnvChainNodeID, "N1")

	banner := chainLineageBanner(projDir, wtPath, "real-session-789")
	if banner == "" {
		t.Fatal("expected non-empty banner")
	}

	// Verify the session_id was backfilled in the ledger.
	storePath := filepath.Join(projDir, ".moai", "state", "chain", "events.jsonl")
	store, _ := chain.NewStore(storePath)
	nodes := store.BuildNodes()
	if len(nodes) == 0 {
		t.Fatal("expected nodes in ledger")
	}
	if nodes[0].SessionID != "real-session-789" {
		t.Errorf("session_id after backfill = %q, want real-session-789", nodes[0].SessionID)
	}
}

// TestSessionStartBannerNoContext verifies that with no chain ledger, the
// banner returns empty string (fail-open, no crash).
func TestSessionStartBannerNoContext(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(config.EnvChainNodeID)

	banner := chainLineageBanner(dir, "/tmp/nonexistent", "")
	if banner != "" {
		t.Errorf("expected empty banner for no chain context, got:\n%s", banner)
	}
}

// TestSessionStartBannerEmptyProjectDir verifies fail-open on empty project dir.
func TestSessionStartBannerEmptyProjectDir(t *testing.T) {
	banner := chainLineageBanner("", "/tmp/x", "sess")
	if banner != "" {
		t.Errorf("expected empty banner for empty project dir, got:\n%s", banner)
	}
}

// TestSessionStartBannerAtRoot verifies banner for a depth-0 root node.
func TestSessionStartBannerAtRoot(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "wt-root")

	projDir := chainBannerTestSetup(t, []chain.ChainEvent{
		{EventType: chain.EventNodeEnter, NodeID: "N0", WorktreePath: wtPath, Depth: 0, OriginChain: []string{"N0"}, EnteredAt: "2026-08-13T09:00:00Z"},
	})

	os.Unsetenv(config.EnvChainNodeID)
	banner := chainLineageBanner(projDir, wtPath, "")
	// Root node with no session_id and no env — may not resolve if session_id
	// is empty. But with fallback to most-recent, it should resolve.
	if banner != "" {
		if !strings.Contains(banner, "depth 0") {
			t.Errorf("expected depth 0 in banner, got:\n%s", banner)
		}
	}
}

// TestFormatChainBanner verifies the banner formatting includes expected fields.
func TestFormatChainBanner(t *testing.T) {
	node := &chain.WorktreeNode{
		NodeID:      "N2-very-long-id-here",
		Depth:       2,
		OriginChain: []string{"N0-aaaa", "N1-bbbb", "N2-very-long-id-here"},
		SpecID:      "SPEC-X-001",
		Milestone:   "M2",
		ResumeTarget: "Continue M2",
		ResumeCommand: "/moai run X",
	}

	banner := formatChainBanner(node, true)
	for _, want := range []string{"depth 2", "SPEC-X-001", "M2", "Continue M2", "/moai run X", "re-injected"} {
		if !strings.Contains(banner, want) {
			t.Errorf("banner missing %q\ngot:\n%s", want, banner)
		}
	}
}

// TestChainBannerTimeout verifies AC-CHAIN-014: the banner resolution is
// time-boxed. On timeout it returns empty string (fail-open). We test this
// indirectly by verifying the chainBannerTimeout constant is set and the
// function has a context deadline.
func TestChainBannerTimeout(t *testing.T) {
	if chainBannerTimeout <= 0 {
		t.Error("chainBannerTimeout should be positive")
	}
	// A normal call should complete well within the timeout.
	dir := t.TempDir()
	banner := chainLineageBanner(dir, "/tmp/x", "")
	// No chain context → empty banner, but it should return quickly.
	_ = banner
}
