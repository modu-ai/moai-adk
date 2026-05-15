# Sprint Contract — Wave A (SPEC-V3R4-HARNESS-003)

Negotiator: evaluator-active (Phase 2.0)
Counterparty: manager-develop (Wave A TDD execution)
Harness: thorough
AC Coverage: AC-HRN-CLS-001 (primary, must-pass) + AC-HRN-CLS-004 (implicit, must-pass)
Round: 1 of max 2
Contract date: 2026-05-15

---

## Scope Confirmation

Wave A = 4 tasks, exactly 5 files touched:

| File | Op |
|------|----|
| `internal/harness/learner.go` | MOD (additive only, after L91) |
| `internal/harness/learner_test.go` | MOD (new test appended) |
| `internal/harness/classifier_cluster.go` | NEW (stubs only) |
| `internal/harness/testdata/stage1_baseline.jsonl` | NEW (~150KB, 1000 lines) |
| `internal/harness/testdata/stage1_baseline_patterns.json` | NEW (~5KB, 10 keys) |

No other files may be modified. See § Hard Thresholds for the file-scope prohibition list.

---

## TDD Cycle Obligation

Wave A MUST follow strict TDD ordering:

1. **Pre-RED (Fixtures)**: Write T-A1 + T-A2 by invoking pre-modification `AggregatePatterns` with synthetic events. Golden pattern JSON generated via `MOAI_REGEN_GOLDEN=1` env gate or `//go:build ignore` helper — implementation's choice; both are acceptable. Keys MUST be sorted before `json.MarshalIndent`.
2. **RED**: Write T-A4 (`TestStage1BackwardCompat_StageDisabled`) first. Test must fail with compile error because `ClassifierConfig` does not yet exist.
3. **GREEN**: Add stubs to `classifier_cluster.go` + seam to `learner.go`. Test must pass.
4. **REFACTOR (MX)**: Add `@MX:NOTE` to the Stage-2 call site in `AggregatePatterns` citing REQ-HRN-CLS-001.

---

## Stub Signature Contract

`classifier_cluster.go` MUST declare these exact exported and unexported identifiers in `package harness`:

```go
// ClassifierConfig holds configuration for the Stage-2 embedding-cluster classifier.
// Stage2Enabled defaults to false (Go zero value), preserving backward compatibility.
type ClassifierConfig struct {
    Stage2Enabled bool
}

// clusterSingletons is the Stage-2 seam. Wave A stub returns patterns unchanged.
// Wave C will implement the full Union-Find Hamming clustering algorithm.
func clusterSingletons(
    patterns map[string]*Pattern,
    cfg ClassifierConfig,
    auditLogPath string,
) (map[string]*Pattern, error) {
    return patterns, nil
}
```

Rationale for no `events []Event` parameter in Wave A: learner.go's scanner loop body (L60-85) MUST NOT be modified in Wave A (criterion 9). Wave C will add event collection to the loop and update the call site signature at that time. Wave A's call site in `AggregatePatterns` passes `(patterns, ClassifierConfig{}, "")`.

`ClassifierConfig` MUST be the zero-value default inside `AggregatePatterns` for Wave A (no config loading until Wave D).

---

## `AggregatePatterns` Seam Placement

The seam call MUST be inserted as follows (pseudocode):

```
scanner loop (L60-85) — UNTOUCHED
scanner.Err() check (L87-89) — UNTOUCHED
[NEW CODE BEGINS HERE — after L89, before existing return]
  cfg := ClassifierConfig{} // zero value = Stage2 disabled
  if cfg.Stage2Enabled {
      var err error
      patterns, err = clusterSingletons(patterns, cfg, "")
      if err != nil {
          return nil, fmt.Errorf("learner: stage-2 clustering failed: %w", err)
      }
  }
[NEW CODE ENDS]
return patterns, nil  // existing return — line number shifts
```

`AggregatePatterns` public signature MUST remain: `func AggregatePatterns(logPath string) (map[string]*Pattern, error)`. No parameter additions.

---

## Done Criteria

Each criterion is an evaluation oracle for Phase 2.8a. Commands are project-root-relative.

### 1. Compilation

```bash
go build ./internal/harness/...
```

Expected: exits 0. Zero errors, zero type errors.

### 2. AC-001 — Backward Compatibility (Stage 2 OFF)

```bash
go test -run TestStage1BackwardCompat_StageDisabled ./internal/harness/ -v
```

Expected:
- exits 0
- `reflect.DeepEqual(actual, golden)` returns true (10 entries, 3-field keys, Count=100, Confidence=1.0)
- `os.Stat(auditLogPath)` returns an error satisfying `os.IsNotExist`

The test MUST use a fixture path from `internal/harness/testdata/stage1_baseline.jsonl` (read via `os.ReadFile` or direct path) and compare against `internal/harness/testdata/stage1_baseline_patterns.json` loaded with `json.Unmarshal`.

### 3. No Regression in Harness Suite

```bash
go test ./internal/harness/...
```

Expected:
- exits 0
- All pre-existing tests pass
- T-A4 new test passes
- Zero failures

### 4. Determinism (5 consecutive passes)

