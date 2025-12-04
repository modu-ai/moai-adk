"""Simple tests for user_behavior_analytics module.

Tests basic analytics functions, session tracking, and metrics
with mocked file operations.
"""

import unittest
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

from src.moai_adk.core.user_behavior_analytics import (
    SessionState,
    UserAction,
    UserActionType,
    UserBehaviorAnalytics,
    UserPreferences,
    UserSession,
)


class TestUserAction(unittest.TestCase):
    """Test UserAction dataclass."""

    def test_user_action_creation(self):
        """Test creating a UserAction instance."""
        # Arrange
        now = datetime.now()
        action = UserAction(
            timestamp=now,
            action_type=UserActionType.COMMAND_EXECUTION,
            user_id="user_001",
            session_id="session_001",
            action_data={"command": "/moai:1-plan", "tool": "spec_builder"},
            duration_ms=1500.0,
            success=True,
        )

        # Act
        result = action.to_dict()

        # Assert
        self.assertEqual(result["user_id"], "user_001")
        self.assertEqual(result["action_type"], "command_execution")
        self.assertEqual(result["duration_ms"], 1500.0)
        self.assertTrue(result["success"])

    def test_user_action_with_tags(self):
        """Test UserAction with tags."""
        # Arrange
        now = datetime.now()
        action = UserAction(
            timestamp=now,
            action_type=UserActionType.TOOL_USAGE,
            user_id="user_001",
            session_id="session_001",
            action_data={"tool": "git"},
            tags={"git_operation", "important"},
        )

        # Act
        result = action.to_dict()

        # Assert
        self.assertEqual(len(result["tags"]), 2)
        self.assertIn("git_operation", result["tags"])


class TestUserSession(unittest.TestCase):
    """Test UserSession dataclass."""

    def test_user_session_creation(self):
        """Test creating a UserSession instance."""
        # Arrange
        now = datetime.now()
        session = UserSession(
            session_id="session_001",
            user_id="user_001",
            start_time=now,
            working_directory="/Users/goos/MoAI/MoAI-ADK",
            git_branch="main",
        )

        # Act
        result = session.to_dict()

        # Assert
        self.assertEqual(result["session_id"], "session_001")
        self.assertEqual(result["user_id"], "user_001")
        self.assertEqual(result["state"], "active")
        self.assertEqual(result["productivity_score"], 0.0)

    def test_user_session_with_end_time(self):
        """Test UserSession with end time."""
        # Arrange
        now = datetime.now()
        end_time = now + timedelta(hours=1)
        session = UserSession(
            session_id="session_001",
            user_id="user_001",
            start_time=now,
            end_time=end_time,
            state=SessionState.PRODUCTIVE,
            productivity_score=85.0,
            total_commands=10,
            total_errors=1,
        )

        # Act
        result = session.to_dict()

        # Assert
        self.assertIsNotNone(result["end_time"])
        self.assertEqual(result["state"], "productive")
        self.assertEqual(result["productivity_score"], 85.0)


class TestUserPreferences(unittest.TestCase):
    """Test UserPreferences dataclass."""

    def test_user_preferences_creation(self):
        """Test creating a UserPreferences instance."""
        # Arrange
        prefs = UserPreferences(
            user_id="user_001",
            preferred_commands={"/moai:1-plan": 5, "/moai:2-run": 3},
            preferred_tools={"git": 10, "pytest": 8},
        )

        # Act
        # Assert
        self.assertEqual(prefs.user_id, "user_001")
        self.assertEqual(len(prefs.preferred_commands), 2)
        self.assertEqual(len(prefs.preferred_tools), 2)

    def test_user_preferences_defaults(self):
        """Test UserPreferences with default values."""
        # Arrange
        prefs = UserPreferences(user_id="user_002")

        # Act
        # Assert
        self.assertEqual(len(prefs.preferred_commands), 0)
        self.assertEqual(len(prefs.preferred_tools), 0)
        self.assertIsNotNone(prefs.last_updated)


