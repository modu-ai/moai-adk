#!/usr/bin/env python3
# @CODE:HOOK-REALTIME-001 | SPEC: TAG-REALTIME-HOOK-001 | TEST: tests/hooks/test_realtime_tag_monitor.py
"""실시간 TAG 모니터링 Hook

지속적인 TAG 상태 모니터링과 실시간 위반 탐지.
PreToolUse 단계에서 프로젝트 전체 TAG 상태를 빠르게 검사.

기능:
- 실시간 TAG 상태 모니터링
- 빠른 위반 탐지 (5초 내)
- 프로젝트 전체 TAG 무결성 검사
- 사용자에게 즉각적인 피드백

사용법:
    python3 pre_tool__realtime_tag_monitor.py <tool_name> <tool_args_json>
"""

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List

# 모듈 경로 추가
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.validator import CentralValidationResult, CentralValidator, ValidationConfig

from ..utils.hook_config import get_graceful_degradation, load_hook_timeout


def load_config() -> Dict[str, Any]:
    """설정 파일 로드

    Returns:
        설정 딕셔너리
    """
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
    except Exception:
        pass

    return {}


def create_validator() -> CentralValidator:
    """중앙 검증기 생성

    Returns:
        CentralValidator 인스턴스
    """
    config_data = load_config()
    tag_policy_config = config_data.get("tags", {}).get("policy", {})

    # ValidationConfig 생성
    validation_config = ValidationConfig(
        strict_mode=tag_policy_config.get("enforcement_mode", "strict") == "strict",
        check_duplicates=True,
        check_orphans=True,
        check_chain_integrity=tag_policy_config.get("realtime_validation", {}).get("enforce_chains", True)
    )

    return CentralValidator(config=validation_config)


def should_monitor(tool_name: str, tool_args: Dict[str, Any]) -> bool:
    """모니터링 대상 툴인지 확인

    Args:
        tool_name: 툴 이름
        tool_args: 툴 인자

    Returns:
        모니터링 대상이면 True
    """
    # 파일 조작 툴만 모니터링
    monitoring_tools = {"Edit", "Write", "MultiEdit"}
    return tool_name in monitoring_tools


def get_project_files_to_scan() -> List[str]:
    """스캔할 프로젝트 파일 목록 가져오기 (최적화)

    선택적 파일 디렉토리를 제외하여 성능 개선.
    필수 파일만 스캔: src/, tests/, .moai/specs/

    Returns:
        파일 경로 목록
    """
    files = []
    # 필수 파일 패턴만 스캔 (선택적 파일 제외)
    important_patterns = [
        "src/**/*.py",           # 구현 코드
        "tests/**/*.py",         # 테스트 코드
        ".moai/specs/**/*.md"    # SPEC 문서
    ]

    # 제외 패턴 (성능 최적화를 위해 제외)
    exclude_patterns = [
        ".claude/",
        ".moai/docs/",
        ".moai/reports/",
        ".moai/analysis/",
        "docs/",
        "templates/",
        "examples/",
        "__pycache__/",
        "node_modules/"
    ]

    # 빠른 스캔을 위해 파일 수 제한 (50개 → 30개로 단축)
    max_files = 30

    for pattern in important_patterns:
        if len(files) >= max_files:
            break

        try:
            for path in Path(".").glob(pattern):
                if len(files) >= max_files:
                    break
                if path.is_file():
                    # 제외 패턴 확인
                    path_str = str(path)
                    if not any(exclude in path_str for exclude in exclude_patterns):
                        files.append(path_str)
        except Exception:
            continue

    return files[:max_files]


def create_quick_scan_result(validation_result: CentralValidationResult,
                           scan_time_ms: float) -> Dict[str, Any]:
    """빠른 스캔 결과 생성

    Args:
        validation_result: 검증 결과
        scan_time_ms: 스캔 시간

    Returns:
        스캔 결과 딕셔너리
    """
    result = {
        "quick_scan_completed": True,
        "scan_time_ms": scan_time_ms,
        "files_scanned": validation_result.statistics.total_files_scanned,
        "tags_found": validation_result.statistics.total_tags_found,
        "total_issues": validation_result.statistics.total_issues,
        "is_valid": validation_result.is_valid
    }

    # 심각한 문제만 요약
    if validation_result.errors:
        result["critical_issues"] = len(validation_result.errors)
        result["error_summary"] = [
            {
                "type": error.type,
                "tag": error.tag,
                "message": error.message
            }
            for error in validation_result.errors[:5]  # 최대 5개만 표시
        ]

    if validation_result.warnings:
        result["warnings"] = len(validation_result.warnings)

    # 커버리지 정보
    result["coverage_percentage"] = validation_result.statistics.coverage_percentage

    # 상태 메시지
    if validation_result.is_valid:
        result["status_message"] = "✅ 프로젝트 TAG 상태 양호"
    elif validation_result.errors:
        result["status_message"] = f"🚨 {len(validation_result.errors)}개 치명적 문제 발견"
    else:
        result["status_message"] = f"⚠️ {len(validation_result.warnings)}개 경고 발견"

    return result


