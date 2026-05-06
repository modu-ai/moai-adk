# Research — SPEC-EVAL-RUBRIC-001 (Verifier Rubric Anchoring 확산)

**SPEC**: SPEC-EVAL-RUBRIC-001
**Wave**: 2 / Tier 1
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Multi-Agent Coordination Patterns":

> "The verifier is only as good as its criteria — vague criteria rubber-stamp outputs."

> "Concrete rubrics with anchored examples produce calibrated scores. Vague rubrics produce drift toward leniency, especially as the verifier sees more outputs and acclimates to mediocrity."

> "Anchor your rubric. Define what a 0.25, 0.50, 0.75, and 1.0 look like with concrete examples drawn from real or simulated outputs. Without anchors, the verifier interpolates from session context, which biases toward the most recent samples seen."

### 1.2 Verbatim from Existing Constitution

`.claude/rules/moai/design/constitution.md §12 Mechanism 1` (FROZEN):

> "Every evaluation criterion has a concrete rubric with examples of scores at 0.25, 0.50, 0.75, and 1.0. evaluator-active MUST reference the rubric when assigning scores. Scores without rubric justification are invalid."

### 1.3 검증 (claude-code-guide 결과)

claude-code-guide:
- **호환성**: ✅ 표준 관행. evaluator agent body에서 rubric 참조는 모범.
- **권고 채택**: ACCEPT — 일반 도메인 evaluator-profiles에 mechanism 적용.

---

## 2. 현재 상태 (As-Is)

### 2.1 default.md (이미 부분 구현)

`.moai/config/evaluator-profiles/default.md` 분석:

```
## Scoring Rubric

### Functionality (40%)
| Score | Description |
|-------|-------------|
| 1.00 | All acceptance criteria pass with edge cases verified |
| 0.75 | All primary acceptance criteria pass; minor edge cases missing |
| 0.50 | Core functionality works; 1-2 acceptance criteria fail or are unverified |
| 0.25 | Basic skeleton present but multiple acceptance criteria fail |
... (Security, Craft, Consistency 모두 동일 형식)
```

**관찰**: ✅ default.md는 이미 4-dimension × 4-anchor 형식으로 구현되어 있음.

### 2.2 strict.md / lenient.md / frontend.md

확인 필요. 본 SPEC의 출발점은 default.md의 표준이 다른 profile에도 일관되게 적용되었는지 검증 + evaluator-active 본문에 "rubric anchoring is mandatory" 절 명시 부재.

### 2.3 evaluator-active.md 본문 (격차 발견)

라인 79-89 (Profile Loading) 절은 잘 작성됨. 그러나:
- "Rubric Anchoring is mandatory" 절이 명시적으로 없음
- "Scores without rubric justification are invalid" 강제력이 design `§12`만큼 명확하지 않음
- 일반 SPEC evaluator가 점수만 출력하고 anchor 인용을 빠뜨릴 수 있는 구조

### 2.4 design `§12` Mechanism 1과의 격차

| 영역 | design `§12` (FROZEN) | 일반 evaluator-profile | 격차 |
|------|---------------------|----------------------|------|
| Anchor 강제력 | "MUST reference" | 권장 수준 | mandatory 절 부재 |
| 점수 무효화 (no anchor) | "Scores without rubric justification are invalid" | 명시 없음 | 강제 메커니즘 부재 |
| 4 anchor (0.25/0.50/0.75/1.0) | ✅ | default.md만 ✅ | 다른 profile 검증 필요 |
| anti-pattern cross-check (Mechanism 5) | ✅ | 없음 | 본 SPEC scope 외 |

---

## 3. 격차 분석

