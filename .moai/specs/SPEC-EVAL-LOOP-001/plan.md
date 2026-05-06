---
id: SPEC-EVAL-LOOP-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-EVAL-LOOP-001

## 1. Overview

evaluator-active의 standard harness 동작을 단일 final-pass에서 최대 2-iteration feedback-loop으로 확장. config-driven routing으로 frontmatter custom field 회피.

## 2. Approach Summary

**전략**: Configuration-First, Behavior-Second, Documentation-Third.

1. `harness.yaml` 스키마 확장 (config가 진실의 단일 출처)
2. `evaluator-active.md` 본문에 protocol 명시 (LLM이 읽고 수행)
3. `workflows/run.md` Phase 2.8a에 분기 로직 추가 (orchestrator routing)
4. Template-First 동기화 + 4개국어 docs sync

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight Verification (Priority: Critical)

- [ ] `harness.yaml` 현재 스키마 verbatim 캡처 (regression baseline)
- [ ] `evaluator-active.md` 현재 Intervention Modes 절 verbatim 캡처
- [ ] 기존 standard SPEC 5개에서 평균 token cost / iteration count baseline 측정
- [ ] 기존 thorough SPEC 1개에서 contract negotiation 동작 verify (regression 방지)

**Exit Criteria**: baseline 메트릭 4종 기록, regression test seed 작성

### M1 — Configuration Schema Extension (Priority: High)

- [ ] `harness.yaml.levels.standard` 절에 다음 키 추가:
  - `evaluator_mode: feedback-loop` (default for standard)
  - `max_iterations: 2`
  - `improvement_threshold: 0.10`
  - `stagnation_consecutive: 1`
- [ ] Template 동기화: `internal/template/templates/.moai/config/sections/harness.yaml` 미러
- [ ] `make build` 실행 후 embedded.go diff 확인
- [ ] 기존 thorough/minimal 절 변경 없음 확인 (regression)

**Exit Criteria**: schema validation pass, embedded.go regenerated, baseline diff = 0 for thorough/minimal

### M2 — Evaluator Body Protocol Documentation (Priority: High)

- [ ] `.claude/agents/moai/evaluator-active.md`에 새 절 추가:
  - **Intervention Modes** 절을 확장: `feedback-loop` 모드 신설
  - protocol 5단계 명시: (1) initial evaluation, (2) feedback generation, (3) generator hand-off, (4) re-evaluation, (5) escalation
  - state persistence 경로 명시: `.moai/state/evaluator-loop/<SPEC-ID>.json`
- [ ] Template 동기화: `internal/template/templates/.claude/agents/moai/evaluator-active.md`

**Exit Criteria**: feedback-loop 절이 evaluator-active.md에 존재 + Template-First 동기화 완료

### M3 — Orchestrator Routing in run.md (Priority: High)

- [ ] `.claude/skills/moai/workflows/run.md` Phase 2.8a (Active Evaluator Pass) 절에 분기 로직 추가:
  - harness level 해석 → standard + feedback-loop 모드면 iteration 루프 진입
  - max_iterations 도달 OR pass score 도달 OR stagnation detected → 종료
- [ ] fallback 경로: `evaluator_loop_disabled: true` SPEC frontmatter 또는 config 부재 시 final-pass 동작
- [ ] state file 갱신 의무 명시 (state/evaluator-loop/ 디렉토리)

**Exit Criteria**: run.md에 routing 분기 명시, fallback 경로 explicit

### M4 — State Persistence + Resume Logic (Priority: Medium)

- [ ] `.moai/state/evaluator-loop/` 디렉토리 schema 정의 (JSON)
  - `iteration_number`, `score`, `feedback_summary`, `last_updated_at`
- [ ] resume 케이스 처리: orchestrator가 state file 발견 시 마지막 iteration 다음부터 재개
- [ ] cleanup: pass 또는 escalation 도달 시 state file 삭제 또는 archive

**Exit Criteria**: state schema 문서화, resume scenario E2E test seed

### M5 — Documentation Sync (Priority: Medium)

