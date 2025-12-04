"""
Windows Compatibility Tests

Tests ensure all file operations and path handling work correctly on Windows.
Validates:
1. No hardcoded Unix paths (/tmp)
2. All file operations use UTF-8 encoding
3. Proper temp directory handling
4. Unicode file content support
"""

import json
import os
import tempfile
from pathlib import Path

import pytest


class TestWindowsPathHandling:
    """Test Windows path compatibility"""

    def test_temp_directory_uses_system_temp(self):
        """Ensure temp directory comes from system, not hardcoded /tmp"""
        from moai_adk.core.performance.cache_system import CacheSystem

        cache = CacheSystem()
        temp_dir = tempfile.gettempdir()

        # Cache directory should use system temp dir, not hardcoded /tmp
        assert temp_dir in cache.cache_dir
        assert "/tmp" not in cache.cache_dir or os.name == "posix"

    def test_error_recovery_temp_directory(self):
        """Test that error recovery system uses proper paths without /tmp hardcoding"""
        from moai_adk.core.error_recovery_system import ErrorRecoverySystem

        system = ErrorRecoverySystem()

        # Error log directory should be within project, not hardcoded /tmp
        # Verify no hardcoded /tmp path is used
        assert "/tmp/moai_error_recovery.log" not in str(system.error_log_dir)
        # The system should have an error_log_dir attribute
        assert hasattr(system, "error_log_dir")
        assert system.error_log_dir is not None


