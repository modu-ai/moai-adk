# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""MoAI-ADK CLI Entry Point

CLI 진입점 구현:
- Click 기반 CLI 프레임워크
- Rich console 터미널 출력
- ASCII 로고 출력
- --version, --help 옵션
- 4개 핵심 명령어: init, doctor, status, restore
"""

import sys

import click
import pyfiglet
from rich.console import Console

from moai_adk import __version__

console = Console()


def show_logo() -> None:
    """MoAI-ADK ASCII 로고 출력 (Pyfiglet)"""
    # Pyfiglet으로 "MoAI-ADK" 텍스트 생성 (ansi_shadow 폰트 사용)
    logo = pyfiglet.figlet_format("MoAI-ADK", font="ansi_shadow")

    # Rich로 스타일 적용하여 출력
    console.print(logo, style="cyan bold", highlight=False)
    console.print("  Modu-AI's Agentic Development Kit w/ SuperAgent 🎩 Alfred", style="yellow bold")
    console.print()
    console.print("  Version: ", style="green", end="")
    console.print(__version__, style="cyan bold")
    console.print()
    console.print("  Tip: Run ", style="yellow", end="")
    console.print("moai-adk --help", style="cyan", end="")
    console.print(" to see available commands", style="yellow")


@click.group(invoke_without_command=True)
@click.version_option(version=__version__, prog_name="MoAI-ADK")
@click.pass_context
def cli(ctx: click.Context) -> None:
    """MoAI Agentic Development Kit

    SPEC-First TDD Framework with Alfred SuperAgent
    """
    # 하위 명령어 없이 실행되면 로고 출력
    if ctx.invoked_subcommand is None:
        show_logo()


# 명령어 등록
from moai_adk.cli.commands.backup import backup
from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import init
from moai_adk.cli.commands.restore import restore
from moai_adk.cli.commands.status import status
from moai_adk.cli.commands.update import update

cli.add_command(init)
cli.add_command(doctor)
cli.add_command(status)
cli.add_command(restore)
cli.add_command(backup)
cli.add_command(update)


def main() -> int:
    """CLI 진입점"""
    try:
        cli(standalone_mode=False)
        return 0
    except click.Abort:
        # 사용자가 Ctrl+C로 취소
        return 130
    except click.ClickException as e:
        e.show()
        return e.exit_code
    except Exception as e:
        console.print(f"[red]Error:[/red] {e}")
        return 1
    finally:
        # 출력 버퍼 명시적 flush
        console.file.flush()


if __name__ == "__main__":
    sys.exit(main())
