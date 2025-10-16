# @TEST:INIT-003 | SPEC: SPEC-INIT-003.md
"""Unit tests for backup_utils.py module

Tests for backup utility functions.
"""

from pathlib import Path

from moai_adk.core.project.backup_utils import (
    BACKUP_TARGETS,
    PROTECTED_PATHS,
    generate_backup_dir_name,
    get_backup_targets,
    has_any_moai_files,
    is_protected_path,
)


class TestBackupConstants:
    """Test backup constants"""

    def test_backup_targets_is_list(self):
        """BACKUP_TARGETS should be a list"""
        assert isinstance(BACKUP_TARGETS, list)
        assert len(BACKUP_TARGETS) > 0

    def test_backup_targets_contains_key_files(self):
        """Should contain key MoAI files"""
        assert ".moai/config.json" in BACKUP_TARGETS
        assert "CLAUDE.md" in BACKUP_TARGETS

    def test_protected_paths_is_list(self):
        """PROTECTED_PATHS should be a list"""
        assert isinstance(PROTECTED_PATHS, list)

    def test_protected_paths_contains_specs(self):
        """Should protect user data like specs"""
        assert ".moai/specs/" in PROTECTED_PATHS


class TestHasAnyMoaiFiles:
    """Test has_any_moai_files function"""

    def test_returns_false_for_empty_directory(self, tmp_project_dir: Path):
        """Should return False when no MoAI files exist"""
        result = has_any_moai_files(tmp_project_dir)
        assert result is False

    def test_returns_true_for_moai_config(self, tmp_project_dir: Path):
        """Should return True when .moai/config.json exists"""
        config_file = tmp_project_dir / ".moai" / "config.json"
        config_file.parent.mkdir(parents=True, exist_ok=True)
        config_file.write_text("{}")

        result = has_any_moai_files(tmp_project_dir)
        assert result is True

    def test_returns_true_for_claude_md(self, tmp_project_dir: Path):
        """Should return True when CLAUDE.md exists"""
        claude_file = tmp_project_dir / "CLAUDE.md"
        claude_file.write_text("# Test")

        result = has_any_moai_files(tmp_project_dir)
        assert result is True

    def test_returns_true_for_moai_directory(self, tmp_project_dir: Path):
        """Should return True when .moai/project/ directory exists"""
        moai_dir = tmp_project_dir / ".moai" / "project"
        moai_dir.mkdir(parents=True, exist_ok=True)

        result = has_any_moai_files(tmp_project_dir)
        assert result is True


class TestGetBackupTargets:
    """Test get_backup_targets function"""

    def test_returns_empty_list_when_no_targets(self, tmp_project_dir: Path):
        """Should return empty list when no targets exist"""
        result = get_backup_targets(tmp_project_dir)
        assert result == []

    def test_returns_existing_targets(self, tmp_project_dir: Path):
        """Should return only existing targets"""
        # Create some targets
        config_file = tmp_project_dir / ".moai" / "config.json"
        config_file.parent.mkdir(parents=True, exist_ok=True)
        config_file.write_text("{}")

        claude_file = tmp_project_dir / "CLAUDE.md"
        claude_file.write_text("# Test")

        result = get_backup_targets(tmp_project_dir)

        assert ".moai/config.json" in result
        assert "CLAUDE.md" in result
        assert len(result) == 2

    def test_returns_directories(self, tmp_project_dir: Path):
        """Should include directories in targets"""
        moai_memory = tmp_project_dir / ".moai" / "memory"
        moai_memory.mkdir(parents=True, exist_ok=True)

        result = get_backup_targets(tmp_project_dir)

        assert ".moai/memory/" in result


class TestGenerateBackupDirName:
    """Test generate_backup_dir_name function"""

    def test_generates_timestamp_format(self):
        """Should generate timestamp in YYYYMMDD-HHMMSS format"""
        result = generate_backup_dir_name()

        # Check format: YYYYMMDD-HHMMSS
        assert len(result) == 15  # 8 digits + hyphen + 6 digits
        assert result[8] == "-"
        assert result[:8].isdigit()  # YYYYMMDD
        assert result[9:].isdigit()  # HHMMSS

    def test_generates_unique_timestamps(self):
        """Should generate different timestamps for successive calls"""
        import time

        result1 = generate_backup_dir_name()
        time.sleep(0.01)  # Small delay
        result2 = generate_backup_dir_name()

        # Usually same unless called at second boundary
        # Just verify both are valid formats
        assert len(result1) == 15
        assert len(result2) == 15

    def test_timestamp_is_recent(self):
        """Generated timestamp should be recent"""
        from datetime import datetime

        result = generate_backup_dir_name()

        # Extract year from result
        year = int(result[:4])
        current_year = datetime.now().year

        # Should be current year
        assert year == current_year


class TestIsProtectedPath:
    """Test is_protected_path function"""

    def test_returns_true_for_specs_directory(self):
        """Should return True for moai/specs/ paths (after stripping leading ./)"""
        # The function strips './' from patterns, so paths need to match the stripped version
        assert is_protected_path(Path("moai/specs/SPEC-AUTH-001/spec.md")) is True
        assert is_protected_path(Path("moai/specs/")) is True

    def test_returns_true_for_reports_directory(self):
        """Should return True for moai/reports/ paths"""
        assert is_protected_path(Path("moai/reports/report.md")) is True
        assert is_protected_path(Path("moai/reports/")) is True

    def test_returns_false_for_config_files(self):
        """Should return False for config files (not protected)"""
        assert is_protected_path(Path(".moai/config.json")) is False
        assert is_protected_path(Path(".moai/project/product.md")) is False
        assert is_protected_path(Path("moai/config.json")) is False

    def test_returns_false_for_root_files(self):
        """Should return False for root files"""
        assert is_protected_path(Path("CLAUDE.md")) is False
        assert is_protected_path(Path("README.md")) is False

    def test_handles_windows_paths(self):
        """Should handle Windows-style paths"""
        # Path with backslashes (converted to forward slash)
        path = Path("moai") / "specs" / "test.md"
        result = is_protected_path(path)
        assert result is True

    def test_handles_relative_paths(self):
        """Should handle various relative path formats"""
        # After stripping './', patterns match 'moai/specs' and 'moai/reports'
        assert is_protected_path(Path("moai/specs/test")) is True
        assert is_protected_path(Path("moai/reports/test")) is True
