# SPEC-LSP-QGATE-004: Acceptance Criteria

## Functional

### AC1: All 11 Orphaned Keys Wired
- [ ] `plan.require_baseline` enforced: plan phase captures baseline state
- [ ] `run.max_errors` enforced: run phase blocks on error threshold breach
- [ ] `run.max_type_errors` enforced separately from max_errors
- [ ] `run.max_lint_errors` enforced separately
- [ ] `run.allow_regression` enforced: regression comparison with baseline
- [ ] `sync.max_errors` enforced: sync phase validates zero errors by default
- [ ] `sync.max_warnings` enforced (already existed)
- [ ] `sync.require_clean_lsp` enforced: blocks sync if non-dismissed diagnostics remain
- [ ] `cache_ttl_seconds` passed to Aggregator
- [ ] `timeout_seconds` passed to Aggregator
- [ ] `regression_detection.error_increase_threshold` applied
- [ ] `regression_detection.warning_increase_threshold` applied
- [ ] `trust5_integration.tested` influences tested-dimension scoring

### AC2: Duplicate Struct Removed
- [ ] `qualityYAMLConfig` struct deleted from `gate.go`
- [ ] All consumers use `pkg/models.QualityConfig`
- [ ] `grep -r qualityYAMLConfig internal/` returns zero results

### AC3: ConfigManager Integration
- [ ] `gate.go` does not call `os.ReadFile` for config
- [ ] Uses `ConfigManager.Load()` passed via DI
- [ ] Env var overrides (e.g., `MOAI_QUALITY_LSP_ENABLED`) work

### AC4: Phase Context
- [ ] Each phase (plan/run/sync) has distinct enforcement code paths
- [ ] Unit tests exist for each phase
- [ ] Hook invocation correctly passes phase context

### AC5: Baseline + Regression
- [ ] `baseline.go` captures LSP state to `.moai/state/lsp-baseline.json`
- [ ] `regression.go` detects error count increase beyond threshold
- [ ] Regression block can be overridden via `allow_regression: true`

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
