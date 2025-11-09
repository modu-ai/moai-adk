#!/usr/bin/env python3
# @CODE:HOOK-SESSION-END-TEMPLATE-001 | SPEC: SESSION-END-HOOK-001

"""SessionEnd Hook: Session cleanup and state persistence on session end

Performs the following tasks on session end:
- Cleanup temporary files and cache
- Save session metrics (for productivity analysis)
- Save work state snapshot (ensure work continuity)
- Warn about uncommitted changes
- Generate session summary

Features:
- Cleanup old temporary files
- Cleanup cache files
- Collect and save session metrics
- Work state snapshot (current SPEC, TodoWrite items, etc)
- Detect uncommitted Git changes
- Generate session summary message
"""

import json
import logging
import shutil
import subprocess
import sys
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional

# Add module path
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

try:
    from moai_adk.utils.common import format_duration, get_summary_stats
except ImportError:
    # Fallback implementations if module not found
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
        import statistics
        return {
            "mean": statistics.mean(values),
            "min": min(values),
            "max": max(values),
            "std": statistics.stdev(values) if len(values) > 1 else 0
        }

logger = logging.getLogger(__name__)


def load_hook_timeout() -> int:
    """Load hook timeout from config.json (default: 5000ms)"""
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
                return config.get("hooks", {}).get("timeout_ms", 5000)
    except Exception:
        pass
    return 5000


def get_graceful_degradation() -> bool:
    """Load graceful_degradation setting from config.json (default: true)"""
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
                return config.get("hooks", {}).get("graceful_degradation", True)
    except Exception:
        pass
    return True


def load_config() -> Dict:
    """Load configuration file"""
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
    except Exception:
        pass

    return {}


def cleanup_old_files(config: Dict) -> Dict[str, int]:
    """Cleanup old files

    Args:
        config: Configuration dictionary

    Returns:
        Cleanup statistics with file counts
    """
    stats = {
        "temp_cleaned": 0,
        "cache_cleaned": 0,
        "total_cleaned": 0
    }

    try:
        cleanup_config = config.get("auto_cleanup", {})
        if not cleanup_config.get("enabled", True):
            return stats

        cleanup_days = cleanup_config.get("cleanup_days", 7)
        cutoff_date = datetime.now() - timedelta(days=cleanup_days)

        # Cleanup temporary files
        temp_dir = Path(".moai/temp")
        if temp_dir.exists():
            stats["temp_cleaned"] = cleanup_directory(
                temp_dir,
                cutoff_date,
                None,
                patterns=["*"]
            )

        # Cleanup cache files
        cache_dir = Path(".moai/cache")
        if cache_dir.exists():
            stats["cache_cleaned"] = cleanup_directory(
                cache_dir,
                cutoff_date,
                None,
                patterns=["*"]
            )

        stats["total_cleaned"] = (
            stats["temp_cleaned"] +
            stats["cache_cleaned"]
        )

    except Exception as e:
        logger.error(f"File cleanup failed: {e}")

    return stats


