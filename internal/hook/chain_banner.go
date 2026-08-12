package hook

// SPEC-CHAIN-CORE-001 REQ-CHAIN-013 — SessionStart lineage banner.
//
// When SessionStart fires and MOAI_CHAIN_NODE_ID is absent from the
// environment (post-/clear env loss), this module resolves the node from
// the chain ledger, re-injects the node ID, and emits a lineage
// system-reminder. It also performs the REQ-CHAIN-021 session_id backfill
// when the env node ID is present but the node's session_id is still empty.
//
// The module is time-boxed via context.WithTimeout and degrades to a no-op
// (empty string) on timeout or any error — fail-open.

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
)

// chainBannerTimeout is the maximum time the lineage banner resolution may
// take before degrading to a no-op. REQ-CHAIN-013 requires time-boxed +
// fail-open.
const chainBannerTimeout = 2 * time.Second

// chainLineageBanner resolves the current chain node and returns a lineage
// reminder string suitable for AdditionalContext injection. It performs:
//
//  1. session_id backfill (REQ-CHAIN-021) when MOAI_CHAIN_NODE_ID is set
//     in env but the node has no session_id yet.
//  2. post-/clear re-injection (REQ-CHAIN-013) when MOAI_CHAIN_NODE_ID is
//     absent, resolving from the ledger via (cwd, sessionID).
//  3. lineage banner emission (depth, parent chain, resume_target).
//
// Returns an empty string on any error or timeout (fail-open).
func chainLineageBanner(projectDir, cwd, sessionID string) string {
	if projectDir == "" {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), chainBannerTimeout)
	defer cancel()

	result := make(chan string, 1)
	go func() {
		result <- resolveChainBanner(ctx, projectDir, cwd, sessionID)
	}()

	select {
	case banner := <-result:
		return banner
	case <-ctx.Done():
		slog.Warn("chain lineage banner: timed out (fail-open to no-op)",
			"timeout", chainBannerTimeout)
		return ""
	}
}

// resolveChainBanner does the actual chain resolution work. It is called
// inside a goroutine with a context deadline.
func resolveChainBanner(ctx context.Context, projectDir, cwd, sessionID string) string {
	chainDir := filepath.Join(projectDir, ".moai", "state", "chain")
	storePath := filepath.Join(chainDir, "events.jsonl")

	store, err := chain.NewStore(storePath)
	if err != nil {
		slog.Debug("chain banner: store unavailable", "error", err)
		return ""
	}

	pop := chain.NewPopulator(store)
	envNodeID := os.Getenv(config.EnvChainNodeID)

	var node *chain.WorktreeNode

	if envNodeID != "" {
		// Fast path: env has the node ID.
		// Perform session_id backfill (REQ-CHAIN-021 Phase 2).
		if sessionID != "" {
			if err := pop.BackfillSessionID(envNodeID, sessionID); err != nil {
				slog.Warn("chain banner: backfill failed (non-blocking)",
					"error", err)
			}
		}
		node, err = pop.ResolveCurrentNode(cwd, sessionID)
		if err != nil {
			slog.Debug("chain banner: resolve from env failed", "error", err)
			return ""
		}
	} else {
		// Slow path: env absent (post-/clear). Resolve from ledger.
		node, err = pop.ResolveCurrentNode(cwd, sessionID)
		if err != nil {
			// No chain context — this is normal for a primary checkout
			// with no worktree lineage.
			return ""
		}
	}

	if node == nil {
		return ""
	}

	return formatChainBanner(node, envNodeID == "")
}

// formatChainBanner builds the lineage system-reminder string from the
// resolved node.
func formatChainBanner(node *chain.WorktreeNode, reInjected bool) string {
	var b strings.Builder

	b.WriteString("🔗 chain lineage: ")
	fmt.Fprintf(&b, "depth %d", node.Depth)

	if len(node.OriginChain) > 1 {
		// Show the abbreviated chain: N0→N1→N2
		chainDisplay := make([]string, 0, len(node.OriginChain))
		for _, id := range node.OriginChain {
			if len(id) > 8 {
				chainDisplay = append(chainDisplay, id[:8])
			} else {
				chainDisplay = append(chainDisplay, id)
			}
		}
		fmt.Fprintf(&b, " of %s", strings.Join(chainDisplay, "→"))
	}

	if node.SpecID != "" {
		fmt.Fprintf(&b, " | spec: %s", node.SpecID)
	}
	if node.Milestone != "" {
		fmt.Fprintf(&b, " | ms: %s", node.Milestone)
	}
	if node.LastCompletedMilestone != "" {
		fmt.Fprintf(&b, " | done: %s", node.LastCompletedMilestone)
	}
	b.WriteByte('\n')

	if node.ResumeTarget != "" {
		fmt.Fprintf(&b, "📍 resume: %s\n", node.ResumeTarget)
	}
	if node.ResumeCommand != "" {
		fmt.Fprintf(&b, "▶ next: %s\n", node.ResumeCommand)
	}

	if reInjected && node.NodeID != "" {
		fmt.Fprintf(&b, "♻ node re-injected: MOAI_CHAIN_NODE_ID=%s (env was lost after /clear)\n", node.NodeID)
	}

	return b.String()
}
