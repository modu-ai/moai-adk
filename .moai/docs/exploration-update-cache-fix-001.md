# Codebase Exploration Report: SPEC-UPDATE-CACHE-FIX-001

## Executive Summary

This report documents the existing code structure, patterns, and infrastructure relevant to implementing the UV tool upgrade cache refresh auto-retry feature (SPEC-UPDATE-CACHE-FIX-001). The exploration covers implementation patterns, test infrastructure, and integration points.

---

## 1. Project Configuration

### Python & Dependencies
- **Python Version**: 3.13+ (from `pyproject.toml`)
- **Test Framework**: pytest 8.4.2+ with coverage
- **Key Dependencies**:
  - `click>=8.1.0` (CLI framework)
  - `rich>=13.0.0` (console output formatting)
  - `packaging>=21.0` (version parsing - already used)
  - `gitpython>=3.1.45`
  - `pyyaml>=6.0`

### Test Configuration
```ini
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = "test_*.py"
python_classes = "Test*"
python_functions = "test_*"
addopts = "-v --cov=src/moai_adk --cov-report=html --cov-report=term-missing"
```

**Coverage Requirement**: 85% minimum

---

## 2. Current `update.py` Implementation Analysis

### Location
- **File**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/cli/commands/update.py`
- **Lines**: 860 (comprehensive implementation)
- **Status**: Production-ready (v0.9.0)

### Key Existing Functions

#### Version Management Functions

**`_get_current_version()` (lines 189-198)**
```python
def _get_current_version() -> str:
    """Get currently installed moai-adk version.
    
    Returns:
        Version string (e.g., "0.6.1")
    
    Raises:
        RuntimeError: If version cannot be determined
    """
    return __version__
```
- **Purpose**: Returns the installed package version from `moai_adk.__version__`
- **Error Handling**: Implicit (no exception handling needed)
- **Integration Point**: Called before latest version check

**`_get_latest_version()` (lines 201-219)**
```python
def _get_latest_version() -> str:
    """Fetch latest moai-adk version from PyPI.
    
    Returns:
        Version string (e.g., "0.6.2")
    
    Raises:
        RuntimeError: If PyPI API unavailable or parsing fails
    """
    try:
        import urllib.error
        import urllib.request

        url = "https://pypi.org/pypi/moai-adk/json"
        with urllib.request.urlopen(url, timeout=5) as response:  # nosec B310
            data = json.loads(response.read().decode("utf-8"))
            return cast(str, data["info"]["version"])
    except (urllib.error.URLError, json.JSONDecodeError, KeyError, TimeoutError) as e:
        raise RuntimeError(f"Failed to fetch latest version from PyPI: {e}") from e
```
- **Pattern**: Uses `urllib` (standard library)
- **Timeout**: 5 seconds hardcoded
- **Error Handling**: Comprehensive with multiple exception types caught
- **Nosec Comment**: Security review already completed

**`_compare_versions(current: str, latest: str) -> int` (lines 222-242)**
```python
def _compare_versions(current: str, latest: str) -> int:
    """Compare semantic versions.
    
    Returns:
        -1 if current < latest (upgrade needed)
        0 if current == latest (up to date)
        1 if current > latest (unusual, already newer)
    """
    current_v = version.parse(current)
    latest_v = version.parse(latest)

    if current_v < latest_v:
        return -1
    elif current_v == latest_v:
        return 0
    else:
        return 1
```
- **Library**: Uses `packaging.version.parse()` (already in dependencies)
- **Returns**: Integer for easy comparison
- **Error Handling**: Can raise `InvalidVersion` - handled in caller at line 805

#### Installer Detection Functions

**`_detect_tool_installer() -> list[str] | None` (lines 149-186)**
```python
def _detect_tool_installer() -> list[str] | None:
    """Detect which tool installed moai-adk.
    
    Checks in priority order:
    1. uv tool (most likely for MoAI-ADK users)
    2. pipx
    3. pip (fallback)
    
    Returns:
        Command list [tool, ...args] ready for subprocess.run()
        or None if detection fails
    """
    if _is_installed_via_uv_tool():
        return UV_TOOL_COMMAND  # ["uv", "tool", "upgrade", "moai-adk"]
    elif _is_installed_via_pipx():
        return PIPX_COMMAND  # ["pipx", "upgrade", "moai-adk"]
    elif _is_installed_via_pip():
        return PIP_COMMAND  # ["pip", "install", "--upgrade", "moai-adk"]
    else:
        return None
