#!/usr/bin/env python3
# @REQ:CLI-COMMANDS-011
"""
CLI Commands for MoAI-ADK

Contains all Click command definitions for the MoAI-ADK CLI including
init, restore, doctor, status, update, and help commands.
"""

import click

from .._version import __version__
from .command_executor import (
    execute_doctor,
    execute_init,
    execute_restore,
)
from .command_operations import (
    execute_status,
    execute_update,
)
from .helpers import print_banner


@click.group(invoke_without_command=True)
@click.option("-V", "--version", is_flag=True, help="output the version number")
@click.option(
    "-h", "--help", "help_flag", is_flag=True, help="display help for command"
)
@click.pass_context
def cli(ctx: click.Context, version: bool, help_flag: bool) -> None:
    """Modu-AI's Agentic Development Kit"""
    if version:
        click.echo(f"MoAI-ADK v{__version__}")
        ctx.exit()

    if help_flag or ctx.invoked_subcommand is None:
        print_banner(show_usage=True)
        click.echo(click.get_current_context().get_help())


@cli.command()
@click.argument("backup_path", type=click.Path())
@click.option(
    "--dry-run", is_flag=True, help="Show what would be restored without making changes"
)
def restore(backup_path: str, dry_run: bool) -> None:
    """Restore MoAI-ADK from a backup directory."""
    execute_restore(backup_path, dry_run)


@cli.command()
@click.option("--list-backups", "-l", is_flag=True, help="List available backups")
def doctor(list_backups: bool) -> None:
    """Diagnose common issues and check system health."""
    execute_doctor(list_backups)


@cli.command()
@click.argument("project_path", type=click.Path(), default=".")
@click.option(
    "--template",
    "-t",
    default="standard",
    help="Template to use (standard, minimal, advanced)",
)
@click.option("--interactive", "-i", is_flag=True, help="Run interactive setup wizard")
@click.option(
    "--backup",
    "-b",
    is_flag=True,
    help="Create backup before installation (recommended)",
)
@click.option(
    "--force",
    "-f",
    is_flag=True,
    help="Force overwrite existing files (주의: 기존 파일을 덮어씁니다)",
)
@click.option(
    "--force-copy",
    is_flag=True,
    help="Force file copying instead of symlinks (recommended for Windows without admin rights)",
)
@click.option("--quiet", "-q", is_flag=True, help="Quiet mode - minimal output")
@click.option(
    "--personal",
    is_flag=True,
    help="Initialize in personal mode (default) - simplified workflow for individual development",
)
@click.option(
    "--team",
    is_flag=True,
    help="Initialize in team mode - full GitFlow with collaboration features",
)
def init(
    project_path: str,
    template: str,
    interactive: bool,
    backup: bool,
    force: bool,
    force_copy: bool,
    quiet: bool,
    personal: bool,
    team: bool,
) -> None:
    """Initialize a new MoAI-ADK project."""
    execute_init(
        project_path, template, interactive, backup, force, force_copy, quiet, personal, team
    )


@cli.command()
@click.argument("command", required=False)
def help(command: str | None) -> None:
    """Show help for MoAI-ADK commands."""
    if command:
        # Show help for specific command
        try:
            cmd = cli.get_command(None, command)
            if cmd:
                click.echo(cmd.get_help(click.Context(cmd)))
            else:
                click.echo(f"❌ Unknown command: {command}")
        except Exception:
            click.echo(f"❌ Unknown command: {command}")
    else:
        # Show general help
        print_banner(show_usage=True)
        click.echo("\nAvailable commands:")
        click.echo("  init      Initialize a new MoAI-ADK project")
        click.echo("  status    Show project status")
        click.echo("  doctor    Diagnose issues and check health")
        click.echo("  restore   Restore from backup")
        click.echo("  update    Update MoAI-ADK")
        click.echo("  help      Show this help message")
        click.echo("\nRun 'moai COMMAND --help' for command-specific help.")


@cli.command()
@click.option("--verbose", "-v", is_flag=True, help="Show detailed status information")
@click.option(
    "--project-path",
    "-p",
    type=click.Path(exists=True),
    help="Path to project directory (default: current directory)",
)
def status(verbose: bool, project_path: str | None) -> None:
    """Show MoAI-ADK project status."""
    execute_status(verbose, project_path)


@cli.command()
@click.option(
    "--check", "-c", is_flag=True, help="Check for updates without installing"
)
@click.option("--no-backup", is_flag=True, help="Skip backup creation before update")
@click.option("--verbose", "-v", is_flag=True, help="Show detailed update information")
@click.option("--package-only", is_flag=True, help="Update only the Python package")
@click.option("--resources-only", is_flag=True, help="Update only project resources")
def update(
    check: bool,
    no_backup: bool,
    verbose: bool,
    package_only: bool,
    resources_only: bool,
) -> None:
    """Update MoAI-ADK to the latest version."""
    execute_update(check, no_backup, verbose, package_only, resources_only)


# Add all commands to the CLI group
cli.add_command(restore)
cli.add_command(doctor)
cli.add_command(init)
cli.add_command(help)
cli.add_command(status)
cli.add_command(update)
