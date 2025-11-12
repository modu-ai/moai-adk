#!/usr/bin/env python3
"""Real-time TAG Monitoring Hook

Continuous TAG status monitoring and real-time violation detection.
Quick inspection of project-wide TAG status during PreToolUse stage.

Features:
- Real-time TAG status monitoring
- Fast violation detection (within 5 seconds)
- Project-wide TAG integrity inspection
- Immediate user feedback

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


# Import utility modules with fallback
try:
    from ..utils.gitignore_parser import get_combined_exclude_patterns
    from ..utils.hook_config import get_graceful_degradation, load_hook_timeout
except ImportError:
    # Fallback for standalone execution
    import sys
    import os
    hook_dir = os.path.dirname(os.path.abspath(__file__))
    sys.path.insert(0, os.path.join(hook_dir, 'utils'))
    from gitignore_parser import get_combined_exclude_patterns
    from hook_config import get_graceful_degradation, load_hook_timeout


def load_config() -> Dict[str, Any]:
    """Load configuration file

    Returns:
        Configuration dictionary
    """
    try:
        config_file = Path(".moai/config/config.json")
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
    """Check if the tool should be monitored

    Args:
        tool_name: Tool name
        tool_args: Tool arguments

    Returns:
        True if the tool should be monitored
    """
    # Monitor only file manipulation tools
    monitoring_tools = {"Edit", "Write", "MultiEdit"}
    return tool_name in monitoring_tools


def get_project_files_to_scan() -> List[str]:
    """Get list of project files to scan (optimized)

    Performance improvement by excluding optional file directories.
    Scan only essential files: src/, tests/, .moai/specs/
    Automatically integrate .gitignore patterns.

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

    # Base exclude patterns
    base_exclude_patterns = [
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

    # Combine with .gitignore patterns (auto-sync)
    exclude_patterns = get_combined_exclude_patterns(base_exclude_patterns)

    # Limit file count for fast scanning (reduced from 50 to 30)
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
        scan_time_ms: Scan time in milliseconds

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
            for error in validation_result.errors[:5]  # Display max 5
        ]

    if validation_result.warnings:
        result["warnings"] = len(validation_result.warnings)

    # Coverage information
    result["coverage_percentage"] = validation_result.statistics.coverage_percentage

    # Status message
    if validation_result.is_valid:
        result["status_message"] = "✅ Project TAG status healthy"
    elif validation_result.errors:
        result["status_message"] = f"🚨 {len(validation_result.errors)} critical issues detected"
    else:
        result["status_message"] = f"⚠️ {len(validation_result.warnings)} warnings detected"

    return result


def create_health_check_result(issues_count: int,
                             coverage_percentage: float,
                             scan_time_ms: float) -> Dict[str, Any]:
    """Create project health status result

    Args:
        issues_count: Number of issues
        coverage_percentage: Coverage percentage
        scan_time_ms: Scan time in milliseconds

    Returns:
        Health status result
    """
    # Calculate health status
    health_score = 100

    # Deduct score for issues
    health_score -= min(issues_count * 5, 50)  # Max 50 points deduction

    # Coverage score
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
        health_message = "Needs Attention"
    else:
        health_grade = "F"
        health_message = "Needs Improvement"

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
        # Load timeout value from config (milliseconds → seconds)
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

        # Execute quick validation (using configured timeout)
        try:
            # Check timeout
            if time.time() - start_time > timeout_seconds:
                raise TimeoutError("Real-time monitoring timeout")

            validation_result = validator.validate_files(files_to_scan)
        except Exception as e:
            # Handle timeout on validation failure
            scan_time = (time.time() - start_time) * 1000
            error_response = {
                "quick_scan_completed": False,
                "error": f"Validation timeout: {str(e)}",
                "scan_time_ms": scan_time,
                "message": "Real-time validation timeout - considered normal operation"
            }

            if graceful_degradation:
                error_response["graceful_degradation"] = True
                error_response["message"] = "Real-time monitoring timeout but continuing due to graceful degradation"

            print(json.dumps(error_response, ensure_ascii=False))
            sys.exit(0)

        scan_time_ms = (time.time() - start_time) * 1000

        # Generate result
        scan_result = create_quick_scan_result(validation_result, scan_time_ms)

        # Health status check
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
            warn_msg = f"Scan time exceeded 80% of configured timeout ({scan_time_ms:.0f}ms / {timeout_warning_ms:.0f}ms)"  # noqa: E501
            response["performance_warning"] = warn_msg

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # Default response on exception
        error_response = {
            "quick_scan_completed": False,
            "error": f"Hook execution error: {str(e)}",
            "message": "Real-time monitoring error - considered normal operation"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True
            error_response["message"] = "Real-time monitoring failed but continuing due to graceful degradation"

        print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()
