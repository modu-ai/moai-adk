package hook

// SPEC-CHAIN-CORE-001 REQ-CHAIN-012 — Mechanical completion hook.
//
// This handler fires on SubagentStop (or a phase-transition signal), reads
// the hook payload from stdin, resolves the chain node from the session_id,
// and appends a completion-edge event to the ledger. The ledger stays
// current even if the spawning session has crashed or been cleared.
//
// The hook is fail-open: any error (missing store, unresolvable node,
// malformed payload) results in a silent no-op (exit 0). The hook NEVER
// blocks or invokes user interaction.

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
)

// chainEventPayload is the subset of the Claude Code SubagentStop / Stop
// hook payload this handler consumes.
type chainEventPayload struct {
	SessionID    string `json:"session_id"`
	CWD          string `json:"cwd"`
	ProjectDir   string `json:"project_dir"`
	SubagentID   string `json:"subagent_id"`
	StopReason   string `json:"stop_reason"`
}

// RunChainEvent reads a hook payload from stdin and appends a completion
// edge to the chain ledger. It is the Go handler invoked by the
// `moai hook chain-event` subcommand.
//
// The function is fail-open: it returns nil on any error, logging a
// warning. It NEVER blocks or returns a non-zero exit.
func RunChainEvent(stdin io.Reader) error {
	data, err := io.ReadAll(stdin)
	if err != nil {
		slog.Debug("chain-event: failed to read stdin (non-blocking)", "error", err)
		return nil
	}

	var payload chainEventPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		slog.Debug("chain-event: malformed payload (non-blocking)", "error", err)
		return nil
	}

	// Resolve project directory: payload.ProjectDir → CLAUDE_PROJECT_DIR → cwd.
	projectDir := payload.ProjectDir
	if projectDir == "" {
		projectDir = os.Getenv(config.EnvClaudeProjectDir)
	}
	if projectDir == "" {
		projectDir = payload.CWD
	}
	if projectDir == "" {
		slog.Debug("chain-event: no project dir resolved (non-blocking)")
		return nil
	}

	// Open the chain store.
	chainDir := filepath.Join(projectDir, ".moai", "state", "chain")
	storePath := filepath.Join(chainDir, "events.jsonl")
	store, err := chain.NewStore(storePath)
	if err != nil {
		slog.Debug("chain-event: store unavailable (non-blocking)", "error", err)
		return nil
	}

	// Resolve the child node from the session_id or env node ID.
	pop := chain.NewPopulator(store)
	cwd := payload.CWD
	if cwd == "" {
		cwd = os.Getenv(config.EnvClaudeProjectDir)
	}

	node, err := pop.ResolveCurrentNode(cwd, payload.SessionID)
	if err != nil {
		// No matching node — this session has no chain context. No-op.
		slog.Debug("chain-event: no matching chain node (non-blocking)",
			"session_id", payload.SessionID)
		return nil
	}

	// Append the completion-edge event. The milestone and resume target
	// are extracted from the node's current state (best-effort).
	completedMilestone := node.LastCompletedMilestone
	if completedMilestone == "" {
		completedMilestone = node.Milestone
	}

	err = pop.AppendCompletionEdge(
		node.ParentNodeID,
		node.NodeID,
		completedMilestone,
		node.ResumeTarget,
	)
	if err != nil {
		slog.Warn("chain-event: failed to append completion edge (non-blocking)",
			"error", err)
		return nil
	}

	slog.Info("chain-event: completion edge appended",
		"parent", node.ParentNodeID,
		"child", node.NodeID,
		"milestone", completedMilestone,
	)

	return nil
}
