# @CODE:UTILS-001 | SPEC: SPEC-CLI-001.md
"""ASCII 배너 모듈

MoAI-ADK ASCII 아트 배너 출력
"""

from rich.console import Console

console = Console()

MOAI_BANNER = """
███╗   ███╗ ██████╗  █████╗ ██╗       █████╗ ██████╗ ██╗  ██╗
████╗ ████║██╔═══██╗██╔══██╗██║      ██╔══██╗██╔══██╗██║ ██╔╝
██╔████╔██║██║   ██║███████║██║█████╗███████║██║  ██║█████╔╝
██║╚██╔╝██║██║   ██║██╔══██║██║╚════╝██╔══██║██║  ██║██╔═██╗
██║ ╚═╝ ██║╚██████╔╝██║  ██║██║      ██║  ██║██████╔╝██║  ██╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝
"""


def print_banner(version: str = "0.3.0") -> None:
    """MoAI-ADK 배너 출력

    Args:
        version: MoAI-ADK 버전
    """
    console.print(f"[cyan]{MOAI_BANNER}[/cyan]")
    console.print(
        "[dim]  Modu-AI's Agentic Development Kit w/ SuperAgent 🎩 Alfred[/dim]\n"
    )
    console.print(f"[dim]  Version: {version}[/dim]\n")


def print_welcome_message() -> None:
    """환영 메시지 출력"""
    console.print("[cyan bold]🚀 Welcome to MoAI-ADK Project Initialization![/cyan bold]\n")
    console.print(
        "[dim]This wizard will guide you through setting up your MoAI-ADK project.[/dim]"
    )
    console.print(
        "[dim]You can press Ctrl+C at any time to cancel.\n[/dim]"
    )
