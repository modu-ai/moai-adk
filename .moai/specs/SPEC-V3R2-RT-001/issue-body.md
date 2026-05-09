## SPEC-V3R2-RT-001 — Hook JSON-OR-ExitCode Dual Protocol

> **Phase**: v3.0.0 — Phase 2 — Runtime Hardening
> **Module**: `internal/hook/`
> **Priority**: P0 Critical
> **Breaking**: **true** (BC-V3R2-001)
> **Lifecycle**: spec-anchored

## Summary

Plan documents for SPEC-V3R2-RT-001 are ready for plan-auditor review. 본 SPEC 는 Claude Code 2026.x 의 `HookJSONOutput` (JSON-OR-ExitCode) 프로토콜을 moai 의 27개 Claude Code 후크 이벤트 전체에 도입한다. typed `HookResponse` JSON 페이로드가 stdout 으로 emit 되며, JSON 파싱 실패 또는 빈 stdout 시 legacy exit-code 의미론으로 fallback 한다 (dual-parse). master §8 BC-V3R2-001 의 wrappers-unchanged + handlers-rewritten 호환성 shim 으로 26개 셸 래퍼는 v3.x 전체 deprecation window 동안 작동.

## Goal (purpose)

- **Structural (master §5.4)**: 5가지 programmable hook 능력 unlock — `additionalContext`, `updatedInput`, `permissionDecision`, `systemMessage`, `continue: false`. exit-code 만으로는 표현 불가능한 의미론.
- **Compatibility (master §8 BC-V3R2-001)**: dual-parse 가 v3.0 → v3.x deprecation window 동안 legacy exit-code-only hook 을 계속 수용. v4.0 에서 exit-code-only 경로 제거 예정.
- **Foundation for downstream Runtime SPECs**: RT-002 (permission stack), RT-003 (sandbox), RT-006 (handler completeness), SPC-002 (@MX routing), HRN-002 (evaluator fresh-memory injection) 모두 본 SPEC 의 protocol 위에서 진행.

## Scope

### In-Scope

- typed `HookResponse` Go struct + 7 canonical fields (`AdditionalContext`, `PermissionDecision`, `UpdatedInput`, `SystemMessage`, `Continue`, `WatchPaths`, `Retry`)
- `PermissionDecision` 4-value enum (allow / ask / deny / defer; "" = no opinion)
- `HookSpecificOutput` discriminator + 27 per-event variant types implementing `HookEventName() string`
- Dual-parse shim: JSON-first then exit-code synthesis (0→allow, 2→deny+stderr, non-zero→user msg)
- validator/v10 schema 검증 (SPEC-V3R2-SCH-001 의존성; M2 에서 직접 추가)
- Context-injection wiring (SessionStart/UserPromptSubmit/PreToolUse/PostToolUse → next model turn)
- Input-mutation wiring (PreToolUse `UpdatedInput` 가 pending tool input 교체)
- `Continue: false` escalation (SubagentStop → teammate idle blocker)
- Once-per-session deprecation banner; `MOAI_HOOK_LEGACY=1` opt-out for CI/air-gapped
- Opt-in strict mode via `.moai/config/sections/system.yaml` `hook.strict_mode: true` (default false)
- HookSpecificOutputMismatch detection — security boundary against forged event names
- 64 KiB AdditionalContext 절단 + truncation systemMessage notice
- api_version 2 opt-in (`# moai-hook-api-version: 2` shell header) — exit-code fallback skip
- WatchPaths registrar interface (SessionStart 등록)
- @MX 마커 라우팅 hook (PostToolUse `additionalContext` → SPEC-V3R2-SPC-002 ingestion)
- Plugin-source bypass (REQ-051) — `ConfigurationSource == "plugin"` 시 strict_mode 우회

### Out-of-Scope

- Handler completeness (27 events 의 비즈니스 로직 완성) — SPEC-V3R2-RT-006
- Permission stack 8-source 해결 — SPEC-V3R2-RT-002
- Sandbox 실행 — SPEC-V3R2-RT-003
- Typed session state (`Checkpoint`, `BlockerReport`) — SPEC-V3R2-RT-004
- Multi-layer settings provenance reader — SPEC-V3R2-RT-005
- Shell wrapper 의 hardcoded path 수정 — SPEC-V3R2-RT-007
- Hook type 확장 (`prompt`/`agent`/`http` beyond `command`) — v3.1+
- @MX 마커 자율 add/update/remove 의미론 — SPEC-V3R2-SPC-002

## Plan Documents

