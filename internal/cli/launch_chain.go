package cli

// launch_chain.go implements the chain population half of the launch path
// (SPEC-CHAIN-CORE-001 REQ-CHAIN-005, design.md §3 Path A): when the launcher
// starts a new session bound to a worktree, a node-enter event is appended to
// the chain ledger and the new node ID is placed on the child environment.
//
// Before this file, CreateNodeAtSpawn had no production caller and nothing
// ever set MOAI_CHAIN_NODE_ID, so the chain ledger could never come into
// existence (card t242).

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// injectChainNodeForTarget records the worktree spawn boundary on the chain
// ledger and returns env with MOAI_CHAIN_NODE_ID set to the new node ID, so
// the child session (and its hooks) inherit the node across the process
// boundary — the invariant design.md §3 calls the thread that connects a
// depth-N child to its parent.
//
// Fail-open: the chain is auxiliary telemetry. Any failure warns on warn and
// the launch proceeds without a node — this never blocks a session start.
// target is the stage-C read of the launch view (extractWorktreeLaunchTarget
// over the consumption loop's passThrough, launch_view.go). Launches without a
// resolvable worktree target are not spawn boundaries for
// the launcher: a bare --worktree lets claude auto-generate the name (which
// the launcher cannot know), and resume (-c/--continue) launches reopen an
// existing session rather than spawning one.
func injectChainNodeForTarget(target string, env []string, warn io.Writer) []string {
	worktreePath, ok := resolveLaunchWorktreePath(target)
	if !ok {
		return env
	}
	store, err := resolveChainStore()
	if err != nil {
		warnChainPopulationFailure(warn, err)
		return env
	}
	nodeID, err := chain.NewPopulator(store).CreateNodeAtSpawn(worktreePath, "", "")
	if err != nil {
		warnChainPopulationFailure(warn, err)
		return env
	}
	return replaceEnvValue(env, config.EnvChainNodeID, nodeID)
}

// resolveLaunchWorktreePath resolves a --worktree value to the filesystem path
// the chain ledger should record. Absolute values (L1 in-repo, L2, or the
// home root) are used as-is. A short name resolves across the three roots
// (SPEC-WORKTREE-HOME-ROOT-001 M3, REQ-WHR-007): a single tree is recorded
// at its actual location, and an ambiguous name records nothing rather than
// guessing (REQ-WHR-022); a name with no tree keeps the legacy .claude
// spelling — branch Y's pass-through default. An empty target resolves to
// false.
func resolveLaunchWorktreePath(target string) (string, bool) {
	if target == "" {
		return "", false
	}
	if filepath.IsAbs(target) {
		return filepath.Clean(target), true
	}
	root, err := findProjectRootFn()
	if err != nil {
		return "", false
	}
	if cands, err := worktreeroot.Candidates(root, target); err == nil {
		if len(cands) == 1 {
			return cands[0].Path, true
		}
		if len(cands) > 1 {
			return "", false
		}
	}
	return filepath.Join(root, ".claude", "worktrees", target), true
}

// warnChainPopulationFailure reports a fail-open chain population failure.
func warnChainPopulationFailure(warn io.Writer, err error) {
	if warn == nil {
		return
	}
	_, _ = fmt.Fprintf(warn, "Warning: chain node not recorded for this launch: %v\n", err)
}

// replaceEnvValue returns env with key=value set, replacing any existing entry
// for key, and appending it when absent.
func replaceEnvValue(env []string, key, value string) []string {
	entry := key + "=" + value
	result := make([]string, 0, len(env)+1)
	replaced := false
	for _, e := range env {
		if strings.HasPrefix(e, key+"=") {
			result = append(result, entry)
			replaced = true
		} else {
			result = append(result, e)
		}
	}
	if !replaced {
		result = append(result, entry)
	}
	return result
}
