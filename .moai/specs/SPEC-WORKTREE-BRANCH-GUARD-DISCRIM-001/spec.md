---
id: SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001
title: "Main-Checkout Branch-State Guard — Worktree Discriminant Directory Correction"
version: "0.1.0"
status: in-progress
created: 2026-08-13
updated: 2026-08-13
author: manager-spec
priority: High
phase: "v3.0.3"
module: internal/hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, branch-guard, worktree, discriminant, bug-fix, sanitized-pair"
depends_on:
  - SPEC-WORKTREE-BRANCH-GUARD-001
  - SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
related_specs:
  - SPEC-WORKTREE-BRANCH-GUARD-001
  - SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
---

# SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 — Worktree Discriminant Directory Correction

## §A. Problem Statement

The Main-Checkout Branch-State Guard landed by **SPEC-WORKTREE-BRANCH-GUARD-001**
(status: completed) and refined by **SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001**
(status: completed) misclassifies a Claude-native worktree
(`.claude/worktrees/<name>/`) as the primary checkout. When the orchestrator
operates inside such a worktree and a legitimate branch-state command
(`git rebase`, `git push --force-with-lease`, `git switch -c`, etc.) is issued
there, the PreToolUse hook denies the command with
`BRANCH_GUARD_VIOLATION: <suffix> in primary checkout`. The
`MOAI_BRANCH_GUARD_EXEMPT=1` env prefix does NOT work around this because the
hook runs in a separate Claude-Code-spawned process that does not inherit the
agent's Bash env (the same env-propagation boundary documented in -001
REQ-WBG-011's threat model).

### Directly observed (PR #1476 session, worktree `.claude/worktrees/hook-pretool-perf`)

Running `git rebase` and `git push --force-with-lease` inside the worktree was
denied. Measured inside the worktree:

- `git rev-parse --git-dir` →
  `/Users/goos/MoAI/moai-adk-go/.git/worktrees/hook-pretool-perf`
- `git rev-parse --git-common-dir` →
  `/Users/goos/MoAI/moai-adk-go/.git`
- `git rev-parse --show-toplevel` → the worktree path

These git-dir / git-common-dir values **differ** → the doctrine's discriminant
classifies the context as a worktree (not primary). Yet the guard still returned
primary. The discriminant was being applied to the **wrong directory**.

### Root cause (measured, not inferred)

