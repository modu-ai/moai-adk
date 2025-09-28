#!/usr/bin/env python3
"""
Configuration management for check_coverage module
"""

import json
from pathlib import Path
from typing import Any, Dict


def load_coverage_config(moai_dir: Path) -> Dict[str, Any]:
    """Load coverage configuration"""
    default_config = {
        "min_coverage": 80.0,
        "min_branch_coverage": 75.0,
        "include_patterns": ["src/**/*.py", "app/**/*.py", "lib/**/*.py"],
        "exclude_patterns": [
            "tests/**/*.py", "test_*.py", "*_test.py", "setup.py",
            "conftest.py", "*/migrations/*", "*/venv/*", "*/node_modules/*",
        ],
        "fail_under": True,
        "show_missing": True,
        "skip_covered": False,
        "precision": 2,
    }

    # Load from .moai/config.json
    config_file = moai_dir / "config.json"
    if config_file.exists():
        try:
            moai_config = json.loads(config_file.read_text())
            coverage_config = moai_config.get("coverage", {})
            default_config.update(coverage_config)
        except Exception as error:
            print(f"Warning: Failed to load coverage config: {error}")

    return default_config


def detect_test_framework(project_root: Path) -> str:
    """Detect test framework in use"""
    # Check for pytest
    if has_pytest(project_root):
        return "pytest"

    # Check for unittest (Python default)
    test_files = list(project_root.rglob("test_*.py"))
    if test_files:
        return "unittest"

    return "pytest"  # Default fallback


def has_pytest(project_root: Path) -> bool:
    """Check if pytest is available"""
    pytest_ini = project_root / "pytest.ini"
    pyproject_toml = project_root / "pyproject.toml"
    setup_cfg = project_root / "setup.cfg"

    return any([
        pytest_ini.exists(),
        pyproject_toml.exists(),
        setup_cfg.exists()
    ])