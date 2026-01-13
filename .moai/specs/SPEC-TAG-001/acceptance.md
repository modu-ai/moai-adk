---
id: SPEC-TAG-001
version: "1.0.0"
status: "draft"
created: "2026-01-13"
updated: "2026-01-13"
author: "Alfred"
priority: "HIGH"
tags: [tag-system, traceability, tdd-integration, code-spec-mapping, pre-commit]
spec_id: SPEC-TAG-001
---

# Acceptance Criteria: SPEC-TAG-001 TAG System v2.0 Phase 1

## Test Strategy Overview

This document defines comprehensive acceptance criteria for TAG System v2.0 Phase 1 implementation, including Given-When-Then test scenarios, edge cases, and TDD workflow integration.

### Test Coverage Goals

- Overall test coverage: 90%+ (unit), 85%+ (integration)
- Core module coverage: 95%+ (parser, validator, linkage)
- Edge case coverage: 80%+
- Mutation testing: Enabled for critical modules

---

## T1: TAG Pattern Definition

### Scenario 1.1: Valid TAG Format Recognition

**Given** a Python file contains valid TAG annotation
```python
# @SPEC SPEC-AUTH-001
def authenticate_user():
    pass
```
**When** TAG parser extracts TAGs from the file
**Then** system **shall** recognize SPEC-ID as `SPEC-AUTH-001` and verb as `impl` (default)

```python
def test_valid_tag_format():
    # Given
    source_code = '''
# @SPEC SPEC-AUTH-001
def authenticate_user():
    pass
'''
    # When
    tags = extract_tags(source_code)

    # Then
    assert len(tags) == 1
    assert tags[0].spec_id == "SPEC-AUTH-001"
    assert tags[0].verb == "impl"
```

### Scenario 1.2: TAG with Optional Verb

**Given** a Python file contains TAG with explicit verb
```python
# @SPEC SPEC-AUTH-001 verify
def test_auth_flow():
    pass
```
**When** TAG parser extracts TAGs from the file
**Then** system **shall** parse verb as `verify`

```python
def test_tag_with_verb():
    # Given
    source_code = '''
# @SPEC SPEC-AUTH-001 verify
def test_auth_flow():
    pass
'''
    # When
    tags = extract_tags(source_code)

    # Then
    assert tags[0].verb == "verify"
```

### Scenario 1.3: Invalid TAG Format Detection

**Given** a Python file contains malformed TAG
```python
# @spec auth-001  # Wrong case and format
def bad_example():
    pass
```
**When** TAG validator checks the TAG format
**Then** system **shall** reject with error message: "Invalid TAG format. Use @SPEC SPEC-ID"

```python
def test_invalid_tag_format():
    # Given
    tag = TAG(spec_id="auth-001", verb="impl", file_path=Path("test.py"), line=1)

    # When & Then
    assert not validate_tag_format(tag)
    assert "Invalid TAG format" in get_validation_error(tag)
```

### Edge Cases

**EC-1.1:** TAG in inline comment vs block comment
```python
def test_inline_vs_block_comment():
    # Both should be extracted
    source_code = '''
# @SPEC SPEC-001  # Inline comment
def func1(): pass

"""
@SPEC SPEC-002  # Block comment
"""
def func2(): pass
'''
    tags = extract_tags(source_code)
    assert len(tags) == 2
```

**EC-1.2:** Multiple TAGs in single file
```python
def test_multiple_tags():
    source_code = '''
# @SPEC SPEC-AUTH-001
def login(): pass

# @SPEC SPEC-AUTH-002 verify
def test_login(): pass
'''
    tags = extract_tags(source_code)
    assert len(tags) == 2
    assert tags[0].spec_id == "SPEC-AUTH-001"
    assert tags[1].spec_id == "SPEC-AUTH-002"
```

**EC-1.3:** TAG with special characters in description
```python
def test_tag_with_description():
    # TAG followed by descriptive text
    source_code = '''
# @SPEC SPEC-AUTH-001 impl - User authentication flow
def authenticate(): pass
'''
    tags = extract_tags(source_code)
    assert tags[0].spec_id == "SPEC-AUTH-001"
    assert tags[0].verb == "impl"
```

