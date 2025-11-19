"""Analysis report module for session_start hook

Handles session log analysis and daily report generation.

Responsibilities:
- Generate daily analysis reports
- Analyze Claude Code session logs
- Format analysis results as markdown reports
"""

import json
import logging
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)

try:
    from moai_adk.utils.common import format_duration, get_summary_stats
except ImportError:
    # Fallback implementations
    def format_duration(seconds: float) -> str:
        """Format duration in seconds to readable string"""
        if seconds < 60:
            return f"{seconds:.1f}s"
        minutes = seconds / 60
        if minutes < 60:
            return f"{minutes:.1f}m"
        hours = minutes / 60
        return f"{hours:.1f}h"

    def get_summary_stats(values: list) -> Dict[str, float]:
        """Get summary statistics for a list of values"""
        if not values:
            return {"mean": 0, "min": 0, "max": 0, "std": 0}
        import statistics

        return {
            "mean": statistics.mean(values),
            "min": min(values),
            "max": max(values),
            "std": statistics.stdev(values) if len(values) > 1 else 0,
        }


class AnalysisError(Exception):
    """Exception raised for analysis-related errors"""

    pass


def generate_daily_analysis(config: Dict[str, Any]) -> Optional[str]:
    """Generate daily session analysis report

    Args:
        config: Configuration dictionary

    Returns:
        Path to generated report file, or None if disabled/failed

    Raises:
        AnalysisError: If analysis operations fail
    """
    try:
        analysis_config = config.get("daily_analysis", {})
        if not analysis_config.get("enabled", True):
            return None

        # Analyze session logs
        report_path = analyze_session_logs(analysis_config)

        # Update last analysis date in config
        if report_path:
            config_file = Path(".moai/config/config.json")
            if config_file.exists():
                with open(config_file, "r", encoding="utf-8") as f:
                    config_data = json.load(f)

                config_data["daily_analysis"]["last_analysis"] = (
                    datetime.now().strftime("%Y-%m-%d")
                )

                with open(config_file, "w", encoding="utf-8") as f:
                    json.dump(config_data, f, indent=2, ensure_ascii=False)

        return report_path

    except Exception as e:
        logger.error(f"Daily analysis failed: {e}")
        raise AnalysisError(f"Failed to generate daily analysis: {e}") from e


def analyze_session_logs(analysis_config: Dict[str, Any]) -> Optional[str]:
    """Analyze Claude Code session logs

    Args:
        analysis_config: Analysis configuration

    Returns:
        Path to generated report file, or None if no logs found

    Raises:
        AnalysisError: If analysis operations fail
    """
    try:
        # Find Claude Code session logs
        session_logs_dir = Path.home() / ".claude" / "projects"
        project_name = Path.cwd().name

        # Collect sessions for current project
        project_sessions = []
        if session_logs_dir.exists():
            for project_dir in session_logs_dir.iterdir():
                if project_dir.is_dir() and project_dir.name.endswith(project_name):
                    session_files = list(project_dir.glob("session-*.json"))
                    project_sessions.extend(session_files)

        if not project_sessions:
            logger.info("No session logs found")
            return None

        # Analyze recent sessions (last 10)
        recent_sessions = sorted(
            project_sessions, key=lambda f: f.stat().st_mtime, reverse=True
        )[:10]

        # Collect analysis data
        analysis_data = {
            "total_sessions": len(recent_sessions),
            "date_range": "",
            "tools_used": {},
            "errors_found": [],
            "duration_stats": {},
            "recommendations": [],
        }

        if recent_sessions:
            first_session = datetime.fromtimestamp(
                recent_sessions[-1].stat().st_mtime
            )
            last_session = datetime.fromtimestamp(recent_sessions[0].stat().st_mtime)
            analysis_data["date_range"] = (
                f"{first_session.strftime('%Y-%m-%d')} ~ "
                f"{last_session.strftime('%Y-%m-%d')}"
            )

            # Analyze each session
            all_durations = []
            for session_file in recent_sessions:
                try:
                    with open(session_file, "r", encoding="utf-8") as f:
                        session_data = json.load(f)

                    # Analyze tool usage
                    if "tool_use" in session_data:
                        for tool_use in session_data["tool_use"]:
                            tool_name = tool_use.get("name", "unknown")
                            analysis_data["tools_used"][tool_name] = (
                                analysis_data["tools_used"].get(tool_name, 0) + 1
                            )

                    # Collect errors
                    if "errors" in session_data:
                        for error in session_data["errors"]:
                            analysis_data["errors_found"].append(
                                {
                                    "timestamp": error.get("timestamp", ""),
                                    "error": error.get("message", "")[:100],
                                }
                            )

                    # Calculate session duration
                    if "start_time" in session_data and "end_time" in session_data:
                        start = session_data["start_time"]
                        end = session_data["end_time"]
                        if start and end:
                            try:
                                duration = float(end) - float(start)
                                all_durations.append(duration)
                            except (ValueError, TypeError):
                                pass

                except json.JSONDecodeError as e:
                    logger.warning(f"Failed to parse session {session_file}: {e}")
                    continue
                except Exception as e:
                    logger.warning(f"Failed to analyze session {session_file}: {e}")
                    continue

            # Calculate duration statistics
            if all_durations:
                analysis_data["duration_stats"] = get_summary_stats(all_durations)

        # Format and save report
        report_content = format_analysis_report(analysis_data)

        # Save report to file
        base_path = Path(".moai/reports")
        base_path.mkdir(exist_ok=True, parents=True)

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        report_file = base_path / f"daily-analysis-{timestamp}.md"

        with open(report_file, "w", encoding="utf-8") as f:
            f.write(report_content)

        logger.info(f"Daily analysis report saved: {report_file}")
        return str(report_file)

    except Exception as e:
        logger.error(f"Session log analysis failed: {e}")
        raise AnalysisError(f"Failed to analyze session logs: {e}") from e


