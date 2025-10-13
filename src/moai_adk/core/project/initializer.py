# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_project_initializer.py
"""프로젝트 초기화 모듈"""

from pathlib import Path

from moai_adk.core.project.detector import LanguageDetector
from moai_adk.core.template.config import ConfigManager


class ProjectInitializer:
    """프로젝트 초기화

    .moai/ 디렉토리 구조를 생성하고 언어별 템플릿을 적용합니다.
    """

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
        """ProjectInitializer 초기화

        Args:
            path: 프로젝트 디렉토리 경로
        """
        self.path = Path(path)
        self.detector = LanguageDetector()

    def initialize(
        self, mode: str = "personal", locale: str = "ko", language: str | None = None
    ) -> dict[str, str]:
        """프로젝트 초기화

        Args:
            mode: 프로젝트 모드 (personal, team)
            locale: 언어 설정 (ko, en)
            language: 강제 언어 지정 (None이면 자동 감지)

        Returns:
            초기화 결과 정보
        """
        # 1. 언어 감지
        if language is None:
            language = self.detector.detect(str(self.path))

        if not language:
            language = "generic"

        # 2. 디렉토리 생성
        self._create_directories()

        # 3. config.json 생성
        config_manager = ConfigManager(str(self.path / ".moai/config.json"))
        config = config_manager.DEFAULT_CONFIG.copy()
        config["mode"] = mode
        config["locale"] = locale
        config["projectName"] = self.path.name
        config_manager.save(config)

        return {
            "path": str(self.path),
            "language": language,
            "mode": mode,
            "locale": locale,
        }

    def _create_directories(self) -> None:
        """디렉토리 구조 생성"""
        for item in self.MOAI_STRUCTURE:
            full_path = self.path / item
            if item.endswith("/"):
                # 디렉토리
                full_path.mkdir(parents=True, exist_ok=True)
            else:
                # 파일 (부모 디렉토리만 생성)
                full_path.parent.mkdir(parents=True, exist_ok=True)

    def is_initialized(self) -> bool:
        """프로젝트 초기화 여부 확인

        Returns:
            .moai/ 디렉토리 존재 여부
        """
        return (self.path / ".moai").exists()
