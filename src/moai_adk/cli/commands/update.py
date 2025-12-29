"""Update command

Update MoAI-ADK to the latest version available on PyPI with 3-stage workflow:
- Stage 1: Package version check (PyPI vs current)
- Stage 2: Config version comparison (template_version in config.json)
- Stage 3: Template sync (only if versions differ)

Includes:
- Automatic installer detection (uv tool, pipx, pip)
- Package upgrade with intelligent re-run prompts
- Template and configuration updates with performance optimization
- Backward compatibility validation
- 70-80% performance improvement for up-to-date projects

## Skill Invocation Guide (English-Only)
# mypy: disable-error-code=return-value

### Related Skills
- **moai-foundation-trust**: For post-update validation
  - Trigger: After updating MoAI-ADK version
  - Invocation: `Skill("moai-foundation-trust")` to verify all toolchains still work

- **moai-foundation-langs**: For language detection after update
  - Trigger: After updating, confirm language stack is intact
  - Invocation: `Skill("moai-foundation-langs")` to re-detect and validate language configuration

### When to Invoke Skills in Related Workflows
1. **After successful update**:
   - Run `Skill("moai-foundation-trust")` to validate all TRUST 4 gates
   - Run `Skill("moai-foundation-langs")` to confirm language toolchain still works
   - Run project doctor command for full system validation

2. **Before updating**:
   - Create backup with `python -m moai_adk backup`

3. **If update fails**:
   - Use backup to restore previous state
   - Debug with `python -m moai_adk doctor --verbose`
"""

# type: ignore

from __future__ import annotations

import json
import logging
import shutil
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Any, Union, cast

import click
import yaml
from packaging import version
from rich.console import Console

from moai_adk import __version__
from moai_adk.core.merge import MergeAnalyzer
from moai_adk.core.migration import VersionMigrator
from moai_adk.core.migration.alfred_to_moai_migrator import AlfredToMoaiMigrator

# Import new custom element restoration modules
from moai_adk.core.migration.custom_element_scanner import create_custom_element_scanner
from moai_adk.core.migration.selective_restorer import create_selective_restorer
from moai_adk.core.migration.user_selection_ui import create_user_selection_ui
from moai_adk.core.template.processor import TemplateProcessor

console = Console()
logger = logging.getLogger(__name__)

# Constants for tool detection
TOOL_DETECTION_TIMEOUT = 5  # seconds
UV_TOOL_COMMAND = ["uv", "tool", "upgrade", "moai-adk"]
PIPX_COMMAND = ["pipx", "upgrade", "moai-adk"]
PIP_COMMAND = ["pip", "install", "--upgrade", "moai-adk"]


# Custom exceptions for better error handling
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


def _get_config_path(project_path: Path) -> tuple[Path, bool]:
    """Get config file path, preferring YAML over JSON.

    Returns:
        Tuple of (config_path, is_yaml)
    """
    yaml_path = project_path / ".moai" / "config" / "config.yaml"
    json_path = project_path / ".moai" / "config" / "config.json"

    if yaml_path.exists():
        return yaml_path, True
    return json_path, False


def _load_config(config_path: Path) -> dict[str, Any]:
    """Load config from YAML or JSON file."""
    if not config_path.exists():
        return {}

    is_yaml = config_path.suffix in (".yaml", ".yml")
    content = config_path.read_text(encoding="utf-8")

    if is_yaml:
        return yaml.safe_load(content) or {}
    return json.loads(content)


def _save_config(config_path: Path, config_data: dict[str, Any]) -> None:
    """Save config to YAML or JSON file."""
    is_yaml = config_path.suffix in (".yaml", ".yml")

    if is_yaml:
        content = yaml.safe_dump(config_data, default_flow_style=False, allow_unicode=True, sort_keys=False)
    else:
        content = json.dumps(config_data, indent=2, ensure_ascii=False) + "\n"

    config_path.write_text(content, encoding="utf-8")


def _is_installed_via_uv_tool() -> bool:
    """Check if moai-adk installed via uv tool.

    Returns:
        True if uv tool list shows moai-adk, False otherwise
    """
    try:
        result = subprocess.run(
            ["uv", "tool", "list"],
            capture_output=True,
            text=True,
            timeout=TOOL_DETECTION_TIMEOUT,
            check=False,
        )
        return result.returncode == 0 and "moai-adk" in result.stdout
    except (FileNotFoundError, subprocess.TimeoutExpired, OSError):
        return False


def _is_installed_via_pipx() -> bool:
    """Check if moai-adk installed via pipx.

    Returns:
        True if pipx list shows moai-adk, False otherwise
    """
    try:
        result = subprocess.run(
            ["pipx", "list"],
            capture_output=True,
            text=True,
            timeout=TOOL_DETECTION_TIMEOUT,
            check=False,
        )
        return result.returncode == 0 and "moai-adk" in result.stdout
    except (FileNotFoundError, subprocess.TimeoutExpired, OSError):
        return False


def _is_installed_via_pip() -> bool:
    """Check if moai-adk installed via pip.

    Returns:
        True if pip show finds moai-adk, False otherwise
    """
    try:
        result = subprocess.run(
            ["pip", "show", "moai-adk"],
            capture_output=True,
            text=True,
            timeout=TOOL_DETECTION_TIMEOUT,
            check=False,
        )
        return result.returncode == 0
    except (FileNotFoundError, subprocess.TimeoutExpired, OSError):
        return False


def _detect_tool_installer() -> list[str] | None:
    """Detect which tool installed moai-adk.

    Checks in priority order:
    1. uv tool (most likely for MoAI-ADK users)
    2. pipx
    3. pip (fallback)

    Returns:
        Command list [tool, ...args] ready for subprocess.run()
        or None if detection fails

    Examples:
        >>> # If uv tool is detected:
        >>> _detect_tool_installer()
        ['uv', 'tool', 'upgrade', 'moai-adk']

        >>> # If pipx is detected:
        >>> _detect_tool_installer()
        ['pipx', 'upgrade', 'moai-adk']

        >>> # If only pip is available:
        >>> _detect_tool_installer()
        ['pip', 'install', '--upgrade', 'moai-adk']

        >>> # If none are detected:
        >>> _detect_tool_installer()
        None
    """
    if _is_installed_via_uv_tool():
        return UV_TOOL_COMMAND
    elif _is_installed_via_pipx():
        return PIPX_COMMAND
    elif _is_installed_via_pip():
        return PIP_COMMAND
    else:
        return None


def _get_current_version() -> str:
    """Get currently installed moai-adk version.

    Returns:
        Version string (e.g., "0.6.1")

    Raises:
        RuntimeError: If version cannot be determined
    """
    return __version__


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
        with urllib.request.urlopen(url, timeout=5) as response:  # nosec B310 - URL is hardcoded HTTPS to PyPI API, no user input
            data = json.loads(response.read().decode("utf-8"))
            return cast(str, data["info"]["version"])
    except (urllib.error.URLError, json.JSONDecodeError, KeyError, TimeoutError) as e:
        raise RuntimeError(f"Failed to fetch latest version from PyPI: {e}") from e


def _compare_versions(current: str, latest: str) -> int:
    """Compare semantic versions.

    Args:
        current: Current version string
        latest: Latest version string

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


def _get_package_config_version() -> str:
    """Get the current package template version.

    This returns the version of the currently installed moai-adk package,
    which is the version of templates that this package provides.

    Returns:
        Version string of the installed package (e.g., "0.6.1")
    """
    # Package template version = current installed package version
    # This is simple and reliable since templates are versioned with the package
    return __version__


def _get_project_config_version(project_path: Path) -> str:
    """Get current project config.json template version.

    This reads the project's .moai/config/config.json to determine the current
    template version that the project is configured with.

    Args:
        project_path: Project directory path (absolute)

    Returns:
        Version string from project's config.json (e.g., "0.6.1")
        Returns "0.0.0" if template_version field not found (indicates no prior sync)

    Raises:
        ValueError: If config file exists but cannot be parsed
    """

    def _is_placeholder_val(value: str) -> bool:
        """Check if value contains unsubstituted template placeholders."""
        return isinstance(value, str) and value.startswith("{{") and value.endswith("}}")

    config_path, _ = _get_config_path(project_path)

    if not config_path.exists():
        # No config yet, treat as version 0.0.0 (needs initial sync)
        return "0.0.0"

    try:
        config_data = _load_config(config_path)
        # Check for template_version in project section
        template_version = config_data.get("project", {}).get("template_version")
        if template_version and not _is_placeholder_val(template_version):
            return template_version

        # Fallback to moai version if no template_version exists
        moai_version = config_data.get("moai", {}).get("version")
        if moai_version and not _is_placeholder_val(moai_version):
            return moai_version

        # If values are placeholders or don't exist, treat as uninitialized (0.0.0 triggers sync)
        return "0.0.0"
    except (json.JSONDecodeError, yaml.YAMLError) as e:
        raise ValueError(f"Failed to parse project config: {e}") from e


def _ask_merge_strategy(yes: bool = False) -> str:
    """
    Ask user to choose merge strategy via CLI prompt.

    Args:
        yes: If True, auto-select "auto" (for --yes flag)

    Returns:
        "auto" or "manual"
    """
    if yes:
        return "auto"

    console.print("\n[cyan]🔀 Choose merge strategy:[/cyan]")
    console.print("[cyan]  [1] Auto-merge (default)[/cyan]")
    console.print("[dim]     → Template installs fresh + user changes preserved + minimal conflicts[/dim]")
    console.print("[cyan]  [2] Manual merge[/cyan]")
    console.print("[dim]     → Backup preserved + merge guide generated + you control merging[/dim]")

    response = click.prompt("Select [1 or 2]", default="1")
    if response == "2":
        return "manual"
    return "auto"


def _generate_manual_merge_guide(backup_path: Path, template_path: Path, project_path: Path) -> Path:
    """
    Generate comprehensive merge guide for manual merging.

    Args:
        backup_path: Path to backup directory
        template_path: Path to template directory
        project_path: Project root path

    Returns:
        Path to generated merge guide
    """
    guide_dir = project_path / ".moai" / "guides"
    guide_dir.mkdir(parents=True, exist_ok=True)

    guide_path = guide_dir / "merge-guide.md"

    # Find changed files
    changed_files = []
    backup_claude = backup_path / ".claude"
    backup_path / ".moai"

    # Compare .claude/
    if backup_claude.exists():
        for file in backup_claude.rglob("*"):
            if file.is_file():
                rel_path = file.relative_to(backup_path)
                current_file = project_path / rel_path
                if current_file.exists():
                    if file.read_text(encoding="utf-8", errors="ignore") != current_file.read_text(
                        encoding="utf-8", errors="ignore"
                    ):
                        changed_files.append(f"  - {rel_path}")
                else:
                    changed_files.append(f"  - {rel_path} (new)")

    # Generate guide
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    guide_content = f"""# Merge Guide - Manual Merge Mode

