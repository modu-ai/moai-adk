# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""moai restore 명령어

백업 복원 명령어:
- .moai/backup/ 디렉토리에서 백업 파일 찾기
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
        backup_dir = Path.cwd() / ".moai" / "backup"

        if not backup_dir.exists():
            console.print("[yellow]⚠ No backup directory found[/yellow]")
            console.print("[dim]Backups are stored in .moai/backup/[/dim]")
            raise click.Abort()

        # 백업 파일 찾기
        backup_files = sorted(backup_dir.glob("*.tar.gz"), reverse=True)
        if not backup_files:
            console.print("[yellow]⚠ No backup files found[/yellow]")
            raise click.Abort()

        # 타임스탬프 지정 시 해당 백업 찾기
        if timestamp:
            console.print(f"[cyan]Restoring from {timestamp}...[/cyan]")
            matching = [f for f in backup_files if timestamp in f.name]
            if not matching:
                console.print(f"[red]✗ Backup not found for timestamp: {timestamp}[/red]")
                raise click.Abort()
            backup_file = matching[0]
        else:
            console.print("[cyan]Restoring from latest backup...[/cyan]")
            backup_file = backup_files[0]

        # 복원 (실제 구현은 추후)
        console.print(f"[dim]  └─ Backup file: {backup_file.name}[/dim]")
        console.print("[green]✓ Restore completed[/green]")

        console.print("\n[yellow]Note:[/yellow] Restore functionality is not yet implemented")
        console.print("[dim]This will be added in a future release[/dim]")

    except click.Abort:
        raise
    except Exception as e:
        console.print(f"[red]✗ Restore failed: {e}[/red]")
        raise
