# `resolveProjectDir()` call-site verdicts

> SPEC-MCP-WORKTREE-ROOT-001 M3 — REQ-4, REQ-5 (AC-4, AC-5).
> Measured in the worktree `.claude/worktrees/t171` at `2c0efade0`
> (post-M2). Line numbers are that commit's.

## Why this table exists

M1 and M2 gave five MCP tools a way for the caller to name its own tree. They
did **not** touch `resolveProjectDir()` itself, deliberately: eighteen other
call sites ride on that function, and moving it would move all of them at once
in a direction nobody had established was right for them.

This is that survey. Its value is the **evidence** column and, for every
deferred row, the "what would settle it" statement. A row reading "unknown"
with no such statement is a gap wearing a verdict's clothes, and fails AC-4.

## The two scopes

- **Project scope** — the artifact belongs to the repository as a whole, and a
  worktree session should read the same one the primary checkout reads.
  `CLAUDE_PROJECT_DIR` is then the *right* answer, not a bug.
- **Tree scope** — the artifact belongs to the working tree the session is
  actually in. `CLAUDE_PROJECT_DIR` is then wrong in a worktree, which is this
  SPEC's defect.

The function cannot serve both, which is why the repair went in above it.

## Verdict table

| # | Call site | Artifact it resolves | Correct scope | Evidence | Disposition |
|---|---|---|---|---|---|
| 1 | `mcp_server.go:105` | tool-enablement map read from `.moai/config/sections/mcp.yaml` at server registration | **project** | `readMCPToolEnablement` reads a config section that is committed and identical across worktrees; the file is deployed by the template, not written per card | correct as-is — no repair |
| 2 | `mcp_server.go:471` (`goal_status`) | `.moai/state/goal/<session-id>.json` | **undecided** | see § The goal rows | deferred |
| 3 | `mcp_server.go:488` (`goal_arm`) | `.moai/state/goal/<session-id>.json` | **undecided** | see § The goal rows | deferred |
| 4 | `mcp_server.go:548` (`verify_snapshot`) | `.moai/state/verify/` snapshot keyed by HEAD digest | **tree** (suspected) | the key is the HEAD SHA, and a worktree has its own HEAD; a snapshot recorded in a worktree and read against the primary's HEAD is a key mismatch by construction | deferred — settled by measuring whether a snapshot written from a worktree is readable from the primary under the same key |
| 5 | `mcp_server.go:578` (`verify_trend`) | the same snapshot store, read-only | **tree** (suspected) | same key argument as row 4; the two must move together or the write and the read land in different stores | deferred — same measurement as row 4 |
| 6 | `mcp_convergence.go:590` | `.moai/state/audit-multi/<session>.json`, the convergence result cache | **undecided** | the state is per-session, and a session is per-tree in Factory/Kanban mode — but the file is a cache of a verdict, and a verdict about which tree is now selectable by `project_root` (M2). The two may need to agree | deferred — settled by deciding whether the cached result is keyed by session alone or by (session, tree); the M5 post-repair run is the first case where a worktree-rooted audit writes this file |
| 7 | `mcp_project_root.go:62` | the fallback inside `resolveToolProjectRoot` | **n/a** | this IS the repaired seam: it returns the old answer only when the caller named nothing (REQ-2) | repaired in M1 |
| 8 | `mcp_codex.go` `handleCodexAudit` | codex review `cwd` | **tree** | lane-9's symptom; a review of the primary checkout issued from a worktree reviews the wrong diff | repaired in M2 (the `resolveProjectDir()` mention at `:1169` is the doc comment for that repair, not a live call) |
| 9 | `goal.go:159` (`goalProjectRoot`) | `.moai/state/goal/` for the `moai goal` CLI | **undecided** | see § The goal rows | deferred |
| 10 | `goal.go:188` | project dir used to count same-project entries in the session registry, for the multi-session arm guard | **undecided** | see § The goal rows | deferred |
| 11 | `launcher_blockcap_infinite.go:139` (`launchProjectRoot`) | the same goal state, read by the launcher's block-cap inject | **undecided** | its own comment states it exists to read the SAME tree `goal.go` writes; whatever rows 2/3/9/10 resolve to, this must match, or the arm and the read diverge | deferred — bound to the goal rows, not independently decidable |
| 12 | `session.go:224` | `.moai/state/current-session-id.txt` side channel written by the SessionStart hook | **undecided** | the hook writes it per project dir and overwrites unconditionally, so under concurrent sessions it already holds only the most recent id — a known weakness independent of worktrees | deferred — settled by determining where the SessionStart hook actually writes the file when the session is in a worktree (read the hook's own resolution, do not infer it from this call site) |
| 13 | `session.go:357` (`moai session doctor`) | `.moai/state/active-sessions.json` registry | **project** | the registry's purpose is cross-session detection *within one repository*; a per-tree registry would make foreign sessions invisible, defeating it | correct as-is — no repair |
| 14 | `todo.go:68` (`resolveTodoQueueRoot`) | the backlog queue root | **project** | the value is immediately passed to `gitcore.ResolveGitDirs` and reduced to the **common** git dir, so every linked worktree resolves to the same queue — the code already normalizes tree → project | correct as-is; the normalization is what makes the "todo is a primary-checkout queue" rule hold |
| 15 | `verify.go:72` | the verify store for the `moai verify` CLI | **tree** (suspected) | same store as rows 4/5; the CLI already offers `--project-root`, so a caller who needs the worktree can say so today | deferred — same measurement as row 4; note the CLI escape hatch already exists, so the CLI is less exposed than the MCP path |
| 16 | `plan.go:67` | `.moai/specs/<SPEC-ID>/` read, `.moai/reports/plan-html/` write | **tree** | a card's SPEC exists only on its branch; this is the same invisibility this SPEC repairs on the MCP side, on the CLI side | deferred — the repair shape is known (a `--project-root` flag, as `verify` has); what is missing is a decision on whether the CLI needs it, since a CLI runs *in* the tree and its cwd is not stale |
| 17 | `memory.go:165` | the auto-memory store, `~/.claude/projects/<slug>/memory` | **project** | the root is reduced to a slug that keys a store outside the repository; lessons are project knowledge, and per-worktree lesson stores would fragment them | correct as-is — no repair |
| 18 | `memory.go:287` (`memory archive`) | the same store | **project** | same reduction, same reasoning as row 17 | correct as-is — no repair |
| 19 | `migrate_profiles.go:378` | migration of that same memory store | **project** | migrates the artifact of rows 17/18; scope follows the artifact | correct as-is — no repair |

**Totals**: 19 rows — 2 repaired (rows 7 and 8; the three SPEC tools repaired in
M1 reach `resolveProjectDir()` only through row 7, so they add no rows of their
own), 6 correct as-is (1, 13, 14, 17, 18, 19), 11 deferred (2, 3, 4, 5, 6, 9, 10,
11, 12, 15, 16).

## The goal rows (REQ-5)

Rows 2, 3, 9, 10, and 11 all resolve `.moai/state/goal/<session-id>.json`, and
they are recorded together because two previously observed defects **may share
this seam**:

- **goal keying is unreliable under worktrees** — an armed goal not being found
  where it was expected when the session was inside a worktree.
- **the multi-session arm parser** — arming landing under a foreign session's id
  when more than one session runs in the same project directory.

**This connection is SUSPECTED, not established.** Nothing in this card measured
it. What makes it worth recording is the cheap-start argument: if the three
symptoms do share one cause, the follow-up is one repair; if they are treated as
three cards, each fixes a third of it and none of them closes.

**What would settle it**: arm a goal from inside a worktree session, then read
`.moai/state/goal/` in BOTH the worktree and the primary checkout, and record
which tree the file landed in and which tree the `stop-goal` evaluator reads at
turn-end. If the write and the read disagree, the seam is confirmed as the
common cause. If they agree, the two prior defects have a different origin and
this note should be struck rather than carried forward.

The row-10 guard is a second, independent reason the goal rows are not simply
"tree scope": it deliberately compares the resolved project dir against session
registry entries, and the registry is project-scoped (row 13). Changing the goal
rows to tree scope without deciding what row 10 compares against would break the
guard silently.

## The stale-cwd branch — recorded, not touched

`resolveProjectDir()` falls back to `os.Getwd()` when `CLAUDE_PROJECT_DIR` is
empty. For an MCP server this fallback is **also** wrong, and wrong in the same
direction: the server is a long-lived subprocess that cannot follow a worktree
switch, so its working directory is stale by construction.

Removing the env branch without also fixing the fallback would therefore not fix
anything — it would land on a different wrong tree. Both branches would have to
move together, which is precisely why this card put the parameter above the
function rather than inside it.

## What this table does not claim

- No row here was verified by running the artifact's own tooling except where the
  evidence column names a mechanical fact (a reduction to a git common dir, a
  slug derivation, a config file's deployment). Every "suspected" and every
  "undecided" is a reading of the code, not a measurement.
- The disposition column records what this card did. It is not a schedule: a
  deferred row is a row a follow-up can pick up with the survey already done, not
  a commitment that anyone will.
