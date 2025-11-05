# @SPEC:UPDATE-REFACTOR-002: Implementation Plan

## Overview

This plan outlines the 5-phase implementation of the moai-adk self-update integration feature. The feature enhances `moai-adk update` command to automatically detect the package manager and execute upgrade, then sync templates in a 2-stage workflow.

---

## Phase 1: Core Logic Implementation (Primary Goal)

### Objective
Implement the 2-stage update workflow with automatic tool detection.

### Step 1.1: Implement `_detect_tool_installer()` Function

**File**: `src/moai_adk/cli/commands/update.py`

**Behavior**:
```
Check if uv tool installed
  ├─ Run: uv tool list | grep moai-adk
  ├─ If found: Return ["uv", "tool", "upgrade", "moai-adk"]
  └─ If not found: Continue to next tool

Check if pipx installed
  ├─ Run: pipx list | grep moai-adk
  ├─ If found: Return ["pipx", "upgrade", "moai-adk"]
  └─ If not found: Continue to fallback

Fallback to pip
  ├─ Return ["pip", "install", "--upgrade", "moai-adk"]
  └─ (pip is assumed to be installed with Python)
```

**Implementation Approach**:
- Use `subprocess.run()` to check each tool
- Capture stdout/stderr to verify moai-adk is installed
- Return command list ready for `subprocess.run()`

**Error Handling**:
- If all tools fail: Return None (handle in main function)
- Suppress command not found errors gracefully

### Step 1.2: Extract `_sync_templates()` Function

**File**: `src/moai_adk/cli/commands/update.py`

**Behavior**:
Separate template synchronization logic from current `update()` function.

**Current Logic** (to be extracted):
1. Check project initialized (`.moai/` exists)
2. Create backup (unless `--force`)
3. Copy templates from package
4. Merge config.json and CLAUDE.md
5. Set `optimized: false` flag
6. Display completion message

**Extracted Signature**:
```python
def _sync_templates(project_path: Path, version_for_config: str,
                    force: bool, check: bool) -> None:
    """Sync templates to project directory."""
```

**Benefits**:
- Reusable in both Stage 1 and Stage 2
- Cleaner main `update()` function
- Easier testing

### Step 1.3: Implement 2-Stage Workflow in `update()` Function

**File**: `src/moai_adk/cli/commands/update.py`

**Main Logic Flow**:

```python
def update(path: str, force: bool, check: bool,
           templates_only: bool, yes: bool) -> None:
    project_path = Path(path).resolve()

    # 1. Check if templates-only mode (skip upgrade)
    if templates_only:
        _sync_templates(project_path, __version__, force, check)
        return

    # 2. Get versions
    current_version = __version__
    latest_version = get_latest_version()  # from PyPI

    # 3. Stage 1: If upgrade needed
    if current_version < latest_version:
        if not yes:
            confirm_prompt()  # User confirmation

        tool_cmd = _detect_tool_installer()
        if not tool_cmd:
            show_error_and_guidance()
            return

        execute_upgrade(tool_cmd)  # Run subprocess
        show_rerun_message()  # Prompt for re-run
        return  # Exit Stage 1

    # 4. Stage 2: If already latest (or after re-run)
    _sync_templates(project_path, current_version, force, check)
```

**Key Design Decisions**:
- **Check mode**: No changes, just report versions
- **Templates-only**: Skip upgrade check entirely
- **Two-stage exit**: After Stage 1 upgrade, exit and let user re-run
- **Idempotent**: Re-running same version executes Stage 2

### Step 1.4: Add CLI Options to Click Command

**File**: `src/moai_adk/cli/commands/update.py`

**New Options**:
```python
@click.command()
@click.option('--path', default='.', help='Project path')
@click.option('--force', is_flag=True, help='Force overwrite without backup')
@click.option('--check', is_flag=True, help='Check for updates without applying')
@click.option('--templates-only', is_flag=True, help='Sync templates, skip package upgrade')
@click.option('--yes', is_flag=True, help='Auto-confirm all prompts')
def update(path, force, check, templates_only, yes):
    """Update MoAI-ADK package and templates."""
```

