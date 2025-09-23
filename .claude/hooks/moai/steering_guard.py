#!/usr/bin/env python3
"""UserPromptSubmit guard enforcing steering safety rules with session notifications."""

from __future__ import annotations

import json
import re
import sys
import os
from pathlib import Path
from typing import Dict

BANNED_PATTERNS = (
    (re.compile(r'(?i)ignore (the )?(claude|constitution|steering|instructions)'), '헌법/지침 무시는 허용되지 않습니다.'),
    (re.compile(r'(?i)disable (all )?(hooks?|guards?|polic(y|ies))'), 'Hook/Guard 해제 요청은 차단되었습니다.'),
    (re.compile(r'(?i)rm -rf'), '위험한 셸 명령을 프롬프트로 제출할 수 없습니다.'),
    (re.compile(r'(?i)drop (all )?safeguards'), '안전장치 제거 요청은 거부됩니다.'),
    (re.compile(r'(?i)clear (all )?(memory|steering)'), 'Steering 메모리를 강제 삭제하는 요청은 지원하지 않습니다.'),
)
# 세션별 알림 상태 추적을 위한 임시 파일
SESSION_NOTIFIED_FILE = "/tmp/moai_session_notified"

def _check_moai_project() -> bool:
    """MoAI 프로젝트 여부 확인"""
    current_dir = Path.cwd()
    return (current_dir / ".moai").exists() and (current_dir / "CLAUDE.md").exists()

def _show_session_notice() -> None:
    """세션 시작 알림 표시 (최초 1회)"""
    if os.path.exists(SESSION_NOTIFIED_FILE):
        return  # 이미 알림을 표시했음

    if not _check_moai_project():
        return  # MoAI 프로젝트가 아님

    # 알림 표시
    print("🚀 MoAI-ADK 프로젝트가 감지되었습니다!", file=sys.stderr)
    print("📖 개발 가이드: CLAUDE.md | TRUST 원칙: .moai/memory/development-guide.md", file=sys.stderr)
    print("⚡ 워크플로우: /moai:1-spec → /moai:2-build → /moai:3-sync", file=sys.stderr)
    print("🔧 디버깅: /moai:4-debug | 설정 관리: @agent-cc-manager", file=sys.stderr)
    print("", file=sys.stderr)

    # 알림 완료 표시
    try:
        with open(SESSION_NOTIFIED_FILE, "w") as f:
            f.write("notified")
    except:
        pass  # 임시 파일 생성 실패는 무시

def _load_input() -> Dict[str, object]:
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        print(f"ERROR steering_guard: 잘못된 JSON 입력 ({exc})", file=sys.stderr)
        sys.exit(1)


def main() -> None:
    # 세션 시작 알림 확인 (첫 번째 프롬프트에서만)
    _show_session_notice()

    data = _load_input()
    prompt = data.get('prompt')
    if not isinstance(prompt, str):
        sys.exit(0)

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
