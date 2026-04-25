---
id: SPEC-WF-AUDIT-GATE-001
version: "1.0.0"
status: draft
created_at: 2026-04-25
updated_at: 2026-04-25
author: GOOS
priority: High
labels: [workflow, plan-audit, gate, governance, dogfood]
issue_number: null
depends_on: []
related_specs: []
---

# SPEC-WF-AUDIT-GATE-001: Plan→Run 전이 의무 감사 게이트

## HISTORY

| Version | Date       | Author | Description |
|---------|------------|--------|-------------|
| 1.0.0   | 2026-04-25 | GOOS   | Initial draft. v3R2 35-SPEC 일괄 점검에서 SPEC-SKILL-001(892B, commit-note 수준)이 plan을 통과해 archive까지 도달한 사고를 반영해 Plan→Run 전이 사이에 plan-auditor 의무 게이트를 신설. |

---

## 1. 개요

본 SPEC은 `/moai run` 호출 시 implementation을 시작하기 직전, plan 산출물(`spec.md`/`plan.md`/`acceptance.md`/(선택) `tasks.md`)을 `plan-auditor` agent로 **자동·의무 감사**하도록 워크플로우를 전면 개편한다. 감사 결과가 `PASS`일 때만 Run 단계가 진행되며, `FAIL` 시에는 감사 보고서를 사용자에게 제시하고 Run을 차단한다.

이는 manager-spec이 자기 산출물을 자기 검증하는 동일 컨텍스트 편향(self-confirmation bias)을 끊고, plan 단계의 누락·약한 EARS·미완성 산출물이 Run 단계에 진입해 구현 비용을 낭비하는 사고를 구조적으로 차단한다. `plan-auditor`는 이미 카탈로그(.claude/agents/moai/plan-auditor.md)에 존재하므로 본 SPEC의 작업은 **호출 지점(gate hook)**과 **차단 의사결정 로직**의 워크플로우 통합에 한정된다.

본 SPEC은 dogfood 원칙에 따라 자기 자신을 첫 번째 감사 대상으로 제출 가능한 수준으로 작성된다.

## 2. 배경 및 문제

### 2.1 사고 사례

2026-04-23 ~ 2026-04-24 사이 v3R2 redesign 35-SPEC 패키지 일괄 작성 직후, SPEC 디렉터리 전수 검사에서 다음 결함이 발견되었다.

- `SPEC-SKILL-001/spec.md` (이후 archive 처리됨): 파일 크기 **892 bytes**. EARS 요구사항 0건, acceptance criteria 0건, 사실상 commit note 수준이었다. 이 SPEC은 plan 단계를 통과해 manifest에 등재되었고, `/moai run` 호출이 발생했다면 Run 단계에서 즉시 좌초했을 것이다.
- 동일 패키지 내 다수 SPEC이 `created_at`/`updated_at` 대신 `created`/`updated` 별칭을 사용하거나, `labels` 배열 누락, version 미인용 등 schema 불일치를 보였다.

### 2.2 근본 원인 (Five Whys)

- 표면: 미완성 SPEC이 manifest에 등재됨.
- Why 1: manager-spec이 산출물 품질을 자체 검증함.
- Why 2: 동일 agent가 작성과 검증을 동시에 수행하여 self-confirmation bias 발생.
- Why 3: plan→run 전이 단계에 독립 감사 게이트가 없음.
- Why 4: `plan-auditor` agent가 카탈로그에 존재하지만 워크플로우에서 호출되지 않음(opt-in 상태).
- Root: Run 단계 진입 전 강제 게이트가 워크플로우 protocol 차원에서 정의되어 있지 않음.

### 2.3 사용자 의도 (verbatim)

> "plan 감사를 run 하기 전에 plan spec 결과물을 모두 전수 조사해서 문제가 없는지 다시 확인 후 진행을 하도록 workflow를 전면 개선을 하도록 하자"

### 2.4 dogfood 역설

본 SPEC 자체가 부실하면, 본 SPEC이 정의하는 게이트를 본 SPEC이 통과하지 못한다. 따라서 본 SPEC은 §3 EARS 7건, §4 명확한 in/out scope, §5 위험·완화, §6 HISTORY를 모두 갖추고 자기 검증 가능한 형태로 작성된다.

---

## 3. 요구사항 (EARS)

### `REQ-WAG-001` (Ubiquitous)

The system shall, on every `/moai run <SPEC-ID>` invocation, automatically invoke the `plan-auditor` subagent against the target SPEC's plan artifacts (`spec.md`, `plan.md`, `acceptance.md`, and `tasks.md` when present) **before** any implementation phase begins.