**Option Behaviors**:
- `--check`: Display version info and exit (no changes)
- `--templates-only`: Skip upgrade, go straight to template sync
- `--yes`: Don't prompt for confirmations
- `--force`: Skip backup creation (use with caution)

---

## Phase 2: CLI Options & User Experience

### Objective
Implement `--templates-only`, `--yes`, `--check`, and `--force` flags with clear messaging.

### Step 2.1: Implement `--check` Mode

**Behavior**:
```
$ moai-adk update --check

🔍 Checking versions...
   Current: 0.6.1
   Latest:  0.6.2

📦 Update available: 0.6.1 → 0.6.2
   Run 'moai-adk update' to upgrade
```

**Implementation**:
- Fetch latest version
- Compare with current
- Display results
- Exit without changes

### Step 2.2: Implement `--templates-only` Flag

**Behavior**:
- Skip tool detection and upgrade
- Go directly to template sync
- Useful for manual upgrade scenarios

**Use Case**:
```bash
# Manual upgrade + template sync
uv tool upgrade moai-adk
moai-adk update --templates-only
```

### Step 2.3: Implement `--yes` Flag

**Behavior**:
- Auto-confirm upgrade prompt
- Auto-confirm template sync
- No user interaction needed

**Use Case**:
```bash
# CI/CD automated update
moai-adk update --yes
moai-adk update --yes  # Re-run for templates
```

### Step 2.4: Rich Terminal Output Improvements

**Stage 1 Output**:
```
🔍 Checking versions...
   Current: 0.6.1
   Latest:  0.6.2

📦 Upgrading package: 0.6.1 → 0.6.2
   Running: uv tool upgrade moai-adk
   [subprocess output...]
   ✅ Package upgraded successfully

📢 Next step:
   Run 'moai-adk update' again to sync templates
```

**Stage 2 Output**:
```
🔍 Checking versions...
   Current: 0.6.2
   Latest:  0.6.2

✓ Package already up to date

💾 Creating backup...
✓ Backup: .moai-backups/20251028-160000/

📄 Syncing templates...
   ✅ .claude/ updated
   ✅ .moai/ updated
   🔄 CLAUDE.md merged
   🔄 config.json merged
   ⚙️  Set optimized=false

✓ Templates synced!

✨ Update complete!
```

---

## Phase 3: Error Handling & Recovery

### Objective
Implement robust error handling with helpful guidance.

### Step 3.1: Installer Detection Failure

**Scenario**: None of uv tool, pipx, pip detected

**Error Message**:
```
❌ Cannot detect package installer

   Installation method not detected.
   To update manually, run:

   • If installed via uv tool:
     uv tool upgrade moai-adk

   • If installed via pipx:
     pipx upgrade moai-adk

   • If installed via pip:
     pip install --upgrade moai-adk

   Then run:
     moai-adk update --templates-only
```

**Implementation**:
- Show detection attempt logs (debug mode)
- Guide user to manual upgrade
- Offer `--templates-only` after manual upgrade

### Step 3.2: Network Failure (PyPI unreachable)

**Scenario**: Cannot fetch latest version from PyPI

**Error Message**:
```
⚠️  Cannot reach PyPI to check latest version

   Options:
   1. Check network connection
   2. Try again with: moai-adk update --force
   3. Proceed with template-only sync: moai-adk update --templates-only
```

**Implementation**:
- Catch `urllib.error.URLError`
- Offer graceful fallback
- Show retry guidance

### Step 3.3: Upgrade Command Failure

**Scenario**: `uv tool upgrade moai-adk` fails

**Error Message**:
```
❌ Upgrade failed: [error message from subprocess]

Troubleshooting:
1. Check network connection
2. Clear cache: uv cache clean
3. Try manually: uv tool upgrade moai-adk
4. Report issue: https://github.com/modu-ai/moai-adk/issues
```

**Implementation**:
- Capture subprocess stderr
- Display relevant parts only
- Show step-by-step recovery

### Step 3.4: Template Sync Failure

**Scenario**: Backup creation or merge fails

**Error Message**:
```
⚠️  Template sync failed: [reason]

Rollback options:
1. Restore from backup: cp -r .moai-backups/TIMESTAMP .moai/
2. Skip backup: moai-adk update --force
3. Manual sync: moai-adk update --templates-only
```

