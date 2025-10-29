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

### Related Skills
- **moai-foundation-trust**: For post-update validation
  - Trigger: After updating MoAI-ADK version
  - Invocation: `Skill("moai-foundation-trust")` to verify all toolchains still work

- **moai-foundation-langs**: For language detection after update
  - Trigger: After updating, confirm language stack is intact
  - Invocation: `Skill("moai-foundation-langs")` to re-detect and validate language configuration

### When to Invoke Skills in Related Workflows
1. **After successful update**:
   - Run `Skill("moai-foundation-trust")` to validate all TRUST 5 gates
   - Run `Skill("moai-foundation-langs")` to confirm language toolchain still works
   - Run project doctor command for full system validation

2. **Before updating**:
   - Create backup with `python -m moai_adk backup`
   - Run `Skill("moai-foundation-tags")` to document current TAG state

3. **If update fails**:
   - Use backup to restore previous state
   - Debug with `python -m moai_adk doctor --verbose`
"""

from __future__ import annotations

import json
import logging
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Any, cast

import click
from packaging import version
from rich.console import Console

from moai_adk import __version__
from moai_adk.core.template.processor import TemplateProcessor

console = Console()
logger = logging.getLogger(__name__)

# Constants for tool detection
TOOL_DETECTION_TIMEOUT = 5  # seconds
UV_TOOL_COMMAND = ["uv", "tool", "upgrade", "moai-adk"]
PIPX_COMMAND = ["pipx", "upgrade", "moai-adk"]
PIP_COMMAND = ["pip", "install", "--upgrade", "moai-adk"]


# @CODE:UPDATE-REFACTOR-002-004
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


def _is_installed_via_uv_tool() -> bool:
    """Check if moai-adk installed via uv tool.

    Returns:
        True if uv tool list shows moai-adk, False otherwise
    """
    try:
        result = subprocess.run(
            ["uv", "tool", "list"], capture_output=True, text=True, timeout=TOOL_DETECTION_TIMEOUT, check=False
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
            ["pipx", "list"], capture_output=True, text=True, timeout=TOOL_DETECTION_TIMEOUT, check=False
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
            ["pip", "show", "moai-adk"], capture_output=True, text=True, timeout=TOOL_DETECTION_TIMEOUT, check=False
        )
        return result.returncode == 0
    except (FileNotFoundError, subprocess.TimeoutExpired, OSError):
        return False


# @CODE:UPDATE-REFACTOR-002-001
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


# @CODE:UPDATE-REFACTOR-002-002
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


# @CODE:UPDATE-REFACTOR-002-006
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


# @CODE:UPDATE-REFACTOR-002-007
def _get_project_config_version(project_path: Path) -> str:
    """Get current project config.json template version.

    This reads the project's .moai/config.json to determine the current
    template version that the project is configured with.

    Args:
        project_path: Project directory path (absolute)

    Returns:
        Version string from project's config.json (e.g., "0.6.1")
        Returns "0.0.0" if template_version field not found (indicates no prior sync)

    Raises:
        ValueError: If config.json exists but cannot be parsed
    """

    def _is_placeholder(value: str) -> bool:
        """Check if value contains unsubstituted template placeholders."""
        return isinstance(value, str) and value.startswith("{{") and value.endswith("}}")

    config_path = project_path / ".moai" / "config.json"

    if not config_path.exists():
        # No config yet, treat as version 0.0.0 (needs initial sync)
        return "0.0.0"

    try:
        config_data = json.loads(config_path.read_text(encoding="utf-8"))
        # Check for template_version in project section
        template_version = config_data.get("project", {}).get("template_version")
        if template_version and not _is_placeholder(template_version):
            return template_version

        # Fallback to moai version if no template_version exists
        moai_version = config_data.get("moai", {}).get("version")
        if moai_version and not _is_placeholder(moai_version):
            return moai_version

        # If values are placeholders or don't exist, treat as uninitialized (0.0.0 triggers sync)
        return "0.0.0"
    except json.JSONDecodeError as e:
        raise ValueError(f"Failed to parse project config.json: {e}") from e


# @CODE:UPDATE-CACHE-FIX-001-001-DETECT-STALE
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


# @CODE:UPDATE-CACHE-FIX-001-002-CLEAR-SUCCESS
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


# @CODE:UPDATE-CACHE-FIX-001-003-RETRY-LOGIC
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
                result = subprocess.run(installer_cmd, capture_output=True, text=True, timeout=60, check=False)

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


def _sync_templates(project_path: Path, force: bool = False) -> bool:
    """Sync templates to project.

    Args:
        project_path: Project path (absolute)
        force: Force update without backup

    Returns:
        True if sync succeeded, False otherwise
    """
    try:
        processor = TemplateProcessor(project_path)

        # Load existing config
        existing_config = _load_existing_config(project_path)

        # Build context
        context = _build_template_context(project_path, existing_config, __version__)
        if context:
            processor.set_context(context)

        # Copy templates
        processor.copy_templates(backup=False, silent=True)

        # Preserve metadata
        _preserve_project_metadata(project_path, context, existing_config, __version__)
        _apply_context_to_file(processor, project_path / "CLAUDE.md")

        # Set optimized=false
        set_optimized_false(project_path)

        return True
    except Exception:
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
    """Set config.json's optimized field to false.

    Args:
        project_path: Project path (absolute).
    """
    config_path = project_path / ".moai" / "config.json"
    if not config_path.exists():
        return

    try:
        config_data = json.loads(config_path.read_text(encoding="utf-8"))
        config_data.setdefault("project", {})["optimized"] = False
        config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    except (json.JSONDecodeError, KeyError):
        # Ignore errors if config.json is invalid
        pass


def _load_existing_config(project_path: Path) -> dict[str, Any]:
    """Load existing config.json if available."""
    config_path = project_path / ".moai" / "config.json"
    if config_path.exists():
        try:
            return json.loads(config_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            console.print("[yellow]⚠ Existing config.json could not be parsed. Proceeding with defaults.[/yellow]")
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

    return {
        "MOAI_VERSION": version_for_config,
        "PROJECT_NAME": project_name,
        "PROJECT_MODE": project_mode,
        "PROJECT_DESCRIPTION": project_description,
        "PROJECT_VERSION": project_version,
        "CREATION_TIMESTAMP": created_at,
    }


def _preserve_project_metadata(
    project_path: Path,
    context: dict[str, str],
    existing_config: dict[str, Any],
    version_for_config: str,
) -> None:
    """Restore project-specific metadata in the new config.json.

    Also updates template_version to track which template version is synchronized.
    """
    config_path = project_path / ".moai" / "config.json"
    if not config_path.exists():
        return

    try:
        config_data = json.loads(config_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        console.print("[red]✗ Failed to parse config.json after template copy[/red]")
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

    # @CODE:UPDATE-REFACTOR-002-008: Update template_version to track sync status
    # This allows Stage 2 to compare package vs project template versions
    project_data["template_version"] = version_for_config

    config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


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


# @CODE:UPDATE-REFACTOR-002-003
def _show_version_info(current: str, latest: str) -> None:
    """Display version information.

    Args:
        current: Current installed version
        latest: Latest available version
    """
    console.print("[cyan]🔍 Checking versions...[/cyan]")
    console.print(f"   Current version: {current}")
    console.print(f"   Latest version:  {latest}")


# @CODE:UPDATE-REFACTOR-002-005
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


@click.command()
@click.option("--path", type=click.Path(exists=True), default=".", help="Project path (default: current directory)")
@click.option("--force", is_flag=True, help="Skip backup and force the update")
@click.option("--check", is_flag=True, help="Only check version (do not update)")
@click.option("--templates-only", is_flag=True, help="Skip package upgrade, sync templates only")
@click.option("--yes", is_flag=True, help="Auto-confirm all prompts (CI/CD mode)")
def update(path: str, force: bool, check: bool, templates_only: bool, yes: bool) -> None:
    """Update command with 3-stage workflow (v0.6.3+).

    Stage 1 (Package Version Check):
    - Fetches current and latest versions from PyPI
    - If current < latest: detects installer (uv tool, pipx, pip) and upgrades package
    - Prompts user to re-run after upgrade completes

    Stage 2 (Config Version Comparison - NEW in v0.6.3):
    - Compares package template_version with project config.json template_version
    - If versions match: skips Stage 3 (already up-to-date)
    - Performance improvement: 70-80% faster for unchanged projects (3-4s vs 12-18s)

    Stage 3 (Template Sync):
    - Syncs templates only if versions differ
    - Updates .claude/, .moai/, CLAUDE.md, config.json
    - Preserves specs and reports
    - Saves new template_version to config.json

    Examples:
        python -m moai_adk update                    # auto 3-stage workflow
        python -m moai_adk update --force            # force template sync
        python -m moai_adk update --check            # check version only
        python -m moai_adk update --templates-only   # skip package upgrade
        python -m moai_adk update --yes              # CI/CD mode (auto-confirm)
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
            try:
                if not _sync_templates(project_path, force):
                    raise TemplateSyncError("Template sync returned False")
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
            console.print("\n[green]✓ Template sync complete![/green]")
            return

        # Compare versions
        comparison = _compare_versions(current, latest)

        # Stage 1: Package Upgrade (if current < latest)
        # @CODE:UPDATE-REFACTOR-002-009: Stage 1 - Package version check and upgrade
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

        # Stage 2: Config Version Comparison
        # @CODE:UPDATE-REFACTOR-002-010: Stage 2 - Compare template versions to determine if sync needed
        console.print(f"✓ Package already up to date ({current})")

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
        # @CODE:UPDATE-REFACTOR-002-011: Stage 3 - Template sync only if versions differ
        console.print(f"\n[cyan]📄 Syncing templates ({project_config_version} → {package_config_version})...[/cyan]")

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

        # Sync templates
        try:
            if not _sync_templates(project_path, force):
                raise TemplateSyncError("Template sync returned False")
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
        console.print("[cyan]ℹ️  Next step: Run /alfred:0-project update to optimize template changes[/cyan]")

    except Exception as e:
        console.print(f"[red]✗ Update failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
