---
id: SPEC-SKILL-DESC-001
status: draft
version: "0.1.0"
priority: Medium
labels: [skill-description, optimization, builder-skill, regression-test, false-positive, false-negative, wave-3, tier-2]
issue_number: null
scope: [.claude/agents/moai/builder-skill.md, internal/skill/optimizer, cmd/moai, .claude/rules/moai/development]
blockedBy: [SPEC-SKILL-TEST-001]
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 3
tier: 2
---

# SPEC-SKILL-DESC-001: Skill Description Optimization

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 3 / Tier 2. Anthropic "Improving Skill Creator: Test, Measure, Refine" 권고에 따라 builder-skill에 description optimization 기능 추가. 의존: SPEC-SKILL-TEST-001 (Wave 2)의 regression test framework 활용.

---

## 1. Goal (목적)

`builder-skill` agent에 skill description 자동 분석/제안 기능을 추가한다. SPEC-SKILL-TEST-001 (Wave 2)이 제공하는 regression test framework를 활용하여 false-positive rate (FP) 및 false-negative rate (FN)을 측정하고, 임계값(FP > 15% 또는 FN > 10%)을 초과하는 skill에 대해 description tightening/broadening 제안을 자동 생성한다. Auto-apply 금지 (사용자 승인 필수).

### 1.1 배경

- Anthropic blog "Improving Skill Creator: Test, Measure, Refine": "Skill-creator's description optimization feature analyzes your current description against sample prompts and suggests edits."
- 본 프로젝트의 100+ skill description은 자동 측정 인프라 부재 → 품질 추정만 가능
- broad description은 false positive, narrow description은 false negative → routing 품질 저하
- SPEC-SKILL-TEST-001 (Wave 2 완료)의 sample prompt 기반 metric 측정 인프라 활용

### 1.2 비목표 (Non-Goals)

