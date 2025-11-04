#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-CLEAN | SPEC: Apply stored learnings to optimize current session
"""SessionStart Hook: Apply Stored Learnings

Claude Code Event: SessionStart
Purpose: Apply previously stored learning patterns to optimize current session
Execution: Triggered when Claude Code session starts (after daily analysis)

Output: Continue execution with learning-based optimizations
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

from handlers import handle_apply_learnings


def main() -> None:
    """Main entry point for SessionStart apply learnings hook

    Functionality:
    - Load stored learning patterns from memory
    - Identify relevant optimizations for current session
    - Apply learning-based adjustments to workflow
    - Provide optimization recommendations

    Exit Codes:
        0: Success (learnings applied or no learnings available)
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
        result = handle_apply_learnings(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Learning application timeout - continuing session",
        }
        print(json.dumps(timeout_response))
        print("SessionStart learnings application timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart learnings application JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"Learnings application error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart learnings application unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()