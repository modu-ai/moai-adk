"""
@TEST:SPEC-COMMAND-REFACTORED-001 Refactored SpecCommand Integration Tests
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:SPEC-REFACTOR-001 → @TEST:SPEC-COMMAND-REFACTORED-001

Tests for refactored SpecCommand integration following TRUST principles:
- T: Test-first orchestration
- R: Readable integration flow
- U: Unified command orchestration
- S: Secure module integration
- T: Trackable command execution
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch, call
from src.moai_adk.commands.spec_command_refactored import SpecCommand
from src.moai_adk.core.exceptions import GitLockedException


class TestSpecCommandRefactored:
    """Test class for refactored SpecCommand following TRUST principles"""

    def setup_method(self):
        """Setup test environment with temporary directory"""
        self.temp_dir = tempfile.mkdtemp()
        self.project_dir = Path(self.temp_dir)
        self.mock_config = Mock()
        self.command = SpecCommand(self.project_dir, self.mock_config)

    def teardown_method(self):
        """Cleanup temporary directory"""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    # Initialization Tests
    def test_should_initialize_with_valid_parameters(self):
        """Test initialization with valid parameters"""
        # Given
        project_dir = self.project_dir
        config = Mock()

        # When
        command = SpecCommand(project_dir, config)

        # Then
        assert command.project_dir == project_dir.resolve()
        assert command.config == config
        assert command.skip_branch is False

    def test_should_reject_invalid_project_dir_type(self):
        """Test rejection of non-Path project directory"""
        # Given
        invalid_dir = "/test/path"  # string instead of Path

        # When & Then
        with pytest.raises(ValueError, match="project_dir must be a Path object"):
            SpecCommand(invalid_dir, Mock())

    def test_should_reject_non_existent_project_dir(self):
        """Test rejection of non-existent project directory"""
        # Given
        non_existent_dir = Path("/non/existent/path")

        # When & Then
        with pytest.raises(ValueError, match="Project directory does not exist"):
            SpecCommand(non_existent_dir, Mock())

    def test_should_initialize_with_skip_branch_option(self):
        """Test initialization with skip_branch option"""
        # Given
        command = SpecCommand(self.project_dir, self.mock_config, skip_branch=True)

        # When & Then
        assert command.skip_branch is True

    # Module Integration Tests
    def test_should_execute_full_workflow_successfully(self):
        """Test successful execution of full SPEC workflow"""
        # Given
        spec_name = "USER-AUTH"
        description = "User authentication system"

        # Setup mocks with dependency injection
        mock_validator = Mock()
        mock_validator.validate_spec_name.return_value = "USER-AUTH"
        mock_validator.validate_description.return_value = "User authentication system"

        mock_file_generator = Mock()
        mock_file_path = self.project_dir / ".moai" / "specs" / "USER-AUTH.md"
        mock_file_generator.create_spec_file.return_value = mock_file_path

        mock_git_handler = Mock()
        mock_git_handler.should_create_branch.return_value = True

        # Create command with injected dependencies
        command = SpecCommand(
            self.project_dir, self.mock_config,
            validator=mock_validator,
            file_generator=mock_file_generator,
            git_handler=mock_git_handler
        )

        # When
        command.execute(spec_name, description)

        # Then
        # Verify validator was called
        mock_validator.validate_spec_name.assert_called_once_with(spec_name)
        mock_validator.validate_description.assert_called_once_with(description)

        # Verify file generator was called
        mock_file_generator.create_spec_file.assert_called_once_with("USER-AUTH", "User authentication system")

        # Verify git handler was called
        mock_git_handler.should_create_branch.assert_called_once()
        mock_git_handler.execute_git_workflow.assert_called_once_with("USER-AUTH")

    def test_should_skip_branch_creation_when_requested(self):
        """Test skipping branch creation when skip_branch=True"""
        # Given
        spec_name = "TEST-SPEC"
        description = "Test specification"

        # Setup mocks
        mock_validator = Mock()
        mock_validator.validate_spec_name.return_value = "TEST-SPEC"
        mock_validator.validate_description.return_value = "Test specification"

        mock_file_generator = Mock()

        mock_git_handler = Mock()

        # Create command with skip_branch=True
        command = SpecCommand(
            self.project_dir, self.mock_config, skip_branch=True,
            validator=mock_validator,
            file_generator=mock_file_generator,
            git_handler=mock_git_handler
        )

        # When
        command.execute(spec_name, description)

        # Then
        # Verify git workflow was not executed
        mock_git_handler.should_create_branch.assert_not_called()
        mock_git_handler.execute_git_workflow.assert_not_called()

    def test_should_skip_branch_creation_when_not_needed(self):
        """Test skipping branch creation when git handler says not needed"""
        # Given
        spec_name = "TEST-SPEC"
        description = "Test specification"

        # Setup mocks
        mock_validator = Mock()
        mock_validator.validate_spec_name.return_value = "TEST-SPEC"
        mock_validator.validate_description.return_value = "Test specification"

        mock_file_generator = Mock()

        mock_git_handler = Mock()
        mock_git_handler.should_create_branch.return_value = False

        # Create command with injected dependencies
        command = SpecCommand(
            self.project_dir, self.mock_config,
            validator=mock_validator,
            file_generator=mock_file_generator,
            git_handler=mock_git_handler
        )

        # When
        command.execute(spec_name, description)

        # Then
        # Verify git workflow was not executed
        mock_git_handler.should_create_branch.assert_called_once()
        mock_git_handler.execute_git_workflow.assert_not_called()

    # Error Handling Tests
    def test_should_propagate_validation_errors(self):
        """Test propagation of validation errors"""
        # Given
        spec_name = "invalid@name"
        description = "Test description"

        mock_validator = Mock()
        mock_validator.validate_spec_name.side_effect = ValueError("Invalid spec name")

        # Create command with injected validator
        command = SpecCommand(
            self.project_dir, self.mock_config,
            validator=mock_validator
        )

        # When & Then
        with pytest.raises(ValueError, match="Invalid spec name"):
            command.execute(spec_name, description)

    def test_should_propagate_file_creation_errors(self):
        """Test propagation of file creation errors"""
        # Given
        spec_name = "TEST-SPEC"
        description = "Test description"

        mock_validator = Mock()
        mock_validator.validate_spec_name.return_value = "TEST-SPEC"
        mock_validator.validate_description.return_value = "Test description"

        mock_file_generator = Mock()
        mock_file_generator.create_spec_file.side_effect = ValueError("File creation failed")

        # Create command with injected dependencies
        command = SpecCommand(
            self.project_dir, self.mock_config,
            validator=mock_validator,
            file_generator=mock_file_generator
        )

        # When & Then
        with pytest.raises(ValueError, match="File creation failed"):
            command.execute(spec_name, description)

    def test_should_handle_git_locked_exception_gracefully(self):
        """Test graceful handling of Git locked exception"""
        # Given
        spec_name = "TEST-SPEC"
        description = "Test description"

        # Setup mocks
        mock_validator = Mock()
        mock_validator.validate_spec_name.return_value = "TEST-SPEC"
        mock_validator.validate_description.return_value = "Test description"

        mock_file_generator = Mock()

        mock_git_handler = Mock()
        mock_git_handler.should_create_branch.return_value = True
        mock_git_handler.execute_git_workflow.side_effect = GitLockedException("Git is locked")

        # Create command with injected dependencies
        command = SpecCommand(
            self.project_dir, self.mock_config,
            validator=mock_validator,
            file_generator=mock_file_generator,
            git_handler=mock_git_handler
        )

        # When (should not raise exception)
        command.execute(spec_name, description)

        # Then
        # Verify file was still created despite Git lock
        mock_file_generator.create_spec_file.assert_called_once()

    # Skip Branch Parameter Tests
    def test_should_override_skip_branch_in_execute(self):
        """Test overriding skip_branch parameter in execute method"""
        # Given
        command = SpecCommand(self.project_dir, self.mock_config, skip_branch=False)

        # When
        with patch.multiple(
            'src.moai_adk.commands.spec_command_refactored',
            SpecValidator=Mock(return_value=Mock(
                validate_spec_name=Mock(return_value="TEST"),
                validate_description=Mock(return_value="Test")
            )),
            SpecFileGenerator=Mock(return_value=Mock()),
            SpecGitHandler=Mock(return_value=Mock())
        ):
            command.execute("TEST", "Test", skip_branch=True)

        # Then
        assert command.skip_branch is True

    # Mode-based Execution Tests
    def test_should_execute_with_personal_mode(self):
        """Test execution with personal mode"""
        # Given
        mode = "personal"
        spec_name = "PERSONAL-SPEC"
        description = "Personal mode test"

        # When
        with patch.multiple(
            'src.moai_adk.commands.spec_command_refactored',
            SpecValidator=Mock(return_value=Mock(
                validate_spec_name=Mock(return_value=spec_name),
                validate_description=Mock(return_value=description)
            )),
            SpecFileGenerator=Mock(return_value=Mock()),
            SpecGitHandler=Mock(return_value=Mock(should_create_branch=Mock(return_value=False)))
        ):
            self.command.execute_with_mode(mode, spec_name, description)

        # Then (should complete without error)

    def test_should_execute_with_team_mode(self):
        """Test execution with team mode"""
        # Given
        mode = "team"
        spec_name = "TEAM-SPEC"
        description = "Team mode test"

        # When
        with patch.multiple(
            'src.moai_adk.commands.spec_command_refactored',
            SpecValidator=Mock(return_value=Mock(
                validate_spec_name=Mock(return_value=spec_name),
                validate_description=Mock(return_value=description)
            )),
            SpecFileGenerator=Mock(return_value=Mock()),
            SpecGitHandler=Mock(return_value=Mock(should_create_branch=Mock(return_value=True)))
        ):
            self.command.execute_with_mode(mode, spec_name, description)

        # Then (should complete without error)

    def test_should_reject_invalid_mode(self):
        """Test rejection of invalid execution mode"""
        # Given
        invalid_mode = "invalid"

        # When & Then
        with pytest.raises(ValueError, match="지원하지 않는 모드입니다"):
            self.command.execute_with_mode(invalid_mode)

    # Status Information Tests
    def test_should_return_command_status(self):
        """Test command status information return"""
        # Given
        mock_git_handler = Mock()
        mock_git_handler.get_handler_status.return_value = {
            "project_dir": str(self.project_dir.resolve()),
            "mode": "personal",
            "git_strategy": "PersonalGitStrategy",
            "repository_status": {"branch": "main", "clean": True}
        }

        # Create command with injected git handler
        command = SpecCommand(
            self.project_dir, self.mock_config,
            git_handler=mock_git_handler
        )

        # When
        status = command.get_command_status()

        # Then
        expected_status = {
            "project_dir": str(self.project_dir.resolve()),
            "skip_branch": False,
            "specs_dir_exists": False,  # temp dir doesn't have specs dir
            "git_handler_status": {
                "project_dir": str(self.project_dir.resolve()),
                "mode": "personal",
                "git_strategy": "PersonalGitStrategy",
                "repository_status": {"branch": "main", "clean": True}
            }
        }
        assert status == expected_status