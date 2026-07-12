---
id: SPEC-DISCOVERY-UNKNOWNS-001
title: "Unknowns framework Tier-1 for Context-First Discovery — Blind Spot Pass + decision-reversibility ordering + 4-quadrant lens"
version: "0.1.0"
status: completed
created: 2026-07-05
updated: 2026-07-13
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/core"
lifecycle: spec-anchored
tier: M
tags: "discovery, unknowns, blind-spot, gears, askuser, planning, context-first, doc"
---

# SPEC-DISCOVERY-UNKNOWNS-001 — Progress Tracking

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-05
plan_author: manager-spec
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
milestones: 4
req_count: 15
ac_count: 16
```

Plan-phase artifacts (spec.md + plan.md + acceptance.md + this progress.md skeleton) authored. Applies the Tier-1 subset of the "Finding Your Unknowns" framework to Context-First Discovery via three genuinely-absent enhancements (T1 Blind Spot Pass, T2 decision-reversibility plan ordering, T3 unknowns 4-quadrant lens). All 3 problem-baseline gaps grep-verified against the current tree (spec.md §A.2). Doc/rule/agent-body level only — no Go code, no new agents/skills/subsystems. Ready for plan-auditor review.

## §E.2 Run-phase Evidence

Run-phase implemented the three Tier-1 enhancements (T1 Blind Spot Pass, T2 decision-reversibility ordering, T3 unknowns 4-quadrant lens) across 4 doctrine surfaces + their 4 template mirrors, `make build` re-embed, and the catalog-hash regen cascade. cycle_type=ddd (behavior-preserving additive doc authoring — PRESERVE existing structure, EXTEND with additive subsections/lines). All 17 ACs (AC-DU-001..017) PASS; the two enforced CI parity guards pass (audit D2 fix). Command evidence written to `/tmp/du-*.log`; grep evidence surfaced in the run report.

### AC PASS/FAIL matrix (17/17 PASS)

| AC | REQ | Status | Evidence (command → observed) |
|----|-----|--------|-------------------------------|
| AC-DU-001 | REQ-DU-001 | PASS | `grep -cE '^#{2,3} .*Blind Spot Pass' askuser-protocol.md` → 1 (named `## Blind Spot Pass` H2 present) |
| AC-DU-016 | REQ-DU-002 | PASS | `grep -A40 'Blind Spot Pass'` → both `unfamiliar` and `before plan-phase entry` matched |
| AC-DU-002 | REQ-DU-003 | PASS | `grep -A40` → `Agent(Explore)` and `read-only` both present |
| AC-DU-003 | REQ-DU-004 | PASS | `grep -A40` → `OPTIONAL` and `not a mandatory gate` present |
| AC-DU-004 | REQ-DU-005 | PASS | `grep -A40` → `AskUserQuestion` and `does not prompt the user` present |
| AC-DU-005 | REQ-DU-006 | PASS | spec-workflow.md `## Plan Phase` section + CLAUDE.md §7 Rule 5 (-A3 window) both carry `Blind Spot Pass` |
| AC-DU-006 | REQ-DU-007 | PASS | manager-spec.md carries `decision-reversibility`/`most likely to change` and `mechanical…bottom` on the plan.md guidance line |
| AC-DU-007 | REQ-DU-008 | PASS | `grep -A2 'Rule 1 — Approach-First' CLAUDE.md` then `tail -n +2` → `most likely to change` on the new Rule 1 sub-bullet (non-vacuous; heading stripped) |
| AC-DU-008 | REQ-DU-009/013 | PASS | `git log --grep=SPEC-DISCOVERY-UNKNOWNS-001 --name-only -- '*.go'` then `grep -c '.go$'` → 0 (verified post-commit) |
| AC-DU-009 | REQ-DU-010 | PASS | all four quadrant terms present in askuser-protocol.md, each count = 2 |
| AC-DU-010 | REQ-DU-011 | PASS | `grep -A6 'Unknown-Unknowns'` → routes suspected Unknown-Unknowns to `Blind Spot Pass` |
| AC-DU-011 | REQ-DU-012 | PASS | CLAUDE.md §7 Rule 5 (-A3) carries `Known-Unknowns`/`4-quadrant`; diet: 3 net lines added to §7 (Rule 1 = 1, Rule 5 = 2) |
| AC-DU-012 | REQ-DU-014 | PASS | per-file mirror token count: askuser=6, manager-spec `most likely to change`=1, CLAUDE BSP=2, spec-workflow BSP=1 — identical in local + mirror |
| AC-DU-013 | REQ-DU-014 | PASS | `make build` exit 0 (embedded template FS + catalog.yaml regenerated) |
| AC-DU-014 | REQ-DU-015 | PASS | §25 neutrality: 0 `SPEC-DISCOVERY-UNKNOWNS` in templates AND 0 `2026-07-05`/`Finding N`/`Audit N` in the 4 mirrors |
| AC-DU-015 | REQ-DU-013 | PASS | `git status --porcelain` then `grep '.go$'` → empty (no `.go` in working tree) |
| AC-DU-017 | REQ-DU-014 | PASS | `go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift\|TestSanitizedPairParity'` → exit 0 (byte-identical spec-workflow + sanitized-pair askuser) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-12
run_author: manager-develop
run_commit_sha: a041f26da
cycle_type: ddd
ac_pass_count: 17
ac_fail_count: 0
preserve_list_post_run_count: intact
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: pass
  windows: pass
