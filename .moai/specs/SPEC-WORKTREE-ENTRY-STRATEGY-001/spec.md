---
id: SPEC-WORKTREE-ENTRY-STRATEGY-001
title: "EnterWorktree-First Worktree Entry Strategy + Launcher L2 Path-Resolution Extension + Web Auto-Toggles Default OFF + Parallel-Session Branch Conflict Auto-Isolation"
version: "0.3.1"
status: draft
created: 2026-07-28
updated: 2026-07-28
author: manager-spec
priority: High
phase: "v3.0.0"
module: "cross-cutting (.claude/rules/moai/workflow, internal/cli, internal/config, internal/web, CLAUDE.local.md)"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "worktree, enterworktree, session-handoff, launcher, web-console, parallel-session, auto-isolation, sanitized-pair, cross-cutting"
related_specs:
  - SPEC-WORKTREE-BRANCH-GUARD-001
  - SPEC-WORKTREE-001
  - SPEC-WORKTREE-002
  - SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001
  - SPEC-WORKTREE-DOCS-001
---

# SPEC-WORKTREE-ENTRY-STRATEGY-001 — EnterWorktree-First Worktree Entry Strategy

## §A. Problem Statement

MoAI-ADK operates across **three distinct worktree-layer mechanisms** whose
boundaries are currently blurred in user-facing docs and orchestrator-emitted
guidance:

1. **`EnterWorktree(path)` / `ExitWorktree`** — Claude Code runtime tools that
   move the *current* session into a worktree (`.claude/worktrees/<name>/` for
   L1, or any path for ad-hoc entry). The canonical mechanism for **current-
   session re-entry** into an existing worktree.
2. **`moai cc -w <worktree-name>` / `moai glm -w` / `moai cg -w`** — launcher
   flag that starts a **new** Claude Code session inside a worktree. The
   canonical Block 0 launcher for paste-ready resume across a `/clear` /
   terminal-restart boundary.
3. **`Agent(isolation: "worktree")`** — spawns a subagent in a **fresh L1
   ephemeral** worktree (`.claude/worktrees/<auto-name>/`). NOT a mechanism for
   re-entering an existing L2 persistent SPEC worktree.

When the orchestrator emits guidance for "enter the worktree", it today
frequently defaults to the **shell `cd` form** (`cd <path> && <launcher>` or
`git -C <path> <cmd>`), which:

- Breaks `Agent(isolation: "worktree")` CWD isolation (the agent's CWD is the
  worktree root; a `cd /absolute/path` bypasses it — see
  `worktree-integration.md` § Prompt Path Rules for Worktree-Isolated Agents).
- Was the root cause of a prior session incident (SPEC-PRETOOL-GATE-MOVE-001
  F1 followup) where a sub-agent used `git -C` / subshell-`cd` instead of
  `EnterWorktree` and was corrected by the user — the canonical form was not
  faithfully applied.
- Produces worktree sprawl: `git worktree list` on 2026-07-28 shows **58
  worktrees total, 31 of which are uncleaned `agent-*` L1 ephemeral** — a
  baseline this SPEC's auto-cleanup-default-OFF + EnterWorktree-first strategy
  directly addresses.

Separately, the `internal/web` console exposes three worktree-automation
toggles (`workflow.worktree.auto_create`, `auto_cleanup`, `auto_merge`) whose
**defaults today are partially ON** (`AutoCleanup: true`, `AutoMerge: true` per
`internal/config/defaults.go:521,523`) — meaning the web console silently
auto-creates / auto-cleans / auto-merges worktrees unless the user explicitly
opts out. The user's stated intent (CLAUDE.local.md §22) is that worktree
automation act only on explicit user choice.

Finally, the existing Pre-Spawn Sync Check (`agent-common-protocol.md` §
Pre-Spawn Sync Check) detects divergence between local and `origin/main`, and
optionally enumerates active sessions on the same SPEC — but it does NOT
auto-isolate when worktree entry is chosen and another session is already on
the same branch. The user must currently resolve the race manually.

