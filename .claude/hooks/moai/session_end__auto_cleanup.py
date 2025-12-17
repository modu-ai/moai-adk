#!/usr/bin/env python3

"""SessionEnd Hook: Cleanup and state saving on session end

Performs the following tasks on session end:
- Clean up temporary files and cache
- Save session metrics (for productivity analysis)
- Save work state snapshot (ensure work continuity)
- Warn uncommitted changes
- Generate session summary

Features:
- Clean up old temporary files
- Clean up cache files
- Collect and save session metrics
- Work state snapshot (current SPEC, TodoWrite items, etc.)
- Detect uncommitted Git changes
- Generate session summary message
"""

import json
import logging
import shutil
import subprocess
import sys
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, List, Optional

# Add module path
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from lib.path_utils import find_project_root  # noqa: E402

# Import unified timeout manager and Git operations manager
try:
    from lib.common import (  # noqa: E402
        format_duration,
        get_summary_stats,
        is_root_whitelisted,
        suggest_moai_location,
    )
    from lib.config_manager import ConfigManager  # noqa: E402
    from lib.config_validator import ValidationIssue, get_config_validator
    from lib.git_operations_manager import GitOperationType, get_git_manager
    from lib.unified_timeout_manager import (
        HookTimeoutConfig,
        HookTimeoutError,
        TimeoutPolicy,
        get_timeout_manager,
        hook_timeout_context,
    )
except ImportError:
    # Fallback implementations if new modules not available
    def get_timeout_manager():
        return None

    def hook_timeout_context(hook_name, config=None):
        import contextlib

        @contextlib.contextmanager
        def dummy_context():
            yield

        return dummy_context()

    class HookTimeoutConfig:  # type: ignore[no-redef]
        def __init__(self, **kwargs):
            pass

    class TimeoutPolicy:  # type: ignore[no-redef]
        FAST = "fast"
        NORMAL = "normal"
        SLOW = "slow"

    class HookTimeoutError(Exception):  # type: ignore[no-redef]
        pass

    def get_git_manager():
        return None

    class GitOperationType:  # type: ignore[no-redef]
        STATUS = "status"
        LOG = "log"

    def get_config_validator():
        return None

    class ValidationIssue:  # type: ignore[no-redef]
        pass

    ConfigManager = None  # type: ignore
    # Fallback implementations if module not found
    import statistics

    def format_duration(seconds):
        """Format duration in seconds to readable string"""
        if seconds < 60:
            return f"{seconds:.1f}s"
        minutes = seconds / 60
        if minutes < 60:
            return f"{minutes:.1f}m"
        hours = minutes / 60
        return f"{hours:.1f}h"

    def get_summary_stats(values):
        """Get summary statistics for a list of values"""
        if not values:
            return {"mean": 0, "min": 0, "max": 0, "std": 0}
        return {
            "mean": statistics.mean(values),
            "min": min(values),
            "max": max(values),
            "std": statistics.stdev(values) if len(values) > 1 else 0,
        }


logger = logging.getLogger(__name__)


def load_hook_timeout() -> int:
    """Load hook timeout from config.json (default: 5000ms)

    Returns:
        Timeout in milliseconds
    """
    try:
        config_file = Path(".moai/config/config.yaml")
        if config_file.exists():
            with open(config_file, "r", encoding="utf-8") as f:
                config: Dict[str, Any] = json.load(f)
                return config.get("hooks", {}).get("timeout_ms", 5000)
    except Exception:
        pass
    return 5000


def get_graceful_degradation() -> bool:
    """Load graceful_degradation setting from config.json (default: true)

    Returns:
        Whether graceful degradation is enabled
    """
    try:
        config_file = Path(".moai/config/config.yaml")
        if config_file.exists():
            with open(config_file, "r", encoding="utf-8") as f:
                config: Dict[str, Any] = json.load(f)
                return config.get("hooks", {}).get("graceful_degradation", True)
    except Exception:
        pass
    return True