`internal/hook/pre_tool.go:487` passes `h.projectRoot()` to `checkBranchState`.
`h.projectRoot()` resolves via `resolveProjectRootFromEnv`
(`internal/hook/path_resolve.go:68`), which prefers `$CLAUDE_PROJECT_DIR` first
and falls back to `os.Getwd()`. `$CLAUDE_PROJECT_DIR` is set by the Claude Code
runtime to the **primary checkout root**, even when the active agent is
operating inside a worktree (this is the documented B7 invariant at
`internal/hook/CLAUDE.md §B7`: *"Never assume cwd equals project root —
worktree hooks run with cwd inside worktrees/"*).

`isPrimaryCheckout(projectDir)` then runs
`git -C <primary-checkout> rev-parse --git-dir --git-common-dir`, which returns
equal paths (both resolve to the primary's `.git`) → classifies as primary →
emits the deny.

But the agent's actual git command runs with cwd = the worktree path, where the
two rev-parse values DIFFER. The discriminant answers *"is the primary checkout
the primary checkout?"* instead of *"is the cwd the agent is about to operate in
the primary checkout?"* — it queries the wrong directory.

### Why candidate causes (1) and (3) are rejected

The motivating brief proposed three candidate causes: (1) `--path-format=absolute`
applying asymmetrically to `--git-dir` vs `--git-common-dir`, (2) projectDir
source, and (3) output-line parsing. Direct measurement in the worktree (the two
rev-parse values printed as distinct absolute paths) disproves (1) and (3): the
flag applied uniformly and the two-line parser split correctly. The bug is (2):
the projectDir passed in is the primary checkout, so the discriminant is
answering the wrong question.

## §B. Scope

### In Scope

1. **Discriminant directory correction.** The git-context directory that
   `isPrimaryCheckout` queries MUST be the cwd the Bash command will actually
   execute in (authoritatively `input.CWD` from the PreToolUse hook payload),
   with the same fallback chain the rest of the hook system already uses when
   `input.CWD` is absent. `$CLAUDE_PROJECT_DIR` MUST NOT be the primary source
   for the git-context query.

2. **Audit-log placement invariant.** `appendBranchGuardAdvisory` MUST continue
   to write to `<primary-checkout>/.moai/logs/branch-guard-audit.log`
  (resolved via `$CLAUDE_PROJECT_DIR` → `os.Getwd()` fallback) — it MUST NOT
   rotate to `<command-cwd>/.moai/logs/`. The git-context directory and the
   audit-log project directory were the same variable before and now MUST
   diverge; the fix MUST separate these two concerns rather than mutate one
   variable to serve both.

3. **Primary-checkout regression.** Primary-checkout detection MUST still
   classify the primary checkout as primary and deny branch-state commands there
   for non-exempt agents. The fix changes behavior ONLY for the worktree case.

4. **Doctrine bump** `main-checkout-branch-guard.md` v1.1.0 → v1.2.0
   documenting the discriminant directory correction, with sanitized-pair mirror
   parity per REQ-WBG-007 of -001.

5. **Worktree fixture unit test** asserting the discriminant returns
   `isPrimaryCheckout == false` for a real worktree git context.

### Out of Scope

#### Out of Scope — Discriminant mechanism redesign

- Replacing the git-dir vs git-common-dir comparison with a different predicate
  (e.g. `git rev-parse --is-inside-work-tree` plus a separate worktree-list
  check). The current discriminant is correct in principle; only the directory
  it was applied to was wrong.
- Adding a `git worktree list` cross-check.

#### Out of Scope — B7 resolution policy change for other handlers

- Modifying `$CLAUDE_PROJECT_DIR`-first resolution for handlers other than the
  branch guard. The B7 priority order is correct for handlers that want the
  project root (post_tool_metrics, observability, session-start env injection,
  etc.). Only the branch guard needs the actual command cwd.

#### Out of Scope — Exemption mechanism

- Touching `MOAI_BRANCH_GUARD_EXEMPT` semantics or the `manager-git` identity
  check. Preserved byte-identical per -OPTIN-001 REQ-6.
- Adding a new exemption surface for worktree-resident agents. The fix is in the
  discriminant, not the exemption logic.

#### Out of Scope — Late-Branch worktree migration

- Migrating `manager-git.md` Phase D (Late-Branch closure) to operate in a
  worktree rather than the primary checkout. Still out of scope per -001 §E.

#### Out of Scope — Legacy EARS migration

- Re-authoring pre-v3 SPECs from EARS to GEARS. The 6-month backward-compatibility
  window remains active; this SPEC uses GEARS but does not migrate legacy SPECs.

#### Out of Scope — Static deny reconciliation

- Reconciling the existing static `deny` entries at `.claude/settings.json` with
  the hook behavior. Out of scope per -001 §E.

## §C. Requirements (GEARS notation)

### REQ-WBG-D-001 — Discriminant Queries the Command CWD (Ubiquitous)

The branch-state guard SHALL determine primary-vs-worktree by querying the git
context of the working directory the Bash command will actually execute in
(authoritatively `input.CWD` from the PreToolUse hook payload), NOT the project
root resolved from `$CLAUDE_PROJECT_DIR`. When `input.CWD` is absent, the guard
SHALL fall back to the same `CLAUDE_PROJECT_DIR` → `os.Getwd()` chain the rest
of the hook system uses, and the choice SHALL be reflected in the audit-log
advisory.

### REQ-WBG-D-002 — Worktree Allow (Event-detected)

**When** the actual command cwd resolves to a git worktree (absolute git-dir
differs from absolute git-common-dir at that cwd), the branch-state guard SHALL
emit `permissionDecision: "allow"` for branch-state commands, regardless of
whether `$CLAUDE_PROJECT_DIR` points at the primary checkout.

### REQ-WBG-D-003 — Primary Deny Preserved (State-driven)

**While** the actual command cwd resolves to the primary checkout (absolute
git-dir equals absolute git-common-dir at that cwd), the branch-state guard
SHALL continue to deny branch-state commands for non-exempt agents — preserving
the -001 / -OPTIN-001 behavior without regression.

### REQ-WBG-D-004 — Audit-Log Placement Invariant (Ubiquitous)

The branch-state guard SHALL write its fail-open advisory audit entries to
`<primary-checkout>/.moai/logs/branch-guard-audit.log` (resolved via
`$CLAUDE_PROJECT_DIR` → `os.Getwd()` fallback), NOT to
`<command-cwd>/.moai/logs/` — independent of which directory the git-context
discriminant queries.

### REQ-WBG-D-005 — Fail-Open Preserved (Event-detected)

**When** the resolved command cwd is not a git repo, the git binary is missing,
OR `git rev-parse` exits non-zero at that cwd, the branch-state guard SHALL emit
`permissionDecision: "allow"` AND write the advisory audit entry per
REQ-WBG-D-004, preserving the -001 REQ-WBG-012 fail-open contract.

### REQ-WBG-D-006 — Doctrine Bump v1.2.0 (Ubiquitous)

The doctrine rule `.claude/rules/moai/workflow/main-checkout-branch-guard.md`
SHALL carry `Version: 1.2.0` and document that the primary-vs-worktree
discriminant queries the actual command cwd (`input.CWD`), not
`$CLAUDE_PROJECT_DIR`. The sanitized-pair mirror at
`internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`
SHALL be kept in parity (the rule path remains enrolled in `sanitizedPairPaths`
per -001 REQ-WBG-007; no enrollment change is required, only content parity).

### REQ-WBG-D-007 — Exemption Boundary Unchanged (Ubiquitous)

The exemption mechanism (`HookInput.AgentType == "manager-git"` OR
`MOAI_BRANCH_GUARD_EXEMPT=1` env var, per -001 REQ-WBG-011) SHALL remain
byte-identical. The fix changes ONLY the discriminant directory; the exemption
evaluation precedes the discriminant and is untouched.

### REQ-WBG-D-008 — Latency Ceiling Preserved (State-driven)

**While** the branch-guard handler runs under the PreToolUse timeout budget,
the git-context query against the actual command cwd SHALL complete within the
same per-invocation latency ceiling as the -001 REQ-WBG-010 budget (measured
wall-time ≤ 500ms). The fix introduces NO additional git subprocess beyond the
existing single `git -C <cwd> rev-parse --path-format=absolute --git-dir
--git-common-dir` call (plus the existing fallback only when the primary path
errors).

## §D. Constraints

- **Tier M**: 3-artifact plan-phase set (spec / plan / acceptance) + progress.md
  §E skeleton.
- **Repo-local PR-mandatory policy** (`repo-local-pr-policy.md`): ALL tier work
  lands via PR (Route B); direct-to-main is disabled by branch protection
  (`enforce_admins: true`).
- **Template-First** (CLAUDE.local.md §2 [HARD]): doctrine rule edits land in
  `internal/template/templates/` first; local mirror follows in the same commit
  for sanitized-pair parity.
- **Language neutrality** (CLAUDE.local.md §15): template portions must not
  elevate any of the 16 supported languages. The rule wording is language-
  neutral (speaks of git, not of Go).
- **Flat hierarchy**: the fix extends the existing `preToolHandler.Handle` →
  `checkBranchState` surface; no new package.
- **verification-claim-integrity**: every AC names a command whose verbatim
  output decides PASS/FAIL; no AC relies on grep token-presence alone.

## §E. Acceptance Criteria (summary — full Given-When-Then in acceptance.md)

| AC | Subject | Verifiable by |
|----|---------|---------------|
| AC-WBG-D-001 | Worktree discriminant: `isPrimaryCheckout(worktree-cwd)` returns `(false, nil)` | Go test fixture with a real worktree git context |
| AC-WBG-D-002 | Worktree allow end-to-end: `git rebase` in worktree cwd → `permissionDecision: "allow"` | PreToolUse handler test with `input.CWD = <worktree>` |
| AC-WBG-D-003 | Primary deny preserved: `git switch` in primary cwd → deny (regression) | PreToolUse handler test with `input.CWD = <primary>` |
| AC-WBG-D-004 | Audit-log placement on primary: fail-open entry lands at `<primary>/.moai/logs/branch-guard-audit.log` even when command cwd is a worktree | Test asserting the audit log path |
| AC-WBG-D-005 | Fail-open preserved at non-git command cwd | Test with `input.CWD = t.TempDir()` (non-git) |
| AC-WBG-D-006 | Exemption unchanged: `MOAI_BRANCH_GUARD_EXEMPT=1` still bypasses | Existing exemption test still passes |
| AC-WBG-D-007 | Latency ceiling ≤ 500ms per invocation | Benchmark / measured wall-time test |
| AC-WBG-D-008 | Doctrine v1.2.0 + sanitized-pair mirror parity | `grep '^Version: 1.2.0'` + sanitized-pair parity test green |

## §F. Cross-References

- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` (v1.1.0 → v1.2.0
  per REQ-WBG-D-006) — doctrine SSOT.
- `internal/hook/branch_guard.go` — `isPrimaryCheckout` (L167), `runGitRevParse`
  (L209), `checkBranchState` (L231), `appendBranchGuardAdvisory` (L276).
- `internal/hook/pre_tool.go:487` — call site passing `h.projectRoot()` (the bug
  locus).
- `internal/hook/path_resolve.go:68` — `resolveProjectRootFromEnv`
  (`CLAUDE_PROJECT_DIR`-first; the helper whose priority is wrong for the
  branch-guard use case).
- `internal/hook/path_resolve.go:88` — `resolveProjectRootFromInputOrEnv`
  (`input.CWD`-first; candidate helper for the fix).
- `internal/hook/types.go:212` — `HookInput.CWD` field (authoritative source for
  the command cwd).
- `internal/hook/CLAUDE.md §B7` — `$CLAUDE_PROJECT_DIR` resolution priority
  (documents the cwd-does-not-equal-project-root worktree invariant).
- `internal/hook/protocol.go:113-119` — `validateInput` `input.CWD` fallback to
  `$CLAUDE_PROJECT_DIR` when Claude Code omits cwd.
- Predecessors: SPEC-WORKTREE-BRANCH-GUARD-001 (status: completed),
  SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 (status: completed).
- PR #1476 session — directly observed denial inside worktree
  `.claude/worktrees/hook-pretool-perf`.

## §G. HISTORY

- 2026-08-13 v0.1.0 — initial draft (manager-spec, plan-phase). Root cause
  confirmed by direct measurement: `$CLAUDE_PROJECT_DIR` (= primary checkout)
  was being passed to `isPrimaryCheckout`, while the agent's git command ran in
  a worktree. Rejected candidate causes (1) `--path-format=absolute` scope and
  (3) output-line parsing per the directly-observed distinct rev-parse values.
  Audit-log placement invariant (REQ-WBG-D-004) separated from git-context
  directory to preserve central logging.
