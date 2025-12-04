"""
Comprehensive tests for SessionManager module.

Tests session tracking, resume logic, and workflow chain management.
"""

import json
import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.session_manager import (
    SessionManager,
    get_resume_id,
    get_session_manager,
    register_agent,
    should_resume,
)


class TestSessionManagerInit:
    """Test SessionManager initialization."""

    def test_init_default_paths(self):
        """Test initialization with default paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.session_manager.Path.cwd", return_value=Path(tmpdir)):
                manager = SessionManager()
                assert manager._session_file.parent.exists()
                assert manager._transcript_dir.exists()

    def test_init_custom_paths(self):
        """Test initialization with custom paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            transcript_dir = Path(tmpdir) / "transcripts"

            manager = SessionManager(
                session_file=session_file,
                transcript_dir=transcript_dir,
            )

            assert manager._session_file == session_file
            assert manager._transcript_dir == transcript_dir
            assert transcript_dir.exists()

    def test_init_creates_directories(self):
        """Test that __init__ creates necessary directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "subdir" / "sessions.json"
            manager = SessionManager(session_file=session_file)

            assert session_file.parent.exists()
            assert manager._transcript_dir.exists()


class TestSessionManagerLoadSave:
    """Test session persistence."""

    def test_load_sessions_empty_file(self):
        """Test loading sessions from non-existent file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            assert manager._sessions == {}
            assert manager._chains == {}

    def test_load_sessions_existing_file(self):
        """Test loading sessions from existing file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            session_file.parent.mkdir(parents=True, exist_ok=True)

            data = {
                "sessions": {"agent1": "id123"},
                "chains": {"chain1": ["id123"]},
                "metadata": {"id123": {"resume_count": 0}},
            }

            with open(session_file, "w") as f:
                json.dump(data, f)

            manager = SessionManager(session_file=session_file)

            assert manager._sessions == {"agent1": "id123"}
            assert manager._chains == {"chain1": ["id123"]}

    def test_save_sessions(self):
        """Test saving sessions to file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager._sessions = {"agent1": "id123"}
            manager._chains = {"chain1": ["id123"]}
            manager._save_sessions()

            assert session_file.exists()

            with open(session_file, "r") as f:
                data = json.load(f)

            assert data["sessions"] == {"agent1": "id123"}
            assert data["chains"] == {"chain1": ["id123"]}

    def test_load_sessions_corrupted_json(self):
        """Test handling corrupted JSON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            session_file.parent.mkdir(parents=True, exist_ok=True)

            with open(session_file, "w") as f:
                f.write("invalid json {")

            manager = SessionManager(session_file=session_file)

            assert manager._sessions == {}
            assert manager._chains == {}


class TestRegisterAgentResult:
    """Test register_agent_result method."""

    def test_register_agent_result_basic(self):
        """Test basic agent result registration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
                chain_id="chain1",
            )

            assert manager._sessions["test-agent"] == "id123"
            assert "id123" in manager._results
            assert manager._results["id123"]["result"]["status"] == "success"

    def test_register_agent_result_multiple_agents(self):
        """Test registering multiple agent results."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="agent1",
                agent_id="id1",
                result={"data": "result1"},
            )
            manager.register_agent_result(
                agent_name="agent2",
                agent_id="id2",
                result={"data": "result2"},
            )

            assert len(manager._sessions) == 2
            assert len(manager._results) == 2

    def test_register_agent_result_creates_metadata(self):
        """Test that registration creates metadata entry."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            assert "id123" in manager._metadata
            assert manager._metadata["id123"]["agent_name"] == "test-agent"
            assert manager._metadata["id123"]["resume_count"] == 0

    def test_register_agent_result_persists_to_disk(self):
        """Test that registration persists to disk."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            assert session_file.exists()

            with open(session_file, "r") as f:
                data = json.load(f)

            assert "test-agent" in data["sessions"]


class TestGetResumeId:
    """Test get_resume_id method."""

    def test_get_resume_id_exists(self):
        """Test getting resume ID for existing session."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            resume_id = manager.get_resume_id("test-agent")
            assert resume_id == "id123"

    def test_get_resume_id_not_found(self):
        """Test getting resume ID for non-existent session."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            resume_id = manager.get_resume_id("nonexistent-agent")
            assert resume_id is None

    def test_get_resume_id_with_chain_validation(self):
        """Test resume ID validation with chain ID."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
                chain_id="chain1",
            )

            # Correct chain
            resume_id = manager.get_resume_id("test-agent", chain_id="chain1")
            assert resume_id == "id123"

            # Wrong chain
            resume_id = manager.get_resume_id("test-agent", chain_id="chain2")
            assert resume_id is None


