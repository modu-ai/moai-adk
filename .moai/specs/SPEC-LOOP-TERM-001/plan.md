---
id: SPEC-LOOP-TERM-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-LOOP-TERM-001

## 1. Overview

신규 canonical reference (`.claude/rules/moai/workflow/iteration-termination.md`)에 표준 termination schema를 정의하고, 4개 iterative workflow (`loop`, `fix`, `coverage`, `e2e`)가 이를 의무 상속하도록 강화한다.

## 2. Approach Summary

**전략**: Single Source of Truth + Mandatory Inheritance.

1. canonical reference 신규 작성 (단일 진실 출처)
2. 4개 workflow에서 reference 의무 인용 + workflow-specific override (max_iterations만)
3. state file schema 통일
4. escalation 경로 표준화 (AskUserQuestion via orchestrator)

## 3. Milestones

### M0 — Pre-flight Verification (Priority: Critical)

- [ ] `loop.md`, `fix.md`, `coverage.md`, `e2e.md` 현재 termination 관련 절 verbatim 캡처
- [ ] `design.yaml §11` GAN Loop Contract 상수 verbatim 캡처 (regression baseline)
- [ ] 기존 `.moai/state/` 사용 패턴 조사

**Exit Criteria**: 4 workflow + 1 design.yaml 베이스라인 기록

### M1 — Canonical Reference Creation (Priority: Critical)

- [ ] 신규 파일 작성: `.claude/rules/moai/workflow/iteration-termination.md`
  - 표준 schema 정의 (max_iterations, stagnation, escalation, state_file)
  - workflow별 default 값 표 (loop=5, fix=3, coverage=3, e2e=2)
  - escalation AskUserQuestion options 명시 (Continue / Adjust / Abort)
  - state file schema (JSON)
  - design `§11`과의 차이점 명시 절 추가 (혼동 방지)
- [ ] Template 동기화: `internal/template/templates/.claude/rules/moai/workflow/iteration-termination.md`

**Exit Criteria**: canonical reference 파일 존재 + Template-First sync

### M2 — Loop Workflow Enhancement (Priority: High)

- [ ] `.claude/skills/moai/workflows/loop.md` 강화:
  - 첫 phase에 "termination schema inheritance" 절 추가
  - canonical reference 참조 명시
  - max_iterations 5 (default 적용)
  - state file 경로: `.moai/state/loop/<run_id>.json`
- [ ] Template 동기화

**Exit Criteria**: loop.md에 termination schema 의무 상속 명시

### M3 — Fix Workflow Enhancement (Priority: High)

- [ ] `.claude/skills/moai/workflows/fix.md` 강화:
  - canonical reference 참조 명시
  - max_iterations 3 (default override)
  - 기존 retry 로직과 통합 (retry는 individual operation, iteration은 fix cycle 전체)
  - state file 경로: `.moai/state/fix/<run_id>.json`
- [ ] Template 동기화

**Exit Criteria**: fix.md에 termination schema 의무 상속 명시

### M4 — Coverage and E2E Workflow Application (Priority: High)

- [ ] `.claude/skills/moai/workflows/coverage.md` 신규 적용:
  - termination schema 절 신설
  - max_iterations 3
  - improvement: coverage% delta가 1.0% 미만일 때 stagnation
  - state file 경로: `.moai/state/coverage/<run_id>.json`
- [ ] `.claude/skills/moai/workflows/e2e.md` 신규 적용:
  - termination schema 절 신설
  - max_iterations 2
  - improvement: 새 통과 시나리오 수가 0인 iteration이 stagnation
  - state file 경로: `.moai/state/e2e/<run_id>.json`
- [ ] Template 동기화

**Exit Criteria**: coverage.md, e2e.md에 termination schema 명시

### M5 — Escalation Standardization (Priority: Medium)

- [ ] AskUserQuestion options 표준 형식 확정:
  - Option 1: "현재 접근으로 계속 (조정 후)" (권장 — 최고빈도 선택)
  - Option 2: "중단하고 상태 보존"
  - Option 3: "새 기준으로 재시작"
  - Option 4: "Other" (자동 추가)
- [ ] orchestrator routing: ToolSearch + AskUserQuestion 호출 보장
- [ ] subagent → orchestrator escalation 보고서 형식 표준화

**Exit Criteria**: escalation 경로 4개 workflow 모두에서 일관

### M6 — State File Schema (Priority: Medium)

