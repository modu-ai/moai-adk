
## Overview

This document defines the acceptance criteria using Given-When-Then (Gherkin) scenarios for the moai-adk self-update integration feature.

---

## Scenario 1: User with uv tool Installation

### Given
- User has moai-adk installed via `uv tool install moai-adk`
- Current version: 0.6.1
- Latest version available on PyPI: 0.6.2
- User is in a project directory with `.moai/` initialized

### When
- User runs `moai-adk update`

### Then
- ✅ System detects uv tool installation method
- ✅ System compares: 0.6.1 < 0.6.2 (upgrade needed)
- ✅ System displays upgrade prompt and asks for confirmation
- ✅ System executes: `uv tool upgrade moai-adk`
- ✅ System shows: "Run 'moai-adk update' again to sync templates"
- ✅ System exits (Stage 1 complete)

### And When
- User runs `moai-adk update` again (second invocation)

### And Then
- ✅ System detects current version is now 0.6.2
- ✅ System recognizes 0.6.2 == 0.6.2 (no upgrade needed)
- ✅ System creates backup in `.moai-backups/20251028-HHMMSS/`
- ✅ System copies new templates to `.claude/` and `.moai/`
- ✅ System merges `config.json` preserving project metadata
- ✅ System merges `CLAUDE.md` preserving custom sections
- ✅ System sets `optimized: false` in `.moai/config.json`
- ✅ System displays: "✓ Templates synced!"

### Test Code
```python
def test_uv_tool_two_stage_workflow(tmp_path, monkeypatch):
    """Test complete 2-stage workflow with uv tool."""
    # Setup: Mock uv tool detection
    monkeypatch.setenv('PATH', mock_uv_path)

    # Stage 1: Upgrade available
    result = runner.invoke(update, [str(tmp_path)])
    assert "Current: 0.6.1" in result.output
    assert "Latest: 0.6.2" in result.output
    assert "Running: uv tool upgrade moai-adk" in result.output
    assert "Run 'moai-adk update' again" in result.output
    assert result.exit_code == 0

    # Simulate upgrade completion
    mock_current_version = "0.6.2"

    # Stage 2: Template sync
    result = runner.invoke(update, [str(tmp_path)])
    assert "Package already up to date" in result.output
    assert "Syncing templates" in result.output
    assert "Templates synced!" in result.output
    assert result.exit_code == 0
```

---

## Scenario 2: User with pipx Installation

### Given
- User has moai-adk installed via `pipx install moai-adk`
- Current version: 0.6.1
- Latest version available on PyPI: 0.6.2

### When
- User runs `moai-adk update`

### Then
- ✅ System detects pipx installation method
- ✅ System displays upgrade confirmation
- ✅ System executes: `pipx upgrade moai-adk`
- ✅ System prompts for re-run

### Test Code
```python
def test_pipx_detection():
    """Test pipx installer detection."""
    # Mock pipx installed with moai-adk
    with patch('subprocess.run') as mock_run:
        mock_run.side_effect = [
            CompletedProcess(returncode=1),  # uv tool not found
            CompletedProcess(returncode=0, stdout="moai-adk 0.6.1"),  # pipx found
        ]

        result = _detect_tool_installer()
        assert result == ["pipx", "upgrade", "moai-adk"]
```

---

## Scenario 3: User with pip Installation

### Given
- User has moai-adk installed via `pip install moai-adk`
- Current version: 0.6.0
- Latest version available: 0.6.2

### When
- User runs `moai-adk update`

### Then
- ✅ System detects pip as fallback method
- ✅ System executes: `pip install --upgrade moai-adk`
- ✅ System prompts for re-run

### Test Code
```python
def test_pip_fallback_detection():
    """Test pip as fallback when uv/pipx not found."""
    # Mock neither uv nor pipx installed
    with patch('subprocess.run') as mock_run:
        mock_run.side_effect = [
            CompletedProcess(returncode=1),  # uv not found
            CompletedProcess(returncode=1),  # pipx not found
        ]

        result = _detect_tool_installer()
        assert result == ["pip", "install", "--upgrade", "moai-adk"]
```

---

## Scenario 4: Templates-Only Update (Manual Upgrade)

### Given
- User manually upgraded package: `uv tool upgrade moai-adk`
- Current version: 0.6.2
- Latest version: 0.6.2
- User wants to sync templates without checking versions

### When
- User runs `moai-adk update --templates-only`

### Then
- ✅ System skips tool detection entirely
- ✅ System skips version checking
- ✅ System goes directly to template sync
- ✅ System creates backup
- ✅ System copies and merges templates
- ✅ System displays completion message

### Test Code
```python
def test_templates_only_flag():
    """Test --templates-only bypasses upgrade."""
    result = runner.invoke(update, [str(project_path), '--templates-only'])

    assert "tool detection" not in result.output
    assert "version check" not in result.output
    assert "Syncing templates" in result.output
    assert "Templates synced!" in result.output
    assert result.exit_code == 0
```

