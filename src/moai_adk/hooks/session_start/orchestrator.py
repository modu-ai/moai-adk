"""Orchestrator module for session_start hook

Coordinates overall hook execution, timeout management, and result formatting.

Responsibilities:
- Manage hook execution lifecycle
- Handle timeouts gracefully
- Coordinate all cleanup and analysis operations
- Format and output hook results
"""

import json
import logging
import signal
import time
from datetime import datetime
from typing import Any, Dict, Optional

from moai_adk.hooks.session_start.analysis_report import generate_daily_analysis
from moai_adk.hooks.session_start.core_cleanup import (
    CleanupError,
    cleanup_old_files,
    update_cleanup_stats,
)
from moai_adk.hooks.session_start.state_cleanup import (
    StateError,
    get_graceful_degradation,
    load_config,
    load_hook_timeout,
    should_cleanup_today,
)

logger = logging.getLogger(__name__)


class OrchestratorError(Exception):
    """Exception raised for orchestrator-related errors"""

    pass


def execute_cleanup_sequence(config: Dict[str, Any]) -> Dict[str, Any]:
    """Execute complete cleanup sequence

    Args:
        config: Configuration dictionary

    Returns:
        Cleanup results dictionary

    Raises:
        OrchestratorError: If sequence execution fails
    """
    result = {
        "cleanup_completed": False,
        "cleanup_stats": {
            "total_cleaned": 0,
            "reports_cleaned": 0,
            "cache_cleaned": 0,
            "temp_cleaned": 0,
        },
        "report_path": None,
    }

    try:
        # Check if cleanup is needed
        last_cleanup = config.get("auto_cleanup", {}).get("last_cleanup")
        cleanup_days = config.get("auto_cleanup", {}).get("cleanup_days", 7)

        if should_cleanup_today(last_cleanup, cleanup_days):
            # Perform cleanup
            cleanup_stats = cleanup_old_files(config)
            result["cleanup_stats"] = cleanup_stats

            # Update cleanup timestamp in config
            from pathlib import Path

            config_file = Path(".moai/config/config.json")
            if config_file.exists():
                with open(config_file, "r", encoding="utf-8") as f:
                    config_data = json.load(f)

                config_data["auto_cleanup"]["last_cleanup"] = (
                    datetime.now().strftime("%Y-%m-%d")
                )

                with open(config_file, "w", encoding="utf-8") as f:
                    json.dump(config_data, f, indent=2, ensure_ascii=False)

            # Update cleanup statistics
            update_cleanup_stats(cleanup_stats)
            result["cleanup_completed"] = True

    except (CleanupError, StateError) as e:
        logger.error(f"Cleanup sequence failed: {e}")
        raise OrchestratorError(f"Cleanup sequence failed: {e}") from e
    except Exception as e:
        logger.error(f"Unexpected error in cleanup sequence: {e}")
        raise OrchestratorError(f"Cleanup sequence failed: {e}") from e

    return result


def execute_analysis_sequence(config: Dict[str, Any]) -> Dict[str, Any]:
    """Execute daily analysis sequence

    Args:
        config: Configuration dictionary

    Returns:
        Analysis results dictionary

    Raises:
        OrchestratorError: If sequence execution fails
    """
    result = {"analysis_completed": False, "report_path": None}

    try:
        # Check if analysis should run (daily)
        last_analysis = config.get("daily_analysis", {}).get("last_analysis")
        if should_cleanup_today(last_analysis, 1):  # Run daily
            report_path = generate_daily_analysis(config)
            result["report_path"] = report_path
            result["analysis_completed"] = report_path is not None

    except Exception as e:
        logger.error(f"Analysis sequence failed: {e}")
        raise OrchestratorError(f"Analysis sequence failed: {e}") from e

    return result


