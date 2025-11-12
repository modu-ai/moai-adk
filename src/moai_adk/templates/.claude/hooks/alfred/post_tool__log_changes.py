#!/usr/bin/env python3
"""PostToolUse Hook: Log Tool Usage and Changes

Claude Code Event: PostToolUse
Purpose: Log tool execution results and track changes for audit trail
Execution: Triggered after Edit, Write, or MultiEdit tools are used
Matcher: Edit|Write|MultiEdit

Output: Continue execution (currently a stub for future enhancements)
"""

import json
import sys
from pathlib import Path
from typing import Any

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

from handlers import handle_post_tool_use  # noqa: E402
from utils.timeout import CrossPlatformTimeout  # noqa: E402
from utils.timeout import TimeoutError as PlatformTimeoutError  # noqa: E402


def main() -> None:
    """Main entry point for PostToolUse hook

    Currently a stub for future functionality:
    - Change tracking and audit logging
    - Metrics collection (files modified, lines changed)
    - Integration with external monitoring systems

    Exit Codes:
        0: Success
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
        result = handle_post_tool_use(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ PostToolUse timeout - continuing",
        }
        print(json.dumps(timeout_response))
        print("PostToolUse hook timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"PostToolUse JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"PostToolUse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"PostToolUse unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
