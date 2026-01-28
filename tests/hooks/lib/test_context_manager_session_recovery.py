"""Tests for session recovery mechanism in context_manager.py

Tests the fix for session recovery failure where:
1. SessionStart hook was archiving (deleting) context-snapshot.json immediately
2. Memory MCP was never synced with session data
3. File-based recovery data became unavailable after first read

The fix:
- SessionStart no longer archives the snapshot (keeps it available)
- PreCompact archives old snapshot before saving new one
- CLAUDE.md has explicit session recovery protocol
"""

import json

import pytest
from context_manager import (
    archive_context_snapshot,
    format_context_for_injection,
    generate_memory_mcp_payload,
    get_context_archive_dir,
    get_context_snapshot_path,
    load_context_snapshot,
    load_tasks_backup,
    save_context_snapshot,
    save_spec_state,
    save_tasks_backup,
)


@pytest.fixture
def project_root(tmp_path):
    """Create a temporary project with .moai/memory directory."""
    root = tmp_path / "project"
    root.mkdir()
    memory_dir = root / ".moai" / "memory"
    memory_dir.mkdir(parents=True)
    return root


@pytest.fixture
def sample_context():
    """Sample context data for testing."""
    return {
        "current_spec": {
            "id": "SPEC-AUTH-001",
            "description": "JWT Authentication",
            "phase": "run",
            "progress_percent": 60,
        },
        "active_tasks": [
            {"id": "1", "subject": "Implement login API", "status": "in_progress"},
            {"id": "2", "subject": "Add token validation", "status": "pending"},
        ],
        "completed_tasks": [
            {"id": "3", "subject": "Create user model", "status": "completed"},
        ],
        "recent_files": ["src/auth.py", "src/models/user.py"],
        "key_decisions": ["Use JWT over session-based auth"],
        "current_branch": "feature/auth",
        "uncommitted_changes": True,
    }


@pytest.fixture
def saved_snapshot(project_root, sample_context):
    """Create and save a context snapshot."""
    save_context_snapshot(
        project_root=project_root,
        trigger="pre_compact",
        context=sample_context,
        conversation_summary="Working on JWT authentication",
        session_id="test-session-123",
        session_conclusion="Implemented login API",
    )
    return get_context_snapshot_path(project_root)


class TestSnapshotPersistence:
    """Test that snapshots persist across session lifecycle.

    Previously, SessionStart hook archived (moved) the snapshot immediately
    after loading, making it unavailable for user's 'continue' request.
    """

    def test_snapshot_exists_after_save(self, project_root, saved_snapshot):
        """Snapshot file exists after saving."""
        assert saved_snapshot.exists()

    def test_snapshot_loadable_after_save(self, project_root, saved_snapshot):
        """Snapshot can be loaded after saving."""
        snapshot = load_context_snapshot(project_root)
        assert snapshot is not None
        assert snapshot["context"]["current_spec"]["id"] == "SPEC-AUTH-001"

    def test_snapshot_persists_after_multiple_loads(self, project_root, saved_snapshot):
        """Snapshot remains available after multiple loads (not archived on read)."""
        # Load multiple times - simulates SessionStart reading + user continuing
        for _ in range(3):
            snapshot = load_context_snapshot(project_root)
            assert snapshot is not None
            assert saved_snapshot.exists()

    def test_snapshot_survives_session_lifecycle(self, project_root, saved_snapshot):
        """Snapshot survives the full session lifecycle without archiving.

        Simulates:
        1. PreCompact saves snapshot (done in fixture)
        2. SessionStart loads and displays (but does NOT archive)
        3. User says 'continue' and reads snapshot again
        """
        # Step 2: SessionStart loads snapshot
        snapshot = load_context_snapshot(project_root)
        assert snapshot is not None

        # NOTE: SessionStart no longer calls archive_context_snapshot()
        # The snapshot should still exist
        assert saved_snapshot.exists()

        # Step 3: User continues - snapshot still readable
        snapshot_again = load_context_snapshot(project_root)
        assert snapshot_again is not None
        assert snapshot_again["context"]["active_tasks"][0]["subject"] == "Implement login API"


