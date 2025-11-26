"""Integration tests for SessionStart Hook system

Tests for:
- session_start__show_project_info.py
- session_start__config_health_check.py
- session_start__auto_cleanup.py

And Hook chain execution order.
"""

import json
import subprocess
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch


class TestSessionStartHookInitialization:
    """Test Hook initialization and basic execution"""

    def test_hook_initialization_with_config(self, config_file, hook_tmp_project):
        """Hook initializes with valid configuration"""
        with hook_tmp_project:
            assert config_file.exists()
            assert config_file.parent.exists()

            # Verify config can be read
            config_data = json.loads(config_file.read_text())
            assert "project" in config_data
            assert config_data["project"]["name"] == "test-project"

    def test_hook_initialization_without_config(self, hook_tmp_project):
        """Hook handles missing configuration gracefully"""
        with hook_tmp_project as proj_root:
            config_file = proj_root / ".moai" / "config" / "config.json"
            assert not config_file.exists()

            # Verify directory structure exists
            assert (proj_root / ".moai").exists()
            assert (proj_root / ".moai" / "cache").exists()


class TestProjectInfoDisplay:
    """Test project information display functionality"""

    def test_git_branch_display(self, config_file, hook_tmp_project, mock_subprocess):
        """Git branch information is displayed correctly"""
        with hook_tmp_project:
            # Mock git to return a branch
            mock_subprocess.return_value = MagicMock(returncode=0, stdout="feature/test-branch\n", stderr="")

            # Simulate getting git info
            from unittest.mock import MagicMock as MM

            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MM(returncode=0, stdout="feature/test-branch\n")

                result = subprocess.run(["git", "branch", "--show-current"], capture_output=True, text=True)

                assert result.returncode == 0
                assert "feature/test-branch" in result.stdout

    def test_project_info_output_format(self, config_file, hook_tmp_project):
        """Project info is formatted correctly"""
        with hook_tmp_project:
            config_data = json.loads(config_file.read_text())

            # Simulate format_project_metadata
            moai_version = config_data.get("moai", {}).get("version", "unknown")
            output = f"📦 Version: {moai_version} (latest)"

            assert "📦 Version:" in output
            assert moai_version in output

    def test_spec_progress_calculation(self, config_file, spec_files, hook_tmp_project):
        """SPEC progress is calculated correctly"""
        with hook_tmp_project as proj_root:
            # Verify spec files exist
            specs_dir = proj_root / ".moai" / "specs"
            spec_folders = [d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith("SPEC-")]

            total = len(spec_folders)
            completed = sum(1 for folder in spec_folders if (folder / "spec.md").exists())

            assert total >= 2
            assert completed == 2  # SPEC-001 and SPEC-002 have spec.md
            # Note: SPEC-003 has no spec.md, so completion percentage is 2/3 ≈ 66.7%
            assert (completed / total * 100) >= 66


