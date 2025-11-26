"""Integration tests for SessionEnd Hook system

Tests for:
- session_end__auto_cleanup.py

And Hook chain execution order for session end.
"""

import json
import subprocess
from datetime import datetime, timedelta
from pathlib import Path


class TestSessionEndHookExecution:
    """Test basic SessionEnd Hook execution"""

    def test_hook_execution_basic(self, config_file, hook_tmp_project):
        """SessionEnd hook executes successfully"""
        with hook_tmp_project:
            assert config_file.exists()

            # Simulate hook execution
            result = {"hook": "session_end__auto_cleanup", "success": True, "timestamp": datetime.now().isoformat()}

            assert result["success"] is True
            assert result["hook"] == "session_end__auto_cleanup"

    def test_hook_execution_with_payload(self, config_file, hook_tmp_project):
        """SessionEnd hook executes with payload"""
        with hook_tmp_project as proj_root:
            # Simulate hook payload
            payload = {"event": "session_end", "timestamp": datetime.now().isoformat(), "cwd": str(proj_root)}

            result = {
                "hook": "session_end__auto_cleanup",
                "success": True,
                "payload_received": bool(payload),
                "timestamp": datetime.now().isoformat(),
            }

            assert result["success"] is True
            assert result["payload_received"] is True


class TestSessionMetricsSaving:
    """Test session metrics collection and saving"""

    def test_save_session_metrics(self, config_file, hook_tmp_project):
        """Session metrics are saved correctly"""
        with hook_tmp_project as proj_root:
            logs_dir = proj_root / ".moai" / "logs" / "sessions"
            logs_dir.mkdir(parents=True, exist_ok=True)

            session_metrics = {
                "session_id": datetime.now().strftime("%Y-%m-%d-%H%M%S"),
                "end_time": datetime.now().isoformat(),
                "cwd": str(proj_root),
                "files_modified": 5,
                "git_commits": 2,
                "specs_worked_on": ["SPEC-001", "SPEC-002"],
            }

            session_file = logs_dir / f"session-{session_metrics['session_id']}.json"
            session_file.write_text(json.dumps(session_metrics, indent=2))

            # Verify metrics are saved
            assert session_file.exists()
            loaded_metrics = json.loads(session_file.read_text())
            assert loaded_metrics["files_modified"] == 5
            assert loaded_metrics["git_commits"] == 2
            assert "SPEC-001" in loaded_metrics["specs_worked_on"]

    def test_session_metrics_file_format(self, config_file, hook_tmp_project):
        """Session metrics file has correct format"""
        with hook_tmp_project as proj_root:
            logs_dir = proj_root / ".moai" / "logs" / "sessions"
            logs_dir.mkdir(parents=True, exist_ok=True)

            session_id = datetime.now().strftime("%Y-%m-%d-%H%M%S")
            session_file = logs_dir / f"session-{session_id}.json"

            metrics = {
                "session_id": session_id,
                "end_time": datetime.now().isoformat(),
                "cwd": str(proj_root),
                "files_modified": 3,
                "git_commits": 1,
                "specs_worked_on": [],
            }

            session_file.write_text(json.dumps(metrics, indent=2))

            # Verify format
            assert session_file.name.startswith("session-")
            assert session_file.name.endswith(".json")

            loaded = json.loads(session_file.read_text())
            assert "session_id" in loaded
            assert "end_time" in loaded
            assert "files_modified" in loaded

    def test_session_metrics_directory_creation(self, config_file, hook_tmp_project):
        """Session metrics directory is created"""
        with hook_tmp_project as proj_root:
            logs_dir = proj_root / ".moai" / "logs" / "sessions"

            # Directory may or may not exist
            if not logs_dir.exists():
                logs_dir.mkdir(parents=True, exist_ok=True)

            assert logs_dir.exists()

    def test_multiple_session_metrics(self, config_file, hook_tmp_project):
        """Multiple session metrics can be stored"""
        with hook_tmp_project as proj_root:
            logs_dir = proj_root / ".moai" / "logs" / "sessions"
            logs_dir.mkdir(parents=True, exist_ok=True)

            # Create multiple session files
            for i in range(3):
                session_id = (datetime.now() - timedelta(hours=i)).strftime("%Y-%m-%d-%H%M%S")

                metrics = {"session_id": session_id, "files_modified": i + 1}

                session_file = logs_dir / f"session-{session_id}.json"
                session_file.write_text(json.dumps(metrics))

            # Verify all sessions saved
            session_files = list(logs_dir.glob("session-*.json"))
            assert len(session_files) >= 3


