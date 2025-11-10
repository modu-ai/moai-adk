#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-CKPT | SPEC: Individual hook files for better UX
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

from core.timeout import CrossPlatformTimeout  # noqa: E402
from core.timeout import TimeoutError as PlatformTimeoutError  # noqa: E402
from core.error_handler import HookErrorHandler  # noqa: E402

# Try to import handler with fallback
try:
    from handlers import handle_pre_tool_use  # noqa: E402
    HANDLER_AVAILABLE = True
except ImportError:
    HANDLER_AVAILABLE = False

    def handle_pre_tool_use(data):
        # Fallback handler that simulates checkpoint creation
        class MockResult:
            def to_dict(self):
                return {
                    "checkpoint_created": False,
                    "message": "Handler not available, skipping checkpoint"
                }
        return MockResult()


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
    # Initialize error handler
    error_handler = HookErrorHandler("pre_tool__auto_checkpoint")

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
        response = error_handler.create_success(
            message="Checkpoint creation completed",
            data=result.to_dict()
        )
        error_handler.print_and_exit(response, exit_code=0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response (allow operation to continue)
        response = error_handler.handle_timeout("checkpoint creation")
        error_handler.print_and_exit(response, exit_code=1)

    except json.JSONDecodeError as e:
        # JSON parse error - allow operation to continue
        response = error_handler.handle_json_error(e, "PreToolUse")
        error_handler.print_and_exit(response, exit_code=1)

    except Exception as e:
        # Unexpected error - allow operation to continue
        response = error_handler.handle_generic_error(e, "PreToolUse")
        error_handler.print_and_exit(response, exit_code=1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
