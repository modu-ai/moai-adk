#!/usr/bin/env python3
"""PostToolUse Hook: LSP Diagnostics after Write/Edit operations.

Claude Code Event: PostToolUse
Matcher: Write|Edit
Purpose: Run LSP diagnostics after file modifications and provide feedback to Claude

This hook integrates with the Ralph Engine feedback loop by providing real-time
LSP diagnostic information after each Write or Edit operation.

Exit Codes:
- 0: Success (no errors found or LSP unavailable)
- 2: Attention needed (errors found, Claude should address)

Output:
- JSON with hookSpecificOutput containing diagnostic summary
- Provides Claude with actionable error information
"""

from __future__ import annotations

import json
import os
import sys
from pathlib import Path
from typing import Any

# Environment variable to disable LSP diagnostics
DISABLE_ENV_VAR = "MOAI_DISABLE_LSP_DIAGNOSTIC"

# Supported file extensions for LSP diagnostics
SUPPORTED_EXTENSIONS = {
    # Python
    ".py": "python",
    ".pyi": "python",
    # TypeScript/JavaScript
    ".ts": "typescript",
    ".tsx": "typescriptreact",
    ".js": "javascript",
    ".jsx": "javascriptreact",
    ".mjs": "javascript",
    ".cjs": "javascript",
    ".mts": "typescript",
    ".cts": "typescript",
    # Go
    ".go": "go",
    # Rust
    ".rs": "rust",
    # Other
    ".java": "java",
    ".kt": "kotlin",
    ".kts": "kotlin",
    ".rb": "ruby",
    ".php": "php",
    ".c": "c",
    ".cpp": "cpp",
    ".h": "c",
    ".hpp": "cpp",
}


def get_project_dir() -> Path:
    """Get the project directory from environment or current working directory."""
    project_dir = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    return Path(project_dir)


def is_supported_file(file_path: str) -> bool:
    """Check if the file is supported for LSP diagnostics.

    Args:
        file_path: Path to the file.

    Returns:
        True if the file extension is supported.
    """
    if not file_path:
        return False

    ext = Path(file_path).suffix.lower()
    return ext in SUPPORTED_EXTENSIONS


def get_language_for_file(file_path: str) -> str | None:
    """Get the language identifier for a file.

    Args:
        file_path: Path to the file.

    Returns:
        Language identifier or None.
    """
    ext = Path(file_path).suffix.lower()
    return SUPPORTED_EXTENSIONS.get(ext)


def load_ralph_config() -> dict[str, Any]:
    """Load Ralph configuration from ralph.yaml.

    Returns:
        Configuration dictionary with defaults.
    """
    config: dict[str, Any] = {
        "enabled": True,
        "hooks": {
            "post_tool_lsp": {
                "enabled": True,
                "severity_threshold": "error",
            }
        },
    }

    # Try to load from config file
    config_path = get_project_dir() / ".moai" / "config" / "sections" / "ralph.yaml"
    if config_path.exists():
        try:
            import yaml

            with open(config_path) as f:
                loaded = yaml.safe_load(f)
                if loaded and "ralph" in loaded:
                    # Merge loaded config with defaults
                    ralph = loaded["ralph"]
                    if "enabled" in ralph:
                        config["enabled"] = ralph["enabled"]
                    if "hooks" in ralph and "post_tool_lsp" in ralph["hooks"]:
                        config["hooks"]["post_tool_lsp"].update(ralph["hooks"]["post_tool_lsp"])
        except Exception:
            # Graceful degradation - use defaults
            pass

    return config


def _run_async_safely(coro):
    """Run an async coroutine safely, handling existing event loops.

    This handles the case where an event loop already exists (e.g., in IDE plugins
    or other async contexts) by using the existing loop or creating a new one.

    Args:
        coro: Async coroutine to run.

    Returns:
        Result of the coroutine.

    Raises:
        Exception: Any exception from the coroutine.
    """
    import asyncio

    try:
        # Try to get the current running loop
        loop = asyncio.get_running_loop()
    except RuntimeError:
        # No running loop - safe to use asyncio.run()
        loop = None

    if loop is not None:
        # Event loop already running - use nest_asyncio if available
        # or create a new loop in a thread
        try:
            import nest_asyncio

            nest_asyncio.apply()
            return loop.run_until_complete(coro)
        except ImportError:
            # nest_asyncio not available, run in separate thread
            import concurrent.futures

            with concurrent.futures.ThreadPoolExecutor() as executor:
                future = executor.submit(asyncio.run, coro)
                return future.result(timeout=60)
    else:
        # No existing loop - safe to use asyncio.run()
        return asyncio.run(coro)


