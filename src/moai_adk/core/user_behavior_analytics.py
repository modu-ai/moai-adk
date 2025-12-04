"""
User Behavior Analytics System

Analyzes user interactions, patterns, and preferences to provide insights
for system optimization and user experience improvement.

Key Features:
- User interaction tracking and pattern analysis
- Command usage frequency and preference analysis
- Session behavior analysis
- User preference learning and adaptation
- Collaboration pattern analysis
- Productivity metrics and insights
- User experience optimization recommendations
"""

import json
import logging
import statistics
import uuid
from collections import Counter, defaultdict, deque
from dataclasses import asdict, dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Set

# Set up logging
logger = logging.getLogger(__name__)


class UserActionType(Enum):
    """Types of user actions tracked"""

    COMMAND_EXECUTION = "command_execution"
    TOOL_USAGE = "tool_usage"
    FILE_OPERATION = "file_operation"
    ERROR_OCCURRED = "error_occurred"
    HELP_REQUESTED = "help_requested"
    SESSION_START = "session_start"
    SESSION_END = "session_end"
    TASK_COMPLETED = "task_completed"
    PHASE_TRANSITION = "phase_transition"


class SessionState(Enum):
    """User session states"""

    ACTIVE = "active"
    IDLE = "idle"
    FOCUSED = "focused"
    STRUGGLING = "struggling"
    PRODUCTIVE = "productive"
    BREAK = "break"


@dataclass
class UserAction:
    """Single user action event"""

    timestamp: datetime
    action_type: UserActionType
    user_id: str
    session_id: str
    action_data: Dict[str, Any]
    context: Dict[str, Any] = field(default_factory=dict)
    duration_ms: Optional[float] = None
    success: bool = True
    tags: Set[str] = field(default_factory=set)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "timestamp": self.timestamp.isoformat(),
            "action_type": self.action_type.value,
            "user_id": self.user_id,
            "session_id": self.session_id,
            "action_data": self.action_data,
            "context": self.context,
            "duration_ms": self.duration_ms,
            "success": self.success,
            "tags": list(self.tags),
        }


@dataclass
class UserSession:
    """User session information"""

    session_id: str
    user_id: str
    start_time: datetime
    end_time: Optional[datetime] = None
    actions: List[UserAction] = field(default_factory=list)
    state: SessionState = SessionState.ACTIVE
    productivity_score: float = 0.0
    total_commands: int = 0
    total_errors: int = 0
    total_duration_ms: float = 0.0
    working_directory: str = ""
    git_branch: str = ""
    modified_files: Set[str] = field(default_factory=set)
    tools_used: Set[str] = field(default_factory=set)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "session_id": self.session_id,
            "user_id": self.user_id,
            "start_time": self.start_time.isoformat(),
            "end_time": self.end_time.isoformat() if self.end_time else None,
            "state": self.state.value,
            "productivity_score": self.productivity_score,
            "total_commands": self.total_commands,
            "total_errors": self.total_errors,
            "total_duration_ms": self.total_duration_ms,
            "working_directory": self.working_directory,
            "git_branch": self.git_branch,
            "modified_files": list(self.modified_files),
            "tools_used": list(self.tools_used),
        }


@dataclass
class UserPreferences:
    """Learned user preferences and patterns"""

    user_id: str
    preferred_commands: Dict[str, int] = field(default_factory=dict)
    preferred_tools: Dict[str, int] = field(default_factory=dict)
    working_hours: Dict[str, int] = field(default_factory=dict)
    peak_productivity_times: List[str] = field(default_factory=list)
    common_workflows: List[List[str]] = field(default_factory=list)
    error_patterns: List[str] = field(default_factory=list)
    success_patterns: List[str] = field(default_factory=list)
    collaboration_patterns: Dict[str, int] = field(default_factory=dict)
    last_updated: datetime = field(default_factory=datetime.now)


