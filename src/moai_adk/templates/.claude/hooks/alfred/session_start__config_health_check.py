#!/usr/bin/env python3
# SessionStart Hook: Configuration Health Check
"""SessionStart Hook: Configuration Health Check

Claude Code Event: SessionStart
Purpose: Automatically detect and propose configuration updates
Execution: Triggered automatically when Claude Code session begins

Features:
- Check if .moai/config.json exists
- Verify configuration completeness
- Detect stale configurations (older than 30 days)
- Suggest updates via interactive prompt
- Propose re-running /alfred:0-project if necessary
"""

import json
import subprocess
import sys
import time
from pathlib import Path
from typing import Any, Optional


def check_config_exists() -> bool:
    """Check if .moai/config.json exists"""
    config_path = Path.cwd() / ".moai" / "config.json"
    return config_path.exists()


def get_config_data() -> Optional[dict[str, Any]]:
    """Read and parse .moai/config.json"""
    try:
        config_path = Path.cwd() / ".moai" / "config.json"
        if not config_path.exists():
            return None
        return json.loads(config_path.read_text())
    except Exception:
        return None


def get_config_age() -> Optional[int]:
    """Get configuration file age in days"""
    try:
        config_path = Path.cwd() / ".moai" / "config.json"
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


def check_moai_version_match() -> tuple[bool, Optional[str], Optional[str]]:
    """Check if .moai/config.json version matches installed moai-adk version

    Returns:
        (is_matched, config_version, installed_version)
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
                    if part[0].isdigit():
                        installed_version = part
                        break

        # Get config version
        config = get_config_data()
        if not config:
            return False, None, installed_version

        config_version = config.get("moai", {}).get("version")

        if installed_version and config_version:
            is_matched = installed_version == config_version
            return is_matched, config_version, installed_version

        return False, config_version, installed_version

    except Exception:
        return False, None, None


def generate_config_report() -> str:
    """Generate configuration health check report"""
    report_lines = []

    # Check 1: Configuration exists
    if not check_config_exists():
        report_lines.append("❌ 프로젝트 설정 없음 - /alfred:0-project을(를) 실행해야 합니다")
        return "\n".join(report_lines)

    config = get_config_data()

    # Check 2: Configuration completeness
    is_complete, missing_fields = check_config_completeness(config or {})
    if not is_complete:
        report_lines.append(f"⚠️  설정 누락: {', '.join(missing_fields)}")
    else:
        report_lines.append("✅ 설정 완성됨")

    # Check 3: Configuration age
    config_age = get_config_age()
    if config_age is not None:
        if config_age > 30:
            report_lines.append(f"⏰ 설정 오래됨: {config_age}일 전 (업데이트 권장)")
        elif config_age > 7:
            report_lines.append(f"⏰ 설정 업데이트: {config_age}일 전")
        else:
            report_lines.append(f"✅ 최근 설정: {config_age}일 전")

    # Check 4: Version match
    is_matched, config_version, installed_version = check_moai_version_match()
    if installed_version and config_version:
        if is_matched:
            report_lines.append(f"✅ 버전 일치: {installed_version}")
        else:
            report_lines.append(
                f"⚠️  버전 불일치: 설정 {config_version} vs 설치됨 {installed_version} "
                f"- /alfred:0-project 실행 권장"
            )

    return "\n".join(report_lines)


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
    is_matched, _, _ = check_moai_version_match()
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
        data = json.loads(input_data) if input_data.strip() else {}

        # Generate configuration report
        config_report = generate_config_report()

        # Check if we should suggest update
        should_update = should_suggest_update()

        # Build system message with health check report
        system_message = f"📋 Configuration Health Check:\n{config_report}"

        if should_update:
            system_message += "\n\n⚠️  Configuration issues detected. Please take action."

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
