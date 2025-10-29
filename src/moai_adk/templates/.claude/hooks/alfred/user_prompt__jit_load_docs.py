#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-001 | SPEC: Individual hook files for better UX
"""UserPromptSubmit Hook: Just-In-Time Document Loading

Claude Code Event: UserPromptSubmit
Purpose: Analyze user prompt and recommend relevant documents to load into context
Execution: Triggered when user submits a prompt

Output: additionalContext with document path suggestions
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

from handlers import handle_user_prompt_submit


class HookTimeoutError(Exception):
    """Hook execution timeout exception"""
    pass


def _timeout_handler(signum, frame):
    """Signal handler for 5-second timeout"""
    raise HookTimeoutError("Hook execution exceeded 5-second timeout")


def main() -> None:
    """Main entry point for UserPromptSubmit hook

    Analyzes user prompt patterns and recommends relevant documents:
    - /alfred:1-plan → spec-metadata.md
    - /alfred:2-run → development-guide.md
    - @TAG mentions → TAG documentation
    - SPEC references → related SPEC files

    Exit Codes:
        0: Success
        1: Error (timeout, JSON parse failure, handler exception)

    Note: Uses special output schema for UserPromptSubmit:
    {
        "continue": true,
        "hookSpecificOutput": {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": "Document path suggestions..."
        }
    }
    """
    # Set 5-second timeout
    signal.signal(signal.SIGALRM, _timeout_handler)
    signal.alarm(5)

    try:
        # Read JSON payload from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Call handler
        result = handle_user_prompt_submit(data)

        # Output result using UserPromptSubmit-specific schema
        print(json.dumps(result.to_user_prompt_submit_dict()))
        sys.exit(0)

    except HookTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": "⚠️ JIT context timeout - continuing without suggestions"
            }
        }
        print(json.dumps(timeout_response))
        print("UserPromptSubmit hook timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "error": f"JSON parse error: {e}"
            }
        }
        print(json.dumps(error_response))
        print(f"UserPromptSubmit JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "error": f"UserPromptSubmit error: {e}"
            }
        }
        print(json.dumps(error_response))
        print(f"UserPromptSubmit unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        signal.alarm(0)


if __name__ == "__main__":
    main()
