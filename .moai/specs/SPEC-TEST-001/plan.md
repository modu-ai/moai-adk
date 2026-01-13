# Implementation Plan: SPEC-TEST-001

> **SPEC ID**: SPEC-TEST-001
> **Document Version**: 1.0
> **Last Updated**: 2026-01-13

---

## Overview

This plan outlines the implementation strategy for comprehensive CLI testing coverage for MoAI-ADK, targeting 85% code coverage across all 27 CLI commands with proper isolation, error handling, and integration testing.

---

## Milestones

### Primary Goal (High Priority)

**Achieve 85% test coverage for all CLI commands**

- Implement tests for all 27 CLI commands
- Fix coverage reporting configuration
- Ensure test isolation and cleanup
- Verify all command options and flags

### Secondary Goal (Medium Priority)

**Comprehensive edge case and error testing**

- Temporary directory testing
- Worktree edge cases
- Error condition handling
- Integration workflow testing

### Final Goal (Low Priority)

**Performance and optimization**

- Parallel test execution support
- Test execution time optimization
- Coverage reporting improvements
- Test documentation and examples

---

## Technical Approach

### Phase 1: Coverage Configuration Fix

**Objective**: Resolve coverage reporting showing 0% despite 898 passing tests

**Tasks**:

1. Investigate pytest coverage configuration
   - Check pytest.ini coverage settings
   - Verify .coveragerc configuration
   - Test with simple coverage example
   - Validate source path mapping

2. Fix configuration issues
   - Update pyproject.toml coverage settings
   - Configure source discovery paths
   - Set proper coverage parameters
   - Test coverage report generation

3. Verify baseline coverage
   - Run coverage on existing tests
   - Generate baseline report
   - Identify actual current coverage
   - Document configuration fixes

**Success Criteria**:
- Coverage report generates without errors
- Baseline coverage accurately reflects existing tests
- Configuration documented for future reference

**Dependencies**: None (can start immediately)

---

### Phase 2: Core Command Testing

**Objective**: Implement tests for core CLI commands (init, update, doctor, status)

**Tasks**:

1. Create test file structure
   - `tests/cli/commands/test_init.py`
   - `tests/cli/commands/test_update.py`
   - `tests/cli/commands/test_doctor.py`
   - `tests/cli/commands/test_status.py`

2. Implement init command tests
   - Test successful initialization
   - Test template creation
   - Test configuration file generation
   - Test directory creation with permissions
   - Test overwrite confirmation prompt
   - Test error handling for invalid paths

3. Implement update command tests
   - Test template synchronization
   - Test config merge behavior
   - Test backup creation before update
   - Test network error handling
   - Test version comparison logic
   - Test dry-run flag

4. Implement doctor command tests
   - Test all diagnostic checks
   - Test check result reporting
   - Test fix suggestion generation
   - Test error handling for corrupted configs
   - Test verbose vs quiet output

5. Implement status command tests
   - Test project status display
   - Test version information display
   - Test configuration display
   - Test template status display
   - Test output formatting

**Success Criteria**:
- All core commands have test coverage
- Normal and error paths tested
- Coverage increases by measurable amount

**Dependencies**: Phase 1 complete

---

### Phase 3: Interactive Command Testing

**Objective**: Implement tests for interactive CLI commands (switch, rank, language)

**Tasks**:

1. Create test files
   - `tests/cli/commands/test_switch.py`
   - `tests/cli/commands/test_rank.py`
   - `tests/cli/commands/test_language.py`

2. Implement switch command tests
   - Test agent switching
   - Test config update persistence
   - Test invalid agent handling
   - Test current agent display
   - Test auto-suggestion behavior

3. Implement rank command tests
   - Test ranking algorithm
   - Test output formatting (table, list, json)
   - Test filtering options
   - Test sorting behavior
   - Test search functionality

4. Implement language command tests
   - Test language switching
   - Test config persistence
   - Test invalid language handling
   - Test current language display
   - Test language list display

**Success Criteria**:
- All interactive commands have tests
- Mock user input properly isolated
- Config changes verified

