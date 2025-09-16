#!/usr/bin/env python3
"""
🗿 MoAI-ADK CLI Interface (Refactored)

Simplified main entry point for MoAI-ADK CLI.
All command definitions and helpers have been moved to separate modules.
"""

import sys
from .cli import cli
from .cli.helpers import validate_environment
from .logger import get_logger

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