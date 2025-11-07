#!/usr/bin/env python3
# @CODE:HOOK-PRE-TAG-001 | SPEC: TAG-PRE-HOOK-001 | TEST: tests/hooks/test_pre_tool_tag_validator.py
"""PreToolUse Hook: TAG 정책 위반 실시간 차단

Edit/Write/MultiEdit 실행 전 TAG 정책 위반을 탐지하고 차단.
SPEC-first 원칙을 강제하여 코드 품질 보증.

기능:
- 파일 생성 전 TAG 정책 검증
- SPEC 없이 CODE 생성 시 차단
- 실시간 위반 보고 및 수정 가이드
- 작업 차단 또는 경고 제공

사용법:
    python3 pre_tool__tag_policy_validator.py <tool_name> <tool_args_json>
"""

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List

# 모듈 경로 추가
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.policy_validator import (
    PolicyValidationConfig,
    PolicyViolation,
    PolicyViolationLevel,
    TagPolicyValidator,
)

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


def create_policy_validator() -> TagPolicyValidator:
    """TAG 정책 검증기 생성

    Returns:
        TagPolicyValidator 인스턴스
    """
    config_data = load_config()
    tag_policy_config = config_data.get("tags", {}).get("policy", {})

    # PolicyValidationConfig 생성
    policy_config = PolicyValidationConfig(
        strict_mode=tag_policy_config.get("enforcement_mode", "strict") == "strict",
        require_spec_before_code=tag_policy_config.get("require_spec_before_code", True),
        require_test_for_code=tag_policy_config.get("require_test_for_code", True),
        allow_duplicate_tags=not tag_policy_config.get("enforce_chains", True),
        validation_timeout=tag_policy_config.get("realtime_validation", {}).get("validation_timeout", 5),
        auto_fix_enabled=tag_policy_config.get("auto_correction", {}).get("enabled", False)
    )

    return TagPolicyValidator(config=policy_config)


def should_validate_tool(tool_name: str, tool_args: Dict[str, Any]) -> bool:
    """검증 대상 툴인지 확인

    Args:
        tool_name: 툴 이름
        tool_args: 툴 인자

    Returns:
        검증 대상이면 True
    """
    # 파일 조작 툴만 검증
    validation_tools = {"Edit", "Write", "MultiEdit"}
    if tool_name not in validation_tools:
        return False

    # 선택적 파일 패턴 (TAG 검증 대상 아님)
    optional_patterns = [
        "CLAUDE.md",
        "README.md",
        "CHANGELOG.md",
        "CONTRIBUTING.md",
        ".claude/",
        ".moai/docs/",
        ".moai/reports/",
        ".moai/analysis/",
        "docs/",
        "templates/",
        "examples/",
    ]

    # Edit/Write의 경우 단일 파일 확인
    if tool_name in {"Edit", "Write"}:
        file_path = tool_args.get("file_path", "")
        if any(pattern in file_path for pattern in optional_patterns):
            return False
        return True

    # MultiEdit의 경우 여러 파일 확인
    if tool_name == "MultiEdit":
        edits = tool_args.get("edits", [])
        for edit in edits:
            file_path = edit.get("file_path", "")
            if any(pattern in file_path for pattern in optional_patterns):
                return False

    return True


def extract_file_paths(tool_name: str, tool_args: Dict[str, Any]) -> List[str]:
    """툴 인자에서 파일 경로 추출

    Args:
        tool_name: 툴 이름
        tool_args: 툴 인자

    Returns:
        파일 경로 목록
    """
    file_paths = []

    if tool_name in {"Edit", "Write"}:
        file_path = tool_args.get("file_path", "")
        if file_path:
            file_paths.append(file_path)

    elif tool_name == "MultiEdit":
        # MultiEdit의 경우 여러 파일 경로 추출
        edits = tool_args.get("edits", [])
        for edit in edits:
            file_path = edit.get("file_path", "")
            if file_path:
                file_paths.append(file_path)

    return file_paths


def get_file_content(tool_name: str, tool_args: Dict[str, Any], file_path: str) -> str:
    """파일 내용 가져오기

    Args:
        tool_name: 툴 이름
        tool_args: 툴 인자
        file_path: 파일 경로

    Returns:
        파일 내용
    """
    # Write: 새 내용
    if tool_name == "Write":
        return tool_args.get("content", "")

    # Edit/MultiEdit: 기존 내용에 수정 적용
    try:
        path = Path(file_path)
        if path.exists():
            return path.read_text(encoding="utf-8", errors="ignore")
    except Exception:
        pass

    return ""


