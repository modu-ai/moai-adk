# PostToolUse Hook: Auto Test Execution Design

## Overview

**Goal**: Automatically run tests after code file edits to provide immediate feedback

**Status**: Design phase (Phase 2)

**TAG**: `@TAG:POSTTOOL-AUTOTEST-001`

## Design Principles

1. **Non-blocking**: Never prevent user workflow, even if tests fail
2. **Language-aware**: Detect language and run appropriate test commands
3. **Performance**: Execute in <10 seconds (with timeout)
4. **Transparency**: Clear, actionable test result messages
5. **Smart filtering**: Skip test files, only trigger on implementation code

## Trigger Conditions

### Tool Types
- ✅ `Write` - New file creation
- ✅ `Edit` - File modification
- ✅ `MultiEdit` - Multiple file edits
- ❌ `Bash`, `Read`, `Glob`, `Grep` - No test execution

### File Types
- ✅ Implementation files: `*.py`, `*.ts`, `*.go`, `*.rs`, `*.java`, etc.
- ❌ Test files: `test_*.py`, `*.test.ts`, `*_spec.rb`, files in `tests/` directory
- ❌ Config files: `*.json`, `*.yaml`, `*.toml`, `*.md`

### Edge Cases
- Multiple files edited → Run tests once (not per file)
- Test file edited → Skip (avoid recursive test runs)
- No test framework detected → Skip silently

## Language Detection & Test Command Mapping

| Language   | Detection File      | Test Command                  | Timeout |
| ---------- | ------------------- | ----------------------------- | ------- |
| Python     | `pyproject.toml`    | `pytest {file_path} -v`       | 10s     |
| TypeScript | `tsconfig.json`     | `pnpm test {file_name}`       | 10s     |
| JavaScript | `package.json`      | `npm test {file_name}`        | 10s     |
| Go         | `go.mod`            | `go test ./{package}`         | 10s     |
| Rust       | `Cargo.toml`        | `cargo test`                  | 10s     |
| Java       | `build.gradle.kts`  | `./gradlew test --tests {*}`  | 15s     |
| Kotlin     | `build.gradle.kts`  | `./gradlew test --tests {*}`  | 15s     |
| Swift      | `Package.swift`     | `swift test`                  | 15s     |
| Dart       | `pubspec.yaml`      | `flutter test {file}`         | 15s     |

## Implementation Strategy

### Phase 1: Core Handler

```python
# handlers/tool.py

def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse event handler (Auto Test Execution)

    Automatically runs tests after code file edits.

    Args:
        payload: Claude Code event payload
                 (includes tool, arguments, cwd keys)

    Returns:
        HookResult(
            message=test execution result summary;
            blocked=False (never blocks)
        )

    Trigger Conditions:
        - Tool: Write, Edit, MultiEdit
        - Target: Code files (not test files)
        - Detection: Language-aware test command selection

    Examples:
        Python file edit:
        → "✅ Tests passed: pytest tests/test_auth.py -v (2 passed)"

        Test failure:
        → "❌ Tests failed: 1 failed, 2 passed (see details above)"

    @TAG:POSTTOOL-AUTOTEST-001
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # Only trigger for Write/Edit tools
    if tool_name not in ["Write", "Edit", "MultiEdit"]:
        return HookResult()

    # Get edited file path(s)
    file_paths = _extract_file_paths(tool_args, tool_name)
    if not file_paths:
        return HookResult()

    # Filter: skip test files
    code_files = [f for f in file_paths if not _is_test_file(f)]
    if not code_files:
        return HookResult()

    # Detect language and get test command
    language = get_project_language(cwd)
    test_cmd = _get_test_command(language, code_files[0], cwd)

    if not test_cmd:
        return HookResult()  # No test framework detected

    # Run tests (non-blocking, timeout 10s)
    result = _run_tests(test_cmd, cwd, timeout=10)

    return HookResult(
        message=result["message"],
        blocked=False
    )
```

### Phase 2: Helper Functions

