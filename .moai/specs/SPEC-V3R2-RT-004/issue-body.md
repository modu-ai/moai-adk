## SPEC-V3R2-RT-004 — Typed Session State + Phase Checkpoint

> **Phase**: v3.0.0 — Phase 2 — Runtime Hardening
> **Module**: `internal/session/`
> **Priority**: P1 High
> **Breaking**: false (additive typed schema on top of existing `.moai/state/`)
> **Lifecycle**: spec-anchored

## Summary

Plan documents for SPEC-V3R2-RT-004 are ready for plan-auditor review. This SPEC formalizes moai's cross-phase and cross-iteration state as a typed Go schema with immutable updates checkpointed at every phase boundary, closes problem-catalog **P-C02** (no sub-agent context isolation primitive) and **P-C05** (no cache-prefix discipline), and adds **`moai state {dump,show-blocker}`** CLI subcommand for human-readable inspection.

## Goal (purpose)

- **Structural (P5 + P6)**: Every phase boundary (`plan` → `run` → `sync`) writes a typed checkpoint file that captures all inputs the next phase needs. The next phase reads the file, not the prior conversation. Closes **P-C02**.
- **Safety (P4 + P11)**: Sprint Contract state (SPEC-V3R2-HRN-002) and evaluator fresh-judgment memory boundaries require a durable substrate that survives agent-level memory resets. Checkpoint files are that substrate.

## Scope

### In-Scope

- `Phase` typed enum (9 values: plan, run, sync, design, review, fix, loop, db, mx)
- `PhaseState` Go struct (`Phase, SPECID, Checkpoint, BlockerRpt, UpdatedAt, Provenance`)
- `Checkpoint` typed interface + `PlanCheckpoint`, `RunCheckpoint`, `SyncCheckpoint` concrete variants
- `BlockerReport` struct for `interrupt()`-equivalent surfacing
- `SessionStore` interface with 6 methods (`Checkpoint`, `Hydrate`, `AppendTaskLedger`, `WriteRunArtifact`, `RecordBlocker`, `ResolveBlocker`)
- Canonical `.moai/state/` file layout per master §5.6
- Provenance tagging on every state mutation (`SrcUser`/`SrcProject`/`SrcLocal`/`SrcSession`/`SrcHook` subset)
- Validator/v10 schema tags on all typed checkpoints; `Checkpoint()` validates before writing
- Crash-resume semantics: STALE_SECONDS primitive (default 3600s, configurable via `ralph.yaml`); on hydration, stale checkpoint prompts user via AskUserQuestion
- `moai state dump` and `moai state show-blocker` CLI subcommands
- `--resume` flag bypassing the staleness AskUserQuestion prompt
- Cross-platform advisory file locking (Unix `flock` + Windows `LockFileEx`) with 3-retry / 10ms-backoff
- Team-mode per-agent checkpoint merge with bubble-mode blocker routing

### Out-of-Scope

- Settings file multi-layer merge (resolver) — SPEC-V3R2-RT-005
- Hook JSON protocol — SPEC-V3R2-RT-001
- Sprint Contract durable state schema — SPEC-V3R2-HRN-002
- Ralph `/moai loop` outer-loop flow — SPEC-V3R2-WF-003
- Memdir typed taxonomy — SPEC-V3R2-EXT-001
- Agent `memory:` field in frontmatter — SPEC-V3R2-ORC-001

## Plan Documents

| Document | Purpose | Size |
|----------|---------|------|
| `spec.md` | EARS requirements + ACs (existing, v0.1.0) | 24180 bytes |
| `plan.md` | Implementation plan (M1-M5 milestones, traceability matrix, mx_plan, plan-audit-ready checklist) | new |
| `research.md` | Deep research (12 sections, 30 file:line anchors, library evaluation, dependency analysis) | new |
| `acceptance.md` | 15 ACs in Given/When/Then format with edge cases + test mapping | new |
| `tasks.md` | 33 tasks (T-RT004-01..33) across M1-M5 with owner roles + dependencies | new |
| `progress.md` | Live progress tracker + paste-ready session-handoff resume message | new |

## EARS Requirements Summary

- **24 REQs across 6 categories**: Ubiquitous (8), Event-Driven (6), State-Driven (3), Optional (4), Unwanted (4), Complex (2)
- **15 ACs all map to REQs (100% coverage)**
- **Traceability matrix**: 27 REQ → 15 AC → 33 tasks (verified in `plan.md` §1.4)

## Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md`.