한국어 설명: `/moai run`이 호출되면 시스템은 어떠한 구현 행위(파일 작성, 테스트 작성, agent 위임)도 시작하기 전에 plan-auditor를 자동으로 호출해야 한다. 사용자 명시 호출이 없어도 기본 동작이다.

### `REQ-WAG-002` (Event-driven)

**When** the plan-audit verdict for `<SPEC-ID>` is `FAIL`, **then** the system shall block the Run phase, surface the audit report path to the user, and require an explicit user decision (revise SPEC / override / abort) **before** any further action.

한국어 설명: FAIL 결과 시 시스템은 Run 진행을 차단하고 보고서 경로(`<SPEC-ID>-review-<iteration>.md`)와 must-pass 실패 항목을 사용자에게 제시한다. 사용자 의사결정은 AskUserQuestion으로 수집한다(orchestrator 책임).

### `REQ-WAG-003` (Event-driven)

**When** the plan-audit verdict for `<SPEC-ID>` is `PASS`, **then** the system shall persist the audit verdict (verdict, report path, audit timestamp, plan-auditor version) into the SPEC's `progress.md` (or create the file if absent) and proceed to the Run phase.

한국어 설명: PASS 시 verdict, 보고서 경로, ISO-8601 timestamp, plan-auditor identifier를 progress.md에 추가 기록(append)하고 Run을 진행한다. 캐싱 검사는 `REQ-WAG-006` 우회 정책과 별개로 본 항에 따른다.

### `REQ-WAG-004` (Ubiquitous)

The system shall persist every plan-audit result (PASS, FAIL, or `INCONCLUSIVE`) as a Markdown file at `.moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md`, in addition to the iteration-scoped report at `.moai/reports/plan-audit/<SPEC-ID>-review-<iteration>.md` already produced by `plan-auditor`.

한국어 설명: plan-auditor가 자체적으로 iteration 단위 보고서를 쓰는 것과 별개로, gate 호출 시점마다 일자 단위 사본(date-stamped copy) 1건을 보존한다. 동일 일자 동일 SPEC 다회 실행 시 파일에 audit run을 append한다.

### `REQ-WAG-005` (State-driven)

**While** `--team` mode is active for `/moai run`, the system shall apply the same audit gate **before** any teammate (role_profiles: implementer, tester, designer, reviewer) is spawned via `Agent(subagent_type: "general-purpose")` or `TeamCreate`.

한국어 설명: 팀 모드라도 게이트는 동일하게 작동한다. teammate spawn 이전에 게이트가 통과되어야 하며, 게이트 자체는 main session에서 단일 plan-auditor로 실행한다(teammate별 중복 호출 금지).

### `REQ-WAG-006` (Optional Feature)

**Where** the user passes the `--skip-audit` flag (or sets `MOAI_SKIP_PLAN_AUDIT=1` env), the system shall record the bypass in `.moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md` with `verdict: BYPASSED`, the bypass timestamp (ISO-8601), the user identifier (`.moai/config/sections/user.yaml#user.name`), and the bypass rationale (prompted via AskUserQuestion when interactive), and proceed under the user's explicit responsibility.

한국어 설명: 사용자가 명시적 우회 플래그를 사용하면 게이트는 우회되지만, 우회 사실 자체는 동일한 보고서 경로에 BYPASSED verdict로 기록된다. 비대화형 환경(예: CI)에서는 rationale 미수집을 허용하되 `bypass_reason: "non-interactive"`를 자동 기록한다.

### `REQ-WAG-007` (Unwanted Behavior)

**If** the `plan-auditor` invocation times out, errors out, returns malformed output, or fails health-check, **then** the system shall **not** auto-PASS, shall classify the run as `INCONCLUSIVE`, and shall fall back to a manual review prompt presented to the user via AskUserQuestion (options: retry audit / proceed-with-acknowledgement / abort).

한국어 설명: 감사 실패는 PASS와 동일하지 않다. 시간 초과·예외·malformed report는 모두 INCONCLUSIVE이며, 자동 통과되지 않는다. 사용자에게 명시적으로 의사결정을 위임한다.

---

## 4. 범위 (Scope)

### 4.1 In Scope