### Success Criteria

- [ ] 100% of valid TAGs are correctly extracted
- [ ] 100% of invalid TAG formats are rejected
- [ ] All 4 optional verbs (impl, verify, depends, related) are supported
- [ ] Default verb (impl) is applied when not specified

---

## T2: TAG Parser (Comment Extraction)

### Scenario 2.1: TAG Extraction from Python File

**Given** a Python file with TAG annotations
**When** parser processes the file
**Then** system **shall** extract all TAGs with file path and line numbers

```python
def test_tag_extraction_from_file():
    # Given
    test_file = Path("/tmp/test.py")
    test_file.write_text('''
# @SPEC SPEC-AUTH-001
def login():
    pass

# @SPEC SPEC-AUTH-002 verify
def test_login():
    pass
''')

    # When
    tags = extract_tags_from_file(test_file)

    # Then
    assert len(tags) == 2
    assert tags[0].file_path == test_file
    assert tags[0].line == 2  # Line number of first TAG
    assert tags[1].line == 7  # Line number of second TAG
```

### Scenario 2.2: Handling Files with Syntax Errors

**Given** a Python file with syntax errors
**When** parser attempts to extract TAGs
**Then** system **shall** log warning and return empty TAG list

```python
def test_syntax_error_handling():
    # Given
    test_file = Path("/tmp/bad_syntax.py")
    test_file.write_text('''
# @SPEC SPEC-001
def incomplete_function(
    # Missing closing parenthesis
''')

    # When
    tags = extract_tags_from_file(test_file)

    # Then
    assert tags == []  # Empty list, no exception
    assert_log_entry(level="WARNING", message="Syntax error in file")
```

### Scenario 2.3: No TAGs in File

**Given** a Python file without any TAG annotations
**When** parser processes the file
**Then** system **shall** return empty TAG list

```python
def test_no_tags():
    # Given
    test_file = Path("/tmp/no_tags.py")
    test_file.write_text('''
def regular_function():
    pass
''')

    # When
    tags = extract_tags_from_file(test_file)

    # Then
    assert tags == []
```

### Edge Cases

**EC-2.1:** TAG in string literal (should NOT be extracted)
```python
def test_tag_in_string_literal():
    source_code = '''
def help_text():
    return "Usage: @SPEC SPEC-001"
'''
    tags = extract_tags(source_code)
    assert len(tags) == 0  # TAGs in strings are ignored
```

**EC-2.2:** TAG with Unicode characters
```python
def test_unicode_tags():
    source_code = '''
# @SPEC SPEC-I18N-001 한글 태그
def internationalized():
    pass
'''
    tags = extract_tags(source_code)
    assert len(tags) == 1
    assert tags[0].spec_id == "SPEC-I18N-001"
```

**EC-2.3:** TAG in very large file (performance test)
```python
def test_large_file_performance():
    # Given - 10,000 line file
    lines = ['# @SPEC SPEC-{:03d}\n'.format(i) for i in range(10000)]
    test_file = create_temp_file('\n'.join(lines))

    # When
    start = time.time()
    tags = extract_tags_from_file(test_file)
    elapsed = time.time() - start

    # Then - Should complete in reasonable time
    assert len(tags) == 10000
    assert elapsed < 5.0  # 5 seconds for 10k TAGs
```

### Success Criteria

- [ ] Parser correctly extracts TAGs from 100% of valid Python files
- [ ] Parser handles syntax errors gracefully (no crashes)
- [ ] TAGs in string literals are correctly ignored
- [ ] Parser performance: <0.1s for typical files, <5s for 10k TAGs

---

## T3: Pre-commit Validation Hook

### Scenario 3.1: Valid TAGs Pass Validation

**Given** staged Python files contain valid TAGs
**When** developer commits changes
**Then** pre-commit hook **shall** exit with success code (0)

```python
def test_valid_tags_commit():
    # Given
    staged_files = [Path("auth.py")]
    staged_files[0].write_text('''
# @SPEC SPEC-AUTH-001
def login():
    pass
''')

    # When
    exit_code = run_pre_commit_hook(staged_files)

    # Then
    assert exit_code == 0
```