def cleanup_old_files(config: Dict[str, Any]) -> Dict[str, int]:
    """Clean up old files

    Args:
        config: Configuration dictionary

    Returns:
        Statistics of cleaned files
    """
    stats = {"temp_cleaned": 0, "cache_cleaned": 0, "total_cleaned": 0}

    try:
        cleanup_config = config.get("auto_cleanup", {})
        if not cleanup_config.get("enabled", True):
            return stats

        cleanup_days = cleanup_config.get("cleanup_days", 7)
        cutoff_date = datetime.now() - timedelta(days=cleanup_days)

        # Clean up temporary files
        temp_dir = Path(".moai/temp")
        if temp_dir.exists():
            stats["temp_cleaned"] = cleanup_directory(temp_dir, cutoff_date, None, patterns=["*"])

        # Clean up cache files
        cache_dir = Path(".moai/cache")
        if cache_dir.exists():
            stats["cache_cleaned"] = cleanup_directory(cache_dir, cutoff_date, None, patterns=["*"])

        stats["total_cleaned"] = stats["temp_cleaned"] + stats["cache_cleaned"]

    except Exception as e:
        logger.error(f"File cleanup failed: {e}")

    return stats


def cleanup_directory(
    directory: Path,
    cutoff_date: datetime,
    max_files: Optional[int],
    patterns: List[str],
) -> int:
    """Clean up directory files

    Args:
        directory: Target directory
        cutoff_date: Cutoff date threshold
        max_files: Maximum number of files to keep
        patterns: List of file patterns to delete

    Returns:
        Number of deleted files
    """
    if not directory.exists():
        return 0

    cleaned_count = 0

    try:
        # Collect files matching patterns
        files_to_check: list[Path] = []
        for pattern in patterns:
            files_to_check.extend(directory.glob(pattern))

        # Sort by date (oldest first)
        files_to_check.sort(key=lambda f: f.stat().st_mtime)

        # Delete files
        for file_path in files_to_check:
            try:
                # Check file modification time
                file_mtime = datetime.fromtimestamp(file_path.stat().st_mtime)

                # Delete if before cutoff date
                if file_mtime < cutoff_date:
                    if file_path.is_file():
                        file_path.unlink()
                        cleaned_count += 1
                    elif file_path.is_dir():
                        shutil.rmtree(file_path)
                        cleaned_count += 1

            except Exception as e:
                logger.warning(f"Failed to delete {file_path}: {e}")
                continue

    except Exception as e:
        logger.error(f"Directory cleanup failed for {directory}: {e}")

    return cleaned_count


def save_session_metrics(payload: Dict[str, Any]) -> bool:
    """Save session metrics (P0-1)

    Args:
        payload: Hook payload

    Returns:
        Success status
    """
    try:
        # Create logs directory
        logs_dir = Path(".moai/logs/sessions")
        logs_dir.mkdir(parents=True, exist_ok=True)

        # Collect session information
        session_metrics = {
            "session_id": datetime.now().strftime("%Y-%m-%d-%H%M%S"),
            "end_time": datetime.now().isoformat(),
            "cwd": str(find_project_root()),
            "files_modified": count_modified_files(),
            "git_commits": count_recent_commits(),
            "specs_worked_on": extract_specs_from_memory(),
        }

        # Save session metrics
        session_file = logs_dir / f"session-{session_metrics['session_id']}.json"
        with open(session_file, "w", encoding="utf-8") as f:
            json.dump(session_metrics, f, indent=2, ensure_ascii=False)

        logger.info(f"Session metrics saved: {session_file}")
        return True

    except Exception as e:
        logger.error(f"Failed to save session metrics: {e}")
        return False


