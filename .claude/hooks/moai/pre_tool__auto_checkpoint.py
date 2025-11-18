#!/usr/bin/env python3
"""PreToolUse Hook: Automatic Safety Checkpoint Creation

Claude Code Event: PreToolUse
Purpose: Detect risky operations and automatically create Git checkpoints before execution
Execution: Triggered before Edit, Write, or MultiEdit tools are used
Matcher: Edit|Write|MultiEdit

Output: System message with checkpoint information (if created)

Risky Operations Detected:
- Bash: rm -rf, git merge, git reset --hard
- Edit/Write: CLAUDE.md, config.json, critical files
- MultiEdit: Operations affecting ≥10 files
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

from handlers import handle_pre_tool_use  # noqa: E402
from utils.timeout import CrossPlatformTimeout  # noqa: E402
from utils.timeout import TimeoutError as PlatformTimeoutError  # noqa: E402


def main() -> None:
    """Main entry point for PreToolUse hook

    Analyzes tool usage and creates checkpoints for risky operations:
    1. Detects dangerous patterns (rm -rf, git reset, etc.)
    2. Creates Git checkpoint: checkpoint/before-{operation}-{timestamp}
    3. Logs checkpoint to .moai/checkpoints.log
    4. Returns guidance message to user

    Exit Codes:
        0: Success (checkpoint created or not needed)
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
        result = handle_pre_tool_use(data)

        # Output result as JSON
        print(json.dumps(result.to_dict()))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response (allow operation to continue)
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Checkpoint creation timeout - operation proceeding without checkpoint",
        }
        print(json.dumps(timeout_response))
        print("PreToolUse hook timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error - allow operation to continue
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"PreToolUse JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error - allow operation to continue
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"PreToolUse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"PreToolUse unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