### Scenario 3.2: Invalid TAG Format Displayed

**Given** staged Python file contains malformed TAG
**When** developer commits changes
**Then** pre-commit hook **shall** display error with file:line and exit with warning (1)

```python
def test_invalid_tag_format_warning():
    # Given
    staged_file = Path("bad.py")
    staged_file.write_text('''
# @spec auth-001  # Wrong format
def bad():
    pass
''')

    # When
    exit_code, output = run_pre_commit_hook([staged_file])

    # Then
    assert exit_code == 1  # Warning code
    assert "Invalid TAG format" in output
    assert "bad.py:2" in output  # File and line number
```

### Scenario 3.3: Non-existent SPEC-ID Warning

**Given** TAG references non-existent SPEC document
**When** developer commits changes
**Then** hook **shall** warn but allow commit (warn mode)

```python
def test_nonexistent_spec_warning():
    # Given
    staged_file = Path("orphan.py")
    staged_file.write_text('''
# @SPEC SPEC-DOES-NOT-EXIST
def orphan():
    pass
''')

    # When
    exit_code, output = run_pre_commit_hook([staged_file])

    # Then
    assert exit_code == 1  # Warning code
    assert "SPEC-DOES-NOT-EXIST not found" in output
    assert "Commit allowed" in output  # Warn mode
```

### Scenario 3.4: No Python Files Staged

**Given** only non-Python files are staged (e.g., .md, .json)
**When** developer commits changes
**Then** hook **shall** exit silently with success code (0)

```python
def test_no_python_files():
    # Given
    staged_files = [Path("README.md"), Path("config.json")]

    # When
    exit_code, output = run_pre_commit_hook(staged_files)

    # Then
    assert exit_code == 0
    assert output == ""  # Silent exit
```

### Edge Cases

**EC-3.1:** Mixed valid and invalid TAGs
```python
def test_mixed_tags():
    # Given - Some valid, some invalid
    staged_files = [
        create_file("valid.py", "# @SPEC SPEC-001\ndef f(): pass"),
        create_file("invalid.py", "# @spec bad\ndef g(): pass")
    ]

    # When
    exit_code, output = run_pre_commit_hook(staged_files)

    # Then - Report all errors
    assert exit_code == 1
    assert "valid.py" not in output  # Valid file not mentioned
    assert "invalid.py:1" in output  # Invalid file reported
```

**EC-3.2:** Hook execution timeout
```python
def test_hook_timeout():
    # Given - Very large file causing timeout
    staged_file = create_giant_file(100000)  # 100k lines

    # When
    exit_code, output = run_pre_commit_hook([staged_file], timeout=5)

    # Then - Graceful handling
    assert exit_code == 1
    assert "Timeout" in output or "skipped" in output
```

**EC-3.3:** Enforce mode (optional feature)
```python
def test_enforce_mode():
    # Given - Enforce mode configured
    set_config_mode("enforce")
    staged_file = create_file("bad.py", "# @spec bad\ndef f(): pass")

    # When
    exit_code, output = run_pre_commit_hook([staged_file])

    # Then - Block commit
    assert exit_code == 1
    assert "Commit blocked" in output
```

### Success Criteria

- [ ] Hook triggers on all Python file commits
- [ ] 100% of invalid TAG formats are detected
- [ ] Non-existent SPEC-IDs generate warnings
- [ ] Hook execution time <2 seconds for typical commits
- [ ] Hook respects warn/enforce/off modes

---

## T4: Linkage Manager (TAG↔CODE Mapping)

### Scenario 4.1: Add TAG to Linkage Database

**Given** TAG is extracted from source file
**When** LinkageManager.add_tag() is called
**Then** system **shall** atomically update linkage database

```python
def test_add_tag_to_linkage():
    # Given
    tag = TAG(spec_id="SPEC-AUTH-001", verb="impl", file_path=Path("auth.py"), line=5)
    manager = LinkageManager(db_path=Path("/tmp/linkage.json"))

    # When
    manager.add_tag(tag)

    # Then
    locations = manager.get_code_locations("SPEC-AUTH-001")
    assert len(locations) == 1
    assert locations[0]["file_path"] == "auth.py"
    assert locations[0]["line"] == 5
```

