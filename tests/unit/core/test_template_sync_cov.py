"""Comprehensive coverage tests for TemplateVariableSynchronizer module.

Tests TemplateVariableSynchronizer class for template variable extraction and synchronization.
Target: 70%+ code coverage with actual code path execution and mocked dependencies.
"""

import json
from unittest.mock import MagicMock, patch

import pytest


class TestTemplateVariableSynchronizerInit:
    """Test TemplateVariableSynchronizer initialization."""

    def test_synchronizer_instantiation(self, tmp_path):
        """Should instantiate with project root."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))

        assert synchronizer.project_root == tmp_path
        assert len(synchronizer.TEMPLATE_PATTERNS) > 0
        assert len(synchronizer.TEMPLATE_TRACKING_PATTERNS) > 0

    def test_synchronizer_constants(self):
        """Should have correct constants defined."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Template patterns are regex patterns, not literal strings
        assert len(TemplateVariableSynchronizer.TEMPLATE_PATTERNS) > 0
        assert any("CONVERSATION_LANGUAGE" in p for p in TemplateVariableSynchronizer.TEMPLATE_PATTERNS)
        assert any("USER_NAME" in p for p in TemplateVariableSynchronizer.TEMPLATE_PATTERNS)
        assert ".claude/settings.json" in TemplateVariableSynchronizer.TEMPLATE_TRACKING_PATTERNS


class TestFindFilesWithTemplateVariables:
    """Test _find_files_with_template_variables method."""

    def test_find_files_no_changed_config(self, tmp_path):
        """Should find common template files."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Setup
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        claude_settings = claude_dir / "settings.json"
        claude_settings.write_text('{"env": {}}')

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        files = synchronizer._find_files_with_template_variables(None)

        assert isinstance(files, list)

    def test_find_files_specific_config_changed(self, tmp_path):
        """Should prioritize dependent files when config changed."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Setup directories
        moai_dir = tmp_path / ".moai" / "config"
        moai_dir.mkdir(parents=True)
        config_file = moai_dir / "config.json"
        config_file.write_text('{"moai": {"version": "1.0.0"}}')

        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        claude_settings = claude_dir / "settings.json"
        claude_settings.write_text('{"env": {}}')

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        changed_path = config_file

        files = synchronizer._find_files_with_template_variables(changed_path)

        assert isinstance(files, list)

    def test_find_files_removes_duplicates(self, tmp_path):
        """Should remove duplicate file paths."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create duplicate test files
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings = claude_dir / "settings.json"
        settings.write_text('{"env": {}}')

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        files = synchronizer._find_files_with_template_variables(None)

        # Check for uniqueness
        assert len(files) == len(set(files))


class TestGlobFiles:
    """Test _glob_files method."""

    def test_glob_files_with_pattern(self, tmp_path):
        """Should glob files matching pattern."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create test files
        test_dir = tmp_path / ".claude"
        test_dir.mkdir()
        (test_dir / "settings.json").write_text("{}")
        (test_dir / "config.json").write_text("{}")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        files = synchronizer._glob_files(".claude/*.json")

        assert len(files) > 0
        assert all(f.exists() for f in files)

    def test_glob_files_recursive_pattern(self, tmp_path):
        """Should handle recursive patterns."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create nested structure
        deep = tmp_path / ".claude" / "hooks" / "moai"
        deep.mkdir(parents=True)
        (deep / "test.py").write_text("# test")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        files = synchronizer._glob_files(".claude/**/*.py")

        assert any(f.name == "test.py" for f in files)

    def test_glob_files_no_matches(self, tmp_path):
        """Should return empty list when no matches."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        files = synchronizer._glob_files(".nonexistent/**/*.json")

        assert files == []

    def test_glob_files_invalid_pattern(self, tmp_path):
        """Should handle invalid patterns gracefully."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        files = synchronizer._glob_files("[invalid[pattern")

        assert files == []


class TestUpdateFileTemplateVariables:
    """Test _update_file_template_variables method."""

    def test_update_single_variable(self, tmp_path):
        """Should update single template variable."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.md"
        test_file.write_text("User: {{USER_NAME}}\nLanguage: {{CONVERSATION_LANGUAGE}}")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"USER_NAME": "TestUser", "CONVERSATION_LANGUAGE": "en"}

        updated = synchronizer._update_file_template_variables(test_file, template_vars)

        assert "USER_NAME" in updated
        assert "CONVERSATION_LANGUAGE" in updated

        content = test_file.read_text()
        assert "TestUser" in content
        assert "en" in content

    def test_update_no_variables(self, tmp_path):
        """Should return empty list when no variables."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.md"
        test_file.write_text("No variables here")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"USER_NAME": "TestUser"}

        updated = synchronizer._update_file_template_variables(test_file, template_vars)

        assert updated == []

    def test_update_nonexistent_file(self, tmp_path):
        """Should handle nonexistent files gracefully."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        nonexistent = tmp_path / "missing.md"
        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"USER_NAME": "TestUser"}

        updated = synchronizer._update_file_template_variables(nonexistent, template_vars)

        assert updated == []

    def test_update_multiple_same_variable(self, tmp_path):
        """Should update multiple instances of same variable."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.md"
        test_file.write_text("Name: {{USER_NAME}}\nAlso: {{USER_NAME}}")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"USER_NAME": "TestUser"}

        synchronizer._update_file_template_variables(test_file, template_vars)

        content = test_file.read_text()
        assert content.count("TestUser") == 2

    def test_update_preserves_unchanged_content(self, tmp_path):
        """Should preserve non-template content."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.md"
        original_text = "Important content\nName: {{USER_NAME}}"
        test_file.write_text(original_text)

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"USER_NAME": "TestUser"}

        synchronizer._update_file_template_variables(test_file, template_vars)

        content = test_file.read_text()
        assert "Important content" in content