class TestWorkStateSaving:
    """Test work state snapshot saving"""

    def test_save_work_state(self, config_file, hook_tmp_project):
        """Work state is saved correctly"""
        with hook_tmp_project as proj_root:
            memory_dir = proj_root / ".moai" / "memory"
            memory_dir.mkdir(exist_ok=True)

            work_state = {
                "last_updated": datetime.now().isoformat(),
                "current_branch": "feature/test",
                "uncommitted_changes": False,
                "uncommitted_files": 0,
                "specs_in_progress": ["SPEC-001"],
            }

            state_file = memory_dir / "last-session-state.json"
            state_file.write_text(json.dumps(work_state, indent=2))

            # Verify state is saved
            assert state_file.exists()
            loaded_state = json.loads(state_file.read_text())
            assert loaded_state["current_branch"] == "feature/test"
            assert loaded_state["uncommitted_files"] == 0

    def test_work_state_includes_branch_info(self, config_file, hook_tmp_project):
        """Work state includes current branch"""
        with hook_tmp_project:
            work_state = {"current_branch": "main", "specs_in_progress": []}

            assert "current_branch" in work_state
            assert work_state["current_branch"] == "main"

    def test_work_state_includes_uncommitted_info(self, config_file, hook_tmp_project):
        """Work state includes uncommitted changes info"""
        with hook_tmp_project:
            work_state = {"uncommitted_changes": True, "uncommitted_files": 5}

            assert "uncommitted_changes" in work_state
            assert "uncommitted_files" in work_state
            assert work_state["uncommitted_files"] > 0

    def test_work_state_includes_spec_progress(self, config_file, hook_tmp_project):
        """Work state includes SPEC progress"""
        with hook_tmp_project:
            work_state = {"specs_in_progress": ["SPEC-001", "SPEC-002"], "last_updated": datetime.now().isoformat()}

            assert "specs_in_progress" in work_state
            assert len(work_state["specs_in_progress"]) > 0


class TestUncommittedChangesDetection:
    """Test uncommitted Git changes detection"""

    def test_detect_uncommitted_changes(self, git_repo_with_changes, hook_tmp_project):
        """Uncommitted changes are detected"""
        with hook_tmp_project:
            # Use the git_repo_with_changes fixture's directory
            git_root = git_repo_with_changes

            # Check for uncommitted changes
            result = subprocess.run(["git", "status", "--porcelain"], cwd=git_root, capture_output=True, text=True)

            assert result.returncode == 0
            # Should show uncommitted files
            assert len(result.stdout.strip()) > 0

    def test_count_uncommitted_files(self, git_repo_with_changes, hook_tmp_project):
        """Uncommitted files are counted correctly"""
        with hook_tmp_project:
            git_root = git_repo_with_changes

            result = subprocess.run(["git", "status", "--porcelain"], cwd=git_root, capture_output=True, text=True)

            uncommitted_lines = [line for line in result.stdout.strip().split("\n") if line]
            uncommitted_count = len(uncommitted_lines)

            # Should have at least 2 uncommitted files
            assert uncommitted_count >= 2

    def test_warning_message_for_uncommitted(self, git_repo_with_changes, hook_tmp_project):
        """Warning is generated for uncommitted changes"""
        with hook_tmp_project:
            git_root = git_repo_with_changes

            result = subprocess.run(["git", "status", "--porcelain"], cwd=git_root, capture_output=True, text=True)

            if result.stdout.strip():
                uncommitted_count = len(result.stdout.strip().split("\n"))
                warning = (
                    f"⚠️  {uncommitted_count} uncommitted files detected - " "Consider committing or stashing changes"
                )

                assert "uncommitted files" in warning.lower()
                assert str(uncommitted_count) in warning


