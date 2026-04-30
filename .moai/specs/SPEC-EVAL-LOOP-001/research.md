# Research — SPEC-EVAL-LOOP-001 (Generator-Verifier Loop in Standard Harness)

**SPEC**: SPEC-EVAL-LOOP-001
**Wave**: 2 / Tier 1 (검증 완료된 단기 권고)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Multi-Agent Coordination Patterns":

> "A generator receives a task and produces an initial output, which it passes to a verifier for evaluation."

> "The verifier is only as good as its criteria — vague criteria rubber-stamp outputs."

> "When the verifier and generator are tightly coupled, you lose the asymmetry that makes the pattern work. The verifier needs an independent vantage point — different model, different prompt, different context — to catch flaws the generator cannot see."

> "Generator-verifier without iteration is just review. The loop emerges when the generator can act on the verifier's feedback and produce a revised output."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 Claude Code 공식 문서 기준으로 본 권고를 검증함. 결론:

- **호환성**: 부분 지원. agent frontmatter에 임의 커스텀 필드 (`evaluator_mode` 등) 추가 불가
- **표준 우회**: `harness.yaml` (config) + skill body (text) + orchestrator (routing) 3-way 조합으로 동등한 동작 확보 가능
- **권고 채택**: ACCEPT — 단, 구현은 frontmatter 필드 의존이 아닌 config-driven routing으로

---

## 2. 현재 상태 (As-Is)

### 2.1 evaluator-active의 두 가지 Intervention Mode

`.claude/agents/moai/evaluator-active.md` (라인 99-102):

```
- final-pass (standard harness): Single evaluation at Phase 2.8a
- per-sprint (thorough harness): Phase 2.0 contract negotiation + Phase 2.8a post-evaluation
```

**관찰**:
- `thorough` 레벨에서만 Generator-Verifier 패턴이 양방향으로 작동 (Phase 2.0 contract → 구현 → Phase 2.8a 평가)
- `standard` 레벨은 final-pass 단일 평가만 수행 → feedback loop 부재
- 일반 SPEC 다수가 `standard`에 속함 → 대다수 워크플로우에서 Generator-Verifier 양방향성 미작동

### 2.2 Profile Loading Mechanism

`.claude/agents/moai/evaluator-active.md` (라인 79-89):

1. SPEC frontmatter `evaluator_profile` 필드 확인
2. 있으면 `.moai/config/evaluator-profiles/{profile}.md` 로드
3. 없으면 `harness.default_profile`로 fallback
4. 파일 없으면 built-in default (Functionality 40%, Security 25%, Craft 20%, Consistency 15%)

**관찰**: profile은 dimension weight + threshold만 다룸. iteration/loop 동작은 profile에 정의 불가.

### 2.3 harness.yaml의 Level 정의

`.moai/config/sections/harness.yaml`에서 `levels.standard`:

```yaml
standard:
  description: "Balanced quality for most development"
  # (evaluator: true 추정, sprint_contract: false)
```

**관찰**: standard 레벨에서는 `evaluator: true`이지만 `sprint_contract: false` → 평가는 하나 contract 협상은 없음. 결과적으로 한 번 평가 후 종료.

### 2.4 GAN Loop Contract (design 도메인 한정)

`.claude/rules/moai/design/constitution.md §11` GAN Loop Contract:

- max_iterations: 5
- pass_threshold: 0.75
- escalation_after: 3
- improvement_threshold: 0.05

**관찰**: 이 contract는 **design 워크플로우 한정**. 일반 SPEC 워크플로우에는 적용되지 않음.

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| standard harness 평가 | 단일 final-pass | 최대 2회 feedback iteration | iteration 미지원 |
| improvement 측정 | 측정 안 함 | iteration 간 0.10 임계값 | 측정 메커니즘 부재 |
| stagnation 감지 | 없음 | 1 consecutive iteration | 감지 로직 부재 |
| escalation 경로 | 없음 (final이라 불필요) | orchestrator로 escalate | 부재 |
| harness.yaml의 evaluator_mode | 키 없음 | 새 키 추가 필요 | 스키마 확장 필요 |
| evaluator-active 본문 protocol | feedback-loop 명시 없음 | 명시적 프로토콜 절 | 문서화 필요 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 수정 대상

