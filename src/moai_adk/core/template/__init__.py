# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_foundation.py
"""Template 모듈: Jinja2 템플릿 엔진

패키지 내부의 템플릿 파일 접근을 위한 유틸리티:
- 템플릿 디렉토리 경로 조회
- 템플릿 파일 존재 확인
- 템플릿 리스트 조회
"""

from pathlib import Path


def get_templates_dir() -> Path:
    """패키지 내부의 templates 디렉토리 경로 반환

    Returns:
        templates 디렉토리 절대 경로
    """
    # src/moai_adk/core/template/__init__.py -> src/moai_adk/templates
    return Path(__file__).parent.parent.parent / "templates"


def get_template_path(relative_path: str) -> Path:
    """템플릿 파일의 절대 경로 반환

    Args:
        relative_path: 템플릿 루트 기준 상대 경로 (예: ".moai/config.json")

    Returns:
        템플릿 파일의 절대 경로
    """
    return get_templates_dir() / relative_path


def template_exists(relative_path: str) -> bool:
    """템플릿 파일 존재 여부 확인

    Args:
        relative_path: 템플릿 루트 기준 상대 경로

    Returns:
        파일 존재 여부
    """
    return get_template_path(relative_path).exists()


__all__ = [
    "get_templates_dir",
    "get_template_path",
    "template_exists",
]
