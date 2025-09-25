#!/usr/bin/env python3
# @SEC:PRE-WRITE-GUARD-011
"""
Optimized PreToolUse guard for MoAI-ADK - v0.2.0
Minimal security checking with essential protection.
"""

import json
import sys

# Essential security patterns
SENSITIVE_KEYWORDS = (".env", "/secrets", "/.git/", "/.ssh")
PROTECTED_PATHS = (".moai/memory/",)

def check_file_safety(file_path: str) -> bool:
    """Check if file is safe to edit"""
    if not file_path:
        return True

    path_lower = file_path.lower()

    # Block sensitive patterns
    for keyword in SENSITIVE_KEYWORDS:
        if keyword in path_lower:
            return False

    # Block protected paths
    for protected in PROTECTED_PATHS:
        if protected in file_path:
            return False

    return True

def main() -> None:
    """Main entry point for Claude Code hook system"""
    try:
        data = json.load(sys.stdin)
        tool_name = data.get("tool_name", "")

        if tool_name not in {"Write", "Edit", "MultiEdit"}:
            sys.exit(0)

        tool_input = data.get("tool_input", {}) or {}
        file_path = (tool_input.get("file_path") or
                    tool_input.get("filePath") or
                    tool_input.get("path"))

        if not check_file_safety(str(file_path) if file_path else ""):
            print("BLOCKED: 민감한 파일은 편집할 수 없습니다.", file=sys.stderr)
            sys.exit(2)

        sys.exit(0)

    except Exception:
        # Silent failure to avoid breaking Claude Code session
        sys.exit(0)

if __name__ == "__main__":
    main()
