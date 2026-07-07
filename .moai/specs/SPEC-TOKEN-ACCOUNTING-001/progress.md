# Progress — SPEC-TOKEN-ACCOUNTING-001

> §E 섹션과 `§E.N` 하위 헤딩은 파서 load-bearing(`internal/spec/era.go`)이다. 헤딩 개명 금지.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-07
- tier: M
- artifacts: spec.md, plan.md, acceptance.md (+ progress.md skeleton)
- owner: manager-spec
- note: Token-Economy Epic 1/4. plan-auditor 게이트 대기.

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소유>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소유>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소유. sync_commit_sha 는 sync commit 시 기록>_

## §I Token Accounting

_<pending sync-phase — 본 SPEC이 도입하는 신규 섹션. sync-close 시 token-accounting
메커니즘이 아래 필드를 채운다. era.go가 grep하지 않는 신규 top-level letter(§I)이므로
§E.N 파서와 무충돌. placeholder only — 값 미기록.>_

<!--
제안 필드 스키마 (run-phase에서 확정, sync-close 시 채움):
- tokens_spent: <int 합산>
- tokens_input: <int>
- tokens_output: <int>
- tokens_cache_creation: <int>
- tokens_cache_read: <int>
- cache_hit_ratio: <float [0,1]>
- token_attribution: session-set
- token_attribution_confidence: high | low
- token_session_count: <int>
-->
