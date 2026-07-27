---
id: SPEC-WORKTREE-BRANCH-GUARD-001
title: Main-Checkout Branch-State Guard via PreToolUse Conditional Deny
version: 0.1.2
status: in-progress
created: 2026-07-28
updated: 2026-07-28
author: manager-spec
priority: High
phase: plan
module: hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, pretool, branch-guard, worktree, main-checkout, template-mirror, sanitized-pair, census-p1b"
related_specs:
  - SPEC-WORKTREE-001
  - SPEC-WORKTREE-002
  - SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001
  - SPEC-V3R5-LATE-BRANCH-001
---

# SPEC-WORKTREE-BRANCH-GUARD-001 — Main-Checkout Branch-State Guard

## §A. Problem Statement

The main project checkout (`/Users/goos/MoAI/moai-adk-go/`) is **shared** mutable
state. Multiple concurrent Claude Code sessions, teammates, hooks, and background
tools operate on the same working tree. A branch-state-changing git command
(`git switch`, `git checkout <branch>`, `git checkout -b`, `git branch`,
`git reset --hard`, `git stash`, `git rebase`, `git merge` onto the checked-out
branch) issued in one session relocates every other session's tree mid-operation,
with no signal to either side. The race is quiet: failures surface later as
"commits I did not make" or "my changes are on the wrong branch" — see
`.claude/rules/moai/workflow/main-checkout-branch-guard.md` (v1.0.0) for the
doctrine that this SPEC mechanically enforces.

The settings.json permission layer CANNOT enforce conditional isolation here:

- Local `.claude/settings.json:309` is `bypassPermissions`; template
  `internal/template/templates/.claude/settings.json.tmpl:439` is `acceptEdits`.
  Both auto-approve.
- A static `deny` entry (already present at `.claude/settings.json:341, 350-353,
  457, 460` for `git branch`, `git checkout`, `git switch`, `git stash`,
  `git merge`, `git reset --hard`, `git rebase -i`) cannot scope to the primary
  checkout only — it fires in worktrees too, causing merge lockout (proven in
  handoff memory `project_worktree_branch_guard_handoff`). It also contradicts
  `manager-git.md:117-123` Phase D (Late-Branch closure), which legitimately
  runs `git checkout main && git reset --hard origin/main` in the primary
  checkout.

The only mechanism that can block branch-state changes in the primary checkout
while permitting them in worktrees and for the trusted git agent is a
**PreToolUse hook with conditional deny** — the existing `handle-pre-tool.sh`
wrapper + `internal/hook/pre_tool.go` handler already provide the infra.

## §B. Scope — 5 Components

1. **PreToolUse hook — conditional deny**: deny branch-state-changing git
   commands when ALL three conditions hold: (a) primary checkout, (b) the
   command matches a branch-state pattern, (c) invoking agent is not exempt.
2. **Discriminant — primary vs worktree**: compare absolute git-dir against
   absolute git-common-dir; equal → primary, different → worktree.
3. **Branch-guard rule v1.0.0 → v1.1.0**: bump doctrine rule + template mirror
   (byte-parity) to document the new mechanical enforcer.
4. **`.worktreeinclude` commit**: reconcile content divergence between repo-root
   and template copies; commit both. Template is SSOT per CLAUDE.local.md §2.
5. **CLI worktree advisory**: `moai init` / `moai update` / `moai web` surface
   advisory wording guiding users toward worktree isolation, consuming
   `workflow.worktree` (`workflow.yaml:112`).

## §C. Requirements (GEARS notation)

### REQ-WBG-001 — Conditional Deny (Event-detected + compound)

**Where** the PreToolUse event fires for a Bash tool invocation whose command
matches a branch-state-changing pattern, **While** the invocation occurs in the
primary checkout (absolute git-dir equals absolute git-common-dir), **When** the
invoking agent identity is not exempt per REQ-WBG-011, the branch-guard handler
shall emit `hookSpecificOutput.permissionDecision: "deny"` with a reason string
prefixed by the sentinel `BRANCH_GUARD_VIOLATION`.

### REQ-WBG-002 — Worktree Allow (Capability gate)

**Where** the invocation occurs in a git worktree (absolute git-dir differs from
absolute git-common-dir), the branch-guard handler shall emit
`permissionDecision: "allow"` regardless of the command pattern or agent
identity.

### REQ-WBG-003 — Exempt-Agent Allow (Capability gate)

**Where** the invoking agent identity matches the `manager-git` trusted agent
(per REQ-WBG-011 resolution), the branch-guard handler shall emit
`permissionDecision: "allow"` regardless of the command pattern or checkout
context.

