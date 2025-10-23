#!/usr/bin/env python3
# @CODE:HOOKS-REFACTOR-001 | SPEC: SPEC-HOOKS-REFACTOR-001.md
"""Alfred Hooks - Main entry point for MoAI-ADK Claude Code Hooks

A main entry point that routes Claude Code events to the appropriate handlers.

🏗️ Architecture:
┌─────────────────────────────────────────────────────────────┐
│ alfred_hooks.py (Router)                                    │
├─────────────────────────────────────────────────────────────┤
│ - CLI argument parsing                                      │
│ - JSON I/O (stdin/stdout)                                   │
│ - Event routing to handlers                                 │
└─────────────────────────────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ handlers/ (Event Handlers)                                  │
├─────────────────────────────────────────────────────────────┤
│ - session.py: SessionStart, SessionEnd                      │
│ - user.py: UserPromptSubmit                                 │
│ - tool.py: PreToolUse, PostToolUse                          │
│ - notification.py: Notification, Stop, SubagentStop         │
└─────────────────────────────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ core/ (Business Logic)                                      │
├─────────────────────────────────────────────────────────────┤
│ - project.py: Language detection, Git info, SPEC progress   │
│ - context.py: JIT Retrieval, workflow context               │
│ - checkpoint.py: Event-Driven Checkpoint system             │
│ - tags.py: TAG search/verification, library version cache   │
└─────────────────────────────────────────────────────────────┘

🛠️ Usage:
    python alfred_hooks.py <event_name> < payload.json

📣 Supported Events:
    - SessionStart: Start Session (display project status)
    - UserPromptSubmit: Prompt submission (JIT document loading)
    - PreToolUse: Before using the tool (automatically creates checkpoint)
    - SessionEnd, PostToolUse, Notification, Stop, SubagentStop

🚦 Exit Codes:
    - 0: Success
    - 1: Error (no arguments, JSON parsing failure, exception thrown)

🧪 TDD History:
    - RED: Module separation design, event routing test
    - GREEN: 1233 LOC → 9 items Module separation implementation (SRP compliance)
    - REFACTOR: Import optimization, enhanced error handling

Setup sys.path for package imports
"""

import json
import sys
from pathlib import Path
from typing import Any

from core import HookResult
from handlers import (
    handle_notification,
    handle_post_tool_use,
    handle_pre_tool_use,
    handle_session_end,
    handle_session_start,
    handle_stop,
    handle_subagent_stop,
    handle_user_prompt_submit,
)

# Add the hooks directory to sys.path to enable package imports
HOOKS_DIR = Path(__file__).parent
if str(HOOKS_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_DIR))


def main() -> None:
    """Main entry point - Claude Code Hook script

    Receives the event name as a CLI argument and reads the JSON payload through stdin.
    Calls the handler appropriate for the event and outputs the results to stdout as JSON.

    🛠️ Usage:
        python alfred_hooks.py <event_name> < payload.json

    📣 Supported Events:
        - SessionStart: Start Session (display project status)
        - UserPromptSubmit: Prompt submission (JIT document loading)
        - SessionEnd, PreToolUse, PostToolUse, Notification, Stop, SubagentStop

    🚦 Exit Codes:
        - 0: Success
        - 1: Error (no arguments, JSON parsing failure, exception thrown)

    📝 Examples:
        $ echo '{"cwd": "."}' | python alfred_hooks.py SessionStart
        {"message": "🚀 MoAI-ADK Session Started\\n...", ...}

    🗒️ Notes:
        - Claude Code is automatically called (no need for direct user execution)
        - JSON I/O processing through stdin/stdout
        - Print error message to stderr
        - UserPromptSubmit uses a special output schema (hookEventName + additionalContext)

    🧪 TDD History:
        - RED: Event routing, JSON I/O, error handling testing
        - GREEN: Handler map-based routing implementation
        - REFACTOR: Error message clarification, exit code standardization, UserPromptSubmit schema separation
    """
    # Check for event argument
    if len(sys.argv) < 2:
        print("Usage: alfred_hooks.py <event>", file=sys.stderr)
        sys.exit(1)

    event_name = sys.argv[1]

    try:
        # Read JSON from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data)

        cwd = data.get("cwd", ".")

        # Route to appropriate handler
        handlers = {
            "SessionStart": handle_session_start,
            "UserPromptSubmit": handle_user_prompt_submit,
            "SessionEnd": handle_session_end,
            "PreToolUse": handle_pre_tool_use,
            "PostToolUse": handle_post_tool_use,
            "Notification": handle_notification,
            "Stop": handle_stop,
            "SubagentStop": handle_subagent_stop,
        }

        handler = handlers.get(event_name)
        result = handler({"cwd": cwd, **data}) if handler else HookResult()

        # Output Hook result as JSON
        # Note: UserPromptSubmit uses to_user_prompt_submit_dict() for special schema
        if event_name == "UserPromptSubmit":
            print(json.dumps(result.to_user_prompt_submit_dict()))
        else:
            print(json.dumps(result.to_dict()))

        sys.exit(0)

    except json.JSONDecodeError as e:
        # Return valid Hook response even on JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"}
        }
        print(json.dumps(error_response))
        print(f"JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        # Return valid Hook response even on unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"Hook error: {e}"}
        }
        print(json.dumps(error_response))
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
