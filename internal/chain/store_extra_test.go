package chain

import (
	"os"
	"strings"
	"testing"
)

// TestResolveByCWDEmptySessionID verifies the CWD-collision fallback: when
// session_id is empty, the most recently entered node for the worktree_path
// is returned.
func TestResolveByCWDEmptySessionID(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)
	wtPath := "/tmp/wt-fallback"

	s.Append(ChainEvent{EventType: EventNodeEnter, NodeID: "old", WorktreePath: wtPath, SessionID: "s1"})
	s.Append(ChainEvent{EventType: EventNodeEnter, NodeID: "new", WorktreePath: wtPath, SessionID: "s2"})

	node, err := s.ResolveByCWD(wtPath, "")
	if err != nil {
		t.Fatalf("ResolveByCWD empty session: %v", err)
	}
	// Should return the most recently entered node.
	if node.NodeID != "new" {
		t.Errorf("got %q, want new (most recent)", node.NodeID)
	}
}

// TestResolveByCWDNotFound verifies ErrNodeNotFound when no node matches.
func TestResolveByCWDNotFound(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)
	s.Append(ChainEvent{EventType: EventNodeEnter, NodeID: "N1", WorktreePath: "/tmp/a"})

	_, err := s.ResolveByCWD("/nonexistent", "sess")
	if err == nil {
		t.Fatal("expected ErrNodeNotFound, got nil")
	}
	if err != ErrNodeNotFound {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

// TestResolveByCWDMismatchFallback verifies fail-open: a session_id that
// doesn't match any node for the worktree_path falls back to the most
// recent node.
func TestResolveByCWDMismatchFallback(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)
	wtPath := "/tmp/wt-mismatch"

	s.Append(ChainEvent{EventType: EventNodeEnter, NodeID: "N1", WorktreePath: wtPath, SessionID: "s1"})

	node, err := s.ResolveByCWD(wtPath, "nonexistent-session")
	if err != nil {
		t.Fatalf("expected fail-open fallback, got error: %v", err)
	}
	if node.NodeID != "N1" {
		t.Errorf("fallback NodeID = %q, want N1", node.NodeID)
	}
}

// TestPathMethod verifies Store.Path() returns the bound file path.
func TestPathMethod(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := dir + "/events.jsonl"
	s, err := NewStore(p)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if got := s.Path(); got != p {
		t.Errorf("Path() = %q, want %q", got, p)
	}
}

// TestBuildNodesUpdateNoOp verifies that a node-update for a non-existent
// node_id is silently skipped (no panic, no phantom node).
func TestBuildNodesUpdateNoOp(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)
	// node-update for a node that was never created by node-enter.
	s.Append(ChainEvent{EventType: EventNodeUpdate, NodeID: "phantom", SessionID: "s"})

	nodes := s.BuildNodes()
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes (phantom update), got %d", len(nodes))
	}
}

// TestBuildNodesMilestoneUpdate verifies node-update overlays milestone
// and last_completed_milestone fields.
func TestBuildNodesMilestoneUpdate(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)
	s.Append(ChainEvent{EventType: EventNodeEnter, NodeID: "N1", Milestone: "M0"})
	s.Append(ChainEvent{
		EventType:             EventNodeUpdate,
		NodeID:                "N1",
		Milestone:             "M1",
		LastCompletedMilestone: "M0-done",
		ResumeTarget:          "Start M2",
		ResumeCommand:         "/moai run X",
	})

	nodes := s.BuildNodes()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n := nodes[0]
	if n.Milestone != "M1" {
		t.Errorf("Milestone = %q, want M1", n.Milestone)
	}
	if n.LastCompletedMilestone != "M0-done" {
		t.Errorf("LastCompletedMilestone = %q, want M0-done", n.LastCompletedMilestone)
	}
	if n.ResumeTarget != "Start M2" {
		t.Errorf("ResumeTarget = %q, want Start M2", n.ResumeTarget)
	}
	if n.ResumeCommand != "/moai run X" {
		t.Errorf("ResumeCommand = %q, want /moai run X", n.ResumeCommand)
	}
}

// TestCompletionEdgeEvent verifies that completion-edge events can be
// appended and read back correctly.
func TestCompletionEdgeEvent(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)

	ev := ChainEvent{
		EventType:         EventCompletionEdge,
		ParentNode:        "N1",
		ChildNode:         "N2",
		CompletedMilestone: "M2",
		CompletedAt:       "2026-08-13T10:00:00Z",
		NextResumeTarget:  "Start M3",
	}
	if err := s.Append(ev); err != nil {
		t.Fatalf("Append completion-edge: %v", err)
	}

	events, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	got := events[0]
	if got.EventType != EventCompletionEdge {
		t.Errorf("EventType = %q, want completion-edge", got.EventType)
	}
	if got.ParentNode != "N1" || got.ChildNode != "N2" {
		t.Errorf("Parent/Child = %q/%q, want N1/N2", got.ParentNode, got.ChildNode)
	}
}

// TestNoAskUserQuestionInChain verifies REQ-CHAIN-018 / AC-CHAIN-019:
// the chain package source must NOT contain AskUserQuestion or
// mcp__askuser tokens. This is the static grep guard.
func TestNoAskUserQuestionInChain(t *testing.T) {
	t.Parallel()
	files := []string{
		"node.go",
		"store.go",
	}
	for _, fname := range files {
		data, err := os.ReadFile(fname)
		if err != nil {
			t.Fatalf("read %s: %v", fname, err)
		}
		if strings.Contains(string(data), "AskUserQuestion") {
			t.Errorf("%s contains AskUserQuestion — orchestrator boundary violated", fname)
		}
		if strings.Contains(string(data), "mcp__askuser") {
			t.Errorf("%s contains mcp__askuser — orchestrator boundary violated", fname)
		}
	}
}
