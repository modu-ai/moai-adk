# SPEC-TEST-001: Comprehensive CLI Testing System

> **SPEC ID**: SPEC-TEST-001
> **Created**: 2026-01-13
> **Status**: Planned
> **Priority**: High
> **Lifecycle Level**: spec-anchored

---

## TAG BLOCK

\`\`\`
@SPEC SPEC-TEST-001
@TYPE test-coverage
@DOMAIN cli-testing
@RELATED SPEC-COVERAGE-001, SPEC-COVERAGE-002, SPEC-COVERAGE-003
@IMPLEMENTATION tests/
\`\`\`

---

## Overview

### Purpose

Develop comprehensive test coverage for the MoAI-ADK CLI toolkit, ensuring all 27 commands across 7 command groups are thoroughly tested with normal operations, edge cases, error conditions, and integration scenarios.

### Scope

- Core Commands: `moai init`, `moai update`, `moai doctor`
- Status Commands: `moai status`
- Switch Commands: `moai switch`
- Rank Commands: `moai rank`
- Language Commands: `moai language`
- Worktree Commands: `moai worktree create|list|remove|prune|switch`
- Options and Flags: All command-line options and parameter combinations
- Temporary Directory Testing: Installation and operations in /tmp environments
- Integration Workflows: End-to-end command chain testing

### Exclusions

- Web UI testing (covered by SPEC-WEB-001, SPEC-WEB-002, SPEC-WEB-003)
- Agent-specific behavior testing (covered by other SPECs)
- Performance benchmarking (separate SPEC required)
- Security penetration testing (separate SPEC required)

---

## Environment

### Development Environment

- **Python Version**: 3.11, 3.12, 3.13, 3.14
- **Operating Systems**: macOS, Linux, Windows (WSL2)
- **Package Manager**: uv (primary), pip (fallback)
- **Test Framework**: pytest 8.4.2+
- **Coverage Tool**: pytest-cov 7.0.0+

### Testing Environment Variables

- `TEMP_DIR`: Temporary directory path for isolation tests
- `MOAI_TEST_HOME`: Mock home directory for config testing
- `MOAI_SKIP_HOOKS`: Disable pre-commit hooks during testing
- `PYTEST_XDIST_WORKER_COUNT`: Parallel test worker count

### Dependencies

- pytest >= 8.4.2
- pytest-cov >= 7.0.0
- pytest-xdist >= 3.8.0
- pytest-asyncio >= 1.2.0
- pytest-mock >= 3.15.1
- click-tester >= 1.0.0 (for CLI testing)

---

## Assumptions

### Technical Assumptions

- **A1**: Click framework provides built-in testing utilities for CLI command isolation
- **A2**: pytest fixtures can create temporary directories and clean up automatically
- **A3**: Git operations can be mocked using GitPython's test utilities
- **A4**: File system operations in /tmp are safe for test isolation

**Confidence Level**: High

**Evidence Basis**: Click and pytest documentation confirms testing utilities

**Risk if Wrong**: Test isolation failures causing intermittent test failures

**Validation Method**: Verify fixture cleanup and isolation in first 5 test files

### Business Assumptions

- **B1**: 85% code coverage is achievable within current sprint capacity
- **B2**: Existing 898 passing tests provide foundation for expansion
- **B3**: Coverage reporting showing 0% is configuration issue, not missing tests

**Confidence Level**: Medium

**Evidence Basis**: User reports 60+ test files exist with passing tests

**Risk if Wrong**: Underestimation of effort required to achieve coverage targets

**Validation Method**: Run coverage report and analyze gaps before implementing new tests

### Team Assumptions

- **T1**: Team has pytest experience for test creation
- **T2**: Team understands CLI testing patterns for Click applications

**Confidence Level**: High

**Evidence Basis**: Existing test suite demonstrates pytest proficiency

**Risk if Wrong**: Slower test implementation velocity

**Validation Method**: Code review first 10 new tests for pattern compliance

---

## Requirements (EARS Format)

### R1: Ubiquitous Requirements - Test Infrastructure

The system **shall always** maintain test isolation between test cases to prevent state leakage.

The system **shall always** clean up temporary directories and files after test execution.

The system **shall always** use pytest fixtures for common test setup and teardown.

The system **shall always** execute tests in random order to detect hidden dependencies.

### R2: Event-Driven Requirements - Command Execution

**WHEN** user executes `moai init` command **THEN** the system **shall** verify template creation in target directory.

**WHEN** user executes `moai update` command **THEN** the system **shall** verify template synchronization from source.

**WHEN** user executes `moai doctor` command **THEN** the system **shall** verify all diagnostic checks execute and report status.

**WHEN** user executes `moai status` command **THEN** the system **shall** display current project configuration and version.

**WHEN** user executes `moai switch <agent>` command **THEN** the system **shall** verify active agent changes to specified agent.

**WHEN** user executes `moai rank` command **THEN** the system **shall** display ranked list of available agents or skills.

**WHEN** user executes `moai language` command **THEN** the system **shall** display or modify current language settings.

**WHEN** user executes `moai worktree create <branch>` command **THEN** the system **shall** create git worktree with specified branch.

**WHEN** user executes `moai worktree list` command **THEN** the system **shall** display all existing worktrees.

**WHEN** user executes `moai worktree remove <worktree>` command **THEN** the system **shall** remove specified worktree directory.

**WHEN** user executes `moai worktree prune` command **THEN** the system **shall** remove all worktrees with deleted branches.

**WHEN** user executes `moai worktree switch <worktree>` command **THEN** the system **shall** change directory to specified worktree.

### R3: State-Driven Requirements - Conditional Behavior

**IF** target directory exists **THEN** `moai init` **shall** prompt for confirmation before overwriting.

**IF** configuration file is corrupted **THEN** `moai doctor` **shall** detect and report corruption.

**IF** network is unavailable **THEN** `moai update` **shall** fail gracefully with error message.

**IF** git repository is not initialized **THEN** worktree commands **shall** fail with appropriate error message.

**IF** temporary directory has insufficient permissions **THEN** tests **shall** skip with informative message.

### R4: Unwanted Requirements - Prohibited Behavior

The system **shall not** modify user's actual home directory during test execution.

The system **shall not** leave test artifacts in /tmp after test suite completion.

The system **shall not** execute commands that require interactive input during automated testing.

The system **shall not** depend on test execution order for test validity.

### R5: Optional Requirements - Enhancement Features

**WHERE POSSIBLE**, the system **shall** provide parallel test execution support using pytest-xdist.

**WHERE POSSIBLE**, the system **shall** provide test coverage reporting for each command group separately.

**WHERE POSSIBLE**, the system **shall** provide performance baseline metrics for command execution time.

---

## Specifications

### S1: Test Structure Organization

Test files **shall** mirror CLI command structure:

```
tests/
├── cli/
│   ├── commands/
│   │   ├── test_init.py           # moai init, moai update
│   │   ├── test_doctor.py         # moai doctor
│   │   ├── test_status.py         # moai status
│   │   ├── test_switch.py         # moai switch
│   │   ├── test_rank.py           # moai rank
│   │   ├── test_language.py       # moai language
│   │   ├── test_worktree.py       # moai worktree subcommands
│   │   └── test_options.py        # global options and flags
│   ├── integration/
│   │   ├── test_init_workflow.py  # end-to-end init workflows
│   │   ├── test_update_workflow.py # end-to-end update workflows
│   │   └── test_worktree_workflow.py # worktree operation chains
│   └── helpers/
│       ├── fixtures.py            # shared pytest fixtures
│       ├── assertions.py          # custom assertion helpers
│       └── mocks.py               # mock objects and utilities
```

### S2: Test Coverage Requirements

Each command group **shall** achieve minimum 85% line coverage:

- Init/Update Commands: Test template copy, config merge, file permissions
- Doctor Command: Test all diagnostic checks, error handling
- Status Command: Test display formats, config reading
- Switch Command: Test agent switching, config updates
- Rank Command: Test ranking algorithms, output formats
- Language Command: Test language switching, config persistence
- Worktree Commands: Test git operations, directory management
- Options: Test all flags, parameter validation

### S3: Temporary Directory Testing

Tests **shall** verify correct behavior in /tmp directory environments:

- **T1**: Test `moai init` in /tmp/test-moai-project-<uuid>
- **T2**: Test `moai update` in temporary project location
- **T3**: Test configuration file creation with non-standard HOME
- **T4**: Test worktree operations in temporary git repositories
- **T5**: Verify cleanup of all temporary files after test completion

### S4: Worktree Edge Cases

Worktree tests **shall** cover edge cases:

- **W1**: Create worktree with non-existent branch (shall fail)
- **W2**: Create worktree with existing directory (shall fail or prompt)
- **W3**: Remove worktree with uncommitted changes (shall warn)
- **W4**: Prune worktrees when no branches deleted (shall report no action)
- **W5**: Switch to worktree that doesn't exist (shall fail)
- **W6**: Create worktree with same branch multiple times (shall fail)

### S5: Error Condition Testing

Tests **shall** verify error handling for:

- **E1**: Invalid command-line arguments
- **E2**: Missing required parameters
- **E3**: Invalid configuration file format
- **E4**: File system permission errors
- **E5**: Git repository corruption
- **E6**: Network errors during update
- **E7**: Insufficient disk space
- **E8**: Concurrent command execution conflicts

### S6: Integration Test Scenarios

Tests **shall** verify command chain workflows:

- **I1**: init → doctor → status (new project verification)
- **I2**: init → update → status (update verification)
- **I3**: worktree create → switch → worktree remove (worktree lifecycle)
- **I4**: language switch → status (language persistence)
- **I5**: init → worktree create → update (parallel development setup)

### S7: Options and Flags Testing

Tests **shall** verify all command options:

- **O1**: --help flag for all commands
- **O2**: --version flag for main CLI
- **O3**: --verbose flag for detailed output
- **O4**: --quiet flag for minimal output
- **O5**: --dry-run flag for preview operations
- **O6**: --force flag for bypassing confirmations
- **O7**: --config flag for custom config path
- **O8**: Command-specific options for each command

---

## Traceability

### Requirements to Test Mapping

| Requirement | Test File | Test Function |
|-------------|-----------|---------------|
| R2 (init) | test_init.py | test_init_creates_templates |
| R2 (update) | test_init.py | test_update_syncs_templates |
| R2 (doctor) | test_doctor.py | test_doctor_runs_all_checks |
| R2 (status) | test_status.py | test_status_shows_config |
| R2 (switch) | test_switch.py | test_switch_changes_agent |
| R2 (rank) | test_rank.py | test_rank_lists_agents |
| R2 (language) | test_language.py | test_language_changes_locale |
| R2 (worktree) | test_worktree.py | test_worktree_operations |
| S3 (/tmp) | test_init.py | test_init_in_temp_directory |
| S4 (edge cases) | test_worktree.py | test_worktree_edge_cases |
| S5 (errors) | test_options.py | test_error_handling |
| S6 (integration) | test_init_workflow.py | test_end_to_end_init |

### Acceptance Criteria Linking

See acceptance.md for detailed Given-When-Then scenarios linked to each requirement.

---

## Success Criteria

### Coverage Metrics

- [ ] Overall test coverage: >= 85%
- [ ] CLI command coverage: 100% (all 27 commands tested)
- [ ] Error path coverage: >= 80%
- [ ] Edge case coverage: >= 75%

### Quality Metrics

- [ ] All tests pass consistently (no flaky tests)
- [ ] Test execution time: < 5 minutes for full suite
- [ ] Zero test isolation failures
- [ ] All temporary files cleaned up

### Functional Metrics

- [ ] All 27 CLI commands have test coverage
- [ ] All command options have test coverage
- [ ] All error conditions have test coverage
- [ ] All integration workflows have test coverage

---

## Dependencies

### Internal Dependencies

- **D1**: MoAI-ADK CLI implementation (src/moai_adk/cli/)
- **D2**: Configuration system (src/moai_adk/project/)
- **D3**: Git utilities (src/moai_adk/foundation/git/)
- **D4**: Template system (src/moai_adk/templates/)

### External Dependencies

- **D5**: pytest >= 8.4.2
- **D6**: pytest-cov >= 7.0.0
- **D7**: pytest-xdist >= 3.8.0

---

## Risks and Mitigation

### Risk R1: Coverage Reporting Configuration Issues

**Risk**: Current 0% coverage report indicates configuration problems

**Probability**: High

**Impact**: Cannot accurately measure coverage progress

**Mitigation**:
- Investigate coverage configuration (pytest.ini, pyproject.toml)
- Verify .coveragerc settings
- Test with simple coverage case first

### Risk R2: Test Isolation Failures

**Risk**: Tests sharing state causing intermittent failures

**Probability**: Medium

**Impact**: Unreliable test suite, blocked CI/CD

**Mitigation**:
- Use pytest fixtures with proper scope
- Implement random test execution order
- Add state cleanup verification

### Risk R3: Temporary Directory Cleanup

**Risk**: Test artifacts left in /tmp consuming disk space

**Probability**: Medium

**Impact**: Disk space exhaustion, CI/CD failures

**Mitigation**:
- Use tmp_path fixture with automatic cleanup
- Add cleanup verification tests
- Implement post-test cleanup hooks

---

## Notes

### Current State Assessment

From Phase 0 exploration:

- 27 CLI commands identified across 7 command groups
- 60+ existing test files
- 898 passing tests
- Coverage reporting shows 0% (configuration issue suspected)
- Missing test areas: statusline, integration workflows

### Implementation Strategy

1. **Phase 1**: Fix coverage reporting configuration
2. **Phase 2**: Implement missing command tests (init, update, doctor, status, switch, rank, language, worktree)
3. **Phase 3**: Add temporary directory tests
4. **Phase 4**: Implement worktree edge case tests
5. **Phase 5**: Add integration test scenarios
6. **Phase 6**: Achieve 85% coverage target

### Testing Best Practices

- Use pytest fixtures for common setup
- Mock external dependencies (Git, network)
- Test both success and failure paths
- Verify side effects (file creation, config changes)
- Use descriptive test names following Given-When-Then pattern
- Add docstrings explaining complex test scenarios

---

**END OF SPEC-TEST-001**
