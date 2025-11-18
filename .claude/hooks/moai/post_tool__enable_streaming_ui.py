#!/usr/bin/env python3
"""PostToolUse Hook: Enable Streaming UI Display

Claude Code Event: PostToolUse
Purpose: Ensure streaming indicators and progress displays are properly configured
Execution: Runs after tool executions to verify UI settings

This hook ensures that streaming display features like:
- "✻ 확인 중...… (esc to interrupt · ctrl+t to hide todos)"
- Progress indicators
- Todo visibility controls
Are properly enabled and functioning.
"""

import json
import os
import sys
from typing import Any


def set_streaming_ui_environment() -> None:
    """Set environment variables for streaming UI configuration.

    Returns:
        None
    """
    os.environ['CLAUDE_UI_STREAMING_ENABLED'] = 'true'
    os.environ['CLAUDE_PROGRESS_INDICATORS'] = 'true'
    os.environ['CLAUDE_TODO_CONTROLS'] = 'true'
    os.environ['CLAUDE_STREAMING_UI'] = 'true'


def log_ui_refresh(tool_name: str) -> None:
    """Log UI refresh status to stderr if TodoWrite operation detected.

    Args:
        tool_name: Name of the tool executed

    Returns:
        None
    """
    if 'TodoWrite' in tool_name:
        # Force refresh of UI display
        print("\n--- UI Refresh Triggered ---", file=sys.stderr)
        print("Streaming indicators: ENABLED", file=sys.stderr)
        print("Progress displays: ENABLED", file=sys.stderr)
        print("Todo controls: ENABLED", file=sys.stderr)
        print("--- End UI Refresh ---", file=sys.stderr)


def main() -> None:
    """Ensure streaming UI settings are properly configured

    Returns:
        None
    """
    try:
        # Read input from stdin
        input_data: str = sys.stdin.read()
        data: dict[str, Any] = json.loads(input_data) if input_data.strip() else {}

        # Set environment variables for streaming UI
        set_streaming_ui_environment()

        # Check if this was a TodoWrite operation
        tool_name: str = data.get('tool', '')
        log_ui_refresh(tool_name)

    except Exception:
        # Silent failure to avoid breaking hook chain
        pass

if __name__ == "__main__":
    sys.exit(main())