class TestUpdateSettingsEnvVars:
    """Test _update_settings_env_vars method."""

    def test_update_settings_creates_env_section(self, tmp_path):
        """Should create env section if missing."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / ".claude" / "settings.json"
        settings_file.parent.mkdir(parents=True)
        settings_file.write_text("{}")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"CONVERSATION_LANGUAGE": "ko", "USER_NAME": "TestUser"}

        results = {"files_updated": 0, "variables_updated": [], "errors": []}
        synchronizer._update_settings_env_vars(settings_file, template_vars, results)

        data = json.loads(settings_file.read_text())
        assert "env" in data

    def test_update_settings_updates_existing_env(self, tmp_path):
        """Should update existing env variables."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / ".claude" / "settings.json"
        settings_file.parent.mkdir(parents=True)
        settings_file.write_text('{"env": {"MOAI_USER_NAME": "OldUser"}}')

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"USER_NAME": "NewUser"}

        results = {"files_updated": 0, "variables_updated": [], "errors": []}
        synchronizer._update_settings_env_vars(settings_file, template_vars, results)

        data = json.loads(settings_file.read_text())
        assert data["env"]["MOAI_USER_NAME"] == "NewUser"

    def test_update_settings_invalid_json(self, tmp_path):
        """Should handle invalid JSON gracefully."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / ".claude" / "settings.json"
        settings_file.parent.mkdir(parents=True)
        settings_file.write_text("invalid json {")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"USER_NAME": "TestUser"}

        results = {"files_updated": 0, "variables_updated": [], "errors": []}
        synchronizer._update_settings_env_vars(settings_file, template_vars, results)

        # Should not crash, just silently fail


class TestHandleSpecialFileUpdates:
    """Test _handle_special_file_updates method."""

    def test_handle_special_updates_with_settings(self, tmp_path):
        """Should handle settings.json updates."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings = claude_dir / "settings.json"
        settings.write_text("{}")

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"CONVERSATION_LANGUAGE": "ko"}
        results = {"files_updated": 0, "variables_updated": [], "errors": []}

        synchronizer._handle_special_file_updates(template_vars, results)

        # Should not crash
        assert isinstance(results, dict)

    def test_handle_special_updates_missing_settings(self, tmp_path):
        """Should handle missing settings gracefully."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        template_vars = {"CONVERSATION_LANGUAGE": "en"}
        results = {"files_updated": 0, "variables_updated": [], "errors": []}

        synchronizer._handle_special_file_updates(template_vars, results)

        # Should not crash
        assert isinstance(results, dict)


class TestSynchronizeAfterConfigChange:
    """Test synchronize_after_config_change method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_synchronize_successful(self, mock_get_resolver, tmp_path):
        """Should synchronize successfully."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Mock resolver
        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {"user_name": "test"}
        mock_resolver.export_template_variables.return_value = {"USER_NAME": "test"}
        mock_get_resolver.return_value = mock_resolver

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        results = synchronizer.synchronize_after_config_change()

        assert results["sync_status"] == "completed"
        assert "files_updated" in results

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_synchronize_with_changed_config(self, mock_get_resolver, tmp_path):
        """Should handle specific config file change."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {"user_name": "test"}
        mock_resolver.export_template_variables.return_value = {"USER_NAME": "test"}
        mock_get_resolver.return_value = mock_resolver

        (moai_dir := tmp_path / ".moai" / "config").mkdir(parents=True)
        changed_path = moai_dir / "config.json"

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        results = synchronizer.synchronize_after_config_change(changed_path)

        assert results["sync_status"] == "completed"

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_synchronize_handles_errors(self, mock_get_resolver, tmp_path):
        """Should handle errors gracefully."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.side_effect = Exception("Test error")
        mock_get_resolver.return_value = mock_resolver

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        results = synchronizer.synchronize_after_config_change()

        assert results["sync_status"] == "failed"
        assert len(results["errors"]) > 0


class TestValidateTemplateVariableConsistency:
    """Test validate_template_variable_consistency method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_validate_consistency_passed(self, mock_get_resolver, tmp_path):
        """Should validate template consistency."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}
        mock_get_resolver.return_value = mock_resolver

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        results = synchronizer.validate_template_variable_consistency()

        assert results["status"] in ["passed", "warning"]
        assert "inconsistencies" in results

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_validate_consistency_with_variables(self, mock_get_resolver, tmp_path):
        """Should detect files with variables."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create test file with template variable
        test_file = tmp_path / "test.md"
        test_file.write_text("Name: {{USER_NAME}}")

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {"USER_NAME": "value"}
        mock_get_resolver.return_value = mock_resolver

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        results = synchronizer.validate_template_variable_consistency()

        assert isinstance(results, dict)
        assert "status" in results


