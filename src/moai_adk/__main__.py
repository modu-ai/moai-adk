# type: ignore
"""MoAI-ADK CLI Entry Point

Implements the CLI entry point:
- Click-based CLI framework with lazy command loading
- Rich console terminal output
- ASCII logo rendering (lazy-loaded)
- --version and --help options
- Five core commands: init, doctor, status, update (lazy-loaded)

Performance optimization: Commands and heavy libraries are lazy-loaded
to reduce CLI startup time by 75% (~400ms → ~100ms).
"""

import sys

import click

from moai_adk import __version__

# Lazy-loaded console (created when needed)
_console = None


def get_console():
    """Get or create Rich Console instance (lazy loading)"""
    global _console
    if _console is None:
        from rich.console import Console

        _console = Console()
    return _console


def show_logo() -> None:
    """Render the MoAI-ADK ASCII logo with Pyfiglet (lazy-loaded)"""
    # Lazy load pyfiglet only when displaying logo
    import pyfiglet

    console = get_console()

    # Generate the "MoAI-ADK" banner using the ansi_shadow font
    logo = pyfiglet.figlet_format("MoAI-ADK", font="ansi_shadow")

    # Print with Rich styling
    console.print(logo, style="cyan bold", highlight=False)
    console.print(
        "  Modu-AI's Agentic Development Kit w/ SuperAgent 🎩 Alfred",
        style="yellow bold",
    )
    console.print()
    console.print("  Version: ", style="green", end="")
    console.print(__version__, style="cyan bold")
    console.print()
    console.print("  Tip: Run ", style="yellow", end="")
    console.print("uv run moai-adk --help", style="cyan", end="")
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


# Lazy-loaded commands (imported only when invoked)
@cli.command()
@click.argument("path", type=click.Path(), default=".")
@click.option(
    "--non-interactive",
    "-y",
    is_flag=True,
    help="Non-interactive mode (use defaults)",
)
@click.option(
    "--mode",
    type=click.Choice(["personal", "team"]),
    default="personal",
    help="Project mode",
)
@click.option(
    "--locale",
    type=click.Choice(["ko", "en", "ja", "zh"]),
    default=None,
    help="Preferred language (ko/en/ja/zh, default: en)",
)
@click.option(
    "--language",
    type=str,
    default=None,
    help="Programming language (auto-detect if not specified)",
)
@click.option(
    "--force",
    is_flag=True,
    help="Force reinitialize without confirmation",
)
@click.pass_context
def init(
    ctx: click.Context,
    path: str,
    non_interactive: bool,
    mode: str,
    locale: str,
    language: str | None,
    force: bool,
) -> None:
    """Initialize a new MoAI-ADK project"""
    from moai_adk.cli.commands.init import init as _init

    ctx.invoke(
        _init,
        path=path,
        non_interactive=non_interactive,
        mode=mode,
        locale=locale,
        language=language,
        force=force,
    )


@cli.command()
@click.pass_context
def doctor(ctx: click.Context, **kwargs) -> None:
    """Run system diagnostics"""
    from moai_adk.cli.commands.doctor import doctor as _doctor

    ctx.invoke(_doctor, **kwargs)


@cli.command()
@click.pass_context
def status(ctx: click.Context, **kwargs) -> None:
    """Show project status"""
    from moai_adk.cli.commands.status import status as _status

    ctx.invoke(_status, **kwargs)


@cli.command()
@click.pass_context
def update(ctx: click.Context, **kwargs) -> None:
    """Update MoAI-ADK to latest version"""
    from moai_adk.cli.commands.update import update as _update

    ctx.invoke(_update, **kwargs)


# statusline command (for Claude Code statusline rendering)
@cli.command(name="statusline")
def statusline() -> None:
    """Render Claude Code statusline (internal use only)"""
    import json

    # Lazy load statusline module
    from moai_adk.statusline.main import build_statusline_data

    try:
        # Read JSON context from stdin
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        context = json.loads(input_data) if input_data else {}
    except (json.JSONDecodeError, EOFError, ValueError):
        context = {}

    # Render statusline
    output = build_statusline_data(context, mode="extended")
    print(output, end="")


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
        console = get_console()
        console.print(f"[red]Error:[/red] {e}")
        return 1
    finally:
        # Flush the output buffer explicitly if console was created
        if _console is not None:
            _console.file.flush()


if __name__ == "__main__":
    sys.exit(main())
