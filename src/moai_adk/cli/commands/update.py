"""Update command"""
from __future__ import annotations

import json
from datetime import datetime
from pathlib import Path
from typing import Any, cast

import click
from packaging import version
from rich.console import Console

from moai_adk import __version__
from moai_adk.core.template.processor import TemplateProcessor

console = Console()


def get_latest_version() -> str | None:
    """Get the latest version from PyPI.

    Returns:
        Latest version string, or None if fetch fails.
    """
    try:
        import urllib.error
        import urllib.request

        url = "https://pypi.org/pypi/moai-adk/json"
        with urllib.request.urlopen(url, timeout=5) as response:  # nosec B310 - URL is hardcoded HTTPS to PyPI API, no user input
            data = json.loads(response.read().decode("utf-8"))
            version_str: str = cast(str, data["info"]["version"])
            return version_str
    except (urllib.error.URLError, json.JSONDecodeError, KeyError, TimeoutError):
        # Return None if PyPI check fails
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
        config_path.write_text(
            json.dumps(config_data, indent=2, ensure_ascii=False) + "\n",
            encoding="utf-8"
        )
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
        existing_config.get("projectName"),
        project_path.name,
    )
    project_mode = _coalesce(
        project_section.get("mode"),
        existing_config.get("mode"),
        default="personal",
    )
    project_description = _coalesce(
        project_section.get("description"),
        existing_config.get("projectDescription"),
        existing_config.get("description"),
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
    """Restore project-specific metadata in the new config.json."""
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
    project_data["moai_adk_version"] = version_for_config

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

    config_path.write_text(
        json.dumps(config_data, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8"
    )


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


@click.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="Project path (default: current directory)"
)
@click.option(
    "--force",
    is_flag=True,
    help="Skip backup and force the update"
)
@click.option(
    "--check",
    is_flag=True,
    help="Only check version (do not update)"
)
def update(path: str, force: bool, check: bool) -> None:
    """Update template files to the latest version.

    Updates include:
    - .claude/ (fully replaced)
    - .moai/ (preserve specs and reports)
    - CLAUDE.md (merged)
    - config.json (smart merge)

    Examples:
        python -m moai_adk update              # update with backup
        python -m moai_adk update --force      # update without backup
        python -m moai_adk update --check      # check version only
    """
    try:
        project_path = Path(path).resolve()

        # Verify the project is initialized
        if not (project_path / ".moai").exists():
            console.print("[yellow]⚠ Project not initialized[/yellow]")
            raise click.Abort()

        existing_config = _load_existing_config(project_path)

        # Phase 1: check versions
        console.print("[cyan]🔍 Checking versions...[/cyan]")
        current_version = __version__
        latest_version = get_latest_version()
        version_for_config = current_version

        # Handle PyPI fetch failure
        if latest_version is None:
            console.print(f"   Current version: {current_version}")
            console.print("   Latest version:  [yellow]Unable to fetch from PyPI[/yellow]")
            if not force:
                console.print("[yellow]⚠ Cannot check for updates. Use --force to update anyway.[/yellow]")
                return
        else:
            console.print(f"   Current version: {current_version}")
            console.print(f"   Latest version:  {latest_version}")

        if check:
            # Exit early when --check is provided
            if latest_version is None:
                console.print("[yellow]⚠ Unable to check for updates[/yellow]")
            elif version.parse(current_version) < version.parse(latest_version):
                console.print("[yellow]⚠ Update available[/yellow]")
            elif version.parse(current_version) > version.parse(latest_version):
                console.print("[green]✓ Development version (newer than PyPI)[/green]")
            else:
                console.print("[green]✓ Already up to date[/green]")
            return

        # Check if update is needed (version only) - skip with --force
        if not force and latest_version is not None:
            current_ver = version.parse(current_version)
            latest_ver = version.parse(latest_version)

            # Don't update if current version is newer
            if current_ver > latest_ver:
                console.print("[green]✓ Development version (newer than PyPI)[/green]")
                return
            # If versions are equal, check if we need to proceed
            elif current_ver == latest_ver:
                # Check if optimized=false (need to update templates)
                config_path = project_path / ".moai" / "config.json"
                if config_path.exists():
                    try:
                        config_data = json.loads(config_path.read_text())
                        is_optimized = config_data.get("project", {}).get("optimized", False)

                        if is_optimized:
                            # Already up to date and optimized - exit silently
                            return
                        else:
                            # Proceed with template update (optimized=false)
                            console.print("[yellow]⚠ Template optimization needed[/yellow]")
                    except (json.JSONDecodeError, KeyError):
                        # If config.json is invalid, proceed with update
                        pass
                else:
                    console.print("[green]✓ Already up to date[/green]")
                    return

        # Phase 2: create a backup unless --force
        if not force:
            console.print("\n[cyan]💾 Creating backup...[/cyan]")
            processor = TemplateProcessor(project_path)
            backup_path = processor.create_backup()
            console.print(f"[green]✓ Backup completed: {backup_path.relative_to(project_path)}/[/green]")
        else:
            console.print("\n[yellow]⚠ Skipping backup (--force)[/yellow]")

        # Phase 3: update templates
        console.print("\n[cyan]📄 Updating templates...[/cyan]")
        processor = TemplateProcessor(project_path)

        context = _build_template_context(project_path, existing_config, version_for_config)
        if context:
            processor.set_context(context)

        processor.copy_templates(backup=False, silent=True)  # Backup already handled

        console.print("   [green]✅ .claude/ update complete[/green]")
        console.print("   [green]✅ .moai/ update complete (specs/reports preserved)[/green]")
        console.print("   [green]🔄 CLAUDE.md merge complete[/green]")
        console.print("   [green]🔄 config.json merge complete[/green]")

        _preserve_project_metadata(project_path, context, existing_config, version_for_config)
        _apply_context_to_file(processor, project_path / "CLAUDE.md")

        # Phase 4: set optimized=false
        set_optimized_false(project_path)
        console.print("   [yellow]⚙️  Set optimized=false (optimization needed)[/yellow]")

        console.print("\n[green]✓ Update complete![/green]")
        if latest_version and version.parse(current_version) < version.parse(latest_version):
            console.print(
                "[yellow]⚠ Python package still on older version. "
                "Run 'pip install --upgrade moai-adk' to upgrade the CLI package.[/yellow]"
            )
        console.print("\n[cyan]ℹ️  Next step: Run /alfred:0-project update to optimize template changes[/cyan]")

    except Exception as e:
        console.print(f"[red]✗ Update failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
