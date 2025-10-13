# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli.py
"""CLI 메인 그룹

Click 기반 CLI 진입점:
- @click.group(): 명령어 그룹화
- show_logo(): ASCII 로고 출력
- Rich Console: 터미널 출력
"""

import click
from rich.console import Console

from moai_adk.cli.commands import doctor, init, restore, status

# Rich Console 인스턴스
console = Console()


def show_logo() -> None:
    """MoAI-ADK ASCII 로고 출력

    Rich를 사용하여 색상이 있는 로고를 출력합니다.
    """
    logo = """
    ▶◀ MoAI-ADK v0.3.0
    ───────────────────────────────────
    SPEC-First TDD Development Framework
    """
    console.print(logo, style="bold cyan")


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


# 명령어 등록
cli.add_command(init.init)
cli.add_command(doctor.doctor)
cli.add_command(status.status)
cli.add_command(restore.restore)
