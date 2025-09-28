#!/usr/bin/env python3
# @REQ:CLI-EXECUTOR-001
"""
CLI command execution logic for MoAI-ADK

Contains the implementation logic for basic CLI commands including
init, restore, and doctor commands.
"""

import shutil
import sys
from pathlib import Path

import click
from colorama import Fore, Style

from ..install.post_install import auto_install_on_first_run
from ..utils.logger import get_logger
from .helpers import validate_environment
from .init_helpers import (
    finalize_installation,
    handle_interactive_mode,
    setup_project_directory,
    validate_initialization,
)

logger = get_logger(__name__)


def execute_restore(backup_path: str, dry_run: bool) -> None:
    """Execute restore command logic."""
    backup_dir = Path(backup_path)

    if not backup_dir.exists():
        click.echo(f"{Fore.RED}❌ Backup path does not exist{Style.RESET_ALL}")
        sys.exit(1)

    if not backup_dir.is_dir():
        click.echo(f"{Fore.RED}❌ Backup path must be a directory{Style.RESET_ALL}")
        sys.exit(1)

    # Validate backup directory structure
    required_items = [".moai", ".claude", "CLAUDE.md"]
    missing_items = [
        item for item in required_items if not (backup_dir / item).exists()
    ]

    if missing_items:
        click.echo(
            f"{Fore.YELLOW}⚠️  Warning: Backup may be incomplete. Missing: {', '.join(missing_items)}{Style.RESET_ALL}"
        )

    current_dir = Path.cwd()

    if dry_run:
        click.echo(
            f"{Fore.CYAN}🔍 Dry run - would restore to: {current_dir}{Style.RESET_ALL}"
        )
        for item in required_items:
            source = backup_dir / item
            target = current_dir / item
            if source.exists():
                click.echo(f"  Would restore: {source} → {target}")
    else:
        click.echo(f"{Fore.CYAN}🔄 Restoring backup to: {current_dir}{Style.RESET_ALL}")

        try:
            # Remove existing directories/files
            for item in required_items:
                target = current_dir / item
                if target.exists():
                    if target.is_dir():
                        shutil.rmtree(target)
                    else:
                        target.unlink()

            # Copy from backup
            for item in required_items:
                source = backup_dir / item
                target = current_dir / item
                if source.exists():
                    if source.is_dir():
                        shutil.copytree(source, target)
                    else:
                        shutil.copy2(source, target)
                    click.echo(f"  Restored: {item}")

            click.echo(f"{Fore.GREEN}✅ Backup restored successfully{Style.RESET_ALL}")

        except Exception as e:
            click.echo(f"{Fore.RED}❌ Failed to restore backup: {e}{Style.RESET_ALL}")
            sys.exit(1)


def execute_doctor(list_backups: bool) -> None:
    """Execute doctor command logic."""
    click.echo(f"{Fore.CYAN}🔍 MoAI-ADK Health Check{Style.RESET_ALL}")

    if list_backups:
        current_dir = Path.cwd()
        backup_dirs = list(current_dir.glob(".moai_backup_*"))

        if backup_dirs:
            click.echo(f"\n📦 Available backups in {current_dir}:")
            for backup_dir in sorted(backup_dirs):
                timestamp = backup_dir.name.split("_", 2)[-1]
                click.echo(f"  {backup_dir.name} (created: {timestamp})")
        else:
            click.echo(f"\n📦 No backups found in {current_dir}")
        return

    # Environment checks
    click.echo("\n🔧 Environment:")
    if validate_environment():
        click.echo(f"  {Fore.GREEN}✅ Python version{Style.RESET_ALL}")
    else:
        click.echo(f"  {Fore.RED}❌ Python version issues{Style.RESET_ALL}")

    # Project checks
    current_dir = Path.cwd()
    click.echo(f"\n📂 Project Status ({current_dir}):")

    moai_exists = (current_dir / ".moai").exists()
    claude_exists = (current_dir / ".claude").exists()
    memory_exists = (current_dir / "CLAUDE.md").exists()

    click.echo(f"  MoAI directory: {'✅' if moai_exists else '❌'}")
    click.echo(f"  Claude directory: {'✅' if claude_exists else '❌'}")
    click.echo(f"  Memory file: {'✅' if memory_exists else '❌'}")

    if not any([moai_exists, claude_exists, memory_exists]):
        click.echo("\n💡 This doesn't appear to be a MoAI-ADK project.")
        click.echo("   Run 'moai init' to initialize.")

    # Resource validation
    if auto_install_on_first_run():
        click.echo(f"  {Fore.GREEN}✅ Resources available{Style.RESET_ALL}")
    else:
        click.echo(f"  {Fore.RED}❌ Resource issues detected{Style.RESET_ALL}")


def execute_init(
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
    """Execute init command logic."""
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


