#!/usr/bin/env python3
# @FEATURE:CLI-ENTRY-011
"""
🗿 MoAI-ADK CLI Main Entry Point

Main entry point for MoAI-ADK CLI when run as a module.
Supports both direct execution and module execution (python -m moai_adk.cli).
"""

import sys

from ..utils.logger import get_logger
from .commands import cli
from .helpers import validate_environment

logger = get_logger(__name__)


def main() -> None:
    """Main entry point for MoAI-ADK CLI."""
    try:
        # Basic environment check on startup
        if not validate_environment():
            sys.exit(1)

        # Run the CLI
        cli()

    except KeyboardInterrupt:
        logger.info("Operation cancelled by user")
        sys.exit(1)
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
