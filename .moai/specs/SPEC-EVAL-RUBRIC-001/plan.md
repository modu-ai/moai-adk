---
id: SPEC-EVAL-RUBRIC-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-EVAL-RUBRIC-001

## 1. Overview

evaluator-active 본문에 "Rubric Anchoring is Mandatory" 절 신설 + 4 evaluator-profile (default/strict/lenient/frontend) 4-anchor 일관성 검증 및 보완.

## 2. Approach Summary

**전략**: Body-First (강제력) + Profile-Second (anchor 정합성) + Output-Third (citation schema).

1. evaluator-active.md에 mandatory 절 추가 (강제력 정립)
2. 4 profile 모두 4-anchor 검증 + 보완
3. 출력 schema 강화 (anchor citation 의무화)
4. invalidation은 warning 단계부터 시작 (점진 hardening)

## 3. Milestones

### M0 — Pre-flight Audit (Priority: Critical)

- [ ] 4 evaluator-profile 파일 verbatim 캡처 (default, strict, lenient, frontend)
- [ ] 각 profile에서 4-dimension × 4-anchor 누락 셀 식별
- [ ] evaluator-active.md 현재 Profile Loading 절 verbatim 캡처
- [ ] sample 5 evaluation outputs (있다면 archive에서) 검토 — 현재 anchor citation 비율 측정

**Exit Criteria**: profile gap matrix 작성, baseline evaluator output 검토 완료

### M1 — Evaluator Body "Rubric Anchoring is Mandatory" Section (Priority: Critical)

- [ ] `.claude/agents/moai/evaluator-active.md`에 신규 절 추가:
  - 위치: "Profile Loading" 절 다음
  - 내용:
    1. Mandatory rule 명시: "When assigning a score, you MUST cite the rubric anchor."
    2. Citation format: "{Dimension} {score} — matches anchor: '{anchor text}'"
    3. Fallback: profile에 anchor 부재 시 default fallback + warning
    4. Invalidation policy: 미인용 점수는 `unanchored` 메타데이터 부착, 초기 severity는 warning
- [ ] Template 동기화: `internal/template/templates/.claude/agents/moai/evaluator-active.md`

**Exit Criteria**: evaluator-active.md에 mandatory 절 명시 + Template-First sync

### M2 — Default Profile Verification (Priority: High)

- [ ] `.moai/config/evaluator-profiles/default.md` 4-anchor 완전성 확인
- [ ] research에서 확인한 대로 이미 4×4=16 anchor 존재 → 누락 시 보완
- [ ] anchor 텍스트가 observable terms로 작성되었는지 점검 (subjective adj 제거)
- [ ] Template 동기화

**Exit Criteria**: default.md 4×4=16 anchor 검증 완료

### M3 — Strict / Lenient / Frontend Profile Verification (Priority: High)

- [ ] strict.md 4-anchor 검증 + strict-tone 차별화 확인 (1.0이 default보다 엄격)
- [ ] lenient.md 4-anchor 검증 + lenient-tone 차별화 확인 (0.50이 default보다 관대)
- [ ] frontend.md 검증:
  - dimension 차이 (Design Quality, Originality, Completeness, Functionality) 존재 시
  - 각 dimension에 4-anchor 적용
- [ ] 누락 anchor 보완 시 design `§12` 톤을 참조하되 일반 도메인에 맞게 조정
- [ ] Template 동기화 (4 profiles)

**Exit Criteria**: 4 profile 모두 4×4=16 anchor (frontend는 dimension 수 다를 수 있음)

### M4 — Output Schema Strengthening (Priority: High)

- [ ] evaluator-active.md의 평가 출력 형식 절 강화:
  - 점수 표 형식 예시 업데이트 (anchor citation 컬럼 추가)
  - Dimensional Score Block 예시:
    ```
    | Dimension | Score | Anchor Citation |
    |-----------|-------|----------------|
    | Functionality | 0.75 | "all primary acceptance criteria pass; minor edge cases missing" |
    | Security | 1.00 | "no findings of any severity; OWASP Top 10 checked" |
    ```
- [ ] Template 동기화

**Exit Criteria**: 출력 schema에 anchor citation 컬럼 명시

### M5 — Invalidation Policy (Priority: Medium)