## §B. Scope — 5 Components

### §B.1 EnterWorktree-First Current-Session Entry

The orchestrator's canonical mechanism for entering an existing worktree in
the **current session** SHALL be `EnterWorktree(path)`. The shell `cd` form,
the `git -C <path>` form, and the subshell-`cd` form (`(cd <path> && ...)`)
SHALL NOT be emitted by the orchestrator for current-session worktree entry.

### §B.2 `moai cc -w` New-Session Launcher Role Preserved + L2 Path-Resolution Extension

`moai cc -w <worktree-name>` (and `moai glm -w` / `moai cg -w`) SHALL remain
the canonical Block 0 launcher for starting a **NEW** Claude Code session
inside a worktree — the paste-ready resume Block 0 new-terminal launcher
role. This is complementary to, not replaced by, `EnterWorktree`:
`EnterWorktree` is for in-session re-entry; `moai cc -w` is for cross-session
(cold-start) launch.

Per Decision Point 1 Resolution OQ-1, the launcher's `-w` / `--worktree`
handling is **EXTENDED** with a new MoAI-side pre-resolution step (additive,
not replacement) that recognizes `~/.moai/worktrees/<project>/...` absolute
paths before the existing token-normalization + claude pass-through. Today
`normalizeWorktreeFlag` (launcher.go:665-702) performs ONLY CLI argument-token
normalization (rewriting `-w` / `--worktree[=name]` forms into the canonical
two-token `--worktree [name]` form for claude pass-through); it does NOT do
path resolution, and claude itself resolves the value against
`.claude/worktrees/<name>`. The extension adds a MoAI-side path-resolution
step (sibling to `normalizeWorktreeFlag`, run before it OR inlined within it)
so `moai cc -w <abs-path-under-~/.moai/worktrees/>` reaches the L2 worktree
rather than being silently interpreted as a `.claude/worktrees/` short name.
This makes `moai cc -w <abs-path-under-~/.moai/worktrees/>` a valid re-entry
path for L2 persistent worktrees (resolving the path-split root cause at the
launcher layer, which is under MoAI control — see REQ-WES-010). The existing
`cleanupMoaiWorktrees` function (launcher.go:357-434) ALREADY recognizes
`~/.moai/worktrees/` paths, confirming the feasibility of path-aware
resolution at the launcher layer. The EnterWorktree RUNTIME TOOL's
`.claude/worktrees/`-only constraint is OUT OF SCOPE (see §F — deferred to a
follow-up runtime-layer SPEC).

### §B.3 `Agent(isolation: "worktree")` Distinction

`Agent(isolation: "worktree")` creates a NEW L1 ephemeral worktree scoped to
a single subagent invocation. It SHALL NOT be used to re-enter an existing L2
persistent SPEC worktree. The two layers (L1 ephemeral vs L2 persistent) are
distinct; conflating them produces the worktree-masked flaky failure mode
documented in `worktree-state-guard.md`.

### §B.4 Web Worktree Auto-Toggles Default OFF

`internal/config/defaults.go` `WorkflowWorktreeConfig` SHALL default all
three worktree-automation toggles to `false`:

- `AutoCleanup: false` (currently `true` at defaults.go:521 — MUTATED)
- `AutoCreate: false` (already correct at defaults.go:522 — preserved)
- `AutoMerge: false` (currently `true` at defaults.go:523 — MUTATED)

Worktree automation SHALL act only on explicit user opt-in via the web
console toggle or the `workflow.worktree.*` YAML keys.

### §B.5 Parallel-Session Branch Conflict Auto-Isolation

When worktree entry is chosen AND the orchestrator detects (via the Pre-Spawn
Sync Check active-sessions registry, `.moai/state/active-sessions.json`)
another active session on the same branch, the orchestrator SHALL
auto-create a new worktree branch (naming scheme `auto-<session-short>-<spec-id>`
or equivalent) to prevent cross-session branch-state interference.

