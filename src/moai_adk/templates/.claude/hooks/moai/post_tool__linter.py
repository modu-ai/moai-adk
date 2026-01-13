#!/usr/bin/env python3
"""PostToolUse Hook: Automatic Linting

Claude Code Event: PostToolUse
Matcher: Write|Edit
Purpose: Automatically run linters after Claude writes or edits files

Provides feedback to Claude about code quality issues:
- Python: ruff check, mypy
- JavaScript/TypeScript: eslint, biome lint
- Go: golangci-lint
- Rust: clippy
- And many more...

Exit Codes:
- 0: Success (linting completed, issues reported as context)
- 2: Critical lint errors (Claude should address immediately)

Output:
- JSON with additionalContext containing lint issues
- Exit code 2 triggers Claude to review and fix issues
"""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

# Try importing tool_registry, with fallback
try:
    import tool_registry  # noqa: F401

    TOOL_REGISTRY_AVAILABLE = True
except ImportError:
    TOOL_REGISTRY_AVAILABLE = False

    # Fallback definitions
    class ToolType:
        LINTER = "linter"

    class ToolResult:
        def __init__(self, **kwargs):
            for k, v in kwargs.items():
                setattr(self, k, v)


# Maximum number of issues to report to Claude
MAX_ISSUES_TO_REPORT = 5

# File extensions that should skip linting
SKIP_EXTENSIONS = {
    ".json",
    ".lock",
    ".min.js",
    ".min.css",
    ".map",
    ".svg",
    ".png",
    ".jpg",
    ".gif",
    ".ico",
    ".woff",
    ".woff2",
    ".ttf",
    ".eot",
    ".md",  # Markdown usually doesn't need linting for code generation
}

# Directories to skip
SKIP_DIRECTORIES = {
    "node_modules",
    ".git",
    ".venv",
    "venv",
    "__pycache__",
    ".cache",
    "dist",
    "build",
    ".next",
    ".nuxt",
    "target",
    "vendor",
}


def should_skip_file(file_path: str) -> tuple[bool, str]:
    """Check if file should be skipped for linting.

    Args:
        file_path: Path to the file

    Returns:
        Tuple of (should_skip, reason)
    """
    path = Path(file_path)

    # Check extension
    if path.suffix.lower() in SKIP_EXTENSIONS:
        return True, f"Skipped: {path.suffix} files are not linted"

    # Check for minified files
    if ".min." in path.name:
        return True, "Skipped: minified file"

    # Check if in skip directory
    for parent in path.parents:
        if parent.name in SKIP_DIRECTORIES:
            return True, f"Skipped: file in {parent.name}/ directory"

    # Check if file exists
    if not path.exists():
        return True, "Skipped: file does not exist"

    return False, ""


def parse_lint_output(output: str, error: str) -> list[dict[str, Any]]:
    """Parse linter output to extract issues.

    Args:
        output: stdout from linter
        error: stderr from linter

    Returns:
        List of parsed issues
    """
    issues: list[dict[str, Any]] = []
    combined = (output + "\n" + error).strip()

    if not combined:
        return issues

    # Parse line by line
    for line in combined.split("\n"):
        line = line.strip()
        if not line:
            continue

        # Skip non-issue lines
        if any(
            skip in line.lower()
            for skip in [
                "warning:",
                "info:",
                "running",
                "checking",
                "finished",
                "success",
                "✓",
                "✔",
            ]
        ):
            continue

        # Look for error patterns
        if "error" in line.lower() or ":" in line:
            issues.append({"message": line[:200]})  # Truncate long messages

        if len(issues) >= MAX_ISSUES_TO_REPORT:
            break

    return issues


def lint_file(file_path: str) -> dict[str, Any]:
    """Lint a single file using the appropriate linter.

    Args:
        file_path: Path to the file to lint

    Returns:
        Dictionary with linting results
    """
    results: dict[str, Any] = {
        "file": file_path,
        "linted": False,
        "issues": [],
        "issues_fixed": 0,
        "tools_run": [],
        "errors": [],
    }

    # Check if we should skip
    should_skip, skip_reason = should_skip_file(file_path)
    if should_skip:
        results["skipped"] = True
        results["skip_reason"] = skip_reason
        return results

    # Get tool registry
    if not TOOL_REGISTRY_AVAILABLE:
        results["errors"].append("Tool registry not available")
        return results

    registry = tool_registry.get_tool_registry()
    if not registry:
        results["errors"].append("Tool registry not available")
        return results

    # Get linters for this file
    linters = registry.get_tools_for_file(file_path, ToolType.LINTER)
    if not linters:
        language = registry.get_language_for_file(file_path)
        if language:
            results["skipped"] = True
            results["skip_reason"] = f"No linter available for {language}"
        else:
            results["skipped"] = True
            results["skip_reason"] = "Unknown file type"
        return results

    # Run linters (use first available)
    for linter in linters[:1]:  # Only run the highest priority linter
        result = registry.run_tool(linter, file_path)
        results["tools_run"].append(
            {
                "name": result.tool_name,
                "success": result.success,
                "fixed": result.file_modified,
            }
        )

        results["linted"] = True

        if result.file_modified:
            results["issues_fixed"] += 1

        # Parse issues from output
        if result.output or result.error:
            parsed_issues = parse_lint_output(result.output, result.error)
            results["issues"].extend(parsed_issues)

        break  # Only use first linter

    return results


def main() -> None:
    """Main entry point for PostToolUse linter hook.

    Reads JSON input from stdin, lints the file if applicable,
    and outputs JSON result for Claude.
    """
    try:
        # Read JSON input from stdin
        input_data = json.load(sys.stdin)
    except json.JSONDecodeError:
        # Invalid JSON input - exit silently
        sys.exit(0)

    # Extract tool information
    tool_name = input_data.get("tool_name", "")
    tool_input = input_data.get("tool_input", {})

    # Only process Write and Edit tools
    if tool_name not in ("Write", "Edit"):
        sys.exit(0)

    # Get file path from tool input
    file_path = tool_input.get("file_path", "")
    if not file_path:
        sys.exit(0)

    # Lint the file
    result = lint_file(file_path)

    # Build output
    output: dict[str, Any] = {}

    if result.get("issues"):
        # Issues found - provide context to Claude
        issue_count = len(result["issues"])
        issue_summary = "; ".join([i["message"] for i in result["issues"][:3]])

        if issue_count > 3:
            issue_summary += f" (+{issue_count - 3} more)"

        # If issues were auto-fixed, just inform
        if result.get("issues_fixed", 0) > 0:
            output = {
                "hookSpecificOutput": {
                    "hookEventName": "PostToolUse",
                    "additionalContext": f"Lint: {result['issues_fixed']} issues auto-fixed",
                }
            }
            print(json.dumps(output))
            sys.exit(0)
        else:
            # Issues need attention - use exit code 2 to alert Claude
            print(f"Lint issues found: {issue_summary}", file=sys.stderr)
            sys.exit(2)

    elif result.get("linted"):
        # Linting passed with no issues
        output = {"suppressOutput": True}
        print(json.dumps(output))
        sys.exit(0)

    else:
        # Skipped or no linter available
        output = {"suppressOutput": True}
        print(json.dumps(output))
        sys.exit(0)


if __name__ == "__main__":
    main()
