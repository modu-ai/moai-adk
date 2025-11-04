"""Test migration functions for v0.17.0 features (report generation and SPEC git workflow).

Tests for:
- migrate_config_schema_v0_17_0
- get_report_generation_config
- get_spec_git_workflow
"""

import pytest
from moai_adk.core.config.migration import (
    migrate_config_schema_v0_17_0,
    get_report_generation_config,
    get_spec_git_workflow,
)


class TestMigrateConfigSchemaV0170:
    """Test migrate_config_schema_v0_17_0 function."""

    def test_add_report_generation_section_to_empty_config(self):
        """Should add report_generation section with defaults to empty config."""
        config = {}
        result = migrate_config_schema_v0_17_0(config)

        assert "report_generation" in result
        assert result["report_generation"]["enabled"] is True
        assert result["report_generation"]["auto_create"] is False
        assert result["report_generation"]["warn_user"] is True
        assert result["report_generation"]["user_choice"] == "Minimal"
        assert ".moai/docs/" in result["report_generation"]["allowed_locations"]

    def test_preserve_existing_report_generation_config(self):
        """Should preserve existing report_generation config."""
        config = {
            "report_generation": {
                "enabled": False,
                "auto_create": True,
                "warn_user": False,
                "user_choice": "Enable",
            }
        }
        result = migrate_config_schema_v0_17_0(config)

        assert result["report_generation"]["enabled"] is False
        assert result["report_generation"]["auto_create"] is True
        assert result["report_generation"]["warn_user"] is False
        assert result["report_generation"]["user_choice"] == "Enable"

    def test_add_github_section_if_missing(self):
        """Should create github section if missing."""
        config = {}
        result = migrate_config_schema_v0_17_0(config)

        assert "github" in result
        assert isinstance(result["github"], dict)

    def test_add_auto_delete_branches_settings(self):
        """Should add auto_delete_branches settings to github section."""
        config = {"github": {}}
        result = migrate_config_schema_v0_17_0(config)

        github = result["github"]
        assert "auto_delete_branches" in github
        assert github["auto_delete_branches"] is None
        assert github["auto_delete_branches_checked"] is False
        assert github["auto_delete_branches_rationale"] == "Not configured"

    def test_add_spec_git_workflow_settings(self):
        """Should add spec_git_workflow settings to github section."""
        config = {"github": {}}
        result = migrate_config_schema_v0_17_0(config)

        github = result["github"]
        assert "spec_git_workflow" in github
        assert github["spec_git_workflow"] == "per_spec"
        assert github["spec_git_workflow_configured"] is False
        assert "Ask per SPEC" in github["spec_git_workflow_rationale"]

    def test_preserve_existing_github_settings(self):
        """Should preserve existing github settings."""
        config = {
            "github": {
                "auto_delete_branches": True,
                "spec_git_workflow": "feature_branch",
            }
        }
        result = migrate_config_schema_v0_17_0(config)

        github = result["github"]
        assert github["auto_delete_branches"] is True
        assert github["spec_git_workflow"] == "feature_branch"

    def test_add_notes_for_new_fields(self):
        """Should add documentation notes for new fields."""
        config = {"github": {}}
        result = migrate_config_schema_v0_17_0(config)

        github = result["github"]
        assert "notes_new_fields" in github
        assert "auto_delete_branches" in github["notes_new_fields"]
        assert "spec_git_workflow" in github["notes_new_fields"]

    def test_backward_compatibility_with_v0_16_0_config(self):
        """Should handle v0.16.0 config format correctly."""
        # Simulate a v0.16.0 config that doesn't have report_generation or enhanced github fields
        config = {
            "project": {"name": "TestProject"},
            "github": {"strategy": "github-pr"},
        }
        result = migrate_config_schema_v0_17_0(config)

        # Should add new sections without removing existing ones
        assert result["project"]["name"] == "TestProject"
        assert result["github"]["strategy"] == "github-pr"
        assert "report_generation" in result
        assert "spec_git_workflow" in result["github"]