def save_work_state(payload: Dict[str, Any]) -> bool:
    """Save work state snapshot (P0-2)

    Args:
        payload: Hook payload

    Returns:
        Success status
    """
    try:
        # Create memory directory
        memory_dir = Path(".moai/memory")
        memory_dir.mkdir(parents=True, exist_ok=True)

        # Collect work state
        work_state = {
            "last_updated": datetime.now().isoformat(),
            "current_branch": get_current_branch(),
            "uncommitted_changes": check_uncommitted_changes(),
            "uncommitted_files": count_uncommitted_files(),
            "specs_in_progress": extract_specs_from_memory(),
        }

        # Save state
        state_file = memory_dir / "last-session-state.json"
        with open(state_file, "w", encoding="utf-8") as f:
            json.dump(work_state, f, indent=2, ensure_ascii=False)

        logger.info(f"Work state saved: {state_file}")
        return True

    except Exception as e:
        logger.error(f"Failed to save work state: {e}")
        return False


def check_uncommitted_changes() -> Optional[str]:
    """Warn uncommitted changes (P0-3) using optimized Git operations

    Returns:
        Warning message or None
    """
    git_manager = get_git_manager()
    if git_manager:
        try:
            # Use optimized Git manager
            from lib.git_operations_manager import GitCommand

            status_result = git_manager.execute_git_command(
                GitCommand(
                    operation_type=GitOperationType.STATUS,
                    args=["status", "--porcelain"],
                    cache_ttl_seconds=5,  # Short TTL for status
                    timeout_seconds=3,
                )
            )

            if status_result.success:
                uncommitted = status_result.stdout.strip()
                if uncommitted:
                    line_count = len(uncommitted.split("\n"))
                    return f"⚠️  {line_count} uncommitted files detected - Consider committing or stashing changes"

        except Exception as e:
            logger.warning(f"Git manager failed for uncommitted changes check: {e}")

    # Fallback to direct Git command
    try:
        result = subprocess.run(["git", "status", "--porcelain"], capture_output=True, text=True, timeout=1)

        if result.returncode == 0:
            uncommitted = result.stdout.strip()
            if uncommitted:
                line_count = len(uncommitted.split("\n"))
                return f"⚠️  {line_count} uncommitted files detected - Consider committing or stashing changes"

    except (subprocess.TimeoutExpired, FileNotFoundError):
        pass
    except Exception as e:
        logger.warning(f"Failed to check uncommitted changes: {e}")

    return None


def get_current_branch() -> Optional[str]:
    """Get current Git branch name using optimized Git operations

    Returns:
        Branch name or None if query fails
    """
    git_manager = get_git_manager()
    if git_manager:
        try:
            from lib.git_operations_manager import GitCommand

            branch_result = git_manager.execute_git_command(
                GitCommand(
                    operation_type=GitOperationType.BRANCH,
                    args=["rev-parse", "--abbrev-ref", "HEAD"],
                    cache_ttl_seconds=30,
                    timeout_seconds=3,
                )
            )

            if branch_result.success:
                return branch_result.stdout.strip()

        except Exception as e:
            logger.warning(f"Git manager failed for branch query: {e}")

    # Fallback to direct Git command
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True,
            text=True,
            timeout=1,
        )
        if result.returncode == 0:
            return result.stdout.strip()
    except Exception:
        pass

    return None


def count_modified_files() -> int:
    """Count number of modified files using optimized Git operations"""
    git_manager = get_git_manager()
    if git_manager:
        try:
            from lib.git_operations_manager import GitCommand

            status_result = git_manager.execute_git_command(
                GitCommand(
                    operation_type=GitOperationType.STATUS,
                    args=["status", "--porcelain"],
                    cache_ttl_seconds=5,
                    timeout_seconds=3,
                )
            )

            if status_result.success:
                return len([line for line in status_result.stdout.strip().split("\n") if line])

        except Exception as e:
            logger.warning(f"Git manager failed for file count: {e}")

    # Fallback to direct Git command
    try:
        result = subprocess.run(["git", "status", "--porcelain"], capture_output=True, text=True, timeout=1)
        if result.returncode == 0:
            return len([line for line in result.stdout.strip().split("\n") if line])
    except Exception:
        pass

    return 0


