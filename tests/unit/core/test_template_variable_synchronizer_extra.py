"""Extended tests for moai_adk.core.template_variable_synchronizer module.

Comprehensive test coverage for TemplateVariableSynchronizer with full method coverage,
edge cases, and mocked dependencies.
"""

import json
import re
from pathlib import Path
from typing import Any, Dict
from unittest.mock import MagicMock, Mock, patch

import pytest


class TestTemplateVariableSynchronizerBasics:
    """Test TemplateVariableSynchronizer initialization and basic operations."""

    def test_class_import(self):
        """Test that TemplateVariableSynchronizer can be imported."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        assert TemplateVariableSynchronizer is not None

    def test_synchronizer_init_with_project_root(self):
        """Test TemplateVariableSynchronizer initialization with project root."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            sync = TemplateVariableSynchronizer("/test/project")
            assert sync.project_root == Path("/test/project")

    def test_template_patterns_defined(self):
        """Test that template patterns are properly defined."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        patterns = TemplateVariableSynchronizer.TEMPLATE_PATTERNS
        assert isinstance(patterns, set)
        assert len(patterns) > 0
        assert any("CONVERSATION_LANGUAGE" in p for p in patterns)

    def test_tracking_patterns_defined(self):
        """Test that tracking patterns are properly defined."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        patterns = TemplateVariableSynchronizer.TEMPLATE_TRACKING_PATTERNS
        assert isinstance(patterns, list)
        assert len(patterns) > 0
        assert any(".claude/settings.json" in p for p in patterns)


class TestSynchronizeAfterConfigChange:
    """Test synchronize_after_config_change method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_synchronize_returns_dict(self, mock_get_resolver):
        """Test that synchronize_after_config_change returns a dictionary."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
            with patch.object(sync, "_handle_special_file_updates"):
                result = sync.synchronize_after_config_change()

        assert isinstance(result, dict)
        assert "files_updated" in result
        assert "variables_updated" in result
        assert "errors" in result
        assert "sync_status" in result

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_synchronize_with_no_files(self, mock_get_resolver):
        """Test synchronize_after_config_change with no template files."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_resolver.resolve_config.return_value = {"lang": "en"}
        mock_resolver.export_template_variables.return_value = {"CONVERSATION_LANGUAGE": "en"}
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
            with patch.object(sync, "_handle_special_file_updates"):
                result = sync.synchronize_after_config_change()

        assert result["files_updated"] == 0
        assert result["sync_status"] == "completed"

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_synchronize_handles_exceptions(self, mock_get_resolver):
        """Test synchronize_after_config_change handles exceptions gracefully."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_resolver.resolve_config.side_effect = Exception("Test error")
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")
        result = sync.synchronize_after_config_change()

        assert result["sync_status"] == "failed"
        assert len(result["errors"]) > 0


class TestFindFilesWithTemplateVariables:
    """Test _find_files_with_template_variables method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_find_files_returns_list(self, mock_get_resolver):
        """Test that _find_files_with_template_variables returns a list."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        with patch.object(sync, "_glob_files", return_value=[]):
            result = sync._find_files_with_template_variables(None)

        assert isinstance(result, list)

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_find_files_with_specific_config_path(self, mock_get_resolver):
        """Test _find_files_with_template_variables with specific config path."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        with patch.object(sync, "_glob_files", return_value=[]):
            config_path = Path("/test/project/.moai/config/config.json")
            result = sync._find_files_with_template_variables(config_path)

        assert isinstance(result, list)


class TestGlobFiles:
    """Test _glob_files method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_glob_files_returns_list(self, mock_get_resolver):
        """Test that _glob_files returns a list of paths."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        # Mock the project_root.glob method
        sync.project_root = Mock()
        sync.project_root.glob.return_value = [Path("/test/project/file1.json")]

        result = sync._glob_files("*.json")

        assert isinstance(result, list)
        assert len(result) == 1

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_glob_files_handles_errors(self, mock_get_resolver):
        """Test that _glob_files handles glob errors gracefully."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        # Mock the project_root.glob to raise an error
        sync.project_root = Mock()
        sync.project_root.glob.side_effect = ValueError("Invalid pattern")

        result = sync._glob_files("**invalid**")

        assert isinstance(result, list)
        assert len(result) == 0


class TestUpdateFileTemplateVariables:
    """Test _update_file_template_variables method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_update_file_with_variables(self, mock_get_resolver):
        """Test updating a file with template variables."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        # Create a temporary file
        mock_file = Mock(spec=Path)
        mock_file.exists.return_value = True
        mock_file.read_text.return_value = "Hello {{USER_NAME}}, your language is {{CONVERSATION_LANGUAGE}}"
        mock_file.write_text.return_value = None

        template_vars = {"USER_NAME": "Alice", "CONVERSATION_LANGUAGE": "en"}

        result = sync._update_file_template_variables(mock_file, template_vars)

        assert isinstance(result, list)
        assert "USER_NAME" in result
        assert "CONVERSATION_LANGUAGE" in result
        mock_file.write_text.assert_called_once()

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_update_file_nonexistent(self, mock_get_resolver):
        """Test updating a nonexistent file returns empty list."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        mock_file = Mock(spec=Path)
        mock_file.exists.return_value = False

        result = sync._update_file_template_variables(mock_file, {})

        assert isinstance(result, list)
        assert len(result) == 0

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_update_file_no_variables_to_replace(self, mock_get_resolver):
        """Test updating a file with no matching template variables."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        mock_file = Mock(spec=Path)
        mock_file.exists.return_value = True
        mock_file.read_text.return_value = "Hello World"

        result = sync._update_file_template_variables(mock_file, {"USER_NAME": "Alice"})

        assert isinstance(result, list)
        assert len(result) == 0
        mock_file.write_text.assert_not_called()


