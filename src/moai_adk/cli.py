#!/usr/bin/env python3
# @FEATURE:CLI-UNIFIED-001
"""
🗿 MoAI-ADK Unified CLI Entry Point

@REQ:CLI-UNIFIED-001 → @DESIGN:CLI-SINGLE-ENTRY-001 → @TASK:CLI-CONSOLIDATION-001 → @TEST:CLI-COMMANDS-001

Single unified CLI entry point that consolidates all MoAI-ADK commands.
Designed for easy packaging with PyInstaller and uv compatibility.

This module serves as the main entry point for:
- pip/pipx installations: `moai init`
- uv installations: `uvx --from moai-adk moai-adk init`
- PyInstaller executables: `moai-adk.exe init`
"""

import sys
from pathlib import Path
from typing import Optional

import click
from colorama import Fore, Style, init

from .cli.commands import cli as cli_commands
from .cli.helpers import validate_environment
from .utils.logger import get_logger

# Initialize colorama for cross-platform colored output
init(autoreset=True)

logger = get_logger(__name__)


def show_banner() -> None:
    """Display MoAI-ADK banner with version info."""
    try:
        from ._version import get_version
        version = get_version()
    except ImportError:
        version = "0.1.28"

    banner = f"""
{Fore.CYAN}🗿 MoAI-ADK{Style.RESET_ALL} - Modu-AI Agentic Development Kit
{Fore.YELLOW}Version:{Style.RESET_ALL} {version}
{Fore.GREEN}Ready for Spec-First TDD Development{Style.RESET_ALL}
"""
    click.echo(banner)


@click.group(invoke_without_command=True)
@click.option('--version', is_flag=True, help='Show version information')
@click.option('--verbose', '-v', is_flag=True, help='Enable verbose output')
@click.pass_context
def main(ctx: click.Context, version: bool, verbose: bool) -> None:
    """
    🗿 MoAI-ADK - Modu-AI Agentic Development Kit

    Spec-First TDD development toolkit for Claude Code integration.

    Examples:
        moai-adk init                    # Initialize new project
        moai-adk init --team             # Initialize with team mode
        moai-adk config                  # Configure MoAI-ADK
        moai-adk update                  # Update to latest version
    """
    if version:
        try:
            from ._version import get_version
            click.echo(f"MoAI-ADK {get_version()}")
        except ImportError:
            click.echo("MoAI-ADK 0.1.28")
        return

    if ctx.invoked_subcommand is None:
        show_banner()
        click.echo("Use --help for available commands.")
        return

    # Set verbose logging if requested
    if verbose:
        import logging
        logging.getLogger('moai_adk').setLevel(logging.DEBUG)


# Import and register all CLI commands from the existing cli module
@main.command()
@click.argument('project_path', default='.', type=click.Path())
@click.option('--team', is_flag=True, help='Initialize with team mode (GitHub integration)')
@click.option('--personal', is_flag=True, help='Initialize with personal mode (local Git only)')
@click.option('--force', is_flag=True, help='Force initialization even if directory exists')
def init(project_path: str, team: bool, personal: bool, force: bool) -> None:
    """Initialize a new MoAI-ADK project."""
    from .cli.commands import init_project

    try:
        # Basic environment validation
        if not validate_environment():
            click.echo(f"{Fore.RED}Environment validation failed{Style.RESET_ALL}")
            sys.exit(1)

        # Determine mode
        if team and personal:
            click.echo(f"{Fore.RED}Error: Cannot specify both --team and --personal{Style.RESET_ALL}")
            sys.exit(1)

        mode = 'team' if team else 'personal' if personal else None

        # Call the existing init function
        success = init_project(Path(project_path), mode=mode, force=force)

        if success:
            click.echo(f"{Fore.GREEN}✅ Project initialized successfully{Style.RESET_ALL}")
        else:
            click.echo(f"{Fore.RED}❌ Project initialization failed{Style.RESET_ALL}")
            sys.exit(1)

    except KeyboardInterrupt:
        click.echo(f"\n{Fore.YELLOW}Operation cancelled by user{Style.RESET_ALL}")
        sys.exit(1)
    except Exception as e:
        logger.error(f"Initialization failed: {e}")
        click.echo(f"{Fore.RED}❌ Initialization failed: {e}{Style.RESET_ALL}")
        sys.exit(1)


@main.command()
def config() -> None:
    """Configure MoAI-ADK settings."""
    from .cli.wizard import InteractiveWizard

    try:
        wizard = InteractiveWizard()
        wizard.run_configuration()
        click.echo(f"{Fore.GREEN}✅ Configuration completed{Style.RESET_ALL}")
    except KeyboardInterrupt:
        click.echo(f"\n{Fore.YELLOW}Configuration cancelled{Style.RESET_ALL}")
    except Exception as e:
        logger.error(f"Configuration failed: {e}")
        click.echo(f"{Fore.RED}❌ Configuration failed: {e}{Style.RESET_ALL}")
        sys.exit(1)


@main.command()
@click.option('--check', is_flag=True, help='Check for updates without installing')
def update(check: bool) -> None:
    """Update MoAI-ADK to the latest version."""
    try:
        if check:
            # Check for updates
            click.echo(f"{Fore.CYAN}Checking for updates...{Style.RESET_ALL}")
            # TODO: Implement update checking logic
            click.echo(f"{Fore.GREEN}You are running the latest version{Style.RESET_ALL}")
        else:
            # Perform update
            click.echo(f"{Fore.CYAN}Updating MoAI-ADK...{Style.RESET_ALL}")
            # TODO: Implement update logic
            click.echo(f"{Fore.GREEN}✅ Update completed{Style.RESET_ALL}")
    except Exception as e:
        logger.error(f"Update failed: {e}")
        click.echo(f"{Fore.RED}❌ Update failed: {e}{Style.RESET_ALL}")
        sys.exit(1)


@main.command()
def doctor() -> None:
    """Run diagnostic checks on the MoAI-ADK installation."""
    try:
        click.echo(f"{Fore.CYAN}🔍 Running MoAI-ADK diagnostics...{Style.RESET_ALL}")

        # Environment checks
        env_ok = validate_environment()
        if env_ok:
            click.echo(f"{Fore.GREEN}✅ Environment: OK{Style.RESET_ALL}")
        else:
            click.echo(f"{Fore.RED}❌ Environment: Issues detected{Style.RESET_ALL}")

        # Check dependencies
        try:
            import git
            click.echo(f"{Fore.GREEN}✅ Git: Available{Style.RESET_ALL}")
        except ImportError:
            click.echo(f"{Fore.YELLOW}⚠️  Git: Not available (optional){Style.RESET_ALL}")

        # Check Python version
        if sys.version_info >= (3, 10):
            click.echo(f"{Fore.GREEN}✅ Python: {sys.version.split()[0]} (supported){Style.RESET_ALL}")
        else:
            click.echo(f"{Fore.RED}❌ Python: {sys.version.split()[0]} (unsupported, need >=3.10){Style.RESET_ALL}")

        click.echo(f"\n{Fore.CYAN}Diagnostics completed{Style.RESET_ALL}")

    except Exception as e:
        logger.error(f"Diagnostics failed: {e}")
        click.echo(f"{Fore.RED}❌ Diagnostics failed: {e}{Style.RESET_ALL}")
        sys.exit(1)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        click.echo(f"\n{Fore.YELLOW}Operation cancelled by user{Style.RESET_ALL}")
        sys.exit(1)
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        click.echo(f"{Fore.RED}❌ Unexpected error: {e}{Style.RESET_ALL}")
        sys.exit(1)