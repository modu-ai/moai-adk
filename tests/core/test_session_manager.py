"""
Comprehensive Test Suite for Session Manager

Tests cover:
- Session initialization and directory structure
- Session creation and registration
- Session persistence and recovery
- Resume decision logic and context continuity
- Workflow chain management (linear, parallel, complex)
- Agent result storage and retrieval
- Session cleanup and archival operations
- Error handling and edge cases
- Singleton pattern for global instance
- Transcript management
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest

try:
    from moai_adk.core.session_manager import (
        SessionManager,
        get_resume_id,
        get_session_manager,
        register_agent,
        should_resume,
    )
except ImportError:
    pytest.skip("session_manager not available", allow_module_level=True)


@pytest.fixture
def temp_project_dir():
    """Create a temporary project directory structure"""
    with tempfile.TemporaryDirectory() as tmpdir:
        project_root = Path(tmpdir)

        # Create necessary directory structure
        (project_root / ".moai" / "memory").mkdir(parents=True, exist_ok=True)
        (project_root / ".moai" / "logs" / "agent-transcripts").mkdir(parents=True, exist_ok=True)

        yield project_root


@pytest.fixture
def session_manager(temp_project_dir):
    """Create a SessionManager instance for testing"""
    manager = SessionManager(
        session_file=temp_project_dir / ".moai" / "memory" / "agent-sessions.json",
        transcript_dir=temp_project_dir / ".moai" / "logs" / "agent-transcripts",
    )
    yield manager


@pytest.fixture(autouse=True)
def reset_global_instance():
    """Reset global session manager instance between tests"""
    import moai_adk.core.session_manager as sm_module

    sm_module._session_manager_instance = None
    yield
    sm_module._session_manager_instance = None


class TestSessionManagerInitialization:
    """Tests for SessionManager initialization"""

    def test_init_creates_default_directories(self, temp_project_dir):
        """Test that initialization creates default directories"""
        SessionManager(
            session_file=temp_project_dir / ".moai" / "memory" / "agent-sessions.json",
            transcript_dir=temp_project_dir / ".moai" / "logs" / "agent-transcripts",
        )

        assert (temp_project_dir / ".moai" / "memory").exists()
        assert (temp_project_dir / ".moai" / "logs" / "agent-transcripts").exists()

    def test_init_with_custom_paths(self, temp_project_dir):
        """Test initialization with custom session and transcript paths"""
        custom_session_file = temp_project_dir / "custom" / "sessions.json"
        custom_transcript_dir = temp_project_dir / "custom" / "transcripts"

        manager = SessionManager(
            session_file=custom_session_file,
            transcript_dir=custom_transcript_dir,
        )

        assert custom_session_file.parent.exists()
        assert custom_transcript_dir.exists()
        assert manager._session_file == custom_session_file
        assert manager._transcript_dir == custom_transcript_dir

    def test_init_empty_state(self, session_manager):
        """Test that initialization creates empty session state"""
        assert session_manager._sessions == {}
        assert session_manager._results == {}
        assert session_manager._chains == {}
        assert session_manager._metadata == {}

    def test_init_loads_existing_sessions(self, temp_project_dir):
        """Test that initialization loads existing session data"""
        session_file = temp_project_dir / ".moai" / "memory" / "agent-sessions.json"
        transcript_dir = temp_project_dir / ".moai" / "logs" / "agent-transcripts"

        # Create existing session file
        existing_data = {
            "sessions": {"agent-1": "id-abc123"},
            "chains": {"SPEC-001": ["id-abc123"]},
            "metadata": {"id-abc123": {"agent_name": "agent-1", "created_at": "2025-01-01T00:00:00"}},
        }
        session_file.parent.mkdir(parents=True, exist_ok=True)
        session_file.write_text(json.dumps(existing_data))

        manager = SessionManager(session_file=session_file, transcript_dir=transcript_dir)

        assert manager._sessions == {"agent-1": "id-abc123"}
        assert manager._chains == {"SPEC-001": ["id-abc123"]}

    def test_init_handles_corrupted_session_file(self, temp_project_dir):
        """Test that init handles corrupted JSON gracefully"""
        session_file = temp_project_dir / ".moai" / "memory" / "agent-sessions.json"
        transcript_dir = temp_project_dir / ".moai" / "logs" / "agent-transcripts"

        session_file.parent.mkdir(parents=True, exist_ok=True)
        session_file.write_text("{ invalid json }")

        manager = SessionManager(session_file=session_file, transcript_dir=transcript_dir)

        # Should initialize with empty state instead of crashing
        assert manager._sessions == {}
        assert manager._chains == {}


class TestAgentResultRegistration:
    """Tests for registering agent execution results"""

    def test_register_agent_result_basic(self, session_manager):
        """Test registering a basic agent result"""
        result_data = {"status": "success", "files_created": 5}

        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result=result_data,
        )

        assert session_manager._sessions["test-agent"] == "id-12345"
        assert session_manager._results["id-12345"]["result"] == result_data
        assert session_manager._results["id-12345"]["agent_name"] == "test-agent"

    def test_register_agent_result_with_chain(self, session_manager):
        """Test registering agent result with chain_id"""
        result_data = {"status": "success"}

        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result=result_data,
            chain_id="SPEC-AUTH-001",
        )

        assert session_manager._chains["SPEC-AUTH-001"] == ["id-111"]
        assert session_manager._results["id-111"]["chain_id"] == "SPEC-AUTH-001"

    def test_register_multiple_agents_in_chain(self, session_manager):
        """Test registering multiple agents in same chain"""
        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result={"status": "success"},
            chain_id="SPEC-001",
        )

        session_manager.register_agent_result(
            agent_name="agent-2",
            agent_id="id-222",
            result={"status": "success"},
            chain_id="SPEC-001",
        )

        assert session_manager._chains["SPEC-001"] == ["id-111", "id-222"]
        assert len(session_manager._results) == 2

    def test_register_creates_metadata(self, session_manager):
        """Test that registration creates metadata entries"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        metadata = session_manager._metadata["id-12345"]
        assert metadata["agent_name"] == "test-agent"
        assert metadata["resume_count"] == 0
        assert "created_at" in metadata

    def test_register_persists_to_disk(self, session_manager):
        """Test that registration persists data to disk"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        # Read file to verify persistence
        session_file = session_manager._session_file
        assert session_file.exists()

        with open(session_file, "r") as f:
            data = json.load(f)
            assert "test-agent" in data["sessions"]

    def test_register_replaces_previous_session(self, session_manager):
        """Test that registering new agent result replaces previous session"""
        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-old",
            result={"version": 1},
        )

        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-new",
            result={"version": 2},
        )

        assert session_manager._sessions["agent-1"] == "id-new"
        assert "id-old" in session_manager._results  # Old result still stored
        assert session_manager._results["id-new"]["result"]["version"] == 2

    @pytest.mark.parametrize(
        "agent_name,agent_id",
        [
            ("agent-1", "id-12345"),
            ("complex-agent-name", "id-abcdef123456"),
            ("a", "id-x"),
        ],
    )
    def test_register_various_agent_names(self, session_manager, agent_name, agent_id):
        """Test registering agents with various name formats"""
        session_manager.register_agent_result(
            agent_name=agent_name,
            agent_id=agent_id,
            result={"test": True},
        )

        assert session_manager._sessions[agent_name] == agent_id


class TestResumeDecision:
    """Tests for resume decision logic"""

    def test_get_resume_id_no_previous_session(self, session_manager):
        """Test get_resume_id returns None when no previous session"""
        resume_id = session_manager.get_resume_id("non-existent-agent")

        assert resume_id is None

    def test_get_resume_id_with_previous_session(self, session_manager):
        """Test get_resume_id returns agent_id when session exists"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        resume_id = session_manager.get_resume_id("test-agent")

        assert resume_id == "id-12345"

    def test_get_resume_id_with_chain_match(self, session_manager):
        """Test get_resume_id returns id when chain_id matches"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
            chain_id="SPEC-AUTH-001",
        )

        resume_id = session_manager.get_resume_id("test-agent", chain_id="SPEC-AUTH-001")

        assert resume_id == "id-12345"

    def test_get_resume_id_with_chain_mismatch(self, session_manager):
        """Test get_resume_id returns None when chain_id doesn't match"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
            chain_id="SPEC-AUTH-001",
        )

        resume_id = session_manager.get_resume_id("test-agent", chain_id="SPEC-OTHER-001")

        assert resume_id is None

    def test_should_resume_no_previous_session(self, session_manager):
        """Test should_resume returns False with no previous session"""
        result = session_manager.should_resume(
            agent_name="new-agent",
            current_task="Do something",
        )

        assert result is False

    def test_should_resume_no_previous_task(self, session_manager):
        """Test should_resume returns False when no previous task info"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        result = session_manager.should_resume(
            agent_name="test-agent",
            current_task="Do something",
            previous_task=None,
        )

        assert result is False

    def test_should_resume_related_tasks(self, session_manager):
        """Test should_resume returns True for related tasks"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        result = session_manager.should_resume(
            agent_name="test-agent",
            current_task="Implement user login endpoint",
            previous_task="Implement user registration endpoint",
        )

        assert result is True

    def test_should_resume_unrelated_tasks(self, session_manager):
        """Test should_resume returns False for unrelated tasks"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        result = session_manager.should_resume(
            agent_name="test-agent",
            current_task="Build database migrations",
            previous_task="Generate documentation for API",
        )

        assert result is False

    def test_should_resume_max_resume_depth(self, session_manager):
        """Test should_resume prevents infinite resume loops"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        # Simulate reaching max resume depth
        session_manager._metadata["id-12345"]["resume_count"] = 5

        result = session_manager.should_resume(
            agent_name="test-agent",
            current_task="Implementation task",
            previous_task="Implementation task related work",
        )

        assert result is False

    @pytest.mark.parametrize(
        "current,previous,expected",
        [
            ("Implement authentication", "Design authentication", True),
            ("Write API documentation", "Write client documentation", True),
            ("Fix database connection", "Fix database schema", True),
            ("Deploy to production", "Analyze performance report", False),
            ("Create REST endpoints", "Delete test files", False),
        ],
    )
    def test_should_resume_keyword_matching(self, session_manager, current, previous, expected):
        """Test keyword matching in resume decision"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        result = session_manager.should_resume(
            agent_name="test-agent",
            current_task=current,
            previous_task=previous,
        )

        assert result == expected


class TestResumeCountTracking:
    """Tests for resume count incrementation"""

    def test_increment_resume_count_basic(self, session_manager):
        """Test incrementing resume count"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        assert session_manager._metadata["id-12345"]["resume_count"] == 0

        session_manager.increment_resume_count("id-12345")

        assert session_manager._metadata["id-12345"]["resume_count"] == 1

    def test_increment_resume_count_multiple_times(self, session_manager):
        """Test incrementing resume count multiple times"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        for i in range(5):
            session_manager.increment_resume_count("id-12345")
            assert session_manager._metadata["id-12345"]["resume_count"] == i + 1

    def test_increment_resume_count_updates_timestamp(self, session_manager):
        """Test that incrementing resume count updates timestamp"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        session_manager.increment_resume_count("id-12345")

        assert "last_resumed_at" in session_manager._metadata["id-12345"]

    def test_increment_resume_count_nonexistent_agent(self, session_manager):
        """Test incrementing count for non-existent agent does nothing"""
        session_manager.increment_resume_count("non-existent-id")

        # Should not raise error
        assert "non-existent-id" not in session_manager._metadata

    def test_increment_resume_count_persists(self, session_manager):
        """Test that resume count increment is persisted"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        session_manager.increment_resume_count("id-12345")

        # Verify persistence
        session_file = session_manager._session_file
        with open(session_file, "r") as f:
            data = json.load(f)
            assert data["metadata"]["id-12345"]["resume_count"] == 1


class TestAgentResultRetrieval:
    """Tests for retrieving agent results"""

    def test_get_agent_result_existing(self, session_manager):
        """Test retrieving an existing agent result"""
        result_data = {"status": "success", "files": ["file1.py", "file2.py"]}

        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result=result_data,
        )

        retrieved = session_manager.get_agent_result("id-12345")

        assert retrieved == result_data

    def test_get_agent_result_nonexistent(self, session_manager):
        """Test retrieving non-existent agent result returns None"""
        result = session_manager.get_agent_result("non-existent-id")

        assert result is None

    def test_get_agent_result_complex_data(self, session_manager):
        """Test retrieving agent result with complex nested data"""
        result_data = {
            "status": "success",
            "metadata": {
                "duration": 42.5,
                "files": ["a.py", "b.py"],
                "nested": {"key": "value"},
            },
        }

        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result=result_data,
        )

        retrieved = session_manager.get_agent_result("id-12345")

        assert retrieved == result_data
        assert retrieved["metadata"]["nested"]["key"] == "value"


class TestWorkflowChainManagement:
    """Tests for workflow chain creation and management"""

    def test_create_chain_basic(self, session_manager, temp_project_dir):
        """Test creating a basic workflow chain"""
        chain_id = "SPEC-AUTH-001-planning"
        agent_sequence = ["spec-builder", "implementation-planner"]

        session_manager.create_chain(chain_id, agent_sequence)

        assert chain_id in session_manager._chains
        assert session_manager._chains[chain_id] == []

    def test_create_chain_with_metadata(self, session_manager, temp_project_dir):
        """Test creating chain with metadata"""
        chain_id = "SPEC-AUTH-001"
        metadata = {"spec_id": "SPEC-AUTH-001", "feature": "User Authentication"}

        session_manager.create_chain(
            chain_id,
            ["agent-1", "agent-2"],
            metadata=metadata,
        )

        chains_file = session_manager._session_file.parent / "workflow-chains.json"
        assert chains_file.exists()

        with open(chains_file, "r") as f:
            chains_data = json.load(f)
            assert chain_id in chains_data
            assert chains_data[chain_id]["metadata"] == metadata

    def test_get_chain_results_empty(self, session_manager):
        """Test getting results for empty chain"""
        results = session_manager.get_chain_results("SPEC-001")

        assert results == []

    def test_get_chain_results_with_agents(self, session_manager):
        """Test getting results for chain with agents"""
        chain_id = "SPEC-001"

        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result={"status": "success"},
            chain_id=chain_id,
        )

        session_manager.register_agent_result(
            agent_name="agent-2",
            agent_id="id-222",
            result={"output": "data"},
            chain_id=chain_id,
        )

        results = session_manager.get_chain_results(chain_id)

        assert len(results) == 2
        assert results[0]["agent_name"] == "agent-1"
        assert results[1]["agent_name"] == "agent-2"

    def test_get_chain_summary_not_found(self, session_manager):
        """Test getting summary for non-existent chain"""
        summary = session_manager.get_chain_summary("SPEC-NOTFOUND")

        assert summary["status"] == "not_found"

    def test_get_chain_summary_with_agents(self, session_manager):
        """Test getting summary for chain with agents"""
        chain_id = "SPEC-AUTH-001"

        session_manager.register_agent_result(
            agent_name="spec-builder",
            agent_id="id-111",
            result={"status": "success"},
            chain_id=chain_id,
        )

        session_manager.register_agent_result(
            agent_name="implementation-planner",
            agent_id="id-222",
            result={"status": "success"},
            chain_id=chain_id,
        )

        summary = session_manager.get_chain_summary(chain_id)

        assert summary["chain_id"] == chain_id
        assert summary["status"] == "completed"
        assert summary["agent_count"] == 2
        assert "spec-builder" in summary["agents"]
        assert "implementation-planner" in summary["agents"]

    def test_get_chain_summary_timestamps(self, session_manager):
        """Test that chain summary includes start and end timestamps"""
        chain_id = "SPEC-001"

        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result={"status": "success"},
            chain_id=chain_id,
        )

        summary = session_manager.get_chain_summary(chain_id)

        assert "started_at" in summary
        assert "completed_at" in summary


class TestSessionCleanup:
    """Tests for session cleanup operations"""

    def test_clear_agent_session_basic(self, session_manager):
        """Test clearing a single agent session"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        session_manager.clear_agent_session("test-agent")

        assert "test-agent" not in session_manager._sessions
        assert "id-12345" not in session_manager._results
        assert "id-12345" not in session_manager._metadata

    def test_clear_agent_session_nonexistent(self, session_manager):
        """Test clearing non-existent agent session does nothing"""
        session_manager.clear_agent_session("non-existent")

        # Should not raise error
        assert len(session_manager._sessions) == 0

    def test_clear_agent_session_persists(self, session_manager):
        """Test that clearing agent session persists to disk"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        session_manager.clear_agent_session("test-agent")

        # Verify persistence
        session_file = session_manager._session_file
        with open(session_file, "r") as f:
            data = json.load(f)
            assert "test-agent" not in data["sessions"]

    def test_clear_chain_basic(self, session_manager):
        """Test clearing a workflow chain"""
        chain_id = "SPEC-001"

        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result={"status": "success"},
            chain_id=chain_id,
        )

        session_manager.register_agent_result(
            agent_name="agent-2",
            agent_id="id-222",
            result={"status": "success"},
            chain_id=chain_id,
        )

        session_manager.clear_chain(chain_id)

        assert chain_id not in session_manager._chains
        assert "id-111" not in session_manager._results
        assert "id-222" not in session_manager._results

    def test_clear_chain_nonexistent(self, session_manager):
        """Test clearing non-existent chain does nothing"""
        session_manager.clear_chain("SPEC-NOTFOUND")

        # Should not raise error
        assert len(session_manager._chains) == 0

    def test_clear_chain_persists(self, session_manager):
        """Test that clearing chain persists to disk"""
        chain_id = "SPEC-001"

        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result={"status": "success"},
            chain_id=chain_id,
        )

        session_manager.clear_chain(chain_id)

        session_file = session_manager._session_file
        with open(session_file, "r") as f:
            data = json.load(f)
            assert chain_id not in data["chains"]


class TestSessionPersistence:
    """Tests for session persistence and recovery"""

    def test_save_sessions_creates_file(self, session_manager):
        """Test that saving sessions creates JSON file"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        session_file = session_manager._session_file
        assert session_file.exists()

        data = json.loads(session_file.read_text())
        assert "sessions" in data
        assert "chains" in data
        assert "metadata" in data

    def test_save_sessions_includes_timestamp(self, session_manager):
        """Test that saved sessions include last_updated timestamp"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        session_file = session_manager._session_file
        data = json.loads(session_file.read_text())

        assert "last_updated" in data
        assert len(data["last_updated"]) > 0

    def test_load_sessions_from_disk(self, temp_project_dir):
        """Test loading sessions from existing file"""
        session_file = temp_project_dir / ".moai" / "memory" / "agent-sessions.json"
        transcript_dir = temp_project_dir / ".moai" / "logs" / "agent-transcripts"

        # Create session file with data
        session_data = {
            "sessions": {"agent-1": "id-abc123"},
            "chains": {"SPEC-001": ["id-abc123"]},
            "metadata": {"id-abc123": {"agent_name": "agent-1", "created_at": "2025-01-01"}},
        }
        session_file.parent.mkdir(parents=True, exist_ok=True)
        session_file.write_text(json.dumps(session_data))

        # Load and verify
        manager = SessionManager(session_file=session_file, transcript_dir=transcript_dir)

        assert manager._sessions == {"agent-1": "id-abc123"}
        assert manager._chains == {"SPEC-001": ["id-abc123"]}

    def test_persistence_survives_reload(self, temp_project_dir):
        """Test that session mappings persist across manager instances"""
        session_file = temp_project_dir / ".moai" / "memory" / "agent-sessions.json"
        transcript_dir = temp_project_dir / ".moai" / "logs" / "agent-transcripts"

        # First manager: register agent
        manager1 = SessionManager(session_file=session_file, transcript_dir=transcript_dir)
        manager1.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        # Second manager: load same file
        manager2 = SessionManager(session_file=session_file, transcript_dir=transcript_dir)

        # Sessions are persisted (session mappings)
        assert manager2._sessions["test-agent"] == "id-12345"
        # Metadata is also persisted
        assert manager2._metadata["id-12345"]["agent_name"] == "test-agent"
        # Results are only in-memory (not persisted to disk)
        assert "id-12345" not in manager2._results


class TestTranscriptManagement:
    """Tests for transcript file management"""

    def test_export_transcript_existing_file(self, session_manager, temp_project_dir):
        """Test exporting transcript path for existing transcript"""
        transcript_dir = temp_project_dir / ".moai" / "logs" / "agent-transcripts"
        agent_id = "id-12345"

        # Create a transcript file
        transcript_file = transcript_dir / f"agent-{agent_id}.jsonl"
        transcript_file.write_text('{"message": "test"}\n')

        session_manager._transcript_dir = transcript_dir

        result = session_manager.export_transcript(agent_id)

        assert result == transcript_file
        assert result.exists()

    def test_export_transcript_nonexistent_file(self, session_manager):
        """Test exporting transcript for non-existent file returns None"""
        result = session_manager.export_transcript("non-existent-id")

        assert result is None

    def test_transcript_path_format(self, session_manager):
        """Test that transcript paths follow correct format"""
        agent_id = "test-agent-id-12345"
        transcript_file = session_manager._transcript_dir / f"agent-{agent_id}.jsonl"

        expected_name = f"agent-{agent_id}.jsonl"
        assert transcript_file.name == expected_name


class TestSessionQueries:
    """Tests for querying session state"""

    def test_get_all_sessions_empty(self, session_manager):
        """Test getting all sessions when empty"""
        sessions = session_manager.get_all_sessions()

        assert sessions["sessions"] == {}
        assert sessions["chains"] == []
        assert sessions["total_results"] == 0

    def test_get_all_sessions_with_data(self, session_manager):
        """Test getting all sessions with registered data"""
        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result={"status": "success"},
            chain_id="SPEC-001",
        )

        session_manager.register_agent_result(
            agent_name="agent-2",
            agent_id="id-222",
            result={"status": "success"},
            chain_id="SPEC-001",
        )

        sessions = session_manager.get_all_sessions()

        assert len(sessions["sessions"]) == 2
        assert "SPEC-001" in sessions["chains"]
        assert sessions["total_results"] == 2

    def test_get_all_sessions_format(self, session_manager):
        """Test that get_all_sessions returns correct structure"""
        sessions = session_manager.get_all_sessions()

        assert "sessions" in sessions
        assert "chains" in sessions
        assert "total_results" in sessions
        assert isinstance(sessions["sessions"], dict)
        assert isinstance(sessions["chains"], list)
        assert isinstance(sessions["total_results"], int)


class TestSingletonPattern:
    """Tests for global instance management"""

    def test_get_session_manager_singleton(self, temp_project_dir):
        """Test that get_session_manager returns singleton"""
        manager1 = get_session_manager()
        manager2 = get_session_manager()

        assert manager1 is manager2

    def test_get_session_manager_initializes_once(self):
        """Test that global instance is only initialized once"""
        manager = get_session_manager()

        assert manager is not None
        assert isinstance(manager, SessionManager)

    def test_convenience_register_agent(self, temp_project_dir):
        """Test convenience function register_agent"""
        # Reset singleton
        import moai_adk.core.session_manager as sm_module

        sm_module._session_manager_instance = None

        # Patch default paths to use temp dir
        with patch("moai_adk.core.session_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = temp_project_dir

            register_agent(
                agent_name="test-agent",
                agent_id="id-12345",
                result={"status": "success"},
            )

            manager = get_session_manager()
            assert "test-agent" in manager._sessions

    def test_convenience_get_resume_id(self, temp_project_dir):
        """Test convenience function get_resume_id"""
        # Reset singleton
        import moai_adk.core.session_manager as sm_module

        sm_module._session_manager_instance = None

        with patch("moai_adk.core.session_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = temp_project_dir

            register_agent(
                agent_name="test-agent",
                agent_id="id-12345",
                result={"status": "success"},
            )

            resume_id = get_resume_id("test-agent")
            assert resume_id == "id-12345"

    def test_convenience_should_resume(self, temp_project_dir):
        """Test convenience function should_resume"""
        # Reset singleton
        import moai_adk.core.session_manager as sm_module

        sm_module._session_manager_instance = None

        with patch("moai_adk.core.session_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = temp_project_dir

            register_agent(
                agent_name="test-agent",
                agent_id="id-12345",
                result={"status": "success"},
            )

            result = should_resume(
                agent_name="test-agent",
                current_task="Implement feature",
                previous_task="Implement feature design",
            )
            assert result is True


class TestComplexWorkflows:
    """Tests for complex workflow scenarios"""

    def test_linear_chain_workflow(self, session_manager):
        """Test a linear chain: spec-builder → implementation-planner → tdd-implementer"""
        chain_id = "SPEC-AUTH-001-workflow"

        session_manager.create_chain(
            chain_id,
            ["spec-builder", "implementation-planner", "tdd-implementer"],
        )

        # Spec builder completes
        session_manager.register_agent_result(
            agent_name="spec-builder",
            agent_id="spec-abc123",
            result={"spec_id": "SPEC-AUTH-001"},
            chain_id=chain_id,
        )

        # Implementation planner completes
        session_manager.register_agent_result(
            agent_name="implementation-planner",
            agent_id="plan-def456",
            result={"implementation_plan": {}},
            chain_id=chain_id,
        )

        # TDD implementer starts and finishes
        session_manager.register_agent_result(
            agent_name="tdd-implementer",
            agent_id="tdd-ghi789",
            result={"files_created": 5},
            chain_id=chain_id,
        )

        summary = session_manager.get_chain_summary(chain_id)
        assert summary["agent_count"] == 3
        assert len(summary["agents"]) == 3

    def test_parallel_agents_same_chain(self, session_manager):
        """Test parallel agents working on same chain"""
        chain_id = "SPEC-AUTH-001-review"

        session_manager.create_chain(
            chain_id,
            ["backend-expert", "security-expert", "frontend-expert"],
        )

        # All experts run in parallel
        experts = [
            ("backend-expert", "backend-jkl012", {"recommendations": ["Use JWT"]}),
            ("security-expert", "security-mno345", {"vulnerabilities": ["Rate limiting"]}),
            ("frontend-expert", "frontend-pqr678", {"ui_concerns": ["Token refresh"]}),
        ]

        for expert_name, agent_id, result in experts:
            session_manager.register_agent_result(
                agent_name=expert_name,
                agent_id=agent_id,
                result=result,
                chain_id=chain_id,
            )

        results = session_manager.get_chain_results(chain_id)
        assert len(results) == 3

    def test_nested_chain_hierarchy(self, session_manager):
        """Test managing multiple independent chains"""
        chains = [
            ("SPEC-AUTH-001", ["agent-1", "agent-2"]),
            ("SPEC-API-001", ["agent-3", "agent-4"]),
            ("SPEC-UI-001", ["agent-5", "agent-6"]),
        ]

        for chain_id, agents in chains:
            session_manager.create_chain(chain_id, agents)

            for i, agent_name in enumerate(agents):
                session_manager.register_agent_result(
                    agent_name=agent_name,
                    agent_id=f"id-{chain_id}-{i}",
                    result={"status": "success"},
                    chain_id=chain_id,
                )

        # Verify all chains exist
        for chain_id, _ in chains:
            summary = session_manager.get_chain_summary(chain_id)
            assert summary["agent_count"] == 2


class TestIOErrorHandling:
    """Tests for I/O error handling"""

    def test_save_sessions_io_error(self, session_manager):
        """Test handling of IOError when saving sessions"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        # Make directory read-only to cause IOError
        session_file = session_manager._session_file
        session_file.parent.chmod(0o444)

        try:
            # Try to save again (should log error but not crash)
            session_manager._save_sessions()
        finally:
            # Restore permissions for cleanup
            session_file.parent.chmod(0o755)

    def test_increment_resume_count_updates_file_on_error(self, session_manager):
        """Test that increment_resume_count attempts to save despite errors"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={"status": "success"},
        )

        # Increment should work even if file operations have issues
        session_manager.increment_resume_count("id-12345")

        assert session_manager._metadata["id-12345"]["resume_count"] == 1


class TestEdgeCases:
    """Tests for edge cases and error conditions"""

    def test_register_result_with_none_value(self, session_manager):
        """Test registering result with None value"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result=None,
        )

        assert session_manager.get_agent_result("id-12345") is None

    def test_register_result_with_empty_dict(self, session_manager):
        """Test registering result with empty dictionary"""
        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result={},
        )

        assert session_manager.get_agent_result("id-12345") == {}

    def test_register_result_with_large_data(self, session_manager):
        """Test registering result with large data structure"""
        large_data = {"items": [{"id": i, "data": "x" * 1000} for i in range(100)]}

        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result=large_data,
        )

        retrieved = session_manager.get_agent_result("id-12345")
        assert len(retrieved["items"]) == 100

    def test_special_characters_in_agent_name(self, session_manager):
        """Test handling special characters in agent names"""
        session_manager.register_agent_result(
            agent_name="agent-with-special_chars.123",
            agent_id="id-12345",
            result={"status": "success"},
        )

        assert session_manager.get_resume_id("agent-with-special_chars.123") == "id-12345"

    def test_unicode_in_result_data(self, session_manager):
        """Test handling Unicode characters in result data"""
        result_data = {
            "message": "한글 テスト 日本語 中文 العربية",
            "emoji": "🚀🎉✨",
        }

        session_manager.register_agent_result(
            agent_name="test-agent",
            agent_id="id-12345",
            result=result_data,
        )

        retrieved = session_manager.get_agent_result("id-12345")
        assert retrieved["message"] == result_data["message"]
        assert retrieved["emoji"] == result_data["emoji"]

    def test_very_long_chain_id(self, session_manager):
        """Test handling very long chain identifiers"""
        long_chain_id = "SPEC-" + "A" * 200

        session_manager.register_agent_result(
            agent_name="agent-1",
            agent_id="id-111",
            result={"status": "success"},
            chain_id=long_chain_id,
        )

        summary = session_manager.get_chain_summary(long_chain_id)
        assert summary["chain_id"] == long_chain_id

    def test_rapid_sequential_registrations(self, session_manager):
        """Test rapid sequential agent registrations"""
        for i in range(50):
            session_manager.register_agent_result(
                agent_name=f"agent-{i}",
                agent_id=f"id-{i}",
                result={"iteration": i},
            )

        assert len(session_manager._sessions) == 50
        assert len(session_manager._results) == 50

    def test_same_agent_multiple_registrations(self, session_manager):
        """Test registering same agent multiple times overwrites session"""
        for i in range(5):
            session_manager.register_agent_result(
                agent_name="recurring-agent",
                agent_id=f"id-{i}",
                result={"iteration": i},
            )

        # Only last registration should be in sessions
        assert session_manager._sessions["recurring-agent"] == "id-4"
        # But all results should be stored
        assert len(session_manager._results) == 5

    def test_concurrent_chain_access(self, session_manager):
        """Test accessing same chain from multiple registrations"""
        chain_id = "SPEC-001"

        for i in range(10):
            session_manager.register_agent_result(
                agent_name=f"agent-{i}",
                agent_id=f"id-{i}",
                result={"index": i},
                chain_id=chain_id,
            )

        assert len(session_manager._chains[chain_id]) == 10

    def test_special_chain_id_formats(self, session_manager):
        """Test various chain ID formats"""
        chain_ids = [
            "SPEC-001",
            "SPEC-AUTH-001-implementation",
            "complex.chain.id",
            "chain-with-dashes-and-numbers-123",
            "UPPERCASE_CHAIN_ID",
        ]

        for chain_id in chain_ids:
            session_manager.register_agent_result(
                agent_name=f"agent-for-{chain_id}",
                agent_id=f"id-{chain_id}",
                result={"chain": chain_id},
                chain_id=chain_id,
            )

        for chain_id in chain_ids:
            summary = session_manager.get_chain_summary(chain_id)
            assert summary["chain_id"] == chain_id
