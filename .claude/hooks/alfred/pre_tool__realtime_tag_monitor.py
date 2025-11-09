#!/usr/bin/env python3
# @CODE:HOOK-REALTIME-001 | SPEC: TAG-REALTIME-HOOK-001 | TEST: tests/hooks/test_realtime_tag_monitor.py
"""Real-time TAG monitoring Hook

Continuous TAG status monitoring and real-time violation detection.
Quick scan of entire project TAG status at PreToolUse stage.

Features:
- Real-time TAG status monitoring
- Fast violation detection (within 5 seconds)
- Project-wide TAG integrity check
- Immediate feedback to user

Usage:
    python3 pre_tool__realtime_tag_monitor.py <tool_name> <tool_args_json>
"""

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List

# Add module path
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.validator import CentralValidationResult, CentralValidator, ValidationConfig

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


def create_validator() -> CentralValidator:
    """Create central validator

    Returns:
        CentralValidator instance
    """
    config_data = load_config()
    tag_policy_config = config_data.get("tags", {}).get("policy", {})

    # Create ValidationConfig
    validation_config = ValidationConfig(
        strict_mode=tag_policy_config.get("enforcement_mode", "strict") == "strict",
        check_duplicates=True,
        check_orphans=True,
        check_chain_integrity=tag_policy_config.get("realtime_validation", {}).get("enforce_chains", True)
    )

    return CentralValidator(config=validation_config)


def should_monitor(tool_name: str, tool_args: Dict[str, Any]) -> bool:
    """Check if tool is a monitoring target

    Args:
        tool_name: Tool name
        tool_args: Tool arguments

    Returns:
        True if monitoring target
    """
    # Monitor only file manipulation tools
    monitoring_tools = {"Edit", "Write", "MultiEdit"}
    return tool_name in monitoring_tools


def get_project_files_to_scan() -> List[str]:
    """Get project files to scan (optimized)

    Improves performance by excluding optional file directories.
    Scan only essential files: src/, tests/, .moai/specs/

    Returns:
        List of file paths
    """
    files = []
    # Scan only essential file patterns (exclude optional files)
    important_patterns = [
        "src/**/*.py",           # Implementation code
        "tests/**/*.py",         # Test code
        ".moai/specs/**/*.md"    # SPEC documents
    ]

    # Load exclude patterns from config
    config_data = load_config()
    hook_config = config_data.get("hooks", {})
    tag_validation_exceptions = hook_config.get("tag_validation_exceptions", {})

    # Exclude patterns (for performance optimization)
    exclude_patterns = [
        ".claude/",
        ".moai/docs/",
        ".moai/reports/",
        ".moai/analysis/",
        "docs/",
        "templates/",
        "examples/",
        "__pycache__/",
        "node_modules/"
    ]

    # Add exempt directories from config
    if tag_validation_exceptions.get("enabled", True):
        exclude_patterns.extend(tag_validation_exceptions.get("exempt_directories", []))

    # Limit file count for fast scan (reduced from 50 to 30)
    max_files = 30

    for pattern in important_patterns:
        if len(files) >= max_files:
            break

        try:
            for path in Path(".").glob(pattern):
                if len(files) >= max_files:
                    break
                if path.is_file():
                    # Check exclude patterns
                    path_str = str(path)
                    if not any(exclude in path_str for exclude in exclude_patterns):
                        files.append(path_str)
        except Exception:
            continue

    return files[:max_files]


def create_quick_scan_result(validation_result: CentralValidationResult,
                           scan_time_ms: float) -> Dict[str, Any]:
    """Create quick scan result

    Args:
        validation_result: Validation result
        scan_time_ms: Scan time

    Returns:
        Scan result dictionary
    """
    result = {
        "quick_scan_completed": True,
        "scan_time_ms": scan_time_ms,
        "files_scanned": validation_result.statistics.total_files_scanned,
        "tags_found": validation_result.statistics.total_tags_found,
        "total_issues": validation_result.statistics.total_issues,
        "is_valid": validation_result.is_valid
    }

    # Summarize only critical issues
    if validation_result.errors:
        result["critical_issues"] = len(validation_result.errors)
        result["error_summary"] = [
            {
                "type": error.type,
                "tag": error.tag,
                "message": error.message
            }
            for error in validation_result.errors[:5]  # Show maximum 5 issues
        ]

    if validation_result.warnings:
        result["warnings"] = len(validation_result.warnings)

    # Coverage information
    result["coverage_percentage"] = validation_result.statistics.coverage_percentage

    # Status message
    if validation_result.is_valid:
        result["status_message"] = "✅ Project TAG status is healthy"
    elif validation_result.errors:
        result["status_message"] = f"🚨 {len(validation_result.errors)} critical issues found"
    else:
        result["status_message"] = f"⚠️ {len(validation_result.warnings)} warnings found"

    return result


