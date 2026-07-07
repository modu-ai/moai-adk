---
id: SPEC-TOKEN-ACCOUNTING-001
title: "Per-SPEC Token Accounting — runtime measurement baseline (Token-Economy Epic 1/4)"
version: "0.1.1"
status: in-progress
created: 2026-07-07
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.1.0"
module: "internal/tokenusage"
lifecycle: spec-anchored
tags: "token, accounting, measurement, telemetry, cost"
era: V3R6
---

# SPEC-TOKEN-ACCOUNTING-001 — Per-SPEC Token Accounting

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-07 | manager-spec | 초안 작성 — Token-Economy Epic 1/4 (Gap A: 런타임 per-SPEC 토큰 측정) |
| 0.1.1 | 2026-07-08 | manager-spec | D1+D2 plan-audit debt 해소 — REQ-TA-014 배포중립성 추가 + AC-TA-012 재연결 + §D.4 DoD 문구 명확화 |

## §A Context / 배경

토큰 과금(pay-per-token) 모델에서는 **"최소 토큰 × 최대 품질"** 이 경쟁 축이 된다.
무엇이든 최적화하려면 먼저 **측정**해야 한다. 현재 리포지토리에는 인메모리 per-agent
토큰 Tracker (`internal/runtime/budget.go` — `Tracker.RecordCall`)가 있으나,
**한 SPEC이 누적 소비한 토큰 비용은 어디에도 영속되지 않는다.** SPEC이 close될 때
"이 SPEC은 N 토큰을 소비했다"는 기록이 없으므로, 토큰-대-품질 KPI를 계산할 수 없고
harness/mode 간 비교도 불가능하다.

Gap A의 목표: **per-SPEC 토큰 소비를 영속·감사 가능한 측정값으로 만들어** 하위
최적화(Gap B 라우팅 매트릭스 / Gap C 검증-출력 다이어트 / Gap D 예산 하드스톱)가
측정 가능한 baseline 위에서 진행되도록 한다.

### §A.1 Epic 위치 (SSOT)

본 SPEC은 4-SPEC **Token-Economy Epic** 의 1/4 (측정 기반). 나머지 3개(라우팅 매트릭스,
검증-출력 다이어트, 예산 하드스톱)는 **아직 미작성**이며 본 SPEC은 이들을 미래/관련
작업으로만 참조한다 (Epic scope SSOT는 본 §A.1). 본 SPEC은 이들을 생성하지 않는다.

### §A.2 측정 소스의 사실 확인 (실측)

플랜 작성 시 실측으로 확인한 사실 (verification-claim-integrity — 관측된 것만 주장):

- Claude Code는 세션별 transcript JSONL(`~/.claude/projects/<hash>/<session-uuid>.jsonl`)의
  각 assistant 턴 레코드에 `message.usage` 객체를 기록한다. 실측한 키:
  `input_tokens`, `output_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens`
  (+ `service_tier`, `cache_creation.*` 등 부가 필드). 활성 세션 하나에서 11개 assistant
  턴 합산: input 414,878 / output 35,555 / cache_read 828,992 / cache_creation 0 (실측).
- transcript 파일명(`<session-uuid>.jsonl`)이 곧 세션 UUID이며, `.moai/state/context-usage.json`
  및 `moai session current`가 노출하는 session_id와 동일 lineage로 상관 가능하다.
- **중요 구분**: statusline / `context-usage.json`의 `tokens_used`는 **현재 컨텍스트 창 점유량
  스냅샷** (`/clear`·compact 시 리셋)이며, 누적 과금 토큰이 아니다. Gap A가 요구하는
  "SPEC이 소비한 토큰"은 transcript `message.usage` 를 턴 전반에 걸쳐 **합산**한 값이다.
  statusline 스냅샷은 baseline 소스가 아니다.

## §B Requirements (GEARS)

> GEARS 표기. `<subject>` 는 일반화된 명사(mechanism/parser/audit 등)를 사용한다.

### §B.1 추출 · 합산 (Extraction)

