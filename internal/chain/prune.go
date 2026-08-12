package chain

// SPEC-CHAIN-CORE-001 REQ-CHAIN-011 — Compaction / prune.
//
// moai chain prune folds exited nodes older than the threshold (default:
// 30 days OR 10 MB, whichever fires first) into a node-summary snapshot at
// .moai/state/chain/archived/node-summary.json, and compacts the active
// events.jsonl to exclude the archived nodes' events. The original events
// are preserved in the audit trail under archived/.
//
// Prune is MANUAL (user-invoked). Automatic compaction is deferred per
// spec.md §I Out of Scope.

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// PruneThreshold defines when a node is eligible for archival.
type PruneThreshold struct {
	// MaxAge: nodes whose EnteredAt is older than this are eligible.
	// Default: 30 days. Zero means no age-based pruning.
	MaxAge time.Duration

	// MaxFileSize: if the events.jsonl file exceeds this size, prune is
	// triggered regardless of node age. Default: 10 MB. Zero means no
	// size-based pruning.
	MaxFileSize int64
}

// DefaultPruneThreshold returns the canonical prune thresholds.
func DefaultPruneThreshold() PruneThreshold {
	return PruneThreshold{
		MaxAge:      30 * 24 * time.Hour, // 30 days
		MaxFileSize: 10 * 1024 * 1024,    // 10 MB
	}
}

// PruneResult summarizes the outcome of a prune operation.
type PruneResult struct {
	ArchivedNodes  int      `json:"archived_nodes"`
	KeptNodes      int      `json:"kept_nodes"`
	OriginalSize   int64    `json:"original_size_bytes"`
	CompactedSize  int64    `json:"compacted_size_bytes"`
	ArchivedPath   string   `json:"archived_path"`
	ArchivedNodeIDs []string `json:"archived_node_ids"`
}

// Prune compacts the chain ledger by archiving old exited nodes.
//
// The operation:
//  1. Reads all events from events.jsonl.
//  2. Identifies nodes eligible for archival (exited + older than threshold,
//     OR file-size threshold exceeded).
//  3. Writes a node-summary snapshot to archived/node-summary.json.
//  4. Backs up the original events.jsonl to archived/events-{timestamp}.jsonl.
//  5. Rewrites events.jsonl with only the kept events.
//
// If dryRun is true, no files are modified — only the result is returned.
func (s *Store) Prune(threshold PruneThreshold, now time.Time, dryRun bool) (*PruneResult, error) {
	events, err := s.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("chain prune: read events: %w", err)
	}

	// Check file size threshold.
	fileInfo, _ := os.Stat(s.path)
	var originalSize int64
	if fileInfo != nil {
		originalSize = fileInfo.Size()
	}

	sizeThresholdHit := threshold.MaxFileSize > 0 && originalSize >= threshold.MaxFileSize

	// Partition events by node.
	nodeEvents := make(map[string][]ChainEvent) // node_id → events
	var nodeOrder []string
	for _, ev := range events {
		id := ev.NodeID
		if id == "" {
			id = ev.ChildNode // completion-edge events reference child_node
		}
		if id == "" {
			continue
		}
		if _, exists := nodeEvents[id]; !exists {
			nodeOrder = append(nodeOrder, id)
		}
		nodeEvents[id] = append(nodeEvents[id], ev)
	}

	// Classify each node: eligible for archive or keep.
	var archiveIDs []string
	var keepIDs []string
	for _, id := range nodeOrder {
		evs := nodeEvents[id]
		eligible := false

		if sizeThresholdHit {
			// File is too big — archive all exited nodes.
			if isExitedNode(evs) {
				eligible = true
			}
		} else if threshold.MaxAge > 0 {
			// Age-based: check EnteredAt.
			enteredAt := extractEnteredAt(evs)
			if !enteredAt.IsZero() {
				age := now.Sub(enteredAt)
				if age > threshold.MaxAge {
					eligible = true
				}
			}
		}

		if eligible {
			archiveIDs = append(archiveIDs, id)
		} else {
			keepIDs = append(keepIDs, id)
		}
	}

	if len(archiveIDs) == 0 {
		return &PruneResult{
			ArchivedNodes: 0,
			KeptNodes:     len(keepIDs),
			OriginalSize:  originalSize,
			CompactedSize: originalSize,
		}, nil
	}

	// Build the set of events to keep (exclude archived nodes' events).
	archiveSet := make(map[string]bool, len(archiveIDs))
	for _, id := range archiveIDs {
		archiveSet[id] = true
	}

	var keptEvents []ChainEvent
	for _, ev := range events {
		id := ev.NodeID
		if id == "" {
			id = ev.ChildNode
		}
		if id == "" || !archiveSet[id] {
			keptEvents = append(keptEvents, ev)
		}
	}

	// Build the archived node summary.
	dir := filepath.Dir(s.path)
	archiveDir := filepath.Join(dir, "archived")

	result := &PruneResult{
		ArchivedNodes:   len(archiveIDs),
		KeptNodes:       len(keepIDs),
		OriginalSize:    originalSize,
		ArchivedNodeIDs: archiveIDs,
	}

	if dryRun {
		return result, nil
	}

	// Create archive directory.
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return nil, fmt.Errorf("chain prune: create archive dir: %w", err)
	}

	// Write node-summary.json (append to existing summaries).
	summaryPath := filepath.Join(archiveDir, "node-summary.json")
	summaries := loadExistingSummaries(summaryPath)
	for _, id := range archiveIDs {
		summaries = append(summaries, summarizeNode(id, nodeEvents[id]))
	}
	if err := writeSummaries(summaryPath, summaries); err != nil {
		return nil, fmt.Errorf("chain prune: write summary: %w", err)
	}
	result.ArchivedPath = summaryPath

	// Back up original events.jsonl to archive.
	backupName := fmt.Sprintf("events-%s.jsonl", now.Format("20060102-150405"))
	backupPath := filepath.Join(archiveDir, backupName)
	if originalSize > 0 {
		if err := copyFile(s.path, backupPath); err != nil {
			slog.Warn("chain prune: failed to back up original events (non-blocking)",
				"error", err)
		}
	}

	// Rewrite events.jsonl with only kept events.
	if err := s.rewriteEvents(keptEvents); err != nil {
		return nil, fmt.Errorf("chain prune: rewrite events: %w", err)
	}

	// Measure compacted size.
	if fi, err := os.Stat(s.path); err == nil {
		result.CompactedSize = fi.Size()
	}

	return result, nil
}

