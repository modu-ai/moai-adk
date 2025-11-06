#!/usr/bin/env python3
"""PostToolUse event handlers

PostToolUse event handling for change logging and audit tracking
"""

from core import HookPayload, HookResult


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse event handler with change logging functionality

    Logs tool usage and tracks changes for audit trail:
    - Record tool execution details
    - Track file modifications
    - Log changes to .moai/changes.log
    - Provide audit trail for compliance

    Args:
        payload: Claude Code event payload with tool information

    Returns:
        HookResult with optional system message for significant changes

    @TAG:CHANGE-LOGGING-001
    """
    import json
    import os
    import sys
    from datetime import datetime
    from pathlib import Path

    cwd = payload.get("cwd", ".")
    project_root = Path(cwd)

    # Extract tool information
    tool_name = payload.get("toolName", "unknown")
    tool_parameters = payload.get("toolParameters", {})
    execution_result = payload.get("executionResult", {})

    # Create change log entry
    change_entry = {
        "timestamp": datetime.now().isoformat(),
        "tool": tool_name,
        "parameters": tool_parameters,
        "result": execution_result,
        "files_affected": [],
        "lines_changed": 0,
        "significance": "minor"
    }

    # Analyze tool-specific changes
    try:
        if tool_name in ["Edit", "Write", "MultiEdit"]:
            files_modified = tool_parameters.get("files", [])
            if isinstance(files_modified, list):
                change_entry["files_affected"] = files_modified
                change_entry["lines_changed"] = len(files_modified)

                # Determine significance
                if len(files_modified) > 10:
                    change_entry["significance"] = "major"
                elif len(files_modified) > 3:
                    change_entry["significance"] = "moderate"

                # Check for critical files
                critical_files = ["CLAUDE.md", ".moai/config.json", "pyproject.toml"]
                for file_path in files_modified:
                    if any(critical in file_path for critical in critical_files):
                        change_entry["significance"] = "critical"
                        break

        elif tool_name == "Bash":
            command = tool_parameters.get("command", "")
            if "git" in command:
                change_entry["significance"] = "moderate"
                if any(ops in command for ops in ["commit", "merge", "reset", "push"]):
                    change_entry["significance"] = "major"

    except Exception as e:
        change_entry["error"] = f"Failed to analyze changes: {str(e)}"

    # Save to change log
    try:
        log_dir = project_root / ".moai"
        log_dir.mkdir(exist_ok=True)
        log_file = log_dir / "changes.log"

        # Format log entry
        log_entry = f"[{change_entry['timestamp']}] {tool_name} ({change_entry['significance']})\n"
        log_entry += f"  Files: {len(change_entry['files_affected'])}\n"
        log_entry += f"  Lines: {change_entry['lines_changed']}\n"

        if change_entry["files_affected"]:
            log_entry += f"  Affected: {', '.join(change_entry['files_affected'][:5])}"
            if len(change_entry["files_affected"]) > 5:
                log_entry += f" ... and {len(change_entry['files_affected']) - 5} more"
            log_entry += "\n"

        if "error" in change_entry:
            log_entry += f"  Error: {change_entry['error']}\n"

        log_entry += "\n"

        # Append to log file
        with open(log_file, 'a', encoding='utf-8') as f:
            f.write(log_entry)

        # Keep log file size manageable (last 1000 entries)
        try:
            with open(log_file, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            # Find entry boundaries
            entries = []
            current_entry = []

            for line in lines:
                if line.startswith('[') and current_entry:
                    entries.append(''.join(current_entry))
                    current_entry = [line]
                else:
                    current_entry.append(line)

            if current_entry:
                entries.append(''.join(current_entry))

            # Keep last 1000 entries
            if len(entries) > 1000:
                entries = entries[-1000:]

                # Write back truncated log
                with open(log_file, 'w', encoding='utf-8') as f:
                    f.writelines(''.join(entries))

        except Exception:
            # If log rotation fails, continue (not critical)
            pass

    except Exception as e:
        # Log failure is not critical
        print(f"Failed to log changes: {e}", file=sys.stderr)

    # Return system message only for significant changes
    if change_entry["significance"] in ["critical", "major"]:
        files_count = len(change_entry["files_affected"])
        message = f"📝 Significant change logged: {tool_name} affected {files_count} files ({change_entry['significance']})"
        return HookResult(system_message=message)

    # For minor changes, return empty result
    return HookResult()


__all__ = ["handle_post_tool_use"]