---
id: SPEC-EVAL-LOOP-001
status: draft
version: "0.1.0"
priority: High
labels: [evaluator, harness, generator-verifier, feedback-loop, wave-2, tier-1]
issue_number: null
scope: [harness.yaml, evaluator-active.md, run.md]
blockedBy: []
dependents: []
related_specs: [SPEC-LOOP-TERM-001, SPEC-EVAL-RUBRIC-001]
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 2
tier: 1
---

# SPEC-EVAL-LOOP-001: Generator-Verifier Loop in Standard Harness

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 2 / Tier 1. Anthropic Generator-Verifier 패턴을 standard harness에 도입하여 단방향 평가의 한계 해소.

---

## 1. Goal (목적)

`standard` harness 레벨에서 evaluator-active의 평가가 단일 final-pass로 종료되어 Anthropic Generator-Verifier 양방향성이 작동하지 않는 문제를 해결한다. config 기반 routing을 통해 standard 레벨에 최대 2회의 feedback iteration을 도입하고, improvement threshold와 stagnation detection을 표준화한다.

### 1.1 배경

- Anthropic blog "Multi-Agent Coordination Patterns": "The verifier needs an independent vantage point to catch flaws the generator cannot see." Generator-Verifier 양방향 루프는 패턴의 핵심.
- 현재 moai-adk-go의 evaluator-active는 `thorough` harness에서만 양방향 (Phase 2.0 contract + Phase 2.8a evaluation) 동작.
- `standard` harness (대다수 SPEC)는 final-pass 단일 평가만 수행 → feedback에 따른 generator 개선 기회 부재.

### 1.2 비목표 (Non-Goals)

- `thorough` 레벨의 sprint contract 메커니즘과의 통합 (별도 의미 유지)
- design 도메인 GAN Loop Contract (`§11`)의 일반 SPEC 워크플로우 적용 (design 한정 유지)
- evaluator-active 외 다른 agent의 feedback-loop 일반화 (본 SPEC은 evaluator 한정)
- frontmatter `evaluator_mode` 커스텀 필드 추가 (claude-code-guide가 거부, 우회 경로 사용)

---

## 2. Scope (범위)

### 2.1 In Scope

- `harness.yaml` `levels.standard` 절에 `evaluator_mode: feedback-loop` 키 + 파라미터 (`max_iterations`, `improvement_threshold`, `stagnation_consecutive`) 추가
- `evaluator-active.md` 본문에 Intervention Mode `feedback-loop` 신설 절
- `workflows/run.md` Phase 2.8a 부근에 feedback-loop 분기 로직 명시
- Template-First 동기화 (`internal/template/templates/`)
- docs-site 4개국어 reference 업데이트 (별도 PR, /moai sync 단계)

### 2.2 Exclusions (What NOT to Build)

- frontmatter `evaluator_mode` 커스텀 필드 (claude-code-guide 거부)
- thorough 레벨의 contract negotiation 변경
- design GAN Loop Contract의 standard 워크플로우 통합
- evaluator-active의 다른 dimension/weight 변경 (별도 SPEC)
- evaluator 외 agent의 feedback-loop 일반화

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.1.111+, Opus 4.7
- 영향 디렉터리: `.moai/config/sections/`, `.claude/agents/moai/`, `.claude/skills/moai/workflows/`
- 템플릿 동기화: `internal/template/templates/`

---

## 4. Assumptions (가정)

- A1: evaluator-active 본문에 명시된 protocol을 orchestrator가 일관되게 준수한다 (LSP 검증 불가, 텍스트 의도 전달)
- A2: 첫 iteration 평가 결과가 actionable feedback으로 변환 가능하다 (점수만이 아닌 구체 피드백 생성)
- A3: standard 레벨 SPEC의 평균 1.0~1.5 iteration 수렴이 가능하다 (acceptance baseline에서 측정)
- A4: token cost 증가는 standard 평균 +20-30% 범위 (acceptance에서 검증)
- A5: improvement_threshold 0.10은 thorough mode 0.05보다 관대 (의도된 차별화)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-EL-001**: THE HARNESS.YAML SHALL define `levels.standard.evaluator_mode = "feedback-loop"` as the default mode for the standard level.
- **REQ-EL-002**: THE HARNESS.YAML SHALL define `levels.standard.max_iterations = 2` as the upper bound for feedback iterations.
- **REQ-EL-003**: THE HARNESS.YAML SHALL define `levels.standard.improvement_threshold = 0.10` for stagnation detection.
- **REQ-EL-004**: THE HARNESS.YAML SHALL define `levels.standard.stagnation_consecutive = 1` as the consecutive-iteration threshold for escalation.
- **REQ-EL-005**: THE EVALUATOR-ACTIVE BODY SHALL document the feedback-loop intervention mode in a dedicated section under "Intervention Modes".
- **REQ-EL-006**: THE FEEDBACK-LOOP PROTOCOL SHALL preserve the existing final-pass behavior of standard mode when `feedback-loop` is disabled via configuration.