- [ ] evaluator-active.md에 "Invalidation" 절 추가:
  - 미인용 점수 → `unanchored: true` 메타데이터 부착
  - severity: warning (초기) → 향후 SPEC에서 error로 escalate 가능
  - report 하단에 "unanchored scores" 카운트 명시
- [ ] orchestrator routing: warning은 non-blocking, 향후 error로 hardening 시 재평가 트리거

**Exit Criteria**: invalidation policy 문서화 + 초기 severity warning

### M6 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md의 Given-When-Then 시나리오 8개 모두 PASS
- [ ] M0 baseline 대비 anchor citation rate 100% 달성
- [ ] token overhead < +5% 확인
- [ ] plan-auditor 검증 PASS
- [ ] Template-First sync clean

**Exit Criteria**: acceptance.md PASS + plan-auditor PASS

## 4. Technical Approach

### 4.1 Mandatory Section Skeleton (evaluator-active.md 추가 예정)

```markdown
## Rubric Anchoring is Mandatory

When you assign a score for any dimension, you MUST cite the rubric anchor
text from the loaded profile. Scores without anchor citation are protocol
violations and SHALL be flagged in the report.

### Citation Format

In the dimensional score breakdown:

| Dimension | Score | Anchor Citation |
|-----------|-------|----------------|
| Functionality | 0.75 | "all primary acceptance criteria pass; minor edge cases missing" |

### Fallback

If the loaded profile lacks an anchor for a dimension at the assigned score
level, fall back to the built-in default profile anchors and emit a warning
in the report's metadata section.

### Invalidation

Scores without anchor citation are flagged as `unanchored: true` in report
metadata. Severity is `warning` for now. The report's footer SHALL include
"Unanchored scores: N" count.
```

### 4.2 Anchor Tone Differentiation Examples

| Profile | Functionality 1.0 anchor (excerpt) |
|---------|-----------------------------------|
| default | "All acceptance criteria pass with edge cases verified" |
| strict | "All acceptance criteria pass with edge cases verified AND every property test/fuzz test passes" |
| lenient | "All primary acceptance criteria pass; edge cases acceptable as documented" |

### 4.3 Frontend Profile Special Handling

frontend.md는 design-specific dimension (Design Quality, Originality 등) 포함 가능. 본 SPEC은:
- 4 default dimension 외 추가 dimension에도 4-anchor 적용 의무
- 각 dimension별 anchor가 (subjective term 회피하며) 작성되었는지 검증

## 5. Risks and Mitigations

| Risk | P | I | Mitigation |
|------|---|---|-----------|
| profile별 anchor 톤 일률 적용으로 strict/lenient 의미 약화 | Medium | Medium | M3에서 톤 차별화 의무 명시 |
| anchor 인용 추가 → 토큰 +5% 초과 | Low | Low | 1줄 인용 (~30 tokens), 측정으로 확인 |
| invalidation false-positive (정당한 평가가 unanchored로 flag) | Medium | Medium | 초기 severity warning만 (non-blocking) |
| frontend.md dimension 수 차이로 schema 불일치 | Medium | Low | frontend는 dimension 수 무관, 각 dimension별 4-anchor만 의무 |
| design `§12` Mechanism 1 의미 충돌 | Low | High | 일반 도메인은 self-contained 절 (design 참조 금지) |

## 6. Dependencies

- 선행 SPEC: 없음
- 동반 SPEC: SPEC-EVAL-LOOP-001 (feedback-loop은 본 SPEC의 anchor citation을 활용)
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1**: strict/lenient 차별화 → ✅ 1.0 / 0.50 anchor에 톤 차이 명시
- **OQ2**: 인용 형식 → ✅ 1줄, "matches anchor: '<text>'" 표준
- **OQ3**: invalidation severity → ✅ 초기 warning, 향후 error hardening (별도 SPEC)
- **OQ4**: frontend.md 검증 → M3에서 dimension 별 4-anchor 적용 의무

## 8. Rollout Plan

1. M1-M5 구현 후 본 SPEC 자체 평가에 적용 (dogfooding)
2. 다음 SPEC 5개에서 anchor citation rate 100% 확인
3. CHANGELOG entry + minor release
4. 추후 SPEC: invalidation severity warning → error hardening

End of plan.md.
