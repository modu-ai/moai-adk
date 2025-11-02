#!/usr/bin/env python3
"""PreToolUse Hook: Risk Detection (Ultra-Lightweight)

Claude Code Event: PreToolUse
Purpose: Detect risky operations before execution
Execution: Triggered before tool use

Features:
- Large file detection (>500 LOC)
- Minimal overhead (<50ms)
- Graceful degradation on failure
"""

import json
import sys
from pathlib import Path
from typing import Any

def get_file_line_count(file_path: str) -> int:
    """Get approximate line count of file"""
    try:
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            return len(f.readlines())
    except Exception:
        return 0

def handle_pre_tool(payload: dict) -> dict[str, Any]:
    """Handle PreToolUse event with risk detection

    Returns: Hook result (continue=True/False)
    """
    tool_name = payload.get("tool", {}).get("name", "")
    tool_args = payload.get("tool", {}).get("arguments", {})

    # Risk Detection Pattern 1: Large file edit
    if tool_name == "Edit" and "file_path" in tool_args:
        file_path = tool_args["file_path"]
        try:
            line_count = get_file_line_count(file_path)
            if line_count > 500:
                return {
                    "continue": True,
                    "systemMessage": f"⚠️ Large file ({line_count} lines). Create checkpoint first?"
                }
        except Exception:
            pass

    # Risk Detection Pattern 2: Force push
    if tool_name == "Bash" and "command" in tool_args:
        cmd = tool_args["command"]
        if "push --force" in cmd or "force-push" in cmd:
            return {
                "continue": True,
                "systemMessage": "🔴 HIGH RISK: Force push. Use --force-with-lease instead?"
            }

    # Default: safe to continue
    return {"continue": True}

def main() -> None:
    """Main entry point"""
    try:
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        result = handle_pre_tool(data)
        print(json.dumps(result))
        sys.exit(0)

    except Exception as e:
        # Graceful degradation: always allow continuation
        error_response = {
            "continue": True,
            "hookSpecificOutput": {"error": str(e)}
        }
        print(json.dumps(error_response))
        print(f"PreToolUse error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