### §B.6 Doc Surfaces Aligned

The following doc surfaces SHALL reflect the EnterWorktree-first policy + the
auto-toggle defaults-OFF policy + the auto-isolation procedure:

- `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume
  Pattern (Block 0) — emit `EnterWorktree(<path>)` for current-session re-entry
  AND `moai cc -w <name>` for new-session launch; document both forms.
- `.claude/rules/moai/workflow/worktree-integration.md` § EnterWorktree /
  ExitWorktree Tools (currently a single paragraph) — expand to carry the
  EnterWorktree-first policy, the L1-vs-L2 distinction, and the auto-isolation
  procedure.
- `CLAUDE.local.md` §22 (dev settings intent) — record that the web worktree
  auto-toggles default OFF is the user-intended state.
- `session-handoff-examples.md` § Worktree-Anchored Resume Pattern — Block 0
  example updated to show both the `EnterWorktree` form and the `moai cc -w`
  form.
- `moai init` / `moai update` doc surfaces — explicitly note that neither
  command auto-enters a worktree on the user's behalf.

## §C. Requirements (GEARS notation)

### REQ-WES-001 (Ubiquitous — EnterWorktree-first current-session entry)

The orchestrator SHALL use `EnterWorktree(<path>)` as the canonical mechanism
for entering an existing worktree in the current session, replacing the
shell-`cd`, `git -C <path>`, and subshell-`cd` patterns for current-session
worktree re-entry.

### REQ-WES-002 (Event-detected — anti-pattern: bare-`cd` Block 0)

**When** the orchestrator emits a paste-ready resume message whose Block 0
instructs the user to `cd <worktree-path>` (the legacy shell-cd form), the
orchestrator SHALL instead emit Block 0 in one of the two canonical forms:
(a) `moai cc -w <worktree-name>` for a NEW Claude Code session inside the
worktree (new-terminal / post-`/clear` boundary); OR (b) `EnterWorktree(<path>)`
for current-session re-entry (no `/clear`, same session continuing). A bare
`cd` instruction SHALL NOT appear in orchestrator-emitted Block 0 guidance.

### REQ-WES-003 (Ubiquitous — L1 vs L2 distinction)

The orchestrator SHALL NOT use `Agent(isolation: "worktree")` to re-enter an
existing L2 persistent SPEC worktree. `Agent(isolation: "worktree")` creates a
NEW L1 ephemeral worktree scoped to a single subagent invocation; it is
categorically distinct from re-entering an existing L2 worktree created via
`moai worktree new <SPEC-ID>` or L3 `--worktree`.

### REQ-WES-004 (Capability-gate — web auto-toggles default OFF)

**Where** the `internal/web` configuration exposes the
`workflow.worktree.auto_create`, `workflow.worktree.auto_cleanup`, and
`workflow.worktree.auto_merge` toggles, all three SHALL default to `false`.
Worktree automation in the web console SHALL act only on explicit user opt-in
(web toggle set to ON, or the corresponding YAML key set to `true` in
`.moai/config/sections/workflow.yaml`).

### REQ-WES-005 (Event-detected — parallel-session branch conflict auto-isolation)

**When** the orchestrator detects that worktree entry is chosen AND ≥1
foreign active session is currently on the same branch (signaled by
`.moai/state/active-sessions.json` reporting one or more foreign session
entries whose recorded branch equals the current branch — per Decision
Point 1 Resolution OQ-2, the **conservative** predicate: ANY foreign
registry entry triggers auto-isolation; false positives are cheap, false
negatives corrupt the working tree), the orchestrator SHALL auto-create
one worktree per foreign registry entry (so N foreign sessions produce N
auto-isolated worktrees — see Edge-3), each named
`auto-<session-short>-<spec-id>` where `<session-short>` is the first 8
characters of THAT entry's session UUID (per Decision Point 1 Resolution
OQ-3 — firm; no "or equivalent" clause). Each auto-created worktree SHALL
land under `.claude/worktrees/` or `~/.moai/worktrees/` so the two
sessions do NOT share a branch-state mutable surface.

### REQ-WES-006 (State-driven — `moai cc -w` role preserved)

**While** `moai cc -w <worktree-name>` remains the canonical launcher for
starting a NEW Claude Code session inside an existing worktree (the Block 0
new-terminal / post-`/clear` launcher role), the orchestrator SHALL NOT use
`moai cc -w` for current-session re-entry into an existing worktree — that
role belongs to `EnterWorktree`.

### REQ-WES-007 (Ubiquitous — moai init/update no auto-entry)

The `moai init` and `moai update` commands SHALL NOT automatically enter a
worktree on the user's behalf. Any worktree entry following `moai init` /
`moai update` MUST be the result of an explicit user action (the `--worktree`
flag, or a subsequent `moai worktree new <SPEC-ID>` invocation).

### REQ-WES-008 (Event-detected — doc surfaces reflect EnterWorktree-first policy)

**When** any of the doc surfaces listed in §B.6 is updated, the update SHALL
reflect the EnterWorktree-first policy: current-session entry uses
`EnterWorktree(<path>)`; new-session launch uses `moai cc -w <name>`; the bare
`cd` shell form is documented as deprecated for orchestrator-emitted guidance
and confined to manual / human-typed contexts only.

### REQ-WES-009 (Ubiquitous — worktree sprawl mitigation via re-entry preference)

The orchestrator SHALL prefer `EnterWorktree` re-entry over new-worktree
creation whenever an existing worktree already covers the target branch,
rather than baking a one-shot sprawl observation into a permanent
state-driven predicate. The historical sprawl baseline (58 worktrees
total / 31 uncleaned `agent-*` L1 ephemeral, observed 2026-07-28 — recorded
in §A Problem Statement as motivation, not as a binding threshold) is
re-baselined at sync-close via the §F Forward-Looking Checks 30-day audit;
the binding intent is "prefer re-entry, do not silently amplify sprawl",
NOT "while count ≥ 50 / agent-* ≥ 30". The web auto-toggles default OFF
(REQ-WES-004) SHALL ensure the web console does not silently amplify the
sprawl. Explicit `moai worktree clean` remains the user-managed disposal
path for accumulated L1 ephemeral worktrees.

### REQ-WES-010 (Capability-gate — launcher `-w` accepts L2 absolute paths)

**Where** the `internal/cli/launcher.go` `-w` / `--worktree` flag handling
is in scope, the launcher SHALL accept `~/.moai/worktrees/<project>/...`
absolute paths as valid re-entry targets for L2 persistent worktrees (per
Decision Point 1 Resolution OQ-1). Today the launcher's
`normalizeWorktreeFlag` (launcher.go:665-702) performs ONLY CLI argument-
token normalization — rewriting `-w` / `--worktree[=name]` forms into the
canonical two-token `--worktree [name]` form that Claude Code accepts — and
does NOT perform path resolution; the value token is passed through to
`claude`, which resolves it against `.claude/worktrees/<name>`. The L2
path-split exists because no MoAI-side path-resolution step routes
`~/.moai/worktrees/` absolute paths before the claude pass-through. This
extension adds that MoAI-side path-resolution step (a new pre-resolution
function sibling to `normalizeWorktreeFlag`, OR an inlined early-return
branch within it) so an absolute path under `~/.moai/worktrees/` is
detected and routed to the L2 worktree rather than being silently treated
as a `.claude/worktrees/` short name. This extension is **additive**: the
existing token-normalization + claude pass-through path for
`.claude/worktrees/<name>` short-name inputs MUST remain behaviorally
intact. Feasibility is confirmed by the existing `cleanupMoaiWorktrees`
function (launcher.go:357-434), which ALREADY recognizes
`~/.moai/worktrees/` paths — the same path-awareness that disposal uses is
extended here to entry. The extension makes `moai cc -w <abs-path-under-
~/.moai/worktrees/>` a valid re-entry path for L2 worktrees, replacing the
bare-`cd` shell form previously required to reach L2 paths from a launcher
entry.

## §D. Constraints

- **GEARS notation** — every requirement uses one of the five GEARS patterns
  (Ubiquitous / Event-detected / State-driven / Capability-gate / Event-
  detected-for-undesired-condition). No legacy `IF/THEN` modality.
- **Cross-cutting scope** — touches `.claude/rules/` (markdown rules) +
  `internal/cli/launcher.go` (Go, `-w` flag behavior unchanged in code;
  documentation alignment only) + `internal/config/defaults.go` (Go, default
  mutation: `AutoCleanup: true → false`, `AutoMerge: true → false`) +
  `internal/web/assets/i18n.js` (i18n keys already present; defaults enforced
  by `defaults.go`) + `CLAUDE.local.md §22` (dev doc).
- **Sanitized-pair obligation** — any change to a file under
  `internal/template/templates/` MUST be mirrored to the corresponding local
  file under the user-facing project tree (`.claude/` or `.moai/`,
  depending on the file's distribution target) per CLAUDE.local.md §2
  [HARD] Template-First Rule and §25 Template Internal-Content Isolation.
  The web i18n.js and config defaults are NOT template-managed (they are
  project-specific), so this constraint binds only on rule-file changes
  that have template mirrors (e.g. `worktree-integration.md` /
  `session-handoff.md` mirrors at both
  `internal/template/templates/.claude/rules/moai/workflow/` and
  `.claude/rules/moai/workflow/`).
- **Launcher `-w` extension boundary (Decision Point 1 Resolution OQ-1)** —
  the existing token-normalization behavior of `normalizeWorktreeFlag`
  (`internal/cli/launcher.go:665-702`) — rewriting `-w` / `--worktree[=name]`
  into the canonical `--worktree [name]` form for claude pass-through,
  which then resolves the value against `.claude/worktrees/<name>` — MUST
  remain behaviorally intact. The extension is ADDITIVE: a new MoAI-side
  path-resolution step (sibling function OR inlined early-return branch
  within `normalizeWorktreeFlag`) detects absolute paths under
  `~/.moai/worktrees/<project>/...` and routes them to the L2 worktree
  BEFORE the token-normalization + claude pass-through path runs. The
  legacy short-name token-normalization path is NOT rewritten. The
  `cleanupMoaiWorktrees` function (the stale `worker-*` cleanup on
  `moai cc` entry, launcher.go:357-434) is preserved as-is.
- **Branch-guard compatibility** — SIBLING SPEC-WORKTREE-BRANCH-GUARD-001
  (#1192 `e89d01461`, merged) mechanically denies branch-state changes in the
  primary checkout. The auto-isolation procedure (REQ-WES-005) MUST NOT
  trigger branch-state changes in the primary checkout; it MUST create a new
  worktree under `.claude/worktrees/` or `~/.moai/worktrees/` (worktree paths
  are exempt from the branch-guard deny per its discriminant).

## §E. Success Criteria

- Every orchestrator-emitted Block 0 in the paste-ready resume uses either
  `moai cc -w <name>` or `EnterWorktree(<path>)` — never bare `cd`.
- The three web auto-toggles all default to `false` in
  `internal/config/defaults.go`.
- The doc surfaces listed in §B.6 reflect the EnterWorktree-first policy with
  no residual `cd`-as-canonical guidance.
- A parallel-session branch conflict is auto-isolated rather than surfaced as
  a manual race.
- Worktree sprawl is mitigated (re-entry preferred over creation; web console
  does not silently amplify).

## §F. Out of Scope

### Out of Scope — `Agent(isolation: "worktree")` L1 ephemeral creation behavior

- The Claude Code runtime's materialization of L1 ephemeral worktrees for
  `Agent(isolation: "worktree")` is owned by the runtime; this SPEC does NOT
  modify L1 creation semantics. It only clarifies that L1 is NOT a re-entry
  mechanism for existing L2 worktrees.
- The `cleanupMoaiWorktrees` function in launcher.go (removes stale `worker-*`
  L1 ephemeral worktrees on `moai cc` entry) is preserved as-is; its disposal
  behavior is NOT modified by this SPEC.

### Out of Scope — `MOAI_BRANCH_GUARD_EXEMPT=1` spawn contract

- The follow-up SPEC for the orchestrator-side
  `MOAI_BRANCH_GUARD_EXEMPT=1` spawn contract (Tier S/M, parent SPEC-
  WORKTREE-BRANCH-GUARD-001 backlog item) is a DIFFERENT SPEC. This SPEC
  references but does NOT absorb it.

### Out of Scope — `moai worktree new` BODP-gated creation

- `moai worktree new <SPEC-ID>` Branch Origin Decision Protocol (BODP) is
  owned by `.claude/rules/moai/development/branch-origin-protocol.md` and the
  `--team` flag P1-P4 launch matrix owned by
  `.claude/skills/moai-workflow-worktree/SKILL.md` § `--team` Flag. This SPEC
  does NOT modify BODP or the P1-P4 matrix.

### Out of Scope — Worktree State Guard forensic primitives

- `moai worktree snapshot|verify|restore` (owned by `worktree-state-guard.md`
  and SPEC forensics) is NOT modified by this SPEC. The State Guard remains
  a user-invoked primitive.

### Out of Scope — Worktree-auto-toggle UI affordance in the web console

- This SPEC changes the DEFAULTS of the three auto-toggles to `false`; it
  does NOT add new UI affordances, new toggle keys, or new i18n strings. The
  i18n keys `f.workflow.worktree.auto_{create,cleanup,merge}` already exist
  in all 4 locales (en/ko/ja/zh) and are reused unchanged.

### Out of Scope — `cd` prohibition in user-authored content

- This SPEC binds the ORCHESTRATOR's emitted guidance. It does NOT prohibit
  the human user from typing `cd <path>` in their own terminal. The bare
  `cd` form remains valid for human-typed, manual-shell contexts; the SPEC
  governs only orchestrator-emitted paste-ready / instruction surfaces.

### Out of Scope — EnterWorktree RUNTIME TOOL path-resolution semantics

- Per Decision Point 1 Resolution OQ-1, the Claude Code **runtime** tool
  `EnterWorktree(path)` (the in-session re-entry tool, owned by the Claude
  Code runtime — distinct from the MoAI-owned `moai cc -w` launcher) has
  path-resolution semantics that are NOT fully described by the existing
  docs. The current text in `.claude/rules/moai/workflow/worktree-integration.md`
  § `EnterWorktree` / `ExitWorktree` Tools (worktree-integration.md:148-150)
  states Claude "can switch directly to another worktree by calling
  `EnterWorktree` with a target path" (loose wording that suggests arbitrary
  paths for SWITCHING), while the canonical L1-creation flow creates a
  worktree under `.claude/worktrees/<name>/`. Which paths the runtime tool
  accepts for ad-hoc ENTRY (as opposed to switching once already inside a
  worktree) is NOT mechanically verified here. THIS RUNTIME-SEMANTICS
  CLARIFICATION IS OUT OF SCOPE for this SPEC. The launcher-layer extension
  (REQ-WES-010) is the MoAI-owned surface extended by this SPEC; clarifying
  / extending the runtime `EnterWorktree` tool's path semantics (including
  the switching-vs-entry distinction and any L2-aware alias) is deferred to
  a follow-up runtime-layer SPEC. A doc-clarification follow-up to align
  `worktree-integration.md:148-150` with whichever behavior the runtime
  actually exhibits is tracked as a sync-phase doc-debt item. The
  EnterWorktree-first policy in this SPEC applies to
  `.claude/worktrees/`-rooted paths, which is the surface the runtime tool
  demonstrably serves for L1 + ad-hoc entry.

### Out of Scope — `TmuxPreferred: true` default mutation

- Per Decision Point 1 Resolution OQ-4, the `TmuxPreferred: true` default
  (`internal/config/defaults.go:525`) is explicitly OUT OF SCOPE for this
  SPEC. It remains `true` unchanged. The web auto-toggles default-OFF
  policy (REQ-WES-004) covers only `auto_create` / `auto_cleanup` /
  `auto_merge` — it does NOT cover `TmuxPreferred`. A follow-up SPEC MAY
  revisit `TmuxPreferred` if the user-intended policy diverges.

## §G. Cross-References

- **Sibling**: `SPEC-WORKTREE-BRANCH-GUARD-001` (#1192 `e89d01461`, merged)
  — PreToolUse conditional-deny guard that mechanically blocks branch-state
  changes in the primary checkout. This SPEC's auto-isolation procedure
  (REQ-WES-005) creates a NEW worktree (worktree paths are exempt from the
  branch-guard deny).
- `.claude/rules/moai/workflow/worktree-integration.md` — Terminology
  Glossary (L1/L2/L3 layer definitions) + `EnterWorktree`/`ExitWorktree`
  section (the surface this SPEC expands).
- `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored
  Resume Pattern (Block 0) — the paste-ready resume surface this SPEC
  aligns.
