#!/usr/bin/env python3
"""
PreToolUse guard for file safety and anti-hallucination controls.

Blocks risky writes/edits by:
- Protecting sensitive paths (.env, .git/, keys, secrets, id_rsa 등)
- Guarding dangerous Bash commands (rm -rf, grep → rg 전환)

New file/내용 크기 제한은 제거되었습니다. (대량 생성 허용)
이 스크립트는 Claude Code Hook JSON을 stdin으로 받습니다.
"""
import json
import os
import re
import sys
from pathlib import Path
from typing import Any, Dict, List

SENSITIVE_PATTERNS = [
    ".env",
    ".git/",
    "/keys/",
    "/secrets/",
    "id_rsa",
]


def load_stdin() -> Dict[str, Any]:
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"Hook error: invalid JSON: {exc}", file=sys.stderr)
        sys.exit(1)


def is_sensitive_path(path: str) -> bool:
    lowered = path.lower()
    return any(pattern in lowered for pattern in SENSITIVE_PATTERNS)


def block(message: str) -> None:
    print(message, file=sys.stderr)
    sys.exit(2)


def handle_file_tools(payload: Dict[str, Any]) -> None:
    tool_name = payload.get("tool_name", "")
    tool_input = payload.get("tool_input", {}) or {}

    def check_one(file_path: str) -> None:
        if not file_path:
            return
        if is_sensitive_path(file_path):
            block(f"Blocked sensitive path: {file_path}")

    if tool_name == "MultiEdit" and isinstance(tool_input, dict) and "files" in tool_input:
        files: List[Dict[str, Any]] = tool_input.get("files") or []
        for entry in files:
            check_one(str(entry.get("file_path", "")))
    else:
        check_one(str(tool_input.get("file_path", "")))

    sys.exit(0)


def handle_bash(payload: Dict[str, Any]) -> None:
    tool_input = payload.get("tool_input", {}) or {}
    command = tool_input.get("command", [])
    command_str = " ".join(command) if isinstance(command, list) else str(command)

    if re.search(r"\brm\s+-rf\s+(/|\.|\*|\$)", command_str):
        block("Refuse dangerous rm -rf. Specify explicit safe target.")
    if re.search(r"(^|\s)grep(\s|$)", command_str) and not re.search(r"(^|\s)rg(\s|$)", command_str):
        block("Use rg (ripgrep) instead of grep for speed/consistency.")

    sys.exit(0)


def main() -> None:
    data = load_stdin()
    if data.get("hook_event_name") != "PreToolUse":
        sys.exit(0)

    tool_name = data.get("tool_name", "")
    if tool_name in {"Write", "Edit", "MultiEdit"}:
        handle_file_tools(data)
    if tool_name == "Bash":
        handle_bash(data)

    sys.exit(0)


if __name__ == "__main__":
    main()
