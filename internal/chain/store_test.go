package chain

import (
	"os"
	"path/filepath"
	"testing"
)

// newTestStore creates a Store bound to a t.TempDir path for test isolation
// (REQ-CHAIN-020). No test writes to the project's real .moai/state/chain/.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore(%s): %v", path, err)
	}
	return s
}

// TestAppendDoesNotOverwrite verifies REQ-CHAIN-002 / AC-CHAIN-002:
// appending a 4th event to a file with 3 existing events preserves all 3.
func TestAppendDoesNotOverwrite(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)

	for i := 0; i < 3; i++ {
		ev := ChainEvent{
			EventType: EventNodeEnter,
			NodeID:    "node-" + string(rune('A'+i)),
			Depth:     i,
		}
		if err := s.Append(ev); err != nil {
			t.Fatalf("Append event %d: %v", i, err)
		}
	}

	// Append the 4th.
	if err := s.Append(ChainEvent{
		EventType: EventNodeEnter,
		NodeID:    "node-D",
		Depth:     3,
	}); err != nil {
		t.Fatalf("Append 4th: %v", err)
	}

	events, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	// Verify original 3 preserved.
	for i, want := range []string{"node-A", "node-B", "node-C"} {
		if events[i].NodeID != want {
			t.Errorf("event[%d].NodeID = %q, want %q", i, events[i].NodeID, want)
		}
	}
	if events[3].NodeID != "node-D" {
		t.Errorf("event[3].NodeID = %q, want node-D", events[3].NodeID)
	}
}

// TestCorruptLineTolerance verifies REQ-CHAIN-003 / AC-CHAIN-003:
// a malformed JSONL line is skipped with a warning, remaining lines processed.
func TestCorruptLineTolerance(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)

	// Write 3 valid events.
	for _, id := range []string{"v1", "v2", "v3"} {
		if err := s.Append(ChainEvent{EventType: EventNodeEnter, NodeID: id}); err != nil {
			t.Fatalf("Append %s: %v", id, err)
		}
	}

	// Inject a corrupt line at position 3 (0-indexed line 3 = 4th line).
	corrupt := []byte("{this is not valid json}\n")
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("open for corrupt inject: %v", err)
	}
	if _, err := f.Write(corrupt); err != nil {
		_ = f.Close()
		t.Fatalf("write corrupt: %v", err)
	}
	_ = f.Close()

	// Write 2 more valid events after the corrupt line.
	for _, id := range []string{"v4", "v5"} {
		if err := s.Append(ChainEvent{EventType: EventNodeEnter, NodeID: id}); err != nil {
			t.Fatalf("Append %s: %v", id, err)
		}
	}

	events, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	// 5 valid events, 1 corrupt skipped.
	if len(events) != 5 {
		t.Fatalf("expected 5 valid events (corrupt skipped), got %d", len(events))
	}
}

// TestCWDCollisionResolution verifies REQ-CHAIN-004 / AC-CHAIN-004:
// two nodes sharing worktree_path are resolved by the (worktree_path,
// session_id) pair.
func TestCWDCollisionResolution(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)

	wtPath := "/tmp/wt-collision"
	// Two nodes, same worktree_path, different session_id.
	_ = s.Append(ChainEvent{
		EventType:    EventNodeEnter,
		NodeID:       "node-A",
		WorktreePath: wtPath,
		SessionID:    "sess-A",
		Depth:        1,
	})
	_ = s.Append(ChainEvent{
		EventType:    EventNodeEnter,
		NodeID:       "node-B",
		WorktreePath: wtPath,
		SessionID:    "sess-B",
		Depth:        1,
	})

	// Resolve by (worktree_path, session_id_A) → node-A.
	node, err := s.ResolveByCWD(wtPath, "sess-A")
	if err != nil {
		t.Fatalf("ResolveByCWD: %v", err)
	}
	if node.NodeID != "node-A" {
		t.Errorf("resolved NodeID = %q, want node-A", node.NodeID)
	}

	// Resolve by (worktree_path, session_id_B) → node-B.
	node, err = s.ResolveByCWD(wtPath, "sess-B")
	if err != nil {
		t.Fatalf("ResolveByCWD B: %v", err)
	}
	if node.NodeID != "node-B" {
		t.Errorf("resolved NodeID = %q, want node-B", node.NodeID)
	}
}