def count_uncommitted_files() -> int:
    """Count number of uncommitted files"""
    return count_modified_files()


def count_recent_commits() -> int:
    """Count recent commits (during this session) using optimized Git operations"""
    git_manager = get_git_manager()
    if git_manager:
        try:
            from lib.git_operations_manager import GitCommand

            log_result = git_manager.execute_git_command(
                GitCommand(
                    operation_type=GitOperationType.LOG,
                    args=["rev-list", "--since=1 hour", "HEAD"],
                    cache_ttl_seconds=60,  # Cache for 1 minute
                    timeout_seconds=5,
                )
            )

            if log_result.success:
                commits = [line for line in log_result.stdout.strip().split("\n") if line]
                return len(commits)

        except Exception as e:
            logger.warning(f"Git manager failed for recent commits: {e}")

    # Fallback to direct Git command
    try:
        result = subprocess.run(
            ["git", "rev-list", "--since=1 hour", "HEAD"],
            capture_output=True,
            text=True,
            timeout=1,
        )
        if result.returncode == 0:
            commits = [line for line in result.stdout.strip().split("\n") if line]
            return len(commits)
    except Exception:
        pass

    return 0


def extract_specs_from_memory() -> List[str]:
    """Extract SPEC information from memory"""
    specs = []

    try:
        # Query recent SPECs from command_execution_state.json
        state_file = Path(".moai/memory/command-execution-state.json")
        if state_file.exists():
            with open(state_file, "r", encoding="utf-8") as f:
                state_data = json.load(f)

            # Extract recent SPEC IDs
            if "last_specs" in state_data:
                specs = state_data["last_specs"][:3]  # Latest 3

    except Exception as e:
        logger.warning(f"Failed to extract specs from memory: {e}")

    return specs


# Note: is_root_whitelisted, get_file_pattern_category, and suggest_moai_location
# are now imported from lib.common (consolidated from duplicate implementations)


def scan_root_violations(config: Dict[str, Any]) -> List[Dict[str, str]]:
    """Scan project root for document management violations

    Args:
        config: Configuration dictionary

    Returns:
        List of violation dictionaries with file info and suggested location
    """
    violations = []

    try:
        # Get project root
        project_root = Path(".moai/config/config.yaml").parent.parent
        if not project_root.exists():
            project_root = find_project_root()

        # Scan root directory
        for item in project_root.iterdir():
            # Skip directories (except backup directories)
            if item.is_dir():
                # Check for backup directories
                if item.name.endswith("-backup") or item.name.endswith("_backup") or "_backup_" in item.name:
                    suggested = suggest_moai_location(item.name, config)
                    violations.append(
                        {
                            "file": item.name + "/",
                            "type": "directory",
                            "suggested": suggested,
                        }
                    )
                continue

            # Skip hidden files and directories
            if item.name.startswith("."):
                continue

            # Check if whitelisted
            if is_root_whitelisted(item.name, config):
                continue

            # Not whitelisted - add to violations
            suggested = suggest_moai_location(item.name, config)
            violations.append({"file": item.name, "type": "file", "suggested": suggested})

    except Exception as e:
        logger.warning(f"Failed to scan root violations: {e}")

    return violations


def generate_migration_report(violations: List[Dict[str, str]]) -> str:
    """Generate migration suggestions report

    Args:
        violations: List of violations

    Returns:
        Formatted report string
    """
    if not violations:
        return ""

    report_lines = [
        "\n⚠️ Document Management Violations Detected",
        f"   Found {len(violations)} misplaced file(s) in project root:\n",
    ]

    for idx, violation in enumerate(violations, 1):
        file_display = violation["file"]
        suggested = violation["suggested"]
        report_lines.append(f"   {idx}. {file_display} → {suggested}")

    report_lines.append("\n   Action: Move files to suggested locations or update root_whitelist")
    report_lines.append('   Guide: Skill("moai-core-document-management")')

    return "\n".join(report_lines)


