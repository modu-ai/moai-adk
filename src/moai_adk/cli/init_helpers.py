"""
Initialization helper functions for MoAI-ADK CLI.

Split from oversized init() function (143 LOC) for TRUST principle compliance.
@FEATURE:CLI-INIT-HELPERS Modular initialization functions
@DESIGN:FUNCTION-SPLIT-001 Extracted from init() to meet 50 LOC limit
"""

import logging
import sys
from datetime import datetime
from pathlib import Path

import click
from colorama import Fore, Style

from ..config import Config
from ..install.installer import SimplifiedInstaller
from ..utils.logger import get_logger
from .helpers import (
    create_installation_backup,
    detect_potential_conflicts,
    validate_environment,
)

logger = get_logger(__name__)


def validate_initialization(
    project_path: str, personal: bool, team: bool, quiet: bool
) -> tuple[Path, str]:
    """
    Validate initialization parameters and determine project mode.

    Args:
        project_path: Target project directory path
        personal: Personal mode flag
        team: Team mode flag
        quiet: Quiet mode flag

    Returns:
        Tuple of (project_dir, project_mode)

    Raises:
        SystemExit: If validation fails
    """
    project_dir = Path(project_path).resolve()

    # Determine project mode
    if team and personal:
        click.echo(
            f"{Fore.RED}❌ Cannot specify both --personal and --team modes{Style.RESET_ALL}"
        )
        sys.exit(1)

    # Default to personal mode if no mode specified
    project_mode = "team" if team else "personal"

    if not quiet:
        mode_icon = "🏢" if project_mode == "team" else "👤"
        click.echo(
            f"{Fore.CYAN}{mode_icon} Initializing in {project_mode} mode{Style.RESET_ALL}"
        )

    # Validate environment
    if not validate_environment():
        if not quiet:
            click.echo(f"{Fore.RED}❌ Environment validation failed{Style.RESET_ALL}")
        sys.exit(1)

    return project_dir, project_mode


def handle_interactive_mode(project_dir: Path, interactive: bool) -> bool:
    """
    Handle interactive wizard mode if requested.

    Args:
        project_dir: Project directory path
        interactive: Whether to use interactive mode

    Returns:
        True if interactive mode was used (and should exit), False otherwise
    """
    if not interactive:
        return False

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
        return True

    except KeyboardInterrupt:
        click.echo(f"\n{Fore.YELLOW}⚠️  Setup cancelled{Style.RESET_ALL}")
        return True

    except Exception as e:
        click.echo(f"{Fore.RED}❌ Interactive setup error: {e}{Style.RESET_ALL}")
        return True


def setup_project_directory(
    project_dir: Path,
    project_mode: str,
    backup: bool,
    force: bool,
    force_copy: bool,
    quiet: bool,
) -> bool:
    """
    Setup project directory structure and files.

    Args:
        project_dir: Target project directory
        project_mode: Project mode ("personal" or "team")
        backup: Whether to create backup
        force: Force overwrite existing files
        force_copy: Force copy resources
        quiet: Quiet mode flag

    Returns:
        True if setup successful, False otherwise
    """
    try:
        # Check for conflicts
        conflicts = detect_potential_conflicts(project_dir)
        if conflicts and not force:
            if not quiet:
                click.echo(
                    f"{Fore.YELLOW}⚠️ Potential conflicts detected:{Style.RESET_ALL}"
                )
                for conflict in conflicts:
                    click.echo(f"   - {conflict}")
                click.echo("Use --force to override or resolve conflicts first.")
            return False

        # Create backup if requested
        if backup and project_dir.exists():
            if not quiet:
                click.echo(f"{Fore.BLUE}📦 Creating backup...{Style.RESET_ALL}")
            create_installation_backup(project_dir)

        # Create project directory if it doesn't exist
        if not project_dir.exists():
            project_dir.mkdir(parents=True, exist_ok=True)
            if not quiet:
                click.echo(
                    f"{Fore.GREEN}📁 Created project directory: {project_dir}{Style.RESET_ALL}"
                )

        return True

    except Exception as e:
        if not quiet:
            click.echo(
                f"{Fore.RED}❌ Error setting up project directory: {e}{Style.RESET_ALL}"
            )
        logger.error(f"Project directory setup failed: {e}")
        return False


