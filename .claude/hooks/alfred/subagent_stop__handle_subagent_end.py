#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-SUBAGENT | SPEC: Individual hook files for better UX
"""SubagentStop Hook: Handle Sub-agent Termination

Claude Code Event: SubagentStop
Purpose: Handle cleanup when a sub-agent execution completes or is terminated
Execution: Triggered when Task tool completes (sub-agent finishes)

Output: Continue execution (currently a stub for future enhancements)
"""

import json
import sys
from pathlib import Path
from typing import Any

from utils.timeout import CrossPlatformTimeout, TimeoutError as PlatformTimeoutError

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

from handlers import handle_subagent_stop




def main() -> None:
    """Main entry point for SubagentStop hook

    Currently a stub for future functionality:
    - Collect sub-agent execution metrics
    - Log sub-agent results and errors
    - Update workflow state based on sub-agent outcome
    - Trigger follow-up actions based on results

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
        result = handle_subagent_stop(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ SubagentStop handler timeout",
        }
        print(json.dumps(timeout_response))
        print("SubagentStop hook timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SubagentStop JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"SubagentStop error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SubagentStop unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