def generate_session_summary(
    cleanup_stats: Dict[str, int], work_state: Dict[str, Any], violations_count: int = 0
) -> str:
    """Generate session summary (P1-3)

    Args:
        cleanup_stats: Cleanup statistics
        work_state: Work state
        violations_count: Number of document management violations

    Returns:
        Summary message
    """
    summary_lines = ["✅ Session Ended"]

    try:
        # Work information
        specs = work_state.get("specs_in_progress", [])
        if specs:
            summary_lines.append(f"   • Worked on: {', '.join(specs)}")

        # File modification information
        files_modified = work_state.get("uncommitted_files", 0)
        if files_modified > 0:
            summary_lines.append(f"   • Files modified: {files_modified}")

        # Cleanup information
        total_cleaned = cleanup_stats.get("total_cleaned", 0)
        if total_cleaned > 0:
            summary_lines.append(f"   • Cleaned: {total_cleaned} temp files")

        # Document management violations
        if violations_count > 0:
            summary_lines.append(f"   ⚠️ {violations_count} root violations detected (see below)")

    except Exception as e:
        logger.warning(f"Failed to generate session summary: {e}")

    return "\n".join(summary_lines)


def execute_session_end_workflow() -> tuple[Dict[str, Any], str]:
    """Execute the session end workflow with proper error handling"""
    start_time = time.time()

    # Load configuration
    if ConfigManager:
        config = ConfigManager().load_config()
    else:
        config = {}

    # Generate hook payload (simple version)
    payload = {"cwd": str(find_project_root())}

    results = {
        "hook": "session_end__auto_cleanup",
        "success": True,
        "execution_time_seconds": 0,
        "cleanup_stats": {"total_cleaned": 0},
        "work_state_saved": False,
        "session_metrics_saved": False,
        "uncommitted_warning": None,
        "session_summary": "",
        "timestamp": datetime.now().isoformat(),
        "performance": {
            "git_manager_used": get_git_manager() is not None,
            "timeout_manager_used": get_timeout_manager() is not None,
            "config_validator_used": get_config_validator() is not None,
        },
    }

    try:
        # P0-1: Save session metrics
        if save_session_metrics(payload):
            results["session_metrics_saved"] = True

        # P0-2: Save work state snapshot
        work_state = {}
        if save_work_state(payload):
            results["work_state_saved"] = True
            work_state = {
                "uncommitted_files": count_uncommitted_files(),
                "specs_in_progress": extract_specs_from_memory(),
            }

        # P0-3: Warn uncommitted changes
        uncommitted_warning = check_uncommitted_changes()
        if uncommitted_warning:
            results["uncommitted_warning"] = uncommitted_warning

        # P1-1: Clean up temporary files
        cleanup_stats = cleanup_old_files(config)
        results["cleanup_stats"] = cleanup_stats

        # P1-2: Document Management - Scan root violations
        violations = []
        migration_report = ""
        doc_mgmt = config.get("document_management", {})
        if doc_mgmt.get("enabled", True):
            violations = scan_root_violations(config)
            if violations:
                migration_report = generate_migration_report(violations)
                results["document_violations"] = {
                    "count": len(violations),
                    "violations": violations,
                }

        # P1-3: Generate session summary
        session_summary = generate_session_summary(cleanup_stats, work_state, len(violations))
        results["session_summary"] = session_summary

        # Add migration report to summary if violations exist
        if migration_report:
            results["migration_report"] = migration_report

        # Record execution time
        execution_time = time.time() - start_time
        results["execution_time_seconds"] = round(execution_time, 2)

        return results, migration_report

    except Exception as e:
        results["success"] = False
        results["error"] = str(e)
        results["execution_time_seconds"] = round(time.time() - start_time, 2)
        return results, ""


