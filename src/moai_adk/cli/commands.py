#!/usr/bin/env python3
"""
CLI Commands for MoAI-ADK

Contains all Click command definitions for the MoAI-ADK CLI including
init, restore, doctor, status, update, and help commands.
"""

import sys
import shutil
from pathlib import Path
from typing import Optional

import click
from colorama import Fore, Style

from .._version import __version__, get_version_format
from ..config import Config, RuntimeConfig
from ..install.installer import SimplifiedInstaller
from ..install.resource_manager import ResourceManager
from ..core.resource_version import ResourceVersionManager
from ..utils.logger import get_logger
from ..install.post_install import auto_install_on_first_run
from ..core.version_sync import VersionSyncManager
from .helpers import (
    create_installation_backup,
    detect_potential_conflicts,
    analyze_existing_project,
    print_banner,
    validate_environment,
    format_project_status
)

logger = get_logger(__name__)


@click.group(invoke_without_command=True)
@click.option("-V", "--version", is_flag=True, help="output the version number")
@click.option("-h", "--help", "help_flag", is_flag=True, help="display help for command")
@click.pass_context
def cli(ctx: click.Context, version: bool, help_flag: bool) -> None:
    """@FEATURE:CLI-001 MoAI-ADK: Agentic Development Kit for Claude Code"""
    if version:
        print(f"MoAI-ADK v{__version__}")
        ctx.exit()

    if help_flag or ctx.invoked_subcommand is None:
        print_banner(show_usage=True)
        print(click.get_current_context().get_help())


@cli.command()
@click.argument("backup_path", type=click.Path(exists=True))
@click.option("--dry-run", is_flag=True, help="Show what would be restored without making changes")
def restore(backup_path: str, dry_run: bool) -> None:
    """@TASK:RESTORE-001 Restore MoAI-ADK from a backup directory."""
    backup_dir = Path(backup_path)

    if not backup_dir.is_dir():
        click.echo(f"{Fore.RED}❌ Backup path must be a directory{Style.RESET_ALL}")
        sys.exit(1)

    # Validate backup directory structure
    required_items = [".moai", ".claude", "CLAUDE.md"]
    missing_items = [item for item in required_items if not (backup_dir / item).exists()]

    if missing_items:
        click.echo(f"{Fore.YELLOW}⚠️  Warning: Backup may be incomplete. Missing: {', '.join(missing_items)}{Style.RESET_ALL}")

    current_dir = Path.cwd()

    if dry_run:
        click.echo(f"{Fore.CYAN}🔍 Dry run - would restore to: {current_dir}{Style.RESET_ALL}")
        for item in required_items:
            source = backup_dir / item
            target = current_dir / item
            if source.exists():
                click.echo(f"  Would restore: {source} → {target}")
    else:
        click.echo(f"{Fore.CYAN}🔄 Restoring backup to: {current_dir}{Style.RESET_ALL}")

        try:
            # Remove existing directories/files
            for item in required_items:
                target = current_dir / item
                if target.exists():
                    if target.is_dir():
                        shutil.rmtree(target)
                    else:
                        target.unlink()

            # Copy from backup
            for item in required_items:
                source = backup_dir / item
                target = current_dir / item
                if source.exists():
                    if source.is_dir():
                        shutil.copytree(source, target)
                    else:
                        shutil.copy2(source, target)
                    click.echo(f"  Restored: {item}")

            click.echo(f"{Fore.GREEN}✅ Backup restored successfully{Style.RESET_ALL}")

        except Exception as e:
            click.echo(f"{Fore.RED}❌ Failed to restore backup: {e}{Style.RESET_ALL}")
            sys.exit(1)


