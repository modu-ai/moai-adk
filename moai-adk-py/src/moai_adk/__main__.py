# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_cli.py
"""MoAI-ADK CLI Entry Point

CLI 진입점 구현:
- Click 기반 CLI 프레임워크
- Rich console 터미널 출력
- ASCII 로고 출력
- --version, --help 옵션
"""

import sys

import click
from rich.console import Console

console = Console()


def show_logo() -> None:
    """MoAI-ADK ASCII 로고 출력"""
    console.print("[cyan]▶◀ MoAI-ADK v0.3.0[/cyan]")
    console.print("[dim]SPEC-First TDD Framework with Alfred SuperAgent[/dim]\n")


@click.group(invoke_without_command=True)
@click.version_option(version="0.3.0", prog_name="MoAI-ADK")
@click.pass_context
def cli(ctx: click.Context) -> None:
    """MoAI Agentic Development Kit

    SPEC-First TDD Framework with Alfred SuperAgent
    """
    # 하위 명령어 없이 실행되면 로고 출력
    if ctx.invoked_subcommand is None:
        show_logo()
        console.print(
            "[yellow]Tip:[/yellow] Run [cyan]moai --help[/cyan] to see available commands"
        )


def main() -> int:
    """CLI 진입점"""
    try:
        cli()
        return 0
    except Exception as e:
        console.print(f"[red]Error:[/red] {e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