**Dependencies**: Phase 2 complete

---

### Phase 4: Worktree Command Testing

**Objective**: Implement comprehensive tests for worktree commands

**Tasks**:

1. Create test file
   - `tests/cli/commands/test_worktree.py`

2. Implement worktree create tests
   - Test successful worktree creation
   - Test branch specification
   - Test directory creation
   - Test duplicate worktree handling
   - Test non-existent branch handling

3. Implement worktree list tests
   - Test worktree listing
   - Test output formatting
   - Test empty worktree list

4. Implement worktree remove tests
   - Test worktree removal
   - Test uncommitted changes warning
   - Test non-existent worktree handling
   - Test directory cleanup

5. Implement worktree prune tests
   - Test pruning deleted branches
   - Test no action when no branches deleted
   - Test confirmation prompt

6. Implement worktree switch tests
   - Test directory switching
   - Test non-existent worktree handling
   - Test shell integration

**Success Criteria**:
- All worktree subcommands tested
- Edge cases covered
- Git operations properly mocked

**Dependencies**: Phase 3 complete

---

### Phase 5: Options and Flags Testing

**Objective**: Test all command-line options and flags

**Tasks**:

1. Create test file
   - `tests/cli/commands/test_options.py`

2. Test global options
   - `--help` flag for all commands
   - `--version` flag
   - `--verbose` flag
   - `--quiet` flag
   - `--config` flag
   - `--dry-run` flag
   - `--force` flag

3. Test command-specific options
   - Init command options
   - Update command options
   - Doctor command options
   - Worktree command options

4. Test parameter validation
   - Required parameters
   - Optional parameters
   - Parameter type validation
   - Parameter range validation

**Success Criteria**:
- All options have test coverage
- Validation logic tested
- Error messages verified

**Dependencies**: Phase 4 complete

---

### Phase 6: Temporary Directory Testing

**Objective**: Test CLI operations in temporary directory environments

**Tasks**:

1. Create test file
   - `tests/cli/test_temp_directory.py`

2. Implement /tmp directory tests
   - Test `moai init` in /tmp/test-project-<uuid>
   - Test `moai update` in temporary location
   - Test config creation with non-standard HOME
   - Test worktree operations in temporary repos
   - Verify cleanup after test completion

3. Implement permission tests
   - Test read-only directory handling
   - Test no-permission directory handling
   - Test cross-device link handling

**Success Criteria**:
- All operations tested in /tmp
- Cleanup verified
- Permission errors handled

**Dependencies**: Phase 5 complete

---

### Phase 7: Integration Testing

**Objective**: Test end-to-end command workflows

**Tasks**:

1. Create test files
   - `tests/cli/integration/test_init_workflow.py`
   - `tests/cli/integration/test_update_workflow.py`
   - `tests/cli/integration/test_worktree_workflow.py`

2. Implement init workflow tests
   - Test init → doctor → status chain
   - Test init with custom options
   - Test init failure recovery

3. Implement update workflow tests
   - Test init → update → status chain
   - Test update with backup
   - Test update rollback

4. Implement worktree workflow tests
   - Test worktree create → switch → remove
   - Test multiple worktrees management
   - Test worktree branch tracking

**Success Criteria**:
- All integration workflows tested
- Command chains verified
- State persistence confirmed

**Dependencies**: Phase 6 complete

---

## Architecture Design

### Test File Organization

```
tests/
├── cli/
│   ├── commands/          # Individual command tests
│   │   ├── test_init.py
│   │   ├── test_update.py
│   │   ├── test_doctor.py
│   │   ├── test_status.py
│   │   ├── test_switch.py
│   │   ├── test_rank.py
│   │   ├── test_language.py
│   │   ├── test_worktree.py
│   │   └── test_options.py
│   ├── integration/        # End-to-end workflow tests
│   │   ├── test_init_workflow.py
│   │   ├── test_update_workflow.py
│   │   └── test_worktree_workflow.py
│   └── helpers/           # Shared test utilities
│       ├── fixtures.py    # Pytest fixtures
│       ├── assertions.py  # Custom assertions
│       └── mocks.py       # Mock objects
└── conftest.py            # Root configuration
```