class TestUserBehaviorAnalyticsInit(unittest.TestCase):
    """Test UserBehaviorAnalytics initialization."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_analytics_initialization(self, mock_mkdir, mock_load_data):
        """Test UserBehaviorAnalytics initialization."""
        # Arrange
        # Act
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Assert
        self.assertIsNotNone(analytics.user_sessions)
        self.assertIsNotNone(analytics.active_sessions)
        self.assertEqual(len(analytics.user_sessions), 0)
        self.assertEqual(len(analytics.active_sessions), 0)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_analytics_action_history(self, mock_mkdir, mock_load_data):
        """Test action history initialization."""
        # Arrange
        # Act
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Assert
        self.assertEqual(len(analytics.action_history), 0)
        self.assertEqual(analytics.action_history.maxlen, 10000)


class TestSessionTracking(unittest.TestCase):
    """Test session tracking functionality."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_start_session(self, mock_mkdir, mock_load_data):
        """Test starting a new session."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Act
        session_id = analytics.start_session(
            "user_001", "/Users/goos/MoAI/MoAI-ADK", "main"
        )

        # Assert
        self.assertIsNotNone(session_id)
        self.assertIn(session_id, analytics.active_sessions)
        self.assertIn(session_id, analytics.user_sessions)
        session = analytics.active_sessions[session_id]
        self.assertEqual(session.user_id, "user_001")
        self.assertEqual(session.state, SessionState.ACTIVE)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._save_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_end_session(self, mock_mkdir, mock_load_data, mock_save_data):
        """Test ending a session."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session_id = analytics.start_session("user_001")

        # Act
        ended_session = analytics.end_session(session_id)

        # Assert
        self.assertIsNotNone(ended_session)
        self.assertNotIn(session_id, analytics.active_sessions)
        self.assertIsNotNone(ended_session.end_time)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_end_nonexistent_session(self, mock_mkdir, mock_load_data):
        """Test ending a nonexistent session."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Act
        result = analytics.end_session("nonexistent_session")

        # Assert
        self.assertIsNone(result)


class TestActionTracking(unittest.TestCase):
    """Test action tracking functionality."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_track_action(self, mock_mkdir, mock_update_prefs, mock_load_data):
        """Test tracking a user action."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session_id = analytics.start_session("user_001")

        # Act
        analytics.track_action(
            UserActionType.COMMAND_EXECUTION,
            "user_001",
            session_id,
            {"command": "/moai:1-plan", "tool": "spec_builder"},
            duration_ms=1500.0,
        )

        # Assert
        self.assertEqual(
            len(analytics.action_history), 2
        )  # SESSION_START + COMMAND_EXECUTION

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_track_multiple_actions(
        self, mock_mkdir, mock_update_prefs, mock_load_data
    ):
        """Test tracking multiple actions in a session."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session_id = analytics.start_session("user_001")

        # Act
        analytics.track_action(
            UserActionType.COMMAND_EXECUTION,
            "user_001",
            session_id,
            {"command": "/moai:1-plan", "tool": "spec_builder"},
        )
        analytics.track_action(
            UserActionType.TOOL_USAGE,
            "user_001",
            session_id,
            {"tool": "git"},
        )
        analytics.track_action(
            UserActionType.FILE_OPERATION,
            "user_001",
            session_id,
            {"files": ["test.py", "config.json"]},
        )

        # Assert
        session = analytics.active_sessions[session_id]
        self.assertEqual(len(session.actions), 4)  # SESSION_START + 3 actions

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_track_action_with_error(
        self, mock_mkdir, mock_update_prefs, mock_load_data
    ):
        """Test tracking a failed action."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session_id = analytics.start_session("user_001")

        # Act
        analytics.track_action(
            UserActionType.COMMAND_EXECUTION,
            "user_001",
            session_id,
            {"command": "/moai:1-plan"},
            success=False,
        )

        # Assert
        session = analytics.active_sessions[session_id]
        self.assertEqual(session.total_errors, 1)


class TestPatternAnalysis(unittest.TestCase):
    """Test user pattern analysis."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_user_patterns_empty(self, mock_mkdir, mock_load_data):
        """Test getting patterns with no sessions."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Act
        patterns = analytics.get_user_patterns("user_001")

        # Assert
        self.assertEqual(patterns["session_count"], 0)
        self.assertEqual(patterns["avg_session_duration"], 0.0)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._save_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_user_patterns_with_sessions(
        self, mock_mkdir, mock_save_data, mock_update_prefs, mock_load_data
    ):
        """Test getting patterns with existing sessions."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session_id = analytics.start_session("user_001")

        # Add some actions
        analytics.track_action(
            UserActionType.COMMAND_EXECUTION,
            "user_001",
            session_id,
            {"command": "/moai:1-plan"},
        )

        analytics.end_session(session_id)

        # Act
        patterns = analytics.get_user_patterns("user_001", days=30)

        # Assert
        self.assertGreater(patterns["session_count"], 0)


class TestInsights(unittest.TestCase):
    """Test user insights generation."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_user_insights_empty(self, mock_mkdir, mock_load_data):
        """Test getting insights with no data."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Act
        insights = analytics.get_user_insights("user_001")

        # Assert
        self.assertIn("productivity_insights", insights)
        self.assertIn("efficiency_recommendations", insights)
        self.assertIn("tool_recommendations", insights)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._save_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_user_insights_with_data(
        self, mock_mkdir, mock_save_data, mock_update_prefs, mock_load_data
    ):
        """Test getting insights with user data."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session_id = analytics.start_session("user_001")

        for i in range(5):
            analytics.track_action(
                UserActionType.COMMAND_EXECUTION,
                "user_001",
                session_id,
                {"command": "/moai:1-plan", "tool": "spec_builder"},
                success=True,
            )

        analytics.end_session(session_id)

        # Act
        insights = analytics.get_user_insights("user_001", days=7)

        # Assert
        self.assertIsInstance(insights, dict)
        self.assertIn("productivity_insights", insights)


class TestTeamAnalytics(unittest.TestCase):
    """Test team-wide analytics."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_team_analytics_empty(self, mock_mkdir, mock_load_data):
        """Test getting team analytics with no data."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Act
        team_metrics = analytics.get_team_analytics()

        # Assert
        self.assertEqual(team_metrics["total_sessions"], 0)
        self.assertEqual(team_metrics["unique_users"], 0)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._save_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_team_analytics_multiple_users(
        self, mock_mkdir, mock_save_data, mock_update_prefs, mock_load_data
    ):
        """Test getting team analytics with multiple users."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Create sessions for multiple users
        for user_id in ["user_001", "user_002"]:
            session_id = analytics.start_session(user_id)
            analytics.track_action(
                UserActionType.COMMAND_EXECUTION,
                user_id,
                session_id,
                {"command": "/moai:1-plan"},
            )
            analytics.end_session(session_id)

        # Act
        team_metrics = analytics.get_team_analytics(days=30)

        # Assert
        self.assertEqual(team_metrics["unique_users"], 2)


