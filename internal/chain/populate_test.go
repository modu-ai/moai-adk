package chain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGenerateNodeIDUniqueness verifies that GenerateNodeID produces unique
// IDs across rapid successive calls.
func TestGenerateNodeIDUniqueness(t *testing.T) {
	t.Parallel()
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := GenerateNodeID()
		if seen[id] {
			t.Fatalf("duplicate node ID generated: %s at iteration %d", id, i)
		}
		seen[id] = true
	}
}

// TestCreateNodeAtSpawnDepth0 verifies that spawning with no parent env
// creates a depth-0 root node.
func TestCreateNodeAtSpawnDepth0(t *testing.T) {

	// Ensure env is unset for this test.
	t.Setenv(config.EnvChainNodeID, "")

	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, err := NewStore(storePath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	pop := NewPopulator(store)

	nodeID, err := pop.CreateNodeAtSpawn("/tmp/wt-root", "SPEC-X", "M0")
	if err != nil {
		t.Fatalf("CreateNodeAtSpawn: %v", err)
	}
	if nodeID == "" {
		t.Fatal("empty nodeID")
	}

	nodes := store.BuildNodes()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n := nodes[0]
	if n.Depth != 1 {
		t.Errorf("Depth = %d, want 1 (first node)", n.Depth)
	}
	if n.ParentNodeID != "" {
		t.Errorf("ParentNodeID = %q, want empty (root)", n.ParentNodeID)
	}
	if n.SessionID != "" {
		t.Errorf("SessionID = %q, want empty (skeleton)", n.SessionID)
	}
	if n.WorktreePath != "/tmp/wt-root" {
		t.Errorf("WorktreePath = %q, want /tmp/wt-root", n.WorktreePath)
	}
}

// TestCreateNodeAtSpawnNestedDepth verifies that spawning from a parent at
// depth 1 produces a child at depth 2 with correct origin_chain.
func TestCreateNodeAtSpawnNestedDepth(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, _ := NewStore(storePath)
	pop := NewPopulator(store)

	// Create parent at depth 1.
	t.Setenv(config.EnvChainNodeID, "")
	parentID, _ := pop.CreateNodeAtSpawn("/tmp/wt-parent", "SPEC-X", "M0")

	// Simulate child spawn with parent env set.
	t.Setenv(config.EnvChainNodeID, parentID)
	childID, err := pop.CreateNodeAtSpawn("/tmp/wt-child", "SPEC-X", "M1")
	if err != nil {
		t.Fatalf("CreateNodeAtSpawn child: %v", err)
	}

	nodes := store.BuildNodes()
	var child *WorktreeNode
	for i := range nodes {
		if nodes[i].NodeID == childID {
			child = &nodes[i]
			break
		}
	}
	if child == nil {
		t.Fatalf("child node %s not found", childID)
	}
	if child.Depth != 2 {
		t.Errorf("child Depth = %d, want 2", child.Depth)
	}
	if child.ParentNodeID != parentID {
		t.Errorf("child ParentNodeID = %q, want %q", child.ParentNodeID, parentID)
	}
	if len(child.OriginChain) != 2 {
		t.Errorf("child OriginChain len = %d, want 2", len(child.OriginChain))
	}
}

// TestBackfillSessionID verifies REQ-CHAIN-021: the two-phase backfill binds
// the real session_id to a skeleton node.
func TestBackfillSessionID(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "events.jsonl"))
	pop := NewPopulator(store)

	t.Setenv(config.EnvChainNodeID, "")
	nodeID, _ := pop.CreateNodeAtSpawn("/tmp/wt-bf", "", "")

	// Node has empty session_id.
	nodes := store.BuildNodes()
	if nodes[0].SessionID != "" {
		t.Fatalf("expected empty SessionID before backfill")
	}

	// Backfill.
	if err := pop.BackfillSessionID(nodeID, "real-sess-456"); err != nil {
		t.Fatalf("BackfillSessionID: %v", err)
	}

	nodes = store.BuildNodes()
	if nodes[0].SessionID != "real-sess-456" {
		t.Errorf("SessionID after backfill = %q, want real-sess-456", nodes[0].SessionID)
	}
}

