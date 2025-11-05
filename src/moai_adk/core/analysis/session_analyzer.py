"""
MoAI-ADK Session Analyzer

Claude Code 세션 로그를 분석하여 데이터 기반 개선 제안 생성

This module provides the SessionAnalyzer class for analyzing Claude Code session logs
and generating improvement suggestions based on usage patterns.
"""

import json
from collections import defaultdict
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, Optional


class SessionAnalyzer:
    """Claude Code 세션 로그 분석기"""

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
            # 파일 수정 시간 확인
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

            # Tool 사용 패턴 추출
            if event_type == "tool_call":
                tool_name = event.get("toolName", "unknown").split("(")[0]
                self.patterns["tool_usage"][tool_name] += 1

            # Tool 오류 패턴
            elif event_type == "tool_error":
                error_msg = event.get("error", "unknown error")
                self.patterns["tool_failures"][error_msg[:50]] += 1  # 처음 50자
                has_error = True

            # 권한 요청
            elif event_type == "permission_request":
                perm_type = event.get("permission_type", "unknown")
                self.patterns["permission_requests"][perm_type] += 1

            # Hook 실패
            elif event_type == "hook_failure":
                hook_name = event.get("hook_name", "unknown")
                self.patterns["hook_failures"][hook_name] += 1
                has_error = True

            # 명령어 사용
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

        report = f"""# MoAI-ADK 세션 메타분석 리포트

**생성 일시**: {timestamp}
**분석 기간**: 최근 {self.days_back}일
**분석 범위**: `~/.claude/projects/`

---

## 📊 전체 메트릭

| 메트릭 | 값 |
|------|-----|
| **총 세션 수** | {total_sessions} |
| **총 이벤트 수** | {self.patterns['total_events']} |
| **성공 세션** | {total_sessions - self.patterns['failed_sessions']} ({success_rate:.1f}%) |
| **실패 세션** | {self.patterns['failed_sessions']} ({100 - success_rate:.1f}%) |
| **평균 세션 길이** | {self.patterns['total_events'] / total_sessions if total_sessions > 0 else 0:.1f} 이벤트 |

---

## 🔧 도구 사용 패턴 (상위 10)

"""

        # 상위 도구 사용
        sorted_tools = sorted(
            self.patterns["tool_usage"].items(), key=lambda x: x[1], reverse=True
        )

        report += "| 도구 | 사용 횟수 |\n|------|----------|\n"
        for tool, count in sorted_tools[:10]:
            report += f"| `{tool}` | {count} |\n"

        # Tool 오류 패턴
        report += "\n## ⚠️ 도구 오류 패턴 (상위 5)\n\n"

        if self.patterns["tool_failures"]:
            sorted_errors = sorted(
                self.patterns["tool_failures"].items(),
                key=lambda x: x[1],
                reverse=True,
            )
            report += "| 오류 | 발생 횟수 |\n|------|----------|\n"
            for error, count in sorted_errors[:5]:
                report += f"| {error}... | {count} |\n"
        else:
            report += "✅ 도구 오류 없음\n"

        # Hook 실패 분석
        report += "\n## 🪝 Hook 실패 분석\n\n"

        if self.patterns["hook_failures"]:
            for hook, count in sorted(
                self.patterns["hook_failures"].items(),
                key=lambda x: x[1],
                reverse=True,
            ):
                report += f"- **{hook}**: {count}회\n"
        else:
            report += "✅ Hook 실패 없음\n"

        # 권한 요청 분석
        report += "\n## 🔐 권한 요청 패턴\n\n"

        if self.patterns["permission_requests"]:
            sorted_perms = sorted(
                self.patterns["permission_requests"].items(),
                key=lambda x: x[1],
                reverse=True,
            )
            report += "| 권한 유형 | 요청 횟수 |\n|---------|----------|\n"
            for perm, count in sorted_perms:
                report += f"| {perm} | {count} |\n"
        else:
            report += "✅ 권한 요청 없음\n"

        # 개선 제안
        report += "\n## 💡 개선 제안\n\n"
        report += self._generate_suggestions()

        return report

    def _generate_suggestions(self) -> str:
        """
        Generate improvement suggestions based on patterns

        Returns:
            Formatted suggestions string
        """
        suggestions = []

        # 높은 권한 요청 → 권한 설정 재검토
        if self.patterns["permission_requests"]:
            top_perm = max(
                self.patterns["permission_requests"].items(),
                key=lambda x: x[1],
            )
            if top_perm[1] >= 5:
                suggestions.append(
                    f"🔐 **{top_perm[0]}** 권한이 자주 요청됨 ({top_perm[1]}회)\n"
                    f"   → `.claude/settings.json`의 `permissions` 재검토 필요\n"
                    f"   → `allow` → `ask`로 변경하거나 새 Bash 도구 규칙 추가"
                )

        # Tool 실패 패턴 → 회피 전략 추가
        if self.patterns["tool_failures"]:
            top_error = max(
                self.patterns["tool_failures"].items(),
                key=lambda x: x[1],
            )
            if top_error[1] >= 3:
                suggestions.append(
                    f"🔧 **도구 오류**: '{top_error[0]}...' ({top_error[1]}회)\n"
                    f"   → CLAUDE.md에 회피 전략 추가\n"
                    f"   → 예: 'X 오류 시 Y를 시도하세요'"
                )

        # Hook 실패 → Hook 로직 검토
        if self.patterns["hook_failures"]:
            for hook, count in sorted(
                self.patterns["hook_failures"].items(),
                key=lambda x: x[1],
                reverse=True,
            )[:3]:
                if count >= 2:
                    suggestions.append(
                        f"🪝 **Hook 실패**: {hook} ({count}회)\n"
                        f"   → `.claude/hooks/alfred/{hook}.py` 디버깅 필요\n"
                        f"   → 타임아웃, 권한, 파일 경로 확인"
                    )

        # 낮은 성공률 → 전반적 진단
        success_rate = (
            ((self.patterns["total_sessions"] - self.patterns["failed_sessions"])
             / self.patterns["total_sessions"] * 100)
            if self.patterns["total_sessions"] > 0
            else 0
        )

        if success_rate < 80 and self.patterns["total_sessions"] >= 5:
            suggestions.append(
                f"📉 **낮은 성공률** ({success_rate:.1f}%)\n"
                f"   → 최근 세션 로그 상세 검토\n"
                f"   → CLAUDE.md의 규칙/제약 재평가\n"
                f"   → Alfred와 Sub-agent 간 컨텍스트 동기화 확인"
            )

        if not suggestions:
            suggestions.append(
                "✅ **No major issues detected**\n"
                "   → 현재 설정과 규칙이 잘 작동 중"
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