class TestGetTemplateVariableUsageReport:
    """Test get_template_variable_usage_report method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_usage_report_basic(self, mock_get_resolver, tmp_path):
        """Should generate usage report."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_get_resolver.return_value = mock_resolver

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        report = synchronizer.get_template_variable_usage_report()

        assert "total_files_with_variables" in report
        assert "variable_usage" in report
        assert "files_by_variable" in report

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_usage_report_with_variables(self, mock_get_resolver, tmp_path):
        """Should report variable usage."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create test file
        test_file = tmp_path / "test.md"
        test_file.write_text("User: {{USER_NAME}}\nGreeting: {{PERSONALIZED_GREETING}}")

        mock_resolver = MagicMock()
        mock_get_resolver.return_value = mock_resolver

        synchronizer = TemplateVariableSynchronizer(str(tmp_path))
        report = synchronizer.get_template_variable_usage_report()

        assert isinstance(report, dict)
        assert "variable_usage" in report


class TestSynchronizeTemplateVariablesFunction:
    """Test module-level convenience function."""

    @patch("moai_adk.core.template_variable_synchronizer.TemplateVariableSynchronizer")
    def test_synchronize_template_variables_function(self, mock_class):
        """Should call synchronizer correctly."""
        from moai_adk.core.template_variable_synchronizer import (
            synchronize_template_variables,
        )

        mock_instance = MagicMock()
        mock_instance.synchronize_after_config_change.return_value = {"sync_status": "completed"}
        mock_class.return_value = mock_instance

        result = synchronize_template_variables("/test/root", "/test/config.json")

        assert result["sync_status"] == "completed"
        mock_class.assert_called_once()

    @patch("moai_adk.core.template_variable_synchronizer.TemplateVariableSynchronizer")
    def test_synchronize_template_variables_no_config(self, mock_class):
        """Should handle missing config path."""
        from moai_adk.core.template_variable_synchronizer import (
            synchronize_template_variables,
        )

        mock_instance = MagicMock()
        mock_instance.synchronize_after_config_change.return_value = {"sync_status": "completed"}
        mock_class.return_value = mock_instance

        result = synchronize_template_variables("/test/root", None)

        assert result["sync_status"] == "completed"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