class TestShouldResume:
    """Test should_resume method."""

    def test_should_resume_no_previous_session(self):
        """Test should_resume returns False with no previous session."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            result = manager.should_resume(
                agent_name="test-agent",
                current_task="Task A",
                previous_task="Task B",
            )

            assert result is False

    def test_should_resume_related_tasks(self):
        """Test should_resume returns True for related tasks."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            result = manager.should_resume(
                agent_name="test-agent",
                current_task="Implement login endpoint",
                previous_task="Implement registration endpoint",
            )

            assert result is True

    def test_should_resume_independent_tasks(self):
        """Test should_resume returns False for independent tasks."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            result = manager.should_resume(
                agent_name="test-agent",
                current_task="Deploy to production",
                previous_task="Fix database schema",
            )

            assert result is False

    def test_should_resume_max_depth_exceeded(self):
        """Test should_resume returns False when max resume depth exceeded."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            # Simulate max resume count
            manager._metadata["id123"]["resume_count"] = 5

            result = manager.should_resume(
                agent_name="test-agent",
                current_task="Implement feature",
                previous_task="Implement endpoint",
            )

            assert result is False


class TestIncrementResumeCount:
    """Test increment_resume_count method."""

    def test_increment_resume_count(self):
        """Test incrementing resume count."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            initial_count = manager._metadata["id123"]["resume_count"]
            manager.increment_resume_count("id123")

            assert manager._metadata["id123"]["resume_count"] == initial_count + 1

    def test_increment_resume_count_updates_timestamp(self):
        """Test that increment_resume_count updates timestamp."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            manager.increment_resume_count("id123")

            assert "last_resumed_at" in manager._metadata["id123"]


class TestGetAgentResult:
    """Test get_agent_result method."""

    def test_get_agent_result_exists(self):
        """Test retrieving existing agent result."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success", "files": ["file1.py"]},
            )

            result = manager.get_agent_result("id123")
            assert result["status"] == "success"
            assert "file1.py" in result["files"]

    def test_get_agent_result_not_found(self):
        """Test getting non-existent agent result."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            result = manager.get_agent_result("nonexistent")
            assert result is None


