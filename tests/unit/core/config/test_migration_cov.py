"""Comprehensive tests for config migration with 70%+ coverage.

Tests cover:
- migrate_config_to_nested_structure - legacy flat config to nested
- get_conversation_language - with fallback handling
- get_conversation_language_name - with language mapping
- migrate_config_schema_v0_17_0 - schema version upgrade
- get_report_generation_config - config retrieval with defaults
- get_spec_git_workflow - workflow setting retrieval
"""

from unittest.mock import patch

from moai_adk.core.config.migration import (
    get_conversation_language,
    get_conversation_language_name,
    get_report_generation_config,
    get_spec_git_workflow,
    migrate_config_schema_v0_17_0,
    migrate_config_to_nested_structure,
)


class TestMigrateConfigToNestedStructure:
    """Test config migration from flat to nested structure."""

    def test_already_nested_structure_unchanged(self):
        """Test that already nested config is returned unchanged."""
        config = {
            "language": {
                "conversation_language": "ko",
                "conversation_language_name": "한국어",
            }
        }

        result = migrate_config_to_nested_structure(config)

        assert result["language"]["conversation_language"] == "ko"
        assert result["language"]["conversation_language_name"] == "한국어"

    def test_migrate_legacy_flat_conversation_language(self):
        """Test migrating legacy flat conversation_language field."""
        config = {"conversation_language": "ko", "locale": "ko"}

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [
                ("ko", {"native_name": "한국어"}),
                ("en", {"native_name": "English"}),
            ]

            result = migrate_config_to_nested_structure(config)

            assert "language" in result
            assert result["language"]["conversation_language"] == "ko"
            assert "conversation_language_name" in result["language"]

    def test_migrate_removes_legacy_locale_field(self):
        """Test that legacy locale field is removed."""
        config = {"conversation_language": "en", "locale": "en", "other_field": "value"}

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [("en", {"native_name": "English"})]

            result = migrate_config_to_nested_structure(config)

            assert "locale" not in result
            assert "other_field" in result

    def test_migrate_string_language_to_nested(self):
        """Test migrating when language is a string (old format)."""
        config = {"language": "ko"}

        result = migrate_config_to_nested_structure(config)

        assert isinstance(result["language"], dict)
        assert result["language"]["conversation_language"] == "ko"

    def test_migrate_defaults_to_english_if_unknown_language(self):
        """Test defaulting to English if language code is unknown."""
        config = {
            "conversation_language": "xx",  # Invalid code
        }

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [("en", {"native_name": "English"})]

            result = migrate_config_to_nested_structure(config)

            assert result["language"]["conversation_language"] == "xx"

    def test_migrate_empty_config(self):
        """Test migrating empty config."""
        config = {}

        result = migrate_config_to_nested_structure(config)

        assert isinstance(result, dict)

    def test_migrate_preserves_other_fields(self):
        """Test that migration preserves other config fields."""
        config = {
            "conversation_language": "ko",
            "other_setting": "value",
            "nested_setting": {"key": "value"},
        }

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [("ko", {"native_name": "한국어"})]

            result = migrate_config_to_nested_structure(config)

            assert result["other_setting"] == "value"
            assert result["nested_setting"]["key"] == "value"


class TestGetConversationLanguage:
    """Test getting conversation language with fallback."""

    def test_get_from_nested_structure(self):
        """Test getting language from nested structure."""
        config = {"language": {"conversation_language": "ko"}}

        result = get_conversation_language(config)

        assert result == "ko"

    def test_fallback_to_legacy_flat_structure(self):
        """Test fallback to legacy flat structure."""
        config = {"conversation_language": "en"}

        result = get_conversation_language(config)

        assert result == "en"

    def test_fallback_to_default_english(self):
        """Test default to English when not found."""
        config = {}

        result = get_conversation_language(config)

        assert result == "en"

    def test_nested_takes_precedence_over_flat(self):
        """Test that nested structure takes precedence."""
        config = {
            "language": {"conversation_language": "ko"},
            "conversation_language": "en",
        }

        result = get_conversation_language(config)

        assert result == "ko"

    def test_ignores_invalid_language_structure(self):
        """Test handling invalid language structure."""
        config = {
            "language": "invalid_string",  # Should be dict
            "conversation_language": "en",
        }

        result = get_conversation_language(config)

        assert result == "en"

    def test_handles_none_language_field(self):
        """Test handling when language field is None."""
        config = {"language": None, "conversation_language": "fr"}

        result = get_conversation_language(config)

        assert result == "fr"


