---
id: SPEC-COVERAGE-003
version: 1.0.0
status: Planned
created: 2026-01-13
updated: 2026-01-13
author: Alfred
priority: HIGH
domain: COVERAGE
lifecycle: spec-first
tags: [tdd, testing, coverage, core-modules, implementation-plan]
---

# Implementation Plan: SPEC-COVERAGE-003

## Overview

This plan outlines the implementation strategy for achieving 85%+ test coverage across five MoAI-ADK core modules using TDD methodology.

---

## Milestones

### Primary Goal (HIGH Priority)

Achieve 85%+ coverage for astgrep/analyzer.py and statusline/main.py.

**Success Criteria**:

- astgrep/analyzer.py: 79.63% -> 85%+ (cover 33+ missing lines)
- statusline/main.py: ~85% -> 90%+ (cover ~50 missing lines)
- All new tests pass
- No regression in existing tests

### Secondary Goal (MEDIUM Priority)

Achieve 95%+ coverage for astgrep/rules.py and 100% for **main**.py.

**Success Criteria**:

- astgrep/rules.py: 90.91% -> 95%+ (cover 6 missing lines)
- **main**.py: 99.08% -> 100% (cover 1 missing line)
- Edge case coverage for models.py
- All error paths tested

### Optional Goal (LOW Priority)

Add performance benchmarks and integration tests.

**Success Criteria**:

- Performance tests for large file scanning
- Integration tests with isolated file system
- Benchmark regression detection

---

## Technical Approach

### Test Architecture

**Test File Organization**:

```
tests/unit/
├── test_astgrep/
│   ├── test_analyzer_coverage.py      # analyzer.py tests
│   ├── test_rules_coverage.py         # rules.py tests
│   └── test_models_coverage.py        # models.py edge cases
├── test_main/
│   └── test_main_coverage.py          # __main__.py tests
└── test_statusline/
    └── test_statusline_coverage.py    # statusline/main.py tests
```

**Test Structure Pattern**:

```python
# AAA Pattern (Arrange-Act-Assert)
def test_feature_with_scenario():
    # Arrange: Set up test data and mocks
    mock_subprocess = Mock()
    analyzer = MoAIASTGrepAnalyzer()

    # Act: Execute the code under test
    result = analyzer.scan_file("test.py")

    # Assert: Verify expected outcomes
    assert result.file_path == "test.py"
    assert result.language == "python"
```

### Mocking Strategy

**Subprocess Mocking**:

```python
from unittest.mock import Mock, patch
import subprocess

@patch('subprocess.run')
def test_sg_scan_timeout(mock_run):
    # Arrange
    mock_run.side_effect = subprocess.TimeoutExpired('sg', 60)
    analyzer = MoAIASTGrepAnalyzer()

    # Act
    result = analyzer._run_sg_scan('test.py', 'python', ScanConfig())

    # Assert
    assert result == []
```

**File System Mocking**:

```python
from pathlib import Path
from unittest.mock import patch, MagicMock

@patch('pathlib.Path.exists')
def test_scan_file_not_found(mock_exists):
    # Arrange
    mock_exists.return_value = False
    analyzer = MoAIASTGrepAnalyzer()

    # Act & Assert
    with pytest.raises(FileNotFoundError):
        analyzer.scan_file('nonexistent.py')
```

### Test Categories

**1. Unit Tests for Private Methods**

Target: Internal helper methods in analyzer.py

```python
class TestDetectLanguage:
    def test_unknown_extension_returns_text(self):
        analyzer = MoAIASTGrepAnalyzer()
        result = analyzer._detect_language('file.unknown')
        assert result == 'text'

    def test_python_extension(self):
        analyzer = MoAIASTGrepAnalyzer()
        result = analyzer._detect_language('script.py')
        assert result == 'python'
```

**2. Edge Case Tests**

Target: Boundary conditions and invalid inputs

```python
class TestScanResultEdgeCases:
    def test_negative_scan_time(self):
        result = ScanResult(
            file_path='test.py',
            matches=[],
            scan_time_ms=-100,  # Invalid
            language='python'
        )
        # Verify behavior with invalid data

    def test_empty_matches_list(self):
        result = ScanResult(
            file_path='test.py',
            matches=[],  # Edge case
            scan_time_ms=50,
            language='python'
        )
        assert len(result.matches) == 0
```

