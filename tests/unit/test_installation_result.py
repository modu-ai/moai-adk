"""
Unit tests for the installation result module.

Tests the InstallationResult dataclass and its methods
to ensure proper data handling and utility functions.
"""

import pytest
from pathlib import Path

from moai_adk.install.installation_result import InstallationResult
from moai_adk.config import Config, RuntimeConfig

class TestInstallationResult:
    """Test cases for InstallationResult dataclass."""

    @pytest.fixture
    def sample_config(self):
        """Create a sample Config instance for testing."""
        return Config(
            name="test-project",
            
            template="standard",
            runtime=RuntimeConfig("python")
        )

    @pytest.fixture
    def success_result(self, sample_config):
        """Create a successful installation result for testing."""
        return InstallationResult(
            success=True,
            project_path="/test/project",
            files_created=["file1.py", "file2.js", "file3.md"],
            next_steps=["Run tests", "Deploy"],
            config=sample_config,
            git_initialized=True,
            backup_created="/backup/location"
        )

    @pytest.fixture
    def failure_result(self, sample_config):
        """Create a failed installation result for testing."""
        return InstallationResult(
            success=False,
            project_path="/test/project",
            files_created=["partial.py"],
            next_steps=[],
            config=sample_config,
            errors=["Permission denied", "Invalid path"]
        )

    def test_init_basic_fields(self, sample_config):
        """Test basic initialization of InstallationResult."""
        result = InstallationResult(
            success=True,
            project_path="/test/path",
            files_created=["file1.py"],
            next_steps=["step1"],
            config=sample_config
        )

        assert result.success is True
        assert result.project_path == "/test/path"
        assert result.files_created == ["file1.py"]
        assert result.next_steps == ["step1"]
        assert result.config == sample_config

    def test_post_init_initializes_optional_fields(self, sample_config):
        """Test that post_init initializes None optional fields to empty lists."""
        result = InstallationResult(
            success=True,
            project_path="/test/path",
            files_created=[],
            next_steps=[],
            config=sample_config
        )

        # Should be initialized to empty lists by post_init
        assert result.errors == []
        assert result.warnings == []

    def test_post_init_preserves_existing_values(self, sample_config):
        """Test that post_init doesn't overwrite existing values."""
        existing_errors = ["existing error"]
        existing_warnings = ["existing warning"]

        result = InstallationResult(
            success=False,
            project_path="/test/path",
            files_created=[],
            next_steps=[],
            config=sample_config,
            errors=existing_errors,
            warnings=existing_warnings
        )

        assert result.errors == existing_errors
        assert result.warnings == existing_warnings

    def test_has_errors_returns_true_with_errors(self, failure_result):
        """Test has_errors returns True when errors exist."""
        assert failure_result.has_errors() is True

    def test_has_errors_returns_false_without_errors(self, success_result):
        """Test has_errors returns False when no errors exist."""
        assert success_result.has_errors() is False

    def test_has_errors_returns_false_with_empty_list(self, sample_config):
        """Test has_errors returns False with empty error list."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=[],
            next_steps=[],
            config=sample_config,
            errors=[]
        )
        assert result.has_errors() is False

    def test_has_warnings_returns_true_with_warnings(self, sample_config):
        """Test has_warnings returns True when warnings exist."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=[],
            next_steps=[],
            config=sample_config,
            warnings=["warning message"]
        )
        assert result.has_warnings() is True

    def test_has_warnings_returns_false_without_warnings(self, success_result):
        """Test has_warnings returns False when no warnings exist."""
        assert success_result.has_warnings() is False

    def test_add_error_to_none_list(self, sample_config):
        """Test adding error when errors list is None."""
        result = InstallationResult(
            success=False,
            project_path="/test",
            files_created=[],
            next_steps=[],
            config=sample_config,
            errors=None
        )

        result.add_error("New error")

        assert result.errors == ["New error"]
        assert result.has_errors() is True

    def test_add_error_to_existing_list(self, failure_result):
        """Test adding error to existing errors list."""
        original_count = len(failure_result.errors)

        failure_result.add_error("Additional error")

        assert len(failure_result.errors) == original_count + 1
        assert "Additional error" in failure_result.errors

    def test_add_warning_to_none_list(self, sample_config):
        """Test adding warning when warnings list is None."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=[],
            next_steps=[],
            config=sample_config,
            warnings=None
        )

        result.add_warning("New warning")

        assert result.warnings == ["New warning"]
        assert result.has_warnings() is True

    def test_add_warning_to_existing_list(self, sample_config):
        """Test adding warning to existing warnings list."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=[],
            next_steps=[],
            config=sample_config,
            warnings=["existing warning"]
        )

        result.add_warning("Additional warning")

        assert len(result.warnings) == 2
        assert "Additional warning" in result.warnings

    def test_get_summary_success_basic(self, success_result):
        """Test get_summary for successful installation."""
        summary = success_result.get_summary()

        assert "✅ Installation successful" in summary
        assert success_result.project_path in summary
        assert "Created 3 files" in summary
        assert "🚀 Git repository initialized" in summary
        assert "💾 Backup created at:" in summary

    def test_get_summary_success_with_warnings(self, sample_config):
        """Test get_summary for successful installation with warnings."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=["file1.py"],
            next_steps=[],
            config=sample_config,
            warnings=["warning1", "warning2"]
        )

        summary = result.get_summary()

        assert "✅ Installation successful" in summary
        assert "⚠️ 2 warning(s) occurred" in summary

    def test_get_summary_failure(self, failure_result):
        """Test get_summary for failed installation."""
        summary = failure_result.get_summary()

        assert "❌ Installation failed" in summary
        assert failure_result.project_path in summary
        assert "🔥 2 error(s) occurred" in summary

    def test_get_file_count_by_type_mixed_extensions(self, success_result):
        """Test file count grouping by extensions."""
        counts = success_result.get_file_count_by_type()

        expected = {'.py': 1, '.js': 1, '.md': 1}
        assert counts == expected

    def test_get_file_count_by_type_no_extension(self, sample_config):
        """Test file count with files that have no extension."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=["README", "LICENSE", "config.py"],
            next_steps=[],
            config=sample_config
        )

        counts = result.get_file_count_by_type()

        expected = {'no_extension': 2, '.py': 1}
        assert counts == expected

    def test_get_file_count_by_type_empty_list(self, sample_config):
        """Test file count with empty files list."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=[],
            next_steps=[],
            config=sample_config
        )

        counts = result.get_file_count_by_type()
        assert counts == {}

    def test_to_dict_complete_data(self, success_result):
        """Test conversion to dictionary with all data."""
        result_dict = success_result.to_dict()

        expected_keys = {
            'success', 'project_path', 'files_created', 'next_steps',
            'errors', 'warnings', 'git_initialized', 'backup_created',
            'file_count', 'file_types'
        }

        assert set(result_dict.keys()) == expected_keys
        assert result_dict['success'] is True
        assert result_dict['project_path'] == "/test/project"
        assert result_dict['file_count'] == 3
        assert result_dict['git_initialized'] is True
        assert isinstance(result_dict['file_types'], dict)

    def test_to_dict_excludes_config_object(self, success_result):
        """Test that to_dict doesn't include the config object directly."""
        result_dict = success_result.to_dict()

        # Config object should not be serialized directly
        assert 'config' not in result_dict

    def test_create_success_with_defaults(self, sample_config):
        """Test create_success class method with default parameters."""
        result = InstallationResult.create_success(
            project_path="/test/project",
            config=sample_config
        )

        assert result.success is True
        assert result.project_path == "/test/project"
        assert result.config == sample_config
        assert result.files_created == []
        assert result.next_steps == []
        assert result.git_initialized is False
        assert result.backup_created is None
        assert result.errors == []
        assert result.warnings == []

    def test_create_success_with_all_parameters(self, sample_config):
        """Test create_success class method with all parameters."""
        files = ["file1.py", "file2.js"]
        steps = ["run tests", "deploy"]

        result = InstallationResult.create_success(
            project_path="/test/project",
            config=sample_config,
            files_created=files,
            next_steps=steps,
            git_initialized=True,
            backup_created="/backup/path"
        )

        assert result.success is True
        assert result.files_created == files
        assert result.next_steps == steps
        assert result.git_initialized is True
        assert result.backup_created == "/backup/path"

    def test_create_failure_with_defaults(self, sample_config):
        """Test create_failure class method with default parameters."""
        error_message = "Installation failed"

        result = InstallationResult.create_failure(
            project_path="/test/project",
            config=sample_config,
            error=error_message
        )

        assert result.success is False
        assert result.project_path == "/test/project"
        assert result.config == sample_config
        assert result.files_created == []
        assert result.next_steps == []
        assert result.errors == [error_message]
        assert result.git_initialized is False
        assert result.backup_created is None

    def test_create_failure_with_files_created(self, sample_config):
        """Test create_failure class method with files_created parameter."""
        error_message = "Partial failure"
        files = ["partial1.py", "partial2.js"]

        result = InstallationResult.create_failure(
            project_path="/test/project",
            config=sample_config,
            error=error_message,
            files_created=files
        )

        assert result.success is False
        assert result.files_created == files
        assert result.errors == [error_message]

    def test_result_can_be_modified_after_creation(self, success_result):
        """Test that result can be modified after creation."""
        # Add error to successful result
        success_result.add_error("Post-creation error")

        assert success_result.has_errors() is True
        assert "Post-creation error" in success_result.errors

        # Add warning
        success_result.add_warning("Post-creation warning")

        assert success_result.has_warnings() is True
        assert "Post-creation warning" in success_result.warnings

    def test_path_handling_in_file_count(self, sample_config):
        """Test that file counting handles various path formats."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=[
                "/absolute/path/file.py",
                "relative/path/file.js",
                "./current/dir/file.css",
                "../parent/dir/file.html"
            ],
            next_steps=[],
            config=sample_config
        )

        counts = result.get_file_count_by_type()

        expected = {'.py': 1, '.js': 1, '.css': 1, '.html': 1}
        assert counts == expected

    def test_case_sensitivity_in_extensions(self, sample_config):
        """Test that file extensions are normalized to lowercase."""
        result = InstallationResult(
            success=True,
            project_path="/test",
            files_created=["file.PY", "file.JS", "file.py", "file.js"],
            next_steps=[],
            config=sample_config
        )

        counts = result.get_file_count_by_type()

        # All should be normalized to lowercase
        assert counts['.py'] == 2
        assert counts['.js'] == 2
        assert '.PY' not in counts
        assert '.JS' not in counts