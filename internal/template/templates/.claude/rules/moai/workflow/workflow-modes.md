# Workflow Modes

Development methodology reference for MoAI-ADK SPEC workflow.

For phase overview, token strategy, and transitions, see @spec-workflow.md

## Methodology Selection

The Run Phase adapts its workflow based on `quality.development_mode` in `.moai/config/sections/quality.yaml`:

| Mode | Workflow Cycle | Best For | Agent Strategy |
|------|---------------|----------|----------------|
| DDD | ANALYZE-PRESERVE-IMPROVE | Existing projects, < 10% coverage | Characterization tests first |
| TDD | RED-GREEN-REFACTOR | New projects, 50%+ coverage | Tests before implementation |
| Hybrid | Mixed per change type | Partial coverage (10-49%) | New code: TDD, Legacy: DDD |

## DDD Mode (default)

Development methodology: Domain-Driven Development (ANALYZE-PRESERVE-IMPROVE)

**ANALYZE**: Understand existing behavior and code structure
- Read existing code and identify dependencies
- Map domain boundaries and interaction patterns
- Identify side effects and implicit contracts

**PRESERVE**: Create characterization tests for existing behavior
- Write characterization tests capturing current behavior
- Create behavior snapshots for regression detection
- Verify test coverage of critical paths

**IMPROVE**: Implement changes with behavior preservation
- Make small, incremental changes
- Run characterization tests after each change
- Refactor with test validation

Success Criteria:
- All SPEC requirements implemented
- Characterization tests passing
- Behavior snapshots stable (no regression)
- 85%+ code coverage achieved
- TRUST 5 quality gates passed

## TDD Mode

Development methodology: Test-Driven Development (RED-GREEN-REFACTOR)

**RED**: Write a failing test
- Write a test that describes the desired behavior
- Verify the test fails (confirms it tests something new)
- One test at a time, focused and specific

**GREEN**: Write minimal code to pass
- Write the simplest implementation that makes the test pass
- No premature optimization or abstraction
- Focus on correctness, not elegance

**REFACTOR**: Improve code quality
- Clean up implementation while keeping tests green
- Extract patterns, remove duplication
- Apply SOLID principles where appropriate

Success Criteria:
- All SPEC requirements implemented
- All tests passing (RED-GREEN-REFACTOR complete)
- Minimum coverage per commit: 80% (configurable)
- No test written after implementation code
- TRUST 5 quality gates passed

## Hybrid Mode

Development methodology: Hybrid (TDD for new + DDD for legacy)

**For NEW code** (new files, new functions):
- Apply TDD workflow (RED-GREEN-REFACTOR)
- Strict test-first requirement
- Coverage target: 85% for new code

**For EXISTING code** (modifications, refactoring):
- Apply DDD workflow (ANALYZE-PRESERVE-IMPROVE)
- Characterization tests before changes
- Coverage target: 85% for modified code

**Classification Logic**:
- New files - TDD rules
- Modified existing files - DDD rules
- New functions in existing files - TDD rules for those functions
- Deleted code - Verify characterization tests still pass

Success Criteria:
- All SPEC requirements implemented
- New code has TDD-level coverage (85%+)
- Modified code has characterization tests
- Overall coverage improvement trend
- TRUST 5 quality gates passed

## Team Mode Methodology

When --team flag is used, the methodology applies at the teammate level:

| Methodology | Team Behavior |
|-------------|---------------|
| DDD | Each teammate applies ANALYZE-PRESERVE-IMPROVE within their file ownership scope |
| TDD | Each teammate applies RED-GREEN-REFACTOR within their module scope |
| Hybrid | team-tester uses TDD for new test files; team-backend-dev and team-frontend-dev use DDD for existing code modifications |

Team-specific rules:
- Methodology is shared across all teammates via the SPEC document
- team-quality agent validates methodology compliance after all implementation completes
- File ownership prevents cross-teammate conflicts during parallel development
- team-tester exclusively owns test files regardless of methodology

## Methodology Selection Guide

### Auto-Detection (via /moai project or /moai init)

The system automatically recommends a methodology based on project analysis:

| Project State | Test Coverage | Recommendation | Rationale |
|--------------|---------------|----------------|-----------|
| Greenfield (new) | N/A | Hybrid | Clean slate, TDD for features + DDD structure |
| Brownfield | >= 50% | TDD | Sufficient test base for test-first development |
| Brownfield | 10-49% | Hybrid | Partial tests, expand with DDD then TDD for new |
| Brownfield | < 10% | DDD | No tests, gradual characterization test creation |

### Manual Override

Users can override the auto-detected methodology:
- During init: Use `moai init --mode <ddd|tdd|hybrid>` flag (default: hybrid)
- After project setup: Re-run `/moai project` to auto-detect based on codebase analysis
- Manual edit: Edit `quality.development_mode` in `.moai/config/sections/quality.yaml`
- Per session: Set `MOAI_DEVELOPMENT_MODE` environment variable

### Methodology Comparison

| Aspect | DDD | TDD | Hybrid |
|--------|-----|-----|--------|
| Test timing | After analysis (PRESERVE) | Before code (RED) | Mixed |
| Coverage approach | Gradual improvement | Strict per-commit | Unified 85% target |
| Best for | Legacy refactoring only | Isolated modules (rare) | All development work |
| Risk level | Low (preserves behavior) | Medium (requires discipline) | Medium |
| Coverage exemptions | Allowed | Not allowed | Allowed for legacy only |
| Run Phase cycle | ANALYZE-PRESERVE-IMPROVE | RED-GREEN-REFACTOR | Both (per change type) |
