#!/usr/bin/env python3
"""
🗿 MoAI-ADK Post-Installation Script

Provides post-installation functionality for MoAI-ADK.
Since pyproject.toml doesn't support post-install hooks directly,
this script serves as an alternative for setting up global resources.

Usage:
    pip install moai-adk
    moai-install  # Run after installation
"""

import sys
from pathlib import Path

import click
from colorama import Fore, Style, init

# Note: global_installer removed in favor of package-based resources
from ..utils.logger import get_logger
from .._version import __version__

# Initialize colorama for cross-platform colored output
init(autoreset=True)

logger = get_logger(__name__)


def print_post_install_banner():
    """Print post-installation banner."""
    banner = f"""
{Fore.CYAN}🗿 MoAI-ADK v{__version__} Post-Installation{Style.RESET_ALL}
{Fore.BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━{Style.RESET_ALL}

Setting up global resources for optimal MoAI-ADK experience...
"""
    print(banner)




@click.command()
@click.option(
    '--force',
    is_flag=True,
    help='Force reinstallation even if already installed'
)
@click.option(
    '--quiet',
    is_flag=True,
    help='Suppress output messages'
)
def main(force: bool, quiet: bool) -> None:
    """
    MoAI-ADK post-installation setup.

    Note: As of v0.1.13+, resources are embedded in the package.
    No separate installation needed.
    """
    try:
        if not quiet:
            print_post_install_banner()

        # Resources are now embedded in package
        if not quiet:
            print(f"{Fore.GREEN}✅ MoAI-ADK resources are embedded in the package.{Style.RESET_ALL}")
            print(f"   No separate installation needed!")
            print(f"   Use {Fore.WHITE}moai init{Style.RESET_ALL} to set up new projects.")

    except KeyboardInterrupt:
        if not quiet:
            print(f"\n{Fore.YELLOW}⚠️  Installation cancelled by user.{Style.RESET_ALL}")
        sys.exit(1)
    except Exception as error:
        logger.error("Post-installation failed: %s", error)
        if not quiet:
            print(f"{Fore.RED}❌ Error: {error}{Style.RESET_ALL}")
        sys.exit(1)


def auto_install_on_first_run() -> bool:
    """
    Check if resources are available (실제 리소스 존재 여부 검증).

    Returns:
        bool: 리소스가 실제로 사용 가능한지 여부
    """
    try:
        # 실제 리소스 존재 여부 검증
        from importlib import resources
        from ..install.resource_manager import ResourceManager

        # 패키지 리소스 접근 가능 여부 확인
        resource_manager = ResourceManager()
        templates = resource_manager.list_templates()

        # 핵심 템플릿들이 존재하는지 확인
        required_templates = ['.claude', '.moai', 'CLAUDE.md']
        for template in required_templates:
            if template not in templates:
                logger.error(f"Required template missing: {template}")
                return False

        logger.debug(f"Resource check passed. Found {len(templates)} templates.")
        return True

    except Exception as error:
        logger.error("Resource check failed: %s", error)
        return False


if __name__ == "__main__":
    main()