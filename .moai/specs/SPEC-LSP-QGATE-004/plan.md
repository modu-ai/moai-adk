# SPEC-LSP-QGATE-004: Implementation Plan

## Phases

### Phase 1: Remove Duplicate Struct
- Delete `qualityYAMLConfig` from `internal/lsp/hook/gate.go`
- Replace all references with `pkg/models.QualityConfig`
- Update imports

### Phase 2: ConfigManager Integration
- Replace `os.ReadFile` + yaml.Unmarshal with `ConfigManager.Load()`
- Pass ConfigManager via dependency injection (deps.go)
- Preserve env var override support

### Phase 3: Phase Context Plumbing
- Add `Phase` enum to `internal/lsp/hook/types.go`
- Plumb phase from hook invocation context
- Update `QualityGate.Run(ctx, phase)` signature

### Phase 4: Enforcement Logic
- Implement `enforcePlanPhase`, `enforceRunPhase`, `enforceSyncPhase`
- Each reads phase-specific config (max_errors, max_type_errors, require_clean_lsp, etc.)
- Return structured GateResult with allow/deny + reason

### Phase 5: Baseline + Regression Detection
- `internal/lsp/hook/baseline.go`: capture and store LSP state
- `internal/lsp/hook/regression.go`: compare current vs baseline
- Apply `regression_detection` thresholds

### Phase 6: trust5_integration Wiring
- Parse `lsp_integration.trust5_integration` config
- Add LSP diagnostics to relevant TRUST 5 dimension scores
- Update evaluator-active profile if present

### Phase 7: Tests + Documentation
- Table-driven tests per phase
- Golden file tests for gate decisions
- Architecture doc in `.claude/rules/moai/core/lsp-quality-gates.md`

## Estimated LOC: +650 new, +150 modified

## Dependencies

- **Hard**: SPEC-LSP-AGG-003 (uses Aggregator for diagnostic collection)
- **Blocks**: None (terminal SPEC in the LSP chain)
