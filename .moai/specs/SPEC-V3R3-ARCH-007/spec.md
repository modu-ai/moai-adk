---
id: SPEC-V3R3-ARCH-007
title: Token Circuit Breaker — runtime.yaml + Go runtime per-agent budget enforcement
version: "1.0.0"
status: implemented
created: 2026-04-25
updated: 2026-04-26
author: manager-spec
priority: P1 High
phase: "v3.0.0 R3 — Phase A — Runtime Safety Net"
module: ".moai/config/sections/runtime.yaml, internal/runtime/budget.go, internal/template/templates/.moai/config/sections/runtime.yaml"
dependencies: []
breaking: false
bc_id: [BC-V3R3-006]
lifecycle: spec-anchored
tags: "token-budget, circuit-breaker, runtime, per-agent, stall-detection, progress-persistence, v3r3, phase-a, safety-net"
related_theme: "Phase A — Iteration 4 Safety Net"
released_in: v2.15.0
---

# SPEC-V3R3-ARCH-007: Token Circuit Breaker

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 1.0.0   | 2026-04-25 | manager-spec | Initial draft. Phase A P1 — runtime.yaml schema 신설 + per-agent budget tracker (warning-first, hard-fail in P5). |

---

## 1. Goal (목적)

MoAI-ADK의 Iteration 4 (Wave 단위 large SPEC 실행) 환경에서 **token budget 초과로 인한 stream stall을 사전 감지하고 자동 복구**하는 Token Circuit Breaker를 구축한다. `.moai/config/sections/runtime.yaml`을 신설하여 per-agent budget을 정의하고, `internal/runtime/budget.go` Go 모듈로 실시간 token tracking과 stall detection을 구현한다. 75% 도달 시 progress.md 자동 저장 + resume message 자동 생성 (warning-first). 90% 또는 stall 감지 시 graceful clear 권고. 본 SPEC은 P5(향후 SPEC)에서 hard-fail로 전환될 예정이며 P3(현재) 시점에는 warning-only.

### 1.1 배경

- 2026-04-25 SPEC-V3R2-WF-001 monolithic delegation 시 Anthropic SSE `stream_idle_partial` stall 발생 (`feedback_large_spec_wave_split.md` 참조).
- 현재 token budget은 `.claude/rules/moai/workflow/file-reading-optimization.md`와 `context-window-management.md`에 가이드라인만 존재, 실시간 enforcement 부재.
- 75% 도달 시 자동 progress.md 저장 + resume message 생성하는 메커니즘 부재 → 사용자가 수동 추적 필요.
- per-agent token budget 부재 → manager-strategy 같은 reasoning-heavy agent와 manager-git 같은 lightweight agent를 동일 기준으로 다룸.

### 1.2 비목표 (Non-Goals)

- /clear 자동 트리거 금지 (HARD constraint per `MEMORY.md` Hard Constraints — 사용자 수동 실행 필수)
- token estimation의 100% 정확도 보장 금지 (heuristic 기반)
- runtime.yaml schema 변경을 다른 config section에 강제 금지
- Go runtime 외 다른 언어로 budget tracker 구현 금지
- evaluator-active scoring에 token budget 관련 항목 추가 금지 (별도 SPEC)
- Claude API 호출 차단 금지 (hard-fail은 P5에서 구현, 본 SPEC은 warning-only)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `.moai/config/sections/runtime.yaml` 신설 (master plan §1.5 schema 준수)
- **Owns**: `internal/template/templates/.moai/config/sections/runtime.yaml` template 동기화
- **Owns**: `internal/runtime/budget.go` Go 모듈 신설 (per-agent token tracker, stall detection)
- **Owns**: SessionStart 훅에서 runtime.yaml 로드 + budget context 초기화
- **Owns**: 75% threshold 도달 시 `.moai/specs/<SPEC-ID>/progress.md` 자동 저장
- **Owns**: resume message 자동 생성 (`.claude/rules/moai/workflow/context-window-management.md` §Resume message format 패턴)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- /clear 명령 자동 실행
- Claude API 호출 가로채기 / proxy 구현
- token estimation을 위한 새로운 API 호출
- runtime.yaml에 token budget 외 필드 추가 (e.g., memory limit, file count)
- TUI dashboard / 시각화
- 다른 config section의 schema 변경
- P5 hard-fail 로직 (별도 SPEC 필요)

