#!/usr/bin/env python3
# @REQ:CLI-MODULAR-001
"""
Command Handlers for Core MoAI-ADK Operations

Handles the main project lifecycle commands: init, update, and status.
These commands manage project initialization, resource updates, and status reporting.
"""

import sys
from pathlib import Path

import click
from colorama import Fore, Style

from .._version import __version__
from ..core.resource_version import ResourceVersionManager
from ..core.version_sync import VersionSyncManager
from ..install.resource_manager import ResourceManager
from ..utils.logger import get_logger
from .helpers import (
    create_installation_backup,
    format_project_status,
)
from .init_helpers import (
    finalize_installation,
    handle_interactive_mode,
    setup_project_directory,
    validate_initialization,
)

logger = get_logger(__name__)


@click.command()
@click.argument("project_path", type=click.Path(), default=".")
@click.option(
    "--template",
    "-t",
    default="standard",
    help="Template to use (standard, minimal, advanced)",
)
@click.option("--interactive", "-i", is_flag=True, help="Run interactive setup wizard")
@click.option(
    "--backup",
    "-b",
    is_flag=True,
    help="Create backup before installation (recommended)",
)
@click.option(
    "--force",
    "-f",
    is_flag=True,
    help="Force overwrite existing files (주의: 기존 파일을 덮어씁니다)",
)
@click.option(
    "--force-copy",
    is_flag=True,
    help="Force file copying instead of symlinks (recommended for Windows without admin rights)",
)
@click.option("--quiet", "-q", is_flag=True, help="Quiet mode - minimal output")
@click.option(
    "--personal",
    is_flag=True,
    help="Initialize in personal mode (default) - simplified workflow for individual development",
)
@click.option(
    "--team",
    is_flag=True,
    help="Initialize in team mode - full GitFlow with collaboration features",
)
def init_command(
    project_path: str,
    template: str,
    interactive: bool,
    backup: bool,
    force: bool,
    force_copy: bool,
    quiet: bool,
    personal: bool,
    team: bool,
) -> None:
    """Initialize a new MoAI-ADK project."""

    # Step 1: Validate initialization parameters
    project_dir, project_mode = validate_initialization(
        project_path, personal, team, quiet
    )

    # Step 2: Handle interactive mode if requested
    if handle_interactive_mode(project_dir, interactive):
        return

    # Step 3: Setup project directory structure
    if not setup_project_directory(
        project_dir, project_mode, backup, force, force_copy, quiet
    ):
        return

    # Step 4: Finalize installation
    finalize_installation(project_dir, project_mode, force_copy, quiet, force, backup)


@click.command()
@click.option("--verbose", "-v", is_flag=True, help="Show detailed status information")
@click.option(
    "--project-path",
    "-p",
    type=click.Path(exists=True),
    help="Path to project directory (default: current directory)",
)
def status_command(verbose: bool, project_path: str | None) -> None:
    """Show MoAI-ADK project status."""
    target_path = Path(project_path) if project_path else Path.cwd()

    click.echo(f"{Fore.CYAN}📊 MoAI-ADK Project Status{Style.RESET_ALL}")

    status_info = format_project_status(target_path)

    click.echo(f"\n📂 Project: {status_info['path']}")
    click.echo(f"   Type: {status_info['project_type']}")

    # Core status
    click.echo("\n🗿 MoAI-ADK Components:")
    click.echo(f"   MoAI System: {'✅' if status_info['moai_initialized'] else '❌'}")
    click.echo(
        f"   Claude Integration: {'✅' if status_info['claude_initialized'] else '❌'}"
    )
    click.echo(f"   Memory File: {'✅' if status_info['memory_file'] else '❌'}")
    click.echo(f"   Git Repository: {'✅' if status_info['git_repository'] else '❌'}")

    versions = status_info.get("versions")
    if versions:
        click.echo("\n🧭 Versions:")
        click.echo(f"   Package: v{versions.get('package', 'unknown')}")
        click.echo(f"   Templates: v{versions.get('resources', 'unknown')}")
        if versions.get("available_resources") and versions.get(
            "available_resources"
        ) != versions.get("resources"):
            click.echo(
                f"   Available template update: v{versions['available_resources']}"
            )
        if versions.get("outdated"):
            click.echo(
                f"{Fore.YELLOW}   ⚠️  Templates are outdated. Run 'moai update' to refresh.{Style.RESET_ALL}"
            )

    if verbose and status_info["file_counts"]:
        click.echo("\n📁 File Counts:")
        for component, count in status_info["file_counts"].items():
            click.echo(f"   {component}: {count} files")

    # Recommendations
    recommendations = []
    if not status_info["moai_initialized"]:
        recommendations.append("Run 'moai init' to initialize MoAI-ADK")
    if not status_info["git_repository"]:
        recommendations.append("Initialize Git repository: git init")

    if recommendations:
        click.echo("\n💡 Recommendations:")
        for rec in recommendations:
            click.echo(f"   - {rec}")