class UserBehaviorAnalytics:
    """Main user behavior analytics system"""

    def __init__(self, storage_path: Optional[Path] = None):
        self.storage_path = storage_path or Path.cwd() / ".moai" / "analytics" / "user_behavior"
        self.storage_path.mkdir(parents=True, exist_ok=True)

        # Data storage
        self.user_sessions: Dict[str, UserSession] = {}
        self.active_sessions: Dict[str, UserSession] = {}
        self.user_preferences: Dict[str, UserPreferences] = {}
        self.action_history: deque = deque(maxlen=10000)

        # Analysis caches
        self._pattern_cache: Dict[str, Any] = {}
        self._insight_cache: Dict[str, Any] = {}
        self._last_cache_update = datetime.now()

        # Load existing data
        self._load_data()

    def track_action(
        self,
        action_type: UserActionType,
        user_id: str,
        session_id: str,
        action_data: Dict[str, Any],
        context: Optional[Dict[str, Any]] = None,
        duration_ms: Optional[float] = None,
        success: bool = True,
    ) -> None:
        """Track a user action"""
        action = UserAction(
            timestamp=datetime.now(),
            action_type=action_type,
            user_id=user_id,
            session_id=session_id,
            action_data=action_data,
            context=context or {},
            duration_ms=duration_ms,
            success=success,
            tags=self._extract_action_tags(action_type, action_data),
        )

        # Store action
        self.action_history.append(action)

        # Update session if active
        if session_id in self.active_sessions:
            session = self.active_sessions[session_id]
            session.actions.append(action)

            # Update session metrics
            if action_type == UserActionType.COMMAND_EXECUTION:
                session.total_commands += 1
                if not success:
                    session.total_errors += 1

                # Track tools used
                if "tool" in action_data:
                    session.tools_used.add(action_data["tool"])

                # Track file modifications
                if "files" in action_data:
                    session.modified_files.update(action_data["files"])

            if duration_ms:
                session.total_duration_ms += duration_ms

            # Update session state
            session.state = self._analyze_session_state(session)

        # Update user preferences
        self._update_user_preferences(user_id, action)

        # Clear caches periodically
        if (datetime.now() - self._last_cache_update).seconds > 300:
            self._pattern_cache.clear()
            self._insight_cache.clear()
            self._last_cache_update = datetime.now()

    def start_session(self, user_id: str, working_directory: str = "", git_branch: str = "") -> str:
        """Start tracking a new user session"""
        session_id = str(uuid.uuid4())

        session = UserSession(
            session_id=session_id,
            user_id=user_id,
            start_time=datetime.now(),
            working_directory=working_directory,
            git_branch=git_branch,
            state=SessionState.ACTIVE,
        )

        self.active_sessions[session_id] = session
        self.user_sessions[session_id] = session

        # Track session start action
        self.track_action(
            UserActionType.SESSION_START,
            user_id,
            session_id,
            {"working_directory": working_directory, "git_branch": git_branch},
        )

        return session_id

    def end_session(self, session_id: str) -> Optional[UserSession]:
        """End a user session"""
        if session_id not in self.active_sessions:
            return None

        session = self.active_sessions[session_id]
        session.end_time = datetime.now()

        # Calculate final productivity score
        session.productivity_score = self._calculate_productivity_score(session)

        # Move from active to completed sessions
        del self.active_sessions[session_id]

        # Track session end action
        self.track_action(
            UserActionType.SESSION_END,
            session.user_id,
            session_id,
            {
                "duration_ms": session.total_duration_ms,
                "commands": session.total_commands,
                "errors": session.total_errors,
                "productivity_score": session.productivity_score,
            },
        )

        # Save data periodically
        self._save_data()

        return session

    def get_user_patterns(self, user_id: str, days: int = 30) -> Dict[str, Any]:
        """Get user behavior patterns and insights"""
        cache_key = f"patterns_{user_id}_{days}"

        if cache_key in self._pattern_cache:
            return self._pattern_cache[cache_key]

        cutoff_date = datetime.now() - timedelta(days=days)

        # Get user sessions in the time range
        user_sessions = [
            session
            for session in self.user_sessions.values()
            if session.user_id == user_id and session.start_time >= cutoff_date
        ]

        # Analyze patterns
        patterns = {
            "session_count": len(user_sessions),
            "avg_session_duration": 0.0,
            "avg_productivity_score": 0.0,
            "most_used_commands": {},
            "most_used_tools": {},
            "peak_productivity_hours": [],
            "common_workflows": [],
            "error_patterns": [],
            "working_preferences": {},
        }

        if user_sessions:
            # Session metrics
            durations = [s.total_duration_ms for s in user_sessions if s.total_duration_ms > 0]
            productivity_scores = [s.productivity_score for s in user_sessions if s.productivity_score > 0]

            if durations:
                patterns["avg_session_duration"] = statistics.mean(durations)
            if productivity_scores:
                patterns["avg_productivity_score"] = statistics.mean(productivity_scores)

            # Command usage
            all_commands = []
            all_tools = []

            for session in user_sessions:
                for action in session.actions:
                    if action.action_type == UserActionType.COMMAND_EXECUTION:
                        if "command" in action.action_data:
                            all_commands.append(action.action_data["command"])
                        if "tool" in action.action_data:
                            all_tools.append(action.action_data["tool"])

            patterns["most_used_commands"] = dict(Counter(all_commands).most_common(10))
            patterns["most_used_tools"] = dict(Counter(all_tools).most_common(10))

            # Peak productivity hours
            hour_productivity = defaultdict(list)

            for session in user_sessions:
                hour_productivity[session.start_time.hour].append(session.productivity_score)

            avg_hourly_productivity = {
                hour: statistics.mean(scores) if scores else 0 for hour, scores in hour_productivity.items()
            }

            # Top 3 most productive hours
            top_hours = sorted(avg_hourly_productivity.items(), key=lambda x: x[1], reverse=True)[:3]
            patterns["peak_productivity_hours"] = [f"{hour:02d}:00" for hour, _ in top_hours]

        # Cache results
        self._pattern_cache[cache_key] = patterns

        return patterns

    def get_user_insights(self, user_id: str, days: int = 7) -> Dict[str, Any]:
        """Get actionable insights and recommendations for a user"""
        cache_key = f"insights_{user_id}_{days}"

        if cache_key in self._insight_cache:
            return self._insight_cache[cache_key]

        patterns = self.get_user_patterns(user_id, days)

        insights: Dict[str, Any] = {
            "productivity_insights": [],
            "efficiency_recommendations": [],
            "collaboration_insights": [],
            "learning_opportunities": [],
            "tool_recommendations": [],
            "workflow_optimizations": [],
        }

        # Productivity insights
        if patterns["avg_productivity_score"] > 80:
            insights["productivity_insights"].append(
                "Excellent productivity score! User consistently performs at high level."
            )
        elif patterns["avg_productivity_score"] < 50:
            insights["productivity_insights"].append(
                "Productivity score could be improved. Consider workflow optimization."
            )

        # Session duration insights
        if patterns["avg_session_duration"] > 1800000:  # > 30 minutes
            insights["productivity_insights"].append(
                "Long session durations detected. Consider taking regular breaks for sustained productivity."
            )

        # Error pattern analysis
        recent_actions = [
            action
            for action in self.action_history
            if action.user_id == user_id and action.timestamp >= datetime.now() - timedelta(days=days)
        ]

        error_actions = [
            action
            for action in recent_actions
            if action.action_type == UserActionType.ERROR_OCCURRED or not action.success
        ]

        if len(error_actions) > len(recent_actions) * 0.1:  # > 10% error rate
            insights["efficiency_recommendations"].append(
                "High error rate detected. Consider additional training or tool familiarization."
            )

        # Tool usage insights
        if patterns["most_used_tools"]:
            top_tool = max(patterns["most_used_tools"].items(), key=lambda x: x[1])
            insights["tool_recommendations"].append(f"Most frequently used tool: {top_tool[0]} ({top_tool[1]} uses)")

        # Command pattern insights
        if patterns["most_used_commands"]:
            commands = list(patterns["most_used_commands"].keys())
            if "/moai:" in " ".join(commands[:3]):  # Top 3 commands
                insights["workflow_optimizations"].append(
                    "Heavy MoAI command usage detected. Consider learning keyboard shortcuts for faster workflow."
                )

        # Peak productivity insights
        if patterns["peak_productivity_hours"]:
            peak_hours = ", ".join(patterns["peak_productivity_hours"])
            insights["efficiency_recommendations"].append(
                f"Peak productivity hours: {peak_hours}. Schedule important work during these times."
            )

        # Learning opportunities
        if patterns["session_count"] < 5:
            insights["learning_opportunities"].append(
                "Consider more frequent sessions to build momentum and consistency."
            )

        # Cache results
        self._insight_cache[cache_key] = insights

        return insights

    def get_team_analytics(self, days: int = 30) -> Dict[str, Any]:
        """Get team-wide analytics and collaboration patterns"""
        cutoff_date = datetime.now() - timedelta(days=days)

        # Get all sessions in time range
        recent_sessions = [session for session in self.user_sessions.values() if session.start_time >= cutoff_date]

        # Aggregate metrics
        team_metrics = {
            "total_sessions": len(recent_sessions),
            "unique_users": len(set(s.user_id for s in recent_sessions)),
            "avg_session_duration": 0.0,
            "avg_productivity_score": 0.0,
            "most_active_users": {},
            "most_used_tools": {},
            "collaboration_patterns": {},
            "productivity_distribution": {},
            "peak_hours": {},
        }

        if recent_sessions:
            # Calculate averages
            durations = [s.total_duration_ms for s in recent_sessions if s.total_duration_ms > 0]
            scores = [s.productivity_score for s in recent_sessions if s.productivity_score > 0]

            if durations:
                team_metrics["avg_session_duration"] = statistics.mean(durations)
            if scores:
                team_metrics["avg_productivity_score"] = statistics.mean(scores)

            # Most active users
            user_session_counts = Counter(s.user_id for s in recent_sessions)
            team_metrics["most_active_users"] = dict(user_session_counts.most_common(10))

            # Most used tools across team
            all_tools: List[str] = []
            for session in recent_sessions:
                all_tools.extend(session.tools_used)

            team_metrics["most_used_tools"] = dict(Counter(all_tools).most_common(10))

            # Productivity distribution
            productivity_ranges = {
                "0-20": 0,
                "21-40": 0,
                "41-60": 0,
                "61-80": 0,
                "81-100": 0,
            }

            for score in scores:
                if score <= 20:
                    productivity_ranges["0-20"] += 1
                elif score <= 40:
                    productivity_ranges["21-40"] += 1
                elif score <= 60:
                    productivity_ranges["41-60"] += 1
                elif score <= 80:
                    productivity_ranges["61-80"] += 1
                else:
                    productivity_ranges["81-100"] += 1

            team_metrics["productivity_distribution"] = productivity_ranges

        return team_metrics

    def _extract_action_tags(self, action_type: UserActionType, action_data: Dict[str, Any]) -> Set[str]:
        """Extract relevant tags from action data"""
        tags = {action_type.value}

        # Extract tool tags
        if "tool" in action_data:
            tags.add(f"tool:{action_data['tool']}")

        # Extract command tags
        if "command" in action_data:
            command = action_data["command"]
            if "/moai:" in command:
                tags.add("moai_command")
            if "git" in command.lower():
                tags.add("git_operation")

        # Extract phase tags
        if "phase" in action_data:
            tags.add(f"phase:{action_data['phase']}")

        # Extract success/failure tags
        if "success" in action_data:
            tags.add("success" if action_data["success"] else "failure")

        return tags

    def _update_user_preferences(self, user_id: str, action: UserAction) -> None:
        """Update learned user preferences based on actions"""
        if user_id not in self.user_preferences:
            self.user_preferences[user_id] = UserPreferences(user_id=user_id)

        prefs = self.user_preferences[user_id]

        # Update command preferences
        if action.action_type == UserActionType.COMMAND_EXECUTION:
            command = action.action_data.get("command", "")
            if command:
                prefs.preferred_commands[command] = prefs.preferred_commands.get(command, 0) + 1

        # Update tool preferences
        if "tool" in action.action_data:
            tool = action.action_data["tool"]
            prefs.preferred_tools[tool] = prefs.preferred_tools.get(tool, 0) + 1

        # Update working hours
        hour = action.timestamp.hour
        prefs.working_hours[str(hour)] = prefs.working_hours.get(str(hour), 0) + 1

        # Update patterns based on success/failure
        if action.success:
            pattern_key = f"{action.action_type.value}_{action.action_data}"
            if pattern_key not in prefs.success_patterns:
                prefs.success_patterns.append(pattern_key)
        else:
            pattern_key = f"{action.action_type.value}_{action.action_data}"
            if pattern_key not in prefs.error_patterns:
                prefs.error_patterns.append(pattern_key)

        prefs.last_updated = datetime.now()

    def _analyze_session_state(self, session: UserSession) -> SessionState:
        """Analyze and determine current session state"""
        if not session.actions:
            return SessionState.ACTIVE

        recent_actions = [
            action
            for action in session.actions[-10:]  # Last 10 actions
            if (datetime.now() - action.timestamp).seconds < 300  # Last 5 minutes
        ]

        if not recent_actions:
            return SessionState.IDLE

        # Calculate metrics
        error_rate = sum(1 for action in recent_actions if not action.success) / len(recent_actions)

        durations_with_ms = [a.duration_ms for a in recent_actions if a.duration_ms is not None]
        avg_duration = statistics.mean(durations_with_ms) if durations_with_ms else 0

        # Determine state
        if error_rate > 0.5:
            return SessionState.STRUGGLING
        elif avg_duration > 10000 and error_rate < 0.1:
            return SessionState.PRODUCTIVE
        elif error_rate < 0.1 and len(recent_actions) > 5:
            return SessionState.FOCUSED
        else:
            return SessionState.ACTIVE

    def _calculate_productivity_score(self, session: UserSession) -> float:
        """Calculate productivity score for a session (0-100)"""
        if not session.actions:
            return 0.0

        # Base score from success rate
        successful_actions = sum(1 for action in session.actions if action.success)
        success_rate = successful_actions / len(session.actions)
        base_score = success_rate * 50

        # Duration factor (optimal session duration)
        duration_hours = session.total_duration_ms / (1000 * 60 * 60)
        duration_factor = 0

        if 0.5 <= duration_hours <= 4:  # Optimal 30 min to 4 hours
            duration_factor = 25
        elif duration_hours > 0.25:  # At least 15 minutes
            duration_factor = 15

        # Tool diversity factor
        tool_diversity = min(len(session.tools_used) / 10, 1) * 15

        # File modifications factor (if it's a coding session)
        file_factor: float = 0.0
        if session.modified_files:
            file_factor = min(len(session.modified_files) / 5, 1) * 10

        total_score = base_score + duration_factor + tool_diversity + file_factor
        return min(total_score, 100.0)

    def _load_data(self) -> None:
        """Load stored analytics data"""
        try:
            # Load user preferences
            prefs_file = self.storage_path / "user_preferences.json"
            if prefs_file.exists():
                with open(prefs_file, "r") as f:
                    prefs_data = json.load(f)

                for user_id, prefs_dict in prefs_data.items():
                    prefs = UserPreferences(user_id=user_id)
                    prefs.preferred_commands = prefs_dict.get("preferred_commands", {})
                    prefs.preferred_tools = prefs_dict.get("preferred_tools", {})
                    prefs.working_hours = prefs_dict.get("working_hours", {})
                    prefs.peak_productivity_times = prefs_dict.get("peak_productivity_times", [])
                    prefs.common_workflows = prefs_dict.get("common_workflows", [])
                    prefs.error_patterns = prefs_dict.get("error_patterns", [])
                    prefs.success_patterns = prefs_dict.get("success_patterns", [])
                    prefs.collaboration_patterns = prefs_dict.get("collaboration_patterns", {})

                    if prefs_dict.get("last_updated"):
                        prefs.last_updated = datetime.fromisoformat(prefs_dict["last_updated"])

                    self.user_preferences[user_id] = prefs

            logger.info(f"Loaded preferences for {len(self.user_preferences)} users")

        except Exception as e:
            logger.error(f"Error loading analytics data: {e}")

    def _save_data(self) -> None:
        """Save analytics data to storage"""
        try:
            # Save user preferences
            prefs_data = {}
            for user_id, prefs in self.user_preferences.items():
                prefs_data[user_id] = {
                    "preferred_commands": prefs.preferred_commands,
                    "preferred_tools": prefs.preferred_tools,
                    "working_hours": prefs.working_hours,
                    "peak_productivity_times": prefs.peak_productivity_times,
                    "common_workflows": prefs.common_workflows,
                    "error_patterns": prefs.error_patterns,
                    "success_patterns": prefs.success_patterns,
                    "collaboration_patterns": prefs.collaboration_patterns,
                    "last_updated": prefs.last_updated.isoformat(),
                }

            prefs_file = self.storage_path / "user_preferences.json"
            with open(prefs_file, "w") as f:
                json.dump(prefs_data, f, indent=2)

            logger.info(f"Saved preferences for {len(self.user_preferences)} users")

        except Exception as e:
            logger.error(f"Error saving analytics data: {e}")

    def export_data(self, output_path: Path, user_id: Optional[str] = None) -> bool:
        """Export analytics data to file"""
        try:
            export_data = {
                "export_timestamp": datetime.now().isoformat(),
                "user_id": user_id or "all_users",
                "data": {},
            }

            if user_id:
                # Export single user data
                patterns = self.get_user_patterns(user_id)
                insights = self.get_user_insights(user_id)
                prefs = self.user_preferences.get(user_id)

                export_data["data"] = {
                    "patterns": patterns,
                    "insights": insights,
                    "preferences": asdict(prefs) if prefs else None,
                }
            else:
                # Export team data
                export_data["data"] = {
                    "team_analytics": self.get_team_analytics(),
                    "user_count": len(self.user_preferences),
                    "session_count": len(self.user_sessions),
                }

            with open(output_path, "w") as f:
                json.dump(export_data, f, indent=2)

            return True

        except Exception as e:
            logger.error(f"Error exporting data: {e}")
            return False

    def get_realtime_metrics(self, user_id: Optional[str] = None) -> Dict[str, Any]:
        """Get real-time analytics metrics"""
        current_time = datetime.now()

        metrics: Dict[str, Any] = {
            "timestamp": current_time.isoformat(),
            "active_sessions": len(self.active_sessions),
            "total_sessions_today": 0,
            "avg_session_duration": 0.0,
            "current_productivity_scores": [],
        }

        today_start = current_time.replace(hour=0, minute=0, second=0, microsecond=0)

        # Filter sessions and calculate metrics
        today_sessions = [session for session in self.user_sessions.values() if session.start_time >= today_start]

        metrics["total_sessions_today"] = len(today_sessions)

        if today_sessions:
            durations = [s.total_duration_ms for s in today_sessions if s.total_duration_ms > 0]
            if durations:
                metrics["avg_session_duration"] = statistics.mean(durations)

        # Current productivity scores
        productivity_scores: List[Dict[str, Any]] = []
        for session in self.active_sessions.values():
            if user_id is None or session.user_id == user_id:
                current_score = self._calculate_productivity_score(session)
                productivity_scores.append(
                    {
                        "user_id": session.user_id,
                        "session_id": session.session_id,
                        "score": current_score,
                        "state": session.state.value,
                    }
                )
        metrics["current_productivity_scores"] = productivity_scores

        return metrics


