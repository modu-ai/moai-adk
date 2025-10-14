# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""moai init 명령어

프로젝트 초기화 명령어:
- .moai/ 디렉토리 구조 생성
- config.json 초기화
- 템플릿 파일 복사
"""

from pathlib import Path

import click
from rich.console import Console

from moai_adk.core.project.initializer import initialize_project

console = Console()


@click.command()
@click.argument("path", type=click.Path(), default=".")
def init(path: str) -> None:
    """Initialize a new MoAI-ADK project

    Args:
        path: Project directory path (default: current directory)
    """
    try:
        console.print(f"[cyan]Initializing project at {path}...[/cyan]")

        # 프로젝트 초기화
        project_path = Path(path).resolve()
        initialize_project(project_path)

        console.print("[green]✓ Project initialized successfully[/green]")
        console.print(f"[dim]  └─ Created .moai/ directory at {project_path}[/dim]")

    except FileExistsError as e:
        console.print("[yellow]⚠ Project already initialized[/yellow]")
        raise click.Abort() from e
    except Exception as e:
        console.print(f"[red]✗ Initialization failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
