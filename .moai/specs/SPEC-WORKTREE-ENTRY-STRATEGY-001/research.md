# research.md — SPEC-WORKTREE-ENTRY-STRATEGY-001

> **Read-only codebase reconnaissance** captured 2026-07-28 (main checkout,
> origin/main, divergence `0 0`). This file is the research artifact per
> spec-workflow.md Tier L. It records current-state evidence that motivates
> the SPEC requirements; it does NOT propose solutions (see design.md).

## §A. Current Worktree Entry Mechanism Inventory

### §A.1 `EnterWorktree` / `ExitWorktree` (Claude Code runtime tools)

**Location**: documented in
`.claude/rules/moai/workflow/worktree-integration.md` § `EnterWorktree` /
`ExitWorktree` Tools (single paragraph, ~10 lines, lines 148-150 of the
rule).

**Current text** (verbatim, lines 148-150):

> Claude can move the session into a worktree mid-session via the
> `EnterWorktree` tool (e.g. when the user says "work in a worktree"),
> creating one under `.claude/worktrees/`. Once inside, Claude can switch
> directly to another worktree by calling `EnterWorktree` with a target path;
> the previous worktree stays on disk untouched. `ExitWorktree` returns to
> the originating checkout. These are Claude Code runtime tools — MoAI does
> not mandate their use; they are the interactive counterpart to the
> `--worktree` launch flag and `isolation: worktree` frontmatter.

**Gap**: the section treats `EnterWorktree` as one option among several; it
does NOT establish `EnterWorktree` as the canonical current-session re-entry
mechanism, nor does it distinguish the L1-ephemeral vs L2-persistent
re-entry cases. This SPEC expands this section to carry that policy.

### §A.2 `moai cc -w <worktree-name>` (launcher flag)

**Location**: `internal/cli/launcher.go` `normalizeWorktreeFlag` (lines
665-702).

**Behavior**: rewrites `-w` / `--worktree[=name]` token forms into the
canonical two-token `--worktree [name]` form that Claude Code accepts.
Preserves argument order; tokens after `--` are verbatim pass-through.

**Current role**: launcher for starting a NEW Claude Code session inside a
worktree. Documented in:

- `worktree-integration.md` § `claude --worktree (-w)` Flag (lines 56-78) —
  user-facing behavior.
- `session-handoff-examples.md` § Worktree-Anchored Resume Pattern — Form A
  (`moai cc -w <worktree-name>` for `.claude/worktrees/<name>/` worktrees)
  and Form B (`cd <path>` for `~/.moai/worktrees/<project>/<spec>/` L2
  worktrees).

**Gap**: Form B currently emits a bare `cd` instruction. Per REQ-WES-002,
the orchestrator-emitted Block 0 SHALL NOT use bare `cd`; the Form B
guidance needs to migrate to either `EnterWorktree(<path>)` (current-session
re-entry) or `moai cc -w` (new-session launch).

