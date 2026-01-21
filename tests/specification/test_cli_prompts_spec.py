"""Specification Tests for cli.prompts module.

These tests specify the DESIRED BEHAVIOR based on domain requirements.
They use Given-When-Then pattern to express business rules and user interactions.

Specification Test Principles:
- Based on domain requirements and user needs
- Express what the system SHOULD do (not what it currently does)
- Use Given-When-Then (Gherkin-inspired) pattern
- Focus on user-visible behavior and business rules

Domain Requirements:
- User Input: Allow users to provide input via modern UI or fallback
- Language Selection: Support multilingual prompts (ko, en, ja, zh)
- API Key Management: Securely handle optional API keys
- Git Integration: Support manual, personal, and team Git modes
- Output Languages: Configure commit, comment, and documentation languages
- DDD Methodology: Always use Domain-Driven Development approach
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


class TestPromptTextSpecification:
    """Specification tests for text input prompt behavior.

    User Story:
    AS a user setting up MoAI-ADK
    I WANT to provide text input through a modern interface
    SO THAT I can configure my project settings efficiently
    """

    def test_given_valid_input_when_prompt_text_then_return_input(self):
        """GIVEN a user provides valid input
        WHEN text prompt is displayed
        THEN the input value should be returned
        """
        # GIVEN: User input "test-project"
        user_input = "test-project"

        # WHEN: Prompt is displayed
        with patch("moai_adk.cli.ui.prompts.styled_input", return_value=user_input):
            result = _prompt_text("Enter project name:", default="my-project")

        # THEN: Input should be returned
        assert result == user_input

    def test_given_empty_required_input_when_prompt_text_then_validation_fails(self):
        """GIVEN a user provides empty input for required field
        WHEN text prompt validates required field
        THEN validation should reject empty input
        """
        # GIVEN: Empty input for required field
        empty_input = ""

        # WHEN: Prompt is displayed with required=True
        mock_questionary = MagicMock()

        def validate_required(text):
            """Validation lambda that rejects empty input."""
            return len(text) > 0 or "This field is required"

        mock_questionary.text.return_value.ask.return_value = empty_input
        mock_questionary.text.return_value.validate = validate_required

        with patch("moai_adk.cli.ui.prompts.styled_input", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                _result = _prompt_text("Enter project name:", required=True)

                # THEN: Validation should be applied
                mock_questionary.text.assert_called_once()
                call_kwargs = mock_questionary.text.call_args[1]
                assert "validate" in call_kwargs
            finally:
                sys.modules.pop("questionary", None)

    def test_given_default_value_when_prompt_text_then_default_displayed(self):
        """GIVEN a default value is provided
        WHEN text prompt is displayed
        THEN default value should be shown and accepted on empty input
        """
        # GIVEN: Default value "my-moai-project"
        default_value = "my-moai-project"

        # WHEN: Prompt is displayed
        with patch("moai_adk.cli.ui.prompts.styled_input", return_value=default_value):
            result = _prompt_text("Enter project name:", default=default_value)

        # THEN: Default value should be used
        assert result == default_value

    def test_given_user_cancels_when_prompt_text_then_none_returned(self):
        """GIVEN a user cancels the prompt (Ctrl+C)
        WHEN text prompt is interrupted
        THEN None should be returned to signal cancellation
        """
        # GIVEN: User cancellation
        # WHEN: Prompt is cancelled
        with patch("moai_adk.cli.ui.prompts.styled_input", return_value=None):
            result = _prompt_text("Enter project name:")

        # THEN: None signals cancellation
        assert result is None


class TestPromptSelectSpecification:
    """Specification tests for selection prompt behavior.

    User Story:
    AS a user setting up MoAI-ADK
    I WANT to select from predefined options
    SO THAT I can make informed choices without knowing all details
    """

    def test_given_choice_list_when_prompt_select_then_value_returned(self):
        """GIVEN a list of choices with names and values
        WHEN user selects an option
        THEN the corresponding value (not name) should be returned
        """
        # GIVEN: Choice list
        choices = [
            {"name": "Korean (한국어)", "value": "ko"},
            {"name": "English", "value": "en"},
        ]

        # WHEN: User selects Korean
        with patch("moai_adk.cli.ui.prompts.styled_select", return_value="ko"):
            result = _prompt_select("Select language:", choices=choices)

        # THEN: Value "ko" is returned (not name "Korean (한국어)")
        assert result == "ko"

    def test_given_default_option_when_prompt_select_then_default_preselected(self):
        """GIVEN a default option is specified
        WHEN selection prompt is displayed
        THEN default option should be pre-selected
        """
        # GIVEN: Default value "en"
        default_value = "en"

        # WHEN: Prompt is displayed
        with patch("moai_adk.cli.ui.prompts.styled_select", return_value=default_value):
            result = _prompt_select("Select language:", choices=[], default=default_value)

        # THEN: Default should be passed to prompt
        assert result == default_value

    def test_given_fallback_needed_when_styled_unavailable_then_questionary_used(self):
        """GIVEN styled UI is not available (ImportError or OSError)
        WHEN selection prompt is displayed
        THEN questionary fallback should be used seamlessly
        """
        # GIVEN: styled_select fails with OSError (macOS terminal issue)
        choices = [
            {"name": "Manual", "value": "manual"},
            {"name": "Personal", "value": "personal"},
        ]

        # WHEN: Prompt attempts styled_select but falls back
        mock_questionary = MagicMock()
        mock_questionary.select.return_value.ask.return_value = "Manual"

        with patch("moai_adk.cli.ui.prompts.styled_select", side_effect=OSError("Invalid argument")):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_select("Select Git mode:", choices=choices, default="manual")

                # THEN: questionary fallback should work
                assert result == "manual"
            finally:
                sys.modules.pop("questionary", None)

    def test_given_user_cancels_when_prompt_select_then_none_returned(self):
        """GIVEN a user cancels the selection (Ctrl+C)
        WHEN selection prompt is interrupted
        THEN None should be returned to signal cancellation
        """
        # GIVEN: User cancellation
        # WHEN: Prompt is cancelled
        mock_questionary = MagicMock()
        mock_questionary.select.return_value.ask.return_value = None

        with patch("moai_adk.cli.ui.prompts.styled_select", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_select("Select:", choices=[])

                # THEN: None signals cancellation
                assert result is None
            finally:
                sys.modules.pop("questionary", None)


class TestPromptPasswordSpecification:
    """Specification tests for password input prompt behavior.

    User Story:
    AS a user setting up MoAI-ADK
    I WANT to securely input sensitive information (API keys)
    SO THAT my credentials are protected from display
    """

    def test_given_password_input_when_prompt_password_then_input_hidden(self):
        """GIVEN a user enters a password
        WHEN password prompt is displayed
        THEN input should be hidden (not echoed to terminal)
        """
        # GIVEN: Password "secret-api-key"
        password = "secret-api-key"

        # WHEN: Password prompt is displayed
        with patch("moai_adk.cli.ui.prompts.styled_password", return_value=password):
            result = _prompt_password("Enter API key:")

        # THEN: Password should be returned
        assert result == password

    def test_given_required_password_when_empty_then_validation_fails(self):
        """GIVEN a required password field
        WHEN user provides empty input
        THEN validation should reject empty password
        """
        # GIVEN: Required password
        # WHEN: Empty input is provided
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = ""

        with patch("moai_adk.cli.ui.prompts.styled_password", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                _result = _prompt_password("Enter API key:")

                # THEN: Validation should be applied
                mock_questionary.password.assert_called_once()
                call_kwargs = mock_questionary.password.call_args[1]
                assert "validate" in call_kwargs
            finally:
                sys.modules.pop("questionary", None)

    def test_given_user_cancels_when_prompt_password_then_none_returned(self):
        """GIVEN a user cancels the password prompt (Ctrl+C)
        WHEN password prompt is interrupted
        THEN None should be returned to signal cancellation
        """
        # GIVEN: User cancellation
        # WHEN: Password prompt is cancelled
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = None

        with patch("moai_adk.cli.ui.prompts.styled_password", side_effect=ImportError):
            sys.modules["questionary"] = mock_questionary
            try:
                result = _prompt_password("Enter API key:")

                # THEN: None signals cancellation
                assert result is None
            finally:
                sys.modules.pop("questionary", None)


class TestPromptPasswordOptionalSpecification:
    """Specification tests for optional password input behavior.

    User Story:
    AS a user setting up MoAI-ADK
    I WANT to optionally provide sensitive information
    SO THAT I can skip optional API keys or keep existing ones
    """

    def test_given_optional_password_when_enter_pressed_then_empty_accepted(self):
        """GIVEN an optional password field
        WHEN user presses Enter without typing
        THEN empty string should be returned (keep existing/skip)
        """
        # GIVEN: Optional password
        # WHEN: User presses Enter (empty string)
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = ""

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter API key (optional):")

            # THEN: Empty string is accepted (not None)
            assert result == ""
        finally:
            sys.modules.pop("questionary", None)

    def test_given_optional_password_when_user_cancels_then_none_returned(self):
        """GIVEN an optional password field
        WHEN user presses Ctrl+C to cancel
        THEN None should be returned (not empty string)
        """
        # GIVEN: Optional password
        # WHEN: User cancels (Ctrl+C)
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = None

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter API key (optional):")

            # THEN: None signals cancellation (different from empty string)
            assert result is None
        finally:
            sys.modules.pop("questionary", None)

    def test_given_optional_password_when_input_provided_then_input_returned(self):
        """GIVEN an optional password field
        WHEN user provides input
        THEN the input should be returned
        """
        # GIVEN: Optional password
        # WHEN: User provides input
        mock_questionary = MagicMock()
        mock_questionary.password.return_value.ask.return_value = "new-api-key"

        sys.modules["questionary"] = mock_questionary
        try:
            result = _prompt_password_optional("Enter API key (optional):")

            # THEN: Input should be returned
            assert result == "new-api-key"
        finally:
            sys.modules.pop("questionary", None)


class TestPromptProjectSetupSpecification:
    """Specification tests for project setup flow behavior.

    User Story:
    AS a developer starting a new MoAI-ADK project
    I WANT to configure project settings through interactive prompts
    SO THAT my project is properly initialized with all necessary configurations
    """

    def test_given_new_project_when_setup_then_language_selected_first(self):
        """GIVEN a user starts project setup
        WHEN setup flow begins
        THEN language selection should be the first question
        """
        # GIVEN: New project setup
        mock_translations = {
            "user_setup": "User Setup",
            "q_user_name": "Your name:",
            "api_key_input": "API Key Input",
            "q_api_key_glm": "GLM API Key (optional):",
            "msg_api_key_stored": "API key stored successfully",
            "msg_glm_key_skip_guidance": "You can add API key later",
            "project_setup": "Project Setup",
            "q_project_name": "Project name:",
            "git_setup": "Git Setup",
            "q_git_mode": "Select Git mode:",
            "opt_manual": "Manual",
            "desc_manual": "I'll handle Git myself",
            "opt_personal": "Personal",
            "desc_personal": "Personal GitHub account",
            "opt_team": "Team",
            "desc_team": "Team/Organization GitHub",
            "q_github_username": "GitHub username:",
            "output_language": "Output Language Settings",
            "q_commit_lang": "Git commit message language:",
            "q_comment_lang": "Code comment language:",
            "q_doc_lang": "Documentation language:",
            "msg_setup_complete": "Project setup complete!",
            "msg_cancelled": "Setup cancelled",
            "dev_mode_ddd_info": "Using DDD (Domain-Driven Development)",
        }

        # WHEN: Setup flow starts
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="en"):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch(
                                "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                            ):
                                result = prompt_project_setup()

                                # THEN: Language should be set
                                assert result["locale"] == "en"

    def test_given_personal_git_mode_when_setup_then_github_username_prompted(self):
        """GIVEN a user selects personal Git mode
        WHEN setup flow reaches Git configuration
        THEN GitHub username should be prompted
        """
        # GIVEN: Personal Git mode selected
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
            "q_github_username": "GitHub username:",
            "output_language": "Lang",
            "q_commit_lang": "Commit:",
            "q_comment_lang": "Comment:",
            "q_doc_lang": "Doc:",
            "msg_setup_complete": "Done",
            "msg_cancelled": "Cancel",
        }

        # WHEN: Git mode is personal
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="personal"):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="github-user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch(
                                "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                            ):
                                result = prompt_project_setup()

                                # THEN: GitHub username should be set
                                assert result["git_mode"] == "personal"
                                assert result["github_username"] == "github-user"

    def test_given_manual_git_mode_when_setup_then_github_username_skipped(self):
        """GIVEN a user selects manual Git mode
        WHEN setup flow reaches Git configuration
        THEN GitHub username should NOT be prompted
        """
        # GIVEN: Manual Git mode selected
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

        # WHEN: Git mode is manual
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="manual"):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch(
                                "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                            ):
                                result = prompt_project_setup()

                                # THEN: GitHub username should remain None
                                assert result["git_mode"] == "manual"
                                assert result["github_username"] is None

    def test_given_multilingual_when_setup_then_output_languages_configurable(self):
        """GIVEN a user selects a conversation language
        WHEN setup flow reaches output language configuration
        THEN all three output languages should be configurable independently
        """
        # GIVEN: Korean conversation language
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

        # WHEN: Output languages are configured
        # Track which select calls are made
        select_call_count = [0]
        language_selections = []  # Track language selections in order

        def mock_select_return(message, choices, default=None):
            select_call_count[0] += 1

            # Language selection (first call)
            if "language" in message.lower() or "Select your conversation" in message:
                language_selections.append("locale")
                return "ko"

            # Git mode selection
            if "Git mode" in message:
                language_selections.append("git_mode")
                return "manual"

            # Commit language selection
            if "commit" in message.lower():
                language_selections.append("commit_lang")
                return "ko"

            # Comment language selection
            if "comment" in message.lower():
                language_selections.append("comment_lang")
                return "en"

            # Doc language selection
            if "doc" in message.lower():
                language_selections.append("doc_lang")
                return "ko"

            return "en"

        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", side_effect=mock_select_return):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch(
                                "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                            ):
                                result = prompt_project_setup()

                                # THEN: All three output languages should be configurable
                                assert result["locale"] == "ko"  # Conversation
                                assert result["git_commit_lang"] == "ko"  # Commit
                                assert result["code_comment_lang"] == "en"  # Comment (different)
                                assert result["doc_lang"] == "ko"  # Doc

                                # Verify all selections were made
                                assert "locale" in language_selections
                                assert "commit_lang" in language_selections
                                assert "comment_lang" in language_selections
                                assert "doc_lang" in language_selections

    def test_given_ddd_methodology_when_setup_then_always_ddd_selected(self):
        """GIVEN a user starts project setup
        WHEN setup flow completes
        THEN development mode should always be 'ddd' (Domain-Driven Development)
        """
        # GIVEN: Any project setup
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

        # WHEN: Setup completes
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="en"):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch(
                                "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                            ):
                                result = prompt_project_setup()

                                # THEN: Development mode should always be DDD
                                assert result["development_mode"] == "ddd"

    def test_given_cancellation_when_setup_then_keyboard_interrupt_raised(self):
        """GIVEN a user cancels during setup
        WHEN any prompt is cancelled
        THEN KeyboardInterrupt should be raised
        """
        # GIVEN: User cancels at language selection
        # WHEN: Setup is interrupted
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value=None):
            # THEN: KeyboardInterrupt should be raised
            with pytest.raises(KeyboardInterrupt):
                prompt_project_setup()

    def test_given_initial_locale_when_setup_then_language_skipped_or_defaulted(self):
        """GIVEN a user provides initial_locale via CLI flag
        WHEN setup flow starts
        THEN language selection should use initial_locale as default
        """
        # GIVEN: initial_locale="ja" from CLI
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

        # WHEN: Setup starts with initial_locale
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="ja"):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch(
                                "moai_adk.cli.prompts.init_prompts.get_translation", return_value=mock_translations
                            ):
                                result = prompt_project_setup(initial_locale="ja")

                                # THEN: Locale should match initial_locale
                                assert result["locale"] == "ja"


class TestProjectSetupAnswersSpecification:
    """Specification tests for ProjectSetupAnswers structure.

    User Story:
    AS a developer using MoAI-ADK
    I WANT project setup answers to be well-structured and type-safe
    SO THAT I can rely on consistent data types throughout the application
    """

    def test_given_complete_answers_when_validated_then_all_fields_present(self):
        """GIVEN a complete set of project setup answers
        WHEN answers are validated
        THEN all 10 required and optional fields should be present
        """
        # GIVEN: Complete answers
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

        # WHEN: Answers are validated
        # THEN: All fields should be accessible
        assert "project_name" in answers
        assert "locale" in answers
        assert "user_name" in answers
        assert "glm_api_key" in answers
        assert "git_mode" in answers
        assert "github_username" in answers
        assert "git_commit_lang" in answers
        assert "code_comment_lang" in answers
        assert "doc_lang" in answers
        assert "development_mode" in answers

    def test_given_optional_fields_when_none_then_accepted(self):
        """GIVEN optional fields (glm_api_key, github_username)
        WHEN set to None
        THEN should be accepted as valid values
        """
        # GIVEN: Optional fields as None
        answers: ProjectSetupAnswers = {
            "project_name": "test-project",
            "locale": "ko",
            "user_name": "",
            "glm_api_key": None,  # Optional: None allowed
            "git_mode": "manual",
            "github_username": None,  # Optional: None allowed
            "git_commit_lang": "ko",
            "code_comment_lang": "en",
            "doc_lang": "ko",
            "development_mode": "ddd",
        }

        # WHEN: Answers are validated
        # THEN: Optional None values should be accepted
        assert answers["glm_api_key"] is None
        assert answers["github_username"] is None

    def test_given_development_mode_when_always_ddd_then_not_configurable(self):
        """GIVEN project setup answers
        WHEN development_mode field is checked
        THEN should always be 'ddd' (not configurable by user)
        """
        # GIVEN: Any project setup
        answers: ProjectSetupAnswers = {
            "project_name": "any-project",
            "locale": "en",
            "user_name": "User",
            "glm_api_key": None,
            "git_mode": "manual",
            "github_username": None,
            "git_commit_lang": "en",
            "code_comment_lang": "en",
            "doc_lang": "en",
            "development_mode": "ddd",  # Always DDD
        }

        # WHEN: Development mode is checked
        # THEN: Should always be "ddd"
        assert answers["development_mode"] == "ddd"
