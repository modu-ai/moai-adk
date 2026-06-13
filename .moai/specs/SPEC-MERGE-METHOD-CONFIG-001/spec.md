---
id: SPEC-MERGE-METHOD-CONFIG-001
title: "Configurable sync-phase PR merge method (squash/merge/rebase)"
version: "0.1.0"
status: implemented
created: 2026-06-13
updated: 2026-06-13
author: manager-spec
priority: P2
phase: "v0.2.0"
module: "internal/config, internal/template/templates/.claude"
lifecycle: spec-anchored
tags: "config, git-strategy, merge-method, dead-config, github-issue-1061"
tier: M
issue_number: 1061
---

# SPEC-MERGE-METHOD-CONFIG-001 — Configurable sync-phase PR merge method

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-13 | manager-spec | Initial plan-phase authoring. Origin: GitHub Issue #1061 (sync-phase merge method hardcoded to `--squash`, not configurable, despite an unused `merge`/`squash`/`rebase` abstraction in `internal/github`). |

## A. Context

### A.1 Problem statement

The `/moai sync` auto-merge step hardcodes `gh pr merge --squash --delete-branch` in the agent-facing template prose. The PR merge method is the one git behavior in `git-strategy.yaml` that is NOT a configuration knob, even though every adjacent behavior (workflow type, automation toggles, branch prefix, hook severity, commit style, required reviews, branch protection) is configurable.

GitHub Issue #1061 reports this as surprising on three counts:

1. The binary already ships a complete, tested `merge`/`squash`/`rebase` abstraction (`internal/github`) that has **zero production callers** — it is exercised only by tests.
2. The workflow supports `gitflow`, whose `release/*` and `hotfix/*` branches the project's own reference doc (`moai-ref-git-workflow/skill.md`) says SHOULD use a merge commit — but the unconditional squash makes that impossible to honor.
3. Squash-only is additionally locked behind a FROZEN rule (`spec-workflow.md` lifecycle table), so the FROZEN policy and the unused abstraction send contradictory signals about whether squash is an intentional invariant.

### A.2 Verified ground-truth (read against current `main` working tree)

The following findings were independently verified before authoring (not assumed from the issue):

| Finding | Verified location | Status |
|---------|-------------------|--------|
| Hardcoded `gh pr merge --squash --delete-branch` | `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md:313, :325` | Confirmed |
| Hardcoded `gh pr merge --squash --delete-branch` | `internal/template/templates/.claude/agents/moai/manager-git.md:140, :176, :204` | Confirmed |
| Squash-merge ref-doc guidance | `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md:133-135` (squash/merge/rebase rows) | Confirmed |
| `MergeMethod` constants + `PRMerge` translation | `internal/github/gh.go:44-55, :241` | Confirmed (exists) |
| `MergeOptions.Method` / `MergeResult.Method` + `PRMerger` | `internal/github/pr_merger.go:11, :32, :73-94` | Confirmed (exists, default `MergeMethodMerge` at `:134`) |
| `internal/github` production callers | grep across `internal/` | Confirmed ZERO (test-only) |
| `merge_method` / `merge_strategy` key in any config section | grep across `internal/config/` | Confirmed ABSENT |
| `git_strategy` section IS wired into loader | `internal/config/loader.go:59, :169-180` (`loadGitStrategySection`) | Confirmed wired (READ path active) |
| FROZEN lifecycle table fixes "squash" for all 3 phases | `spec-workflow.md` lifecycle table, rows for plan/feat/sync | Confirmed `[ZONE:Frozen] [HARD]` |

### A.3 Relationship to the PREPUSH dead-config line

This SPEC is structurally adjacent to the completed PREPUSH dead-config wiring line (`SPEC-PREPUSH-WIRING-001` → `-MODE-WIRING-001` → `-LOADER-WIRING-001` → `-SAVE-WIRING-001`), which wired `git_strategy.<mode>.hooks.pre_push` from dead YAML into the runtime. The key difference: the `git_strategy` section loader is **already wired** here (unlike the prepush loader gap), so this SPEC adds a NEW field to an already-wired section rather than repairing a broken loader. The consumer of this new field is the **sync agent prose** (`gh pr merge`), not Go runtime code — so the wiring is config-field-to-template-prose, not config-field-to-Go-consumer.

### A.4 In-scope module boundary

