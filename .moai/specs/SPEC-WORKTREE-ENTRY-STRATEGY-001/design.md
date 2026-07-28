# design.md — SPEC-WORKTREE-ENTRY-STRATEGY-001

> Design decisions worth recording for posterity. Tier L requires this
> artifact per spec-workflow.md § SPEC Complexity Tier. Records the WHY
> behind non-obvious choices; consult alongside plan.md §E (Milestones).

## §A. Decision Records

### §A.1 DR-1: `EnterWorktree` for current-session, `moai cc -w` for new-session

**Decision**: the two mechanisms are complementary, NOT interchangeable.

- `EnterWorktree(<path>)` — current-session re-entry. Used when the
  orchestrator is mid-session and needs to switch the CWD to an existing
  worktree without losing prompt-cache context.
- `moai cc -w <worktree-name>` — new-session launcher. Used when the
  paste-ready resume crosses a `/clear` boundary or a terminal restart,
  starting a fresh Claude Code session inside the worktree.

**Rationale**: conflating the two produces either (a) lost prompt-cache
context (using `moai cc -w` mid-session needlessly cold-starts), or (b)
failed cold-start (using `EnterWorktree` after a `/clear` finds no live
session to relocate). The two forms serve distinct continuity regimes.

**Alternatives considered**:
- (a) Unify on `moai cc -w` only — rejected: cold-starts mid-session are
  wasteful and break in-flight reasoning.
- (b) Unify on `EnterWorktree` only — rejected: no live session exists
  after `/clear` or terminal restart.
- (c) Bare `cd` shell form — rejected per REQ-WES-002: bypasses CWD
  isolation for `Agent(isolation: "worktree")` subagents and produces
  worktree-masked failures.

### §A.2 DR-2: web auto-toggles default OFF (user-intent alignment)

**Decision**: all three web worktree auto-toggles (`AutoCreate`,
`AutoCleanup`, `AutoMerge`) default to `false`.

**Rationale**: the user's stated intent (CLAUDE.local.md §22, dev-settings-
intent) is that worktree automation act only on explicit opt-in. The
current defaults (`AutoCleanup: true`, `AutoMerge: true`) silently perform
actions the user did not request — a violation of the explicit-opt-in
principle. Web console automation that mutates the working tree without an
explicit user toggle is the named anti-pattern.

**Alternatives considered**:
- (a) Keep `AutoCleanup: true`, flip only `AutoMerge` — rejected: half-
  measure that leaves the silent-cleanup hazard in place.
- (b) Add a confirmation prompt before each auto action — rejected: the
  web console is a settings surface, not an interactive dialog; defaults
  are the correct intervention point.
- (c) Remove the toggles entirely — rejected: the toggles are useful when
  the user explicitly opts in; removing them eliminates a legitimate
  power-user feature.

### §A.3 DR-3: parallel-session auto-isolation (conservative predicate)

**Decision**: when worktree entry is chosen AND the active-sessions
registry reports ≥1 foreign entry whose recorded branch equals the current
branch, the orchestrator auto-creates an `auto-<session-short>-<spec-id>`
worktree. The predicate is conservative (any foreign entry triggers; no
PID-liveness check).

**Rationale**: false positives are cheap (an unnecessary auto-isolated
worktree is a few MB on disk + a deleted branch); false negatives corrupt
the working tree (the named race that produced the branch-guard sibling
SPEC). Per memory `feedback_session_registry_stale_race_detection.md`, the
registry CAN produce stale entries — but the cost differential favors
conservative firing.

**Alternatives considered**:
- (a) Strict predicate (live PID + cwd match) — rejected as the default
  (would miss real races when the registry is stale); remains a future
  tuning option via OQ-2.
- (b) Surface as AskUserQuestion (manual resolution) — rejected: the
  Pre-Spawn Sync Check already does this; auto-isolation is the next
  layer that reduces user interrupt fatigue.
- (c) Deny worktree entry on race — rejected: too restrictive; the user
  should be able to proceed in an isolated branch.

### §A.4 DR-4: auto-isolated branch naming `auto-<session-short>-<spec-id>`

**Decision**: the auto-isolated worktree's branch follows the pattern
`auto-<session-short>-<spec-id>`, where `<session-short>` is the first 8
chars of the foreign session's UUID and `<spec-id>` is the active SPEC ID
(or `ad-hoc` if no SPEC is active).

**Rationale**:
- `auto-` prefix: visually distinguishes auto-isolated branches from
  user-created `feat/` / `fix/` branches in `git branch -a` output; makes
  them easy to bulk-delete via `git branch | grep ^auto- | xargs git branch -D`.
- `<session-short>` (8 chars): enough to disambiguate two concurrent
  sessions without leaking the full UUID into the branch name (which would
  be unwieldy).
- `<spec-id>`: ties the auto-isolated branch to a SPEC context for
  traceability; falls back to `ad-hoc` when no SPEC is active.