---

## Scenario 5: Check Mode (Preview Update)

### Given
- Current version: 0.6.1
- Latest version: 0.6.2
- User wants to check for available updates

### When
- User runs `moai-adk update --check`

### Then
- ✅ System displays current version: 0.6.1
- ✅ System displays latest version: 0.6.2
- ✅ System indicates upgrade is available
- ✅ System exits without making any changes
- ✅ System suggests: "Run 'moai-adk update' to upgrade"

### Test Code
```python
def test_check_mode():
    """Test --check displays versions without changes."""
    result = runner.invoke(update, [str(project_path), '--check'])

    assert "Current: 0.6.1" in result.output
    assert "Latest: 0.6.2" in result.output
    assert "Update available" in result.output
    assert "Templates synced!" not in result.output  # No changes made
    assert result.exit_code == 0
```

---

## Scenario 6: Auto-Confirm Mode (CI/CD)

### Given
- User is running in CI/CD pipeline
- Current version: 0.6.1
- Latest version: 0.6.2

### When
- User runs `moai-adk update --yes`

### Then
- ✅ System auto-confirms upgrade without prompt
- ✅ System executes upgrade
- ✅ System prompts user to run again (can't avoid, requires re-run)
- ✅ System does NOT wait for confirmation

### And When
- User runs `moai-adk update --yes` again (for Stage 2)

### And Then
- ✅ System auto-confirms template sync without prompt
- ✅ System syncs templates
- ✅ System exits successfully
- ✅ Full update completes in 2 automated invocations

### Test Code
```python
def test_yes_flag_auto_confirms():
    """Test --yes flag auto-confirms prompts."""
    # Stage 1
    result = runner.invoke(update, [str(project_path), '--yes'])
    assert "Continue?" not in result.output  # No confirmation prompt
    assert "Running: uv tool upgrade" in result.output
    assert result.exit_code == 0

    # Stage 2
    result = runner.invoke(update, [str(project_path), '--yes'])
    assert "Syncing templates" in result.output
    assert result.exit_code == 0
```

---

## Scenario 7: Force Mode (Skip Backup)

### Given
- User wants to update without creating backup
- Current version is up to date
- Templates need sync

### When
- User runs `moai-adk update --force`

### Then
- ✅ System recognizes `--force` flag
- ✅ System skips backup creation
- ✅ System overwrites templates directly
- ✅ System displays: "Skipping backup (--force)"
- ✅ System syncs templates without merge

### Test Code
```python
def test_force_skip_backup():
    """Test --force skips backup creation."""
    result = runner.invoke(update, [str(project_path), '--force'])

    assert "backup" not in result.output.lower() or "skip" in result.output.lower()
    assert "Syncing templates" in result.output
    assert os.path.exists(backup_path) is False  # No backup created
    assert result.exit_code == 0
```

---

## Scenario 8: No Installation Method Detected

### Given
- User has moai-adk installed (somehow, not via standard tools)
- uv tool detection fails
- pipx detection fails
- pip detection fails (unusual, but possible in custom environments)

### When
- User runs `moai-adk update`

### Then
- ✅ System displays error message
- ✅ System shows diagnostic info (which tools were checked)
- ✅ System provides fallback instructions
- ✅ System suggests manual upgrade commands
- ✅ System exits gracefully (exit code != 0)

### And When
- User manually upgrades: `pip install --upgrade moai-adk`
- User runs: `moai-adk update --templates-only`

### And Then
- ✅ System skips tool detection (due to --templates-only)
- ✅ System syncs templates successfully

### Test Code
```python
def test_installer_detection_failure():
    """Test graceful handling when no installer detected."""
    with patch('subprocess.run') as mock_run:
        # All tool detection fails
        mock_run.side_effect = FileNotFoundError()

        result = runner.invoke(update, [str(project_path)])

        assert "Cannot detect package installer" in result.output
        assert "uv tool" in result.output
        assert "pipx" in result.output
        assert "pip" in result.output
        assert result.exit_code != 0
```

---

## Scenario 9: Network Failure (PyPI Unreachable)

### Given
- Network connection is unavailable
- User runs update command
- PyPI API is unreachable

### When
- User runs `moai-adk update`

### Then
- ✅ System attempts to fetch latest version from PyPI
- ✅ System detects network failure (timeout or connection error)
- ✅ System displays helpful error message
- ✅ System offers recovery options
- ✅ System suggests: `moai-adk update --force` or `moai-adk update --templates-only`

### Test Code
```python
def test_network_failure_handling():
    """Test graceful handling of network failures."""
    with patch('urllib.request.urlopen') as mock_urlopen:
        mock_urlopen.side_effect = urllib.error.URLError("Network unreachable")

        result = runner.invoke(update, [str(project_path)])

        assert "Cannot reach PyPI" in result.output
        assert "Check network connection" in result.output
        assert "--force" in result.output or "--templates-only" in result.output
```

---

## Scenario 10: Upgrade Command Fails

### Given
- Tool is detected correctly
- Upgrade command fails (e.g., `uv tool upgrade moai-adk` returns error)
- User needs recovery guidance

### When
- User runs `moai-adk update`

### Then
- ✅ System executes upgrade command
- ✅ System detects subprocess failure (return code != 0)
- ✅ System captures stderr from subprocess
- ✅ System displays helpful error message with:
  - What failed
  - Troubleshooting steps
  - Manual workaround
  - Link to GitHub issues
- ✅ System exits with appropriate error code

### Test Code
```python
def test_upgrade_failure_handling():
    """Test recovery when upgrade command fails."""
    with patch('subprocess.run') as mock_run:
        mock_run.return_value = CompletedProcess(
            returncode=1,
            stdout="",
            stderr="uv tool not found"
        )

        result = runner.invoke(update, [str(project_path)])

        assert "Upgrade failed" in result.output
        assert "uv tool not found" in result.output
        assert "Troubleshooting" in result.output
        assert "manually" in result.output.lower()
        assert result.exit_code != 0
```

---

## Scenario 11: Already Latest Version

### Given
- Current version: 0.6.2
- Latest version: 0.6.2
- User runs update

### When
- User runs `moai-adk update`

### Then
- ✅ System compares versions: 0.6.2 >= 0.6.2
- ✅ System recognizes: already latest, skip upgrade
- ✅ System directly proceeds to Stage 2 (template sync)
- ✅ System creates backup
- ✅ System syncs templates
- ✅ System displays: "Already up to date"

### Test Code
```python
def test_already_latest_version():
    """Test behavior when already on latest version."""
    # Current == Latest
    result = runner.invoke(update, [str(project_path)])

    assert "Already up to date" in result.output or "latest" in result.output.lower()
    assert "uv tool upgrade" not in result.output  # No upgrade
    assert "Syncing templates" in result.output
    assert result.exit_code == 0
```

---

## Scenario 12: Backup and Merge Integrity

### Given
- User has customized configuration files:
  - `.moai/config.json` with custom project metadata
  - `.claude/agents/custom-agent.md` with custom agent
  - User's edited sections in `CLAUDE.md`

### When
- User runs `moai-adk update` (Stage 2 - template sync)

### Then
- ✅ System creates backup: `.moai-backups/20251028-HHMMSS/`
- ✅ System backup includes all `.moai/` and `.claude/` contents
- ✅ System intelligently merges:
  - `config.json`: Preserves project metadata (name, author, locale)
  - `CLAUDE.md`: Preserves custom ## Project Information sections
  - New template files are added
  - Unchanged template files remain unchanged
- ✅ System sets `optimized: false` for CodeRabbit review
- ✅ User can restore backup if needed

### Test Code
```python
def test_backup_and_merge_integrity():
    """Test backup creation and smart merge logic."""
    # Setup: Create custom files
    custom_agent = project_path / ".claude/agents/custom.md"
    custom_agent.write_text("# Custom Agent")

    # Run update
    result = runner.invoke(update, [str(project_path)])

    # Verify backup exists
    assert backup_path.exists()
    assert (backup_path / "agents/custom.md").exists()

    # Verify merge preserved custom file
    assert (project_path / ".claude/agents/custom.md").exists()
    assert (project_path / ".claude/agents/custom.md").read_text() == "# Custom Agent"

    # Verify config.json metadata preserved
    config = json.loads((project_path / ".moai/config.json").read_text())
    assert config['name'] == original_project_name
    assert config['author'] == original_author
    assert config['optimized'] is False

    assert result.exit_code == 0
```

---

## Platform-Specific Tests

### macOS
- ✅ uv tool detection works
- ✅ pipx detection works
- ✅ pip detection works
- ✅ Path handling correct (/)
- ✅ Subprocess execution succeeds

### Linux
- ✅ All detection methods work
- ✅ File permissions preserved
- ✅ Symlinks handled correctly
- ✅ Command execution succeeds

### Windows
- ✅ Command detection works (.exe, .bat)
- ✅ Path handling correct (\)
- ✅ Environment variables accessible
- ✅ Command execution succeeds

---

## Definition of Acceptance

✅ **All Scenarios Pass**:
- Scenario 1-12 all execute without errors
- Expected outputs match actual outputs
- No data loss or corruption
- Backups created and restorable

✅ **Platform Coverage**:
- macOS: Full test suite passing
- Linux: Full test suite passing
- Windows: Full test suite passing

✅ **Error Handling**:
- All error paths tested
- User guidance clear and actionable
- Recovery options provided

✅ **Quality Metrics**:
- Code passes ruff and mypy
- Test coverage ≥85%
- All integration tests passing
- No regression in existing tests

✅ **User Experience**:
- Messages are clear and helpful
- 2-stage workflow is intuitive
- Recovery from errors is straightforward
- No unexpected behavior