---

## 3. Environment

- Go 1.24+ (`internal/runtime/` 모듈)
- File system writable: `.moai/config/sections/`, `.moai/specs/`, `internal/runtime/`, `internal/template/templates/.moai/config/sections/`
- SessionStart 훅 실행 권한
- `make build` 가능 환경

## 4. Assumptions

- 현재 `internal/config/` 모듈이 YAML config 로드 인프라 제공
- SessionStart 훅이 Go bin에서 호출됨 (`moai hook session-start`)
- progress.md 파일은 `.moai/specs/<SPEC-ID>/progress.md` 표준 위치
- token estimation은 input/output character count 기반 heuristic으로 충분

## 5. Requirements (EARS)

### REQ-ARCH007-001 (Ubiquitous)

The system **shall** provide a `.moai/config/sections/runtime.yaml` configuration file with the following keys:
  - `pre_clear_threshold: 0.75` (warning trigger)
  - `hard_clear_threshold: 0.90` (mandatory clear suggestion)
  - `per_agent_budget` (default 30000, manager-strategy 60000, expert-* 40000, evaluator-* 20000)
  - `circuit_breaker.stall_detection_seconds: 60`
  - `circuit_breaker.retry_max: 3`
  - `circuit_breaker.fallback: split_into_waves`

### REQ-ARCH007-002 (Ubiquitous)

The `internal/runtime/budget.go` module **shall** expose a `Tracker` type with the following methods:
  - `RecordCall(agentName string, tokensIn int, tokensOut int)`
  - `Usage(agentName string) (current int, budget int, ratio float64)`
  - `IsApproachingLimit(agentName string) bool` (returns true at >= 75% of budget)
  - `IsAtHardLimit(agentName string) bool` (returns true at >= 90% of budget)
  - `DetectStall(agentName string) bool` (returns true if no progress for `stall_detection_seconds`)

### REQ-ARCH007-003 (Ubiquitous)

The `internal/template/templates/.moai/config/sections/runtime.yaml` **shall** be identical to `.moai/config/sections/runtime.yaml` content (template-local sync).

### REQ-ARCH007-004 (Event-Driven)

**When** the SessionStart hook executes, the system **shall** load `runtime.yaml` and initialize a per-session `Tracker` instance.

### REQ-ARCH007-005 (Event-Driven)

**When** the cumulative token usage for a SPEC reaches 75% of total context budget, the system **shall**:
  1. Persist current SPEC progress to `.moai/specs/<SPEC-ID>/progress.md`
  2. Emit a structured resume message to the orchestrator output (paste-ready format per `.claude/rules/moai/workflow/context-window-management.md` §Resume message format)
  3. Log a warning to stderr (no API call interruption)

### REQ-ARCH007-006 (Event-Driven)

**When** a stall is detected (no progress for `stall_detection_seconds: 60`), the system **shall**:
  1. Log a warning with the stalled agent name
  2. Increment retry counter; on reaching `retry_max: 3`, emit a fallback recommendation (`split_into_waves`)
  3. Do NOT auto-clear or auto-restart

### REQ-ARCH007-007 (State-Driven)

**While** an agent is invoked with `Agent()`, the system **shall** track its token usage against `per_agent_budget` for that agent's role profile.

### REQ-ARCH007-008 (Event-Driven)

**When** a per-agent budget is exceeded (current >= budget), the system **shall** emit a warning naming the agent and budget value (warning-only in P3, hard-fail deferred to P5).

### REQ-ARCH007-009 (Unwanted)

The system **shall not** automatically execute `/clear` under any condition (HARD constraint per project memory).

### REQ-ARCH007-010 (Unwanted)

The Token Circuit Breaker **shall not** intercept or modify Claude API requests/responses.

### REQ-ARCH007-011 (Optional)

**Where** runtime.yaml is missing, the system **shall** use built-in default values matching the schema in REQ-ARCH007-001.

### REQ-ARCH007-012 (Complex / Unwanted)

