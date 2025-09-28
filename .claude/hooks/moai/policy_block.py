#!/usr/bin/env python3
# @SEC:POLICY-BLOCK-011
"""Policy guard for Bash commands in MoAI-ADK."""

from __future__ import annotations

import json
import sys

DANGEROUS_COMMANDS = (
    "rm -rf /",
    "rm -rf --no-preserve-root",
    "sudo rm",
    "dd if=/dev/zero",
    ":(){:|:&};:",
    "mkfs.",
)
DEPRECATED_PATTERNS = (
    # grep/find 차단 제거 - 가이드라인으로만 권장
)
PROTECTED_SEGMENTS = (
    # .moai/project/ 보호 해제 - 모든 접근 허용
    # .moai/memory/ 읽기는 허용, 쓰기만 차단 (별도 로직 필요)
)
ALLOWED_PREFIXES = (
    "git ",
    "python",
    "pytest",
    "npm ",
    "node ",
    "go ",
    "cargo ",
    "poetry ",
    "pnpm ",
    "rg ",
    "ls ",
    "cat ",
    "echo ",
    "which ",
    "make ",
    "moai ",
)


def _load_input() -> dict[str, object]:
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"ERROR policy_block: 잘못된 JSON 입력 ({exc})", file=sys.stderr)
        sys.exit(1)


def _extract_command(tool_input: object) -> str | None:
    if not isinstance(tool_input, dict):
        return None
    raw = tool_input.get("command") or tool_input.get("cmd")
    if isinstance(raw, list):
        return " ".join(str(item) for item in raw)
    if isinstance(raw, str):
        return raw.strip()
    return None


def _is_allowed_prefix(command: str) -> bool:
    return any(command.startswith(prefix) for prefix in ALLOWED_PREFIXES)


def main() -> None:
    data = _load_input()
    if data.get("tool_name") != "Bash":
        sys.exit(0)

    command = _extract_command(data.get("tool_input", {}))
    if not command:
        sys.exit(0)

    command_lower = command.lower()
    for token in DANGEROUS_COMMANDS:
        if token in command_lower:
            print(f"BLOCKED: 위험 명령이 감지되었습니다 ({token}).", file=sys.stderr)
            sys.exit(2)

    # 보호 세그먼트 검사 제거 - .moai/project/ 모든 접근 허용
    # 패턴 검사 제거 - grep/find 사용 가능 (rg 권장만)

    if not _is_allowed_prefix(command):
        print(
            "NOTICE: 등록되지 않은 명령입니다. 필요 시 settings.json 의 allow 목록을 갱신하세요.",
            file=sys.stderr,
        )

    sys.exit(0)


if __name__ == "__main__":
    main()