[NEEDS CLARIFICATION: which form should Form B (L2 worktree at ~/.moai/worktrees/) emit by default — EnterWorktree for current-session continuation, or moai cc -w for new-session launch? The plan.md default is EnterWorktree (in-session continuation is the common case mid-SPEC); confirm at Implementation Kickoff Approval.] → [RESOLVED v0.2.0 — Form B default is `EnterWorktree(<path>)` for current-session continuation; `moai cc -w <abs-path-under-~/.moai/worktrees/>` is the new-session-launch counterpart (REQ-WES-010's launcher L2 path-resolution extension, additively landed, legacy token-normalization path preserved). The EnterWorktree RUNTIME TOOL's `.claude/worktrees/`-only constraint is OUT OF SCOPE (deferred to a follow-up runtime-layer SPEC). See plan.md §I OQ-1 / spec.md HISTORY 0.2.0 + REQ-WES-010 + §B.2.]

### §A.3 `Agent(isolation: "worktree")` (L1 ephemeral)

**Location**: documented in `worktree-integration.md` § `isolation: worktree`
in Agent Frontmatter (lines 82-99) and § Worktree Selection Rules.

**Current role**: spawns a subagent in a fresh L1 ephemeral worktree at
`.claude/worktrees/<auto-name>/`. The runtime decides whether to materialize
the worktree; MoAI does not mandate it.

**HARD rule** (worktree-integration.md § Worktree Selection Rules, line 178):
implementation teammates in team mode (write-capable implementation roles)
MUST use `isolation: "worktree"`. Read-only teammates MUST NOT.

**Gap**: the L1-vs-L2 distinction is implicit. This SPEC's REQ-WES-003 makes
it explicit that `Agent(isolation: "worktree")` is NOT a re-entry mechanism
for existing L2 worktrees.

## §B. Current Web Worktree Auto-Toggle Defaults

### §B.1 Config struct

**Location**: `internal/config/types.go:476-483`.

```go
// WorkflowWorktreeConfig mirrors workflow.worktree.* — worktree automation settings.
// Distinct from GitStrategyConfig.WorktreeRoot (different key domain, no conflict).
type WorkflowWorktreeConfig struct {
    AutoCleanup        bool   `yaml:"auto_cleanup"`
    AutoCreate         bool   `yaml:"auto_create"`
    AutoMerge          bool   `yaml:"auto_merge"`
    SessionNamePattern string `yaml:"session_name_pattern"`
    TmuxPreferred      bool   `yaml:"tmux_preferred"`
}
```

### §B.2 Current defaults (defaults.go:520-527)

```go
Worktree: WorkflowWorktreeConfig{
    AutoCleanup:        true,   // ← MUTATE to false (REQ-WES-004)
    AutoCreate:         false,  // ← already correct
    AutoMerge:          true,   // ← MUTATE to false (REQ-WES-004)
    SessionNamePattern: "moai-{ProjectName}-{SPEC-ID}",
    TmuxPreferred:      true,   // out of scope (not in SPEC)
},
```

**Baseline drift confirmed**: 2 of 3 toggles targeted by REQ-WES-004 are
currently `true`. The SPEC's AC-WES-004 must observe the post-mutation
state (`AutoCleanup: false`, `AutoCreate: false`, `AutoMerge: false`),
NOT the current state.

### §B.3 i18n keys

**Location**: `internal/web/assets/i18n.js` — all 4 locales (en/ko/ja/zh)
carry the three toggle title+desc keys:

- `f.workflow.worktree.auto_cleanup.{title,desc}` — en lines 288-289, ko
  680-681, ja 1044-1045, zh 1408-1409
- `f.workflow.worktree.auto_create.{title,desc}` — en 290-291, ko 682-683,
  ja 1046-1047, zh 1410-1411
- `f.workflow.worktree.auto_merge.{title,desc}` — en 292-293, ko 684-685,
  ja 1048-1049, zh 1412-1413

**No i18n changes needed** — the keys already exist; defaults enforcement
happens in `defaults.go`.

### §B.4 `SessionNamePattern` field

Discovered during reconnaissance: `SessionNamePattern: "moai-{ProjectName}-{SPEC-ID}"`
exists on the struct but is NOT one of the three auto-toggles targeted by
this SPEC. It is left untouched. Worth noting because REQ-WES-005's auto-
isolation naming scheme (`auto-<session-short>-<spec-id>`) is conceptually
adjacent but distinct (it names a worktree BRANCH, not a session).

## §C. Worktree Sprawl Baseline (2026-07-28)

**Command**: `git worktree list | wc -l` → **58 entries**

**Command**: `git worktree list | grep -c "agent-"` → **31 entries** (L1
ephemeral `agent-*` worktrees, uncleaned)

**Sample (first 5)**:

```
/Users/goos/MoAI/moai-adk-go                                                a75ae654c [main]
/Users/goos/.moai/worktrees/moai-adk-go/console-impl                        d900d992b [console-m5]
/Users/goos/.moai/worktrees/moai-adk/branch-guard                           6e155954a [docs/main-checkout-branch-guard]
/Users/goos/.moai/worktrees/moai-adk/ci-merge-convergence                   3f0aaa337 [docs/ci-merge-convergence]
/Users/goos/.moai/worktrees/moai-adk/fix-dead-cache-injector                272bafb74 [fix/dead-cache-injector]
```

**Interpretation**: 58 total worktrees, 31 of which are L1 ephemeral
`agent-*` (the runtime's auto-naming for `Agent(isolation: "worktree")`
worktrees). The remaining 27 are L2 persistent SPEC worktrees or named L1
worktrees. The sprawl signature confirms the motivation for REQ-WES-009
(prefer re-entry over creation) and REQ-WES-004 (web console must not
silently amplify via auto-creation).

## §D. Active-Sessions Registry (auto-isolation input)

### §D.1 Location

- `.moai/state/active-sessions.json` — per-project session registry
- Populated by `moai session register` / `moai session list`
- Consumed by the Pre-Spawn Sync Check (`.claude/rules/moai/core/agent-
  common-protocol.md` § Pre-Spawn Sync Check, 3rd command)

### §D.2 Current interpretation matrix (Pre-Spawn Sync Check)

| Output | Meaning | Action |
|--------|---------|--------|
| `[]` | No other session on this SPEC | Proceed normally |
| `[{...}]` (≥1 foreign entry) | Concurrent session race detected | STOP, surface via AskUserQuestion |

**Gap**: the current matrix surfaces the race as an AskUserQuestion (manual
resolution). REQ-WES-005 adds an auto-isolation path: when worktree entry is
chosen, the orchestrator auto-creates `auto-<session-short>-<spec-id>` so the
two sessions do NOT share a branch.

### §D.3 Caveats from the field (memory)

Per the orchestrator memory entry
`feedback_session_registry_stale_race_detection.md`:

> moai session PID·lsof cwd 오탐 가능 (내 PID조차 DEAD); git status 미커밋
> 증분 + index.lock이 진짜 신호. worktree(origin/main) 격리 우회.

The registry can produce **stale entries** (recorded PIDs of sessions that
have since exited). The auto-isolation procedure MUST therefore treat a
registry entry as a *hypothesis* confirmed by `git status` porcelain +
`index.lock` presence, NOT as definitive evidence of an active race. False
positives are preferable to false negatives here (an unnecessary auto-
isolated branch is cheap; a missed race corrupts the working tree).

[NEEDS CLARIFICATION: the auto-isolation procedure's confidence predicate — should it fire on ANY foreign registry entry, or only on entries whose recorded PID is alive AND whose recorded cwd equals this checkout? The conservative default is "any foreign entry" (false-positive-tolerant); a stricter predicate reduces churn but risks missing a real race when the registry is stale.] → [RESOLVED v0.2.0 — conservative predicate confirmed: the procedure fires on ANY foreign registry entry (Recorded-PID liveness + cwd match are NOT required preconditions); false positives are cheap, false negatives corrupt the working tree. The acceptance.md §F Forward-Looking Checks 30-day false-positive audit MAY motivate a follow-up SPEC tightening the predicate. See plan.md §I OQ-2 / spec.md HISTORY 0.2.0 + REQ-WES-005.]

## §E. Branch-Guard Compatibility

### §E.1 Sibling SPEC-WORKTREE-BRANCH-GUARD-001 (#1192)

**Status**: merged `e89d01461` (2026-07-28).

**Mechanism**: PreToolUse conditional deny via `internal/hook/branch_guard.go`.
Denies branch-state-changing git commands (`git switch`, `git checkout -b`,
`git reset --hard`, `git stash`, `git rebase`, `git merge`) when ALL three
conditions hold: (a) primary checkout, (b) branch-state pattern matched,
(c) agent is not exempt (`HookInput.AgentType == "manager-git"` OR
`MOAI_BRANCH_GUARD_EXEMPT=1`).

### §E.2 Discriminant

The guard classifies via comparing `git rev-parse --git-dir` (absolute)
against `git rev-parse --git-common-dir`:

- Equal paths → primary checkout (deny applies)
- Differing paths → worktree (deny suppressed)

### §E.3 Compatibility with REQ-WES-005

REQ-WES-005's auto-isolation creates a NEW worktree at
`.claude/worktrees/auto-<session-short>-<spec-id>/` or
`~/.moai/worktrees/<project>/auto-<session-short>-<spec-id>/`. Both are
worktree paths per the guard's discriminant, so the branch-state changes
(`git worktree add -b auto-...`) occur in a worktree context and are NOT
denied by the primary-checkout guard.

**Verified compatible**: REQ-WES-005 does NOT trigger the branch-guard deny.

## §F. CLAUDE.local.md §22 Current State

The dev-settings-intent section §22 currently does NOT mention the web
worktree auto-toggles. The closest reference is §22.3 (`teammateMode`) and
§22.5 (operating principle: machine-specific keys are intentionally
absent from templates).

**Gap**: §22 needs a new subsection (proposed §22.8 — web worktree auto-
toggles default OFF) recording the user intent that aligns defaults.go with
the "explicit user opt-in" policy.

## §G. Known Cross-References

| Surface | Path | Touch type |
|---------|------|-----------|
| `EnterWorktree`/`ExitWorktree` docs | `.claude/rules/moai/workflow/worktree-integration.md` § lines 148-150 | Expand section |
| Block 0 (Worktree-Anchored) | `.claude/rules/moai/workflow/session-handoff.md` § lines 161-163 + `session-handoff-examples.md` § Worktree-Anchored | Update both forms |
| `-w` flag behavior | `internal/cli/launcher.go:665-702` | Doc alignment only (no code change) |
| `cleanupMoaiWorktrees` | `internal/cli/launcher.go:354-434` | No change (out of scope) |
| Web worktree defaults | `internal/config/defaults.go:520-527` | MUTATE 2 lines (`AutoCleanup`, `AutoMerge`) |
| Web config struct | `internal/config/types.go:476-483` | No change |
| i18n keys (4 locales) | `internal/web/assets/i18n.js` lines 288-293 (en) + ko/ja/zh | No change (keys exist) |
| Dev settings intent | `CLAUDE.local.md §22` | Add §22.8 |
| Active-sessions registry | `.moai/state/active-sessions.json` + `agent-common-protocol.md` § Pre-Spawn Sync Check | Consume (no schema change) |
| Branch-guard (sibling) | `internal/hook/branch_guard.go` (via SPEC-WORKTREE-BRANCH-GUARD-001) | Cross-reference only (no change) |

## §H. Open Questions Carried to plan.md

- **OQ-1**: Block 0 Form B default — `EnterWorktree` vs `moai cc -w`?
- **OQ-2**: Auto-isolation confidence predicate — strict (live PID + cwd
  match) vs conservative (any foreign registry entry)?
- **OQ-3**: Auto-isolated branch naming — confirm `auto-<session-short>-<spec-id>`
  scheme, or accept the existing `SessionNamePattern` field as the source?
- **OQ-4**: `TmuxPreferred: true` default — out of scope per §B.4, but worth
  confirming at Implementation Kickoff Approval whether it should also flip
  to `false` for consistency with the auto-toggles.
