---
id: SPEC-V3R5-CORE-SLIM-B-001
title: "Category B Dead-Weight Skill Retire (Phase 0+1)"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: "internal/template/templates/.claude/skills"
lifecycle: spec-anchored
tags: "core-slim, skill-retire, dead-weight, template, w4"
tier: S
---

# SPEC-V3R5-CORE-SLIM-B-001 — Category B Dead-Weight Skill Retire

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-05-20 | 0.1.0 | Initial draft. Tier S LEAN dogfooding 2nd cycle. Predecessor audit: `.moai/research/core-slimming-audit-2026-05-20.md` §3.2 Category B + §4 Phase 0+1 + §5 권장안 A row 1. |

---

## 1. Identity

- **SPEC ID**: SPEC-V3R5-CORE-SLIM-B-001
- **Title**: Category B Dead-Weight Skill Retire (Phase 0 + Phase 1)
- **Status**: draft
- **Version**: 0.1.0
- **Author**: GOOS Kim
- **Tier**: S (LEAN — 2 artifacts: spec.md + plan.md)
- **Predecessor**: `.moai/research/core-slimming-audit-2026-05-20.md` (Core Slimming Audit, 2026-05-20)
- **Decomposition basis**: Audit §5 권장안 A row 1 (4-SPEC LEAN split). This SPEC covers Phase 0 (dead reference cleanup) + Phase 1 (Category B dead-weight retire).

### 1.1 Background

The Core Slimming audit measured 10 migration candidates (3 Category A domain skills + 4 Category B platform/framework skills + 2 Category C expert agents + 1 dead reference). Category B comprises 4 skills (`moai-framework-electron`, `moai-platform-auth`, `moai-platform-chrome-extension`, `moai-platform-deployment`) totalling **1,432 LOC** with **0 workflow/agent invocations** measured across the codebase — classified as TRIVIAL-risk dead weight. A separate dead reference (`expert-mobile` line at `moai-meta-harness/SKILL.md:203`, left over from W0 expert-mobile hard-delete) and 5 language rules carrying the typo `moai-platform-deploy` (referencing the to-be-retired `moai-platform-deployment` skill) are also cleaned up in this SPEC.

### 1.2 Goal

Retire 4 dead-weight skills + cleanup 6 cross-references in a single Tier S SPEC, validated by binary AC commands and cross-platform build pass.

---

## 2. EARS Requirements

- **REQ-CSB-001**: The system SHALL remove the `moai-framework-electron` skill directory from both `internal/template/templates/.claude/skills/` (template source) and `.claude/skills/` (local mirror).
- **REQ-CSB-002**: The system SHALL remove the `moai-platform-auth` skill directory from both template source and local mirror.
- **REQ-CSB-003**: The system SHALL remove the `moai-platform-chrome-extension` skill directory from both template source and local mirror.
- **REQ-CSB-004**: The system SHALL remove the `moai-platform-deployment` skill directory from both template source and local mirror.
- **REQ-CSB-005**: The system SHALL remove all `moai-platform-deploy` cross-references from 5 language rules (`.claude/rules/moai/languages/{elixir,csharp,kotlin,swift,flutter}.md`). The references are removed entirely (not renamed to `moai-platform-deployment`) because the underlying skill is also retired by REQ-CSB-004.
- **REQ-CSB-006**: The system SHALL remove the dead `expert-mobile` reference at `.claude/skills/moai-meta-harness/SKILL.md` line 203 (`- \`expert-mobile\` — Mobile domain harness templates`), which was orphaned when W0 (SPEC-V3R5-CLAUDE-REFRESH-001) hard-deleted the `expert-mobile` agent.
- **REQ-CSB-007**: WHEN `make build` is invoked after the deletions in REQ-CSB-001..004, the regenerated `internal/template/embedded.go` SHALL NOT contain any file path or file content from the 4 retired skill directories.

---

## 3. Acceptance Criteria

Each AC is binary PASS/FAIL via a single shell command. Evaluator runs each command verbatim from the project root `/Users/goos/MoAI/moai-adk-go`.

