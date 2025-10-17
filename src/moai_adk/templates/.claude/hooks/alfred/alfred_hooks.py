#!/usr/bin/env python3
# @CODE:HOOKS-REFACTOR-001 | SPEC: SPEC-HOOKS-REFACTOR-001.md
"""Alfred Hooks - Main entry point for MoAI-ADK Claude Code Hooks

Claude Code 이벤트를 적절한 핸들러로 라우팅하는 메인 진입점

Setup sys.path for package imports
"""
import sys
from pathlib import Path

# Add the hooks directory to sys.path to enable package imports
HOOKS_DIR = Path(__file__).parent
if str(HOOKS_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_DIR))

# Now we can import from the package

"""
Architecture:
┌─────────────────────────────────────────────────────────────┐
│ alfred_hooks.py (Router)                                    │
├─────────────────────────────────────────────────────────────┤
│ - CLI argument parsing                                      │
│ - JSON I/O (stdin/stdout)                                   │
│ - Event routing to handlers                                 │
└─────────────────────────────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ handlers/ (Event Handlers)                                  │
├─────────────────────────────────────────────────────────────┤
│ - session.py: SessionStart, SessionEnd                      │
│ - user.py: UserPromptSubmit                                 │
│ - tool.py: PreToolUse, PostToolUse                          │
│ - notification.py: Notification, Stop, SubagentStop         │
└─────────────────────────────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ core/ (Business Logic)                                      │
├─────────────────────────────────────────────────────────────┤
│ - project.py: Language detection, Git info, SPEC progress   │
│ - context.py: JIT Retrieval, workflow context               │
│ - checkpoint.py: Event-Driven Checkpoint system             │
│ - tags.py: TAG search/verification, library version cache   │
└─────────────────────────────────────────────────────────────┘

Usage:
    python alfred_hooks.py <event_name> < payload.json

Supported Events:
    - SessionStart: 세션 시작 (프로젝트 상태 표시)
    - UserPromptSubmit: 프롬프트 제출 (JIT 문서 로딩)
    - PreToolUse: Tool 사용 전 (Checkpoint 자동 생성)
    - SessionEnd, PostToolUse, Notification, Stop, SubagentStop

Exit Codes:
    - 0: 성공
    - 1: 에러 (인수 없음, JSON 파싱 실패, 예외 발생)

TDD History:
    - RED: 모듈 분리 설계, 이벤트 라우팅 테스트
    - GREEN: 1233 LOC → 9개 모듈 분리 구현 (SRP 준수)
    - REFACTOR: Import 최적화, 에러 처리 강화
"""

import json

from core import HookResult
from handlers import (
    handle_notification,
    handle_post_tool_use,
    handle_pre_tool_use,
    handle_session_end,
    handle_session_start,
    handle_stop,
    handle_subagent_stop,
    handle_user_prompt_submit,
)


def main() -> None:
    """메인 진입점 - Claude Code Hook 스크립트

    CLI 인수로 이벤트명을 받고, stdin으로 JSON 페이로드를 읽습니다.
    이벤트에 맞는 핸들러를 호출하고, 결과를 JSON으로 stdout에 출력합니다.

    Usage:
        python alfred_hooks.py <event_name> < payload.json

    Supported Events:
        - SessionStart: 세션 시작 (프로젝트 상태 표시)
        - UserPromptSubmit: 프롬프트 제출 (JIT 문서 로딩)
        - SessionEnd, PreToolUse, PostToolUse, Notification, Stop, SubagentStop

    Exit Codes:
        - 0: 성공
        - 1: 에러 (인수 없음, JSON 파싱 실패, 예외 발생)

    Examples:
        $ echo '{"cwd": "."}' | python alfred_hooks.py SessionStart
        {"message": "🚀 MoAI-ADK Session Started\\n...", ...}

    Notes:
        - Claude Code가 자동으로 호출 (사용자 직접 실행 불필요)
        - stdin/stdout으로 JSON I/O 처리
        - stderr로 에러 메시지 출력
        - UserPromptSubmit은 특별한 출력 스키마 사용 (hookEventName + additionalContext)

    TDD History:
        - RED: 이벤트 라우팅, JSON I/O, 에러 처리 테스트
        - GREEN: 핸들러 맵 기반 라우팅 구현
        - REFACTOR: 에러 메시지 명확화, exit code 표준화, UserPromptSubmit 스키마 분리
    """
    # Check for event argument
    if len(sys.argv) < 2:
        print("Usage: alfred_hooks.py <event>", file=sys.stderr)
        sys.exit(1)

    event_name = sys.argv[1]

    try:
        # Read JSON from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data)

        cwd = data.get("cwd", ".")

        # Route to appropriate handler
        handlers = {
            "SessionStart": handle_session_start,
            "UserPromptSubmit": handle_user_prompt_submit,
            "SessionEnd": handle_session_end,
            "PreToolUse": handle_pre_tool_use,
            "PostToolUse": handle_post_tool_use,
            "Notification": handle_notification,
            "Stop": handle_stop,
            "SubagentStop": handle_subagent_stop,
        }

        handler = handlers.get(event_name)
        result = handler({"cwd": cwd, **data}) if handler else HookResult()

        # UserPromptSubmit은 특별한 출력 스키마 사용
        if event_name == "UserPromptSubmit":
            print(json.dumps(result.to_user_prompt_submit_dict()))
        else:
            print(json.dumps(result.to_dict()))

        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
