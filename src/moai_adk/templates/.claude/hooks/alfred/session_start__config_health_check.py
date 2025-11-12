#!/usr/bin/env python3
# SessionStart Hook: Configuration Health Check
"""SessionStart Hook: Configuration Health Check

Claude Code Event: SessionStart
Purpose: Automatically detect and propose configuration updates
Execution: Triggered automatically when Claude Code session begins

Features:
- Check if .moai/config/config.json exists
- Verify configuration completeness
- Detect stale configurations (older than 30 days)
- Suggest updates via interactive prompt
- Propose re-running /alfred:0-project if necessary
"""

import json
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Optional


def check_config_exists() -> bool:
    """Check if .moai/config/config.json exists"""
    config_path = Path.cwd() / ".moai" / "config" / "config.json"
    return config_path.exists()


def get_config_data() -> Optional[dict[str, Any]]:
    """Read and parse .moai/config/config.json"""
    try:
        config_path = Path.cwd() / ".moai" / "config" / "config.json"
        if not config_path.exists():
            return None
        return json.loads(config_path.read_text())
    except Exception:
        return None


def get_config_age() -> Optional[int]:
    """Get configuration file age in days"""
    try:
        config_path = Path.cwd() / ".moai" / "config" / "config.json"
        if not config_path.exists():
            return None

        current_time = time.time()
        config_time = config_path.stat().st_mtime
        age_seconds = current_time - config_time
        age_days = age_seconds / (24 * 3600)

        return int(age_days)
    except Exception:
        return None


def check_config_completeness(config: dict[str, Any]) -> tuple[bool, list[str]]:
    """Check if configuration has all required fields

    Returns:
        (is_complete, missing_fields)
    """
    required_sections = ["project", "language", "git_strategy", "constitution"]
    missing_fields = []

    for section in required_sections:
        if section not in config:
            missing_fields.append(section)

    # Check critical fields
    if config.get("project", {}).get("name") == "":
        missing_fields.append("project.name (empty)")

    if config.get("language", {}).get("conversation_language") is None:
        missing_fields.append("language.conversation_language")

    return len(missing_fields) == 0, missing_fields


def get_latest_pypi_version() -> Optional[str]:
    """Check latest moai-adk version on PyPI

    Returns:
        Latest version string or None if check fails
    """
    try:
        result = subprocess.run(
            ["pip", "index", "versions", "moai-adk"],
            capture_output=True,
            text=True,
            timeout=5
        )

        if result.returncode == 0:
            # Parse version from pip index output
            for line in result.stdout.split('\n'):
                if 'Available versions:' in line:
                    # Extract first version (latest)
                    parts = line.split(':')
                    if len(parts) > 1:
                        versions = parts[1].strip().split(',')
                        if versions:
                            return versions[0].strip()

        return None
    except Exception:
        return None


def check_moai_version_match() -> tuple[bool, Optional[str], Optional[str], Optional[str]]:
    """Check installed moai-adk version and available updates

    Returns:
        (is_matched, installed_version, latest_version, status_message)
    """
    try:
        # Get installed moai-adk version
        result = subprocess.run(
            ["moai-adk", "--version"],
            capture_output=True,
            text=True,
            timeout=3
        )

        installed_version = None
        if result.returncode == 0:
            # Parse version from output (e.g., "0.21.1")
            version_line = result.stdout.strip()
            if "version" in version_line.lower():
                parts = version_line.split()
                for part in parts:
                    if part and part[0].isdigit():
                        installed_version = part
                        break

        if not installed_version:
            return False, None, None, None

        # Get latest version from PyPI
        latest_version = get_latest_pypi_version()

        if latest_version:
            is_matched = installed_version == latest_version
            if is_matched:
                status = f"✅ Version: {installed_version} (latest)"
            else:
                status = f"⚠️  Version: {installed_version} → {latest_version} update available (run moai-adk update)"
            return is_matched, installed_version, latest_version, status
        else:
            # Can't check PyPI, just show installed version
            return True, installed_version, None, f"✅ Version: {installed_version}"

    except Exception:
        return False, None, None, None


