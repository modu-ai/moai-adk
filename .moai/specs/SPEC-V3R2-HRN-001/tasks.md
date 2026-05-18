# SPEC-V3R2-HRN-001 — Task Decomposition (tasks.md)

> Companion to `plan.md` (M1-M5 milestones).
> Companion to `acceptance.md` (10 binary AC).
> Companion to `research.md` (file:line anchors).

This file decomposes the M1-M5 milestones into atomic tasks `T-HRN001-NN`. Each task has owner role, file target with line anchor where applicable, dependencies, REQ↔AC references, and LOC estimate.

---

## HISTORY

| Version | Date       | Author | Description |
|---------|------------|--------|-------------|
| 0.1.0   | 2026-05-18 | manager-spec | Initial plan-phase decomposition: 25 tasks across M1-M5 |

---

## Task Format

| Field | Description |
|-------|-------------|
| ID | `T-HRN001-NN` (zero-padded 2-digit) |
| Subject | One-line task description |
| Owner | Role profile (tester / implementer / reviewer) |
| File | Target `internal/...` path with line anchor or NEW |
| Depends on | Prior task IDs (DAG) |
| REQ refs | REQ-HRN-001-NNN identifiers |
| AC refs | AC-HRN-001-NN identifiers |
| LOC est. | Lines-of-code estimate (rough) |

---

## M1 — RED Phase (Tests First)

### T-HRN001-01

| | |
|---|---|
| Subject | Create test fixture directory `internal/config/testdata/harness-valid/harness.yaml` mirroring shipping schema |
| Owner | tester |
| File | NEW: `internal/config/testdata/harness-valid/harness.yaml` |
| Depends on | — |
| REQ refs | REQ-HRN-001-001, REQ-HRN-001-002 |
| AC refs | AC-HRN-001-01, AC-HRN-001-04 |
| LOC est. | ~127 (mirror of shipping `.moai/config/sections/harness.yaml`) |

### T-HRN001-02

| | |
|---|---|
| Subject | Create test fixture `harness-invalid-threshold/` with lenient.md anchor at 0.5 |
| Owner | tester |
| File | NEW: `internal/config/testdata/harness-invalid-threshold/harness.yaml` + `evaluator-profiles/lenient.md` + `expected_error.txt` |
| Depends on | T-HRN001-01 |
| REQ refs | REQ-HRN-001-010, REQ-HRN-001-012 |
| AC refs | AC-HRN-001-05 |
| LOC est. | ~140 (yaml + md fixture + 5-line expected_error.txt) |

### T-HRN001-03

| | |
|---|---|
| Subject | Create test fixture `harness-invalid-level/` with `levels.expert:` unknown enum |
| Owner | tester |
| File | NEW: `internal/config/testdata/harness-invalid-level/harness.yaml` + `expected_error.txt` |
| Depends on | T-HRN001-01 |
| REQ refs | REQ-HRN-001-017 |
| AC refs | (regression) |
| LOC est. | ~130 |

### T-HRN001-04

| | |
|---|---|
| Subject | Create test fixture `harness-drift-strict/` with `unknown_top_field: 1` |
| Owner | tester |
| File | NEW: `internal/config/testdata/harness-drift-strict/harness.yaml` |
| Depends on | T-HRN001-01 |
| REQ refs | REQ-HRN-001-019 |
| AC refs | AC-HRN-001-07 |
| LOC est. | ~130 |

### T-HRN001-05

| | |
|---|---|
| Subject | Write failing loader test suite covering all 4 fixtures + golden path |
| Owner | tester |
| File | NEW: `internal/config/loader_harness_extended_test.go` |
| Depends on | T-HRN001-01, T-HRN001-02, T-HRN001-03, T-HRN001-04 |
| REQ refs | REQ-HRN-001-001, REQ-HRN-001-002, REQ-HRN-001-010, REQ-HRN-001-012, REQ-HRN-001-017, REQ-HRN-001-019 |
| AC refs | AC-HRN-001-01, AC-HRN-001-04, AC-HRN-001-05, AC-HRN-001-07 |
| LOC est. | ~240 (5 test funcs × ~45 LOC + helpers; mirrors `internal/config/loader_test.go:33-72` patterns) |

---

## M2 — GREEN Phase Part 1: Struct + Loader

### T-HRN001-06

| | |
|---|---|
| Subject | Extend `HarnessConfig` struct in `internal/config/types.go:401-404` with 6 new fields + sub-structs (`AutoDetectionConfig`, `EscalationConfig`, `LevelConfig`, `ModelUpgradeReviewConfig`, `PlanAuditGlobalConfig`, `ReviewChecklistItem`) |
| Owner | implementer |
| File | MODIFIED: `internal/config/types.go:401` (extension), `internal/config/types.go:480+` (new sub-structs appended) |
| Depends on | T-HRN001-05 (RED tests target this surface) |
| REQ refs | REQ-HRN-001-001 |
| AC refs | AC-HRN-001-01 |
| LOC est. | ~130 (6 fields + ~5 sub-structs with validator/v10 tags) |

