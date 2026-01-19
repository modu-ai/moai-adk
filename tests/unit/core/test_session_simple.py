"""Comprehensive tests for moai_adk.core.session_manager module.

Tests SessionManager class with full coverage of:
- Session initialization and loading
- Agent result registration
- Resume logic and decision making
- Chain management
- Session persistence
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.session_manager import (
    SessionManager,
    get_resume_id,
    register_agent,
    should_resume,
)


class TestSessionManagerInitialization:
    """Test SessionManager initialization and setup."""

    def test_init_with_default_paths(self):
        """Test SessionManager initializes with default paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.session_manager.Path.cwd") as mock_cwd:
                mock_cwd.return_value = Path(tmpdir)
                manager = SessionManager()
                assert manager is not None
                assert manager._sessions == {}
                assert manager._results == {}
                assert manager._chains == {}

    def test_init_with_custom_paths(self):
        """Test SessionManager initializes with custom paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            transcript_dir = Path(tmpdir) / "transcripts"

            manager = SessionManager(session_file=session_file, transcript_dir=transcript_dir)

            assert manager._session_file == session_file
            assert manager._transcript_dir == transcript_dir
            assert transcript_dir.exists()

    def test_init_creates_directories(self):
        """Test SessionManager creates required directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "subdir" / "sessions.json"

            manager = SessionManager(session_file=session_file)

            assert session_file.parent.exists()
            assert manager._transcript_dir.exists()


class TestSessionLoading:
    """Test loading existing sessions from file."""

    def test_load_sessions_from_existing_file(self):
        """Test loading sessions from existing JSON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"

            # Create test data
            test_data = {
                "sessions": {"agent1": "id123"},
                "chains": {"chain1": ["id123"]},
                "metadata": {"id123": {"agent_name": "agent1"}},
            }

            with open(session_file, "w") as f:
                json.dump(test_data, f)

            manager = SessionManager(session_file=session_file)

            assert manager._sessions == {"agent1": "id123"}
            assert "chain1" in manager._chains
            assert "id123" in manager._metadata

    def test_load_sessions_with_invalid_json(self):
        """Test loading sessions handles invalid JSON gracefully."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"

            # Write invalid JSON
            with open(session_file, "w") as f:
                f.write("{ invalid json")

            with patch("moai_adk.core.session_manager.logger") as mock_logger:
                manager = SessionManager(session_file=session_file)

                assert manager._sessions == {}
                assert manager._chains == {}
                mock_logger.warning.assert_called_once()

    def test_load_sessions_file_not_found(self):
        """Test loading sessions when file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "nonexistent.json"

            manager = SessionManager(session_file=session_file)

            assert manager._sessions == {}
            assert manager._chains == {}


class TestAgentRegistration:
    """Test registering agent execution results."""

    def test_register_agent_result(self):
        """Test registering an agent result."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"

            manager = SessionManager(session_file=session_file)

            result_data = {"status": "success", "files": ["test.py"]}
            manager.register_agent_result(
                agent_name="ddd-implementer",
                agent_id="agent-abc123",
                result=result_data,
                chain_id="SPEC-001-impl",
            )

            assert manager._sessions["ddd-implementer"] == "agent-abc123"
            assert manager._results["agent-abc123"]["result"] == result_data
            assert "SPEC-001-impl" in manager._chains
            assert "agent-abc123" in manager._chains["SPEC-001-impl"]

    def test_register_agent_without_chain(self):
        """Test registering agent result without chain ID."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(
                agent_name="backend-expert",
                agent_id="backend-123",
                result={"recommendation": "Use FastAPI"},
                chain_id=None,
            )

            assert manager._sessions["backend-expert"] == "backend-123"
            assert manager._results["backend-123"]["chain_id"] is None

    def test_register_multiple_agents_in_chain(self):
        """Test registering multiple agents in same chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            chain_id = "SPEC-AUTH-001-flow"

            # Register first agent
            manager.register_agent_result(
                agent_name="spec-builder",
                agent_id="spec-001",
                result={"spec_created": True},
                chain_id=chain_id,
            )

            # Register second agent in same chain
            manager.register_agent_result(
                agent_name="ddd-implementer",
                agent_id="impl-001",
                result={"code_written": True},
                chain_id=chain_id,
            )

            assert len(manager._chains[chain_id]) == 2
            assert manager._chains[chain_id] == ["spec-001", "impl-001"]


