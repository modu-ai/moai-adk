"""
@TEST:BACKUP-SCENARIOS-001 Backup scenario tests for MoAI-ADK

Tests the backup functionality and force overwrite scenarios
to ensure CLAUDE.md files are properly backed up during refactoring mode.
"""

import shutil
import tempfile
from pathlib import Path

import pytest

from moai_adk.cli.helpers import create_installation_backup, detect_potential_conflicts
from moai_adk.config import Config
from moai_adk.install.installer import SimplifiedInstaller


class TestBackupScenarios:
    """Test backup functionality and force overwrite scenarios."""

    def setup_method(self):
        """Set up test environment before each test."""
        # Create temporary directory for testing
        self.temp_dir = Path(tempfile.mkdtemp())
        self.project_dir = self.temp_dir / "test_project"
        self.project_dir.mkdir(parents=True, exist_ok=True)

    def teardown_method(self):
        """Clean up test environment after each test."""
        if self.temp_dir.exists():
            shutil.rmtree(self.temp_dir)

    def create_existing_moai_project(self):
        """Create an existing MoAI project structure for testing."""
        # Create .moai directory
        moai_dir = self.project_dir / ".moai"
        moai_dir.mkdir(exist_ok=True)
        (moai_dir / "config.json").write_text('{"test": "config"}')

        # Create .claude directory
        claude_dir = self.project_dir / ".claude"
        claude_dir.mkdir(exist_ok=True)
        (claude_dir / "settings.json").write_text('{"test": "settings"}')

        # Create CLAUDE.md file
        claude_md = self.project_dir / "CLAUDE.md"
        claude_md.write_text("# Original CLAUDE.md Content\n\nThis is existing content.")

        return moai_dir, claude_dir, claude_md

    def test_backup_creation_with_existing_files(self):
        """Test that backup is properly created when files exist."""
        # Arrange
        moai_dir, claude_dir, claude_md = self.create_existing_moai_project()

        # Act
        result = create_installation_backup(self.project_dir)

        # Assert
        assert result is True, "Backup creation should succeed"

        # Check backup directory exists (now in parent directory)
        backup_pattern = f"{self.project_dir.name}_moai_backup_*"
        backup_dirs = list(self.project_dir.parent.glob(backup_pattern))
        assert len(backup_dirs) == 1, "Exactly one backup directory should be created"

        backup_dir = backup_dirs[0]

        # Verify all items were backed up
        assert (backup_dir / ".moai").exists(), ".moai directory should be backed up"
        assert (backup_dir / ".claude").exists(), ".claude directory should be backed up"
        assert (backup_dir / "CLAUDE.md").exists(), "CLAUDE.md should be backed up"

        # Verify content is preserved
        backup_claude_md = backup_dir / "CLAUDE.md"
        original_content = claude_md.read_text()
        backup_content = backup_claude_md.read_text()
        assert backup_content == original_content, "CLAUDE.md content should be preserved"

        # Check backup info file
        backup_info = backup_dir / "backup_info.txt"
        assert backup_info.exists(), "backup_info.txt should be created"
        info_content = backup_info.read_text()
        assert "CLAUDE.md" in info_content, "Backup info should mention CLAUDE.md"
        assert "Items Backed Up: 3" in info_content, "Should report 3 backed up items"

    def test_backup_empty_project(self):
        """Test backup behavior with empty project directory."""
        # Act
        result = create_installation_backup(self.project_dir)

        # Assert
        assert result is True, "Backup should succeed even with empty project"

        # Check backup directory exists (now in parent directory)
        backup_pattern = f"{self.project_dir.name}_moai_backup_*"
        backup_dirs = list(self.project_dir.parent.glob(backup_pattern))
        assert len(backup_dirs) == 1, "Backup directory should still be created"

        backup_dir = backup_dirs[0]
        backup_info = backup_dir / "backup_info.txt"
        assert backup_info.exists(), "backup_info.txt should be created"

        info_content = backup_info.read_text()
        assert "Items Backed Up: 0" in info_content, "Should report 0 backed up items"

    def test_conflict_detection(self):
        """Test that conflicts are properly detected before installation."""
        # Arrange
        self.create_existing_moai_project()

        # Act
        conflicts = detect_potential_conflicts(self.project_dir)

        # Assert
        assert len(conflicts) > 0, "Should detect conflicts with existing files"
        assert any("CLAUDE.md" in conflict for conflict in conflicts), "Should detect CLAUDE.md conflict"
        assert any(".moai" in conflict for conflict in conflicts), "Should detect .moai conflict"
        assert any(".claude" in conflict for conflict in conflicts), "Should detect .claude conflict"

    def test_force_overwrite_with_backup(self):
        """Test that force overwrite works correctly with backup enabled."""
        # Arrange
        moai_dir, claude_dir, claude_md = self.create_existing_moai_project()
        original_content = "# Original CLAUDE.md Content\n\nThis is existing content."

        config = Config(
            name=self.project_dir.name,
            path=str(self.project_dir),
            force_overwrite=True,
            backup_enabled=True,
            silent=True,
        )

        # Act
        installer = SimplifiedInstaller(config)
        result = installer.install()

        # Assert
        assert result.success, f"Installation should succeed. Errors: {result.errors}"

        # Verify backup was created (now in parent directory)
        backup_pattern = f"{self.project_dir.name}_moai_backup_*"
        backup_dirs = list(self.project_dir.parent.glob(backup_pattern))
        assert len(backup_dirs) >= 1, "Backup should be created when backup_enabled=True"

        if backup_dirs:
            backup_claude_md = backup_dirs[0] / "CLAUDE.md"
            if backup_claude_md.exists():
                backup_content = backup_claude_md.read_text()
                assert backup_content == original_content, "Original CLAUDE.md should be preserved in backup"

        # Verify new CLAUDE.md was installed
        new_claude_md = self.project_dir / "CLAUDE.md"
        assert new_claude_md.exists(), "New CLAUDE.md should be installed"

        new_content = new_claude_md.read_text()
        assert new_content != original_content, "New CLAUDE.md should have different content"
        assert "MoAI-ADK" in new_content, "New CLAUDE.md should contain MoAI-ADK content"

    def test_force_overwrite_without_backup(self):
        """Test that force overwrite works without backup."""
        # Arrange
        moai_dir, claude_dir, claude_md = self.create_existing_moai_project()

        config = Config(
            name=self.project_dir.name,
            path=str(self.project_dir),
            force_overwrite=True,
            backup_enabled=False,
            silent=True,
        )

        # Act
        installer = SimplifiedInstaller(config)
        result = installer.install()

        # Assert
        assert result.success, f"Installation should succeed. Errors: {result.errors}"

        # Verify new CLAUDE.md was installed
        new_claude_md = self.project_dir / "CLAUDE.md"
        assert new_claude_md.exists(), "New CLAUDE.md should be installed"

        new_content = new_claude_md.read_text()
        assert "MoAI-ADK" in new_content, "New CLAUDE.md should contain MoAI-ADK content"

    def test_no_force_no_overwrite(self):
        """Test that files are not overwritten without force flag."""
        # Arrange
        moai_dir, claude_dir, claude_md = self.create_existing_moai_project()
        original_content = claude_md.read_text()

        config = Config(
            name=self.project_dir.name,
            path=str(self.project_dir),
            force_overwrite=False,
            backup_enabled=False,
            silent=True,
        )

        # Act
        installer = SimplifiedInstaller(config)
        result = installer.install()

        # Assert - installation should succeed but files should not be overwritten
        assert result.success, f"Installation should succeed. Errors: {result.errors}"

        # Verify original CLAUDE.md is preserved
        current_content = claude_md.read_text()
        assert current_content == original_content, "Original CLAUDE.md should be preserved when force=False"

    @pytest.mark.parametrize("force,backup,expected_backup_dirs", [
        (True, True, 1),    # Force with backup should create backup
        (True, False, 0),   # Force without backup should not create backup
        (False, True, 1),   # No force with backup should create backup
        (False, False, 0),  # No force without backup should not create backup
    ])
    def test_backup_behavior_combinations(self, force, backup, expected_backup_dirs):
        """Test different combinations of force and backup flags."""
        # Arrange
        self.create_existing_moai_project()

        config = Config(
            name=self.project_dir.name,
            path=str(self.project_dir),
            force_overwrite=force,
            backup_enabled=backup,
            silent=True,
        )

        # Act
        installer = SimplifiedInstaller(config)
        result = installer.install()

        # Assert
        assert result.success, f"Installation should succeed. Errors: {result.errors}"

        # Check backup directory count (now in parent directory)
        backup_pattern = f"{self.project_dir.name}_moai_backup_*"
        backup_dirs = list(self.project_dir.parent.glob(backup_pattern))
        assert len(backup_dirs) == expected_backup_dirs, \
            f"Expected {expected_backup_dirs} backup dirs, got {len(backup_dirs)} for force={force}, backup={backup}"