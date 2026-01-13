#!/usr/bin/env python3
"""PostToolUse Hook: Automatic AST-Grep Security Scan

Automatically runs AST-Grep security scanning after Write/Edit operations
to detect potential security vulnerabilities in real-time.

Hook Protocol:
- Input: JSON from stdin with tool_name, tool_input
- Output: JSON to stdout with hookSpecificOutput
- Exit codes: 0 = success, 2 = attention needed
"""

import json
import logging
import os
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Any, Optional

# Configure logger for AST-Grep scanner (H4: structured logging)
logger = logging.getLogger(__name__)

# Supported extensions for AST-Grep scanning
SUPPORTED_EXTENSIONS = {
    ".py",
    ".pyi",  # Python
    ".js",
    ".jsx",
    ".mjs",
    ".cjs",  # JavaScript
    ".ts",
    ".tsx",
    ".mts",
    ".cts",  # TypeScript
    ".go",  # Go
    ".rs",  # Rust
    ".java",  # Java
    ".kt",
    ".kts",  # Kotlin
    ".c",
    ".cpp",
    ".cc",
    ".h",
    ".hpp",  # C/C++
    ".rb",  # Ruby
    ".php",  # PHP
}

# Environment variable to disable AST-Grep scanning
DISABLE_ENV_VAR = "MOAI_DISABLE_AST_GREP_SCAN"


def get_project_dir() -> Path:
    """Get the project directory from environment or current working directory."""
    project_dir = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    return Path(project_dir)


def get_rules_config_path() -> Optional[Path]:
    """Find the AST-Grep rules configuration file."""
    project_dir = get_project_dir()

    # Check common locations for sgconfig.yml
    possible_paths = [
        project_dir / ".claude" / "skills" / "moai-tool-ast-grep" / "rules" / "sgconfig.yml",
        project_dir / "sgconfig.yml",
        project_dir / ".ast-grep" / "sgconfig.yml",
    ]

    for path in possible_paths:
        if path.exists():
            return path

    return None


def is_scannable_file(file_path: str) -> bool:
    """Check if the file is supported for AST-Grep scanning."""
    if not file_path:
        return False

    ext = Path(file_path).suffix.lower()
    return ext in SUPPORTED_EXTENSIONS


def run_ast_grep_scan(file_path: str, config_path: Optional[Path] = None) -> dict[str, Any]:
    """Run AST-Grep scan on a file and return results."""
    logger.debug(f"Starting AST-Grep scan for: {file_path}")

    result: dict[str, Any] = {
        "scanned": False,
        "issues_found": 0,
        "error_count": 0,
        "warning_count": 0,
        "info_count": 0,
        "details": [],
        "error": None,
    }

    # Check if sg (ast-grep) is available
    if not shutil.which("sg"):
        result["error"] = "ast-grep (sg) not installed"
        logger.warning("AST-Grep (sg) not installed, skipping scan")
        return result

    try:
        # Build command
        cmd = ["sg", "scan", "--json"]

        if config_path and config_path.exists():
            cmd.extend(["--config", str(config_path)])
            logger.debug(f"Using config: {config_path}")

        cmd.append(file_path)

        # Run scan with timeout
        logger.debug(f"Running AST-Grep command: {' '.join(cmd)}")
        proc = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=30,
            cwd=str(get_project_dir()),
        )

        result["scanned"] = True

        # Parse JSON output
        if proc.stdout.strip():
            try:
                findings = json.loads(proc.stdout)
                if isinstance(findings, list):
                    for finding in findings:
                        severity = finding.get("severity", "info").lower()
                        if severity == "error":
                            result["error_count"] += 1
                        elif severity == "warning":
                            result["warning_count"] += 1
                        else:
                            result["info_count"] += 1

                        result["details"].append(
                            {
                                "rule": finding.get("ruleId", "unknown"),
                                "severity": severity,
                                "message": finding.get("message", ""),
                                "line": finding.get("range", {}).get("start", {}).get("line", 0),
                            }
                        )

                    result["issues_found"] = len(findings)
                    if result["issues_found"] > 0:
                        logger.info(
                            f"AST-Grep found {result['issues_found']} issues "
                            f"({result['error_count']} errors, {result['warning_count']} warnings)"
                        )
                    else:
                        logger.debug("AST-Grep scan completed: no issues found")
            except json.JSONDecodeError:
                # Non-JSON output (possibly no issues found)
                logger.debug("AST-Grep returned non-JSON output")
                pass

    except subprocess.TimeoutExpired:
        result["error"] = "AST-Grep scan timed out"
        logger.warning(f"AST-Grep scan timed out for: {file_path}")
    except FileNotFoundError:
        result["error"] = "ast-grep (sg) command not found"
        logger.error("ast-grep (sg) command not found")
    except Exception as e:
        result["error"] = str(e)
        logger.error(f"AST-Grep scan error: {e}", exc_info=True)

    return result


def format_scan_result(result: dict[str, Any], file_path: str) -> str:
    """Format scan result for Claude's additional context."""
    if result.get("error"):
        return f"AST-Grep scan skipped: {result['error']}"

    if not result.get("scanned"):
        return "AST-Grep scan not performed"

    if result["issues_found"] == 0:
        return f"AST-Grep: No security issues found in {Path(file_path).name}"

    # Build summary
    parts = []

    if result["error_count"] > 0:
        parts.append(f"{result['error_count']} error(s)")
    if result["warning_count"] > 0:
        parts.append(f"{result['warning_count']} warning(s)")
    if result["info_count"] > 0:
        parts.append(f"{result['info_count']} info")

    summary = f"AST-Grep found {', '.join(parts)} in {Path(file_path).name}"

    # Add top 3 issues
    if result["details"]:
        issues = []
        for detail in result["details"][:3]:
            severity = detail["severity"].upper()
            rule = detail["rule"]
            message = detail["message"]
            line_num = detail["line"]
            issues.append(f"  - [{severity}] {rule}: {message} (line {line_num})")
        summary += "\n" + "\n".join(issues)

        if len(result["details"]) > 3:
            summary += f"\n  ... and {len(result['details']) - 3} more"

    return summary


def main() -> None:
    """Main hook entry point."""
    # Check if scanning is disabled
    if os.environ.get(DISABLE_ENV_VAR, "").lower() in ("1", "true", "yes"):
        sys.exit(0)

    # Read input from stdin
    try:
        input_data = json.load(sys.stdin)
    except (json.JSONDecodeError, OSError):
        sys.exit(0)

    # Extract file path from tool input
    tool_name = input_data.get("tool_name", "")
    tool_input = input_data.get("tool_input", {})
    file_path = tool_input.get("file_path", "")

    # Only process Write and Edit tools
    if tool_name not in ("Write", "Edit"):
        sys.exit(0)

    # Check if file is scannable
    if not is_scannable_file(file_path):
        sys.exit(0)

    # Check if file exists
    if not Path(file_path).exists():
        sys.exit(0)

    # Find rules configuration
    config_path = get_rules_config_path()

    # Run scan
    result = run_ast_grep_scan(file_path, config_path)

    # Format output
    context = format_scan_result(result, file_path)

    # Prepare hook output
    output = {
        "hookSpecificOutput": {
            "hookEventName": "PostToolUse",
            "additionalContext": context,
        }
    }

    # Output JSON
    print(json.dumps(output))

    # Exit with attention code if errors found
    if result.get("error_count", 0) > 0:
        sys.exit(2)

    sys.exit(0)


if __name__ == "__main__":
    main()
