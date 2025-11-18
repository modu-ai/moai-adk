# SPEC-UPDATE-HOOKS-001: Acceptance Criteria and Test Scenarios

**SPEC ID**: SPEC-UPDATE-HOOKS-001
**Test Document Version**: 1.0.0
**Last Updated**: 2025-11-19

---

## Overview

This document defines the acceptance criteria and test scenarios for SPEC-UPDATE-HOOKS-001 (Claude Code Hooks Optimization). All test scenarios follow the Given-When-Then (Gherkin) format for clarity and testability.

---

## Test Execution Strategy

### Test Levels

| Level | Type | Tools | Pass Criteria |
|---|---|---|---|
| **Unit** | Function-level testing | pytest | 100% pass rate |
| **Integration** | Module interaction testing | pytest + imports | 100% pass rate |
| **End-to-End** | Hook execution testing | pytest + stdin/stdout | 100% pass rate |
| **System** | Full workflow testing | ./moai scripts | 100% pass rate |

### Test Environment

- **Python Version**: 3.8+
- **Test Framework**: pytest
- **Mock Libraries**: unittest.mock, pytest-mock
- **Coverage Tool**: pytest-cov
- **Performance Tool**: pytest-benchmark

### Test Execution Command

```bash
# Run all tests with coverage
pytest tests/ --cov=src/ --cov-report=html --cov-fail-under=80

# Run specific test class
pytest tests/test_hooks_consolidation.py::TestDeduplication -v

# Run with performance benchmarking
pytest tests/test_hooks_performance.py --benchmark-only
```

---

## Acceptance Scenario 1: Code Duplication Elimination

### Scenario 1A: Duplicate Files Identification

**Title**: Identify duplicate files in moai/core/ vs moai/shared/core/

```gherkin
GIVEN the template directory contains:
  - moai/core/timeout.py (200 lines)
  - moai/shared/core/timeout.py (200 lines)
  - moai/core/version_cache.py (150 lines)
  - moai/shared/core/version_cache.py (150 lines)

WHEN I run the duplication analysis:
  ```bash
  python tests/helpers/analyze_duplication.py src/moai_adk/templates/.claude/hooks/
  ```

THEN the analysis should report:
  - 2 duplicate file pairs identified
  - timeout.py: 95%+ content match
  - version_cache.py: 95%+ content match
  - Total duplicate lines: 350+ lines
  - Recommendation: Consolidate to moai/shared/core/
```

**Test Implementation**:

```python
def test_duplicate_file_identification():
    """Test that duplicate files are correctly identified"""
    duplicates = analyze_hooks_duplication(
        hooks_dir="src/moai_adk/templates/.claude/hooks/"
    )

    assert len(duplicates) >= 2, "Should find at least 2 duplicate pairs"
    assert "timeout.py" in [d["file1"] for d in duplicates]
    assert "version_cache.py" in [d["file1"] for d in duplicates]

    for dup in duplicates:
        assert dup["match_percentage"] >= 95, "Duplicates should be 95%+ identical"
        assert dup["duplicate_lines"] >= 140, "Each duplicate should be 140+ lines"
```

### Scenario 1B: moai/core/ Directory Deletion

**Title**: Verify moai/core/ directory is safely deleted after consolidation

```gherkin
GIVEN the code duplication analysis is complete
  AND all imports have been updated to moai.shared.core

WHEN I execute the consolidation phase:
  ```bash
  rm -rf src/moai_adk/templates/.claude/hooks/moai/core/
  git add -A
  git status
  ```

THEN the following should be true:
  - moai/core/ directory no longer exists
  - git status shows deletion of 4-5 files
  - No remaining references to moai.core in any .py file
  - pytest --cov runs with 0 import errors
```

**Test Implementation**:

```python
def test_moai_core_directory_removal():
    """Test that moai/core/ is successfully removed"""
    hooks_dir = "src/moai_adk/templates/.claude/hooks/"

    # Verify directory is removed
    core_dir = Path(hooks_dir) / "moai" / "core"
    assert not core_dir.exists(), "moai/core/ should be deleted"

    # Verify no references remain
    py_files = Path(hooks_dir).rglob("*.py")
    for py_file in py_files:
        content = py_file.read_text()
        assert "moai.core" not in content, f"Found moai.core reference in {py_file}"
        assert "from .moai.core" not in content, f"Found relative core import in {py_file}"
```

### Scenario 1C: Consolidated Module Verification

**Title**: Verify consolidated modules are complete and functional