```

**Detection Helper Functions** (lines 91-146)
- `_is_installed_via_uv_tool()`: Runs `uv tool list`
- `_is_installed_via_pipx()`: Runs `pipx list`
- `_is_installed_via_pip()`: Runs `pip show moai-adk`

**Timeout Constants**:
```python
TOOL_DETECTION_TIMEOUT = 5  # seconds
TOOL_DETECTION_TIMEOUT = 60  # for upgrade itself (line 323)
```

#### Subprocess Execution Pattern

**`_execute_upgrade(installer_cmd: list[str]) -> bool` (lines 305-330)**
```python
def _execute_upgrade(installer_cmd: list[str]) -> bool:
    """Execute package upgrade using detected installer.
    
    Args:
        installer_cmd: Command list from _detect_tool_installer()
                      e.g., ["uv", "tool", "upgrade", "moai-adk"]
    
    Returns:
        True if upgrade succeeded, False otherwise
    
    Raises:
        subprocess.TimeoutExpired: If upgrade times out
    """
    try:
        result = subprocess.run(
            installer_cmd,
            capture_output=True,
            text=True,
            timeout=60,
            check=False
        )
        return result.returncode == 0
    except subprocess.TimeoutExpired:
        raise  # Re-raise timeout for caller to handle
    except Exception:
        return False
```

**Key Pattern Elements**:
- `capture_output=True`: Captures stdout and stderr
- `text=True`: Returns strings instead of bytes
- `timeout=60`: 60-second timeout for package upgrade
- `check=False`: Does not raise exception on non-zero return code
- **Exception Handling**: 
  - `TimeoutExpired` is re-raised
  - Other exceptions return False

### Constants

```python
TOOL_DETECTION_TIMEOUT = 5  # seconds
UV_TOOL_COMMAND = ["uv", "tool", "upgrade", "moai-adk"]
PIPX_COMMAND = ["pipx", "upgrade", "moai-adk"]
PIP_COMMAND = ["pip", "install", "--upgrade", "moai-adk"]
```

### Console Output Patterns

The code uses `rich.console.Console` for formatted output:

```python
from rich.console import Console
console = Console()

# Examples from update.py:
console.print("[cyan]🔍 Checking versions...[/cyan]")
console.print(f"   Current version: {current}")
console.print(f"   Latest version:  {latest}")

console.print("[yellow]⚠️  Cannot reach PyPI[/yellow]")
console.print("[red]❌ Cannot detect package installer[/red]")
console.print("[green]✓ Upgrade complete![/green]")
console.print("[cyan]📢 Run 'moai-adk update' again[/cyan]")
```

**Color Codes Used**:
- `[red]` - Error messages
- `[yellow]` - Warnings
- `[cyan]` - Info/actions
- `[green]` - Success messages

---

## 3. Test Infrastructure & Patterns

### Test File Structure

**Main Test Files** (from `tests/unit/`):
1. `test_update.py` - Primary integration tests (833 lines)
2. `test_update_error_handling.py` - Error scenario tests (364 lines)
3. `test_update_workflow.py` - Workflow tests (393 lines)
4. `test_update_options.py` - CLI option tests
5. `test_update_tool_detection.py` - Tool detection tests
6. `test_version_cache.py` - Version caching tests (219 lines)
7. `test_version_check_config.py` - Version check configuration tests

### Mock Patterns

**CliRunner Usage** (from Click testing):
```python
from click.testing import CliRunner

runner = CliRunner()
with runner.isolated_filesystem(temp_dir=tmp_path):
    # Create .moai directory
    Path(".moai").mkdir()
    
    result = runner.invoke(update, ["--check"])
    assert result.exit_code == 0
    assert "Checking versions" in result.output
```

**Patch Patterns** (from unittest.mock):
```python
# Single patch
with patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest:
    mock_latest.return_value = "0.6.2"
    # test code

