#!/usr/bin/env python3
"""SessionEnd Hook: Session Cleanup and Finalization

Claude Code Event: SessionEnd
Purpose: Clean up resources and finalize session when Claude Code exits
Execution: Triggered when Claude Code session ends

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

from handlers import handle_session_end  # noqa: E402
from utils.timeout import CrossPlatformTimeout  # noqa: E402
from utils.timeout import TimeoutError as PlatformTimeoutError  # noqa: E402


def main() -> None:
    """Main entry point for SessionEnd hook

    Currently a stub for future functionality:
    - Clear temporary caches
    - Save session metrics
    - Upload analytics (if enabled)
    - Cleanup background processes

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
        result = handle_session_end(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ SessionEnd cleanup timeout - session ending anyway",
        }
        print(json.dumps(timeout_response))
        print("SessionEnd hook timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionEnd JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"SessionEnd error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionEnd unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
