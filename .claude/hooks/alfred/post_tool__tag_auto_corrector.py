#!/usr/bin/env python3
# @CODE:HOOK-POST-TAG-001 | SPEC: TAG-POST-HOOK-001 | TEST: tests/hooks/test_post_tool_tag_corrector.py
"""PostToolUse Hook: TAG 자동 교정 및 모니터링

Edit/Write/MultiEdit 실행 후 TAG 상태를 검증하고 자동 교정.
실시간 모니터링으로 누락된 TAG를 탐지하고 수정 제안.

기능:
- 파일 수정 후 TAG 상태 검증
- 누락된 TAG 자동 생성 제안
- TAG 체인 무결성 검사
- 자동 수정 기능 (설정에 따라)

사용법:
    python3 post_tool__tag_auto_corrector.py <tool_name> <tool_args_json> <result_json>
"""

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List

# 모듈 경로 추가
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.auto_corrector import AutoCorrection, AutoCorrectionConfig, TagAutoCorrector
from moai_adk.core.tags.policy_validator import PolicyValidationConfig, PolicyViolation, TagPolicyValidator
from moai_adk.core.tags.rollback_manager import RollbackConfig, RollbackManager

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

    policy_config = PolicyValidationConfig(
        strict_mode=tag_policy_config.get("enforcement_mode", "strict") == "strict",
        require_spec_before_code=tag_policy_config.get("require_spec_before_code", True),
        require_test_for_code=tag_policy_config.get("require_test_for_code", True),
        allow_duplicate_tags=not tag_policy_config.get("enforce_chains", True),
        validation_timeout=tag_policy_config.get("realtime_validation", {}).get("validation_timeout", 5),
        auto_fix_enabled=tag_policy_config.get("auto_correction", {}).get("enabled", False)
    )

    return TagPolicyValidator(config=policy_config)


def create_auto_corrector() -> TagAutoCorrector:
    """TAG 자동 수정기 생성

    Returns:
        TagAutoCorrector 인스턴스
    """
    config_data = load_config()
    auto_correction_config = config_data.get("tags", {}).get("policy", {}).get("auto_correction", {})

    correction_config = AutoCorrectionConfig(
        enable_auto_fix=auto_correction_config.get("enabled", False),
        confidence_threshold=auto_correction_config.get("confidence_threshold", 0.8),
        create_missing_specs=auto_correction_config.get("create_missing_specs", False),
        create_missing_tests=auto_correction_config.get("create_missing_tests", False),
        remove_duplicates=auto_correction_config.get("remove_duplicates", True),
        backup_before_fix=auto_correction_config.get("backup_before_fix", True)
    )

    return TagAutoCorrector(config=correction_config)


def create_rollback_manager() -> RollbackManager:
    """롤백 관리자 생성

    Returns:
        RollbackManager 인스턴스
    """
    config_data = load_config()
    rollback_config = config_data.get("tags", {}).get("policy", {}).get("rollback", {})

    config = RollbackConfig(
        checkpoints_dir=rollback_config.get("checkpoints_dir", ".moai/checkpoints"),
        max_checkpoints=rollback_config.get("max_checkpoints", 10),
        auto_cleanup=rollback_config.get("auto_cleanup", True),
        backup_before_rollback=rollback_config.get("backup_before_rollback", True),
        rollback_timeout=rollback_config.get("rollback_timeout", 30)
    )

    return RollbackManager(config=config)


def should_monitor_tool(tool_name: str, tool_args: Dict[str, Any], result: Dict[str, Any]) -> bool:
    """모니터링 대상 툴인지 확인

    Args:
        tool_name: 툴 이름
        tool_args: 툴 인자
        result: 툴 실행 결과

    Returns:
        모니터링 대상이면 True
    """
    # 파일 조작 툴만 모니터링
    monitoring_tools = {"Edit", "Write", "MultiEdit"}

    # 성공한 경우에만 모니터링
    if tool_name not in monitoring_tools:
        return False

    if result.get("success", True):  # 기본값은 True
        return True

    return False


