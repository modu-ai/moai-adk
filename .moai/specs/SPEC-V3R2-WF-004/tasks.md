# SPEC-V3R2-WF-004 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.2.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-02 | MoAI Plan Workflow (Phase 1B)     | Initial task breakdown — 21 tasks (T-WF004-01..21) across M1-M5         |

---

## Task ID Convention

- ID format: `T-WF004-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `manager-docs`, `expert-backend` (Go test), `manager-git` (commit/PR boundary)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- DDD/TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M4 (GREEN) precedes M5 (REFACTOR + Trackable)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding (RED phase)

Goal: Create the audit test that will fail until M2-M4 add the required content. Per `spec-workflow.md:60-65` TDD: write failing test first.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF004-01 | Create `internal/template/agentless_audit_test.go` skeleton (package, imports, no test bodies yet) | expert-backend | `internal/template/agentless_audit_test.go` (new file, ~15 LOC scaffold) | none | 1 file (create) | RED setup |
| T-WF004-02 | Implement `TestAgentlessUtilityNoLLMControlFlow` walking 5 utility skills + applying 4 forbidden regexes per research.md §6.2 | expert-backend | `internal/template/agentless_audit_test.go` (test func, ~60 LOC) | T-WF004-01 | 1 file (extend) | RED — test must compile and fail (no Pipeline Contract sections yet) |
| T-WF004-03 | Implement `TestUtilitySkillsContainModeFlagIgnoredSentinel` asserting 5 utility skills contain `MODE_FLAG_IGNORED_FOR_UTILITY` | expert-backend | `internal/template/agentless_audit_test.go` (test func, ~30 LOC) | T-WF004-01 | 1 file (extend) | RED — must compile and fail (sentinel not yet added to any skill) |
| T-WF004-04 | Implement `TestImplementationSkillsContainPipelineRejectionSentinel` asserting 4 impl skills contain `MODE_PIPELINE_ONLY_UTILITY` | expert-backend | `internal/template/agentless_audit_test.go` (test func, ~30 LOC) | T-WF004-01 | 1 file (extend) | RED — must compile and fail (sentinel not yet added to any impl skill) |
| T-WF004-05 | Run `go test ./internal/template/ -run TestAgentless` and confirm RED state (3 functions × 5+5+4 subtests = 14 failing assertions, but compilation succeeds) | manager-tdd | n/a (verification only) | T-WF004-02, T-WF004-03, T-WF004-04 | 0 files | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones. M1 must complete before M2/M3/M4 begin (TDD discipline).

---

## M2: Pipeline Contract headers in 5 utility skills (GREEN, part 1)

Goal: Make `TestUtilitySkillsContainModeFlagIgnoredSentinel` pass and partially relieve `TestAgentlessUtilityNoLLMControlFlow` (which should already pass given research.md §4 confirms no current violations).

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF004-06 | Add `## Pipeline Contract (Agentless Classification)` section to `.claude/skills/moai/workflows/fix.md` after line 44 (end of Supported Flags). Phase mapping: localize ← Phase 1+2+2.5; repair ← Phase 3; validate ← Phase 4. Include `MODE_FLAG_IGNORED_FOR_UTILITY` sentinel verbatim. | manager-docs | `.claude/skills/moai/workflows/fix.md` (insert ~25 lines after L44) | T-WF004-05 | 1 file (edit) | GREEN |
| T-WF004-07 | Same as T-WF004-06 for `coverage.md` (after line 42). Mapping: localize ← Phase 1+2; repair ← Phase 3; validate ← Phase 4. | manager-docs | `.claude/skills/moai/workflows/coverage.md` | T-WF004-05 | 1 file (edit) | GREEN |
| T-WF004-08 | Same for `mx.md` (after line 59). Mapping: localize ← Pass 1+2; repair ← Pass 3; validate ← post-edit MX scan. | manager-docs | `.claude/skills/moai/workflows/mx.md` | T-WF004-05 | 1 file (edit) | GREEN |
| T-WF004-09 | Same for `codemaps.md` (after line 40). Mapping: localize ← Phase 1; repair ← Phase 2+3; validate ← Phase 4. | manager-docs | `.claude/skills/moai/workflows/codemaps.md` | T-WF004-05 | 1 file (edit) | GREEN |
| T-WF004-10 | Same for `clean.md` (after line 41). Mapping: localize ← Phase 1+2; repair ← Phase 4; validate ← Phase 5+5.5. | manager-docs | `.claude/skills/moai/workflows/clean.md` | T-WF004-05 | 1 file (edit) | GREEN |
| T-WF004-11 | Mirror T-WF004-06..10 edits into `internal/template/templates/.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md` (5 files) | manager-docs | `internal/template/templates/.claude/skills/moai/workflows/*.md` | T-WF004-06, 07, 08, 09, 10 | 5 files (edit, parity) | Embedded-template parity |
| T-WF004-12 | Run `make build` in worktree to regenerate `internal/template/embedded.go`. Verify diff is exactly the 5 utility skill content additions. | manager-docs | `internal/template/embedded.go` (regenerated) | T-WF004-11 | 1 file (regenerated) | Build verification |
| T-WF004-13 | Run `go test ./internal/template/ -run TestUtilitySkillsContainModeFlagIgnoredSentinel` and confirm PASS (5 of 5 utility subtests GREEN) | manager-tdd | n/a (verification only) | T-WF004-12 | 0 files | GREEN gate part 1 |

