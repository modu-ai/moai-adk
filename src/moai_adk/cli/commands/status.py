# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_commands.py
"""moai status 명령어

프로젝트 상태 표시:
- 프로젝트 모드 (personal/team)
- Locale 설정
- SPEC 개수
- 최근 작업 내역
"""

import click
from rich.console import Console

from moai_adk.core.project import get_project_status

console = Console()


@click.command()
def status() -> None:
    """Show current project status"""
    try:
        status_data = get_project_status()

        console.print("[bold]Project Status:[/bold]")
        console.print(f"  Mode: {status_data['mode']}")
        console.print(f"  Locale: {status_data['locale']}")
        console.print(f"  SPECs: {status_data['spec_count']}")

    except FileNotFoundError as e:
        console.print(f"[red]✗ Error: {e}[/red]")
        console.print("[yellow]Tip:[/yellow] Run [cyan]moai init .[/cyan] to initialize a project")
        raise click.Abort() from e
    except Exception as e:
        console.print(f"[red]✗ Error: {e}[/red]")
        raise
