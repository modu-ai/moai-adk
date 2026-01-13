#!/usr/bin/env python3
"""PostToolUse Hook: Universal Test Coverage Guard (Python, TypeScript, JavaScript)

Checks test coverage after code writes for Python, TypeScript, and JavaScript.
Supports multiple testing frameworks and coverage tools.

Exit Codes:
- 0: Success (coverage checked, results reported)
- 2: Coverage below target (get user attention)
"""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path
from typing import Any

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

try:
    from language_detector import (
        LANGUAGE_CONFIGS,
        detect_language,
        get_coverage_target,
        get_test_file_path,
    )
except ImportError:
    # Fallback implementations
    LANGUAGE_CONFIGS = {
        "Python": {
            "extensions": [".py"],
            "test_command": "uv run pytest -q tests/",
            "coverage_command": (
                "uv run coverage run --branch -m pytest -q tests/ && uv run coverage report --format=json"
            ),
        },
        "TypeScript": {
            "extensions": [".ts", ".tsx"],
            "test_command": "npm test -- --run",
            "coverage_command": (
                "npm run test:coverage -- --reporter=json 2>/dev/null || "
                "npm test -- --coverage --coverageReporters=json-summary"
            ),
        },
        "JavaScript": {
            "extensions": [".js", ".jsx"],
            "test_command": "npm test -- --run",
            "coverage_command": "npm run test:coverage -- --coverageReporters=json-summary",
        },
    }

    def detect_language(file_path: Path) -> str | None:
        ext = file_path.suffix.lower()
        for lang, config in LANGUAGE_CONFIGS.items():
            if ext in config.get("extensions", []):
                return lang
        return None

    def get_test_file_path(_source_file: Path) -> Path | None:
        return None

    def get_coverage_target() -> int:
        return 100


class CoverageChecker:
    """Coverage checker for multiple languages."""

    def __init__(self, source_file: Path):
        self.source_file = source_file
        self.language = detect_language(source_file)
        self.config = LANGUAGE_CONFIGS.get(self.language, {}) if self.language else {}

    def check_coverage(self) -> dict[str, Any]:
        """Run language-specific coverage check.

        Returns:
            Dictionary with coverage results
        """
        if not self.language:
            return {"error": "Unsupported language", "skipped": True}

        checkers = {
            "Python": self._check_python_coverage,
            "TypeScript": self._check_typescript_coverage,
            "JavaScript": self._check_javascript_coverage,
        }

        checker = checkers.get(self.language)
        if checker:
            return checker()

        return {"error": f"No coverage checker for {self.language}", "skipped": True}

    def _check_python_coverage(self) -> dict[str, Any]:
        """Check Python test coverage using coverage.py."""
        try:
            # Run pytest with coverage
            subprocess.run(
                ["uv", "run", "coverage", "run", "--branch", "-m", "pytest", "-q", "tests/"],
                capture_output=True,
                timeout=60,
                check=False,
            )

            # Get JSON report
            result = subprocess.run(
                ["uv", "run", "coverage", "report", "--format=json"],
                capture_output=True,
                timeout=30,
                text=True,
            )

            if result.returncode == 0:
                data = json.loads(result.stdout)
                files = data.get("files", [])

                # Find coverage for this specific file
                source_name = self.source_file.name
                for f in files:
                    if source_name in f.get("file", ""):
                        summary = f.get("summary", {})
                        return {
                            "percent_covered": summary.get("percent_covered", 0),
                            "tool": "coverage.py",
                            "lines_covered": summary.get("covered_lines", 0),
                            "lines_missing": summary.get("missing_lines", 0),
                        }

                # No specific file coverage, return total
                totals = data.get("totals", {})
                return {
                    "percent_covered": totals.get("percent_covered", 0),
                    "tool": "coverage.py",
                }

            return {"percent_covered": 0, "tool": "coverage.py"}
        except FileNotFoundError:
            return {"error": "uv or coverage.py not found", "skipped": True}
        except subprocess.TimeoutExpired:
            return {"error": "Coverage check timed out", "skipped": True}
        except Exception as e:
            return {"error": str(e), "skipped": True}

    def _check_typescript_coverage(self) -> dict[str, Any]:
        """Check TypeScript/JavaScript coverage using vitest or jest."""
        try:
            # Check package.json for test script
            package_json = Path("package.json")
            if not package_json.exists():
                return {"error": "package.json not found", "skipped": True}

            # Try vitest first (modern, faster)
            result = subprocess.run(
                ["npm", "run", "test:coverage", "--", "--reporter=json-summary", "--reporter=json"],
                capture_output=True,
                timeout=90,
                text=True,
            )

            if result.returncode == 0:
                # Parse vitest coverage
                return self._parse_coverage_json(result.stdout, "vitest")

            # Fallback to jest
            result = subprocess.run(
                ["npm", "test", "--", "--coverage", "--coverageReporters=json-summary"],
                capture_output=True,
                timeout=90,
                text=True,
            )

            if result.returncode == 0:
                return self._parse_coverage_json(result.stdout, "jest")

            return {"percent_covered": 0, "tool": "vitest/jest", "skipped": True}

        except FileNotFoundError:
            return {"error": "npm not found", "skipped": True}
        except subprocess.TimeoutExpired:
            return {"error": "Coverage check timed out", "skipped": True}
        except Exception as e:
            return {"error": str(e), "skipped": True}

    def _check_javascript_coverage(self) -> dict[str, Any]:
        """JavaScript uses same tools as TypeScript."""
        return self._check_typescript_coverage()

    def _parse_coverage_json(self, output: str, tool: str) -> dict[str, Any]:
        """Parse coverage JSON output from vitest or jest."""
        try:
            data = json.loads(output)

            # Vitest format
            if "total" in data:
                total = data.get("total", {})
                lines = total.get("lines", {})
                return {
                    "percent_covered": lines.get("pct", 0),
                    "tool": tool,
                }

            # Jest format (JSON-summary)
            if "total" in data:
                total = data.get("total", {})
                lines = total.get("lines", {})
                return {
                    "percent_covered": lines.get("pct", 0) * 100 if lines.get("pct") <= 1 else lines.get("pct", 0),
                    "tool": tool,
                }

            return {"percent_covered": 0, "tool": tool}

        except json.JSONDecodeError:
            # Try to extract percentage from text
            import re

            match = re.search(r"Lines\s*:\s*([\d.]+)%", output)
            if match:
                return {"percent_covered": float(match.group(1)), "tool": tool}
            return {"percent_covered": 0, "tool": tool}


