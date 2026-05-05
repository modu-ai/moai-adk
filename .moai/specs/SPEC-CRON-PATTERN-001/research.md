# Research — SPEC-CRON-PATTERN-001 (Routines/Cron Pattern Catalog)

**SPEC**: SPEC-CRON-PATTERN-001
**Wave**: 4 / Tier 3 (장기/폴리싱)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

**Source**: Anthropic blog "Introducing Routines in Claude Code"
**URL**: https://claude.com/blog/introducing-routines-in-claude-code
**Accessed**: 2026-04-30 (verified via WebFetch)

### 1.1 Verbatim 인용 (§ "How it works")

> "A routine is a Claude Code automation you configure once — including a prompt, repo, and connectors — and then run on a schedule, from an API call, or in response to an event."

— Section: "How it works" (definition)

> "Give Claude Code a prompt and a cadence (hourly, nightly, or weekly) and it runs on that schedule"

— Section: "How it works" → "Scheduled routines"

> "You can also configure routines to be triggered by API calls. Every routine gets its own endpoint and auth token."

— Section: "How it works" → "API routines"

> "Subscribe a routine to automatically kick off in response to GitHub repository events."

— Section: "How it works" → "Webhook routines"

### 1.2 Verbatim 사용 예 (§ "Scheduled routines")

> "Every night at 2am: pull the top bug from Linear, attempt a fix, and open a draft PR."

— Section: "Scheduled routines" (example use case)

### 1.3 Anthropic 권고의 핵심 포인트

- **One-time configuration**: 단일 설정 + 반복 실행 = ad-hoc 실행 대비 토큰/시간 효율 ↑
- **Three trigger modes**: schedule, API, webhook — 각각 다른 use case
- **Plan-tier limits**: Pro 5/day, Max 15/day, Team 25/day, Enterprise unlimited
- **Routine is durable**: configuration이 한 번 성공하면 indefinite하게 가치 발생

---

## 2. 현재 상태 (As-Is)

### 2.1 moai-adk-go의 Cron 인프라

기존 도구:
- `CronCreate(prompt, schedule, repo, connectors)` — Claude Code native 도구 보유
- `CronList()` — 등록된 cron job 목록
- `CronDelete(id)` — cron job 제거

운영 현황:
- 위 도구는 정상 작동하나 **표준 routine 패턴 카탈로그 부재**
- 사용 예시 0건 (in-tree documentation에 등재된 routine 없음)
- 사용자가 어떤 cron job을 만들어야 효과적인지 학습 자료 없음

### 2.2 moai-adk-go에서 자동화 가능한 반복 작업 식별

후보 1: SPEC backlog triage (현재 수동, 매주 GOOS 확인)
후보 2: 코드-문서 drift detection (현재 ad-hoc, /moai sync 시점만)
후보 3: dependency update notification (Dependabot이 일부 처리, 정책 신호 부재)
후보 4: CI failure pattern aggregation (현재 부재, 디버그 시간 증가)
후보 5: stale memory cleanup (현재 부재, MEMORY.md 누적 → 토큰 비용 증가)

### 2.3 운영 격차

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|--------------|--------------|------|
| 표준 routine 카탈로그 | 부재 | 5+ 검증된 패턴 | 신규 문서 |
| 사용자 학습 자료 | 부재 | 패턴별 rationale + 예시 | 신규 카탈로그 |
| Plan-tier 안내 | 부재 | Pro/Max/Team/Enterprise 제한 표 | 신규 |
| 실패 자동 일시정지 | 부재 | 3회 연속 실패 시 pause | 신규 정책 |
| 실행 이력 영속화 | 부재 | `.moai/routines/history/` JSONL | 신규 |

---

## 3. 코드베이스 분석 (Affected Files)

### 3.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/rules/moai/workflow/routine-patterns.md` | 신규 | 5+ 표준 routine 패턴 카탈로그 |
| `internal/template/templates/.claude/rules/moai/workflow/routine-patterns.md` | 신규 | Template-First |
| `.moai/routines/history/.gitkeep` | 신규 | 실행 이력 디렉토리 placeholder |

### 3.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/skills/moai-workflow-project/SKILL.md` | cross-ref 추가 | Routine 패턴 사용 안내 |
| `CLAUDE.md` §3 또는 §16 | cross-ref 추가 | routine-patterns.md 위치 안내 |

### 3.3 표준 패턴 5종 후보 상세

**Pattern P1 — Backlog Triage (Daily)**
- Schedule: `0 2 * * *` (매일 02:00)
- Prompt: "List top 5 SPECs by priority in `.moai/specs/`. For each, summarize blockers and recommend next action."
- Output: `.moai/routines/history/triage-YYYY-MM-DD.jsonl`
- Rationale: 매일 새 SPEC을 plan-auditor 없이 자동 분류

