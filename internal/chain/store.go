package chain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Store is the append-only JSONL ledger for the origin-trail chain.
// The file is opened with O_APPEND on every write — the kernel serializes
// concurrent appends. No read-modify-write cycle exists; the store never
// loads the full file to mutate it.
//
// SPEC-CHAIN-CORE-001 REQ-CHAIN-002 (append-only), REQ-CHAIN-003
// (corrupt-line tolerance), REQ-CHAIN-004 (CWD-collision resolution).
type Store struct {
	path string
}

// NewStore creates a Store bound to the given JSONL file path. The directory
// is created if it does not exist. The file itself is NOT created here — it
// is created lazily on the first Append (O_APPEND|O_CREATE).
//
// @MX:ANCHOR [AUTO]: NewStore — chain ledger public API boundary
// @MX:REASON: 3 cross-package callers (cli/chain.go, hook/chain_event.go, hook/chain_banner.go)
func NewStore(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("chain store: create dir %s: %w", dir, err)
	}
	return &Store{path: path}, nil
}

// Append writes exactly one JSONL line to the ledger. The file is opened with
// O_APPEND|O_CREATE|O_WRONLY and permission 0o600 (the ledger may carry
// session context). The store NEVER overwrites or truncates existing lines.
//
// SPEC-CHAIN-CORE-001 REQ-CHAIN-002.
func (s *Store) Append(event ChainEvent) error {
	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("chain store: marshal event: %w", err)
	}
	line = append(line, '\n')

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("chain store: open %s: %w", s.path, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("chain store: write %s: %w", s.path, err)
	}
	return nil
}

// ReadAll reads every valid JSONL line from the ledger, skipping corrupt
// lines with a diagnostic warning. A missing file returns an empty slice
// (not an error), so callers can query an absent ledger without special
// handling.
//
// SPEC-CHAIN-CORE-001 REQ-CHAIN-003 (corrupt-line tolerance).
func (s *Store) ReadAll() ([]ChainEvent, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("chain store: open %s for read: %w", s.path, err)
	}
	defer func() { _ = f.Close() }()

	var events []ChainEvent
	scanner := bufio.NewScanner(f)
	// Allow lines up to 1 MB (a node line with full paths + origin_chain
	// can exceed the default 64 KB scanner buffer at depth > 10).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev ChainEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			slog.Warn("chain: skipping corrupt line",
				"file", s.path, "line", lineNum, "error", err)
			continue
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		return events, fmt.Errorf("chain store: scan %s: %w", s.path, err)
	}
	return events, nil
}

// BuildNodes derives the current-state WorktreeNode slice from the flat event
// stream. node-enter events create skeleton nodes; node-update events
// backfill session_id and other mutable fields. The result is ordered
// root-to-leaf by depth.
func (s *Store) BuildNodes() []WorktreeNode {
	events, err := s.ReadAll()
	if err != nil {
		return nil
	}

	// Index by node_id; node-enter establishes the base, node-update
	// overlays updates.
	index := make(map[string]*WorktreeNode)
	var order []string // preserve first-seen order

	for _, ev := range events {
		switch ev.EventType {
		case EventNodeEnter:
			n := &WorktreeNode{
				NodeID:                 ev.NodeID,
				ParentNodeID:           ev.ParentNodeID,
				Depth:                  ev.Depth,
				OriginChain:            ev.OriginChain,
				WorktreePath:           ev.WorktreePath,
				SessionID:              ev.SessionID,
				SpecID:                 ev.SpecID,
				Milestone:              ev.Milestone,
				EnteredAt:              ev.EnteredAt,
				LastCompletedMilestone: ev.LastCompletedMilestone,
				ResumeTarget:           ev.ResumeTarget,
				ResumeCommand:          ev.ResumeCommand,
			}
			index[ev.NodeID] = n
			order = append(order, ev.NodeID)
		case EventNodeUpdate:
			if n, ok := index[ev.NodeID]; ok {
				if ev.SessionID != "" {
					n.SessionID = ev.SessionID
				}
				if ev.LastCompletedMilestone != "" {
					n.LastCompletedMilestone = ev.LastCompletedMilestone
				}
				if ev.Milestone != "" {
					n.Milestone = ev.Milestone
				}
				if ev.ResumeTarget != "" {
					n.ResumeTarget = ev.ResumeTarget
				}
				if ev.ResumeCommand != "" {
					n.ResumeCommand = ev.ResumeCommand
				}
			}
		}
	}

	result := make([]WorktreeNode, 0, len(order))
	for _, id := range order {
		if n, ok := index[id]; ok {
			result = append(result, *n)
		}
	}
	return result
}

// ResolveByCWD resolves the current node for a given (worktreePath, sessionID)
// pair. This is the CWD-collision resolution path (REQ-CHAIN-004):
//
//  1. Primary key: (worktree_path, session_id) pair. If multiple nodes match,
//     the most recently entered (highest depth, last in first-seen order) wins.
//  2. If session_id is empty, fall back to the most recently entered node for
//     that worktree_path.
//
// Returns ErrNodeNotFound if no node matches.
func (s *Store) ResolveByCWD(worktreePath, sessionID string) (*WorktreeNode, error) {
	nodes := s.BuildNodes()

	var bestMatch *WorktreeNode
	var fallback *WorktreeNode
	for i := range nodes {
		n := &nodes[i]
		if n.WorktreePath != worktreePath {
			continue
		}
		// Track most recent entry as fallback (BuildNodes is first-seen
		// ordered, so later nodes with the same path are more recent).
		fallback = n
		// Primary match: (worktree_path, session_id) pair. Prefer the most
		// recent match (higher depth / later in order).
		if sessionID != "" && n.SessionID == sessionID {
			bestMatch = n
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	if fallback != nil {
		if sessionID == "" {
			return fallback, nil
		}
		// session_id provided but no exact match — fall back to most recent.
		slog.Warn("chain: CWD-collision session_id mismatch, using most recent",
			"worktree_path", worktreePath,
			"requested_session_id", sessionID,
			"resolved_node_id", fallback.NodeID)
		return fallback, nil
	}

	return nil, ErrNodeNotFound
}

// ErrNodeNotFound is returned when ResolveByCWD finds no matching node.
var ErrNodeNotFound = fmt.Errorf("chain: node not found")

// Path returns the JSONL file path the store is bound to.
func (s *Store) Path() string {
	return s.path
}
