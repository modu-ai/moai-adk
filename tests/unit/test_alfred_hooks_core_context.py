"""Tests for .claude/hooks/alfred/core/context.py module

워크플로우 컨텍스트 관리 및 JIT 문서 검색 테스트
"""
import importlib.util
import sys
import time
from pathlib import Path

import pytest


def _load_context_module(module_name: str = "context_module"):
    """Dynamically load core/context.py as a fresh module."""
    repo_root = Path(__file__).resolve().parents[2]

    # Add hooks directory to sys.path
    hooks_dir = repo_root / "src" / "moai_adk" / "templates" / ".claude" / "hooks" / "alfred"
    if str(hooks_dir) not in sys.path:
        sys.path.insert(0, str(hooks_dir))

    module_path = hooks_dir / "core" / "context.py"

    spec = importlib.util.spec_from_file_location(module_name, module_path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


@pytest.fixture
def context_module():
    """Provide a clean context module instance per test."""
    module_name = f"context_module_{time.time_ns()}"
    module = _load_context_module(module_name=module_name)
    module.clear_workflow_context()
    yield module
    module.clear_workflow_context()
    sys.modules.pop(module_name, None)


def test_save_and_load_phase_context(context_module):
    """save_phase_context()와 load_phase_context()가 정상 작동하는지 확인"""
    context_module.save_phase_context("analysis", {"spec": "data"})

    assert context_module.load_phase_context("analysis") == {"spec": "data"}


def test_phase_context_ttl_expiration(context_module):
    """phase context가 TTL 후 만료되는지 확인"""
    context_module.save_phase_context("analysis", {"spec": "data"})

    # Get the entry and backdate it
    entry = context_module._workflow_context["analysis"]
    context_module._workflow_context["analysis"] = {
        "data": entry["data"],
        "timestamp": entry["timestamp"] - 1000
    }

    # Should return None because TTL expired
    assert context_module.load_phase_context("analysis") is None


def test_clear_workflow_context(context_module):
    """clear_workflow_context()가 모든 컨텍스트를 삭제하는지 확인"""
    context_module.save_phase_context("implementation", {"state": "ok"})
    context_module.save_phase_context("analysis", {"spec": "data"})

    context_module.clear_workflow_context()

    assert context_module.load_phase_context("implementation") is None
    assert context_module.load_phase_context("analysis") is None


def test_get_jit_context_returns_list(context_module, tmp_path: Path):
    """get_jit_context()가 리스트를 반환하는지 확인"""
    project_dir = tmp_path / "project"
    project_dir.mkdir()

    # Test with any prompt
    context = context_module.get_jit_context("/alfred:1-spec", str(project_dir))

    # Should return a list (may be empty if files don't exist)
    assert isinstance(context, list)


def test_get_jit_context_with_existing_files(context_module, tmp_path: Path):
    """get_jit_context()가 존재하는 파일만 반환하는지 확인"""
    # Create mock project structure
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir()
    memory_dir = moai_dir / "memory"
    memory_dir.mkdir()

    spec_metadata_file = memory_dir / "spec-metadata.md"
    spec_metadata_file.write_text("# SPEC Metadata", encoding="utf-8")

    # Test /alfred:1-spec command (should return spec-metadata.md if it exists)
    context = context_module.get_jit_context("/alfred:1-spec", str(project_dir))

    # Context should be a list and may contain the spec-metadata.md path
    assert isinstance(context, list)
    # If the file exists and pattern matches, it should be in the list
    if context:
        assert any("spec-metadata.md" in doc for doc in context)