### 5.2 Event-Driven Requirements

- **REQ-EL-007**: WHEN harness level is resolved to "standard" AND `evaluator_mode = "feedback-loop"`, THE EVALUATOR-ACTIVE SHALL execute up to `max_iterations` evaluation cycles (default 2).
- **REQ-EL-008**: WHEN an iteration completes with score below the configured pass threshold, THE EVALUATOR SHALL produce actionable feedback for the next iteration before terminating the cycle.
- **REQ-EL-009**: WHEN `max_iterations` is reached without achieving pass threshold, THE EVALUATOR-ACTIVE SHALL emit a structured escalation report to the orchestrator including final score, dimension breakdown, and unresolved findings.
- **REQ-EL-010**: WHEN feedback is consumed by the generator (manager-ddd/tdd) and a revised artifact is produced, THE EVALUATOR SHALL re-score using the same evaluator profile loaded in iteration 1.

### 5.3 State-Driven Requirements

- **REQ-EL-011**: WHILE iteration count is below `max_iterations` AND current score < pass_threshold, THE EVALUATOR SHALL provide actionable feedback for the next iteration.
- **REQ-EL-012**: WHILE feedback-loop is in progress, THE EVALUATOR SHALL persist iteration state (iteration_number, score, feedback_summary) to `.moai/state/evaluator-loop/<SPEC-ID>.json` for resume capability.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-EL-013**: WHERE improvement between iterations is less than `improvement_threshold` (0.10), THE EVALUATOR SHALL flag the loop as stagnating in the iteration report.
- **REQ-EL-014**: IF stagnation persists for `stagnation_consecutive` (1) consecutive iterations in standard harness, THEN THE EVALUATOR-ACTIVE SHALL escalate to the orchestrator without performing additional iterations.
- **REQ-EL-015**: WHERE the SPEC frontmatter contains `evaluator_loop_disabled: true`, THE FEEDBACK-LOOP SHALL be skipped and final-pass behavior SHALL be applied.
- **REQ-EL-016**: IF the orchestrator detects feedback-loop unavailability (config absent or evaluator unable to score), THEN THE WORKFLOW SHALL fall back to final-pass with a non-blocking warning.

### 5.5 Unwanted (Negative) Requirements

- **REQ-EL-017**: THE FEEDBACK-LOOP SHALL NOT exceed `max_iterations` even if the score is still below pass threshold; escalation is the terminal action.
- **REQ-EL-018**: THE EVALUATOR-ACTIVE SHALL NOT modify or rewrite the generator's artifact directly; feedback must be returned for the generator to apply.

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| feedback-loop activates in standard harness | E2E test with sample SPEC | PASS |
| Token cost increase | average token measurement on 5 sample SPECs | < +25% vs baseline |
| Stagnation detection accuracy | controlled test cases | false-positive rate < 20% |
| Backward compatibility | existing thorough/minimal levels unchanged | unchanged |
| Template-First sync | `make build` produces no diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Frontmatter custom field 추가 금지 (claude-code-guide 결정)
- C2: thorough/minimal 레벨 동작 변경 금지
- C3: design domain GAN Loop과 의미 충돌 방지 (별도 의미 유지)
- C4: standard 레벨 default behavior 변경은 backward-compatible해야 함 (기본 활성화는 OK, 단 disable 경로 제공 필수)

---

## 9. Frontmatter Field Semantics (Wave 2 Tier 1 Standard)

This section defines the canonical meaning of inter-SPEC reference fields used in `.moai/specs/*/spec.md` frontmatter. All 5 SPECs in Wave 2 Tier 1 (EVAL-LOOP-001, LOOP-TERM-001, EVAL-RUBRIC-001, REVIEW-MULTI-001, SKILL-TEST-001) follow this standard.

| Field | Semantic | Blocking? |
|-------|----------|-----------|
| `blockedBy: [SPEC-X-001, ...]` | This SPEC's implementation cannot start until the listed SPECs are completed. HARD dependency. | Yes |
| `dependents: [SPEC-Y-001, ...]` | The listed SPECs are blocked by this SPEC (inverse of `blockedBy`). Forward declarations to future SPECs are allowed. | Yes (transitively) |
| `related_specs: [SPEC-Z-001, ...]` | Semantic association only; reference for context. NOT blocking. Cross-references for design coherence. | No |

### Application to this SPEC

- `blockedBy: []` — No prior SPEC must be completed first.
- `dependents: []` — No SPEC currently waits on this one for unblocking.
- `related_specs: [SPEC-LOOP-TERM-001, SPEC-EVAL-RUBRIC-001]` — These SPECs share the iterative-evaluation problem space and are designed to compose; they are NOT blocked by this SPEC and can develop in parallel.

---

End of spec.md (SPEC-EVAL-LOOP-001 v0.1.0).