# Multiple patches (chained)
with patch("moai_adk.cli.commands.update._get_current_version") as mock_current, \
     patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest, \
     patch("moai_adk.cli.commands.update._execute_upgrade") as mock_upgrade:
    # test code
```

**Mock Object Patterns**:
```python
from unittest.mock import Mock

mock_instance = Mock()
mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup"
mock_instance.copy_templates.return_value = None
mock_processor.return_value = mock_instance

# Verify calls
mock_instance.create_backup.assert_called_once()
mock_instance.create_backup.assert_not_called()
```

### Test Naming Conventions

**Pattern**: `test_<feature>_<scenario>`

Examples:
- `test_update_help` - Basic functionality
- `test_update_not_initialized` - Error condition
- `test_update_check_only` - CLI option
- `test_upgrade_failure_shows_error` - Error handling
- `test_subprocess_timeout_handling` - Timeout scenario

### Parametrization Pattern

From existing tests:
```python
@pytest.mark.parametrize("version,expected", [
    ("0.6.1", -1),
    ("0.6.2", 0),
    ("0.6.3", 1),
])
def test_compare_versions(version, expected):
    result = _compare_versions("0.6.2", version)
    assert result == expected
```

### Subprocess Mocking Pattern

```python
import subprocess

with patch("subprocess.run") as mock_run:
    mock_run.return_value = MagicMock(returncode=0, stdout="Success")
    # test code
    mock_run.assert_called_once_with(
        ["command"],
        capture_output=True,
        text=True,
        timeout=60,
        check=False
    )
```

### Timeout Exception Testing

```python
with patch("subprocess.run") as mock_run:
    mock_run.side_effect = subprocess.TimeoutExpired(cmd="test", timeout=60)
    
    with pytest.raises(subprocess.TimeoutExpired):
        _execute_upgrade(["uv", "tool", "upgrade", "moai-adk"])
```

---

## 4. Existing Error Handling Mechanisms

### Custom Exception Hierarchy

```python
class UpdateError(Exception):
    """Base exception for update operations."""
    pass

class InstallerNotFoundError(UpdateError):
    """Raised when no package installer detected."""
    pass

class NetworkError(UpdateError):
    """Raised when network operation fails."""
    pass

class UpgradeError(UpdateError):
    """Raised when package upgrade fails."""
    pass

class TemplateSyncError(UpdateError):
    """Raised when template sync fails."""
    pass
```

**Location**: Lines 65-88 in `update.py`

### Error Handling in Main Function

The `update()` command handler (lines 684-859) shows:

1. **Try-except wrapper** around entire workflow (lines 684-859)
2. **Graceful fallback** for version check failures (lines 694-705)
3. **Specific exception handlers** for different stages:
   ```python
   except InstallerNotFoundError:
       _show_installer_not_found_help()
       raise click.Abort()
   
   except subprocess.TimeoutExpired:
       _show_timeout_error_help()
       raise click.Abort()
   
   except TemplateSyncError:
       _show_template_sync_failure_help()
       raise click.Abort()
   ```

### User-Facing Error Messages

Helper functions provide formatted error messages:
- `_show_installer_not_found_help()` (lines 578-589)
- `_show_upgrade_failure_help()` (lines 592-603)
- `_show_network_error_help()` (lines 606-612)
- `_show_template_sync_failure_help()` (lines 615-621)
- `_show_timeout_error_help()` (lines 624-628)

---

## 5. Integration Points for SPEC-UPDATE-CACHE-FIX-001

### Where to Add New Functions

**Option 1: Add near existing upgrade functions** (Recommended)
- Location: After `_execute_upgrade()` (line 330)
- New functions:
  1. `_detect_stale_cache()` - Lines 332-355 (estimated)
  2. `_clear_uv_package_cache()` - Lines 357-380 (estimated)

**Option 2: Create wrapper for execute_upgrade**
- Replace direct call to `_execute_upgrade()` with a new wrapper
- Wrapper `_execute_upgrade_with_retry()` handles cache refresh logic

### Integration with Main Update Function

**Current upgrade flow** (lines 746-784):
```python
if comparison < 0:  # Package upgrade needed
    # Confirm upgrade
    if not yes:
        if not click.confirm(...):
            return
    
    # Detect installer
    installer_cmd = _detect_tool_installer()
    if not installer_cmd:
        _show_installer_not_found_help()
        raise click.Abort()
    
    # Execute upgrade
    upgrade_result = _execute_upgrade(installer_cmd)  # <-- HERE
    if not upgrade_result:
        _show_upgrade_failure_help(installer_cmd)
        raise click.Abort()
