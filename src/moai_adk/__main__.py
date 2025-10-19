# @CODE:CLI-001 | SPEC: SPEC-CLI-001/spec.md | TEST: tests/unit/test_cli_commands.py
"""MoAI-ADK CLI Entry Point

Implements the CLI entry point:
- Click-based CLI framework
- Rich console terminal output
- ASCII logo rendering
- --version and --help options
- Five core commands: init, doctor, status, backup, update
"""

import sys

import click
import pyfiglet
from rich.console import Console

from moai_adk import __version__
from moai_adk.cli.commands.backup import backup
from moai_adk.cli.commands.doctor import doctor
from moai_adk.cli.commands.init import init
from moai_adk.cli.commands.status import status
from moai_adk.cli.commands.update import update

console = Console()


def show_logo() -> None:
    """Render the MoAI-ADK ASCII logo with Pyfiglet"""
    # Generate the "MoAI-ADK" banner using the ansi_shadow font
    logo = pyfiglet.figlet_format("MoAI-ADK", font="ansi_shadow")

    # Print with Rich styling
    console.print(logo, style="cyan bold", highlight=False)
    console.print("  Modu-AI's Agentic Development Kit w/ SuperAgent 🎩 Alfred", style="yellow bold")
    console.print()
    console.print("  Version: ", style="green", end="")
    console.print(__version__, style="cyan bold")
    console.print()
    console.print("  Tip: Run ", style="yellow", end="")
    console.print("python -m moai_adk --help", style="cyan", end="")
    console.print(" to see available commands", style="yellow")


@click.group(invoke_without_command=True)
@click.version_option(version=__version__, prog_name="MoAI-ADK")
@click.pass_context
def cli(ctx: click.Context) -> None:
    """MoAI Agentic Development Kit

    SPEC-First TDD Framework with Alfred SuperAgent
    """
    # Display the logo when no subcommand is invoked
    if ctx.invoked_subcommand is None:
        show_logo()

cli.add_command(init)
cli.add_command(doctor)
cli.add_command(status)
cli.add_command(backup)
cli.add_command(update)


def main() -> int:
    """CLI entry point"""
    try:
        cli(standalone_mode=False)
        return 0
    except click.Abort:
        # User cancelled with Ctrl+C
        return 130
    except click.ClickException as e:
        e.show()
        return e.exit_code
    except Exception as e:
        console.print(f"[red]Error:[/red] {e}")
        return 1
    finally:
        # Flush the output buffer explicitly
        console.file.flush()


if __name__ == "__main__":
    sys.exit(main())
