#!/usr/bin/env python3
# @TASK:VALIDATE-PROJECT-001
"""
Project Validation Module

Validates project structure, configuration, and readiness for development.
Focuses on project-level validation checks.
"""

from pathlib import Path
from typing import Dict
import click
from colorama import Fore, Style
from ...utils.logger import get_logger

logger = get_logger(__name__)


def validate_project_structure(project_path: Path) -> Dict[str, bool]:
    """Validate basic project structure requirements."""
    checks = {}

    # Check essential files
    essential_files = [
        "README.md",
        "pyproject.toml",
        ".gitignore",
    ]

    for file_name in essential_files:
        file_path = project_path / file_name
        checks[f"has_{file_name.replace('.', '_')}"] = file_path.exists()

    # Check directory structure
    essential_dirs = [
        "src",
        "tests",
    ]

    for dir_name in essential_dirs:
        dir_path = project_path / dir_name
        checks[f"has_{dir_name}_dir"] = dir_path.exists() and dir_path.is_dir()

    # Calculate pass rate
    passed = sum(checks.values())
    total = len(checks)
    checks["structure_valid"] = passed >= (total * 0.7)  # 70% pass rate

    logger.info(f"Project structure validation: {passed}/{total} checks passed")
    return checks


def validate_project_readiness(project_path: Path) -> bool:
    """Validate project readiness for MoAI-ADK integration."""
    try:
        # Validate basic structure
        structure_checks = validate_project_structure(project_path)

        # Check for existing MoAI setup
        moai_dir = project_path / ".moai"
        has_moai = moai_dir.exists()

        # Check for Claude Code setup
        claude_dir = project_path / ".claude"
        has_claude = claude_dir.exists()

        # Determine readiness
        structure_ready = structure_checks.get("structure_valid", False)
        integration_ready = has_moai or has_claude

        overall_ready = structure_ready and integration_ready

        if overall_ready:
            click.echo(f"{Fore.GREEN}✅ Project is ready for MoAI-ADK{Style.RESET_ALL}")
        elif structure_ready:
            click.echo(f"{Fore.YELLOW}⚠️ Project structure is good, needs MoAI integration{Style.RESET_ALL}")
        else:
            click.echo(f"{Fore.RED}❌ Project needs structure improvements{Style.RESET_ALL}")

        return overall_ready

    except Exception as e:
        logger.error(f"Error validating project readiness: {e}")
        return False