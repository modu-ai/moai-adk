package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
)

// setupChainTestDir creates a temp project root with .moai/state/chain/
// and returns the path. Sets CLAUDE_PROJECT_DIR for test isolation.
func setupChainTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	chainDir := filepath.Join(dir, ".moai", "state", "chain")
	if err := os.MkdirAll(chainDir, 0o755); err != nil {
		t.Fatalf("mkdir chain dir: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, dir)
	return dir
}

// populateChainLedger writes the given events to the chain store under the
// test project root.
func populateChainLedger(t *testing.T, dir string, events []chain.ChainEvent) {
	t.Helper()
	storePath := filepath.Join(dir, ".moai", "state", "chain", "events.jsonl")
	store, err := chain.NewStore(storePath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	for _, ev := range events {
		if err := store.Append(ev); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
}

// TestChainStatus verifies AC-CHAIN-007: moai chain status prints the current
// node summary with depth, parent, spec, milestone, resume.
func TestChainStatus(t *testing.T) {
	dir := setupChainTestDir(t)
	populateChainLedger(t, dir, []chain.ChainEvent{
		{
			EventType:   chain.EventNodeEnter,
			NodeID:      "N1",
			WorktreePath: dir,
			SpecID:       "SPEC-AUTH-001",
			Milestone:    "M2",
			Depth:       1,
			EnteredAt:    "2026-08-13T10:00:00Z",
			ResumeTarget: "Continue M2",
		},
	})
	t.Setenv(config.EnvChainNodeID, "N1")

	var buf bytes.Buffer
	if err := runChainStatus(&buf); err != nil {
		t.Fatalf("runChainStatus: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"depth:", "SPEC-AUTH-001", "M2", "Continue M2"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

// TestChainLineage verifies AC-CHAIN-008: moai chain lineage prints
// root-to-leaf chain with node details.
func TestChainLineage(t *testing.T) {
	dir := setupChainTestDir(t)
	populateChainLedger(t, dir, []chain.ChainEvent{
		{EventType: chain.EventNodeEnter, NodeID: "N0", WorktreePath: "/primary", SpecID: "SPEC-X", Depth: 0, OriginChain: []string{"N0"}},
		{EventType: chain.EventNodeEnter, NodeID: "N1", WorktreePath: "/wt-1", ParentNodeID: "N0", SpecID: "SPEC-X", Depth: 1, OriginChain: []string{"N0", "N1"}},
		{EventType: chain.EventNodeEnter, NodeID: "N2", WorktreePath: dir, ParentNodeID: "N1", SpecID: "SPEC-X", Depth: 2, OriginChain: []string{"N0", "N1", "N2"}},
	})
	t.Setenv(config.EnvChainNodeID, "N2")

	var buf bytes.Buffer
	if err := runChainLineage(&buf); err != nil {
		t.Fatalf("runChainLineage: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"N0", "N1", "N2", "origin chain"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

// TestChainBack verifies AC-CHAIN-009: moai chain back prints parent resume
// target and command.
func TestChainBack(t *testing.T) {
	dir := setupChainTestDir(t)
	populateChainLedger(t, dir, []chain.ChainEvent{
		{
			EventType:     chain.EventNodeEnter,
			NodeID:        "N1",
			WorktreePath:  "/wt-parent",
			Depth:         1,
			ResumeTarget:  "Continue M2",
			ResumeCommand: "/moai run SPEC-AUTH-001",
		},
		{
			EventType:   chain.EventNodeEnter,
			NodeID:      "N2",
			WorktreePath: dir,
			ParentNodeID: "N1",
			Depth:       2,
		},
	})
	t.Setenv(config.EnvChainNodeID, "N2")

	var buf bytes.Buffer
	if err := runChainBack(&buf); err != nil {
		t.Fatalf("runChainBack: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Continue M2", "/moai run SPEC-AUTH-001"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

// TestChainBackAtRoot verifies the edge case: at root, prints "no parent".
func TestChainBackAtRoot(t *testing.T) {
	dir := setupChainTestDir(t)
	populateChainLedger(t, dir, []chain.ChainEvent{
		{EventType: chain.EventNodeEnter, NodeID: "N0", WorktreePath: dir, Depth: 0},
	})
	t.Setenv(config.EnvChainNodeID, "N0")

	var buf bytes.Buffer
	_ = runChainBack(&buf)
	if !strings.Contains(buf.String(), "no parent") {
		t.Errorf("expected 'no parent', got:\n%s", buf.String())
	}
}

// TestChainList verifies AC-CHAIN-010: moai chain list enumerates all nodes.
func TestChainList(t *testing.T) {
	dir := setupChainTestDir(t)
	populateChainLedger(t, dir, []chain.ChainEvent{
		{EventType: chain.EventNodeEnter, NodeID: "N0", WorktreePath: "/primary", SessionID: "s0", Depth: 0},
		{EventType: chain.EventNodeEnter, NodeID: "N1", WorktreePath: "/wt-1", SessionID: "s1", Depth: 1, SpecID: "SPEC-A"},
		{EventType: chain.EventNodeEnter, NodeID: "N2", WorktreePath: "/wt-2", SessionID: "", Depth: 2, SpecID: "SPEC-B"},
	})

	var buf bytes.Buffer
	if err := runChainList(&buf); err != nil {
		t.Fatalf("runChainList: %v", err)
	}
	out := buf.String()
	// All 3 nodes should appear.
	for _, want := range []string{"N0", "N1", "N2", "DEPTH"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

// TestChainStatusNoContext verifies the edge case: empty ledger prints
// "no chain context".
func TestChainStatusNoContext(t *testing.T) {
	setupChainTestDir(t)
	t.Setenv(config.EnvChainNodeID, "")

	var buf bytes.Buffer
	_ = runChainStatus(&buf)
	if !strings.Contains(buf.String(), "no chain context") {
		t.Errorf("expected 'no chain context', got:\n%s", buf.String())
	}
}

// TestChainRemoteCWD verifies AC-CHAIN-017: remote CWD surfaces limitation.
func TestChainRemoteCWD(t *testing.T) {
	setupChainTestDir(t)
	// Simulate a remote CWD that doesn't exist locally.
	t.Setenv(config.EnvClaudeProjectDir, "ssh://remote-host/project")

	var buf bytes.Buffer
	_ = runChainStatus(&buf)
	out := buf.String()
	if !strings.Contains(out, "single-host") || !strings.Contains(out, "remote") {
		t.Errorf("expected single-host limitation notice, got:\n%s", out)
	}
}

// TestChainPruneDryRun verifies prune dry-run identifies exited nodes.
func TestChainPruneDryRun(t *testing.T) {
	dir := setupChainTestDir(t)
	// Create old exited nodes (31 days ago, no session_id = exited).
	oldTime := time.Now().UTC().AddDate(0, 0, -31).Format(time.RFC3339)
	populateChainLedger(t, dir, []chain.ChainEvent{
		{EventType: chain.EventNodeEnter, NodeID: "N0", WorktreePath: "/p", SessionID: "", Depth: 0, EnteredAt: oldTime},
		{EventType: chain.EventNodeEnter, NodeID: "N1", WorktreePath: "/w", SessionID: "", Depth: 1, EnteredAt: oldTime},
	})

	var buf bytes.Buffer
	if err := runChainPrune(&buf, true); err != nil {
		t.Fatalf("runChainPrune: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected dry-run output, got:\n%s", out)
	}
}

// TestNoAskUserQuestionInChain verifies AC-CHAIN-019: the chain CLI source
// must NOT contain AskUserQuestion or mcp__askuser tokens.
func TestNoAskUserQuestionInChain(t *testing.T) {
	data, err := os.ReadFile("chain.go")
	if err != nil {
		t.Fatalf("read chain.go: %v", err)
	}
	if strings.Contains(string(data), "AskUserQuestion") {
		t.Error("chain.go contains AskUserQuestion — orchestrator boundary violated")
	}
	if strings.Contains(string(data), "mcp__askuser") {
		t.Error("chain.go contains mcp__askuser — orchestrator boundary violated")
	}
}

// TestChainFlagAgnostic verifies AC-CHAIN-018: no -k/-f/manager-lead
// dependency in chain source.
func TestChainFlagAgnostic(t *testing.T) {
	data, err := os.ReadFile("chain.go")
	if err != nil {
		t.Fatalf("read chain.go: %v", err)
	}
	for _, token := range []string{"MOAI_KANBAN", "MOAI_FACTORY", "manager.lead", "managerLead"} {
		if strings.Contains(string(data), token) {
			t.Errorf("chain.go contains %q — flag-agnostic violation", token)
		}
	}
}

// TestClassifyStaleness verifies the staleness classification logic.
func TestClassifyStaleness(t *testing.T) {
	entries := []struct {
		name     string
		sessionID string
		expected nodeStaleness
	}{
		{"empty_session", "", stalenessExited},
		{"not_in_registry", "unknown-sess", stalenessExited},
	}
	for _, tc := range entries {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyStaleness(tc.sessionID, nil)
			if got != tc.expected {
				t.Errorf("classifyStaleness(%q) = %q, want %q", tc.sessionID, got, tc.expected)
			}
		})
	}
}

// TestTruncateID verifies the ID truncation helper.
func TestTruncateID(t *testing.T) {
	short := "abc123"
	if got := truncateID(short); got != short {
		t.Errorf("truncateID(%q) = %q, want %q", short, got, short)
	}
	long := "0123456789abcdef0123456789abcdef0123456789abcdef"
	got := truncateID(long)
	if len(got) >= len(long) {
		t.Errorf("truncateID should shorten long IDs, got %q (len %d)", got, len(got))
	}
}

// TestIsRemoteCWD verifies the remote CWD detection.
func TestIsRemoteCWD(t *testing.T) {
	if isRemoteCWD("") {
		t.Error("empty CWD should not be remote")
	}
	if isRemoteCWD("/tmp/local") {
		t.Error("local path should not be remote")
	}
	if !isRemoteCWD("ssh://host/path") {
		t.Error("ssh:// URL should be remote")
	}
}
