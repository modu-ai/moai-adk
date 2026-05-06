---
id: SPEC-LOOP-TERM-001
status: draft
version: "0.1.0"
priority: High
labels: [workflow, termination, iteration, loop, escalation, wave-2, tier-1]
issue_number: null
scope: [loop.md, fix.md, coverage.md, e2e.md, iteration-termination.md]
blockedBy: []
dependents: []
related_specs: [SPEC-EVAL-LOOP-001]
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 2
tier: 1
---

# SPEC-LOOP-TERM-001: Iterative Workflow Termination Conditions Standardization

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 2 / Tier 1. Anthropic "first-class termination conditions" 권고를 일반 워크플로우 4개 (loop / fix / coverage / e2e)에 표준 schema 적용.

---

## 1. Goal (목적)

`.moai/config/sections/design.yaml §11`이 design 도메인 한정으로 잘 정의한 GAN Loop Contract (max_iterations + improvement_threshold + escalation)를 일반 iterative workflow 4개 (`/moai loop`, `/moai fix`, `/moai coverage`, `/moai e2e`)에 표준화된 형태로 확산한다. 이로써 Anthropic이 명시한 "Reactive loops are a behavioral problem requiring first-class termination conditions" 원칙을 준수한다.

### 1.1 배경

- Anthropic blog "Multi-Agent Coordination Patterns": "The hardest part of building agentic loops is not getting them to start — it's getting them to stop."
- 현재 design 도메인은 `§11` GAN Loop Contract로 termination 정의됨 (모범).
- 일반 workflow 4개는 termination schema가 약하거나 부재 → context window 소진 또는 무한 변형 위험.

### 1.2 비목표 (Non-Goals)

- `design.yaml §11` GAN Loop Contract의 의미 변경 (FROZEN 영역, 변경 금지)
- workflow 단계별 detail 변경 (본 SPEC은 termination policy 한정)
- Hook-level enforcement (hooks는 termination logic에 부적합)
- agent의 내부 retry 정책 (agent specific, 본 SPEC scope 외)

---

## 2. Scope (범위)

### 2.1 In Scope

- 신규 canonical reference: `.claude/rules/moai/workflow/iteration-termination.md`
  - max_iterations, stagnation, escalation, state_file 4-field 표준 schema
- 4개 workflow 의무 상속:
  - `loop.md` 강화
  - `fix.md` 강화
  - `coverage.md` 신규 적용
  - `e2e.md` 신규 적용
- state persistence: `.moai/state/<workflow>/<run_id>.json`
- escalation 경로: AskUserQuestion via orchestrator
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- design domain `§11` GAN Loop Contract 변경 (FROZEN)
- 신규 termination 종류 (예: time-based timeout, memory-based limit) — 본 SPEC은 iteration count + score-based 한정
- workflow별 max_iterations 차별 (default는 단일 schema, override는 individual workflow에서)
- Hook-level enforcement
- agent 내부 retry 정책 변경
- evaluator-active의 평가 동작 변경 (SPEC-EVAL-LOOP-001 영역)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.1.111+
- 영향 디렉터리: `.claude/skills/moai/workflows/`, `.claude/rules/moai/workflow/`, `.moai/state/`
- 템플릿 동기화: `internal/template/templates/`

---

## 4. Assumptions (가정)

- A1: 4개 workflow 모두 동일 schema 적용 가능 (구조적 차이 흡수 가능)
- A2: state file은 SPEC-ID 또는 run_id (UUID) 기반으로 격리 가능
- A3: escalation은 orchestrator만 AskUserQuestion 호출 가능 (subagent 직접 호출 금지)
- A4: design `§11`과의 키 이름 차별화로 의미 충돌 회피 (design은 `gan_loop`, 일반은 `termination`)
- A5: workflow마다 default max_iterations 다르게 설정 가능 (loop=5, fix=3, coverage=3, e2e=2)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-LT-001**: THE NEW CANONICAL REFERENCE `.claude/rules/moai/workflow/iteration-termination.md` SHALL define a standard termination schema with the following fields (single source of truth — all field names use dot notation):

  ```yaml
  termination:
    max_iterations: integer            # required; upper bound on iterations
    stagnation:
      detect_after: integer            # iteration index at which delta measurement begins (default 2)
      improvement_min: float           # minimum acceptable score delta in absolute units (default 0.05)
      consecutive: integer             # number of consecutive stagnating iterations required to trigger escalation (default 1)
    escalation:
      target: enum [user, log, abort]  # destination for escalation; default "user"
      reason_required: boolean         # whether the escalation report MUST include a structured rationale (default true)
    state_file: string                 # path template, default ".moai/state/{workflow}/{run_id}.json"
  ```

  All references to these fields throughout this SPEC, plan.md, and acceptance.md MUST use this dot notation (e.g., `stagnation.consecutive`, NOT `stagnation_consecutive`).
