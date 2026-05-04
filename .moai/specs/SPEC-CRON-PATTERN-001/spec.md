---
id: SPEC-CRON-PATTERN-001
status: draft
version: "0.1.0"
priority: Medium
labels: [routine, cron, automation, pattern, catalog, wave-4, tier-3]
issue_number: null
scope: [.claude/rules/moai/workflow, .claude/skills/moai-workflow-project, .moai/routines, CLAUDE.md]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 4
tier: 3
---

# SPEC-CRON-PATTERN-001: Routines/Cron Pattern Catalog

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 4 / Tier 3. Anthropic "Introducing Routines in Claude Code" 권고를 본 프로젝트의 표준 routine 패턴 카탈로그로 변환. 5+ 패턴 신설.

---

## 1. Goal (목적)

`CronCreate` / `CronList` / `CronDelete` 도구가 이미 존재하나 **표준 routine 패턴 카탈로그가 부재**하다. 본 SPEC은 5개 검증된 routine 패턴 (Backlog Triage / Documentation Drift / Dependency Notifier / CI Failure Aggregator / Memory Hygiene)을 카탈로그화하여 사용자가 즉시 활용할 수 있도록 한다.

### 1.1 배경

- Anthropic blog "Introducing Routines in Claude Code": "A routine is a Claude Code automation you configure once — including a prompt, repo, and connectors — and then run on a schedule, from an API call, or in response to an event."
- 본 프로젝트의 cron 도구는 정상 작동 → 그러나 사용 예시 0건, 학습 자료 부재
- 사용자가 매주/매일 수행하는 반복 작업 5종 식별 (research.md §2.2)

### 1.2 비목표 (Non-Goals)

- `moai routine create <pattern>` CLI 신규 구현 (별도 SPEC 후보)
- 5+ 패턴 자동 등록 (Plan-tier limit 위반 위험, 사용자 선택)
- Cron 실행 도구 자체 변경 (`CronCreate`는 그대로 사용)
- 자동 패턴 추천 시스템 (사용자가 직접 선택)
- `.moai/routines/history/` 자동 retention 도구 (사용자 책임)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/rules/moai/workflow/routine-patterns.md` 신규 작성
- 5개 표준 패턴 (P1-P5) 명시: schedule, prompt, output, rationale
- Plan-tier limit 표 (Pro 5/day, Max 15/day, Team 25/day, Enterprise unlimited)
- 3회 연속 실패 → pause 정책
- 실행 이력 영속화 schema (`.moai/routines/history/<routine-id>-YYYY-MM-DD.jsonl`)
- 패턴 prompt 작성 가이드 (200-500 tokens 권장)
- `moai-workflow-project` SKILL.md cross-ref
- CLAUDE.md cross-ref
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- `moai routine create <pattern>` CLI 명령어
- 자동 패턴 등록
- Cron 실행 엔진 자체 변경
- `.moai/routines/history/` 자동 cleanup 도구
- Plan-tier 자동 detection / enforce
- 사용자 자기 패턴 contribution 자동화

---

## 3. Environment (환경)

- 런타임: Claude Code v2.1.111+ Cron 도구 (CronCreate, CronList, CronDelete)
- Plan: Pro / Max / Team / Enterprise (각각 5/15/25/unlimited per day)
- 영향 파일: `.claude/rules/moai/workflow/`, `.claude/skills/moai-workflow-project/`, `.moai/routines/`, `CLAUDE.md`
- Connectors (옵션): Linear, Slack, GitHub (사용자 organization별 customization)

---

## 4. Assumptions (가정)

- A1: `CronCreate` 도구는 v2.1.111+에서 안정 유지
- A2: 5개 패턴이 80% 이상 use case 커버
- A3: 사용자가 Plan-tier limit를 자기 plan에서 확인 (orchestrator는 정보 제공만)
- A4: `.moai/routines/history/` 디렉토리는 git-ignored 권장 (사용자 선택)
- A5: 정책은 living document — 사용자가 자기 패턴 추가 가능

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-CRON-001**: THE FILE `.claude/rules/moai/workflow/routine-patterns.md` SHALL exist and document at least 5 standard routine patterns.
- **REQ-CRON-002**: EACH STANDARD PATTERN SHALL specify schedule, prompt, output target, and rationale.
- **REQ-CRON-003**: THE CATALOG SHALL include a Plan-tier limits table for Pro, Max, Team, and Enterprise.
- **REQ-CRON-004**: THE CATALOG SHALL define routine execution history schema at `.moai/routines/history/<routine-id>-YYYY-MM-DD.jsonl`.

### 5.2 Event-Driven Requirements

- **REQ-CRON-005**: WHEN a user invokes `CronCreate` for a standard pattern, THE PATTERN SHALL be referenced by ID (P1-P5) in the cron job description.
- **REQ-CRON-006**: WHEN a routine fails 3 consecutive runs, THE ROUTINE SHALL be paused and the orchestrator SHALL surface the failure on next session via natural-language status announcement.
- **REQ-CRON-007**: WHEN a routine completes execution, THE ROUTINE SHALL append a JSONL entry to `.moai/routines/history/<routine-id>-YYYY-MM-DD.jsonl` with timestamp, exit status, and summary.

### 5.3 State-Driven Requirements

- **REQ-CRON-008**: WHILE a routine is paused due to consecutive failures, THE ROUTINE SHALL NOT be re-triggered until user explicitly resumes via `CronCreate` re-registration.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-CRON-009**: WHERE a user's Plan-tier limit is reached for the day, THE CATALOG SHALL warn the user about further routine creation impact.
- **REQ-CRON-010**: WHERE `.moai/routines/history/` directory does not exist, THE FIRST ROUTINE EXECUTION SHALL create the directory automatically.
- **REQ-CRON-011**: IF a routine prompt exceeds 500 tokens, THE PATTERN AUTHORING GUIDE SHALL warn against potential token cost amplification.
- **REQ-CRON-012**: WHERE a routine uses external connectors (Linear, Slack), THE PATTERN DOCUMENTATION SHALL note that connector configuration is user/organization responsibility.

### 5.5 Unwanted (Negative) Requirements

- **REQ-CRON-013**: THE CATALOG SHALL NOT auto-register any routine without explicit user action.
- **REQ-CRON-014**: THE CATALOG SHALL NOT include patterns that bypass Plan-tier limits.
- **REQ-CRON-015**: THE PATTERN PROMPTS SHALL NOT include hardcoded credentials, API keys, or secrets.
- **REQ-CRON-016**: THE CATALOG SHALL NOT mandate `.moai/routines/history/` retention policy (user responsibility).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 카탈로그 문서 존재 | file existence | EXISTS |
| 표준 패턴 수 | 패턴 카운트 | >= 5 |
| 패턴별 4-항목 명시 | schedule + prompt + output + rationale | 100% |
| Plan-tier 표 | 4 plan 명시 | EXISTS |
| Failure pause 정책 | 3-consecutive 명시 | EXISTS |
| History schema | JSONL 정의 | EXISTS |
| Cross-ref | grep | >= 2 |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Go 코드 변경 없음 (정책 + 카탈로그 문서만)
- C2: `CronCreate` 도구 자체는 변경 없음
- C3: 패턴 prompt 200-500 tokens 권장 (hard limit 아님)
- C4: `.moai/routines/history/` 디렉토리는 git-ignored 권장 (.gitignore 안내만)
- C5: Template-First Rule 준수

End of spec.md (SPEC-CRON-PATTERN-001 v0.1.0).