| Document | Purpose | Notes |
|----------|---------|-------|
| `spec.md` | EARS requirements + ACs (existing, v0.1.0) | 25 REQs across 6 categories, 15 ACs, breaking: true |
| `plan.md` | Implementation plan (M1-M5 milestones, traceability matrix, mx_plan, plan-audit-ready checklist) | new |
| `research.md` | Deep research (12 sections, 35 file:line anchors, library evaluation, BC-V3R2-001 분석) | new |
| `acceptance.md` | 15 ACs in Given/When/Then format with edge cases + test mapping | new |
| `tasks.md` | 43 tasks (T-RT001-01..43) across M1-M5 with owner roles + dependencies | new |
| `progress.md` | Live progress tracker + paste-ready session-handoff resume message | new |

## EARS Requirements Summary

- **25 REQs across 6 categories**: Ubiquitous (7), Event-Driven (7), State-Driven (3), Optional (3), Unwanted (3), Complex (2)
- **15 ACs all map to REQs (100% coverage)**
- **Traceability matrix**: 25 REQ → 15 AC → 43 tasks (verified in `plan.md` §1.4)

## Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md`.

- **M1 (RED)**: 18개 실패 테스트 작성 — strict_mode/banner/api_version/mismatch detection/AdditionalContext routing/64 KiB truncation/Continue:false escalation/27 variant round-trip/plugin bypass.
- **M2 (GREEN part 1)**: validator/v10 직접 의존성 + HookResponse Validate() 메서드. AC-12 GREEN.
- **M3 (GREEN part 2)**: strict_mode 분기 + MOAI_HOOK_LEGACY env + once-per-session banner + HookConfig + system.yaml template. AC-04, AC-06, AC-15 GREEN.
- **M4 (GREEN part 3)**: registry.go 의 4가지 의미 강화 — mismatch detection + UpdatedInput-then-Decision ordering + AdditionalContext routing + Continue:false escalation. AC-02, AC-05, AC-08, AC-09, AC-10 GREEN.
- **M5 (GREEN part 4 + REFACTOR + Trackable)**: api_version 2 + WatchPaths registrar + @MX routing + plugin bypass + CHANGELOG + 7 MX tags + make build + go test + vet + lint. AC-01, AC-03, AC-07, AC-11, AC-13, AC-14 GREEN.

## Breaking Change (BC-V3R2-001)

본 SPEC 는 frontmatter `breaking: true` + `bc_id: [BC-V3R2-001]`. 호환성 shim 의 핵심 약속:

- **Wrappers-unchanged**: `.claude/hooks/moai/*.sh` 26개(또는 27개) 셸 래퍼 모두 보존. 어느 것도 수정하지 않는다. master §8 BC-V3R2-001 의 핵심 약속.
- **Handlers-rewritten**: moai binary 의 hook subcommand 가 typed `HookResponse` JSON 을 stdout 에 emit. wrapper 는 stdin/stdout 을 transparent forwarding 하므로 수정 불요.
- **Dual-parse window**: v3.0 → v3.x 전체 minor cycle 동안 legacy exit-code 출력 수용. `MOAI_HOOK_LEGACY=1` opt-out (CI/air-gapped) + `hook.strict_mode: true` opt-in (early failure).
- **Sunset**: v4.0 에서 `synthesizeFromExitCode()`, `MOAI_HOOK_LEGACY` 분기, deprecation banner 제거 예정.

사용자 영향 매트릭스 (research.md §8 참고):

| 카테고리 | v3.0 동작 | 영향 |
|---------|----------|------|
| 일반 사용자 (moai binary 만) | JSON 자동 마이그레이션. 동일 기능. | None |
| Plugin author (외부 hook) | exit-code-only 가 deprecation window 동안 작동. session-당 1회 banner. | Cosmetic (banner). Suppressible via env. |
| CI / air-gapped | banner 노출 가능. `MOAI_HOOK_LEGACY=1` 으로 suppress. | Suppressible. |
| strict-mode 채택 팀 | `hook.strict_mode: true` 시 exit-code-only hook 즉시 reject. | Opt-in. Default off. |

## Risks (Top 3)

1. **validator/v10 SPEC-V3R2-SCH-001 미머지** (M, M) — Mitigation: M2 에서 직접 의존성 추가. SCH-001 후속 머지 시 dedup 자동.
2. **Discriminated-union mismatch 에러 메시지 가독성 부족** (M, M) — Mitigation: REQ-040 + AC-05 의 `HookSpecificOutputMismatch{Expected, Actual}` struct + `.moai/logs/hook.log` append.
3. **PreToolUse UpdatedInput 과 PermissionDecision ordering 모호성** (L, M) — Mitigation: REQ-041 + AC-10 의 deterministic order (UpdatedInput first, then PermissionDecision against updated input).

