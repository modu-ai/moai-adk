"""
MoAI-ADK Session Analyzer

Analyzes Claude Code session logs to generate data-driven improvement suggestions

This module provides the SessionAnalyzer class for analyzing Claude Code session logs
and generating improvement suggestions based on usage patterns.
"""

import json
from collections import defaultdict
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, Optional


class SessionAnalyzer:
    """Claude Code session log analyzer"""

    def __init__(self, days_back: int = 7, verbose: bool = False):
        """
        Initialize SessionAnalyzer

        Args:
            days_back: Number of days to analyze (default: 7)
            verbose: Enable verbose output (default: False)
        """
        self.claude_projects = Path.home() / ".claude" / "projects"
        self.days_back = days_back
        self.verbose = verbose

        self.patterns = {
            "total_sessions": 0,
            "total_events": 0,
            "tool_usage": defaultdict(int),
            "tool_failures": defaultdict(int),
            "error_patterns": defaultdict(int),
            "permission_requests": defaultdict(int),
            "hook_failures": defaultdict(int),
            "command_frequency": defaultdict(int),
            "average_session_length": 0,
            "success_rate": 0.0,
            "failed_sessions": 0,
        }

        self.sessions_data = []

    def parse_sessions(self) -> Dict[str, Any]:
        """
        Parse all session logs from the last N days

        Returns:
            Dictionary containing analysis patterns and metrics
        """
        if not self.claude_projects.exists():
            if self.verbose:
                print(f"⚠️ Claude projects directory not found: {self.claude_projects}")
            return self.patterns

        cutoff_date = datetime.now() - timedelta(days=self.days_back)

        # Look for both session-*.json and UUID.jsonl files
        session_files = []
        session_files.extend(self.claude_projects.glob("*/session-*.json"))
        session_files.extend(self.claude_projects.glob("*/*.jsonl"))

        if self.verbose:
            print(f"Found {len(session_files)} session files")

        for session_file in session_files:
            # Check file modification time
            if datetime.fromtimestamp(session_file.stat().st_mtime) < cutoff_date:
                continue

            try:
                # Handle both JSON and JSONL formats
                if session_file.suffix == '.jsonl':
                    # JSONL format: read line by line
                    sessions = []
                    with open(session_file, encoding="utf-8") as f:
                        for line_num, line in enumerate(f, 1):
                            line = line.strip()
                            if line:
                                try:
                                    session = json.loads(line)
                                    sessions.append(session)
                                except json.JSONDecodeError as e:
                                    if self.verbose:
                                        print(f"⚠️ Error reading line {line_num} in {session_file}: {e}")

                    # Analyze each session from the JSONL file
                    for session in sessions:
                        self._analyze_session(session)
                        self.sessions_data.append(session)
                else:
                    # JSON format: single session per file
                    with open(session_file, encoding="utf-8") as f:
                        session = json.load(f)
                        self._analyze_session(session)
                        self.sessions_data.append(session)
            except (json.JSONDecodeError, IOError) as e:
                if self.verbose:
                    print(f"⚠️ Error reading {session_file}: {e}")

        self.patterns["total_sessions"] = len(self.sessions_data)
        return self.patterns

    def _analyze_session(self, session: Dict[str, Any]):
        """
        Analyze individual session

        Args:
            session: Session data dictionary from Claude Code
        """
        # Handle session summary format (current JSONL format)
        if session.get("type") == "summary":
            # Count session types by summary content
            summary = session.get("summary", "").lower()

            # Simple analysis of session summaries
            if any(keyword in summary for keyword in ["error", "fail", "issue", "problem"]):
                self.patterns["failed_sessions"] += 1
                self.patterns["tool_failures"]["session_error_in_summary"] += 1

            # Extract potential tool usage from summary
            tool_keywords = ["test", "build", "deploy", "analyze", "create", "update", "fix", "check"]
            for keyword in tool_keywords:
                if keyword in summary:
                    self.patterns["tool_usage"][f"summary_{keyword}"] += 1

            # Track session summaries as events
            self.patterns["total_events"] += 1
            return

        # Handle detailed event format (legacy session-*.json format)
        events = session.get("events", [])
        self.patterns["total_events"] += len(events)

        has_error = False

        for event in events:
            event_type = event.get("type", "unknown")

            # Extract tool usage patterns
            if event_type == "tool_call":
                tool_name = event.get("toolName", "unknown").split("(")[0]
                self.patterns["tool_usage"][tool_name] += 1

            # Tool error patterns
            elif event_type == "tool_error":
                error_msg = event.get("error", "unknown error")
                self.patterns["tool_failures"][error_msg[:50]] += 1  # First 50 characters
                has_error = True

            # Permission requests
            elif event_type == "permission_request":
                perm_type = event.get("permission_type", "unknown")
                self.patterns["permission_requests"][perm_type] += 1

            # Hook failures
            elif event_type == "hook_failure":
                hook_name = event.get("hook_name", "unknown")
                self.patterns["hook_failures"][hook_name] += 1
                has_error = True

            # Command usage
            if "command" in event:
                cmd = event.get("command", "").split()[0]
                if cmd:
                    self.patterns["command_frequency"][cmd] += 1

        if has_error:
            self.patterns["failed_sessions"] += 1

    def generate_report(self) -> str:
        """
        Generate markdown report

        Returns:
            Formatted markdown report string
        """
        timestamp = datetime.now().isoformat()
        total_sessions = self.patterns["total_sessions"]
        success_rate = (
            ((total_sessions - self.patterns["failed_sessions"]) / total_sessions * 100)
            if total_sessions > 0
            else 0
        )

        report = f"""# MoAI-ADK Session Meta-Analysis Report

**Generated at**: {timestamp}
**Analysis period**: Last {self.days_back} days
**Analysis scope**: `~/.claude/projects/`

---

## Overall Metrics

| Metric | Value |
|--------|-------|
| **Total sessions** | {total_sessions} |
| **Total events** | {self.patterns['total_events']} |
| **Successful sessions** | {total_sessions - self.patterns['failed_sessions']} ({success_rate:.1f}%) |
| **Failed sessions** | {self.patterns['failed_sessions']} ({100 - success_rate:.1f}%) |
| **Average session length** | {self.patterns['total_events'] / total_sessions if total_sessions > 0 else 0:.1f} events |

---

## Tool Usage Patterns (Top 10)

"""

        # Top tool usage
        sorted_tools = sorted(
            self.patterns["tool_usage"].items(), key=lambda x: x[1], reverse=True
        )

        report += "| Tool | Usage Count |\n|------|----------|\n"
        for tool, count in sorted_tools[:10]:
            report += f"| `{tool}` | {count} |\n"

        # Tool error patterns
        report += "\n## Tool Error Patterns (Top 5)\n\n"

        if self.patterns["tool_failures"]:
            sorted_errors = sorted(
                self.patterns["tool_failures"].items(),
                key=lambda x: x[1],
                reverse=True,
            )
            report += "| Error | Occurrence Count |\n|--------|----------|\n"
            for error, count in sorted_errors[:5]:
                report += f"| {error}... | {count} |\n"
        else:
            report += "✅ No tool errors\n"

        # Hook failure analysis
        report += "\n## Hook Failure Analysis\n\n"

        if self.patterns["hook_failures"]:
            for hook, count in sorted(
                self.patterns["hook_failures"].items(),
                key=lambda x: x[1],
                reverse=True,
            ):
                report += f"- **{hook}**: {count} times\n"
        else:
            report += "✅ No hook failures\n"

        # Permission request analysis
        report += "\n## Permission Request Patterns\n\n"

        if self.patterns["permission_requests"]:
            sorted_perms = sorted(
                self.patterns["permission_requests"].items(),
                key=lambda x: x[1],
                reverse=True,
            )
            report += "| Permission Type | Request Count |\n|---------|----------|\n"
            for perm, count in sorted_perms:
                report += f"| {perm} | {count} |\n"
        else:
            report += "✅ No permission requests\n"

        # Improvement suggestions
        report += "\n## Improvement Suggestions\n\n"
        report += self._generate_suggestions()

        return report

    def _generate_suggestions(self) -> str:
        """
        Generate improvement suggestions based on patterns

        Returns:
            Formatted suggestions string
        """
        suggestions = []

        # High permission requests - review permission settings
        if self.patterns["permission_requests"]:
            top_perm = max(
                self.patterns["permission_requests"].items(),
                key=lambda x: x[1],
            )
            if top_perm[1] >= 5:
                suggestions.append(
                    f"Permission: **{top_perm[0]}** requested frequently ({top_perm[1]} times)\n"
                    f"   - Review `permissions` in `.claude/settings.json`\n"
                    f"   - Change `allow` to `ask` or add new Bash tool rules"
                )

        # Tool failure patterns - add fallback strategies
        if self.patterns["tool_failures"]:
            top_error = max(
                self.patterns["tool_failures"].items(),
                key=lambda x: x[1],
            )
            if top_error[1] >= 3:
                suggestions.append(
                    f"Tool error: **{top_error[0]}...** ({top_error[1]} times)\n"
                    f"   - Add fallback strategy to CLAUDE.md\n"
                    f"   - Example: 'If X error occurs, try Y'"
                )

        # Hook failures - review hook logic
        if self.patterns["hook_failures"]:
            for hook, count in sorted(
                self.patterns["hook_failures"].items(),
                key=lambda x: x[1],
                reverse=True,
            )[:3]:
                if count >= 2:
                    suggestions.append(
                        f"Hook failure: **{hook}** ({count} times)\n"
                        f"   - Debug `.claude/hooks/alfred/{hook}.py`\n"
                        f"   - Check timeouts, permissions, and file paths"
                    )

        # Low success rate - general diagnosis
        success_rate = (
            ((self.patterns["total_sessions"] - self.patterns["failed_sessions"])
             / self.patterns["total_sessions"] * 100)
            if self.patterns["total_sessions"] > 0
            else 0
        )

        if success_rate < 80 and self.patterns["total_sessions"] >= 5:
            suggestions.append(
                f"Low success rate: **{success_rate:.1f}%**\n"
                f"   - Review recent session logs in detail\n"
                f"   - Re-evaluate rules and constraints in CLAUDE.md\n"
                f"   - Verify context synchronization between Alfred and Sub-agents"
            )

        if not suggestions:
            suggestions.append(
                "✅ No major issues detected\n"
                "   - Current settings and rules are working well"
            )

        return "\n\n".join(suggestions)

    def save_report(self, output_path: Optional[Path] = None, project_path: Optional[Path] = None) -> Path:
        """
        Save report to file

        Args:
            output_path: Custom output file path (optional)
            project_path: Project root path (defaults to current working directory)

        Returns:
            Path to the saved report file
        """
        if output_path is None:
            if project_path is None:
                project_path = Path.cwd()

            output_dir = project_path / ".moai" / "reports"
            output_dir.mkdir(parents=True, exist_ok=True)
            output_path = output_dir / f"daily-{datetime.now().strftime('%Y-%m-%d')}.md"

        report = self.generate_report()
        output_path.write_text(report, encoding="utf-8")

        if self.verbose:
            print(f"📄 Report saved: {output_path}")

        return output_path

    def get_metrics(self) -> Dict[str, Any]:
        """
        Get analysis metrics as dictionary

        Returns:
            Dictionary containing analysis metrics
        """
        total_sessions = self.patterns["total_sessions"]
        if total_sessions > 0:
            self.patterns["success_rate"] = (
                (total_sessions - self.patterns["failed_sessions"]) / total_sessions * 100
            )
            self.patterns["average_session_length"] = (
                self.patterns["total_events"] / total_sessions
            )

        return self.patterns.copy()
