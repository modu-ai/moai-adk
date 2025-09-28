#!/usr/bin/env python3
"""
Test runner for check_coverage module
"""

import subprocess
import json
from pathlib import Path
from typing import Dict, Any, List
from .models import CoverageResult


def run_coverage_test(project_root: Path, config: Dict[str, Any]) -> CoverageResult:
    """Run coverage test with pytest-cov"""
    if not has_pytest_cov():
        raise RuntimeError(
            "pytest-cov not available. Install with: pip install pytest-cov"
        )

    # Determine coverage target directories
    src_dirs = [
        pattern.split("/")[0] for pattern in config["include_patterns"]
        if (project_root / pattern.split("/")[0]).exists()
    ] or ["src"]

    # Build pytest command
    cmd = [
        "python", "-m", "pytest",
        "--cov=" + ",".join(src_dirs),
        "--cov-report=term-missing",
        "--cov-report=json:coverage.json",
        "--cov-report=html:htmlcov",
        "--cov-branch",
        f"--cov-fail-under={config['min_coverage']}",
        "-v",
    ]

    # Add test directories
    test_dirs = [
        test_dir for test_dir in ["tests", "test"]
        if (project_root / test_dir).exists()
    ]

    cmd.extend(test_dirs if test_dirs else ["-k", "test_"])

    print(f"Running: {' '.join(cmd)}")
    try:
        result = subprocess.run(
            cmd, cwd=project_root, capture_output=True, text=True, timeout=300
        )
        from .parser import parse_json_coverage, parse_text_coverage
        coverage_file = project_root / "coverage.json"
        if coverage_file.exists():
            return parse_json_coverage(coverage_file)
        else:
            return parse_text_coverage(result.stdout)
    except Exception as error:
        raise RuntimeError(f"Coverage test failed: {error}")


def has_pytest_cov() -> bool:
    """Check if pytest-cov is available"""
    try:
        import pytest_cov
        return True
    except ImportError:
        return False