### T-HRN001-07

| | |
|---|---|
| Subject | Add new sentinel errors: `ErrUnknownLevel`, `ErrPassThresholdFloor`, `ErrSchemaDrift`, `ErrEscalationCapExceeded` |
| Owner | implementer |
| File | MODIFIED: `internal/config/errors.go:68` (append after existing sentinels) |
| Depends on | T-HRN001-06 |
| REQ refs | REQ-HRN-001-010, REQ-HRN-001-012, REQ-HRN-001-017, REQ-HRN-001-019 |
| AC refs | AC-HRN-001-05, AC-HRN-001-07 |
| LOC est. | ~20 (4 sentinels × ~5 LOC docstrings + literal) |

### T-HRN001-08

| | |
|---|---|
| Subject | Extend `LoadHarnessConfig()` in `internal/config/loader.go:223-262` with validator/v10 invariants + MOAI_CONFIG_STRICT env probe + level enum check |
| Owner | implementer |
| File | MODIFIED: `internal/config/loader.go:223` |
| Depends on | T-HRN001-06, T-HRN001-07 |
| REQ refs | REQ-HRN-001-002, REQ-HRN-001-010, REQ-HRN-001-017, REQ-HRN-001-019 |
| AC refs | AC-HRN-001-01, AC-HRN-001-04, AC-HRN-001-07 |
| LOC est. | ~80 (extended function body + validator integration + drift detection helper) |

### T-HRN001-09

| | |
|---|---|
| Subject | Implement `validatePassThresholdFloor()` helper invoking `ParseRubricMarkdown()` for each `cfg.Levels[X].EvaluatorProfile` reference |
| Owner | implementer |
| File | NEW: `internal/harness/router/floor_validation.go` (helper); MODIFIED: `internal/config/loader.go` (calls helper from LoadHarnessConfig) |
| Depends on | T-HRN001-08 |
| REQ refs | REQ-HRN-001-012 |
| AC refs | AC-HRN-001-05 |
| LOC est. | ~60 (parse 4 profile files, check each anchor level >= 0.60) |

---

## M3 — GREEN Phase Part 2: Routing Logic

### T-HRN001-10

| | |
|---|---|
| Subject | Define `Level` type + constants in `internal/harness/router/router.go`; define `Rationale` struct + `Router` interface |
| Owner | implementer |
| File | NEW: `internal/harness/router/router.go` |
| Depends on | T-HRN001-06 (HarnessConfig types) |
| REQ refs | REQ-HRN-001-003 |
| AC refs | AC-HRN-001-02, AC-HRN-001-03 |
| LOC est. | ~60 (type defs + interface; impl in T-HRN001-11) |

### T-HRN001-11

| | |
|---|---|
| Subject | Implement `defaultRouter.Route(spec *spec.SPECDoc, cfg *config.HarnessConfig)` invoking complexity estimator + override checks in priority order |
| Owner | implementer |
| File | NEW: `internal/harness/router/router.go` (continues T-10) |
| Depends on | T-HRN001-10, T-HRN001-12, T-HRN001-13 |
| REQ refs | REQ-HRN-001-003, REQ-HRN-001-007, REQ-HRN-001-008, REQ-HRN-001-015 |
| AC refs | AC-HRN-001-02, AC-HRN-001-03, AC-HRN-001-09 |
| LOC est. | ~120 (Route method + helper functions) |

### T-HRN001-12

| | |
|---|---|
| Subject | Implement Complexity Estimator helpers: `signalFileCount()`, `signalDomainCount()`, `signalSpecType()`, `signalSpecPriority()` |
| Owner | implementer |
| File | NEW: `internal/harness/router/complexity.go` |
| Depends on | T-HRN001-10 |
| REQ refs | REQ-HRN-001-007, REQ-HRN-001-014 |
| AC refs | AC-HRN-001-02 |
| LOC est. | ~140 (4 signal functions + path-pattern regex matchers) |

### T-HRN001-13

| | |
|---|---|
| Subject | Implement keyword matchers + force-thorough override in `internal/harness/router/keywords.go` + spec_override hook |
| Owner | implementer |
| File | NEW: `internal/harness/router/keywords.go`; MODIFIED: `internal/spec/lint.go:268` (add `HarnessLevel string \`yaml:"harness_level"\`` to `SPECFrontmatter`) |
| Depends on | T-HRN001-10 |
| REQ refs | REQ-HRN-001-008, REQ-HRN-001-015 |
| AC refs | AC-HRN-001-03, AC-HRN-001-09 |
| LOC est. | ~80 (canonical keyword sets + matcher func + 1 LOC SPECFrontmatter extension) |

