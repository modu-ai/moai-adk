# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_project_initializer.py
"""프로젝트 초기화 모듈

.moai/ 디렉토리 구조를 생성하고 config.json을 작성합니다.
"""

# REFACTOR: Improve type hints and docstrings for better readability

import json
from pathlib import Path

from moai_adk.core.project.detector import LanguageDetector


class ProjectInitializer:
    """프로젝트 초기화"""

    MOAI_STRUCTURE: list[str] = [
        ".moai/config.json",
        ".moai/project/product.md",
        ".moai/project/structure.md",
        ".moai/project/tech.md",
        ".moai/specs/",
        ".moai/memory/",
        ".moai/backup/",
    ]

    def __init__(self, path: str | Path = ".") -> None:
        """초기화

        Args:
            path: 프로젝트 루트 디렉토리
        """
        self.path = Path(path)
        self.detector = LanguageDetector()

    def initialize(
        self,
        mode: str = "personal",
        locale: str = "ko",
        language: str | None = None,
    ) -> dict[str, str]:
        """프로젝트 초기화 실행

        Args:
            mode: 프로젝트 모드 (personal/team)
            locale: 로케일 (ko/en/ja/zh)
            language: 강제 언어 지정 (None이면 자동 감지)

        Returns:
            초기화 결과 딕셔너리
        """
        # 1. 언어 감지
        detected_language = language or self.detector.detect(self.path) or "generic"

        # 2. 디렉토리 생성
        self._create_directories()

        # 3. config.json 생성
        config = {
            "projectName": self.path.name,
            "mode": mode,
            "locale": locale,
            "language": detected_language,
        }
        self._write_config(config)

        # 4. 결과 반환
        return {
            "path": str(self.path),
            "language": detected_language,
            "mode": mode,
            "locale": locale,
        }

    def is_initialized(self) -> bool:
        """.moai/ 디렉토리 존재 여부 확인

        Returns:
            초기화 여부
        """
        return (self.path / ".moai").exists()

    def _create_directories(self) -> None:
        """디렉토리 구조 생성"""
        for item in self.MOAI_STRUCTURE:
            target_path = self.path / item

            # 디렉토리인 경우 (끝이 /로 끝남)
            if item.endswith("/"):
                target_path.mkdir(parents=True, exist_ok=True)
            # 파일인 경우 부모 디렉토리만 생성
            else:
                target_path.parent.mkdir(parents=True, exist_ok=True)

    def _write_config(self, config: dict[str, str]) -> None:
        """config.json 작성

        Args:
            config: 설정 딕셔너리
        """
        config_path = self.path / ".moai" / "config.json"
        with open(config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=2, ensure_ascii=False)


def initialize_project(project_path: Path) -> None:
    """프로젝트 초기화 (CLI 명령어용)

    Args:
        project_path: 프로젝트 디렉토리 경로
    """
    initializer = ProjectInitializer(project_path)
    initializer.initialize()