**Generated**: {timestamp}
**Backup Location**: `{backup_path.relative_to(project_path)}/`

## Summary

During this update, the following files were changed:

{chr(10).join(changed_files) if changed_files else "  (No changes detected)"}

## How to Merge

### Option 1: Using diff (Terminal)

```bash
# Compare specific files
diff {backup_path.name}/.claude/settings.json .claude/settings.json

# View all differences
diff -r {backup_path.name}/ .
```

### Option 2: Using Visual Merge Tool

```bash
# macOS/Linux - Using meld
meld {backup_path.relative_to(project_path)}/ .

# Using VSCode
code --diff {backup_path.relative_to(project_path)}/.claude/settings.json .claude/settings.json
```

### Option 3: Manual Line-by-Line

1. Open backup file in your editor
2. Open current file side-by-side
3. Manually copy your customizations

## Key Files to Review

### .claude/settings.json
- Contains MCP servers, hooks, environment variables
- **Action**: Restore any custom MCP servers and environment variables
- **Location**: {backup_path.relative_to(project_path)}/.claude/settings.json

### .moai/config/config.json
- Contains project configuration and metadata
- **Action**: Verify user-specific settings are preserved
- **Location**: {backup_path.relative_to(project_path)}/.moai/config/config.json

### .claude/commands/, .claude/agents/, .claude/hooks/
- Contains custom scripts and automation
- **Action**: Restore any custom scripts outside of /moai/ folders
- **Location**: {backup_path.relative_to(project_path)}/.claude/

## Migration Checklist

- [ ] Compare `.claude/settings.json`
  - [ ] Restore custom MCP servers
  - [ ] Restore environment variables
  - [ ] Verify hooks are properly configured

- [ ] Review `.moai/config/config.json`
  - [ ] Check version was updated
  - [ ] Verify user settings preserved

- [ ] Restore custom scripts
  - [ ] Any custom commands outside /moai/
  - [ ] Any custom agents outside /moai/
  - [ ] Any custom hooks outside /moai/

- [ ] Run tests
  ```bash
  uv run pytest
  moai-adk validate
  ```

- [ ] Commit changes
  ```bash
  git add .
  git commit -m "merge: Update templates with manual merge"
  ```

## Rollback if Needed

If you want to cancel and restore the backup:

```bash
# Restore everything from backup
cp -r {backup_path.relative_to(project_path)}/.claude .
cp -r {backup_path.relative_to(project_path)}/.moai .
cp {backup_path.relative_to(project_path)}/CLAUDE.md .

# Or restore specific files
cp {backup_path.relative_to(project_path)}/.claude/settings.json .claude/
```

## Questions?

If you encounter merge conflicts or issues:

1. Check the backup folder for original files
2. Compare line-by-line using diff tools
3. Consult documentation: https://adk.mo.ai.kr/update-merge

---

