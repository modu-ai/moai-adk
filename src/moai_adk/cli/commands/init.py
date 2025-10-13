# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_commands.py
"""moai init 명령어

프로젝트 초기화:
- .moai/ 디렉토리 생성
- config.json, product.md 등 기본 파일 생성
- 템플릿 복사
"""

import click
from rich.console import Console

from moai_adk.core.project import initialize_project

console = Console()


@click.command()
@click.argument("path", type=click.Path(), default=".")
def init(path: str) -> None:
    """Initialize a new MoAI-ADK project

    Args:
        path: 프로젝트 경로 (기본값: 현재 디렉토리)
    """
    try:
        console.print(f"[cyan]Initializing project at {path}...[/cyan]")
        initialize_project(path)
        console.print("[green]✓ Project initialized successfully[/green]")
    except FileNotFoundError as e:
        console.print(f"[red]✗ File not found: {e}[/red]")
        raise click.Abort() from e
    except Exception as e:
        console.print(f"[red]✗ Error: {e}[/red]")
        raise