**M2 priority: P0** — blocks M5 (CHANGELOG cannot reference unfinished work).

T-WF004-06 through T-WF004-10 may execute in parallel — they touch independent files. T-WF004-11 must wait for all 5 source-of-truth edits (it depends on the content to mirror). T-WF004-12 and T-WF004-13 are sequential.

---

## M3: Mode Flag Compatibility clauses in 4 implementation skills (GREEN, part 2)

Goal: Make `TestImplementationSkillsContainPipelineRejectionSentinel` pass.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF004-14 | Add `## Mode Flag Compatibility` section to `.claude/skills/moai/workflows/{plan,run,sync,design}.md` (4 files). Section template per plan.md §2 M3, including `MODE_PIPELINE_ONLY_UTILITY` sentinel verbatim. | manager-docs | 4 skill files (insert ~12 lines each) | T-WF004-13 | 4 files (edit) | GREEN |
| T-WF004-15 | Mirror T-WF004-14 edits into `internal/template/templates/.claude/skills/moai/workflows/{plan,run,sync,design}.md` (4 files) | manager-docs | 4 template files | T-WF004-14 | 4 files (edit, parity) | Embedded-template parity |
| T-WF004-16 | Run `make build` to regenerate `internal/template/embedded.go`. Verify diff. | manager-docs | `internal/template/embedded.go` | T-WF004-15 | 1 file (regenerated) | Build verification |
| T-WF004-17 | Run `go test ./internal/template/ -run TestImplementationSkillsContainPipelineRejectionSentinel` and confirm PASS (4 of 4 impl subtests GREEN) | manager-tdd | n/a (verification only) | T-WF004-16 | 0 files | GREEN gate part 2 |

**M3 priority: P0** — required for cross-SPEC consistency with WF-003 REQ-WF003-016 sentinel.

[HARD] T-WF004-14 sub-edits MUST be insert-only. Existing flow declarations, phase orderings, agent delegations in plan/run/sync/design MUST NOT be modified.

---

## M4: Subcommand Classification matrix in `spec-workflow.md` (GREEN, part 3)

Goal: Publish the single source of truth matrix per REQ-WF004-005.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF004-18 | Insert `## Subcommand Classification (Pipeline vs Multi-Agent)` section in `.claude/rules/moai/workflow/spec-workflow.md` immediately after line 17 (end of Phase Overview table). Body per plan.md §2 M4 template. | manager-docs | `.claude/rules/moai/workflow/spec-workflow.md` (insert ~50 lines after L17) | T-WF004-17 | 1 file (edit) | GREEN |
| T-WF004-19 | Mirror T-WF004-18 edit into `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`. Run `make build`. | manager-docs | template file + `internal/template/embedded.go` | T-WF004-18 | 2 files (edit + regenerated) | Embedded-template parity |
| T-WF004-20 | Run full `go test ./...` from worktree root. Verify ALL 3 audit tests pass + 0 cascading failures across the rest of the repo (per `CLAUDE.local.md` §6 HARD rule "After fixing ANY test, run the FULL test suite"). | manager-tdd | n/a (verification only) | T-WF004-19 | 0 files | GREEN final gate |

