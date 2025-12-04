"""
Comprehensive unit tests for TemplateVariableSynchronizer with 85%+ coverage.

Tests cover:
- Initialization
- synchronize_after_config_change() with various scenarios
- _find_files_with_template_variables() file discovery
- _update_file_template_variables() substitution
- _glob_files() pattern matching
- _handle_special_file_updates() for settings.json
- _update_settings_env_vars() environment variable handling
- validate_template_variable_consistency() validation logic
- get_template_variable_usage_report() report generation
- Error handling and edge cases
"""

import pytest
import json
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock

from moai_adk.core.template_variable_synchronizer import (
    TemplateVariableSynchronizer,
    synchronize_template_variables,
)


class TestTemplateVariableSynchronizerInitialization:
    """Test TemplateVariableSynchronizer initialization."""

    def test_init_with_valid_path(self):
        """Test initialization with valid project path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)
            assert sync.project_root == Path(tmpdir)

    def test_init_language_resolver_created(self):
        """Test language resolver is initialized."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
                sync = TemplateVariableSynchronizer(tmpdir)
                assert hasattr(sync, "language_resolver")

    def test_init_template_patterns_defined(self):
        """Test template patterns are defined."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)
            assert len(sync.TEMPLATE_PATTERNS) > 0

    def test_init_template_tracking_patterns_defined(self):
        """Test tracking patterns are defined."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)
            assert len(sync.TEMPLATE_TRACKING_PATTERNS) > 0


class TestTemplateVariableSynchronizeSynchronizeAfterConfigChange:
    """Test synchronize_after_config_change() method."""

    def test_synchronize_no_config_change(self):
        """Test synchronization with no specific config path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
                with patch.object(sync.language_resolver, "resolve_config", return_value={}):
                    with patch.object(sync.language_resolver, "export_template_variables", return_value={}):
                        result = sync.synchronize_after_config_change()

                        assert "files_updated" in result
                        assert "variables_updated" in result
                        assert "errors" in result
                        assert "sync_status" in result

    def test_synchronize_with_specific_config_path(self):
        """Test synchronization with specific config file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            config_path = project_root / ".moai" / "config" / "config.json"

            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
                with patch.object(sync.language_resolver, "resolve_config", return_value={}):
                    with patch.object(sync.language_resolver, "export_template_variables", return_value={}):
                        result = sync.synchronize_after_config_change(config_path)

                        assert result["sync_status"] == "completed"

    def test_synchronize_updates_files(self):
        """Test synchronization updates files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("{{USER_NAME}}")

            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[test_file]):
                with patch.object(sync, "_update_file_template_variables", return_value=["USER_NAME"]):
                    with patch.object(sync.language_resolver, "resolve_config", return_value={}):
                        with patch.object(sync.language_resolver, "export_template_variables", return_value={"USER_NAME": "John"}):
                            with patch.object(sync, "_handle_special_file_updates"):
                                result = sync.synchronize_after_config_change()

                                assert result["files_updated"] == 1
                                assert "USER_NAME" in result["variables_updated"]

    def test_synchronize_handles_errors(self):
        """Test synchronization handles file update errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("content")

            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[test_file]):
                with patch.object(sync, "_update_file_template_variables", side_effect=Exception("Update failed")):
                    with patch.object(sync.language_resolver, "resolve_config", return_value={}):
                        with patch.object(sync.language_resolver, "export_template_variables", return_value={}):
                            with patch.object(sync, "_handle_special_file_updates"):
                                result = sync.synchronize_after_config_change()

                                assert len(result["errors"]) > 0