```

**Integration point for cache-aware retry**:
- Replace the `_execute_upgrade(installer_cmd)` call
- Or wrap it with cache detection/retry logic

### Required Imports (Already Present)

```python
import json
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Any, cast

import click
from packaging import version
from rich.console import Console
```

**Additional imports needed for cache fix**:
- Already imported: `subprocess`, `version`, `console`
- May need: logging (optional, for debug info)

---

## 6. Existing Version Cache System

### Version Cache Infrastructure

From `test_version_cache.py`, a `VersionCache` class already exists:

**Location**: `src/moai_adk/templates/.claude/hooks/alfred/core/version_cache.py`

**Features**:
- TTL-based caching (24 hours default)
- Cache file location: `.moai/cache/version-check.json`
- Methods: `save()`, `load()`, `is_valid()`, `clear()`, `get_age_hours()`

**Cache Data Structure**:
```python
{
    "last_check": "2025-10-30T10:30:00",
    "current_version": "0.8.1",
    "latest_version": "0.9.0",
    "update_available": True,
    "upgrade_command": "uv tool upgrade moai-adk",
    "release_notes_url": "https://github.com/modu-ai/moai-adk/releases/tag/v0.9.0",
    "is_major_update": False
}
```

**Note**: This is used by SessionStart hook, not the `update` command itself.

---

## 7. Implementation Guidelines for New Functions

### Pattern for _detect_stale_cache()

Should follow this structure:

```python
def _detect_stale_cache(
    upgrade_output: str,
    current_version: str,
    latest_version: str
) -> bool:
    """Detect if uv cache is stale by comparing versions.
    
    @CODE:UPDATE-CACHE-FIX-001-001
    
    Args:
        upgrade_output: Output from subprocess.run
        current_version: Currently installed version
        latest_version: Latest available version
    
    Returns:
        True if cache appears stale, False otherwise
    
    Note:
        Returns False gracefully on version parsing errors
    """
    try:
        # Check if "Nothing to upgrade" in output
        if "Nothing to upgrade" not in upgrade_output:
            return False
        
        # Check if actual newer version exists
        if _compare_versions(current_version, latest_version) < 0:
            return True
        
        return False
    except Exception:
        # Graceful degradation: treat as not stale
        return False