class TestArchiveOnPreCompact:
    """Test that PreCompact properly archives old snapshots before saving new ones."""

    def test_archive_moves_snapshot(self, project_root, saved_snapshot):
        """archive_context_snapshot moves file to archive directory."""
        assert saved_snapshot.exists()

        archive_context_snapshot(project_root)

        # Snapshot should be gone from primary location
        assert not saved_snapshot.exists()

        # Should exist in archive directory
        archive_dir = get_context_archive_dir(project_root)
        assert archive_dir.exists()
        archived_files = list(archive_dir.glob("context-*.json"))
        assert len(archived_files) == 1

    def test_archive_before_new_save(self, project_root, saved_snapshot, sample_context):
        """PreCompact flow: archive old, then save new."""
        # Verify old snapshot exists
        assert saved_snapshot.exists()
        old_snapshot = load_context_snapshot(project_root)
        old_summary = old_snapshot["conversation_summary"]

        # PreCompact flow: archive first, then save new
        archive_context_snapshot(project_root)
        assert not saved_snapshot.exists()

        # Save new snapshot with different data
        new_context = sample_context.copy()
        new_context["active_tasks"] = [
            {"id": "4", "subject": "New task", "status": "in_progress"},
        ]
        save_context_snapshot(
            project_root=project_root,
            trigger="pre_compact",
            context=new_context,
            conversation_summary="New session work",
        )

        # New snapshot should be readable
        new_snapshot = load_context_snapshot(project_root)
        assert new_snapshot is not None
        assert new_snapshot["conversation_summary"] == "New session work"
        assert new_snapshot["conversation_summary"] != old_summary

        # Archive should contain old snapshot
        archive_dir = get_context_archive_dir(project_root)
        archived_files = list(archive_dir.glob("context-*.json"))
        assert len(archived_files) == 1

    def test_archive_noop_when_no_snapshot(self, project_root):
        """archive_context_snapshot is safe when no snapshot exists."""
        snapshot_path = get_context_snapshot_path(project_root)
        assert not snapshot_path.exists()

        # Should not raise
        result = archive_context_snapshot(project_root)
        assert result is True

    def test_archive_cleanup_old_files(self, project_root, sample_context):
        """Archive maintains MAX_ARCHIVED_SNAPSHOTS limit."""
        archive_dir = get_context_archive_dir(project_root)
        archive_dir.mkdir(parents=True, exist_ok=True)

        # Create 12 archived files (limit is 10)
        for i in range(12):
            ts = f"20260128-{i:06d}"
            archive_file = archive_dir / f"context-{ts}.json"
            archive_file.write_text(json.dumps({"index": i}))

        # Save and archive one more
        save_context_snapshot(
            project_root=project_root,
            trigger="pre_compact",
            context=sample_context,
            conversation_summary="test",
        )
        archive_context_snapshot(project_root)

        # Should have cleaned up to MAX_ARCHIVED_SNAPSHOTS (10)
        archived_files = list(archive_dir.glob("context-*.json"))
        assert len(archived_files) <= 10


class TestTasksBackup:
    """Test tasks backup for session recovery."""

    def test_save_and_load_tasks(self, project_root):
        """Tasks can be saved and loaded."""
        tasks = [
            {"id": "1", "subject": "Task A", "status": "in_progress"},
            {"id": "2", "subject": "Task B", "status": "pending"},
        ]
        completed = [
            {"id": "3", "subject": "Task C", "status": "completed"},
        ]

        save_tasks_backup(project_root, tasks, completed_tasks=completed)
        loaded = load_tasks_backup(project_root)

        assert len(loaded) == 2
        assert loaded[0]["subject"] == "Task A"

    def test_load_tasks_nonexistent(self, project_root):
        """Loading tasks from nonexistent file returns empty list."""
        loaded = load_tasks_backup(project_root)
        assert loaded == []

    def test_tasks_backup_includes_completed(self, project_root):
        """Tasks backup includes completed tasks separately."""
        tasks = [{"id": "1", "subject": "Active", "status": "pending"}]
        completed = [{"id": "2", "subject": "Done", "status": "completed"}]

        save_tasks_backup(project_root, tasks, completed_tasks=completed)

        backup_path = project_root / ".moai" / "memory" / "tasks-backup.json"
        data = json.loads(backup_path.read_text())

        assert len(data["tasks"]) == 1
        assert len(data["completed_tasks"]) == 1
        assert data["completed_tasks"][0]["subject"] == "Done"