### Scenario 4.2: Query TAGs by SPEC-ID

**Given** multiple code locations reference same SPEC-ID
**When** LinkageManager.get_code_locations() is called
**Then** system **shall** return all locations

```python
def test_query_by_spec_id():
    # Given - Multiple locations for same SPEC
    manager = LinkageManager(db_path=Path("/tmp/linkage.json"))
    manager.add_tag(TAG("SPEC-001", "impl", Path("a.py"), 10))
    manager.add_tag(TAG("SPEC-001", "verify", Path("b.py"), 20))
    manager.add_tag(TAG("SPEC-001", "impl", Path("c.py"), 30))

    # When
    locations = manager.get_code_locations("SPEC-001")

    # Then
    assert len(locations) == 3
    assert any(loc["file_path"] == "a.py" and loc["line"] == 10 for loc in locations)
    assert any(loc["file_path"] == "b.py" and loc["line"] == 20 for loc in locations)
    assert any(loc["file_path"] == "c.py" and loc["line"] == 30 for loc in locations)
```

### Scenario 4.3: Remove TAGs for Deleted File

**Given** code file is deleted from repository
**When** LinkageManager.remove_file_tags() is called
**Then** system **shall** remove all TAG entries for that file

```python
def test_remove_file_tags():
    # Given
    manager = LinkageManager(db_path=Path("/tmp/linkage.json"))
    manager.add_tag(TAG("SPEC-001", "impl", Path("deleted.py"), 5))
    manager.add_tag(TAG("SPEC-002", "impl", Path("deleted.py"), 10))

    # When
    manager.remove_file_tags(Path("deleted.py"))

    # Then
    assert manager.get_code_locations("SPEC-001") == []
    assert manager.get_code_locations("SPEC-002") == []
```

### Edge Cases

**EC-4.1:** Atomic write failure recovery
```python
def test_atomic_write_recovery():
    # Given
    manager = LinkageManager(db_path=Path("/tmp/linkage.json"))
    tag = TAG("SPEC-001", "impl", Path("test.py"), 1)

    # When - Simulate write failure
    with patch("pathlib.Path.rename", side_effect=OSError("Disk full")):
        with pytest.raises(OSError):
            manager.add_tag(tag)

    # Then - Original database intact
    assert manager.db_path.exists()  # Original file not corrupted
    assert manager.get_code_locations("SPEC-001") == []  # No partial data
```

**EC-4.2:** Concurrent database access
```python
def test_concurrent_access():
    # Given
    manager1 = LinkageManager(Path("/tmp/linkage.json"))
    manager2 = LinkageManager(Path("/tmp/linkage.json"))

    # When - Concurrent writes
    def add_tag(manager, spec_id):
        manager.add_tag(TAG(spec_id, "impl", Path("test.py"), 1))

    with ThreadPoolExecutor(max_workers=2) as executor:
        futures = [
            executor.submit(add_tag, manager1, "SPEC-001"),
            executor.submit(add_tag, manager2, "SPEC-002")
        ]
        wait(futures)

    # Then - Both writes succeed (no corruption)
    manager = LinkageManager(Path("/tmp/linkage.json"))
    assert len(manager.get_code_locations("SPEC-001")) == 1
    assert len(manager.get_code_locations("SPEC-002")) == 1
```

**EC-4.3:** Orphaned TAG detection
```python
def test_orphaned_tag_detection():
    # Given - TAG referencing deleted SPEC
    manager = LinkageManager(Path("/tmp/linkage.json"))
    manager.add_tag(TAG("SPEC-DELETED", "impl", Path("orphan.py"), 1))

    # When - SPEC is deleted
    delete_spec_document("SPEC-DELETED")

    # Then - Detect orphaned TAGs
    orphans = manager.find_orphaned_tags()
    assert len(orphans) == 1
    assert orphans[0].spec_id == "SPEC-DELETED"
```

### Success Criteria

