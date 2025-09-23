#!/usr/bin/env python3
"""Policy guard for Bash commands in MoAI-ADK."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path
from typing import Dict, Optional

DANGEROUS_COMMANDS = (
    'rm -rf /',
    'rm -rf --no-preserve-root',
    'sudo rm',
    'dd if=/dev/zero',
    ':(){:|:&};:',
    'mkfs.',
)
DEPRECATED_PATTERNS = (
    (re.compile(r'(^|\s)grep(\s|$)'), "'rg' 명령을 사용해주세요."),
    (re.compile(r'(^|\s)find\s+[^|]*-name'), "'rg --files -g' 패턴을 사용해주세요."),
)
PROTECTED_SEGMENTS = (
    '.moai/project/',
    '.moai/memory/',
)
ALLOWED_PREFIXES = (
    'git ',
    'python',
    'pytest',
    'npm ',
    'node ',
    'go ',
    'cargo ',
    'poetry ',
    'pnpm ',
)


def _load_input() -> Dict[str, object]:
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"ERROR policy_block: 잘못된 JSON 입력 ({exc})", file=sys.stderr)
        sys.exit(1)


def _extract_command(tool_input: object) -> Optional[str]:
    if not isinstance(tool_input, dict):
        return None
    raw = tool_input.get('command') or tool_input.get('cmd')
    if isinstance(raw, list):
        return ' '.join(str(item) for item in raw)
    if isinstance(raw, str):
        return raw.strip()
    return None


def _has_protected_segment(command: str) -> Optional[str]:
    lowered = command.lower()
    for segment in PROTECTED_SEGMENTS:
        if segment in lowered:
            return segment
    return None


def _is_allowed_prefix(command: str) -> bool:
    return any(command.startswith(prefix) for prefix in ALLOWED_PREFIXES)


def main() -> None:
    data = _load_input()
    if data.get('tool_name') != 'Bash':
        sys.exit(0)

    command = _extract_command(data.get('tool_input', {}))
    if not command:
        sys.exit(0)

    command_lower = command.lower()
    for token in DANGEROUS_COMMANDS:
        if token in command_lower:
            print(f"BLOCKED: 위험 명령이 감지되었습니다 ({token}).", file=sys.stderr)
            sys.exit(2)

    protected = _has_protected_segment(command)
    if protected:
        print('BLOCKED: Steering/메모리 문서는 전용 커맨드로 수정하세요.', file=sys.stderr)
        print('HINT: /moai:0-project, /moai:1-spec, /moai:3-sync 커맨드를 사용해주세요.', file=sys.stderr)
        sys.exit(2)

    for pattern, message in DEPRECATED_PATTERNS:
        if pattern.search(command):
            print(f"BLOCKED: {message}", file=sys.stderr)
            sys.exit(2)

    if not _is_allowed_prefix(command):
        print('NOTICE: 등록되지 않은 명령입니다. 필요 시 settings.json 의 allow 목록을 갱신하세요.', file=sys.stderr)

    sys.exit(0)


if __name__ == '__main__':
    main()
