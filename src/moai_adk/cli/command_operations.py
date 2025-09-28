#!/usr/bin/env python3
# @REQ:CLI-OPERATIONS-001
"""
CLI command operations for complex commands (status, update)

Contains the implementation logic for complex CLI commands that require
extensive resource management and version handling.
"""

import sys
from pathlib import Path

import click
from colorama import Fore, Style

from .._version import __version__
from ..core.resource_version import ResourceVersionManager
from ..core.version_sync import VersionSyncManager
from ..install.resource_manager import ResourceManager
from .helpers import create_installation_backup, format_project_status


def execute_status(verbose: bool, project_path: str | None) -> None:
    """Execute status command logic."""
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


def execute_update(
    check: bool,
    no_backup: bool,
    verbose: bool,
    package_only: bool,
    resources_only: bool,
) -> None:
    """Execute update command logic."""
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