```

### Pattern for _clear_uv_package_cache()

Should handle subprocess error cases:

```python
def _clear_uv_package_cache(package_name: str = "moai-adk") -> bool:
    """Clear uv cache for specific package.
    
    @CODE:UPDATE-CACHE-FIX-001-002
    
    Args:
        package_name: Package to clear cache for
    
    Returns:
        True if cache clear succeeded, False otherwise
    
    Note:
        Does not raise exceptions; returns False on any error
    """
    try:
        result = subprocess.run(
            ["uv", "cache", "clean", package_name],
            capture_output=True,
            text=True,
            timeout=10,  # Cache clean should be fast
            check=False
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return False
```

### Test Pattern Example

```python
class TestDetectStaleCache:
    """Test _detect_stale_cache function
    
    @TEST:UPDATE-CACHE-FIX-001-001
    """
    
    def test_detect_stale_cache_true_when_nothing_to_upgrade_and_newer_exists(self):
        """Detects stale cache when Nothing to upgrade but newer version exists"""
        output = "Nothing to upgrade"
        current = "0.8.3"
        latest = "0.9.0"
        
        result = _detect_stale_cache(output, current, latest)
        
        assert result is True
    
    def test_detect_stale_cache_false_when_versions_equal(self):
        """Does not detect stale cache when versions are equal"""
        output = "Nothing to upgrade"
        current = "0.9.0"
        latest = "0.9.0"
        
        result = _detect_stale_cache(output, current, latest)
        
        assert result is False
    
    def test_detect_stale_cache_graceful_on_parse_error(self):
        """Returns False gracefully on version parsing error"""
        output = "Nothing to upgrade"
        current = "{{INVALID}}"
        latest = "0.9.0"
        
        result = _detect_stale_cache(output, current, latest)
        
        assert result is False
```

---

## 8. Subprocess Usage Patterns in Codebase

### Standard Pattern from update.py

All subprocess calls follow this pattern:

```python
result = subprocess.run(
    command_list,              # List of strings, NOT shell=True
    capture_output=True,       # Get stdout/stderr
    text=True,                 # Return strings, not bytes
    timeout=seconds,           # Prevent hanging
    check=False                # Don't raise on non-zero exit
)

# Check result
if result.returncode == 0:
    process_output(result.stdout)
else:
    handle_error(result.stderr)
```

### Subprocess Calls in Tags Module

From `src/moai_adk/core/tags/generator.py`:

```python
import subprocess

# Using ripgrep for performance
def find_existing_tags(pattern: str) -> List[str]:
    try:
        result = subprocess.run(
            ["rg", "--no-heading", pattern],
            capture_output=True,
            text=True,
            timeout=5
        )
        if result.returncode == 0:
            return result.stdout.strip().split('\n')
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass
    return []
```

---

## 9. Key Files for Implementation Reference

### Core Implementation File
- **`src/moai_adk/cli/commands/update.py`** (860 lines)
  - Contains all functions needing modification
  - Already has subprocess patterns, error handling, console output

### Test Files to Study
- **`tests/unit/test_update.py`** (833 lines) - Main test patterns
- **`tests/unit/test_update_error_handling.py`** (364 lines) - Error scenario patterns
- **`tests/unit/test_update_workflow.py`** (393 lines) - Workflow integration

### Configuration Files
- **`pyproject.toml`** - Dependencies, test config
- **`.moai/specs/SPEC-UPDATE-CACHE-FIX-001/spec.md`** - Full requirements

---

## 10. Summary: Ready-to-Use Components

### Already Available Functions
1. `_get_current_version()` - Get installed version
2. `_get_latest_version()` - Fetch PyPI version
3. `_compare_versions()` - Compare version strings
4. `_detect_tool_installer()` - Detect uv/pip/pipx
5. `_execute_upgrade()` - Run upgrade command

### Already Available Infrastructure
1. **Exception hierarchy** - UpdateError and subclasses
2. **Console output** - Rich formatting with colors
3. **Error message helpers** - _show_*_help() functions
4. **Test framework** - CliRunner, mocks, pytest fixtures
5. **Subprocess patterns** - capture_output, timeout, check=False

### Import Statements Ready
```python
import subprocess  # Already imported
from packaging import version  # Already imported
from rich.console import Console  # Already imported (console)
```

### No New External Dependencies Needed
- subprocess (stdlib)
- packaging (already in dependencies)
- pathlib (stdlib)

---

## 11. Implementation Checklist for Developer

### Phase 1: Implementation (TDD-RED)
- [ ] Create `test_update_uv_cache_fix.py` with all test cases
- [ ] Implement `_detect_stale_cache()` function
- [ ] Implement `_clear_uv_package_cache()` function
- [ ] Modify `_execute_upgrade()` or create wrapper for retry logic
- [ ] Integration with main `update()` command flow

### Phase 2: Testing (TDD-GREEN)
- [ ] Run pytest to verify all tests pass
- [ ] Coverage > 85%
- [ ] Test subprocess mocking patterns

### Phase 3: Refinement (TDD-REFACTOR)
- [ ] Code review against SPEC requirements
- [ ] Add inline comments with @CODE tags
- [ ] Test error edge cases
- [ ] Performance testing (timeout values)

### Phase 4: Integration
- [ ] Update main update.py with @CODE:UPDATE-CACHE-FIX-001 tags
- [ ] Create integration test in test_update_workflow.py
- [ ] Document in SPEC-UPDATE-CACHE-FIX-001/acceptance.md

---

## Conclusion

The codebase is well-structured for implementing the cache fix. All necessary infrastructure exists:
- Version comparison functions
- Subprocess execution patterns
- Comprehensive error handling
- Extensive test infrastructure
- Rich console output formatting

The implementation can follow existing patterns without introducing new external dependencies or architecture changes.

