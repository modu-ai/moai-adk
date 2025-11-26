"""Integration tests for PhaseExecutor

Tests 5-phase installation workflow:
- Phase 1: Preparation and backup
- Phase 2: Directory creation
- Phase 3: Resource installation
- Phase 4: Configuration generation
- Phase 5: Validation and finalization
"""

import json
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from moai_adk.core.project.phase_executor import PhaseExecutor
from moai_adk.core.project.validator import ProjectValidator


@pytest.fixture
def validator(tmp_path: Path) -> ProjectValidator:
    """Create ProjectValidator instance"""
    return ProjectValidator()


@pytest.fixture
def executor(validator: ProjectValidator) -> PhaseExecutor:
    """Create PhaseExecutor instance"""
    return PhaseExecutor(validator)


class TestPhaseExecutorInit:
    """Test PhaseExecutor initialization"""

    def test_init_sets_validator(self, validator: ProjectValidator) -> None:
        """Should set validator instance"""
        executor = PhaseExecutor(validator)
        assert executor.validator is validator

    def test_init_sets_total_phases(self, validator: ProjectValidator) -> None:
        """Should set total phases to 5"""
        executor = PhaseExecutor(validator)
        assert executor.total_phases == 5

    def test_init_sets_current_phase(self, validator: ProjectValidator) -> None:
        """Should initialize current phase to 0"""
        executor = PhaseExecutor(validator)
        assert executor.current_phase == 0