// rewriteEvents replaces the events file with only the given events. This is
// the ONE exception to the append-only invariant (REQ-CHAIN-002) — it is a
// maintenance compaction invoked only by the explicit user-facing prune
// command (REQ-CHAIN-011).
func (s *Store) rewriteEvents(events []ChainEvent) error {
	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("create events file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return fmt.Errorf("encode event: %w", err)
		}
	}
	return nil
}

// isExitedNode returns true if the node has no active session (session_id is
// empty or the node was never backfilled).
func isExitedNode(events []ChainEvent) bool {
	for _, ev := range events {
		if ev.EventType == EventNodeUpdate && ev.SessionID != "" {
			return false // has a session_id → not exited
		}
		if ev.EventType == EventNodeEnter && ev.SessionID != "" {
			return false
		}
	}
	return true
}

// extractEnteredAt parses the EnteredAt timestamp from a node's events.
func extractEnteredAt(events []ChainEvent) time.Time {
	for _, ev := range events {
		if ev.EventType == EventNodeEnter && ev.EnteredAt != "" {
			t, err := time.Parse(time.RFC3339, ev.EnteredAt)
			if err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// ArchivedNodeSummary is the compact representation of an archived node.
type ArchivedNodeSummary struct {
	NodeID      string   `json:"node_id"`
	ParentNodeID string  `json:"parent_node_id,omitempty"`
	Depth       int      `json:"depth"`
	WorktreePath string  `json:"worktree_path,omitempty"`
	SpecID      string   `json:"spec_id,omitempty"`
	EnteredAt   string   `json:"entered_at,omitempty"`
	EventCount  int      `json:"event_count"`
	ArchivedAt  string   `json:"archived_at"`
}

// summarizeNode creates an ArchivedNodeSummary from a node's events.
func summarizeNode(nodeID string, events []ChainEvent) ArchivedNodeSummary {
	s := ArchivedNodeSummary{
		NodeID:     nodeID,
		EventCount: len(events),
		ArchivedAt: time.Now().UTC().Format(time.RFC3339),
	}
	for _, ev := range events {
		if ev.EventType == EventNodeEnter {
			s.ParentNodeID = ev.ParentNodeID
			s.Depth = ev.Depth
			s.WorktreePath = ev.WorktreePath
			s.SpecID = ev.SpecID
			s.EnteredAt = ev.EnteredAt
		}
	}
	return s
}

// loadExistingSummaries reads existing archived summaries (if any).
func loadExistingSummaries(path string) []ArchivedNodeSummary {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var summaries []ArchivedNodeSummary
	json.Unmarshal(data, &summaries)
	return summaries
}

// writeSummaries writes the summary list to the archive file.
func writeSummaries(path string, summaries []ArchivedNodeSummary) error {
	data, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
