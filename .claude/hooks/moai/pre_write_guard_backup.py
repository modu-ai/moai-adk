#!/usr/bin/env python3
"""PreToolUse guard for MoAI-ADK.

Blocks edits to sensitive files and enforces steering workflow usage.
"""

from __future__ import annotations

import json
import os
import sys
from pathlib import Path
from typing import Dict, Optional

try:
    project_root = Path(__file__).resolve().parents[3]
    src_dir = project_root / "src"
    if str(src_dir) not in sys.path:
        sys.path.insert(0, str(src_dir))
    from moai_adk.core.security import SecurityManager, SecurityError  # type: ignore
except Exception:  # pragma: no cover - fallback when package missing
    SecurityManager = None

    class SecurityError(Exception):
        """Fallback security error."""
        pass

EDIT_TOOLS = {"Write", "Edit", "MultiEdit"}
SENSITIVE_KEYWORDS = (
    ".env",
    "/secrets",
    "/.git/",
    "/.ssh",
)
PROTECTED_PREFIXES = (
    ".moai/memory/",
)

SECURITY_MANAGER = SecurityManager() if SecurityManager else None


def _load_input() -> Dict[str, object]:
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"ERROR pre_write_guard: 잘못된 JSON 입력 ({exc})", file=sys.stderr)
        sys.exit(1)


def _project_root() -> Path:
    env = os.environ.get("CLAUDE_PROJECT_DIR")
    return Path(env).resolve() if env else Path.cwd().resolve()


def _normalize_path(raw_path: object, root: Path) -> Optional[Path]:
    if not isinstance(raw_path, str) or not raw_path.strip():
        return None
    candidate = Path(raw_path)
    if not candidate.is_absolute():
        candidate = (root / candidate).resolve()
    return candidate


def _validate_with_security(path: Path, root: Path) -> Optional[str]:
    if SECURITY_MANAGER:
        try:
            if not SECURITY_MANAGER.validate_path_safety(path, root):
                return "프로젝트 루트 밖 경로는 수정할 수 없습니다."
        except SecurityError as exc:
            return f"보안 매니저가 경로를 거부했습니다: {exc}"
    return None


def _check_path_rules(path: Path, root: Path) -> Optional[str]:
    try:
        path.relative_to(root)
    except ValueError:
        return "프로젝트 루트 밖 경로는 수정할 수 없습니다."

    path_posix = path.as_posix().lower()

    # .moai/project/ 경로는 모든 편집 허용
    project_docs_root = (root / ".moai" / "project").resolve()
    try:
        path.resolve().relative_to(project_docs_root)
        return None
    except ValueError:
        pass

    for keyword in SENSITIVE_KEYWORDS:
        if keyword in path_posix:
            return f"민감 경로({keyword})는 편집할 수 없습니다."
    for prefix in PROTECTED_PREFIXES:
        if prefix in path.as_posix():
            return "Steering 문서는 전용 명령으로만 편집하세요."
    return None


def main() -> None:
    data = _load_input()
    tool_name = str(data.get("tool_name", ""))
    if tool_name not in EDIT_TOOLS:
        sys.exit(0)

    root = _project_root()
    tool_input = data.get("tool_input", {}) or {}
    raw_path = (
        tool_input.get("file_path")
        or tool_input.get("filePath")
        or tool_input.get("path")
    )

    if isinstance(raw_path, str) and '..' in Path(raw_path).parts:
        print('BLOCKED: 상대 경로 ".." 는 허용되지 않습니다.', file=sys.stderr)
        sys.exit(2)

    target = _normalize_path(raw_path, root)
    if not target:
        sys.exit(0)

    message = _validate_with_security(target, root) or _check_path_rules(target, root)
    if message:
        print(f"BLOCKED: {message}", file=sys.stderr)
        print("HINT: /moai:0-project 또는 /moai:1-spec 커맨드로 문서를 갱신해주세요.", file=sys.stderr)
        sys.exit(2)

    sys.exit(0)


if __name__ == "__main__":
    main()