@cli.command()
@click.option("--list-backups", "-l", is_flag=True, help="List available backups")
def doctor(list_backups: bool) -> None:
    """@TASK:HEALTH-001 Diagnose common issues and check system health."""
    click.echo(f"{Fore.CYAN}🔍 MoAI-ADK Health Check{Style.RESET_ALL}")

    if list_backups:
        current_dir = Path.cwd()
        backup_dirs = list(current_dir.glob(".moai_backup_*"))

        if backup_dirs:
            click.echo(f"\n📦 Available backups in {current_dir}:")
            for backup_dir in sorted(backup_dirs):
                timestamp = backup_dir.name.split("_", 2)[-1]
                click.echo(f"  {backup_dir.name} (created: {timestamp})")
        else:
            click.echo(f"\n📦 No backups found in {current_dir}")
        return

    # Environment checks
    click.echo("\n🔧 Environment:")
    if validate_environment():
        click.echo(f"  {Fore.GREEN}✅ Python version{Style.RESET_ALL}")
    else:
        click.echo(f"  {Fore.RED}❌ Python version issues{Style.RESET_ALL}")

    # Project checks
    current_dir = Path.cwd()
    click.echo(f"\n📂 Project Status ({current_dir}):")

    moai_exists = (current_dir / ".moai").exists()
    claude_exists = (current_dir / ".claude").exists()
    memory_exists = (current_dir / "CLAUDE.md").exists()

    click.echo(f"  MoAI directory: {'✅' if moai_exists else '❌'}")
    click.echo(f"  Claude directory: {'✅' if claude_exists else '❌'}")
    click.echo(f"  Memory file: {'✅' if memory_exists else '❌'}")

    if not any([moai_exists, claude_exists, memory_exists]):
        click.echo(f"\n💡 This doesn't appear to be a MoAI-ADK project.")
        click.echo(f"   Run 'moai init' to initialize.")

    # Resource validation
    if auto_install_on_first_run():
        click.echo(f"  {Fore.GREEN}✅ Resources available{Style.RESET_ALL}")
    else:
        click.echo(f"  {Fore.RED}❌ Resource issues detected{Style.RESET_ALL}")