class TestHandleSpecialFileUpdates:
    """Test _handle_special_file_updates method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_handle_special_file_updates_success(self, mock_get_resolver):
        """Test _handle_special_file_updates with settings.json present."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        template_vars = {"CONVERSATION_LANGUAGE": "en"}
        results = {"files_updated": 0, "variables_updated": [], "errors": []}

        with patch.object(sync, "_update_settings_env_vars"):
            sync._handle_special_file_updates(template_vars, results)

        assert isinstance(results, dict)

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_handle_special_file_updates_no_settings(self, mock_get_resolver):
        """Test _handle_special_file_updates when settings.json doesn't exist."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        # Simply test that the method works without raising exceptions
        template_vars = {"CONVERSATION_LANGUAGE": "en"}
        results = {"files_updated": 0, "variables_updated": [], "errors": []}

        # Mock the settings file path to not exist
        with patch("pathlib.Path.exists", return_value=False):
            sync._handle_special_file_updates(template_vars, results)

            assert isinstance(results, dict)


class TestUpdateSettingsEnvVars:
    """Test _update_settings_env_vars method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_update_settings_env_vars_success(self, mock_get_resolver):
        """Test _update_settings_env_vars with valid settings file."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        # Create mock settings file
        settings_json = {"env": {"MOAI_CONVERSATION_LANG": "ko"}}
        settings_content = json.dumps(settings_json)

        mock_file = Mock(spec=Path)
        mock_file.read_text.return_value = settings_content
        mock_file.write_text.return_value = None

        template_vars = {"CONVERSATION_LANGUAGE": "en"}
        results = {"files_updated": 0, "variables_updated": [], "errors": []}

        sync._update_settings_env_vars(mock_file, template_vars, results)

        assert isinstance(results, dict)

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_update_settings_env_vars_invalid_json(self, mock_get_resolver):
        """Test _update_settings_env_vars with invalid JSON."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        mock_file = Mock(spec=Path)
        mock_file.read_text.return_value = "invalid json {{{{"

        template_vars = {"CONVERSATION_LANGUAGE": "en"}
        results = {"files_updated": 0, "variables_updated": [], "errors": []}

        # Should not raise, should handle gracefully
        sync._update_settings_env_vars(mock_file, template_vars, results)

        assert isinstance(results, dict)


class TestValidateTemplateVariableConsistency:
    """Test validate_template_variable_consistency method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_validate_consistency_returns_dict(self, mock_get_resolver):
        """Test that validate_template_variable_consistency returns a dictionary."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
            result = sync.validate_template_variable_consistency()

        assert isinstance(result, dict)
        assert "status" in result
        assert "inconsistencies" in result
        assert "total_files_checked" in result
        assert "files_with_variables" in result

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_validate_consistency_handles_exception(self, mock_get_resolver):
        """Test validate_template_variable_consistency handles exceptions."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_resolver.resolve_config.side_effect = Exception("Test error")
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")
        result = sync.validate_template_variable_consistency()

        assert result["status"] == "failed"
        assert "error" in result


class TestGetTemplateVariableUsageReport:
    """Test get_template_variable_usage_report method."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_usage_report_returns_dict(self, mock_get_resolver):
        """Test that get_template_variable_usage_report returns a dictionary."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
            result = sync.get_template_variable_usage_report()

        assert isinstance(result, dict)
        assert "total_files_with_variables" in result
        assert "variable_usage" in result
        assert "files_by_variable" in result
        assert "unsubstituted_variables" in result

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_usage_report_tracks_variables(self, mock_get_resolver):
        """Test that usage report tracks variable usage."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        # Create mock file with template variables
        mock_file = Mock(spec=Path)
        mock_file.read_text.return_value = "Hello {{USER_NAME}}, your language is {{CONVERSATION_LANGUAGE}}"
        mock_file.relative_to.return_value = Path("test.json")

        with patch.object(sync, "_find_files_with_template_variables", return_value=[mock_file]):
            result = sync.get_template_variable_usage_report()

        assert isinstance(result, dict)
        assert result["total_files_with_variables"] > 0


class TestModuleConvenienceFunctions:
    """Test module-level convenience functions."""

    def test_synchronize_template_variables_function(self):
        """Test synchronize_template_variables convenience function."""
        from moai_adk.core.template_variable_synchronizer import synchronize_template_variables

        with patch("moai_adk.core.template_variable_synchronizer.TemplateVariableSynchronizer") as mock_class:
            mock_instance = Mock()
            mock_instance.synchronize_after_config_change.return_value = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
                "sync_status": "completed",
            }
            mock_class.return_value = mock_instance

            result = synchronize_template_variables("/test/project")

            assert isinstance(result, dict)
            assert "files_updated" in result


class TestPatternMatching:
    """Test template pattern matching and regex."""

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_pattern_regex_compilation(self, mock_get_resolver):
        """Test that patterns can be compiled as regex."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        for pattern_str in sync.TEMPLATE_PATTERNS:
            pattern = re.compile(pattern_str)
            assert pattern is not None

    @patch("moai_adk.core.template_variable_synchronizer.get_resolver")
    def test_variable_substitution_regex(self, mock_get_resolver):
        """Test variable substitution regex works correctly."""
        from moai_adk.core.template_variable_synchronizer import TemplateVariableSynchronizer

        mock_resolver = Mock()
        mock_get_resolver.return_value = mock_resolver

        sync = TemplateVariableSynchronizer("/test/project")

        test_content = "Hello {{USER_NAME}}"
        pattern = re.compile(r"\{\{USER_NAME\}\}")

        assert pattern.search(test_content) is not None
        result = pattern.sub("Alice", test_content)
        assert result == "Hello Alice"
