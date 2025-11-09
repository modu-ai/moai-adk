#!/usr/bin/env python3
# @CODE:HOOK-PRE-TAG-001 | SPEC: TAG-PRE-HOOK-001 | TEST: tests/hooks/test_pre_tool_tag_validator.py
"""PreToolUse Hook: Real-time TAG Policy Violation Detection

Detects and blocks TAG policy violations before Edit/Write/MultiEdit execution.
Enforces SPEC-first principle to guarantee code quality.

Features:
- TAG policy validation before file creation
- Blocks CODE creation without SPEC
- Real-time violation reporting and fix guidance
- Work blocking or warning provision

Usage:
    python3 pre_tool__tag_policy_validator.py <tool_name> <tool_args_json>
"""

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List

# Add module path
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.policy_validator import (
    PolicyValidationConfig,
    PolicyViolation,
    PolicyViolationLevel,
    TagPolicyValidator,
)

from ..utils.hook_config import get_graceful_degradation, load_hook_timeout


def load_config() -> Dict[str, Any]:
    """Load configuration file

    Returns:
        Configuration dictionary
    """
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
    except Exception:
        pass

    return {}


def create_policy_validator() -> TagPolicyValidator:
    """Create TAG policy validator

    Returns:
        TagPolicyValidator instance
    """
    config_data = load_config()
    tag_policy_config = config_data.get("tags", {}).get("policy", {})

    # Create PolicyValidationConfig
    policy_config = PolicyValidationConfig(
        strict_mode=tag_policy_config.get("enforcement_mode", "strict") == "strict",
        require_spec_before_code=tag_policy_config.get("require_spec_before_code", True),
        require_test_for_code=tag_policy_config.get("require_test_for_code", True),
        allow_duplicate_tags=not tag_policy_config.get("enforce_chains", True),
        validation_timeout=tag_policy_config.get("realtime_validation", {}).get("validation_timeout", 5),
        auto_fix_enabled=tag_policy_config.get("auto_correction", {}).get("enabled", False)
    )

    return TagPolicyValidator(config=policy_config)


def should_validate_tool(tool_name: str, tool_args: Dict[str, Any]) -> bool:
    """Check if tool is a validation target

    Args:
        tool_name: Tool name
        tool_args: Tool arguments

    Returns:
        True if validation target
    """
    # Validate only file manipulation tools
    validation_tools = {"Edit", "Write", "MultiEdit"}
    if tool_name not in validation_tools:
        return False

    # Optional file patterns (not a TAG validation target)
    optional_patterns = [
        "CLAUDE.md",
        "README.md",
        "CHANGELOG.md",
        "CONTRIBUTING.md",
        ".claude/",
        ".moai/docs/",
        ".moai/reports/",
        ".moai/analysis/",
        "docs/",
        "templates/",
        "examples/",
    ]

    # For Edit/Write, check single file
    if tool_name in {"Edit", "Write"}:
        file_path = tool_args.get("file_path", "")
        if any(pattern in file_path for pattern in optional_patterns):
            return False
        return True

    # For MultiEdit, check multiple files
    if tool_name == "MultiEdit":
        edits = tool_args.get("edits", [])
        for edit in edits:
            file_path = edit.get("file_path", "")
            if any(pattern in file_path for pattern in optional_patterns):
                return False

    return True


def extract_file_paths(tool_name: str, tool_args: Dict[str, Any]) -> List[str]:
    """Extract file paths from tool arguments

    Args:
        tool_name: Tool name
        tool_args: Tool arguments

    Returns:
        List of file paths
    """
    file_paths = []

    if tool_name in {"Edit", "Write"}:
        file_path = tool_args.get("file_path", "")
        if file_path:
            file_paths.append(file_path)

    elif tool_name == "MultiEdit":
        # For MultiEdit, extract multiple file paths
        edits = tool_args.get("edits", [])
        for edit in edits:
            file_path = edit.get("file_path", "")
            if file_path:
                file_paths.append(file_path)

    return file_paths


