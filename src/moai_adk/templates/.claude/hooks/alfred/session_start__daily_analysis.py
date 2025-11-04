#!/usr/bin/env python3
# @CODE:HOOKS-DAILY-ANALYSIS-001 | Daily session meta-analysis
"""SessionStart Hook: Run Daily Session Analysis

Claude Code Event: SessionStart
Purpose: Automatically run daily session analysis on previous day's Claude Code session logs
Execution: Triggered automatically when Claude Code session begins (silent background operation)

Output: No user-visible message (analysis report saved to .moai/reports/daily-analysis-YYYY-MM-DD.md)
"""

import json
import subprocess
import sys
from pathlib import Path
from datetime import datetime
from typing import Any

from utils.timeout import CrossPlatformTimeout
from utils.timeout import TimeoutError as PlatformTimeoutError

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

from handlers.daily_analysis import handle_daily_analysis


def main() -> None:
    """Main entry point for Daily Analysis SessionStart hook

    Runs session analysis for the previous day's Claude Code session logs.
    Silent operation - no user-visible message unless error occurs.

    Exit Codes:
        0: Success or skipped (already analyzed today)
        1: Error (timeout, subprocess failure, etc.)
    """
    # Set 5-second timeout (inclusive of subprocess execution)
    timeout = CrossPlatformTimeout(5)
    timeout.start()

    try:
        # Read JSON payload from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Call handler - silent operation
        result = handle_daily_analysis(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response (silent)
        timeout_response: dict[str, Any] = {
            "continue": True,
        }
        print(json.dumps(timeout_response))
        print("[DEBUG] Daily analysis timeout (5s exceeded)", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error (silent)
        error_response: dict[str, Any] = {
            "continue": True,
        }
        print(json.dumps(error_response))
        print(f"[DEBUG] Daily analysis JSON error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error (silent)
        error_response: dict[str, Any] = {
            "continue": True,
        }
        print(json.dumps(error_response))
        print(f"[DEBUG] Daily analysis error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