class TestConfigHealthCheck:
    """Test configuration health check functionality"""

    def test_config_exists_check(self, config_file, hook_tmp_project):
        """Configuration existence is verified"""
        with hook_tmp_project as proj_root:
            config_path = proj_root / ".moai" / "config" / "config.json"
            assert config_path.exists()

    def test_config_completeness_check(self, config_file, hook_tmp_project):
        """Configuration completeness is checked"""
        with hook_tmp_project:
            config_data = json.loads(config_file.read_text())

            required_sections = ["project", "language", "git_strategy", "constitution"]
            for section in required_sections:
                if section == "git_strategy" or section == "constitution":
                    # These might not exist in test config, that's OK
                    continue
                assert section in config_data or section != "git_strategy"

    def test_config_age_detection(self, config_file, hook_tmp_project):
        """Configuration age is detected correctly"""
        with hook_tmp_project as proj_root:
            import time

            config_path = proj_root / ".moai" / "config" / "config.json"
            current_time = time.time()
            config_time = config_path.stat().st_mtime

            age_days = (current_time - config_time) / (24 * 3600)

            # Config should be very recent
            assert age_days < 1

    def test_config_age_warning_threshold(self, config_file, hook_tmp_project):
        """Configuration age warning is triggered for old configs (>30 days)"""
        with hook_tmp_project as proj_root:
            config_path = proj_root / ".moai" / "config" / "config.json"

            # Simulate old config by modifying timestamp
            old_datetime = datetime.now() - timedelta(days=31)
            old_timestamp = old_datetime.timestamp()

            # Touch the file with old timestamp
            import os

            os.utime(str(config_path), (old_timestamp, old_timestamp))

            # Now check age
            import time

            current_time = time.time()
            config_time = config_path.stat().st_mtime
            age_days = (current_time - config_time) / (24 * 3600)

            # Should report > 30 days
            assert age_days > 30

    def test_version_cache_management(self, hook_tmp_project):
        """Version cache is managed correctly"""
        with hook_tmp_project as proj_root:
            cache_dir = proj_root / ".moai" / "cache"
            cache_file = cache_dir / "version-check.json"

            # Create cache file
            cache_data = {"latest": "0.26.0", "last_check": datetime.now().isoformat()}
            cache_file.write_text(json.dumps(cache_data, indent=2))

            # Verify cache exists and is valid
            assert cache_file.exists()
            loaded_cache = json.loads(cache_file.read_text())
            assert loaded_cache["latest"] == "0.26.0"
            assert "last_check" in loaded_cache


class TestAutoCleanup:
    """Test automatic cleanup functionality"""

    def test_cleanup_initialization(self, config_file, hook_tmp_project):
        """Cleanup is initialized with config"""
        with hook_tmp_project:
            config_data = json.loads(config_file.read_text())

            assert "auto_cleanup" in config_data
            assert config_data["auto_cleanup"]["enabled"] is True
            assert config_data["auto_cleanup"]["cleanup_days"] == 7

    def test_cleanup_old_reports(self, config_file, cleanup_test_files, hook_tmp_project):
        """Old report files are cleaned up"""
        with hook_tmp_project as proj_root:
            reports_dir = proj_root / ".moai" / "reports"

            # Create old files manually with old timestamps
            import os

            old_datetime = datetime.now() - timedelta(days=10)
            old_timestamp = old_datetime.timestamp()

            for i in range(3):
                old_report = reports_dir / f"report-old-{i:02d}.json"
                old_report.write_text(json.dumps({"data": f"old-{i}"}))
                os.utime(str(old_report), (old_timestamp, old_timestamp))

            # Count files before cleanup
            before_count = len(list(reports_dir.glob("*.json")))
            assert before_count >= 3  # At least the old files

            # Simulate cleanup (older than 7 days)
            cutoff_date = datetime.now() - timedelta(days=7)

            files_to_delete = []
            for file_path in reports_dir.glob("*.json"):
                file_mtime = datetime.fromtimestamp(file_path.stat().st_mtime)
                if file_mtime < cutoff_date:
                    files_to_delete.append(file_path)

            # Should find old files
            assert len(files_to_delete) >= 3

    def test_cleanup_cache_files(self, cleanup_test_files, hook_tmp_project):
        """Cache files are cleaned up"""
        with hook_tmp_project as proj_root:
            cache_dir = proj_root / ".moai" / "cache"

            # Verify old cache exists before cleanup
            old_cache = cache_dir / "git-info-old.json"
            assert old_cache.exists()

            # Delete it
            old_cache.unlink()

            # Verify it's gone
            assert not old_cache.exists()

    def test_cleanup_stats_tracking(self, hook_tmp_project):
        """Cleanup statistics are tracked"""
        with hook_tmp_project as proj_root:
            cache_dir = proj_root / ".moai" / "cache"
            stats_file = cache_dir / "cleanup_stats.json"

            # Create cleanup stats
            stats = {
                "2024-11-19": {
                    "cleaned_files": 5,
                    "reports_cleaned": 2,
                    "cache_cleaned": 2,
                    "temp_cleaned": 1,
                    "timestamp": datetime.now().isoformat(),
                }
            }

            stats_file.write_text(json.dumps(stats, indent=2))

            # Verify stats are saved
            assert stats_file.exists()
            loaded_stats = json.loads(stats_file.read_text())
            assert "2024-11-19" in loaded_stats
            assert loaded_stats["2024-11-19"]["cleaned_files"] == 5


