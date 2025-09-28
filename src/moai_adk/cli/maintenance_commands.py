#!/usr/bin/env python3
# @REQ:CLI-MAINTENANCE-001
"""
Maintenance Commands for MoAI-ADK

Handles system maintenance operations including health checks and backup restoration.
These commands help diagnose issues and recover from problematic states.
"""

import shutil
import sys
from pathlib import Path

import click
from colorama import Fore, Style

from ..install.post_install import auto_install_on_first_run
from ..utils.logger import get_logger
from .helpers import validate_environment

logger = get_logger(__name__)


@click.command()
@click.argument("backup_path", type=click.Path(exists=True))
@click.option(
    "--dry-run", is_flag=True, help="Show what would be restored without making changes"
)
def restore_command(backup_path: str, dry_run: bool) -> None:
    """Restore MoAI-ADK from a backup directory."""
    backup_dir = Path(backup_path)

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


@click.command()
@click.option("--list-backups", "-l", is_flag=True, help="List available backups")
def doctor_command(list_backups: bool) -> None:
    """Diagnose common issues and check system health."""
    click.echo(f"{Fore.CYAN}🔍 MoAI-ADK Health Check{Style.RESET_ALL}")

    if list_backups:
        _list_available_backups()
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


def _list_available_backups() -> None:
    """List available backup directories in current location."""
    current_dir = Path.cwd()
    backup_dirs = list(current_dir.glob(".moai_backup_*"))

    if backup_dirs:
        click.echo(f"\n📦 Available backups in {current_dir}:")
        for backup_dir in sorted(backup_dirs):
            timestamp = backup_dir.name.split("_", 2)[-1]
            click.echo(f"  {backup_dir.name} (created: {timestamp})")
    else:
        click.echo(f"\n📦 No backups found in {current_dir}")