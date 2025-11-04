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

    Args:
        payload: Claude Code PostToolUse 이벤트 데이터

    Returns:
        Hook 결과 (blocked=False, message 포함)
    """
    try:
        # 1. TRUST 검증 필요 여부 판단
        if not is_trust_validation_needed(payload):
            return {
                "blocked": False,
                "message": None,
            }

        # 2. 검증 도구 존재 확인
        validate_script = (
            project_root / "src" / "moai_adk" / "cli" / "validate_trust.py"
        )
        if not validate_script.exists():
            return {
                "blocked": False,
                "message": (
                    "ℹ️ TRUST 검증 도구를 찾을 수 없습니다. "
                    "src/moai_adk/cli/validate_trust.py 필요"
                ),
            }

        # 3. 비동기 검증 실행
        try:
            process = trigger_trust_validation()

            # 프로세스 ID를 메모리 파일에 저장 (다음 Hook에서 수집)
            pid_file = project_root / ".moai" / "memory" / "validation_pids.json"
            pid_file.parent.mkdir(parents=True, exist_ok=True)

            pids = []
            if pid_file.exists():
                try:
                    pids = json.loads(pid_file.read_text())
                except (json.JSONDecodeError, OSError):
                    pids = []

            pids.append(process.pid)
            pid_file.write_text(json.dumps(pids))

            return {
                "blocked": False,
                "message": "🔍 TRUST 원칙 검증 중... (백그라운드 실행)",
            }

        except Exception as e:
            return {
                "blocked": False,
                "message": f"⚠️ TRUST 검증 시작 실패: {str(e)}",
            }

    except Exception as e:
        return {
            "blocked": False,
            "message": f"❌ Hook 실행 중 오류: {str(e)}",
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
