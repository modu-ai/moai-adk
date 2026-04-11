# SPEC-LSP-QGATE-004: Implementation Plan

## Phases

### Phase 1: Remove Duplicate Struct

**Goal**: Eliminate the `qualityYAMLConfig` duplicate and align all gate code with `pkg/models.QualityConfig`.

**Files**:
- `internal/lsp/hook/gate.go` (modify, ~50 LOC removed):
  - Delete `type qualityYAMLConfig struct { ... }`
  - Replace all references with `pkg/models.QualityConfig`
- `internal/lsp/hook/gate_test.go` (modify, ~30 LOC):
  - Update fixtures to use canonical QualityConfig

**Estimated LOC**: -50 removed, +30 modified

### Phase 2: ConfigManager Integration

**Goal**: Route all config reads through `ConfigManager` instead of direct yaml parsing.

**Files**:
- `internal/lsp/hook/gate.go` (modify, ~80 LOC):
  - Remove `os.ReadFile`/`yaml.Unmarshal` blocks
  - Accept `*config.Manager` via constructor (dependency injection)
  - Preserve environment variable overrides (MOAI_QUALITY_LSP_ENABLED, etc.)
- `internal/cli/deps.go` (modify, ~30 LOC):
  - Construct `QualityGate` with shared ConfigManager instance
- `internal/lsp/hook/gate_test.go` (modify, ~50 LOC):
  - Inject mock ConfigManager in tests

**Estimated LOC**: +160 modified

### Phase 3: Phase Context Plumbing

**Goal**: Give the gate awareness of which workflow phase (plan/run/sync) it is enforcing.

**Files**:
- `internal/lsp/hook/types.go` (new, ~60 LOC):
  - `type Phase int`
  - Constants: `PhasePlan`, `PhaseRun`, `PhaseSync`, `PhaseAuto`
  - `func (p Phase) String() string`
  - `func PhaseFromHookContext(ctx) Phase` — reads workflow state
- `internal/lsp/hook/gate.go` (modify, ~40 LOC):
  - Update `QualityGate.Run(ctx)` to `QualityGate.Run(ctx, phase Phase)`
- `internal/hook/post_tool.go` (modify, ~20 LOC):
  - Pass phase from hook context to gate
- `internal/lsp/hook/types_test.go` (new, ~80 LOC):
  - Phase enum serialization + context discovery tests

**Estimated LOC**: +140 new, +60 modified

### Phase 4: Enforcement Logic

**Goal**: Implement per-phase threshold enforcement using canonical `QualityConfig`.

**Files**:
- `internal/lsp/hook/enforce_plan.go` (new, ~100 LOC):
  - `enforcePlanPhase(ctx, cfg, diags) GateResult`
  - Reads `plan.require_baseline`; captures baseline if set
- `internal/lsp/hook/enforce_run.go` (new, ~150 LOC):
  - `enforceRunPhase(ctx, cfg, diags) GateResult`
  - Reads `run.max_errors`, `run.max_type_errors`, `run.max_lint_errors`, `run.allow_regression`
  - Classifies diagnostics by source (compiler vs linter) for separate thresholds
- `internal/lsp/hook/enforce_sync.go` (new, ~120 LOC):
  - `enforceSyncPhase(ctx, cfg, diags) GateResult`
  - Reads `sync.max_errors`, `sync.max_warnings`, `sync.require_clean_lsp`
  - Fails closed when non-dismissed diagnostics remain
- `internal/lsp/hook/gate_result.go` (new, ~80 LOC):
  - `type GateResult struct { Allow bool; Reason string; Details []FailureDetail }`
- `internal/lsp/hook/enforce_*_test.go` (new, ~300 LOC combined):
  - Table-driven tests per phase

**Estimated LOC**: +750 new

### Phase 5: Baseline + Regression Detection

**Goal**: Persist LSP baseline state and compare against current during run phase.

**Files**:
- `internal/lsp/hook/baseline.go` (new, ~180 LOC):
  - `CaptureBaseline(ctx, diags) error` — writes `.moai/state/lsp-baseline.json`
  - `LoadBaseline() (*Baseline, error)` — reads baseline, handles missing file
  - Atomic write via tmp file + rename
- `internal/lsp/hook/regression.go` (new, ~150 LOC):
  - `DetectRegression(baseline, current, cfg) *RegressionReport`
  - Applies `regression_detection.error_increase_threshold`
  - Applies `regression_detection.warning_increase_threshold`
- `internal/lsp/hook/baseline_test.go`, `regression_test.go` (new, ~250 LOC combined):
  - Round-trip tests, threshold edge cases, atomic write tests

**Estimated LOC**: +580 new

### Phase 6: trust5_integration Wiring

**Goal**: Map LSP diagnostics to TRUST 5 dimension scores per config.

**Files**:
- `internal/lsp/hook/trust5.go` (new, ~120 LOC):
  - `type Trust5Mapping struct { Tested, Readable, Understandable, Secured, Trackable bool }`
  - `ApplyTrust5(diags, mapping) Trust5Impact` — shapes diagnostic impact per dimension
- `internal/quality/evaluator.go` (modify if present, ~40 LOC):
  - Accept LSP-derived Trust5Impact as input to scoring
- `internal/lsp/hook/trust5_test.go` (new, ~150 LOC):
  - Each mapping combination tested

**Estimated LOC**: +270 new, +40 modified

### Phase 7: Tests + Documentation

**Goal**: Golden file tests, CHANGELOG, architecture doc.

**Files**:
- `internal/lsp/hook/golden/` (new test fixtures, ~20 JSON files):
  - One per phase × threshold combination
  - Expected GateResult JSON per fixture
- `internal/lsp/hook/golden_test.go` (new, ~200 LOC):
  - Reads golden fixtures, validates gate decisions
- `.claude/rules/moai/core/lsp-quality-gates.md` (new, ~250 LOC):
  - Architecture diagram
  - Phase semantics table
  - Config key → enforcement mapping (all 11 keys listed)
  - Baseline/regression flow
- `CHANGELOG.md` (modify, ~10 LOC)

**Estimated LOC**: +450 new

## Total Estimate

| Phase | New LOC | Modified LOC | Notes |
|-------|---------|--------------|-------|
| 1 | 0 | -20 | Struct removal |
| 2 | 0 | 160 | ConfigManager wiring |
| 3 | 140 | 60 | Phase plumbing |
| 4 | 750 | 0 | Enforcement logic |
| 5 | 580 | 0 | Baseline + regression |
| 6 | 270 | 40 | TRUST 5 wiring |
| 7 | 450 | 10 | Tests + docs |
| **Total** | **~2190** | **~250** | — |

Note: frontmatter `estimated_loc: 800` was based on a narrower initial scope. Actual breakdown including test code and golden fixtures brings ceiling closer to 2400, still within P2 budget.

## Dependencies

- **Hard**: SPEC-LSP-AGG-003 (uses Aggregator for diagnostic collection)
- **Blocks**: None (terminal SPEC in the LSP chain)

## Risks

| Risk | Mitigation |
|------|------------|
| Phase detection ambiguity | Phase 3 defines explicit enum + context discovery |
| Baseline file corruption | Atomic write via tmp + rename in Phase 5 |
| TRUST 5 scoring regression | Phase 6 gates behind config flag; default preserves current behavior |
| Orphan keys re-emerging | Each of the 11 keys has a dedicated AC in acceptance.md |