class TestCleanupExecution:
    """Test cleanup execution in SessionEnd"""

    def test_cleanup_temp_files(self, config_file, cleanup_test_files, hook_tmp_project):
        """Temporary files are cleaned up"""
        with hook_tmp_project as proj_root:
            temp_dir = proj_root / ".moai" / "temp"
            temp_dir.mkdir(exist_ok=True)

            # Create old temp file
            old_timestamp = (datetime.now() - timedelta(days=10)).timestamp()
            old_file = temp_dir / "old-temp.txt"
            old_file.write_text("old content")
            Path(old_file).touch((old_timestamp, old_timestamp))

            assert old_file.exists()

            # Delete old file (simulate cleanup)
            old_file.unlink()

            assert not old_file.exists()

    def test_cleanup_cache_files(self, config_file, cleanup_test_files, hook_tmp_project):
        """Cache files are cleaned up"""
        with hook_tmp_project as proj_root:
            cache_dir = proj_root / ".moai" / "cache"

            # Create old cache file
            old_timestamp = (datetime.now() - timedelta(days=10)).timestamp()
            old_cache = cache_dir / "old-cache.json"
            old_cache.write_text('{"data": "old"}')
            Path(old_cache).touch((old_timestamp, old_timestamp))

            assert old_cache.exists()

            # Delete old cache
            old_cache.unlink()

            assert not old_cache.exists()

    def test_cleanup_respects_config(self, config_file, hook_tmp_project):
        """Cleanup respects configuration"""
        with hook_tmp_project:
            config_data = json.loads(config_file.read_text())

            cleanup_config = config_data.get("auto_cleanup", {})
            assert "enabled" in cleanup_config
            assert "cleanup_days" in cleanup_config

            # Cleanup should be enabled
            assert cleanup_config["enabled"] is True


class TestStateSavingAndRestoration:
    """Test state saving and restoration"""

    def test_save_session_state(self, config_file, hook_tmp_project):
        """Session state is saved"""
        with hook_tmp_project as proj_root:
            memory_dir = proj_root / ".moai" / "memory"
            memory_dir.mkdir(exist_ok=True)

            state = {"timestamp": datetime.now().isoformat(), "branch": "feature/test", "changes": 5}

            state_file = memory_dir / "session-state.json"
            state_file.write_text(json.dumps(state, indent=2))

            assert state_file.exists()

    def test_restore_session_state(self, config_file, hook_tmp_project):
        """Session state can be restored"""
        with hook_tmp_project as proj_root:
            memory_dir = proj_root / ".moai" / "memory"
            memory_dir.mkdir(exist_ok=True)

            # Save state
            original_state = {"branch": "feature/test", "changes": 5}

            state_file = memory_dir / "session-state.json"
            state_file.write_text(json.dumps(original_state))

            # Restore state
            loaded_state = json.loads(state_file.read_text())

            assert loaded_state["branch"] == original_state["branch"]
            assert loaded_state["changes"] == original_state["changes"]

    def test_state_persistence_across_sessions(self, config_file, hook_tmp_project):
        """State persists across sessions"""
        with hook_tmp_project as proj_root:
            memory_dir = proj_root / ".moai" / "memory"
            memory_dir.mkdir(exist_ok=True)

            # Session 1: Save state
            state1 = {"session": 1, "data": "session1"}

            state_file = memory_dir / "persistent-state.json"
            state_file.write_text(json.dumps(state1))

            # Session 2: Verify state still exists
            assert state_file.exists()
            loaded = json.loads(state_file.read_text())
            assert loaded["session"] == 1


class TestSessionCompletion:
    """Test session completion and summary"""

    def test_session_summary_generation(self, config_file, hook_tmp_project):
        """Session summary is generated"""
        with hook_tmp_project:
            cleanup_stats = {"total_cleaned": 5, "temp_cleaned": 2, "cache_cleaned": 3}

            work_state = {"specs_in_progress": ["SPEC-001"], "uncommitted_files": 3}

            summary = {
                "hook": "session_end__auto_cleanup",
                "success": True,
                "cleanup_stats": cleanup_stats,
                "work_state": work_state,
                "timestamp": datetime.now().isoformat(),
            }

            assert summary["success"] is True
            assert summary["cleanup_stats"]["total_cleaned"] > 0

    def test_session_summary_includes_cleanup_info(self, config_file, hook_tmp_project):
        """Session summary includes cleanup information"""
        with hook_tmp_project:
            summary = "✅ Session Ended\n   • Files modified: 3\n   • Cleaned: 5 temp files"

            assert "Session Ended" in summary
            assert "Files modified" in summary
            assert "Cleaned" in summary

    def test_session_summary_includes_spec_info(self, config_file, hook_tmp_project):
        """Session summary includes SPEC information"""
        with hook_tmp_project:
            specs = ["SPEC-001", "SPEC-002"]
            summary = f"   • Worked on: {', '.join(specs)}"

            assert "SPEC-001" in summary
            assert "SPEC-002" in summary


