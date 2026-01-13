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
tags: [tdd, testing, coverage, core-modules, acceptance-criteria]
---

# Acceptance Criteria: SPEC-COVERAGE-003

## Overview

This document defines the acceptance criteria for achieving 85%+ test coverage across five MoAI-ADK core modules using TDD methodology.

---

## Quality Gates

### Gate 1: Coverage Threshold (CRITICAL)

**GIVEN** the test suite is executed with pytest-cov
**WHEN** coverage report is generated for all targeted modules
**THEN** each module **MUST** meet or exceed its target coverage:

- astgrep/analyzer.py: 85%+ (currently 79.63%)
- astgrep/rules.py: 95%+ (currently 90.91%)
- astgrep/models.py: 100% (maintain current)
- **main**.py: 100% (currently 99.08%)
- statusline/main.py: 90%+ (currently ~85%)
- **Overall project**: 85%+ (currently 83.44%)

**Verification Command**:

```bash
pytest --cov=src/moai_adk --cov-report=term-missing --cov-fail-under=85
```

### Gate 2: All Tests Pass (CRITICAL)

**GIVEN** the test suite includes all newly created tests
**WHEN** pytest is executed with verbose output
**THEN** all tests **MUST** pass with zero failures

- No FAILED test results
- No ERROR test results
- No SKIP test results (unless platform-specific)

**Verification Command**:

```bash
pytest tests/unit/test_astgrep/ tests/unit/test_main/ tests/unit/test_statusline/ -v
```

### Gate 3: No Regression (CRITICAL)

**GIVEN** the existing test suite
**WHEN** new tests are added
**THEN** all existing tests **MUST** continue to pass

- No regression in existing functionality
- No breaking changes to public APIs
- No changes to existing behavior

**Verification Command**:

```bash
pytest tests/ -v --cov=src/moai_adk --cov-report=term-missing
```

### Gate 4: Code Quality (HIGH)

**GIVEN** the new test files
**WHEN** code quality tools are executed
**THEN** all quality checks **MUST** pass:

- Zero ruff linter warnings
- Zero mypy type checking errors
- Zero black formatting issues
- Zero bandit security warnings

**Verification Commands**:

```bash
# Linting
ruff check src/moai_adk/ tests/

# Type checking
mypy src/moai_adk/

# Security
bandit -r src/moai_adk/
```

---

## Module-Specific Acceptance Criteria

### AC-001: astgrep/analyzer.py Coverage

**Scenario 1: Private Helper Methods**

**GIVEN** the MoAIASTGrepAnalyzer class
**WHEN** private methods are tested with various inputs
**THEN** the following methods **MUST** have test coverage:

- `_detect_language()`: Test with unknown extensions returns "text"
- `_should_include_file()`: Test with overlapping include/exclude patterns
- `_parse_sg_match()`: Test with malformed JSON output
- `_parse_pattern_search_output()`: Test with invalid JSON structure

**Example Test**:

```python
def test_detect_language_unknown_extension():
    analyzer = MoAIASTGrepAnalyzer()
    result = analyzer._detect_language('file.unknown')
    assert result == 'text'
```

**Scenario 2: Timeout Behavior**

**GIVEN** subprocess.run() is mocked to raise TimeoutExpired
**WHEN** \_run_sg_scan() is called
**THEN** the method **MUST** return empty list gracefully
**AND** no exception **MUST** propagate to caller

**Example Test**:

```python
@patch('subprocess.run')
def test_run_sg_scan_timeout(mock_run):
    mock_run.side_effect = subprocess.TimeoutExpired('sg', 60)
    analyzer = MoAIASTGrepAnalyzer()
    result = analyzer._run_sg_scan('test.py', 'python', ScanConfig())
    assert result == []
```

**Scenario 3: File Permission Errors**

**GIVEN** a file path that does not exist
**WHEN** scan_file() is called
**THEN** FileNotFoundError **MUST** be raised
**AND** error message **MUST** include the file path

**Example Test**:

```python
def test_scan_file_not_found():
    analyzer = MoAIASTGrepAnalyzer()
    with pytest.raises(FileNotFoundError, match="nonexistent.py"):
        analyzer.scan_file('nonexistent.py')
```

