#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-001 | SPEC: Individual hook files for better UX
"""SessionStart Hook: Show Project Information

Claude Code Event: SessionStart
Purpose: Display project status, language, Git info, and SPEC progress when session starts
Execution: Triggered automatically when Claude Code session begins

Output: System message with formatted project summary
"""

import json
import signal
import sys
from pathlib import Path
from typing import Any

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

from handlers import handle_session_start


class HookTimeoutError(Exception):
    """Hook execution timeout exception"""
    pass


def _timeout_handler(signum, frame):
    """Signal handler for 5-second timeout"""
    raise HookTimeoutError("Hook execution exceeded 5-second timeout")


def main() -> None:
    """Main entry point for SessionStart hook

    Displays project information including:
    - Programming language
    - Git branch and status
    - SPEC progress (completed/total)
    - Recent checkpoints

    Exit Codes:
        0: Success
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Set 5-second timeout
    signal.signal(signal.SIGALRM, _timeout_handler)
    signal.alarm(5)

    try:
        # Read JSON payload from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Call handler
        result = handle_session_start(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except HookTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Session start timeout - continuing without project info"
        }
        print(json.dumps(timeout_response))
        print("SessionStart hook timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"}
        }
        print(json.dumps(error_response))
        print(f"SessionStart JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"SessionStart error: {e}"}
        }
        print(json.dumps(error_response))
        print(f"SessionStart unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        signal.alarm(0)


if __name__ == "__main__":
    main()
