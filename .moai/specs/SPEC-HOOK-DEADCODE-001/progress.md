---
id: SPEC-HOOK-DEADCODE-001
title: "internal/hook package dead-code cleanup (3 corroborated scopes)"
version: "0.1.0"
status: completed
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

### M2 — COMPLETED 2026-07-04 (commit `f7bfa86a9`)

- **Commit**: `f7bfa86a9` — `refactor(SPEC-HOOK-DEADCODE-001): M2 dual_parse.go 무호출 함수5개+테스트 삭제 (HookResponse 타입 보존)` (2026-07-04)
- **Stats** (`git show --stat f7bfa86a9`): 4 files changed, 635 deletions(−)
- **Files deleted** (2 whole files):
  - `internal/hook/dual_parse.go` (172 LOC — 5 dead funcs: ParseHookOutput, synthesizeFromExitCode, ValidateHookResponse, ToHookOutput, ToHookResponse; 2 package-level error vars: ErrHookProtocolLegacyRejected, ErrHookInvalidPermissionDecision)
  - `internal/hook/dual_parse_test.go` (375 LOC — dedicated test for the deleted functions)
- **Files edited** (2):
  - `internal/hook/response.go` (2 LOC) — line-9 doc-comment reference to `ParseHookOutput` removed; `type HookResponse struct {` now at line 9
  - `internal/hook/response_test.go` (86 LOC) — 3 dependent test functions removed: `TestPermissionDecisionValues`, `TestHookResponseContinue`, `TestHookResponseAdditionalContextTruncation`
- **PRESERVE confirmed** (orchestrator independent verification 2026-07-12): `HookResponse` type (`response.go:9`) intact; `internal/permission/resolver.go` HookResponse references = **9** (unchanged baseline).
- **Orchestrator independent verification** (2026-07-12, local main `89a5dd086`): `go build ./...` exit 0; `go test ./internal/hook/...` all `ok` (green); dead-func caller grep EMPTY (exit 1); stale-import grep EMPTY (exit 1); `response_test.go` retains only 3 unrelated tests (`TestHookResponseEventNames`, `TestHookResponseMarshalUnmarshal`, `TestRetryHint`).

### M3 — COMPLETED 2026-07-12 (orchestrator-direct; commit pending this session)

- **Scope** (narrowed per plan-audit D3 remediation — the hooks-system.md:322 false-attribution item was removed as already-correct, verified by orchestrator cross-check 2026-07-12: live line 323 carries the correct `handle-agent-hook.sh` PreToolUse/PostToolUse/SubagentStop attribution):
  1. `internal/cli/hook.go` `runAgentHook`: removed the dead `input.Data = actionJSON` injection block (the `json.Marshal` struct + assignment, ~7 LOC incl. 2-line comment). The `HookInput.Data` field declaration (`internal/hook/types.go:316`) is RETAINED per spec §A.3 — the writer was removed because the field has zero readers repo-wide.
  2. `agent-hooks.md` "Agent Hook Actions" table: added the missing `| sync-auditor | - | - | evaluator-completion |` row — in BOTH the live copy (`.claude/rules/moai/core/agent-hooks.md`) AND the template mirror (`internal/template/templates/.claude/rules/moai/core/agent-hooks.md`), byte-identical.
- **PRESERVE confirmed** (orchestrator verification 2026-07-12): `HookInput.Data` (`types.go:316`) + `HookOutput.Data` (`types.go:380`) field declarations both intact; `moai hook agent` subcommand registration + `registry.Dispatch` call unchanged; 4 agent-frontmatter `hooks:` blocks untouched.
- **Orchestrator independent verification** (2026-07-12): `go build ./...` exit 0; `go test ./internal/hook/... ./internal/cli/...` all `ok` (green); `input.Data = ` writer grep EMPTY; `diff live template` IDENTICAL; `evaluator-completion` row count = 1 in both copies.

## §E.3 Run-phase Audit-Ready Signal

Run-phase COMPLETE — all 3 milestones executed and independently verified by the orchestrator (2026-07-12):
- **M1** `43c996bb7` — agents/+lifecycle/ deletion (18 files / 2874 LOC) + agent-hooks.md Handler Architecture correction.
- **M2** `f7bfa86a9` — dual_parse.go whole-file deletion (5 funcs + 2 error vars) + 3 dependent response_test.go functions + response.go:9 doc.
- **M3** (this session, orchestrator-direct) — HookInput.Data dead-injection removal + evaluator-completion doc row (live + template mirror).

Aggregate verification: `go build ./...` exit 0, `go test ./internal/hook/... ./internal/cli/...` green, dead-func caller grep EMPTY, stale-import grep EMPTY, mirror parity IDENTICAL. All PRESERVE targets intact across M1-M3 (HookResponse type `response.go`, `resolver.go` 9 references, `moai hook agent` subcommand, HookInput/HookOutput.Data field declarations). Ready for sync-phase (manager-docs).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-07-12
sync_commit_sha: pending-backfill-hook-deadcode
```

All 12 acceptance criteria (AC-HDC-001..012) confirmed PASS against `acceptance.md` (SSOT). CHANGELOG.md `[Unreleased] § Removed` entry added (M1/M2/M3 summary, 3500 LOC net removal, zero behavior change). Frontmatter `status: in-progress → completed` applied to spec.md, plan.md, acceptance.md, progress.md in the single sync commit (merged 3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001) — no README change (pure internal dead-code cleanup, zero user-facing CLI surface). `sync_commit_sha` is a placeholder pending backfill in a follow-up commit (self-referential-hazard exemption per spec-frontmatter-schema.md).

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
