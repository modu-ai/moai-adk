"""
@CODE:HOOKS-003-CORE - TRUST 원칙 자동 검증 시스템

PostToolUse Hook에서 TDD 완료를 감지하고
TRUST 5 원칙 자동 검증을 수행합니다.
"""

import json
import subprocess
from pathlib import Path
from typing import Any, Dict, Optional
import sys


def detect_tdd_completion() -> bool:
    """
    Git 로그 분석하여 TDD 완료 여부 확인.

    최근 5개 커밋에서 🟢 GREEN 또는 ♻️ REFACTOR 키워드 검색.

    Returns:
        True: GREEN 또는 REFACTOR 단계 감지
        False: TDD 구현 미완료
    """
    try:
        result = subprocess.run(
            ["git", "log", "-5", "--pretty=format:%s"],
            capture_output=True,
            text=True,
            timeout=0.5,  # 성능: 500ms로 단축
        )

        if result.returncode != 0:
            return False

        commit_messages = result.stdout.strip().split("\n")
        tdd_keywords = ["🟢 GREEN:", "♻️ REFACTOR:"]

        # 성능: 첫 번째 매치 발견 시 즉시 반환
        for msg in commit_messages:
            if msg and any(keyword in msg for keyword in tdd_keywords):
                return True

        return False

    except subprocess.TimeoutExpired:
        # 성능: Git 타임아웃 시 False 반환 (non-blocking)
        return False
    except Exception:
        # 에러 처리: 모든 예외는 조용히 처리
        return False


def trigger_trust_validation() -> subprocess.Popen:
    """
    TRUST 검증을 백그라운드 프로세스로 실행.

    Returns:
        subprocess.Popen 객체 (비동기 실행)
    """
    project_root = Path(__file__).parent.parent.parent.parent

    process = subprocess.Popen(
        [sys.executable, "-m", "moai_adk.cli.validate_trust", "--json"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        cwd=str(project_root),
    )

    return process


def collect_validation_result(
    process: subprocess.Popen,
    timeout: float = 30.0,
) -> Dict[str, Any]:
    """
    TRUST 검증 결과를 수집하고 파싱.

    Args:
        process: 실행 중인 검증 프로세스
        timeout: 프로세스 타임아웃 (초)

    Returns:
        JSON 형식 검증 보고서
    """
    try:
        stdout, stderr = process.communicate(timeout=timeout)

        if process.returncode != 0:
            return {
                "status": "failed",
                "error": stderr if stderr else "Unknown error",
                "exit_code": process.returncode,
            }

        return json.loads(stdout)

    except subprocess.TimeoutExpired:
        process.kill()
        return {
            "status": "failed",
            "error": f"Validation timeout after {timeout} seconds",
            "exit_code": -1,
        }
    except json.JSONDecodeError:
        return {
            "status": "failed",
            "error": "Failed to parse validation output",
            "exit_code": -1,
        }


def format_validation_result(result: Dict[str, Any]) -> str:
    """
    TRUST 검증 결과를 Markdown 형식으로 변환.

    Args:
        result: JSON 형식 검증 보고서

    Returns:
        Markdown 형식 알림 메시지
    """
    if result.get("status") == "passed":
        return (
            f"✅ **TRUST 원칙 검증 통과**\n"
            f"- 테스트 커버리지: {result.get('test_coverage', 'N/A')}%\n"
            f"- 코드 제약 준수: {result.get('code_constraints_passed', 0)}/{result.get('code_constraints_total', 0)}\n"
            f"- TAG 체인 무결성: OK"
        )

    else:
        return (
            f"❌ **TRUST 원칙 검증 실패**\n"
            f"- 실패 원인: {result.get('error', 'Unknown error')}\n"
            f"- 테스트 커버리지: {result.get('test_coverage', 'N/A')}% (목표 85%)\n"
            f"- 권장 조치: {result.get('recommendation', 'scripts/validate_trust.py 실행하여 상세 확인')}"
        )


def is_alfred_build_command(payload: Dict[str, Any]) -> bool:
    """
    PostToolUse payload에서 alfred:2-run 실행 여부 확인.

    Args:
        payload: PostToolUse 이벤트 데이터

    Returns:
        True: alfred:2-run 실행됨
        False: 다른 명령 실행됨
    """
    tool_name = payload.get("tool", "")
    tool_input = payload.get("input", {})

    command = tool_input.get("command", "")
    description = tool_input.get("description", "")

    return "alfred:2-run" in command or "alfred:2-run" in description


def is_trust_validation_needed(payload: Dict[str, Any]) -> bool:
    """
    TRUST 검증이 필요한지 판단.

    Args:
        payload: PostToolUse 이벤트 데이터

    Returns:
        True: 검증 필요 (TDD 완료 또는 alfred:2-run)
        False: 검증 불필요
    """
    return detect_tdd_completion() or is_alfred_build_command(payload)