// TestBackfillSessionIDEmptyNodeID verifies that backfill with empty nodeID
// returns an error.
func TestBackfillSessionIDEmptyNodeID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "events.jsonl"))
	pop := NewPopulator(store)

	err := pop.BackfillSessionID("", "sess")
	if err == nil {
		t.Fatal("expected error for empty nodeID")
	}
}

// TestResolveCurrentNodeFromEnv verifies that ResolveCurrentNode returns the
// node from env when MOAI_CHAIN_NODE_ID is set.
func TestResolveCurrentNodeFromEnv(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "events.jsonl"))
	pop := NewPopulator(store)

	t.Setenv(config.EnvChainNodeID, "")
	nodeID, _ := pop.CreateNodeAtSpawn("/tmp/wt-env", "", "")

	// Env has node ID → fast path.
	node, err := pop.ResolveCurrentNode("/tmp/wt-env", "")
	if err != nil {
		t.Fatalf("ResolveCurrentNode: %v", err)
	}
	if node.NodeID != nodeID {
		t.Errorf("resolved NodeID = %q, want %q", node.NodeID, nodeID)
	}
}

// TestResolveCurrentNodeFromLedger verifies that ResolveCurrentNode falls
// back to ledger resolution when env is unset (post-/clear simulation).
func TestResolveCurrentNodeFromLedger(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "events.jsonl"))
	pop := NewPopulator(store)

	t.Setenv(config.EnvChainNodeID, "")
	nodeID, _ := pop.CreateNodeAtSpawn("/tmp/wt-ledger", "", "")
	_ = pop.BackfillSessionID(nodeID, "sess-ledger")

	// Simulate /clear: unset env.
	_ = os.Unsetenv(config.EnvChainNodeID)

	node, err := pop.ResolveCurrentNode("/tmp/wt-ledger", "sess-ledger")
	if err != nil {
		t.Fatalf("ResolveCurrentNode from ledger: %v", err)
	}
	if node.NodeID != nodeID {
		t.Errorf("resolved NodeID = %q, want %q", node.NodeID, nodeID)
	}
}

// TestAppendCompletionEdge verifies the completion-edge append path.
func TestAppendCompletionEdge(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "events.jsonl"))
	pop := NewPopulator(store)

	err := pop.AppendCompletionEdge("N1", "N2", "M2-done", "Start M3")
	if err != nil {
		t.Fatalf("AppendCompletionEdge: %v", err)
	}

	events, _ := store.ReadAll()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev.EventType != EventCompletionEdge {
		t.Errorf("EventType = %q, want completion-edge", ev.EventType)
	}
	if ev.CompletedMilestone != "M2-done" {
		t.Errorf("CompletedMilestone = %q, want M2-done", ev.CompletedMilestone)
	}
}

// TestUpdateMilestone verifies milestone update path.
func TestUpdateMilestone(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "events.jsonl"))
	pop := NewPopulator(store)

	t.Setenv(config.EnvChainNodeID, "")
	nodeID, _ := pop.CreateNodeAtSpawn("/tmp/wt-ms", "SPEC-X", "M0")

	err := pop.UpdateMilestone(nodeID, "M1", "M0-complete")
	if err != nil {
		t.Fatalf("UpdateMilestone: %v", err)
	}

	nodes := store.BuildNodes()
	if nodes[0].Milestone != "M1" {
		t.Errorf("Milestone = %q, want M1", nodes[0].Milestone)
	}
	if nodes[0].LastCompletedMilestone != "M0-complete" {
		t.Errorf("LastCompletedMilestone = %q, want M0-complete", nodes[0].LastCompletedMilestone)
	}
}
