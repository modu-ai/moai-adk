#!/usr/bin/env python3

"""PreToolUse Hook: Agent Permission Validation

Claude Code Event: PreToolUse
Purpose: Validate agent permissionMode values before Task delegation
Execution: Triggered before Task tool is used
Matcher: Task

Output: System message with validation errors and fix instructions

Validation Rules:
- Check if Task delegates to a MoAI agent
- Read agent file's YAML frontmatter
- Validate permissionMode value against allowed list
- Block execution if invalid permissionMode detected
- Provide fix script instructions
"""

import json
import re
import sys
from pathlib import Path
from typing import Any, Dict, Optional

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

try:
    from lib.timeout import CrossPlatformTimeout  # noqa: E402
    from lib.timeout import TimeoutError as PlatformTimeoutError  # noqa: E402
except ImportError:
    # Fallback for timeout if shared module unavailable
    import signal

    class PlatformTimeoutError(Exception):  # type: ignore[no-redef]
        pass

    class CrossPlatformTimeout:  # type: ignore[no-redef]
        def __init__(self, seconds: int) -> None:
            self.seconds = seconds

        def start(self) -> None:
            signal.alarm(int(self.seconds))

        def cancel(self) -> None:
            signal.alarm(0)


# Valid permissionMode values per Claude Code spec
VALID_PERMISSION_MODES = {
    'acceptEdits',
    'bypassPermissions',
    'default',
    'dontAsk',
    'plan'
}

# Invalid values and their recommended replacements
INVALID_MODE_MAPPING = {
    'auto': 'dontAsk',
    'ask': 'default'
}


def extract_permission_mode(agent_file_path: Path) -> Optional[str]:
    """Extract permissionMode from agent YAML frontmatter

    Args:
        agent_file_path: Path to agent .md file

    Returns:
        permissionMode value or None if not found
    """
    try:
        content = agent_file_path.read_text(encoding='utf-8')

        # Match permissionMode in YAML frontmatter
        match = re.search(r'^permissionMode:\s*(\S+)\s*$', content, re.MULTILINE)

        if match:
            return match.group(1).strip()

        return None

    except Exception:
        return None


def validate_agent_permission_mode(subagent_type: str) -> Dict[str, Any]:
    """Validate agent's permissionMode value

    Args:
        subagent_type: Agent identifier (e.g., 'spec-builder')

    Returns:
        Validation result dictionary
    """
    result: Dict[str, Any] = {
        "valid": True,
        "agent_found": False,
        "permission_mode": None,
        "is_invalid": False,
        "suggested_fix": None,
        "error_message": None
    }

    # Locate agent file
    agent_file = Path(f".claude/agents/moai/{subagent_type}.md")

    if not agent_file.exists():
        # Agent file not found - allow operation (might be built-in agent)
        result["valid"] = True
        return result

    result["agent_found"] = True

    # Extract permissionMode
    permission_mode = extract_permission_mode(agent_file)

    if permission_mode is None:
        # No permissionMode specified - allow operation
        result["valid"] = True
        return result

    result["permission_mode"] = permission_mode

    # Validate against allowed values
    if permission_mode in VALID_PERMISSION_MODES:
        result["valid"] = True
        return result

    # Invalid permissionMode detected
    result["valid"] = False
    result["is_invalid"] = True

    # Determine suggested fix
    suggested = INVALID_MODE_MAPPING.get(permission_mode, "default")
    result["suggested_fix"] = suggested

    result["error_message"] = (
        f"❌ Agent '{subagent_type}' has invalid permissionMode: '{permission_mode}'\n"
        f"\n"
        f"Valid options:\n"
        f"  - acceptEdits      (auto-accept file edits)\n"
        f"  - bypassPermissions (skip permission checks)\n"
        f"  - default          (standard permission flow)\n"
        f"  - dontAsk          (auto-proceed with restrictions)\n"
        f"  - plan             (plan mode only)\n"
        f"\n"
        f"Suggested replacement: '{suggested}'\n"
        f"\n"
        f"Fix this agent:\n"
        f"  uv run .moai/scripts/fix-agent-permissions.py\n"
        f"\n"
        f"Or manually edit:\n"
        f"  {agent_file}\n"
        f"  Change: permissionMode: {permission_mode}\n"
        f"  To:     permissionMode: {suggested}"
    )

    return result


def handle_pre_tool_use(payload: Dict) -> Dict[str, Any]:
    """Handle PreToolUse event for agent permission validation

    Args:
        payload: Hook payload containing tool name and parameters

    Returns:
        Hook response dictionary
    """
    # Get tool name and parameters
    tool_name = payload.get("tool", {}).get("name", "")
    parameters = payload.get("tool", {}).get("parameters", {})

    # Only validate Task operations
    if tool_name != "Task":
        return {"continue": True}

    # Extract subagent_type
    subagent_type = parameters.get("subagent_type")

    if not subagent_type:
        # No subagent specified - allow operation
        return {"continue": True}

    # Validate agent permissionMode
    validation = validate_agent_permission_mode(subagent_type)

    # If validation passed, allow operation
    if validation["valid"]:
        return {"continue": True}

    # Validation failed - block operation
    response = {
        "continue": False,
        "systemMessage": validation["error_message"]
    }

    return response


def main() -> None:
    """Main entry point for PreToolUse hook

    Validates agent permissionMode before Task delegation:
    1. Extract subagent_type from Task parameters
    2. Locate agent file in .claude/agents/moai/
    3. Extract permissionMode from YAML frontmatter
    4. Validate against allowed values
    5. Block operation if invalid mode detected

    Exit Codes:
        0: Success (validation complete)
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Set 2-second timeout
    timeout = CrossPlatformTimeout(2)
    timeout.start()

    try:
        # Read JSON payload from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Call handler
        result = handle_pre_tool_use(data)

        # Output result as JSON
        print(json.dumps(result, ensure_ascii=False))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - allow operation to continue
        timeout_response: Dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Agent permission validation timeout - operation proceeding",
        }
        print(json.dumps(timeout_response, ensure_ascii=False))
        print("PreToolUse agent permission hook timeout after 2 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error - allow operation to continue
        json_error_response: Dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(json_error_response, ensure_ascii=False))
        print(f"PreToolUse agent permission JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error - allow operation to continue
        unexpected_error_response: Dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"Agent permission validation error: {e}"},
        }
        print(json.dumps(unexpected_error_response, ensure_ascii=False))
        print(f"PreToolUse agent permission unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