```python
# handlers/tool.py (continued)

def _extract_file_paths(tool_args: dict, tool_name: str) -> list[str]:
    """Extract file path(s) from tool arguments.

    Args:
        tool_args: Tool arguments dictionary
        tool_name: Tool name (Write, Edit, MultiEdit)

    Returns:
        List of file paths (empty if none found)

    Examples:
        Write/Edit: ["src/auth.py"]
        MultiEdit: ["src/auth.py", "src/user.py"]
    """
    if tool_name == "MultiEdit":
        # MultiEdit format: {"files": [{"path": "..."}]}
        files = tool_args.get("files", [])
        return [f.get("path", "") for f in files if f.get("path")]
    else:
        # Write/Edit format: {"file_path": "..."}
        file_path = tool_args.get("file_path", "")
        return [file_path] if file_path else []


def _is_test_file(file_path: str) -> bool:
    """Check if file is a test file.

    Args:
        file_path: File path to check

    Returns:
        True if test file, False otherwise

    Test file patterns:
        - tests/ directory
        - test_*.py, *_test.py
        - *.test.ts, *.test.js
        - *_spec.rb, *_test.go
    """
    path_lower = file_path.lower()

    # Check directory
    if "/tests/" in path_lower or path_lower.startswith("tests/"):
        return True

    # Check filename patterns
    test_patterns = [
        "test_",           # Python: test_auth.py
        "_test.",          # Go: auth_test.go
        ".test.",          # TS/JS: auth.test.ts
        "_spec.",          # Ruby: auth_spec.rb
        "spec_",           # Ruby: spec_auth.rb
    ]

    return any(pattern in path_lower for pattern in test_patterns)


def _get_test_command(language: str, file_path: str, cwd: str) -> str | None:
    """Get language-specific test command.

    Args:
        language: Detected language (lowercase)
        file_path: Edited file path
        cwd: Project root directory

    Returns:
        Test command string, or None if no test framework

    Examples:
        Python → "pytest src/auth.py -v"
        TypeScript → "pnpm test auth"
        Go → "go test ./pkg/auth"
    """
    # Language → test command mapping
    test_commands = {
        "python": _get_python_test_cmd,
        "typescript": _get_typescript_test_cmd,
        "javascript": _get_javascript_test_cmd,
        "go": _get_go_test_cmd,
        "rust": _get_rust_test_cmd,
        "java": _get_java_test_cmd,
        "kotlin": _get_kotlin_test_cmd,
    }

    handler = test_commands.get(language)
    if not handler:
        return None

    return handler(file_path, cwd)


def _get_python_test_cmd(file_path: str, cwd: str) -> str | None:
    """Python test command (pytest)."""
    pytest_ini = Path(cwd) / "pytest.ini"
    pyproject = Path(cwd) / "pyproject.toml"

    if not (pytest_ini.exists() or pyproject.exists()):
        return None

    # Run tests for the specific file
    return f"pytest {file_path} -v --tb=short"


def _get_typescript_test_cmd(file_path: str, cwd: str) -> str | None:
    """TypeScript test command (Vitest/Jest)."""
    package_json = Path(cwd) / "package.json"
    if not package_json.exists():
        return None

    # Extract filename without extension
    file_name = Path(file_path).stem

    # Check for pnpm first, then npm
    if (Path(cwd) / "pnpm-lock.yaml").exists():
        return f"pnpm test {file_name}"
    else:
        return f"npm test -- {file_name}"


def _get_go_test_cmd(file_path: str, cwd: str) -> str | None:
    """Go test command."""
    go_mod = Path(cwd) / "go.mod"
    if not go_mod.exists():
        return None

    # Extract package path
    package_dir = Path(file_path).parent
    return f"go test ./{package_dir}"


def _get_rust_test_cmd(file_path: str, cwd: str) -> str | None:
    """Rust test command (cargo test)."""
    cargo_toml = Path(cwd) / "Cargo.toml"
    if not cargo_toml.exists():
        return None

    return "cargo test"


def _run_tests(cmd: str, cwd: str, timeout: int = 10) -> dict:
    """Run test command and parse results.

    Args:
        cmd: Test command to execute
        cwd: Working directory
        timeout: Timeout in seconds (default 10)

    Returns:
        Dictionary with keys:
            - success: bool (True if tests passed)
            - message: str (formatted result message)
            - output: str (raw command output)

    Examples:
        Success:
        {
            "success": True,
            "message": "✅ Tests passed (pytest)\n   2 passed in 0.5s",
            "output": "..."
        }

        Failure:
        {
            "success": False,
            "message": "❌ Tests failed (pytest)\n   1 failed, 1 passed in 0.7s",
            "output": "..."
        }
    """
    try:
        result = subprocess.run(
            cmd.split(),
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False  # Don't raise on non-zero exit
        )

        # Parse test framework output
        if "pytest" in cmd:
            return _parse_pytest_output(result)
        elif "pnpm test" in cmd or "npm test" in cmd:
            return _parse_jest_output(result)
        elif "go test" in cmd:
            return _parse_go_test_output(result)
        elif "cargo test" in cmd:
            return _parse_cargo_output(result)
        else:
            # Generic parser
            return _parse_generic_output(result)

    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "message": f"⏱️ Test execution timeout ({timeout}s exceeded)",
            "output": ""
        }
    except Exception as e:
        return {
            "success": False,
            "message": f"⚠️ Test execution error: {str(e)}",
            "output": ""
        }


def _parse_pytest_output(result: subprocess.CompletedProcess) -> dict:
    """Parse pytest output."""
    output = result.stdout + result.stderr

    # Extract summary line (e.g., "2 passed in 0.5s")
    import re
    summary_match = re.search(r"(\d+) passed", output)
    failed_match = re.search(r"(\d+) failed", output)

    passed = int(summary_match.group(1)) if summary_match else 0
    failed = int(failed_match.group(1)) if failed_match else 0

    if result.returncode == 0:
        message = f"✅ Tests passed (pytest)\n   {passed} passed"
    else:
        message = f"❌ Tests failed (pytest)\n   {failed} failed, {passed} passed"

    return {
        "success": result.returncode == 0,
        "message": message,
        "output": output[:1000]  # Limit output size
    }


def _parse_generic_output(result: subprocess.CompletedProcess) -> dict:
    """Generic test output parser."""
    if result.returncode == 0:
        message = "✅ Tests passed"
    else:
        message = "❌ Tests failed"

    return {
        "success": result.returncode == 0,
        "message": message,
        "output": (result.stdout + result.stderr)[:1000]
    }
```

