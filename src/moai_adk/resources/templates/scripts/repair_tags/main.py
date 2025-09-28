#!/usr/bin/env python3
# @TASK:TAG-REPAIR-MAIN-001
"""
TAG Repair Main Entry Point

Command-line interface for the modularized TAG repair system.
Provides backward compatibility with the original script.
"""

import argparse
from pathlib import Path
from .core import TagRepairer


def main():
    """Main entry point for TAG repair system."""
    parser = argparse.ArgumentParser(description="MoAI-ADK TAG Auto-Repair System")
    parser.add_argument(
        "--project-root",
        type=Path,
        default=Path.cwd(),
        help="Project root directory (default: current directory)"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview repairs without applying them"
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose output"
    )

    args = parser.parse_args()

    # Initialize repairer
    repairer = TagRepairer(args.project_root)

    # Execute repair
    try:
        success = repairer.auto_repair_tags(dry_run=args.dry_run)
        if success:
            print("✅ TAG repair completed successfully")
        else:
            print("❌ TAG repair failed")
            return 1
    except Exception as e:
        print(f"❌ Error during TAG repair: {e}")
        return 1

    return 0


if __name__ == "__main__":
    exit(main())