- **REQ-TA-001** (Ubiquitous): The token-accounting mechanism **shall** extract the four
  per-turn usage fields — `input_tokens`, `output_tokens`, `cache_creation_input_tokens`,
  `cache_read_input_tokens` — from each assistant-type record of a Claude Code session
  transcript JSONL file.
- **REQ-TA-002** (Ubiquitous): The mechanism **shall** sum the four usage fields across all
  assistant-type turns of a given session transcript to produce a session-total token figure
  (`tokens_input`, `tokens_output`, `tokens_cache_creation`, `tokens_cache_read`, and their
  sum `tokens_spent`).
- **REQ-TA-003** (Ubiquitous): The mechanism **shall** compute a cache-hit ratio as
  `cache_read / (input + cache_creation + cache_read)`, bounded to `[0, 1]`, and **shall**
  define the ratio as `0` when the denominator is `0`.
- **While** the mechanism reads transcript files, **REQ-TA-004** (State-driven): it **shall**
  treat `~/.claude/projects/**` as read-only and **shall not** create, modify, or delete any
  file there.

### §B.2 귀속 (Attribution)

- **REQ-TA-005** (Ubiquitous): The mechanism **shall** attribute token spend to a SPEC via the
  **session-set summation** approach — summing the transcript usage of the session-UUID set
  recorded in the SPEC's `progress.md` lineage plus the active sync session — and **shall**
  record the method as `token_attribution: session-set`.
- **Where** the SPEC's session lineage is unavailable (environment-fallback session IDs, or no
  recorded session UUID), **REQ-TA-006** (Capability gate): the mechanism **shall** fall back to
  current-session-only measurement and **shall** record `token_attribution_confidence: low`.
- **REQ-TA-007** (Ubiquitous): The mechanism **shall** record an explicit
  `token_attribution_confidence` qualifier (`high` when every contributing session is
  SPEC-dedicated; `low` when any contributing session is shared across SPECs or lineage is
  unavailable) so that no precision is claimed beyond what the session-set approach delivers
  (verification-claim-integrity §1.1).

### §B.3 영속 (Persistence)

- **When** a SPEC reaches sync-close, **REQ-TA-008** (Event-driven): the mechanism **shall**
  persist the measured token figures to the SPEC's `progress.md` **`## §I Token Accounting`**
  section (a fresh top-level section letter, NOT an `§E.N` sub-heading).
- **REQ-TA-009** (Unwanted behavior): The mechanism **shall not** rename, remove, or collide
  with any `internal/spec/era.go` parser-load-bearing token — the literal headings `§E.2`,
  `§E.3`, `§E.4`, `§E.5` and the literal field names `sync_commit_sha`, `mx_commit_sha`.
- **REQ-TA-010** (Ubiquitous): The `§I` field set **shall** use new field names distinct from
  the two parsed SHA fields: `tokens_spent`, `tokens_input`, `tokens_output`,
  `tokens_cache_creation`, `tokens_cache_read`, `cache_hit_ratio`, `token_attribution`,
  `token_attribution_confidence`, `token_session_count`.

### §B.4 감사 표면 (Audit Surface)

- **When** `moai spec audit --json` runs, **REQ-TA-011** (Event-driven): the audit output
  **shall** surface a `tokens_spent` value per SPEC that carries a populated `§I Token
  Accounting` section (JSON field; a table column is a nice-to-have MVP addition).
- **Where** a SPEC has no populated `§I` section, **REQ-TA-012** (Capability gate): the audit
  output **shall** emit `tokens_spent: null` (or omit the field) rather than fabricating a value.

### §B.5 견고성 (Robustness)

- **Where** a transcript JSONL line is malformed or a transcript file is absent,
  **REQ-TA-013** (Capability gate): the mechanism **shall** skip the offending line/file and
  continue, **shall not** panic, and **shall not** abort the surrounding audit run.

### §B.6 배포 중립성 (Deployment Neutrality)

- **REQ-TA-014** (Unwanted behavior): The token-accounting source — the new `internal/tokenusage`
  package and the `internal/spec` / `internal/cli` extensions — **shall not** reside anywhere under
  `internal/template/templates/`, preserving 16-language 배포(deployment) 중립성. 본 SPEC의 코드는
  런타임/개발 도구이며 배포 템플릿 콘텐츠가 아니므로 template tree에 두지 않는다.