def format_hook_result(
    success: bool,
    execution_time: float,
    cleanup_stats: Optional[Dict[str, int]] = None,
    report_path: Optional[str] = None,
    error: Optional[str] = None,
    graceful_degradation: bool = True,
) -> Dict[str, Any]:
    """Format hook execution result as JSON

    Args:
        success: Whether execution was successful
        execution_time: Total execution time in seconds
        cleanup_stats: Cleanup statistics (if applicable)
        report_path: Path to generated report (if applicable)
        error: Error message (if failed)
        graceful_degradation: Whether graceful degradation is enabled

    Returns:
        Formatted result dictionary
    """
    result = {
        "hook": "session_start__auto_cleanup",
        "success": success,
        "execution_time_seconds": round(execution_time, 3),
        "timestamp": datetime.now().isoformat(),
    }

    if success:
        result["cleanup_stats"] = cleanup_stats or {
            "total_cleaned": 0,
            "reports_cleaned": 0,
            "cache_cleaned": 0,
            "temp_cleaned": 0,
        }
        result["daily_analysis_report"] = report_path
    else:
        result["error"] = error or "Unknown error"
        result["graceful_degradation"] = graceful_degradation
        if graceful_degradation:
            result["message"] = "Hook failed but continuing due to graceful degradation"

    return result


def handle_timeout(
    graceful_degradation: bool = True,
) -> Dict[str, Any]:
    """Handle hook timeout

    Args:
        graceful_degradation: Whether to continue despite timeout

    Returns:
        Timeout error result

    Raises:
        TimeoutError: If graceful degradation is disabled
    """
    error_msg = "Hook execution timeout"

    if not graceful_degradation:
        raise TimeoutError(error_msg)

    return format_hook_result(
        success=False,
        execution_time=0,
        error=error_msg,
        graceful_degradation=graceful_degradation,
    )


def main() -> None:
    """Main hook execution function

    Executes the complete session_start hook lifecycle:
    1. Load configuration
    2. Setup timeout handler
    3. Execute cleanup sequence
    4. Execute analysis sequence
    5. Format and output results

    Outputs JSON result to stdout.
    """
    graceful_degradation = False
    try:
        start_time = time.time()

        # Load timeout settings
        timeout_ms = load_hook_timeout()
        timeout_seconds = timeout_ms / 1000
        graceful_degradation = get_graceful_degradation()

        # Setup timeout handler
        def timeout_handler(signum, frame) -> None:
            raise TimeoutError("Hook execution timeout")

        signal.signal(signal.SIGALRM, timeout_handler)
        signal.alarm(int(timeout_seconds))

        try:
            # Load configuration
            config = load_config()

            cleanup_stats = {
                "total_cleaned": 0,
                "reports_cleaned": 0,
                "cache_cleaned": 0,
                "temp_cleaned": 0,
            }
            report_path = None

            # Execute cleanup sequence
            try:
                cleanup_result = execute_cleanup_sequence(config)
                cleanup_stats = cleanup_result["cleanup_stats"]
            except OrchestratorError as e:
                logger.warning(f"Cleanup sequence skipped: {e}")

            # Execute analysis sequence
            try:
                analysis_result = execute_analysis_sequence(config)
                report_path = analysis_result["report_path"]
            except OrchestratorError as e:
                logger.warning(f"Analysis sequence skipped: {e}")

            # Calculate execution time
            execution_time = time.time() - start_time

            # Format and output success result
            result = format_hook_result(
                success=True,
                execution_time=execution_time,
                cleanup_stats=cleanup_stats,
                report_path=report_path,
            )

            print(json.dumps(result, ensure_ascii=False, indent=2))

        finally:
            signal.alarm(0)  # Disable timeout

    except TimeoutError as e:
        # Handle timeout
        execution_time = time.time() - start_time
        result = format_hook_result(
            success=False,
            execution_time=execution_time,
            error=f"Hook execution timeout: {str(e)}",
            graceful_degradation=graceful_degradation,
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        # Handle unexpected exceptions
        execution_time = time.time() - start_time
        result = format_hook_result(
            success=False,
            execution_time=execution_time,
            error=f"Hook execution failed: {str(e)}",
            graceful_degradation=graceful_degradation,
        )
        print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
