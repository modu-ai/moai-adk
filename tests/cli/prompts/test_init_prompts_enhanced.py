"""Comprehensive test suite for init_prompts.py

Tests cover:
- Project name handling (provided, prompted, current directory)
- Language selection (Korean, English, Japanese, Chinese, Other)
- Custom language input
- User input validation
- Keyboard interrupt handling
- Default value behavior
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.cli.prompts.init_prompts import prompt_project_setup


class TestProjectNameHandling:
    """Test project name input scenarios."""

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_project_name_provided_not_current_dir(self, mock_select):
        """Test when project name is provided and not using current directory."""
        # Mock language selection
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False)

        assert result["project_name"] == "test-project"
        assert result["mode"] == "personal"
        assert result["language"] is None
        assert result["author"] == ""

    @patch("moai_adk.cli.prompts.init_prompts.questionary.text")
    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_project_name_prompted_valid_input(self, mock_select, mock_text):
        """Test prompting for project name with valid user input."""
        # Mock project name prompt
        mock_text_instance = MagicMock()
        mock_text_instance.ask.return_value = "my-custom-project"
        mock_text.return_value = mock_text_instance

        # Mock language selection (English)
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name=None, is_current_dir=False)

        assert result["project_name"] == "my-custom-project"
        mock_text.assert_called_once()

    @patch("moai_adk.cli.prompts.init_prompts.questionary.text")
    def test_project_name_prompted_cancelled(self, mock_text):
        """Test when user cancels project name prompt (Ctrl+C)."""
        mock_text_instance = MagicMock()
        mock_text_instance.ask.return_value = None  # User cancelled
        mock_text.return_value = mock_text_instance

        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(project_name=None, is_current_dir=False)

    @patch("moai_adk.cli.prompts.init_prompts.Path.cwd")
    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_project_name_current_dir_with_path(self, mock_select, mock_cwd):
        """Test using current directory name with provided project_path."""
        mock_cwd.return_value = Path("/fallback/path")
        project_path = Path("/user/execution/custom-project")

        # Mock language selection
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name=None, is_current_dir=True, project_path=project_path)

        assert result["project_name"] == "custom-project"
        mock_cwd.assert_not_called()  # Should not use fallback

    @patch("moai_adk.cli.prompts.init_prompts.Path.cwd")
    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_project_name_current_dir_fallback(self, mock_select, mock_cwd):
        """Test using current directory name without project_path (fallback)."""
        mock_cwd.return_value = Path("/fallback/directory-name")

        # Mock language selection
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name=None, is_current_dir=True, project_path=None)

        assert result["project_name"] == "directory-name"
        mock_cwd.assert_called_once()


class TestLanguageSelection:
    """Test language selection scenarios."""

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_selection_korean(self, mock_select):
        """Test selecting Korean language."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Korean (한국어)"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False)

        assert result["locale"] == "ko"
        assert result["custom_language"] is None

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_selection_english(self, mock_select):
        """Test selecting English language."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False)

        assert result["locale"] == "en"
        assert result["custom_language"] is None

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_selection_japanese(self, mock_select):
        """Test selecting Japanese language."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Japanese (日本語)"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False)

        assert result["locale"] == "ja"
        assert result["custom_language"] is None

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_selection_chinese(self, mock_select):
        """Test selecting Chinese language."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Chinese (中文)"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False)

        assert result["locale"] == "zh"
        assert result["custom_language"] is None

    @patch("moai_adk.cli.prompts.init_prompts.questionary.text")
    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_selection_other_with_custom_input(self, mock_select, mock_text):
        """Test selecting 'Other' and providing custom language."""
        # Mock language selection (Other)
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Other - Manual input"
        mock_select.return_value = mock_select_instance

        # Mock custom language input
        mock_text_instance = MagicMock()
        mock_text_instance.ask.return_value = "Spanish"
        mock_text.return_value = mock_text_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False)

        assert result["locale"] == "other"
        assert result["custom_language"] == "Spanish"

    @patch("moai_adk.cli.prompts.init_prompts.questionary.text")
    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_selection_other_cancelled_custom_input(self, mock_select, mock_text):
        """Test cancelling custom language input after selecting 'Other'."""
        # Mock language selection (Other)
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Other - Manual input"
        mock_select.return_value = mock_select_instance

        # Mock custom language input (cancelled)
        mock_text_instance = MagicMock()
        mock_text_instance.ask.return_value = None
        mock_text.return_value = mock_text_instance

        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(project_name="test-project", is_current_dir=False)

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_selection_cancelled(self, mock_select):
        """Test when user cancels language selection."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = None
        mock_select.return_value = mock_select_instance

        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(project_name="test-project", is_current_dir=False)


