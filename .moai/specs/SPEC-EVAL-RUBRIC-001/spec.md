---
id: SPEC-EVAL-RUBRIC-001
status: draft
version: "0.1.0"
priority: High
labels: [evaluator, rubric, anchoring, profile, quality, wave-2, tier-1]
issue_number: null
scope: [evaluator-active.md, evaluator-profiles/]
blockedBy: []
dependents: []
related_specs: [SPEC-EVAL-LOOP-001]
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 2
tier: 1
---

# SPEC-EVAL-RUBRIC-001: Verifier Rubric Anchoring 확산 (일반 도메인)

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 2 / Tier 1. design `§12` Mechanism 1 (Rubric Anchoring)을 일반 evaluator-profiles에 확산하고 evaluator-active 본문에 mandatory 절 명시.

---

## 1. Goal (목적)

Anthropic blog "The verifier is only as good as its criteria"와 design constitution `§12` Mechanism 1에 따라, 일반 도메인 evaluator의 4-dimension scoring에 4-anchor (0.25/0.50/0.75/1.0) rubric anchoring을 의무화한다. evaluator-active 본문에 "Rubric Anchoring is mandatory" 절을 명시하고, 4개 evaluator-profile (default, strict, lenient, frontend)이 모두 4-anchor를 일관되게 정의하도록 한다.

### 1.1 배경

- Anthropic: "Vague criteria rubber-stamp outputs."
- design `§12` Mechanism 1 (FROZEN): "Scores without rubric justification are invalid."
- 현재 default.md는 4-anchor 정의되어 있으나, evaluator-active 본문에 mandatory 강제력 부재
- strict.md / lenient.md / frontend.md는 검증 필요 (구조 다를 수 있음)

### 1.2 비목표 (Non-Goals)

- design `§12` Mechanism 2-5 (Regression Baseline, Must-Pass Firewall, Independent Re-evaluation, Anti-Pattern Cross-check) 일반화 — Mechanism 1만 적용
- evaluator dimension 추가 (Functionality, Security, Craft, Consistency 4종 유지)
- evaluator weight 조정 (40/25/20/15 default 유지)
- frontend profile의 design-specific dimension 변경 (필요 시 4-anchor 적용만)
- 새 evaluator agent 신설

---

## 2. Scope (범위)

### 2.1 In Scope

- `evaluator-active.md`에 "Rubric Anchoring is Mandatory" 절 신설
- 점수 출력 schema 강화: 각 dimension 점수 옆 1줄 anchor 인용 의무화
- 4개 evaluator-profile (default / strict / lenient / frontend) 4-anchor 검증 + 보완
- "Scores without rubric justification" invalidation 메커니즘 명시 (단, soft warning부터 시작)
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- design `§12` Mechanism 2-5 일반화
- 4 dimension 외 신규 dimension
- evaluator weight 변경
- 별도 evaluator agent 신설
- LSP-level 검증 (anchor 인용은 텍스트 차원, hook이 enforce 불가)
- evaluator 출력의 다른 부분 (findings, recommendations) 변경

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.1.111+
- 영향 디렉터리: `.claude/agents/moai/`, `.moai/config/evaluator-profiles/`
- 템플릿 동기화: `internal/template/templates/`

---

## 4. Assumptions (가정)

- A1: evaluator-active가 본문에 명시된 protocol을 충실히 준수 (LLM 신뢰)
- A2: anchor 인용 1줄 추가는 평가 출력 토큰 < +5% 영향 (acceptance에서 검증)
- A3: 4-anchor (0.25/0.50/0.75/1.0)는 4-dimension 모두에 일관되게 적용 가능
- A4: profile별 anchor 톤은 strict/lenient에서 차별화 가능 (예: strict의 1.0은 lenient의 1.0보다 엄격)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-ER-001**: THE EVALUATOR-ACTIVE BODY SHALL contain a dedicated section titled "Rubric Anchoring is Mandatory" describing the requirement to reference rubric anchors when assigning scores.
- **REQ-ER-002**: THE EVALUATOR-ACTIVE OUTPUT SCHEMA SHALL require a one-line rubric anchor citation per dimension when reporting scores (e.g., "Functionality 0.75 — matches anchor: 'all primary acceptance criteria pass; minor edge cases missing'").
- **REQ-ER-003**: THE FOUR EVALUATOR PROFILES (default, strict, lenient, frontend) SHALL each provide rubric anchors at 0.25, 0.50, 0.75, 1.0 for ALL their evaluation dimensions.
- **REQ-ER-004**: THE BUILT-IN DEFAULT PROFILE (used when no profile file is found) SHALL provide rubric anchors for all 4 default dimensions (Functionality, Security, Craft, Consistency).
- **REQ-ER-005**: THE RUBRIC ANCHORS SHALL be expressed in concrete, observable terms (avoid subjective adjectives like "good", "acceptable" without qualifying detail).
- **REQ-ER-006**: THE SCORING RUBRIC SHALL be the canonical reference; arbitrary scoring without anchor reference SHALL be classified as a protocol violation in the evaluator output.

### 5.2 Event-Driven Requirements

- **REQ-ER-007**: WHEN evaluator-active assigns a score for any dimension, THE EVALUATOR SHALL reference rubric anchors at 0.25, 0.50, 0.75, 1.0 from the loaded profile.
- **REQ-ER-008**: WHEN the loaded profile lacks an anchor for a particular dimension, THE EVALUATOR SHALL fall back to the built-in default profile anchors and emit a non-blocking warning.
- **REQ-ER-009**: WHEN producing the final evaluation report, THE EVALUATOR SHALL include the anchor citation in the dimension breakdown table.

### 5.3 State-Driven Requirements