```bash
go test -run TestStage1BackwardCompat_StageDisabled -count=5 ./internal/harness/
```

Expected: exits 0 across all 5 invocations. Golden fixture must be deterministic (sorted keys, `time.Time{}` zero values in synthetic events).

### 5. Vet + Race Detector

```bash
go vet ./internal/harness/...
go test -race ./internal/harness/ -short
```

Expected: both exit 0.

### 6. No Stage-2 in Observer Hot Path

```bash
grep -c 'classifier_simhash\|classifier_cluster' internal/cli/hook.go internal/harness/observer.go
```

Expected: returns `0` (zero matches). Stage-2 symbols MUST NOT appear in the observer hot path or CLI hook handler.

### 7. No PromptContent in Classifier

```bash
grep -c 'PromptContent' internal/harness/classifier_cluster.go 2>/dev/null || echo 0
```

Expected: returns `0`. `PromptContent` field MUST NOT be referenced in the stub file.

### 8. Fixture Sanity

```bash
wc -l < internal/harness/testdata/stage1_baseline.jsonl
jq 'keys | length' internal/harness/testdata/stage1_baseline_patterns.json
```

Expected:
- First command: `1000` (exactly 1000 JSONL lines)
- Second command: `10` (exactly 10 top-level keys)

### 9. Diff Scope — Seam Placement

```bash
git diff main -- internal/harness/learner.go
```

Expected: additions appear ONLY after the `scanner.Err()` block (currently L87-89). NO additions inside the scanner loop body (L60-85). The diff MUST NOT show any `-` (deletion) lines in existing logic.

### 10. Stubs in Correct Location

```bash
grep -c 'type ClassifierConfig struct' internal/harness/classifier_cluster.go
grep -c 'func clusterSingletons' internal/harness/classifier_cluster.go
```

Expected: both return `1`.

### 11. Package Declaration

```bash
head -1 internal/harness/classifier_cluster.go
```

Expected: `package harness` (same package as learner.go — no sub-package, no import cycle).

### 12. `AggregatePatterns` Signature Unchanged

```bash
grep 'func AggregatePatterns' internal/harness/learner.go
```

Expected: `func AggregatePatterns(logPath string) (map[string]*Pattern, error)` — exactly one parameter, signature unchanged from pre-V3R4-003 baseline.

---

## Edge Cases Required Coverage

All 4 edge cases MUST be handled or confirmed safe before Phase 2.8a:

### EC-A1: Empty Pattern Map (Zero Events)

Input: `stage1_baseline.jsonl`-style path pointing to an empty file (0 bytes or 0 valid JSONL lines).

Behavior: `AggregatePatterns` returns empty `map[string]*Pattern{}`. The seam call `if cfg.Stage2Enabled` evaluates false. `clusterSingletons` is NOT invoked. Audit log does NOT exist.

Verification: `TestAggregatePatterns_EmptyFile` (existing test) already covers this path. No additional test needed, but the seam must not panic on empty input.

### EC-A2: Single-Line JSONL (1 Event)

Input: JSONL file with exactly one event line.

Behavior: `AggregatePatterns` returns `map[string]*Pattern` with 1 entry (Count=1). Stage-2 not invoked (`Stage2Enabled=false`). No audit log.

Verification: Existing `TestAggregatePatterns_SingleEntry` or equivalent covers this. If not present, T-A4 need not add it — confirm existing coverage handles it.

### EC-A3: Map Iteration Order Non-Determinism

Risk: `json.MarshalIndent` on `map[string]*Pattern` in golden fixture generator could produce non-deterministic key order.

Mitigation: Golden fixture generator MUST sort map keys before marshalling. Verified by criterion 4 (`-count=5` all pass).

### EC-A4: Test-Isolated Audit Log Path

T-A4 MUST pass the audit log path as `filepath.Join(t.TempDir(), ".moai/harness/cluster-merges.jsonl")` when constructing the expected-absent path. The test must assert `os.IsNotExist` on THIS path, NOT on the project-root `.moai/harness/cluster-merges.jsonl`. If Wave A's seam hardwires `""` as `auditLogPath`, the test can assert the project-root path is absent — both approaches are acceptable, but the project-root assertion is riskier if any pre-existing file exists there. Use `t.TempDir()`-based isolation.

### EC-A5: `ClassifierConfig` Zero-Value Safety (Evaluator-Added)

Risk: If `ClassifierConfig` fields are added in Wave C with non-zero defaults, Wave A's hardwired `ClassifierConfig{}` could silently change behavior.

Mitigation: Wave A's contract requires `Stage2Enabled bool` only. No other fields. Wave C extends the struct — implementer MUST ensure new fields default to safe values (false/zero for booleans/ints) so that Wave A's `ClassifierConfig{}` literal remains a valid "Stage2 fully disabled" configuration.

### EC-A6: `clusterSingletons` Stub Error Path (Evaluator-Added)

Risk: T-A4 exercises only the `Stage2Enabled=false` branch. The error-return path in the seam (`if err != nil { return nil, ... }`) has 0% coverage in Wave A.