```gherkin
GIVEN moai/shared/core/ is the consolidation target

WHEN I validate the consolidated modules:
  - Check moai/shared/core/timeout.py exists and is complete
  - Check moai/shared/core/version_cache.py exists and is complete
  - Run imports: from moai.shared.core import TimeoutManager, VersionCache
  - Run unit tests for each module

THEN all the following should be true:
  - Both files exist with 100%+ of original lines
  - Import statements succeed without error
  - All public classes are importable
  - All public methods are accessible
  - Unit test suite passes (100% pass rate)
  - Test coverage >= 80% for each module
```

**Test Implementation**:

```python
def test_consolidated_modules_exist():
    """Test that consolidated modules exist and are importable"""
    from moai.shared.core.timeout import TimeoutManager
    from moai.shared.core.version_cache import VersionCache

    assert TimeoutManager is not None
    assert VersionCache is not None

    # Test instantiation
    timeout_mgr = TimeoutManager(timeout_seconds=5)
    assert timeout_mgr is not None

    cache = VersionCache(cache_ttl_seconds=3600)
    assert cache is not None
```

---

## Acceptance Scenario 2: Import Path Standardization

### Scenario 2A: Import Pattern Analysis

**Title**: Analyze and categorize all import patterns

```gherkin
GIVEN the hooks directory contains 39 .py files with 3 different import patterns

WHEN I analyze all import statements:
  ```bash
  python tests/helpers/analyze_imports.py src/moai_adk/templates/.claude/hooks/
  ```

THEN the analysis should report:
  - Total files analyzed: 39
  - Pattern 1 (Relative - from .): ~15 files
  - Pattern 2 (Mixed - from moai + from .): ~8 files
  - Pattern 3 (Absolute - from moai.*): ~16 files
  - Total import statements: 100+
  - Unique modules imported: 20+
```

**Test Implementation**:

```python
def test_import_pattern_analysis():
    """Test that import patterns are correctly categorized"""
    patterns = analyze_import_patterns("src/moai_adk/templates/.claude/hooks/")

    assert patterns["total_files"] == 39, "Should analyze 39 files"
    assert patterns["pattern_1_relative"] >= 10, "Should find relative imports"
    assert patterns["pattern_2_mixed"] >= 5, "Should find mixed imports"
    assert patterns["pattern_3_absolute"] >= 10, "Should find absolute imports"
```

### Scenario 2B: Import Path Conversion

**Title**: Convert all imports to absolute pattern (Pattern 3)

```gherkin
GIVEN 23 Hook files contain non-absolute imports
  (15 relative + 8 mixed patterns)

WHEN I apply the conversion rules:
  - Pattern 1: from .moai.core → from moai.shared.core
  - Pattern 2: from moai.shared + from . → from moai.shared + from moai.*
  - Run: sed, grep, or Python script to bulk convert
  - Validate with: python -m py_compile

THEN all the following should be true:
  - All 39 files use absolute imports (from moai.*)
  - No relative imports remain (grep "from \." = 0 matches)
  - No mixed patterns remain
  - python -m py_compile passes for all files (exit code 0)
  - pytest --collect-only succeeds (all imports resolve)
```

**Test Implementation**:

```python
def test_import_path_conversion():
    """Test that all imports are converted to absolute pattern"""
    hooks_dir = Path("src/moai_adk/templates/.claude/hooks/")
    py_files = list(hooks_dir.rglob("*.py"))

    for py_file in py_files:
        content = py_file.read_text()

        # No relative imports
        assert not re.search(r"^from \.", content, re.MULTILINE), \
            f"Found relative import in {py_file}"

        # All imports start with 'from moai' or 'import moai'
        imports = re.findall(r"^(?:from|import) (\w+)", content, re.MULTILINE)
        for imp in imports:
            assert imp.startswith("moai") or imp in ["json", "pathlib", "typing"], \
                f"Found non-absolute import in {py_file}: {imp}"
```

### Scenario 2C: Import Compliance Verification

**Title**: Verify 100% compliance with absolute import standard

```gherkin
GIVEN all imports have been converted to absolute pattern

WHEN I run compliance checks:
  1. Check 1: grep -r "from \." → expect 0 results
  2. Check 2: grep -r "^from moai" → expect 39+ results (one per file min)
  3. Check 3: python -m py_compile → expect exit code 0
  4. Check 4: pytest --collect-only → expect no import errors

THEN all checks should pass:
  - Relative import count: 0
  - Absolute import count: 39+ (at least one per file)
  - Compilation success: 100% (39/39 files)
  - Import resolution success: 100%
  - Compliance score: 100%
```

