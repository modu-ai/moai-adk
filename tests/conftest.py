# @TEST:CONFTEST-001 | SPEC: SPEC-TEST-COVERAGE-001.md
"""Pytest configuration and shared fixtures for MoAI-ADK tests"""

import shutil
import sys
import tempfile
from pathlib import Path
from typing import Generator

import pytest

# Hook 디렉토리를 sys.path에 추가 (hooks 테스트가 모듈을 찾을 수 있도록)
HOOKS_DIR = Path(__file__).parent.parent / ".claude" / "hooks" / "alfred"
SHARED_DIR = HOOKS_DIR / "shared"
UTILS_DIR = HOOKS_DIR / "utils"

if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))
if str(HOOKS_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_DIR))
if str(UTILS_DIR) not in sys.path:
    sys.path.insert(0, str(UTILS_DIR))


@pytest.fixture
def tmp_project_dir() -> Generator[Path, None, None]:
    """Create a temporary directory for testing project operations.

    Yields:
        Path: Temporary project directory path

    Cleanup:
        Automatically removes directory after test
    """
    tmp_dir = Path(tempfile.mkdtemp(prefix="moai_test_"))
    try:
        yield tmp_dir
    finally:
        if tmp_dir.exists():
            shutil.rmtree(tmp_dir)


@pytest.fixture
def tmp_git_repo(tmp_project_dir: Path) -> Path:
    """Create a temporary Git repository for testing.

    Args:
        tmp_project_dir: Temporary directory fixture

    Returns:
        Path: Initialized Git repository path
    """
    import subprocess

    subprocess.run(["git", "init"], cwd=tmp_project_dir, check=True, capture_output=True)
    subprocess.run(
        ["git", "config", "user.name", "Test User"],
        cwd=tmp_project_dir,
        check=True,
        capture_output=True,
    )
    subprocess.run(
        ["git", "config", "user.email", "test@example.com"],
        cwd=tmp_project_dir,
        check=True,
        capture_output=True,
    )

    return tmp_project_dir


@pytest.fixture
def sample_moai_config() -> dict:
    """Provide sample .moai/config.json structure.

    Returns:
        dict: Sample configuration dictionary
    """
    return {
        "project": {
            "name": "test-project",
            "version": "0.0.1",
            "mode": "personal",
            "locale": "ko",
        },
        "git": {
            "default_branch": "main",
            "feature_prefix": "feature/",
        },
    }