class TestGetReportGenerationConfig:
    """Test get_report_generation_config function."""

    def test_returns_default_config_when_missing(self):
        """Should return default config when report_generation is missing."""
        config = {}
        result = get_report_generation_config(config)

        assert result["enabled"] is True
        assert result["auto_create"] is False
        assert result["warn_user"] is True
        assert result["user_choice"] == "Minimal"

    def test_returns_default_locations(self):
        """Should return default allowed_locations when missing."""
        config = {}
        result = get_report_generation_config(config)

        assert ".moai/docs/" in result["allowed_locations"]
        assert ".moai/reports/" in result["allowed_locations"]
        assert ".moai/analysis/" in result["allowed_locations"]
        assert ".moai/specs/SPEC-*/" in result["allowed_locations"]

    def test_merges_with_existing_config(self):
        """Should merge with existing config, preserving values."""
        config = {
            "report_generation": {
                "enabled": False,
                "user_choice": "Enable",
            }
        }
        result = get_report_generation_config(config)

        assert result["enabled"] is False
        assert result["user_choice"] == "Enable"
        # Should preserve defaults for missing keys
        assert result["auto_create"] is False
        assert result["warn_user"] is True

    def test_overwrites_defaults_with_user_values(self):
        """Should overwrite defaults with user-configured values."""
        config = {
            "report_generation": {
                "enabled": True,
                "auto_create": True,
                "warn_user": False,
                "user_choice": "Enable",
                "allowed_locations": [".moai/custom/"],
            }
        }
        result = get_report_generation_config(config)

        assert result["enabled"] is True
        assert result["auto_create"] is True
        assert result["warn_user"] is False
        assert result["user_choice"] == "Enable"
        assert result["allowed_locations"] == [".moai/custom/"]

    def test_handles_non_dict_report_generation(self):
        """Should return defaults if report_generation is not a dict."""
        config = {"report_generation": "invalid"}
        result = get_report_generation_config(config)

        # Should return defaults, not crash
        assert result["enabled"] is True
        assert result["auto_create"] is False

    def test_empty_report_generation_dict(self):
        """Should apply defaults to empty report_generation dict."""
        config = {"report_generation": {}}
        result = get_report_generation_config(config)

        assert result["enabled"] is True
        assert result["auto_create"] is False
        assert len(result["allowed_locations"]) == 4


class TestGetSpecGitWorkflow:
    """Test get_spec_git_workflow function."""

    def test_returns_configured_workflow(self):
        """Should return configured spec_git_workflow value."""
        config = {
            "github": {
                "spec_git_workflow": "feature_branch",
            }
        }
        result = get_spec_git_workflow(config)

        assert result == "feature_branch"

    def test_returns_per_spec_default(self):
        """Should return 'per_spec' as default when not configured."""
        config = {}
        result = get_spec_git_workflow(config)

        assert result == "per_spec"

    def test_supports_all_workflow_options(self):
        """Should support all valid workflow options."""
        valid_workflows = ["per_spec", "feature_branch", "develop_direct"]

        for workflow in valid_workflows:
            config = {"github": {"spec_git_workflow": workflow}}
            result = get_spec_git_workflow(config)
            assert result == workflow

    def test_returns_default_for_invalid_workflow(self):
        """Should return default for invalid workflow values."""
        config = {
            "github": {
                "spec_git_workflow": "invalid_workflow",
            }
        }
        result = get_spec_git_workflow(config)

        # Should return safe default instead of invalid value
        assert result == "per_spec"

    def test_handles_missing_github_section(self):
        """Should handle missing github section gracefully."""
        config = {}
        result = get_spec_git_workflow(config)

        assert result == "per_spec"

    def test_handles_non_dict_github_section(self):
        """Should handle non-dict github section gracefully."""
        config = {"github": "invalid"}
        result = get_spec_git_workflow(config)

        assert result == "per_spec"

    def test_case_sensitive_workflow_matching(self):
        """Should only match exact workflow names (case-sensitive)."""
        config = {
            "github": {
                "spec_git_workflow": "Feature_Branch",  # Wrong case
            }
        }
        result = get_spec_git_workflow(config)

        # Should return default for case mismatch
        assert result == "per_spec"


class TestMigrationIntegration:
    """Integration tests for migration functions working together."""

    def test_full_migration_workflow(self):
        """Should handle complete v0.16.0 to v0.17.0 migration."""
        # Simulate v0.16.0 config
        old_config = {
            "project": {"name": "TestProject"},
            "github": {"strategy": "github-pr"},
        }

        # Run migration
        migrated_config = migrate_config_schema_v0_17_0(old_config)

        # Verify report generation config
        report_gen = get_report_generation_config(migrated_config)
        assert report_gen["enabled"] is True
        assert report_gen["user_choice"] == "Minimal"

        # Verify workflow config
        workflow = get_spec_git_workflow(migrated_config)
        assert workflow == "per_spec"

        # Verify original settings preserved
        assert migrated_config["project"]["name"] == "TestProject"
        assert migrated_config["github"]["strategy"] == "github-pr"

    def test_user_configured_v0_17_0_config(self):
        """Should preserve user-configured values during access."""
        # User has configured everything
        config = {
            "project": {"name": "MyProject"},
            "report_generation": {
                "enabled": True,
                "auto_create": True,
                "user_choice": "Enable",
            },
            "github": {
                "spec_git_workflow": "feature_branch",
            },
        }

        # Verify all values are preserved
        report_gen = get_report_generation_config(config)
        assert report_gen["enabled"] is True
        assert report_gen["auto_create"] is True

        workflow = get_spec_git_workflow(config)
        assert workflow == "feature_branch"
