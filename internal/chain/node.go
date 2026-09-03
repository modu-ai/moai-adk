// Package chain implements the Worktree Session Origin-Trail Chain — an
// append-only JSONL lineage tree persisted at .moai/state/chain/events.jsonl.
//
// The chain records worktree spawn boundaries, session_id backfills, and
// completion edges so that a maintainer re-entering a depth-N worktree after
// /clear can immediately recover origin, completion, and resume-target without
// grep or scrollback archaeology.
//
// SPEC-CHAIN-CORE-001.
package chain

// EventType discriminates the three lifecycle events the JSONL stream carries.
type EventType string

const (
	// EventNodeEnter is appended at a spawn boundary (moai cc -w,
	// EnterWorktree, Agent isolation:worktree). Carries node_id,
	// parent_node_id, depth, origin_chain, worktree_path, spec_id, milestone,
	// entered_at. session_id is left empty pending REQ-CHAIN-021 backfill.
	EventNodeEnter EventType = "node-enter"

	// EventNodeUpdate is appended at child SessionStart (session_id backfill)
	// or milestone completion. Carries node_id plus the field(s) being updated
	// (session_id OR last_completed_milestone).
	EventNodeUpdate EventType = "node-update"

	// EventCompletionEdge is appended by the chain-event.sh hook on
	// SubagentStop. Carries parent_node, child_node, completed_milestone,
	// completed_at, next_resume_target.
	EventCompletionEdge EventType = "completion-edge"
)

// WorktreeNode is the derived current-state view of a single node in the
// origin-trail chain. The tree is derived at read time from the flat event
// stream — no mutable tree file exists.
//
// All 13 named fields are populated from node-enter and node-update events.
// Fields not yet backfilled carry their zero value.
//
// @MX:ANCHOR [AUTO]: WorktreeNode — 13-field invariant contract (REQ-CHAIN-001)
// @MX:REASON: central data structure consumed by CLI, hook, and banner; struct shape frozen by SPEC
type WorktreeNode struct {
	// NodeID is the unique identifier of this node (ULID for monotonic
	// ordering).
	NodeID string `json:"node_id"`

	// ParentNodeID is the node_id of the spawning context. Empty for a
	// depth-0 root node.
	ParentNodeID string `json:"parent_node_id"`

	// Depth is the nesting level: 0 for the primary checkout, 1 for the
	// first worktree layer, 2 for a nested worktree, etc.
	Depth int `json:"depth"`

	// OriginChain is the ordered root-to-leaf NodeID list, denormalized on
	// each node for O(1) lineage queries without tree traversal.
	OriginChain []string `json:"origin_chain"`

	// WorktreePath is the absolute filesystem path of the worktree this node
	// represents.
	WorktreePath string `json:"worktree_path"`

	// SessionID is the runtime-assigned Claude Code session ID. Empty until
	// the REQ-CHAIN-021 two-phase backfill completes at child SessionStart.
	SessionID string `json:"session_id"`

	// SpecID is the SPEC identifier this node's work targets, if any.
	SpecID string `json:"spec_id"`

	// Milestone is the current milestone label within the SPEC.
	Milestone string `json:"milestone"`

	// EnteredAt is the RFC 3339 timestamp when the node was created.
	EnteredAt string `json:"entered_at"`

	// ExitedAt is the RFC 3339 timestamp when the node's session ended, or
	// empty if still active. Derived from heartbeat staleness, not an exit
	// event.
	ExitedAt string `json:"exited_at"`

	// LastCompletedMilestone is the most recent milestone marked complete by
	// a completion-edge event.
	LastCompletedMilestone string `json:"last_completed_milestone"`

	// ResumeTarget is a one-line human intent describing what to resume.
	ResumeTarget string `json:"resume_target"`

	// ResumeCommand is the single primary action to resume at this node.
	ResumeCommand string `json:"resume_command"`
}

// ChainEvent is a single raw event line in the JSONL stream. Each event
// becomes exactly one line in events.jsonl.
//
// node-enter events populate the node-identity fields.
// node-update events carry only node_id + the field(s) being updated.
// completion-edge events carry parent_node, child_node, completed_*.
type ChainEvent struct {
	// EventType discriminates the three lifecycle event types.
	EventType EventType `json:"event_type"`

	// NodeID identifies the node this event concerns (node-enter, node-update).
	NodeID string `json:"node_id"`

	// ParentNodeID is set on node-enter events.
	ParentNodeID string `json:"parent_node_id,omitempty"`

	// Depth is set on node-enter events.
	Depth int `json:"depth,omitempty"`

	// OriginChain is set on node-enter events.
	OriginChain []string `json:"origin_chain,omitempty"`

	// WorktreePath is set on node-enter events.
	WorktreePath string `json:"worktree_path,omitempty"`

	// SessionID is backfilled on node-update events (and optionally set on
	// node-enter if already known).
	SessionID string `json:"session_id,omitempty"`

	// SpecID is set on node-enter events.
	SpecID string `json:"spec_id,omitempty"`

	// Milestone is set on node-enter events and updated on node-update events.
	Milestone string `json:"milestone,omitempty"`

	// EnteredAt is set on node-enter events.
	EnteredAt string `json:"entered_at,omitempty"`

	// LastCompletedMilestone is updated on node-update events.
	LastCompletedMilestone string `json:"last_completed_milestone,omitempty"`

	// ResumeTarget is set/updated on node-enter and node-update events.
	ResumeTarget string `json:"resume_target,omitempty"`

	// ResumeCommand is set/updated on node-enter and node-update events.
	ResumeCommand string `json:"resume_command,omitempty"`

	// --- completion-edge specific fields ---

	// ParentNode identifies the parent node in a completion edge.
	ParentNode string `json:"parent_node,omitempty"`

	// ChildNode identifies the child node in a completion edge.
	ChildNode string `json:"child_node,omitempty"`

	// CompletedMilestone names the milestone that was completed.
	CompletedMilestone string `json:"completed_milestone,omitempty"`

	// CompletedAt is the RFC 3339 timestamp of completion.
	CompletedAt string `json:"completed_at,omitempty"`

	// NextResumeTarget is the one-line intent for the next resume.
	NextResumeTarget string `json:"next_resume_target,omitempty"`
}

// CompletionEdge is a convenience view of a completion-edge event for display.
type CompletionEdge struct {
	ParentNode         string `json:"parent_node"`
	ChildNode          string `json:"child_node"`
	CompletedMilestone string `json:"completed_milestone"`
	CompletedAt        string `json:"completed_at"`
	NextResumeTarget   string `json:"next_resume_target"`
}