**M4 priority: P1** — required deliverable but only blocks M5 (not M2/M3 which run in parallel).

T-WF004-18 may technically execute in parallel with M3 (it touches `spec-workflow.md`, not skills) — but plan.md sequences M4 after M3 so all GREEN gates pass simultaneously at T-WF004-20.

---

## M5: Documentation Sync + MX Tags + Cross-Links (REFACTOR + Trackable)

Goal: TRUST 5 Trackable + MX tag insertion + cross-link footers.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF004-21 | Add CHANGELOG entry under `## [Unreleased]` per plan.md §2 M5. Add cross-link footer to 9 skill files (`fix.md`, `coverage.md`, `mx.md`, `codemaps.md`, `clean.md`, `plan.md`, `run.md`, `sync.md`, `design.md`). Insert 10 MX tags across 8 distinct file locations per plan.md §6. Mirror to template tree. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`. Run `make build` + `go test ./...` final verification. | manager-docs | `CHANGELOG.md` (~5 lines), 9 skill files (cross-link footers, ~3 lines each), 8 files (MX tag inserts), 9 template mirrors, `progress.md` update, `internal/template/embedded.go` (regenerated) | T-WF004-20 | ~28 files | REFACTOR / Trackable |

**M5 priority: P2** — quality polish; could be split if size becomes unwieldy in the run phase.

[HARD] T-WF004-21's "9 skill files" cross-link footer is the M5 deliverable, NOT a 10th utility skill. The 5 utility + 4 implementation = 9 skills total (per plan.md §1.3).

[HARD] T-WF004-21 ends with a final `go test ./...` after `make build`. No tests may regress (per CLAUDE.local.md §6 HARD).

---

## Aggregate Statistics

- **Total tasks**: 21
- **Total milestones**: 5 (M1: 5 tasks, M2: 8 tasks, M3: 4 tasks, M4: 3 tasks, M5: 1 task)
- **Files created**: 1 (`internal/template/agentless_audit_test.go`)
- **Files modified (source-of-truth)**: 11
  - 5 utility skills (`fix.md`, `coverage.md`, `mx.md`, `codemaps.md`, `clean.md`)
  - 4 implementation skills (`plan.md`, `run.md`, `sync.md`, `design.md`)
  - 1 rule file (`spec-workflow.md`)
  - 1 changelog (`CHANGELOG.md`)
- **Files modified (embedded-template parity)**: 10 (mirrors of all 10 above except CHANGELOG, plus `internal/template/embedded.go` regenerated)
- **MX tag insertions**: 10 across 8 distinct file locations (per plan.md §6)
- **Owner-role distribution**:
  - `expert-backend`: 4 tasks (T-WF004-01..04, all Go test work)
  - `manager-tdd`: 4 tasks (T-WF004-05, 13, 17, 20 — TDD gate verification)
  - `manager-docs`: 13 tasks (all content authoring + template parity + final sync)

---

## Owner-Role Rationale

- **Go test work** (`expert-backend`): the audit test is Go code; needs Go expertise (regex, fs.WalkDir, embedded FS).
- **TDD gate verification** (`manager-tdd`): the project's `quality.yaml` declares `development_mode: tdd`, so manager-tdd is the methodology owner. Each gate (RED, GREEN parts 1/2/3, final GREEN) is a manager-tdd checkpoint.
- **Content authoring** (`manager-docs`): all skill/rule/CHANGELOG edits are documentation per `.claude/rules/moai/development/coding-standards.md` (skills are config-as-code documents). manager-docs is the owner per `CLAUDE.md` §4 Manager Agents catalog.

---

## Parallel Execution Opportunities

These task groups have no inter-dependencies and may run in parallel within their milestone:

- **M1 parallel**: T-WF004-02, T-WF004-03, T-WF004-04 (all extend the same file; cannot truly parallelize without merge conflict — execute sequentially in same edit pass)
- **M2 parallel**: T-WF004-06, T-WF004-07, T-WF004-08, T-WF004-09, T-WF004-10 (touch 5 independent files — true parallelism possible, but per `CLAUDE.md` §14 Multi-File Decomposition HARD rule with 3+ files, decomposition + sequencing recommended)
- **M3**: T-WF004-14 touches 4 files but is a single edit pass per plan.md §2 M3 (single task ID).

Per `CLAUDE.md` §1 HARD rule "Multi-File Decomposition: Split work when modifying 3+ files," M2 SHOULD be split into per-file sub-edits even if executed sequentially. The task IDs T-WF004-06..10 already encode this.

---

## Cross-Reference Map

Each task references which REQ(s) and which AC(s) it advances toward DoD:

| Task ID | REQ coverage | AC coverage |
|---------|--------------|-------------|
| T-WF004-01 | (scaffold) | (scaffold) |
| T-WF004-02 | REQ-WF004-013 | AC-WF004-12 |
| T-WF004-03 | REQ-WF004-011 | AC-WF004-10 |
| T-WF004-04 | REQ-WF004-014 | AC-WF004-11 |
| T-WF004-05 | (gate) | (gate) |
| T-WF004-06 | REQ-WF004-001, 003, 004, 006, 007, 008, 011, 015 | AC-WF004-01, 07, 08, 10 |
| T-WF004-07 | REQ-WF004-001, 003, 011 | AC-WF004-02, 10 |
| T-WF004-08 | REQ-WF004-001, 003, 011 | AC-WF004-03, 10 |
| T-WF004-09 | REQ-WF004-001, 003, 011 | AC-WF004-04, 10 |
| T-WF004-10 | REQ-WF004-001, 003, 011 | AC-WF004-05, 10 |
| T-WF004-11 | (parity) | (parity) |
| T-WF004-12 | (build) | (build) |
| T-WF004-13 | (gate) | AC-WF004-10 GREEN |
| T-WF004-14 | REQ-WF004-002, 010, 014 | AC-WF004-11, 13 |
| T-WF004-15 | (parity) | (parity) |
| T-WF004-16 | (build) | (build) |
| T-WF004-17 | (gate) | AC-WF004-11 GREEN |
| T-WF004-18 | REQ-WF004-005 | AC-WF004-06 |
| T-WF004-19 | (parity + build) | AC-WF004-06 (embedded parity) |
| T-WF004-20 | (final gate) | All 13 ACs |
| T-WF004-21 | (Trackable / MX) | (DoD steps 8-10) |

REQ coverage summary: 15 REQs, all referenced by at least one task (verified by transitive lookup against `spec.md` §5).

AC coverage summary: 13 ACs, all advanced toward DoD by tasks (AC-WF004-09 is implicit via the 5 utility Pipeline Contract sections in M2 + the test in T-WF004-02).

---

## Final Verification Pass (subset of T-WF004-21)

Before marking SPEC implementation complete, execute these checks in order:

1. **Audit tests green**: `go test ./internal/template/ -run TestAgentless -v` — 3 funcs, 14 subtests, all PASS.
2. **Full repo green**: `go test ./...` — 0 failures (per `CLAUDE.local.md` §6 HARD).
3. **Lint clean**: `golangci-lint run` — 0 errors, ≤10 warnings (per `quality.yaml` `lsp_quality_gates.sync`).
4. **Vet clean**: `go vet ./...` — 0 issues.
5. **Embedded parity**: diff `internal/template/embedded.go` after `make build` shows only the 11 expected file content changes (no spurious diffs).
6. **DoD checklist**: all 10 items in `acceptance.md` §"Definition of Done" checked.

If any step fails, return to the appropriate milestone for remediation. Do NOT advance to `/moai sync` until all 6 checks pass.

---

End of tasks.md.