def generate_config_report() -> str:
    """Generate configuration health check report

    Only shows warnings if problems exist:
    - Missing configuration sections
    - Configuration file age > 30 days
    - Version mismatch (only if configuration incomplete)
    """
    report_lines = []

    # Check 1: Configuration exists
    if not check_config_exists():
        report_lines.append("❌ Configuration not found - run /alfred:0-project to initialize")
        return "\n".join(report_lines)

    config = get_config_data()

    # Check 2: Configuration completeness
    is_complete, missing_fields = check_config_completeness(config or {})
    if not is_complete:
        report_lines.append(f"⚠️ Missing configuration: {', '.join(missing_fields)}")

    # Check 3: Configuration age (only warn if > 30 days)
    config_age = get_config_age()
    if config_age is not None and config_age > 30:
        report_lines.append(f"⏰ Configuration outdated: {config_age} days old (update recommended)")

    # Check 4: Version status (only show if configuration is incomplete)
    if not is_complete:
        is_matched, installed_version, latest_version, status_message = check_moai_version_match()
        if status_message:
            report_lines.append(status_message)

    # If no issues found, return empty (will not display health check section)
    return "\n".join(report_lines)


def should_suppress_setup() -> bool:
    """Determine whether setup messages should be suppressed.

    Returns:
        bool: True if setup should be suppressed, False otherwise
    """
    config = get_config_data()
    if not config:
        # No config, don't suppress
        return False

    session_config = config.get("session", {})
    suppress = session_config.get("suppress_setup_messages", False)

    if not suppress:
        # Flag is False, don't suppress
        return False

    # Flag is True, check time threshold (7 days)
    suppressed_at_str = session_config.get("setup_messages_suppressed_at")
    if not suppressed_at_str:
        # No timestamp, suppress
        return True

    try:
        suppressed_at = datetime.fromisoformat(suppressed_at_str)
        now = datetime.now(suppressed_at.tzinfo) if suppressed_at.tzinfo else datetime.now()
        days_passed = (now - suppressed_at).days

        # Suppress if less than 7 days have passed
        return days_passed < 7
    except (ValueError, TypeError):
        # If timestamp is invalid, don't suppress (show messages)
        return False


def should_suggest_update() -> bool:
    """Determine if we should suggest configuration update

    Returns True if:
    - Config doesn't exist
    - Config is incomplete
    - Config is older than 30 days
    - Version mismatch detected
    """
    if not check_config_exists():
        return True

    config = get_config_data()
    if not config:
        return True

    # Check completeness
    is_complete, _ = check_config_completeness(config)
    if not is_complete:
        return True

    # Check age (suggest if > 30 days)
    config_age = get_config_age()
    if config_age and config_age > 30:
        return True

    # Check version match
    is_matched, _, _, _ = check_moai_version_match()
    if not is_matched:
        return True

    return False


def main() -> None:
    """Main entry point for configuration health check hook

    Displays configuration status and suggests updates if needed.
    If configuration issues detected, prompts user for action via AskUserQuestion.

    Exit Codes:
        0: Success
        1: Error
    """
    try:
        # Read JSON payload from stdin (for compatibility)
        input_data = sys.stdin.read()
        _data = json.loads(input_data) if input_data.strip() else {}

        # Check if setup messages should be suppressed
        is_suppressed = should_suppress_setup()

        # Generate configuration report
        config_report = generate_config_report()

        # Determine system message
        should_update = should_suggest_update()

        # Build system message based on report content
        if config_report.strip():
            # Report has content, show it with proper formatting
            system_message = f"📋 Configuration Health Check\n{config_report}"

            if should_update:
                system_message += "\n⚠️ Configuration issues detected. Please take action."
        else:
            # No issues found, return empty message (suppresses health check section)
            system_message = ""

        if is_suppressed and not should_update:
            # Suppressed and no issues detected
            system_message = ""

        # Prepare response
        result: dict[str, Any] = {
            "continue": True,
            "systemMessage": system_message
        }

        # If issues detected, add AskUserQuestion to prompt user for action
        if should_update:
            # Build question choices
            question_data = {
                "questions": [
                    {
                        "question": "Configuration issues detected. Select an action to proceed:",
                        "header": "Project Configuration",
                        "multiSelect": False,
                        "options": [
                            {
                                "label": "Initialize Project",
                                "description": "Run /alfred:0-project to initialize new project configuration"
                            },
                            {
                                "label": "Update Settings",
                                "description": "Run /alfred:0-project to update/verify existing configuration"
                            },
                            {
                                "label": "Skip for Now",
                                "description": "Continue without configuration update (not recommended)"
                            }
                        ]
                    }
                ]
            }

            # Add prompt data to result
            result["askUserQuestion"] = question_data

        print(json.dumps(result))
        sys.exit(0)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"ConfigHealthCheck JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error - don't block session
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"ConfigHealthCheck error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"ConfigHealthCheck unexpected error: {e}", file=sys.stderr)
        sys.exit(0)  # Exit 0 to not block session


if __name__ == "__main__":
    main()