class TestDefaultLocaleHandling:
    """Test default locale selection behavior."""

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_initial_locale_korean(self, mock_select):
        """Test default choice when initial_locale is 'ko'."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Korean (한국어)"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False, initial_locale="ko")

        # Verify questionary.select was called with Korean as default
        call_args = mock_select.call_args
        assert call_args.kwargs["default"] == "Korean (한국어)"
        assert result["locale"] == "ko"

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_initial_locale_english(self, mock_select):
        """Test default choice when initial_locale is 'en'."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False, initial_locale="en")

        call_args = mock_select.call_args
        assert call_args.kwargs["default"] == "English"
        assert result["locale"] == "en"

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_initial_locale_japanese(self, mock_select):
        """Test default choice when initial_locale is 'ja'."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Japanese (日本語)"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False, initial_locale="ja")

        call_args = mock_select.call_args
        assert call_args.kwargs["default"] == "Japanese (日本語)"
        assert result["locale"] == "ja"

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_initial_locale_chinese(self, mock_select):
        """Test default choice when initial_locale is 'zh'."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Chinese (中文)"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False, initial_locale="zh")

        call_args = mock_select.call_args
        assert call_args.kwargs["default"] == "Chinese (中文)"
        assert result["locale"] == "zh"

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_initial_locale_invalid_defaults_to_english(self, mock_select):
        """Test when initial_locale is invalid/unknown, defaults to English."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False, initial_locale="fr")

        call_args = mock_select.call_args
        assert call_args.kwargs["default"] == "English"  # Should fallback to index 1
        assert result["locale"] == "en"

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_initial_locale_none_defaults_to_english(self, mock_select):
        """Test when initial_locale is None, defaults to English."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False, initial_locale=None)

        call_args = mock_select.call_args
        assert call_args.kwargs["default"] == "English"
        assert result["locale"] == "en"


class TestAnswersStructure:
    """Test ProjectSetupAnswers structure."""

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_answers_default_values(self, mock_select):
        """Test that all fields have correct default values."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name="test-project", is_current_dir=False)

        assert isinstance(result, dict)
        assert "project_name" in result
        assert "mode" in result
        assert "locale" in result
        assert "language" in result
        assert "author" in result
        assert "custom_language" in result

        assert result["mode"] == "personal"
        assert result["language"] is None
        assert result["author"] == ""


class TestKeyboardInterruptHandling:
    """Test keyboard interrupt (Ctrl+C) handling throughout the flow."""

    @patch("moai_adk.cli.prompts.init_prompts.questionary.text")
    def test_keyboard_interrupt_during_project_name_prompt(self, mock_text):
        """Test KeyboardInterrupt during project name input."""
        mock_text_instance = MagicMock()
        mock_text_instance.ask.return_value = None
        mock_text.return_value = mock_text_instance

        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(project_name=None, is_current_dir=False)

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_keyboard_interrupt_during_language_selection(self, mock_select):
        """Test KeyboardInterrupt during language selection."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = None
        mock_select.return_value = mock_select_instance

        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(project_name="test-project", is_current_dir=False)

    @patch("moai_adk.cli.prompts.init_prompts.questionary.text")
    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_keyboard_interrupt_during_custom_language_input(self, mock_select, mock_text):
        """Test KeyboardInterrupt during custom language input."""
        # Mock language selection (Other)
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "Other - Manual input"
        mock_select.return_value = mock_select_instance

        # Mock custom language input (cancelled)
        mock_text_instance = MagicMock()
        mock_text_instance.ask.return_value = None
        mock_text.return_value = mock_text_instance

        with pytest.raises(KeyboardInterrupt):
            prompt_project_setup(project_name="test-project", is_current_dir=False)


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_language_mapping_all_choices(self, mock_select):
        """Test that all language choices map correctly."""
        language_choices = [
            ("Korean (한국어)", "ko"),
            ("English", "en"),
            ("Japanese (日本語)", "ja"),
            ("Chinese (中文)", "zh"),
            ("Other - Manual input", "other"),
        ]

        for choice_name, expected_locale in language_choices:
            mock_select_instance = MagicMock()
            mock_select_instance.ask.return_value = choice_name
            mock_select.return_value = mock_select_instance

            # For "Other", need to mock text input
            if choice_name == "Other - Manual input":
                with patch("moai_adk.cli.prompts.init_prompts.questionary.text") as mock_text:
                    mock_text_instance = MagicMock()
                    mock_text_instance.ask.return_value = "Test Language"
                    mock_text.return_value = mock_text_instance

                    result = prompt_project_setup(project_name="test-project", is_current_dir=False)
                    assert result["locale"] == expected_locale
                    assert result["custom_language"] == "Test Language"
            else:
                result = prompt_project_setup(project_name="test-project", is_current_dir=False)
                assert result["locale"] == expected_locale
                assert result["custom_language"] is None

    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_project_name_with_special_characters(self, mock_select):
        """Test project name with special characters."""
        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        special_names = [
            "my-project-123",
            "project_with_underscores",
            "project.with.dots",
            "UPPERCASE-PROJECT",
        ]

        for name in special_names:
            result = prompt_project_setup(project_name=name, is_current_dir=False)
            assert result["project_name"] == name

    @patch("moai_adk.cli.prompts.init_prompts.Path.cwd")
    @patch("moai_adk.cli.prompts.init_prompts.questionary.select")
    def test_current_dir_with_complex_path(self, mock_select, mock_cwd):
        """Test current directory with complex path structure."""
        mock_cwd.return_value = Path("/very/long/nested/path/with/many/directories/project-name")

        mock_select_instance = MagicMock()
        mock_select_instance.ask.return_value = "English"
        mock_select.return_value = mock_select_instance

        result = prompt_project_setup(project_name=None, is_current_dir=True, project_path=None)

        assert result["project_name"] == "project-name"
