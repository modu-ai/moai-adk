#!/usr/bin/env python3
# @CODE:HOOK-SESSION-START-TEMPLATE-001 | SPEC: SESSION-START-HOOK-001

"""SessionStart Hook: Auto cleanup and report generation

Cleans old temporary files and reports on session start
and generates daily analysis reports.

Features:
- Auto cleanup old report files
- Cleanup temporary files
- Generate daily analysis report
- Analyze session logs
"""

import json
import logging
import shutil
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
    """Load hook timeout from config.json (default: 3000ms)"""
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
                return config.get("hooks", {}).get("timeout_ms", 3000)
    except Exception:
        pass
    return 3000


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


def should_cleanup_today(last_cleanup: Optional[str], cleanup_days: int = 7) -> bool:
    """Check if cleanup is needed today

    Args:
        last_cleanup: Last cleanup date (YYYY-MM-DD)
        cleanup_days: Cleanup period (days)

    Returns:
        Cleanup needed status
    """
    if not last_cleanup:
        return True

    try:
        last_date = datetime.strptime(last_cleanup, "%Y-%m-%d")
        next_cleanup = last_date + timedelta(days=cleanup_days)
        return datetime.now() >= next_cleanup
    except Exception:
        return True


def cleanup_old_files(config: Dict) -> Dict[str, int]:
    """Cleanup old files

    Args:
        config: Configuration dictionary

    Returns:
        Cleanup statistics
    """
    stats = {
        "reports_cleaned": 0,
        "cache_cleaned": 0,
        "temp_cleaned": 0,
        "total_cleaned": 0
    }

    try:
        cleanup_config = config.get("auto_cleanup", {})
        if not cleanup_config.get("enabled", True):
            return stats

        cleanup_days = cleanup_config.get("cleanup_days", 7)
        max_reports = cleanup_config.get("max_reports", 10)
        cleanup_targets = cleanup_config.get("cleanup_targets", [])

        cutoff_date = datetime.now() - timedelta(days=cleanup_days)

        # Cleanup report files
        reports_dir = Path(".moai/reports")
        if reports_dir.exists():
            stats["reports_cleaned"] = cleanup_directory(
                reports_dir,
                cutoff_date,
                max_reports,
                patterns=["*.json", "*.md"]
            )

        # Cleanup cache files
        cache_dir = Path(".moai/cache")
        if cache_dir.exists():
            stats["cache_cleaned"] = cleanup_directory(
                cache_dir,
                cutoff_date,
                None,  # Cache has no file count limit
                patterns=["*"]
            )

        # Cleanup temp files
        temp_dir = Path(".moai/temp")
        if temp_dir.exists():
            stats["temp_cleaned"] = cleanup_directory(
                temp_dir,
                cutoff_date,
                None,
                patterns=["*"]
            )

        stats["total_cleaned"] = (
            stats["reports_cleaned"] +
            stats["cache_cleaned"] +
            stats["temp_cleaned"]
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

                # Enforce maximum file count limit
                elif max_files is not None:
                    remaining_files = len([f for f in files_to_check
                                         if f.exists() and
                                         datetime.fromtimestamp(f.stat().st_mtime) >= cutoff_date])
                    if remaining_files > max_files:
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


def generate_daily_analysis(config: Dict) -> Optional[str]:
    """Generate daily analysis report

    Args:
        config: Configuration dictionary

    Returns:
        Generated report file path or None
    """
    try:
        analysis_config = config.get("daily_analysis", {})
        if not analysis_config.get("enabled", True):
            return None

        # Analyze session logs
        report_path = analyze_session_logs(analysis_config)

        # Update last analysis date in config
        if report_path:
            config_file = Path(".moai/config.json")
            if config_file.exists():
                with open(config_file, 'r', encoding='utf-8') as f:
                    config_data = json.load(f)

                config_data["daily_analysis"]["last_analysis"] = datetime.now().strftime("%Y-%m-%d")

                with open(config_file, 'w', encoding='utf-8') as f:
                    json.dump(config_data, f, indent=2, ensure_ascii=False)

        return report_path

    except Exception as e:
        logger.error(f"Daily analysis failed: {e}")
        return None


def analyze_session_logs(analysis_config: Dict) -> Optional[str]:
    """Analyze session logs

    Args:
        analysis_config: Analysis configuration

    Returns:
        Report file path or None
    """
    try:
        # Claude Code session logs path
        session_logs_dir = Path.home() / ".claude" / "projects"
        project_name = Path.cwd().name

        # Find session logs for current project
        project_sessions = []
        for project_dir in session_logs_dir.iterdir():
            if project_dir.name.endswith(project_name):
                session_files = list(project_dir.glob("session-*.json"))
                project_sessions.extend(session_files)

        if not project_sessions:
            return None

        # Analyze recent session logs
        recent_sessions = sorted(project_sessions, key=lambda f: f.stat().st_mtime, reverse=True)[:10]

        # Collect analysis data
        analysis_data = {
            "total_sessions": len(recent_sessions),
            "date_range": "",
            "tools_used": {},
            "errors_found": [],
            "duration_stats": {},
            "recommendations": []
        }

        if recent_sessions:
            first_session = datetime.fromtimestamp(recent_sessions[-1].stat().st_mtime)
            last_session = datetime.fromtimestamp(recent_sessions[0].stat().st_mtime)
            analysis_data["date_range"] = f"{first_session.strftime('%Y-%m-%d')} ~ {last_session.strftime('%Y-%m-%d')}"

            # Analyze each session
            all_durations = []
            for session_file in recent_sessions:
                try:
                    with open(session_file, 'r', encoding='utf-8') as f:
                        session_data = json.load(f)

                    # Analyze tool usage
                    if "tool_use" in session_data:
                        for tool_use in session_data["tool_use"]:
                            tool_name = tool_use.get("name", "unknown")
                            analysis_data["tools_used"][tool_name] = analysis_data["tools_used"].get(tool_name, 0) + 1

                    # Analyze errors
                    if "errors" in session_data:
                        for error in session_data["errors"]:
                            analysis_data["errors_found"].append({
                                "timestamp": error.get("timestamp", ""),
                                "error": error.get("message", "")[:100]  # First 100 characters only
                            })

                    # Analyze session duration
                    if "start_time" in session_data and "end_time" in session_data:
                        start = session_data["start_time"]
                        end = session_data["end_time"]
                        if start and end:
                            try:
                                duration = float(end) - float(start)
                                all_durations.append(duration)
                            except (ValueError, TypeError):
                                pass

                except Exception as e:
                    logger.warning(f"Failed to analyze session {session_file}: {e}")
                    continue

            # Session duration statistics
            if all_durations:
                analysis_data["duration_stats"] = get_summary_stats(all_durations)

        # Generate report
        report_content = format_analysis_report(analysis_data)

        # Save report
        report_location = analysis_config.get("report_location", ".moai/reports/daily-")
        base_path = Path(".moai/reports")
        base_path.mkdir(exist_ok=True)

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        report_file = base_path / f"daily-analysis-{timestamp}.md"

        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(report_content)

        return str(report_file)

    except Exception as e:
        logger.error(f"Session log analysis failed: {e}")
        return None


def format_analysis_report(analysis_data: Dict) -> str:
    """Convert analysis results to report format

    Args:
        analysis_data: Analysis data

    Returns:
        Formatted report content
    """
    report_lines = [
        "# Daily Session Analysis Report",
        "",
        f"Generated time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
        f"Analysis period: {analysis_data.get('date_range', 'N/A')}",
        f"Total sessions: {analysis_data.get('total_sessions', 0)}",
        "",
        "## Tool Usage Report",
        ""
    ]

    # Tool usage ranking
    tools_used = analysis_data.get("tools_used", {})
    if tools_used:
        sorted_tools = sorted(tools_used.items(), key=lambda x: x[1], reverse=True)
        for tool_name, count in sorted_tools[:10]:  # TOP 10
            report_lines.append(f"- **{tool_name}**: {count} times")
    else:
        report_lines.append("- No tools used")

    report_lines.extend([
        "",
        "## Error Report",
        ""
    ])

    # Error status
    errors = analysis_data.get("errors_found", [])
    if errors:
        for i, error in enumerate(errors[:5], 1):  # Last 5
            report_lines.append(f"{i}. {error.get('error', 'N/A')} ({error.get('timestamp', 'N/A')})")
    else:
        report_lines.append("- No errors found")

    # Session duration statistics
    duration_stats = analysis_data.get("duration_stats", {})
    if duration_stats.get("mean", 0) > 0:
        report_lines.extend([
            "",
            "## Session Duration Statistics",
            "",
            f"- Average: {format_duration(duration_stats['mean'])}",
            f"- Minimum: {format_duration(duration_stats['min'])}",
            f"- Maximum: {format_duration(duration_stats['max'])}",
            f"- Std Dev: {format_duration(duration_stats['std'])}"
        ])

    # Improvement recommendations
    report_lines.extend([
        "",
        "## Improvement Recommendations",
        ""
    ])

    # Recommendations based on tool usage patterns
    if tools_used:
        most_used_tool = max(tools_used.items(), key=lambda x: x[1])[0]
        if "Bash" in most_used_tool and tools_used[most_used_tool] > 10:
            report_lines.append("- Consider script automation - frequent Bash command usage detected")

    if len(errors) > 3:
        report_lines.append("- Review stability - frequent errors detected")

    if duration_stats.get("mean", 0) > 1800:  # More than 30 minutes
        report_lines.append("- Consider task splitting - session time exceeds 30 minutes")

    if not report_lines[-1].startswith("-"):
        report_lines.append("- Current session patterns look good")

    report_lines.extend([
        "",
        "---",
        "*보고서는 Alfred의 SessionStart Hook으로 자동 생성되었습니다*",
        "*분석 설정은 `.moai/config.json`의 `daily_analysis` 섹션에서 관리할 수 있습니다*"
    ])

    return "\n".join(report_lines)


def update_cleanup_stats(cleanup_stats: Dict[str, int]):
    """Update cleanup statistics

    Args:
        cleanup_stats: Cleanup statistics
    """
    try:
        stats_file = Path(".moai/cache/cleanup_stats.json")
        stats_file.parent.mkdir(exist_ok=True)

        # Load existing statistics
        existing_stats = {}
        if stats_file.exists():
            with open(stats_file, 'r', encoding='utf-8') as f:
                existing_stats = json.load(f)

        # Add new statistics
        today = datetime.now().strftime("%Y-%m-%d")
        existing_stats[today] = {
            "cleaned_files": cleanup_stats["total_cleaned"],
            "reports_cleaned": cleanup_stats["reports_cleaned"],
            "cache_cleaned": cleanup_stats["cache_cleaned"],
            "temp_cleaned": cleanup_stats["temp_cleaned"],
            "timestamp": datetime.now().isoformat()
        }

        # Keep only last 30 days of statistics
        cutoff_date = datetime.now() - timedelta(days=30)
        filtered_stats = {}
        for date, stats in existing_stats.items():
            try:
                stat_date = datetime.strptime(date, "%Y-%m-%d")
                if stat_date >= cutoff_date:
                    filtered_stats[date] = stats
            except ValueError:
                continue

        # Save statistics
        with open(stats_file, 'w', encoding='utf-8') as f:
            json.dump(filtered_stats, f, indent=2, ensure_ascii=False)

    except Exception as e:
        logger.error(f"Failed to update cleanup stats: {e}")


def main():
    """Main function"""
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

            # Check last cleanup date
            last_cleanup = config.get("auto_cleanup", {}).get("last_cleanup")
            cleanup_days = config.get("auto_cleanup", {}).get("cleanup_days", 7)

            cleanup_stats = {"total_cleaned": 0, "reports_cleaned": 0, "cache_cleaned": 0, "temp_cleaned": 0}
            report_path = None

            # Execute cleanup if needed
            if should_cleanup_today(last_cleanup, cleanup_days):
                cleanup_stats = cleanup_old_files(config)

                # Update last cleanup date
                config_file = Path(".moai/config.json")
                if config_file.exists():
                    with open(config_file, 'r', encoding='utf-8') as f:
                        config_data = json.load(f)

                    config_data["auto_cleanup"]["last_cleanup"] = datetime.now().strftime("%Y-%m-%d")

                    with open(config_file, 'w', encoding='utf-8') as f:
                        json.dump(config_data, f, indent=2, ensure_ascii=False)

                # Update cleanup statistics
                update_cleanup_stats(cleanup_stats)

            # Generate daily analysis report
            last_analysis = config.get("daily_analysis", {}).get("last_analysis")
            if should_cleanup_today(last_analysis, 1):  # Run daily
                report_path = generate_daily_analysis(config)

            # Record execution time
            execution_time = time.time() - start_time

            # Print results
            result = {
                "hook": "session_start__auto_cleanup",
                "success": True,
                "execution_time_seconds": round(execution_time, 2),
                "cleanup_stats": cleanup_stats,
                "daily_analysis_report": report_path,
                "timestamp": datetime.now().isoformat()
            }

            print(json.dumps(result, ensure_ascii=False, indent=2))

        finally:
            signal.alarm(0)  # Disable timeout

    except TimeoutError as e:
        # Timeout handling
        result = {
            "hook": "session_start__auto_cleanup",
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
            "hook": "session_start__auto_cleanup",
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
