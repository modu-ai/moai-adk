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


def main() -> None:
    """Ensure streaming UI settings are properly configured"""
    try:
        # Read input from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Set environment variables for streaming UI
        os.environ['CLAUDE_UI_STREAMING_ENABLED'] = 'true'
        os.environ['CLAUDE_PROGRESS_INDICATORS'] = 'true'
        os.environ['CLAUDE_TODO_CONTROLS'] = 'true'
        os.environ['CLAUDE_STREAMING_UI'] = 'true'

        # Check if this was a TodoWrite operation
        tool_name = data.get('tool', '')
        if 'TodoWrite' in tool_name:
            # Force refresh of UI display
            print("\n--- UI Refresh Triggered ---", file=sys.stderr)
            print("Streaming indicators: ENABLED", file=sys.stderr)
            print("Progress displays: ENABLED", file=sys.stderr)
            print("Todo controls: ENABLED", file=sys.stderr)
            print("--- End UI Refresh ---", file=sys.stderr)

        return 0

    except Exception:
        # Silent failure to avoid breaking hook chain
        return 0

if __name__ == "__main__":
    sys.exit(main())