# Global instance for easy access
_user_analytics: Optional[UserBehaviorAnalytics] = None


def get_user_analytics() -> UserBehaviorAnalytics:
    """Get or create global user behavior analytics instance"""
    global _user_analytics
    if _user_analytics is None:
        _user_analytics = UserBehaviorAnalytics()
    return _user_analytics


# Convenience functions
def track_user_action(
    action_type: UserActionType,
    user_id: str,
    session_id: str,
    action_data: Dict[str, Any],
    context: Optional[Dict[str, Any]] = None,
    duration_ms: Optional[float] = None,
    success: bool = True,
) -> None:
    """Track a user action"""
    analytics = get_user_analytics()
    analytics.track_action(action_type, user_id, session_id, action_data, context, duration_ms, success)


def start_user_session(user_id: str, working_directory: str = "", git_branch: str = "") -> str:
    """Start tracking a user session"""
    analytics = get_user_analytics()
    return analytics.start_session(user_id, working_directory, git_branch)


def end_user_session(session_id: str) -> Optional[UserSession]:
    """End a user session"""
    analytics = get_user_analytics()
    return analytics.end_session(session_id)


def get_user_patterns(user_id: str, days: int = 30) -> Dict[str, Any]:
    """Get user behavior patterns"""
    analytics = get_user_analytics()
    return analytics.get_user_patterns(user_id, days)


