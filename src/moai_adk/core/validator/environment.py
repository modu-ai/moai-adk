#!/usr/bin/env python3
# @TASK:VALIDATE-ENVIRONMENT-001
"""
Environment Validation Module

Validates system environment requirements like Python, Claude Code, and Git.
Focuses on external dependency verification.
"""

import subprocess
import sys
from pathlib import Path
import click
from colorama import Fore, Style
from ...utils.logger import get_logger

logger = get_logger(__name__)


def validate_python_version(min_version: tuple[int, int] = (3, 8)) -> bool:
    """Validate Python version meets minimum requirements."""
    current = sys.version_info[:2]
    if current >= min_version:
        logger.info(f"✅ Python {current[0]}.{current[1]} meets requirement ≥{min_version[0]}.{min_version[1]}")
        return True
    else:
        logger.error(f"❌ Python {current[0]}.{current[1]} below requirement ≥{min_version[0]}.{min_version[1]}")
        return False


def validate_claude_code() -> bool:
    """Validate Claude Code environment availability."""
    try:
        # Check if we're running in Claude Code (simplified check)
        import os
        if os.environ.get('CLAUDE_PROJECT_DIR'):
            logger.info("✅ Claude Code environment detected")
            return True
        else:
            logger.warning("⚠️ Claude Code environment not detected")
            return False
    except Exception as e:
        logger.error(f"❌ Error checking Claude Code: {e}")
        return False


def validate_git_repository(path: Path) -> bool:
    """Validate Git repository exists and is functional."""
    git_dir = path / '.git'
    if not git_dir.exists():
        logger.error(f"❌ No Git repository found at {path}")
        return False

    try:
        result = subprocess.run(['git', 'status'], cwd=path, capture_output=True, timeout=5)
        if result.returncode == 0:
            logger.info("✅ Git repository is functional")
            return True
        else:
            logger.error("❌ Git repository has issues")
            return False
    except (subprocess.TimeoutExpired, FileNotFoundError):
        logger.error("❌ Git command failed or not available")
        return False


def validate_environment() -> bool:
    """Validate overall environment readiness."""
    checks = [
        validate_python_version(),
        validate_claude_code(),
        validate_git_repository(Path.cwd()),
    ]

    passed = sum(checks)
    total = len(checks)

    if passed == total:
        click.echo(f"{Fore.GREEN}✅ Environment validation passed ({passed}/{total}){Style.RESET_ALL}")
        return True
    else:
        click.echo(f"{Fore.YELLOW}⚠️ Environment validation partial ({passed}/{total}){Style.RESET_ALL}")
        return passed >= 2  # At least 2 out of 3 should pass