def create_block_response(violations: List[PolicyViolation]) -> Dict[str, Any]:
    """작업 차단 응답 생성

    Args:
        violations: 정책 위반 목록

    Returns:
        차단 응답 딕셔너리
    """
    critical_violations = [v for v in violations if v.level == PolicyViolationLevel.CRITICAL]
    blocking_violations = [v for v in violations if v.should_block_operation()]

    response = {
        "block_execution": True,
        "reason": "TAG 정책 위반",
        "violations": [v.to_dict() for v in blocking_violations],
        "message": "❌ TAG 정책 위반으로 작업이 차단되었습니다.",
        "guidance": []
    }

    if critical_violations:
        response["message"] = "🚨 치명적인 TAG 정책 위반입니다. 작업을 진행할 수 없습니다."
        response["critical_violations"] = [v.to_dict() for v in critical_violations]

    # 수정 가이드 추가
    for violation in blocking_violations:
        if violation.guidance:
            response["guidance"].append(f"• {violation.guidance}")

    return response


def create_warning_response(violations: List[PolicyViolation]) -> Dict[str, Any]:
    """경고 응답 생성

    Args:
        violations: 정책 위반 목록

    Returns:
        경고 응답 딕셔너리
    """
    response = {
        "block_execution": False,
        "reason": "TAG 정책 경고",
        "violations": [v.to_dict() for v in violations],
        "message": "⚠️ TAG 정책 경고가 있지만 작업을 진행할 수 있습니다.",
        "guidance": []
    }

    # 수정 가이드 추가
    for violation in violations:
        if violation.guidance:
            response["guidance"].append(f"• {violation.guidance}")

    return response


def create_success_response() -> Dict[str, Any]:
    """성공 응답 생성

    Returns:
        성공 응답 딕셔너리
    """
    return {
        "block_execution": False,
        "reason": "TAG 정책 준수",
        "violations": [],
        "message": "✅ TAG 정책 검증 통과",
        "guidance": []
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
                "block_execution": False,
                "error": "Invalid arguments. Usage: python3 pre_tool__tag_policy_validator.py <tool_name> <tool_args_json>"
            }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
        except json.JSONDecodeError:
            print(json.dumps({
                "block_execution": False,
                "error": "Invalid tool_args JSON"
            }))
            sys.exit(0)

        # 시작 시간 기록 (타임아웃 체크용)
        start_time = time.time()

        # 툴 검증 여부 확인
        if not should_validate_tool(tool_name, tool_args):
            print(json.dumps(create_success_response()))
            sys.exit(0)

        # 파일 경로 추출
        file_paths = extract_file_paths(tool_name, tool_args)
        if not file_paths:
            print(json.dumps(create_success_response()))
            sys.exit(0)

        # 정책 검증기 생성
        validator = create_policy_validator()

        # 모든 파일에 대해 검증
        all_violations = []
        for file_path in file_paths:
            # 타임아웃 체크
            if time.time() - start_time > timeout_seconds:
                break

            # 파일 내용 가져오기
            content = get_file_content(tool_name, tool_args, file_path)

            # 정책 검증
            violations = validator.validate_before_creation(file_path, content)
            all_violations.extend(violations)

        # 위반 수준별 분류
        blocking_violations = [v for v in all_violations if v.should_block_operation()]
        warning_violations = [v for v in all_violations if not v.should_block_operation()]

        # 응답 생성
        if blocking_violations:
            response = create_block_response(blocking_violations)
        elif warning_violations:
            response = create_warning_response(warning_violations)
        else:
            response = create_success_response()

        # 검증 보고서 추가
        if all_violations:
            validation_report = validator.create_validation_report(all_violations)
            response["validation_report"] = validation_report

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # 예외 발생 시 차단하지 않고 로그만 남김
        error_response = {
            "block_execution": False,
            "error": f"Hook execution error: {str(e)}",
            "message": "Hook 실행 중 오류가 발생했지만 작업을 진행합니다."
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True
            error_response["message"] = "Hook failed but continuing due to graceful degradation"

        print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()