- **M1 (RED)**: ~11 failing tests covering AC-04, AC-05, AC-06, AC-08, AC-09, AC-10, AC-14, AC-15 + audit lint for AC-11
- **M2 (GREEN part 1)**: validator/v10 tags + atomic-write helper. AC-01, AC-09, AC-15 GREEN.
- **M3 (GREEN part 2)**: cross-platform advisory locking. AC-10 GREEN.
- **M4 (GREEN part 3)**: blocker-file scan + filename rename + HydrateWithOpts + STALE_SECONDS config + in-flight detection + team merge. AC-04, AC-05, AC-06, AC-07, AC-08, AC-14 GREEN.
- **M5 (GREEN part 4 + REFACTOR + Trackable)**: `moai state` CLI + cache-prefix invariant + retention_days + AskUserQuestion audit + CHANGELOG + 7 MX tags. AC-11, AC-12, AC-13 GREEN.

## Risks (Top 3)

1. **Validator/v10 dependency from SPEC-V3R2-SCH-001 not yet merged** (M, H) — Mitigation: check `go.mod` at run-phase plan-audit gate; add direct dependency if absent.
2. **Cache-prefix invariant comment ignored by future contributors** (L, H) — Mitigation: `@MX:ANCHOR` tag on `hydrate.go`; future CI grep can verify the comment string preservation.
3. **Concurrent test races flaky on slow Windows runners** (M, L) — Mitigation: `t.Skip("requires fast disk")` if timing margin insufficient.

Full 11-risk table in `plan.md` §5.

## Dependencies

### Blocked by

- SPEC-V3R2-SCH-001 (validator/v10 — at-risk, mitigation defined)
- SPEC-V3R2-RT-005 (Source enum subset — referenced not blocking)
- SPEC-V3R2-CON-001 (FROZEN zone declaration — completed per Wave 6)

### Blocks

- SPEC-V3R2-HRN-002 (Sprint Contract durable state)
- SPEC-V3R2-WF-003 (Multi-mode router `/moai loop` Ralph layout)
- SPEC-V3R2-WF-004 (Agentless fixed pipelines audit trail)

### Related (non-blocking)

- SPEC-V3R2-RT-001, RT-002, EXT-001, EXT-004, ORC-001, CON-003

## Plan-Audit-Ready Checklist

All 15 criteria PASS per `plan.md` §8:

- [x] Frontmatter v0.1.0 schema
- [x] HISTORY entry for v0.1.0
- [x] 24 EARS REQs across 6 categories
- [x] 15 ACs all map to REQs (100% coverage)
- [x] BC scope clarity (`breaking: false`)
- [x] File:line anchors ≥10 (research.md: 30, plan.md: 18)
- [x] Exclusions section present (6 entries explicitly mapped to other SPECs)
- [x] TDD methodology declared
- [x] mx_plan section (7 MX tag insertions across 4 categories)
- [x] Risk table with mitigations (spec.md: 7, plan.md: 11)
- [x] Worktree mode path discipline
- [x] No implementation code in plan documents
- [x] Acceptance.md G/W/T format with edge cases
- [x] tasks.md owner roles aligned with TDD methodology
- [x] Cross-SPEC consistency (dependency status verified)

## Worktree Discipline

[HARD] All run-phase work executes in `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004` on branch `plan/SPEC-V3R2-RT-004` (initial) or sibling `feature/SPEC-V3R2-RT-004-typed-state` (run phase).

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`; tests use `t.TempDir()` per `CLAUDE.local.md` §6.

[HARD] Code comments in Korean; commit messages in Korean (per `.moai/config/sections/language.yaml`).

## Test Plan

- [ ] M1 RED gate: ~11 new tests fail with documented sentinels; existing tests still GREEN
- [ ] M2 GREEN part 1: AC-01, AC-09, AC-15 turn GREEN
- [ ] M3 GREEN part 2: AC-10 turns GREEN on Linux/macOS/Windows CI matrix
- [ ] M4 GREEN part 3: AC-04, AC-05, AC-06, AC-07, AC-08, AC-14 turn GREEN
- [ ] M5 GREEN part 4: AC-11, AC-12, AC-13 turn GREEN
- [ ] Full `go test ./...` passes with zero failures and zero cascading regressions
- [ ] `make build` regenerates `internal/template/embedded.go` cleanly
- [ ] `go vet ./...` and `golangci-lint run` pass with zero warnings
- [ ] All required CI checks green: Lint, Test (ubuntu/macos/windows), Build (5 platforms), CodeQL

## Next Action

After plan-auditor PASS on this PR:
- Merge plan PR to `main`
- Switch to run phase: `/moai run SPEC-V3R2-RT-004` (paste-ready resume in `progress.md`)

🗿 MoAI <email@mo.ai.kr>
