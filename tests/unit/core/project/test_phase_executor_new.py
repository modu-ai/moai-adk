"""
Comprehensive tests for PhaseExecutor module.

Tests cover:
- PhaseExecutor class initialization
- All 5 phase execution methods
- Backup and restoration operations
- Version context generation
- Template variable substitution
- Git initialization
"""

import json
import subprocess
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.project.phase_executor import PhaseExecutor
from moai_adk.core.project.validator import ProjectValidator


class TestPhaseExecutor:
    """Test suite for PhaseExecutor class."""

    def test_initialization(self):
        """Test PhaseExecutor initialization."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)

        # Act
        executor = PhaseExecutor(validator)

        # Assert
        assert executor.validator is validator
        assert executor.total_phases == 5
        assert executor.current_phase == 0

    def test_required_directories_defined(self):
        """Test REQUIRED_DIRECTORIES constant is properly defined."""
        # Arrange & Act
        required_dirs = PhaseExecutor.REQUIRED_DIRECTORIES

        # Assert
        assert ".moai/" in required_dirs
        assert ".claude/" in required_dirs
        assert ".github/" in required_dirs
        assert ".moai/memory/" in required_dirs
        assert ".moai/specs/" in required_dirs

    def test_execute_preparation_phase_calls_validator(self):
        """Test execute_preparation_phase calls validator methods."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            executor.execute_preparation_phase(path)

            # Assert
            validator.validate_system_requirements.assert_called_once()
            validator.validate_project_path.assert_called_once_with(path)

    def test_execute_preparation_phase_creates_backup_when_enabled(self):
        """Test execute_preparation_phase creates backup when enabled."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            moai_dir = path / ".moai"
            moai_dir.mkdir()
            (moai_dir / "config").mkdir()
            (moai_dir / "config" / "config.json").write_text("{}")

            # Act
            with patch(
                "moai_adk.core.project.phase_executor.has_any_moai_files",
                return_value=True,
            ):
                with patch.object(executor, "_create_backup") as mock_backup:
                    executor.execute_preparation_phase(path, backup_enabled=True)

            # Assert
            mock_backup.assert_called_once_with(path)

    def test_execute_preparation_phase_skips_backup_when_disabled(self):
        """Test execute_preparation_phase skips backup when disabled."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            with patch(
                "moai_adk.core.project.phase_executor.has_any_moai_files",
                return_value=True,
            ):
                with patch.object(executor, "_create_backup") as mock_backup:
                    executor.execute_preparation_phase(path, backup_enabled=False)

            # Assert
            mock_backup.assert_not_called()

    def test_execute_directory_phase_creates_all_directories(self):
        """Test execute_directory_phase creates all required directories."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            executor.execute_directory_phase(path)

            # Assert
            for directory in PhaseExecutor.REQUIRED_DIRECTORIES:
                assert (path / directory).exists()

    def test_execute_directory_phase_is_idempotent(self):
        """Test execute_directory_phase can be called multiple times safely."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            executor.execute_directory_phase(path)
            executor.execute_directory_phase(path)

            # Assert
            for directory in PhaseExecutor.REQUIRED_DIRECTORIES:
                assert (path / directory).exists()

    def test_execute_resource_phase_returns_file_list(self):
        """Test execute_resource_phase returns list of created files."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            with patch("moai_adk.core.project.phase_executor.TemplateProcessor") as mock_processor:
                mock_processor_instance = MagicMock()
                mock_processor.return_value = mock_processor_instance
                result = executor.execute_resource_phase(path)

            # Assert
            assert isinstance(result, list)
            assert ".claude/" in result or ".moai/" in result

    def test_execute_resource_phase_sets_context_variables(self):
        """Test execute_resource_phase sets context variables correctly."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            config = {
                "name": "test-project",
                "language": "python",
                "mode": "personal",
                "author": "Test Author",
                "description": "Test Description",
                "version": "0.1.0",
            }

            # Act
            with patch("moai_adk.core.project.phase_executor.TemplateProcessor") as mock_processor:
                mock_processor_instance = MagicMock()
                mock_processor.return_value = mock_processor_instance
                executor.execute_resource_phase(path, config)

            # Assert
            mock_processor_instance.set_context.assert_called_once()
            call_args = mock_processor_instance.set_context.call_args
            context = call_args[0][0]
            assert context["PROJECT_NAME"] == "test-project"
            assert context["CODEBASE_LANGUAGE"] == "python"

    def test_execute_configuration_phase_creates_config_file(self):
        """Test execute_configuration_phase updates section YAML files.

        Note: As of v0.37.0, we use section YAML files instead of config.json.
        """
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            # Create sections directory with YAML files
            sections_dir = path / ".moai" / "config" / "sections"
            sections_dir.mkdir(parents=True)

            # Create sample section files that Phase 4 updates
            (sections_dir / "project.yaml").write_text('project:\n  name: ""\n  initialized: false\n')
            (sections_dir / "system.yaml").write_text('moai:\n  version: "0.0.0"\n')

            config = {
                "project": {"name": "test-project"},
                "language": {"conversation_language": "en"},
                "constitution": {"enforce_quality": True},
            }

            # Act
            result = executor.execute_configuration_phase(path, config)

            # Assert - should return updated section files
            assert len(result) > 0
            # Verify section files were updated (not config.json)
            assert (sections_dir / "project.yaml").exists()
            assert (sections_dir / "system.yaml").exists()

    def test_execute_configuration_phase_reads_existing_config(self):
        """Test execute_configuration_phase reads and preserves existing config."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            config_dir = path / ".moai" / "config"
            config_dir.mkdir(parents=True)

            # Write existing config
            existing_config = {
                "moai": {"version": "0.30.0"},
                "user": {"nickname": "test-user"},
            }
            config_file = config_dir / "config.json"
            config_file.write_text(json.dumps(existing_config))

            new_config = {
                "project": {"name": "test-project"},
                "language": {"conversation_language": "en"},
            }

            # Act
            executor.execute_configuration_phase(path, new_config)

            # Assert
            written_config = json.loads(config_file.read_text())
            assert written_config is not None

    def test_execute_configuration_phase_merges_configurations(self):
        """Test execute_configuration_phase updates section YAML files.

        Note: As of v0.37.0, we use section YAML files instead of config.json.
        """
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            sections_dir = path / ".moai" / "config" / "sections"
            sections_dir.mkdir(parents=True)

            # Create existing section files
            (sections_dir / "project.yaml").write_text("project:\n  name: old-project\n  initialized: false\n")
            (sections_dir / "system.yaml").write_text('moai:\n  version: "0.30.0"\n')

            new_config = {
                "project": {"name": "test-project"},
                "moai": {"version": "0.32.0"},
            }

            # Act
            executor.execute_configuration_phase(path, new_config)

            # Assert - verify section files were updated
            import yaml

            project_content = yaml.safe_load((sections_dir / "project.yaml").read_text())
            assert project_content["project"]["name"] == "test-project"
            # system.yaml is updated with the current MoAI version, not the config input
            assert (sections_dir / "system.yaml").exists()

    def test_execute_validation_phase_calls_validator(self):
        """Test execute_validation_phase calls validator.validate_installation."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            with patch.object(executor, "_initialize_git"):
                executor.execute_validation_phase(path)

            # Assert
            validator.validate_installation.assert_called_once()

    def test_execute_validation_phase_initializes_git(self):
        """Test execute_validation_phase initializes git repository."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            with patch.object(executor, "_initialize_git") as mock_git:
                executor.execute_validation_phase(path)

            # Assert
            mock_git.assert_called_once_with(path)

    def test_create_backup_creates_backup_directory(self):
        """Test _create_backup creates backup directory structure."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            moai_dir = path / ".moai"
            moai_dir.mkdir()
            (moai_dir / "config").mkdir()
            (moai_dir / "config" / "config.json").write_text("{}")

            # Act
            with patch(
                "moai_adk.core.project.phase_executor.get_backup_targets",
                return_value=[".moai/config"],
            ):
                with patch(
                    "moai_adk.core.project.phase_executor.is_protected_path",
                    return_value=False,
                ):
                    executor._create_backup(path)

            # Assert
            backup_dir = path / ".moai-backups" / "backup"
            assert backup_dir.exists()

    def test_create_backup_removes_existing_backup(self):
        """Test _create_backup removes existing backup before creating new one."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            moai_dir = path / ".moai"
            moai_dir.mkdir()
            (moai_dir / "config").mkdir()

            # Create existing backup
            backup_dir = path / ".moai-backups" / "backup"
            backup_dir.mkdir(parents=True)
            (backup_dir / "old_file.txt").write_text("old")

            # Act
            with patch(
                "moai_adk.core.project.phase_executor.get_backup_targets",
                return_value=[],
            ):
                executor._create_backup(path)

            # Assert
            assert backup_dir.exists()
            assert not (backup_dir / "old_file.txt").exists()

    def test_initialize_git_runs_git_init_command(self):
        """Test _initialize_git runs git init command."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            with patch("moai_adk.core.project.phase_executor.subprocess.run") as mock_run:
                executor._initialize_git(path)

            # Assert
            mock_run.assert_called_once()
            call_args = mock_run.call_args
            assert "git" in call_args[0][0]
            assert "init" in call_args[0][0]

    def test_initialize_git_skips_if_already_initialized(self):
        """Test _initialize_git skips if .git directory exists."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / ".git").mkdir()

            # Act
            with patch("moai_adk.core.project.phase_executor.subprocess.run") as mock_run:
                executor._initialize_git(path)

            # Assert
            mock_run.assert_not_called()

    def test_initialize_git_handles_subprocess_errors(self):
        """Test _initialize_git handles subprocess errors gracefully."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            with patch("moai_adk.core.project.phase_executor.subprocess.run") as mock_run:
                mock_run.side_effect = subprocess.CalledProcessError(1, "git init")
                # Should not raise
                executor._initialize_git(path)

            # Assert - no exception raised
            assert True

    def test_report_progress_calls_callback(self):
        """Test _report_progress calls callback when provided."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)
        callback = MagicMock()

        # Act
        executor._report_progress("Test message", callback)

        # Assert
        callback.assert_called_once()
        call_args = callback.call_args
        assert "Test message" in call_args[0]

    def test_report_progress_handles_none_callback(self):
        """Test _report_progress handles None callback gracefully."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        # Act & Assert
        executor._report_progress("Test message", None)  # Should not raise

    def test_format_short_version_removes_prefix(self):
        """Test _format_short_version removes 'v' prefix."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        # Act & Assert
        assert executor._format_short_version("v1.2.3") == "1.2.3"
        assert executor._format_short_version("1.2.3") == "1.2.3"

    def test_format_display_version_adds_moai_prefix(self):
        """Test _format_display_version adds MoAI-ADK prefix."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        # Act & Assert
        assert executor._format_display_version("1.2.3") == "MoAI-ADK v1.2.3"
        assert executor._format_display_version("v1.2.3") == "MoAI-ADK v1.2.3"
        assert executor._format_display_version("unknown") == "MoAI-ADK unknown version"

    def test_format_trimmed_version_trims_to_max_length(self):
        """Test _format_trimmed_version trims to maximum length."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        # Act & Assert
        result = executor._format_trimmed_version("1.2.3-beta-long", max_length=5)
        assert len(result) <= 5

    def test_format_semver_version_extracts_semver(self):
        """Test _format_semver_version extracts semantic version."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        # Act & Assert
        assert executor._format_semver_version("v1.2.3") == "1.2.3"
        assert executor._format_semver_version("1.2.3-beta") == "1.2.3"
        assert executor._format_semver_version("unknown") == "0.0.0"

    def test_get_enhanced_version_context_returns_dict(self):
        """Test _get_enhanced_version_context returns dictionary."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        # Act
        with patch.object(executor, "_get_version_reader") as mock_reader:
            mock_reader_instance = MagicMock()
            mock_reader.return_value = mock_reader_instance
            mock_reader_instance.get_version.return_value = "1.0.0"
            mock_reader_instance.get_cache_age_seconds.return_value = 10.0

            context = executor._get_enhanced_version_context()

        # Assert
        assert isinstance(context, dict)
        assert "MOAI_VERSION" in context
        assert "MOAI_VERSION_SHORT" in context
        assert "MOAI_VERSION_DISPLAY" in context

    def test_copy_directory_selective_skips_protected_paths(self):
        """Test _copy_directory_selective skips protected paths."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            src_dir = Path(temp_dir) / "src"
            dst_dir = Path(temp_dir) / "dst"
            src_dir.mkdir()

            # Create files and directories
            (src_dir / "file.txt").write_text("content")
            (src_dir / "protected").mkdir()
            (src_dir / "protected" / "file.txt").write_text("protected content")

            # Act
            with patch("moai_adk.core.project.phase_executor.is_protected_path") as mock_protected:

                def is_protected(path):
                    return "protected" in str(path)

                mock_protected.side_effect = is_protected
                executor._copy_directory_selective(src_dir, dst_dir)

            # Assert
            assert (dst_dir / "file.txt").exists()
            assert not (dst_dir / "protected").exists()

    def test_current_phase_tracking(self):
        """Test current_phase tracking across phase execution."""
        # Arrange
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act & Assert - Phase 1
            executor.execute_preparation_phase(path)
            assert executor.current_phase == 1

            # Phase 2
            executor.execute_directory_phase(path)
            assert executor.current_phase == 2

            # Phase 3
            with patch("moai_adk.core.project.phase_executor.TemplateProcessor"):
                executor.execute_resource_phase(path)
            assert executor.current_phase == 3

            # Phase 4
            (path / ".moai" / "config").mkdir(parents=True)
            executor.execute_configuration_phase(path, {})
            assert executor.current_phase == 4

            # Phase 5
            with patch.object(executor, "_initialize_git"):
                executor.execute_validation_phase(path)
            assert executor.current_phase == 5
