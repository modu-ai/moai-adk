#!/usr/bin/env python3
"""Daily Analysis Handler: Trigger daily session analysis and self-learning

This handler implements daily session analysis with self-learning capabilities
for continuous improvement and adaptation of Claude Code usage patterns.
"""

import json
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any, Dict


@dataclass
class AnalysisResult:
    """Result of daily analysis operation"""
    continue_execution: bool = True
    analysis_completed: bool = False
    days_since_last_analysis: int = 0
    report_path: str = ""
    message: str = ""


def handle_daily_analysis(session_data: Dict[str, Any]) -> AnalysisResult:
    """Handle daily session analysis

    Args:
        session_data: Session start data from Claude Code

    Returns:
        AnalysisResult: Analysis operation result
    """
    project_root = Path(session_data.get("projectPath", "."))
    moai_dir = project_root / ".moai"

    if not moai_dir.exists():
        return AnalysisResult(
            message="No .moai directory found - skipping daily analysis"
        )

    # Check last analysis date
    last_analysis_file = moai_dir / "last_analysis_date.json"
    cutoff_days = 1  # Run analysis every day

    today = datetime.now()
    last_analysis_date = None

    if last_analysis_file.exists():
        try:
            last_data = json.loads(last_analysis_file.read_text(encoding="utf-8"))
            last_analysis_str = last_data.get("last_analysis_date")
            if last_analysis_str:
                last_analysis_date = datetime.fromisoformat(last_analysis_str)
        except (json.JSONDecodeError, ValueError, KeyError):
            pass  # If file is corrupt, ignore and proceed

    days_since_last = (today - last_analysis_date).days if last_analysis_date else 999

    if days_since_last < cutoff_days:
        return AnalysisResult(
            analysis_completed=False,
            days_since_last_analysis=days_since_last,
            message=f"Analysis not due yet (last: {days_since_last} days ago)"
        )

    # Run analysis
    return _run_daily_analysis(project_root, moai_dir, today)


def _run_daily_analysis(project_root: Path, moai_dir: Path, analysis_date: datetime) -> AnalysisResult:
    """Run the actual daily analysis

    Args:
        project_root: Project root directory
        moai_dir: .moai directory
        analysis_date: Current analysis date

    Returns:
        AnalysisResult: Analysis operation result
    """
    try:
        # Import session analyzer from moai-adk package
        from moai_adk.core.analysis import SessionAnalyzer

        # Create analyzer for last 1 day
        analyzer = SessionAnalyzer(days_back=1, verbose=False)

        # Run analysis
        analysis_results = analyzer.parse_sessions()

        # Generate report
        reports_dir = moai_dir / "reports"
        reports_dir.mkdir(exist_ok=True)

        report_filename = f"daily-{analysis_date.strftime('%Y-%m-%d')}.md"
        report_path = reports_dir / report_filename

        # Generate report content
        report_content = _generate_daily_report(analysis_results, analysis_date)
        report_path.write_text(report_content, encoding="utf-8")

        # Update last analysis date
        last_analysis_file = moai_dir / "last_analysis_date.json"
        last_analysis_data = {
            "last_analysis_date": analysis_date.isoformat(),
            "report_file": report_filename,
            "metrics": {
                "total_sessions": analysis_results.get("total_sessions", 0),
                "total_events": analysis_results.get("total_events", 0),
                "success_rate": analysis_results.get("success_rate", 0.0)
            }
        }
        last_analysis_file.write_text(json.dumps(last_analysis_data, indent=2), encoding="utf-8")

        # Self-learning: Store patterns and insights
        _store_learning_patterns(moai_dir, analysis_results, analysis_date)

        return AnalysisResult(
            continue_execution=True,
            analysis_completed=True,
            days_since_last_analysis=1,
            report_path=str(report_path.relative_to(project_root)),
            message=f"Daily analysis completed: {len(analysis_results.get('tool_usage', {}))} tools analyzed"
        )

    except Exception as e:
        return AnalysisResult(
            analysis_completed=False,
            message=f"Analysis failed: {str(e)}"
        )


def _generate_daily_report(analysis_results: Dict[str, Any], analysis_date: datetime) -> str:
    """Generate daily analysis report content

    Args:
        analysis_results: Results from session analysis
        analysis_date: Analysis date

    Returns:
        str: Markdown report content
    """
    report_lines = [
        "# Daily Session Analysis Report",
        "",
        f"**Analysis Date**: {analysis_date.strftime('%Y-%m-%d %H:%M:%S')}",
        "**Analysis Period**: Last 24 hours",
        "",
        "## Summary Metrics",
        "",
        f"- **Total Sessions**: {analysis_results.get('total_sessions', 0)}",
        f"- **Total Events**: {analysis_results.get('total_events', 0)}",
        f"- **Success Rate**: {analysis_results.get('success_rate', 0):.1%}",
        f"- **Failed Sessions**: {analysis_results.get('failed_sessions', 0)}",
        "",
        "## Tool Usage",
        ""
    ]

    tool_usage = analysis_results.get("tool_usage", {})
    if tool_usage:
        for tool, count in sorted(tool_usage.items(), key=lambda x: x[1], reverse=True)[:10]:
            report_lines.append(f"- **{tool}**: {count} uses")

    report_lines.extend([
        "",
        "## Error Patterns",
        ""
    ])

    error_patterns = analysis_results.get("error_patterns", {})
    if error_patterns:
        for error, count in sorted(error_patterns.items(), key=lambda x: x[1], reverse=True)[:5]:
            report_lines.append(f"- **{error}**: {count} occurrences")
    else:
        report_lines.append("- No error patterns detected")

    report_lines.extend([
        "",
        "## Daily Insights",
        "",
        "1. Monitor tool usage patterns for daily optimization opportunities",
        "2. Address any recurring error patterns immediately",
        "3. Review frequently used tools for performance optimization",
        "4. Track daily productivity and workflow efficiency",
        "",
        "---",
        "*Generated by MoAI-ADK Daily Analysis*"
    ])

    return "\n".join(report_lines)


