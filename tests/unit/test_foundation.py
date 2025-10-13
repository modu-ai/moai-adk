# @TEST:PY314-001 | SPEC: SPEC-PY314-001.md
"""Test PY314-001: Python 3.14 Foundation & Build System

Phase 1 Tests: 프로젝트 구조 검증
- pyproject.toml 존재 확인
- 필수 필드 검증 (name, version, dependencies)
- 디렉토리 구조 확인
"""

import tomllib
from pathlib import Path


class TestProjectStructure:
    """프로젝트 구조 검증 테스트"""

    def test_pyproject_toml_exists(self):
        """pyproject.toml 파일이 존재해야 한다"""
        project_root = Path(__file__).parent.parent.parent
        pyproject_path = project_root / "pyproject.toml"
        assert pyproject_path.exists(), "pyproject.toml 파일이 없습니다"

    def test_pyproject_metadata(self):
        """pyproject.toml에 필수 메타데이터가 포함되어야 한다"""
        project_root = Path(__file__).parent.parent.parent
        pyproject_path = project_root / "pyproject.toml"

        with open(pyproject_path, "rb") as f:
            config = tomllib.load(f)

        # 필수 필드 검증
        assert "project" in config, "project 섹션이 없습니다"
        project = config["project"]

        assert project["name"] == "moai-adk", "패키지명이 moai-adk가 아닙니다"
        assert project["version"] == "0.3.0", "버전이 0.3.0이 아닙니다"
        assert ">=3.13" in project["requires-python"], "Python 3.13+ 요구사항이 없습니다"

    def test_core_dependencies(self):
        """핵심 6개 의존성이 정의되어야 한다"""
        project_root = Path(__file__).parent.parent.parent
        pyproject_path = project_root / "pyproject.toml"

        with open(pyproject_path, "rb") as f:
            config = tomllib.load(f)

        dependencies = config["project"]["dependencies"]
        required_deps = ["click", "rich", "questionary", "gitpython", "jinja2", "pyyaml"]

        for dep in required_deps:
            assert any(dep.lower() in d.lower() for d in dependencies), f"{dep} 의존성이 없습니다"

    def test_moai_adk_directory_structure(self):
        """moai-adk-py/src/moai_adk/ 디렉토리가 존재해야 한다"""
        project_root = Path(__file__).parent.parent.parent
        src_dir = project_root / "moai-adk-py" / "src" / "moai_adk"
        assert src_dir.exists(), "moai-adk-py/src/moai_adk/ 디렉토리가 없습니다"

    def test_init_file_exists(self):
        """__init__.py 파일이 존재해야 한다"""
        project_root = Path(__file__).parent.parent.parent
        init_file = project_root / "moai-adk-py" / "src" / "moai_adk" / "__init__.py"
        assert init_file.exists(), "__init__.py 파일이 없습니다"

    def test_main_file_exists(self):
        """__main__.py 파일이 존재해야 한다"""
        project_root = Path(__file__).parent.parent.parent
        main_file = project_root / "moai-adk-py" / "src" / "moai_adk" / "__main__.py"
        assert main_file.exists(), "__main__.py 파일이 없습니다"

    def test_build_system_configured(self):
        """빌드 시스템이 hatchling으로 설정되어야 한다"""
        project_root = Path(__file__).parent.parent.parent
        pyproject_path = project_root / "pyproject.toml"

        with open(pyproject_path, "rb") as f:
            config = tomllib.load(f)

        assert "build-system" in config, "build-system 섹션이 없습니다"
        build_system = config["build-system"]

        assert "hatchling" in build_system["requires"], "hatchling이 빌드 백엔드가 아닙니다"
        assert build_system["build-backend"] == "hatchling.build", (
            "빌드 백엔드가 hatchling.build가 아닙니다"
        )
