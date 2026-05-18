# SPEC-V3R2-HRN-001 Run-Phase Progress

## Status

acceptance_phase_status: all-pass (with documented divergences)

## Milestones

- [x] M1 RED: Test fixtures + failing tests (T-01~05)
- [x] M2 GREEN: HarnessConfig struct + LoadHarnessConfig() (T-06~09)
- [x] M3 GREEN: Router logic (T-10~13)
- [x] M4 GREEN: Escalation + Effort + CLI (T-14~21)
- [x] M5 REFACTOR: MX tags + CHANGELOG + verification (T-22~25)

## Acceptance Criteria Results

| AC | Description | Status | Evidence |
|----|-------------|--------|----------|
| AC-01 | Tests pass ≥ 85% coverage | PASS | `go test -cover ./internal/harness/router/...` → 88.8% |
| AC-02 | Standard routing for multi-domain SPEC | PASS (with caveat) | SPEC-TEST-ORC-LIKE-001 → standard; ORC-001 routes thorough (P0 Critical triggers force_thorough — correct behavior) |
| AC-03 | Thorough routing for security keywords | PASS | SPEC-TEST-AUTH-001 → thorough, keywords=[oauth,jwt,session,encrypt] |
| AC-04 | Shipping harness.yaml validates clean | PASS | `moai harness validate` → exit 0 |
| AC-05 | pass_threshold < 0.60 → ErrPassThresholdFloor | PASS | Unit test `TestLoadHarnessConfigExtended_InvalidThreshold` |
| AC-06 | JSON output schema conformance | PASS | `moai harness route --spec SPEC-TEST-ORC-LIKE-001 --json` → all fields present |
| AC-07 | MOAI_CONFIG_STRICT=1 → HRN_SCHEMA_DRIFT error | PASS | exit 1 + HRN_SCHEMA_DRIFT error message |
| AC-08 | Escalation cap enforcement | PASS | `TestEscalationCapEnforcement` PASS |
| AC-09 | spec_override matched_rule | PASS | SPEC-TEST-OVERRIDE-001 → thorough + spec_override |
| AC-10 | Effort mapping: medium/high/xhigh | PASS | `TestEffortForLevel` PASS |

## Known Divergences

### AC-02: ORC-001 routes thorough (not standard as AC specifies)

The acceptance.md says "SPEC-V3R2-ORC-001 with frontmatter priority: P1" but the real SPEC has "P0 Critical".
"P0 Critical" triggers `isCriticalPriority()` → force_thorough. This is CORRECT behavior per REQ-HRN-001-008.
AC-02 used SPEC-TEST-ORC-LIKE-001 (P1, multi-domain) as verified fixture.

### AC-05: LoadHarnessConfig floor validation

The FROZEN floor validation happens via `validateEvaluatorProfileFloors()` in the CLI validate command
(when profile files are accessible), and via unit tests for the loader. The loader alone does not
parse .md profile files — it validates YAML structure and level enums. The full floor enforcement
requires both the loader (level enum) and the CLI (profile file parsing).

## Quality Gates

- [x] `golangci-lint run ./internal/config/... ./internal/harness/router/... ./internal/cli/...` → 0 issues
- [x] `go test -race -count=1 ./internal/config/... ./internal/harness/router/...` → all PASS
- [x] `go test -cover ./internal/harness/router/...` → 88.8% (≥ 85% target)
- [x] `moai spec lint --strict` → 0 errors
- [x] `moai harness validate` → exit 0
- [x] `TestHarnessRetirement` CI guard → PASS (retired lifecycle verbs not registered)
- [x] CHANGELOG.md updated
- [x] MX tags added to all new exported symbols

## New Files Created

- `internal/harness/router/router.go` — Level, Rationale, Router interface, defaultRouter
- `internal/harness/router/complexity.go` — ComplexitySignals, ExtractSignals
- `internal/harness/router/keywords.go` — security/payment keyword sets, matchForceThoroughKeywords
- `internal/harness/router/escalation.go` — EscalationManager with max_escalations cap
- `internal/harness/router/effort.go` — EffortForLevel, ParseProfileFloor
- `internal/harness/router/router_test.go` — Route tests
- `internal/harness/router/escalation_test.go` — Escalation tests
- `internal/harness/router/effort_test.go` — Effort tests
- `internal/harness/router/complexity_test.go` — Complexity signal tests
- `internal/cli/harness_route.go` — `moai harness route` CLI
- `internal/cli/harness_validate.go` — `moai harness validate` CLI
- `internal/cli/harness_route_test.go` — CLI tests
- `internal/config/loader_harness_extended_test.go` — Extended loader tests
- `internal/config/testdata/harness-valid/` — Valid fixture
- `internal/config/testdata/harness-invalid-threshold/` — Floor violation fixture
- `internal/config/testdata/harness-invalid-level/` — Unknown level fixture
- `internal/config/testdata/harness-drift-strict/` — Schema drift fixture
- `internal/harness/router/testdata/` — Router test SPEC fixtures

## Modified Files

- `internal/config/types.go` — HarnessConfig extended (7 new fields + sub-structs)
- `internal/config/errors.go` — 4 new sentinels (ErrUnknownLevel, ErrPassThresholdFloor, ErrSchemaDrift, ErrEscalationCapExceeded)
- `internal/config/loader.go` — LoadHarnessConfig() extended (level enum, drift detection)
- `internal/spec/lint.go` — SPECFrontmatter.HarnessLevel field + ExtractFrontmatter export
- `internal/cli/root.go` — newHarnessRouterCmd() registered
- `internal/cli/harness_retirement_test.go` — Updated to allow route/validate verbs
- `CHANGELOG.md` — HRN-001 entry added