### REQ-WBG-004 — Discriminant Comparison (Ubiquitous)

The branch-guard handler shall determine primary-vs-worktree by comparing the
absolute path of `git rev-parse --git-dir` against the absolute path of
`git rev-parse --git-common-dir`; equal paths classify as primary checkout,
differing paths classify as worktree.

### REQ-WBG-005 — Discriminant Portability (State-driven)

**While** the host git supports `--path-format=absolute` (git 2.31+, March 2021),
the branch-guard handler shall use it for both rev-parse invocations; **When**
the host git rejects `--path-format=absolute` (older git or Apple Git fallback),
the handler shall fall back to `git rev-parse --absolute-git-dir` for git-dir and
cwd-normalized `git rev-parse --git-common-dir` for the common dir.

### REQ-WBG-006 — Rule v1.1.0 Bump (Ubiquitous)

The doctrine rule `.claude/rules/moai/workflow/main-checkout-branch-guard.md`
shall carry `Version: 1.1.0` and document the PreToolUse conditional-deny hook
as the mechanical enforcer, superseding the v1.0.0 doctrine-only text.

### REQ-WBG-007 — Template Mirror Sanitized-Pair (Ubiquitous)

The template mirror
`internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`
shall be **§25-sanitized**: the SPEC-ID line and the cross-reference bullet
naming `SPEC-WORKTREE-BRANCH-GUARD-001` carried in the source rule shall be
replaced in the mirror with generic prose
(`Origin: the run-phase SPEC that landed the v1.1.0 mechanical enforcer.`)
and the SPEC-ID cross-reference bullet shall be dropped. The source rule
`.claude/rules/moai/workflow/main-checkout-branch-guard.md` RETAINS the
SPEC-ID + REQ tokens for traceability (REQ-WBG-006). The mirror SHALL be
enrolled in the `sanitizedPairPaths` registry in
`internal/template/sanitized_pair_parity_test.go` so
`TestSanitizedPairParity` enforces doctrine parity between source and
sanitized mirror, and `TestTemplateNoInternalContentLeak` enforces the
mirror's §25 cleanliness. The rule is intentionally excluded from the
byte-parity allowlist `workflowOptMirroredPaths` in
`internal/template/rule_template_mirror_test.go` (a NOTE comment in that
file records the intentional exclusion).

**Rationale — §25 forbids a SPEC-ID in the template mirror**: every existing
byte-mirrored rule (`spec-workflow.md`, `session-handoff.md`, `hooks-system.md`,
`model-policy.md`) achieves byte-parity because the SOURCE is also §25-clean.
This rule's source is NOT §25-clean (it carries the SPEC-ID for REQ-WBG-006
traceability), so byte-parity is impossible without violating §25's
`TestTemplateNoInternalContentLeak` guard. The sanitized-pair pattern is the
established resolution — it matches the `runtime-recovery-doctrine.md` /
`zone-registry.md` precedent.

### REQ-WBG-008 — `.worktreeinclude` Reconciliation (Ubiquitous)

The repo-root `.worktreeinclude` and the template
`internal/template/templates/.worktreeinclude` shall be reconciled to
byte-identity against the template SSOT, and both committed in the same
sync-phase commit. The `teammateMode` mention in the local copy is redundant
prose (§22.3 documents teammateMode as a `settings.local.json` field, and
`.worktreeinclude` already copies `settings.local.json`).

### REQ-WBG-009 — CLI Worktree Advisory (Ubiquitous)

The `moai init`, `moai update`, and `moai web` CLI commands shall surface
advisory wording that (a) names the shared-checkout hazard, (b) recommends
worktree isolation for branch-changing work, and (c) consumes the
`workflow.worktree` config key (`workflow.yaml:112`) to honor the project's
auto-create setting when surfacing the recommendation.

### REQ-WBG-010 — Census P1-B Timeout-Safety (State-driven)

**While** the PreToolUse timeout budget is 5 seconds, the branch-guard handler
shall emit its deny decision on the new `checkBranchState` regex path
(microseconds, structurally analogous to but distinct from `checkBashCommand`),
NOT on the quality-gate path (`quality.IsGitCommit` + `quality.NewQualityGate`).
The existing `checkBashCommand` iterates `SecurityPolicy.DangerousBashPatterns`
which does NOT include branch-state commands — the new deny comes from the NEW
`checkBranchState` regex introduced by M2 of this SPEC. The quality-gate path
is structurally inaccessible to branch-state commands because `IsGitCommit`
only matches `git commit` invocations (census P1-B measured 10.58s on that path
— out of reach of the 5s budget). The branch-guard deny SHALL complete in
measured wall-time ≤ 500ms per invocation.

