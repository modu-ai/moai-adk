---
id: SPEC-V3R5-CORE-SLIM-B-001
title: "Category B Dead-Weight Skill Retire (Phase 0+1)"
version: "0.2.0"
status: implemented
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.0.0 — Round 5"
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
| 2026-05-20 | 0.1.1 | Iter 2 revision (5 BLOCKING + 3 SHOULD resolved). F-001: dropped phantom REQ-CSB-006/AC-CSB-006 (expert-mobile already absent — verified grep count 0). F-002: REQ-CSB-005 scope corrected to 2 language rules (elixir + csharp) plus template-tree mirror (4 files total), not 5. F-003: "typo" framing removed — `moai-platform-deployment` is the full canonical name. F-004: template-tree paths hardcoded as mandatory tasks (verified the template tree DOES mirror language rules). F-005: new REQ-CSB-006 + AC-CSB-007 added for `agents-reference.md` dead-ref cleanup (local + template mirror). S1: AC-CSB-005 extended to cover template tree. S2: R-CSB-004 [LOW] documentation residual risk added. S3: `.claple` typo in AC-CSB-002 fixed verbatim. REQ count 7→7 (renumbered contiguous), AC count 8→8 (renumbered contiguous). |
| 2026-05-20 | 0.2.0 | Sync-phase: status `draft → implemented`. Run-phase commits on main: `12a66b514` (M1 — 4 Category B skills retire, 1,432 LOC delete) + `07b709a3e` (M2 — moai-platform-deployment cross-refs removed in 2 language rules + template mirror) + `9d4ab401a` (M3 — embedded.go regeneration + catalog.yaml cleanup) + `e044dbcc3` (iter2 revise — SPEC artifact + agents-reference.md cleanup). 8/8 ACs binary PASS. Cross-platform Windows build PASS. No NEW lint regressions. LEAN dogfooding 2nd cycle: ~6min run-phase wall-time, Tier S minimal delegation prompt successful (Section A-E template OPTIONAL applied). Phase 0+5 dead-reference cleanup pre-processed by `f499746d3` (Vision §3.5 Determinism removed + meta-harness expert-mobile + 2 language rule platform-deploy typo fix). Total: 6 commits on main since `c0eb30da6` (WORKFLOW-LEAN-001 머지). |

---

## 1. Identity

- **SPEC ID**: SPEC-V3R5-CORE-SLIM-B-001
- **Title**: Category B Dead-Weight Skill Retire (Phase 0 + Phase 1)
- **Status**: implemented
- **Version**: 0.2.0
- **Author**: GOOS Kim
- **Tier**: S (LEAN — 2 artifacts: spec.md + plan.md)
- **Predecessor**: `.moai/research/core-slimming-audit-2026-05-20.md` (Core Slimming Audit, 2026-05-20)
- **Decomposition basis**: Audit §5 권장안 A row 1 (4-SPEC LEAN split). This SPEC covers Phase 0 (dead reference cleanup) + Phase 1 (Category B dead-weight retire).

### 1.1 Background

The Core Slimming audit measured 10 migration candidates (3 Category A domain skills + 4 Category B platform/framework skills + 2 Category C expert agents + 1 dead reference). Category B comprises 4 skills (`moai-framework-electron`, `moai-platform-auth`, `moai-platform-chrome-extension`, `moai-platform-deployment`) totalling **1,432 LOC** with **0 workflow/agent invocations** measured across the codebase — classified as TRIVIAL-risk dead weight.

Additionally, 2 language rules (`elixir.md`, `csharp.md`) and their template-tree mirrors carry cross-references to the `moai-platform-deployment` skill being retired by REQ-CSB-004; these references must be removed in the same pass. Furthermore, 3 cross-reference lines in `.claude/skills/moai-foundation-core/modules/agents-reference.md` (and its template-tree mirror) reference the retired skills and become orphaned after M1; these must also be cleaned up.

Note: The predecessor audit doc framed the language-rule references as a `moai-platform-deploy` "typo". That framing is corrected here: the actual canonical name is `moai-platform-deployment` (full name), and the references are removed entirely (not renamed) because the underlying skill itself is retired by REQ-CSB-004.