enforced_parity_guards:
  TestRuleTemplateMirrorDrift: pass
  TestSanitizedPairParity: pass
template_neutrality_25: clean
total_run_phase_files: 9
m1_to_mN_commit_strategy: single run-phase commit (M1-M4 folded; doc/rule/agent-body only)
```

Run-phase complete on Route A (Hybrid Trunk main-direct, Tier M). Implementation executed in a runtime-materialized stale L1 worktree (branch `worktree-agent-aceee5b3caefd3877`), fast-forwarded to main `b3a410c8b` before editing so the 17-AC remediated criteria applied; the orchestrator integrates the run-phase commit into `main` via cherry-pick/FF and backfills the post-integration `run_commit_sha`. Ready for sync-phase.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-07-13
sync_author: manager-docs
sync_commit_sha: pending-backfill-discovery-unknowns-001
changelog_entry_position: "[Unreleased] > ### Changed (top entry)"
frontmatter_status_transitions:
  spec.md: "in-progress -> completed"
  plan.md: "in-progress -> completed"
  acceptance.md: "in-progress -> completed"
  progress.md: "in-progress -> completed"
push_state: deferred (local main-direct commit only; batch-push by orchestrator after parallel session ends)
```

Sync-phase actions performed on this commit: (1) merged `in-progress -> implemented -> completed` frontmatter status transition applied atomically across all 4 SPEC artifacts (spec.md / plan.md / acceptance.md / progress.md), with `updated:` refreshed to 2026-07-13 in each; (2) one `### Changed` entry appended under `CHANGELOG.md` `[Unreleased]`, describing the actually-shipped T1 Blind Spot Pass + T2 decision-reversibility ordering + T3 4-quadrant lens behavior verified by reading the 4 changed rule/agent-body surfaces (`askuser-protocol.md`, `spec-workflow.md`, `manager-spec.md`, `CLAUDE.md`), pre-emission duplicate-check `grep -c 'SPEC-DISCOVERY-UNKNOWNS' CHANGELOG.md` → 0 confirmed before append; (3) README/README.ko.md assessed — no user-facing CLI surface or feature-list change, skipped per B12 discipline; (4) `sync_commit_sha` recorded as a `pending-backfill-*` placeholder per the SHA placeholder backfill exemption (spec-frontmatter-schema.md) — this sync commit cannot know its own SHA, real SHA to be backfilled in a follow-up commit by the orchestrator. Push is deferred: a parallel session holds the shared checkout; the orchestrator will batch-push after that session ends. This SPEC is doc/rule/agent-body only (zero Go source touched by the sync commit), consistent with the run-phase scope (REQ-DU-013).
