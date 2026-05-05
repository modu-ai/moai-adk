---
id: SPEC-CRON-PATTERN-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-CRON-PATTERN-001

## 1. Overview

Anthropic의 Routines 기능을 본 프로젝트의 5개 표준 패턴 카탈로그로 변환. `CronCreate` 도구가 이미 존재하므로 **카탈로그 + 정책 문서**만으로 사용자가 즉시 활용 가능. 5 패턴: Backlog Triage / Doc Drift / Dep Notifier / CI Failure Aggregator / Memory Hygiene.

## 2. Approach Summary

**전략**: Catalog-First, Documentation-Driven, User-Activated.

1. `.claude/rules/moai/workflow/routine-patterns.md` 신규 작성
2. 5 패턴 (P1-P5) 명시: schedule + prompt + output + rationale
3. Plan-tier limits 표 + 실패 정책 + history schema
4. SKILL.md / CLAUDE.md cross-ref
5. Template-First 동기화

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] 본 프로젝트의 현재 cron 사용 사례 0건 확인 (베이스라인)
- [ ] 사용자가 매일/매주 수행하는 반복 작업 5종 검증 (research.md §2.2 매핑)
- [ ] Anthropic blog "Introducing Routines in Claude Code" verbatim 재확인
- [ ] Plan-tier limits 정확성 검증 (Pro 5/day, Max 15/day, Team 25/day, Enterprise unlimited)
- [ ] `CronCreate` 도구 schedule 표현식 호환성 확인 (cron syntax)

**Exit Criteria**: 5 패턴 후보 검증 + 베이스라인 확보

### M1 — Catalog Document 작성 (Priority: High)

- [ ] `.claude/rules/moai/workflow/routine-patterns.md` 신규 작성
  - §1 Overview + scope
  - §2 5 Standard Patterns (P1-P5):
    - P1 Backlog Triage (daily 02:00)
    - P2 Documentation Drift (weekly Mon 09:00)
    - P3 Dependency Notifier (daily 06:00)
    - P4 CI Failure Aggregator (weekly Fri 10:00)
    - P5 Memory Hygiene Sweep (weekly Sun 03:00)
  - §3 Each pattern: schedule + prompt + output + rationale (4 fields)
  - §4 Plan-Tier Limits Table
  - §5 Failure Policy (3 consecutive → pause + next-session announce)
  - §6 History Schema (`.moai/routines/history/<routine-id>-YYYY-MM-DD.jsonl`)
  - §7 Pattern Authoring Guide (200-500 tokens prompt 권장)
  - §8 Anti-Patterns (hardcoded credentials, etc.)
- [ ] 카탈로그 문서 3-4KB 범위로 압축

**Exit Criteria**: 8 절 모두 작성, 5 패턴 명시

### M2 — Pattern Specifications (Priority: High)

- [ ] **P1 Backlog Triage**:
  - Schedule: `0 2 * * *`
  - Prompt: "List top 5 SPECs by priority in `.moai/specs/`. For each, summarize blockers (file:line if available) and recommend next action. Output: structured markdown."
  - Output: `.moai/routines/history/triage-YYYY-MM-DD.jsonl`
  - Rationale: Plan phase 자동 분류, 사용자 매주 수동 작업 대체
- [ ] **P2 Documentation Drift**:
  - Schedule: `0 9 * * 1`
  - Prompt: "Compare `.moai/specs/SPEC-*/spec.md` modification dates against last commit on referenced files. Report mismatches with file paths."
  - Output: `.moai/routines/history/drift-YYYY-MM-DD.jsonl`
  - Rationale: 코드-문서 동기화 모니터링
- [ ] **P3 Dependency Notifier**:
  - Schedule: `0 6 * * *`
  - Prompt: "Run `go list -u -m all` and identify modules with new minor/patch versions. Create GitHub issue with summary if any updates available."
  - Output: GitHub issue + `.moai/routines/history/dep-YYYY-MM-DD.jsonl`
  - Rationale: Dependabot 보완 (정책 신호)
- [ ] **P4 CI Failure Aggregator**:
  - Schedule: `0 10 * * 5`
  - Prompt: "Analyze last 7 days of GitHub Actions failure logs via `gh run list --status failure`. Identify recurring patterns. Recommend mitigations."
  - Output: `.moai/routines/history/ci-failures-YYYY-MM-DD.jsonl`
  - Rationale: 주간 CI 안정성 회고
- [ ] **P5 Memory Hygiene Sweep**:
  - Schedule: `0 3 * * 0`
  - Prompt: "Audit `~/.claude/projects/<hash>/memory/MEMORY.md`. Identify entries older than 30 days that have not been referenced. Propose archive."
  - Output: `.moai/routines/history/memory-sweep-YYYY-MM-DD.jsonl`
  - Rationale: MEMORY 신선도 + 토큰 비용 감소

**Exit Criteria**: 5 패턴 모두 4-field 명시 (schedule + prompt + output + rationale)

### M3 — Plan-Tier Limits + Failure Policy (Priority: High)

- [ ] Plan-Tier Limits Table:
  | Plan | Routines/day | Recommended Patterns |
  |------|--------------|---------------------|
  | Pro | 5 | P1 + P2 (선택 3개) |
  | Max | 15 | All 5 + 추가 가능 |
  | Team | 25 | All 5 + 팀 공유 |
  | Enterprise | unlimited | All 5 + organization-wide |
- [ ] Failure Policy 명시:
  - 3 consecutive failures → automatic pause
  - Pause 상태 = next-session에서 orchestrator가 자연어 status 안내
  - User가 명시 `CronCreate` 재등록 시까지 routine 미실행
  - "실패" 정의: routine 응답이 error indicator 포함 (exit code != 0 또는 "ERROR" prefix)

