"""Enhanced tests for cli.prompts.init_prompts module - Extended coverage.

This module provides additional test coverage for:
- End-to-end flow with complete prompt sequence
- Multilingual support (ko, ja, zh)
- GLM API key handling with existing key
- Cancellation handling at various stages
- Edge cases and error conditions
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.cli.prompts.init_prompts import (
    prompt_project_setup,
)


# Fixture: Complete mock translation dictionary
@pytest.fixture
def mock_translations():
    """Provide complete mock translation dictionary."""
    return {
        # Headers
        "user_setup": "User Setup",
        "service_selection": "Service Selection",
        "pricing_selection": "Pricing Plan",
        "api_key_input": "API Key Input",
        "project_setup": "Project Setup",
        "git_setup": "Git Setup",
        "output_language": "Output Language",
        "claude_auth_selection": "Claude Authentication",
        # Questions
        "q_language": "Select language:",
        "q_user_name": "Your name:",
        "q_service": "Select service:",
        "q_claude_auth_type": "Select Claude auth type:",
        "q_pricing_claude": "Select Claude pricing:",
        "q_pricing_glm": "Select GLM pricing:",
        "q_api_key_anthropic": "Anthropic API key:",
        "q_api_key_glm": "GLM Key:",
        "q_project_name": "Project name:",
        "q_git_mode": "Git mode:",
        "q_github_username": "GitHub username:",
        "q_commit_lang": "Commit language:",
        "q_comment_lang": "Comment language:",
        "q_doc_lang": "Doc language:",
        # Options - Service
        "opt_claude_subscription": "Claude Subscription",
        "opt_claude_api": "Claude API",
        "opt_glm": "GLM CodePlan",
        "opt_hybrid": "Claude + GLM (Hybrid)",
        # Options - Claude Auth Type
        "opt_claude_sub": "Subscription",
        "opt_claude_api_key": "API Key",
        "desc_claude_sub": "Use Claude Code subscription",
        "desc_claude_api_key": "Enter API key directly",
        # Options - Pricing Claude
        "opt_pro": "Pro ($20/mo)",
        "opt_max5": "Max5 ($100/mo)",
        "opt_max20": "Max20 ($200/mo)",
        # Options - Pricing GLM
        "opt_basic": "Basic",
        "opt_glm_pro": "Pro",
        "opt_enterprise": "Enterprise",
        # Options - Git
        "opt_manual": "manual (local only)",
        "opt_personal": "personal (GitHub personal)",
        "opt_team": "team (GitHub team)",
        # Descriptions
        "desc_claude_subscription": "Claude Code subscriber - No API key needed",
        "desc_claude_api": "Enter API key directly",
        "desc_glm": "Use GLM CodePlan service",
        "desc_hybrid": "Cost-optimized automatic allocation",
        "desc_pro": "Sonnet-focused, basic usage",
        "desc_max5": "Opus partially available",
        "desc_max20": "Opus freely available",
        "desc_basic": "Basic features",
        "desc_glm_pro": "Advanced features",
        "desc_enterprise": "Full features",
        "desc_manual": "Local repository only",
        "desc_personal": "GitHub personal account",
        "desc_team": "GitHub team/organization",
        # Messages
        "msg_api_key_stored": "Key stored",
        "msg_glm_key_found": "GLM API key found",
        "msg_glm_key_keep_prompt": "Press Enter to keep existing key",
        "msg_glm_key_skip_guidance": "Skip guidance",
        "msg_setup_complete": "Complete",
        "msg_cancelled": "Cancelled",
        "msg_current_dir": "(current directory)",
        "msg_skip_same_lang": "Set to same as conversation language",
    }


class TestPromptProjectSetupGLMKeyHandling:
    """Test GLM API key handling in prompt_project_setup."""

    def test_glm_api_key_with_existing_key(self, tmp_path: Path, mock_translations: dict) -> None:
        """Test GLM API key prompt when key already exists."""
        mock_existing_key = "sk-existing-key-1234567890"

        with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
            # Sequence: language, git_mode, commit_lang, comment_lang, doc_lang
            mock_select.side_effect = ["en", "manual", "en", "en", "en"]

            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="test-user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=True):
                        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=mock_existing_key):
                            with patch("moai_adk.cli.prompts.init_prompts.console"):
                                with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
                                    with patch("moai_adk.cli.prompts.init_prompts.get_translation") as mock_t:
                                        mock_t.return_value = mock_translations

                                        result = prompt_project_setup(project_name="test-project")

                                        assert result["glm_api_key"] == mock_existing_key

    def test_glm_api_key_new_key_provided(self, tmp_path: Path, mock_translations: dict) -> None:
        """Test GLM API key prompt when user provides new key."""
        new_key = "sk-new-key-0987654321"

        with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
            # Sequence: language, git_mode, commit_lang, comment_lang, doc_lang
            mock_select.side_effect = ["en", "manual", "en", "en", "en"]

            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="test-user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=new_key):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.core.credentials.save_glm_key_to_env") as mock_save:
                            with patch("moai_adk.cli.prompts.init_prompts.console"):
                                with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
                                    with patch("moai_adk.cli.prompts.init_prompts.get_translation") as mock_t:
                                        mock_t.return_value = mock_translations

                                        result = prompt_project_setup(project_name="test-project")

                                        mock_save.assert_called_once_with(new_key)
                                        assert result["glm_api_key"] == new_key

    def test_glm_api_key_skipped_no_existing(self, tmp_path: Path, mock_translations: dict) -> None:
        """Test GLM API key prompt when user skips with no existing key."""
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
            # Sequence: language, git_mode, commit_lang, comment_lang, doc_lang
            mock_select.side_effect = ["en", "manual", "en", "en", "en"]

            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="test-user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
                                with patch("moai_adk.cli.prompts.init_prompts.get_translation") as mock_t:
                                    mock_t.return_value = mock_translations

                                    result = prompt_project_setup(project_name="test-project")

                                    assert result["glm_api_key"] is None


class TestPromptProjectSetupGitHubUsername:
    """Test GitHub username prompting for personal/team git modes."""

    def test_git_mode_personal_prompts_username(self, tmp_path: Path, mock_translations: dict) -> None:
        """Test that personal git mode prompts for GitHub username."""
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
            # Sequence: language, personal, commit_lang, comment_lang, doc_lang
            mock_select.side_effect = ["en", "personal", "en", "en", "en"]

            # Create a text prompt mock that returns values sequentially
            text_values = ["test-user", "test-project", "mygithub"]
            text_mock = MagicMock(side_effect=text_values)

            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", text_mock):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
                                with patch("moai_adk.cli.prompts.init_prompts.get_translation") as mock_t:
                                    mock_t.return_value = mock_translations

                                    result = prompt_project_setup()

                                    assert result["git_mode"] == "personal"
                                    assert result["github_username"] == "mygithub"


class TestPromptProjectSetupCancellation:
    """Test cancellation handling at various stages."""

    def test_cancellation_at_language_selection(self, tmp_path: Path) -> None:
        """Test cancellation at language selection."""
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value=None):
            with patch("moai_adk.cli.prompts.init_prompts.console"):
                with pytest.raises(KeyboardInterrupt):
                    prompt_project_setup()

    def test_cancellation_at_user_name(self, tmp_path: Path) -> None:
        """Test cancellation at user name prompt."""
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="en"):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value=None):
                with patch("moai_adk.cli.prompts.init_prompts.console"):
                    with pytest.raises(KeyboardInterrupt):
                        prompt_project_setup()

    def test_cancellation_at_glm_key(self, tmp_path: Path) -> None:
        """Test cancellation at GLM API key prompt."""
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select", return_value="en"):
            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", return_value="test-user"):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=None):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with pytest.raises(KeyboardInterrupt):
                                prompt_project_setup()


class TestPromptProjectSetupEdgeCases:
    """Test edge cases and special scenarios."""

    def test_empty_user_name_allowed(self, tmp_path: Path, mock_translations: dict) -> None:
        """Test that empty user name is allowed (optional field)."""
        with patch("moai_adk.cli.prompts.init_prompts._prompt_select") as mock_select:
            # Sequence: language, manual, commit_lang, comment_lang, doc_lang
            mock_select.side_effect = ["en", "manual", "en", "en", "en"]

            # Create a text prompt mock that returns values sequentially
            text_values = ["", "test-project"]
            text_mock = MagicMock(side_effect=text_values)

            with patch("moai_adk.cli.prompts.init_prompts._prompt_text", text_mock):
                with patch("moai_adk.cli.prompts.init_prompts._prompt_password_optional", return_value=""):
                    with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
                        with patch("moai_adk.cli.prompts.init_prompts.console"):
                            with patch("moai_adk.cli.prompts.init_prompts._prompt_confirm", return_value=True):
                                with patch("moai_adk.cli.prompts.init_prompts.get_translation") as mock_t:
                                    mock_t.return_value = mock_translations

                                    result = prompt_project_setup()

                                    assert result["user_name"] == ""