| AC | REQ | Binary Command | PASS condition |
|----|-----|----------------|----------------|
| AC-CSB-001 | REQ-CSB-001 | `test ! -e internal/template/templates/.claude/skills/moai-framework-electron && test ! -e .claude/skills/moai-framework-electron` | exit 0 (both directories absent) |
| AC-CSB-002 | REQ-CSB-002 | `test ! -e internal/template/templates/.claude/skills/moai-platform-auth && test ! -e .claple/skills/moai-platform-auth` (corrected: `.claude/skills/moai-platform-auth`) | exit 0 (both directories absent) |
| AC-CSB-003 | REQ-CSB-003 | `test ! -e internal/template/templates/.claude/skills/moai-platform-chrome-extension && test ! -e .claude/skills/moai-platform-chrome-extension` | exit 0 (both directories absent) |
| AC-CSB-004 | REQ-CSB-004 | `test ! -e internal/template/templates/.claude/skills/moai-platform-deployment && test ! -e .claude/skills/moai-platform-deployment` | exit 0 (both directories absent) |
| AC-CSB-005 | REQ-CSB-005 | `grep -rn "moai-platform-deploy" .claude/rules/moai/languages/ \| wc -l \| tr -d ' '` | output `0` |
| AC-CSB-006 | REQ-CSB-006 | `grep -c "expert-mobile" .claude/skills/moai-meta-harness/SKILL.md` | output `0` |
| AC-CSB-007 | REQ-CSB-007 | `make build && go test ./...` | exit 0 (build succeeds + full test suite passes, no regression on retired-skill-related tests) |
| AC-CSB-008 | REQ-CSB-001..004 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (cross-platform build PASS, known issue B1) |

**Note on AC-CSB-002**: The orchestrator-prepared draft contained a typo (`.claple`). The canonical command uses `.claude` — corrected here. AC commands are paste-ready.

---

## 4. Risks

- **R-CSB-001** [MEDIUM]: User projects that explicitly invoke `Skill("moai-platform-auth")`, `Skill("moai-framework-electron")`, `Skill("moai-platform-chrome-extension")`, or `Skill("moai-platform-deployment")` will fail after this SPEC merges. **Mitigation**: Release notes entry + deprecation note in CHANGELOG.md scheduled for v3.5.0 release tag (deferred — separate doc commit at release time per C-CSB-003). Audit §3.2 §6 Regression 위험 row 1 explicitly notes this.
- **R-CSB-002** [LOW]: `internal/template/embedded.go` is auto-generated by Go's `//go:embed` directive against `internal/template/templates/`. After directory deletion, `make build` must succeed cleanly and the regenerated `embedded.go` must contain no stale paths. **Mitigation**: AC-CSB-007 (`make build && go test ./...`) and AC-CSB-008 (`GOOS=windows GOARCH=amd64 go build ./...`) both gate this. If build fails, the deletion is reverted.
- **R-CSB-003** [LOW]: Template-local mirror divergence — template-tree change without matching local-tree change (or vice versa) would corrupt `moai update` flow on subsequent project init. **Mitigation**: enforced by REQ-CSB-001..004 (each requirement mandates BOTH paths) and plan.md M1 task list enumerates both deletes per skill.

---

## 5. Constraints

- **C-CSB-001** [Template-First mirror]: Every change in `internal/template/templates/.claude/skills/` MUST have a matching change in `.claude/skills/` (and vice versa). Per CLAUDE.local.md §2 [HARD] Template-First Rule. Verified by AC-CSB-007 (`make build` exit 0 implies template tree is internally consistent and regenerable).
- **C-CSB-002** [LATE-BRANCH commit]: This SPEC's plan-phase commit goes directly to `main`. No feature branch is created during plan-phase. Late-branch is created at PR time per SPEC-V3R5-LATE-BRANCH-001 4-phase procedure (Phase A: SPEC main commit → Phase B: implementation main commits → Phase C: `git switch -c` + push + squash → Phase D: local main reset).
- **C-CSB-003** [No release-notes coupling]: This SPEC does NOT include CHANGELOG.md or release-notes edits. Those are deferred to v3.5.0 release tag time. R-CSB-001 mitigation is acknowledged via release notes but not delivered by this SPEC.

