"""Update command"""
import json
from pathlib import Path

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
        from typing import cast

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

        # Phase 1: check versions
        console.print("[cyan]🔍 Checking versions...[/cyan]")
        current_version = __version__
        latest_version = get_latest_version()

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

            # Don't update if current version is newer or equal
            if current_ver >= latest_ver:
                console.print("[green]✓ Already up to date[/green]")
                return

            # If new version available, guide user to upgrade package first
            if current_ver < latest_ver:
                console.print(f"\n[yellow]⚠ New version available: {latest_version}[/yellow]")
                console.print("[yellow]📦 Please upgrade the package first:[/yellow]")
                console.print("\n   [cyan]# Recommended (uv tool):[/cyan]")
                console.print("   uv tool upgrade moai-adk")
                console.print("\n   [cyan]# Alternative (uv pip):[/cyan]")
                console.print("   uv pip install --upgrade moai-adk")
                console.print("\n   [cyan]# Alternative (pip):[/cyan]")
                console.print("   pip install --upgrade moai-adk")
                console.print(f"\n[dim]💡 After upgrade, run 'moai-adk update' again to update templates.[/dim]")
                return

        # Phase 2: create a backup unless --force
        if not force:
            console.print("\n[cyan]💾 Creating backup...[/cyan]")
            processor = TemplateProcessor(project_path)
            backup_path = processor.create_backup()
            console.print(f"[green]✓ Backup completed: {backup_path.relative_to(project_path)}[/green]")
        else:
            console.print("\n[yellow]⚠ Skipping backup (--force)[/yellow]")

        # Phase 3: update templates
        console.print("\n[cyan]📄 Updating templates...[/cyan]")
        processor = TemplateProcessor(project_path)
        processor.copy_templates(backup=False, silent=True)  # Backup already handled

        console.print("   [green]✅ .claude/ update complete[/green]")
        console.print("   [green]✅ .moai/ update complete (specs/reports preserved)[/green]")
        console.print("   [green]🔄 CLAUDE.md merge complete[/green]")
        console.print("   [green]🔄 config.json merge complete[/green]")

        # Phase 4: set optimized=false
        set_optimized_false(project_path)
        console.print("   [yellow]⚙️  Set optimized=false (optimization needed)[/yellow]")

        console.print("\n[green]✓ Update complete![/green]")
        console.print("\n[cyan]ℹ️  Next step: Run /alfred:0-project update to optimize template changes[/cyan]")

    except Exception as e:
        console.print(f"[red]✗ Update failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