def get_user_insights(user_id: str, days: int = 7) -> Dict[str, Any]:
    """Get user insights and recommendations"""
    analytics = get_user_analytics()
    return analytics.get_user_insights(user_id, days)


if __name__ == "__main__":
    # Example usage
    print("Testing User Behavior Analytics...")

    analytics = UserBehaviorAnalytics()

    # Simulate user session
    user_id = "test_user"
    session_id = analytics.start_session(user_id, "/Users/goos/MoAI/MoAI-ADK", "main")

    # Track some actions
    analytics.track_action(
        UserActionType.COMMAND_EXECUTION,
        user_id,
        session_id,
        {"command": "/moai:1-plan test feature", "tool": "spec_builder"},
        duration_ms=1500,
        success=True,
    )

    analytics.track_action(
        UserActionType.TOOL_USAGE,
        user_id,
        session_id,
        {"tool": "git", "operation": "commit"},
        duration_ms=800,
        success=True,
    )

    # End session
    ended_session = analytics.end_session(session_id)

    if ended_session:
        print("Session completed:")
        print(f"  Duration: {ended_session.total_duration_ms / 1000:.1f} seconds")
        print(f"  Commands: {ended_session.total_commands}")
        print(f"  Productivity Score: {ended_session.productivity_score}")
        print(f"  Tools Used: {list(ended_session.tools_used)}")

    # Get patterns
    patterns = analytics.get_user_patterns(user_id, days=1)
    print(f"User patterns: {patterns}")

    print("User Behavior Analytics test completed!")