@click.command()
@click.option(
    "--check", "-c", is_flag=True, help="Check for updates without installing"
)
@click.option("--no-backup", is_flag=True, help="Skip backup creation before update")
@click.option("--verbose", "-v", is_flag=True, help="Show detailed update information")
@click.option("--package-only", is_flag=True, help="Update only the Python package")
@click.option("--resources-only", is_flag=True, help="Update only project resources")
def update_command(
    check: bool,
    no_backup: bool,
    verbose: bool,
    package_only: bool,
    resources_only: bool,
) -> None:
    """Update MoAI-ADK to the latest version."""
    current_version = __version__
    project_path = Path.cwd()

    resource_manager = ResourceManager()
    version_manager = ResourceVersionManager(project_path)
    version_info = version_manager.read()
    current_resource_version = version_info.get("template_version") or "unknown"
    available_resource_version = resource_manager.get_version()

    if check:
        click.echo(f"{Fore.CYAN}🔍 Checking for updates...{Style.RESET_ALL}")
        click.echo(f"Current version: v{current_version}")
        click.echo(f"Installed template version: {current_resource_version}")
        click.echo(f"Available template version: {available_resource_version}")

        if not (project_path / ".moai").exists():
            click.echo(
                f"{Fore.YELLOW}⚠️  This directory does not appear to be a MoAI-ADK project{Style.RESET_ALL}"
            )
        elif current_resource_version != available_resource_version:
            click.echo(
                f"{Fore.YELLOW}⚠️  Templates are outdated. Run 'moai update' to refresh.{Style.RESET_ALL}"
            )
        else:
            click.echo(
                f"{Fore.GREEN}✅ Project resources are up to date{Style.RESET_ALL}"
            )
        return

    if package_only and resources_only:
        click.echo(
            f"{Fore.RED}❌ Cannot use --package-only and --resources-only together{Style.RESET_ALL}"
        )
        sys.exit(1)

    # Check if this is a MoAI-ADK project
    if not (project_path / ".moai").exists():
        click.echo(
            f"{Fore.YELLOW}⚠️  This doesn't appear to be a MoAI-ADK project{Style.RESET_ALL}"
        )
        click.echo("Run 'moai init' to initialize a new project")
        return

    # Create backup unless explicitly disabled
    if not no_backup:
        click.echo(f"{Fore.CYAN}📦 Creating backup...{Style.RESET_ALL}")
        if not create_installation_backup(project_path):
            click.echo(f"{Fore.RED}❌ Backup failed - aborting update{Style.RESET_ALL}")
            sys.exit(1)
        click.echo(f"{Fore.GREEN}✅ Backup created{Style.RESET_ALL}")

    try:
        if not package_only:
            # Update resources
            click.echo(f"{Fore.CYAN}🔄 Updating project resources...{Style.RESET_ALL}")
            resource_manager.copy_claude_resources(project_path, overwrite=True)
            resource_manager.copy_moai_resources(project_path, overwrite=True)
            resource_manager.copy_project_memory(project_path, overwrite=True)
            version_manager.write(available_resource_version, __version__)
            click.echo(f"   Templates updated to v{available_resource_version}")

        if not resources_only:
            # Check if package update is needed
            click.echo(f"{Fore.CYAN}📦 Checking package version...{Style.RESET_ALL}")
            click.echo(f"   Current package version: v{current_version}")

            # For now, we assume the current version is the latest
            # In the future, this could check PyPI for newer versions
            if current_version == available_resource_version:
                click.echo(f"{Fore.GREEN}   ✅ Package is up to date{Style.RESET_ALL}")
            else:
                click.echo(
                    f"{Fore.YELLOW}   💡 Manual upgrade recommended: pip install --upgrade moai-adk{Style.RESET_ALL}"
                )

        # Version synchronization
        if verbose:
            sync_manager = VersionSyncManager(str(project_path))
            click.echo(
                f"{Fore.CYAN}🔄 Synchronizing version information...{Style.RESET_ALL}"
            )
            results = sync_manager.sync_all_versions(dry_run=True)
            for pattern, files in results.items():
                if files:
                    click.echo(f"   {pattern}: {len(files)} files")

        click.echo(f"\n{Fore.GREEN}✅ Update completed successfully{Style.RESET_ALL}")
        click.echo(f"Package version: v{current_version}")
        click.echo(f"Template version: v{available_resource_version}")

    except Exception as e:
        click.echo(f"{Fore.RED}❌ Update failed: {e}{Style.RESET_ALL}")
        if not no_backup:
            click.echo("You can restore from the backup if needed")
        sys.exit(1)