## §C Success Criteria

1. 임의 세션 transcript에 대해 `tokens_spent` 가 4개 필드의 산술 합과 일치한다 (단위 테스트).
2. per-SPEC `§I Token Accounting` 필드가 sync-close 시 progress.md에 기록되고 grep으로 검증된다.
3. `§I` 필드 추가 전후로 `ClassifyEra()` 결과가 불변이다 (파서 무충돌 회귀 테스트).
4. `moai spec audit --json` 출력에 `tokens_spent` 필드가 노출된다.
5. 귀속 신뢰도 qualifier가 lineage 가용성에 따라 `high`/`low` 로 정확히 기록된다.

## §D Out of Scope (Exclusions)

본 SPEC은 **측정 baseline**만 구축한다. 다음은 명시적으로 범위 밖이다.

### Out of Scope — 토큰 예산 하드스톱 / 차단 (enforcement)

- 토큰 예산 초과 시 실행을 **차단**하거나 하드-페일하는 로직은 만들지 않는다. 이는 Gap D
  (예산 하드스톱, 미작성 Epic SPEC 4/4)의 소관이다. 기존 `internal/runtime/budget.go`
  Tracker의 warning-first 정책(BC-V3R3-006)은 그대로 유지되며 본 SPEC은 이를 변경하지 않는다.

### Out of Scope — mode/harness 라우팅 최적화 (routing)

- 측정된 토큰-대-품질 KPI에 기반해 mode/model/harness를 **선택**하는 라우팅 매트릭스는
  만들지 않는다. 이는 Gap B (라우팅 매트릭스, 미작성 Epic SPEC 2/4)의 소관이다.

### Out of Scope — 검증-출력 다이어트 (verification-output reduction)

- 검증 배치/출력 토큰을 **줄이는** 최적화는 만들지 않는다. 이는 Gap C (검증-출력 다이어트,
  미작성 Epic SPEC 3/4)의 소관이다.

### Out of Scope — 서브에이전트-내부 토큰의 정밀 분해 (sub-agent internal accounting)

- manager-spec/develop/docs 등 서브에이전트 **내부** 턴의 토큰을 per-agent로 정밀 분해하는
  것은 범위 밖이다. transcript `message.usage`가 서브에이전트 내부 토큰을 포함하는지는
  플랜 시점에 미검증이며, 본 SPEC은 이를 **residual risk**로 명시하고 세션-집합 총합만
  측정한다 (verification-claim-integrity — 검증하지 않은 정밀도를 주장하지 않는다).

### Out of Scope — 기존 정적 가드와의 중복 (no duplication)

- SPEC-TOKEN-EFFICIENCY-001의 always-loaded 75K 트립-와이어 가드
  (`internal/config/token_budget_guard.go`, 정적 회귀 가드)와 SPEC-TOKEN-001의 skill-count
  축소는 본 SPEC과 **상보적·별개**이며 재구현하지 않는다.

### Out of Scope — 실시간 statusline 스냅샷 변경 (statusline)

- `context-usage.json`/statusline의 현재-컨텍스트-점유 스냅샷 로직은 변경하지 않는다.
  본 SPEC은 그것을 소스로 쓰지 않고 transcript `message.usage` 합산을 별도로 읽는다.

## §E Cross-References

- `internal/runtime/budget.go` — 인메모리 per-agent Tracker (EXTEND 대상, 재구축 아님)
- `internal/statusline/usage.go`, `internal/statusline/context_usage.go` — 기존 usage/context 읽기 (구분 대상)
- `internal/spec/era.go` — 파서 load-bearing 토큰 (충돌 금지 대상)
- `internal/spec/audit.go`, `internal/cli/spec_audit.go` — audit 표면 확장 지점
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map — §I 배정 근거 SSOT
- `.claude/rules/moai/core/verification-claim-integrity.md` — 귀속 정밀도 주장 제약
- SPEC-TOKEN-EFFICIENCY-001, SPEC-TOKEN-001 — 상보적 기존 SPEC (dedup 확인 완료)