class TestPreparationPhase:
    """Test Phase 1: Preparation and backup"""

    def test_preparation_phase_updates_current_phase(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should update current phase to 1"""
        executor.execute_preparation_phase(tmp_path, backup_enabled=False)
        assert executor.current_phase == 1

    def test_preparation_phase_calls_progress_callback(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should call progress callback with phase 1"""
        callback = Mock()
        executor.execute_preparation_phase(tmp_path, backup_enabled=False, progress_callback=callback)
        callback.assert_called_once_with("Phase 1: Preparation and backup...", 1, 5)

    def test_preparation_phase_validates_system(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should validate system requirements"""
        with patch.object(executor.validator, "validate_system_requirements") as mock_validate:
            executor.execute_preparation_phase(tmp_path, backup_enabled=False)
            mock_validate.assert_called_once()

    def test_preparation_phase_validates_project_path(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should validate project path"""
        with patch.object(executor.validator, "validate_project_path") as mock_validate:
            executor.execute_preparation_phase(tmp_path, backup_enabled=False)
            mock_validate.assert_called_once_with(tmp_path)

    def test_preparation_phase_creates_backup_when_files_exist(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should create backup when MoAI files exist"""
        # Create existing .moai directory with config
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config").mkdir()
        (tmp_path / ".moai" / "config" / "config.json").write_text("{}")

        executor.execute_preparation_phase(tmp_path, backup_enabled=True)

        # Backup should be created in .moai-backups/ directory
        backup_dir = tmp_path / ".moai-backups"
        assert backup_dir.exists()
        backup_timestamps = list(backup_dir.iterdir())
        assert len(backup_timestamps) > 0
        assert backup_timestamps[0].is_dir()

    def test_preparation_phase_skips_backup_when_disabled(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should skip backup when backup_enabled=False"""
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config").mkdir()
        (tmp_path / ".moai" / "config" / "config.json").write_text("{}")

        executor.execute_preparation_phase(tmp_path, backup_enabled=False)

        # No backup should be created
        backup_dir = tmp_path / ".moai" / "backups"
        if backup_dir.exists():
            assert len(list(backup_dir.iterdir())) == 0


class TestDirectoryPhase:
    """Test Phase 2: Directory creation"""

    def test_directory_phase_updates_current_phase(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should update current phase to 2"""
        executor.execute_directory_phase(tmp_path)
        assert executor.current_phase == 2

    def test_directory_phase_calls_progress_callback(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should call progress callback with phase 2"""
        callback = Mock()
        executor.execute_directory_phase(tmp_path, progress_callback=callback)
        callback.assert_called_once_with("Phase 2: Creating directory structure...", 2, 5)

    def test_directory_phase_creates_required_directories(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should create all required directories"""
        executor.execute_directory_phase(tmp_path)

        # Check all required directories exist
        for directory in executor.REQUIRED_DIRECTORIES:
            dir_path = tmp_path / directory
            assert dir_path.exists(), f"Directory {directory} should exist"
            assert dir_path.is_dir(), f"{directory} should be a directory"

    def test_directory_phase_creates_nested_directories(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should create nested directory structures"""
        executor.execute_directory_phase(tmp_path)

        # Check nested structures
        assert (tmp_path / ".moai" / "project").exists()
        assert (tmp_path / ".moai" / "specs").exists()
        assert (tmp_path / ".moai" / "reports").exists()
        assert (tmp_path / ".moai" / "memory").exists()
        assert (tmp_path / ".claude" / "logs").exists()
        assert (tmp_path / ".github").exists()


class TestResourcePhase:
    """Test Phase 3: Resource installation"""

    def test_resource_phase_updates_current_phase(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should update current phase to 3"""
        executor.execute_resource_phase(tmp_path)
        assert executor.current_phase == 3

    def test_resource_phase_calls_progress_callback(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should call progress callback with phase 3"""
        callback = Mock()
        executor.execute_resource_phase(tmp_path, progress_callback=callback)
        callback.assert_called_once_with("Phase 3: Installing resources...", 3, 5)

    def test_resource_phase_returns_created_files(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should return list of created files"""
        created_files = executor.execute_resource_phase(tmp_path)

        assert isinstance(created_files, list)
        assert len(created_files) > 0
        assert ".claude/" in created_files
        assert ".moai/" in created_files
        assert ".github/" in created_files
        assert "CLAUDE.md" in created_files

    def test_resource_phase_copies_template_files(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should copy template files to project"""
        # Execute resource phase
        created_files = executor.execute_resource_phase(tmp_path)

        # Should return list of created files
        assert isinstance(created_files, list)
        assert len(created_files) > 0

        # Key template directories/files should be listed
        assert any(".claude" in f for f in created_files)
        assert any(".moai" in f for f in created_files)
        assert any(".github" in f for f in created_files)
        assert any("CLAUDE.md" in f for f in created_files)


class TestConfigurationPhase:
    """Test Phase 4: Configuration generation"""

    def test_configuration_phase_updates_current_phase(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should update current phase to 4"""
        (tmp_path / ".moai").mkdir(parents=True)
        config = {"projectName": "test", "mode": "personal"}
        executor.execute_configuration_phase(tmp_path, config)
        assert executor.current_phase == 4

    def test_configuration_phase_calls_progress_callback(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should call progress callback with phase 4"""
        callback = Mock()
        (tmp_path / ".moai").mkdir(parents=True)
        config = {"projectName": "test"}
        executor.execute_configuration_phase(tmp_path, config, progress_callback=callback)
        callback.assert_called_once_with("Phase 4: Generating configurations...", 4, 5)

    def test_configuration_phase_creates_config_file(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should create config.json file"""
        config = {
            "projectName": "TestProject",
            "mode": "personal",
            "locale": "ko",
            "language": "python",
        }

        (tmp_path / ".moai").mkdir(parents=True)
        created_files = executor.execute_configuration_phase(tmp_path, config)

        # Config file should be created
        config_path = tmp_path / ".moai" / "config" / "config.json"
        assert config_path.exists()
        assert str(config_path) in created_files

    def test_configuration_phase_writes_correct_config(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should write config data correctly"""
        config = {
            "project": {
                "name": "TestProject",
                "mode": "team",
                "locale": "en",
                "language": "typescript",
            }
        }

        (tmp_path / ".moai").mkdir(parents=True)
        executor.execute_configuration_phase(tmp_path, config)

        # Verify config content
        config_path = tmp_path / ".moai" / "config" / "config.json"
        saved_config = json.loads(config_path.read_text())

        # Check project section
        assert "project" in saved_config
        assert saved_config["project"]["name"] == "TestProject"
        assert saved_config["project"]["mode"] == "team"
        assert saved_config["project"]["locale"] == "en"
        assert saved_config["project"]["language"] == "typescript"


class TestValidationPhase:
    """Test Phase 5: Validation and finalization"""

    def test_validation_phase_updates_current_phase(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should update current phase to 5"""
        # Setup complete installation
        for directory in executor.REQUIRED_DIRECTORIES:
            (tmp_path / directory).mkdir(parents=True, exist_ok=True)
        (tmp_path / "CLAUDE.md").write_text("# Project")
        (tmp_path / ".moai" / "config").mkdir(parents=True, exist_ok=True)
        (tmp_path / ".moai" / "config" / "config.json").write_text("{}")

        # Create Alfred command files (SPEC-INIT-004)
        alfred_dir = tmp_path / ".claude" / "commands" / "alfred"
        alfred_dir.mkdir(parents=True, exist_ok=True)
        for cmd in ["0-project.md", "1-plan.md", "2-run.md", "3-sync.md"]:
            (alfred_dir / cmd).write_text("# Command")

        executor.execute_validation_phase(tmp_path, mode="personal")
        assert executor.current_phase == 5

    def test_validation_phase_calls_progress_callback(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should call progress callback with phase 5"""
        callback = Mock()

        # Setup complete installation
        for directory in executor.REQUIRED_DIRECTORIES:
            (tmp_path / directory).mkdir(parents=True, exist_ok=True)
        (tmp_path / "CLAUDE.md").write_text("# Project")
        (tmp_path / ".moai" / "config").mkdir(parents=True, exist_ok=True)
        (tmp_path / ".moai" / "config" / "config.json").write_text("{}")

        # Create Alfred command files (SPEC-INIT-004)
        alfred_dir = tmp_path / ".claude" / "commands" / "alfred"
        alfred_dir.mkdir(parents=True, exist_ok=True)
        for cmd in ["0-project.md", "1-plan.md", "2-run.md", "3-sync.md"]:
            (alfred_dir / cmd).write_text("# Command")

        executor.execute_validation_phase(tmp_path, mode="personal", progress_callback=callback)
        callback.assert_called_once_with("Phase 5: Validation and finalization...", 5, 5)

    def test_validation_phase_validates_installation(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should validate installation completeness"""
        # Setup complete installation
        for directory in executor.REQUIRED_DIRECTORIES:
            (tmp_path / directory).mkdir(parents=True, exist_ok=True)
        (tmp_path / "CLAUDE.md").write_text("# Project")

        with patch.object(executor.validator, "validate_installation") as mock_validate:
            executor.execute_validation_phase(tmp_path, mode="personal")
            mock_validate.assert_called_once_with(tmp_path)

    @patch("subprocess.run")
    def test_validation_phase_initializes_git_in_team_mode(
        self, mock_run: Mock, executor: PhaseExecutor, tmp_path: Path
    ) -> None:
        """Should initialize Git repository in team mode"""
        # Setup complete installation
        for directory in executor.REQUIRED_DIRECTORIES:
            (tmp_path / directory).mkdir(parents=True, exist_ok=True)
        (tmp_path / "CLAUDE.md").write_text("# Project")
        (tmp_path / ".moai" / "config").mkdir(parents=True, exist_ok=True)
        (tmp_path / ".moai" / "config" / "config.json").write_text("{}")

        # Create Alfred command files (SPEC-INIT-004)
        alfred_dir = tmp_path / ".claude" / "commands" / "alfred"
        alfred_dir.mkdir(parents=True, exist_ok=True)
        for cmd in ["0-project.md", "1-plan.md", "2-run.md", "3-sync.md"]:
            (alfred_dir / cmd).write_text("# Command")

        executor.execute_validation_phase(tmp_path, mode="team")

        # Git init should be called
        mock_run.assert_called_once()
        call_args = mock_run.call_args
        assert call_args[0][0] == ["git", "init"]
        assert call_args[1]["cwd"] == tmp_path

    @patch("subprocess.run")
    def test_validation_phase_skips_git_in_personal_mode(
        self, mock_run: Mock, executor: PhaseExecutor, tmp_path: Path
    ) -> None:
        """Should not initialize Git in personal mode"""
        # Setup complete installation
        for directory in executor.REQUIRED_DIRECTORIES:
            (tmp_path / directory).mkdir(parents=True, exist_ok=True)
        (tmp_path / "CLAUDE.md").write_text("# Project")
        (tmp_path / ".moai" / "config").mkdir(parents=True, exist_ok=True)
        (tmp_path / ".moai" / "config" / "config.json").write_text("{}")

        # Create Alfred command files (SPEC-INIT-004)
        alfred_dir = tmp_path / ".claude" / "commands" / "alfred"
        alfred_dir.mkdir(parents=True, exist_ok=True)
        for cmd in ["0-project.md", "1-plan.md", "2-run.md", "3-sync.md"]:
            (alfred_dir / cmd).write_text("# Command")

        executor.execute_validation_phase(tmp_path, mode="personal")

        # Git init should NOT be called
        mock_run.assert_not_called()


class TestCreateBackup:
    """Test backup creation helper"""

    def test_create_backup_creates_timestamped_directory(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should create backup with timestamp"""
        # Create files to backup
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config.json").write_text("{}")

        executor._create_backup(tmp_path)

        # Backup directory should exist in .moai-backups/
        backup_dir = tmp_path / ".moai-backups"
        assert backup_dir.exists()
        backup_timestamps = list(backup_dir.iterdir())
        assert len(backup_timestamps) > 0
        assert backup_timestamps[0].is_dir()

    def test_create_backup_excludes_protected_paths(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should exclude protected paths from backup"""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "config.json").write_text("{}")

        # Create protected paths
        (moai_dir / "specs").mkdir()
        (moai_dir / "specs" / "test.md").write_text("# SPEC")

        executor._create_backup(tmp_path)

        # Find backup directory in .moai-backups/
        backup_root = tmp_path / ".moai-backups"
        backup_dirs = list(backup_root.iterdir())
        assert len(backup_dirs) > 0

        # Protected paths should not be in backup
        backup_dir = backup_dirs[0]
        assert not (backup_dir / ".moai" / "specs").exists()


class TestCopyDirectorySelective:
    """Test selective directory copying"""

    def test_copy_directory_selective_copies_allowed_files(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should copy allowed files in directory"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir(parents=True)

        # Create regular files (should be copied)
        (src / "file1.txt").write_text("content1")
        (src / "file2.json").write_text('{"key": "value"}')

        # Create subdirectory with files
        subdir = src / "subdir"
        subdir.mkdir()
        (subdir / "nested.txt").write_text("nested content")

        executor._copy_directory_selective(src, dst)

        # Regular files should be copied
        assert (dst / "file1.txt").exists()
        assert (dst / "file2.json").exists()
        assert (dst / "subdir" / "nested.txt").exists()

        # Content should match
        assert "content1" in (dst / "file1.txt").read_text()
        assert "value" in (dst / "file2.json").read_text()
        assert "nested" in (dst / "subdir" / "nested.txt").read_text()


class TestInitializeGit:
    """Test Git initialization"""

    @patch("subprocess.run")
    def test_initialize_git_runs_git_init(self, mock_run: Mock, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Should run git init command"""
        executor._initialize_git(tmp_path)

        mock_run.assert_called_once_with(["git", "init"], cwd=tmp_path, check=True, capture_output=True, timeout=30)

    @patch("subprocess.run")
    def test_initialize_git_handles_errors_gracefully(
        self, mock_run: Mock, executor: PhaseExecutor, tmp_path: Path
    ) -> None:
        """Should handle git init errors without raising"""
        import subprocess

        mock_run.side_effect = subprocess.CalledProcessError(1, ["git", "init"])

        # Should not raise exception
        executor._initialize_git(tmp_path)


class TestReportProgress:
    """Test progress reporting"""

    def test_report_progress_calls_callback(self, executor: PhaseExecutor) -> None:
        """Should call callback with progress info"""
        callback = Mock()
        executor.current_phase = 3
        executor._report_progress("Test message", callback)

        callback.assert_called_once_with("Test message", 3, 5)

    def test_report_progress_handles_none_callback(self, executor: PhaseExecutor) -> None:
        """Should handle None callback without error"""
        executor.current_phase = 1
        # Should not raise exception
        executor._report_progress("Test message", None)