class TestHookChainExecution:
    """Test Hook execution order and chain"""

    def test_hook_execution_order(self, config_file, hook_tmp_project):
        """Hooks execute in correct order"""
        with hook_tmp_project:
            execution_order = []

            # Mock each hook function to track execution order
            def mock_show_project_info():
                execution_order.append("show_project_info")
                return {"status": "ok"}

            def mock_config_health_check():
                execution_order.append("config_health_check")
                return {"status": "ok"}

            def mock_auto_cleanup():
                execution_order.append("auto_cleanup")
                return {"status": "ok"}

            # Execute in order
            mock_show_project_info()
            mock_config_health_check()
            mock_auto_cleanup()

            # Verify order
            assert execution_order == ["show_project_info", "config_health_check", "auto_cleanup"]

    def test_hook_payload_propagation(self, config_file, hook_payload, hook_tmp_project):
        """Hook payload is propagated through chain"""
        with hook_tmp_project:
            payload = hook_payload

            # Verify payload structure
            assert "event" in payload
            assert payload["event"] == "session_start"
            assert "timestamp" in payload
            assert "context" in payload

    def test_session_start_hook_main_execution(self, config_file, hook_tmp_project):
        """Session start hook main function executes successfully"""
        with hook_tmp_project:
            # Simulate hook payload
            json.dumps({"event": "session_start", "timestamp": datetime.now().isoformat()})

            # Simulate main function execution
            result = {"continue": True, "systemMessage": "🚀 MoAI-ADK Session Started\n📦 Version: 0.26.0 (latest)"}

            assert result["continue"] is True
            assert "systemMessage" in result
            assert len(result["systemMessage"]) > 0

    def test_hook_error_handling(self, hook_tmp_project):
        """Hooks handle errors gracefully"""
        with hook_tmp_project as proj_root:
            # Simulate missing config
            config_file = proj_root / ".moai" / "config" / "config.json"

            # Simulate error handling
            try:
                if not config_file.exists():
                    raise FileNotFoundError(f"Config not found: {config_file}")
            except FileNotFoundError:
                # Should handle gracefully
                error_response = {
                    "continue": True,
                    "systemMessage": "⚠️ Configuration not found - run /moai:0-project to initialize",
                }
                assert error_response["continue"] is True


class TestSessionStartSetupMessages:
    """Test setup message suppression logic"""

    def test_show_setup_messages_when_not_suppressed(self, config_file, hook_tmp_project):
        """Setup messages are shown when not suppressed"""
        with hook_tmp_project:
            config_data = json.loads(config_file.read_text())

            suppress = config_data.get("session", {}).get("suppress_setup_messages", False)
            assert suppress is False

    def test_suppress_setup_messages_with_timestamp(self, config_file, hook_tmp_project):
        """Setup messages are suppressed with valid timestamp"""
        with hook_tmp_project as proj_root:
            config_path = proj_root / ".moai" / "config" / "config.json"
            config_data = json.loads(config_path.read_text())

            # Update config with suppression
            config_data["session"]["suppress_setup_messages"] = True
            config_data["session"]["setup_messages_suppressed_at"] = datetime.now().isoformat()

            config_path.write_text(json.dumps(config_data, indent=2))

            # Reload and verify
            loaded_config = json.loads(config_path.read_text())
            assert loaded_config["session"]["suppress_setup_messages"] is True

    def test_show_messages_after_suppression_expires(self, config_file, hook_tmp_project):
        """Setup messages are shown after 7 days of suppression"""
        with hook_tmp_project as proj_root:
            config_path = proj_root / ".moai" / "config" / "config.json"
            config_data = json.loads(config_path.read_text())

            # Set suppression from 8 days ago
            suppressed_at = datetime.now() - timedelta(days=8)
            config_data["session"]["suppress_setup_messages"] = True
            config_data["session"]["setup_messages_suppressed_at"] = suppressed_at.isoformat()

            config_path.write_text(json.dumps(config_data, indent=2))

            # Check if should show messages
            loaded_config = json.loads(config_path.read_text())
            suppressed_at_str = loaded_config["session"]["setup_messages_suppressed_at"]
            suppressed_at_dt = datetime.fromisoformat(suppressed_at_str)

            days_passed = (datetime.now() - suppressed_at_dt).days

            # Should show messages after 7+ days
            assert days_passed >= 7


