#!/usr/bin/env python3
"""
PreToolUse guard for file safety and anti-hallucination controls.

Blocks risky writes/edits and enforces simple creation/count/size limits:
- Sensitive path protection (.env, .git/, keys, secrets)
- New file count per session <= 5 (approximate)
- Content size per write <= 200 KiB
- Bash safety: block dangerous rm -rf, encourage rg over grep

This script expects Claude Code hook JSON on stdin.
"""
import json
import os
import re
import sys
from pathlib import Path
from typing import Any, Dict, List


MAX_NEW_FILES_PER_SESSION = 5
MAX_CONTENT_BYTES = 200 * 1024
STATE_DIRNAME = ".claude/.hook_state"


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
    except json.JSONDecodeError as e:
        print(f"Hook error: invalid JSON: {e}", file=sys.stderr)
        sys.exit(1)


def ensure_state_file(project_dir: Path, session_id: str) -> Path:
    state_dir = project_dir / STATE_DIRNAME
    state_dir.mkdir(parents=True, exist_ok=True)
    return state_dir / f"session_{session_id}.json"


def read_state(path: Path) -> Dict[str, Any]:
    if not path.exists():
        return {"new_files": 0}
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception:
        return {"new_files": 0}


def write_state(path: Path, data: Dict[str, Any]) -> None:
    try:
        path.write_text(json.dumps(data, ensure_ascii=False), encoding="utf-8")
    except Exception as e:
        print(f"Hook warn: failed to persist state: {e}", file=sys.stderr)


def is_sensitive_path(p: str) -> bool:
    lp = p.lower()
    return any(sp in lp for sp in SENSITIVE_PATTERNS)


def block(msg: str) -> None:
    print(msg, file=sys.stderr)
    sys.exit(2)


def handle_file_tools(data: Dict[str, Any]) -> None:
    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {}) or {}
    session_id = data.get("session_id", "session")
    project_dir_str = data.get("cwd") or os.environ.get("CLAUDE_PROJECT_DIR") or os.getcwd()
    project_dir = Path(project_dir_str)

    state_path = ensure_state_file(project_dir, session_id)
    state = read_state(state_path)

    def check_one(file_path: str, content: str) -> None:
        if not file_path:
            return
        if is_sensitive_path(file_path):
            block(f"Blocked sensitive path: {file_path}")
        if content is not None and isinstance(content, str):
            if len(content.encode("utf-8")) > MAX_CONTENT_BYTES:
                block(f"Content too large (> {MAX_CONTENT_BYTES} bytes): {file_path}")
        # new file detection
        abs_path = (project_dir / file_path).resolve()
        if not abs_path.exists() and tool_name in {"Write", "Edit"}:
            state["new_files"] = int(state.get("new_files", 0)) + 1
            if state["new_files"] > MAX_NEW_FILES_PER_SESSION:
                block(
                    f"New file creation limit exceeded (>{MAX_NEW_FILES_PER_SESSION}). "
                    f"Use /moai:6-sync force or batch your changes."
                )

    # MultiEdit might include multiple files
    if tool_name == "MultiEdit" and isinstance(tool_input, dict) and "files" in tool_input:
        files: List[Dict[str, Any]] = tool_input.get("files") or []
        for entry in files:
            check_one(str(entry.get("file_path", "")), entry.get("content", ""))
    else:
        check_one(str(tool_input.get("file_path", "")), tool_input.get("content", ""))

    # persist state
    write_state(state_path, state)
    # allow
    sys.exit(0)


def handle_bash(data: Dict[str, Any]) -> None:
    tool_input = data.get("tool_input", {}) or {}
    cmd = tool_input.get("command", [])
    s = " ".join(cmd) if isinstance(cmd, list) else str(cmd)

    if re.search(r"\brm\s+-rf\s+(/|\.|\*|\$)", s):
        block("Refuse dangerous rm -rf. Specify explicit safe target.")
    if re.search(r"(^|\s)grep(\s|$)", s) and not re.search(r"(^|\s)rg(\s|$)", s):
        block("Use rg (ripgrep) instead of grep for speed/consistency.")

    # allow
    sys.exit(0)


def main() -> None:
    data = load_stdin()
    event = data.get("hook_event_name", "")
    if event != "PreToolUse":
        sys.exit(0)
    tool_name = data.get("tool_name", "")
    if tool_name in {"Write", "Edit", "MultiEdit"}:
        handle_file_tools(data)
    if tool_name == "Bash":
        handle_bash(data)
    sys.exit(0)


if __name__ == "__main__":
    main()