| 영역 | As-Is | To-Be | 격차 |
|------|-------|-------|------|
| default.md 4-anchor | ✅ | ✅ | 없음 (확인만) |
| strict.md 4-anchor | TBD | ✅ | 검증 + 보완 |
| lenient.md 4-anchor | TBD | ✅ | 검증 + 보완 |
| frontend.md 4-anchor | TBD | ✅ | 검증 + 보완 |
| evaluator-active.md "Rubric Anchoring mandatory" 절 | ❌ | ✅ | 신설 |
| 점수 출력 시 anchor 인용 의무화 | ❌ | ✅ | 출력 schema 강화 |
| 미인용 시 invalidation 메커니즘 | ❌ | ✅ | protocol 명시 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/agents/moai/evaluator-active.md` | 추가 | "Rubric Anchoring is Mandatory" 절 신설, 점수 출력 schema 강화 |
| `.moai/config/evaluator-profiles/default.md` | 보완 (필요 시) | 4-anchor 완전성 검증 |
| `.moai/config/evaluator-profiles/strict.md` | 검증 + 보완 | 4-anchor 통일 |
| `.moai/config/evaluator-profiles/lenient.md` | 검증 + 보완 | 4-anchor 통일 |
| `.moai/config/evaluator-profiles/frontend.md` | 검증 + 보완 | 4-anchor 통일 |

### 4.2 Templates (Template-First)

- `internal/template/templates/.moai/config/evaluator-profiles/*.md`: 동일 변경
- `internal/template/templates/.claude/agents/moai/evaluator-active.md`: 동일 변경

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| 4 profile마다 anchor 톤이 달라야 하는데 일률 적용 | Medium | Medium | strict는 더 엄격, lenient는 더 관대한 anchor 톤 유지 |
| evaluator 출력에 anchor 인용 필수 → 토큰 비용 증가 | Medium | Low | 인용은 dimension당 1줄 (~30 tokens) 충분 |
| invalidation 메커니즘 → re-evaluation loop 가능성 | Low | High | invalidation은 critical violation에 한정, soft warning은 non-blocking |
| design `§12` Mechanism 1과의 의미 충돌 | Low | High | 일반 profile은 mandatory만 적용, anti-pattern cross-check 등은 design 한정 유지 |

### 5.2 Assumptions

- A1: 4-anchor 형식 (0.25/0.50/0.75/1.0)이 모든 dimension에 적합함 (Functionality, Security, Craft, Consistency 외에도)
- A2: evaluator-active가 텍스트 protocol을 충실히 준수 (LLM 신뢰)
- A3: anchor 인용 1줄 추가는 평가 비용에 무의미한 수준

---

## 6. 측정 계획

| Metric | 측정 방법 | 목표 |
|--------|----------|------|
| anchor coverage in profiles | 4 dimension × 4 anchor = 16 cells per profile | 100% (4 profiles × 16 = 64) |
| anchor citation in evaluator output | sample 10 evaluation outputs | 100% (각 dimension별 1줄 인용) |
| invalidation false-positive rate | controlled test | < 10% |
| evaluator output token overhead | sample 5 evaluations | < +5% |

---

## 7. 대안 검토

| 대안 | 채택 | 이유 |
|------|-----|------|
| design `§12` 통째로 일반 도메인 적용 | ❌ | Mechanism 2-5는 design 특화, scope creep |
| profile-level rubric만 강화 (evaluator body 변경 없음) | ❌ | 강제력 없음, 점수 출력에서 인용 의무화 안 됨 |
| Mechanism 1 (Rubric Anchoring)만 일반화 + body 강화 | ✅ | 최소 변경으로 핵심 효과 |
| 새 evaluator agent (`evaluator-strict`) 신설 | ❌ | 복잡도 증가, 기존 default 강화가 더 단순 |

---

## 8. 참고 SPEC

- SPEC-AGENCY-ABSORB-001 §12: design `§12` Mechanism 1 출처 (FROZEN)
- SPEC-EVAL-LOOP-001 (Wave 2 sibling): standard feedback-loop과 짝
- SPEC-V3R3-HARNESS-001: harness 통합

---

## 9. Open Questions

- OQ1: profile별 anchor 톤 차별화 범위 (strict는 0.75 기준 더 엄격하게? lenient는 0.50 기준 관대?) → plan에서 결정
- OQ2: anchor 인용 형식 표준 (1줄? 인용 전체 anchor 텍스트?) → spec에서 결정
- OQ3: invalidation은 hard FAIL인가 warning인가? → spec에서 결정
- OQ4: frontend.md는 dimension 자체가 다른가 (Design Quality, Originality 등 design 차원 포함)? → 검증 필요

---

End of research.md (SPEC-EVAL-RUBRIC-001).
