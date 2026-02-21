---
name: moai-workflow-coverage
description: >
  Analyze test coverage gaps and generate missing tests.
  Supports coverage target enforcement, file-specific analysis, and report-only mode.
  Uses language-specific coverage tools for accurate measurement.
  Use when improving test coverage, identifying gaps, or generating tests.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "coverage, testing, test-generation, gaps, quality"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4000

# MoAI Extension: Triggers
triggers:
  keywords: ["coverage", "test coverage", "coverage gap", "missing tests", "coverage target"]
  agents: ["expert-testing"]
  phases: ["coverage"]
---

# Workflow: Coverage - Test Coverage Analysis

Purpose: Analyze test coverage, identify gaps, and generate missing tests to meet coverage targets. Uses language-specific coverage tools for accurate measurement.

Flow: Measure Coverage -> Identify Gaps -> Generate Tests -> Verify -> Report

## Supported Flags

- --target N: Coverage target percentage (default: from quality.yaml test_coverage_target, typically 85)
- --file PATH: Analyze specific file or directory only
- --report: Generate coverage report only, do not generate tests
- --package PKG: Analyze specific package (Go) or module
- --uncovered: Show only uncovered lines/functions
- --critical: Focus on critical paths (high fan_in, public API)

## Phase 1: Coverage Measurement

[HARD] Delegate coverage measurement to the expert-testing subagent.

Language-specific coverage tools:

- Go: `go test -coverprofile=coverage.out -covermode=atomic ./...` then `go tool cover -func=coverage.out`
- Python: `pytest --cov --cov-report=json` or `coverage run -m pytest && coverage json`
- TypeScript/JavaScript: `vitest run --coverage` or `jest --coverage --json`
- Rust: `cargo llvm-cov --json`

If --file flag: Limit measurement to the specified file/directory.
If --package flag: Limit measurement to the specified package.

Expected Output:

- Overall coverage percentage
- Per-file coverage percentages
- Per-function coverage data (covered/uncovered lines)
- Branch coverage where available

## Phase 2: Gap Analysis

[HARD] Delegate gap analysis to the expert-testing subagent.

Analysis Tasks:

- Identify files below the coverage target
- List uncovered functions and methods
- Prioritize gaps by risk:
  - P1 (Critical): Public API functions, high fan_in (>=3), functions with @MX:ANCHOR
  - P2 (High): Business logic, error handling paths
  - P3 (Medium): Internal utilities, helper functions
  - P4 (Low): Generated code, configuration, trivial getters/setters

If --uncovered flag: Output only uncovered lines with file:line references.
If --critical flag: Focus analysis on P1 and P2 priority gaps only.

Gap Report Structure:

```markdown
## Coverage Gap Analysis

### Current Coverage: XX.X% (target: YY%)

### Critical Gaps (P1)
- file.go:FunctionName (0% covered, fan_in: 5, @MX:ANCHOR)

### High Priority Gaps (P2)
- file.go:BusinessLogic (30% covered, complex error handling)

### Medium Priority Gaps (P3)
- file.go:HelperFunc (0% covered, internal utility)

### Low Priority Gaps (P4)
- file_generated.go (excluded from target)
```

## Phase 3: Test Generation

If --report flag: Skip this phase. Display gap report and exit.

[HARD] Delegate test generation to the expert-testing subagent.

Test Generation Strategy (based on quality.yaml development_mode):

If TDD mode: Generate tests following RED-GREEN-REFACTOR pattern
- Write failing test first (RED)
- Verify test fails
- Note: Implementation already exists, so GREEN phase is verification

If DDD mode: Generate characterization tests
- Capture existing behavior as test assertions
- Create behavior snapshots for regression detection

Test Generation Order:
1. P1 critical gaps first (public API, high fan_in)
2. P2 high priority gaps (business logic, error handling)
3. P3 medium priority gaps (if target not yet met)
4. Skip P4 low priority gaps

For each gap:
- Generate table-driven tests (Go) or parameterized tests (Python/TS)
- Include edge cases and error scenarios
- Follow existing test patterns in the codebase
- Respect file naming conventions (*_test.go, *.test.ts, test_*.py)

## Phase 4: Verification

After test generation:
- Run the full test suite to ensure no regressions
- Re-measure coverage to confirm improvement
- Compare before/after coverage percentages
- Verify coverage target is met (or report remaining gap)

## Phase 5: Report

Display coverage report in user's conversation_language:

```markdown
## Coverage Report

### Before: XX.X% -> After: YY.Y%
### Target: ZZ% - ACHIEVED/REMAINING: N.N%

### Tests Generated: N
- file_test.go: TestFunctionA (covers P1 gap)
- file_test.go: TestFunctionB (covers P2 gap)

### Coverage by Package
| Package | Before | After | Target | Status |
|---------|--------|-------|--------|--------|
| pkg/api | 70% | 88% | 85% | PASS |
| pkg/core | 45% | 82% | 85% | FAIL |

### Remaining Gaps
- pkg/core: 3% remaining (2 functions uncovered)
```

Next Steps (AskUserQuestion):

- Fix remaining gaps (Recommended): Continue generating tests for uncovered areas until the target is met. MoAI will prioritize the highest-risk gaps.
- Run full test suite: Execute the complete test suite with race detection to verify all new tests pass reliably.
- Review generated tests: Open the generated test files for manual review and adjustment before committing.

## Task Tracking

[HARD] Task management tools mandatory:
- Each coverage gap tracked as a pending task via TaskCreate
- Before test generation: change to in_progress via TaskUpdate
- After test passes: change to completed via TaskUpdate

## Agent Chain Summary

- Phase 1-2: expert-testing subagent (measurement and analysis)
- Phase 3: expert-testing subagent (test generation)
- Phase 4: expert-testing subagent (verification)
- Phase 5: MoAI orchestrator (report and user interaction)

## Execution Summary

1. Parse arguments (extract flags: --target, --file, --report, --package, --uncovered, --critical)
2. Read coverage target from quality.yaml if --target not specified
3. Delegate coverage measurement to expert-testing subagent
4. Delegate gap analysis to expert-testing subagent
5. If --report: Display gap report and exit
6. Delegate test generation to expert-testing subagent
7. Verify tests pass and re-measure coverage
8. TaskCreate/TaskUpdate for all gaps and generated tests
9. Report results with next step options

---

Version: 1.0.0