**Backup**: `{backup_path}/`
**Generated**: {timestamp}
"""

    guide_path.write_text(guide_content, encoding="utf-8")
    logger.info(f"✅ Merge guide created: {guide_path}")
    return guide_path


def _migrate_legacy_logs(project_path: Path, dry_run: bool = False) -> bool:
    """Migrate legacy log files to unified directory structure.

    Creates new unified directory structure (.moai/docs/, .moai/logs/archive/) and
    migrates files from legacy locations to new unified structure:
    - .moai/memory/last-session-state.json → .moai/logs/sessions/
    - .moai/error_logs/ → .moai/logs/errors/
    - .moai/reports/ → .moai/docs/reports/

    Args:
        project_path: Project directory path (absolute)
        dry_run: If True, only simulate migration without making changes

    Returns:
        True if migration succeeded or no migration needed, False otherwise

    Raises:
        Exception: If migration fails during actual execution
    """
    try:
        # Define source and target directories
        legacy_memory = project_path / ".moai" / "memory"
        legacy_error_logs = project_path / ".moai" / "error_logs"
        legacy_reports = project_path / ".moai" / "reports"

        # Create new unified directory structure
        new_logs_dir = project_path / ".moai" / "logs"
        new_docs_dir = project_path / ".moai" / "docs"
        new_sessions_dir = new_logs_dir / "sessions"
        new_errors_dir = new_logs_dir / "errors"
        new_archive_dir = new_logs_dir / "archive"
        new_docs_reports_dir = new_docs_dir / "reports"

        migration_log = []
        files_migrated = 0
        files_skipped = 0

        # Check if any legacy directories exist
        has_legacy_files = legacy_memory.exists() or legacy_error_logs.exists() or legacy_reports.exists()

        if not has_legacy_files:
            if not dry_run:
                # Create new directory structure anyway for consistency
                new_logs_dir.mkdir(parents=True, exist_ok=True)
                new_docs_dir.mkdir(parents=True, exist_ok=True)
                new_sessions_dir.mkdir(parents=True, exist_ok=True)
                new_errors_dir.mkdir(parents=True, exist_ok=True)
                new_archive_dir.mkdir(parents=True, exist_ok=True)
                new_docs_reports_dir.mkdir(parents=True, exist_ok=True)
            return True

        if dry_run:
            console.print("[cyan]🔍 Legacy log migration (dry run):[/cyan]")

        # Create new directories if not dry run
        if not dry_run:
            new_logs_dir.mkdir(parents=True, exist_ok=True)
            new_docs_dir.mkdir(parents=True, exist_ok=True)
            new_sessions_dir.mkdir(parents=True, exist_ok=True)
            new_errors_dir.mkdir(parents=True, exist_ok=True)
            new_archive_dir.mkdir(parents=True, exist_ok=True)
            new_docs_reports_dir.mkdir(parents=True, exist_ok=True)

        # Migration 1: .moai/memory/last-session-state.json → .moai/logs/sessions/
        if legacy_memory.exists():
            session_file = legacy_memory / "last-session-state.json"
            if session_file.exists():
                target_file = new_sessions_dir / "last-session-state.json"

                if target_file.exists():
                    files_skipped += 1
                    migration_log.append(f"Skipped: {session_file.relative_to(project_path)} (target already exists)")
                else:
                    if not dry_run:
                        shutil.copy2(session_file, target_file)
                        # Preserve original timestamp
                        shutil.copystat(session_file, target_file)
                        src_path = session_file.relative_to(project_path)
                        dst_path = target_file.relative_to(project_path)
                        migration_log.append(f"Migrated: {src_path} → {dst_path}")
                    else:
                        src_path = session_file.relative_to(project_path)
                        dst_path = target_file.relative_to(project_path)
                        migration_log.append(f"Would migrate: {src_path} → {dst_path}")
                    files_migrated += 1

        # Migration 2: .moai/error_logs/ → .moai/logs/errors/
        if legacy_error_logs.exists() and legacy_error_logs.is_dir():
            for error_file in legacy_error_logs.rglob("*"):
                if error_file.is_file():
                    relative_path = error_file.relative_to(legacy_error_logs)
                    target_file = new_errors_dir / relative_path

                    # Ensure target directory exists
                    if not dry_run:
                        target_file.parent.mkdir(parents=True, exist_ok=True)

                    if target_file.exists():
                        files_skipped += 1
                        error_path = error_file.relative_to(project_path)
                        migration_log.append(f"Skipped: {error_path} (target already exists)")
                    else:
                        if not dry_run:
                            shutil.copy2(error_file, target_file)
                            shutil.copystat(error_file, target_file)
                            error_path = error_file.relative_to(project_path)
                            target_path = target_file.relative_to(project_path)
                            migration_log.append(f"Migrated: {error_path} → {target_path}")
                        else:
                            error_path = error_file.relative_to(project_path)
                            target_path = target_file.relative_to(project_path)
                            migration_log.append(f"Would migrate: {error_path} → {target_path}")
                        files_migrated += 1

        # Migration 3: .moai/reports/ → .moai/docs/reports/
        if legacy_reports.exists() and legacy_reports.is_dir():
            for report_file in legacy_reports.rglob("*"):
                if report_file.is_file():
                    relative_path = report_file.relative_to(legacy_reports)
                    target_file = new_docs_reports_dir / relative_path

                    # Ensure target directory exists
                    if not dry_run:
                        target_file.parent.mkdir(parents=True, exist_ok=True)

                    if target_file.exists():
                        files_skipped += 1
                        report_path = report_file.relative_to(project_path)
                        migration_log.append(f"Skipped: {report_path} (target already exists)")
                    else:
                        if not dry_run:
                            shutil.copy2(report_file, target_file)
                            shutil.copystat(report_file, target_file)
                            report_path = report_file.relative_to(project_path)
                            target_path = target_file.relative_to(project_path)
                            migration_log.append(f"Migrated: {report_path} → {target_path}")
                        else:
                            report_path = report_file.relative_to(project_path)
                            target_path = target_file.relative_to(project_path)
                            migration_log.append(f"Would migrate: {report_path} → {target_path}")
                        files_migrated += 1

        # Create migration log
        migration_log_path = new_logs_dir / "migration-log.json"
        if not dry_run and files_migrated > 0:
            migration_data = {
                "migration_timestamp": datetime.now().isoformat(),
                "moai_adk_version": __version__,
                "files_migrated": files_migrated,
                "files_skipped": files_skipped,
                "migration_log": migration_log,
                "legacy_directories_found": [
                    str(d.relative_to(project_path))
                    for d in [legacy_memory, legacy_error_logs, legacy_reports]
                    if d.exists()
                ],
            }
            json_content = json.dumps(migration_data, indent=2, ensure_ascii=False)
            migration_log_path.write_text(json_content + "\n", encoding="utf-8")

        # Display results
        if files_migrated > 0 or files_skipped > 0:
            if dry_run:
                console.print(f"   [yellow]Would migrate {files_migrated} files, skip {files_skipped} files[/yellow]")
            else:
                console.print(f"   [green]✓ Migrated {files_migrated} legacy log files[/green]")
                if files_skipped > 0:
                    console.print(f"   [yellow]⚠ Skipped {files_skipped} files (already exist)[/yellow]")
                console.print(f"   [dim]   Migration log: {migration_log_path.relative_to(project_path)}[/dim]")
        elif has_legacy_files:
            console.print("   [dim]   No files to migrate[/dim]")

        return True

    except Exception as e:
        console.print(f"   [red]✗ Log migration failed: {e}[/red]")
        logger.error(f"Legacy log migration failed: {e}", exc_info=True)
        return False


def _detect_stale_cache(upgrade_output: str, current_version: str, latest_version: str) -> bool:
    """
    Detect if uv cache is stale by comparing versions.

    A stale cache occurs when PyPI metadata is outdated, causing uv to incorrectly
    report "Nothing to upgrade" even though a newer version exists. This function
    detects this condition by:
    1. Checking if upgrade output contains "Nothing to upgrade"
    2. Verifying that latest version is actually newer than current version

    Uses packaging.version.parse() for robust semantic version comparison that
    handles pre-releases, dev versions, and other PEP 440 version formats correctly.

    Args:
        upgrade_output: Output from uv tool upgrade command
        current_version: Currently installed version (string, e.g., "0.8.3")
        latest_version: Latest version available on PyPI (string, e.g., "0.9.0")

    Returns:
        True if cache is stale (output shows "Nothing to upgrade" but current < latest),
        False otherwise

    Examples:
        >>> _detect_stale_cache("Nothing to upgrade", "0.8.3", "0.9.0")
        True
        >>> _detect_stale_cache("Updated moai-adk", "0.8.3", "0.9.0")
        False
        >>> _detect_stale_cache("Nothing to upgrade", "0.9.0", "0.9.0")
        False
    """
    # Check if output indicates no upgrade needed
    if not upgrade_output or "Nothing to upgrade" not in upgrade_output:
        return False

    # Compare versions using packaging.version
    try:
        current_v = version.parse(current_version)
        latest_v = version.parse(latest_version)
        return current_v < latest_v
    except (version.InvalidVersion, TypeError) as e:
        # Graceful degradation: if version parsing fails, assume cache is not stale
        logger.debug(f"Version parsing failed: {e}")
        return False


def _clear_uv_package_cache(package_name: str = "moai-adk") -> bool:
    """
    Clear uv cache for specific package.

    Executes `uv cache clean <package>` with 10-second timeout to prevent
    hanging on network issues. Provides user-friendly error handling for
    various failure scenarios (timeout, missing uv, etc.).

    Args:
        package_name: Package name to clear cache for (default: "moai-adk")

    Returns:
        True if cache cleared successfully, False otherwise

    Exceptions:
        - subprocess.TimeoutExpired: Logged as warning, returns False
        - FileNotFoundError: Logged as warning, returns False
        - Exception: Logged as warning, returns False

    Examples:
        >>> _clear_uv_package_cache("moai-adk")
        True  # If uv cache clean succeeds
    """
    try:
        result = subprocess.run(
            ["uv", "cache", "clean", package_name],
            capture_output=True,
            text=True,
            timeout=10,  # 10 second timeout
            check=False,
        )

        if result.returncode == 0:
            logger.debug(f"UV cache cleared for {package_name}")
            return True
        else:
            logger.warning(f"Failed to clear UV cache: {result.stderr}")
            return False

    except subprocess.TimeoutExpired:
        logger.warning(f"UV cache clean timed out for {package_name}")
        return False
    except FileNotFoundError:
        logger.warning("UV command not found. Is uv installed?")
        return False
    except Exception as e:
        logger.warning(f"Unexpected error clearing cache: {e}")
        return False


def _execute_upgrade_with_retry(installer_cmd: list[str], package_name: str = "moai-adk") -> bool:
    """
    Execute upgrade with automatic cache retry on stale detection.

    Implements a robust 7-stage upgrade flow that handles PyPI cache staleness:

    Stage 1: First upgrade attempt (up to 60 seconds)
    Stage 2: Check success condition (returncode=0 AND no "Nothing to upgrade")
    Stage 3: Detect stale cache using _detect_stale_cache()
    Stage 4: Show user feedback if stale cache detected
    Stage 5: Clear cache using _clear_uv_package_cache()
    Stage 6: Retry upgrade with same command
    Stage 7: Return final result (success or failure)

    Retry Logic:
    - Only ONE retry is performed to prevent infinite loops
    - Retry only happens if stale cache is detected AND cache clear succeeds
    - Cache clear failures are reported to user with manual workaround

    User Feedback:
    - Shows emoji-based status messages for each stage
    - Clear guidance on manual workaround if automatic retry fails
    - All errors logged at WARNING level for debugging

    Args:
        installer_cmd: Command list from _detect_tool_installer()
                      e.g., ["uv", "tool", "upgrade", "moai-adk"]
        package_name: Package name for cache clearing (default: "moai-adk")

    Returns:
        True if upgrade succeeded (either first attempt or after retry),
        False otherwise

    Examples:
        >>> # First attempt succeeds
        >>> _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])
        True

        >>> # First attempt stale, retry succeeds
        >>> _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])
        True  # After cache clear and retry

    Raises:
        subprocess.TimeoutExpired: Re-raised if upgrade command times out
    """
    # Stage 1: First upgrade attempt
    try:
        result = subprocess.run(installer_cmd, capture_output=True, text=True, timeout=60, check=False)
    except subprocess.TimeoutExpired:
        raise  # Re-raise timeout for caller to handle
    except Exception:
        return False

    # Stage 2: Check if upgrade succeeded without stale cache
    if result.returncode == 0 and "Nothing to upgrade" not in result.stdout:
        return True

    # Stage 3: Detect stale cache
    try:
        current_version = _get_current_version()
        latest_version = _get_latest_version()
    except RuntimeError:
        # If version check fails, return original result
        return result.returncode == 0

    if _detect_stale_cache(result.stdout, current_version, latest_version):
        # Stage 4: User feedback
        console.print("[yellow]⚠️ Cache outdated, refreshing...[/yellow]")

        # Stage 5: Clear cache
        if _clear_uv_package_cache(package_name):
            console.print("[cyan]♻️ Cache cleared, retrying upgrade...[/cyan]")

            # Stage 6: Retry upgrade
            try:
                result = subprocess.run(
                    installer_cmd,
                    capture_output=True,
                    text=True,
                    timeout=60,
                    check=False,
                )

                if result.returncode == 0:
                    return True
                else:
                    console.print("[red]✗ Upgrade failed after retry[/red]")
                    return False
            except subprocess.TimeoutExpired:
                raise  # Re-raise timeout
            except Exception:
                return False
        else:
            # Cache clear failed
            console.print("[red]✗ Cache clear failed. Manual workaround:[/red]")
            console.print("  [cyan]uv cache clean moai-adk && moai-adk update[/cyan]")
            return False

    # Stage 7: Cache is not stale, return original result
    return result.returncode == 0


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
        result = subprocess.run(installer_cmd, capture_output=True, text=True, timeout=60, check=False)
        return result.returncode == 0
    except subprocess.TimeoutExpired:
        raise  # Re-raise timeout for caller to handle
    except Exception:
        return False


def _preserve_user_settings(project_path: Path) -> dict[str, Path | None]:
    """Back up user-specific settings files before template sync.

    Args:
        project_path: Project directory path

    Returns:
        Dictionary with backup paths of preserved files
    """
    preserved = {}
    claude_dir = project_path / ".claude"

    # Preserve settings.local.json (user MCP and GLM configuration)
    settings_local = claude_dir / "settings.local.json"
    if settings_local.exists():
        try:
            backup_dir = project_path / ".moai-backups" / "settings-backup"
            backup_dir.mkdir(parents=True, exist_ok=True)
            backup_path = backup_dir / "settings.local.json"
            backup_path.write_text(settings_local.read_text(encoding="utf-8"))
            preserved["settings.local.json"] = backup_path
            console.print("   [cyan]💾 Backed up user settings[/cyan]")
        except Exception as e:
            logger.warning(f"Failed to backup settings.local.json: {e}")
            preserved["settings.local.json"] = None
    else:
        preserved["settings.local.json"] = None

    return preserved


def _restore_user_settings(project_path: Path, preserved: dict[str, Path | None]) -> bool:
    """Restore user-specific settings files after template sync.

    Args:
        project_path: Project directory path
        preserved: Dictionary of backup paths from _preserve_user_settings()

    Returns:
        True if restoration succeeded, False otherwise
    """
    claude_dir = project_path / ".claude"
    claude_dir.mkdir(parents=True, exist_ok=True)

    success = True

    # Restore settings.local.json
    if preserved.get("settings.local.json"):
        try:
            backup_path = preserved["settings.local.json"]
            settings_local = claude_dir / "settings.local.json"
            settings_local.write_text(backup_path.read_text(encoding="utf-8"))
            console.print("   [cyan]✓ Restored user settings[/cyan]")
        except Exception as e:
            console.print(f"   [yellow]⚠️ Failed to restore settings.local.json: {e}[/yellow]")
            logger.warning(f"Failed to restore settings.local.json: {e}")
            success = False

    return success


def _get_template_skill_names() -> set[str]:
    """Get set of skill folder names from installed template.

    Returns:
        Set of skill folder names that are part of the template package.
    """
    template_path = Path(__file__).parent.parent.parent / "templates"
    skills_path = template_path / ".claude" / "skills"

    if not skills_path.exists():
        return set()

    return {d.name for d in skills_path.iterdir() if d.is_dir()}


def _get_template_command_names() -> set[str]:
    """Get set of command file names from installed template.

    Returns:
        Set of .md command file names from .claude/commands/moai/ in template.
    """
    template_path = Path(__file__).parent.parent.parent / "templates"
    commands_path = template_path / ".claude" / "commands" / "moai"

    if not commands_path.exists():
        return set()

    return {f.name for f in commands_path.iterdir() if f.is_file() and f.suffix == ".md"}


def _get_template_agent_names() -> set[str]:
    """Get set of agent file names from installed template.

    Returns:
        Set of agent file names from .claude/agents/ in template.
    """
    template_path = Path(__file__).parent.parent.parent / "templates"
    agents_path = template_path / ".claude" / "agents"

    if not agents_path.exists():
        return set()

    return {f.name for f in agents_path.iterdir() if f.is_file()}


def _get_template_hook_names() -> set[str]:
    """Get set of hook file names from installed template.

    Returns:
        Set of .py hook file names from .claude/hooks/moai/ in template.
    """
    template_path = Path(__file__).parent.parent.parent / "templates"
    hooks_path = template_path / ".claude" / "hooks" / "moai"

    if not hooks_path.exists():
        return set()

    return {f.name for f in hooks_path.iterdir() if f.is_file() and f.suffix == ".py"}


def _detect_custom_commands(project_path: Path, template_commands: set[str]) -> list[str]:
    """Detect custom commands NOT in template (user-created).

    Args:
        project_path: Project path (absolute)
        template_commands: Set of template command file names

    Returns:
        Sorted list of custom command file names.
    """
    commands_path = project_path / ".claude" / "commands" / "moai"

    if not commands_path.exists():
        return []

    project_commands = {f.name for f in commands_path.iterdir() if f.is_file() and f.suffix == ".md"}
    custom_commands = project_commands - template_commands

    return sorted(custom_commands)


def _detect_custom_agents(project_path: Path, template_agents: set[str]) -> list[str]:
    """Detect custom agents NOT in template (user-created).

    Args:
        project_path: Project path (absolute)
        template_agents: Set of template agent file names

    Returns:
        Sorted list of custom agent file names.
    """
    agents_path = project_path / ".claude" / "agents"

    if not agents_path.exists():
        return []

    project_agents = {f.name for f in agents_path.iterdir() if f.is_file()}
    custom_agents = project_agents - template_agents

    return sorted(custom_agents)


def _detect_custom_hooks(project_path: Path, template_hooks: set[str]) -> list[str]:
    """Detect custom hooks NOT in template (user-created).

    Args:
        project_path: Project path (absolute)
        template_hooks: Set of template hook file names

    Returns:
        Sorted list of custom hook file names.
    """
    hooks_path = project_path / ".claude" / "hooks" / "moai"

    if not hooks_path.exists():
        return []

    project_hooks = {f.name for f in hooks_path.iterdir() if f.is_file() and f.suffix == ".py"}
    custom_hooks = project_hooks - template_hooks

    return sorted(custom_hooks)


def _group_custom_files_by_type(
    custom_commands: list[str],
    custom_agents: list[str],
    custom_hooks: list[str],
) -> dict[str, list[str]]:
    """Group custom files by type for UI display.

    Args:
        custom_commands: List of custom command file names
        custom_agents: List of custom agent file names
        custom_hooks: List of custom hook file names

    Returns:
        Dictionary with keys: commands, agents, hooks
    """
    return {
        "commands": custom_commands,
        "agents": custom_agents,
        "hooks": custom_hooks,
    }


def _prompt_custom_files_restore(
    custom_commands: list[str],
    custom_agents: list[str],
    custom_hooks: list[str],
    yes: bool = False,
) -> dict[str, list[str]]:
    """Interactive fuzzy checkbox for custom files restore with search support.

    Args:
        custom_commands: List of custom command file names
        custom_agents: List of custom agent file names
        custom_hooks: List of custom hook file names
        yes: Auto-confirm flag (skips restoration in CI/CD mode)

    Returns:
        Dictionary with selected files grouped by type.
    """
    # If no custom files, skip UI
    if not (custom_commands or custom_agents or custom_hooks):
        return {
            "commands": [],
            "agents": [],
            "hooks": [],
        }

    # In --yes mode, skip restoration (safest default)
    if yes:
        console.print("\n[dim]   Skipping custom files restoration (--yes mode)[/dim]\n")
        return {
            "commands": [],
            "agents": [],
            "hooks": [],
        }

    # Try to use new UI, fallback to questionary if import fails
    try:
        from moai_adk.cli.ui.prompts import create_grouped_choices, fuzzy_checkbox

        # Build grouped choices for fuzzy checkbox
        groups: dict[str, list[dict[str, str]]] = {}

        if custom_commands:
            groups["Commands (.claude/commands/moai/)"] = [
                {"name": cmd, "value": f"cmd:{cmd}"} for cmd in custom_commands
            ]

        if custom_agents:
            groups["Agents (.claude/agents/)"] = [{"name": agent, "value": f"agent:{agent}"} for agent in custom_agents]

        if custom_hooks:
            groups["Hooks (.claude/hooks/moai/)"] = [{"name": hook, "value": f"hook:{hook}"} for hook in custom_hooks]

        choices = create_grouped_choices(groups)

        console.print("\n[#DA7756]📦 Custom files detected in backup:[/#DA7756]")
        console.print("[dim]   Use fuzzy search to find files quickly[/dim]\n")

        selected = fuzzy_checkbox(
            "Select custom files to restore:",
            choices=choices,
            instruction="[Space] Toggle  [Tab] All  [Enter] Confirm  [Type to search]",
        )

    except ImportError:
        # Fallback to questionary if new UI not available
        import questionary
        from questionary import Choice, Separator

        choices_legacy: list[Union[Separator, Choice]] = []

        if custom_commands:
            choices_legacy.append(Separator("Commands (.claude/commands/moai/)"))
            for cmd in custom_commands:
                choices_legacy.append(Choice(title=cmd, value=f"cmd:{cmd}"))

        if custom_agents:
            choices_legacy.append(Separator("Agents (.claude/agents/)"))
            for agent in custom_agents:
                choices_legacy.append(Choice(title=agent, value=f"agent:{agent}"))

        if custom_hooks:
            choices_legacy.append(Separator("Hooks (.claude/hooks/moai/)"))
            for hook in custom_hooks:
                choices_legacy.append(Choice(title=hook, value=f"hook:{hook}"))

        console.print("\n[cyan]📦 Custom files detected in backup:[/cyan]")
        console.print("[dim]   Select files to restore (none selected by default)[/dim]\n")

        selected = questionary.checkbox(
            "Select custom files to restore:",
            choices=choices_legacy,
        ).ask()

    # Parse results
    result_commands = []
    result_agents = []
    result_hooks = []

    if selected:
        for item in selected:
            if item.startswith("cmd:"):
                result_commands.append(item[4:])
            elif item.startswith("agent:"):
                result_agents.append(item[6:])
            elif item.startswith("hook:"):
                result_hooks.append(item[5:])

    return {
        "commands": result_commands,
        "agents": result_agents,
        "hooks": result_hooks,
    }


def _restore_custom_files(
    project_path: Path,
    backup_path: Path,
    selected_commands: list[str],
    selected_agents: list[str],
    selected_hooks: list[str],
) -> bool:
    """Restore selected custom files from backup to project.

    Args:
        project_path: Project directory path
        backup_path: Backup directory path
        selected_commands: List of command files to restore
        selected_agents: List of agent files to restore
        selected_hooks: List of hook files to restore

    Returns:
        True if all restorations succeeded, False otherwise.
    """
    import shutil

    success = True

    # Restore commands
    if selected_commands:
        commands_dst = project_path / ".claude" / "commands" / "moai"
        commands_dst.mkdir(parents=True, exist_ok=True)

        for cmd_file in selected_commands:
            src = backup_path / ".claude" / "commands" / "moai" / cmd_file
            dst = commands_dst / cmd_file

            if src.exists():
                try:
                    shutil.copy2(src, dst)
                except Exception as e:
                    logger.warning(f"Failed to restore command {cmd_file}: {e}")
                    success = False
            else:
                logger.warning(f"Command file not in backup: {cmd_file}")
                success = False

    # Restore agents
    if selected_agents:
        agents_dst = project_path / ".claude" / "agents"
        agents_dst.mkdir(parents=True, exist_ok=True)

        for agent_file in selected_agents:
            src = backup_path / ".claude" / "agents" / agent_file
            dst = agents_dst / agent_file

            if src.exists():
                try:
                    shutil.copy2(src, dst)
                except Exception as e:
                    logger.warning(f"Failed to restore agent {agent_file}: {e}")
                    success = False
            else:
                logger.warning(f"Agent file not in backup: {agent_file}")
                success = False

    # Restore hooks
    if selected_hooks:
        hooks_dst = project_path / ".claude" / "hooks" / "moai"
        hooks_dst.mkdir(parents=True, exist_ok=True)

        for hook_file in selected_hooks:
            src = backup_path / ".claude" / "hooks" / "moai" / hook_file
            dst = hooks_dst / hook_file

            if src.exists():
                try:
                    shutil.copy2(src, dst)
                except Exception as e:
                    logger.warning(f"Failed to restore hook {hook_file}: {e}")
                    success = False
            else:
                logger.warning(f"Hook file not in backup: {hook_file}")
                success = False

    return success


def _detect_custom_skills(project_path: Path, template_skills: set[str]) -> list[str]:
    """Detect skills NOT in template (user-created).

    Args:
        project_path: Project path (absolute)
        template_skills: Set of template skill names

    Returns:
        Sorted list of custom skill names.
    """
    skills_path = project_path / ".claude" / "skills"

    if not skills_path.exists():
        return []

    project_skills = {d.name for d in skills_path.iterdir() if d.is_dir()}
    custom_skills = project_skills - template_skills

    return sorted(custom_skills)


def _prompt_skill_restore(custom_skills: list[str], yes: bool = False) -> list[str]:
    """Interactive fuzzy checkbox for skill restore with search support.

    Args:
        custom_skills: List of custom skill names
        yes: Auto-confirm flag (skips restoration in CI/CD mode)

    Returns:
        List of skills user selected to restore.
    """
    if not custom_skills:
        return []

    console.print("\n[#DA7756]📦 Custom skills detected in backup:[/#DA7756]")
    for skill in custom_skills:
        console.print(f"   • {skill}")
    console.print()

    if yes:
        console.print("[dim]   Skipping restoration (--yes mode)[/dim]\n")
        return []

    # Try new UI, fallback to questionary
    try:
        from moai_adk.cli.ui.prompts import fuzzy_checkbox

        choices = [{"name": skill, "value": skill} for skill in custom_skills]

        selected = fuzzy_checkbox(
            "Select skills to restore (type to search):",
            choices=choices,
            instruction="[Space] Toggle  [Tab] All  [Enter] Confirm  [Type to search]",
        )

    except ImportError:
        import questionary

        selected = questionary.checkbox(
            "Select skills to restore (none selected by default):",
            choices=[questionary.Choice(title=skill, checked=False) for skill in custom_skills],
        ).ask()

    return selected if selected else []


def _restore_selected_skills(skills: list[str], backup_path: Path, project_path: Path) -> bool:
    """Restore selected skills from backup.

    Args:
        skills: List of skill names to restore
        backup_path: Backup directory path
        project_path: Project path (absolute)

    Returns:
        True if all restorations succeeded.
    """
    import shutil

    if not skills:
        return True

    console.print("\n[cyan]📥 Restoring selected skills...[/cyan]")
    skills_dst = project_path / ".claude" / "skills"
    skills_dst.mkdir(parents=True, exist_ok=True)

    success = True
    for skill_name in skills:
        src = backup_path / ".claude" / "skills" / skill_name
        dst = skills_dst / skill_name

        if src.exists():
            try:
                shutil.copytree(src, dst, dirs_exist_ok=True)
                console.print(f"   [green]✓ Restored: {skill_name}[/green]")
            except Exception as e:
                console.print(f"   [red]✗ Failed: {skill_name} - {e}[/red]")
                success = False
        else:
            console.print(f"   [yellow]⚠ Not in backup: {skill_name}[/yellow]")
            success = False

    return success


def _show_post_update_guidance(backup_path: Path) -> None:
    """Show post-update guidance for running /moai:0-project update.

    Args:
        backup_path: Backup directory path for reference
    """
    console.print("\n" + "[cyan]" + "=" * 60 + "[/cyan]")
    console.print("[green]✅ Update complete![/green]")
    console.print("\n[yellow]📝 IMPORTANT - Next step:[/yellow]")
    console.print("   Run [cyan]/moai:0-project update[/cyan] in Claude Code")
    console.print("\n   This will:")
    console.print("   • Merge your settings with new templates")
    console.print("   • Validate configuration compatibility")
    console.print("\n[dim]💡 Personal instructions should go in CLAUDE.local.md[/dim]")
    console.print(f"[dim]📂 Backup location: {backup_path}[/dim]")
    console.print("[cyan]" + "=" * 60 + "[/cyan]\n")


def _sync_templates(project_path: Path, force: bool = False, yes: bool = False) -> bool:
    """Sync templates to project with rollback mechanism.

    Args:
        project_path: Project path (absolute)
        force: Force update without backup
        yes: Auto-confirm flag (skips interactive prompts)

    Returns:
        True if sync succeeded, False otherwise
    """
    from moai_adk.core.template.backup import TemplateBackup

    backup_path = None
    try:
        # NEW: Detect custom files and skills BEFORE backup/sync
        template_skills = _get_template_skill_names()
        _detect_custom_skills(project_path, template_skills)

        # Detect custom commands, agents, and hooks
        template_commands = _get_template_command_names()
        _detect_custom_commands(project_path, template_commands)

        template_agents = _get_template_agent_names()
        _detect_custom_agents(project_path, template_agents)

        template_hooks = _get_template_hook_names()
        _detect_custom_hooks(project_path, template_hooks)

        processor = TemplateProcessor(project_path)

        # Create pre-sync backup for rollback
        if not force:
            backup = TemplateBackup(project_path)
            if backup.has_existing_files():
                backup_path = backup.create_backup()
                console.print(f"💾 Created backup: {backup_path.name}")

                # Merge analysis using Pure Python semantic heuristics
                try:
                    analyzer = MergeAnalyzer(project_path)
                    # Template source path from installed package
                    template_path = Path(__file__).parent.parent.parent / "templates"

                    console.print("\n[cyan]🔍 Starting merge analysis...[/cyan]")
                    console.print("[dim]   Analyzing templates with semantic heuristics.[/dim]\n")
                    analysis = analyzer.analyze_merge(backup_path, template_path)

                    # Ask user confirmation
                    if not analyzer.ask_user_confirmation(analysis):
                        console.print("[yellow]⚠️  User cancelled the update.[/yellow]")
                        backup.restore_backup(backup_path)
                        return False
                except Exception as e:
                    console.print(f"[yellow]⚠️  Merge analysis failed: {e}[/yellow]")
                    console.print("[yellow]Proceeding with automatic merge.[/yellow]")

        # Load existing config
        existing_config = _load_existing_config(project_path)

        # Build context
        context = _build_template_context(project_path, existing_config, __version__)
        if context:
            processor.set_context(context)

        # Copy templates (including moai folder)
        processor.copy_templates(backup=False, silent=True)

        # Stage 1.5: Alfred → Moai migration (AFTER template sync)
        # Execute migration after template copy (moai folders must exist first)
        migrator = AlfredToMoaiMigrator(project_path)
        if migrator.needs_migration():
            console.print("\n[cyan]🔄 Migrating folder structure: Alfred → Moai[/cyan]")
            try:
                if not migrator.execute_migration(backup_path):
                    console.print("[red]❌ Alfred → Moai migration failed[/red]")
                    if backup_path:
                        console.print("[yellow]🔄 Restoring from backup...[/yellow]")
                        backup = TemplateBackup(project_path)
                        backup.restore_backup(backup_path)
                    return False
            except Exception as e:
                console.print(f"[red]❌ Error during migration: {e}[/red]")
                if backup_path:
                    backup = TemplateBackup(project_path)
                    backup.restore_backup(backup_path)
                return False

        # Validate template substitution
        validation_passed = _validate_template_substitution_with_rollback(project_path, backup_path)
        if not validation_passed:
            if backup_path:
                console.print(f"[yellow]🔄 Rolling back to backup: {backup_path.name}[/yellow]")
                backup.restore_backup(backup_path)
            return False

        # Preserve metadata
        _preserve_project_metadata(project_path, context, existing_config, __version__)
        _apply_context_to_file(processor, project_path / "CLAUDE.md")

        # Set optimized=false
        set_optimized_false(project_path)

        # Update companyAnnouncements in settings.local.json
        try:
            import sys

            utils_dir = (
                Path(__file__).parent.parent.parent / "templates" / ".claude" / "hooks" / "moai" / "shared" / "utils"
            )

            if utils_dir.exists():
                sys.path.insert(0, str(utils_dir))
                try:
                    from announcement_translator import auto_translate_and_update

                    console.print("[cyan]Updating announcements...[/cyan]")
                    auto_translate_and_update(project_path)
                    console.print("[green]✓ Announcements updated[/green]")
                except Exception as e:
                    console.print(f"[yellow]⚠️  Announcement update failed: {e}[/yellow]")
                finally:
                    sys.path.remove(str(utils_dir))

        except Exception as e:
            console.print(f"[yellow]⚠️  Announcement module not available: {e}[/yellow]")

        # NEW: Interactive custom element restore using new system
        _handle_custom_element_restoration(project_path, backup_path, yes)

        # NEW: Migrate legacy logs to unified structure
        console.print("\n[cyan]📁 Migrating legacy log files...[/cyan]")
        if not _migrate_legacy_logs(project_path):
            console.print("[yellow]⚠️ Legacy log migration failed, but update continuing[/yellow]")

        # Clean up legacy presets directory
        _cleanup_legacy_presets(project_path)

        # NEW: Show post-update guidance
        if backup_path:
            _show_post_update_guidance(backup_path)

        return True
    except Exception as e:
        console.print(f"[red]✗ Template sync failed: {e}[/red]")
        if backup_path:
            console.print(f"[yellow]🔄 Rolling back to backup: {backup_path.name}[/yellow]")
            try:
                backup = TemplateBackup(project_path)
                backup.restore_backup(backup_path)
                console.print("[green]✅ Rollback completed[/green]")
            except Exception as rollback_error:
                console.print(f"[red]✗ Rollback failed: {rollback_error}[/red]")
        return False


def get_latest_version() -> str | None:
    """Get the latest version from PyPI.

    DEPRECATED: Use _get_latest_version() for new code.
    This function is kept for backward compatibility.

    Returns:
        Latest version string, or None if fetch fails.
    """
    try:
        return _get_latest_version()
    except RuntimeError:
        # Return None if PyPI check fails (backward compatibility)
        return None


def set_optimized_false(project_path: Path) -> None:
    """Set config's optimized field to false.

    Args:
        project_path: Project path (absolute).
    """
    config_path, _ = _get_config_path(project_path)
    if not config_path.exists():
        return

    try:
        config_data = _load_config(config_path)
        config_data.setdefault("project", {})["optimized"] = False
        _save_config(config_path, config_data)
    except (json.JSONDecodeError, yaml.YAMLError, KeyError):
        # Ignore errors if config is invalid
        pass


def _load_existing_config(project_path: Path) -> dict[str, Any]:
    """Load existing config (YAML or JSON) if available."""
    config_path, _ = _get_config_path(project_path)
    if config_path.exists():
        try:
            return _load_config(config_path)
        except (json.JSONDecodeError, yaml.YAMLError):
            console.print("[yellow]⚠ Existing config could not be parsed. Proceeding with defaults.[/yellow]")
    return {}


def _is_placeholder(value: Any) -> bool:
    """Check if a string value is an unsubstituted template placeholder."""
    return isinstance(value, str) and value.strip().startswith("{{") and value.strip().endswith("}}")


def _coalesce(*values: Any, default: str = "") -> str:
    """Return the first non-empty, non-placeholder string value."""
    for value in values:
        if isinstance(value, str):
            if not value.strip():
                continue
            if _is_placeholder(value):
                continue
            return value
    for value in values:
        if value is not None and not isinstance(value, str):
            return str(value)
    return default


def _extract_project_section(config: dict[str, Any]) -> dict[str, Any]:
    """Return the nested project section if present."""
    project_section = config.get("project")
    if isinstance(project_section, dict):
        return project_section
    return {}


def _build_template_context(
    project_path: Path,
    existing_config: dict[str, Any],
    version_for_config: str,
) -> dict[str, str]:
    """Build substitution context for template files."""
    import platform

    project_section = _extract_project_section(existing_config)

    project_name = _coalesce(
        project_section.get("name"),
        existing_config.get("projectName"),  # Legacy fallback
        project_path.name,
    )
    project_mode = _coalesce(
        project_section.get("mode"),
        existing_config.get("mode"),  # Legacy fallback
        default="personal",
    )
    project_description = _coalesce(
        project_section.get("description"),
        existing_config.get("projectDescription"),  # Legacy fallback
        existing_config.get("description"),  # Legacy fallback
    )
    project_version = _coalesce(
        project_section.get("version"),
        existing_config.get("projectVersion"),
        existing_config.get("version"),
        default="0.1.0",
    )
    created_at = _coalesce(
        project_section.get("created_at"),
        existing_config.get("created_at"),
        default=datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
    )

    # Detect OS for cross-platform Hook path configuration
    hook_project_dir = "%CLAUDE_PROJECT_DIR%" if platform.system() == "Windows" else "$CLAUDE_PROJECT_DIR"

    # Detect OS for cross-platform statusline command
    # Windows: Use python -m for better PATH compatibility
    # Unix: Use moai-adk directly (assumes installed via uv tool)
    if platform.system() == "Windows":
        statusline_command = "python -m moai_adk statusline"
    else:
        statusline_command = "moai-adk statusline"

    # Extract and resolve language configuration using centralized resolver
    try:
        from moai_adk.core.language_config_resolver import get_resolver

        # Use language resolver to get complete configuration
        resolver = get_resolver(str(project_path))
        resolved_config = resolver.resolve_config()

        # Extract language configuration with environment variable priority
        language_config = {
            "conversation_language": resolved_config.get("conversation_language", "en"),
            "conversation_language_name": resolved_config.get("conversation_language_name", "English"),
            "agent_prompt_language": resolved_config.get("agent_prompt_language", "en"),
        }

        # Extract user personalization
        user_name = resolved_config.get("user_name", "")
        personalized_greeting = resolver.get_personalized_greeting(resolved_config)
        config_source = resolved_config.get("config_source", "config_file")

    except ImportError:
        # Fallback to basic language config extraction if resolver not available
        language_config = existing_config.get("language", {})
        if not isinstance(language_config, dict):
            language_config = {}

        user_name = existing_config.get("user", {}).get("name", "")
        conv_lang = language_config.get("conversation_language")
        personalized_greeting = f"{user_name}님" if user_name and conv_lang == "ko" else user_name
        config_source = "config_file"

    # Enhanced version formatting (matches TemplateProcessor.get_enhanced_version_context)
    def format_short_version(v: str) -> str:
        """Remove 'v' prefix if present."""
        return v[1:] if v.startswith("v") else v

    def format_display_version(v: str) -> str:
        """Format display version with proper formatting."""
        if v == "unknown":
            return "MoAI-ADK unknown version"
        elif v.startswith("v"):
            return f"MoAI-ADK {v}"
        else:
            return f"MoAI-ADK v{v}"

    def format_trimmed_version(v: str, max_length: int = 10) -> str:
        """Format version with maximum length for UI displays."""
        if v == "unknown":
            return "unknown"
        clean_version = v[1:] if v.startswith("v") else v
        if len(clean_version) > max_length:
            return clean_version[:max_length]
        return clean_version

    def format_semver_version(v: str) -> str:
        """Format version as semantic version."""
        if v == "unknown":
            return "0.0.0"
        clean_version = v[1:] if v.startswith("v") else v
        import re

        semver_match = re.match(r"^(\d+\.\d+\.\d+)", clean_version)
        if semver_match:
            return semver_match.group(1)
        return "0.0.0"

    return {
        "MOAI_VERSION": version_for_config,
        "MOAI_VERSION_SHORT": format_short_version(version_for_config),
        "MOAI_VERSION_DISPLAY": format_display_version(version_for_config),
        "MOAI_VERSION_TRIMMED": format_trimmed_version(version_for_config),
        "MOAI_VERSION_SEMVER": format_semver_version(version_for_config),
        "MOAI_VERSION_VALID": "true" if version_for_config != "unknown" else "false",
        "MOAI_VERSION_SOURCE": "config_cached",
        "PROJECT_NAME": project_name,
        "PROJECT_MODE": project_mode,
        "PROJECT_DESCRIPTION": project_description,
        "PROJECT_VERSION": project_version,
        "CREATION_TIMESTAMP": created_at,
        "PROJECT_DIR": hook_project_dir,
        "CONVERSATION_LANGUAGE": language_config.get("conversation_language", "en"),
        "CONVERSATION_LANGUAGE_NAME": language_config.get("conversation_language_name", "English"),
        "AGENT_PROMPT_LANGUAGE": language_config.get("agent_prompt_language", "en"),
        "GIT_COMMIT_MESSAGES_LANGUAGE": language_config.get("git_commit_messages", "en"),
        "CODE_COMMENTS_LANGUAGE": language_config.get("code_comments", "en"),
        "DOCUMENTATION_LANGUAGE": language_config.get(
            "documentation", language_config.get("conversation_language", "en")
        ),
        "ERROR_MESSAGES_LANGUAGE": language_config.get(
            "error_messages", language_config.get("conversation_language", "en")
        ),
        "USER_NAME": user_name,
        "PERSONALIZED_GREETING": personalized_greeting,
        "LANGUAGE_CONFIG_SOURCE": config_source,
        "CODEBASE_LANGUAGE": project_section.get("language", "generic"),
        "PROJECT_OWNER": project_section.get("author", "@user"),
        "AUTHOR": project_section.get("author", "@user"),
        "STATUSLINE_COMMAND": statusline_command,
    }


def _preserve_project_metadata(
    project_path: Path,
    context: dict[str, str],
    existing_config: dict[str, Any],
    version_for_config: str,
) -> None:
    """Restore project-specific metadata in the new config (YAML or JSON).

    Also updates template_version to track which template version is synchronized.
    """
    config_path, _ = _get_config_path(project_path)
    if not config_path.exists():
        return

    try:
        config_data = _load_config(config_path)
    except (json.JSONDecodeError, yaml.YAMLError):
        console.print("[red]✗ Failed to parse config after template copy[/red]")
        return

    project_data = config_data.setdefault("project", {})
    project_data["name"] = context["PROJECT_NAME"]
    project_data["mode"] = context["PROJECT_MODE"]
    project_data["description"] = context["PROJECT_DESCRIPTION"]
    project_data["created_at"] = context["CREATION_TIMESTAMP"]

    if "optimized" not in project_data and isinstance(existing_config, dict):
        existing_project = _extract_project_section(existing_config)
        if isinstance(existing_project, dict) and "optimized" in existing_project:
            project_data["optimized"] = bool(existing_project["optimized"])

    # Preserve locale and language preferences when possible
    existing_project = _extract_project_section(existing_config)
    locale = _coalesce(existing_project.get("locale"), existing_config.get("locale"))
    if locale:
        project_data["locale"] = locale

    language = _coalesce(existing_project.get("language"), existing_config.get("language"))
    if language:
        project_data["language"] = language

    config_data.setdefault("moai", {})
    config_data["moai"]["version"] = version_for_config

    # This allows Stage 2 to compare package vs project template versions
    project_data["template_version"] = version_for_config

    _save_config(config_path, config_data)


def _apply_context_to_file(processor: TemplateProcessor, target_path: Path) -> None:
    """Apply the processor context to an existing file (post-merge pass)."""
    if not processor.context or not target_path.exists():
        return

    try:
        content = target_path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return

    substituted, warnings = processor._substitute_variables(content)  # pylint: disable=protected-access
    if warnings:
        console.print("[yellow]⚠ Template warnings:[/yellow]")
        for warning in warnings:
            console.print(f"   {warning}")

    target_path.write_text(substituted, encoding="utf-8")


def _validate_template_substitution(project_path: Path) -> None:
    """Validate that all template variables have been properly substituted."""
    import re

    # Files to check for unsubstituted variables
    files_to_check = [
        project_path / ".claude" / "settings.json",
        project_path / "CLAUDE.md",
    ]

    issues_found = []

    for file_path in files_to_check:
        if not file_path.exists():
            continue

        try:
            content = file_path.read_text(encoding="utf-8")
            # Look for unsubstituted template variables
            unsubstituted = re.findall(r"\{\{([A-Z_]+)\}\}", content)
            if unsubstituted:
                unique_vars = sorted(set(unsubstituted))
                issues_found.append(f"{file_path.relative_to(project_path)}: {', '.join(unique_vars)}")
        except Exception as e:
            console.print(f"[yellow]⚠️ Could not validate {file_path.relative_to(project_path)}: {e}[/yellow]")

    if issues_found:
        console.print("[red]✗ Template substitution validation failed:[/red]")
        for issue in issues_found:
            console.print(f"   {issue}")
        console.print("[yellow]💡 Run '/moai:0-project' to fix template variables[/yellow]")
    else:
        console.print("[green]✅ Template substitution validation passed[/green]")


def _validate_template_substitution_with_rollback(project_path: Path, backup_path: Path | None) -> bool:
    """Validate template substitution with rollback capability.

    Returns:
        True if validation passed, False if failed (rollback handled by caller)
    """
    import re

    # Files to check for unsubstituted variables
    files_to_check = [
        project_path / ".claude" / "settings.json",
        project_path / "CLAUDE.md",
    ]

    issues_found = []

    for file_path in files_to_check:
        if not file_path.exists():
            continue

        try:
            content = file_path.read_text(encoding="utf-8")
            # Look for unsubstituted template variables
            unsubstituted = re.findall(r"\{\{([A-Z_]+)\}\}", content)
            if unsubstituted:
                unique_vars = sorted(set(unsubstituted))
                issues_found.append(f"{file_path.relative_to(project_path)}: {', '.join(unique_vars)}")
        except Exception as e:
            console.print(f"[yellow]⚠️ Could not validate {file_path.relative_to(project_path)}: {e}[/yellow]")

    if issues_found:
        console.print("[red]✗ Template substitution validation failed:[/red]")
        for issue in issues_found:
            console.print(f"   {issue}")

        if backup_path:
            console.print("[yellow]🔄 Rolling back due to validation failure...[/yellow]")
        else:
            console.print("[yellow]💡 Run '/moai:0-project' to fix template variables[/yellow]")
            console.print("[red]⚠️ No backup available - manual fix required[/red]")

        return False
    else:
        console.print("[green]✅ Template substitution validation passed[/green]")
        return True


def _show_version_info(current: str, latest: str) -> None:
    """Display version information.

    Args:
        current: Current installed version
        latest: Latest available version
    """
    console.print("[cyan]🔍 Checking versions...[/cyan]")
    console.print(f"   Current version: {current}")
    console.print(f"   Latest version:  {latest}")


def _show_installer_not_found_help() -> None:
    """Show help when installer not found."""
    console.print("[red]❌ Cannot detect package installer[/red]\n")
    console.print("Installation method not detected. To update manually:\n")
    console.print("  • If installed via uv tool:")
    console.print("    [cyan]uv tool upgrade moai-adk[/cyan]\n")
    console.print("  • If installed via pipx:")
    console.print("    [cyan]pipx upgrade moai-adk[/cyan]\n")
    console.print("  • If installed via pip:")
    console.print("    [cyan]pip install --upgrade moai-adk[/cyan]\n")
    console.print("Then run:")
    console.print("  [cyan]moai-adk update --templates-only[/cyan]")


def _show_upgrade_failure_help(installer_cmd: list[str]) -> None:
    """Show help when upgrade fails.

    Args:
        installer_cmd: The installer command that failed
    """
    console.print("[red]❌ Upgrade failed[/red]\n")
    console.print("Troubleshooting:")
    console.print("  1. Check network connection")
    console.print(f"  2. Clear cache: {installer_cmd[0]} cache clean")
    console.print(f"  3. Try manually: {' '.join(installer_cmd)}")
    console.print("  4. Report issue: https://github.com/modu-ai/moai-adk/issues")


def _show_network_error_help() -> None:
    """Show help for network errors."""
    console.print("[yellow]⚠️  Cannot reach PyPI to check latest version[/yellow]\n")
    console.print("Options:")
    console.print("  1. Check network connection")
    console.print("  2. Try again with: [cyan]moai-adk update --force[/cyan]")
    console.print("  3. Skip version check: [cyan]moai-adk update --templates-only[/cyan]")


def _show_template_sync_failure_help() -> None:
    """Show help when template sync fails."""
    console.print("[yellow]⚠️  Template sync failed[/yellow]\n")
    console.print("Rollback options:")
    console.print("  1. Restore from backup: [cyan]cp -r .moai-backups/TIMESTAMP .moai/[/cyan]")
    console.print("  2. Skip backup and retry: [cyan]moai-adk update --force[/cyan]")
    console.print("  3. Report issue: https://github.com/modu-ai/moai-adk/issues")


def _show_timeout_error_help() -> None:
    """Show help for timeout errors."""
    console.print("[red]❌ Error: Operation timed out[/red]\n")
    console.print("Try again with:")
    console.print("  [cyan]moai-adk update --yes --force[/cyan]")


def _execute_migration_if_needed(project_path: Path, yes: bool = False) -> bool:
    """Check and execute migration if needed.

    Args:
        project_path: Project directory path
        yes: Auto-confirm without prompting

    Returns:
        True if no migration needed or migration succeeded, False if migration failed
    """
    try:
        migrator = VersionMigrator(project_path)

        # Check if migration is needed
        if not migrator.needs_migration():
            return True

        # Get migration info
        info = migrator.get_migration_info()
        console.print("\n[cyan]🔄 Migration Required[/cyan]")
        console.print(f"   Current version: {info['current_version']}")
        console.print(f"   Target version:  {info['target_version']}")
        console.print(f"   Files to migrate: {info['file_count']}")
        console.print()
        console.print("   This will migrate configuration files to new locations:")
        console.print("   • .moai/config.json → .moai/config/config.json")
        console.print("   • .claude/statusline-config.yaml → .moai/config/statusline-config.yaml")
        console.print()
        console.print("   A backup will be created automatically.")
        console.print()

        # Confirm with user (unless --yes)
        if not yes:
            if not click.confirm("Do you want to proceed with migration?", default=True):
                console.print("[yellow]⚠️  Migration skipped. Some features may not work correctly.[/yellow]")
                console.print("[cyan]💡 Run 'moai-adk migrate' manually when ready[/cyan]")
                return False

        # Execute migration
        console.print("[cyan]🚀 Starting migration...[/cyan]")
        success = migrator.migrate_to_v024(dry_run=False, cleanup=True)

        if success:
            console.print("[green]✅ Migration completed successfully![/green]")
            return True
        else:
            console.print("[red]❌ Migration failed[/red]")
            console.print("[cyan]💡 Use 'moai-adk migrate --rollback' to restore from backup[/cyan]")
            return False

    except Exception as e:
        console.print(f"[red]❌ Migration error: {e}[/red]")
        logger.error(f"Migration failed: {e}", exc_info=True)
        return False


@click.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="Project path (default: current directory)",
)
@click.option("--force", is_flag=True, help="Skip backup and force the update")
@click.option("--check", is_flag=True, help="Only check version (do not update)")
@click.option("--templates-only", is_flag=True, help="Skip package upgrade, sync templates only")
@click.option("--yes", is_flag=True, help="Auto-confirm all prompts (CI/CD mode)")
@click.option(
    "--merge",
    "merge_strategy",
    flag_value="auto",
    help="Auto-merge: Apply template + preserve user changes",
)
@click.option(
    "--manual",
    "merge_strategy",
    flag_value="manual",
    help="Manual merge: Preserve backup, generate merge guide",
)
def update(
    path: str,
    force: bool,
    check: bool,
    templates_only: bool,
    yes: bool,
    merge_strategy: str | None,
) -> None:
    """Update command with 3-stage workflow + merge strategy selection (v0.26.0+).

    Stage 1 (Package Version Check):
    - Fetches current and latest versions from PyPI
    - If current < latest: detects installer (uv tool, pipx, pip) and upgrades package
    - Prompts user to re-run after upgrade completes

    Stage 2 (Config Version Comparison - NEW in v0.6.3):
    - Compares package template_version with project config.json template_version
    - If versions match: skips Stage 3 (already up-to-date)
    - Performance improvement: 70-80% faster for unchanged projects (3-4s vs 12-18s)

    Stage 3 (Template Sync with Merge Strategy - NEW in v0.26.0):
    - Syncs templates only if versions differ
    - User chooses merge strategy:
      * Auto-merge (default): Template + preserved user changes
      * Manual merge: Backup + comprehensive merge guide (full control)
    - Updates .claude/, .moai/, CLAUDE.md, config.json
    - Preserves specs and reports
    - Saves new template_version to config.json

    Examples:
        python -m moai_adk update                    # interactive merge strategy selection
        python -m moai_adk update --merge            # auto-merge (template + user changes)
        python -m moai_adk update --manual           # manual merge (backup + guide)
        python -m moai_adk update --force            # force template sync (no backup)
        python -m moai_adk update --check            # check version only
        python -m moai_adk update --templates-only   # skip package upgrade
        python -m moai_adk update --yes              # CI/CD mode (auto-confirm + auto-merge)

    Merge Strategies:
        --merge:  Auto-merge applies template + preserves your changes (default)
                  Generated files: backup, merge report
        --manual: Manual merge preserves backup + generates comprehensive guide
                  Generated files: backup, merge guide

    Generated Files:
        - Backup: .moai-backups/pre-update-backup_{timestamp}/
        - Report: .moai/reports/merge-report.md (auto-merge only)
        - Guide:  .moai/guides/merge-guide.md (manual merge only)
    """
    try:
        project_path = Path(path).resolve()

        # Verify the project is initialized
        if not (project_path / ".moai").exists():
            console.print("[yellow]⚠ Project not initialized[/yellow]")
            raise click.Abort()

        # Get versions (needed for --check and normal workflow, but not for --templates-only alone)
        # Note: If --check is used, always fetch versions even if --templates-only is also present
        if check or not templates_only:
            try:
                # Try to use new spinner UI
                try:
                    from moai_adk.cli.ui.progress import SpinnerContext

                    with SpinnerContext("Checking for updates...") as spinner:
                        current = _get_current_version()
                        spinner.update("Fetching latest version from PyPI...")
                        latest = _get_latest_version()
                        spinner.success("Version check complete")
                except ImportError:
                    # Fallback to simple console output
                    console.print("[dim]Checking for updates...[/dim]")
                    current = _get_current_version()
                    latest = _get_latest_version()
            except RuntimeError as e:
                console.print(f"[red]Error: {e}[/red]")
                if not force:
                    console.print("[yellow]⚠ Cannot check for updates. Use --force to update anyway.[/yellow]")
                    raise click.Abort()
                # With --force, proceed to Stage 2 even if version check fails
                current = __version__
                latest = __version__

            _show_version_info(current, latest)

        # Step 1: Handle --check (preview mode, no changes) - takes priority
        if check:
            comparison = _compare_versions(current, latest)
            if comparison < 0:
                console.print(f"\n[yellow]📦 Update available: {current} → {latest}[/yellow]")
                console.print("   Run 'moai-adk update' to upgrade")
            elif comparison == 0:
                console.print(f"[green]✓ Already up to date ({current})[/green]")
            else:
                console.print(f"[cyan]ℹ️  Dev version: {current} (latest: {latest})[/cyan]")
            return

        # Step 2: Handle --templates-only (skip upgrade, go straight to sync)
        if templates_only:
            console.print("[cyan]📄 Syncing templates only...[/cyan]")

            # Preserve user-specific settings before sync
            console.print("   [cyan]💾 Preserving user settings...[/cyan]")
            preserved_settings = _preserve_user_settings(project_path)

            try:
                if not _sync_templates(project_path, force, yes):
                    raise TemplateSyncError("Template sync returned False")
            except TemplateSyncError:
                console.print("[red]Error: Template sync failed[/red]")
                _show_template_sync_failure_help()
                raise click.Abort()
            except Exception as e:
                console.print(f"[red]Error: Template sync failed - {e}[/red]")
                _show_template_sync_failure_help()
                raise click.Abort()

            # Restore user-specific settings after sync
            _restore_user_settings(project_path, preserved_settings)

            console.print("   [green]✅ .claude/ update complete[/green]")
            console.print("   [green]✅ .moai/ update complete (specs/reports preserved)[/green]")
            console.print("   [green]🔄 CLAUDE.md merge complete[/green]")
            console.print("   [green]🔄 config.json merge complete[/green]")
            console.print("\n[green]✓ Template sync complete![/green]")
            return

        # Compare versions
        comparison = _compare_versions(current, latest)

        # Stage 1: Package Upgrade (if current < latest)
        if comparison < 0:
            console.print(f"\n[cyan]📦 Upgrading: {current} → {latest}[/cyan]")

            # Confirm upgrade (unless --yes)
            if not yes:
                if not click.confirm(f"Upgrade {current} → {latest}?", default=True):
                    console.print("Cancelled")
                    return

            # Detect installer
            try:
                installer_cmd = _detect_tool_installer()
                if not installer_cmd:
                    raise InstallerNotFoundError("No package installer detected")
            except InstallerNotFoundError:
                _show_installer_not_found_help()
                raise click.Abort()

            # Display upgrade command
            console.print(f"Running: {' '.join(installer_cmd)}")

            # Execute upgrade with timeout handling
            try:
                upgrade_result = _execute_upgrade(installer_cmd)
                if not upgrade_result:
                    raise UpgradeError(f"Upgrade command failed: {' '.join(installer_cmd)}")
            except subprocess.TimeoutExpired:
                _show_timeout_error_help()
                raise click.Abort()
            except UpgradeError:
                _show_upgrade_failure_help(installer_cmd)
                raise click.Abort()

            # Prompt re-run
            console.print("\n[green]✓ Upgrade complete![/green]")
            console.print("[cyan]📢 Run 'moai-adk update' again to sync templates[/cyan]")
            return

        # Stage 1.5: Migration Check (NEW in v0.24.0)
        console.print(f"✓ Package already up to date ({current})")

        # Execute migration if needed
        if not _execute_migration_if_needed(project_path, yes):
            console.print("[yellow]⚠️  Update continuing without migration[/yellow]")
            console.print("[cyan]💡 Some features may require migration to work correctly[/cyan]")

        # Migrate config.json → config.yaml (v0.32.0+)
        console.print("\n[cyan]🔍 Checking for config format migration...[/cyan]")
        if not _migrate_config_json_to_yaml(project_path):
            console.print("[yellow]⚠️  Config migration failed, continuing with existing format[/yellow]")

        # Stage 2: Config Version Comparison
        try:
            package_config_version = _get_package_config_version()
            project_config_version = _get_project_config_version(project_path)
        except ValueError as e:
            console.print(f"[yellow]⚠ Warning: {e}[/yellow]")
            # On version detection error, proceed with template sync (safer choice)
            package_config_version = __version__
            project_config_version = "0.0.0"

        console.print("\n[cyan]🔍 Comparing config versions...[/cyan]")
        console.print(f"   Package template: {package_config_version}")
        console.print(f"   Project config:   {project_config_version}")

        try:
            config_comparison = _compare_versions(package_config_version, project_config_version)
        except version.InvalidVersion as e:
            # Handle invalid version strings (e.g., unsubstituted template placeholders, corrupted configs)
            console.print(f"[yellow]⚠ Invalid version format in config: {e}[/yellow]")
            console.print("[cyan]ℹ️  Forcing template sync to repair configuration...[/cyan]")
            # Force template sync by treating project version as outdated
            config_comparison = 1  # package_config_version > project_config_version

        # If versions are equal, no sync needed
        if config_comparison <= 0:
            console.print(f"\n[green]✓ Project already has latest template version ({project_config_version})[/green]")
            console.print("[cyan]ℹ️  Templates are up to date! No changes needed.[/cyan]")
            return

        # Stage 3: Template Sync (Only if package_config_version > project_config_version)
        console.print(f"\n[cyan]📄 Syncing templates ({project_config_version} → {package_config_version})...[/cyan]")

        # Determine merge strategy (default: auto-merge)
        final_merge_strategy = merge_strategy or "auto"

        # Handle merge strategy
        if final_merge_strategy == "manual":
            # Manual merge mode: Create full backup + generate guide, no template sync
            console.print("\n[cyan]🔀 Manual merge mode selected[/cyan]")

            # Create full project backup
            console.print("   [cyan]💾 Creating full project backup...[/cyan]")
            try:
                from moai_adk.core.migration.backup_manager import BackupManager

                backup_manager = BackupManager(project_path)
                full_backup_path = backup_manager.create_full_project_backup(description="pre-update-backup")
                console.print(f"   [green]✓ Backup: {full_backup_path.relative_to(project_path)}/[/green]")

                # Generate merge guide
                console.print("   [cyan]📋 Generating merge guide...[/cyan]")
                template_path = Path(__file__).parent.parent.parent / "templates"
                guide_path = _generate_manual_merge_guide(full_backup_path, template_path, project_path)
                console.print(f"   [green]✓ Guide: {guide_path.relative_to(project_path)}[/green]")

                # Summary
                console.print("\n[green]✓ Manual merge setup complete![/green]")
                console.print(f"[cyan]📍 Backup location: {full_backup_path.relative_to(project_path)}/[/cyan]")
                console.print(f"[cyan]📋 Merge guide: {guide_path.relative_to(project_path)}[/cyan]")
                console.print("\n[yellow]⚠️  Next steps:[/yellow]")
                console.print("[yellow]  1. Review the merge guide[/yellow]")
                console.print("[yellow]  2. Compare files using diff or visual tools[/yellow]")
                console.print("[yellow]  3. Manually merge your customizations[/yellow]")
                console.print("[yellow]  4. Test and commit changes[/yellow]")

            except Exception as e:
                console.print(f"[red]Error: Manual merge setup failed - {e}[/red]")
                raise click.Abort()

            return

        # Auto merge mode: Preserve user-specific settings before sync
        console.print("\n[cyan]🔀 Auto-merge mode selected[/cyan]")
        console.print("   [cyan]💾 Preserving user settings...[/cyan]")
        preserved_settings = _preserve_user_settings(project_path)

        # Create backup unless --force
        if not force:
            console.print("   [cyan]💾 Creating backup...[/cyan]")
            try:
                processor = TemplateProcessor(project_path)
                backup_path = processor.create_backup()
                console.print(f"   [green]✓ Backup: {backup_path.relative_to(project_path)}/[/green]")
            except Exception as e:
                console.print(f"   [yellow]⚠ Backup failed: {e}[/yellow]")
                console.print("   [yellow]⚠ Continuing without backup...[/yellow]")
        else:
            console.print("   [yellow]⚠ Skipping backup (--force)[/yellow]")

        # Sync templates (NO spinner - user interaction may be required)
        # SpinnerContext blocks stdin, causing hang when click.confirm() is called
        try:
            console.print("   [cyan]Syncing templates...[/cyan]")
            if not _sync_templates(project_path, force, yes):
                raise TemplateSyncError("Template sync returned False")
            _restore_user_settings(project_path, preserved_settings)
            console.print("   [green]✓ Template sync complete[/green]")
        except TemplateSyncError:
            console.print("[red]Error: Template sync failed[/red]")
            _show_template_sync_failure_help()
            raise click.Abort()
        except Exception as e:
            console.print(f"[red]Error: Template sync failed - {e}[/red]")
            _show_template_sync_failure_help()
            raise click.Abort()

        console.print("   [green]✅ .claude/ update complete[/green]")
        console.print("   [green]✅ .moai/ update complete (specs/reports preserved)[/green]")
        console.print("   [green]🔄 CLAUDE.md merge complete[/green]")
        console.print("   [green]🔄 config.json merge complete[/green]")
        console.print("   [yellow]⚙️  Set optimized=false (optimization needed)[/yellow]")

        console.print("\n[green]✓ Update complete![/green]")
        console.print("[cyan]ℹ️  Next step: Run /moai:0-project update to optimize template changes[/cyan]")

    except Exception as e:
        console.print(f"[red]✗ Update failed: {e}[/red]")
        raise click.ClickException(str(e)) from e


def _handle_custom_element_restoration(project_path: Path, backup_path: Path | None, yes: bool = False) -> None:
    """Handle custom element restoration using the enhanced system.

    This function provides an improved interface for restoring user-created custom elements
    (agents, commands, skills, hooks) from backup during MoAI-ADK updates.

    Key improvements:
    - Preserves unselected elements (fixes disappearing issue)
    - Only overwrites/creates selected elements from backup
    - Interactive checkbox selection with arrow key navigation
    - Includes all categories (Agents, Commands, Skills, Hooks)

    Args:
        project_path: Path to the MoAI-ADK project directory
        backup_path: Path to the backup directory (None if no backup)
        yes: Whether to automatically accept defaults (non-interactive mode)
    """
    if not backup_path:
        # No backup available, cannot restore
        return

    try:
        # Create scanner to find custom elements in backup (not current project)
        backup_scanner = create_custom_element_scanner(backup_path)

        # Get count of custom elements in backup
        backup_element_count = backup_scanner.get_element_count()

        if backup_element_count == 0:
            # No custom elements found in backup
            console.print("[green]✓ No custom elements found in backup to restore[/green]")
            return

        # Create enhanced user selection UI
        ui = create_user_selection_ui(project_path)

        console.print(f"\n[cyan]🔍 Found {backup_element_count} custom elements in backup[/cyan]")

        # If yes mode is enabled, restore all elements automatically
        if yes:
            console.print(f"[cyan]🔄 Auto-restoring {backup_element_count} custom elements...[/cyan]")
            backup_custom_elements = backup_scanner.scan_custom_elements()
            selected_elements = []

            # Collect all element paths from backup
            for element_type, elements in backup_custom_elements.items():
                if element_type == "skills":
                    for skill in elements:
                        selected_elements.append(str(skill.path))
                else:
                    for element_path in elements:
                        selected_elements.append(str(element_path))
        else:
            # Interactive mode - prompt user for selection using enhanced UI
            selected_elements = ui.prompt_user_selection(backup_available=True)

            if not selected_elements:
                console.print("[yellow]⚠ No elements selected for restoration[/yellow]")
                console.print("[green]✓ All existing custom elements will be preserved[/green]")
                return

            # Confirm selection
            if not ui.confirm_selection(selected_elements):
                console.print("[yellow]⚠ Restoration cancelled by user[/yellow]")
                console.print("[green]✓ All existing custom elements will be preserved[/green]")
                return

        # Perform selective restoration - ONLY restore selected elements
        if selected_elements:
            console.print(f"[cyan]🔄 Restoring {len(selected_elements)} selected elements from backup...[/cyan]")
            restorer = create_selective_restorer(project_path, backup_path)
            success, stats = restorer.restore_elements(selected_elements)

            if success:
                console.print(f"[green]✅ Successfully restored {stats['success']} custom elements[/green]")
                console.print("[green]✓ All unselected elements remain preserved[/green]")
            else:
                console.print(f"[yellow]⚠️ Partial restoration: {stats['success']}/{stats['total']} elements[/yellow]")
                if stats["failed"] > 0:
                    console.print(f"[red]❌ Failed to restore {stats['failed']} elements[/red]")
                console.print("[yellow]⚠️ All other elements remain preserved[/yellow]")
        else:
            console.print("[green]✓ No elements selected, all custom elements preserved[/green]")

    except Exception as e:
        console.print(f"[yellow]⚠️ Custom element restoration failed: {e}[/yellow]")
        logger.warning(f"Custom element restoration error: {e}")
        console.print("[yellow]⚠️ All existing custom elements remain as-is[/yellow]")
        # Don't fail the entire update process, just log the error
        pass


def _cleanup_legacy_presets(project_path: Path) -> None:
    """Remove legacy presets directory entirely.

    This function removes the entire .moai/config/presets/ directory as it is
    no longer used. All preset settings are now consolidated in sections/git-strategy.yaml.

    Args:
        project_path: Project directory path (absolute)
    """
    import shutil

    presets_dir = project_path / ".moai" / "config" / "presets"

    if not presets_dir.exists() or not presets_dir.is_dir():
        return

    try:
        # Remove entire presets directory (no longer needed)
        shutil.rmtree(presets_dir)
        console.print("   [cyan]🧹 Removed legacy presets directory (now in sections/git-strategy.yaml)[/cyan]")
        logger.info(f"Removed legacy presets directory: {presets_dir}")
    except Exception as e:
        logger.warning(f"Failed to remove legacy presets directory {presets_dir}: {e}")


def _migrate_config_json_to_yaml(project_path: Path) -> bool:
    """Migrate legacy config.json to config.yaml format.

    This function:
    1. Checks if config.json exists
    2. Converts it to config.yaml using YAML format
    3. Removes the old config.json file
    4. Also migrates preset files from JSON to YAML

    Args:
        project_path: Project directory path (absolute)

    Returns:
        bool: True if migration successful or not needed, False on error
    """
    try:
        import yaml
    except ImportError:
        console.print("   [yellow]⚠️ PyYAML not available, skipping config migration[/yellow]")
        return True  # Not a critical error

    config_dir = project_path / ".moai" / "config"
    json_path = config_dir / "config.json"
    yaml_path = config_dir / "config.yaml"

    # Check if migration needed
    if not json_path.exists():
        # No JSON file, migration not needed
        return True

    if yaml_path.exists():
        # YAML already exists, just remove JSON
        try:
            json_path.unlink()
            console.print("   [cyan]🔄 Removed legacy config.json (YAML version exists)[/cyan]")
            logger.info(f"Removed legacy config.json: {json_path}")
            return True
        except Exception as e:
            console.print(f"   [yellow]⚠️ Failed to remove legacy config.json: {e}[/yellow]")
            logger.warning(f"Failed to remove {json_path}: {e}")
            return True  # Not critical

    # Perform migration
    try:
        # Read JSON config
        with open(json_path, "r", encoding="utf-8") as f:
            config_data = json.load(f)

        # Write YAML config
        with open(yaml_path, "w", encoding="utf-8") as f:
            yaml.safe_dump(
                config_data,
                f,
                default_flow_style=False,
                allow_unicode=True,
                sort_keys=False,
            )

        # Remove old JSON file
        json_path.unlink()

        console.print("   [green]✓ Migrated config.json → config.yaml[/green]")
        logger.info(f"Migrated config from JSON to YAML: {json_path} → {yaml_path}")

        # Migrate preset files if they exist
        _migrate_preset_files_to_yaml(config_dir)

        return True

    except Exception as e:
        console.print(f"   [red]✗ Config migration failed: {e}[/red]")
        logger.error(f"Failed to migrate config.json to YAML: {e}")
        return False


def _migrate_preset_files_to_yaml(config_dir: Path) -> None:
    """Migrate preset files from JSON to YAML format.

    Args:
        config_dir: .moai/config directory path
    """
    try:
        import yaml
    except ImportError:
        return

    presets_dir = config_dir / "presets"
    if not presets_dir.exists():
        return

    migrated_count = 0
    for json_file in presets_dir.glob("*.json"):
        yaml_file = json_file.with_suffix(".yaml")

        # Skip if YAML already exists
        if yaml_file.exists():
            # Just remove the JSON file
            try:
                json_file.unlink()
                migrated_count += 1
            except Exception as e:
                logger.warning(f"Failed to remove {json_file}: {e}")
            continue

        # Migrate JSON → YAML
        try:
            with open(json_file, "r", encoding="utf-8") as f:
                preset_data = json.load(f)

            with open(yaml_file, "w", encoding="utf-8") as f:
                yaml.safe_dump(
                    preset_data,
                    f,
                    default_flow_style=False,
                    allow_unicode=True,
                    sort_keys=False,
                )

            json_file.unlink()
            migrated_count += 1

        except Exception as e:
            logger.warning(f"Failed to migrate preset {json_file}: {e}")

    if migrated_count > 0:
        console.print(f"   [cyan]🔄 Migrated {migrated_count} preset file(s) to YAML[/cyan]")
        logger.info(f"Migrated {migrated_count} preset files to YAML")