// TestStoreIsolation verifies REQ-CHAIN-020 / AC-CHAIN-021:
// all chain store tests use t.TempDir, never the real .moai/state/chain/.
func TestStoreIsolation(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)

	// The store path must NOT be under .moai/state/chain/.
	realPath := filepath.Join(".moai", "state", "chain", "events.jsonl")
	if s.path == realPath {
		t.Fatalf("store path is the real project path %s — test isolation violated", realPath)
	}

	// The path must be under the temp dir.
	tmpDir := t.TempDir()
	if !filepath.IsAbs(s.path) && s.path != filepath.Join(tmpDir, "events.jsonl") {
		// newTestStore uses t.TempDir internally; verify it's not the cwd.
		cwd, _ := os.Getwd()
		if filepath.Dir(s.path) == filepath.Join(cwd, ".moai", "state", "chain") {
			t.Fatalf("store path under real project .moai/state/chain — isolation violated")
		}
	}
}

// TestEmptyLedgerReturnsEmpty verifies the edge case from §D.2:
// an absent ledger file returns an empty slice, not an error.
func TestEmptyLedgerReturnsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")
	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	events, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll on absent file: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events from absent file, got %d", len(events))
	}
}

// TestSpawnBoundaryNodeCreation verifies REQ-CHAIN-005 / AC-CHAIN-005:
// a spawn-boundary creates a child node with correct depth and origin_chain.
func TestSpawnBoundaryNodeCreation(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)

	// Parent at depth 1.
	parentChain := []string{"N0", "N1"}
	_ = s.Append(ChainEvent{
		EventType:    EventNodeEnter,
		NodeID:       "N1",
		ParentNodeID: "N0",
		Depth:        1,
		OriginChain:  parentChain,
	})

	// Spawn child N2 from N1.
	childChain := append([]string{}, parentChain...)
	childChain = append(childChain, "N2")
	_ = s.Append(ChainEvent{
		EventType:    EventNodeEnter,
		NodeID:       "N2",
		ParentNodeID: "N1",
		Depth:        2,
		OriginChain:  childChain,
	})

	nodes := s.BuildNodes()
	if len(nodes) < 2 {
		t.Fatalf("expected >= 2 nodes, got %d", len(nodes))
	}
	var n2 *WorktreeNode
	for i := range nodes {
		if nodes[i].NodeID == "N2" {
			n2 = &nodes[i]
			break
		}
	}
	if n2 == nil {
		t.Fatalf("N2 not found in built nodes")
	}
	if n2.ParentNodeID != "N1" {
		t.Errorf("N2.ParentNodeID = %q, want N1", n2.ParentNodeID)
	}
	if n2.Depth != 2 {
		t.Errorf("N2.Depth = %d, want 2", n2.Depth)
	}
	if len(n2.OriginChain) != 3 || n2.OriginChain[2] != "N2" {
		t.Errorf("N2.OriginChain = %v, want [N0 N1 N2]", n2.OriginChain)
	}
}

// TestSessionIDBackfill verifies REQ-CHAIN-021 / AC-CHAIN-023:
// a node-update event backfills session_id on a skeleton node.
func TestSessionIDBackfill(t *testing.T) {
	t.Parallel()
	s := newTestStore(t)

	// Phase 1: skeleton node-enter with session_id="".
	_ = s.Append(ChainEvent{
		EventType:    EventNodeEnter,
		NodeID:       "N1",
		SessionID:    "",
		WorktreePath: "/tmp/wt-1",
		Depth:        1,
	})

	// Phase 2: node-update backfills session_id.
	_ = s.Append(ChainEvent{
		EventType: EventNodeUpdate,
		NodeID:    "N1",
		SessionID: "real-session-123",
	})

	nodes := s.BuildNodes()
	for _, n := range nodes {
		if n.NodeID == "N1" {
			if n.SessionID != "real-session-123" {
				t.Errorf("N1.SessionID after backfill = %q, want real-session-123", n.SessionID)
			}
			return
		}
	}
	t.Fatalf("N1 not found in built nodes")
}