### REQ-WBG-011 — manager-git Exemption Boundary (Ubiquitous, design-decision)

The exemption for the `manager-git` trusted agent shall be **identity-based and
unconditional within the agent identity**, NOT scoped to Phase D. The agent
identity is determined by EITHER:

- (a) `HookInput.AgentType == "manager-git"` (main-thread `claude --agent
  manager-git` invocation), OR
- (b) environment variable `MOAI_BRANCH_GUARD_EXEMPT=1` (orchestrator-set when
  spawning `Agent(manager-git, ...)` for Phase D work).

**Rationale — research finding that contradicts a verified precondition**: the
resume states "The hook receives the agent type via the PreToolUse stdin JSON".
Research confirmed `HookInput.AgentType` exists (JSON tag `agent_type`,
`internal/hook/types.go:226`) but its docstring reads "Custom agent name if
`--agent` flag used" — main-thread `--agent` invocation only. For `Agent(manager-git,
...)` sub-agent spawns, Claude Code does NOT reliably populate `agent_type` in
the subagent's PreToolUse events. The sentinel env var (b) closes the gap.

**Env-var threat model**: hooks are spawned by Claude Code (not by agent Bash
subprocesses), so an agent's `export MOAI_BRANCH_GUARD_EXEMPT=1` via Bash does
NOT propagate to the hook's env — narrow exploit surface. However, any process
that can mutate Claude Code's own env (a SessionStart hook, or the
`settings.json` `env` block) could set the var globally and silently disable
the guard. This boundary is accepted: the `settings.json` `env` block is a
trusted surface (CLAUDE.local.md §22), and SessionStart hooks are MoAI-managed.
A future SPEC may add a denylist guard against `MOAI_BRANCH_GUARD_EXEMPT` in
the `settings.json` `env` block if threat modeling warrants.

**Phase-D scoping rejected**: scoping the exemption to "Phase D only" would
require the hook to parse agent-internal state (which phase it is in), which is
not observable at PreToolUse time. The exemption is identity-based because
`manager-git.md` Phase D is the ONLY place the project authorizes branch-state
changes in the primary checkout.

### REQ-WBG-012 — Fail-Open with Advisory (Event-detected)

**When** the branch-guard handler cannot determine the git context (not a git
repo, git binary missing, `git rev-parse` exits non-zero) OR cannot determine
the agent identity (both `HookInput.AgentType` and
`MOAI_BRANCH_GUARD_EXEMPT` env var are empty/absent), the handler shall emit
`permissionDecision: "allow"` AND write an advisory message to stderr AND append
a structured entry to `.moai/logs/branch-guard-audit.log`.

**Rationale**: aligns with the Bash Risk-Amplifier Doctrine (WARN-ONLY,
FAIL-OPEN in `.claude/rules/moai/development/coding-standards.md`) and the
Recovery-Signal Carve-Out. The deny is emitted only on positive evidence
(primary checkout confirmed AND branch-state pattern matched AND agent not
exempt); uncertainty falls through to allow with an auditable advisory.

### REQ-WBG-013 — Non-reliance on Static Deny (Unwanted)

The system shall NOT rely on a static `settings.json` `deny` entry as the
primary enforcement for branch-state commands in the primary checkout, because
static deny cannot scope to the primary checkout (it fires in worktrees too)
and causes merge lockout for legitimate worktree flows. The existing static
`deny` entries at `.claude/settings.json:341, 350-353, 457, 460` are retained
as defense-in-depth for non-hook-enabled contexts only.

## §D. Constraints

- **Tier M**: 4-artifact plan-phase set (spec/plan/acceptance/progress).
- **Template-First** (CLAUDE.local.md §2 [HARD]): every template change lands in
  `internal/template/templates/` first; the local mirror follows in the same
  commit for byte-parity.
- **Language neutrality** (CLAUDE.local.md §15): the template portions must not
  elevate any of the 16 supported languages. The rule wording is
  language-neutral (it speaks of git, not of Go).
- **Flat hierarchy**: the hook logic extends the existing `preToolHandler` in
  `internal/hook/pre_tool.go`; no new top-level package.
- **verification-claim-integrity**: every AC names a command whose verbatim
  output decides PASS/FAIL; no AC relies on grep token-presence alone.

## §E. Out of Scope

### Out of Scope — settings.json static-deny approach

- Migrating branch-state enforcement to a settings.json-only solution. The
  static deny is retained as defense-in-depth for non-hook-enabled contexts but
  is NOT the primary enforcer (REQ-WBG-013).
- Removing the existing static `deny` entries at `.claude/settings.json:341,
  350-353, 457, 460`. Their reconciliation with `manager-git.md` Phase D is a
  separate concern tracked in census P1-B residual debt.

### Out of Scope — Sub-agent spawn metadata plumbing

- Modifying Claude Code runtime to populate `agent_type` in PreToolUse events
  for `Agent(manager-git, ...)` sub-agent spawns. The SPEC uses the sentinel
  env var `MOAI_BRANCH_GUARD_EXEMPT=1` workaround (REQ-WBG-011b); deeper
  runtime plumbing is out of scope.
- Adding new fields to the `HookInput` struct beyond reading the existing
  `AgentType` field and the env var.

### Out of Scope — manager-git Phase D worktree migration

- Migrating `manager-git.md` Phase D (Late-Branch closure) to operate in a
  worktree rather than the primary checkout. The SPEC's exemption mechanism
  (REQ-WBG-011) keeps Phase D working as documented; a future SPEC may migrate
  Phase D to a worktree to eliminate the exemption entirely.

### Out of Scope — Worktree auto-creation

- Flipping `workflow.worktree.auto_create` from `false` to `true`
  (`workflow.yaml:117`). The 2026-05-22 user-policy decision
  (`feedback_worktree_autonomous`) keeps L1 worktree creation as a Claude Code
  runtime-autonomous step; MoAI does not auto-create.

### Out of Scope — Legacy EARS migration

- Re-authoring the 88 pre-v3 SPECs from EARS to GEARS notation. The 6-month
  backward-compatibility window remains active; this SPEC uses GEARS but does
  not migrate legacy SPECs.

### Out of Scope — Bash Risk-Amplifier soft-cap warn signal

- Modifying the existing Bash subcommand-count warn signal in
  `handle-pre-tool.sh` (lines 33-58). That signal is WARN-ONLY and
  independent; this SPEC adds a deny path, not a warn path.

## §F. Cross-References

- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` (v1.0.0 — to be
  bumped to v1.1.0 per REQ-WBG-006)