async def _get_lsp_diagnostics_async(file_path: str) -> dict[str, Any]:
    """Async implementation of LSP diagnostics retrieval.

    Args:
        file_path: Path to the file.

    Returns:
        Dictionary with diagnostic information.
    """
    result: dict[str, Any] = {
        "available": False,
        "error_count": 0,
        "warning_count": 0,
        "info_count": 0,
        "hint_count": 0,
        "diagnostics": [],
        "error": None,
    }

    try:
        # Try to import and use the LSP client
        project_root = get_project_dir()
        sys.path.insert(0, str(project_root / "src"))

        from moai_adk.lsp.client import MoAILSPClient
        from moai_adk.lsp.models import DiagnosticSeverity

        client = MoAILSPClient(project_root)
        language = client.get_language_for_file(file_path)

        if language is None:
            result["error"] = f"No LSP support for file type: {Path(file_path).suffix}"
            return result

        # Ensure server is running
        await client.ensure_server_running(language)

        # Get diagnostics
        diagnostics = await client.get_diagnostics(file_path)
        result["available"] = True

        # Count by severity
        for diag in diagnostics:
            if diag.severity == DiagnosticSeverity.ERROR:
                result["error_count"] += 1
            elif diag.severity == DiagnosticSeverity.WARNING:
                result["warning_count"] += 1
            elif diag.severity == DiagnosticSeverity.INFORMATION:
                result["info_count"] += 1
            elif diag.severity == DiagnosticSeverity.HINT:
                result["hint_count"] += 1

            result["diagnostics"].append(
                {
                    "severity": diag.severity.name.lower(),
                    "message": diag.message,
                    "line": diag.range.start.line + 1,  # Convert to 1-based
                    "source": diag.source,
                    "code": diag.code,
                }
            )

    except ImportError:
        result["error"] = "LSP client not available"
    except Exception as e:
        result["error"] = f"LSP error: {str(e)}"

    return result


def get_lsp_diagnostics(file_path: str) -> dict[str, Any]:
    """Get LSP diagnostics for a file (sync wrapper).

    This function attempts to get diagnostics from the LSP client.
    If the LSP infrastructure is not available, it gracefully degrades.

    Args:
        file_path: Path to the file.

    Returns:
        Dictionary with diagnostic information.
    """
    try:
        return _run_async_safely(_get_lsp_diagnostics_async(file_path))
    except Exception as e:
        return {
            "available": False,
            "error_count": 0,
            "warning_count": 0,
            "info_count": 0,
            "hint_count": 0,
            "diagnostics": [],
            "error": f"Async execution error: {str(e)}",
        }


def run_fallback_diagnostics(file_path: str) -> dict[str, Any]:
    """Run fallback diagnostics using external tools.

    When LSP is not available, use command-line tools for basic diagnostics.

    Args:
        file_path: Path to the file.

    Returns:
        Dictionary with diagnostic information.
    """
    import shutil
    import subprocess

    result: dict[str, Any] = {
        "available": False,
        "error_count": 0,
        "warning_count": 0,
        "info_count": 0,
        "hint_count": 0,
        "diagnostics": [],
        "error": None,
        "fallback": True,
    }

    language = get_language_for_file(file_path)
    if not language:
        return result

    # Try language-specific linters as fallback
    try:
        if language == "python":
            # Try ruff first, then flake8
            if shutil.which("ruff"):
                proc = subprocess.run(
                    ["ruff", "check", "--output-format=json", file_path],
                    capture_output=True,
                    text=True,
                    timeout=30,
                )
                if proc.stdout.strip():
                    try:
                        issues = json.loads(proc.stdout)
                        result["available"] = True
                        for issue in issues[:10]:  # Limit to 10 issues
                            severity = "error" if issue.get("code", "").startswith("E") else "warning"
                            if severity == "error":
                                result["error_count"] += 1
                            else:
                                result["warning_count"] += 1
                            result["diagnostics"].append(
                                {
                                    "severity": severity,
                                    "message": issue.get("message", ""),
                                    "line": issue.get("location", {}).get("row", 0),
                                    "source": "ruff",
                                    "code": issue.get("code"),
                                }
                            )
                    except json.JSONDecodeError:
                        pass

        elif language in ("typescript", "typescriptreact", "javascript", "javascriptreact"):
            # Try tsc for TypeScript
            if language.startswith("typescript") and shutil.which("tsc"):
                proc = subprocess.run(
                    ["tsc", "--noEmit", "--pretty", "false", file_path],
                    capture_output=True,
                    text=True,
                    timeout=60,
                )
                if proc.stderr or proc.stdout:
                    result["available"] = True
                    output = proc.stderr or proc.stdout
                    for line in output.split("\n")[:10]:
                        if "error" in line.lower():
                            result["error_count"] += 1
                            result["diagnostics"].append(
                                {
                                    "severity": "error",
                                    "message": line.strip(),
                                    "line": 0,
                                    "source": "tsc",
                                    "code": None,
                                }
                            )

    except subprocess.TimeoutExpired:
        result["error"] = "Diagnostic command timed out"
    except Exception as e:
        result["error"] = f"Fallback diagnostic error: {str(e)}"

    return result