@cli.command()
@click.argument("project_path", type=click.Path(), default=".")
@click.option("--template", "-t", default="standard", help="Template to use (standard, minimal, advanced)")
@click.option("--interactive", "-i", is_flag=True, help="Run interactive setup wizard")
@click.option("--backup", "-b", is_flag=True, help="Create backup before installation (recommended)")
@click.option("--force", "-f", is_flag=True, help="Force overwrite existing files (주의: 기존 파일을 덮어씁니다)")
@click.option("--force-copy", is_flag=True, help="Force file copying instead of symlinks (recommended for Windows without admin rights)")
@click.option("--quiet", "-q", is_flag=True, help="Quiet mode - minimal output")
@click.option("--personal", is_flag=True, help="Initialize in personal mode (default) - simplified workflow for individual development")
@click.option("--team", is_flag=True, help="Initialize in team mode - full GitFlow with collaboration features")
def init(project_path: str, template: str, interactive: bool, backup: bool, force: bool, force_copy: bool, quiet: bool, personal: bool, team: bool) -> None:
    """@TASK:INIT-001 Initialize a new MoAI-ADK project."""
    project_dir = Path(project_path).resolve()

    # Determine project mode
    if team and personal:
        click.echo(f"{Fore.RED}❌ Cannot specify both --personal and --team modes{Style.RESET_ALL}")
        sys.exit(1)

    # Default to personal mode if no mode specified
    project_mode = "team" if team else "personal"

    if not quiet:
        mode_icon = "🏢" if project_mode == "team" else "👤"
        click.echo(f"{Fore.CYAN}{mode_icon} Initializing in {project_mode} mode{Style.RESET_ALL}")

    # Interactive mode
    if interactive:
        from .wizard import InteractiveWizard
        wizard = InteractiveWizard()
        try:
            result = wizard.run(project_dir)
            if result.success:
                click.echo(f"{Fore.GREEN}🎉 Interactive setup completed!{Style.RESET_ALL}")
                for step in result.next_steps:
                    click.echo(f"   {step}")
            else:
                click.echo(f"{Fore.RED}❌ Interactive setup failed{Style.RESET_ALL}")
                for error in result.errors:
                    click.echo(f"   {error}")
            return
        except KeyboardInterrupt:
            click.echo(f"\n{Fore.YELLOW}⚠️  Setup cancelled{Style.RESET_ALL}")
            return
        except Exception as e:
            click.echo(f"{Fore.RED}❌ Interactive setup error: {e}{Style.RESET_ALL}")
            return

    # Set logging level based on quiet mode
    if quiet:
        import logging
        # Silence all moai_adk loggers
        logging.getLogger("moai_adk").setLevel(logging.CRITICAL)
        for logger_name in logging.Logger.manager.loggerDict:
            if logger_name.startswith("moai_adk"):
                logging.getLogger(logger_name).setLevel(logging.CRITICAL)

    # Validate environment
    if not validate_environment():
        if not quiet:
            click.echo(f"{Fore.RED}❌ Environment validation failed{Style.RESET_ALL}")
        sys.exit(1)

    # Check for conflicts
    conflicts = detect_potential_conflicts(project_dir)
    if conflicts and not force:
        if not quiet:
            click.echo(f"{Fore.YELLOW}⚠️  Potential conflicts detected:{Style.RESET_ALL}")
            for conflict in conflicts:
                click.echo(f"   - {conflict}")

        if backup:
            if not quiet:
                click.echo(f"\n{Fore.CYAN}📦 Creating backup...{Style.RESET_ALL}")
            if create_installation_backup(project_dir):
                if not quiet:
                    click.echo(f"{Fore.GREEN}✅ Backup created{Style.RESET_ALL}")
            else:
                if not quiet:
                    click.echo(f"{Fore.RED}❌ Backup failed{Style.RESET_ALL}")
                sys.exit(1)
        else:
            if not quiet:
                click.echo(f"\n💡 Use --backup to create a backup first, or --force to overwrite")
            sys.exit(1)

    # Project analysis
    analysis = analyze_existing_project(project_dir)
    if analysis["project_type"] != "unknown" and not quiet:
        click.echo(f"{Fore.CYAN}📋 Detected {analysis['project_type']} project{Style.RESET_ALL}")

    # Create configuration
    config = Config(
        project_path=str(project_dir),
        name=project_dir.name,
        template=template,
        runtime=RuntimeConfig("python"),
        force_overwrite=force
    )

    # Run installation
    installer = SimplifiedInstaller(config)

    def progress_callback(message: str, current: int, total: int):
        if not quiet:
            # Clean progress output without colors/emojis
            click.echo(f"  {message}")

    try:
        # Show header
        if not quiet:
            click.echo(f"\n{Fore.CYAN}Creating MoAI-ADK project in {project_dir.name}...{Style.RESET_ALL}")

        result = installer.install(progress_callback)

        # Create mode-specific configuration
        if result.success:
            create_mode_configuration(project_dir, project_mode, quiet)

        if result.success:
            if quiet:
                # Minimal output for quiet mode
                click.echo(f"✓ MoAI-ADK project created: {result.project_path}")
            else:
                click.echo(f"\n{Fore.GREEN}✓{Style.RESET_ALL} MoAI-ADK initialized successfully!")
                click.echo(f"  Project: {result.project_path}")
                click.echo(f"  Files created: {len(result.files_created)}")

                # Show next steps
                click.echo(f"\n{Fore.CYAN}Next Steps:{Style.RESET_ALL}")
                for step in result.next_steps:
                    if step.strip():  # Skip empty lines
                        if step.startswith(("1.", "2.", "3.", "4.")):
                            click.echo(f"  {step}")
                        elif step.startswith("   "):
                            click.echo(f"    {step.strip()}")
                        else:
                            click.echo(f"  {step}")

        else:
            if quiet:
                click.echo(f"✗ Installation failed: {'; '.join(result.errors)}")
            else:
                click.echo(f"\n{Fore.RED}✗{Style.RESET_ALL} Installation failed")
                for error in result.errors:
                    click.echo(f"  {error}")
            sys.exit(1)

    except KeyboardInterrupt:
        click.echo(f"\n{Fore.YELLOW}⚠️  Installation cancelled{Style.RESET_ALL}")
    except Exception as e:
        click.echo(f"{Fore.RED}❌ Installation error: {e}{Style.RESET_ALL}")
        sys.exit(1)