**While** a SPEC is being executed via `/moai run`, **if** the cumulative token usage exceeds `hard_clear_threshold: 0.90`, **then** the system **shall** emit a strong recommendation to `/clear` and STOP issuing new `Agent()` invocations until the user manually clears.

---

## 6. Acceptance Criteria (요약)

전체 acceptance.md 참조. 핵심:

- AC-ARCH007-01: runtime.yaml 신설 (local + template), schema 일치
- AC-ARCH007-02: `internal/runtime/budget.go` Tracker 인터페이스 5 메서드 구현
- AC-ARCH007-03: SessionStart 훅에서 runtime.yaml 로드 확인
- AC-ARCH007-04: 75% 도달 시 progress.md 자동 저장 + resume message 출력
- AC-ARCH007-05: stall detection 60s + retry max 3
- AC-ARCH007-06: /clear 자동 트리거 부재 검증
- AC-ARCH007-07: BC-V3R3-006 warning-first 동작 검증

---

## 7. Constraints

- **C1**: HARD — /clear 자동 트리거 절대 금지 (`MEMORY.md` Hard Constraints)
- **C2**: HARD — Go binary는 hot-reload 없음. `internal/runtime/*.go` 변경 후 `make build && make install` + Claude Code 재시작 필요 (`MEMORY.md` Hard Constraints)
- **C3**: BC-V3R3-006 (warning-first): 본 SPEC 시점에는 모든 budget 초과는 warning. P5(향후 SPEC)에서 hard-fail로 전환 예정. 본 SPEC은 hard-fail 코드를 작성하지 않음.
- **C4**: 16-language neutrality — runtime.yaml은 언어 중립적 schema
- **C5**: token estimation은 heuristic OK (정확도 ±10% 허용)
- **C6**: progress.md 자동 저장은 SPEC 디렉터리 존재 시에만 동작 (디렉터리 부재 시 silent skip)

---

## 8. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| token estimation 부정확 | Medium | heuristic 기반 ±10% 허용; runtime.yaml에 추가 calibration 키 future SPEC에서 |
| stall detection false positive | Medium | retry_max 3으로 완충, fallback recommendation만 emit |
| progress.md 자동 저장 race condition | Medium | file lock 또는 atomic write 사용 |
| runtime.yaml schema 변경이 다른 config 영향 | Low | runtime.yaml은 새 section, 다른 config 무영향 |
| Go binary 미빌드 시 동작 안 함 | High | DoD에 `make build && make install` 명시 |
| BC-V3R3-006 warning이 오히려 signal noise | Low | log level 분리 (WARN vs INFO), 향후 P5 hard-fail로 강화 |

---

## 9. Dependencies

없음. (Phase A 독립 진행 가능)

후속:
- P5 (향후 SPEC): warning → hard-fail 전환
- 별도 SPEC: TUI dashboard, evaluator-active token criteria 추가

---

## 10. Traceability

| REQ ID | Acceptance Criteria | Source |
|--------|---------------------|--------|
| REQ-ARCH007-001 | AC-ARCH007-01 | Master plan §1.5 schema |
| REQ-ARCH007-002 | AC-ARCH007-02 | Tracker interface design |
| REQ-ARCH007-003 | AC-ARCH007-01 (template sync) | CLAUDE.local.md §2 Template-First Rule |
| REQ-ARCH007-004 | AC-ARCH007-03 | SessionStart hook integration |
| REQ-ARCH007-005 | AC-ARCH007-04 | context-window-management.md §75% behavior |
| REQ-ARCH007-006 | AC-ARCH007-05 | Stall detection contract |
| REQ-ARCH007-007 | AC-ARCH007-02 (per-agent tracking) | per-agent budget design |
| REQ-ARCH007-008 | AC-ARCH007-07 | warning-first policy (BC-V3R3-006) |
| REQ-ARCH007-009 | AC-ARCH007-06 | HARD: MEMORY.md /clear constraint |
| REQ-ARCH007-010 | (no acceptance, non-functional) | API non-interception policy |
| REQ-ARCH007-011 | AC-ARCH007-01 (default fallback) | Robustness |
| REQ-ARCH007-012 | AC-ARCH007-04 (90% behavior) | Hard threshold guidance |