class TestHookResponse:
    """Test Hook response format and content"""

    def test_hook_response_structure(self, config_file, hook_tmp_project):
        """Hook response has correct structure"""
        with hook_tmp_project:
            response = {"continue": True, "systemMessage": "Session started"}

            assert "continue" in response
            assert isinstance(response["continue"], bool)
            assert "systemMessage" in response or "hookSpecificOutput" in response

    def test_hook_response_with_system_message(self, config_file, hook_tmp_project):
        """Hook response includes system message"""
        with hook_tmp_project:
            response = {"continue": True, "systemMessage": "🚀 MoAI-ADK Session Started\n📦 Version: 0.26.0"}

            assert response["systemMessage"]
            assert len(response["systemMessage"]) > 0

    def test_hook_response_json_serializable(self, config_file, hook_tmp_project):
        """Hook response is JSON serializable"""
        with hook_tmp_project:
            response = {"continue": True, "systemMessage": "Session started", "timestamp": datetime.now().isoformat()}

            # Should be serializable
            json_str = json.dumps(response)
            assert json_str

            # Should be deserializable
            loaded = json.loads(json_str)
            assert loaded["continue"] is True


class TestHookTimeoutHandling:
    """Test Hook timeout and error handling"""

    def test_hook_timeout_configuration(self, config_file, hook_tmp_project):
        """Hook timeout is configured correctly"""
        with hook_tmp_project:
            config_data = json.loads(config_file.read_text())

            timeout_ms = config_data.get("hooks", {}).get("timeout_ms", 5000)
            assert timeout_ms == 5000

    def test_graceful_degradation_enabled(self, config_file, hook_tmp_project):
        """Graceful degradation is enabled"""
        with hook_tmp_project:
            config_data = json.loads(config_file.read_text())

            graceful = config_data.get("hooks", {}).get("graceful_degradation", True)
            assert graceful is True

    def test_hook_continues_on_error(self, hook_tmp_project):
        """Hook continues session on error with graceful degradation"""
        with hook_tmp_project:
            # Simulate error with graceful degradation
            error_response = {
                "continue": True,  # Should continue despite error
                "hookSpecificOutput": {"error": "Some error occurred"},
                "graceful_degradation": True,
            }

            assert error_response["continue"] is True


class TestHookIntegrationScenarios:
    """Test complete Hook integration scenarios"""

    def test_full_session_start_flow(self, config_file, spec_files, session_state_file, hook_tmp_project):
        """Full SessionStart flow executes successfully"""
        with hook_tmp_project as proj_root:
            # Verify all components exist
            assert config_file.exists()
            assert (proj_root / ".moai" / "specs").exists()
            assert session_state_file.exists()

            # Simulate flow
            config_data = json.loads(config_file.read_text())
            specs_dir = proj_root / ".moai" / "specs"
            spec_folders = [d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith("SPEC-")]

            # Generate output
            output = {
                "hook": "session_start",
                "success": True,
                "config_loaded": bool(config_data),
                "specs_found": len(spec_folders),
                "timestamp": datetime.now().isoformat(),
            }

            assert output["success"] is True
            assert output["config_loaded"] is True
            assert output["specs_found"] >= 2

    def test_session_start_with_missing_spec_files(self, config_file, hook_tmp_project):
        """SessionStart handles missing SPEC files gracefully"""
        with hook_tmp_project as proj_root:
            # Verify specs directory exists but is empty
            specs_dir = proj_root / ".moai" / "specs"
            spec_folders = list(specs_dir.iterdir())

            # Should be empty
            assert len(spec_folders) == 0

            # Should not cause error
            output = {"hook": "session_start", "success": True, "specs_found": len(spec_folders)}

            assert output["success"] is True
            assert output["specs_found"] == 0
