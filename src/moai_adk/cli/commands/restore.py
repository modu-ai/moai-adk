# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_commands.py
"""moai restore 명령어

백업 복원:
- 최신 백업 복원
- 특정 시간 백업 복원
- 백업 목록 조회
"""


import click
from rich.console import Console

from moai_adk.core.backup import restore_backup

console = Console()


@click.command()
@click.option("--timestamp", help="Specific backup timestamp", default=None)
def restore(timestamp: str | None) -> None:
    """Restore from backup

    Args:
        timestamp: 복원할 백업 타임스탬프 (None이면 최신)
    """
    try:
        if timestamp:
            console.print(f"[cyan]Restoring from {timestamp}...[/cyan]")
        else:
            console.print("[cyan]Restoring from latest backup...[/cyan]")

        restore_backup(timestamp)
        console.print("[green]✓ Restore completed[/green]")

    except FileNotFoundError as e:
        console.print(f"[red]✗ Backup not found: {e}[/red]")
        raise click.Abort() from e
    except Exception as e:
        console.print(f"[red]✗ Error: {e}[/red]")
        raise
