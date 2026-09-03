package chain

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestChainPruneAgeThreshold verifies AC-CHAIN-011: exited nodes older than
// the age threshold are archived.
func TestChainPruneAgeThreshold(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, _ := NewStore(storePath)

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	oldTime := now.AddDate(0, 0, -31) // 31 days ago — exceeds 30-day threshold

	// Old exited node (no session_id backfill → exited).
	_ = store.Append(ChainEvent{
		EventType: EventNodeEnter,
		NodeID:    "old-node",
		Depth:     1,
		EnteredAt: oldTime.Format(time.RFC3339),
	})

	// Active node (has session_id).
	_ = store.Append(ChainEvent{
		EventType: EventNodeEnter,
		NodeID:    "active-node",
		Depth:     1,
		EnteredAt: now.Format(time.RFC3339),
	})
	_ = store.Append(ChainEvent{
		EventType: EventNodeUpdate,
		NodeID:    "active-node",
		SessionID: "sess-active",
	})

	result, err := store.Prune(PruneThreshold{
		MaxAge:      30 * 24 * time.Hour,
		MaxFileSize: 0, // disable size threshold
	}, now, false)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	if result.ArchivedNodes != 1 {
		t.Errorf("ArchivedNodes = %d, want 1", result.ArchivedNodes)
	}
	if result.KeptNodes != 1 {
		t.Errorf("KeptNodes = %d, want 1", result.KeptNodes)
	}

	// Verify the active read path no longer has the old node.
	events, _ := store.ReadAll()
	for _, ev := range events {
		if ev.NodeID == "old-node" {
			t.Error("old-node should have been pruned from active path")
		}
	}

	// Verify the archive summary exists.
	summaryPath := filepath.Join(dir, "archived", "node-summary.json")
	if _, err := os.Stat(summaryPath); err != nil {
		t.Errorf("archive summary not created: %v", err)
	}
}

// TestChainPruneDryRun verifies dry-run mode doesn't modify files.
func TestChainPruneDryRun(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, _ := NewStore(storePath)

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	oldTime := now.AddDate(0, 0, -31)

	_ = store.Append(ChainEvent{
		EventType: EventNodeEnter,
		NodeID:    "old-node",
		Depth:     1,
		EnteredAt: oldTime.Format(time.RFC3339),
	})

	originalEvents, _ := store.ReadAll()

	result, err := store.Prune(PruneThreshold{
		MaxAge: 30 * 24 * time.Hour,
	}, now, true) // dry-run
	if err != nil {
		t.Fatalf("Prune dry-run: %v", err)
	}

	if result.ArchivedNodes != 1 {
		t.Errorf("dry-run ArchivedNodes = %d, want 1", result.ArchivedNodes)
	}

	// Verify events.jsonl was NOT modified.
	eventsAfter, _ := store.ReadAll()
	if len(eventsAfter) != len(originalEvents) {
		t.Errorf("dry-run modified events: before=%d, after=%d", len(originalEvents), len(eventsAfter))
	}

	// Verify no archive directory was created.
	archiveDir := filepath.Join(dir, "archived")
	if _, err := os.Stat(archiveDir); !os.IsNotExist(err) {
		t.Error("dry-run should not create archive directory")
	}
}

// TestChainPruneNoEligibleNodes verifies prune with no eligible nodes is a no-op.
func TestChainPruneNoEligibleNodes(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, _ := NewStore(storePath)

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)

	// Active node only.
	_ = store.Append(ChainEvent{
		EventType: EventNodeEnter,
		NodeID:    "active",
		Depth:     1,
		EnteredAt: now.Format(time.RFC3339),
		SessionID: "sess",
	})

	result, err := store.Prune(DefaultPruneThreshold(), now, false)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if result.ArchivedNodes != 0 {
		t.Errorf("ArchivedNodes = %d, want 0", result.ArchivedNodes)
	}
}

// TestChainPruneSizeThreshold verifies the file-size threshold.
func TestChainPruneSizeThreshold(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, _ := NewStore(storePath)

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)

	// Exited node (no session_id).
	_ = store.Append(ChainEvent{
		EventType: EventNodeEnter,
		NodeID:    "exited",
		Depth:     1,
		EnteredAt: now.Format(time.RFC3339),
	})

	// Check original file size.
	fi, _ := os.Stat(storePath)
	originalSize := fi.Size()

	// Set a very small size threshold (1 byte) to trigger size-based pruning.
	result, err := store.Prune(PruneThreshold{
		MaxAge:      0, // disable age threshold
		MaxFileSize: 1, // trigger on any file
	}, now, false)
	if err != nil {
		t.Fatalf("Prune size threshold: %v", err)
	}

	if result.ArchivedNodes != 1 {
		t.Errorf("ArchivedNodes = %d, want 1 (size threshold hit)", result.ArchivedNodes)
	}
	if result.OriginalSize != originalSize {
		t.Errorf("OriginalSize = %d, want %d", result.OriginalSize, originalSize)
	}
}

