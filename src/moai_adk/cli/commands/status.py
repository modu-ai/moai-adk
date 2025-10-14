# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""moai status 명령어

프로젝트 상태 표시 명령어:
- config.json에서 프로젝트 정보 읽기
- SPEC 문서 개수 표시
- Git 상태 요약
"""

import json
from pathlib import Path

import click
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

console = Console()


@click.command()
def status() -> None:
    """Show current project status

    Displays:
    - Project mode (personal/team)
    - Locale setting
    - Number of SPEC documents
    - Git branch and status
    """
    try:
        # config.json 읽기
        config_path = Path.cwd() / ".moai" / "config.json"
        if not config_path.exists():
            console.print("[yellow]⚠ No .moai/config.json found[/yellow]")
            console.print("[dim]Run [cyan]moai init .[/cyan] to initialize project[/dim]")
            raise click.Abort()

        with open(config_path) as f:
            config = json.load(f)

        # SPEC 문서 개수 세기
        specs_dir = Path.cwd() / ".moai" / "specs"
        spec_count = len(list(specs_dir.glob("SPEC-*/spec.md"))) if specs_dir.exists() else 0

        # 상태 정보 테이블
        table = Table(show_header=False, box=None, padding=(0, 2))
        table.add_column("Key", style="cyan")
        table.add_column("Value", style="bold")

        table.add_row("Mode", config.get("mode", "unknown"))
        table.add_row("Locale", config.get("locale", "unknown"))
        table.add_row("SPECs", str(spec_count))

        # Git 정보 추가 (선택적)
        try:
            from git import Repo

            repo = Repo(Path.cwd())
            table.add_row("Branch", repo.active_branch.name)
            table.add_row("Git Status", "Clean" if not repo.is_dirty() else "Modified")
        except Exception:
            pass

        # 패널로 출력
        panel = Panel(
            table,
            title="[bold]Project Status[/bold]",
            border_style="cyan",
            expand=False,
        )

        console.print(panel)

    except click.Abort:
        raise
    except Exception as e:
        console.print(f"[red]✗ Failed to get status: {e}[/red]")
        raise