**Exit Criteria**: 4 plan tier 명시 + failure 정의 명시

### M4 — History Schema + Output Standards (Priority: Medium)

- [ ] History JSONL schema 정의:
  ```jsonl
  {"timestamp": "ISO-8601", "routine_id": "P1", "exit_status": "success|error|skipped", "summary": "1-line", "output_size_bytes": <int>, "duration_ms": <int>}
  ```
- [ ] `.moai/routines/history/.gitkeep` 추가 (디렉토리 placeholder)
- [ ] `.gitignore` 권장 entry: `/.moai/routines/history/*.jsonl` (사용자 선택)
- [ ] Retention 권장: 90일 (사용자 수동 정리)

**Exit Criteria**: schema + placeholder + gitignore 권장 명시

### M5 — Cross-Reference + Skill Integration (Priority: Medium)

- [ ] `.claude/skills/moai-workflow-project/SKILL.md`에 cross-ref 추가:
  ```
  Routine 자동화: .claude/rules/moai/workflow/routine-patterns.md (5 표준 패턴)
  ```
- [ ] `CLAUDE.md` §3 또는 §16에 카탈로그 위치 안내 추가
- [ ] Pattern P5 (Memory Hygiene)의 cross-ref:
  - `~/.claude/projects/<hash>/memory/MEMORY.md` 형식 인용
- [ ] Pattern P3 (Dependency)의 cross-ref:
  - GitHub Issue label `area:deps` 활용

**Exit Criteria**: 2+ cross-ref 추가, P3/P5 통합 명시

### M6 — Template-First Sync + Documentation (Priority: Medium)

- [ ] `internal/template/templates/.claude/rules/moai/workflow/routine-patterns.md` 동기화
- [ ] `internal/template/templates/.claude/skills/moai-workflow-project/SKILL.md` 동기화 (M5 cross-ref 반영)
- [ ] `internal/template/templates/CLAUDE.md` 동기화 (M5 cross-ref)
- [ ] `make build` 실행 → embedded.go 재생성
- [ ] CHANGELOG entry under Unreleased

**Exit Criteria**: Template-First sync clean

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] 5 패턴 모두 cron syntax 검증 (manual)
- [ ] history schema JSONL parse 검증 (sample 5 entries)
- [ ] cross-ref 2+ 검증 (grep)
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 카탈로그 구조 (요약)

```markdown
# Routine Patterns Catalog

## 1. Overview
This catalog documents 5 standard routine patterns for use with CronCreate...

## 2. Standard Patterns (5)

### P1 — Backlog Triage
- Schedule: 0 2 * * *
- Prompt: ...
- Output: .moai/routines/history/triage-YYYY-MM-DD.jsonl
- Rationale: ...

### P2 — Documentation Drift
... (similarly)

## 3. Plan-Tier Limits
(table)

## 4. Failure Policy
3 consecutive failures → pause...

## 5. History Schema
(JSONL example)

## 6. Pattern Authoring Guide
200-500 tokens recommended...

## 7. Anti-Patterns
- Hardcoded credentials
- Connector secrets in prompt
- Routines that mutate code (use webhook instead)
```

### 4.2 Routine 등록 사용자 워크플로우 (안내)

```bash
# Pro plan, P1 등록 예
CronCreate --pattern P1 --schedule "0 2 * * *" --prompt "..." --output ".moai/routines/history/triage-{date}.jsonl"
```

본 SPEC은 CLI 신규 구현 X — `CronCreate` 도구 직접 호출 안내.

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Routine 토큰 비용 누적 (Pro 5/day) | High | Medium | Plan-tier 표 명시, 사용자가 자기 plan 확인 |
| 실패 시 silent skip | Medium | High | 3-consecutive → pause + next-session announce |
| 패턴 prompt 토큰 폭발 | Medium | Medium | 200-500 tokens 권장 명시 |
| `.moai/routines/history/` 누적 | Low | Low | 90일 retention 권장 (manual) |
| connector 설정 조직별 차이 | High | Low | 카탈로그는 generic, customization 사용자 책임 |
| Cron syntax 호환성 | Low | Medium | M0에서 검증 |

## 6. Dependencies

- 선행 SPEC: 없음 (standalone)
- 의존 입력: `CronCreate` 도구 (Claude Code v2.1.111+)
- sibling SPEC: SPEC-MEMO-001 (P5 입력), SPEC-METRICS-001 (P4 입력 후보)
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (history git-tracked vs ignored): `.gitignore` 권장 entry 명시 (사용자 선택)
- **OQ2** (5 외 추가 후보): v2.x.0 후속 SPEC, 현재 5 패턴이 80% use case 커버
- **OQ3** ("실패" 정의): routine 응답이 error indicator 포함 (exit code != 0 또는 "ERROR" prefix). M3에서 명시.
- **OQ4** (cron syntax 호환성): M0에서 검증, standard cron syntax 가정
- **OQ5** (사용자 패턴 contribution): 본 SPEC scope 외, 추후 PR 통한 카탈로그 확장

## 8. Rollout Plan

1. M1-M6 구현 후 본 프로젝트의 다음 sync에 P2 (Doc Drift)부터 dogfooding
2. P2 검증 후 P1, P3 순차 활성화 (priority order)
3. CHANGELOG + v2.x.0 minor release
4. Documentation site 사용자 가이드 업데이트 (별도 PR)

End of plan.md (SPEC-CRON-PATTERN-001).
