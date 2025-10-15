# @CODE:CORE-PROJECT-003 | SPEC: SPEC-CORE-PROJECT-001.md
"""프로젝트 초기화 검증 모듈

시스템 요구사항 및 설치 결과 검증
"""

import shutil
from pathlib import Path


class ValidationError(Exception):
    """검증 실패 예외"""

    pass


class ProjectValidator:
    """프로젝트 초기화 검증"""

    # 필수 디렉토리 구조
    REQUIRED_DIRECTORIES = [
        ".moai/",
        ".moai/project/",
        ".moai/specs/",
        ".moai/memory/",
        ".claude/",
    ]

    # 필수 파일
    REQUIRED_FILES = [
        ".moai/config.json",
        "CLAUDE.md",
    ]

    def validate_system_requirements(self) -> None:
        """시스템 요구사항 검증

        Raises:
            ValidationError: 요구사항 미충족 시
        """
        # Git 설치 확인
        if not shutil.which("git"):
            raise ValidationError("Git is not installed")

        # Python 버전 확인 (3.10+)
        import sys

        if sys.version_info < (3, 10):
            raise ValidationError(
                f"Python 3.10+ required (current: {sys.version_info.major}.{sys.version_info.minor})"
            )

    def validate_project_path(self, project_path: Path) -> None:
        """프로젝트 경로 검증

        Args:
            project_path: 프로젝트 경로

        Raises:
            ValidationError: 경로가 유효하지 않은 경우
        """
        # 절대 경로 확인
        if not project_path.is_absolute():
            raise ValidationError(f"Project path must be absolute: {project_path}")

        # 부모 디렉토리 존재 확인
        if not project_path.parent.exists():
            raise ValidationError(f"Parent directory does not exist: {project_path.parent}")

        # MoAI-ADK 패키지 내부 확인
        if self._is_inside_moai_package(project_path):
            raise ValidationError(
                "Cannot initialize inside MoAI-ADK package directory"
            )

    def validate_installation(self, project_path: Path) -> None:
        """설치 결과 검증

        Args:
            project_path: 프로젝트 경로

        Raises:
            ValidationError: 설치가 완료되지 않은 경우
        """
        # 필수 디렉토리 확인
        for directory in self.REQUIRED_DIRECTORIES:
            dir_path = project_path / directory
            if not dir_path.exists():
                raise ValidationError(f"Required directory not found: {directory}")

        # 필수 파일 확인
        for file in self.REQUIRED_FILES:
            file_path = project_path / file
            if not file_path.exists():
                raise ValidationError(f"Required file not found: {file}")

    def _is_inside_moai_package(self, project_path: Path) -> bool:
        """MoAI-ADK 패키지 내부 경로 여부 확인

        Args:
            project_path: 확인할 경로

        Returns:
            패키지 내부이면 True
        """
        # pyproject.toml에 moai-adk가 있으면 패키지 루트
        current = project_path.resolve()
        while current != current.parent:
            pyproject = current / "pyproject.toml"
            if pyproject.exists():
                try:
                    content = pyproject.read_text(encoding="utf-8")
                    if "name = \"moai-adk\"" in content or 'name = "moai-adk"' in content:
                        return True
                except Exception:
                    pass
            current = current.parent
        return False