### Fixture Design

Key fixtures to implement:

- `temp_dir`: Temporary directory with auto-cleanup
- `mock_home`: Mock home directory for config isolation
- `mock_git_repo`: Mock git repository for worktree tests
- `cli_runner`: Click CLI test runner
- `sample_config`: Sample configuration for testing
- `mock_templates`: Mock template directory

### Testing Patterns

1. **Command Testing Pattern**:
   - Arrange: Setup test environment, fixtures, mocks
   - Act: Execute CLI command via CliRunner
   - Assert: Verify exit code, output, side effects

2. **Integration Testing Pattern**:
   - Setup: Create isolated test environment
   - Execute: Run command chain in sequence
   - Verify: Check final state and intermediate states

3. **Mock Strategy**:
   - Mock Git operations using GitPython test utilities
   - Mock network calls for update commands
   - Mock file system for permission tests
   - Mock user input for interactive commands

---

## Risks and Response Plans

### Risk R1: Coverage Configuration Complexity

**Risk**: Coverage configuration may require significant debugging

**Probability**: High

**Impact**: Delays Phase 2 start

**Response Plan**:
- Allocate dedicated time for Phase 1 (2-3 days)
- Create minimal test case for validation
- Consult pytest-cov documentation
- Consider coverage.py alternative if needed

### Risk R2: Test Isolation Challenges

**Risk**: Tests may share state causing flaky failures

**Probability**: Medium

**Impact**: Unreliable test suite

**Response Plan**:
- Use pytest fixtures with function scope by default
- Implement random test execution order detection
- Add state cleanup verification tests
- Use separate temp directories for each test

### Risk R3: Mock Complexity for Git Operations

**Risk**: Git operations may be difficult to mock properly

**Probability**: Medium

**Impact**: Worktree tests may be unreliable

**Response Plan**:
- Use GitPython's test utilities
- Create real temporary git repositories for tests
- Avoid mocking when feasible (use temp repos instead)
- Document git repository setup procedures

### Risk R4: 85% Coverage May Not Be Achievable

**Risk**: Some code paths may be difficult to test

**Probability**: Low

**Impact**: Coverage target not met

**Response Plan**:
- Identify difficult-to-test code paths early
- Refactor code for better testability if needed
- Document exceptions with justification
- Consider coverage exceptions if properly justified

---

## Quality Assurance

### Code Review Checklist

- [ ] Tests follow pytest best practices
- [ ] Test names clearly describe what is being tested
- [ ] Fixtures used appropriately
- [ ] Mocks properly isolated
- [ ] Assertions verify expected behavior
- [ ] Error cases tested alongside success cases
- [ ] Temporary files cleaned up
- [ ] Tests execute quickly (< 5 seconds per test)

### Coverage Validation

- [ ] Coverage report generates without errors
- [ ] Coverage >= 85% for each command group
- [ ] Coverage >= 80% for error paths
- [ ] Coverage >= 75% for edge cases
- [ ] No critical code paths untested

### Test Execution Validation

- [ ] All tests pass consistently
- [ ] No flaky tests (random failures)
- [ ] Test execution time acceptable
- [ ] Parallel execution works correctly
- [ ] Tests can run in isolation

---

## Definition of Done

A command is considered fully tested when:

- [ ] Normal operation path tested
- [ ] Error conditions tested
- [ ] Edge cases tested
- [ ] Options and flags tested
- [ ] Integration scenarios tested
- [ ] Coverage >= 85% for command module
- [ ] Tests pass consistently
- [ ] Test execution time < 5 seconds

The complete testing system is done when:

- [ ] All 27 CLI commands have tests
- [ ] Overall coverage >= 85%
- [ ] All integration workflows tested
- [ ] Zero flaky tests
- [ ] Full test suite executes in < 5 minutes
- [ ] Coverage report generates correctly
- [ ] Documentation updated

---

**END OF IMPLEMENTATION PLAN**