def extract_modified_files(tool_name: str, tool_args: Dict[str, Any]) -> List[str]:
    """수정된 파일 경로 추출

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
        edits = tool_args.get("edits", [])
        for edit in edits:
            file_path = edit.get("file_path", "")
            if file_path:
                file_paths.append(file_path)

    return file_paths


def get_current_file_content(file_path: str) -> str:
    """현재 파일 내용 가져오기

    Args:
        file_path: 파일 경로

    Returns:
        파일 내용
    """
    try:
        path = Path(file_path)
        if path.exists():
            return path.read_text(encoding="utf-8", errors="ignore")
    except Exception:
        pass

    return ""


def create_checkpoint_if_needed(rollback_manager: RollbackManager, file_paths: List[str]) -> Optional[str]:
    """필요시 체크포인트 생성

    Args:
        rollback_manager: 롤백 관리자
        file_paths: 파일 경로 목록

    Returns:
        체크포인트 ID 또는 None
    """
    try:
        # 중요 파일이 변경된 경우에만 체크포인트 생성
        important_files = [fp for fp in file_paths if any(
            pattern in fp for pattern in ["src/", "tests/", ".moai/", ".claude/"]
        )]

        if important_files:
            description = f"TAG 시스템 체크포인트: {len(important_files)}개 파일 수정"
            return rollback_manager.create_checkpoint(
                description=description,
                files=important_files,
                metadata={"tool": "post_tool_tag_corrector"}
            )
    except Exception:
        pass

    return None


def create_corrections_summary(corrections: List[AutoCorrection]) -> Dict[str, Any]:
    """수정 내용 요약 생성

    Args:
        corrections: 수정 목록

    Returns:
        수정 요약 딕셔너리
    """
    if not corrections:
        return {
            "total_corrections": 0,
            "applied_corrections": 0,
            "corrections": []
        }

    applied_corrections = [c for c in corrections if c.confidence >= 0.8]

    summary = {
        "total_corrections": len(corrections),
        "applied_corrections": len(applied_corrections),
        "high_confidence_corrections": len([c for c in corrections if c.confidence >= 0.9]),
        "corrections": []
    }

    for correction in corrections:
        summary["corrections"].append({
            "file_path": correction.file_path,
            "description": correction.description,
            "confidence": correction.confidence,
            "requires_review": correction.requires_review,
            "applied": correction.confidence >= 0.8
        })

    return summary


def create_monitoring_response(
    violations: List[PolicyViolation],
    corrections: List[AutoCorrection],
    checkpoint_id: Optional[str] = None
) -> Dict[str, Any]:
    """모니터링 결과 응답 생성

    Args:
        violations: 정책 위반 목록
        corrections: 수정 목록
        checkpoint_id: 체크포인트 ID

    Returns:
        모니터링 응답 딕셔너리
    """
    response = {
        "monitoring_completed": True,
        "timestamp": time.time(),
        "violations_found": len(violations),
        "corrections_available": len(corrections),
        "checkpoint_created": checkpoint_id is not None,
        "checkpoint_id": checkpoint_id
    }

    # 위반 정보 추가
    if violations:
        response["violations"] = [v.to_dict() for v in violations]
        response["violation_summary"] = {
            "critical": len([v for v in violations if v.level.value == "critical"]),
            "high": len([v for v in violations if v.level.value == "high"]),
            "medium": len([v for v in violations if v.level.value == "medium"]),
            "low": len([v for v in violations if v.level.value == "low"])
        }

    # 수정 정보 추가
    if corrections:
        response["corrections"] = create_corrections_summary(corrections)

    return response


def main() -> None:
    """메인 함수"""
    try:
        # 설정에서 타임아웃 값 로드 (밀리초 → 초)
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # 인자 파싱
        if len(sys.argv) < 4:
            print(json.dumps({
                "monitoring_completed": False,
                "error": "Invalid arguments. Usage: python3 post_tool__tag_auto_corrector.py <tool_name> <tool_args_json> <result_json>"
            }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
            tool_result = json.loads(sys.argv[3])
        except json.JSONDecodeError as e:
            print(json.dumps({
                "monitoring_completed": False,
                "error": f"Invalid JSON: {str(e)}"
            }))
            sys.exit(0)

        # 시작 시간 기록
        start_time = time.time()

        # 모니터링 대상 확인
        if not should_monitor_tool(tool_name, tool_args, tool_result):
            print(json.dumps({
                "monitoring_completed": True,
                "message": "툴 실행 결과 모니터링 대상 아님"
            }))
            sys.exit(0)

        # 수정된 파일 경로 추출
        file_paths = extract_modified_files(tool_name, tool_args)
        if not file_paths:
            print(json.dumps({
                "monitoring_completed": True,
                "message": "수정된 파일 없음"
            }))
            sys.exit(0)

        # 구성요소 생성
        policy_validator = create_policy_validator()
        auto_corrector = create_auto_corrector()
        rollback_manager = create_rollback_manager()

        # 체크포인트 생성 (중요 파일인 경우)
        checkpoint_id = create_checkpoint_if_needed(rollback_manager, file_paths)

        # 모든 파일에 대해 검증 및 수정
        all_violations = []
        all_corrections = []

        for file_path in file_paths:
            # 타임아웃 체크
            if time.time() - start_time > timeout_seconds:
                break

            # 현재 파일 내용 가져오기
            content = get_current_file_content(file_path)
            if not content:
                continue

            # 정책 위반 검증
            violations = policy_validator.validate_after_modification(file_path, content)
            all_violations.extend(violations)

            # 자동 수정 생성
            if violations:
                corrections = auto_corrector.generate_corrections(violations)
                all_corrections.extend(corrections)

        # 자동 수정 적용
        applied_corrections = []
        if all_corrections and auto_corrector.config.enable_auto_fix:
            success = auto_corrector.apply_corrections(all_corrections)
            if success:
                applied_corrections = [c for c in all_corrections
                                     if c.confidence >= auto_corrector.config.confidence_threshold]

        # 응답 생성
        response = create_monitoring_response(all_violations, all_corrections, checkpoint_id)

        # 추가 정보
        if applied_corrections:
            response["auto_corrections_applied"] = len(applied_corrections)
            response["message"] = f"✅ {len(applied_corrections)}개 자동 수정 적용 완료"
        elif all_corrections:
            response["message"] = f"💡 {len(all_corrections)}개 수정 제안 생성 (자동 적용 비활성화)"
        elif all_violations:
            response["message"] = f"⚠️ {len(all_violations)}개 TAG 정책 위반 발견"
        else:
            response["message"] = "✅ TAG 정책 준수 확인"

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # 예외 발생 시 로그만 남기고 계속 진행
        error_response = {
            "monitoring_completed": False,
            "error": f"Hook execution error: {str(e)}",
            "message": "Hook 실행 중 오류가 발생했지만 정상 처리됨"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True
            error_response["message"] = "Hook failed but continuing due to graceful degradation"

        print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()
