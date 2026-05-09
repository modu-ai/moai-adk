// Package worktree provides working tree state guard primitives for the MoAI
// orchestrator. It captures Snapshots of working tree state, computes Divergence
// between pre/post states, logs divergences to .moai/reports/worktree-guard/,
// and writes SuspectFlags when an Agent(isolation: "worktree") response shows
// an empty worktreePath.
//
// This package is consumed by internal/cli/worktree (for `moai worktree
// snapshot|verify|restore` subcommands) and is invoked by the orchestrator
// (Claude Code runtime) via Bash before/after each Agent(isolation:) call.
//
// SPEC: SPEC-V3R3-CI-AUTONOMY-001 Wave 5 (T6 Worktree State Guard).
//
// Key constraints (per .claude/rules/moai/core/agent-common-protocol.md):
//   - This package does NOT call AskUserQuestion. Decisions surface as exit
//     codes and JSON reports for the orchestrator to act upon.
//   - Snapshots are persisted to .moai/state/ as JSON; they capture paths only,
//     never file contents (untracked files cannot be restored from git).
package worktree