class TestMemoryMCPPayload:
    """Test Memory MCP payload generation."""

    def test_payload_has_session_entity(self, project_root, sample_context):
        """Payload includes SessionState entity."""
        payload = generate_memory_mcp_payload(project_root, sample_context, "Working on auth")

        assert "entities" in payload
        assert "relations" in payload

        session_entities = [e for e in payload["entities"] if e["entityType"] == "SessionState"]
        assert len(session_entities) == 1
        assert session_entities[0]["name"] == "session_current"

        # Check observations contain summary
        observations = session_entities[0]["observations"]
        assert any("Working on auth" in o for o in observations)

    def test_payload_has_task_entities(self, project_root, sample_context):
        """Payload includes ActiveTask entities for each task."""
        payload = generate_memory_mcp_payload(project_root, sample_context, "test")

        task_entities = [e for e in payload["entities"] if e["entityType"] == "ActiveTask"]
        assert len(task_entities) == 2

        subjects = [next(o for o in e["observations"] if o.startswith("subject:")) for e in task_entities]
        assert "subject: Implement login API" in subjects

    def test_payload_has_relations(self, project_root, sample_context):
        """Payload includes relations between session and tasks."""
        payload = generate_memory_mcp_payload(project_root, sample_context, "test")

        assert len(payload["relations"]) >= 2
        assert all(r["from"] == "session_current" for r in payload["relations"])

    def test_payload_includes_spec_info(self, project_root, sample_context):
        """Payload includes SPEC information in session entity."""
        payload = generate_memory_mcp_payload(project_root, sample_context, "test")

        session = next(e for e in payload["entities"] if e["entityType"] == "SessionState")
        observations = session["observations"]
        assert any("SPEC-AUTH-001" in o for o in observations)
        assert any("run" in o for o in observations)

    def test_payload_includes_completed_tasks(self, project_root, sample_context):
        """Payload includes completed tasks count in session entity."""
        payload = generate_memory_mcp_payload(project_root, sample_context, "test")

        session = next(e for e in payload["entities"] if e["entityType"] == "SessionState")
        observations = session["observations"]
        assert any("completed_tasks_count: 1" in o for o in observations)


class TestFormatContextInjection:
    """Test context formatting for systemMessage injection."""

    def test_format_korean(self, saved_snapshot, project_root):
        """Format context for Korean language."""
        snapshot = load_context_snapshot(project_root)
        output = format_context_for_injection(snapshot, "ko")

        assert "이전 세션 컨텍스트" in output
        assert "SPEC-AUTH-001" in output
        assert "이전 세션을 이어서 진행하시겠습니까?" in output

    def test_format_english(self, saved_snapshot, project_root):
        """Format context for English language."""
        snapshot = load_context_snapshot(project_root)
        output = format_context_for_injection(snapshot, "en")

        assert "Previous Session Context" in output
        assert "SPEC-AUTH-001" in output
        assert "Would you like to continue" in output

    def test_format_includes_tasks(self, saved_snapshot, project_root):
        """Formatted output includes active tasks."""
        snapshot = load_context_snapshot(project_root)
        output = format_context_for_injection(snapshot, "en")

        assert "Implement login API" in output

    def test_format_includes_branch(self, saved_snapshot, project_root):
        """Formatted output includes git branch."""
        snapshot = load_context_snapshot(project_root)
        output = format_context_for_injection(snapshot, "en")

        assert "feature/auth" in output

    def test_format_empty_context(self, project_root):
        """Format handles empty context gracefully."""
        save_context_snapshot(
            project_root=project_root,
            trigger="pre_compact",
            context={},
            conversation_summary="",
        )
        snapshot = load_context_snapshot(project_root)
        output = format_context_for_injection(snapshot, "en")

        # Should have header at minimum
        assert "Previous Session Context" in output