- `.claude/rules/moai/workflow/session-handoff-examples.md` § Worktree-
  Anchored Resume Pattern — Block 0 example forms.
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync
  Check — the active-sessions registry this SPEC's auto-isolation procedure
  consumes.
- `CLAUDE.local.md` §22 (dev settings intent) — records the user-intended
  defaults for the web worktree auto-toggles.
- `internal/cli/launcher.go` — `normalizeWorktreeFlag` (the `-w` / `--worktree`
  token-normalization logic, baseline cite `launcher.go:665-702` — currently
  performs ONLY CLI argument-token normalization with NO path resolution;
  EXTENDED ADDITIVELY per REQ-WES-010 with a new sibling pre-resolution step
  that recognizes `~/.moai/worktrees/` absolute paths) and
  `cleanupMoaiWorktrees` (the stale-`worker-*` cleanup at launcher.go:357-434,
  unchanged by this SPEC).
- `internal/config/types.go:478` `WorkflowWorktreeConfig` struct +
  `internal/config/defaults.go:521,523` defaults — the surface this SPEC
  mutates (AutoCleanup @ L521, AutoCreate @ L522, AutoMerge @ L523).

## §H. HISTORY

- **0.1.0** (2026-07-28) — initial plan-phase draft. Tier L (5-artifact set:
  spec.md + plan.md + acceptance.md + design.md + research.md). Cross-
  references sibling SPEC-WORKTREE-BRANCH-GUARD-001. Baseline evidence:
  58 worktrees / 31 `agent-*` uncleaned observed 2026-07-28;
  `AutoCleanup: true` / `AutoMerge: true` current defaults require mutation
  to `false`.