class TestFileEncodingUTF8:
    """Test UTF-8 encoding on all file operations"""

    def test_error_recovery_system_write_utf8(self, tmp_path):
        """Test error_recovery_system writes JSON with UTF-8 encoding"""
        from datetime import datetime, timezone

        from moai_adk.core.error_recovery_system import (
            ErrorCategory,
            ErrorRecoverySystem,
            ErrorReport,
            ErrorSeverity,
        )

        # Create system with temp log directory
        system = ErrorRecoverySystem()
        system.error_log_dir = tmp_path

        # Create error report with Unicode content
        error = ErrorReport(
            id="test_001",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.SYSTEM,
            message="테스트 에러 메시지 with émojis 🎯",
            details={"key": "값 with special chars éàü"},
            stack_trace=None,
            context={},
        )

        # Log the error
        system._log_error(error)

        # Verify file was created with UTF-8 encoding
        error_file = tmp_path / f"error_{error.id}.json"
        assert error_file.exists()

        # Read and verify content
        with open(error_file, "r", encoding="utf-8") as f:
            data = json.load(f)
            assert data["message"] == "테스트 에러 메시지 with émojis 🎯"
            assert data["details"]["key"] == "값 with special chars éàü"

    def test_rollback_manager_read_utf8(self, tmp_path):
        """Test rollback_manager reads JSON with UTF-8 encoding"""
        from moai_adk.core.rollback_manager import RollbackManager

        manager = RollbackManager(project_root=tmp_path)

        # Create a registry with Unicode content
        registry_data = {
            "rollback_001": {
                "description": "테스트 롤백 with émojis 🔄",
                "timestamp": "2025-11-16T00:00:00Z",
            }
        }

        manager.registry_file.parent.mkdir(parents=True, exist_ok=True)
        with open(manager.registry_file, "w", encoding="utf-8") as f:
            json.dump(registry_data, f, ensure_ascii=False)

        # Load registry
        loaded = manager._load_registry()
        assert loaded["rollback_001"]["description"] == "테스트 롤백 with émojis 🔄"

    def test_rollback_manager_write_utf8(self, tmp_path):
        """Test rollback_manager writes JSON with UTF-8 encoding"""
        from moai_adk.core.rollback_manager import RollbackManager

        manager = RollbackManager(project_root=tmp_path)

        # Add registry with Unicode content
        manager.registry = {
            "rollback_001": {
                "description": "테스트 with special chars: éàü",
                "timestamp": "2025-11-16T00:00:00Z",
            }
        }

        # Save registry
        manager._save_registry()

        # Verify file was created with UTF-8 encoding
        assert manager.registry_file.exists()

        with open(manager.registry_file, "r", encoding="utf-8") as f:
            data = json.load(f)
            assert data["rollback_001"]["description"] == "테스트 with special chars: éàü"

    def test_backup_manager_read_utf8(self, tmp_path):
        """Test backup_manager reads metadata with UTF-8 encoding"""
        from moai_adk.core.migration.backup_manager import BackupManager

        manager = BackupManager(project_root=tmp_path)

        # Create backup metadata with Unicode content
        backup_dir = manager.backup_base_dir / "test_backup"
        backup_dir.mkdir(parents=True, exist_ok=True)

        metadata = {
            "timestamp": "20251116_000000",
            "description": "테스트 백업 with émojis 📦",
            "backed_up_files": ["config.json"],
        }

        metadata_path = backup_dir / "backup_metadata.json"
        with open(metadata_path, "w", encoding="utf-8") as f:
            json.dump(metadata, f, ensure_ascii=False, indent=2)

        # List backups (reads metadata)
        backups = manager.list_backups()
        assert len(backups) > 0
        assert "테스트 백업 with émojis 📦" in backups[0]["description"]

    def test_backup_manager_write_utf8(self, tmp_path):
        """Test backup_manager writes metadata with UTF-8 encoding"""
        from moai_adk.core.migration.backup_manager import BackupManager

        # Create config file for backup
        config_file = tmp_path / ".moai" / "config" / "config.json"
        config_file.parent.mkdir(parents=True, exist_ok=True)
        with open(config_file, "w", encoding="utf-8") as f:
            json.dump({"project": "테스트 프로젝트"}, f, ensure_ascii=False)

        manager = BackupManager(project_root=tmp_path)
        backup_dir = manager.create_backup(description="테스트 with émojis 🔄")

        # Verify metadata was written with UTF-8 encoding
        metadata_path = backup_dir / "backup_metadata.json"
        assert metadata_path.exists()

        with open(metadata_path, "r", encoding="utf-8") as f:
            data = json.load(f)
            assert data["description"] == "테스트 with émojis 🔄"

    def test_session_manager_read_utf8(self, tmp_path):
        """Test session_manager reads sessions with UTF-8 encoding"""
        from moai_adk.core.session_manager import SessionManager

        session_file = tmp_path / "sessions.json"
        session_data = {
            "sessions": {"agent_1": "session_001"},
            "metadata": {"session_001": {"note": "테스트 세션 with émojis 📝"}},
            "chains": {},
        }

        with open(session_file, "w", encoding="utf-8") as f:
            json.dump(session_data, f, ensure_ascii=False)

        manager = SessionManager(session_file=session_file)
        assert manager._metadata["session_001"]["note"] == "테스트 세션 with émojis 📝"

    def test_session_manager_write_utf8(self, tmp_path):
        """Test session_manager writes sessions with UTF-8 encoding"""
        from moai_adk.core.session_manager import SessionManager

        session_file = tmp_path / "sessions.json"
        manager = SessionManager(session_file=session_file)

        # Add session with Unicode content
        manager._sessions["agent_1"] = "session_001"
        manager._metadata["session_001"] = {"note": "테스트 with éàü"}
        manager._save_sessions()

        # Verify UTF-8 encoding
        assert session_file.exists()
        with open(session_file, "r", encoding="utf-8") as f:
            data = json.load(f)
            assert data["metadata"]["session_001"]["note"] == "테스트 with éàü"

    def test_command_helpers_read_utf8(self, tmp_path):
        """Test command_helpers reads config with UTF-8 encoding"""
        from moai_adk.core.command_helpers import extract_project_metadata

        # Create config with Unicode content
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_file = config_dir / "config.json"

        config_data = {
            "project": {
                "name": "테스트 프로젝트 with émojis 🎯",
                "owner": "GoosLab",
            }
        }

        with open(config_file, "w", encoding="utf-8") as f:
            json.dump(config_data, f, ensure_ascii=False)

        # Extract metadata (should use UTF-8)
        metadata = extract_project_metadata(str(tmp_path))
        assert metadata["project_name"] == "테스트 프로젝트 with émojis 🎯"

    def test_context_manager_write_utf8(self, tmp_path):
        """Test context_manager writes phase result with UTF-8 encoding"""
        from moai_adk.core.context_manager import save_phase_result

        target_path = tmp_path / "phase_result.json"
        data = {"phase": "test", "message": "테스트 결과 with émojis ✅"}

        save_phase_result(data, str(target_path))

        # Verify UTF-8 encoding
        assert target_path.exists()
        with open(target_path, "r", encoding="utf-8") as f:
            loaded = json.load(f)
            assert loaded["message"] == "테스트 결과 with émojis ✅"


