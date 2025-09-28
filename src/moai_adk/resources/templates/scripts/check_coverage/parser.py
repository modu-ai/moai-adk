#!/usr/bin/env python3
"""
Coverage result parsing for check_coverage module
"""

import json
import re
from pathlib import Path
from typing import Dict, List
from .models import CoverageResult, FileCoverage


def parse_json_coverage(coverage_file: Path) -> CoverageResult:
    """Parse coverage.json file"""
    try:
        data = json.loads(coverage_file.read_text())

        total_statements = data["totals"]["num_statements"]
        covered_statements = data["totals"]["covered_lines"]
        coverage_percentage = data["totals"]["percent_covered"]
        branch_coverage = data["totals"].get("percent_covered_display")

        missing_lines = {}
        for file_path, file_data in data["files"].items():
            if file_data["missing_lines"]:
                missing_lines[file_path] = file_data["missing_lines"]

        return CoverageResult(
            total_statements=total_statements,
            covered_statements=covered_statements,
            coverage_percentage=coverage_percentage,
            missing_lines=missing_lines,
            branch_coverage=float(branch_coverage) if branch_coverage else None
        )
    except Exception as error:
        raise RuntimeError(f"Failed to parse coverage JSON: {error}")


def parse_text_coverage(stdout: str) -> CoverageResult:
    """Parse coverage from text output"""
    # Extract coverage percentage from output
    coverage_match = re.search(r"TOTAL.*?(\d+)%", stdout)
    if coverage_match:
        coverage_percentage = float(coverage_match.group(1))
    else:
        coverage_percentage = 0.0

    # Extract statement counts (simplified)
    statements_match = re.search(r"(\d+)\s+statements", stdout)
    total_statements = int(statements_match.group(1)) if statements_match else 0

    covered_statements = int(total_statements * coverage_percentage / 100) if total_statements > 0 else 0

    return CoverageResult(
        total_statements=total_statements,
        covered_statements=covered_statements,
        coverage_percentage=coverage_percentage,
        missing_lines={},
        branch_coverage=None
    )


def parse_file_coverage(file_data: Dict) -> FileCoverage:
    """Parse individual file coverage data"""
    return FileCoverage(
        file_path=file_data.get("file_path", ""),
        statements=file_data.get("num_statements", 0),
        missing=len(file_data.get("missing_lines", [])),
        coverage=file_data.get("percent_covered", 0.0),
        missing_lines=file_data.get("missing_lines", [])
    )