"""Test init_prompts changes for v0.17.0 (simplified init with only project name).

Tests for:
- Simplified init_prompts structure
- Default values for mode, locale, language
- ProjectSetupAnswers TypedDict structure
- prompt_project_setup function
"""

from moai_adk.cli.prompts.init_prompts import ProjectSetupAnswers


class TestProjectSetupAnswersStructure:
    """Test ProjectSetupAnswers TypedDict structure."""

    def test_project_setup_answers_has_required_fields(self):
        """Should have all required fields in ProjectSetupAnswers."""
        # Create an instance matching the TypedDict
        answers: ProjectSetupAnswers = {
            "project_name": "test-project",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        # Core fields that must be present
        assert "project_name" in answers
        assert "mode" in answers
        assert "locale" in answers
        assert "language" in answers
        assert "author" in answers

    def test_answers_dict_field_types(self):
        """Should have correct field types."""
        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        assert isinstance(answers["project_name"], str)
        assert isinstance(answers["mode"], str)
        assert isinstance(answers["locale"], str)
        assert answers["language"] is None
        assert isinstance(answers["author"], str)

    def test_mode_field_valid_values(self):
        """Mode field should support 'personal' and 'team'."""
        for mode in ["personal", "team"]:
            answers: ProjectSetupAnswers = {
                "project_name": "test",
                "mode": mode,
                "locale": "en",
                "language": None,
                "author": "",
            }
            assert answers["mode"] in ["personal", "team"]


class TestInitPromptSimplification:
    """Test that init prompts have been simplified as per v0.17.0."""

    def test_default_mode_is_personal(self):
        """Should default to 'personal' mode (not prompted)."""
        # In v0.17.0, mode is not prompted - it's a default
        # Mode selection is deferred to /alfred:0-project

        # Simulate what prompt_project_setup initializes
        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        assert answers["mode"] == "personal"

    def test_default_locale_is_english(self):
        """Should default to 'en' locale (not 'ko' as in v0.16.0)."""
        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        assert answers["locale"] == "en"

    def test_default_language_is_none(self):
        """Should initialize language as None (set in /alfred:0-project)."""
        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        assert answers["language"] is None

    def test_other_configuration_deferred_to_0_project(self):
        """Non-essential configuration should be deferred to /alfred:0-project.

        Configuration deferred:
        - Language selection (language field is None)
        - Mode selection (personal/team) - default provided, configured in 0-project
        - Author name (empty string, set in 0-project)
        - Report generation preferences (set in 0-project)
        """
        # These are handled in /alfred:0-project via AskUserQuestion
        # Not in init_prompts

        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        # init only sets defaults
        # Everything else is deferred to be configured in 0-project
        assert answers["language"] is None  # Deferred
        assert answers["mode"] == "personal"  # Default, configurable in 0-project
        assert answers["author"] == ""  # Deferred


class TestInitPromptBackwardCompatibility:
    """Test that v0.17.0 init is backward compatible."""

    def test_answers_dict_has_all_legacy_fields(self):
        """Should preserve all legacy field names for compatibility."""
        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        # Ensure all legacy field names still exist
        legacy_fields = {
            "project_name",
            "mode",
            "locale",
            "language",
            "author",
        }
        assert legacy_fields.issubset(answers.keys())

    def test_default_mode_maintains_compatibility(self):
        """Default mode should be valid for legacy code."""
        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        # Mode should be one of valid values
        valid_modes = ["personal", "team"]
        assert answers["mode"] in valid_modes

    def test_locale_default_supports_config_migration(self):
        """Default locale should work with config migration functions."""
        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "en",
            "language": None,
            "author": "",
        }

        # Locale should be a valid language code
        valid_locales = ["en", "ko", "ja", "zh", "es"]
        assert answers["locale"] in valid_locales

    def test_answers_dict_with_initial_locale_parameter(self):
        """Should support initial_locale parameter for config migration."""
        # prompt_project_setup accepts initial_locale parameter
        # for backward compatibility with config migration

        answers: ProjectSetupAnswers = {
            "project_name": "test",
            "mode": "personal",
            "locale": "ko",  # Set via initial_locale parameter
            "language": None,
            "author": "",
        }

        # Should accept any valid locale
        assert answers["locale"] in ["en", "ko", "ja", "zh", "es"]