- [ ] JSON schema 정의:
  ```json
  {
    "workflow": "loop|fix|coverage|e2e",
    "run_id": "uuid",
    "spec_id": "SPEC-XXX (optional)",
    "iterations": [{"n": 1, "score": 0.62, "delta": null, "summary": "..."}],
    "status": "in-progress|completed|escalated|aborted",
    "started_at": "ISO-8601",
    "last_updated_at": "ISO-8601"
  }
  ```
- [ ] resume scenario E2E test seed
- [ ] cleanup 정책: completed 상태 30일 후 archive, escalated/aborted 보존

**Exit Criteria**: state file schema 문서화 + 4 workflow 모두 동일 schema 사용

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md의 Given-When-Then 시나리오 9개 모두 PASS
- [ ] design `§11` regression: zero behavior change verified
- [ ] plan-auditor 검증 PASS
- [ ] Template-First sync: clean diff

**Exit Criteria**: acceptance.md PASS + plan-auditor PASS

## 4. Technical Approach

### 4.1 Canonical Reference Schema

```yaml
# .claude/rules/moai/workflow/iteration-termination.md content (excerpt)
termination:
  max_iterations: 5            # default; workflows may override
  stagnation:
    detect_after: 2            # iteration index to start measuring delta
    improvement_min: 0.05      # below this delta = stagnating
    consecutive: 1             # consecutive stagnating iterations to trigger escalation
  escalation:
    target: "user"
    via: "AskUserQuestion (orchestrator-only)"
    reason_required: true
    options:
      - "현재 접근으로 계속 (조정 후) (권장)"
      - "중단하고 상태 보존"
      - "새 기준으로 재시작"
  state_file: ".moai/state/{workflow}/{run_id}.json"
```

### 4.2 Workflow Override Pattern

```markdown
# loop.md (excerpt)
This workflow inherits termination policy from
`.claude/rules/moai/workflow/iteration-termination.md` with:
- max_iterations: 5
- state_file: `.moai/state/loop/{run_id}.json`
All other fields use the canonical default.
```

### 4.3 Escalation Sequence

```
[iteration n hits stagnation/max_iter]
  ↓
subagent: returns blocker report to orchestrator
  ↓
orchestrator: ToolSearch(query: "select:AskUserQuestion")
  ↓
orchestrator: AskUserQuestion with 3 standardized options + Other
  ↓
user: selects option
  ↓
orchestrator: routes to next action deterministically
```

## 5. Risks and Mitigations

| Risk | P | I | Mitigation |
|------|---|---|-----------|
| 4 workflow의 inconsistent 적용 | High | Medium | canonical reference를 단일 진실 출처로, individual override 최소화 |
| design `§11`과 의미 충돌 | Low | High | 키 prefix 분리 (termination.* vs gan_loop.*) + design.md에 차이점 절 명시 |
| state file race condition | Medium | Medium | run_id (UUID) 격리, atomic write |
| AskUserQuestion deferred tool 미선로드 | High | Medium | escalation 절차에 ToolSearch 의무 명시 |
| 기존 workflow 사용자 워크플로우 단절 | Medium | Medium | feature flag (`enforce_termination_schema: false`) 도입 가능, 단 default true |
| coverage workflow의 stagnation 측정 (% delta 0.01) 너무 보수적 | Medium | Low | 사용 후 조정 (config 변경만으로 가능) |

## 6. Dependencies

- 선행 SPEC: 없음
- 동반 SPEC: SPEC-EVAL-LOOP-001 (evaluator-active의 standard feedback-loop이 본 schema 활용 가능)
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1**: 키 이름 design `§11`과 동일 vs 다르게 → ✅ 다르게 (`termination.*` vs `gan_loop.*`)
- **OQ2**: state file 위치 통합 vs 격리 → ✅ workflow별 격리 (`.moai/state/<workflow>/`)
- **OQ3**: workflow별 max_iterations 차별화 → ✅ 차별화 (loop=5, fix=3, coverage=3, e2e=2)
- **OQ4**: escalation options 표준화 → ✅ 3-option 표준 + Other (자동)

## 8. Rollout Plan

1. M1 canonical reference 작성 후 dogfooding (본 SPEC 자체에 termination 적용)
2. M2-M4 4개 workflow 점진 적용 (one-at-a-time, regression 확인)
3. M5-M6 escalation/state 통일
4. CHANGELOG entry + minor release

End of plan.md.
