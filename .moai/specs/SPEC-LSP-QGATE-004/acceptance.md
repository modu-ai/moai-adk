# SPEC-LSP-QGATE-004: Acceptance Criteria

## Functional

### AC1: All 11 Orphaned Keys Wired
- [ ] `plan.require_baseline` enforced: plan phase captures baseline state (REQ-QG-002)
- [ ] `run.max_errors` enforced: run phase blocks on error threshold breach (REQ-QG-003)
- [ ] `run.max_type_errors` enforced separately from max_errors (REQ-QG-003)
- [ ] `run.max_lint_errors` enforced separately (REQ-QG-003)
- [ ] `run.allow_regression` enforced: regression comparison with baseline (REQ-QG-003, REQ-QG-007)
- [ ] `sync.max_errors` enforced: sync phase validates zero errors by default (REQ-QG-004)
- [ ] `sync.max_warnings` enforced (already existed) (REQ-QG-004)
- [ ] `sync.require_clean_lsp` enforced: blocks sync if non-dismissed diagnostics remain (REQ-QG-004)
- [ ] `cache_ttl_seconds` passed to Aggregator (REQ-QG-008)
- [ ] `timeout_seconds` passed to Aggregator (REQ-QG-008)
- [ ] `regression_detection.error_increase_threshold` applied (REQ-QG-007)
- [ ] `regression_detection.warning_increase_threshold` applied (REQ-QG-007)
- [ ] `trust5_integration.tested` influences tested-dimension scoring (REQ-QG-009)

### AC2: Duplicate Struct Removed
- [ ] `qualityYAMLConfig` struct deleted from `gate.go` (REQ-QG-006)
- [ ] All consumers use `pkg/models.QualityConfig` (REQ-QG-006)
- [ ] `grep -r qualityYAMLConfig internal/` returns zero results (REQ-QG-006)

### AC3: ConfigManager Integration
- [ ] `gate.go` does not call `os.ReadFile` for config (REQ-QG-005)
- [ ] Uses `ConfigManager.Load()` passed via DI (REQ-QG-005)
- [ ] Env var overrides (e.g., `MOAI_QUALITY_LSP_ENABLED`) work (REQ-QG-005)

### AC4: Phase Context
- [ ] Each phase (plan/run/sync) has distinct enforcement code paths (REQ-QG-001, REQ-QG-002, REQ-QG-003, REQ-QG-004)
- [ ] Unit tests exist for each phase (REQ-QG-001)
- [ ] Hook invocation correctly passes phase context (REQ-QG-010)

### AC5: Baseline + Regression
- [ ] `baseline.go` captures LSP state to `.moai/state/lsp-baseline.json` (REQ-QG-002)
- [ ] `regression.go` detects error count increase beyond threshold (REQ-QG-007)
- [ ] Regression block can be overridden via `allow_regression: true` (REQ-QG-003)

## Quality (TRUST 5)

### Tested
- [ ] ≥ 85% coverage for `internal/lsp/hook/`
- [ ] Golden file tests for each gate decision path
- [ ] Integration test with Aggregator mock

### Readable
- [ ] Phase enum has godoc explaining semantics
- [ ] `enforce*Phase` functions documented with enforcement rules

### Unified
- [ ] Uses canonical `pkg/models.QualityConfig`
- [ ] Phase enum reused across hooks where applicable

### Secured
- [ ] Config validation rejects invalid phase values
- [ ] Baseline file path is sandboxed to `.moai/state/`

### Trackable
- [ ] `@MX:ANCHOR` on `QualityGate.Run`
- [ ] SPEC-LSP-QGATE-004 in commit scopes

## Deliverables

- [ ] 5+ new Go files (baseline, regression, phase-specific enforcers)
- [ ] 3+ modified files (gate.go, deps.go, types.go)
- [ ] Deletion of qualityYAMLConfig
- [ ] 1 architecture doc
- [ ] All 11 orphaned config keys covered by enforcement or explicitly deprecated
