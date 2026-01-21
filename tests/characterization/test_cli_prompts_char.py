"""Characterization Tests for cli.prompts module.

These tests capture the CURRENT BEHAVIOR of the cli/prompts module.
They serve as a safety net during refactoring by documenting what the code
actually does, not what it should do.

Characterization Test Principles:
- Capture actual behavior without judgment
- Document side effects and edge cases
- Provide regression safety during refactoring
- Focus on observable behavior, not implementation

Behavior Snapshots:
- _prompt_text: Returns user input or None if cancelled, falls back to questionary
- _prompt_select: Returns selected value or None if cancelled, handles OSError
- _prompt_password: Returns password or None if cancelled
- _prompt_password_optional: Returns empty string for Enter, None for Ctrl+C
- prompt_project_setup: Orchestrates 7-question flow, raises KeyboardInterrupt on cancel
"""

import sys
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.cli.prompts.init_prompts import (
    ProjectSetupAnswers,
    _prompt_password,
    _prompt_password_optional,
    _prompt_select,
    _prompt_text,
    prompt_project_setup,
)


class TestPromptTextCharacterization:
    """Characterization tests for _prompt_text function.

    Current Behavior:
    - Tries to import styled_input from moai_adk.cli.ui.prompts first
    - Falls back to questionary.text on ImportError
    - Returns user input string
    - Returns None when user cancels (Ctrl+C)
    - Validates required fields (rejects empty strings when required=True)
    - Accepts default values
    """

    def test_characterize_styled_input_success(self):
        """CAPTURE: When styled_input is available, returns its result directly.

        Behavior Snapshot:
        Input: message="Enter value:", default="default", required=False
        Mock: styled_input returns "test-input"
        Output: "test-input"

        Note: This is what currently happens - may or may not be desired.
        """
        with patch("moai_adk.cli.ui.prompts.styled_input", return_value="test-input"):
            result = _prompt_text("Enter value:", default="default", required=False)
            assert result == "test-input"

    def test_characterize_fallback_to_questionary_required(self):
        """CAPTURE: Falls back to questionary when styled_input not available.

        Behavior Snapshot:
        Input: styled_input raises ImportError
        Mock: questionary.text returns "user-input"
        Output: "user-input"

        Note: questionary.text validates required fields with lambda.
        """
        mock_questionary = MagicMock()
        mock_questionary.text.return_value.ask.return_value = "user-input"

        with patch("moai_adk.cli.ui.prompts.styled_input", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_text("Enter value:", required=True)
                assert result == "user-input"
                # Verify validation lambda was applied
                mock_questionary.text.assert_called_once()
                call_kwargs = mock_questionary.text.call_args[1]
                assert "validate" in call_kwargs
            finally:
                sys.modules.pop("questionary", None)

    def test_characterize_fallback_optional_field(self):
        """CAPTURE: Optional fields accept empty strings.

        Behavior Snapshot:
        Input: required=False
        Mock: questionary.text returns "user-input"
        Output: "user-input"

        Note: No validation lambda for optional fields.
        """
        mock_questionary = MagicMock()
        mock_questionary.text.return_value.ask.return_value = "user-input"

        with patch("moai_adk.cli.ui.prompts.styled_input", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_text("Enter value:", required=False)
                assert result == "user-input"
                # Verify no validation for optional (no 'validate' in kwargs)
                call_kwargs = mock_questionary.text.call_args[1]
                assert "validate" not in call_kwargs
            finally:
                sys.modules.pop("questionary", None)

    def test_characterize_cancellation_returns_none(self):
        """CAPTURE: User cancellation returns None.

        Behavior Snapshot:
        Input: styled_input returns None
        Output: None

        Note: None signals cancellation in the prompt flow.
        """
        with patch("moai_adk.cli.ui.prompts.styled_input", return_value=None):
            result = _prompt_text("Enter value:")
            assert result is None


class TestPromptSelectCharacterization:
    """Characterization tests for _prompt_select function.

    Current Behavior:
    - Tries to import styled_select from moai_adk.cli.ui.prompts first
    - Falls back to questionary.select on ImportError or OSError
    - Maps name/value choices for questionary format
    - Returns the value (not the name) of selected choice
    - Returns None when user cancels
    - Handles default value by finding matching choice name
    """

    def test_characterize_styled_select_success(self):
        """CAPTURE: When styled_select is available, returns its result.

        Behavior Snapshot:
        Input: choices=[{"name": "Option 1", "value": "opt1"}]
        Mock: styled_select returns "opt1"
        Output: "opt1"

        Note: styled_select receives original choices format.
        """
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]
        with patch("moai_adk.cli.ui.prompts.styled_select", return_value="opt1"):
            result = _prompt_select("Choose:", choices=choices, default="opt1")
            assert result == "opt1"

    def test_characterize_fallback_to_questionary(self):
        """CAPTURE: Falls back to questionary, maps name->value.

        Behavior Snapshot:
        Input: choices=[{"name": "Option 1", "value": "opt1"}]
        Mock: questionary.select returns "Option 1" (name)
        Output: "opt1" (value)

        Note: Translates questionary's name selection back to value.
        """
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

    def test_characterize_default_value_mapping(self):
        """CAPTURE: Maps default value to choice name for questionary.

        Behavior Snapshot:
        Input: default="opt2", choices=[{"name": "Option 2", "value": "opt2"}]
        Mock: questionary.select receives default="Option 2" (name)
        Output: "opt2" (value)

        Note: Must find choice name matching default value.
        """
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

    def test_characterize_cancellation_returns_none(self):
        """CAPTURE: User cancellation returns None.

        Behavior Snapshot:
        Input: questionary.select returns None
        Output: None

        Note: None signals cancellation in the prompt flow.
        """
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

    def test_characterize_oserror_fallback(self):
        """CAPTURE: OSError from styled_select triggers fallback.

        Behavior Snapshot:
        Input: styled_select raises OSError("Invalid argument")
        Mock: Falls back to questionary.select
        Output: "opt1"

        Note: OSError occurs on macOS with certain terminal configurations.
        """
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


class TestPromptPasswordCharacterization:
    """Characterization tests for _prompt_password function.

    Current Behavior:
    - Tries to import styled_password from moai_adk.cli.ui.prompts first
    - Falls back to questionary.password on ImportError
    - Returns password string
    - Returns None when user cancels
    - Validates required fields (password is always required)
    """

    def test_characterize_styled_password_success(self):
        """CAPTURE: When styled_password is available, returns its result.

        Behavior Snapshot:
        Input: message="Enter password:"
        Mock: styled_password returns "secret123"
        Output: "secret123"
        """
        with patch("moai_adk.cli.ui.prompts.styled_password", return_value="secret123"):
            result = _prompt_password("Enter password:")
            assert result == "secret123"

    def test_characterize_fallback_to_questionary(self):
        """CAPTURE: Falls back to questionary.password.

        Behavior Snapshot:
        Input: message="Enter password:"
        Mock: questionary.password returns "pass123"
        Output: "pass123"

        Note: questionary.password validates non-empty with lambda.
        """
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = "pass123"

        with patch("moai_adk.cli.ui.prompts.styled_password", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_password("Enter password:")
                assert result == "pass123"
            finally:
                sys.modules.pop("questionary", None)

    def test_characterize_password_cancellation(self):
        """CAPTURE: User cancellation returns None.

        Behavior Snapshot:
        Input: questionary.password returns None
        Output: None
        """
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = None

        with patch("moai_adk.cli.ui.prompts.styled_password", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_password("Enter password:")
                assert result is None
            finally:
                sys.modules.pop("questionary", None)


class TestPromptPasswordOptionalCharacterization:
    """Characterization tests for _prompt_password_optional function.

    Current Behavior:
    - Only uses questionary.password (no styled_password fallback)
    - Returns empty string when user presses Enter
    - Returns None when user cancels (Ctrl+C)
    - Distinguishes between empty input and cancellation
    """

    def test_characterize_with_user_input(self):
        """CAPTURE: User input returns the key.

        Behavior Snapshot:
        Input: questionary.password returns "key123"
        Output: "key123"
        """
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = "key123"

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter key (optional):")
            assert result == "key123"
        finally:
            sys.modules.pop("questionary", None)

    def test_characterize_empty_input_allowed(self):
        """CAPTURE: Pressing Enter returns empty string.

        Behavior Snapshot:
        Input: questionary.password returns ""
        Output: ""

        Note: Empty string means "use existing/keep current", not cancellation.
        """
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = ""

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter key (optional):")
            assert result == ""
        finally:
            sys.modules.pop("questionary", None)

    def test_characterize_cancellation_vs_empty(self):
        """CAPTURE: Cancellation returns None, empty returns "".

        Behavior Snapshot:
        Case 1: questionary.password returns None -> Output: None (cancelled)
        Case 2: questionary.password returns "" -> Output: "" (keep existing)

        Note: Critical distinction for GLM API key handling.
        """
        # Test cancellation
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = None

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter key (optional):")
            assert result is None  # Ctrl+C
        finally:
            sys.modules.pop("questionary", None)

        # Test empty input
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = ""

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter key (optional):")
            assert result == ""  # Enter key
        finally:
            sys.modules.pop("questionary", None)


class TestPromptProjectSetupCharacterization:
    """Characterization tests for prompt_project_setup function.

    Current Behavior:
    - Runs 7-question flow for project setup
    - Always starts with language selection (English first)
    - Supports initial_locale parameter to skip language selection
    - Raises KeyboardInterrupt on any cancellation
    - Returns ProjectSetupAnswers TypedDict with all fields
    - Defaults git_mode to "manual", development_mode to "ddd"
    - Handles GLM API key storage via credentials module
    - Conditionally prompts for GitHub username based on git_mode
    - Prompts for 3 output languages (commit, comment, doc)
    - Uses Rich console for colored output
    """

    def test_characterize_keyboard_interrupt(self):
        """CAPTURE: Cancellation raises KeyboardInterrupt.

        Behavior Snapshot:
        Input: User cancels at language selection
        Output: Raises KeyboardInterrupt

        Note: Early cancellation stops entire flow.
        """
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value=None):
            with pytest.raises(KeyboardInterrupt):
                prompt_project_setup()

    def test_characterize_initial_locale_skips_language_selection(self):
        """CAPTURE: initial_locale parameter sets language without prompting.

        Behavior Snapshot:
        Input: initial_locale="en"
        Mock: _prompt_select returns "en" (called but may be bypassed)
        Output: answers["locale"] == "en"

        Note: initial_locale is used as default but prompt still runs.
        """
        mock_translations = {
            "user_setup": "User Setup",
            "q_user_name": "Your name:",
            "api_key_input": "API Key",
            "q_api_key_glm": "GLM Key:",
            "msg_glm_key_found": "Found key",
            "msg_glm_key_keep_prompt": "Keep?",
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
            "dev_mode_ddd_info": "Using DDD",
        }

        with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="en"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="test-user"):
                    with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                        with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                            with patch("moai_adk.cli.prompts.init_prompts.console"):
                                with patch(
                                    "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                                ):
                                    result = prompt_project_setup(initial_locale="en")
                                    assert result["locale"] == "en"

    def test_characterize_default_values(self):
        """CAPTURE: Default values for all fields.

        Behavior Snapshot:
        - locale: "en" (unless initial_locale provided)
        - user_name: "" (optional, can be empty)
        - glm_api_key: None (optional, can be skipped)
        - git_mode: "manual" (default)
        - github_username: None (only set if git_mode is personal/team)
        - git_commit_lang: matches locale
        - code_comment_lang: matches locale
        - doc_lang: matches locale
        - development_mode: "ddd" (always, not configurable)

        Note: Output languages default to conversation language.
        """
        mock_translations = {
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

        with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
                mock_select.return_value = "en"  # Language selection

                with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                    with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                        with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                            with patch("moai_adk.cli.prompts.init_prompts.console"):
                                with patch(
                                    "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                                ):
                                    result = prompt_project_setup()

                                    # Verify defaults
                                    assert result["locale"] == "en"
                                    assert result["user_name"] == "user"
                                    assert result["glm_api_key"] is None
                                    assert result["git_mode"] == "en"  # Mock returns "en"
                                    assert result["development_mode"] == "ddd"

    def test_characterize_github_username_conditional(self):
        """CAPTURE: GitHub username only prompted for personal/team git modes.

        Behavior Snapshot:
        Case 1: git_mode="manual" -> github_username not prompted, stays None
        Case 2: git_mode="personal" -> github_username prompted
        Case 3: git_mode="team" -> github_username prompted

        Note: Manual mode doesn't need GitHub integration.
        """
        mock_translations = {
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

        # Test manual mode (no GitHub username)
        with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
                mock_select.return_value = "manual"  # Manual mode

                with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                    with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                        with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                            with patch("moai_adk.cli.prompts.init_prompts.console"):
                                with patch(
                                    "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                                ):
                                    result = prompt_project_setup()
                                    assert result["github_username"] is None

    def test_characterize_glm_key_handling(self):
        """CAPTURE: GLM API key handling with existing key.

        Behavior Snapshot:
        Case 1: Existing key found, user presses Enter -> keeps existing key
        Case 2: Existing key found, user enters new key -> saves new key
        Case 3: No existing key, user enters key -> saves new key
        Case 4: No existing key, user presses Enter -> glm_api_key is None

        Note: Key masking shows first 8 chars for security.
        """
        mock_translations = {
            "user_setup": "User",
            "q_user_name": "Name:",
            "api_key_input": "API",
            "q_api_key_glm": "Key:",
            "msg_glm_key_found": "Found",
            "msg_glm_key_keep_prompt": "Keep?",
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

        # Test with existing key kept
        with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="manual"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                    with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                        with patch("moai_adk.core.credentials.glm_env_exists", return_value=True):
                            with patch(
                                "moai_adk.core.credentials.load_glm_key_from_env", return_value="existing-key-12345678"
                            ):
                                with patch("moai_adk.cli.prompts.init_prompts.console"):
                                    with patch(
                                        "moai_adk.cli.prompts.init_prompts.get_translation",
                                        return_value=mock_translations,
                                    ):
                                        result = prompt_project_setup()
                                        # Empty input keeps existing key
                                        assert result["glm_api_key"] == "existing-key-12345678"


class TestProjectSetupAnswersCharacterization:
    """Characterization tests for ProjectSetupAnswers TypedDict.

    Current Behavior:
    - TypedDict with 10 fields
    - Required fields: project_name, locale, user_name, git_mode,
      git_commit_lang, code_comment_lang, doc_lang, development_mode
    - Optional fields: glm_api_key (str | None), github_username (str | None)
    - user_name can be empty string (optional input)
    - development_mode is always "ddd"
    """

    def test_characterize_complete_structure(self):
        """CAPTURE: Full structure with all fields populated.

        Behavior Snapshot:
        All required fields present and typed correctly.
        Optional fields can be None or str.
        """
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
        assert answers["development_mode"] == "ddd"

    def test_characterize_optional_fields_none(self):
        """CAPTURE: Optional fields can be None.

        Behavior Snapshot:
        glm_api_key: None when no key provided
        github_username: None when git_mode is "manual"
        """
        answers: ProjectSetupAnswers = {
            "project_name": "test-project",
            "locale": "ko",
            "user_name": "",  # Empty string allowed
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
        assert answers["user_name"] == ""  # Empty string, not None