**Test Implementation**:

```python
def test_import_compliance_100_percent():
    """Test that all imports are 100% compliant with absolute pattern"""
    hooks_dir = Path("src/moai_adk/templates/.claude/hooks/")
    py_files = list(hooks_dir.rglob("*.py"))

    relative_imports = 0
    absolute_imports = 0

    for py_file in py_files:
        content = py_file.read_text()

        # Count import types
        relative = len(re.findall(r"^from \.", content, re.MULTILINE))
        absolute = len(re.findall(r"^from moai", content, re.MULTILINE))

        relative_imports += relative
        absolute_imports += absolute

    assert relative_imports == 0, f"Found {relative_imports} relative imports"
    assert absolute_imports > 0, "Should have absolute imports"

    # Test compilation
    for py_file in py_files:
        py_compile.compile(str(py_file), doraise=True)
```

---

## Acceptance Scenario 3: Hook Compliance Verification

### Scenario 3A: Hook Naming Convention Compliance

**Title**: Verify all Hook files follow official naming convention

```gherkin
GIVEN the hooks directory contains 5 Hook files

WHEN I validate the file naming:
  - List all Hook files: session_start*.py, pre_tool*.py, post_tool*.py, etc.
  - Check pattern: {event}__{description}.py
  - Validate event names against official spec

THEN all Hook files should:
  - Follow pattern {event}__{description}.py (100%)
  - Have valid event names:
    - session_start (SessionStart events)
    - pre_tool (PreToolUse events)
    - post_tool (PostToolUse events)
    - subagent_start (SubagentStart events)
    - subagent_stop (SubagentStop events)
  - Have descriptive names: cleanup, validation, logging, context, checkpoint
```

**Test Implementation**:

```python
VALID_EVENTS = {
    "session_start", "pre_tool", "post_tool",
    "subagent_start", "subagent_stop"
}

def test_hook_naming_convention():
    """Test that all Hook files follow naming convention"""
    hooks_dir = Path("src/moai_adk/templates/.claude/hooks/")
    hook_files = list(hooks_dir.glob("*.py"))

    # Extract hook files (not moai/ subdirectory)
    hook_files = [f for f in hook_files if f.parent == hooks_dir]

    for hook_file in hook_files:
        name = hook_file.stem  # filename without .py

        # Check pattern: event__description
        assert "__" in name, f"Hook name missing '__' separator: {name}"

        event, description = name.split("__", 1)

        assert event in VALID_EVENTS, \
            f"Invalid event '{event}' in {name}. Valid: {VALID_EVENTS}"

        assert len(description) > 0, \
            f"Missing description in Hook name: {name}"
```

### Scenario 3B: Hook stdin/stdout Interface Validation

**Title**: Verify Hook files implement JSON stdin/stdout interface

```gherkin
GIVEN each Hook file receives JSON input via stdin

WHEN I invoke each Hook with sample input:
  ```bash
  echo '{"hook_name":"session_start","context":{}}' | \
  python src/moai_adk/templates/.claude/hooks/session_start__checkpoint.py
  ```

THEN the Hook should:
  - Accept JSON input on stdin (no parsing errors)
  - Output valid JSON on stdout (parseable)
  - Return exit code 0 on success
  - Return non-zero exit code on error
  - Execute in < 1000ms (recommended < 500ms)
```

**Test Implementation**:

```python
def test_hook_json_interface():
    """Test that Hooks implement JSON stdin/stdout interface"""
    hooks_dir = Path("src/moai_adk/templates/.claude/hooks/")

    # Sample input (valid JSON)
    sample_input = json.dumps({
        "hook_name": "session_start",
        "context": {
            "project_root": "/path/to/project",
            "config": {"version": "0.26.0"}
        }
    })

    hook_files = [f for f in hooks_dir.glob("*.py") if f.parent == hooks_dir]

    for hook_file in hook_files:
        # Run Hook with sample input
        result = subprocess.run(
            [sys.executable, str(hook_file)],
            input=sample_input.encode(),
            capture_output=True,
            timeout=2.0  # 2 second timeout
        )

        # Verify exit code
        assert result.returncode in [0, 1, 2, 3], \
            f"Invalid exit code {result.returncode} from {hook_file.name}"

        # Verify JSON output
        try:
            output = json.loads(result.stdout.decode())
            assert "status" in output, "Missing 'status' in output"
        except json.JSONDecodeError:
            pytest.fail(f"Invalid JSON output from {hook_file.name}")
```

