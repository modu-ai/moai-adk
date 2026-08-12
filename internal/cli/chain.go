package cli

// SPEC-CHAIN-CORE-001 — Chain CLI subcommands.
//
// moai chain status   — print current node summary (depth, parent, spec, milestone, resume)
// moai chain lineage  — print root-to-leaf origin chain
// moai chain back     — print parent resume target + command
// moai chain list     — enumerate all nodes with staleness status
// moai chain prune    — fold old exited nodes into archive (M5)
//
// All subcommands are flag-agnostic (no kanban/factory/lead dependency).
// The CLI preserves the orchestrator-only interaction boundary (REQ-CHAIN-018).

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// ChainStateDir is the project-relative path for the chain state directory.
const ChainStateDir = ".moai/state/chain"

// ChainEventsFile is the JSONL ledger filename within the chain state dir.
const ChainEventsFile = "events.jsonl"

// chainStaleThreshold is the default staleness window for heartbeat-based
// classification (REQ-CHAIN-015: 15 minutes).
const chainStaleThreshold = 15 * time.Minute

// newChainCmd creates the chain subcommand tree.
func newChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain",
		Short: "Worktree session origin-trail chain",
		Long: `Query and manage the worktree session origin-trail chain.

The chain records worktree spawn lineage so that a maintainer re-entering
a depth-N worktree after /clear can recover origin, completion, and
resume-target without grep or scrollback archaeology.`,
		GroupID: "tools",
	}

	cmd.AddCommand(newChainStatusCmd())
	cmd.AddCommand(newChainLineageCmd())
	cmd.AddCommand(newChainBackCmd())
	cmd.AddCommand(newChainListCmd())
	cmd.AddCommand(newChainPruneCmd())

	return cmd
}

// resolveChainStore finds or creates the chain Store from the project root.
// Prefers CLAUDE_PROJECT_DIR; falls back to findStateDir() directory walk.
func resolveChainStore() (*chain.Store, error) {
	var chainDir string
	if projDir := os.Getenv(config.EnvClaudeProjectDir); projDir != "" {
		chainDir = filepath.Join(projDir, ChainStateDir)
	} else {
		stateDir, err := findStateDir()
		if err != nil {
			return nil, fmt.Errorf("chain: locate project state dir: %w", err)
		}
		chainDir = filepath.Join(filepath.Dir(stateDir), "chain")
	}
	if err := os.MkdirAll(chainDir, 0o755); err != nil {
		return nil, fmt.Errorf("chain: create state dir: %w", err)
	}
	storePath := filepath.Join(chainDir, ChainEventsFile)
	return chain.NewStore(storePath)
}

// resolveCWD returns the current working directory, preferring
// CLAUDE_PROJECT_DIR when set.
func resolveCWD() string {
	if dir := os.Getenv(config.EnvClaudeProjectDir); dir != "" {
		return dir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

// isRemoteCWD returns true if the CWD looks like a remote path (not on the
// local filesystem). This surfaces the single-host v1 limitation
// (REQ-CHAIN-016).
func isRemoteCWD(cwd string) bool {
	if cwd == "" {
		return false
	}
	// Check for common remote indicators: ssh/scp syntax, UNC paths that
	// don't resolve locally.
	if strings.Contains(cwd, "://") || strings.HasPrefix(cwd, "ssh://") {
		return true
	}
	// If the path doesn't exist locally, it's likely remote.
	if _, err := os.Stat(cwd); err != nil && os.IsNotExist(err) {
		// Only flag as remote if it looks like a remote path pattern.
		if strings.Contains(cwd, ":") && !strings.HasPrefix(cwd, "/") {
			return true
		}
	}
	return false
}

// --- status ---

func newChainStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print the current chain node summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChainStatus(cmd.OutOrStdout())
		},
	}
}