- [ ] `.claude/rules/moai/workflow/spec-workflow.md`의 Phase 2.8a 절 업데이트
- [ ] docs-site 4개국어 reference (`content/{ko,en,ja,zh}/.../harness.md`) — /moai sync 단계
- [ ] CHANGELOG entry 작성 (Unreleased)

**Exit Criteria**: 모든 reference doc에 feedback-loop 모드 설명 등재

### M6 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md의 Given-When-Then 시나리오 8개 모두 PASS
- [ ] M0 baseline 대비 token cost 증가 < +25%
- [ ] thorough/minimal regression: zero behavior change
- [ ] plan-auditor 검증 PASS (frontmatter, EARS keyword count, ID uniqueness)

**Exit Criteria**: acceptance.md 모든 시나리오 PASS + plan-auditor PASS

## 4. Technical Approach

### 4.1 Config Schema (harness.yaml)

```yaml
levels:
  standard:
    description: "Balanced quality for most development"
    evaluator: true
    sprint_contract: false
    # NEW: feedback-loop config
    evaluator_mode: feedback-loop  # alternatives: final-pass, feedback-loop
    max_iterations: 2
    improvement_threshold: 0.10
    stagnation_consecutive: 1
```

### 4.2 Protocol Pseudocode (evaluator-active body 문서화 예정)

```
iteration = 1
state = {scores: [], feedback: []}
while iteration <= max_iterations:
  score, findings = evaluate(artifact, profile)
  state.scores.append(score)
  if score >= pass_threshold:
    return PASS
  if iteration > 1:
    delta = score - state.scores[-2]
    if delta < improvement_threshold:
      return ESCALATE("stagnation")
  feedback = generate_feedback(findings)
  state.feedback.append(feedback)
  hand_off_to_generator(feedback)
  iteration += 1
return ESCALATE("max_iterations_reached")
```

### 4.3 State File (JSON)

```json
{
  "spec_id": "SPEC-XXX",
  "harness_level": "standard",
  "evaluator_mode": "feedback-loop",
  "iterations": [
    {"n": 1, "score": 0.62, "feedback_summary": "..."}
  ],
  "status": "in-progress",
  "last_updated_at": "2026-04-30T12:34:56Z"
}
```

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Token cost +50% (worst case) | Low | High | M0 baseline 측정 후 max_iterations 1로 축소 가능 (config 변경만) |
| feedback 품질 불충분 → 무의미한 iteration | Medium | High | improvement_threshold 0.10 + stagnation 1회로 즉시 cut |
| state file race (parallel evaluation) | Low | Medium | SPEC-ID당 단일 evaluator instance 보장 (orchestrator 설계상) |
| thorough mode 동작 회귀 | Low | Critical | M0 verification + M1 diff guard |
| frontmatter `evaluator_loop_disabled` 키 미지원 → schema warning | Medium | Low | optional 키로 정의, 없으면 default behavior |

## 6. Dependencies

- 선행 SPEC: 없음 (독립 가능)
- 동반 SPEC: SPEC-LOOP-TERM-001 (termination 표준화 — 본 SPEC의 max_iterations / stagnation 패턴이 그곳 standard schema 기준)
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1**: feedback-loop을 standard에 default ON → ✅ 채택 (기존 standard 사용자에 자동 적용, opt-out 경로 제공)
- **OQ2**: max_iterations 기본 2 → ✅ 채택 (대부분 1-2회 수렴, 3회는 thorough 영역)
- **OQ3**: improvement_threshold 0.10 → ✅ 채택 (thorough 0.05의 2배, 표준 SPEC에 더 관대)
- **OQ4**: feedback 생성 LLM 비용 → 측정으로 결정 (M6 acceptance에서 cost validation)

## 8. Rollout Plan

1. M1-M3 구현 후 dogfooding: 본 SPEC 자체에 standard + feedback-loop 적용
2. 5개 sample SPEC에서 baseline 비교
3. CHANGELOG에 명시 후 v2.x.0 minor release
4. 부정적 시그널 (token cost > +30% 또는 stagnation false-positive > 30%) 발견 시 default 비활성화 hotfix

End of plan.md.