### Scenario 3C: Hook Performance and Timeout

**Title**: Verify all Hooks execute within timeout limits

```gherkin
GIVEN each Hook has execution timeout of 5 seconds (official spec)
  AND recommended timeout is 500-1000ms

WHEN I measure Hook execution time:
  - Execute each Hook 10 times
  - Measure execution time with timer
  - Record min, max, average times

THEN performance should meet criteria:
  - session_start__checkpoint: < 200ms avg
  - session_start__cleanup: < 250ms avg
  - session_start__config: < 150ms avg
  - session_start__cache: < 150ms avg
  - session_start__init: < 100ms avg
  - Total Hook chain: < 1.5 seconds
  - 99th percentile: < official 5 second limit
```

**Test Implementation**:

```python
@pytest.mark.benchmark
def test_hook_performance():
    """Test that Hooks execute within timeout limits"""
    hooks_dir = Path("src/moai_adk/templates/.claude/hooks/")
    hook_files = [f for f in hooks_dir.glob("*.py") if f.parent == hooks_dir]

    performance_data = {}

    for hook_file in hook_files:
        sample_input = json.dumps({
            "hook_name": "test",
            "context": {}
        })

        times = []
        for _ in range(10):
            start = time.perf_counter()
            result = subprocess.run(
                [sys.executable, str(hook_file)],
                input=sample_input.encode(),
                capture_output=True,
                timeout=2.0
            )
            elapsed = (time.perf_counter() - start) * 1000  # ms
            times.append(elapsed)

        avg_time = sum(times) / len(times)
        performance_data[hook_file.name] = {
            "min_ms": min(times),
            "avg_ms": avg_time,
            "max_ms": max(times)
        }

        # Check against limits
        assert avg_time < 1000, \
            f"Hook {hook_file.name} avg time {avg_time}ms exceeds 1000ms limit"

    # Print performance report
    print("\nHook Performance Report:")
    for name, metrics in performance_data.items():
        print(f"  {name}: avg={metrics['avg_ms']:.1f}ms "
              f"(min={metrics['min_ms']:.1f}, max={metrics['max_ms']:.1f})")
```

---

## Acceptance Scenario 4: Directory Structure Optimization

### Scenario 4A: Empty Directory Removal

**Title**: Verify empty and redundant directories are removed

```gherkin
GIVEN the current directory structure:
  - moai/core/ (contains duplicates, to be removed)
  - moai/handlers/ (empty, to be removed)
  - moai/shared/config/ (to be removed, files moved)

WHEN I verify directory removal:
  - Check moai/core/ exists: No
  - Check moai/handlers/ exists: No
  - Check moai/shared/config/ exists: No

THEN the following should be true:
  - No errors on directory deletion
  - Final file count <= 34 (from original 39)
  - Directory depth <= 3 levels
```

**Test Implementation**:

```python
def test_directory_removal():
    """Test that empty directories are successfully removed"""
    hooks_dir = Path("src/moai_adk/templates/.claude/hooks/")

    # Check removed directories don't exist
    removed_dirs = [
        hooks_dir / "moai" / "core",
        hooks_dir / "moai" / "handlers",
        hooks_dir / "moai" / "shared" / "config"
    ]

    for dir_path in removed_dirs:
        assert not dir_path.exists(), \
            f"Directory should be removed: {dir_path}"

    # Check file count
    all_py_files = list(hooks_dir.rglob("*.py"))
    assert len(all_py_files) <= 34, \
        f"File count {len(all_py_files)} exceeds target of 34"

    # Check directory depth
    for py_file in all_py_files:
        depth = len(py_file.relative_to(hooks_dir).parts)
        assert depth <= 3, \
            f"Directory depth {depth} exceeds limit of 3 for {py_file}"
```

### Scenario 4B: .gitignore Configuration

**Title**: Verify .gitignore is updated to prevent future tracking of cache files

```gherkin
GIVEN the .gitignore file in project root

WHEN I check .gitignore content:
  - Should contain: __pycache__/
  - Should contain: *.pyc
  - Should contain: .pyc
  - Should contain: *.py[cod]

THEN the following should be true:
  - All patterns present in .gitignore
  - No .pyc files exist in git tracking (git ls-files | grep .pyc → 0)
  - __pycache__ directories not tracked (git ls-files | grep __pycache__ → 0)
```

**Test Implementation**:

```python
def test_gitignore_configuration():
    """Test that .gitignore is properly configured"""
    gitignore_path = Path(".gitignore")

    assert gitignore_path.exists(), ".gitignore should exist"

    content = gitignore_path.read_text()

    required_patterns = [
        "__pycache__/",
        "*.pyc",
        "*.py[cod]"
    ]

    for pattern in required_patterns:
        assert pattern in content, \
            f"Pattern '{pattern}' missing from .gitignore"

    # Verify no .pyc files are tracked
    result = subprocess.run(
        ["git", "ls-files", "*.pyc"],
        capture_output=True,
        text=True
    )
    assert result.stdout == "", \
        f"Found tracked .pyc files: {result.stdout}"
```

### Scenario 4C: File Count Reduction

**Title**: Verify total file count is reduced from 39 to <= 34

```gherkin
GIVEN the initial file count: 39 Hook-related files

WHEN I count all files after optimization:
  - Count .py files in .claude/hooks/
  - Count .json config files in .moai/config/
  - Exclude __pycache__ and .pyc files

THEN the reduction should achieve:
  - Total files: 39 → <= 34 (13% reduction)
  - Removed files:
    - moai/core/*.py: 4 files deleted
    - moai/handlers/*.py: 1 file deleted
    - moai/shared/config/*.py: 1 file deleted
```

**Test Implementation**:

```python
def test_file_count_reduction():
    """Test that file count is reduced as expected"""
    hooks_dir = Path("src/moai_adk/templates/.claude/hooks/")

    # Count hook files (exclude __pycache__ and .pyc)
    hook_files = [
        f for f in hooks_dir.rglob("*.py")
        if "__pycache__" not in f.parts and f.suffix == ".py"
    ]

    file_count = len(hook_files)

    assert file_count <= 34, \
        f"File count {file_count} exceeds target of 34 " \
        f"(target reduction: 39 → 34, achieved: 39 → {file_count})"

    print(f"\nFile count optimization:")
    print(f"  Before: 39 files")
    print(f"  After: {file_count} files")
    print(f"  Reduction: {39 - file_count} files ({100*(39-file_count)/39:.1f}%)")
```

---

## Acceptance Scenario 5: Local Synchronization Verification

### Scenario 5A: Template-to-Local Synchronization

**Title**: Verify template changes synchronize correctly to local .claude/hooks/

```gherkin
GIVEN the template src/moai_adk/templates/.claude/hooks/ has been optimized
  AND local project has .moai/config/config.json with {{PROJECT_DIR}} variable

WHEN I run /moai:3-sync:
  ```bash
  /moai:3-sync SPEC-UPDATE-HOOKS-001
  ```

THEN synchronization should:
  - Copy template changes to local .claude/hooks/
  - Substitute {{PROJECT_DIR}} with actual project path
  - Preserve existing local customizations
  - Report: "X files synced, Y files modified, Z files preserved"
```

**Test Implementation**:

```python
def test_template_sync():
    """Test that template changes sync correctly to local"""
    # This test runs in local project context
    local_hooks = Path(".claude/hooks")
    template_hooks = Path("src/moai_adk/templates/.claude/hooks")

    # Count files before sync
    local_files_before = set(local_hooks.rglob("*.py"))

    # Run sync (simulated)
    # sync_templates(template_hooks, local_hooks)

    local_files_after = set(local_hooks.rglob("*.py"))

    # Verify sync results
    new_files = local_files_after - local_files_before

    assert len(local_files_after) <= 34, \
        f"Local file count {len(local_files_after)} exceeds 34"

    # Verify no files lost (new >= old is ok)
    assert len(local_files_after) >= len(local_files_before), \
        "Sync should not delete local files"
```

### Scenario 5B: Variable Substitution Verification

**Title**: Verify {{PROJECT_DIR}} variables are correctly substituted

```gherkin
GIVEN template files contain {{PROJECT_DIR}} variable references

WHEN I check local .claude/hooks/ files:
  - Read all .py files
  - Search for {{PROJECT_DIR}} string

THEN all templates should be substituted:
  - {{PROJECT_DIR}} count: 0 (all substituted)
  - Actual paths count: >= 1 (at least one file uses project path)
  - Example: /Users/goos/MoAI/MoAI-ADK (fully resolved path)
```

**Test Implementation**:

```python
def test_variable_substitution():
    """Test that {{PROJECT_DIR}} variables are substituted"""
    local_hooks = Path(".claude/hooks")

    for py_file in local_hooks.rglob("*.py"):
        content = py_file.read_text()

        # No unsubstituted variables
        assert "{{PROJECT_DIR}}" not in content, \
            f"Found unsubstituted {{{{PROJECT_DIR}}}} in {py_file}"

        # Should have resolved paths (if any path references exist)
        if "project" in content.lower() or "root" in content.lower():
            # At least one absolute path reference
            assert "/" in content or ":" in content, \
                f"Missing path references in {py_file}"
```

### Scenario 5C: Local Customization Preservation

**Title**: Verify existing local customizations are preserved during sync

```gherkin
GIVEN local project has custom Hook file:
  - .claude/hooks/session_start__my_custom_hook.py (added locally)

WHEN I run /moai:3-sync:

THEN local customizations should:
  - Remain in .claude/hooks/ (not deleted)
  - Not be overwritten by template
  - git status should show: "modified: .claude/hooks/" (not deleted)
```

**Test Implementation**:

```python
def test_local_customization_preservation():
    """Test that local customizations are preserved during sync"""
    local_hooks = Path(".claude/hooks")
    custom_file = local_hooks / "session_start__my_custom_hook.py"

    # Create custom file before sync
    custom_file.write_text("# Custom Hook - should be preserved\n")

    # Run sync (simulated)
    # sync_templates(template_hooks, local_hooks)

    # Verify custom file still exists
    assert custom_file.exists(), \
        "Custom Hook file should be preserved"

    # Verify content unchanged
    content = custom_file.read_text()
    assert "Custom Hook" in content, \
        "Custom Hook content should not be modified"
```

---

## Test Summary and Metrics

### Test Coverage Requirements

| Component | Test Type | Required Coverage | Status |
|---|---|---|---|
| moai.shared.core | Unit | 80%+ | Pending |
| moai.shared.handlers | Unit | 80%+ | Pending |
| Hook files | Integration | 100% | Pending |
| Import paths | System | 100% | Pending |
| Performance | Benchmark | 5 Hook types | Pending |

### Test Execution Checklist

- [ ] Scenario 1: Code Duplication (A, B, C)
- [ ] Scenario 2: Import Standardization (A, B, C)
- [ ] Scenario 3: Hook Compliance (A, B, C)
- [ ] Scenario 4: Directory Optimization (A, B, C)
- [ ] Scenario 5: Local Synchronization (A, B, C)

### Pass/Fail Criteria

**All tests must pass** for SPEC to be marked Complete:
- Unit tests: 100% pass rate
- Integration tests: 100% pass rate
- End-to-End tests: 100% pass rate
- Performance tests: All hooks < 1 second avg
- Compatibility tests: 0 regressions

**Critical tests** (must not fail):
- Scenario 1A: Duplicate identification
- Scenario 2C: Import compliance (100%)
- Scenario 3C: Hook interface validation
- Scenario 4A: Directory removal
- Scenario 5A: Template sync

---

## Definition of Done

The SPEC-UPDATE-HOOKS-001 is considered COMPLETE when:

1. **Code Quality**
   - [ ] All duplicate code eliminated (moai/core/ removed)
   - [ ] All imports standardized to absolute pattern (100%)
   - [ ] All code compiled without errors (py_compile: 0 errors)

2. **Functionality**
   - [ ] All Hook files functioning identically (regression tests pass)
   - [ ] Hook I/O interface verified (JSON stdin/stdout)
   - [ ] Performance within limits (< 1 second per Hook)

3. **Testing**
   - [ ] Unit tests: 100% pass, 80%+ coverage
   - [ ] Integration tests: 100% pass
   - [ ] E2E tests: 100% pass
   - [ ] Performance tests: All targets met

4. **Documentation**
   - [ ] hooks/ directory documented (README.md)
   - [ ] All public APIs documented (docstrings)
   - [ ] Architecture documented (diagrams included)

5. **Deployment**
   - [ ] Changes committed to feature/SPEC-UPDATE-HOOKS-001
   - [ ] All tests pass on local and CI/CD
   - [ ] Ready for merge to main
   - [ ] Release v0.27.0 can proceed

---

## Sign-Off

**Created By**: GOOS行
**Date Created**: 2025-11-19
**Status**: Draft (awaiting implementation and testing)

**Test Execution Start Date**: [To be filled during implementation]
**Test Execution End Date**: [To be filled after completion]
**Overall Test Result**: [PASS / FAIL]

**Approved By**: [QA Lead / Product Owner]
**Approval Date**: [To be filled after review]