**Scenario 4: Malformed Output Parsing**

**GIVEN** sg CLI returns invalid JSON
**WHEN** \_parse_sg_output() is called
**THEN** empty list **MUST** be returned
**AND** no exception **MUST** propagate

**Example Test**:

```python
def test_parse_sg_output_invalid_json():
    analyzer = MoAIASTGrepAnalyzer()
    result = analyzer._parse_sg_output('invalid json', 'test.py')
    assert result == []
```

**Acceptance Metrics**:

- Coverage: 85%+ (up from 79.63%)
- New tests: 40-50 tests
- All tests pass: 100%
- No regression: ✓

### AC-002: astgrep/rules.py Coverage

**Scenario 1: Unicode YAML Content**

**GIVEN** a YAML file with UTF-8 encoded content
**WHEN** load_from_file() is called
**THEN** rules **MUST** be loaded successfully
**AND** multi-byte characters **MUST** be preserved

**Example Test**:

```python
def test_load_unicode_yaml(tmp_path):
    yaml_file = tmp_path / "rules.yaml"
    yaml_file.write_text("id: test\nmessage: 'Hello 世界'", encoding='utf-8')
    loader = RuleLoader()
    rules = loader.load_from_file(str(yaml_file))
    assert len(rules) == 1
    assert '世界' in rules[0].message
```

**Scenario 2: Duplicate Rule IDs**

**GIVEN** multiple YAML files with duplicate rule IDs
**WHEN** rules are loaded from directory
**THEN** all rules **MUST** be loaded
**AND** duplicates **MUST** be allowed (no collision detection)

**Example Test**:

```python
def test_load_duplicate_rule_ids(tmp_path):
    file1 = tmp_path / "rules1.yaml"
    file2 = tmp_path / "rules2.yaml"
    file1.write_text("id: test\nlanguage: python\npattern: test")
    file2.write_text("id: test\nlanguage: javascript\npattern: test")
    loader = RuleLoader()
    rules = loader.load_from_directory(str(tmp_path))
    assert len(rules) == 2
```

**Scenario 3: Language Alias Support**

**GIVEN** a rule with language "js"
**WHEN** get_rules_for_language("javascript") is called
**THEN** the rule **MUST NOT** be returned (no alias support)

**Example Test**:

```python
def test_get_rules_language_no_alias():
    loader = RuleLoader()
    loader._rules = [Rule(id="test", language="js", severity="warning", message="test", pattern="test")]
    rules = loader.get_rules_for_language("javascript")
    assert len(rules) == 0  # No alias support
```

**Scenario 4: Concurrent Access Safety**

**GIVEN** multiple threads loading rules simultaneously
**WHEN** load_from_directory() is called concurrently
**THEN** all rules **MUST** be loaded
**AND** no race conditions **MUST** occur

**Example Test**:

```python
import threading

def test_concurrent_rule_loading(tmp_path):
    yaml_file = tmp_path / "rules.yaml"
    yaml_file.write_text("id: test\nlanguage: python\npattern: test")
    loader = RuleLoader()

    def load_rules():
        loader.load_from_file(str(yaml_file))

    threads = [threading.Thread(target=load_rules) for _ in range(10)]
    for t in threads:
        t.start()
    for t in threads:
        t.join()

    assert len(loader._rules) == 10  # All threads loaded successfully
```

**Acceptance Metrics**:

- Coverage: 95%+ (up from 90.91%)
- New tests: 15-20 tests
- All tests pass: 100%
- No regression: ✓

### AC-003: astgrep/models.py Edge Cases

**Scenario 1: Negative Scan Time**

**GIVEN** a ScanResult with negative scan_time_ms
**WHEN** the result is created
**THEN** the value **MUST** be accepted (no validation)

**Example Test**:

```python
def test_scan_result_negative_time():
    result = ScanResult(
        file_path='test.py',
        matches=[],
        scan_time_ms=-100,
        language='python'
    )
    assert result.scan_time_ms == -100  # No validation
```

**Scenario 2: Invalid Severity**