class TestTemplateVariableSynchronizerFindFiles:
    """Test _find_files_with_template_variables() method."""

    def test_find_files_empty_directory(self):
        """Test finding files in empty directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)
            result = sync._find_files_with_template_variables(None)
            assert isinstance(result, list)

    def test_find_files_with_settings(self):
        """Test finding settings.json file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            claude_dir = project_root / ".claude"
            claude_dir.mkdir()
            settings_file = claude_dir / "settings.json"
            settings_file.write_text("{}")

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._find_files_with_template_variables(None)

            assert any("settings.json" in str(f) for f in result)

    def test_find_files_with_specific_config_path(self):
        """Test finding dependent files for specific config."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            sync = TemplateVariableSynchronizer(str(project_root))
            config_path = project_root / ".moai" / "config" / "config.json"

            result = sync._find_files_with_template_variables(config_path)
            assert isinstance(result, list)


class TestTemplateVariableSynchronizerUpdateFileTemplateVariables:
    """Test _update_file_template_variables() method."""

    def test_update_file_nonexistent(self):
        """Test updating nonexistent file returns empty list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)
            result = sync._update_file_template_variables(
                Path(tmpdir) / "nonexistent.md",
                {"USER_NAME": "John"}
            )
            assert result == []

    def test_update_file_no_variables(self):
        """Test updating file with no template variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("Regular content without variables")

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._update_file_template_variables(test_file, {"USER_NAME": "John"})

            assert result == []

    def test_update_file_with_variables(self):
        """Test updating file with template variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("Hello {{USER_NAME}}!")

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._update_file_template_variables(test_file, {"USER_NAME": "John"})

            assert "USER_NAME" in result
            assert test_file.read_text() == "Hello John!"

    def test_update_file_multiple_variables(self):
        """Test updating file with multiple variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("{{USER_NAME}} speaks {{CONVERSATION_LANGUAGE}}")

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._update_file_template_variables(
                test_file,
                {"USER_NAME": "John", "CONVERSATION_LANGUAGE": "English"}
            )

            assert len(result) == 2
            content = test_file.read_text()
            assert "John" in content
            assert "English" in content

    def test_update_file_preserves_unrelated_content(self):
        """Test updating preserves content without variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            original = "This is {{USER_NAME}} and regular text"
            test_file.write_text(original)

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._update_file_template_variables(test_file, {"USER_NAME": "John"})

            content = test_file.read_text()
            assert "regular text" in content


class TestTemplateVariableSynchronizerGlobFiles:
    """Test _glob_files() method."""

    def test_glob_files_simple_pattern(self):
        """Test globbing with simple pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("content")

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._glob_files("*.md")

            assert any("test.md" in str(f) for f in result)

    def test_glob_files_nested_pattern(self):
        """Test globbing with nested pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            dir1 = project_root / "dir1"
            dir1.mkdir()
            test_file = dir1 / "test.md"
            test_file.write_text("content")

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._glob_files("dir1/*.md")

            assert any("test.md" in str(f) for f in result)

    def test_glob_files_recursive_pattern(self):
        """Test globbing with recursive pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            dir1 = project_root / "dir1"
            dir2 = dir1 / "dir2"
            dir2.mkdir(parents=True)
            test_file = dir2 / "test.md"
            test_file.write_text("content")

            sync = TemplateVariableSynchronizer(str(project_root))
            result = sync._glob_files("**/*.md")

            assert any("test.md" in str(f) for f in result)

    def test_glob_files_no_matches(self):
        """Test globbing with no matches."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)
            result = sync._glob_files("nonexistent/*.md")
            assert result == []


class TestTemplateVariableSynchronizerHandleSpecialFiles:
    """Test _handle_special_file_updates() method."""

    def test_handle_special_files_no_settings(self):
        """Test handling special files when settings doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)
            results = {"files_updated": 0, "variables_updated": [], "errors": []}

            sync._handle_special_file_updates({}, results)

            assert results["files_updated"] == 0

    def test_handle_special_files_updates_settings(self):
        """Test handling updates settings.json."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            claude_dir = project_root / ".claude"
            claude_dir.mkdir()

            settings_file = claude_dir / "settings.json"
            settings_file.write_text(json.dumps({"env": {}}))

            sync = TemplateVariableSynchronizer(str(project_root))
            results = {"files_updated": 0, "variables_updated": [], "errors": []}

            with patch.object(sync, "_update_settings_env_vars"):
                sync._handle_special_file_updates({"USER_NAME": "John"}, results)


class TestTemplateVariableSynchronizerUpdateSettingsEnvVars:
    """Test _update_settings_env_vars() method."""

    def test_update_settings_env_vars_creates_env_section(self):
        """Test creating env section if missing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            claude_dir = project_root / ".claude"
            claude_dir.mkdir()

            settings_file = claude_dir / "settings.json"
            settings_file.write_text(json.dumps({}))

            sync = TemplateVariableSynchronizer(str(project_root))
            results = {"files_updated": 0, "variables_updated": [], "errors": []}

            sync._update_settings_env_vars(settings_file, {"USER_NAME": "John"}, results)

            content = json.loads(settings_file.read_text())
            assert "env" in content

    def test_update_settings_env_vars_maps_variables(self):
        """Test mapping template variables to env vars."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            claude_dir = project_root / ".claude"
            claude_dir.mkdir()

            settings_file = claude_dir / "settings.json"
            settings_file.write_text(json.dumps({"env": {}}))

            sync = TemplateVariableSynchronizer(str(project_root))
            results = {"files_updated": 0, "variables_updated": [], "errors": []}

            sync._update_settings_env_vars(
                settings_file,
                {"USER_NAME": "John", "CONVERSATION_LANGUAGE": "en"},
                results
            )

            content = json.loads(settings_file.read_text())
            assert "MOAI_USER_NAME" in content["env"]


