# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_commands.py
"""moai doctor 명령어

환경 검증:
- Python 3.14+ 확인
- Git 설치 확인
- uv 설치 확인
- .moai/ 디렉토리 확인
"""

import click
from rich.console import Console

from moai_adk.core.project import check_environment

console = Console()


@click.command()
def doctor() -> None:
    """Check system requirements and project health"""
    try:
        console.print("[cyan]Running system diagnostics...[/cyan]")
        results = check_environment()

        for check, status in results.items():
            icon = "✓" if status else "✗"
            color = "green" if status else "red"
            console.print(f"[{color}]{icon} {check}[/{color}]")

    except Exception as e:
        console.print(f"[red]✗ Error: {e}[/red]")
        raise
