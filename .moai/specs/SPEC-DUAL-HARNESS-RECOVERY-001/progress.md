---
id: SPEC-DUAL-HARNESS-RECOVERY-001
document: progress
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
card: t1100
---

# Progress — SPEC-DUAL-HARNESS-RECOVERY-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC status: `draft` (Tier L, 산출물 spec.md, plan.md, acceptance.md, design.md, research.md, progress.md).
- 카드·워크트리·브랜치: `t1100` / `.claude/worktrees/t1100-recovery` / `WT-dual-harness-recovery`. plan 기준 HEAD `d87e9af2e`.
- REQ 24개(REQ-DHR-001 ~ 024), AC 19개(AC-DHR-001 ~ 019, 그중 LIVE 2개: 012, 018).
- 설계 기준 매핑: AC-MIG-01 → AC-DHR-001~005, AC-WT-01 → 006~009, AC-AGENT-01 → 010~013, AC-MSG-01 → 014~016, AC-FACT-01 → 017~019.
- 운영자 결정 대기 4건: plan.md §B의 `unwire-trigger`, `codex-kanban-roles`, `idempotency-scope`, `live-budget`.
- 구현: 시작하지 않음. Implementation Kickoff Approval: 요청하지 않음. 독립 plan 감사: 아직 없음.
- plan 단계 기준 측정(이 트리, 이번 실행):

```text
$ go test ./internal/codexwiring ./internal/factorymsg ./internal/template/agentemit -count=1
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.608s
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	3.044s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.384s
$ go test ./internal/cli -run '^TestFactoryLive(CodexCodex|CodexClaude|ClaudeCodex|ClaudeClaudeCompletionSeparation)$' -count=1 -v
--- SKIP x4, 패키지 결과 ok (SKIP은 PASS 아님)
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
