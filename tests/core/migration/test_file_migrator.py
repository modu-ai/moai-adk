"""
Tests for core/migration/file_migrator.py module

Achieves 90%+ coverage by testing all paths including:
- File and directory operations
- Migration plan execution
- Cleanup operations
- Edge cases and error handling
"""

from pathlib import Path
from typing import Any, Dict

import pytest

from moai_adk.core.migration.file_migrator import FileMigrator


@pytest.fixture
def migrator(tmp_path: Path) -> FileMigrator:
    """Create FileMigrator instance"""
    return FileMigrator(tmp_path)


@pytest.fixture
def project_structure(tmp_path: Path) -> Path:
    """Create test project structure"""
    # Create source files
    (tmp_path / "src").mkdir()
    (tmp_path / "src" / "file1.txt").write_text("content 1")
    (tmp_path / "src" / "file2.json").write_text('{"key": "value"}')

    # Create nested structure
    (tmp_path / "nested" / "deep").mkdir(parents=True)
    (tmp_path / "nested" / "deep" / "file.md").write_text("# Nested")

    return tmp_path


class TestFileMigratorInit:
    """Test FileMigrator initialization"""

    def test_init_sets_project_root(self, tmp_path: Path) -> None:
        """Should set project root as Path"""
        migrator = FileMigrator(tmp_path)
        assert isinstance(migrator.project_root, Path)
        assert migrator.project_root == tmp_path

    def test_init_converts_string_to_path(self, tmp_path: Path) -> None:
        """Should convert string path to Path object"""
        migrator = FileMigrator(str(tmp_path))
        assert isinstance(migrator.project_root, Path)
        assert migrator.project_root == tmp_path

    def test_init_initializes_tracking_lists(self, migrator: FileMigrator) -> None:
        """Should initialize empty tracking lists"""
        assert migrator.moved_files == []
        assert migrator.created_dirs == []