- [ ] Linkage database accurately tracks all TAGs
- [ ] Query performance <100ms for SPEC-ID lookups
- [ ] Atomic writes prevent database corruption
- [ ] Concurrent access handled safely
- [ ] Orphaned TAG detection works correctly

---

## T5: Quality Configuration Integration

### Scenario 5.1: Load TAG Configuration from quality.yaml

**Given** `.moai/config/sections/quality.yaml` contains TAG settings
```yaml
tag_validation:
  enabled: true
  mode: warn
```
**When** TAG system loads configuration
**Then** system **shall** apply settings correctly

```python
def test_load_tag_configuration():
    # Given
    config_yaml = create_quality_config('''
tag_validation:
  enabled: true
  mode: warn
''')

    # When
    config = load_tag_configuration(config_yaml)

    # Then
    assert config.enabled == True
    assert config.mode == "warn"
```

### Scenario 5.2: Disabled TAG Validation

**Given** configuration sets `tag_validation.enabled: false`
**When** pre-commit hook runs
**Then** system **shall** skip all TAG validation

```python
def test_disabled_validation():
    # Given
    set_config(enabled=False, mode="warn")
    staged_file = create_file("test.py", "# @spec bad\ndef f(): pass")

    # When
    exit_code, output = run_pre_commit_hook([staged_file])

    # Then - Skipped
    assert exit_code == 0  # Success even with invalid TAG
    assert "TAG validation disabled" in output
```

### Scenario 5.3: Mode Switching (warn/enforce/off)

**Given** configuration changes from `warn` to `enforce`
**When** pre-commit hook runs with invalid TAG
**Then** system **shall** block commit (enforce mode)

```python
def test_mode_switching():
    # Given
    staged_file = create_file("test.py", "# @spec bad\ndef f(): pass")

    # When - Warn mode
    set_config(enabled=True, mode="warn")
    exit_code_warn, output_warn = run_pre_commit_hook([staged_file])

    # Then - Warn allows commit
    assert exit_code_warn == 1  # Warning
    assert "Commit allowed" in output_warn

    # When - Enforce mode
    set_config(enabled=True, mode="enforce")
    exit_code_enforce, output_enforce = run_pre_commit_hook([staged_file])

    # Then - Enforce blocks commit
    assert exit_code_enforce == 1  # Error
    assert "Commit blocked" in output_enforce
```

### Edge Cases

**EC-5.1:** Invalid configuration value
```python
def test_invalid_configuration():
    # Given
    config_yaml = create_quality_config('''
tag_validation:
  enabled: true
  mode: invalid_mode  # Invalid mode
''')

    # When & Then
    with pytest.raises(ValueError, match="Invalid mode: invalid_mode"):
        load_tag_configuration(config_yaml)
```

**EC-5.2:** Missing configuration (use defaults)
```python
def test_missing_configuration():
    # Given - No TAG configuration in quality.yaml
    config_yaml = create_quality_config('''
# tag_validation section missing
''')

    # When
    config = load_tag_configuration(config_yaml)

    # Then - Use defaults
    assert config.enabled == True  # Default: enabled
    assert config.mode == "warn"  # Default: warn
```

**EC-5.3:** Configuration hot-reload
```python
def test_configuration_hot_reload():
    # Given
    set_config(enabled=True, mode="warn")
    manager = LinkageManager(Path("/tmp/linkage.json"))

    # When - Reload configuration
    set_config(enabled=True, mode="enforce")
    manager.reload_configuration()

    # Then - New settings applied immediately
    assert manager.config.mode == "enforce"
```

### Success Criteria

- [ ] Configuration loaded correctly from quality.yaml
- [ ] Default settings applied when configuration missing
- [ ] Invalid configuration values rejected with clear error
- [ ] Mode switching works dynamically (no restart required)
- [ ] Disabled validation skips all TAG checks

---

## Integration Test Scenarios

### IT-1: Complete TAG Workflow

**Given** developer adds TAG to new feature code
```python
# @SPEC SPEC-FEATURE-001
def new_feature():
    pass
```
**When** developer commits changes
**Then** complete workflow executes successfully:
1. Pre-commit hook validates TAG format
2. Linkage database updated with TAG location
3. SPEC existence verified (warning if not exists)
4. Commit allowed (warn mode)