def format_diagnostic_output(result: dict[str, Any], file_path: str) -> str:
    """Format diagnostic results for Claude's additional context.

    Args:
        result: Diagnostic results dictionary.
        file_path: Path to the file.

    Returns:
        Formatted string for additionalContext.
    """
    filename = Path(file_path).name

    if result.get("error"):
        # Graceful degradation - just note that diagnostics were skipped
        return f"LSP: Diagnostics unavailable for {filename} ({result['error']})"

    if not result.get("available"):
        return f"LSP: No diagnostics available for {filename}"

    # No issues found
    total = result["error_count"] + result["warning_count"] + result["info_count"] + result["hint_count"]
    if total == 0:
        return f"LSP: No issues in {filename}"

    # Build summary
    parts = []
    if result["error_count"] > 0:
        parts.append(f"{result['error_count']} error(s)")
    if result["warning_count"] > 0:
        parts.append(f"{result['warning_count']} warning(s)")
    if result["info_count"] > 0:
        parts.append(f"{result['info_count']} info")
    if result["hint_count"] > 0:
        parts.append(f"{result['hint_count']} hint(s)")

    summary = f"LSP: {', '.join(parts)} in {filename}"

    # Add top diagnostics (errors first, then warnings)
    diagnostics = result.get("diagnostics", [])

    # Sort by severity (errors first)
    severity_order = {"error": 0, "warning": 1, "information": 2, "info": 2, "hint": 3}
    sorted_diags = sorted(diagnostics, key=lambda d: severity_order.get(d["severity"], 4))

    if sorted_diags:
        issues = []
        for diag in sorted_diags[:5]:  # Show top 5
            sev = diag["severity"].upper()
            msg = diag["message"][:100]  # Truncate long messages
            line = diag.get("line", "?")
            source = diag.get("source", "")
            source_info = f" [{source}]" if source else ""
            issues.append(f"  - [{sev}] Line {line}: {msg}{source_info}")
        summary += "\n" + "\n".join(issues)

        if len(sorted_diags) > 5:
            summary += f"\n  ... and {len(sorted_diags) - 5} more"

    return summary


def main() -> None:
    """Main hook entry point."""
    # Check if diagnostics are disabled
    if os.environ.get(DISABLE_ENV_VAR, "").lower() in ("1", "true", "yes"):
        sys.exit(0)

    # Load configuration
    config = load_ralph_config()

    # Check if Ralph or LSP hook is disabled
    if not config.get("enabled", True):
        sys.exit(0)

    hook_config = config.get("hooks", {}).get("post_tool_lsp", {})
    if not hook_config.get("enabled", True):
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

    # Check if file is supported
    if not is_supported_file(file_path):
        sys.exit(0)

    # Check if file exists
    if not Path(file_path).exists():
        sys.exit(0)

    # Try to get LSP diagnostics (get_lsp_diagnostics is now a sync wrapper)
    result = get_lsp_diagnostics(file_path)

    # If LSP failed and no fallback, try fallback explicitly
    if not result.get("available") and not result.get("fallback"):
        result = run_fallback_diagnostics(file_path)

    # Format output
    context = format_diagnostic_output(result, file_path)

    # Prepare hook output
    output = {
        "hookSpecificOutput": {
            "hookEventName": "PostToolUse",
            "additionalContext": context,
        }
    }

    # Output JSON
    print(json.dumps(output))

    # Determine exit code based on severity threshold
    severity_threshold = hook_config.get("severity_threshold", "error")

    if severity_threshold == "error" and result.get("error_count", 0) > 0:
        sys.exit(2)  # Attention needed
    elif severity_threshold == "warning" and (result.get("error_count", 0) > 0 or result.get("warning_count", 0) > 0):
        sys.exit(2)  # Attention needed
    elif severity_threshold == "info" and (
        result.get("error_count", 0) > 0 or result.get("warning_count", 0) > 0 or result.get("info_count", 0) > 0
    ):
        sys.exit(2)  # Attention needed

    sys.exit(0)


if __name__ == "__main__":
    main()