// TestChainPrunePreservesAuditTrail verifies the original events are backed up.
func TestChainPrunePreservesAuditTrail(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, _ := NewStore(storePath)

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	oldTime := now.AddDate(0, 0, -31)

	_ = store.Append(ChainEvent{
		EventType: EventNodeEnter,
		NodeID:    "old-node",
		Depth:     1,
		EnteredAt: oldTime.Format(time.RFC3339),
	})

	_, err := store.Prune(DefaultPruneThreshold(), now, false)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	// Verify backup exists in archive.
	archiveDir := filepath.Join(dir, "archived")
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("read archive dir: %v", err)
	}

	var hasBackup, hasSummary bool
	for _, e := range entries {
		if e.Name() == "node-summary.json" {
			hasSummary = true
		}
		if len(e.Name()) > 12 && e.Name()[:7] == "events-" {
			hasBackup = true
		}
	}
	if !hasBackup {
		t.Error("audit trail backup not found in archive")
	}
	if !hasSummary {
		t.Error("node-summary.json not found in archive")
	}
}

// TestChainPruneMultipleOldNodes verifies pruning many old nodes.
func TestChainPruneMultipleOldNodes(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "events.jsonl")
	store, _ := NewStore(storePath)

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	oldTime := now.AddDate(0, 0, -40)

	// 10 old exited nodes.
	for i := 0; i < 10; i++ {
		_ = store.Append(ChainEvent{
			EventType: EventNodeEnter,
			NodeID:    "old-" + string(rune('A'+i)),
			Depth:     1,
			EnteredAt: oldTime.Format(time.RFC3339),
		})
	}

	// 5 active nodes.
	for i := 0; i < 5; i++ {
		_ = store.Append(ChainEvent{
			EventType: EventNodeEnter,
			NodeID:    "act-" + string(rune('A'+i)),
			Depth:     1,
			EnteredAt: now.Format(time.RFC3339),
			SessionID: "sess-" + string(rune('A'+i)),
		})
	}

	result, err := store.Prune(DefaultPruneThreshold(), now, false)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if result.ArchivedNodes != 10 {
		t.Errorf("ArchivedNodes = %d, want 10", result.ArchivedNodes)
	}
	if result.KeptNodes != 5 {
		t.Errorf("KeptNodes = %d, want 5", result.KeptNodes)
	}

	// Active path should have only 5 nodes.
	events, _ := store.ReadAll()
	// Each active node has 1 event → 5 events total.
	if len(events) != 5 {
		t.Errorf("active path events = %d, want 5", len(events))
	}
}

// TestIsExitedNode verifies the exited-node classification.
func TestIsExitedNode(t *testing.T) {
	tests := []struct {
		name   string
		events []ChainEvent
		exited bool
	}{
		{
			name:   "skeleton no session",
			events: []ChainEvent{{EventType: EventNodeEnter, NodeID: "N1"}},
			exited: true,
		},
		{
			name: "has session via update",
			events: []ChainEvent{
				{EventType: EventNodeEnter, NodeID: "N1"},
				{EventType: EventNodeUpdate, NodeID: "N1", SessionID: "s1"},
			},
			exited: false,
		},
		{
			name: "has session via enter",
			events: []ChainEvent{
				{EventType: EventNodeEnter, NodeID: "N1", SessionID: "s1"},
			},
			exited: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isExitedNode(tc.events); got != tc.exited {
				t.Errorf("isExitedNode() = %v, want %v", got, tc.exited)
			}
		})
	}
}

// TestSummarizeNode verifies the node summary extraction.
func TestSummarizeNode(t *testing.T) {
	events := []ChainEvent{
		{
			EventType:    EventNodeEnter,
			NodeID:       "N1",
			ParentNodeID: "N0",
			Depth:        2,
			WorktreePath: "/tmp/wt",
			SpecID:       "SPEC-X",
			EnteredAt:    "2026-08-13T10:00:00Z",
		},
		{EventType: EventNodeUpdate, NodeID: "N1", SessionID: "s1"},
	}

	s := summarizeNode("N1", events)
	if s.NodeID != "N1" {
		t.Errorf("NodeID = %q", s.NodeID)
	}
	if s.Depth != 2 {
		t.Errorf("Depth = %d, want 2", s.Depth)
	}
	if s.WorktreePath != "/tmp/wt" {
		t.Errorf("WorktreePath = %q", s.WorktreePath)
	}
	if s.EventCount != 2 {
		t.Errorf("EventCount = %d, want 2", s.EventCount)
	}
	if s.EnteredAt != "2026-08-13T10:00:00Z" {
		t.Errorf("EnteredAt = %q", s.EnteredAt)
	}
}