@cli.command()
@click.argument("command", required=False)
def help(command: Optional[str]) -> None:
    """@TASK:HELP-001 Show help for MoAI-ADK commands."""
    if command:
        # Show help for specific command
        try:
            cmd = cli.get_command(None, command)
            if cmd:
                click.echo(cmd.get_help(click.Context(cmd)))
            else:
                click.echo(f"{Fore.RED}❌ Unknown command: {command}{Style.RESET_ALL}")
        except Exception:
            click.echo(f"{Fore.RED}❌ Unknown command: {command}{Style.RESET_ALL}")
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
@click.option(
    "--verbose", "-v",
    is_flag=True,
    help="Show detailed status information"
)
@click.option(
    "--project-path", "-p",
    type=click.Path(exists=True),
    help="Path to project directory (default: current directory)"
)
def status(verbose: bool, project_path: Optional[str]) -> None:
    """@TASK:STATUS-001 Show MoAI-ADK project status."""
    target_path = Path(project_path) if project_path else Path.cwd()

    click.echo(f"{Fore.CYAN}📊 MoAI-ADK Project Status{Style.RESET_ALL}")

    status_info = format_project_status(target_path)

    click.echo(f"\n📂 Project: {status_info['path']}")
    click.echo(f"   Type: {status_info['project_type']}")

    # Core status
    click.echo(f"\n🗿 MoAI-ADK Components:")
    click.echo(f"   MoAI System: {'✅' if status_info['moai_initialized'] else '❌'}")
    click.echo(f"   Claude Integration: {'✅' if status_info['claude_initialized'] else '❌'}")
    click.echo(f"   Memory File: {'✅' if status_info['memory_file'] else '❌'}")
    click.echo(f"   Git Repository: {'✅' if status_info['git_repository'] else '❌'}")

    versions = status_info.get('versions')
    if versions:
        click.echo(f"\n🧭 Versions:")
        click.echo(f"   Package: v{versions.get('package', 'unknown')}")
        click.echo(f"   Templates: v{versions.get('resources', 'unknown')}")
        if versions.get('available_resources') and versions.get('available_resources') != versions.get('resources'):
            click.echo(f"   Available template update: v{versions['available_resources']}")
        if versions.get('outdated'):
            click.echo(f"{Fore.YELLOW}   ⚠️  Templates are outdated. Run 'moai update' to refresh.{Style.RESET_ALL}")

    if verbose and status_info['file_counts']:
        click.echo(f"\n📁 File Counts:")
        for component, count in status_info['file_counts'].items():
            click.echo(f"   {component}: {count} files")

    # Recommendations
    recommendations = []
    if not status_info['moai_initialized']:
        recommendations.append("Run 'moai init' to initialize MoAI-ADK")
    if not status_info['git_repository']:
        recommendations.append("Initialize Git repository: git init")

    if recommendations:
        click.echo(f"\n💡 Recommendations:")
        for rec in recommendations:
            click.echo(f"   - {rec}")


@cli.command()
@click.option(
    "--check", "-c",
    is_flag=True,
    help="Check for updates without installing"
)
@click.option(
    "--no-backup",
    is_flag=True,
    help="Skip backup creation before update"
)
@click.option(
    "--verbose", "-v",
    is_flag=True,
    help="Show detailed update information"
)
@click.option(
    "--package-only",
    is_flag=True,
    help="Update only the Python package"
)
@click.option(
    "--resources-only",
    is_flag=True,
    help="Update only project resources"
)
def update(check: bool, no_backup: bool, verbose: bool, package_only: bool, resources_only: bool) -> None:
    """@TASK:UPDATE-001 Update MoAI-ADK to the latest version."""
    current_version = __version__
    project_path = Path.cwd()

    resource_manager = ResourceManager()
    version_manager = ResourceVersionManager(project_path)
    version_info = version_manager.read()
    current_resource_version = version_info.get('template_version') or "unknown"
    available_resource_version = resource_manager.get_version()

    if check:
        click.echo(f"{Fore.CYAN}🔍 Checking for updates...{Style.RESET_ALL}")
        click.echo(f"Current version: v{current_version}")
        click.echo(f"Installed template version: {current_resource_version}")
        click.echo(f"Available template version: {available_resource_version}")

        if not (project_path / ".moai").exists():
            click.echo(f"{Fore.YELLOW}⚠️  This directory does not appear to be a MoAI-ADK project{Style.RESET_ALL}")
        elif current_resource_version != available_resource_version:
            click.echo(f"{Fore.YELLOW}⚠️  Templates are outdated. Run 'moai update' to refresh.{Style.RESET_ALL}")
        else:
            click.echo(f"{Fore.GREEN}✅ Project resources are up to date{Style.RESET_ALL}")
        return

    if package_only and resources_only:
        click.echo(f"{Fore.RED}❌ Cannot use --package-only and --resources-only together{Style.RESET_ALL}")
        sys.exit(1)

    # Check if this is a MoAI-ADK project
    if not (project_path / ".moai").exists():
        click.echo(f"{Fore.YELLOW}⚠️  This doesn't appear to be a MoAI-ADK project{Style.RESET_ALL}")
        click.echo("Run 'moai init' to initialize a new project")
        return

    # Create backup unless explicitly disabled
    if not no_backup:
        click.echo(f"{Fore.CYAN}📦 Creating backup...{Style.RESET_ALL}")
        if not create_installation_backup(project_path):
            click.echo(f"{Fore.RED}❌ Backup failed - aborting update{Style.RESET_ALL}")
            sys.exit(1)
        click.echo(f"{Fore.GREEN}✅ Backup created{Style.RESET_ALL}")

    try:
        if not package_only:
            # Update resources
            click.echo(f"{Fore.CYAN}🔄 Updating project resources...{Style.RESET_ALL}")
            resource_manager.copy_claude_resources(project_path, overwrite=True)
            resource_manager.copy_moai_resources(project_path, overwrite=True)
            resource_manager.copy_project_memory(project_path, overwrite=True)
            version_manager.write(available_resource_version, __version__)
            click.echo(f"   Templates updated to v{available_resource_version}")

        if not resources_only:
            # Update package
            click.echo(f"{Fore.CYAN}📦 Package update requires manual intervention{Style.RESET_ALL}")
            click.echo("   Run: pip install --upgrade moai-adk")

        # Version synchronization
        if verbose:
            sync_manager = VersionSyncManager(str(project_path))
            click.echo(f"{Fore.CYAN}🔄 Synchronizing version information...{Style.RESET_ALL}")
            results = sync_manager.sync_all_versions(dry_run=True)
            for pattern, files in results.items():
                if files:
                    click.echo(f"   {pattern}: {len(files)} files")

        click.echo(f"\n{Fore.GREEN}✅ Update completed successfully{Style.RESET_ALL}")
        click.echo(f"Package version: v{current_version}")
        click.echo(f"Template version: v{available_resource_version}")

    except Exception as e:
        click.echo(f"{Fore.RED}❌ Update failed: {e}{Style.RESET_ALL}")
        if not no_backup:
            click.echo("You can restore from the backup if needed")
        sys.exit(1)


