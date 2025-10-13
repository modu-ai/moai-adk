# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_project_*.py
"""프로젝트 초기화 및 관리 모듈

언어 감지, 시스템 체크, 프로젝트 초기화 기능 제공
"""

from pathlib import Path

from moai_adk.core.project.checker import SystemChecker
from moai_adk.core.project.detector import LanguageDetector
from moai_adk.core.project.initializer import ProjectInitializer

__all__ = [
    "LanguageDetector",
    "SystemChecker",
    "ProjectInitializer",
    "check_environment",
    "initialize_project",
    "get_project_status",
]


def check_environment() -> dict[str, bool]:
    """환경 검증 (CLI doctor 명령어용)

    Returns:
        도구명: 설치 여부 딕셔너리
    """
    checker = SystemChecker()
    return checker.check_all()


def initialize_project(path: str) -> None:
    """프로젝트 초기화 (CLI init 명령어용)

    Args:
        path: 프로젝트 경로
    """
    initializer = ProjectInitializer(path)
    initializer.initialize()


def get_project_status() -> dict[str, str | int]:
    """프로젝트 상태 조회 (CLI status 명령어용)

    Returns:
        프로젝트 상태 딕셔너리 (mode, locale, spec_count 등)
    """
    # TODO: 실제 config.json 읽기 구현 필요
    moai_dir = Path(".moai")

    if not moai_dir.exists():
        raise FileNotFoundError(
            "No .moai directory found. Run 'moai init .' to initialize a project."
        )

    # Placeholder: 기본값 반환
    return {
        "mode": "personal",
        "locale": "ko",
        "spec_count": 0,
    }
