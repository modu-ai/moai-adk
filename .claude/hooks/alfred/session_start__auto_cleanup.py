#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-CLEAN | SPEC: Auto-cleanup old reports on session start
"""SessionStart Hook: Automatic Report Cleanup

Claude Code Event: SessionStart
Purpose: Automatically clean up old reports and maintain .moai directory health
Execution: Triggered when Claude Code session starts (if enabled in config)

Output: Continue execution with cleanup status notification
"""

import json
import sys
from pathlib import Path
from typing import Any

from utils.timeout import CrossPlatformTimeout
from utils.timeout import TimeoutError as PlatformTimeoutError

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

from handlers import handle_auto_cleanup


def main() -> None:
    """Main entry point for SessionStart auto-cleanup hook

    Functionality:
    - Check if auto-cleanup is enabled in .moai/config.json
    - Run report cleanup if enabled and needed
    - Maintain .moai directory health
    - Provide cleanup status feedback

    Exit Codes:
        0: Success (cleanup completed or skipped)
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Set 5-second timeout
    timeout = CrossPlatformTimeout(5)
    timeout.start()

    try:
        # Read JSON payload from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Call handler
        result = handle_auto_cleanup(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Auto-cleanup timeout - continuing session",
        }
        print(json.dumps(timeout_response))
        print("SessionStart auto-cleanup timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart auto-cleanup JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"Auto-cleanup error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart auto-cleanup unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()