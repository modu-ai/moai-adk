---
id: SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001
title: "plan-auditor GEARS-aware rubric 정렬"
version: "0.2.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents/meta, internal/template/templates"
lifecycle: spec-anchored
tags: "gears, ears, plan-auditor, rubric, alignment, sprint-10, v3.0.0"
tier: S
issue_number: null
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001]
related_specs: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001]
---

# SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 — plan-auditor GEARS-aware rubric 정렬

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.2.0 | 2026-05-25 | manager-docs | Sync-phase close (plan `906f9285e` + run `6366d7428` + sync `7cd25e386`): 4-artifact frontmatter `status: in-progress → implemented` + HISTORY v0.2.0 entry + CHANGELOG entry. M1-M3 plan.md execution VERIFIED: 8/8 AC PASS (AC-PAGA-001..008). Plan-auditor iter-1 0.913 PASS skip-eligible. All 2 artifacts (`.claude/agents/meta/plan-auditor.md` + template mirror) byte-identical post-sync. Self-dogfood lint confirms 0 LegacyEARSKeyword + 0 FrontmatterInvalid on this SPEC's own GEARS-notation REQs (REQ-PAGA-001..009). |
| 0.1.0 | 2026-05-25 | manager-spec | Initial draft — Sprint 10 GEARS sweep cohort P1 entry (1st of 7 follow-up SPECs after `SPEC-V3R6-GEARS-MIGRATION-001` v0.2.0 PR #1046 + `SPEC-V3R6-SKILL-GEARS-ALIGN-001` v0.2.0 just-merged `7b8f939b7`+`263db0600`). Baseline: `.claude/agents/meta/plan-auditor.md` 463 lines / 14 EARS refs / 0 GEARS refs; M3 Score 1.0 anchor (lines 58-64) uses 5 EARS patterns with Unwanted in legacy `If [undesired condition], then [action]` form; MP-2 label (line 126) reads "EARS Format Compliance"; M2 failure modes list (lines 41-47) lacks IF/THEN deprecation marker entry. Template mirror at `internal/template/templates/.claude/agents/meta/plan-auditor.md` (26469B) byte-identical. Tier S minimal (~150 LOC markdown across 2 files / 2 artifacts spec.md + plan.md inline AC, NO acceptance.md per Tier S inline pattern). Self-dogfooding: this SPEC's REQs use GEARS notation as canonical exemplar; the transitional asymmetry (plan-auditor judges this SPEC against current EARS-only rubric before the update lands) is acknowledged in plan.md M0 — Opus 4.7 natural-language mapping of GEARS `When` / `While` / `Where` / `shall not` to EARS Event-driven / State-driven / Optional / Unwanted patterns provides PASS-equivalence (already empirically proven on `SPEC-V3R6-SKILL-GEARS-ALIGN-001` plan-auditor 0.892 PASS run). Predecessor SPECs both `status: implemented`; this SPEC enters as `status: draft`. |

## 1. Goal

`.claude/agents/meta/plan-auditor.md` 에이전트 본문을 GEARS (Generalized EARS) 표기와 일관되게 정렬한다. 구체적으로 (a) MP-2 must-pass criterion 라벨을 `EARS/GEARS Format Compliance` 로 갱신하고, (b) M3 Score 1.0 5-pattern rubric anchor 를 GEARS canonical form 으로 갱신 (Ubiquitous / `When` Event-driven / `While` State-driven / `Where` capability-gate / Unwanted = `<subject> shall not`) 하며, (c) M2 Adversarial Stance failure-modes list 에 "IF/THEN syntax without deprecation marker (post-6-month window)" 후보를 추가하고, (d) docs-site `moai-plan.md#gears-notation` 4-locale 마이그레이션 가이드로의 cross-reference 를 본문에 명시한다. 88개 기존 EARS SPEC 의 6-month backward-compatibility window 와 GEARS generalized `<subject>` 치환 정책을 명문화하여, plan-auditor 가 신규 GEARS SPEC 과 legacy EARS SPEC 양쪽 모두에 대해 일관된 판정을 수행하도록 한다.

### 1.1 Why Now

- 선행 SPEC 2개 (`SPEC-V3R6-GEARS-MIGRATION-001` PR #1046 + `SPEC-V3R6-SKILL-GEARS-ALIGN-001` v0.2.0) 가 lint engine + 4-locale docs-site + 5개 authoring guide + 5개 template mirror 를 GEARS-first 로 정렬 완료한 상태에서, plan-auditor.md 만 EARS-only 표기로 잔존하면 plan-phase auditor 가 GEARS REQ 작성자에게 false-negative 판정을 내릴 수 있는 정책 일관성 갭이 남는다. 본 SPEC 은 그 갭을 닫는다.
- GEARS sweep cohort 진행 중 (전체 88 SPEC 대상 6-month window) — plan-auditor 가 cohort 종료 전에 GEARS-aware 판정 능력을 갖춰야 향후 GEARS SPEC 작성자 onboarding 비용이 최소화된다.

### 1.2 Non-goals

본 SPEC 의 작업 범위에 **포함되지 않는** 항목:

#### 1.2.1 Out of Scope

- `internal/spec/lint.go` 의 `LegacyEARSKeyword` rule 수정 (`SPEC-V3R6-GEARS-MIGRATION-001` M2 에서 완료됨)
- 4-locale docs-site `moai-plan.md` 본문 수정 (`SPEC-V3R6-GEARS-MIGRATION-001` M3 에서 완료됨)
- 5개 authoring guide skill 본문 수정 (`SPEC-V3R6-SKILL-GEARS-ALIGN-001` M1-M5 에서 완료됨)
- 88개 기존 EARS SPEC 본문의 GEARS retroactive 재작성 (provisional `SPEC-V3R6-GEARS-SWEEP-001` 으로 deferred)
- 다른 meta agent (`.claude/agents/meta/evaluator-active.md`, `.claude/agents/meta/claude-code-guide.md`) 의 EARS/GEARS 참조 정렬 (별도 scope, Sprint 10 cohort 의 P2-P7 SPEC 에서 처리 검토)
- `.claude/agents/core/manager-spec.md` 본문 수정 (SKILL-GEARS-ALIGN-001 v0.2.0 M5 에서 완료됨)

## 2. Requirements

본 SPEC 의 REQ 는 GEARS 표기로 작성되며, 자기 자신을 self-dogfood 케이스로 사용한다 (REQ-PAGA-009 가 lint engine 으로 자체 검증).

### REQ-PAGA-001 (Ubiquitous, generalized subject)

The agent body of `plan-auditor` **shall** present "EARS/GEARS Format Compliance" as the canonical label of the MP-2 must-pass criterion (currently "EARS Format Compliance" at line 126).

### REQ-PAGA-002 (Event-driven)

**When** the M3 rubric of `plan-auditor` anchors Score 1.0 for EARS/GEARS Format Compliance, the agent **shall** enumerate the GEARS Five Patterns as the canonical reference: (1) Ubiquitous `The <subject> shall <behavior>`, (2) `When` Event-driven, (3) `While` State-driven, (4) `Where` capability-gate (representing capability gate / feature flag / static config; no longer "Optional"), and (5) Unwanted in negative form `<subject> shall not <action>`.

### REQ-PAGA-003 (Event-driven)

**When** the M3 Score 1.0 anchor lists the Unwanted pattern, the agent **shall** present `<subject> shall not <action>` as the GEARS canonical form, retaining the legacy form `If [undesired condition], then [action]` only with an inline `[DEPRECATED — use shall not]` annotation per the 6-month backward-compatibility window.

### REQ-PAGA-004 (Where, capability-gate)

**Where** a SPEC author writes a compound clause using the unified GEARS form `[Where ...][While ...][When ...] The <subject> shall <behavior>`, the agent **shall** recognize this as PASS-equivalent to the corresponding EARS pattern (Event-driven / State-driven / Optional) at Score 1.0.

### REQ-PAGA-005 (Event-driven)

**When** the agent loads its M2 Adversarial Stance failure-modes list, the agent **shall** include "ACs use IF/THEN syntax without deprecation marker (post-6-month backward-compatibility window)" as a candidate failure mode for SPECs created after the cohort closure date.

### REQ-PAGA-006 (Ubiquitous)

The agent body **shall** include a cross-reference link to `docs-site/content/en/workflow-commands/moai-plan.md#gears-notation` (or the 4-locale canonical equivalent) as the GEARS migration guide.

### REQ-PAGA-007 (Ubiquitous, generalized subject)

The agent body **shall** acknowledge that GEARS generalizes the EARS hardcoded "the system" subject to `<subject>` (any noun: system, component, service, agent, function, artifact), while preserving backward compatibility for the 88 existing EARS SPECs which keep "The system" as the default subject for readability.

### REQ-PAGA-008 (Ubiquitous, mirror parity)

The local agent body `.claude/agents/meta/plan-auditor.md` and the template mirror `internal/template/templates/.claude/agents/meta/plan-auditor.md` **shall** be byte-identical per the Template-First Rule.

### REQ-PAGA-009 (Event-driven, self-dogfood)

**When** the spec linter executes `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md`, the linter **shall** emit zero `LegacyEARSKeyword` findings and zero `FrontmatterInvalid` findings on this SPEC's own REQs and acceptance criteria.

## 3. Acceptance Criteria (inline per Tier S minimal pattern)

본 Tier S SPEC 은 별도 `acceptance.md` 파일 없이 spec.md §3 에 binary AC 를 inline 으로 명시한다 (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S 권장 패턴).

### AC-PAGA-001 (REQ-PAGA-001 + REQ-PAGA-002)

`grep -c "GEARS" .claude/agents/meta/plan-auditor.md` ≥ 5. Verifies the agent body explicitly references GEARS in at least 5 distinct locations (MP-2 label + M3 5-pattern anchor + cross-reference + generalized subject paragraph + Unwanted form note minimum).

### AC-PAGA-002 (REQ-PAGA-001)

`grep -c "EARS/GEARS\|EARS+GEARS" .claude/agents/meta/plan-auditor.md` ≥ 1. Verifies the MP-2 label uses the combined "EARS/GEARS" form, signaling that both notations are first-class for plan-auditor judgment.

### AC-PAGA-003 (REQ-PAGA-003)

`grep -cE "shall not|negative form" .claude/agents/meta/plan-auditor.md` ≥ 1. Verifies the Unwanted pattern is presented in GEARS canonical `shall not` negative form.

### AC-PAGA-004 (REQ-PAGA-005)

`grep -cE "IF/THEN.*deprecation|IF/THEN.*backward|backward-compatibility window" .claude/agents/meta/plan-auditor.md` ≥ 1. Verifies the M2 failure-modes list includes the IF/THEN deprecation marker check.

### AC-PAGA-005 (REQ-PAGA-006)

`grep -cE "#gears-notation|moai-plan.*gears" .claude/agents/meta/plan-auditor.md` ≥ 1. Verifies the agent body contains a cross-reference link to the docs-site GEARS migration guide anchor.

### AC-PAGA-006 (REQ-PAGA-008)

`diff -q .claude/agents/meta/plan-auditor.md internal/template/templates/.claude/agents/meta/plan-auditor.md` returns empty output (zero diff). Verifies template mirror parity.

### AC-PAGA-007 (REQ-PAGA-009, self-dogfood)

`go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md 2>&1 | grep -c "LegacyEARSKeyword"` = 0 **AND** the same command's output `grep -c "FrontmatterInvalid"` = 0. Self-dogfood verifies this SPEC's own REQs are GEARS-clean and the canonical 12-field frontmatter is intact.

### AC-PAGA-008 (cross-cutting PRESERVE invariant)

Post-commit working tree count ≤ baseline + 4 entries. The 4 expected additions: spec.md (new) + plan.md (new) + plan-auditor.md edit (modified) + template mirror edit (modified). Baseline at plan-phase commit time = 10 (4 modified config + 6 untracked from prior sessions). Asserts no scope drift beyond the declared file pairs.

## 4. Risks

### R1 — Self-lint failure (REQ-PAGA-009 / AC-PAGA-007)

**Risk**: If REQ-PAGA-001..009 contain literal `IF/THEN` tokens or modality keywords that violate GEARS conventions, the self-dogfood lint fails (AC-PAGA-007 = FAIL).

**Mitigation**: REQs use only GEARS keywords `When` / `While` / `Where` / `Ubiquitous (no keyword)` and the Unwanted form uses `shall not`. The legacy `IF` token appears only inside quoted markup (e.g., the line `the legacy form 'If [undesired condition], then [action]'` describing the deprecation case) where it is contextualized as a quoted deprecation example. Final pre-commit verification via `go run ./cmd/moai spec lint` confirms zero LegacyEARSKeyword findings.

### R2 — Template-mirror drift (REQ-PAGA-008 / AC-PAGA-006)

**Risk**: Standard Template-First Rule risk — local-only edits leave template mirror stale, breaking `moai update` baseline parity.

**Mitigation**: M2 milestone copies local body to template path in a single Bash invocation; AC-PAGA-006 `diff -q` assertion is part of pre-commit self-verification.

### R3 — Recursive self-audit paradox (transitional asymmetry)

**Risk**: This SPEC's plan-auditor run during plan-phase will use the *current* (EARS-only rubric) plan-auditor body since the update has NOT yet landed at plan-phase. Recursive paradox: how does an EARS-only auditor judge a GEARS-only SPEC?

**Mitigation**: Empirically resolved via Opus 4.7 natural-language mapping. The predecessor `SPEC-V3R6-SKILL-GEARS-ALIGN-001` already proved this path: plan-auditor iter-1 PASS 0.892 on a GEARS-notation SPEC under the unchanged EARS-only rubric (manager-docs sync `7b8f939b7` 2026-05-25). GEARS `When` → EARS Event-driven; GEARS `While` → EARS State-driven; GEARS `Where` capability-gate → EARS Optional; GEARS `shall not` → EARS Unwanted (informal mapping). Post-merge plan-auditor runs (e.g., on the next GEARS SPEC) will use the updated GEARS-aware rubric directly. plan.md M0 documents this transitional asymmetry as an explicit acknowledged invariant.

### R4 — Scope drift via interleaved-strategy mirror copy

**Risk**: When applying the same edits to local + template mirror sequentially (instead of as `cp local template`), human/agent error introduces non-identical bodies, breaking AC-PAGA-006.

**Mitigation**: M2 uses single-direction overwrite `cp -f .claude/agents/meta/plan-auditor.md internal/template/templates/.claude/agents/meta/plan-auditor.md` after all M1 edits complete; `diff -q` assertion blocks commit if non-empty.

## 5. Dependencies

- **Predecessor SPECs** (both `status: implemented`, blocking — both MUST be merged before this SPEC's run-phase):
  - `SPEC-V3R6-GEARS-MIGRATION-001` v0.2.0 — PR #1046 merged `134a43fac` 2026-05-22 — lint engine + 4-locale docs-site
  - `SPEC-V3R6-SKILL-GEARS-ALIGN-001` v0.2.0 — merged `7b8f939b7` + chore backfill `263db0600` 2026-05-25 — 5 guide files + 5 template mirrors GEARS-first
- **Tooling**: `go run ./cmd/moai spec lint` (LegacyEARSKeyword + FrontmatterInvalid rules) — both rules already production-ready from predecessor closures
- **Source files**:
  - `.claude/agents/meta/plan-auditor.md` (463 lines)
  - `internal/template/templates/.claude/agents/meta/plan-auditor.md` (mirror, 26469B baseline)
- **Reference documentation**: 4-locale `docs-site/content/{en,ko,ja,zh}/workflow-commands/moai-plan.md#gears-notation` (predecessor M3 deliverable)

## 6. Cross-References

- Canonical GEARS notation: `docs-site/content/en/workflow-commands/moai-plan.md#gears-notation` (4-locale)
- Lint engine SSOT: `internal/spec/lint.go` `LegacyEARSKeyword` rule (predecessor `SPEC-V3R6-GEARS-MIGRATION-001` M2)
- Authoring guide SSOT: `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Five Patterns (predecessor `SPEC-V3R6-SKILL-GEARS-ALIGN-001` M1)
- Frontmatter schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` (canonical 12 + optional fields)
- SPEC complexity tier policy: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier S inline AC pattern justification)
- Agent authoring rules: `.claude/rules/moai/development/agent-authoring.md` § Template-First Rule (mirror parity obligation)