**GIVEN** an ASTMatch with severity not in allowed values
**WHEN** the match is created
**THEN** the value **MUST** be accepted (no validation)

**Example Test**:

```python
def test_ast_match_invalid_severity():
    match = ASTMatch(
        rule_id='test',
        severity='invalid',  # Not in error/warning/info/hint
        message='test',
        file_path='test.py',
        range=Range(start=Position(0, 0), end=Position(0, 10)),
        suggested_fix=None
    )
    assert match.severity == 'invalid'  # No validation
```

**Scenario 3: Empty Collections**

**GIVEN** a ProjectScanResult with empty results
**WHEN** the result is created
**THEN** empty collections **MUST** be accepted

**Example Test**:

```python
def test_project_scan_result_empty():
    result = ProjectScanResult(
        project_path='/test',
        files_scanned=0,
        total_matches=0,
        results_by_file={},
        summary_by_severity={'error': 0, 'warning': 0, 'info': 0, 'hint': 0},
        scan_time_ms=0
    )
    assert result.files_scanned == 0
    assert result.total_matches == 0
    assert len(result.results_by_file) == 0
```

**Acceptance Metrics**:

- Coverage: 100% (maintain current)
- New tests: 10-15 edge case tests
- All tests pass: 100%
- Edge cases documented: ✓

### AC-004: **main**.py Coverage

**Scenario 1: Generic Exception Handling**

**GIVEN** an unexpected exception in main()
**WHEN** the exception occurs
**THEN** the exception **MUST** be caught
**AND** error message **MUST** be printed
**AND** exit code 1 **MUST** be returned

**Example Test**:

```python
@patch('moai_adk.__main__.cli')
def test_main_exception_handling(mock_cli):
    mock_cli.side_effect = Exception('Test error')
    from moai_adk.__main__ import main
    exit_code = main()
    assert exit_code == 1
```

**Scenario 2: Lazy Loading Verification**

**GIVEN** the CLI entry point
**WHEN** a command is imported
**THEN** the module **MUST NOT** be imported until command is invoked

**Example Test**:

```python
def test_init_command_lazy_loading():
    import sys
    # Clear module if already imported
    sys.modules.pop('moai_adk.cli.commands.init', None)

    from moai_adk.__main__ import cli
    # Init module should not be imported yet
    assert 'moai_adk.cli.commands.init' not in sys.modules
```

**Scenario 3: Windows Encoding Handling**

**GIVEN** Windows platform
**WHEN** console output contains emoji
**THEN** output **MUST** be rendered without encoding errors

**Example Test**:

```python
@pytest.mark.skipif(sys.platform != 'win32', reason='Windows only')
def test_windows_emoji_rendering():
    from moai_adk.__main__ import show_logo
    # Should not raise UnicodeEncodeError
    show_logo()
```

**Acceptance Metrics**:

- Coverage: 100% (up from 99.08%)
- New tests: 10-15 tests
- All tests pass: 100%
- No regression: ✓

### AC-005: statusline/main.py Coverage

**Scenario 1: Invalid JSON from stdin**

**GIVEN** stdin contains invalid JSON
**WHEN** read_session_context() is called
**THEN** empty dictionary **MUST** be returned
**AND** no exception **MUST** propagate

**Example Test**:

```python
@patch('sys.stdin.read')
def test_read_session_context_invalid_json(mock_read):
    mock_read.return_value = '{invalid json}'
    from moai_adk.statusline.main import read_session_context
    result = read_session_context()
    assert result == {}
```

**Scenario 2: Git Collector Fallback**

**GIVEN** git repository access fails
**WHEN** safe_collect_git_info() is called
**THEN** "N/A" **MUST** be returned for branch
**AND** empty string **MUST** be returned for status

**Example Test**:

```python
@patch('moai_adk.statusline.main.GitCollector')
def test_safe_collect_git_info_error(mock_collector):
    mock_collector.side_effect = OSError('Git error')
    from moai_adk.statusline.main import safe_collect_git_info
    branch, status = safe_collect_git_info()
    assert branch == "N/A"
    assert status == ""
```

**Scenario 3: Empty Session Context**