---

## M4 — GREEN Phase Part 3: Escalation + Effort + CLI

### T-HRN001-14

| | |
|---|---|
| Subject | Implement `EscalationManager` with `MaxEscalations` cap, `CheckTriggers(ctx)` method, cap-reached log emission |
| Owner | implementer |
| File | NEW: `internal/harness/router/escalation.go` |
| Depends on | T-HRN001-10 |
| REQ refs | REQ-HRN-001-004, REQ-HRN-001-009, REQ-HRN-001-013, REQ-HRN-001-018 |
| AC refs | AC-HRN-001-08 |
| LOC est. | ~120 (struct + 3 methods + log helpers) |

### T-HRN001-15

| | |
|---|---|
| Subject | Implement `EffortForLevel(level Level, cfg *config.HarnessConfig) string` with fallback defaults |
| Owner | implementer |
| File | NEW: `internal/harness/router/effort.go` |
| Depends on | T-HRN001-06 |
| REQ refs | REQ-HRN-001-005 |
| AC refs | AC-HRN-001-10 |
| LOC est. | ~40 |

### T-HRN001-16

| | |
|---|---|
| Subject | Implement `newHarnessRouterCmd()` cobra parent factory + register `route` and `validate` subcommands |
| Owner | implementer |
| File | NEW: `internal/cli/cmd/harness_route.go` + `internal/cli/cmd/harness_validate.go` |
| Depends on | T-HRN001-11, T-HRN001-08 |
| REQ refs | REQ-HRN-001-006 |
| AC refs | AC-HRN-001-02, AC-HRN-001-04 |
| LOC est. | ~80 (parent factory + 2 child factories with persistent flags) |

### T-HRN001-17

| | |
|---|---|
| Subject | Implement `route` command output (plaintext + `--json` mode per documented schema) |
| Owner | implementer |
| File | NEW: `internal/cli/cmd/harness_route.go` (continues T-16) |
| Depends on | T-HRN001-16 |
| REQ refs | REQ-HRN-001-011 |
| AC refs | AC-HRN-001-02, AC-HRN-001-03, AC-HRN-001-06, AC-HRN-001-09 |
| LOC est. | ~110 (JSON marshaling, error handling, plaintext formatting) |

### T-HRN001-18

| | |
|---|---|
| Subject | Implement `validate` command with `--path` flag, exit code semantics (0/1), model_upgrade_review reminder emission |
| Owner | implementer |
| File | NEW: `internal/cli/cmd/harness_validate.go` |
| Depends on | T-HRN001-08, T-HRN001-09 |
| REQ refs | REQ-HRN-001-006, REQ-HRN-001-016 |
| AC refs | AC-HRN-001-04, AC-HRN-001-05, AC-HRN-001-07 |
| LOC est. | ~100 |

### T-HRN001-19

| | |
|---|---|
| Subject | Register `newHarnessRouterCmd()` in `internal/cli/root.go` (after line 92); ensure retired `newHarnessCmd` remains unregistered |
| Owner | implementer |
| File | MODIFIED: `internal/cli/root.go:95` |
| Depends on | T-HRN001-16 |
| REQ refs | REQ-HRN-001-006 |
| AC refs | AC-HRN-001-02 |
| LOC est. | ~3 (1 import + 1 AddCommand line + comment) |

### T-HRN001-20

| | |
|---|---|
| Subject | Write CLI golden tests + JSON schema validation tests in `internal/cli/cmd/harness_route_test.go` |
| Owner | tester |
| File | NEW: `internal/cli/cmd/harness_route_test.go` + `internal/cli/cmd/harness_validate_test.go` |
| Depends on | T-HRN001-17, T-HRN001-18 |
| REQ refs | REQ-HRN-001-006, REQ-HRN-001-011 |
| AC refs | AC-HRN-001-02, AC-HRN-001-03, AC-HRN-001-04, AC-HRN-001-06, AC-HRN-001-09 |
| LOC est. | ~280 (table-driven tests covering all 10 ACs that exercise CLI) |

### T-HRN001-21

| | |
|---|---|
| Subject | Write router_test.go + escalation_test.go + effort_test.go with full AC coverage |
| Owner | tester |
| File | NEW: `internal/harness/router/router_test.go`, `internal/harness/router/escalation_test.go`, `internal/harness/router/effort_test.go`, `internal/harness/router/complexity_test.go` |
| Depends on | T-HRN001-11, T-HRN001-14, T-HRN001-15 |
| REQ refs | REQ-HRN-001-003, REQ-HRN-001-004, REQ-HRN-001-005, REQ-HRN-001-007, REQ-HRN-001-008, REQ-HRN-001-009, REQ-HRN-001-013, REQ-HRN-001-015, REQ-HRN-001-018 |
| AC refs | AC-HRN-001-02, AC-HRN-001-03, AC-HRN-001-08, AC-HRN-001-09, AC-HRN-001-10 |
| LOC est. | ~420 (4 test files × ~100 LOC avg) |

