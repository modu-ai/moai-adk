"""Tests for .claude/hooks/alfred/core/project.py module

프로젝트 언어 감지, Git 정보, SPEC 카운팅 테스트
"""
import importlib.util
import json
import sys
import time
from pathlib import Path

import pytest


def _load_project_module(module_name: str = "project_module"):
    """Dynamically load core/project.py as a fresh module."""
    repo_root = Path(__file__).resolve().parents[2]

    # Add hooks directory to sys.path
    hooks_dir = repo_root / "src" / "moai_adk" / "templates" / ".claude" / "hooks" / "alfred"
    if str(hooks_dir) not in sys.path:
        sys.path.insert(0, str(hooks_dir))

    module_path = hooks_dir / "core" / "project.py"

    spec = importlib.util.spec_from_file_location(module_name, module_path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


@pytest.fixture
def project_module():
    """Provide a clean project module instance per test."""
    module_name = f"project_module_{time.time_ns()}"
    module = _load_project_module(module_name=module_name)
    yield module
    sys.modules.pop(module_name, None)


def test_get_project_language_prefers_config(monkeypatch: pytest.MonkeyPatch, tmp_path: Path, project_module):
    """get_project_language()가 .moai/config.json을 우선하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    config_path = moai_dir / "config.json"
    config_path.write_text(json.dumps({"language": "python"}), encoding="utf-8")

    assert project_module.get_project_language(str(project_dir)) == "python"


def test_get_project_language_falls_back_to_detect(monkeypatch: pytest.MonkeyPatch, tmp_path: Path, project_module):
    """get_project_language()가 config 없을 때 detect_language()로 fallback하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()

    config_path = moai_dir / "config.json"
    config_path.write_text("not-json", encoding="utf-8")

    monkeypatch.setattr(project_module, "detect_language", lambda cwd: "fallback")
    assert project_module.get_project_language(str(project_dir)) == "fallback"

    config_path.unlink()
    assert project_module.get_project_language(str(project_dir)) == "fallback"


def test_detect_language_python(tmp_path: Path, project_module):
    """detect_language()가 Python 프로젝트를 감지하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Create pyproject.toml
    (project_dir / "pyproject.toml").write_text("[tool.poetry]\nname = 'test'", encoding="utf-8")

    assert project_module.detect_language(str(project_dir)) == "python"


def test_detect_language_typescript_with_tsconfig(tmp_path: Path, project_module):
    """detect_language()가 tsconfig.json이 있는 TypeScript 프로젝트를 감지하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Create both package.json and tsconfig.json
    (project_dir / "package.json").write_text(
        json.dumps({"dependencies": {"typescript": "^5.0.0"}}),
        encoding="utf-8"
    )
    (project_dir / "tsconfig.json").write_text("{}", encoding="utf-8")

    assert project_module.detect_language(str(project_dir)) == "typescript"


def test_detect_language_javascript_without_tsconfig(tmp_path: Path, project_module):
    """detect_language()가 tsconfig.json 없는 JavaScript 프로젝트를 감지하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Create package.json without tsconfig.json
    (project_dir / "package.json").write_text(
        json.dumps({"dependencies": {"react": "^18.0.0"}}),
        encoding="utf-8"
    )

    assert project_module.detect_language(str(project_dir)) == "javascript"


def test_count_specs(tmp_path: Path, project_module):
    """count_specs()가 SPEC 파일들을 정확히 카운트하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()
    specs_dir = moai_dir / "specs"
    specs_dir.mkdir()

    # Create SPEC directories
    for i in range(1, 4):
        spec_dir = specs_dir / f"SPEC-AUTH-{i:03d}"
        spec_dir.mkdir()
        spec_file = spec_dir / "spec.md"
        spec_file.write_text(f"---\nstatus: {'completed' if i == 1 else 'active'}\n---\n", encoding="utf-8")

    result = project_module.count_specs(str(project_dir))

    assert result["total"] == 3
    assert result["completed"] == 1
    # Calculate expected percentage: 1/3 * 100 = 33
    assert result["percentage"] == 33