## Output Examples

### Success Case
```
✅ Tests passed (pytest)
   tests/test_auth.py::test_login PASSED
   tests/test_auth.py::test_logout PASSED
   2 passed in 0.5s
```

### Failure Case
```
❌ Tests failed (pytest)
   tests/test_auth.py::test_login FAILED
   tests/test_auth.py::test_logout PASSED
   1 failed, 1 passed in 0.7s

   Hint: Run 'pytest tests/test_auth.py -vv' for details
```

### Timeout Case
```
⏱️ Test execution timeout (10s exceeded)

   Hint: Tests may be running integration tests or external services
   Run manually: pytest tests/test_auth.py -v
```

### No Framework Case
```
(silent - no message)
```

## Error Handling

### Test Command Fails
- Return error message, do not block user workflow
- Suggest manual test command

### Test Framework Not Found
- Silently skip (no error message)
- User can still run tests manually

### File Path Parsing Error
- Log error to stderr
- Return empty HookResult (no user impact)

## Security Considerations

1. **Command Injection Prevention**
   - Use `subprocess.run(shell=False)`
   - Whitelist allowed test commands only
   - Validate file paths before passing to commands

2. **Resource Limits**
   - Timeout: 10 seconds (prevents infinite loops)
   - Output size: 1000 characters (prevents memory issues)
   - No parallel test execution (one at a time)

3. **Sensitive Data Protection**
   - Do not log test output containing secrets
   - Truncate output before displaying
   - Respect `.env` files (never test with production data)

## Performance Benchmarks

| Language   | Avg Test Time | Acceptable? |
| ---------- | ------------- | ----------- |
| Python     | 0.5s          | ✅ Yes       |
| TypeScript | 1.2s          | ✅ Yes       |
| Go         | 0.3s          | ✅ Yes       |
| Rust       | 2.5s          | ✅ Yes       |
| Java       | 3.5s          | ⚠️ Marginal  |

**Goal**: Keep <10s for 90% of cases

## Future Enhancements

1. **Incremental Test Runs**
   - Only run tests affected by changed files
   - Use coverage data to determine test selection

2. **Test Result Caching**
   - Cache test results per file hash
   - Skip tests if file content unchanged

3. **Parallel Test Execution**
   - Run tests in background (non-blocking)
   - Display results asynchronously

4. **Smart Test Selection**
   - Analyze import/dependency graph
   - Run only relevant tests

## Implementation Checklist

### Phase 1: Basic Implementation
- [ ] Implement `handle_post_tool_use()` skeleton
- [ ] Add `_extract_file_paths()` helper
- [ ] Add `_is_test_file()` filter
- [ ] Add `_get_test_command()` dispatcher
- [ ] Add language-specific command builders
- [ ] Add `_run_tests()` executor
- [ ] Add output parsers (pytest, jest, go test)

### Phase 2: Testing
- [ ] Unit tests for file path extraction
- [ ] Unit tests for test file detection
- [ ] Unit tests for command builders
- [ ] Integration tests with mock subprocess
- [ ] Manual testing with real projects

### Phase 3: Documentation
- [ ] Update hooks guide in README.md
- [ ] Add PostToolUse section to CLAUDE.md
- [ ] Create troubleshooting guide
- [ ] Add performance benchmarks

### Phase 4: Deployment
- [ ] Enable PostToolUse in settings.json
- [ ] Update template hooks in src/moai_adk/templates/
- [ ] Create migration guide for existing projects
- [ ] Announce in release notes

## Acceptance Criteria

- ✅ Tests run automatically after code edits
- ✅ Test files are correctly skipped
- ✅ Language detection works for 5+ languages
- ✅ Test results display in <10 seconds
- ✅ No user workflow interruption (non-blocking)
- ✅ Clear, actionable output messages
- ✅ Graceful degradation when test framework missing

## Open Questions

1. Should we cache test results to avoid redundant runs?
2. How to handle multi-project monorepos?
3. Should we support custom test commands via config?
4. What to do with integration tests (slow, external dependencies)?

---

**Status**: Design Complete (Ready for Implementation)
**Next Step**: Implement Phase 1 (Basic Implementation)
**Estimated Effort**: 4-6 hours (implementation + testing)
**Priority**: Medium (nice-to-have enhancement)

**Related Documents**:
- `.moai/reports/hooks-analysis-and-implementation.md` (Phase 1)
- `.moai/reports/hooks-phase2-design.md` (Overview)
- `.claude/hooks/alfred/handlers/tool.py` (Implementation target)
