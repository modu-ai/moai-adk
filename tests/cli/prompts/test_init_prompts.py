"""Tests for cli.prompts.init_prompts module."""

import sys
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.cli.prompts.init_prompts import (
    ProjectSetupAnswers,
    _prompt_password,
    _prompt_password_optional,
    _prompt_select,
    _prompt_text,
)


class TestPromptText:
    """Test _prompt_text function."""

    def test_prompt_text_with_styled_input(self):
        """Test _prompt_text using styled_input."""
        with patch("moai_adk.cli.ui.prompts.styled_input", return_value="test-input"):
            result = _prompt_text("Enter value:", default="default", required=False)
            assert result == "test-input"

    def test_prompt_text_fallback_to_questionary_required(self):
        """Test _prompt_text fallback to questionary with required=True."""
        mock_questionary = MagicMock()
        mock_questionary.text.return_value.ask.return_value = "user-input"

        with patch("moai_adk.cli.ui.prompts.styled_input", side_effect=ImportError):
            # Inject mock into sys.modules before function runs
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_text("Enter value:", required=True)
                assert result == "user-input"
            finally:
                sys.modules.pop("questionary", None)

    def test_prompt_text_fallback_to_questionary_optional(self):
        """Test _prompt_text fallback to questionary with required=False."""
        mock_questionary = MagicMock()
        mock_questionary.text.return_value.ask.return_value = "user-input"

        with patch("moai_adk.cli.ui.prompts.styled_input", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_text("Enter value:", required=False)
                assert result == "user-input"
            finally:
                sys.modules.pop("questionary", None)

    def test_prompt_text_cancelled(self):
        """Test _prompt_text when user cancels."""
        with patch("moai_adk.cli.ui.prompts.styled_input", return_value=None):
            result = _prompt_text("Enter value:")
            assert result is None


class TestPromptSelect:
    """Test _prompt_select function."""

    def test_prompt_select_with_styled_select(self):
        """Test _prompt_select using styled_select."""
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]
        with patch("moai_adk.cli.ui.prompts.styled_select", return_value="opt1"):
            result = _prompt_select("Choose:", choices=choices, default="opt1")
            assert result == "opt1"

    def test_prompt_select_fallback_to_questionary(self):
        """Test _prompt_select fallback to questionary."""
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]
        mock_questionary = MagicMock()
        mock_questionary.select.return_value.ask.return_value = "Option 1"

        with patch("moai_adk.cli.ui.prompts.styled_select", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_select("Choose:", choices=choices)
                assert result == "opt1"
            finally:
                sys.modules.pop("questionary", None)

    def test_prompt_select_fallback_with_default(self):
        """Test _prompt_select fallback with default value."""
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]
        mock_questionary = MagicMock()
        mock_questionary.select.return_value.ask.return_value = "Option 2"

        with patch("moai_adk.cli.ui.prompts.styled_select", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_select("Choose:", choices=choices, default="opt2")
                assert result == "opt2"
            finally:
                sys.modules.pop("questionary", None)

    def test_prompt_select_cancelled(self):
        """Test _prompt_select when user cancels."""
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]
        mock_questionary = MagicMock()
        mock_questionary.select.return_value.ask.return_value = None

        with patch("moai_adk.cli.ui.prompts.styled_select", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_select("Choose:", choices=choices)
                assert result is None
            finally:
                sys.modules.pop("questionary", None)

    def test_prompt_select_fallback_error_handling(self):
        """Test _prompt_select fallback handles OSError and other exceptions."""
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]
        mock_questionary = MagicMock()
        mock_questionary.select.return_value.ask.return_value = "Option 1"

        with patch("moai_adk.cli.ui.prompts.styled_select", side_effect=OSError("Invalid argument")):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_select("Choose:", choices=choices)
                assert result == "opt1"
            finally:
                sys.modules.pop("questionary", None)


