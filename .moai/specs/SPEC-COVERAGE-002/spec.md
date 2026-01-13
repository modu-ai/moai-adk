---
id: SPEC-COVERAGE-002
version: "1.0.1"
status: "complete-partial"
created: "2026-01-13"
updated: "2026-01-13"
author: "Alfred"
priority: "HIGH"
tags: [test-coverage, tdd, cli-commands, quality-gates, pytest]
completion_date: "2026-01-13"
achievement_level: "60%"
---

# SPEC-COVERAGE-002: Comprehensive TDD Test Coverage for CLI User Commands

## HISTORY

| Version | Date       | Changes                                      | Author |
| ------- | ---------- | -------------------------------------------- | ------ |
| 1.0.1   | 2026-01-13 | Implementation complete, partial achievement | Alfred |
| 1.0.0   | 2026-01-13 | Initial SPEC creation                        | Alfred |

## IMPLEMENTATION RESULTS

### Summary

- **Status**: PARTIAL SUCCESS (60% achievement)
- **Target**: 5 files to 85%+ coverage
- **Achieved**: 4 out of 5 files passed coverage target
- **Overall Coverage**: 71.4% -> 78.6% (Target: 85%, Gap: -6.4%)
- **Total Tests Created**: 205 tests
- **Test Files Created**: 6 comprehensive test files

### Coverage Results by Module

| Module      | Target | Original | Final | Delta  | Status  |
| ----------- | ------ | -------- | ----- | ------ | ------- |
| rank.py     | 85%+   | 63.11%   | 87.5% | +24.4% | PASS    |
| switch.py   | 95%+   | 95.00%   | 96.2% | +1.2%  | PASS    |
| status.py   | 97%+   | 97.67%   | 98.1% | +0.4%  | PASS    |
| language.py | 98%+   | 98.79%   | 99.0% | +0.2%  | PASS    |
| update.py   | 85%+   | 65.94%   | 72.8% | +6.9%  | FAIL    |
| **OVERALL** | 85%+   | 71.40%   | 78.6% | +7.2%  | PARTIAL |

### Remaining Work

The update.py module requires additional coverage (Target: 85%, Achieved: 72.8%, Gap: -12.2%).

**Uncovered areas**:

- Complex migration edge cases (legacy to current format)
- Backup creation failure scenarios (disk full, permission denied)
- Template sync merge conflict resolution
- Concurrent update attempt locking
- Custom element preservation during sync
- Advanced PyPI network error scenarios

**Recommendation**: Create SPEC-COVERAGE-004 to focus specifically on update.py remaining coverage gaps.

---

---

## Overview

### Purpose

Achieve 85%+ overall test coverage for MoAI-ADK CLI user commands through comprehensive TDD implementation, focusing on critical user-facing command modules (rank.py, update.py, switch.py, language.py, status.py) with complete test scenario coverage including error paths, edge cases, and integration tests.

### Scope

- **rank.py** (Current: 63.11%, Target: 85%): OAuth registration flow, status API calls, sync operations, credential management
- **update.py** (Current: 65.94%, Target: 85%): Migration functions, settings preservation/restore, upgrade workflows, installer detection
- **switch.py** (Current: 95.00%, Target: 95%): Error paths, credential edge cases, environment variable handling
- **language.py** (Current: 98.79%, Target: 98%): Windows-specific paths, encoding handling
- **status.py** (Current: 97.67%, Target: 97%): Status display edge cases, Windows compatibility

### Background

MoAI-ADK enforces TRUST-5 framework with 85% minimum test coverage threshold. Current overall coverage is 71.4%, below the target. CLI user commands are the primary user interaction points and require comprehensive testing to ensure reliability across different platforms (Windows, macOS, Linux) and usage scenarios.

---

## Environment and Assumptions

### Environment

- Python 3.11+ execution environment with pytest 8.4+
- Coverage measurement via pytest-cov with 85% fail_under threshold
- Platform-specific tests for Windows (win32), macOS (darwin), Linux
- Mock frameworks available: unittest.mock, pytest-mock
- Network mocking for external API calls (rank status API, PyPI version checks)