def cleanup_directory(
    directory: Path,
    cutoff_date: datetime,
    max_files: Optional[int],
    patterns: List[str]
) -> int:
    """Cleanup directory files

    Args:
        directory: Target directory
        cutoff_date: Cutoff date for deletion
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
        files_to_check = []
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


def save_session_metrics(payload: Dict) -> bool:
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
            "cwd": str(Path.cwd()),
            "files_modified": count_modified_files(),
            "git_commits": count_recent_commits(),
            "specs_worked_on": extract_specs_from_memory(),
        }

        # Save session metrics
        session_file = logs_dir / f"session-{session_metrics['session_id']}.json"
        with open(session_file, 'w', encoding='utf-8') as f:
            json.dump(session_metrics, f, indent=2, ensure_ascii=False)

        logger.info(f"Session metrics saved: {session_file}")
        return True

    except Exception as e:
        logger.error(f"Failed to save session metrics: {e}")
        return False


def save_work_state(payload: Dict) -> bool:
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
        with open(state_file, 'w', encoding='utf-8') as f:
            json.dump(work_state, f, indent=2, ensure_ascii=False)

        logger.info(f"Work state saved: {state_file}")
        return True

    except Exception as e:
        logger.error(f"Failed to save work state: {e}")
        return False


def check_uncommitted_changes(config: Dict) -> Optional[str]:
    """Warn about uncommitted changes (P0-3)

    Args:
        config: Configuration dictionary

    Returns:
        Warning message or None
    """
    try:
        warnings_config = config.get("session_end", {}).get("warnings", {})
        if not warnings_config.get("uncommitted_changes", True):
            return None

        # Execute git command
        try:
            result = subprocess.run(
                ["git", "status", "--porcelain"],
                capture_output=True,
                text=True,
                timeout=1
            )

            if result.returncode == 0:
                uncommitted = result.stdout.strip()
                if uncommitted:
                    line_count = len(uncommitted.split('\n'))
                    return f"⚠️  {line_count} uncommitted files detected - Consider committing or stashing changes"

        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

    except Exception as e:
        logger.warning(f"Failed to check uncommitted changes: {e}")

    return None


def get_current_branch() -> Optional[str]:
    """Get current Git branch name"""
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True,
            text=True,
            timeout=1
        )

        if result.returncode == 0:
            return result.stdout.strip()

    except Exception:
        pass

    return None


def count_modified_files() -> int:
    """Count modified files"""
    try:
        result = subprocess.run(
            ["git", "status", "--porcelain"],
            capture_output=True,
            text=True,
            timeout=1
        )

        if result.returncode == 0:
            return len([line for line in result.stdout.strip().split('\n') if line])

    except Exception:
        pass

    return 0


def count_uncommitted_files() -> int:
    """Count uncommitted files"""
    return count_modified_files()


def count_recent_commits() -> int:
    """Count recent commits (this session)"""
    try:
        # Get commits from last 1 hour
        result = subprocess.run(
            ["git", "rev-list", "--since=1 hour", "HEAD"],
            capture_output=True,
            text=True,
            timeout=1
        )

        if result.returncode == 0:
            commits = [line for line in result.stdout.strip().split('\n') if line]
            return len(commits)

    except Exception:
        pass

    return 0


def extract_specs_from_memory() -> List[str]:
    """Extract SPEC information from memory"""
    specs = []

    try:
        # Get recent SPEC from command_execution_state.json
        state_file = Path(".moai/memory/command-execution-state.json")
        if state_file.exists():
            with open(state_file, 'r', encoding='utf-8') as f:
                state_data = json.load(f)

            # Extract recent SPEC IDs
            if "last_specs" in state_data:
                specs = state_data["last_specs"][:3]  # Last 3

    except Exception as e:
        logger.warning(f"Failed to extract specs from memory: {e}")

    return specs


def generate_session_summary(cleanup_stats: Dict, work_state: Dict) -> str:
    """Generate session summary (P1-3)

    Args:
        cleanup_stats: Cleanup statistics
        work_state: Work state

    Returns:
        Summary message
    """
    summary_lines = ["✅ Session Ended"]

    try:
        # Work information
        specs = work_state.get("specs_in_progress", [])
        if specs:
            summary_lines.append(f"   • Worked on: {', '.join(specs)}")

        # File change information
        files_modified = work_state.get("uncommitted_files", 0)
        if files_modified > 0:
            summary_lines.append(f"   • Files modified: {files_modified}")

        # Cleanup information
        total_cleaned = cleanup_stats.get("total_cleaned", 0)
        if total_cleaned > 0:
            summary_lines.append(f"   • Cleaned: {total_cleaned} temp files")

    except Exception as e:
        logger.warning(f"Failed to generate session summary: {e}")

    return "\n".join(summary_lines)


def main():
    """Main function"""
    graceful_degradation = False

    try:
        # Load hook timeout setting
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # Timeout check
        import signal
        import time

        def timeout_handler(signum, frame):
            raise TimeoutError("Hook execution timeout")

        signal.signal(signal.SIGALRM, timeout_handler)
        signal.alarm(int(timeout_seconds))

        try:
            start_time = time.time()

            # Load configuration
            config = load_config()

            # Create hook payload (simple version)
            payload = {"cwd": str(Path.cwd())}

            results = {
                "hook": "session_end__auto_cleanup",
                "success": True,
                "execution_time_seconds": 0,
                "cleanup_stats": {"total_cleaned": 0},
                "work_state_saved": False,
                "session_metrics_saved": False,
                "uncommitted_warning": None,
                "session_summary": "",
                "timestamp": datetime.now().isoformat()
            }

            # P0-1: Save session metrics
            if save_session_metrics(payload):
                results["session_metrics_saved"] = True

            # P0-2: Save work state snapshot
            work_state = {}
            if save_work_state(payload):
                results["work_state_saved"] = True
                work_state = {
                    "uncommitted_files": count_uncommitted_files(),
                    "specs_in_progress": extract_specs_from_memory()
                }

            # P0-3: Warn uncommitted changes
            uncommitted_warning = check_uncommitted_changes(config)
            if uncommitted_warning:
                results["uncommitted_warning"] = uncommitted_warning

            # P1-1: Cleanup temporary files
            cleanup_stats = cleanup_old_files(config)
            results["cleanup_stats"] = cleanup_stats

            # P1-3: Generate session summary
            session_summary = generate_session_summary(cleanup_stats, work_state)
            results["session_summary"] = session_summary

            # Record execution time
            execution_time = time.time() - start_time
            results["execution_time_seconds"] = round(execution_time, 2)

            # Print results
            print(json.dumps(results, ensure_ascii=False, indent=2))

        finally:
            signal.alarm(0)  # Disable timeout

    except TimeoutError as e:
        # Timeout handling
        result = {
            "hook": "session_end__auto_cleanup",
            "success": False,
            "error": f"Hook execution timeout: {str(e)}",
            "graceful_degradation": graceful_degradation,
            "timestamp": datetime.now().isoformat()
        }

        if graceful_degradation:
            result["message"] = "Hook timeout but continuing due to graceful degradation"

        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        # Exception handling
        result = {
            "hook": "session_end__auto_cleanup",
            "success": False,
            "error": f"Hook execution failed: {str(e)}",
            "graceful_degradation": graceful_degradation,
            "timestamp": datetime.now().isoformat()
        }

        if graceful_degradation:
            result["message"] = "Hook failed but continuing due to graceful degradation"

        print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