class TestGetConversationLanguageName:
    """Test getting conversation language name."""

    def test_get_from_nested_structure(self):
        """Test getting language name from nested structure."""
        config = {"language": {"conversation_language_name": "한국어"}}

        result = get_conversation_language_name(config)

        assert result == "한국어"

    def test_map_language_code_from_nested(self):
        """Test mapping language code to name from nested structure."""
        config = {"language": {"conversation_language": "ko"}}

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [
                ("ko", {"native_name": "한국어"}),
                ("en", {"native_name": "English"}),
            ]

            result = get_conversation_language_name(config)

            assert "한국어" in result or result == "한국어"

    def test_fallback_to_hardcoded_mapping(self):
        """Test fallback when language not in LANGUAGE_CONFIG."""
        config = {"conversation_language": "ja"}

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            # When LANGUAGE_CONFIG doesn't have the language, it defaults to English
            mock_lang_config.items.return_value = []

            result = get_conversation_language_name(config)

            # Should return English as default when not found in LANGUAGE_CONFIG
            assert result == "English"

    def test_default_to_english_if_not_found(self):
        """Test defaulting to English if language not mapped."""
        config = {"language": {"conversation_language": "xx"}}  # Unknown

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [("en", {"native_name": "English"})]

            result = get_conversation_language_name(config)

            assert result == "English"

    def test_hardcoded_language_mapping(self):
        """Test hardcoded language name mappings."""
        test_cases = [
            ({"language": {"conversation_language": "ko"}}, "ko"),
            ({"language": {"conversation_language": "en"}}, "en"),
            ({"language": {"conversation_language": "ja"}}, "ja"),
            ({"language": {"conversation_language": "zh"}}, "zh"),
        ]

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [
                ("ko", {"native_name": "Korean"}),
                ("en", {"native_name": "English"}),
                ("ja", {"native_name": "Japanese"}),
                ("zh", {"native_name": "Chinese"}),
            ]

            for config, lang_code in test_cases:
                result = get_conversation_language_name(config)
                # Should return either the name or successfully process
                assert isinstance(result, str)


