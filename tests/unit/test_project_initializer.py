# @TEST:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md
"""ProjectInitializer 테스트 스위트

프로젝트 초기화 및 .moai/ 디렉토리 구조 생성을 검증합니다.
"""

import json
from pathlib import Path

from moai_adk.core.project.initializer import ProjectInitializer


class TestProjectInitializer:
    """ProjectInitializer 클래스 테스트"""

    def test_initialize_creates_moai_directory(self, tmp_path: Path) -> None:
        """.moai/ 디렉토리 생성"""
        initializer = ProjectInitializer(tmp_path)

        initializer.initialize()

        assert (tmp_path / ".moai").exists()
        assert (tmp_path / ".moai").is_dir()

    def test_initialize_creates_all_subdirectories(self, tmp_path: Path) -> None:
        """모든 하위 디렉토리 생성"""
        initializer = ProjectInitializer(tmp_path)

        initializer.initialize()

        assert (tmp_path / ".moai" / "project").exists()
        assert (tmp_path / ".moai" / "specs").exists()
        assert (tmp_path / ".moai" / "memory").exists()
        assert (tmp_path / ".moai" / "backup").exists()

    def test_initialize_creates_config_json(self, tmp_path: Path) -> None:
        """config.json 파일 생성"""
        initializer = ProjectInitializer(tmp_path)

        initializer.initialize()

        config_path = tmp_path / ".moai" / "config.json"
        assert config_path.exists()

        with open(config_path, encoding="utf-8") as f:
            config = json.load(f)
            assert "mode" in config
            assert "locale" in config

    def test_initialize_with_personal_mode(self, tmp_path: Path) -> None:
        """personal 모드로 초기화"""
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize(mode="personal")

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, encoding="utf-8") as f:
            config = json.load(f)
            assert config["mode"] == "personal"
        assert result["mode"] == "personal"

    def test_initialize_with_team_mode(self, tmp_path: Path) -> None:
        """team 모드로 초기화"""
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize(mode="team")

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, encoding="utf-8") as f:
            config = json.load(f)
            assert config["mode"] == "team"
        assert result["mode"] == "team"

    def test_initialize_with_korean_locale(self, tmp_path: Path) -> None:
        """한국어 로케일로 초기화"""
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize(locale="ko")

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, encoding="utf-8") as f:
            config = json.load(f)
            assert config["locale"] == "ko"
        assert result["locale"] == "ko"

    def test_initialize_with_english_locale(self, tmp_path: Path) -> None:
        """영어 로케일로 초기화"""
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize(locale="en")

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, encoding="utf-8") as f:
            config = json.load(f)
            assert config["locale"] == "en"
        assert result["locale"] == "en"

    def test_initialize_detects_python_language(self, tmp_path: Path) -> None:
        """Python 프로젝트 감지"""
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize()

        assert result["language"] == "python"

    def test_initialize_detects_typescript_language(self, tmp_path: Path) -> None:
        """TypeScript 프로젝트 감지"""
        (tmp_path / "tsconfig.json").write_text("{}")
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize()

        assert result["language"] == "typescript"

    def test_initialize_uses_generic_for_unknown_language(self, tmp_path: Path) -> None:
        """알 수 없는 언어일 때 generic 사용"""
        # 빈 디렉토리
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize()

        assert result["language"] == "generic"

    def test_initialize_with_forced_language(self, tmp_path: Path) -> None:
        """강제 언어 지정"""
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize(language="rust")

        assert result["language"] == "rust"

    def test_initialize_sets_project_name(self, tmp_path: Path) -> None:
        """프로젝트명 설정"""
        initializer = ProjectInitializer(tmp_path)

        initializer.initialize()

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, encoding="utf-8") as f:
            config = json.load(f)
            assert config["projectName"] == tmp_path.name

    def test_initialize_returns_result_dict(self, tmp_path: Path) -> None:
        """초기화 결과 딕셔너리 반환"""
        initializer = ProjectInitializer(tmp_path)

        result = initializer.initialize(mode="team", locale="en")

        assert "path" in result
        assert "language" in result
        assert "mode" in result
        assert "locale" in result
        assert result["path"] == str(tmp_path)

    def test_is_initialized_returns_true_when_moai_exists(self, tmp_path: Path) -> None:
        """.moai/ 디렉토리 존재 시 True 반환"""
        (tmp_path / ".moai").mkdir()
        initializer = ProjectInitializer(tmp_path)

        assert initializer.is_initialized() is True

    def test_is_initialized_returns_false_when_moai_not_exists(
        self, tmp_path: Path
    ) -> None:
        """.moai/ 디렉토리 미존재 시 False 반환"""
        initializer = ProjectInitializer(tmp_path)

        assert initializer.is_initialized() is False

    def test_initialize_twice_does_not_fail(self, tmp_path: Path) -> None:
        """중복 초기화 시 실패하지 않음 (exist_ok=True)"""
        initializer = ProjectInitializer(tmp_path)

        initializer.initialize()
        initializer.initialize()  # 두 번째 초기화

        assert (tmp_path / ".moai").exists()

    def test_moai_structure_constant_completeness(self) -> None:
        """MOAI_STRUCTURE 상수가 모든 필수 항목 포함"""
        structure = ProjectInitializer.MOAI_STRUCTURE

        assert ".moai/config.json" in structure
        assert ".moai/specs/" in structure
        assert ".moai/memory/" in structure
        assert ".moai/backup/" in structure

    def test_create_directories_handles_file_paths(self, tmp_path: Path) -> None:
        """파일 경로는 부모 디렉토리만 생성"""
        initializer = ProjectInitializer(tmp_path)

        initializer._create_directories()

        # .moai/config.json은 파일이므로 부모 디렉토리만 생성됨
        assert (tmp_path / ".moai").exists()
        # 파일 자체는 생성되지 않음
        assert not (tmp_path / ".moai" / "config.json").exists()

    def test_create_directories_handles_directory_paths(self, tmp_path: Path) -> None:
        """디렉토리 경로는 디렉토리 생성"""
        initializer = ProjectInitializer(tmp_path)

        initializer._create_directories()

        # .moai/specs/는 디렉토리이므로 디렉토리 생성됨
        assert (tmp_path / ".moai" / "specs").exists()
        assert (tmp_path / ".moai" / "specs").is_dir()
