#!/usr/bin/env python3
"""
@CODE:HOOKS-003-HANDLER - PostToolUse Hook: TRUST 원칙 자동 검증

TDD 완료를 감지하여 TRUST 5 원칙 자동 검증을 비동기로 실행합니다.
"""

import json
import sys
from pathlib import Path
from typing import Any, Dict

# Hook 시스템 경로 추가
hook_root = Path(__file__).parent
sys.path.insert(0, str(hook_root))

# 패키지 경로 추가
project_root = hook_root.parent.parent.parent
sys.path.insert(0, str(project_root / "src"))

from moai_adk.core.validation import (
    is_trust_validation_needed,
    trigger_trust_validation,
    format_validation_result,
)


def handle_post_tool_use(payload: Dict[str, Any]) -> Dict[str, Any]:
    """
    PostToolUse Hook 핸들러: TRUST 검증 자동 트리거

    최적화: 100ms 제약 준수를 위해 최소 작업만 수행합니다.

    Args:
        payload: Claude Code PostToolUse 이벤트 데이터

    Returns:
        Hook 결과 (blocked=False, message 포함)
    """
    try:
        # 1. TRUST 검증 필요 여부 판단 (<100ms)
        if not is_trust_validation_needed(payload):
            return {
                "blocked": False,
                "message": None,
            }

        # 2. 비동기 검증 실행 (검증 도구 존재 확인은 프로세스 실행 시)
        try:
            process = trigger_trust_validation()

            # 성능: 프로세스 ID만 저장하고 즉시 반환
            try:
                pid_file = project_root / ".moai" / "memory" / "validation_pids.json"
                pid_file.parent.mkdir(parents=True, exist_ok=True)

                pids = []
                if pid_file.exists():
                    try:
                        pids = json.loads(pid_file.read_text())
                    except (json.JSONDecodeError, OSError):
                        pass

                pids.append(process.pid)
                pid_file.write_text(json.dumps(pids))
            except Exception:
                # PID 저장 실패는 무시 (검증 자체는 백그라운드 실행)
                pass

            return {
                "blocked": False,
                "message": "🔍 TRUST 원칙 검증 중... (백그라운드 실행)",
            }

        except Exception:
            # 검증 실행 실패는 silent (non-blocking)
            return {
                "blocked": False,
                "message": None,
            }

    except Exception:
        # 모든 예외는 조용히 처리 (Hook은 절대 blocked되면 안 됨)
        return {
            "blocked": False,
            "message": None,
        }


# Hook 진입점
if __name__ == "__main__":
    # 표준입력에서 페이로드 읽기
    try:
        payload = json.loads(sys.stdin.read())
    except json.JSONDecodeError:
        payload = {}

    result = handle_post_tool_use(payload)
    print(json.dumps(result))