- `.claude/skills/moai/workflows/run.md`에 **Phase 0.5: Plan Audit Gate** 단락 신설 (Phase 0과 Phase 1 사이).
- `.claude/skills/moai/team/run.md`에 동일 게이트 단락 신설 (team 모드 parity, `REQ-WAG-005`).
- `.claude/skills/moai/workflows/plan.md` 종료 단락에 **audit-ready 선언 출력** 추가 (Run으로의 전이 신호: progress.md에 `plan_complete_at` 기록).
- `internal/template/templates/.claude/skills/moai/workflows/{run,team/run,plan}.md` 동기화 (Template-First, `make build` 의무).
- `.moai/reports/plan-audit/.keep` 디렉터리 생성 (`.gitignore` 보강 포함).
- `.claude/rules/moai/workflow/spec-workflow.md`에 Phase 0.5 단락 추가 (워크플로우 protocol 문서).
- `--skip-audit` 플래그와 `MOAI_SKIP_PLAN_AUDIT` env 처리 로직 (현 단계: 워크플로우 skill 차원의 routing 합의; CLI flag binding은 별도 SPEC).
- 7일 grace window: 본 SPEC merge 후 7일간 게이트는 **warn-only** 모드로 동작 (FAIL이어도 차단하지 않고 경고만 출력). grace window 종료 후 자동으로 차단 모드로 전환.

### 4.2 Out of Scope

- `plan-auditor` agent 내부 채점 알고리즘·rubric 변경 (현행 4 must-pass + 4 scored 유지).
- `/moai plan` UI/CLI prompt 변경 (annotation cycle은 그대로).
- 새 agent 정의 (plan-auditor를 그대로 호출).
- `plan-auditor` 결과 캐시 invalidation 로직의 일반화 (본 SPEC은 24h 단순 TTL만 다룸; 정교한 invalidation은 후속 SPEC).
- 과거 SPEC 일괄 retroactive audit 자동 실행 (수동 트리거만 지원).

---

## 5. 위험 및 완화

| Risk ID | Risk | Likelihood | Impact | Mitigation |
|---------|------|------------|--------|------------|
| R-WAG-1 | 기존 SPEC 다수가 신규 게이트 통과 실패 → Run 단계 일제 차단 | High | High | (a) 7일 grace window: warn-only 모드. (b) `--skip-audit` 우회 + 기록. (c) Migration script(out-of-scope, 후속) 대신 수동 경고 일람 제공. |
| R-WAG-2 | plan-auditor 호출 token 비용으로 `/moai run` 지연·예산 초과 | Medium | Medium | (a) audit report 24h validity 캐시: 동일 SPEC, 동일 plan 산출물 hash, 24h 이내 PASS는 재호출 생략. (b) plan-auditor는 단일 호출(teammate 중복 호출 금지). (c) progress.md에 `audit_cache_hit` 표기. |
| R-WAG-3 | dogfood paradox — 본 SPEC 자체가 audit FAIL | Medium | High | (a) 본 SPEC을 §3 7건 EARS, §4 명시 in/out, §5 위험표, §6 HISTORY로 작성. (b) acceptance.md를 REQ별 1:1 AC로 작성. (c) self-audit summary를 본 응답에 포함(아래 §self-audit). |
| R-WAG-4 | plan-auditor가 PASS인데 사용자가 SPEC 결함을 인지해 차단을 원할 경우 | Low | Low | annotation cycle(`/moai plan`)에서 차단 가능. gate는 PASS 보장만 제공하며, 사용자 거부권은 별도. |
| R-WAG-5 | warn-only mode에서 사용자가 경고를 무시한 채 Run 진행 | Medium | Medium | warn-only 출력에 "기간 종료 D-N일" 카운트다운 명시. grace 종료 시 자동 차단으로 전환됨을 매 호출 표기. |
| R-WAG-6 | `--skip-audit` 남용 (게이트 형해화) | Medium | High | bypass 사용 횟수를 일자별 보고서에 누적 기록. 30일 누적 N회 초과 시 사용자에게 경고(post-MVP, out-of-scope). |
| R-WAG-7 | INCONCLUSIVE 처리 시 사용자가 일관되게 proceed-with-acknowledgement만 선택 | Low | Medium | proceed 선택 시 progress.md에 `inconclusive_acknowledged_by` 사용자 식별자 강제 기록. |

---

## 6. HISTORY

- 2026-04-25: SPEC-WF-AUDIT-GATE-001 v1.0.0 created. v3R2 35-SPEC 패키지 작성 직후 진행한 전수 점검에서 SPEC-SKILL-001(892B archived) 사고를 발견함에 따라, plan→run 전이 사이에 plan-auditor 의무 게이트를 신설하기 위해 작성. 본 SPEC은 dogfood 대상이며 자기 자신을 첫 번째 감사 대상으로 제출 가능한 수준으로 작성됨.