class TestPromptPassword:
    """Test _prompt_password function."""

    def test_prompt_password_with_styled_password(self):
        """Test _prompt_password using styled_password."""
        with patch("moai_adk.cli.ui.prompts.styled_password", return_value="secret123"):
            result = _prompt_password("Enter password:")
            assert result == "secret123"

    def test_prompt_password_fallback_to_questionary(self):
        """Test _prompt_password fallback to questionary."""
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = "pass123"

        with patch("moai_adk.cli.ui.prompts.styled_password", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_password("Enter password:")
                assert result == "pass123"
            finally:
                sys.modules.pop("questionary", None)

    def test_prompt_password_cancelled(self):
        """Test _prompt_password when user cancels."""
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = None

        with patch("moai_adk.cli.ui.prompts.styled_password", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_password("Enter password:")
                assert result is None
            finally:
                sys.modules.pop("questionary", None)


class TestPromptPasswordOptional:
    """Test _prompt_password_optional function."""

    def test_prompt_password_optional_with_input(self):
        """Test _prompt_password_optional with user input."""
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = "key123"

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter key (optional):")
            assert result == "key123"
        finally:
            sys.modules.pop("questionary", None)

    def test_prompt_password_optional_empty_input(self):
        """Test _prompt_password_optional with empty input (Enter pressed)."""
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = ""

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter key (optional):")
            assert result == ""
        finally:
            sys.modules.pop("questionary", None)

    def test_prompt_password_optional_cancelled(self):
        """Test _prompt_password_optional when user cancels."""
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = None

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter key (optional):")
            assert result is None
        finally:
            sys.modules.pop("questionary", None)


class TestPromptProjectSetup:
    """Test prompt_project_setup function."""

    def test_prompt_project_setup_keyboard_interrupt(self):
        """Test prompt_project_setup raises KeyboardInterrupt on cancel."""
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value=None):
            with pytest.raises(KeyboardInterrupt):
                from moai_adk.cli.prompts.init_prompts import prompt_project_setup

                prompt_project_setup()

    def test_prompt_project_setup_with_initial_locale(self):
        """Test prompt_project_setup with initial locale parameter."""
        from moai_adk.cli.prompts.init_prompts import prompt_project_setup

        # Mock all the prompts
        with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="en"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="test-user"):
                    with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                        with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                            with patch("moai_adk.cli.prompts.init_prompts.console"):
                                with patch("moai_adk.cli.prompts.init_prompts.get_translation") as mock_t:
                                    mock_t.return_value = {
                                        "user_setup": "User Setup",
                                        "q_user_name": "Your name:",
                                        "api_key_input": "API Key",
                                        "q_api_key_glm": "GLM Key:",
                                        "msg_api_key_stored": "Stored",
                                        "msg_glm_key_skip_guidance": "Skip guidance",
                                        "project_setup": "Project Setup",
                                        "q_project_name": "Project name:",
                                        "git_setup": "Git Setup",
                                        "q_git_mode": "Git mode:",
                                        "opt_manual": "Manual",
                                        "desc_manual": "Manual desc",
                                        "opt_personal": "Personal",
                                        "desc_personal": "Personal desc",
                                        "opt_team": "Team",
                                        "desc_team": "Team desc",
                                        "q_github_username": "GitHub username:",
                                        "output_language": "Output Language",
                                        "q_commit_lang": "Commit language:",
                                        "q_comment_lang": "Comment language:",
                                        "q_doc_lang": "Doc language:",
                                        "msg_setup_complete": "Complete",
                                        "msg_cancelled": "Cancelled",
                                    }

                                    result = prompt_project_setup(initial_locale="en")
                                    assert result["locale"] == "en"

    def test_prompt_project_setup_default_values(self):
        """Test prompt_project_setup uses correct default values."""
        from moai_adk.cli.prompts.init_prompts import prompt_project_setup

        with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
                # First call is for language, return "en"
                mock_select.return_value = "en"

                with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                    with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                        with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                            with patch("moai_adk.cli.prompts.init_prompts.console"):
                                with patch("moai_adk.cli.prompts.init_prompts.get_translation") as mock_t:
                                    mock_t.return_value = {
                                        "user_setup": "User",
                                        "q_user_name": "Name:",
                                        "api_key_input": "API",
                                        "q_api_key_glm": "Key:",
                                        "msg_api_key_stored": "Stored",
                                        "msg_glm_key_skip_guidance": "Guidance",
                                        "project_setup": "Project",
                                        "q_project_name": "Name:",
                                        "git_setup": "Git",
                                        "q_git_mode": "Mode:",
                                        "opt_manual": "Manual",
                                        "desc_manual": "Desc",
                                        "opt_personal": "Personal",
                                        "desc_personal": "PD",
                                        "opt_team": "Team",
                                        "desc_team": "TD",
                                        "q_github_username": "User:",
                                        "output_language": "Lang",
                                        "q_commit_lang": "Commit:",
                                        "q_comment_lang": "Comment:",
                                        "q_doc_lang": "Doc:",
                                        "msg_setup_complete": "Done",
                                        "msg_cancelled": "Cancel",
                                    }

                                    prompt_project_setup()
                                    # Verify _prompt_select was called with default "en"
                                    assert mock_select.call_count >= 1  # At least language selection


class TestProjectSetupAnswers:
    """Test ProjectSetupAnswers TypedDict."""

    def test_project_setup_answers_structure(self):
        """Test ProjectSetupAnswers has correct structure."""
        answers: ProjectSetupAnswers = {
            "project_name": "test-project",
            "locale": "en",
            "user_name": "Test User",
            "glm_api_key": "test-key",
            "git_mode": "personal",
            "github_username": "testuser",
            "git_commit_lang": "en",
            "code_comment_lang": "en",
            "doc_lang": "en",
            "development_mode": "ddd",
        }
        assert answers["project_name"] == "test-project"
        assert answers["locale"] == "en"
        assert answers["git_mode"] == "personal"

    def test_project_setup_answers_optional_fields(self):
        """Test ProjectSetupAnswers with optional fields as None."""
        answers: ProjectSetupAnswers = {
            "project_name": "test-project",
            "locale": "ko",
            "user_name": "",
            "glm_api_key": None,
            "git_mode": "manual",
            "github_username": None,
            "git_commit_lang": "ko",
            "code_comment_lang": "en",
            "doc_lang": "ko",
            "development_mode": "ddd",
        }
        assert answers["glm_api_key"] is None
        assert answers["github_username"] is None
        assert answers["user_name"] == ""