**Alternatives considered**:
- (a) `auto-<timestamp>` — rejected: timestamp does not identify the
  foreign session, harder to debug after the fact.
- (b) `auto-<foreign-pid>` — rejected: PID is not stable across reboots
  and not meaningful to humans.
- (c) Use the existing `SessionNamePattern` field (`moai-{ProjectName}-{SPEC-ID}`)
  — rejected: that field names SESSIONS, not branches; conflating them
  breaks the existing semantics.

### §A.5 DR-5: `Agent(isolation: "worktree")` explicitly NOT a re-entry mechanism

**Decision**: the docs explicitly call out that
`Agent(isolation: "worktree")` creates a NEW L1 ephemeral worktree and is
NOT a mechanism for re-entering an existing L2 persistent SPEC worktree.

**Rationale**: the implicit distinction was the root cause of past confusion
(where a user attempted to "re-enter" an L2 worktree by spawning an
isolation-marked subagent, which silently created a NEW L1 worktree
instead, causing the subagent's writes to land in the wrong tree). An
explicit callout eliminates the ambiguity.

**Alternatives considered**:
- (a) Leave the distinction implicit — rejected: empirically produces
  confusion (prior incident evidence).
- (b) Make `Agent(isolation: "worktree")` re-entry-aware (accept a path
  argument to re-enter existing L2) — rejected: would require a runtime
  change in Claude Code, out of scope for this SPEC.

## §B. Architecture

### §B.1 Component relationships

```
User
 │
 ├─ Pastes paste-ready resume Block 0
 │   ├─ Form A: moai cc -w <name>      → NEW Claude Code session inside worktree
 │   └─ Form B: EnterWorktree(<path>)  → CURRENT session relocates CWD to worktree
 │
 ├─ Mid-session worktree re-entry
 │   └─ EnterWorktree(<path>)          ← canonical per REQ-WES-001
 │
 ├─ Parallel-session race detected
 │   ├─ Pre-Spawn Sync Check (existing)
 │   └─ Auto-isolation (NEW per REQ-WES-005)
 │       └─ Creates auto-<session-short>-<spec-id> worktree
 │
 └─ Web console worktree automation
     ├─ AutoCreate  (default false per REQ-WES-004)
     ├─ AutoCleanup (default false per REQ-WES-004)
     └─ AutoMerge   (default false per REQ-WES-004)
```

### §B.2 Branch-guard sibling compatibility

The sibling SPEC-WORKTREE-BRANCH-GUARD-001 denies branch-state changes in
the **primary checkout** only. The auto-isolation procedure (REQ-WES-005)
creates worktrees at `.claude/worktrees/` or `~/.moai/worktrees/` — both
are worktree paths per the guard's discriminant (`git rev-parse --git-dir`
≠ `git rev-parse --git-common-dir`). The guard's deny is suppressed in
worktree contexts; auto-isolation proceeds unimpeded.

## §C. Trade-offs

| Trade-off | Choice | Cost |
|-----------|--------|------|
| Cold-start vs cache-preservation for mid-session entry | Cache-preservation (`EnterWorktree`) | None (canonical form); only alternative cost is the cold-start that `moai cc -w` would impose |
| Conservative vs strict auto-isolation predicate | Conservative | Extra worktree churn on stale registry entries; offset by the info log + the cheap cost of an unnecessary worktree |
| Doc-rule edits vs code changes | Both (doc-rule edits for policy; 2-line code change for defaults) | Sanitized-pair obligation on doc-rule edits (template mirror verification) |

## §D. Forward Compatibility

- The `WorkflowWorktreeConfig` struct (types.go:476) is unchanged — future
  SPECs that add new toggles (e.g. `AutoSync`) can do so without
  coordinating with this SPEC.
- The auto-isolation naming scheme (`auto-<session-short>-<spec-id>`) is
  deterministic and parseable; future tooling (e.g. `moai worktree
  list-auto`) can identify auto-isolated branches by prefix.
- The EnterWorktree-first policy is runtime-agnostic — if Claude Code
  introduces additional worktree tools (e.g. `SwitchWorktree`), this SPEC's
  policy extends naturally to treat them as canonical for current-session
  re-entry.

## §E. Out-of-Scope Design Decisions (deferred)

- The `MOAI_BRANCH_GUARD_EXEMPT=1` spawn contract (the orchestrator's
  follow-up SPEC for `manager-git` Late-Branch closure) — different SPEC,
  tracked as backlog.
- A runtime-layer hook that mechanically enforces the EnterWorktree-first
  policy (e.g. a PreToolUse guard that denies `cd <worktree-path>` in
  orchestrator context) — out of scope; this SPEC establishes the policy
  via documentation + AC grep evidence.
- A unified `worktree entry` subcommand that wraps both `EnterWorktree`
  and `moai cc -w` behind a single CLI verb — out of scope; the two forms
  serve distinct continuity regimes and SHOULD remain separate per DR-1.