func runChainStatus(out interface{ Write([]byte) (int, error) }) error {
	cwd := resolveCWD()

	if isRemoteCWD(cwd) {
		fmt.Fprintf(out, "chain: single-host v1 limitation — CWD %q appears remote\n", cwd)
		fmt.Fprintln(out, "chain: cross-machine lineage is not supported in v1")
		return nil
	}

	store, err := resolveChainStore()
	if err != nil {
		fmt.Fprintln(out, "no chain context (state dir not found)")
		return nil
	}

	pop := chain.NewPopulator(store)
	node, err := pop.ResolveCurrentNode(cwd, os.Getenv("CLAUDE_SESSION_ID"))
	if err != nil {
		fmt.Fprintln(out, "no chain context (no matching node)")
		return nil
	}

	fmt.Fprintf(out, "depth:     %d\n", node.Depth)
	fmt.Fprintf(out, "node:      %s\n", node.NodeID)
	if node.ParentNodeID != "" {
		fmt.Fprintf(out, "parent:    %s\n", node.ParentNodeID)
	}
	if node.SpecID != "" {
		fmt.Fprintf(out, "spec:      %s\n", node.SpecID)
	}
	if node.Milestone != "" {
		fmt.Fprintf(out, "milestone: %s\n", node.Milestone)
	}
	if node.LastCompletedMilestone != "" {
		fmt.Fprintf(out, "completed: %s\n", node.LastCompletedMilestone)
	}
	if node.ResumeTarget != "" {
		fmt.Fprintf(out, "resume:    %s\n", node.ResumeTarget)
	}
	if node.SessionID != "" {
		fmt.Fprintf(out, "session:   %s\n", node.SessionID)
	}
	fmt.Fprintf(out, "worktree:  %s\n", node.WorktreePath)

	return nil
}

// --- lineage ---

func newChainLineageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lineage",
		Short: "Print the root-to-leaf origin chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChainLineage(cmd.OutOrStdout())
		},
	}
}

func runChainLineage(out interface{ Write([]byte) (int, error) }) error {
	cwd := resolveCWD()
	store, err := resolveChainStore()
	if err != nil {
		fmt.Fprintln(out, "no chain context")
		return nil
	}

	pop := chain.NewPopulator(store)
	node, err := pop.ResolveCurrentNode(cwd, os.Getenv("CLAUDE_SESSION_ID"))
	if err != nil {
		fmt.Fprintln(out, "no chain context (no matching node)")
		return nil
	}

	if len(node.OriginChain) == 0 {
		fmt.Fprintln(out, "at root — no ancestors")
		return nil
	}

	nodes := store.BuildNodes()
	nodeMap := make(map[string]*chain.WorktreeNode, len(nodes))
	for i := range nodes {
		nodeMap[nodes[i].NodeID] = &nodes[i]
	}

	fmt.Fprintf(out, "origin chain (%d nodes):\n", len(node.OriginChain))
	for i, id := range node.OriginChain {
		n := nodeMap[id]
		prefix := "  "
		if id == node.NodeID {
			prefix = "> "
		}
		if n != nil {
			fmt.Fprintf(out, "%s[%d] %s\n", prefix, n.Depth, id)
			if n.WorktreePath != "" {
				fmt.Fprintf(out, "     path: %s\n", n.WorktreePath)
			}
			if n.SpecID != "" {
				fmt.Fprintf(out, "     spec: %s\n", n.SpecID)
			}
			if n.Milestone != "" {
				fmt.Fprintf(out, "     ms:   %s\n", n.Milestone)
			}
			if n.EnteredAt != "" {
				fmt.Fprintf(out, "     at:   %s\n", n.EnteredAt)
			}
		} else {
			fmt.Fprintf(out, "%s[?] %s (not in ledger)\n", prefix, id)
		}
		_ = i
	}

	return nil
}

// --- back ---

func newChainBackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "back",
		Short: "Print the parent node's resume target and command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChainBack(cmd.OutOrStdout())
		},
	}
}

func runChainBack(out interface{ Write([]byte) (int, error) }) error {
	cwd := resolveCWD()
	store, err := resolveChainStore()
	if err != nil {
		fmt.Fprintln(out, "no chain context")
		return nil
	}

	pop := chain.NewPopulator(store)
	node, err := pop.ResolveCurrentNode(cwd, os.Getenv("CLAUDE_SESSION_ID"))
	if err != nil {
		fmt.Fprintln(out, "no chain context (no matching node)")
		return nil
	}

	if node.ParentNodeID == "" {
		fmt.Fprintln(out, "at root — no parent")
		return nil
	}

	nodes := store.BuildNodes()
	for _, n := range nodes {
		if n.NodeID == node.ParentNodeID {
			fmt.Fprintf(out, "parent: %s\n", n.NodeID)
			if n.ResumeTarget != "" {
				fmt.Fprintf(out, "resume target: %s\n", n.ResumeTarget)
			}
			if n.ResumeCommand != "" {
				fmt.Fprintf(out, "resume cmd:    %s\n", n.ResumeCommand)
			}
			if n.WorktreePath != "" {
				fmt.Fprintf(out, "worktree:      %s\n", n.WorktreePath)
			}
			return nil
		}
	}

	fmt.Fprintf(out, "parent node %s not found in ledger\n", node.ParentNodeID)
	return nil
}