class TestResumeLogic:
    """Test resume decision and ID retrieval."""

    def test_get_resume_id_existing_session(self):
        """Test getting resume ID for existing session."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(
                agent_name="ddd-implementer",
                agent_id="tdd-xyz789",
                result={"phase": "1"},
                chain_id="SPEC-001-impl",
            )

            resume_id = manager.get_resume_id(agent_name="ddd-implementer", chain_id="SPEC-001-impl")

            assert resume_id == "tdd-xyz789"

    def test_get_resume_id_no_previous_session(self):
        """Test getting resume ID when no previous session exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            resume_id = manager.get_resume_id(agent_name="unknown-agent", chain_id="SPEC-001")

            assert resume_id is None

    def test_get_resume_id_chain_mismatch(self):
        """Test resume ID returns None for chain mismatch."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={}, chain_id="CHAIN-A")

            resume_id = manager.get_resume_id(agent_name="agent1", chain_id="CHAIN-B")

            assert resume_id is None

    def test_should_resume_no_previous_session(self):
        """Test should_resume returns False for new agent."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            result = manager.should_resume(agent_name="new-agent", current_task="Task 1", previous_task="Task 0")

            assert result is False

    def test_should_resume_no_previous_task(self):
        """Test should_resume returns False without previous task."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={})

            result = manager.should_resume(agent_name="agent1", current_task="Task 1", previous_task=None)

            assert result is False

    def test_should_resume_with_related_tasks(self):
        """Test should_resume returns True for related tasks."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="ddd-implementer", agent_id="tdd1", result={})

            result = manager.should_resume(
                agent_name="ddd-implementer",
                current_task="Implement user login endpoint",
                previous_task="Implement user registration endpoint",
            )

            assert result is True

    def test_should_resume_max_depth_exceeded(self):
        """Test should_resume returns False at max resume depth."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={})

            # Set resume count to max
            manager._metadata["id1"]["resume_count"] = 5

            result = manager.should_resume(agent_name="agent1", current_task="Task B", previous_task="Task A")

            assert result is False


class TestSessionPersistence:
    """Test saving and loading session data."""

    def test_save_sessions_to_file(self):
        """Test sessions are saved to persistent storage."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={"key": "value"})

            assert session_file.exists()

            with open(session_file) as f:
                data = json.load(f)
                assert "sessions" in data
                assert data["sessions"]["agent1"] == "id1"

    def test_increment_resume_count(self):
        """Test incrementing resume count."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={})

            initial_count = manager._metadata["id1"]["resume_count"]
            manager.increment_resume_count("id1")

            assert manager._metadata["id1"]["resume_count"] == initial_count + 1
            assert "last_resumed_at" in manager._metadata["id1"]


class TestChainManagement:
    """Test workflow chain management."""

    def test_create_chain(self):
        """Test creating a workflow chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.create_chain(
                chain_id="SPEC-001-workflow",
                agent_sequence=["spec-builder", "ddd-implementer", "quality-checker"],
                metadata={"spec_id": "SPEC-001"},
            )

            assert "SPEC-001-workflow" in manager._chains

    def test_get_chain_results(self):
        """Test retrieving all results in a chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            chain_id = "SPEC-001-impl"

            manager.register_agent_result(
                agent_name="agent1",
                agent_id="id1",
                result={"step": 1},
                chain_id=chain_id,
            )
            manager.register_agent_result(
                agent_name="agent2",
                agent_id="id2",
                result={"step": 2},
                chain_id=chain_id,
            )

            results = manager.get_chain_results(chain_id)

            assert len(results) == 2
            assert results[0]["agent_name"] == "agent1"
            assert results[1]["agent_name"] == "agent2"

    def test_get_chain_summary(self):
        """Test getting chain summary."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            chain_id = "SPEC-001-workflow"

            manager.register_agent_result(agent_name="spec-builder", agent_id="id1", result={}, chain_id=chain_id)
            manager.register_agent_result(
                agent_name="ddd-implementer",
                agent_id="id2",
                result={},
                chain_id=chain_id,
            )

            summary = manager.get_chain_summary(chain_id)

            assert summary["chain_id"] == chain_id
            assert summary["status"] == "completed"
            assert summary["agent_count"] == 2
            assert len(summary["agents"]) == 2

    def test_get_chain_summary_not_found(self):
        """Test chain summary for non-existent chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            summary = manager.get_chain_summary("NONEXISTENT")

            assert summary["status"] == "not_found"

    def test_clear_chain(self):
        """Test clearing all sessions in a chain."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            chain_id = "SPEC-001"

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={}, chain_id=chain_id)

            assert chain_id in manager._chains

            manager.clear_chain(chain_id)

            assert chain_id not in manager._chains
            assert "id1" not in manager._results


