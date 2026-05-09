## SPEC-V3R2-RT-002 — Permission Stack + Bubble Mode

> **Phase**: v3.0.0 — Phase 2 — Runtime Hardening
> **Module**: `internal/permission/`
> **Priority**: P0 Critical
> **Breaking**: false (additive — non-breaking on top of master §8 BC-V3R2-015 multi-layer settings reader)
> **Lifecycle**: spec-anchored

## Summary

Plan documents for SPEC-V3R2-RT-002 are ready for plan-auditor review. This SPEC replaces moai's flat `permissions.allow` list with a **typed 8-source permission stack** (`SrcPolicy > SrcUser > SrcProject > SrcLocal > SrcPlugin > SrcSkill > SrcSession > SrcBuiltin`) that resolves every tool invocation through an ordered precedence chain and carries provenance (`Origin`/`ResolvedBy` per rule). Introduces `bubble` as a first-class `PermissionMode` value so fork agents that inherit parent context can escalate permission decisions to the **parent terminal's user via AskUserQuestion** rather than silently default-allow or deny in an isolated mailbox. Closes problem-catalog **P-C01** (no permission bubble, CRITICAL) and **P-C04** (no config provenance, HIGH).

## Goal (purpose)

- **Structural (P6 Permission Bubble)**: 8-tier resolver replaces flat allowlist; every decision carries Origin (which file) + ResolvedBy (which tier).
- **Safety (P7 Sandbox Default)**: `bypassPermissions` rejected in `security.yaml strict_mode: true` environment.
- **UX (r3 §4 Adopt 2)**: Pre-allowlist of 8 dev-op patterns (`Bash(go test:*)`, `Read(*)`, `Glob(*)`, `Grep(*)`, etc.) absorbs ~80% of bubble-fatigue.
- **Hooks (P8 Hook JSON)**: PreToolUse hook PermissionDecision overlays at session-tier; UpdatedInput triggers re-match.

## Scope

### In-Scope

- `Source` 8-tier enum (consumed from `internal/config/source.go` — RT-005 ownership)
- `PermissionMode` 5-enum (`default`, `acceptEdits`, `bypassPermissions`, `plan`, `bubble`) — `bubble` is first-class peer
- `PermissionRule{Pattern, Action, Source, Origin}` Go struct
- `PermissionResolver.Resolve(tool, input, ctx)` walking 8-source stack in priority order
- Bubble-mode semantics: fork agent → parent AskUserQuestion (orchestrator-only per agent-common-protocol.md)
- PreToolUse hook integration (overlay above SrcSession tier)
- Pre-allowlist 8 patterns (`Bash(go test:*)`, `Bash(golangci-lint run:*)`, `Bash(ruff check:*)`, `Bash(npm test:*)`, `Bash(pytest:*)`, `Read(*)`, `Glob(*)`, `Grep(*)`)
- `moai doctor permission --all-tiers --tool --input --trace --dry-run --mode --fork --format` CLI subcommand
- Agent frontmatter `permissionMode` strict-validation (CI lint rejects unknown values)
- Plan-mode write-deny gate
- Strict_mode bypassPermissions reject
- Fork depth >3 → bubble degrade with sentinel
- Same-tier conflict resolution: specificity-then-fs-order tiebreak
- Legacy v2 `bypassPermissions` action migration with deprecation warning
- Non-interactive mode fail-closed (ask → deny + log)

### Out-of-Scope

- Hook JSON protocol wire format — SPEC-V3R2-RT-001
- Sandbox execution backends (bwrap/seatbelt) — SPEC-V3R2-RT-003
- Settings file provenance reader + multi-layer merge — SPEC-V3R2-RT-005
- Typed session state for resumable permission context — SPEC-V3R2-RT-004
- Plugin-origin settings ingestion — deferred to v3.1+ per master §7 plugin NOT-NOW
- UI prompt rendering for bubble (AskUserQuestion owns the UX)

## Non-Breaking Compatibility Note

**RT-002 is additive (`breaking: false`, `bc_id: []`).** v2.x users gain the typed layer transparently:

