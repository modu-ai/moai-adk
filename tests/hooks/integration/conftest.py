"""Shared fixtures for Hook integration tests"""

import json
import subprocess
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest


@pytest.fixture
def hook_tmp_project(tmp_path):
    """Create a temporary project structure with .moai directories"""
    project_root = tmp_path / "test_project"
    project_root.mkdir()

    # Create .moai directory structure
    moai_dir = project_root / ".moai"
    moai_dir.mkdir()

    (moai_dir / "config").mkdir()
    (moai_dir / "cache").mkdir()
    (moai_dir / "logs").mkdir()
    (moai_dir / "logs" / "sessions").mkdir()
    (moai_dir / "reports").mkdir()
    (moai_dir / "specs").mkdir()
    (moai_dir / "temp").mkdir()
    (moai_dir / "memory").mkdir()

    # Change to project root for tests
    original_cwd = Path.cwd()

    class ChangeDir:
        def __enter__(self):
            import os

            os.chdir(project_root)
            return project_root

        def __exit__(self, *args):
            import os

            os.chdir(original_cwd)

    yield ChangeDir()

    # Cleanup (restore cwd)
    import os

    if os.getcwd() != str(original_cwd):
        os.chdir(original_cwd)


@pytest.fixture
def valid_config_dict():
    """Return a valid MoAI config dictionary"""
    return {
        "moai": {"version": "0.26.0"},
        "project": {"name": "test-project", "initialized": True},
        "language": {"conversation_language": "en"},
        "session": {"suppress_setup_messages": False, "setup_messages_suppressed_at": None},
        "auto_cleanup": {"enabled": True, "cleanup_days": 7, "max_reports": 10, "last_cleanup": None},
        "daily_analysis": {"enabled": True, "last_analysis": None},
        "hooks": {"timeout_ms": 5000, "graceful_degradation": True},
        "document_management": {
            "enabled": True,
            "root_whitelist": ["README.md", "LICENSE", "pyproject.toml", "package.json", ".env.example"],
            "file_patterns": {
                "spec": {"spec_docs": ["SPEC-*.md"]},
                "docs": {"implementation": ["*.implementation.md"]},
            },
            "directories": {"spec": {"base": ".moai/specs/"}, "docs": {"base": ".moai/docs/"}},
        },
    }


@pytest.fixture
def config_file(hook_tmp_project, valid_config_dict):
    """Create a config.json file in .moai/config"""
    with hook_tmp_project as proj_root:
        config_path = proj_root / ".moai" / "config" / "config.json"
        config_path.write_text(json.dumps(valid_config_dict, indent=2))
        return config_path


@pytest.fixture
def hook_payload():
    """Return a typical hook payload"""
    return {
        "event": "session_start",
        "timestamp": datetime.now().isoformat(),
        "context": {"project": "test-project", "cwd": str(Path.cwd())},
    }


@pytest.fixture
def mock_subprocess():
    """Mock subprocess.run for git commands"""
    with patch("subprocess.run") as mock_run:
        # Default git command responses
        def git_command_response(cmd, *args, **kwargs):
            result = MagicMock()
            result.returncode = 0
            result.stdout = ""
            result.stderr = ""

            if isinstance(cmd, list) and len(cmd) > 0:
                if cmd[0] == "git":
                    if len(cmd) > 1:
                        if cmd[1] == "branch":
                            result.stdout = "main\n"
                        elif cmd[1] == "log":
                            result.stdout = "abc1234 Initial commit\n"
                        elif cmd[1] == "status":
                            result.stdout = ""
                        elif cmd[1] == "rev-parse":
                            result.stdout = "main\n"
                        elif cmd[1] == "rev-list":
                            result.stdout = ""

            return result

        mock_run.side_effect = git_command_response
        yield mock_run


@pytest.fixture
def cleanup_test_files(hook_tmp_project):
    """Create test files for cleanup operations"""
    with hook_tmp_project as proj_root:
        # Create old files (older than 7 days)
        old_date = datetime.now() - timedelta(days=10)
        timestamp_old = old_date.timestamp()

        # Old report files
        reports_dir = proj_root / ".moai" / "reports"
        for i in range(3):
            old_report = reports_dir / f"report-20231201-{i:02d}.json"
            old_report.write_text(json.dumps({"data": f"old-{i}"}))
            Path(old_report).touch((timestamp_old, timestamp_old))

        # Recent report files
        recent_report = reports_dir / "report-recent.json"
        recent_report.write_text(json.dumps({"data": "recent"}))

        # Old cache files
        cache_dir = proj_root / ".moai" / "cache"
        old_cache = cache_dir / "git-info-old.json"
        old_cache.write_text(json.dumps({"branch": "old"}))
        Path(old_cache).touch((timestamp_old, timestamp_old))

        yield {"reports_dir": reports_dir, "cache_dir": cache_dir, "old_timestamp": timestamp_old}


@pytest.fixture
def git_repo_with_changes(hook_tmp_project):
    """Create a git repository with uncommitted changes"""
    with hook_tmp_project as proj_root:
        # Initialize git repo
        subprocess.run(["git", "init"], cwd=proj_root, capture_output=True)

        # Configure git
        subprocess.run(["git", "config", "user.email", "test@example.com"], cwd=proj_root, capture_output=True)

        subprocess.run(["git", "config", "user.name", "Test User"], cwd=proj_root, capture_output=True)

        # Create initial commit
        (proj_root / "README.md").write_text("# Test Project")
        subprocess.run(["git", "add", "README.md"], cwd=proj_root, capture_output=True)
        subprocess.run(["git", "commit", "-m", "Initial commit"], cwd=proj_root, capture_output=True)

        # Create uncommitted changes
        (proj_root / "modified.txt").write_text("Modified content")
        (proj_root / "untracked.txt").write_text("Untracked file")

        yield proj_root


@pytest.fixture
def spec_files(hook_tmp_project):
    """Create SPEC files for testing"""
    with hook_tmp_project as proj_root:
        specs_dir = proj_root / ".moai" / "specs"

        # Create SPEC-001
        spec_001_dir = specs_dir / "SPEC-001"
        spec_001_dir.mkdir()
        (spec_001_dir / "spec.md").write_text("# SPEC-001: Test Feature")

        # Create SPEC-002
        spec_002_dir = specs_dir / "SPEC-002"
        spec_002_dir.mkdir()
        (spec_002_dir / "spec.md").write_text("# SPEC-002: Another Feature")

        # Create incomplete SPEC-003 (no spec.md)
        spec_003_dir = specs_dir / "SPEC-003"
        spec_003_dir.mkdir()

        yield specs_dir


@pytest.fixture
def session_state_file(hook_tmp_project):
    """Create a session state file"""
    with hook_tmp_project as proj_root:
        memory_dir = proj_root / ".moai" / "memory"
        state_file = memory_dir / "command-execution-state.json"

        state_data = {
            "last_specs": ["SPEC-001", "SPEC-002"],
            "last_command": "/moai:2-run",
            "timestamp": datetime.now().isoformat(),
        }

        state_file.write_text(json.dumps(state_data, indent=2))
        return state_file
