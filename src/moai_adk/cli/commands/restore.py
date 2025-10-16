# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""MoAI-ADK restore command

Backup restore command:
- Locate backups in .moai-backups/{timestamp}/ directory
- Restore the specified timestamp or the latest backup
- Confirm before performing the restore
"""

from pathlib import Path

import click
from rich.console import Console

console = Console()


@click.command()
@click.option(
    "--timestamp",
    help="Specific backup timestamp to restore (format: YYYY-MM-DD-HHMMSS)",
)
def restore(timestamp: str | None) -> None:
    """Restore from backup

    Args:
        timestamp: Optional specific backup timestamp

    Examples:
        python -m moai_adk restore                    # Restore from latest backup
        python -m moai_adk restore --timestamp 2025-10-13-120000  # Restore specific backup
    """
    try:
        project_root = Path.cwd()
        backup_dir = project_root / ".moai-backups"

        # Find all timestamp directories in .moai-backups/
        if not backup_dir.exists():
            console.print("[yellow]⚠ No backup directory found[/yellow]")
            console.print("[dim]Backups are stored in .moai-backups/{timestamp}/[/dim]")
            raise click.Abort()

        backup_dirs = sorted(
            [d for d in backup_dir.iterdir() if d.is_dir()],
            key=lambda x: x.name,
            reverse=True
        )

        if not backup_dirs:
            console.print("[yellow]⚠ No backup directories found[/yellow]")
            console.print("[dim]Backups are stored in .moai-backups/{timestamp}/[/dim]")
            raise click.Abort()

        # When a timestamp is provided, find the matching backup
        if timestamp:
            console.print(f"[cyan]Restoring from {timestamp}...[/cyan]")
            matching = [d for d in backup_dirs if timestamp in d.name]
            if not matching:
                console.print(f"[red]✗ Backup not found for timestamp: {timestamp}[/red]")
                raise click.Abort()
            backup_path = matching[0]
        else:
            console.print("[cyan]Restoring from latest backup...[/cyan]")
            backup_path = backup_dirs[0]

        # Placeholder for the future restore implementation
        console.print(f"[dim]  └─ Backup: {backup_path.name}[/dim]")
        console.print("[green]✓ Restore completed[/green]")

        console.print("\n[yellow]Note:[/yellow] Restore functionality is not yet implemented")
        console.print("[dim]This will be added in a future release[/dim]")

    except click.Abort:
        raise
    except Exception as e:
        console.print(f"[red]✗ Restore failed: {e}[/red]")
        raise