```python
def test_complete_tag_workflow():
    # Given
    create_spec_document("SPEC-FEATURE-001")
    staged_file = create_file("feature.py", "# @SPEC SPEC-FEATURE-001\ndef f(): pass")

    # When
    exit_code, output = run_pre_commit_hook([staged_file])

    # Then
    assert exit_code == 0  # Success

    # Verify linkage database updated
    manager = LinkageManager(Path("/tmp/linkage.json"))
    locations = manager.get_code_locations("SPEC-FEATURE-001")
    assert len(locations) == 1
    assert locations[0]["file_path"] == "feature.py"
```

### IT-2: Multi-SPEC Traceability

**Given** feature implementation with multiple related TAGs
```python
# @SPEC SPEC-AUTH-001 impl
# @SPEC SPEC-AUTH-002 depends
# @SPEC SPEC-AUTH-003 verify
def complex_auth():
    pass
```
**When** TAG system processes file
**Then** all TAGs extracted and linked correctly

```python
def test_multi_spec_traceability():
    # Given
    source_code = '''
# @SPEC SPEC-AUTH-001 impl
# @SPEC SPEC-AUTH-002 depends
# @SPEC SPEC-AUTH-003 verify
def complex_auth():
    pass
'''
    # When
    tags = extract_tags(source_code)

    # Then
    assert len(tags) == 3
    assert tags[0].spec_id == "SPEC-AUTH-001" and tags[0].verb == "impl"
    assert tags[1].spec_id == "SPEC-AUTH-002" and tags[1].verb == "depends"
    assert tags[2].spec_id == "SPEC-AUTH-003" and tags[2].verb == "verify"
```

### IT-3: TDD Integration (RED-GREEN-REFACTOR)

**Given** SPEC-001 defines authentication feature
**When** developer follows TDD workflow:
1. **RED**: Write failing test with TAG `@SPEC SPEC-001 verify`
2. **GREEN**: Implement feature with TAG `@SPEC SPEC-001 impl`
3. **REFACTOR**: Refactor with TAGs preserved

**Then** TAG traceability maintained throughout TDD cycle

```python
def test_tdd_tag_integration():
    # Given - RED phase
    test_file = create_file("test_auth.py", "# @SPEC SPEC-001 verify\ndef test_login(): pass")

    # When - GREEN phase
    impl_file = create_file("auth.py", "# @SPEC SPEC-001 impl\ndef login(): pass")

    # Then - Both TAGs linked to SPEC-001
    manager = LinkageManager(Path("/tmp/linkage.json"))
    locations = manager.get_code_locations("SPEC-001")
    assert len(locations) == 2
    assert any(loc["file_path"] == "test_auth.py" and loc["verb"] == "verify" for loc in locations)
    assert any(loc["file_path"] == "auth.py" and loc["verb"] == "impl" for loc in locations)
```

---

## TDD Test Scenarios (RED-GREEN-REFACTOR)

### TDD-1: Parser Implementation (RED Phase)

**Given** TAG parser does not exist
**When** developer writes failing test
```python
def test_extract_nonexistent_tags():
    tags = extract_tags("# @SPEC SPEC-001")
    assert tags == []  # Fails initially
```
**Then** test fails (RED)

### TDD-2: Parser Implementation (GREEN Phase)

**Given** failing test for TAG extraction
**When** developer implements minimal parser code
```python
def extract_tags(source_code: str) -> List[TAG]:
    # Minimal implementation to pass test
    return []
```
**Then** test passes (GREEN)

### TDD-3: Parser Implementation (REFACTOR Phase)

**Given** working but unoptimized parser
**When** developer refactors for clarity and performance
**Then** all tests still pass and TAG extraction is efficient

---

## Performance Tests

### PT-1: Large Codebase TAG Extraction

**Given** codebase with 1000 Python files, average 500 TAGs
**When** TAG system scans entire codebase
**Then** extraction completes in <30 seconds