**Implementation**:
- Create rollback instructions
- Show backup location
- Offer force option

---

## Phase 4: Testing

### Objective
Achieve 85%+ test coverage with unit, integration, and platform tests.

### Step 4.1: Unit Tests

**File**: `tests/cli/commands/test_update.py`

**Test Cases**:
```python
def test_detect_uv_tool():
    """Test detection of uv tool installation."""

def test_detect_pipx():
    """Test detection of pipx installation."""

def test_detect_pip_fallback():
    """Test fallback to pip."""

def test_detect_none_available():
    """Test when no tools available."""

def test_sync_templates_basic():
    """Test basic template sync."""

def test_sync_templates_with_backup():
    """Test backup creation during sync."""

def test_sync_templates_no_backup():
    """Test --force flag skips backup."""

def test_update_check_mode():
    """Test --check mode displays versions."""

def test_update_templates_only():
    """Test --templates-only skips upgrade."""

def test_update_yes_flag():
    """Test --yes auto-confirms prompts."""
```

**Coverage Target**: 85%+

### Step 4.2: Integration Tests

**Test Scenarios**:
1. Full 2-stage workflow simulation
2. Version comparison logic
3. Tool command execution
4. Backup and merge logic

### Step 4.3: Platform Tests

**Platforms**: macOS, Linux, Windows

**Test Focus**:
- Tool detection accuracy per platform
- Command execution differences
- Path handling (/ vs \)
- File permissions

---

## Phase 5: Documentation

### Step 5.1: README.md Update

**Section**: "Updating MoAI-ADK"

**Changes**:
- Explain automatic tool detection
- Show 2-stage workflow
- Document --templates-only, --yes, --check, --force options
- Before/after comparison

### Step 5.2: CHANGELOG.md Update

**Entry for v0.6.2**:
```
## [0.6.2] - 2025-10-28

### Added
- **Self-update support in `moai-adk update`**
  - Automatically detects installer (uv tool, pipx, pip)
  - Executes package upgrade
  - Two-stage process for safety
  - New `--templates-only` flag to skip upgrade
  - New `--yes` flag to auto-confirm
  - New `--check` flag to preview updates
  - New `--force` flag to skip backups

### Changed
- `moai-adk update` now handles package upgrade + template sync
- Improved error messages with helpful recovery steps

### Fixed
- Better handling of network failures
- Improved installer detection robustness
```

### Step 5.3: Docstring Updates

**Update**:
- Main `update()` function docstring
- New helper functions docstrings
- CLI option descriptions

---

## Dependencies & Blockers

**No Blocking Dependencies**:
- ✅ SPEC-INIT-001 (initialization) already completed
- ✅ All required libraries available (subprocess, urllib, click, rich)

**Related SPECs**:
- SPEC-INIT-001: Initialization system (provides template processor)

---

## Success Metrics

| Metric | Target | Validation |
|--------|--------|------------|
| Tool detection accuracy | 100% | All 3 tools (uv, pipx, pip) detected |
| 2-stage workflow | 100% | User can re-run for templates |
| Error handling | 95%+ | All error paths tested |
| Test coverage | 85%+ | pytest coverage report |
| Platform support | 100% | macOS, Linux, Windows passing |
| User experience | Subjective | Clear messages, helpful guidance |

---

## Implementation Sequence

1. **Step 1.1**: `_detect_tool_installer()` + unit tests
2. **Step 1.2**: Extract `_sync_templates()` + unit tests
3. **Step 1.3**: Implement 2-stage workflow logic
4. **Step 1.4**: Add CLI options to Click command
5. **Phase 2**: Error handling + user messaging
6. **Phase 3**: Recovery paths and fallbacks
7. **Phase 4**: Full test suite
8. **Phase 5**: Documentation updates

---

## Definition of Done

- ✅ All code passes ruff and mypy
- ✅ Test coverage ≥85%
- ✅ All tests passing (pytest)
- ✅ README updated with examples
- ✅ CHANGELOG documented
- ✅ Docstrings complete
- ✅ Platform-tested on macOS/Linux/Windows
- ✅ SPEC document linked via @TAG in code
- ✅ Ready for `/alfred:3-sync` documentation sync
