"""
@FEATURE:LOGGER-UTILITIES-001 🗿 MoAI-ADK Logging Utilities
@TASK:STRUCTURED-LOGGING-001 Provides structured logging functionality for MoAI-ADK operations

This module provides:
- Configured logger instances with color formatting
- Project-specific logging setup
- Silent mode support for automated operations
- Structured logging for debugging and audit trails
"""

import logging
import sys
from pathlib import Path

from colorama import Fore, Style


def get_logger(name: str = "moai-adk", level: str = "WARNING") -> logging.Logger:
    """
    @TASK:GET-LOGGER-001 Get a configured logger instance

    Args:
        name: Logger name
        level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)

    Returns:
        Configured logger instance
    """
    logger = logging.getLogger(name)

    if not logger.handlers:
        handler = logging.StreamHandler(sys.stdout)
        formatter = logging.Formatter(
            f'{Fore.BLUE}%(asctime)s{Style.RESET_ALL} - '
            f'{Fore.GREEN}%(name)s{Style.RESET_ALL} - '
            f'%(levelname)s - %(message)s',
            datefmt='%H:%M:%S'
        )
        handler.setFormatter(formatter)
        logger.addHandler(handler)
        logger.setLevel(getattr(logging, level.upper()))
        logger.propagate = False

    return logger


def setup_project_logging(project_path: Path, silent: bool = False) -> logging.Logger:
    """
    @TASK:PROJECT-LOGGING-001 Setup project-specific logging

    Args:
        project_path: Path to the project
        silent: Whether to enable silent mode

    Returns:
        Project logger instance
    """
    logger = get_logger(f"moai-adk.{project_path.name}")

    if silent:
        logger.setLevel(logging.ERROR)
    else:
        logger.setLevel(logging.INFO)

    return logger