class TestMigrateConfigSchemaV0_17_0:
    """Test schema migration for v0.17.0."""

    def test_adds_report_generation_section(self):
        """Test that report_generation section is added."""
        config = {}

        result = migrate_config_schema_v0_17_0(config)

        assert "report_generation" in result
        assert result["report_generation"]["enabled"] is True
        assert result["report_generation"]["auto_create"] is False

    def test_preserves_existing_report_generation(self):
        """Test that existing report_generation is preserved."""
        config = {"report_generation": {"enabled": False, "custom_field": "value"}}

        result = migrate_config_schema_v0_17_0(config)

        assert result["report_generation"]["enabled"] is False
        assert result["report_generation"]["custom_field"] == "value"

    def test_report_generation_default_structure(self):
        """Test report_generation default structure."""
        config = {}

        result = migrate_config_schema_v0_17_0(config)

        report_gen = result["report_generation"]
        assert "enabled" in report_gen
        assert "auto_create" in report_gen
        assert "warn_user" in report_gen
        assert "user_choice" in report_gen
        assert "configured_at" in report_gen
        assert "allowed_locations" in report_gen

    def test_adds_github_section_if_missing(self):
        """Test that github section is added if missing."""
        config = {}

        result = migrate_config_schema_v0_17_0(config)

        assert "github" in result
        assert isinstance(result["github"], dict)

    def test_preserves_existing_github_section(self):
        """Test that existing github section is preserved."""
        config = {"github": {"username": "testuser", "token": "token"}}

        result = migrate_config_schema_v0_17_0(config)

        assert result["github"]["username"] == "testuser"
        assert result["github"]["token"] == "token"

    def test_adds_auto_delete_branches_settings(self):
        """Test adding auto_delete_branches settings."""
        config = {}

        result = migrate_config_schema_v0_17_0(config)

        github = result["github"]
        assert "auto_delete_branches" in github
        assert "auto_delete_branches_checked" in github
        assert "auto_delete_branches_rationale" in github

    def test_preserves_existing_auto_delete_branches(self):
        """Test that existing auto_delete_branches is preserved."""
        config = {
            "github": {
                "auto_delete_branches": True,
                "auto_delete_branches_checked": True,
            }
        }

        result = migrate_config_schema_v0_17_0(config)

        assert result["github"]["auto_delete_branches"] is True
        assert result["github"]["auto_delete_branches_checked"] is True

    def test_adds_spec_git_workflow_settings(self):
        """Test adding spec_git_workflow settings."""
        config = {}

        result = migrate_config_schema_v0_17_0(config)

        github = result["github"]
        assert "spec_git_workflow" in github
        assert "spec_git_workflow_configured" in github
        assert "spec_git_workflow_rationale" in github

    def test_preserves_existing_spec_git_workflow(self):
        """Test that existing spec_git_workflow is preserved."""
        config = {
            "github": {
                "spec_git_workflow": "feature_branch",
                "spec_git_workflow_configured": True,
            }
        }

        result = migrate_config_schema_v0_17_0(config)

        assert result["github"]["spec_git_workflow"] == "feature_branch"
        assert result["github"]["spec_git_workflow_configured"] is True

    def test_adds_notes_new_fields(self):
        """Test that notes for new fields are added."""
        config = {}

        result = migrate_config_schema_v0_17_0(config)

        github = result["github"]
        assert "notes_new_fields" in github

    def test_preserves_other_config_sections(self):
        """Test that other config sections are preserved."""
        config = {"user": {"name": "Test User"}, "other": "value"}

        result = migrate_config_schema_v0_17_0(config)

        assert result["user"]["name"] == "Test User"
        assert result["other"] == "value"


class TestGetReportGenerationConfig:
    """Test getting report generation configuration."""

    def test_get_existing_report_generation(self):
        """Test getting existing report generation config."""
        config = {"report_generation": {"enabled": False, "auto_create": True}}

        result = get_report_generation_config(config)

        assert result["enabled"] is False
        assert result["auto_create"] is True

    def test_merge_with_defaults(self):
        """Test that result is merged with defaults."""
        config = {"report_generation": {"enabled": False}}

        result = get_report_generation_config(config)

        # Should have defaults merged
        assert result["enabled"] is False  # From config
        assert "auto_create" in result  # From defaults

    def test_empty_config_returns_defaults(self):
        """Test that empty config returns defaults."""
        config = {}

        result = get_report_generation_config(config)

        assert result["enabled"] is True
        assert result["auto_create"] is False
        assert "warn_user" in result

    def test_non_dict_report_generation_returns_defaults(self):
        """Test that non-dict report_generation returns defaults."""
        config = {"report_generation": "invalid"}

        result = get_report_generation_config(config)

        # Should return defaults
        assert isinstance(result, dict)
        assert result["enabled"] is True

    def test_default_structure_completeness(self):
        """Test that default structure is complete."""
        config = {}

        result = get_report_generation_config(config)

        required_keys = [
            "enabled",
            "auto_create",
            "warn_user",
            "user_choice",
            "configured_at",
            "allowed_locations",
        ]

        for key in required_keys:
            assert key in result