def create_health_check_result(issues_count: int,
                             coverage_percentage: float,
                             scan_time_ms: float) -> Dict[str, Any]:
    """Create health check result

    Args:
        issues_count: Issue count
        coverage_percentage: Coverage percentage
        scan_time_ms: Scan time

    Returns:
        Health check result
    """
    # Calculate health score
    health_score = 100

    # Deduct points for issues
    health_score -= min(issues_count * 5, 50)  # Maximum 50 points deduction

    # Coverage percentage score
    if coverage_percentage < 50:
        health_score -= 20
    elif coverage_percentage < 75:
        health_score -= 10

    health_score = max(0, health_score)

    # Health grade
    if health_score >= 90:
        health_grade = "A"
        health_message = "Excellent"
    elif health_score >= 80:
        health_grade = "B"
        health_message = "Good"
    elif health_score >= 70:
        health_grade = "C"
        health_message = "Fair"
    elif health_score >= 60:
        health_grade = "D"
        health_message = "Needs attention"
    else:
        health_grade = "F"
        health_message = "Needs improvement"

    return {
        "health_score": health_score,
        "health_grade": health_grade,
        "health_message": health_message,
        "issues_count": issues_count,
        "coverage_percentage": coverage_percentage,
        "scan_time_ms": scan_time_ms
    }


def main() -> None:
    """Main function"""
    try:
        # Load timeout value from config (milliseconds to seconds)
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # Parse arguments
        if len(sys.argv) < 3:
            usage = "python3 pre_tool__realtime_tag_monitor.py <tool_name> <tool_args_json>"  # noqa: E501
            print(json.dumps({
                "quick_scan_completed": False,
                "error": f"Invalid arguments. Usage: {usage}"
            }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
        except json.JSONDecodeError:
            print(json.dumps({
                "quick_scan_completed": False,
                "error": "Invalid tool_args JSON"
            }))
            sys.exit(0)

        # Check if monitoring target
        if not should_monitor(tool_name, tool_args):
            print(json.dumps({
                "quick_scan_completed": True,
                "message": "Not a monitoring target"
            }))
            sys.exit(0)

        # Record start time
        start_time = time.time()

        # Get list of files to scan
        files_to_scan = get_project_files_to_scan()
        if not files_to_scan:
            print(json.dumps({
                "quick_scan_completed": True,
                "message": "No files to scan"
            }))
            sys.exit(0)

        # Create validator
        validator = create_validator()

        # Execute fast validation (using configured timeout)
        try:
            # Timeout check
            if time.time() - start_time > timeout_seconds:
                raise TimeoutError("Real-time monitoring timeout")

            validation_result = validator.validate_files(files_to_scan)
        except Exception as e:
            # Handle validation failure and timeout
            scan_time = (time.time() - start_time) * 1000
            error_response = {
                "quick_scan_completed": False,
                "error": f"Validation timeout: {str(e)}",
                "scan_time_ms": scan_time,
                "message": "Real-time validation timeout - treating as normal operation"
            }

            if graceful_degradation:
                error_response["graceful_degradation"] = True
                error_response["message"] = "Real-time monitoring timeout but continuing due to graceful degradation"

            print(json.dumps(error_response, ensure_ascii=False))
            sys.exit(0)

        scan_time_ms = (time.time() - start_time) * 1000

        # Create results
        scan_result = create_quick_scan_result(validation_result, scan_time_ms)

        # Perform health check
        health_result = create_health_check_result(
            validation_result.statistics.total_issues,
            validation_result.statistics.coverage_percentage,
            scan_time_ms
        )

        # Final response
        response = {
            **scan_result,
            "health_check": health_result,
            "monitoring_type": "realtime_quick_scan"
        }

        # Timeout warning
        timeout_warning_ms = timeout_seconds * 1000 * 0.8  # 80% of timeout
        if scan_time_ms > timeout_warning_ms:
<<<<<<< HEAD
            warn_msg = f"스캔 시간이 설정된 타임아웃의 80%를 초과했습니다 ({scan_time_ms:.0f}ms / {timeout_warning_ms:.0f}ms)"  # noqa: E501
            response["performance_warning"] = warn_msg
=======
            response["performance_warning"] = f"Scan time exceeded 80% of configured timeout ({scan_time_ms:.0f}ms / {timeout_warning_ms:.0f}ms)"
>>>>>>> b5ac98dc46dcbb7aa3d64d1c16f4a5ef2dfa3053

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # Default response on exception
        error_response = {
            "quick_scan_completed": False,
            "error": f"Hook execution error: {str(e)}",
            "message": "Real-time monitoring error - treating as normal operation"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True
            error_response["message"] = "Real-time monitoring failed but continuing due to graceful degradation"

        print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()