def get_file_content(tool_name: str, tool_args: Dict[str, Any], file_path: str) -> str:
    """Get file content

    Args:
        tool_name: Tool name
        tool_args: Tool arguments
        file_path: File path

    Returns:
        파일 내용
    """
    # Write: new content
    if tool_name == "Write":
        return tool_args.get("content", "")

    # Edit/MultiEdit: apply modifications to existing content
    try:
        path = Path(file_path)
        if path.exists():
            return path.read_text(encoding="utf-8", errors="ignore")
    except Exception:
        pass

    return ""


def create_block_response(violations: List[PolicyViolation]) -> Dict[str, Any]:
    """Create work blocking response

    Args:
        violations: List of policy violations

    Returns:
        Blocking response dictionary
    """
    critical_violations = [v for v in violations if v.level == PolicyViolationLevel.CRITICAL]
    blocking_violations = [v for v in violations if v.should_block_operation()]

    response = {
        "block_execution": True,
        "reason": "TAG policy violation",
        "violations": [v.to_dict() for v in blocking_violations],
        "message": "❌ Work blocked due to TAG policy violation",
        "guidance": []
    }

    if critical_violations:
        response["message"] = "🚨 치명적인 TAG 정책 위반입니다. 작업을 진행할 수 없습니다."
        response["critical_violations"] = [v.to_dict() for v in critical_violations]

    # Add fix guidance
    for violation in blocking_violations:
        if violation.guidance:
            response["guidance"].append(f"• {violation.guidance}")

    return response


def create_warning_response(violations: List[PolicyViolation]) -> Dict[str, Any]:
    """Create warning response

    Args:
        violations: List of policy violations

    Returns:
        경고 응답 딕셔너리
    """
    response = {
        "block_execution": False,
        "reason": "TAG policy warning",
        "violations": [v.to_dict() for v in violations],
        "message": "⚠️ TAG policy warnings detected but work can proceed",
        "guidance": []
    }

    # Add fix guidance
    for violation in violations:
        if violation.guidance:
            response["guidance"].append(f"• {violation.guidance}")

    return response


def create_success_response() -> Dict[str, Any]:
    """Create success response

    Returns:
        Success response dictionary
    """
    return {
        "block_execution": False,
        "reason": "TAG policy compliant",
        "violations": [],
        "message": "✅ TAG policy validation passed",
        "guidance": []
    }


def main() -> None:
    """Main function"""
    try:
        # Load timeout value from config (milliseconds → seconds)
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # Parse arguments
        if len(sys.argv) < 3:
            print(json.dumps({
                "block_execution": False,
                "error": "Invalid arguments. Usage: python3 pre_tool__tag_policy_validator.py <tool_name> <tool_args_json>"
            }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
        except json.JSONDecodeError:
            print(json.dumps({
                "block_execution": False,
                "error": "Invalid tool_args JSON"
            }))
            sys.exit(0)

        # Record start time (for timeout check)
        start_time = time.time()

        # Check if tool should be validated
        if not should_validate_tool(tool_name, tool_args):
            print(json.dumps(create_success_response()))
            sys.exit(0)

        # Extract file paths
        file_paths = extract_file_paths(tool_name, tool_args)
        if not file_paths:
            print(json.dumps(create_success_response()))
            sys.exit(0)

        # Create policy validator
        validator = create_policy_validator()

        # Validate all files
        all_violations = []
        for file_path in file_paths:
            # Timeout check
            if time.time() - start_time > timeout_seconds:
                break

            # Get file content
            content = get_file_content(tool_name, tool_args, file_path)

            # Validate policy
            violations = validator.validate_before_creation(file_path, content)
            all_violations.extend(violations)

        # Classify violations by level
        blocking_violations = [v for v in all_violations if v.should_block_operation()]
        warning_violations = [v for v in all_violations if not v.should_block_operation()]

        # Create response
        if blocking_violations:
            response = create_block_response(blocking_violations)
        elif warning_violations:
            response = create_warning_response(warning_violations)
        else:
            response = create_success_response()

        # Add validation report
        if all_violations:
            validation_report = validator.create_validation_report(all_violations)
            response["validation_report"] = validation_report

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # On exception, do not block but only log
        error_response = {
            "block_execution": False,
            "error": f"Hook execution error: {str(e)}",
            "message": "Hook execution encountered error but work will proceed."
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True
            error_response["message"] = "Hook failed but continuing due to graceful degradation"

        print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()