```python
def test_large_codebase_performance():
    # Given - 1000 files with varying TAG counts
    create_test_codebase(file_count=1000, avg_tags_per_file=0.5)

    # When
    start = time.time()
    all_tags = extract_tags_from_codebase("/tmp/test_codebase")
    elapsed = time.time() - start

    # Then
    assert len(all_tags) == 500  # ~500 TAGs total
    assert elapsed < 30.0  # 30 seconds for entire codebase
```

### PT-2: Linkage Database Query Performance

**Given** linkage database with 10,000 TAG entries
**When** querying code locations for SPEC-ID with 100 matches
**Then** query completes in <100ms

```python
def test_linkage_query_performance():
    # Given - 10,000 TAG entries
    manager = LinkageManager(Path("/tmp/linkage.json"))
    for i in range(10000):
        spec_id = f"SPEC-{i % 100:03d}"  # 100 unique SPECs
        manager.add_tag(TAG(spec_id, "impl", Path(f"file{i}.py"), 1))

    # When - Query SPEC with 100 matches
    start = time.time()
    locations = manager.get_code_locations("SPEC-001")
    elapsed = time.time() - start

    # Then
    assert len(locations) == 100
    assert elapsed < 0.1  # 100ms
```

---

## Security Tests

### ST-1: Path Traversal Prevention

**Given** malicious TAG with path traversal attempt
```python
# @SPEC ../../../../../etc/passwd
def malicious():
    pass
```
**When** validator checks TAG format
**Then** TAG is rejected as invalid SPEC-ID format

```python
def test_path_traversal_prevention():
    # Given
    tag = TAG(spec_id="../../../../etc/passwd", verb="impl", file_path=Path("test.py"), line=1)

    # When & Then
    assert not validate_tag_format(tag)
    assert "Invalid SPEC-ID format" in get_validation_error(tag)
```

### ST-2: Code Injection Prevention

**Given** TAG with code injection attempt
```python
# @SPEC SPEC-001; rm -rf /
def dangerous():
    pass
```
**When** parser extracts TAG
**Then** only SPEC-ID portion is extracted, command ignored

```python
def test_code_injection_prevention():
    # Given
    source_code = "# @SPEC SPEC-001; rm -rf /\ndef dangerous(): pass"

    # When
    tags = extract_tags(source_code)

    # Then - Only valid SPEC-ID extracted
    assert len(tags) == 1
    assert tags[0].spec_id == "SPEC-001"
    assert ";" not in tags[0].spec_id  # Command separator not included
```

---

## Final Acceptance Criteria

### Quality Gates (TRUST-5 Framework)

#### Test-first Pillar
- [ ] Overall test coverage 90%+ (unit), 85%+ (integration)
- [ ] All 5 requirement types (T1-T5) have comprehensive test coverage
- [ ] Edge case coverage 80%+
- [ ] Mutation testing enabled for critical modules

#### Readable Pillar
- [ ] All code passes ruff linter (zero warnings)
- [ ] Function average length <20 lines
- [ ] Clear naming conventions (TAG, LinkageManager, extract_tags)

#### Unified Pillar
- [ ] Consistent code formatting (black)
- [ ] Consistent import ordering (isort)
- [ ] Unified TAG annotation format across codebase

#### Secured Pillar
- [ ] Path traversal prevention tests pass
- [ ] Code injection prevention tests pass
- [ ] Atomic writes prevent database corruption

#### Trackable Pillar
- [ ] All TAGs traceable to SPEC documents
- [ ] Linkage database enables reverse lookup
- [ ] Commit history includes TAG validation results

### Definition of Done

- [ ] All T1-T5 requirements implemented and tested
- [ ] Acceptance criteria 100% satisfied
- [ ] Unit test coverage 90%+
- [ ] Integration test coverage 85%+
- [ ] Performance benchmarks met (<2s hook, <100ms query)
- [ ] Security tests pass (path traversal, code injection)
- [ ] Documentation complete (README, API docs, TAG guide)
- [ ] CLI tools functional (moai-tag command)
- [ ] Quality configuration integrated with quality.yaml

---

## Next Steps

```bash
# TDD Execution
/moai:2-run SPEC-TAG-001

# Documentation Sync
/moai:3-sync SPEC-TAG-001
```