- Auto-apply (사용자 승인 없이 description 자동 변경 금지)
- skill body까지 동시 최적화 (별도 SPEC)
- frontmatter 다른 metadata (allowed-tools 등) 최적화 (별도 SPEC)
- 임계값 자동 조정 알고리즘
- LLM 호출 없는 rule-based suggestion 단독 (LLM 보강 사용)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/agents/moai/builder-skill.md` 본문 확장 (description optimizer protocol 절)
- `internal/skill/optimizer/analyzer.go` — sample prompts → FP/FN metric 측정 로직 (SPEC-SKILL-TEST-001 framework 호출)
- `internal/skill/optimizer/suggester.go` — tightening/broadening 제안 생성 (LLM call 또는 advisor)
- `cmd/moai/skill.go`에 `moai skill optimize <name>` 서브커맨드 추가
- `.claude/rules/moai/development/skill-description.md` 작성 가이드 신규
- before/after diff 출력
- optimization report `.moai/reports/skill-optimization-<NAME>-<DATE>.md`
- 임계값: FP > 15% (tightening), FN > 10% (broadening)
- 사용자 승인 게이트 (AskUserQuestion 또는 `--apply` flag)
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- Auto-apply 메커니즘
- skill body / 다른 metadata 최적화
- 임계값 자동 조정
- LLM 미사용 단독 rule-based optimizer
- Multi-skill batch optimization (단일 skill 단위만)
- description history / version tracking (별도 SPEC)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- 의존: SPEC-SKILL-TEST-001 framework (Wave 2 산출물)
- 영향 디렉터리: `internal/skill/optimizer/`, `cmd/moai/`, `.claude/agents/moai/`, `.claude/rules/moai/development/`, `.claude/skills/moai/builder-skill/`
- 의존 LLM: builder-skill의 advisor (Opus 또는 Sonnet+advisor) — SPEC-ADVISOR-001 패턴 활용 시 효율적

---

## 4. Assumptions (가정)

- A1: SPEC-SKILL-TEST-001 framework 안정 작동 (Wave 2 완료 후)
- A2: 각 skill에 최소 5-10 sample prompts 보유
- A3: builder-skill agent가 description rewriting LLM call 가능
- A4: 임계값 15%/10%는 baseline 출발점, 측정 후 조정
- A5: 사용자 승인 게이트는 본 SPEC의 핵심 안전장치 (auto-apply 절대 금지)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-SD-001**: THE COMMAND `moai skill optimize <name>` SHALL produce a description optimization report at `.moai/reports/skill-optimization-<NAME>-<DATE>.md`.
- **REQ-SD-002**: THE OPTIMIZATION SHALL run regression tests via the SPEC-SKILL-TEST-001 framework before and after suggesting changes.
- **REQ-SD-003**: THE BUILDER-SKILL BODY SHALL document the description optimizer protocol in a dedicated section.
- **REQ-SD-004**: THE OPTIMIZATION SHALL require explicit user approval before applying any description change to skill frontmatter.

### 5.2 Event-Driven Requirements

- **REQ-SD-005**: WHEN `moai skill optimize <name>` is invoked, THE OPTIMIZER SHALL run sample prompts (provided by SPEC-SKILL-TEST-001 framework) against the current description AND compute false_positive_rate (FP) and false_negative_rate (FN).
- **REQ-SD-006**: WHEN FP exceeds 15%, THE OPTIMIZER SHALL produce a tightening suggestion (narrower description) using LLM call.
- **REQ-SD-007**: WHEN FN exceeds 10%, THE OPTIMIZER SHALL produce a broadening suggestion (wider description) using LLM call.
- **REQ-SD-008**: WHEN both FP and FN are below thresholds, THE OPTIMIZER SHALL emit "no optimization needed" message.
- **REQ-SD-009**: WHEN the user approves the suggestion, THE OPTIMIZER SHALL apply the change to the skill frontmatter AND re-run regression tests to measure improvement.

### 5.3 State-Driven Requirements

- **REQ-SD-010**: WHILE optimization is in progress, THE OPTIMIZER SHALL persist intermediate state (current description, suggested description, FP/FN measurements) to `.moai/state/skill-optimizer/<NAME>.json` for resume capability.
- **REQ-SD-011**: WHILE the skill has fewer than 5 sample prompts in its test framework, THE OPTIMIZER SHALL halt with a "data insufficient" message.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-SD-012**: WHERE per-skill threshold overrides are declared in the skill frontmatter (e.g., `optimization_thresholds.fp: 0.20`), THE OPTIMIZER SHALL use those overrides instead of global defaults.
- **REQ-SD-013**: IF the user provides `--apply` flag, THEN THE OPTIMIZER SHALL skip interactive approval and apply the suggestion directly (audit log entry mandatory).
- **REQ-SD-014**: WHERE SPEC-SKILL-TEST-001 framework is unavailable (file missing), THE OPTIMIZER SHALL emit a blocker error referencing the dependency.
- **REQ-SD-015**: IF the same description is suggested twice consecutively (loop detection), THEN THE OPTIMIZER SHALL halt with "convergence reached" message.

### 5.5 Unwanted (Negative) Requirements

- **REQ-SD-016**: THE OPTIMIZER SHALL NOT apply description changes without user approval (`--apply` flag is the only exception, with mandatory audit log).
- **REQ-SD-017**: THE OPTIMIZER SHALL NOT modify any field other than `description` in skill frontmatter.
- **REQ-SD-018**: THE OPTIMIZER SHALL NOT run optimization on multiple skills in batch (single skill per invocation).
- **REQ-SD-019**: THE OPTIMIZER SHALL NOT skip cross-effect verification (full skill catalog regression run mandatory after apply).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| optimization 정확도 | 5 skill 시범 적용 acceptance rate | >= 60% |
| FP 감소 (tightening 후) | 재측정 | -50% 이상 |
| FN 감소 (broadening 후) | 재측정 | -50% 이상 |
| Cross-effect | full catalog regression | <= 5% routing 변동 |
| Auto-apply 차단 | unit test (no-approval scenario) | 100% block |
| Loop 방지 | convergence detection | works |
| Cross-platform | 3 platforms | 3/3 PASS |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: blockedBy `SPEC-SKILL-TEST-001` (Wave 2 산출물 완성 필수)
- C2: Auto-apply 절대 금지 (사용자 승인 또는 명시 `--apply` flag)
- C3: 단일 skill optimization만 (batch 금지)
- C4: full catalog regression 의무 (cross-effect 보호)
- C5: Template-First Rule 준수
- C6: 임계값 default (FP 15%, FN 10%)는 측정 후 조정 가능

End of spec.md (SPEC-SKILL-DESC-001 v0.1.0).