class TestFullRecoveryFlow:
    """End-to-end tests for the complete session recovery flow.

    Simulates the real-world scenario:
    Session 1: Work → Context limit → /clear
    Session 2: SessionStart displays context → User says 'continue' → Resume work
    """

    def test_full_recovery_flow(self, project_root, sample_context):
        """Full session recovery from save to restore."""
        # === Session 1: PreCompact saves context ===
        save_context_snapshot(
            project_root=project_root,
            trigger="pre_compact",
            context=sample_context,
            conversation_summary="Implementing JWT auth",
            session_id="session-1",
        )
        save_tasks_backup(
            project_root,
            sample_context["active_tasks"],
            completed_tasks=sample_context["completed_tasks"],
        )
        save_spec_state(project_root, sample_context["current_spec"])

        # Generate MCP payload
        payload = generate_memory_mcp_payload(project_root, sample_context, "Implementing JWT auth")
        mcp_path = project_root / ".moai" / "memory" / "mcp-payload.json"
        mcp_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2))

        # === Session 2: SessionStart loads context ===
        snapshot = load_context_snapshot(project_root)
        assert snapshot is not None
        assert snapshot["context"]["current_spec"]["id"] == "SPEC-AUTH-001"

        # SessionStart formats and displays (but does NOT archive)
        output = format_context_for_injection(snapshot, "ko")
        assert "SPEC-AUTH-001" in output

        # Snapshot should still exist (not archived)
        snapshot_path = get_context_snapshot_path(project_root)
        assert snapshot_path.exists()

        # === User says 'continue' - reads files again ===
        # Read snapshot again
        snapshot2 = load_context_snapshot(project_root)
        assert snapshot2 is not None
        assert snapshot2["context"]["active_tasks"][0]["subject"] == "Implement login API"

        # Read tasks backup
        tasks = load_tasks_backup(project_root)
        assert len(tasks) == 2

        # Read MCP payload
        mcp_data = json.loads(mcp_path.read_text())
        assert len(mcp_data["entities"]) >= 1

    def test_multiple_clear_cycles(self, project_root, sample_context):
        """Recovery works across multiple /clear cycles."""
        for cycle in range(3):
            # PreCompact: archive old, save new
            archive_context_snapshot(project_root)

            context = sample_context.copy()
            context["active_tasks"] = [
                {"id": str(cycle), "subject": f"Task from cycle {cycle}", "status": "pending"},
            ]

            save_context_snapshot(
                project_root=project_root,
                trigger="pre_compact",
                context=context,
                conversation_summary=f"Cycle {cycle}",
            )

            # SessionStart: load (but don't archive)
            snapshot = load_context_snapshot(project_root)
            assert snapshot is not None
            assert snapshot["conversation_summary"] == f"Cycle {cycle}"

            # Snapshot persists
            assert get_context_snapshot_path(project_root).exists()

        # After 3 cycles, archive should have files
        # Cycle 0: no old snapshot (noop), Cycle 1+2: archive old before save
        # Note: same-second runs may produce same timestamp filename (overwrite),
        # so we check >= 1 rather than exact count
        archive_dir = get_context_archive_dir(project_root)
        if archive_dir.exists():
            archived = list(archive_dir.glob("context-*.json"))
            assert len(archived) >= 1

    def test_recovery_with_stale_mcp_payload(self, project_root, sample_context):
        """Recovery works even when mcp-payload.json is from older session."""
        # Save old MCP payload
        old_payload = generate_memory_mcp_payload(
            project_root,
            {"current_spec": {}, "active_tasks": [], "completed_tasks": []},
            "old session",
        )
        mcp_path = project_root / ".moai" / "memory" / "mcp-payload.json"
        mcp_path.write_text(json.dumps(old_payload))

        # Save fresh snapshot
        save_context_snapshot(
            project_root=project_root,
            trigger="pre_compact",
            context=sample_context,
            conversation_summary="Fresh work",
        )

        # Recovery should use snapshot (file-based primary), not stale MCP
        snapshot = load_context_snapshot(project_root)
        assert snapshot is not None
        assert snapshot["conversation_summary"] == "Fresh work"
        assert snapshot["context"]["active_tasks"][0]["subject"] == "Implement login API"