def main() -> None:
    """Main function

    SessionEnd Hook entry point for cleanup and work state tracking.
    Cleans up temporary files, saves session metrics, and warns uncommitted changes.

    Features:
    - Optimized timeout handling with unified manager
    - Enhanced error handling with graceful degradation
    - Resource monitoring and cleanup
    - Performance optimization with Git operations manager

    Returns:
        None
    """
    # Configure timeout for session end hook
    timeout_config = HookTimeoutConfig(
        policy=TimeoutPolicy.NORMAL,
        custom_timeout_ms=5000,  # 5 seconds
        retry_count=1,
        retry_delay_ms=500,
        graceful_degradation=True,
        memory_limit_mb=150,  # Higher memory limit for cleanup operations
    )

    # Use unified timeout manager if available
    timeout_manager = get_timeout_manager()
    if timeout_manager:
        try:
            results, migration_report = timeout_manager.execute_with_timeout(
                "session_end__auto_cleanup",
                execute_session_end_workflow,
                config=timeout_config,
            )

            # Print results
            output_lines = [json.dumps(results, ensure_ascii=False, indent=2)]

            # Print migration report separately for visibility
            if migration_report:
                output_lines.append(migration_report)

            print("\n".join(output_lines))

        except HookTimeoutError as e:
            # Enhanced timeout error handling
            timeout_response = {
                "hook": "session_end__auto_cleanup",
                "success": False,
                "error": f"Hook execution timeout: {str(e)}",
                "error_details": {
                    "hook_id": e.hook_id,
                    "timeout_seconds": e.timeout_seconds,
                    "execution_time": e.execution_time,
                    "will_retry": e.will_retry,
                },
                "graceful_degradation": True,
                "timestamp": datetime.now().isoformat(),
            }
            timeout_response["message"] = "Hook timeout but continuing due to graceful degradation"
            print(json.dumps(timeout_response, ensure_ascii=False, indent=2))

        except Exception as e:
            # Enhanced error handling with context
            error_response = {
                "hook": "session_end__auto_cleanup",
                "success": False,
                "error": f"Hook execution failed: {str(e)}",
                "error_details": {
                    "error_type": type(e).__name__,
                    "message": str(e),
                    "graceful_degradation": True,
                },
                "timestamp": datetime.now().isoformat(),
            }
            error_response["message"] = "Hook failed but continuing due to graceful degradation"
            print(json.dumps(error_response, ensure_ascii=False, indent=2))

    else:
        # Fallback to legacy timeout handling
        try:
            timeout_seconds = load_hook_timeout() / 1000
            graceful_degradation = get_graceful_degradation()

            # Legacy timeout implementation
            import signal

            def timeout_handler(signum, frame):
                raise TimeoutError("Hook execution timeout")

            signal.signal(signal.SIGALRM, timeout_handler)
            signal.alarm(int(timeout_seconds))

            try:
                results, migration_report = execute_session_end_workflow()

                # Print results
                output_lines = [json.dumps(results, ensure_ascii=False, indent=2)]

                # Print migration report separately for visibility
                if migration_report:
                    output_lines.append(migration_report)

                print("\n".join(output_lines))

            finally:
                signal.alarm(0)  # Clear timeout

        except TimeoutError as e:
            # Handle timeout with graceful degradation
            result = {
                "hook": "session_end__auto_cleanup",
                "success": False,
                "error": f"Hook execution timeout: {str(e)}",
                "graceful_degradation": graceful_degradation,
                "timestamp": datetime.now().isoformat(),
            }

            if graceful_degradation:
                result["message"] = "Hook timeout but continuing due to graceful degradation"

            print(json.dumps(result, ensure_ascii=False, indent=2))

        except Exception as e:
            # Handle exceptions with graceful degradation
            result = {
                "hook": "session_end__auto_cleanup",
                "success": False,
                "error": f"Hook execution failed: {str(e)}",
                "graceful_degradation": graceful_degradation,
                "timestamp": datetime.now().isoformat(),
            }

            if graceful_degradation:
                result["message"] = "Hook failed but continuing due to graceful degradation"

            print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