**Pattern P2 — Documentation Drift Detection (Weekly)**
- Schedule: `0 9 * * 1` (매주 월요일 09:00)
- Prompt: "Compare `.moai/specs/SPEC-*/spec.md` modification dates against last commit on referenced files. Report mismatches."
- Output: `.moai/routines/history/drift-YYYY-MM-DD.jsonl`
- Rationale: 주간 리듬으로 코드-문서 동기화 모니터링

**Pattern P3 — Dependency Update Notifier (Daily)**
- Schedule: `0 6 * * *`
- Prompt: "Run `go list -u -m all` and report modules with new minor/patch versions. Create issue with summary."
- Output: GitHub issue + history log
- Rationale: Dependabot 보완 (정책 신호 + 우선순위)

**Pattern P4 — CI Failure Aggregator (Weekly)**
- Schedule: `0 10 * * 5` (매주 금요일 10:00)
- Prompt: "Analyze last 7 days of GitHub Actions failure logs. Identify recurring patterns. Recommend mitigations."
- Output: `.moai/routines/history/ci-failures-YYYY-MM-DD.jsonl`
- Rationale: 주간 CI 안정성 회고

**Pattern P5 — Memory Hygiene Sweep (Weekly)**
- Schedule: `0 3 * * 0` (매주 일요일 03:00)
- Prompt: "Audit `~/.claude/projects/<hash>/memory/MEMORY.md`. Identify entries older than 30 days that have not been referenced. Propose archive."
- Output: `.moai/routines/history/memory-sweep-YYYY-MM-DD.jsonl`
- Rationale: 토큰 비용 감소 + MEMORY 신선도

---

## 4. 위험 및 가정

### 4.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| Routine 토큰 비용 누적 (Pro: 5/day) | High | Medium | Plan-tier 표 명시, 사용자가 자기 plan 확인 |
| 실패 시 silent skip → 사용자 인지 못함 | Medium | High | 3회 연속 실패 → pause + 다음 세션 AskUserQuestion 알림 |
| 패턴 prompt 너무 광범위 → 토큰 폭발 | Medium | Medium | 패턴 prompt 작성 가이드 (200-500 tokens 권장) |
| `.moai/routines/history/` 누적 → disk overflow | Low | Low | 90일 retention 권장 (GOOS 수동 정리) |
| connector 설정 (Linear, Slack) 조직별 차이 | High | Low | 패턴 카탈로그는 generic, organization customization은 사용자 책임 |

### 4.2 Assumptions

- A1: `CronCreate` 도구는 안정적이며 v2.1.111+에서 유지됨
- A2: 5개 패턴이 moai-adk-go 사용자의 80% 이상 use case 커버
- A3: Plan-tier limit는 사용자가 자기 plan에서 확인 (orchestrator는 정보 제공만)
- A4: `.moai/routines/history/` 디렉토리는 git-ignored 권장 (사용자 선택)
- A5: 정책은 living document — 사용자가 자기 패턴 추가 가능

---

## 5. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| 표준 패턴 카탈로그 | 패턴 수 | >= 5 |
| 패턴별 rationale | 문서 검토 | 모두 명시 |
| Plan-tier 제한 표 | 문서 검토 | 4 plan 모두 |
| 실행 이력 파일 형식 | JSONL schema 검증 | 정확 |
| 3회 연속 실패 정책 | 문서 검토 | 명시 |
| Cross-ref 카운트 | grep | >= 2 |
| Template-First sync | `make build` diff | clean |

---

## 6. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|-----------|------|
| Routine pattern을 Go 코드로 hard-code | ❌ | living document 원칙 위배, 사용자 customization 불가 |
| 5개 패턴 모두 자동 등록 | ❌ | Plan-tier limit (Pro 5/day) 위반 위험 |
| `.moai/routines/` 대신 `.claude/routines/` | ❌ | `.moai/`가 SSOT (project-scope), `.claude/`는 config-only |
| 실행 이력을 GitHub issue로 영속화 | ❌ | issue 폭발 위험, JSONL이 더 적합 |
| `moai routine create` CLI 신설 | ✅ (별도 SPEC) | 현재 SPEC scope 외 |

---

## 7. 참고 SPEC

- SPEC-MEMO-001: 기존 메모리 시스템 (Pattern P5의 입력)
- SPEC-METRICS-001 (이번 wave sibling): 통계 수집 (Pattern P4의 입력 일부)
- SPEC-V3R2-WF-001: 워크플로우 표준 (Pattern P2와 정합)

---

## 8. Open Questions (Plan 단계 해결 대상)

- OQ1: `.moai/routines/history/` 디렉토리는 git-tracked인가 git-ignored인가? → plan.md에서 결정
- OQ2: 패턴 5개 외 추가 후보 (예: PR review aggregator)는 v2.x.0 후속 SPEC?
- OQ3: 3회 연속 실패 정책의 "실패" 정의 (exit code? error log?) → plan.md에서 명시
- OQ4: CronCreate 도구의 schedule 표현식 호환성 (cron syntax)? → research 필요
- OQ5: 사용자 자기 패턴 추가 시 카탈로그 contribution 절차?

---

End of research.md (SPEC-CRON-PATTERN-001).