def create_health_check_result(issues_count: int,
                             coverage_percentage: float,
                             scan_time_ms: float) -> Dict[str, Any]:
    """프로젝트 건강 상태 결과 생성

    Args:
        issues_count: 문제 수
        coverage_percentage: 커버리지
        scan_time_ms: 스캔 시간

    Returns:
        건강 상태 결과
    """
    # 건강 상태 계산
    health_score = 100

    # 문제 점수 차감
    health_score -= min(issues_count * 5, 50)  # 최대 50점 차감

    # 커버리지 점수
    if coverage_percentage < 50:
        health_score -= 20
    elif coverage_percentage < 75:
        health_score -= 10

    health_score = max(0, health_score)

    # 건강 등급
    if health_score >= 90:
        health_grade = "A"
        health_message = "매우 좋음"
    elif health_score >= 80:
        health_grade = "B"
        health_message = "좋음"
    elif health_score >= 70:
        health_grade = "C"
        health_message = "보통"
    elif health_score >= 60:
        health_grade = "D"
        health_message = "주의 필요"
    else:
        health_grade = "F"
        health_message = "개선 필요"

    return {
        "health_score": health_score,
        "health_grade": health_grade,
        "health_message": health_message,
        "issues_count": issues_count,
        "coverage_percentage": coverage_percentage,
        "scan_time_ms": scan_time_ms
    }


def main() -> None:
    """메인 함수"""
    try:
        # 설정에서 타임아웃 값 로드 (밀리초 → 초)
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # 인자 파싱
        if len(sys.argv) < 3:
            print(json.dumps({
                "quick_scan_completed": False,
                "error": "Invalid arguments. Usage: python3 pre_tool__realtime_tag_monitor.py <tool_name> <tool_args_json>"
            }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
        except json.JSONDecodeError:
            print(json.dumps({
                "quick_scan_completed": False,
                "error": "Invalid tool_args JSON"
            }))
            sys.exit(0)

        # 모니터링 대상 확인
        if not should_monitor(tool_name, tool_args):
            print(json.dumps({
                "quick_scan_completed": True,
                "message": "모니터링 대상 아님"
            }))
            sys.exit(0)

        # 시작 시간 기록
        start_time = time.time()

        # 스캔할 파일 목록 가져오기
        files_to_scan = get_project_files_to_scan()
        if not files_to_scan:
            print(json.dumps({
                "quick_scan_completed": True,
                "message": "스캔할 파일 없음"
            }))
            sys.exit(0)

        # 검증기 생성
        validator = create_validator()

        # 빠른 검증 실행 (설정된 타임아웃 사용)
        try:
            # 타임아웃 체크
            if time.time() - start_time > timeout_seconds:
                raise TimeoutError("Real-time monitoring timeout")

            validation_result = validator.validate_files(files_to_scan)
        except Exception as e:
            # 검증 실패시 타임아웃 처리
            scan_time = (time.time() - start_time) * 1000
            error_response = {
                "quick_scan_completed": False,
                "error": f"검증 타임아웃: {str(e)}",
                "scan_time_ms": scan_time,
                "message": "실시간 검증 타임아웃 - 정상 작동으로 간주"
            }

            if graceful_degradation:
                error_response["graceful_degradation"] = True
                error_response["message"] = "Real-time monitoring timeout but continuing due to graceful degradation"

            print(json.dumps(error_response, ensure_ascii=False))
            sys.exit(0)

        scan_time_ms = (time.time() - start_time) * 1000

        # 결과 생성
        scan_result = create_quick_scan_result(validation_result, scan_time_ms)

        # 건강 상태 검사
        health_result = create_health_check_result(
            validation_result.statistics.total_issues,
            validation_result.statistics.coverage_percentage,
            scan_time_ms
        )

        # 최종 응답
        response = {
            **scan_result,
            "health_check": health_result,
            "monitoring_type": "realtime_quick_scan"
        }

        # 타임아웃 경고
        timeout_warning_ms = timeout_seconds * 1000 * 0.8  # 80% of timeout
        if scan_time_ms > timeout_warning_ms:
            response["performance_warning"] = f"스캔 시간이 설정된 타임아웃의 80%를 초과했습니다 ({scan_time_ms:.0f}ms / {timeout_warning_ms:.0f}ms)"

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # 예외 발생시 기본 응답
        error_response = {
            "quick_scan_completed": False,
            "error": f"Hook execution error: {str(e)}",
            "message": "실시간 모니터링 오류 - 정상 작동으로 간주"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True
            error_response["message"] = "Real-time monitoring failed but continuing due to graceful degradation"

        print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()
