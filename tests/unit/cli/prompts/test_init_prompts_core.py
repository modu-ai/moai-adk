"""
High-coverage unit tests for prompt_project_setup and helper functions.

Tests focus on real execution paths including user input handling,
language selection, and error recovery.
"""

from pathlib import Path
from unittest.mock import patch

import pytest

from moai_adk.cli.prompts.init_prompts import (
    _prompt_select,
    _prompt_text,
    prompt_project_setup,
)


class TestPromptProjectSetup:
    """Test main project setup prompt function."""

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value="")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_current_directory_no_name_input(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup with current directory uses dirname."""
        # Arrange - current flow needs 5 select values:
        # 1. locale, 2. git_mode, 3. git_commit_lang, 4. code_comment_lang, 5. doc_lang
        mock_select.side_effect = [
            "en",  # locale
            "manual",  # git_mode
            "en",  # git_commit_lang
            "en",  # code_comment_lang
            "en",  # doc_lang
        ]
        mock_text.return_value = "my-project"  # user_name and project_name

        # Act
        result = prompt_project_setup(is_current_dir=True, project_path=Path("/test/my-project"))

        # Assert
        assert result["project_name"] == "my-project"
        assert result["locale"] == "en"
        assert result["git_mode"] == "manual"

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value="")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_provided_project_name(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup with provided project name."""
        # Arrange - current flow needs 5 select values
        mock_select.side_effect = [
            "ko",  # locale
            "manual",  # git_mode
            "ko",  # git_commit_lang
            "en",  # code_comment_lang
            "ko",  # doc_lang
        ]
        mock_text.return_value = "test-project"

        # Act
        result = prompt_project_setup(
            project_name="test-project",
            is_current_dir=False,
        )

        # Assert
        assert result["project_name"] == "test-project"
        assert result["locale"] == "ko"
        mock_text.assert_called()

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value="")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_prompt_project_name(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup prompts for project name when not provided."""
        # Arrange
        mock_text.return_value = "my-new-project"
        mock_select.return_value = "en"

        # Act
        result = prompt_project_setup(is_current_dir=False)

        # Assert
        assert result["project_name"] == "my-new-project"
        mock_text.assert_called()

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value="")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_language_korean(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup with Korean language selection."""
        # Arrange - current flow needs 5 select values:
        # 1. locale, 2. git_mode, 3. git_commit_lang, 4. code_comment_lang, 5. doc_lang
        mock_select.side_effect = [
            "ko",  # locale
            "manual",  # git_mode
            "ko",  # git_commit_lang
            "en",  # code_comment_lang
            "ko",  # doc_lang
        ]
        mock_text.return_value = "test"  # user_name and project_name

        # Act
        result = prompt_project_setup(
            project_name="test",
            is_current_dir=True,
            project_path=Path("/test/project"),
        )

        # Assert
        assert result["locale"] == "ko"
        assert result["git_commit_lang"] == "ko"

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value="")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_language_japanese(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup with Japanese language selection."""
        # Arrange
        mock_select.return_value = "ja"

        # Act
        result = prompt_project_setup(
            project_name="test",
            is_current_dir=True,
            project_path=Path("/test/project"),
        )

        # Assert
        assert result["locale"] == "ja"

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value="")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_language_chinese(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup with Chinese language selection."""
        # Arrange
        mock_select.return_value = "zh"

        # Act
        result = prompt_project_setup(
            project_name="test",
            is_current_dir=True,
            project_path=Path("/test/project"),
        )

        # Assert
        assert result["locale"] == "zh"

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_glm_api_key(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup with GLM API key input."""
        # Arrange - current flow needs 5 select values:
        # 1. locale, 2. git_mode, 3. git_commit_lang, 4. code_comment_lang, 5. doc_lang
        mock_select.side_effect = [
            "zh",  # locale
            "manual",  # git_mode
            "zh",  # git_commit_lang
            "en",  # code_comment_lang
            "zh",  # doc_lang
        ]
        mock_text.return_value = "test"  # user_name and project_name
        mock_password.return_value = "glm-api-key"  # GLM API key

        # Act
        result = prompt_project_setup(
            project_name="test",
            is_current_dir=True,
            project_path=Path("/test/project"),
        )

        # Assert
        assert result["locale"] == "zh"
        assert result["glm_api_key"] == "glm-api-key"
        mock_text.assert_called()

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value="")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_with_initial_locale(self, mock_text, mock_select, mock_password, mock_glm_exists):
        """Test setup respects initial_locale preference."""
        # Arrange
        mock_select.return_value = "ja"

        # Act
        result = prompt_project_setup(
            project_name="test",
            is_current_dir=True,
            project_path=Path("/test/project"),
            initial_locale="ja",
        )

        # Assert
        assert result["locale"] == "ja"

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_keyboard_interrupt_text(self, mock_text, mock_select, mock_glm_exists):
        """Test setup handles KeyboardInterrupt from text prompt."""
        # Arrange
        mock_text.side_effect = KeyboardInterrupt()

        # Act & Assert
        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(
                is_current_dir=False,
            )

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_keyboard_interrupt_select(self, mock_text, mock_select, mock_glm_exists):
        """Test setup handles KeyboardInterrupt from select prompt."""
        # Arrange
        mock_text.return_value = "test"
        mock_select.side_effect = KeyboardInterrupt()

        # Act & Assert
        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(
                is_current_dir=False,
            )

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_none_text_input(self, mock_text, mock_select, mock_glm_exists):
        """Test setup handles None from text prompt (cancellation)."""
        # Arrange
        mock_text.return_value = None

        # Act & Assert
        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(
                is_current_dir=False,
            )

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_none_select_input(self, mock_text, mock_select, mock_glm_exists):
        """Test setup handles None from select prompt (cancellation)."""
        # Arrange
        mock_text.return_value = "test"
        mock_select.return_value = None

        # Act & Assert
        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(
                is_current_dir=False,
            )

    @patch("moai_adk.core.credentials.glm_env_exists", return_value=False)
    @patch("moai_adk.cli.prompts.init_prompts._prompt_select")
    @patch("moai_adk.cli.prompts.init_prompts._prompt_text")
    def test_setup_none_custom_language(self, mock_text, mock_select, mock_glm_exists):
        """Test setup handles None when custom language prompt cancelled."""
        # Arrange
        mock_select.return_value = "other"
        mock_text.return_value = None

        # Act & Assert
        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(
                project_name="test",
                is_current_dir=True,
                project_path=Path("/test/project"),
            )

    def test_setup_returns_typed_dict(self):
        """Test setup returns proper ProjectSetupAnswers dict."""
        # Arrange
        with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
            with patch(
                "moai_adk.cli.prompts.init_prompts._prompt_password_optional",
                return_value="",
            ):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_text") as mock_text:
                    with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
                        mock_text.return_value = "test-project"
                        # Current flow needs 5 select values:
                        # 1. locale, 2. git_mode, 3. git_commit_lang, 4. code_comment_lang, 5. doc_lang
                        mock_select.side_effect = [
                            "en",  # locale
                            "manual",  # git_mode
                            "en",  # git_commit_lang
                            "en",  # code_comment_lang
                            "en",  # doc_lang
                        ]

                        # Act
                        result = prompt_project_setup(is_current_dir=False)

                        # Assert - current ProjectSetupAnswers fields
                        assert isinstance(result, dict)
                        assert "project_name" in result
                        assert "locale" in result
                        assert "user_name" in result
                        assert "glm_api_key" in result
                        assert "git_mode" in result
                        assert "git_commit_lang" in result
                        assert "code_comment_lang" in result
                        assert "doc_lang" in result


class TestPromptText:
    """Test _prompt_text helper function."""

    @patch("moai_adk.cli.ui.prompts.styled_input")
    def test_prompt_text_styled_input_success(self, mock_styled):
        """Test _prompt_text uses styled_input when available."""
        # Arrange
        mock_styled.return_value = "user input"

        # Act
        result = _prompt_text("Enter name:", default="default", required=True)

        # Assert
        assert result == "user input"
        mock_styled.assert_called_once()

    @patch("moai_adk.cli.ui.prompts.styled_input")
    def test_prompt_text_returns_input(self, mock_styled):
        """Test _prompt_text returns input from styled_input."""
        # Arrange
        mock_styled.return_value = "user input"

        # Act
        result = _prompt_text("Enter name:")

        # Assert
        assert result == "user input"
        mock_styled.assert_called_once()

    @patch("moai_adk.cli.ui.prompts.styled_input")
    def test_prompt_text_styled_input_with_validation(self, mock_styled):
        """Test _prompt_text with required parameter."""
        # Arrange
        mock_styled.return_value = "validated input"

        # Act
        result = _prompt_text("Enter name:", required=True)

        # Assert
        assert result == "validated input"
        call_args = mock_styled.call_args
        assert call_args[1]["required"] is True

    @patch("moai_adk.cli.ui.prompts.styled_input")
    def test_prompt_text_returns_none(self, mock_styled):
        """Test _prompt_text returns None when cancelled."""
        # Arrange
        mock_styled.return_value = None

        # Act
        result = _prompt_text("Enter name:")

        # Assert
        assert result is None

    @patch("moai_adk.cli.ui.prompts.styled_input")
    def test_prompt_text_default_value(self, mock_styled):
        """Test _prompt_text passes default value."""
        # Arrange
        mock_styled.return_value = "default-value"

        # Act
        result = _prompt_text("Enter name:", default="default-value")

        # Assert
        assert result == "default-value"
        call_args = mock_styled.call_args
        assert call_args[1]["default"] == "default-value"


class TestPromptSelect:
    """Test _prompt_select helper function."""

    @patch("moai_adk.cli.ui.prompts.styled_select")
    def test_prompt_select_styled_success(self, mock_styled):
        """Test _prompt_select uses styled_select when available."""
        # Arrange
        choices = [
            {"name": "Option 1", "value": "val1"},
            {"name": "Option 2", "value": "val2"},
        ]
        mock_styled.return_value = "val1"

        # Act
        result = _prompt_select("Choose:", choices=choices)

        # Assert
        assert result == "val1"
        mock_styled.assert_called_once()

    @patch("moai_adk.cli.ui.prompts.styled_select")
    def test_prompt_select_styled_with_default(self, mock_styled):
        """Test _prompt_select passes default value."""
        # Arrange
        choices = [
            {"name": "Option 1", "value": "val1"},
            {"name": "Option 2", "value": "val2"},
        ]
        mock_styled.return_value = "val2"

        # Act
        result = _prompt_select("Choose:", choices=choices, default="val2")

        # Assert
        assert result == "val2"
        call_args = mock_styled.call_args
        assert call_args[1]["default"] == "val2"

    @patch("moai_adk.cli.ui.prompts.styled_select")
    def test_prompt_select_passes_choices(self, mock_styled):
        """Test _prompt_select passes choices correctly."""
        # Arrange
        mock_styled.return_value = "val1"
        choices = [
            {"name": "Option 1", "value": "val1"},
            {"name": "Option 2", "value": "val2"},
        ]

        # Act
        result = _prompt_select("Choose:", choices=choices)

        # Assert
        assert result == "val1"
        mock_styled.assert_called_once()

    @patch("moai_adk.cli.ui.prompts.styled_select")
    def test_prompt_select_returns_none(self, mock_styled):
        """Test _prompt_select returns None when cancelled."""
        # Arrange
        mock_styled.return_value = None
        choices = [
            {"name": "Option 1", "value": "val1"},
            {"name": "Option 2", "value": "val2"},
        ]

        # Act
        result = _prompt_select("Choose:", choices=choices)

        # Assert
        assert result is None

    @patch("moai_adk.cli.ui.prompts.styled_select")
    def test_prompt_select_with_none(self, mock_styled):
        """Test _prompt_select returns None when cancelled."""
        # Arrange
        mock_styled.return_value = None
        choices = [
            {"name": "Option 1", "value": "val1"},
            {"name": "Option 2", "value": "val2"},
        ]

        # Act
        result = _prompt_select("Choose:", choices=choices)

        # Assert
        assert result is None

    @patch("moai_adk.cli.ui.prompts.styled_select")
    def test_prompt_select_multiple_choices(self, mock_styled):
        """Test _prompt_select with multiple choices."""
        # Arrange
        mock_styled.return_value = "en"
        choices = [
            {"name": "Korean", "value": "ko"},
            {"name": "English", "value": "en"},
            {"name": "Japanese", "value": "ja"},
        ]

        # Act
        result = _prompt_select("Choose language:", choices=choices)

        # Assert
        assert result == "en"

    @patch("moai_adk.cli.ui.prompts.styled_select")
    def test_prompt_select_with_default_parameter(self, mock_styled):
        """Test _prompt_select with default parameter."""
        # Arrange
        mock_styled.return_value = "val2"
        choices = [
            {"name": "Option 1", "value": "val1"},
            {"name": "Option 2", "value": "val2"},
            {"name": "Option 3", "value": "val3"},
        ]

        # Act
        result = _prompt_select("Choose:", choices=choices, default="val2")

        # Assert
        assert result == "val2"
        call_kwargs = mock_styled.call_args[1]
        assert call_kwargs["default"] == "val2"