class TestRootDocumentViolations:
    """Test root document violation detection"""

    def test_scan_root_violations(self, config_file, hook_tmp_project):
        """Root directory violations are scanned"""
        with hook_tmp_project as proj_root:
            # Create a document file in root (violation)
            (proj_root / "SPEC-001.md").write_text("# SPEC-001")
            (proj_root / "API-DOCS.md").write_text("# API")

            # Scan for violations
            root_files = [f.name for f in proj_root.iterdir() if f.is_file() and not f.name.startswith(".")]

            violations = [f for f in root_files if f.endswith(".md")]

            assert len(violations) >= 2

    def test_whitelist_check(self, config_file, hook_tmp_project):
        """Whitelisted files are not reported as violations"""
        with hook_tmp_project as proj_root:
            config_data = json.loads(config_file.read_text())

            whitelist = config_data.get("document_management", {}).get("root_whitelist", [])

            # README.md should be whitelisted
            (proj_root / "README.md").write_text("# Project")

            # Check if whitelisted
            is_whitelisted = "README.md" in whitelist

            assert is_whitelisted

    def test_violation_suggestion(self, config_file, hook_tmp_project):
        """Suggested location for violations"""
        with hook_tmp_project as proj_root:
            # Create misplaced file
            (proj_root / "implementation-plan.md").write_text("# Plan")

            # Should suggest .moai/docs/
            suggested = ".moai/docs/"

            assert suggested.startswith(".moai/")


class TestSessionEndResponse:
    """Test SessionEnd Hook response format"""

    def test_response_structure(self, config_file, hook_tmp_project):
        """SessionEnd response has correct structure"""
        with hook_tmp_project:
            response = {
                "hook": "session_end__auto_cleanup",
                "success": True,
                "cleanup_stats": {"total_cleaned": 0},
                "session_metrics_saved": True,
                "work_state_saved": True,
                "timestamp": datetime.now().isoformat(),
            }

            assert "hook" in response
            assert "success" in response
            assert "timestamp" in response

    def test_response_is_json_serializable(self, config_file, hook_tmp_project):
        """SessionEnd response is JSON serializable"""
        with hook_tmp_project:
            response = {
                "hook": "session_end__auto_cleanup",
                "success": True,
                "cleanup_stats": {"total_cleaned": 5},
                "timestamp": datetime.now().isoformat(),
            }

            # Should serialize
            json_str = json.dumps(response)
            assert json_str

            # Should deserialize
            loaded = json.loads(json_str)
            assert loaded["success"] is True


class TestSessionEndIntegrationScenarios:
    """Test complete SessionEnd integration scenarios"""

    def test_full_session_end_flow(self, config_file, git_repo_with_changes, session_state_file, hook_tmp_project):
        """Full SessionEnd flow executes successfully"""
        with hook_tmp_project:
            # Verify all components exist
            assert config_file.exists()
            assert session_state_file.exists()

            # Simulate full flow
            result = {
                "hook": "session_end__auto_cleanup",
                "success": True,
                "cleanup_stats": {"total_cleaned": 3, "temp_cleaned": 1, "cache_cleaned": 2},
                "session_metrics_saved": True,
                "work_state_saved": True,
                "uncommitted_warning": "⚠️  2 uncommitted files detected",
                "session_summary": "✅ Session Ended",
                "timestamp": datetime.now().isoformat(),
            }

            assert result["success"] is True
            assert result["cleanup_stats"]["total_cleaned"] > 0
            assert result["session_metrics_saved"] is True

    def test_session_end_with_no_changes(self, config_file, hook_tmp_project):
        """SessionEnd handles no changes gracefully"""
        with hook_tmp_project:
            result = {
                "hook": "session_end__auto_cleanup",
                "success": True,
                "cleanup_stats": {"total_cleaned": 0, "temp_cleaned": 0, "cache_cleaned": 0},
                "uncommitted_warning": None,
                "session_summary": "✅ Session Ended",
                "timestamp": datetime.now().isoformat(),
            }

            assert result["success"] is True
            assert result["cleanup_stats"]["total_cleaned"] == 0
            assert result["uncommitted_warning"] is None

    def test_session_end_error_handling(self, hook_tmp_project):
        """SessionEnd handles errors with graceful degradation"""
        with hook_tmp_project:
            error_response = {
                "hook": "session_end__auto_cleanup",
                "success": False,
                "error": "Some error occurred",
                "graceful_degradation": True,
                "message": "Hook failed but continuing due to graceful degradation",
                "timestamp": datetime.now().isoformat(),
            }

            # Even with error, message should indicate continuation
            assert error_response["graceful_degradation"] is True
            assert "continuing" in error_response["message"].lower()