class TestProductivityScore(unittest.TestCase):
    """Test productivity score calculation."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_calculate_productivity_score_no_actions(
        self, mock_mkdir, mock_update_prefs, mock_load_data
    ):
        """Test productivity score with no actions."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session = UserSession("session_001", "user_001", datetime.now())

        # Act
        score = analytics._calculate_productivity_score(session)

        # Assert
        self.assertEqual(score, 0.0)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_calculate_productivity_score_with_success(
        self, mock_mkdir, mock_update_prefs, mock_load_data
    ):
        """Test productivity score with successful actions."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session = UserSession("session_001", "user_001", datetime.now())

        action = UserAction(
            timestamp=datetime.now(),
            action_type=UserActionType.COMMAND_EXECUTION,
            user_id="user_001",
            session_id="session_001",
            action_data={"command": "/moai:1-plan"},
            success=True,
        )
        session.actions.append(action)
        session.total_duration_ms = 2000000  # 33 minutes (optimal range)

        # Act
        score = analytics._calculate_productivity_score(session)

        # Assert
        self.assertGreater(score, 0)
        self.assertLessEqual(score, 100.0)


class TestSessionState(unittest.TestCase):
    """Test session state analysis."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_analyze_session_state_no_actions(self, mock_mkdir, mock_load_data):
        """Test analyzing session state with no actions."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session = UserSession("session_001", "user_001", datetime.now())

        # Act
        state = analytics._analyze_session_state(session)

        # Assert
        self.assertEqual(state, SessionState.ACTIVE)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_analyze_session_state_productive(self, mock_mkdir, mock_load_data):
        """Test analyzing session state for productive session."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session = UserSession("session_001", "user_001", datetime.now())

        # Add successful actions with good duration
        for i in range(6):
            action = UserAction(
                timestamp=datetime.now() - timedelta(minutes=i),
                action_type=UserActionType.COMMAND_EXECUTION,
                user_id="user_001",
                session_id="session_001",
                action_data={"command": "/moai:1-plan"},
                duration_ms=15000,
                success=True,
            )
            session.actions.append(action)

        # Act
        state = analytics._analyze_session_state(session)

        # Assert
        self.assertEqual(state, SessionState.PRODUCTIVE)


class TestRealTimeMetrics(unittest.TestCase):
    """Test real-time metrics."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_realtime_metrics_empty(self, mock_mkdir, mock_load_data):
        """Test getting real-time metrics with no active sessions."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))

        # Act
        metrics = analytics.get_realtime_metrics()

        # Assert
        self.assertEqual(metrics["active_sessions"], 0)
        self.assertEqual(metrics["total_sessions_today"], 0)

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch(
        "src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._update_user_preferences"
    )
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_get_realtime_metrics_with_session(
        self, mock_mkdir, mock_update_prefs, mock_load_data
    ):
        """Test getting real-time metrics with active session."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        session_id = analytics.start_session("user_001")

        # Act
        metrics = analytics.get_realtime_metrics()

        # Assert
        self.assertGreater(metrics["active_sessions"], 0)
        self.assertGreater(len(metrics["current_productivity_scores"]), 0)


class TestExportData(unittest.TestCase):
    """Test data export functionality."""

    @patch("src.moai_adk.core.user_behavior_analytics.UserBehaviorAnalytics._load_data")
    @patch("builtins.open", create=True)
    @patch("src.moai_adk.core.user_behavior_analytics.Path.mkdir")
    def test_export_data(self, mock_mkdir, mock_open_file, mock_load_data):
        """Test exporting analytics data."""
        # Arrange
        analytics = UserBehaviorAnalytics(storage_path=Path("/test/analytics"))
        mock_file = MagicMock()
        mock_open_file.return_value.__enter__.return_value = mock_file

        # Act
        result = analytics.export_data(Path("/test/export.json"), user_id="user_001")

        # Assert
        self.assertTrue(result)


if __name__ == "__main__":
    unittest.main()
