---
id: SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001
title: "Main-Checkout Branch-State Guard — Convert to Default-OFF (Opt-In) + Read-Only Pattern Refinement"
version: 0.1.0
status: completed
created: 2026-07-30
updated: 2026-07-30
author: manager-spec
priority: high
phase: "v3.0.2"
module: hook
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, branch-guard, config, template-neutrality, opt-in"
related_specs: [SPEC-WORKTREE-BRANCH-GUARD-001]
---

# SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001

## §A Problem / Motivation

The Main-Checkout Branch-State Guard landed by **SPEC-WORKTREE-BRANCH-GUARD-001** (status: completed) is wired through the TEMPLATE — meaning it ships to ALL moai-adk users, not just maintainers of multi-session shared checkouts. Two problems were directly observed during this session:

### Problem A — ships default-on to all users (no config gate)

The guard fires in every user's primary checkout. A 1-person developer (the common distribution case) cannot freely run `git stash` / `git switch` / `git reset --hard` / `git checkout <branch>` in their own single-checkout repo. The guard's intent (shared-checkout multi-session protection) does not apply to single-developer repos, yet they receive the block.

**Verified absence of a config gate**: `internal/config/`, `internal/hook/branch_guard.go`, and `internal/hook/pre_tool.go` contain no `enabled` flag today — `checkBranchState` is called unconditionally at `pre_tool.go:455`.

### Problem B — over-broad patterns block read-only commands

The patterns in `internal/hook/branch_guard.go:81-88` match read-only subcommands. Observed directly this session:

- `\bgit\s+stash(\s+(push|pop|apply|drop)\b)?` matches `git stash list` and `git stash show` (both read-only). **Observed**: `git stash list` returned `BRANCH_GUARD_VIOLATION: git stash in primary checkout`.
- `\bgit\s+merge\b` matches `git merge-base` (read-only). **Observed**: `git merge-base` returned `BRANCH_GUARD_VIOLATION: git merge in primary checkout`.

### Decision (confirmed with maintainer GOOS)

Convert the guard to **DEFAULT-OFF (opt-in)**. Intent (shared-checkout multi-session protection) is preserved; only the default flips. Maintainer enables via LOCAL config (`enabled: true`); distributed users get `enabled: false` → guard does not fire. Independently of the default flip, the read-only false-positive patterns are refined out so that even when the guard is enabled, it does not block read-only inspection commands.

## §B Scope

**In Scope**:
- Add a config gate (`BranchGuardConfig{ Enabled bool }`, default `false`) below the existing exemption logic.
- Refine the `branchStatePatterns` regex set to exclude read-only forms (`git stash list`, `git stash show`, `git merge-base`).
- Wire the gate through the `preToolHandler` via the existing `ConfigProvider`.
- Document the opt-in default and pattern refinement in the rule mirror (`main-checkout-branch-guard.md`) and the template copy.
- Document maintainer intent in `CLAUDE.local.md §22`-family.
- Preserve `MOAI_BRANCH_GUARD_EXEMPT=1` and the `manager-git` agent exemption unchanged (REQ-6 backward compat).

**Out of Scope**: redesigning the exemption mechanism, removing the guard, or changing the hook wiring in `settings.json.tmpl` / `handle-pre-tool.sh.tmpl`.

### Out of Scope — Guard Removal
- The guard is NOT removed or gutted. The hook stays wired and the `checkBranchState` entrypoint stays in place. Only the default + pattern refinement change.

### Out of Scope — Exemption Logic
- The `MOAI_BRANCH_GUARD_EXEMPT=1` env var and the `AgentType == "manager-git"` identity check remain byte-identical. This SPEC does not weaken, refactor, or extend them.

### Out of Scope — Hook Wiring
- The PreToolUse matcher (`"Write|Edit|Bash"`) in `internal/template/templates/.claude/settings.json.tmpl:51-65` and the `moai hook pre-tool` dispatch in `handle-pre-tool.sh.tmpl:67` are unchanged.

## §C Requirements (GEARS)

### REQ-1 — Config Gate, Default OFF

**Where** the branch-state guard is not explicitly enabled via configuration, `checkBranchState` SHALL return the allow fall-through (no-op) without evaluating patterns, primary-checkout state, or exemption logic.

**Rationale**: the guard ships through the template to all users; the default MUST be inert. The gate is ADDITIVE below the existing exemption logic — `MOAI_BRANCH_GUARD_EXEMPT` and the `manager-git` agent exemption remain unchanged (see REQ-6).

### REQ-2 — Pattern Refinement (Exclude Read-Only Forms)

**Where** the guard is enabled, `branchStatePatterns` SHALL NOT match the read-only forms `git stash list`, `git stash show`, and `git merge-base`.

The genuinely dangerous forms — `git switch`, `git checkout <branch>`, `git branch <name>`, `git reset --hard`, `git rebase`, bare/mutating `git stash` (push/pop/apply/drop), and actual `git merge` — SHALL remain matched.