def to_dict(self) -> Dict[str, Any]:
    """Convert AnalysisResult to dictionary for JSON serialization"""
    return {
        "continue": self.continue_execution,
        "hookSpecificOutput": {
            "analysis_completed": self.analysis_completed,
            "days_since_last_analysis": self.days_since_last_analysis,
            "report_path": self.report_path,
            "message": self.message
        }
    }


def _store_learning_patterns(moai_dir: Path, analysis_results: Dict[str, Any], analysis_date: datetime) -> None:
    """Store daily analysis results for self-learning and pattern recognition

    Args:
        moai_dir: .moai directory
        analysis_results: Results from session analysis
        analysis_date: Analysis date
    """
    memory_dir = moai_dir / "memory"
    patterns_file = memory_dir / "daily-patterns.json"

    # Load existing patterns
    if patterns_file.exists():
        patterns_data = json.loads(patterns_file.read_text(encoding="utf-8"))
    else:
        patterns_data = {
            "version": "1.0.0",
            "created_at": analysis_date.isoformat(),
            "last_updated": analysis_date.isoformat(),
            "patterns": {
                "tool_usage": {},
                "error_patterns": {},
                "success_patterns": {},
                "workflow_optimizations": {}
            },
            "insights": {
                "most_used_tools": [],
                "common_errors": [],
                "optimization_suggestions": []
            },
            "metadata": {
                "total_days_analyzed": 0,
                "patterns_identified": 0,
                "insights_generated": 0
            }
        }

    # Update tool usage patterns
    tool_usage = analysis_results.get("tool_usage", {})
    for tool, count in tool_usage.items():
        if tool not in patterns_data["patterns"]["tool_usage"]:
            patterns_data["patterns"]["tool_usage"][tool] = []
        patterns_data["patterns"]["tool_usage"][tool].append({
            "date": analysis_date.isoformat(),
            "count": count,
            "type": "usage"
        })

    # Update error patterns
    error_patterns = analysis_results.get("error_patterns", {})
    for error, count in error_patterns.items():
        if error not in patterns_data["patterns"]["error_patterns"]:
            patterns_data["patterns"]["error_patterns"][error] = []
        patterns_data["patterns"]["error_patterns"][error].append({
            "date": analysis_date.isoformat(),
            "count": count,
            "type": "error"
        })

    # Update success rate patterns
    success_rate = analysis_results.get("success_rate", 0.0)
    if "daily_success_rates" not in patterns_data["patterns"]:
        patterns_data["patterns"]["daily_success_rates"] = []
    patterns_data["patterns"]["daily_success_rates"].append({
        "date": analysis_date.isoformat(),
        "rate": success_rate
    })

    # Generate insights
    insights = _generate_insights(patterns_data, analysis_results)
    patterns_data["insights"].update(insights)

    # Update metadata
    patterns_data["last_updated"] = analysis_date.isoformat()
    patterns_data["metadata"]["total_days_analyzed"] += 1
    patterns_data["metadata"]["patterns_identified"] = len(tool_usage) + len(error_patterns)
    patterns_data["metadata"]["insights_generated"] = len([v for v in insights.values() if v])

    # Save updated patterns
    patterns_file.write_text(json.dumps(patterns_data, indent=2), encoding="utf-8")


def _generate_insights(patterns_data: Dict[str, Any], current_results: Dict[str, Any]) -> Dict[str, Any]:
    """Generate insights from analysis patterns

    Args:
        patterns_data: Existing patterns data
        current_results: Current day's analysis results

    Returns:
        Dict[str, Any]: Generated insights
    """
    insights = {}

    # Most used tools insight
    tool_usage = current_results.get("tool_usage", {})
    if tool_usage:
        most_used = max(tool_usage.items(), key=lambda x: x[1])
        insights["most_used_tools"] = [{
            "tool": most_used[0],
            "count": most_used[1],
            "date": datetime.now().isoformat()
        }]

    # Common errors insight
    error_patterns = current_results.get("error_patterns", {})
    if error_patterns:
        most_common_error = max(error_patterns.items(), key=lambda x: x[1])
        insights["common_errors"] = [{
            "error": most_common_error[0],
            "count": most_common_error[1],
            "date": datetime.now().isoformat()
        }]

    # Optimization suggestions
    suggestions = []

    # Suggest optimization for frequently used tools
    if tool_usage:
        top_tools = sorted(tool_usage.items(), key=lambda x: x[1], reverse=True)[:3]
        suggestions.append({
            "type": "tool_optimization",
            "description": f"Focus on optimizing {top_tools[0][0]} usage ({top_tools[0][1]} uses)",
            "priority": "high"
        })

    # Suggest error resolution
    if error_patterns:
        for error, count in error_patterns.items():
            if count >= 3:  # Errors occurring 3+ times
                suggestions.append({
                    "type": "error_resolution",
                    "description": f"Address recurring error: {error} ({count} occurrences)",
                    "priority": "critical"
                })

    insights["optimization_suggestions"] = suggestions

    return insights


# Add to_dict method to AnalysisResult
AnalysisResult.to_dict = to_dict