// --- list ---

// nodeStaleness classifies a chain node's liveness via the overlay join on
// the session registry's LastHeartbeat (REQ-CHAIN-015).
type nodeStaleness string

const (
	stalenessActive nodeStaleness = "active"
	stalenessStale  nodeStaleness = "stale"
	stalenessExited nodeStaleness = "exited"
)

func newChainListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Enumerate all chain nodes with staleness status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChainList(cmd.OutOrStdout())
		},
	}
}

func runChainList(out interface{ Write([]byte) (int, error) }) error {
	store, err := resolveChainStore()
	if err != nil {
		fmt.Fprintln(out, "no chain context")
		return nil
	}

	nodes := store.BuildNodes()
	if len(nodes) == 0 {
		fmt.Fprintln(out, "no chain nodes in ledger")
		return nil
	}

	// Overlay join: read registry entries for staleness classification.
	registryEntries := loadRegistryForOverlay()

	fmt.Fprintf(out, "%-20s %-5s %-12s %-12s %s\n", "NODE", "DEPTH", "SESSION", "STATUS", "WORKTREE")
	for _, n := range nodes {
		status := classifyStaleness(n.SessionID, registryEntries)
		sess := n.SessionID
		if sess == "" {
			sess = "(pending)"
		}
		wt := n.WorktreePath
		if len(wt) > 40 {
			wt = "..." + wt[len(wt)-37:]
		}
		fmt.Fprintf(out, "%-20s %-5d %-12s %-12s %s\n", truncateID(n.NodeID), n.Depth, sess, status, wt)
	}

	return nil
}

// loadRegistryForOverlay reads the session registry for the staleness overlay
// join. Returns nil if the registry is unavailable (fail-open).
func loadRegistryForOverlay() []session.Entry {
	stateDir, err := findStateDir()
	if err != nil {
		return nil
	}
	regPath := filepath.Join(stateDir, "active-sessions.json")
	reg := session.NewRegistry(regPath, realClock{})
	entries, err := reg.Query("")
	if err != nil {
		return nil
	}
	return entries
}

// classifyStaleness determines a node's liveness from the registry overlay.
func classifyStaleness(sessionID string, entries []session.Entry) nodeStaleness {
	if sessionID == "" {
		return stalenessExited
	}
	for _, e := range entries {
		if e.SessionID == sessionID {
			age := time.Since(e.LastHeartbeat)
			if age > chainStaleThreshold {
				return stalenessStale
			}
			return stalenessActive
		}
	}
	return stalenessExited
}

// realClock implements session.Clock using time.Now().
type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// --- prune (basic — full implementation in M5) ---

func newChainPruneCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Fold old exited nodes into archive",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChainPrune(cmd.OutOrStdout(), dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "Dry-run mode (default true; use --no-dry-run to execute)")
	return cmd
}

func runChainPrune(out interface{ Write([]byte) (int, error) }, dryRun bool) error {
	store, err := resolveChainStore()
	if err != nil {
		fmt.Fprintln(out, "no chain context")
		return nil
	}

	threshold := chain.DefaultPruneThreshold()
	result, err := store.Prune(threshold, time.Now().UTC(), dryRun)
	if err != nil {
		fmt.Fprintf(out, "prune error: %v\n", err)
		return nil
	}

	if result.ArchivedNodes == 0 {
		fmt.Fprintln(out, "no nodes eligible for pruning")
		return nil
	}

	if dryRun {
		fmt.Fprintf(out, "[dry-run] %d nodes would be archived (%d kept)\n",
			result.ArchivedNodes, result.KeptNodes)
	} else {
		fmt.Fprintf(out, "archived %d nodes (%d kept)\n",
			result.ArchivedNodes, result.KeptNodes)
		fmt.Fprintf(out, "  original size: %d bytes → compacted: %d bytes\n",
			result.OriginalSize, result.CompactedSize)
		if result.ArchivedPath != "" {
			fmt.Fprintf(out, "  archive: %s\n", result.ArchivedPath)
		}
	}

	for _, id := range result.ArchivedNodeIDs {
		fmt.Fprintf(out, "  %s\n", truncateID(id))
	}

	return nil
}

// truncateID shortens a node ID for display.
func truncateID(id string) string {
	if len(id) <= 18 {
		return id
	}
	return id[:8] + "..." + id[len(id)-6:]
}
