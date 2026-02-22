---
paths: "**/.moai/specs/**,**/.moai/config/sections/quality.yaml"
---

# Workflow Modes

Development methodology reference for MoAI-ADK SPEC workflow.

For phase overview, token strategy, and transitions, see @spec-workflow.md

## Methodology Selection

The Run Phase adapts its workflow based on `quality.development_mode` in `.moai/config/sections/quality.yaml`:

| Mode | Workflow Cycle | Best For | Agent Strategy |
|------|---------------|----------|----------------|
| DDD | ANALYZE-PRESERVE-IMPROVE | Existing projects, < 10% coverage | Characterization tests first |
| TDD | RED-GREEN-REFACTOR | All development work, new projects, 10%+ coverage (default) | Tests before implementation |

## DDD Mode

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
- TRUST 5 gates passed (see @.claude/rules/moai/core/moai-constitution.md)

## TDD Mode (default)

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
- TRUST 5 gates passed (see @.claude/rules/moai/core/moai-constitution.md)

### Brownfield Enhancement (for existing codebases)

When TDD is selected for a project with existing code, the RED phase is enhanced:

1. (Pre-RED) Read existing code in the target area to understand current behavior
2. RED: Write a failing test informed by existing code understanding
3. GREEN: Write minimal code to pass
4. REFACTOR: Improve while keeping tests green

This ensures TDD on brownfield projects still respects existing behavior without requiring a separate methodology mode.

## Team Mode Methodology

When --team flag is used, the methodology applies at the teammate level:

| Methodology | Team Behavior |
|-------------|---------------|
| DDD | Each teammate applies ANALYZE-PRESERVE-IMPROVE within their file ownership scope |
| TDD | Each teammate applies RED-GREEN-REFACTOR within their module scope |

Team-specific rules:
- Methodology is shared across all teammates via the SPEC document
- team-quality agent validates methodology compliance after all implementation completes
- File ownership prevents cross-teammate conflicts during parallel development
- team-tester exclusively owns test files regardless of methodology

## MX Tag Integration

Both methodologies include @MX tag management:

### TDD Mode MX Tags

| Phase | MX Action |
|-------|-----------|
| RED | Add `@MX:TODO` for test requirements |
| GREEN | Remove `@MX:TODO` when test passes |
| REFACTOR | Add `@MX:NOTE` for refactored logic |

### DDD Mode MX Tags

| Phase | MX Action |
|-------|-----------|
| ANALYZE | Run 3-Pass scan, identify tag targets |
| PRESERVE | Validate existing tags, add `@MX:LEGACY` for legacy code |
| IMPROVE | Update tags, add `@MX:NOTE` for new logic |

### MX Tag Priority by Methodology

| Tag Type | TDD Trigger | DDD Trigger |
|----------|-------------|-------------|
| `@MX:TODO` | Missing test | SPEC not implemented |
| `@MX:NOTE` | Complex logic | Business rule discovered |
| `@MX:WARN` | Complexity >= 15 | Goroutine without context |
| `@MX:ANCHOR` | fan_in >= 3 | Public API boundary |

## Methodology Selection Guide

### Auto-Detection (via /moai project or /moai init)

The system automatically recommends a methodology based on project analysis:

| Project State | Test Coverage | Recommendation | Rationale |
|--------------|---------------|----------------|-----------|
| Greenfield (new) | N/A | TDD | Clean slate, test-first development |
| Brownfield | >= 50% | TDD | Strong test base for test-first development |
| Brownfield | 10-49% | TDD | Partial tests, expand with test-first development |
| Brownfield | < 10% | DDD | No tests, gradual characterization test creation |

### Manual Override

Users can override the auto-detected methodology:
- During init: Use `moai init --mode <ddd|tdd>` flag (default: tdd)
- After project setup: Re-run `/moai project` to auto-detect based on codebase analysis
- Manual edit: Edit `quality.development_mode` in `.moai/config/sections/quality.yaml`
- Per session: Set `MOAI_DEVELOPMENT_MODE` environment variable

### Methodology Comparison

| Aspect | DDD | TDD |
|--------|-----|-----|
| Test timing | After analysis (PRESERVE) | Before code (RED) |
| Coverage approach | Gradual improvement | Strict per-commit |
| Best for | Existing projects with < 10% coverage | All development work (default) |
| Risk level | Low (preserves behavior) | Medium (requires discipline) |
| Coverage exemptions | Allowed | Not allowed |
| Run Phase cycle | ANALYZE-PRESERVE-IMPROVE | RED-GREEN-REFACTOR |