- `internal/config/` — add `MergeMethod` field to `ModeProfile`, default factory, validation, loader (READ path).
- `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` — add `merge_method: squash` key per mode.
- `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` — replace hardcoded `--squash` with config-driven method selection.
- `internal/template/templates/.claude/agents/moai/manager-git.md` — same.
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` + the SSOT mirror `.claude/rules/moai/workflow/spec-workflow.md` — FROZEN→CONFIGURABLE lifecycle-table amendment.

## B. Requirements (GEARS)

### B.1 Config schema

- **REQ-MMC-001** (Ubiquitous): The `git_strategy.<mode>` config schema shall expose a `merge_method` field accepting one of `squash`, `merge`, `rebase`.
- **REQ-MMC-002** (Ubiquitous): The `merge_method` field shall default to `squash` in every mode profile (manual, personal, team) so that current behavior is preserved when the field is absent.
- **REQ-MMC-003** (Event-driven): When the config loader reads `git-strategy.yaml` and the file omits `merge_method`, the loader shall retain the compiled default (`squash`) per the existing partial-override contract.
- **REQ-MMC-004** (Event-driven): When the config loader reads a `git-strategy.yaml` that sets `git_strategy.<mode>.merge_method`, the loader shall populate `ModeProfile.MergeMethod` with that value (file value, not compiled default).

### B.2 Validation

- **REQ-MMC-005** (Event-detected): When validation runs against a `merge_method` value outside the set `{squash, merge, rebase}`, the validator shall emit a validation error naming the offending field path (`git_strategy.<mode>.merge_method`).
- **REQ-MMC-006** (State-driven): While the `merge_method` field is empty (unset by both user and default), validation shall treat it as the compiled default `squash` and shall not emit an error (fail-safe to current behavior).

### B.3 Template / agent honoring

- **REQ-MMC-007** (Ubiquitous): The sync-phase delivery workflow (`sync/delivery.md`) shall select the `gh pr merge` method flag from the active mode's `merge_method` config value rather than hardcoding `--squash`.
- **REQ-MMC-008** (Ubiquitous): The `manager-git` agent shall select the `gh pr merge` method flag from the active mode's `merge_method` config value rather than hardcoding `--squash`.
- **REQ-MMC-009** (State-driven): While `merge_method` resolves to `squash` (the default), the rendered merge command shall be `gh pr merge --squash --delete-branch` — byte-equivalent to current behavior — so that no existing user observes a behavior change without opting in.

### B.4 FROZEN rule reconciliation

- **REQ-MMC-010** (Ubiquitous): The FROZEN lifecycle table in `spec-workflow.md` (and its SSOT mirror) shall state that the PR merge strategy column is the configured `merge_method` (default `squash`), rather than a hardcoded literal `squash`, while preserving the documented rationale that squash is the default for clean SPEC history.
- **REQ-MMC-011** (Ubiquitous): The FROZEN-rule amendment shall be acknowledged explicitly as a FROZEN-zone edit in the plan-phase artifacts and shall require GATE-2 human approval before any run-phase implementation (this requirement is a process gate, not a code behavior).

### B.5 Template neutrality

- **REQ-MMC-012** (Ubiquitous): All edits under `internal/template/templates/` shall remain language-neutral and free of internal SPEC IDs, internal dates, commit SHAs, and CLAUDE.local references per CLAUDE.local.md §15/§25 and the CI guard `template-neutrality-check.yaml`.

## C. Design Considerations and Risks

### C.1 FROZEN→CONFIGURABLE policy decision (PRIMARY RISK)

`spec-workflow.md` line 25 is `[ZONE:Frozen] [HARD]` and its lifecycle table fixes "squash" for all three SPEC phases (plan/feat/sync). Making the merge method configurable touches a FROZEN-zone document.

The proposed reconciliation is **non-destructive**: the default stays `squash`, and the FROZEN rationale (one squash commit per phase → clean, revertable SPEC history) remains the documented recommendation. The amendment changes the lifecycle table cell from the literal `squash` to "configured `merge_method` (default `squash`)". This is a widening, not a removal — the FROZEN invariant's intent (predictable default) is preserved while removing the contradiction with the gitflow ref-doc and the unused abstraction.

[HARD] This FROZEN-zone edit MUST NOT be implemented without explicit GATE-2 human approval. The plan-auditor and the orchestrator's GATE-2 gate are the control points. If the maintainer decides squash MUST remain a non-negotiable invariant, the alternative resolution is: keep the FROZEN rule as-is, scope this SPEC down to config + abstraction wiring for the gitflow `release/*` exception only, OR reject the SPEC and instead remove the misleading unused abstraction. This decision is surfaced to the orchestrator for GATE-2.

#### C.1.1 Why all 3 lifecycle rows are widened (not just the sync row)

Issue #1061 is reported against the sync phase only (`/moai sync` Step 3.4). However, the FROZEN lifecycle table fixes "squash" for all THREE phase rows (plan/feat/sync), and this SPEC widens all three. This is deliberate, for two reasons:

1. **The merge sites are generic, not sync-exclusive.** The hardcoded `gh pr merge --squash` lives in `manager-git.md` (3 sites: `:140`, `:176`, `:204`), which is the single agent that performs PR merges for ALL phases — plan PRs, run/feat PRs, and sync PRs alike. There is no phase-specific merge codepath. Wiring `merge_method` into `manager-git` necessarily affects every phase's merge, so leaving the plan/feat rows reading a hardcoded `squash` while the sync row reads "configured" would create a table that contradicts the actual (now uniform) behavior.

2. **A per-phase split would be incoherent.** The `merge_method` config field is per-MODE (manual/personal/team), not per-PHASE. There is no config surface for "squash plan PRs but merge sync PRs". Widening only the sync row would imply a per-phase override that this SPEC does not deliver (and that is explicitly excluded — see EX-2 deferral of per-branch-type overrides). Widening all three rows uniformly is the only formulation consistent with the per-mode config model.

The default stays `squash` for all three rows, so the FROZEN invariant's intent (clean per-phase SPEC history by default) is preserved uniformly. The widening removes the literal-vs-configurable contradiction across the whole table, not just the one cell the issue happened to name.

### C.2 Consumer is agent prose, not Go runtime

Unlike the PREPUSH line (where the consumer was Go hook code), the `gh pr merge` consumer is the sync agent executing a shell command from template prose. The Go-side change (config field + loader + validator) is necessary for the value to be readable, but the actual method-selection logic lives in agent instructions. This means the run-phase implementation cannot be unit-tested end-to-end at the Go level for the prose-honoring requirement (REQ-MMC-007/008/009) — those are verified by grep/structural assertions on the rendered template, not by Go behavior tests. The Go-level requirements (REQ-MMC-001..006) ARE unit-testable, mirroring `git_strategy_loader_test.go`.

### C.3 Optional PRMerger connection (deferred scope candidate)

The issue suggests routing the merge through the existing `internal/github` `PRMerger`/`PRMerge`. This is NOT in the minimal scope of this SPEC: `PRMerger` has zero callers and wiring a Go-level PR merge path would require a new CLI subcommand or hook, a much larger change with its own design surface. The minimal honest scope here is config-field-to-template-prose. Connecting `PRMerger` to a real caller is a candidate FOLLOW-UP SPEC and is listed under Exclusions. Note also that `PRMerger`'s own default is `merge` (not `squash`), so if a future SPEC wires it as the merge path, that default MUST be reconciled with the `squash` default established here.

### C.4 gitflow per-branch-type override (deferred scope candidate)

The issue's strongest case (release branches should preserve history) implies a per-branch-type override (`release/* → merge`) beyond a single per-mode `merge_method`. This SPEC delivers the per-mode field only. Per-branch-type overrides for gitflow are a candidate FOLLOW-UP and are listed under Exclusions, to keep this SPEC at Tier M scope.

### C.5 Edge cases

- Empty string `merge_method` → treated as default `squash` (REQ-MMC-006), no error.
- `merge_method: rebase` under branch-protection requiring linear history → valid config; runtime `gh pr merge --rebase` may fail at GitHub if the branch is not rebaseable. This is a GitHub-side runtime concern, NOT a config-validation concern; the validator only checks enum membership.
- Mixed: a mode sets `merge_method` but the workflow is `main_direct` (no PR) → the field is simply unused in that workflow; no error.

## D. Acceptance Criteria

See `acceptance.md` for the full Given-When-Then matrix and grep-verifiable assertions.

## E. Exclusions (What NOT to Build)

### E.1 Out of Scope

- **EX-1**: Connecting the `internal/github` `PRMerger`/`PRMerge` abstraction to a real production caller (new CLI subcommand or hook). The minimal scope is config-field-to-template-prose; the Go merge-path wiring is a candidate FOLLOW-UP SPEC. (See C.3.)
- **EX-2**: Per-branch-type merge-method overrides for gitflow (e.g., `release/* → merge`, `hotfix/* → merge`). This SPEC delivers a single per-mode `merge_method` field only. Per-branch-type overrides are a candidate FOLLOW-UP. (See C.4.)
- **EX-3**: Removing the unused `internal/github` `PRMerger`/`MergeMethod` abstraction. This SPEC wires the config to honor it conceptually but does not delete or refactor that code.
- **EX-4**: Changing the DEFAULT merge method away from `squash`. The default MUST stay `squash` to preserve current behavior; only the OPT-IN to `merge`/`rebase` is added.
- **EX-5**: Modifying the `internal/cli/worktree/sync.go` `--strategy merge|rebase` flag. That flag syncs a worktree against its base branch — a different operation from the PR merge method, explicitly out of scope per the issue's own note.
- **EX-6**: Auto-detecting the merge method from branch type at runtime in Go. Method selection is config-driven and resolved by the agent from the active mode profile.
- **EX-7**: Any deprecation or removal of the FROZEN squash rationale. The amendment widens the rule (default squash + configurable opt-in); it does not remove the documented recommendation.