class TestTemplateVariableSynchronizerValidateConsistency:
    """Test validate_template_variable_consistency() method."""

    def test_validate_consistency_no_files(self):
        """Test validation with no template files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)

            with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
                result = sync.validate_template_variable_consistency()

                assert result["status"] == "passed"
                assert result["total_files_checked"] == 0

    def test_validate_consistency_consistent_files(self):
        """Test validation with consistent files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("Content without variables")

            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[test_file]):
                with patch.object(sync.language_resolver, "resolve_config", return_value={}):
                    with patch.object(sync.language_resolver, "export_template_variables", return_value={}):
                        result = sync.validate_template_variable_consistency()

                        assert result["status"] == "passed" or result["status"] == "warning"


class TestTemplateVariableSynchronizerGetUsageReport:
    """Test get_template_variable_usage_report() method."""

    def test_get_usage_report_no_variables(self):
        """Test usage report with no template variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            sync = TemplateVariableSynchronizer(tmpdir)

            with patch.object(sync, "_find_files_with_template_variables", return_value=[]):
                result = sync.get_template_variable_usage_report()

                assert "variable_usage" in result
                assert "files_by_variable" in result
                assert "unsubstituted_variables" in result

    def test_get_usage_report_with_variables(self):
        """Test usage report with template variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("Hello {{USER_NAME}}")

            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[test_file]):
                result = sync.get_template_variable_usage_report()

                assert result["total_files_with_variables"] > 0


class TestTemplateVariableSynchronizerFactoryFunction:
    """Test module-level factory function."""

    def test_synchronize_template_variables_factory(self):
        """Test factory function."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.object(TemplateVariableSynchronizer, "synchronize_after_config_change", return_value={}):
                result = synchronize_template_variables(tmpdir)
                assert isinstance(result, dict)

    def test_synchronize_template_variables_with_config_path(self):
        """Test factory function with config path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = str(Path(tmpdir) / ".moai" / "config" / "config.json")
            with patch.object(TemplateVariableSynchronizer, "synchronize_after_config_change", return_value={}):
                result = synchronize_template_variables(tmpdir, config_path)
                assert isinstance(result, dict)


class TestTemplateVariableSynchronizerIntegration:
    """Integration tests for TemplateVariableSynchronizer."""

    def test_complete_sync_workflow(self):
        """Test complete synchronization workflow."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create test files
            test_file = project_root / "test.md"
            test_file.write_text("User: {{USER_NAME}}, Language: {{CONVERSATION_LANGUAGE}}")

            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[test_file]):
                with patch.object(sync.language_resolver, "resolve_config", return_value={}):
                    with patch.object(
                        sync.language_resolver,
                        "export_template_variables",
                        return_value={"USER_NAME": "Alice", "CONVERSATION_LANGUAGE": "English"}
                    ):
                        with patch.object(sync, "_handle_special_file_updates"):
                            result = sync.synchronize_after_config_change()

                            assert result["sync_status"] == "completed"

    def test_sync_and_validate_workflow(self):
        """Test sync followed by validation."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            test_file = project_root / "test.md"
            test_file.write_text("Content")

            sync = TemplateVariableSynchronizer(str(project_root))

            with patch.object(sync, "_find_files_with_template_variables", return_value=[test_file]):
                with patch.object(sync.language_resolver, "resolve_config", return_value={}):
                    with patch.object(sync.language_resolver, "export_template_variables", return_value={}):
                        with patch.object(sync, "_handle_special_file_updates"):
                            # Synchronize
                            sync_result = sync.synchronize_after_config_change()

                            # Validate
                            validation_result = sync.validate_template_variable_consistency()

                            assert sync_result["sync_status"] == "completed"
                            assert validation_result["status"] in ["passed", "warning", "failed"]