---

## 6. Out of Scope

### 6.1 Out of Scope

This SPEC is intentionally narrow. The following items are explicitly NOT in scope and will be addressed by separate SPECs:

- **Category A domain skill retire** (`moai-domain-backend`, `moai-domain-frontend`, `moai-domain-database`) — separate **SPEC-V3R5-CORE-SLIM-A-001** (Tier M, Phase 2 per audit §4). Category A requires baseline extraction into `moai-meta-harness/templates/` plus 9 language-rule cross-reference updates, which is fundamentally a different shape of work than Category B's pure delete.
- **Category C expert agent retire** (`expert-backend`, `expert-frontend`) — separate **SPEC-V3R5-CORE-SLIM-C-001** (Tier M-L, Phase 3 per audit §4). Category C requires 24 workflow invocation replacements + 9 agent cross-reference cleanups + baseline extraction prerequisite. Pre-requisite: 1+ web/frontend project E2E meta-harness validation (currently 0% per audit §1.2).
- **Seed library bootstrap** (8 baseline seeds — Go-cli, Go-web, React, Vue, etc.) — separate **SPEC-V3R5-PROJECT-MEGA-001** (Tier S, Phase 4 per audit §4). Additive only, independent of retire work.
- **`/moai project --refresh` subcommand** — part of SPEC-V3R5-PROJECT-MEGA-001. Adds new CLI verb, no dependency on this SPEC.
- **Determinism guarantee** (Seed=SHA256 + temperature=0 + bit-exact CI test + semantic ≥0.95) — REMOVED from W4 scope per user directive 2026-05-20 (`feedback_w4_no_determinism` memory). Vision §3.5 will be deleted in Phase 5 documentation refresh, separate from this SPEC.

---

## 7. References

- **Predecessor audit**: `.moai/research/core-slimming-audit-2026-05-20.md` — Core Slimming Audit (2026-05-20). This SPEC implements §4 Phase 0 + Phase 1 and §5 권장안 A row 1.
- **W4 vision (will be revised in Phase 5)**: `.moai/research/harness-autonomy-vision-2026-05-18.md` — §5 W4 scope. §3.5 Determinism section will be removed per `feedback_w4_no_determinism`.
- **LEAN workflow precedent**: SPEC-V3R5-WORKFLOW-LEAN-001 (PR #1030, merged 2026-05-20). plan-auditor 0.92 PASS in 1 iter for Tier S, 2 artifacts, 7 commits. This SPEC is the 2nd LEAN dogfooding cycle.
- **Late-branch workflow**: SPEC-V3R5-LATE-BRANCH-001 (PR #1029, merged 2026-05-20). Default `current` path for new SPECs per user policy 2026-05-17.
- **No GitHub Issue policy**: `feedback_no_github_issue_for_specs` memory. `issue_number` field omitted from frontmatter per REQ-LB-009 + EXCL-LB-008.
- **Frontmatter schema SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12 fields + optional `tier:`. Backward compat: absence of `tier:` = Tier L (this SPEC opts in to Tier S explicitly).
- **Template-First rule**: CLAUDE.local.md §2 [HARD] Template-First Rule. Every new/deleted file under `.claude/skills/` MUST have matching change under `internal/template/templates/.claude/skills/`.
- **Embedded template system**: CLAUDE.local.md §2 Embedded Template System. `internal/template/embedded.go` is auto-generated by `make build` from `internal/template/templates/`. Never edit directly.
- **Spec-lint h3 Exclusions pattern**: `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section B Known Issues B6. `## Out of Scope` (h2) alone produces `MissingExclusions` ERROR — at least one `### X.Y Out of Scope` (h3) sub-section is required. This spec.md complies via §6.1.