def finalize_installation(
    project_dir: Path, project_mode: str, force_copy: bool, quiet: bool
) -> None:
    """
    Finalize installation with installer and provide completion feedback.

    Args:
        project_dir: Project directory path
        project_mode: Project mode ("personal" or "team")
        force_copy: Force copy resources flag
        quiet: Quiet mode flag
    """
    try:
        # Configure logging for quiet mode
        if quiet:
            # Silence all moai_adk loggers
            logging.getLogger("moai_adk").setLevel(logging.CRITICAL)
            for logger_name in logging.Logger.manager.loggerDict:
                if logger_name.startswith("moai_adk"):
                    logging.getLogger(logger_name).setLevel(logging.CRITICAL)

        # Run installation
        # @FEATURE:CONFIG-BASED-INIT-001 Create Config object for SimplifiedInstaller
        try:
            config = Config(
                name=project_dir.name,
                path=str(project_dir),
                force_copy=force_copy,
                silent=quiet,  # Map quiet parameter to silent attribute
            )
            installer = SimplifiedInstaller(config)
        except Exception as e:
            if not quiet:
                click.echo(f"{Fore.RED}❌ Configuration error: {e}{Style.RESET_ALL}")
            logger.error(f"Failed to create installation config: {e}")
            sys.exit(1)

        result = installer.install()

        if result.success:
            # Create mode-specific configuration
            create_mode_configuration(project_dir, project_mode, quiet)

            if not quiet:
                click.echo(
                    f"\n{Fore.GREEN}🎉 MoAI-ADK initialization completed!{Style.RESET_ALL}"
                )
                click.echo(f"{Fore.CYAN}📍 Project: {project_dir}{Style.RESET_ALL}")
                click.echo(f"{Fore.CYAN}⚙️  Mode: {project_mode}{Style.RESET_ALL}")

                if result.next_steps:
                    click.echo(f"\n{Fore.YELLOW}📋 Next steps:{Style.RESET_ALL}")
                    for step in result.next_steps:
                        click.echo(f"   {step}")
        else:
            if not quiet:
                click.echo(f"{Fore.RED}❌ Installation failed{Style.RESET_ALL}")
                if result.errors:
                    for error in result.errors:
                        click.echo(f"   {error}")

            # Exit with error code
            sys.exit(1)

    except Exception as e:
        if not quiet:
            click.echo(f"{Fore.RED}❌ Installation error: {e}{Style.RESET_ALL}")
        logger.error(f"Installation failed: {e}")
        sys.exit(1)


def create_mode_configuration(
    project_dir: Path, project_mode: str, quiet: bool = False
) -> None:
    """
    Create mode-specific configuration for the project.

    Args:
        project_dir: Project directory path
        project_mode: Project mode ("personal" or "team")
        quiet: Quiet mode flag
    """
    try:
        # Create mode-specific configuration
        config = {
            "project": {
                "name": project_dir.name,
                "mode": project_mode,
                "version": "0.1.9",
                "created": datetime.now().isoformat(),
                "constitution_version": "2.1",
            },
            "git_strategy": {
                project_mode: {
                    "auto_commit": project_mode == "personal",
                    "auto_pr": project_mode == "team",
                    "develop_branch": "develop" if project_mode == "team" else "main",
                    "feature_prefix": "feature/SPEC-"
                    if project_mode == "team"
                    else "feature/",
                    "use_gitflow": project_mode == "team",
                }
            },
            "created_at": datetime.now().isoformat(),
            "moai_adk_version": "0.1.9",
        }

        config_path = project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        import json

        with open(config_path, "w") as f:
            json.dump(config, f, indent=2)

        if not quiet:
            logger.info(f"Created {project_mode} mode configuration")

    except Exception as e:
        logger.error(f"Failed to create mode configuration: {e}")
        if not quiet:
            click.echo(
                f"{Fore.YELLOW}⚠️ Warning: Failed to create mode configuration{Style.RESET_ALL}"
            )