- **REQ-LT-002**: THE TERMINATION SCHEMA SHALL be inheritable by any moai workflow that performs iterative computation.
- **REQ-LT-003**: THE WORKFLOWS LOOP, FIX, COVERAGE, AND E2E SHALL all comply with the standardized termination schema.
- **REQ-LT-004**: THE STATE FILE SHALL be persisted to `.moai/state/<workflow>/<run_id>.json` for resume capability.
- **REQ-LT-005**: THE ESCALATION TARGET SHALL be `user` via AskUserQuestion invoked by the orchestrator (subagents MUST NOT invoke AskUserQuestion).
- **REQ-LT-006**: THE TERMINATION SCHEMA KEYS SHALL NOT collide with `design.yaml §11` GAN Loop Contract keys (use `termination.*` prefix vs `gan_loop.*`).

### 5.2 Event-Driven Requirements

- **REQ-LT-007**: WHEN any moai iterative workflow enters its iteration phase, THE WORKFLOW SHALL declare a termination schema in its skill body or first execution step.
- **REQ-LT-008**: WHEN iteration count reaches `max_iterations` without success, THE WORKFLOW SHALL escalate via AskUserQuestion with options: `Continue with adjustments`, `Abort and preserve state`, `Restart with new criteria`.
- **REQ-LT-009**: WHEN the orchestrator receives an escalation request from a subagent, THE ORCHESTRATOR SHALL invoke `ToolSearch(query: "select:AskUserQuestion")` before opening the AskUserQuestion round.
- **REQ-LT-010**: WHEN the user selects an option in the escalation AskUserQuestion, THE WORKFLOW SHALL act on that option deterministically (no further LLM judgment on the choice).

### 5.3 State-Driven Requirements

- **REQ-LT-011**: WHILE workflow is iterating, THE STATE SHALL be persisted to `.moai/state/<workflow>/<run_id>.json` after each iteration completes.
- **REQ-LT-012**: WHILE iteration count <= `max_iterations` AND stagnation NOT detected, THE WORKFLOW SHALL continue to the next iteration.

### 5.4 Conditional Requirements

- **REQ-LT-013**: WHERE iteration count is greater than or equal to `stagnation.detect_after` (default 2), THE WORKFLOW SHALL compute the score delta against the previous iteration.
- **REQ-LT-014**: IF the score delta is less than `stagnation.improvement_min` (default 0.05) for `stagnation.consecutive` (default 1) consecutive iterations, THEN THE WORKFLOW SHALL be flagged as stagnating and escalation SHALL be triggered.
- **REQ-LT-015**: WHERE a state file from a previous run exists at `.moai/state/<workflow>/<run_id>.json`, THE WORKFLOW SHALL offer resume option via AskUserQuestion before starting fresh.
- **REQ-LT-016**: IF `escalation.reason_required = true`, THEN THE ESCALATION REPORT SHALL include a structured rationale section explaining why escalation was triggered.
- **REQ-LT-017**: WHERE workflow is `loop`, THE DEFAULT max_iterations SHALL be 5.
- **REQ-LT-018**: WHERE workflow is `fix`, THE DEFAULT max_iterations SHALL be 3.
- **REQ-LT-019**: WHERE workflow is `coverage`, THE DEFAULT max_iterations SHALL be 3.
- **REQ-LT-020**: WHERE workflow is `e2e`, THE DEFAULT max_iterations SHALL be 2.

### 5.5 Unwanted (Negative) Requirements

- **REQ-LT-021**: THE WORKFLOWS SHALL NOT iterate beyond `max_iterations` regardless of progress.
- **REQ-LT-022**: SUBAGENTS SHALL NOT invoke AskUserQuestion directly for escalation; they MUST return a blocker report to the orchestrator.
- **REQ-LT-023**: THE TERMINATION SCHEMA SHALL NOT silently override `design.yaml §11` GAN Loop Contract values for design domain workflows.

---

## 6. Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| canonical reference adopted by all 4 workflows | grep for `iteration-termination.md` | 4/4 |
| state file persistence | E2E test of one iteration cycle | file exists with valid JSON |
| escalation reaches user via AskUserQuestion | E2E test triggering max_iterations | AskUserQuestion observed |
| design `§11` regression | regression tests on design workflow | zero behavior change |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: design.yaml `§11` GAN Loop Contract는 FROZEN — 변경 불가
- C2: subagent의 AskUserQuestion 직접 호출 금지 (HARD rule)
- C3: state file는 atomic write (race 방지)
- C4: termination schema의 default 값은 보수적 (max_iter 작게, stagnation 임계 명확)

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
- `related_specs: [SPEC-EVAL-LOOP-001]` — Shares iterative-evaluation problem space; not blocked by or blocking this SPEC.

End of spec.md (SPEC-LOOP-TERM-001 v0.1.0).
