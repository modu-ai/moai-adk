# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""moai doctor 명령어

시스템 진단 명령어:
- Python 버전 확인
- Git 설치 확인
- 프로젝트 구조 검증
- 의존성 체크
"""

import click
from rich.console import Console
from rich.table import Table

from moai_adk.core.project.checker import check_environment

console = Console()


@click.command()
def doctor() -> None:
    """Check system requirements and project health

    Verifies:
    - Python version (>= 3.13)
    - Git installation
    - Project structure (.moai directory)
    - Required dependencies
    """
    try:
        console.print("[cyan]Running system diagnostics...[/cyan]\n")

        # 환경 체크
        results = check_environment()

        # 결과 테이블 생성
        table = Table(show_header=True, header_style="bold magenta")
        table.add_column("Check", style="dim", width=40)
        table.add_column("Status", justify="center")

        for check_name, status in results.items():
            icon = "✓" if status else "✗"
            color = "green" if status else "red"
            table.add_row(check_name, f"[{color}]{icon}[/{color}]")

        console.print(table)

        # 전체 결과 요약
        all_passed = all(results.values())
        if all_passed:
            console.print("\n[green]✓ All checks passed[/green]")
        else:
            console.print("\n[yellow]⚠ Some checks failed[/yellow]")
            console.print("[dim]Run [cyan]moai doctor --help[/cyan] for troubleshooting tips[/dim]")

    except Exception as e:
        console.print(f"[red]✗ Diagnostic failed: {e}[/red]")
        raise
