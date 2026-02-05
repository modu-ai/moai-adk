"""Standalone CLI entry point for moai-worktree command

This module provides a standalone entry point for the moai-worktree CLI,
allowing users to invoke worktree commands directly via `moai-worktree`
instead of `moai-adk worktree`.

Usage:
    moai-worktree new SPEC-001
    moai-worktree list
    moai-worktree go SPEC-001
    etc.
"""

import sys

# Import the worktree CLI group from the main module
from moai_adk.cli.worktree.cli import worktree


def main() -> int:
    """Standalone entry point for moai-worktree CLI"""
    try:
        worktree(standalone_mode=False)
        return 0
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