def create_mode_configuration(project_dir: Path, project_mode: str, quiet: bool = False) -> None:
    """@TASK:CONFIG-001 Create mode-specific configuration for MoAI-ADK project."""
    import json
    from datetime import datetime

    # Create .moai directory if it doesn't exist
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir(exist_ok=True)

    # Create mode-specific configuration
    config = {
        "project": {
            "name": project_dir.name,
            "mode": project_mode,
            "version": "0.1.0",
            "created": datetime.now().isoformat(),
            "constitution_version": "2.1"
        },
        "git_strategy": {
            "personal": {
                "auto_checkpoint": True,
                "checkpoint_interval": 300,  # 5 minutes
                "simplified_commits": True,
                "auto_push": False,
                "backup_before_sync": True,
                "conflict_strategy": "local_priority",
                "max_checkpoints": 50,
                "cleanup_days": 7
            },
            "team": {
                "auto_checkpoint": False,
                "structured_commits": True,
                "required_reviews": True,
                "auto_pr": True,
                "auto_sync": True,
                "sync_interval": 1800,  # 30 minutes
                "conflict_strategy": "remote_priority",
                "gitflow_strict": True
            }
        },
        "workflow": {
            "spec_auto_commit": True,
            "build_tdd_commits": True,
            "sync_living_docs": True,
            "constitution_check": True,
            "tag_tracking": True
        },
        "features": {
            "checkpoint_system": project_mode == "personal",
            "auto_rollback": project_mode == "personal",
            "smart_sync": True,
            "branch_management": True,
            "commit_automation": True
        }
    }

    # Save configuration
    config_file = moai_dir / "config.json"
    with open(config_file, 'w', encoding='utf-8') as f:
        json.dump(config, f, indent=2, ensure_ascii=False)

    if not quiet:
        mode_desc = {
            "personal": "개인 개발 최적화 (체크포인트, 간소화된 워크플로우)",
            "team": "팀 협업 최적화 (GitFlow, 구조화된 커밋, PR 자동화)"
        }
        click.echo(f"  {Fore.GREEN}✓{Style.RESET_ALL} {project_mode.title()} mode configured: {mode_desc[project_mode]}")

    # Create checkpoints directory for personal mode
    if project_mode == "personal":
        checkpoints_dir = moai_dir / "checkpoints"
        checkpoints_dir.mkdir(exist_ok=True)

        # Initialize checkpoint metadata
        metadata = {"checkpoints": []}
        metadata_file = checkpoints_dir / "metadata.json"
        with open(metadata_file, 'w', encoding='utf-8') as f:
            json.dump(metadata, f, indent=2)

        if not quiet:
            click.echo(f"  {Fore.GREEN}✓{Style.RESET_ALL} Checkpoint system initialized")


# Add all commands to the CLI group
cli.add_command(restore)
cli.add_command(doctor)
cli.add_command(init)
cli.add_command(help)
cli.add_command(status)
cli.add_command(update)