- **0.2.0** (2026-07-28) — Decision Point 1 Resolutions folded in (plan→run
  HUMAN GATE clarifications resolved). Four open-question markers (formerly
  the `NEEDS-CLARIFICATION` inline form in v0.1.0 research.md) promoted to
  FIRM decisions: (OQ-1) launcher `-w` extended additively to
  accept `~/.moai/worktrees/` absolute paths → new REQ-WES-010 + new §B.2
  extension clause + constraint rewrite (the legacy resolution path remains
  behaviorally intact; only the new sibling branch resolves L2 absolute
  paths); the EnterWorktree RUNTIME TOOL's `.claude/worktrees/`-only
  constraint is OUT OF SCOPE (deferred to a follow-up runtime-layer SPEC);
  (OQ-2) conservative predicate confirmed (any foreign registry entry
  triggers auto-isolation; false positives are cheap); (OQ-3) naming scheme
  `auto-<session-short>-<spec-id>` confirmed (`<session-short>` = first 8
  chars of the foreign session's UUID; no "or equivalent" clause);
  (OQ-4) `TmuxPreferred: true` default explicitly OUT OF SCOPE. Title
  updated to surface the launcher extension. Spec scope now includes a
  real Go code change to `internal/cli/launcher.go` (REQ-WES-010), not
  only documentation alignment.
- **0.3.0** (2026-07-28) — plan-audit iter-1 (FAIL 0.72) delta fixes D1-D9:
  (D1) research.md NEEDS-CLARIFICATION inline markers at L57 / L194 marked
  `[RESOLVED per Decision Point 1]` — zero active markers remain;
  (D2) acceptance.md keeps Given-When-Then (GWT is the de facto project
  standard — sibling `SPEC-WORKTREE-BRANCH-GUARD-001/acceptance.md` uses
  GWT) with an explicit convention note in §A.2; (D3) §B.4 per-line cites
  corrected to AutoCleanup@521 / AutoCreate@522 / AutoMerge@523 (the §A
  range cite was already correct); (D4) REQ-WES-010 + §B.2 + §D constraint
  reframed to accurately describe `normalizeWorktreeFlag` as token-only
  normalization (no path resolution today; the extension adds a MoAI-side
  pre-resolution step — not a "sibling branch in an existing resolver",
  since no path-resolution branch currently exists); (D5) §F OOS
  EnterWorktree runtime-tool clause rewritten to reconcile with
  `worktree-integration.md:148-150` (switching-vs-entry semantics is the
  deferred runtime-layer SPEC; a sync-phase doc-clarification is tracked);
  (D6) §D sanitized-pair copy/paste error fixed (destination was identical
  to source); (D7) REQ-WES-010 relocated to after REQ-WES-009 (numeric
  order restored); (D8) REQ-WES-005 body clarified for the N-foreign-
  session case (one worktree per foreign entry, matching Edge-3);
  (D9) REQ-WES-009 reframed from State-driven (hardcoded 50/30 thresholds
  that would stop binding once sprawl is cleaned) to Ubiquitous (prefer
  re-entry; sprawl baseline moved to §A motivation + §F 30-day re-baseline).
- **0.3.1** (2026-07-28) — plan-auditor iter-1 (FAIL 0.84, Tier L 0.85)
  remediation targeted at iter-2 PASS: D1 research.md NEEDS-CLARIFICATION
  inline markers at L57 / L194 re-authored per auditor option (a) — the
  literal NEEDS-CLARIFICATION prefix is preserved (grep still locates both
  markers for the audit trail) and the resolution is co-located via an
  inline RESOLVED v0.2.0 clause citing plan.md §I OQ-1 / OQ-2;
  D2 acceptance.md AC-WES-006 verification command anchored to the EXISTING
  `TestNormalizeWorktreeFlag` only (dropped the non-existent
  `TestLauncherWorktreeFlag` from the alternation regex — the AC's intent
  "normalizeWorktreeFlag byte-identical, no code change" is fully covered
  by the existing test at launcher_test.go:248); the same vacuous-selector
  pattern in AC-WES-010b's verification command was also dropped for
  consistency; D4 spec.md §A + §G cites anchored to the ACTUAL mutation lines
  `defaults.go:521,523` (AutoCleanup @ L521, AutoMerge @ L523; AutoCreate @ L522
  already `false`) — the plan-auditor iter-1 D4 finding (`:520,522`) was a
  line misread of defaults.go; corrected after manager-spec flagged the
  live-file discrepancy. cleanupMoaiWorktrees cite already at `launcher.go:357`
  — no drift to correct); D5 REQ-WES-010 already in numerical position
  (after REQ-WES-009, before §D Constraints) from the prior 0.3.0 cycle —
  no relocation needed. D3 (AC-WES-005b doc-grep-only verification for the
  branch-guard exemption) ACCEPTED as residual risk — the sibling
  SPEC-WORKTREE-BRANCH-GUARD-001 owns the discriminant; a runtime test is
  out of scope for plan-phase.
