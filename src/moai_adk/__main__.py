# type: ignore
"""MoAI-ADK CLI Entry Point

Implements the CLI entry point:
- Click-based CLI framework with lazy command loading
- Rich console terminal output
- ASCII logo rendering (lazy-loaded)
- --version and --help options
- Core commands: init, doctor, status, update, switch, rank (lazy-loaded)

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
@click.option(
    "--verbose",
    "-v",
    is_flag=True,
    help="Show detailed tool versions and language detection",
)
@click.option("--fix", is_flag=True, help="Suggest fixes for missing tools")
@click.option("--export", type=click.Path(), help="Export diagnostics to JSON file")
@click.option("--check", type=str, help="Check specific tool only")
@click.option("--check-commands", is_flag=True, help="Diagnose slash command loading issues")
@click.pass_context
def doctor(
    ctx: click.Context,
    verbose: bool,
    fix: bool,
    export: str | None,
    check: str | None,
    check_commands: bool,
) -> None:
    """Run system diagnostics"""
    from moai_adk.cli.commands.doctor import doctor as _doctor

    ctx.invoke(
        _doctor,
        verbose=verbose,
        fix=fix,
        export=export,
        check=check,
        check_commands=check_commands,
    )


@cli.command()
@click.pass_context
def status(ctx: click.Context, **kwargs) -> None:
    """Show project status"""
    from moai_adk.cli.commands.status import status as _status

    ctx.invoke(_status, **kwargs)


@cli.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="Project path (default: current directory)",
)
@click.option("--force", is_flag=True, help="Skip backup and force the update")
@click.option("--check", is_flag=True, help="Only check version (do not update)")
@click.option("--templates-only", is_flag=True, help="Skip package upgrade, sync templates only")
@click.option("--yes", is_flag=True, help="Auto-confirm all prompts (CI/CD mode)")
@click.option(
    "-c",
    "--config",
    "edit_config",
    is_flag=True,
    help="Edit project configuration (same as init wizard)",
)
@click.option(
    "--merge",
    "merge_strategy",
    flag_value="auto",
    help="Auto-merge: Apply template + preserve user changes",
)
@click.option(
    "--manual",
    "merge_strategy",
    flag_value="manual",
    help="Manual merge: Preserve backup, generate merge guide",
)
@click.pass_context
def update(
    ctx: click.Context,
    path: str,
    force: bool,
    check: bool,
    templates_only: bool,
    yes: bool,
    edit_config: bool,
    merge_strategy: str | None,
) -> None:
    """Update MoAI-ADK to latest version"""
    from moai_adk.cli.commands.update import update as _update

    ctx.invoke(
        _update,
        path=path,
        force=force,
        check=check,
        templates_only=templates_only,
        yes=yes,
        edit_config=edit_config,
        merge_strategy=merge_strategy,
    )


@cli.command(name="claude")
def claude() -> None:
    """Switch to Claude backend (Anthropic API)"""
    from moai_adk.cli.commands.switch import switch_to_claude

    switch_to_claude()


# Alias: moai cc -> moai claude
@cli.command(name="cc", hidden=True)
def cc() -> None:
    """Alias for 'claude' command"""
    from moai_adk.cli.commands.switch import switch_to_claude

    switch_to_claude()


@cli.command()
@click.argument("api_key", required=False, default=None)
def glm(api_key: str | None) -> None:
    """Switch to GLM backend (cost-effective) or update API key

    Usage:
      moai glm              # Switch to GLM backend (requires key set)
      moai glm <api-key>    # Update API key (without switching)
    """
    from moai_adk.cli.commands.switch import switch_to_glm, update_glm_key

    if api_key:
        # Update API key only, no backend switch
        update_glm_key(api_key)
    else:
        # Switch to GLM backend
        switch_to_glm()


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


# Rank command group (lazy-loaded)
def _load_rank_group():
    """Lazy load the rank command group."""
    from moai_adk.cli.commands.rank import rank

    return rank


# Register rank command group
cli.add_command(_load_rank_group(), name="rank")


# Worktree command group (lazy-loaded)
def _load_worktree_group():
    """Lazy load the worktree command group."""
    from moai_adk.cli.worktree.cli import worktree

    return worktree


# Register worktree command group
cli.add_command(_load_worktree_group(), name="worktree")


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