def format_analysis_report(analysis_data: Dict[str, Any]) -> str:
    """Format analysis results as markdown report

    Args:
        analysis_data: Analysis data dictionary

    Returns:
        Formatted markdown report content
    """
    report_lines = [
        "# Daily Session Analysis Report",
        "",
        f"Generated at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
        f"Analysis Period: {analysis_data.get('date_range', 'N/A')}",
        f"Total Sessions: {analysis_data.get('total_sessions', 0)}",
        "",
        "## 📊 Tool Usage",
        "",
    ]

    # Add tool usage
    tools_used = analysis_data.get("tools_used", {})
    if tools_used:
        sorted_tools = sorted(tools_used.items(), key=lambda x: x[1], reverse=True)
        for tool_name, count in sorted_tools[:10]:  # TOP 10
            report_lines.append(f"- **{tool_name}**: {count} times")
    else:
        report_lines.append("- No tools used")

    report_lines.extend(
        [
            "",
            "## ⚠️ Errors",
            "",
        ]
    )

    # Add error summary
    errors = analysis_data.get("errors_found", [])
    if errors:
        for i, error in enumerate(errors[:5], 1):  # Recent 5 errors
            report_lines.append(
                f"{i}. {error.get('error', 'N/A')} ({error.get('timestamp', 'N/A')})"
            )
    else:
        report_lines.append("- No errors found")

    # Add session duration statistics
    duration_stats = analysis_data.get("duration_stats", {})
    if duration_stats.get("mean", 0) > 0:
        report_lines.extend(
            [
                "",
                "",
                "## ⏱️ Session Duration Statistics",
                "",
                f"- Mean: {format_duration(duration_stats['mean'])}",
                f"- Min: {format_duration(duration_stats['min'])}",
                f"- Max: {format_duration(duration_stats['max'])}",
                f"- Std Dev: {format_duration(duration_stats['std'])}",
            ]
        )

    # Add recommendations
    report_lines.extend(
        [
            "",
            "",
            "## 💡 Recommendations",
            "",
        ]
    )

    # Tool usage based recommendations
    if tools_used:
        most_used_tool = max(tools_used.items(), key=lambda x: x[1])[0]
        if "Bash" in most_used_tool and tools_used[most_used_tool] > 10:
            report_lines.append(
                "- 🔧 Frequent Bash command usage. Consider script automation."
            )

    if len(errors) > 3:
        report_lines.append("- ⚠️ Frequent errors detected. Stability review recommended.")

    if duration_stats.get("mean", 0) > 1800:  # >30 min
        report_lines.append("- ⏰ Long session duration. Consider breaking down tasks.")

    if not report_lines[-1].startswith("-"):
        report_lines.append("- Current session pattern is good.")

    report_lines.extend(
        [
            "",
            "---",
            "---",
            "*Report automatically generated by Alfred's SessionStart Hook*",
            "*Analysis settings can be managed in `daily_analysis` section of `.moai/config/config.json`*",
        ]
    )

    return "\n".join(report_lines)