### Assumptions

1. All existing test patterns in `tests/unit/cli/commands/` are reference standards
2. AAA (Arrange-Act-Assert) test pattern is the default structure
3. Mock objects are used for subprocess calls, file operations, and network requests
4. Platform-specific tests use `@pytest.mark.skipif` decorators
5. Coverage reports include HTML output for detailed analysis
6. Error paths are tested with explicit exception assertions

---

## Requirements (EARS Format)

### T1: rank.py Test Coverage Enhancement (Current: 63.11% -> Target: 85%)

#### Ubiquitous Requirements (Always Active)

- The system **shall always** validate OAuth flow state transitions (pending, authorized, failed)
- The system **shall always** handle API network errors gracefully with retry logic
- The system **shall always** verify credential storage format (JSON structure, encryption)

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** user executes `moai rank register` **THEN** the system **shall** test browser opening mechanism
- **WHEN** OAuth callback is received **THEN** the system **shall** validate API key storage
- **WHEN** `moai rank status` is called **THEN** the system **shall** mock API responses and test display formatting
- **WHEN** background sync is triggered **THEN** the system **shall** test async session upload behavior
- **WHEN** logout command executes **THEN** the system **shall** verify credential removal

#### State-Driven Requirements (Conditional)

- **IF** user is already registered **THEN** the system **shall** test re-registration flow (confirmation prompt)
- **IF** network request fails **THEN** the system **shall** test retry behavior and error messages
- **IF** credential file is corrupted **THEN** the system **shall** test error recovery
- **IF** sync operation is interrupted **THEN** the system **shall** test partial state handling

#### Unwanted Requirements (Prohibited)

- The system **shall not** store API keys in plaintext without encryption
- The system **shall not** proceed with sync if previous session data is invalid
- The system **shall not** expose OAuth tokens in logs or error messages

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** test rate limiting behavior for API calls
- **Where possible**, the system **should** validate token formatting functions (K/M suffixes)

---

### T2: update.py Test Coverage Enhancement (Current: 65.94% -> Target: 85%)

#### Ubiquitous Requirements (Always Active)

- The system **shall always** validate installer detection (uv tool, pipx, pip)
- The system **shall always** test version comparison logic (current vs latest)
- The system **shall always** verify template sync backup operations

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** `moai update` is executed **THEN** the system **shall** test version checking workflow
- **WHEN** new version is available **THEN** the system **shall** test upgrade command execution
- **WHEN** template sync runs **THEN** the system **shall** test merge conflict detection
- **WHEN** migration is required **THEN** the system **shall** test config version migration
- **WHEN** stale cache is detected **THEN** the system **shall** test cache clearing workflow

#### State-Driven Requirements (Conditional)

- **IF** no installer is found **THEN** the system **shall** test error message display
- **IF** backup creation fails **THEN** the system **shall** test abort behavior
- **IF** network is unavailable **THEN** the system **shall** test offline mode behavior
- **IF** manual merge is requested **THEN** the system **shall** test merge guide generation

#### Unwanted Requirements (Prohibited)

- The system **shall not** proceed with update without backup confirmation
- The system **shall not** delete user configuration during template sync
- The system **shall not** allow downgrade without explicit user confirmation

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** test performance optimization skip logic for up-to-date projects
- **Where possible**, the system **should** validate custom element preservation during sync

---

### T3: switch.py Test Coverage Enhancement (Current: 95.00% -> Target: 95%)

#### Ubiquitous Requirements (Always Active)

- The system **shall always** validate credential source priority (.env.glm > credentials.yaml > environment)
- The system **shall always** test environment variable substitution in configuration files
- The system **shall always** verify LLM backend switching logic (Claude <-> GLM)

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** `moai glm` is executed **THEN** the system **shall** test GLM credential storage
- **WHEN** `moai claude` is executed **THEN** the system **shall** test Claude credential restoration
- **WHEN** configuration contains `${VAR}` patterns **THEN** the system **shall** test substitution logic
- **WHEN** credential is missing **THEN** the system **shall** test prompt for credential input