1. v2.x flat `permissions.allow` lists continue to read correctly via the tier-aware merge layer (master §8 BC-V3R2-015 — RT-005 reader auto-absorbs as `SrcProject`/`SrcLocal` tier rules).
2. The resolver answer becomes provenance-aware (carries `Origin`/`ResolvedBy`), but the *decision* derived from a v2 single-source allowlist matches the v3 resolver's output for identical input.
3. v2.x users see no behaviour change; v3 users gain `moai doctor permission --trace` for debugging.
4. Migration of `bypassPermissions` action (legacy v2 form) emits a deprecation warning naming the origin file — non-blocking for runtime; user can update at leisure.

This SPEC layers `bubble` mode + 8-tier walk + provenance on top of an existing v3 reader contract. No file format changes are required for v2 users.

## Plan Documents

| Document | Purpose | Size |
|----------|---------|------|
| `spec.md` | EARS requirements + ACs (existing, v0.1.0) | ~22 KB |
| `plan.md` | Implementation plan (M1-M5 milestones, traceability matrix, mx_plan, plan-audit-ready checklist) | new |
| `research.md` | Deep research (12 sections, 48 file:line anchors, library evaluation, dependency analysis, performance budget) | new |
| `acceptance.md` | 15 ACs in Given/When/Then format with edge cases + test mapping | new |
| `tasks.md` | 35 tasks (T-RT002-01..35) across M1-M5 with owner roles + dependencies | new |
| `progress.md` | Live progress tracker + paste-ready session-handoff resume message | new |

## EARS Requirements Summary

- **24 REQs across 6 categories**: Ubiquitous (8), Event-Driven (6), State-Driven (4), Optional (3), Unwanted (4), Complex (2)
- **15 ACs all map to REQs (100% coverage)**
- **Traceability matrix**: 27 REQ → 15 AC → 35 tasks (verified in `plan.md` §1.5)

## Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md`.

- **M1 (RED)**: ~10 failing tests covering AC-04, AC-05, AC-07, AC-08, AC-10, AC-11, AC-12, AC-13, AC-14, AC-15 + frontmatter audit lint for AC-09
- **M2 (GREEN part 1)**: PreAllowlist sync.Once + RejectIfStrict spawn helper + frontmatter strict-validation lint + security.yaml new keys + template mirror. AC-07, AC-09 GREEN.
- **M3 (GREEN part 2)**: hook UpdatedInput single-execution guard + IsWriteOperation pattern refinement + conflict.go specificity-then-fs-order tiebreak. AC-04, AC-10, AC-12, AC-13 GREEN.
- **M4 (GREEN part 3)**: bubble.go DispatchToParent IPC contract freezing + fork depth sentinel alignment + non-interactive fail-closed log. AC-08, AC-14, AC-15 GREEN.
- **M5 (GREEN part 4 + REFACTOR + Trackable)**: `moai doctor permission --all-tiers --mode --fork --format` CLI + migration.go (legacy bypassPermissions) + CHANGELOG + 7 MX tags. AC-05, AC-11 GREEN.

## Risks (Top 4)

1. **Bubble dispatch IPC channel undefined for v3.0** (M, M) — Mitigation: M4 freezes contract API only; actual wire deferred to orchestrator integration (skill body) coordinated with RT-001.
2. **Pre-allowlist 8 patterns insufficient for non-Go/JS/Python projects** (M, L) — Mitigation: post-beta.1 telemetry-driven amendment v0.2.0 (additional patterns for Rust/Java/Ruby/C++).
3. **Conflict tiebreak specificity formula may diverge from user expectation** (M, L) — Mitigation: explicit specificity formula in godoc + `--trace` JSON exposes specificity score.
4. **`internal/permission/` resolver mutex (RWMutex) deadlock under hook UpdatedInput re-entry** (M, H) — Mitigation: M3 inline single-execution guard verified by `TestResolve_HookUpdatedInputReMatch`.

Full 11-risk table in `plan.md` §5.

## Dependencies

### Blocked by