- `internal/hook/pre_tool.go` (`preToolHandler.Handle`, `checkBashCommand`,
  `SecurityPolicy.DangerousBashPatterns`)
- `internal/hook/types.go:208-280` (`HookInput` struct, `AgentType` field)
- `internal/hook/quality/gate.go:623-625` (`IsGitCommit` — the census P1-B slow
  path that branch-state commands structurally avoid)
- `internal/template/rule_template_mirror_test.go` (`workflowOptMirroredPaths`
  byte-parity allowlist — REQ-WBG-007 intentionally EXCLUDES the rule path; see
  the NOTE comment in that file)
- `internal/template/sanitized_pair_parity_test.go` (`sanitizedPairPaths`
  registry — REQ-WBG-007 enrolls the rule path at line 91)
- `.claude/agents/moai/manager-git.md:117-127` (Phase D — Late-Branch closure)
- `.moai/config/sections/workflow.yaml:112-120` (`workflow.worktree` config)
- Census P1-B: handoff memory `project_census_priority1_handoff` (5s budget
  destroys deny on the quality-gate path)
- Handoff memory `project_worktree_branch_guard_handoff` (motivating analysis)

## §G. HISTORY

- 2026-07-28 v0.1.0 — initial draft (manager-spec, plan-phase). Verified all 5
  preconditions; resolved all 3 design questions; surfaced 1 research
  contradiction (agent_type field is main-thread-only — closed with sentinel
  env var per REQ-WBG-011).
- 2026-07-28 v0.1.1 — plan-audit iter-2 fixes (PASS-With-Debt 0.84 → fixes
  applied): D1 honest Phase-D break framing (F-1/B-1 aligned with §A valence);
  D2 AC-WBG-005 mock-exec.Command dispatcher injection (matches plan M1); D3
  REQ-WBG-010 `checkBranchState` wording + deny-origin AC arm; D4 E-3
  false-positive DENY (not fail-open); D5 AC-WBG-009 temp-dir for init (no
  --dry-run flag); D6 AC-WBG-010 self-consistent Test* + internal loop; D7
  `[NEEDS CLARIFICATION: ...]` marker convention; N1 word-boundary regex;
  N2 env-var threat model.
- 2026-07-28 v0.1.2 — REQ-WBG-007/AC-WBG-007 corrected from byte-parity to
  sanitized-pair (§25 forbids a SPEC-ID in the template mirror; the
  orchestrator's M3 delegation premise that mirrored rules may carry
  SPEC-IDs was false — verified via `TestTemplateNoInternalContentLeak`.
  Sanitized-pair matches the runtime-recovery-doctrine / zone-registry
  precedent.).
