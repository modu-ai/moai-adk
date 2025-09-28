"""
@TEST:SPEC-GIT-HANDLER-001 SpecGitHandler Unit Tests
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:GIT-HANDLER-001 → @TEST:SPEC-GIT-HANDLER-001

Tests for Git workflow handling module following TRUST principles:
- T: Test-first development
- R: Readable test structure
- U: Unified Git workflow responsibility
- S: Secure Git operations
- T: Trackable Git workflow execution
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch, call
from src.moai_adk.commands.spec_git_handler import SpecGitHandler
from src.moai_adk.core.exceptions import GitLockedException


class TestSpecGitHandler:
    """Test class for SpecGitHandler following TRUST principles"""

    def setup_method(self):
        """Setup test environment with temporary directory"""
        self.temp_dir = tempfile.mkdtemp()
        self.project_dir = Path(self.temp_dir)
        self.mock_config = Mock()
        self.handler = SpecGitHandler(self.project_dir, self.mock_config)

    def teardown_method(self):
        """Cleanup temporary directory"""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    # Initialization Tests
    def test_should_initialize_with_valid_parameters(self):
        """Test initialization with valid parameters"""
        # Given
        project_dir = self.project_dir  # Use existing temp dir
        config = Mock()

        # When
        handler = SpecGitHandler(project_dir, config)

        # Then
        assert handler.project_dir == project_dir.resolve()
        assert handler.config == config
        assert handler._git_strategy is None

    def test_should_reject_invalid_project_dir_type(self):
        """Test rejection of non-Path project directory"""
        # Given
        invalid_dir = "/test/path"  # string instead of Path

        # When & Then
        with pytest.raises(ValueError, match="project_dir must be a Path object"):
            SpecGitHandler(invalid_dir, Mock())

    def test_should_reject_non_existent_project_dir(self):
        """Test rejection of non-existent project directory"""
        # Given
        non_existent_dir = Path("/non/existent/path")

        # When & Then
        with pytest.raises(ValueError, match="Project directory does not exist"):
            SpecGitHandler(non_existent_dir, Mock())

    # Mode Detection Tests
    def test_should_detect_personal_mode_by_default(self):
        """Test default mode detection as personal"""
        # Given
        handler = SpecGitHandler(self.project_dir, None)

        # When
        mode = handler._get_current_mode()

        # Then
        assert mode == "personal"

    def test_should_detect_team_mode_from_config_manager(self):
        """Test team mode detection from ConfigManager"""
        # Given
        mock_config = Mock()
        mock_config.get_mode.return_value = "team"
        handler = SpecGitHandler(self.project_dir, mock_config)

        # When
        mode = handler._get_current_mode()

        # Then
        assert mode == "team"

    def test_should_detect_mode_from_dict_config(self):
        """Test mode detection from dictionary config"""
        # Given
        dict_config = {"mode": "team"}
        handler = SpecGitHandler(self.project_dir, dict_config)

        # When
        mode = handler._get_current_mode()

        # Then
        assert mode == "team"

    def test_should_handle_config_attribute_error(self):
        """Test handling of config AttributeError"""
        # Given
        mock_config = Mock()
        mock_config.get_mode.side_effect = AttributeError("No such method")
        handler = SpecGitHandler(self.project_dir, mock_config)

        # When
        mode = handler._get_current_mode()

        # Then
        assert mode == "personal"

    # Git Strategy Tests
    @patch('src.moai_adk.commands.spec_git_handler.PersonalGitStrategy')
    def test_should_create_personal_git_strategy_for_personal_mode(self, mock_personal_strategy):
        """Test personal Git strategy creation for personal mode"""
        # Given
        mock_config = {"mode": "personal"}
        handler = SpecGitHandler(self.project_dir, mock_config)
        mock_strategy_instance = Mock()
        mock_personal_strategy.return_value = mock_strategy_instance

        # When
        strategy = handler._get_git_strategy()

        # Then
        mock_personal_strategy.assert_called_once_with(self.project_dir.resolve(), mock_config)
        assert strategy == mock_strategy_instance

    @patch('src.moai_adk.commands.spec_git_handler.TeamGitStrategy')
    def test_should_create_team_git_strategy_for_team_mode(self, mock_team_strategy):
        """Test team Git strategy creation for team mode"""
        # Given
        mock_config = {"mode": "team"}
        handler = SpecGitHandler(self.project_dir, mock_config)
        mock_strategy_instance = Mock()
        mock_team_strategy.return_value = mock_strategy_instance

        # When
        strategy = handler._get_git_strategy()

        # Then
        mock_team_strategy.assert_called_once_with(self.project_dir.resolve(), mock_config)
        assert strategy == mock_strategy_instance

    def test_should_cache_git_strategy_instance(self):
        """Test Git strategy instance caching"""
        # Given
        handler = SpecGitHandler(self.project_dir, {"mode": "personal"})

        # When
        strategy1 = handler._get_git_strategy()
        strategy2 = handler._get_git_strategy()

        # Then
        assert strategy1 is strategy2

    # Branch Creation Decision Tests
    def test_should_create_branch_in_team_mode(self):
        """Test branch creation decision in team mode"""
        # Given
        mock_config = {"mode": "team"}
        handler = SpecGitHandler(self.project_dir, mock_config)

        # When
        should_create = handler.should_create_branch()

        # Then
        assert should_create is True

    def test_should_not_create_branch_in_personal_mode(self):
        """Test no branch creation in personal mode"""
        # Given
        mock_config = {"mode": "personal"}
        handler = SpecGitHandler(self.project_dir, mock_config)

        # When
        should_create = handler.should_create_branch()

        # Then
        assert should_create is False

    def test_should_not_create_branch_with_no_config(self):
        """Test no branch creation with no config (defaults to personal)"""
        # Given
        handler = SpecGitHandler(self.project_dir, None)

        # When
        should_create = handler.should_create_branch()

        # Then
        assert should_create is False

    # Git Workflow Execution Tests
    def test_should_execute_git_workflow_successfully(self):
        """Test successful Git workflow execution"""
        # Given
        spec_name = "TEST-SPEC"
        mock_strategy = Mock()
        mock_context_manager = Mock()
        mock_context_manager.__enter__ = Mock(return_value=mock_context_manager)
        mock_context_manager.__exit__ = Mock(return_value=None)
        mock_strategy.work_context.return_value = mock_context_manager

        with patch.object(self.handler, '_get_git_strategy', return_value=mock_strategy):
            # When
            self.handler.execute_git_workflow(spec_name)

            # Then
            mock_strategy.work_context.assert_called_once_with("spec-test-spec")
            mock_context_manager.__enter__.assert_called_once()
            mock_context_manager.__exit__.assert_called_once()

    def test_should_handle_git_locked_exception_gracefully(self):
        """Test graceful handling of Git locked exception"""
        # Given
        spec_name = "TEST-SPEC"
        mock_strategy = Mock()
        mock_strategy.work_context.side_effect = GitLockedException("Git is locked")

        with patch.object(self.handler, '_get_git_strategy', return_value=mock_strategy):
            # When & Then (should re-raise GitLockedException)
            with pytest.raises(GitLockedException):
                self.handler.execute_git_workflow(spec_name)

    def test_should_propagate_other_git_exceptions(self):
        """Test propagation of non-lock Git exceptions"""
        # Given
        spec_name = "TEST-SPEC"
        mock_strategy = Mock()
        mock_strategy.work_context.side_effect = RuntimeError("Other Git error")

        with patch.object(self.handler, '_get_git_strategy', return_value=mock_strategy):
            # When & Then
            with pytest.raises(RuntimeError, match="Other Git error"):
                self.handler.execute_git_workflow(spec_name)

    def test_should_normalize_spec_name_for_branch(self):
        """Test spec name normalization for branch creation"""
        # Given
        spec_name = "USER-AUTH-V2"
        mock_strategy = Mock()
        mock_context_manager = Mock()
        mock_context_manager.__enter__ = Mock(return_value=mock_context_manager)
        mock_context_manager.__exit__ = Mock(return_value=None)
        mock_strategy.work_context.return_value = mock_context_manager

        with patch.object(self.handler, '_get_git_strategy', return_value=mock_strategy):
            # When
            self.handler.execute_git_workflow(spec_name)

            # Then
            mock_strategy.work_context.assert_called_once_with("spec-user-auth-v2")

    # Status and Information Tests
    def test_should_return_handler_status(self):
        """Test handler status information return"""
        # Given
        mock_config = {"mode": "team"}
        handler = SpecGitHandler(self.project_dir, mock_config)
        mock_strategy = Mock()
        mock_strategy.__class__.__name__ = "TeamGitStrategy"
        mock_strategy.get_repository_status.return_value = {"branch": "main", "clean": True}

        with patch.object(handler, '_get_git_strategy', return_value=mock_strategy):
            # When
            status = handler.get_handler_status()

            # Then
            expected_status = {
                "project_dir": str(self.project_dir.resolve()),
                "mode": "team",
                "git_strategy": "TeamGitStrategy",
                "repository_status": {"branch": "main", "clean": True}
            }
            assert status == expected_status

    # Edge Cases
    def test_should_handle_empty_spec_name(self):
        """Test handling of empty spec name"""
        # Given
        spec_name = ""
        mock_strategy = Mock()
        mock_context_manager = Mock()
        mock_context_manager.__enter__ = Mock(return_value=mock_context_manager)
        mock_context_manager.__exit__ = Mock(return_value=None)
        mock_strategy.work_context.return_value = mock_context_manager

        with patch.object(self.handler, '_get_git_strategy', return_value=mock_strategy):
            # When
            self.handler.execute_git_workflow(spec_name)

            # Then
            mock_strategy.work_context.assert_called_once_with("spec-")

    def test_should_handle_special_characters_in_spec_name(self):
        """Test handling of special characters in spec name"""
        # Given
        spec_name = "USER@AUTH#SPEC"
        mock_strategy = Mock()
        mock_context_manager = Mock()
        mock_context_manager.__enter__ = Mock(return_value=mock_context_manager)
        mock_context_manager.__exit__ = Mock(return_value=None)
        mock_strategy.work_context.return_value = mock_context_manager

        with patch.object(self.handler, '_get_git_strategy', return_value=mock_strategy):
            # When
            self.handler.execute_git_workflow(spec_name)

            # Then
            mock_strategy.work_context.assert_called_once_with("spec-user@auth#spec")

    def test_should_handle_none_spec_name(self):
        """Test handling of None spec name"""
        # Given
        spec_name = None
        mock_strategy = Mock()
        mock_context_manager = Mock()
        mock_strategy.work_context.return_value = mock_context_manager

        with patch.object(self.handler, '_get_git_strategy', return_value=mock_strategy):
            # When & Then
            with pytest.raises(AttributeError):
                self.handler.execute_git_workflow(spec_name)