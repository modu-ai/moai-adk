#!/usr/bin/env python3
"""PostToolUse Hook: Automatic Code Formatting

Claude Code Event: PostToolUse
Matcher: Write|Edit
Purpose: Automatically format code after Claude writes or edits files

Supports 16+ languages with automatic tool detection:
- Python: ruff format, black, isort
- JavaScript/TypeScript: biome, prettier, eslint
- Go: gofmt, goimports
- Rust: rustfmt
- And many more...

Exit Codes:
- 0: Success (formatting applied or skipped gracefully)
- 2: Error (stderr shown to Claude for context)

Output:
- JSON with additionalContext for Claude feedback
- suppressOutput: true when nothing to report
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

try:
    from lib.tool_registry import ToolResult, ToolType, get_tool_registry
except ImportError:
    # Fallback if module not available
    def get_tool_registry():
        return None

    class ToolType:
        FORMATTER = "formatter"

    class ToolResult:
        def __init__(self, **kwargs):
            for k, v in kwargs.items():
                setattr(self, k, v)


# File extensions that should skip formatting
SKIP_EXTENSIONS = {
    ".json",  # JSON formatting can break configs
    ".lock",  # Lock files should not be modified
    ".min.js",  # Minified files
    ".min.css",
    ".map",  # Source maps
    ".svg",  # SVG files
    ".png",
    ".jpg",
    ".gif",  # Binary files
    ".ico",
    ".woff",
    ".woff2",
    ".ttf",
    ".eot",
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
    "target",  # Rust/Java
    "vendor",  # PHP/Go
}


def should_skip_file(file_path: str) -> tuple[bool, str]:
    """Check if file should be skipped for formatting.

    Args:
        file_path: Path to the file

    Returns:
        Tuple of (should_skip, reason)
    """
    path = Path(file_path)

    # Check extension
    if path.suffix.lower() in SKIP_EXTENSIONS:
        return True, f"Skipped: {path.suffix} files are not formatted"

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

    # Check if file is binary
    try:
        with open(path, "rb") as f:
            chunk = f.read(8192)
            if b"\x00" in chunk:  # Binary file indicator
                return True, "Skipped: binary file"
    except (OSError, PermissionError):
        return True, "Skipped: cannot read file"

    return False, ""


def format_file(file_path: str) -> dict[str, Any]:
    """Format a single file using the appropriate formatter.

    Args:
        file_path: Path to the file to format

    Returns:
        Dictionary with formatting results
    """
    results: dict[str, Any] = {
        "file": file_path,
        "formatted": False,
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
    registry = get_tool_registry()
    if not registry:
        results["errors"].append("Tool registry not available")
        return results

    # Get formatters for this file
    formatters = registry.get_tools_for_file(file_path, ToolType.FORMATTER)
    if not formatters:
        language = registry.get_language_for_file(file_path)
        if language:
            results["skipped"] = True
            results["skip_reason"] = f"No formatter available for {language}"
        else:
            results["skipped"] = True
            results["skip_reason"] = "Unknown file type"
        return results

    # Run formatters (use first available)
    for formatter in formatters[:1]:  # Only run the highest priority formatter
        result = registry.run_tool(formatter, file_path)
        results["tools_run"].append(
            {
                "name": result.tool_name,
                "success": result.success,
                "modified": result.file_modified,
            }
        )

        if result.success:
            results["formatted"] = result.file_modified
            if result.file_modified:
                results["message"] = f"Formatted with {result.tool_name}"
            break
        else:
            results["errors"].append(f"{result.tool_name}: {result.error}")

    return results


def main() -> None:
    """Main entry point for PostToolUse formatter hook.

    Reads JSON input from stdin, formats the file if applicable,
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
    input_data.get("tool_response", {})

    # Only process Write and Edit tools
    if tool_name not in ("Write", "Edit"):
        sys.exit(0)

    # Get file path from tool input
    file_path = tool_input.get("file_path", "")
    if not file_path:
        sys.exit(0)

    # Format the file
    result = format_file(file_path)

    # Build output
    output: dict[str, Any] = {}

    if result.get("formatted"):
        # File was formatted - provide context to Claude
        output = {
            "hookSpecificOutput": {
                "hookEventName": "PostToolUse",
                "additionalContext": f"Auto-formatted: {result.get('message', 'File formatted')}",
            }
        }
        print(json.dumps(output))
        sys.exit(0)

    elif result.get("errors"):
        # Formatting failed - let Claude know
        error_msg = "; ".join(result["errors"])
        output = {
            "hookSpecificOutput": {
                "hookEventName": "PostToolUse",
                "additionalContext": f"Format warning: {error_msg}",
            }
        }
        print(json.dumps(output))
        sys.exit(0)

    else:
        # Nothing to report (skipped or no change)
        output = {"suppressOutput": True}
        print(json.dumps(output))
        sys.exit(0)


if __name__ == "__main__":
    main()