---

## M5 — REFACTOR + Final Verification

### T-HRN001-22

| | |
|---|---|
| Subject | Add `@MX:NOTE`, `@MX:ANCHOR`, `@MX:WARN` tags to all new exported symbols per `mx-tag-protocol.md` (code_comments: ko) |
| Owner | implementer |
| File | MODIFIED: all NEW files from M2-M4 |
| Depends on | T-HRN001-21 (all impl done) |
| REQ refs | (cross-cutting, all REQ) |
| AC refs | (DoD) |
| LOC est. | ~60 (annotations only; not net new code) |

### T-HRN001-23

| | |
|---|---|
| Subject | Update `CHANGELOG.md` `[Unreleased]` section with SPEC-V3R2-HRN-001 entry |
| Owner | reviewer |
| File | MODIFIED: `CHANGELOG.md` |
| Depends on | T-HRN001-21 |
| REQ refs | (DoD) |
| AC refs | (DoD) |
| LOC est. | ~12 |

### T-HRN001-24

| | |
|---|---|
| Subject | Run quality gates: `make preflight` + `golangci-lint run` + `go test -race -cover` ≥ 85% coverage |
| Owner | reviewer |
| File | (verification only — no code changes) |
| Depends on | T-HRN001-22, T-HRN001-23 |
| REQ refs | (DoD) |
| AC refs | AC-HRN-001-01 (coverage), AC-HRN-001-04 (lint clean) |
| LOC est. | 0 (verification task) |

### T-HRN001-25

| | |
|---|---|
| Subject | Run all 10 acceptance.md verification commands; record results in `progress.md` `acceptance_phase_status: all-pass` |
| Owner | reviewer |
| File | MODIFIED: `.moai/specs/SPEC-V3R2-HRN-001/progress.md` |
| Depends on | T-HRN001-24 |
| REQ refs | (DoD) |
| AC refs | AC-HRN-001-01 through AC-HRN-001-10 |
| LOC est. | ~30 (progress.md update with timestamps + evidence references) |

---

## Task DAG (dependency graph)

```
T-01 ─┐
T-02 ─┤
T-03 ─┤── T-05 (RED tests)
T-04 ─┘     │
            ▼
T-06 (struct) ── T-07 (errors) ── T-08 (loader) ── T-09 (floor)
                                       │
                                       ▼
                              T-10 (router types)
                                       │
                                       ├── T-12 (complexity)
                                       ├── T-13 (keywords + spec_override)
                                       │
                                       ▼
                              T-11 (Route() impl) ────┐
                                                       │
                              T-14 (Escalation) ──────┤
                              T-15 (EffortForLevel) ──┤
                                                       │
                                                       ▼
                              T-16 (CLI parent) ── T-17 (route subcmd)
                                                ── T-18 (validate subcmd)
                                                        │
                              T-19 (CLI root register) ─┤
                                                        │
                              T-20 (CLI golden tests) ──┤
                              T-21 (unit tests) ────────┤
                                                        ▼
                              T-22 (MX tags)
                                    │
                              T-23 (CHANGELOG)
                                    │
                              T-24 (preflight + lint + coverage)
                                    │
                              T-25 (acceptance run + progress.md)
```

---

## LOC Budget Summary

| Milestone | Net LOC estimate | Cumulative |
|-----------|------------------|------------|
| M1 (RED tests + fixtures) | ~770 | ~770 |
| M2 (struct + loader) | ~290 | ~1060 |
| M3 (router logic) | ~340 | ~1400 |
| M4 (escalation + effort + CLI) | ~960 | ~2360 |
| M5 (refactor + verification) | ~102 | ~2462 |

Total estimate: **~2,460 LOC** across ~14 new files + 5 modified files. Test code is ~40% (~970 LOC); production code is ~60% (~1,490 LOC).

[NOTE] This estimate is plan-phase rough sizing only — not a commitment. Actual LOC may diverge during M2-M4 implementation; run-phase pre-submission self-review (per `spec-workflow.md` § Run Phase) will reconcile against the SPEC acceptance criteria.

---

## Task Ownership Summary

| Owner role | Tasks | Total tasks |
|-----------|-------|-------------|
| tester | T-01, T-02, T-03, T-04, T-05, T-20, T-21 | 7 |
| implementer | T-06, T-07, T-08, T-09, T-10, T-11, T-12, T-13, T-14, T-15, T-16, T-17, T-18, T-19, T-22 | 15 |
| reviewer | T-23, T-24, T-25 | 3 |

---

End of tasks.md.