Full 11-risk table in `plan.md` §5.

## Dependencies

### Blocked by

- SPEC-V3R2-SCH-001 (validator/v10 — at-risk, M2 에서 직접 추가 mitigation)
- SPEC-V3R2-CON-001 (FROZEN zone — Wave 6 history 기준 완료)
- SPEC-V3R2-RT-005 (Source provenance — 본 SPEC 는 string literal `"plugin"` 으로 충족; 비-blocking)

### Blocks

- SPEC-V3R2-RT-002 (permission stack PermissionDecision consumer)
- SPEC-V3R2-RT-003 (sandbox UpdatedInput consumer)
- SPEC-V3R2-RT-006 (27 handler 비즈니스 로직 완성)
- SPEC-V3R2-SPC-002 (@MX 자율 add/update/remove)
- SPEC-V3R2-HRN-002 (evaluator fresh-memory injection)

### Related (non-blocking)

- SPEC-V3R2-MIG-001 (v2→v3 migrator)
- SPEC-V3R2-CON-003 (constitution consolidation)

## Plan-Audit-Ready Checklist

All 15 criteria PASS per `plan.md` §8:

- [x] Frontmatter v0.1.0 schema (id/title/version/status/created/updated/author/priority/phase/module/dependencies/bc_id/related_*/breaking/lifecycle/tags)
- [x] HISTORY entry for v0.1.0
- [x] 25 EARS REQs across 6 categories
- [x] 15 ACs all map to REQs (100% coverage)
- [x] BC scope clarity (`breaking: true`, `bc_id: [BC-V3R2-001]`)
- [x] File:line anchors ≥10 (research.md: 35, plan.md: 30)
- [x] Exclusions section present (8 entries explicitly mapped to other SPECs)
- [x] TDD methodology declared
- [x] mx_plan section (7 MX tag insertions across 4 categories: 3 ANCHOR + 2 NOTE + 2 WARN + 0 TODO)
- [x] Risk table with mitigations (spec.md: 6 risks, plan.md: 11 risks)
- [x] Worktree mode path discipline
- [x] No implementation code in plan documents
- [x] Acceptance.md G/W/T format with edge cases (15 ACs × 1-3 edge cases each)
- [x] tasks.md owner roles aligned with TDD methodology
- [x] Cross-SPEC consistency (3 blocked-by + 5 blocks + 5 related verified)

## Worktree Discipline

[HARD] All run-phase work executes in `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-001` on branch `plan/SPEC-V3R2-RT-001` (initial) or sibling `feat/SPEC-V3R2-RT-001-dual-protocol` (run phase).

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`; tests use `t.TempDir()` per `CLAUDE.local.md` §6.

[HARD] Code comments in Korean; commit messages in Korean (per `.moai/config/sections/language.yaml`). Godoc / 외부 노출 식별자 docstring 은 영어 유지.

[HARD] No `.claude/hooks/moai/*.sh` 파일 수정 — wrappers-unchanged. master §8 BC-V3R2-001 의 핵심 약속.

## Test Plan

- [ ] M1 RED gate: 18개 새 테스트 실패 + 기존 테스트 GREEN 회귀 0
- [ ] M2 GREEN part 1: AC-12 GREEN (validator/v10 통합)
- [ ] M3 GREEN part 2: AC-04, AC-06, AC-15 GREEN (strict_mode/banner/env)
- [ ] M4 GREEN part 3: AC-02, AC-05, AC-08, AC-09, AC-10 GREEN (registry 강화)
- [ ] M5 GREEN part 4: AC-01, AC-03, AC-07, AC-11, AC-13, AC-14 GREEN (api_version/WatchPaths/@MX/plugin)
- [ ] Full `go test ./...` passes with zero failures and zero cascading regressions
- [ ] `make build` regenerates `internal/template/embedded.go` cleanly (system.yaml template diff only)
- [ ] `go vet ./...` and `golangci-lint run` pass with zero warnings
- [ ] All required CI checks green: Lint, Test (ubuntu/macos/windows), Build (5 platforms), CodeQL

## Next Action

After plan-auditor PASS on this PR:
- Merge plan PR to `main`
- Switch to run phase: `/moai run SPEC-V3R2-RT-001` (paste-ready resume in `progress.md`)

🗿 MoAI <email@mo.ai.kr>