Mitigation: Wave A error-path coverage is explicitly deferred to Wave C tests. The contract acknowledges this as intentional (stub returns nil error always). This is exempt from the 85% threshold per § Coverage Clarification below.

---

## Hard Thresholds

### Coverage

- **Overall `internal/harness/` package coverage: ≥ 85%** (matches `test_coverage_target: 85` in quality.yaml)
- **Coverage clarification**: `clusterSingletons` stub body and the `if cfg.Stage2Enabled == true` branch in `AggregatePatterns` are NOT callable in Wave A (Stage2 hardwired off). These lines are exempt from Wave A's 85% target. Evaluator verifies package-level coverage, not line-level stub coverage.
- Wave C tests will bring these lines to ≥ 85% coverage.

### LSP Quality Gates (from quality.yaml lsp_quality_gates.run)

- LSP errors: 0 (max_errors: 0)
- LSP type errors: 0 (max_type_errors: 0)
- LSP lint errors: 0 (max_lint_errors: 0)
- LSP regression: not allowed (allow_regression: false)

### LOC Budget

- Wave A LOC: ≤ 200 LOC excluding fixture files (testdata/)
- Strategy estimate: ~100 LOC (learner.go ~20 + classifier_cluster.go ~20 + learner_test.go ~30 + minor overhead)
- Exceeding 200 LOC (non-fixture) triggers Simplicity Review before Phase 2.8a proceeds

### File Scope (HARD — zero exceptions)

Wave A MUST NOT modify any file outside the 5 listed in § Scope Confirmation. Specifically PROHIBITED:

| File / Pattern | Reason |
|----------------|--------|
| `internal/harness/observer.go` | REQ-HRN-CLS-011: hot path preserved |
| `internal/harness/types.go` | REQ-HRN-CLS-006/007/008: FROZEN structs |
| `internal/harness/frozen_guard.go` | REQ-HRN-CLS-009: FROZEN zone guard |
| `internal/harness/safety/` (all files) | REQ-HRN-CLS-008: 5-Layer Safety untouched |
| `internal/cli/hook.go` | REQ-HRN-CLS-011: hot path preserved |
| `.moai/config/sections/harness.yaml` | Wave D scope |
| `internal/config/loader.go` | Wave D scope |
| Any `.claude/` path | FROZEN |

---

## Security Must-Pass

All 3 security criteria are must-pass (blocking). Failure = Overall FAIL regardless of other scores.

1. **PromptContent-free classifier**: `grep -c 'PromptContent' internal/harness/classifier_cluster.go 2>/dev/null || echo 0` returns `0`. Criterion 7.
2. **No new AskUserQuestion call sites**: `grep -rn 'AskUserQuestion' internal/harness/classifier_cluster.go 2>/dev/null` returns empty. Subagent prohibition (agent-common-protocol).
3. **No new network calls or external dependencies**: `classifier_cluster.go` imports only stdlib packages (`fmt`, etc.). No `net/`, `net/http`, third-party modules. Verified via `go list -deps ./internal/harness/...` — no new external imports vs pre-Wave-A baseline.

---

## Performance Targets

- Wave A introduces no performance regressions in the observer hot path (`observer.go` unmodified).
- The seam in `AggregatePatterns` adds O(1) work when `Stage2Enabled=false` (single boolean check). No impact on aggregation throughput.
- AC-HRN-CLS-005 (p99 ≤ 25ms) is Wave D scope — informational only, NOT Wave A blocking.

---

## MX Tag Obligation

Wave A REFACTOR phase MUST add:

```go
// @MX:NOTE: [AUTO] Stage-2 seam — REQ-HRN-CLS-001 backward compat gate.
// @MX:SPEC: REQ-HRN-CLS-001, REQ-HRN-CLS-004
// cfg.Stage2Enabled is false by default; no clustering until Wave C + D config wiring.
```

Placed immediately before the `if cfg.Stage2Enabled {` seam in `AggregatePatterns`. Language: Korean is permitted per language.yaml `code_comments: ko` but English is also acceptable for this annotation.

---

## Contract Acceptance

**manager-develop SHALL**:
- Implement Wave A within the 5-file scope, TDD ordering, and stub signature specified above
- Pass all 12 Done Criteria before marking Wave A complete
- Verify all 6 Edge Cases are handled
- Report PR diff summary to user for checkpoint before proceeding to Wave B

**evaluator-active SHALL**:
- Evaluate Phase 2.8a ONLY against this contract's Done Criteria, Edge Cases, Hard Thresholds, and Security Must-Pass
- Score using 4-dimension model: Functionality (40%), Security (25%), Craft (20%), Consistency (15%)
- Must-pass dimensions: Functionality + Security (per HRN-003 firewall)
- AC-HRN-CLS-001 is a must-pass sub-criterion under Functionality
- Disputes resolved by reference to spec-compact.md + strategy-notes.md

**Neither party SHALL**:
- Interpret ambiguous situations in their own favor without first checking spec-compact.md
- Exceed Wave A file scope unilaterally
- Change the `AggregatePatterns` public signature

---

End of Sprint Contract — Wave A. Owned by evaluator-active. Binding on Wave A execution.