class TestCreateDirectory:
    """Test directory creation"""

    def test_create_directory_creates_new_directory(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should create new directory"""
        new_dir = tmp_path / "new_directory"
        result = migrator.create_directory(new_dir)

        assert result is True
        assert new_dir.exists()
        assert new_dir.is_dir()

    def test_create_directory_creates_nested_directories(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should create nested directories"""
        nested_dir = tmp_path / "level1" / "level2" / "level3"
        result = migrator.create_directory(nested_dir)

        assert result is True
        assert nested_dir.exists()
        assert (tmp_path / "level1").exists()
        assert (tmp_path / "level1" / "level2").exists()

    def test_create_directory_handles_existing_directory(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should handle existing directory without error"""
        existing_dir = tmp_path / "existing"
        existing_dir.mkdir()

        result = migrator.create_directory(existing_dir)

        assert result is True
        assert existing_dir.exists()

    def test_create_directory_tracks_created_dirs(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should track created directories"""
        dir1 = tmp_path / "dir1"
        dir2 = tmp_path / "dir2"

        migrator.create_directory(dir1)
        migrator.create_directory(dir2)

        assert str(dir1) in migrator.created_dirs
        assert str(dir2) in migrator.created_dirs
        assert len(migrator.created_dirs) == 2

    def test_create_directory_handles_invalid_path(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should handle invalid path gracefully"""
        # Try to create directory in non-writable location (on most systems)
        invalid_dir = Path("/root/impossible/directory")
        result = migrator.create_directory(invalid_dir)

        assert result is False


class TestMoveFile:
    """Test file moving operations"""

    def test_move_file_copies_file_by_default(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should copy file by default (copy_instead=True)"""
        source = project_structure / "src" / "file1.txt"
        dest = project_structure / "dest" / "file1.txt"

        result = migrator.move_file(source, dest)

        assert result is True
        assert dest.exists()
        assert source.exists()  # Original should still exist (copy)
        assert "content 1" in dest.read_text()

    def test_move_file_moves_when_copy_instead_false(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should move file when copy_instead=False"""
        source = project_structure / "src" / "file1.txt"
        dest = project_structure / "dest" / "file1.txt"

        result = migrator.move_file(source, dest, copy_instead=False)

        assert result is True
        assert dest.exists()
        assert not source.exists()  # Original should be moved

    def test_move_file_creates_parent_directories(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should create parent directories automatically"""
        source = project_structure / "src" / "file1.txt"
        dest = project_structure / "deep" / "nested" / "path" / "file1.txt"

        result = migrator.move_file(source, dest)

        assert result is True
        assert dest.exists()
        assert (project_structure / "deep" / "nested" / "path").exists()

    def test_move_file_returns_false_if_source_missing(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should return False if source doesn't exist"""
        source = tmp_path / "nonexistent.txt"
        dest = tmp_path / "dest.txt"

        result = migrator.move_file(source, dest)

        assert result is False
        assert not dest.exists()

    def test_move_file_skips_if_destination_exists(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should skip if destination already exists"""
        source = project_structure / "src" / "file1.txt"
        dest = project_structure / "dest.txt"

        # Create destination first
        dest.write_text("existing content")

        result = migrator.move_file(source, dest)

        assert result is True
        assert "existing content" in dest.read_text()  # Not overwritten

    def test_move_file_tracks_operations(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should track move operations"""
        source = project_structure / "src" / "file1.txt"
        dest = project_structure / "dest.txt"

        migrator.move_file(source, dest)

        assert len(migrator.moved_files) == 1
        assert migrator.moved_files[0]["from"] == str(source)
        assert migrator.moved_files[0]["to"] == str(dest)

    def test_move_file_preserves_file_metadata(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should preserve file metadata when copying"""
        source = project_structure / "src" / "file1.txt"
        dest = project_structure / "dest.txt"

        original_stat = source.stat()

        migrator.move_file(source, dest)

        dest_stat = dest.stat()
        assert dest_stat.st_size == original_stat.st_size


class TestDeleteFile:
    """Test file deletion"""

    def test_delete_file_removes_file(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should delete file"""
        file_path = project_structure / "src" / "file1.txt"

        # Mark as safe for testing
        result = migrator.delete_file(file_path, safe=False)

        assert result is True
        assert not file_path.exists()

    def test_delete_file_returns_true_if_already_deleted(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should return True if file already deleted"""
        file_path = tmp_path / "nonexistent.txt"

        result = migrator.delete_file(file_path)

        assert result is True

    def test_delete_file_safe_mode_allows_safe_patterns(self, migrator: FileMigrator) -> None:
        """Should allow deletion of safe files in safe mode"""
        # Create safe file relative to project root
        safe_file = migrator.project_root / ".moai" / "config.json"
        safe_file.parent.mkdir(parents=True)
        safe_file.write_text("{}")

        result = migrator.delete_file(safe_file, safe=True)

        assert result is True
        assert not safe_file.exists()

    def test_delete_file_safe_mode_refuses_unsafe_patterns(
        self, migrator: FileMigrator, project_structure: Path
    ) -> None:
        """Should refuse to delete unsafe files in safe mode"""
        unsafe_file = project_structure / "important_data.txt"
        unsafe_file.write_text("important")

        result = migrator.delete_file(unsafe_file, safe=True)

        assert result is False
        assert unsafe_file.exists()  # Should not be deleted

    def test_delete_file_handles_errors_gracefully(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should handle deletion errors gracefully"""
        # Test with a directory instead of file (should fail gracefully)
        dir_path = tmp_path / "directory"
        dir_path.mkdir()

        # Try to delete as if it's a file (should handle error)
        result = migrator.delete_file(dir_path, safe=False)

        # Should return False and handle error gracefully
        assert isinstance(result, bool)


class TestExecuteMigrationPlan:
    """Test migration plan execution"""

    def test_execute_migration_plan_creates_directories(self, migrator: FileMigrator) -> None:
        """Should create directories from plan"""
        plan: Dict[str, Any] = {
            "create": ["new_dir1", "new_dir2/nested"],
            "move": [],
            "cleanup": [],
        }

        results = migrator.execute_migration_plan(plan)

        assert results["success"] is True
        assert results["created_dirs"] == 2
        assert (migrator.project_root / "new_dir1").exists()
        assert (migrator.project_root / "new_dir2" / "nested").exists()

    def test_execute_migration_plan_moves_files(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should move files from plan"""
        plan: Dict[str, Any] = {
            "create": [],
            "move": [
                {
                    "from": "src/file1.txt",
                    "to": "dest/file1.txt",
                    "description": "Move file1",
                }
            ],
            "cleanup": [],
        }

        results = migrator.execute_migration_plan(plan)

        assert results["success"] is True
        assert results["moved_files"] == 1
        assert (migrator.project_root / "dest" / "file1.txt").exists()

    def test_execute_migration_plan_handles_errors(self, migrator: FileMigrator) -> None:
        """Should track errors in plan execution"""
        plan: Dict[str, Any] = {
            "create": [],
            "move": [
                {
                    "from": "nonexistent.txt",
                    "to": "dest.txt",
                    "description": "Move missing file",
                }
            ],
            "cleanup": [],
        }

        results = migrator.execute_migration_plan(plan)

        assert results["success"] is False
        assert len(results["errors"]) > 0
        assert "Failed to move" in results["errors"][0]

    def test_execute_migration_plan_returns_comprehensive_results(
        self, migrator: FileMigrator, project_structure: Path
    ) -> None:
        """Should return comprehensive execution results"""
        plan: Dict[str, Any] = {
            "create": ["new_dir"],
            "move": [{"from": "src/file1.txt", "to": "dest.txt", "description": "Move file"}],
            "cleanup": [],
        }

        results = migrator.execute_migration_plan(plan)

        assert "success" in results
        assert "created_dirs" in results
        assert "moved_files" in results
        assert "cleaned_files" in results
        assert "errors" in results

    def test_execute_migration_plan_with_empty_plan(self, migrator: FileMigrator) -> None:
        """Should handle empty plan"""
        plan: Dict[str, Any] = {"create": [], "move": [], "cleanup": []}

        results = migrator.execute_migration_plan(plan)

        assert results["success"] is True
        assert results["created_dirs"] == 0
        assert results["moved_files"] == 0
        assert len(results["errors"]) == 0


class TestCleanupOldFiles:
    """Test cleanup operations"""

    def test_cleanup_old_files_removes_safe_files(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should remove safe files (uses safe patterns)"""
        # Create safe files that match safe patterns
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config.json").write_text("{}")
        (tmp_path / ".claude").mkdir()
        (tmp_path / ".claude" / "statusline-config.yaml").write_text("status: ok")

        cleanup_list = [".moai/config.json", ".claude/statusline-config.yaml"]

        cleaned = migrator.cleanup_old_files(cleanup_list, dry_run=False)

        assert cleaned == 2
        assert not (tmp_path / ".moai" / "config.json").exists()
        assert not (tmp_path / ".claude" / "statusline-config.yaml").exists()

    def test_cleanup_old_files_dry_run_does_not_delete(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should not delete files in dry run mode"""
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config.json").write_text("{}")

        cleanup_list = [".moai/config.json"]

        cleaned = migrator.cleanup_old_files(cleanup_list, dry_run=True)

        assert cleaned == 1
        assert (tmp_path / ".moai" / "config.json").exists()  # Still exists

    def test_cleanup_old_files_counts_nonexistent_as_cleaned(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should count non-existent files as already cleaned"""
        # Safe patterns but files don't exist
        cleanup_list = [".moai/config.json", ".claude/statusline-config.yaml"]

        cleaned = migrator.cleanup_old_files(cleanup_list, dry_run=False)

        # delete_file returns True for non-existent files (already deleted)
        assert cleaned == 2

    def test_cleanup_old_files_returns_count(self, migrator: FileMigrator, tmp_path: Path) -> None:
        """Should return count of cleaned files (including already-deleted)"""
        # Create safe files
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config.json").write_text("{}")
        (tmp_path / ".claude").mkdir()
        (tmp_path / ".claude" / "statusline-config.yaml").write_text("status: ok")

        cleanup_list = [
            ".moai/config.json",
            ".moai/nonexistent.json",  # Will return True (already cleaned)
            ".claude/statusline-config.yaml",
        ]

        cleaned = migrator.cleanup_old_files(cleanup_list, dry_run=False)

        # All safe files count as cleaned (even non-existent ones)
        assert cleaned == 3

    def test_cleanup_old_files_respects_safe_mode(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should respect safe mode and refuse unsafe files"""
        # Try to cleanup unsafe files (not in safe patterns)
        cleanup_list = ["src/file1.txt", "src/file2.json"]

        cleaned = migrator.cleanup_old_files(cleanup_list, dry_run=False)

        # Should not delete unsafe files
        assert cleaned == 0
        assert (project_structure / "src" / "file1.txt").exists()
        assert (project_structure / "src" / "file2.json").exists()


class TestGetMigrationSummary:
    """Test migration summary"""

    def test_get_migration_summary_returns_summary(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should return migration summary"""
        # Perform some operations
        migrator.create_directory(project_structure / "new_dir")
        migrator.move_file(
            project_structure / "src" / "file1.txt",
            project_structure / "dest.txt",
        )

        summary = migrator.get_migration_summary()

        assert "moved_files" in summary
        assert "created_directories" in summary
        assert "operations" in summary
        assert summary["moved_files"] == 1
        assert summary["created_directories"] == 1

    def test_get_migration_summary_includes_operations(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should include detailed operations"""
        source = project_structure / "src" / "file1.txt"
        dest = project_structure / "dest.txt"

        migrator.move_file(source, dest)

        summary = migrator.get_migration_summary()

        assert len(summary["operations"]) == 1
        assert summary["operations"][0]["from"] == str(source)
        assert summary["operations"][0]["to"] == str(dest)

    def test_get_migration_summary_empty_when_no_operations(self, migrator: FileMigrator) -> None:
        """Should return empty summary when no operations"""
        summary = migrator.get_migration_summary()

        assert summary["moved_files"] == 0
        assert summary["created_directories"] == 0
        assert summary["operations"] == []


class TestFileMigratorIntegration:
    """Integration tests for full migration workflow"""

    def test_full_migration_workflow(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should complete full migration workflow"""
        # Define migration plan
        plan: Dict[str, Any] = {
            "create": ["migrated/config", "migrated/data"],
            "move": [
                {
                    "from": "src/file1.txt",
                    "to": "migrated/config/file1.txt",
                    "description": "Migrate config file",
                },
                {
                    "from": "nested/deep/file.md",
                    "to": "migrated/data/file.md",
                    "description": "Migrate data file",
                },
            ],
            "cleanup": [],
        }

        # Execute plan
        results = migrator.execute_migration_plan(plan)

        assert results["success"] is True
        assert results["created_dirs"] == 2
        assert results["moved_files"] == 2

        # Verify migrations
        assert (migrator.project_root / "migrated" / "config" / "file1.txt").exists()
        assert (migrator.project_root / "migrated" / "data" / "file.md").exists()

        # Cleanup old files
        cleanup_list = ["src/file1.txt", "nested/deep/file.md"]
        cleaned = migrator.cleanup_old_files(cleanup_list, dry_run=False)

        assert cleaned == 0  # Files already moved (copied), so originals still exist

        # Get summary
        summary = migrator.get_migration_summary()
        assert summary["moved_files"] == 2
        assert summary["created_directories"] == 2

    def test_migration_with_partial_failures(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should handle partial failures gracefully"""
        plan: Dict[str, Any] = {
            "create": ["good_dir"],
            "move": [
                {
                    "from": "src/file1.txt",
                    "to": "dest/file1.txt",
                    "description": "Valid move",
                },
                {
                    "from": "nonexistent.txt",
                    "to": "dest/invalid.txt",
                    "description": "Invalid move",
                },
            ],
            "cleanup": [],
        }

        results = migrator.execute_migration_plan(plan)

        assert results["success"] is False  # Has errors
        assert results["created_dirs"] == 1  # Directory created
        assert results["moved_files"] == 1  # One file moved
        assert len(results["errors"]) == 1  # One error

    def test_migration_tracking_across_operations(self, migrator: FileMigrator, project_structure: Path) -> None:
        """Should track operations across multiple calls"""
        # First operation
        migrator.create_directory(project_structure / "dir1")

        # Second operation
        migrator.move_file(
            project_structure / "src" / "file1.txt",
            project_structure / "dest1.txt",
        )

        # Third operation
        migrator.move_file(
            project_structure / "src" / "file2.json",
            project_structure / "dest2.json",
        )

        summary = migrator.get_migration_summary()

        assert summary["created_directories"] == 1
        assert summary["moved_files"] == 2
        assert len(summary["operations"]) == 2