**GIVEN** empty session context dictionary
**WHEN** build_statusline_data() is called
**THEN** statusline **MUST** be generated with fallback values
**AND** no exception **MUST** occur

**Example Test**:

```python
def test_build_statusline_empty_context():
    from moai_adk.statusline.main import build_statusline_data
    result = build_statusline_data({}, mode='compact')
    assert isinstance(result, str)
    assert len(result) > 0
```

**Scenario 4: Debug Mode**

**GIVEN** MOAI_STATUSLINE_DEBUG environment variable is set
**WHEN** main() is executed
**THEN** debug information **MUST** be written to stderr

**Example Test**:

```python
@patch.dict(os.environ, {'MOAI_STATUSLINE_DEBUG': '1'})
@patch('sys.stderr.write')
def test_main_debug_mode(mock_write):
    from moai_adk.statusline.main import main
    main()
    assert mock_write.called
    assert '[DEBUG]' in str(mock_write.call_args)
```

**Acceptance Metrics**:

- Coverage: 90%+ (up from ~85%)
- New tests: 40-50 tests
- All tests pass: 100%
- No regression: ✓

---

## Definition of Done

A module is considered complete when:

1. **Coverage**: Target coverage percentage achieved
2. **Tests Pass**: All new tests pass consistently
3. **No Regression**: Existing tests continue to pass
4. **Code Quality**: Zero linter and type checker warnings
5. **Documentation**: Edge cases and error handling documented
6. **Review**: Code reviewed and approved

The entire SPEC is complete when:

1. **All Modules**: All five modules meet acceptance criteria
2. **Overall Coverage**: Project coverage >= 85%
3. **Quality Gates**: All quality gates pass
4. **CI/CD**: Pipeline passes all checks
5. **Documentation**: Documentation updated with `/moai:3-sync`

---

## Test Execution Checklist

### Before Committing

- [ ] Run full test suite: `pytest tests/ -v`
- [ ] Check coverage: `pytest --cov=src/moai_adk --cov-report=term-missing`
- [ ] Verify coverage threshold: `pytest --cov-fail-under=85`
- [ ] Run linting: `ruff check src/ tests/`
- [ ] Run type checking: `mypy src/`
- [ ] Run security check: `bandit -r src/`

### Before Merging

- [ ] All acceptance criteria met
- [ ] Coverage report reviewed
- [ ] Test quality verified
- [ ] Documentation updated
- [ ] Code review approved
- [ ] CI/CD pipeline passing

---

## Verification Commands

### Quick Verification

```bash
# Run all tests with coverage
pytest tests/ -v --cov=src/moai_adk --cov-report=term-missing

# Check specific module coverage
pytest tests/unit/test_astgrep/test_analyzer_coverage.py -v --cov=src/moai_adk/astgrep/analyzer --cov-report=term-missing
```

### Comprehensive Verification

```bash
# Generate HTML coverage report
pytest --cov=src/moai_adk --cov-report=html

# Open coverage report
open htmlcov/index.html

# Check coverage for all targeted modules
pytest tests/unit/test_astgrep/ tests/unit/test_main/ tests/unit/test_statusline/ -v --cov=src/moai_adk --cov-report=term-missing
```

### Quality Verification

```bash
# Run all quality checks
pytest tests/ -v && ruff check src/ tests/ && mypy src/

# Run security scan
bandit -r src/moai_adk/

# Check for common issues
pytest tests/ --cache-clear --disable-warnings
```

---

## Success Criteria Summary

| Criterion            | Target                   | Status |
| -------------------- | ------------------------ | ------ |
| astgrep/analyzer.py  | 85%+ coverage            |        |
| astgrep/rules.py     | 95%+ coverage            |        |
| astgrep/models.py    | 100% coverage (maintain) |        |
| **main**.py          | 100% coverage            |        |
| statusline/main.py   | 90%+ coverage            |        |
| **Overall Coverage** | **85%+**                 |        |
| **All Tests Pass**   | **100%**                 |        |
| **No Regression**    | **0 failures**           |        |
| **Code Quality**     | **0 warnings/errors**    |        |