Note: An earlier orchestrator-prepared draft contemplated cleanup of an `expert-mobile` line at `moai-meta-harness/SKILL.md` line 203. Grep verification at plan-revision time confirmed `expert-mobile` already produces **0 matches** across both the local file and its template-tree mirror — the cleanup has already happened (likely as W0 SPEC-V3R5-CLAUDE-REFRESH-001 collateral on the expert-mobile hard-delete). Consequently, this SPEC does NOT include any meta-harness edit and any prior phantom requirement/AC targeting it has been dropped in iter 2.

### 1.2 Goal

Retire 4 dead-weight skills + cleanup orphan cross-references in language rules (2 files + template mirrors = 4 paths) and `agents-reference.md` (1 file + template mirror = 2 paths, 3 reference lines each) in a single Tier S SPEC, validated by binary AC commands and cross-platform build pass.

---

## 2. EARS Requirements

- **REQ-CSB-001**: The system SHALL remove the `moai-framework-electron` skill directory from both `internal/template/templates/.claude/skills/` (template source) and `.claude/skills/` (local mirror).
- **REQ-CSB-002**: The system SHALL remove the `moai-platform-auth` skill directory from both template source and local mirror.
- **REQ-CSB-003**: The system SHALL remove the `moai-platform-chrome-extension` skill directory from both template source and local mirror.
- **REQ-CSB-004**: The system SHALL remove the `moai-platform-deployment` skill directory from both template source and local mirror.
- **REQ-CSB-005**: The system SHALL remove all `moai-platform-deployment` cross-references from 2 language rules (`.claude/rules/moai/languages/elixir.md`, `.claude/rules/moai/languages/csharp.md`) AND their template-tree mirrors (`internal/template/templates/.claude/rules/moai/languages/elixir.md`, `internal/template/templates/.claude/rules/moai/languages/csharp.md`). The references are removed entirely (not renamed) because the underlying skill is retired by REQ-CSB-004. Verified at plan-time: only elixir and csharp carry the reference; kotlin, swift, and flutter have zero matches (full-tree grep at iter 2).
- **REQ-CSB-006**: The system SHALL remove dead skill references in `.claude/skills/moai-foundation-core/modules/agents-reference.md` AND its template-tree mirror (`internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md`), specifically: (a) remove `moai-platform-auth` and `moai-platform-deploy` from line 269 (leave `moai-platform-database` intact — out of this SPEC's scope); (b) remove the entire `moai-platform-auth` row at line 288; (c) remove the entire `moai-platform-deploy` row at line 290. The Category B audit scope covers exactly these 4 skills: `moai-framework-electron`, `moai-platform-auth`, `moai-platform-chrome-extension`, `moai-platform-deployment`. `moai-platform-database` is NOT in scope of this SPEC and MUST be preserved in line 269.
- **REQ-CSB-007**: WHEN `make build` is invoked after the deletions in REQ-CSB-001..004, the regenerated `internal/template/embedded.go` SHALL NOT contain any file path or file content from the 4 retired skill directories.

---

## 3. Acceptance Criteria

Each AC is binary PASS/FAIL via a single shell command. Evaluator runs each command verbatim from the project root `/Users/goos/MoAI/moai-adk-go`.

| AC | REQ | Binary Command | PASS condition |
|----|-----|----------------|----------------|
| AC-CSB-001 | REQ-CSB-001 | `test ! -e internal/template/templates/.claude/skills/moai-framework-electron && test ! -e .claude/skills/moai-framework-electron` | exit 0 (both directories absent) |
| AC-CSB-002 | REQ-CSB-002 | `test ! -e internal/template/templates/.claude/skills/moai-platform-auth && test ! -e .claude/skills/moai-platform-auth` | exit 0 (both directories absent) |
| AC-CSB-003 | REQ-CSB-003 | `test ! -e internal/template/templates/.claude/skills/moai-platform-chrome-extension && test ! -e .claude/skills/moai-platform-chrome-extension` | exit 0 (both directories absent) |
| AC-CSB-004 | REQ-CSB-004 | `test ! -e internal/template/templates/.claude/skills/moai-platform-deployment && test ! -e .claude/skills/moai-platform-deployment` | exit 0 (both directories absent) |
| AC-CSB-005 | REQ-CSB-005 | `[ $(grep -rcE "moai-platform-deploy" .claude/rules/moai/languages/ internal/template/templates/.claude/rules/moai/languages/ \| awk -F: '{sum+=$2} END {print sum+0}') -eq 0 ] && echo PASS` | output `PASS` (combined count across local + template tree is 0) |
| AC-CSB-006 | REQ-CSB-007 | `make build && go test ./...` | exit 0 (build succeeds + full test suite passes, no regression on retired-skill-related tests) |
| AC-CSB-007 | REQ-CSB-006 | `[ $(grep -cE "moai-platform-auth\|moai-framework-electron\|moai-platform-chrome-extension\|moai-platform-deploy" .claude/skills/moai-foundation-core/modules/agents-reference.md internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md \| awk -F: '{sum+=$2} END {print sum+0}') -eq 0 ] && echo PASS` | output `PASS` (combined grep count for the 4 retired-skill names across both paths is 0) |
| AC-CSB-008 | REQ-CSB-001..004 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (cross-platform build PASS, known issue B1) |

**Note on AC-CSB-007 scope**: The grep pattern explicitly excludes `moai-platform-database`, which line 269 of `agents-reference.md` also references. `moai-platform-database` is OUT of this SPEC's scope (audit Category B covers only 4 skills) and is preserved untouched by REQ-CSB-006.

**Note on AC-CSB-005 scope extension**: AC-CSB-005 now covers BOTH local rules (`.claude/rules/moai/languages/`) AND their template-tree mirrors (`internal/template/templates/.claude/rules/moai/languages/`). Verified at plan-time: 4 matches total exist pre-implementation (elixir local + csharp local + elixir template + csharp template). All 4 must become 0 to PASS.

**Provenance note**: Iter 2 fixes a `.claple` typo in the iter 1 AC-CSB-002 cell (corrected to `.claude`) and renumbers ACs after dropping a phantom expert-mobile AC. ACs are paste-ready.

---

## 4. Risks

- **R-CSB-001** [MEDIUM]: User projects that explicitly invoke `Skill("moai-platform-auth")`, `Skill("moai-framework-electron")`, `Skill("moai-platform-chrome-extension")`, or `Skill("moai-platform-deployment")` will fail after this SPEC merges. **Mitigation**: Release notes entry + deprecation note in CHANGELOG.md scheduled for v3.5.0 release tag (deferred — separate doc commit at release time per C-CSB-003). Audit §3.2 §6 Regression 위험 row 1 explicitly notes this.
- **R-CSB-002** [LOW]: `internal/template/embedded.go` is auto-generated by Go's `//go:embed` directive against `internal/template/templates/`. After directory deletion, `make build` must succeed cleanly and the regenerated `embedded.go` must contain no stale paths. **Mitigation**: AC-CSB-006 (`make build && go test ./...`) and AC-CSB-008 (`GOOS=windows GOARCH=amd64 go build ./...`) both gate this. If build fails, the deletion is reverted.
- **R-CSB-003** [LOW]: Template-local mirror divergence — template-tree change without matching local-tree change (or vice versa) would corrupt `moai update` flow on subsequent project init. **Mitigation**: enforced by REQ-CSB-001..006 (each requirement mandates BOTH paths) and plan.md M1/M2/M3 task tables enumerate both edits per file.
- **R-CSB-004** [LOW]: After M1 deletions, additional documentation cross-references may exist beyond the 4 files explicitly covered by REQ-CSB-005/006 (e.g., `README.md`, `CHANGELOG.md`, blog posts under `.moai/docs/`, or other skill bundles that may reference the retired skill names in their bodies). **Mitigation**: project-wide grep audit during sync-phase (`grep -rn "moai-platform-auth\|moai-framework-electron\|moai-platform-chrome-extension\|moai-platform-deployment" .` excluding `.git/`, `.moai/research/`, `.moai/specs/`); residual hits are documented but not blocked — they are deferred to v3.5.0 release notes (C-CSB-003) for the user-facing migration guide. No AC is added for this risk because the residual surface is unbounded and documenting it is a release-time concern, not an implementation gate.

---

## 5. Constraints

- **C-CSB-001** [Template-First mirror]: Every change in `internal/template/templates/.claude/skills/`, `internal/template/templates/.claude/rules/`, etc. MUST have a matching change in `.claude/skills/`, `.claude/rules/`, etc. (and vice versa). Per CLAUDE.local.md §2 [HARD] Template-First Rule. Plan-time verification confirmed the template tree DOES mirror language rules (`internal/template/templates/.claude/rules/moai/languages/`) and the `moai-foundation-core/modules/agents-reference.md` file. Verified by AC-CSB-006 (`make build` exit 0 implies template tree is internally consistent and regenerable) and reinforced by AC-CSB-005 + AC-CSB-007, which both grep across local + template paths.
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