def main():
    """Main hook logic."""
    # Read hook input
    try:
        input_data = json.loads(sys.stdin.read())
    except json.JSONDecodeError:
        print(json.dumps({}))
        sys.exit(0)

    tool_name = input_data.get("tool_name", "")
    tool_input = input_data.get("tool_input", {})

    # Only check Write operations
    if tool_name != "Write":
        print(json.dumps({}))
        sys.exit(0)

    file_path = tool_input.get("file_path", "")
    if not file_path:
        print(json.dumps({}))
        sys.exit(0)

    source_file = Path(file_path)

    # Skip test files
    if any(pattern in file_path.lower() for pattern in ["test", "spec", "__mocks__"]):
        print(json.dumps({}))
        sys.exit(0)

    # Check language
    checker = CoverageChecker(source_file)

    if not checker.language:
        print(json.dumps({}))
        sys.exit(0)

    # Only support Python, TypeScript, JavaScript
    if checker.language not in ("Python", "TypeScript", "JavaScript"):
        print(json.dumps({}))
        sys.exit(0)

    # Check if test file exists
    test_file = get_test_file_path(source_file)
    output = {"hookSpecificOutput": {"hookEventName": "PostToolUse"}}

    if not test_file or not Path(test_file).exists():
        output["hookSpecificOutput"]["additionalContext"] = (
            f"TDD ({checker.language}): No test file found.\n"
            f"Expected: {test_file if test_file else 'N/A'}\n"
            f"Create test file to enable coverage checking."
        )
        print(json.dumps(output))
        sys.exit(2)  # Warning but don't block

    # Run coverage check
    coverage = checker.check_coverage()

    if coverage.get("skipped"):
        # Coverage check failed - don't block
        output["hookSpecificOutput"]["additionalContext"] = (
            f"Coverage check unavailable: {coverage.get('error', 'Unknown')}"
        )
        print(json.dumps(output))
        sys.exit(0)

    if "error" in coverage:
        # Error but not skipped - report but don't block
        output["hookSpecificOutput"]["additionalContext"] = f"Coverage check failed: {coverage['error']}"
        print(json.dumps(output))
        sys.exit(0)

    coverage_percent = coverage.get("percent_covered", 0)
    target = get_coverage_target()
    tool = coverage.get("tool", "unknown")

    if coverage_percent < target:
        output["hookSpecificOutput"]["additionalContext"] = (
            f"Coverage: {coverage_percent:.1f}% (target: {target}%)\n"
            f"Tool: {tool}\n"
            f"Language: {checker.language}\n"
            f"File: {source_file.name}\n"
            f"Test: {test_file}"
        )
        print(json.dumps(output))
        sys.exit(2)  # Warn but don't block

    # All good
    output["hookSpecificOutput"]["additionalContext"] = (
        f"Coverage: {coverage_percent:.1f}% ({checker.language}, {tool})"
    )
    print(json.dumps(output))
    sys.exit(0)


if __name__ == "__main__":
    main()