### REQ-3 — Config Wiring

**Where** the `preToolHandler` is constructed with a `ConfigProvider`, the handler SHALL read the `BranchGuardConfig.Enabled` flag and gate the `checkBranchState` call accordingly.

The config struct `BranchGuardConfig{ Enabled bool }` SHALL live near `WorkflowWorktreeConfig` (`internal/config/types.go:476`) with its default in `NewDefaultWorkflowConfig` (`internal/config/defaults.go:520`). The flag MUST default to `false`.

### REQ-4 — Maintainer Local Opt-In, NOT Template

**Where** the maintainer's local environment needs the guard active, the `enabled: true` value SHALL live in the LOCAL-only `.moai/config` family (gitignored / `CLAUDE.local.md`-documented), NEVER in the template defaults.

The template default in `internal/template/templates/.moai/config/sections/workflow.yaml` (and the Go default `NewDefaultWorkflowConfig`) SHALL remain `enabled: false` — satisfying CLAUDE.local.md §25 template-neutrality.

### REQ-5 — Rule Mirror Documentation

When this SPEC lands, the rule document `main-checkout-branch-guard.md` SHALL be updated in BOTH the source tree (`.claude/rules/moai/workflow/`) AND the template tree (`internal/template/templates/.claude/rules/moai/workflow/`) to document the opt-in default and the pattern refinement.

Mirror parity (`internal/template/rule_template_mirror_test.go`) SHALL hold for the enrolled rule files. Note: `main-checkout-branch-guard.md` is a §25-sanitized pair (NOT in the byte-parity allowlist) — the source retains SPEC/REQ tokens for traceability; the template mirror is held sanitized; doctrine parity is enforced by `TestSanitizedPairParity` and mirror cleanliness by `TestTemplateNoInternalContentLeak`.

### REQ-6 — Backward Compatibility

**Where** the guard is enabled, the existing `MOAI_BRANCH_GUARD_EXEMPT=1` env var and the `HookInput.AgentType == "manager-git"` identity check SHALL continue to bypass the guard regardless of the config flag.

The fail-open norm (uncertainty → allow) SHALL be preserved: any git-context uncertainty (non-git dir, missing git binary, rev-parse non-zero) continues to return allow + stderr advisory + audit-log append, regardless of the config flag.

## §D Evidence (Observed This Session)

Problem B was observed directly:

- `git stash list` → `BRANCH_GUARD_VIOLATION: git stash in primary checkout` — the `\bgit\s+stash(\s+(push|pop|apply|drop)\b)?` pattern at `branch_guard.go:85` matches `git stash list` because the optional group matches zero times (none of `push|pop|apply|drop` is required), so the bare `\bgit\s+stash\b` prefix matches. The stash pattern carries no `list`/`show` exclusion clause.
- `git merge-base` → `BRANCH_GUARD_VIOLATION: git merge in primary checkout` — `\bgit\s+merge\b` at `branch_guard.go:87` matches `git merge-base` because `\b` after `merge` sits between `e` and `-` (a word/non-word boundary), so the pattern matches the prefix.

Problem A was verified by grep: no `enabled` flag exists in `internal/config/`, `branch_guard.go`, or `pre_tool.go` as of this session.

## §E Constraints / Non-Goals

- **Template neutrality (CLAUDE.local.md §25)**: the template default MUST stay `enabled: false`. No `enabled: true` anywhere in `internal/template/templates/`.
- **Additive gate**: do NOT touch the exemption logic in `isExemptAgent` (`branch_guard.go:118-129`).
- **Fail-open preserved**: the deny fires ONLY on positive evidence; uncertainty never denies. The config flag does not change fail-open behavior — when disabled, the guard returns allow BEFORE reaching any uncertainty path, so fail-open is trivially preserved; when enabled, the existing fail-open path at `branch_guard.go:215-221` is unchanged.
- **Dangerous forms stay matched**: pattern refinement removes ONLY the read-only false positives. Switch / checkout / branch / reset --hard / rebase / bare+mutating stash / merge remain matched.

## §F Cross-References

- Related SPEC: **SPEC-WORKTREE-BRANCH-GUARD-001** (the completed SPEC that landed the guard) — this SPEC is a behavior-tuning follow-up, NOT a supersession.
- `internal/hook/branch_guard.go` — guard logic.
- `internal/hook/pre_tool.go:455` — `checkBranchState` call site.
- `internal/config/types.go:476` — `WorkflowWorktreeConfig` (placement neighbor).
- `internal/config/defaults.go:520` — `Workflow.Worktree` literal (placement neighbor).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — doctrine mirror (source side).
- `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md` — doctrine mirror (template side, §25-sanitized).
- `CLAUDE.local.md §22` — dev settings intent (target of maintainer-opt-in documentation).

## §G HISTORY

- **2026-07-30** v0.1.0 — initial draft (plan-phase). Author: manager-spec. Evidence (Problem B) observed directly during the session that produced this SPEC.