class TestWindowsPathSeparators:
    """Test proper path separator handling"""

    def test_rollback_manager_path_conversion(self, tmp_path):
        """Test rollback manager handles path separators correctly"""
        from moai_adk.core.rollback_manager import RollbackManager

        manager = RollbackManager(project_root=tmp_path)

        # Paths should be converted to Path objects for cross-platform support
        assert isinstance(manager.backup_root, Path)
        assert isinstance(manager.registry_file, Path)

    def test_backup_manager_path_conversion(self, tmp_path):
        """Test backup manager handles path separators correctly"""
        from moai_adk.core.migration.backup_manager import BackupManager

        manager = BackupManager(project_root=tmp_path)

        # Paths should be converted to Path objects
        assert isinstance(manager.backup_base_dir, Path)

    def test_session_manager_path_conversion(self, tmp_path):
        """Test session manager handles path separators correctly"""
        from moai_adk.core.session_manager import SessionManager

        session_file = tmp_path / "sessions.json"
        manager = SessionManager(session_file=session_file)

        # Paths should be converted to Path objects
        assert isinstance(manager._session_file, Path)
        assert isinstance(manager._transcript_dir, Path)


class TestUnicodeErrorMessages:
    """Test proper handling of Unicode in error messages"""

    def test_error_recovery_unicode_message(self, tmp_path):
        """Test error messages with Unicode characters are logged correctly"""
        from datetime import datetime, timezone

        from moai_adk.core.error_recovery_system import (
            ErrorCategory,
            ErrorRecoverySystem,
            ErrorReport,
            ErrorSeverity,
        )

        system = ErrorRecoverySystem()
        system.error_log_dir = tmp_path

        # Create error with Unicode message
        error = ErrorReport(
            id="test_001",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.SYSTEM,
            message="시스템 오류 🚨 with émojis",
            details={"error": "값이 없습니다"},
            stack_trace="Stack: 값 에러",
            context={"user": "사용자"},
        )

        system._log_error(error)

        # Verify file exists and contains Unicode
        error_file = tmp_path / f"error_{error.id}.json"
        assert error_file.exists()

        content = error_file.read_text(encoding="utf-8")
        assert "시스템 오류" in content
        assert "émojis" in content
        assert "값이 없습니다" in content


class TestNoHardcodedPaths:
    """Test that no hardcoded Unix paths exist"""

    def test_error_recovery_no_tmp_hardcoding(self):
        """Verify error_recovery_system doesn't hardcode /tmp"""
        import inspect

        from moai_adk.core.error_recovery_system import ErrorRecoverySystem

        source = inspect.getsource(ErrorRecoverySystem.__init__)
        # Should use tempfile.gettempdir(), not "/tmp"
        assert '"/tmp' not in source or "tempfile" in source

    def test_cache_system_no_tmp_hardcoding(self):
        """Verify cache_system doesn't hardcode /tmp"""
        import inspect

        from moai_adk.core.performance.cache_system import CacheSystem

        source = inspect.getsource(CacheSystem)
        # Should use tempfile.gettempdir(), not hardcoded path
        assert "tempfile.gettempdir()" in source


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