**3. Error Handling Tests**

Target: Exception paths and fallback behavior

```python
class TestErrorHandling:
    def test_json_decode_error_in_parse_sg_output(self):
        analyzer = MoAIASTGrepAnalyzer()
        result = analyzer._parse_sg_output('invalid json', 'test.py')
        assert result == []

    def test_subprocess_timeout(self):
        with patch('subprocess.run') as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired('sg', 60)
            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer._run_sg_scan('test.py', 'python', ScanConfig())
            assert result == []
```

**4. Platform-Specific Tests**

Target: Windows, macOS, Linux variations

```python
import pytest
import sys

@pytest.mark.skipif(sys.platform != 'win32', reason='Windows only')
def test_windows_encoding_handling():
    # Windows-specific encoding tests
    pass

@pytest.mark.skipif(sys.platform == 'win32', reason='Non-Windows only')
def test_unix_path_handling():
    # Unix-specific path tests
    pass
```

---

## Implementation Steps

### Phase 1: astgrep/analyzer.py Coverage (HIGH Priority)

**Step 1.1: Private Method Tests**

- Create test_analyzer_coverage.py
- Test \_detect_language() with unknown extensions
- Test \_should_include_file() with pattern combinations
- Test \_parse_sg_match() with malformed output

**Step 1.2: Timeout Behavior Tests**

- Test subprocess.TimeoutExpired handling
- Test timeout in \_run_sg_scan()
- Test timeout in pattern_search()
- Test timeout in pattern_replace()

**Step 1.3: File Permission Tests**

- Test FileNotFoundError in scan_file()
- Test FileNotFoundError in scan_project()
- Test permission denied error messages

**Step 1.4: Malformed Output Tests**

- Test JSONDecodeError in \_parse_sg_output()
- Test missing fields in sg match output
- Test invalid range data structures

**Step 1.5: Verification**

- Run pytest with coverage
- Verify 85%+ coverage achieved
- Review coverage report for remaining gaps

### Phase 2: astgrep/rules.py Coverage (MEDIUM Priority)

**Step 2.1: Unicode YAML Tests**

- Create test_rules_coverage.py
- Test loading UTF-8 encoded YAML files
- Test multi-byte characters in rule messages
- Test Unicode pattern strings

**Step 2.2: Duplicate Rule ID Tests**

- Test loading multiple files with duplicate IDs
- Test rule ID collision behavior
- Test get_rules_for_language() with duplicates

**Step 2.3: Language Alias Tests**

- Test case-insensitive language matching
- Test language name variations (js vs javascript)
- Test invalid language names

**Step 2.4: Concurrent Access Tests**

- Test thread-safe rule loading
- Test concurrent file reading
- Test shared \_rules list access

**Step 2.5: Verification**

- Run pytest with coverage
- Verify 95%+ coverage achieved
- Review coverage report for remaining gaps

### Phase 3: astgrep/models.py Edge Cases (LOW Priority)

**Step 3.1: Dataclass Validation Tests**

- Create test_models_coverage.py
- Test negative scan_time_ms values
- Test invalid severity values
- Test empty collections

**Step 3.2: Type Safety Tests**

- Test integer overflow scenarios
- Test string encoding in file_path
- Test range boundary conditions

**Step 3.3: Verification**

- Run pytest with coverage
- Verify 100% coverage maintained
- Document edge case behavior

### Phase 4: **main**.py Coverage (MEDIUM Priority)

**Step 4.1: Exception Handling Tests**

- Create test_main_coverage.py
- Test main() generic exception handling
- Test console.flush() exception handling
- Test lazy loading import failures

**Step 4.2: Lazy Loading Tests**

- Test init command lazy loading
- Test doctor command lazy loading
- Test status command lazy loading
- Verify modules imported only when invoked

**Step 4.3: Platform-Specific Tests**

- Test Windows console encoding
- Test emoji rendering on different platforms
- Test path handling on Windows

**Step 4.4: Verification**

- Run pytest with coverage
- Verify 100% coverage achieved
- Review coverage report for remaining gaps

### Phase 5: statusline/main.py Coverage (HIGH Priority)

**Step 5.1: Renderer Constraint Tests**

- Create test_statusline_coverage.py
- Test mode parameter variations
- Test missing fields in StatuslineData
- Test invalid JSON from stdin

