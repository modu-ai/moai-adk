"""Tests for legacy log migration functionality."""

import json
from unittest.mock import patch

from moai_adk.cli.commands.update import _migrate_legacy_logs


class TestLegacyLogMigration:
    """Test legacy log migration functionality"""

    def test_no_legacy_files(self, tmp_path):
        """Test migration when no legacy files exist"""
        # Should succeed and create directories
        result = _migrate_legacy_logs(tmp_path)

        assert result is True
        assert (tmp_path / ".moai" / "logs").exists()
        assert (tmp_path / ".moai" / "docs").exists()

    def test_legacy_session_file_migration(self, tmp_path):
        """Test migration of session state file"""
        # Create legacy structure
        legacy_memory = tmp_path / ".moai" / "memory"
        legacy_memory.mkdir(parents=True)
        session_file = legacy_memory / "last-session-state.json"
        session_file.write_text('{"test": "data"}')

        # Touch the file to set access time
        session_file.touch()

        # Run migration
        result = _migrate_legacy_logs(tmp_path)

        assert result is True
        assert (
            tmp_path / ".moai" / "logs" / "sessions" / "last-session-state.json"
        ).exists()

        # Verify content is preserved
        migrated_content = (
            tmp_path / ".moai" / "logs" / "sessions" / "last-session-state.json"
        ).read_text()
        assert '"test": "data"' in migrated_content

    def test_legacy_error_logs_migration(self, tmp_path):
        """Test migration of error logs"""
        # Create legacy error logs
        legacy_error_logs = tmp_path / ".moai" / "error_logs"
        legacy_error_logs.mkdir(parents=True)

        error_file1 = legacy_error_logs / "error1.log"
        error_file1.write_text("Error 1")

        error_file2 = legacy_error_logs / "subdir" / "error2.log"
        error_file2.parent.mkdir()
        error_file2.write_text("Error 2")

        # Run migration
        result = _migrate_legacy_logs(tmp_path)

        assert result is True
        assert (tmp_path / ".moai" / "logs" / "errors" / "error1.log").exists()
        assert (
            tmp_path / ".moai" / "logs" / "errors" / "subdir" / "error2.log"
        ).exists()

    def test_legacy_reports_migration(self, tmp_path):
        """Test migration of reports"""
        # Create legacy reports
        legacy_reports = tmp_path / ".moai" / "reports"
        legacy_reports.mkdir(parents=True)

        report_file = legacy_reports / "report.md"
        report_file.write_text("# Test Report")

        # Run migration
        result = _migrate_legacy_logs(tmp_path)

        assert result is True
        assert (tmp_path / ".moai" / "docs" / "reports" / "report.md").exists()

    def test_skip_existing_files(self, tmp_path):
        """Test that existing files are not overwritten"""
        # Create legacy file
        legacy_memory = tmp_path / ".moai" / "memory"
        legacy_memory.mkdir(parents=True)
        session_file = legacy_memory / "last-session-state.json"
        session_file.write_text('{"legacy": "data"}')

        # Create target file that should not be overwritten
        target_dir = tmp_path / ".moai" / "logs" / "sessions"
        target_dir.mkdir(parents=True)
        target_file = target_dir / "last-session-state.json"
        target_file.write_text('{"existing": "data"}')

        # Run migration
        result = _migrate_legacy_logs(tmp_path)

        assert result is True
        # Target file should remain unchanged
        content = target_file.read_text()
        assert '"existing": "data"' in content
        assert '"legacy": "data"' not in content

    def test_migration_log_creation(self, tmp_path):
        """Test that migration log is created"""
        # Create some legacy files
        legacy_memory = tmp_path / ".moai" / "memory"
        legacy_memory.mkdir(parents=True)
        session_file = legacy_memory / "last-session-state.json"
        session_file.write_text('{"test": "data"}')

        # Run migration
        result = _migrate_legacy_logs(tmp_path)

        assert result is True
        migration_log_path = tmp_path / ".moai" / "logs" / "migration-log.json"
        assert migration_log_path.exists()

        # Verify migration log content
        log_data = json.loads(migration_log_path.read_text())
        assert "migration_timestamp" in log_data
        assert "files_migrated" in log_data
        assert "migration_log" in log_data
        assert log_data["files_migrated"] == 1

    def test_dry_run_mode(self, tmp_path):
        """Test dry run mode without actual file operations"""
        # Create legacy files
        legacy_memory = tmp_path / ".moai" / "memory"
        legacy_memory.mkdir(parents=True)
        session_file = legacy_memory / "last-session-state.json"
        session_file.write_text('{"test": "data"}')

        # Run dry run
        with patch("moai_adk.cli.commands.update.console") as mock_console:
            result = _migrate_legacy_logs(tmp_path, dry_run=True)

        assert result is True
        # Target file should not exist in dry run
        assert not (
            tmp_path / ".moai" / "logs" / "sessions" / "last-session-state.json"
        ).exists()

        # Console should show dry run message
        mock_console.print.assert_any_call(
            "[cyan]🔍 Legacy log migration (dry run):[/cyan]"
        )

    def test_error_handling(self, tmp_path):
        """Test graceful error handling"""
        # Create legacy file with problematic content
        legacy_memory = tmp_path / ".moai" / "memory"
        legacy_memory.mkdir(parents=True)
        session_file = legacy_memory / "last-session-state.json"
        session_file.write_text('{"test": "data"}')

        # Mock shutil.copy2 to raise an exception
        with patch("shutil.copy2", side_effect=Exception("Simulated error")):
            with patch("moai_adk.cli.commands.update.console") as mock_console:
                result = _migrate_legacy_logs(tmp_path)

        assert result is False
        # Should log error message
        mock_console.print.assert_any_call(
            "   [red]✗ Log migration failed: Simulated error[/red]"
        )

    def test_complex_directory_structure(self, tmp_path):
        """Test migration with nested directories"""
        # Create complex legacy structure
        legacy_error_logs = tmp_path / ".moai" / "error_logs"
        legacy_error_logs.mkdir(parents=True)

        # Create nested files
        nested_dir = legacy_error_logs / "level1" / "level2"
        nested_dir.mkdir(parents=True, exist_ok=True)
        error_file = nested_dir / "error.log"
        error_file.write_text("Nested error")

        # Run migration
        result = _migrate_legacy_logs(tmp_path)

        assert result is True
        nested_target = (
            tmp_path / ".moai" / "logs" / "errors" / "level1" / "level2" / "error.log"
        )
        assert nested_target.exists()
        assert nested_target.read_text() == "Nested error"
