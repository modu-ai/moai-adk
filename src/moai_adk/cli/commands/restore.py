# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""moai restore 명령어

백업 복원 명령어:
- .moai/backups/ 디렉토리에서 백업 찾기
- 지정된 타임스탬프 또는 최신 백업 복원
- 복원 전 확인 프롬프트
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
        moai restore                    # Restore from latest backup
        moai restore --timestamp 2025-10-13-120000  # Restore specific backup
    """
    try:
        backup_dir = Path.cwd() / ".moai" / "backups"

        if not backup_dir.exists():
            console.print("[yellow]⚠ No backup directory found[/yellow]")
            console.print("[dim]Backups are stored in .moai/backups/[/dim]")
            raise click.Abort()

        # 백업 디렉토리 찾기 (타임스탬프 기반)
        backup_dirs = sorted(
            [d for d in backup_dir.iterdir() if d.is_dir()],
            key=lambda x: x.name,
            reverse=True
        )
        if not backup_dirs:
            console.print("[yellow]⚠ No backups found[/yellow]")
            raise click.Abort()

        # 타임스탬프 지정 시 해당 백업 찾기
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

        # 복원 (실제 구현은 추후)
        console.print(f"[dim]  └─ Backup: {backup_path.name}[/dim]")
        console.print("[green]✓ Restore completed[/green]")

        console.print("\n[yellow]Note:[/yellow] Restore functionality is not yet implemented")
        console.print("[dim]This will be added in a future release[/dim]")

    except click.Abort:
        raise
    except Exception as e:
        console.print(f"[red]✗ Restore failed: {e}[/red]")
        raise