class TestChainResults:
    """Test chain result operations."""

    def test_get_chain_results(self):
        """Test retrieving all results in a chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="agent1",
                agent_id="id1",
                result={"status": "phase1"},
                chain_id="chain1",
            )
            manager.register_agent_result(
                agent_name="agent2",
                agent_id="id2",
                result={"status": "phase2"},
                chain_id="chain1",
            )

            results = manager.get_chain_results("chain1")
            assert len(results) == 2
            assert results[0]["agent_name"] == "agent1"
            assert results[1]["agent_name"] == "agent2"

    def test_get_chain_results_not_found(self):
        """Test getting results for non-existent chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            results = manager.get_chain_results("nonexistent")
            assert results == []

    def test_get_chain_summary(self):
        """Test getting chain summary."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="agent1",
                agent_id="id1",
                result={"status": "phase1"},
                chain_id="chain1",
            )
            manager.register_agent_result(
                agent_name="agent2",
                agent_id="id2",
                result={"status": "phase2"},
                chain_id="chain1",
            )

            summary = manager.get_chain_summary("chain1")
            assert summary["chain_id"] == "chain1"
            assert summary["agent_count"] == 2
            assert "agent1" in summary["agents"]
            assert summary["status"] == "completed"


class TestClearSessions:
    """Test session clearing methods."""

    def test_clear_agent_session(self):
        """Test clearing a single agent session."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="test-agent",
                agent_id="id123",
                result={"status": "success"},
            )

            assert "test-agent" in manager._sessions
            manager.clear_agent_session("test-agent")
            assert "test-agent" not in manager._sessions
            assert "id123" not in manager._results

    def test_clear_chain(self):
        """Test clearing all sessions in a chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="agent1",
                agent_id="id1",
                result={"status": "success"},
                chain_id="chain1",
            )
            manager.register_agent_result(
                agent_name="agent2",
                agent_id="id2",
                result={"status": "success"},
                chain_id="chain1",
            )

            assert len(manager._chains["chain1"]) == 2
            manager.clear_chain("chain1")
            assert "chain1" not in manager._chains


class TestGetAllSessions:
    """Test get_all_sessions method."""

    def test_get_all_sessions_empty(self):
        """Test getting sessions when none exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            sessions = manager.get_all_sessions()
            assert sessions["sessions"] == {}
            assert sessions["chains"] == []
            assert sessions["total_results"] == 0

    def test_get_all_sessions_populated(self):
        """Test getting sessions with data."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.register_agent_result(
                agent_name="agent1",
                agent_id="id1",
                result={"status": "success"},
                chain_id="chain1",
            )

            sessions = manager.get_all_sessions()
            assert "agent1" in sessions["sessions"]
            assert "chain1" in sessions["chains"]
            assert sessions["total_results"] == 1


class TestExportTranscript:
    """Test transcript export functionality."""

    def test_export_transcript_exists(self):
        """Test exporting transcript that exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            transcript_dir = Path(tmpdir) / "transcripts"
            transcript_dir.mkdir()
            transcript_file = transcript_dir / "agent-id123.jsonl"
            transcript_file.write_text("line1\nline2\n")

            manager = SessionManager(
                session_file=Path(tmpdir) / "sessions.json",
                transcript_dir=transcript_dir,
            )

            result = manager.export_transcript("id123")
            assert result == transcript_file

    def test_export_transcript_not_found(self):
        """Test exporting transcript that doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(
                session_file=Path(tmpdir) / "sessions.json",
                transcript_dir=Path(tmpdir) / "transcripts",
            )

            result = manager.export_transcript("nonexistent")
            assert result is None


class TestCreateChain:
    """Test create_chain method."""

    def test_create_chain(self):
        """Test creating a workflow chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.create_chain(
                chain_id="chain1",
                agent_sequence=["agent1", "agent2"],
                metadata={"type": "spec"},
            )

            assert "chain1" in manager._chains
            chains_file = manager._session_file.parent / "workflow-chains.json"
            assert chains_file.exists()

    def test_create_chain_persists(self):
        """Test that chain creation persists to disk."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = SessionManager(session_file=Path(tmpdir) / "sessions.json")

            manager.create_chain(
                chain_id="chain1",
                agent_sequence=["agent1", "agent2"],
            )

            chains_file = manager._session_file.parent / "workflow-chains.json"
            with open(chains_file, "r") as f:
                data = json.load(f)

            assert "chain1" in data


class TestGlobalSingleton:
    """Test global SessionManager singleton pattern."""

    def test_get_session_manager_returns_same_instance(self):
        """Test that get_session_manager returns the same instance."""
        # Clear the global instance
        import moai_adk.core.session_manager as sm_module

        sm_module._session_manager_instance = None

        manager1 = get_session_manager()
        manager2 = get_session_manager()

        assert manager1 is manager2

    def test_convenience_functions(self):
        """Test convenience functions."""
        import moai_adk.core.session_manager as sm_module

        sm_module._session_manager_instance = None

        register_agent("test-agent", "id123", {"status": "success"})
        resume_id = get_resume_id("test-agent")

        assert resume_id == "id123"

        should_resume_result = should_resume(
            "test-agent",
            "Task A",
            "Task B",
        )

        assert isinstance(should_resume_result, bool)