| 파일 | 수정 유형 | 라인 범위 (예상) | 변경 사유 |
|------|----------|-----------------|----------|
| `.moai/config/sections/harness.yaml` | 추가 | levels.standard 절 (~5 lines) | `evaluator_mode`, `max_iterations`, `improvement_threshold` 키 추가 |
| `.claude/agents/moai/evaluator-active.md` | 추가 | Intervention Modes 절 확장 (~30 lines) | feedback-loop 모드 신설 절 |
| `.claude/skills/moai/workflows/run.md` | 수정 | Phase 2.8a 부근 | feedback-loop 분기 로직 |

### 4.2 Secondary 영향 (테스트/문서)

- `internal/config/harness.go` (있을 시): YAML schema 확장
- `docs-site/content/{ko,en,ja,zh}/.../harness.md`: feedback-loop 모드 설명 추가

### 4.3 Templates (Template-First Rule 준수)

- `internal/template/templates/.moai/config/sections/harness.yaml`: 동일 변경
- `internal/template/templates/.claude/agents/moai/evaluator-active.md`: 동일 변경

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Iteration 1회 추가로 토큰 비용 증가 | High | Medium | acceptance에 token-cost baseline 측정 단계 명시 |
| feedback-loop 도입으로 thorough 모드와 의미 충돌 | Low | High | thorough = 2단계 contract+evaluation, standard = 단순 iterative — 명확히 분리 |
| stagnation 감지가 너무 보수적 → 즉시 escalation 폭주 | Medium | Medium | improvement_threshold 0.10을 보수적 출발점으로, 측정 후 조정 |
| 기존 standard SPEC들이 갑자기 두 배 평가 시간 → DX 저하 | Medium | High | feature flag (`enable_feedback_loop`) 도입, 점진 롤아웃 |

### 5.2 Assumptions

- evaluator-active 본문에 명시한 protocol을 orchestrator/skill이 일관되게 따름 (LSP 검증으로는 보장 불가, 텍스트 수준 의도 전달)
- 첫 iteration의 평가 결과가 "actionable feedback"으로 변환 가능 (점수만이 아닌 구체 피드백)
- improvement_threshold 0.10은 thorough mode의 0.05보다 큼 → standard는 더 관대한 정체 판정 (의도)

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| Token cost per evaluation | 현재 standard 평가 5회 평균 토큰 | < +25% (acceptance) |
| Pass rate after iteration | 첫 iter pass + 2nd iter pass / 전체 | > 첫 iter pass 단독 |
| Stagnation false-positive rate | 정체 판정 / 전체 정체 의심 | < 20% |
| Average iteration count | feedback-loop 켰을 때 평균 | <= 1.5 (1회 충분 표준) |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| frontmatter `evaluator_mode` 커스텀 필드 추가 | ❌ | claude-code-guide: Claude Code agent schema 위반 |
| harness.yaml에 `evaluator_mode` config 추가 | ✅ | config 영역은 자유롭게 확장 가능 |
| design.yaml의 GAN Loop을 일반 도메인으로 일반화 | ❌ | design 한정 의도가 강함, 의미 충돌 우려 |
| thorough 레벨로 모든 SPEC 강제 승격 | ❌ | 토큰 비용 4배+ 폭증, DX 악화 |
| 새 harness 레벨 (`enhanced-standard`) 신설 | ❌ | 레벨 수 증가 = 복잡도 증가, 기존 standard 업그레이드가 더 단순 |

---

## 8. 참고 SPEC

- SPEC-V3R3-HARNESS-001: harness 시스템 기반 SPEC (선행)
- SPEC-V3R3-HARNESS-LEARNING-001: harness 진화 메커니즘
- SPEC-AGENCY-ABSORB-001 §11: GAN Loop Contract 원형 (design 한정)
- SPEC-LOOP-TERM-001 (Wave 2 sibling): termination 조건 표준화 (본 SPEC과 짝)

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: `feedback-loop` 모드를 standard에 default ON으로 둘 것인가, opt-in으로 둘 것인가? → plan.md에서 결정
- OQ2: max_iterations 기본값 2가 적절한가, 3까지 허용할 것인가? → acceptance baseline 측정 후 결정
- OQ3: improvement_threshold 0.10이 표준 SPEC에 적절한가, 더 작게 설정할 것인가?
- OQ4: feedback 자체를 evaluator-active가 생성할 때 LLM call 추가 비용은? → acceptance에 포함

---

End of research.md (SPEC-EVAL-LOOP-001).