- SPEC-V3R2-RT-001 (Hook JSON protocol — provides `internal/hook.HookResponse` import for REQ-V3R2-RT-002-011/-013) — **at-risk**, mitigation: hardcoded fixture if unmerged
- SPEC-V3R2-RT-005 (Settings reader + 8-tier `internal/config.Source` enum) — **at-risk**, mitigation: hardcoded fixture if unmerged
- SPEC-V3R2-CON-001 (FROZEN zone declaration of 8-source ordering as constitutional clause) — completed per Wave 6

### Blocks

- SPEC-V3R2-RT-003 (Sandbox launcher consults PermissionMode for bwrap/seatbelt wrapping)
- SPEC-V3R2-ORC-001 (Agent roster reduction consumes validated permissionMode enum from REQ-008)
- SPEC-V3R2-ORC-004 (Worktree MUST for implementers pairs with permissionMode: acceptEdits at project tier)

### Related (non-blocking)

- SPEC-V3R2-MIG-001, SPEC-V3R2-SPC-004, SPEC-V3R2-CON-003, SPEC-V3R2-RT-004

## Plan-Audit-Ready Checklist

All 15 criteria PASS per `plan.md` §8:

- [x] Frontmatter v0.1.0 schema
- [x] HISTORY entry for v0.1.0
- [x] 24 EARS REQs across 6 categories
- [x] 15 ACs all map to REQs (100% coverage)
- [x] BC scope clarity (`breaking: false`, `bc_id: []` — additive)
- [x] File:line anchors ≥10 (research.md: 48, plan.md: 30)
- [x] Exclusions section present (6 entries explicitly mapped to other SPECs / X-4 deferred)
- [x] TDD methodology declared
- [x] mx_plan section (7 MX tag insertions across 4 categories — 3 ANCHOR + 2 NOTE + 2 WARN + 0 TODO)
- [x] Risk table with mitigations (spec.md: 6, plan.md: 11)
- [x] Worktree mode path discipline
- [x] No implementation code in plan documents
- [x] Acceptance.md G/W/T format with edge cases
- [x] tasks.md owner roles aligned with TDD methodology
- [x] Cross-SPEC consistency (RT-001/RT-005/CON-001 dependency status verified; RT-003/ORC-001/ORC-004 blocking confirmed)

## Worktree Discipline

[HARD] All run-phase work executes in `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002` on branch `plan/SPEC-V3R2-RT-002` (initial) or sibling `feature/SPEC-V3R2-RT-002-permission-stack` (run phase).

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`; tests use `t.TempDir()` per `CLAUDE.local.md` §6.

[HARD] Code comments in Korean; commit messages in Korean (per `.moai/config/sections/language.yaml`).

[HARD] `internal/template/templates/.moai/config/sections/security.yaml` mirror MUST be updated alongside the project-side `security.yaml` (Template-First Rule per CLAUDE.local.md §2).

## Test Plan

- [ ] M1 RED gate: ~10 new tests fail with documented sentinels; existing 69 tests in resolver_test.go/stack_test.go/bubble_test.go remain GREEN
- [ ] M2 GREEN part 1: AC-07, AC-09 turn GREEN
- [ ] M3 GREEN part 2: AC-04, AC-10, AC-12, AC-13 turn GREEN
- [ ] M4 GREEN part 3: AC-08, AC-14, AC-15 turn GREEN
- [ ] M5 GREEN part 4: AC-05, AC-11 turn GREEN; AC-01, AC-02, AC-03, AC-06 verified by existing tests + extension
- [ ] Full `go test ./...` passes with zero failures and zero cascading regressions
- [ ] `make build` regenerates `internal/template/embedded.go` cleanly (security.yaml mirror reflected)
- [ ] `go vet ./...` and `golangci-lint run` pass with zero warnings
- [ ] All required CI checks green: Lint, Test (ubuntu/macos/windows), Build (5 platforms), CodeQL

## Next Action

After plan-auditor PASS on this PR:
- Merge plan PR to `main`
- Verify SPEC-V3R2-RT-001 / SPEC-V3R2-RT-005 dependency status (gh pr list)
- Switch to run phase: `/moai run SPEC-V3R2-RT-002` (paste-ready resume in `progress.md`)

🗿 MoAI <email@mo.ai.kr>
