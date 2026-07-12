---
id: SPEC-HOOK-DEADCODE-001
title: "internal/hook package dead-code cleanup (3 corroborated scopes)"
version: "0.1.0"
status: in-progress
created: 2026-07-03
updated: 2026-07-12
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tier: M
tags: "cleanup, dead-code, hook, refactor, go, internal-hook"
---

# SPEC-HOOK-DEADCODE-001 — Progress Tracking

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-03
plan_author: manager-spec
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
milestones: 3
```

Plan-phase artifacts (spec.md + plan.md + acceptance.md + this progress.md skeleton) authored and independently evidence-verified per `verification-claim-integrity.md` (see spec.md §A.3 for the full re-verification log and §A.4 for a material correction to the original Scope 3 evidence handoff). Ready for plan-auditor review.

## §E.2 Run-phase Evidence

### M1 — COMPLETED 2026-07-04 (commit `43c996bb7`)

- **Commit**: `43c996bb7c315ddc639c1754f13bd7f608c3012f` — `refactor(SPEC-HOOK-DEADCODE-001): M1 internal/hook/agents+lifecycle 미연결 패키지 삭제` (2026-07-04 18:57:50 +0900)
- **Stats** (`git show --stat 43c996bb7`): 21 files changed, 3 insertions(+), 2874 deletions(−)
- **Files deleted** (18 `.go` files = 13 production + 5 test):
  - `internal/hook/agents/{backend_handler,cycle_handler,default_handler,devops_handler,docs_handler,factory,frontend_handler,quality_handler,retired_handler,spec_handler}.go` (10 production)
  - `internal/hook/agents/{base_handler_test,factory_test}.go` (2 test)
  - `internal/hook/lifecycle/{cleanup,persistence,types}.go` (3 production)
  - `internal/hook/lifecycle/{cleanup_coverage_test,cleanup_test,persistence_test}.go` (3 test)
- **Doc hygiene edit** (2 files, +1/−1 each): `agent-hooks.md` "Handler Architecture" section corrected in BOTH live copy (`.claude/rules/moai/core/agent-hooks.md`) AND template mirror (`internal/template/templates/.claude/rules/moai/core/agent-hooks.md`) — dangling `internal/hook/agents/factory.go` reference replaced with generic `EventType`-based dispatch description.
- **Commit-time verification** (per commit context): `go build ./...` exit 0; `go list -deps ./cmd/moai | grep -E 'internal/hook/(agents|lifecycle)'` empty (exit 1 — both packages absent from binary dependency graph); `agent-hooks.md` live ↔ template mirror byte-identical; 20-package test subset green.

### M2 — _<pending run-phase>_

_<pending run-phase>_

### M3 — _<pending run-phase>_

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 0.95 Mode Selection

**Input parameters** (orchestrator, plan→run boundary):
- tier: M · scope: ~4-6 files · domain count: 1 (internal/hook + internal/cli/hook.go — Go dead-code cleanup, single domain)
- file language mix: Go (production + test) + 1 Markdown edit (agent-hooks.md Actions table row — M3)
- concurrency benefit: LOW (coding-heavy deletion with PRESERVE targets; Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: not met (Mode 3 retired)

**Mode evaluation**:
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | 다중 파일 deletion + doc edits (typo 아님) |
| 2 background | no | Write/Edit 수반 (read-only 아님) |
| 3 agent-team | no | RETIRED (MoAI static team layer) |
| 4 parallel | no | coding-heavy 단일 도메인 → Anthropic coding parallelism caveat; sequential이 deletion safety에 유리 |
| 5 sub-agent | **selected** | coding-heavy, per-milestone sequential 구현 (M2→M3), PRESERVE 검증 per-milestone |
| 6 workflow | no | <30 파일, mechanical-uniform 아님 (semantic deletion + PRESERVE) |

**Decision: sub-agent** (sequential manager-develop, M2→M3, per-milestone orchestrator verification)

**Justification**: 순수 dead-code deletion이나 PRESERVE 대상(HookResponse type `response.go:11`, `resolver.go` 9 references, `moai hook agent` subcommand)이 명확해 per-milestone `go build`/`go test`/`deadcode` 검증이 필수. coding-heavy + 단일 도메인이라 병렬화 이득이 낮고(Anthropic coding-task parallelism caveat), 순차 sub-agent가 deletion safety에 가장 적합. Implementation Kickoff Approval은 사용자 AskUserQuestion 승인 완료(반자율 진행 모드 — per-milestone 검증).
