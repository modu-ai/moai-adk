# Research — SPEC-LOOP-TERM-001 (Termination Conditions Standardization)

**SPEC**: SPEC-LOOP-TERM-001
**Wave**: 2 / Tier 1
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Multi-Agent Coordination Patterns":

> "Reactive loops are a behavioral problem requiring first-class termination conditions."

> "The hardest part of building agentic loops is not getting them to start — it's getting them to stop."

> "An agent that re-prompts itself indefinitely is not exploring the solution space; it is producing variations of the same misunderstanding."

Anthropic blog "Common Workflow Patterns":

> "Set clear stopping criteria before you start iterating. Stopping criteria are the contract between you and the loop."

> "Anti-pattern: Running iterations without defined stopping criteria or maximum counts. The loop will continue until the context window or budget runs out, neither of which is a quality signal."

> "Termination conditions should be testable from outside the loop — counters, thresholds, explicit pass/fail signals — not subjective judgments embedded inside the loop."

### 1.2 검증 (claude-code-guide 결과)

claude-code-guide:
- **호환성**: ✅ 이미 부분 구현 (design.yaml `§11` GAN Loop Contract). 확산만 필요.
- **권고 채택**: ACCEPT — 단, 일반 워크플로우용 standard schema 신설.

---

## 2. 현재 상태 (As-Is)

### 2.1 Design GAN Loop Contract (잘 구현됨, design 한정)

`.claude/rules/moai/design/constitution.md §11` Frozen 절:

```
Loop Mechanics:
- max_iterations: 5
- pass_threshold: 0.75
- escalation_after: 3
- improvement_threshold: 0.05
- strict_mode: true → individual must-pass criteria
```

→ design 도메인은 모범적.

### 2.2 일반 Workflow termination 조건 (부재 또는 약함)

| Workflow | Termination 명시 | 비고 |
|----------|----------------|------|
| `loop.md` | ⚠️ 약함 | iteration max 명시 있으나 stagnation/escalation 부재 |
| `fix.md` | ⚠️ 약함 | 재시도 횟수만 명시, improvement 측정 없음 |
| `coverage.md` | ❌ 없음 | iteration 자체가 명시 안 됨 |
| `e2e.md` | ❌ 없음 | retry 정책만 |
| `review.md` | N/A | 단일 패스 (loop 아님) |

### 2.3 fix.md 상세 (loop 5분의 1)

`.claude/skills/moai/workflows/fix.md` 분석:
- max retries 언급 있음 (3회 추정)
- failure pattern detection 명시
- 그러나 improvement_threshold 없음, stagnation 정의 없음
- escalation 경로: 명확하지 않음

### 2.4 loop.md 상세

`.claude/skills/moai/workflows/loop.md`:
- iteration limit 추정 5
- per-iteration progress comparison 부재
- AskUserQuestion으로의 escalation 경로 미명시

---

## 3. 격차 분석

| 영역 | As-Is | To-Be | 격차 |
|------|-------|-------|------|
| 표준 termination schema 정의 | 없음 | 1 schema (max_iter + stagnation + escalation) | 신설 필요 |
| `/moai loop` 적용 | 약함 | 표준 schema 적용 | 강화 필요 |
| `/moai fix` 적용 | 약함 | 표준 schema 적용 | 강화 필요 |
| `/moai coverage` 적용 | 없음 | 표준 schema 적용 | 신규 추가 |
| `/moai e2e` 적용 | 없음 | 표준 schema 적용 | 신규 추가 |
| state persistence (resume) | 없음 | `.moai/state/<workflow>/` | 신설 |
| escalation via AskUserQuestion | 없음 | 명시 | 신설 |

---

## 4. 표준 Schema 설계 초안

```yaml
termination:
  max_iterations: 5          # 절대 한계
  stagnation:
    detect_after: 2          # 정체 감지 시작 iteration
    improvement_min: 0.05    # 최소 개선 임계값
  escalation:
    target: "user"           # AskUserQuestion 호출 대상
    reason_required: true    # 에스컬레이션 사유 명시 필수
  state_file: ".moai/state/{workflow}/{run_id}.json"  # resume 지원
```

---

## 5. 코드베이스 분석 (Affected Files)

### 5.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/skills/moai/workflows/loop.md` | 강화 | termination schema 적용 |
| `.claude/skills/moai/workflows/fix.md` | 강화 | termination schema 적용 |
| `.claude/skills/moai/workflows/coverage.md` | 추가 | termination schema 신규 |
| `.claude/skills/moai/workflows/e2e.md` | 추가 | termination schema 신규 |
| `.claude/rules/moai/workflow/iteration-termination.md` | 신규 | 표준 schema 정의 (canonical reference) |

### 5.2 Templates (Template-First)

- `internal/template/templates/.claude/skills/moai/workflows/{loop,fix,coverage,e2e}.md`: 동일 변경
- `internal/template/templates/.claude/rules/moai/workflow/iteration-termination.md`: 신규

---

## 6. 위험 및 가정

### 6.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| 4개 workflow에 schema 적용 → 일관성 낮은 구현 위험 | High | Medium | canonical reference 1개 + workflow별 의무 상속 |
| design GAN Loop Contract와 키 충돌 | Low | High | schema 키 prefix 분리 또는 design은 §11 그대로 유지 |
| resume state file race | Medium | Medium | run_id (UUID) 기반 격리, single-active 가정 |
| AskUserQuestion 의존 (deferred tool preload) | Medium | Medium | escalation 시점 ToolSearch 자동 트리거 명시 |

### 6.2 Assumptions

- A1: 4개 workflow 모두 동일 schema 적용 가능 (구조적 차이 무시할 수 있음)
- A2: state file은 SPEC-ID 또는 run_id 단위로 격리 가능
- A3: AskUserQuestion 경유 escalation은 orchestrator만 가능 (subagent 직접 호출 금지)

---

## 7. 대안 검토

| 대안 | 채택 | 이유 |
|------|-----|------|
| design.yaml `§11`을 표준으로 직접 참조 | ❌ | design 한정 의도 강함, 일반화 시 의미 충돌 |
| workflow별 개별 schema (현 상태 유지) | ❌ | 일관성 결여, 학습 비용 증가 |
| 새 canonical reference 파일 + 4개 workflow 의무 상속 | ✅ | DRY, 한 곳에서 정책 관리 |
| Hooks-level termination enforcement | ❌ | hook은 hard-coded 동작이라 workflow마다 다른 schema에 부적합 |

---

## 8. 참고 SPEC

- SPEC-EVAL-LOOP-001 (Wave 2 sibling): standard harness feedback-loop과 짝
- SPEC-V3R3-HARNESS-001: harness routing
- SPEC-AGENCY-ABSORB-001 §11: design GAN Loop (canonical 출처)

---

## 9. Open Questions

- OQ1: standard schema의 키 이름은 design `§11`과 같이 가야 하는가, 다르게 가야 하는가? → plan에서 결정
- OQ2: state file 위치를 `.moai/state/iteration-loops/` 단일 디렉토리로 통합할 것인가, workflow별로 격리할 것인가?
- OQ3: 각 workflow의 max_iterations 기본값을 통일할 것인가, workflow별로 다르게 할 것인가? (loop=5, fix=3, coverage=3, e2e=2 권장)
- OQ4: escalation 시 AskUserQuestion options를 표준화할 것인가? (Continue / Adjust / Abort)

---

End of research.md (SPEC-LOOP-TERM-001).
