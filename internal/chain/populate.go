package chain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// GenerateNodeID creates a monotonically-sortable node ID without external
// dependencies. Format: <unix-millis-hex>-<random-4-bytes-hex>. The
// millisecond prefix ensures chronological ordering; the random suffix
// ensures uniqueness within the same millisecond.
func GenerateNodeID() string {
	ts := uint64(time.Now().UnixMilli())
	var randBytes [4]byte
	_, _ = rand.Read(randBytes[:])
	return fmt.Sprintf("%013x-%s", ts, hex.EncodeToString(randBytes[:]))
}

// nowFunc is the time provider, overridable for tests.
var nowFunc = func() time.Time { return time.Now().UTC() }

// Populator creates chain nodes at spawn boundaries and handles session_id
// backfill. It is the population path that connects the chain store to the
// CLI/worktree subsystem.
//
// SPEC-CHAIN-CORE-001 REQ-CHAIN-005 (spawn-boundary population),
// REQ-CHAIN-021 (session_id two-phase backfill).
type Populator struct {
	store *Store
}

// NewPopulator creates a Populator bound to the given store.
func NewPopulator(store *Store) *Populator {
	return &Populator{store: store}
}

// CreateNodeAtSpawn creates a skeleton node-enter event at a spawn boundary.
//
// The parent node ID is read from MOAI_CHAIN_NODE_ID in the environment
// (the spawning context's node ID). If the env is unset (primary checkout
// with no chain context), a depth-0 root node is created.
//
// worktreePath is the resolved worktree filesystem path. specID and milestone
// are optional context (may be empty at spawn time).
//
// Returns the new node_id, which the caller MUST set as MOAI_CHAIN_NODE_ID
// on the child process environment.
func (p *Populator) CreateNodeAtSpawn(worktreePath, specID, milestone string) (string, error) {
	parentID := os.Getenv(config.EnvChainNodeID)

	var parentDepth int
	var parentChain []string

	if parentID != "" {
		// Look up parent node to derive depth and origin_chain.
		nodes := p.store.BuildNodes()
		for _, n := range nodes {
			if n.NodeID == parentID {
				parentDepth = n.Depth
				parentChain = make([]string, len(n.OriginChain))
				copy(parentChain, n.OriginChain)
				break
			}
		}
		// If parent not found in ledger, treat as depth-0 root (degraded
		// path — the spawner's env pointed at a stale/unknown node).
	}

	nodeID := GenerateNodeID()
	depth := parentDepth + 1

	// Build origin_chain: parent_chain + [nodeID].
	originChain := make([]string, 0, len(parentChain)+1)
	originChain = append(originChain, parentChain...)
	originChain = append(originChain, nodeID)

	ev := ChainEvent{
		EventType:    EventNodeEnter,
		NodeID:       nodeID,
		ParentNodeID: parentID,
		Depth:        depth,
		OriginChain:  originChain,
		WorktreePath: worktreePath,
		SessionID:    "", // skeleton — backfilled at child SessionStart
		SpecID:       specID,
		Milestone:    milestone,
		EnteredAt:    nowFunc().Format(time.RFC3339),
	}

	if err := p.store.Append(ev); err != nil {
		return "", fmt.Errorf("chain populate: append node-enter: %w", err)
	}

	return nodeID, nil
}

// BackfillSessionID appends a node-update event binding a node to its
// runtime-assigned session_id. Called at child SessionStart when the runtime
// assigns the real SessionID.
//
// SPEC-CHAIN-CORE-001 REQ-CHAIN-021.
func (p *Populator) BackfillSessionID(nodeID, sessionID string) error {
	if nodeID == "" {
		return fmt.Errorf("chain populate: BackfillSessionID requires non-empty nodeID")
	}
	return p.store.Append(ChainEvent{
		EventType: EventNodeUpdate,
		NodeID:    nodeID,
		SessionID: sessionID,
	})
}

// UpdateMilestone appends a node-update event recording milestone progress.
// This is used to track milestone transitions without a completion edge.
func (p *Populator) UpdateMilestone(nodeID, milestone, lastCompleted string) error {
	if nodeID == "" {
		return fmt.Errorf("chain populate: UpdateMilestone requires non-empty nodeID")
	}
	return p.store.Append(ChainEvent{
		EventType:             EventNodeUpdate,
		NodeID:                nodeID,
		Milestone:             milestone,
		LastCompletedMilestone: lastCompleted,
	})
}

// AppendCompletionEdge appends a completion-edge event. This is the
// mechanical write path used by the chain-event hook (REQ-CHAIN-012).
func (p *Populator) AppendCompletionEdge(parentNode, childNode, completedMilestone, nextResumeTarget string) error {
	return p.store.Append(ChainEvent{
		EventType:         EventCompletionEdge,
		ParentNode:        parentNode,
		ChildNode:         childNode,
		CompletedMilestone: completedMilestone,
		CompletedAt:       nowFunc().Format(time.RFC3339),
		NextResumeTarget:  nextResumeTarget,
	})
}

// ResolveCurrentNode resolves the current node from the environment. If
// MOAI_CHAIN_NODE_ID is set, it returns that node directly. If unset
// (post-/clear), it resolves via (worktreePath, sessionID) from the ledger.
//
// SPEC-CHAIN-CORE-001 REQ-CHAIN-013 (re-injection after /clear).
//
// @MX:ANCHOR [AUTO]: ResolveCurrentNode — chain node resolution entry point
// @MX:REASON: 5 cross-package callers (cli chain status/lineage/back + hook chain_event/chain_banner)
func (p *Populator) ResolveCurrentNode(worktreePath, sessionID string) (*WorktreeNode, error) {
	// Fast path: env has the node ID.
	if envID := os.Getenv(config.EnvChainNodeID); envID != "" {
		nodes := p.store.BuildNodes()
		for i := range nodes {
			if nodes[i].NodeID == envID {
				return &nodes[i], nil
			}
		}
	}

	// Slow path: resolve from ledger by (worktreePath, sessionID).
	return p.store.ResolveByCWD(worktreePath, sessionID)
}
