#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-CLEAN | SPEC: Daily session analysis on session start
"""SessionStart Hook: Daily Session Analysis

Claude Code Event: SessionStart
Purpose: Trigger daily session analysis if enough time has passed
Execution: Triggered when Claude Code session starts (daily intervals)

Output: Continue execution with analysis status notification
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

from handlers import handle_daily_analysis


def main() -> None:
    """Main entry point for SessionStart daily analysis hook

    Functionality:
    - Check if daily analysis is due (1+ day since last analysis)
    - Run session analysis for the past day
    - Generate improvement suggestions
    - Update analysis status in .moai/

    Exit Codes:
        0: Success (analysis completed or skipped)
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Set 10-second timeout for analysis
    timeout = CrossPlatformTimeout(10)
    timeout.start()

    try:
        # Read JSON payload from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Call handler
        result = handle_daily_analysis(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Daily analysis timeout - continuing session",
        }
        print(json.dumps(timeout_response))
        print("SessionStart daily analysis timeout after 10 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart daily analysis JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"Daily analysis error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart daily analysis unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()