**Step 5.2: Fallback Behavior Tests**

- Test safe_collect_git_info() error handling
- Test safe_collect_duration() error handling
- Test safe_collect_alfred_task() error handling
- Test safe_collect_version() error handling
- Test safe_collect_memory() error handling
- Test safe_check_update() error handling

**Step 5.3: Performance Tests**

- Test large session context processing
- Test large git repository processing
- Test long duration formatting

**Step 5.4: Edge Case Tests**

- Test empty session context
- Test missing model information
- Test invalid UTF-8 in stdin
- Test debug mode environment variable

**Step 5.5: Verification**

- Run pytest with coverage
- Verify 90%+ coverage achieved
- Review coverage report for remaining gaps

---

## Quality Assurance

### Pre-Commit Checks

```bash
# Run tests before commit
pytest tests/unit/test_astgrep/ -v --cov=src/moai_adk/astgrep --cov-report=term-missing

# Check coverage threshold
pytest --cov=src/moai_adk --cov-fail-under=85

# Run linting
ruff check src/moai_adk/ tests/

# Run type checking
mypy src/moai_adk/
```

### Coverage Verification

```bash
# Generate HTML coverage report
pytest --cov=src/moai_adk --cov-report=html

# Open report in browser
open htmlcov/index.html

# Check specific module coverage
pytest --cov=src/moai_adk/astgrep/analyzer --cov-report=term-missing
```

### Test Quality Criteria

- All tests use AAA pattern (Arrange-Act-Assert)
- Test names clearly describe what is being tested
- Mock objects used for external dependencies
- No actual subprocess calls or file system modifications
- Tests are independent and can run in any order
- Edge cases and error paths covered

---

## Risks and Mitigation

### Risk 1: Flaky Tests Due to Subprocess Mocking

**Mitigation**:

- Use unittest.mock consistently
- Verify mock return values match real subprocess behavior
- Add integration tests for critical subprocess paths

### Risk 2: Platform-Specific Test Failures

**Mitigation**:

- Use pytest.mark.skipif for platform-specific tests
- Test on all target platforms (Windows, macOS, Linux)
- Mock platform-specific behavior where appropriate

### Risk 3: Coverage Regression in Existing Modules

**Mitigation**:

- Run full test suite before committing
- Monitor coverage trends across commits
- Set up CI coverage regression detection

### Risk 4: Mock Behavior Mismatch

**Mitigation**:

- Document expected subprocess behavior
- Add integration tests for verification
- Review mock implementations regularly

---

## Dependencies

### Internal Dependencies

- astgrep/models.py - Data models used by analyzer.py and rules.py
- lsp/models.py - Position and Range models used by ASTMatch
- cli/commands/\* - Lazy-loaded commands in **main**.py
- statusline/\* - Collector and renderer modules used by statusline/main.py

### External Dependencies

- pytest >= 8.4.2
- pytest-cov >= 7.0.0
- pytest-mock >= 3.15.1
- unittest.mock (Python standard library)
- freezegun (for time-related testing)
- pytest-timeout (for subprocess timeout testing)

---

## Success Metrics

### Coverage Metrics

| Module              | Before | After | Delta  | Status |
| ------------------- | ------ | ----- | ------ | ------ |
| astgrep/analyzer.py | 79.63% | 85%+  | +5.37% |        |
| astgrep/rules.py    | 90.91% | 95%+  | +4.09% |        |
| astgrep/models.py   | 100%   | 100%  | 0%     |        |
| **main**.py         | 99.08% | 100%  | +0.92% |        |
| statusline/main.py  | ~85%   | 90%+  | +5%+   |        |
| **Overall**         | 83.44% | 85%+  | +1.56% |        |

### Test Count Metrics

- New tests created: ~150-200 tests
- Test execution time: < 30 seconds
- Test pass rate: 100%

### Quality Metrics

- Zero linter warnings
- Zero type checking errors
- All tests pass consistently
- No test interdependence

---

## Next Steps

1. Review and approve this implementation plan
2. Execute TDD workflow: `/moai:2-run SPEC-COVERAGE-003`
3. Monitor coverage progress during implementation
4. Verify quality gates before merging
5. Update documentation: `/moai:3-sync SPEC-COVERAGE-003`