class TestSessionRetrieval:
    """Test retrieving session information."""

    def test_get_agent_result(self):
        """Test retrieving agent result."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            test_result = {"files": ["test.py"], "tests": 10}

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result=test_result)

            retrieved = manager.get_agent_result("id1")

            assert retrieved == test_result

    def test_get_agent_result_not_found(self):
        """Test retrieving non-existent agent result."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            result = manager.get_agent_result("nonexistent")

            assert result is None

    def test_get_all_sessions(self):
        """Test getting all active sessions."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={}, chain_id="chain1")
            manager.register_agent_result(agent_name="agent2", agent_id="id2", result={}, chain_id="chain2")

            sessions = manager.get_all_sessions()

            assert len(sessions["sessions"]) == 2
            assert len(sessions["chains"]) == 2
            assert sessions["total_results"] == 2


class TestSessionClearance:
    """Test clearing session data."""

    def test_clear_agent_session(self):
        """Test clearing specific agent session."""
        with tempfile.TemporaryDirectory() as tmpdir:
            session_file = Path(tmpdir) / "sessions.json"
            manager = SessionManager(session_file=session_file)

            manager.register_agent_result(agent_name="agent1", agent_id="id1", result={})

            assert "agent1" in manager._sessions

            manager.clear_agent_session("agent1")

            assert "agent1" not in manager._sessions
            assert "id1" not in manager._results


class TestTranscriptManagement:
    """Test transcript file management."""

    def test_export_transcript_exists(self):
        """Test exporting transcript when file exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            transcript_dir = Path(tmpdir) / "transcripts"
            transcript_dir.mkdir(parents=True)

            session_file = Path(tmpdir) / "sessions.json"

            manager = SessionManager(session_file=session_file, transcript_dir=transcript_dir)

            # Create a mock transcript file
            transcript_file = transcript_dir / "agent-abc123.jsonl"
            transcript_file.touch()

            result = manager.export_transcript("abc123")

            assert result == transcript_file

    def test_export_transcript_not_found(self):
        """Test exporting transcript when file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            transcript_dir = Path(tmpdir) / "transcripts"
            session_file = Path(tmpdir) / "sessions.json"

            manager = SessionManager(session_file=session_file, transcript_dir=transcript_dir)

            result = manager.export_transcript("nonexistent")

            assert result is None


class TestConvenienceFunctions:
    """Test module-level convenience functions."""

    def test_register_agent_convenience(self):
        """Test register_agent convenience function."""
        with tempfile.TemporaryDirectory():
            with patch("moai_adk.core.session_manager.get_session_manager") as mock_get:
                mock_manager = MagicMock()
                mock_get.return_value = mock_manager

                register_agent(
                    agent_name="test-agent",
                    agent_id="id123",
                    result={"data": "value"},
                    chain_id="chain1",
                )

                mock_manager.register_agent_result.assert_called_once_with(
                    "test-agent", "id123", {"data": "value"}, "chain1"
                )

    def test_get_resume_id_convenience(self):
        """Test get_resume_id convenience function."""
        with patch("moai_adk.core.session_manager.get_session_manager") as mock_get:
            mock_manager = MagicMock()
            mock_manager.get_resume_id.return_value = "resume123"
            mock_get.return_value = mock_manager

            result = get_resume_id("agent1", "chain1")

            assert result == "resume123"
            mock_manager.get_resume_id.assert_called_once_with("agent1", "chain1")

    def test_should_resume_convenience(self):
        """Test should_resume convenience function."""
        with patch("moai_adk.core.session_manager.get_session_manager") as mock_get:
            mock_manager = MagicMock()
            mock_manager.should_resume.return_value = True
            mock_get.return_value = mock_manager

            result = should_resume("agent1", "task1", "task0")

            assert result is True
            mock_manager.should_resume.assert_called_once_with("agent1", "task1", "task0")
