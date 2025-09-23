#!/usr/bin/env python3
"""UserPromptSubmit guard enforcing steering safety rules."""

from __future__ import annotations

import json
import re
import sys
from typing import Dict

BANNED_PATTERNS = (
    (re.compile(r'(?i)ignore (the )?(claude|constitution|steering|instructions)'), '헌법/지침 무시는 허용되지 않습니다.'),
    (re.compile(r'(?i)disable (all )?(hooks?|guards?|polic(y|ies))'), 'Hook/Guard 해제 요청은 차단되었습니다.'),
    (re.compile(r'(?i)rm -rf'), '위험한 셸 명령을 프롬프트로 제출할 수 없습니다.'),
    (re.compile(r'(?i)drop (all )?safeguards'), '안전장치 제거 요청은 거부됩니다.'),
    (re.compile(r'(?i)clear (all )?(memory|steering)'), 'Steering 메모리를 강제 삭제하는 요청은 지원하지 않습니다.'),
)
MAX_PROMPT_LEN = 6000


def _load_input() -> Dict[str, object]:
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"ERROR steering_guard: 잘못된 JSON 입력 ({exc})", file=sys.stderr)
        sys.exit(1)


def main() -> None:
    data = _load_input()
    prompt = data.get('prompt')
    if not isinstance(prompt, str):
        sys.exit(0)

    if len(prompt) > MAX_PROMPT_LEN:
        print('BLOCKED: 프롬프트가 너무 깁니다. 핵심 요구만 정리해 다시 제출해주세요.', file=sys.stderr)
        sys.exit(2)

    for pattern, message in BANNED_PATTERNS:
        if pattern.search(prompt):
            print(f'BLOCKED: {message}', file=sys.stderr)
            print('HINT: CLAUDE.md와 @.moai/project/* 문서를 기반으로 목표/제약을 명시해 주세요.', file=sys.stderr)
            sys.exit(2)

    # Provide lightweight steering context back to Claude.
    print('Steering Guard: 개발 가이드과 TAG 규칙을 준수하며 작업을 진행합니다.', flush=True)
    sys.exit(0)


if __name__ == '__main__':
    main()