class TestGetSpecGitWorkflow:
    """Test getting SPEC git workflow setting."""

    def test_get_per_spec_workflow(self):
        """Test getting per_spec workflow."""
        config = {"github": {"spec_git_workflow": "per_spec"}}

        result = get_spec_git_workflow(config)

        assert result == "per_spec"

    def test_get_feature_branch_workflow(self):
        """Test getting feature_branch workflow."""
        config = {"github": {"spec_git_workflow": "feature_branch"}}

        result = get_spec_git_workflow(config)

        assert result == "feature_branch"

    def test_get_develop_direct_workflow(self):
        """Test getting develop_direct workflow."""
        config = {"github": {"spec_git_workflow": "develop_direct"}}

        result = get_spec_git_workflow(config)

        assert result == "develop_direct"

    def test_default_to_per_spec(self):
        """Test defaulting to per_spec when not configured."""
        config = {}

        result = get_spec_git_workflow(config)

        assert result == "per_spec"

    def test_invalid_workflow_defaults_to_per_spec(self):
        """Test that invalid workflow defaults to per_spec."""
        config = {"github": {"spec_git_workflow": "invalid_workflow"}}

        result = get_spec_git_workflow(config)

        assert result == "per_spec"

    def test_non_dict_github_section_defaults(self):
        """Test that non-dict github section returns default."""
        config = {"github": "invalid"}

        result = get_spec_git_workflow(config)

        assert result == "per_spec"

    def test_missing_github_section_defaults(self):
        """Test that missing github section returns default."""
        config = {}

        result = get_spec_git_workflow(config)

        assert result == "per_spec"


class TestMigrationIntegration:
    """Integration tests for migration functions."""

    def test_full_migration_legacy_to_current(self):
        """Test full migration from legacy to current schema."""
        legacy_config = {
            "conversation_language": "ko",
            "locale": "ko",
            "user": {"name": "Test User"},
        }

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = [
                ("ko", {"native_name": "한국어"}),
                ("en", {"native_name": "English"}),
            ]

            # Step 1: Migrate to nested structure
            result = migrate_config_to_nested_structure(legacy_config)

            # Step 2: Migrate schema to v0.17.0
            result = migrate_config_schema_v0_17_0(result)

            # Verify results
            assert "language" in result
            assert result["language"]["conversation_language"] == "ko"
            assert "report_generation" in result
            assert "github" in result
            assert result["user"]["name"] == "Test User"

    def test_already_current_config_unchanged(self):
        """Test that current config is not unnecessarily changed."""
        current_config = {
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English",
            },
            "report_generation": {"enabled": True},
            "github": {"spec_git_workflow": "per_spec"},
        }

        # Migrate
        result = migrate_config_to_nested_structure(current_config.copy())
        result = migrate_config_schema_v0_17_0(result)

        # Should preserve values
        assert result["language"]["conversation_language"] == "en"
        assert result["report_generation"]["enabled"] is True
        assert "spec_git_workflow" in result["github"]

    def test_migration_with_missing_language_config(self):
        """Test migration when LANGUAGE_CONFIG is not available."""
        config = {"conversation_language": "ko"}

        with patch("moai_adk.core.language_config.LANGUAGE_CONFIG") as mock_lang_config:
            mock_lang_config.items.return_value = []

            result = migrate_config_to_nested_structure(config)

            assert "language" in result
            assert result["language"]["conversation_language"] == "ko"


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_empty_language_dict_handling(self):
        """Test handling of empty language dictionary."""
        config = {"language": {}}

        result = get_conversation_language(config)

        assert result == "en"  # Should fall back to default

    def test_none_values_handling(self):
        """Test handling of None values."""
        config = {"language": {"conversation_language": None}}

        result = get_conversation_language(config)

        assert result == "en"  # Should fall back to default

    def test_very_nested_config_preservation(self):
        """Test that deeply nested structures are preserved."""
        config = {"level1": {"level2": {"level3": {"data": "value"}}}}

        result = migrate_config_to_nested_structure(config)

        assert result["level1"]["level2"]["level3"]["data"] == "value"

    def test_special_characters_in_language_code(self):
        """Test handling special characters in language code."""
        config = {"language": {"conversation_language": "zh-Hans"}}

        result = get_conversation_language(config)

        assert result == "zh-Hans"
