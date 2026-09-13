# Card t689 Verdict — EnterWorktree PostToolUse hook: real-session firing observed; `re-stamped: false` is design, not defect

Date: 2026-09-13 · Lane: lane-3 · Branch: `WT-moai-project-dir-restamp` (base `bf27af2cb` = origin/develop)
Card: t689 (GH #1640 residual, Tier S, Class B — investigation; t236 verdict Gap 1 closure)

## Claim

1. The Enter/ExitWorktree PostToolUse hook (`internal/hook/post_tool_worktree.go`, `handleWorktreeMove`) **fires in a real Claude Code session** — observed on every tree move, in this very session, across an isolated worktree pair plus the moves between them.
2. `MOAI_PROJECT_DIR re-stamped in the env file: false` is **an honest report of a provisioned-file absence, not a hook defect**: the hook stamps only when Claude Code has provisioned `CLAUDE_ENV_FILE`, and this session has none. The stamp mechanism itself demonstrably works in sessions where the file exists.
3. Catalog tools called WITHOUT `project_root` read the **primary checkout** (server spawn-frozen env), with the designed warning — observed directly.
4. NEW observation: the session's Bash environment carries a **stale** `MOAI_PROJECT_DIR` pointing at `.claude/worktrees/t681` — a tree this session left hours ago — and it did not track any of the subsequent five tree moves. **Zero Go-code impact** (re-verified: no consumer), so a curiosity rather than a defect; the operative guidance remains "pass project_root explicitly".

## Evidence

**Hook firing (real session, two isolated worktrees + moves between)** — the PostToolUse system message "Session tree moved to <path>. … (MOAI_PROJECT_DIR re-stamped in the env file: false.)" was emitted verbatim on: `EnterWorktree(t689)`, `ExitWorktree`→primary, `EnterWorktree(t689-b)`, `ExitWorktree`→primary, `EnterWorktree(t689)` — plus earlier the same day on every lane-3 card-tree move (≥10 firings total). The message text matches `post_tool_worktree.go:47-54` byte-for-byte.

**`stamped=false` root cause** — `post_tool_worktree.go:41-45`:

```go
stamped := false
if envFile := os.Getenv(config.EnvClaudeEnvFile); envFile != "" {
    stampProjectDirEnv(envFile, newCwd)
    stamped = true
}
```

Observed in the session's Bash environment (same process env the hooks inherit from the runtime):

```
CLAUDE_ENV_FILE=[UNSET]
CLAUDE_PROJECT_DIR=[UNSET]
MOAI_PROJECT_DIR=[/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t681]   ← stale (see claim 4)
```

and no per-session env directory exists for this session: `ls ~/.claude/session-env/ | grep f2050582` → 0 matches (204 other session dirs exist).

**The mechanism works when the file is provisioned** — 13 sibling sessions carry the stamp the producer writes (`cwd_changed.go` `stampProjectDirEnv`, "sole producer of the MOAI_PROJECT_DIR export"), e.g.:

```
$ cat ~/.claude/session-env/1190d23b-afb5-49a7-9dd1-4799696da747/cwdchanged-hook-0.sh
export MOAI_PROJECT_DIR="/Users/goos/.moai/worktrees/home-isolation-guard"
```

**Catalog tool without project_root** — `mcp__moai__spec_progress` called with `project_root` absent from inside `.claude/worktrees/t689`:

```
{"_root":{"dir":"/Users/goos/MoAI/moai-adk-go","source":"env:CLAUDE_PROJECT_DIR",
 "warning":"project_root not passed — resolved from env:CLAUDE_PROJECT_DIR, which froze at server spawn; …"},
 "count":672, …}
```

→ it read the PRIMARY checkout's SPEC catalogue (672 specs), not the worktree's — exactly the documented spawn-frozen behavior, with the designed warning naming the corrective.

**Stale MOAI_PROJECT_DIR (claim 4)** — observed identical value `.claude/worktrees/t681` from inside t689 and again from inside t689-b, after the session had moved through t682, t673, t567, develop, t689, t689-b. Writer not identified: no on-disk env file in `~/.claude/session-env/` carries the t681 path, no settings file carries it, no shell rc sets it (`grep` over all four, exit 2); `ps eww $$` shows it in the Bash process env. Consumer check re-run today: **no Go code reads `MOAI_PROJECT_DIR`** (`grep` over internal/ pkg/ cmd/ → only the producer and this hook's message string), so the stale value reaches shell scripts and humans only.

## Baseline-attribution

Every observation above was made in this session, today (2026-09-13), on the live worktree set of this machine, with the hook source read at `bf27af2cb`. The firing evidence is this session's own PostToolUse system messages; the env readings are direct `echo`/`ps`/`ls`/`grep` outputs from this session's Bash processes.

## Judgment

**`re-stamped: false` is design (fail-open honesty), not a defect.** The hook cannot stamp without a Claude-Code-provisioned env file, reports that condition truthfully, and still performs its other half (session-registry relocation) unconditionally. The lead's reference observation ("false 반복 표시") is explained: this session was provisioned without `CLAUDE_ENV_FILE`, as many are.

**GH #1640's real-session verification gap is closed**: the hook is live, correct in both provisioned and unprovisioned sessions, and the documented project_root guidance remains the operative mitigation.

## Improvement note (non-defect, optional follow-up)

The message's `false` cannot distinguish "no env file provisioned" from "provisioned but the write failed". One clause would: `(CLAUDE_ENV_FILE not provisioned)` vs `(write failed)`. Filed here as a suggestion only — no behavior change, so not repaired on this card.

The stale `MOAI_PROJECT_DIR` observation is recorded for the same reason: today it has no consumer, but the day Go code starts reading it (it is exported in the session env with a tree-relative value), it becomes the "은퇴한 위치를 보는" canary hazard. If a consumer ever lands, the stamp must track moves or be removed.

## Gaps (explicitly NOT observed)

- The writer of the stale Bash-env `MOAI_PROJECT_DIR=t681` was not observed in the act (no on-disk artifact found; hypotheses: a since-deleted session env file, or Claude Code in-memory env application at tree entry).
- Whether Claude Code provisions `CLAUDE_ENV_FILE` for sessions launched via the `moai cc`/`moai glm` launchers was not tested (this session was not launcher-fresh); the 13 stamped sibling sessions suggest it happens under some launch conditions.
- `stampProjectDirEnv`'s write path itself was not exercised in a provisioned session (none available); its unit tests in `internal/hook/post_tool_worktree_test.go` cover it.

## Residual-risk

- The `false` message will keep appearing in unprovisioned sessions and may keep drawing maintainer attention as a suspected defect; the improvement note above is the cheap durable fix.
- The stale env value is harmless today precisely because no Go consumer exists — this is a verified-today fact, not a permanent property.