- **REQ-ER-010**: WHILE scoring is in progress, THE EVALUATOR SHALL maintain a working association between assigned scores and their cited anchors for the duration of the report generation.

### 5.4 Conditional Requirements

- **REQ-ER-011**: IF rubric justification is absent for a score in the evaluator output, THEN the score SHALL be flagged as `unanchored` in the report metadata, with severity `warning` initially and a path to escalation in future iterations.
- **REQ-ER-012**: WHERE evaluator-profile is not loaded (file missing or unreadable), THE BUILT-IN DEFAULT PROFILE SHALL provide rubric anchors for all 4 default dimensions.
- **REQ-ER-013**: WHERE the profile is `strict`, THE 1.0 ANCHOR SHALL be more demanding than the corresponding 1.0 anchor in `default` (e.g., requires zero findings of any severity vs. no Critical/High findings).
- **REQ-ER-014**: WHERE the profile is `lenient`, THE 0.50 ANCHOR SHALL be more permissive than the corresponding 0.50 anchor in `default`.
- **REQ-ER-015**: WHERE the profile is `frontend`, THE PROFILE'S DIMENSIONS (which may include Design Quality, Originality, Completeness, Functionality) SHALL each have 4-anchor definitions.
- **REQ-ER-016**: IF a dimension lacks an anchor at one or more of the 4 levels (0.25/0.50/0.75/1.0), THEN that profile SHALL be flagged as schema-violating and the evaluator-active SHALL emit a warning.

### 5.5 Unwanted (Negative) Requirements

- **REQ-ER-017**: THE EVALUATOR SHALL NOT assign a score without anchor citation in the production output (warning initially, error in future hardening).
- **REQ-ER-018**: THE PROFILES SHALL NOT contain anchors expressed solely as subjective adjectives without observable criteria.
- **REQ-ER-019**: THE evaluator-active body SHALL NOT delegate rubric anchoring to design `§12` Mechanism 1 reference (must be self-contained for general domain).

---

## 6. Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 4 profiles × 4 dim × 4 anchor | manual count | 64 anchors total (per default schema) |
| evaluator output anchor citation rate | sample 10 evaluations | 100% citation per dimension |
| token overhead | per-evaluation token measurement | < +5% |
| invalidation false-positive rate | controlled test | < 10% |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: design `§12` Mechanism 1만 일반화 (Mechanism 2-5 일반화 금지)
- C2: 4 dimension 외 신규 dimension 추가 금지
- C3: invalidation은 초기엔 warning (점진 hardening)
- C4: frontend profile의 design-specific dimension은 보존 (4-anchor 적용만)

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
- `related_specs: [SPEC-EVAL-LOOP-001]` — Shares the rubric-anchoring problem space (iteration-aware evaluators benefit from anchored rubrics); not blocked by or blocking this SPEC.

---

## 10. Rubric Anchor Self-Contained Reference (Default Profile)

This SPEC is self-verifying: the 16 anchors that comprise the canonical "default" profile are reproduced here verbatim so that Acceptance Scenario 1 ("16 anchor 검증") can be performed against this SPEC alone, without external file dependency. Strict / lenient / frontend profile anchors are derived from the default by the rules in REQ-ER-013/014/015 and are validated by Acceptance Scenarios 2/3/4 against their respective profile files.

The 4 dimensions and their relative weights are: Functionality (40%), Security (25%), Craft (20%), Consistency (15%). Each dimension MUST have anchors at exactly four score levels: 0.25, 0.50, 0.75, 1.00. Total = 4 × 4 = 16 anchors.

### 10.1 Functionality — Default Anchors (4)

| Score | Anchor (observable criterion) |
|-------|-------------------------------|
| 1.00 | All acceptance criteria pass with edge cases verified |
| 0.75 | All primary acceptance criteria pass; minor edge cases missing |
| 0.50 | Core functionality works; 1-2 acceptance criteria fail or are unverified |
| 0.25 | Basic skeleton present but multiple acceptance criteria fail |

### 10.2 Security — Default Anchors (4)

| Score | Anchor (observable criterion) |
|-------|-------------------------------|
| 1.00 | No findings of any severity; OWASP Top 10 checked |
| 0.75 | No Critical/High findings; Medium findings documented with mitigations |
| 0.50 | No Critical findings; High findings present but contained |
| 0.25 | Critical or multiple High findings present (triggers overall FAIL) |

### 10.3 Craft — Default Anchors (4)

| Score | Anchor (observable criterion) |
|-------|-------------------------------|
| 1.00 | Coverage >= 85%, clean code, no duplication, clear naming |
| 0.75 | Coverage >= 80%, minor style issues, acceptable naming |
| 0.50 | Coverage >= 70%, some duplication or unclear naming |
| 0.25 | Coverage < 70% or significant code quality issues |

### 10.4 Consistency — Default Anchors (4)

| Score | Anchor (observable criterion) |
|-------|-------------------------------|
| 1.00 | Fully consistent with project conventions and existing patterns |
| 0.75 | Minor deviations from conventions; no structural inconsistencies |
| 0.50 | Some pattern violations; deviations are localized |
| 0.25 | Significant inconsistencies with existing codebase patterns |

### 10.5 Verification Note

These 16 anchors are the canonical reference for the default profile. The implementation in `.moai/config/evaluator-profiles/default.md` MUST match this table verbatim. Any divergence between this SPEC's §10 table and the file content is a schema violation and SHALL be reported by Acceptance Scenario 1.

Profile-specific divergences (strict / lenient / frontend) are bounded by REQ-ER-013/014/015 and verified by Scenarios 2/3/4 against their respective profile files; the default profile remains the comparison baseline.

End of spec.md (SPEC-EVAL-RUBRIC-001 v0.1.0).