#### State-Driven Requirements (Conditional)

- **IF** .env.glm file exists **THEN** the system **shall** test priority over environment variable
- **IF** credential file is corrupted **THEN** the system **shall** test error handling
- **IF** both GLM and Claude credentials are missing **THEN** the system **shall** test error message

#### Unwanted Requirements (Prohibited)

- The system **shall not** store credentials in plaintext without user confirmation
- The system **shall not** expose API keys in error messages or logs

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** test credential validation (format check for API keys)
- **Where possible**, the system **should** test configuration file backup before modification

---

### T4: language.py Test Coverage Enhancement (Current: 98.79% -> Target: 98%)

#### Ubiquitous Requirements (Always Active)

- The system **shall always** validate language configuration file parsing
- The system **shall always** test language detection from user input

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** `moai language` is executed **THEN** the system **shall** test language selection workflow
- **WHEN** configuration file is updated **THEN** the system **shall** test YAML serialization

#### State-Driven Requirements (Conditional)

- **IF** on Windows platform **THEN** the system **shall** test UTF-8 encoding handling
- **IF** invalid language code is provided **THEN** the system **shall** test error message

#### Unwanted Requirements (Prohibited)

- The system **shall not** allow invalid language codes (must be in supported list)

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** test language auto-detection from system locale

---

### T5: status.py Test Coverage Enhancement (Current: 97.67% -> Target: 97%)

#### Ubiquitous Requirements (Always Active)

- The system **shall always** validate status information gathering from all modules
- The system **shall always** test status display formatting (tables, panels, colors)

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** `moai status` is executed **THEN** the system **shall** test status collection
- **WHEN** module status is unavailable **THEN** the system **shall** test graceful degradation

#### State-Driven Requirements (Conditional)

- **IF** on Windows platform **THEN** the system **shall** test console encoding handling
- **IF** configuration is missing **THEN** the system **shall** test default status display

#### Unwanted Requirements (Prohibited)

- The system **shall not** crash when individual module status checks fail

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** test status caching for performance

---

## Traceability

### Related Documents

- [CLAUDE.md](../../../../CLAUDE.md) - Execution directives and TRUST-5 framework
- [moai-foundation-core](../../../../.claude/skills/moai-foundation-core) - SPEC-First TDD methodology
- [moai-workflow-spec](../../../../.claude/skills/moai-workflow-spec) - SPEC creation workflow
- [moai-lang-python](../../../../.claude/skills/moai-lang-python) - Python 3.13 patterns and pytest
- [pyproject.toml](../../../../pyproject.toml) - Coverage configuration (fail_under = 85)

### Related SPECs

- SPEC-TAG-001: TAG System v2.0 Phase 1 (Traceability for test annotations)
- SPEC-COVERAGE-001: Existing coverage improvement SPEC

### Test Pattern References

- `tests/unit/cli/commands/test_update_coverage.py` - AAA pattern with @patch decorators
- `tests/unit/cli/commands/test_update_comprehensive.py` - Comprehensive edge case coverage
- `tests/unit/cli/commands/test_update_final.py` - Final validation and integration tests

### Next Steps

```bash
# TDD Execution
/moai:2-run SPEC-COVERAGE-002

# Documentation Sync
/moai:3-sync SPEC-COVERAGE-002
```

---

## References

- pytest Documentation: [pytest.org](https://docs.pytest.org/)
- pytest-cov Documentation: [pytest-cov.readthedocs.io](https://pytest-cov.readthedocs.io/)
- AAA Testing Pattern: Industry standard for test structure
- Python unittest.mock: [docs.python.org/library/unittest.mock.html](https://docs.python.org/3/library/unittest.mock.html)
- MoAI-ADK Testing Guide: Internal documentation for test patterns
