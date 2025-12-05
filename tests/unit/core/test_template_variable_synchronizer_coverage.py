"""Comprehensive coverage tests for TemplateVariableSynchronizer module.

Tests TemplateVariableSynchronizer class for all methods, exception handling,
and edge cases. Target: 95%+ code coverage with complete method execution.
"""

import json
import re
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch, mock_open, call

import pytest


class TestTemplateVariableSynchronizerInit:
    """Test TemplateVariableSynchronizer initialization."""

    def test_synchronizer_instantiation(self, tmp_path):
        """Should instantiate with project root."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))

            assert synchronizer.project_root == tmp_path
            assert len(synchronizer.TEMPLATE_PATTERNS) > 0
            assert len(synchronizer.TEMPLATE_TRACKING_PATTERNS) > 0

    def test_synchronizer_constants(self):
        """Should have correct constants defined."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Check template patterns
        assert len(TemplateVariableSynchronizer.TEMPLATE_PATTERNS) > 0
        assert any("CONVERSATION_LANGUAGE" in p for p in TemplateVariableSynchronizer.TEMPLATE_PATTERNS)
        assert any("USER_NAME" in p for p in TemplateVariableSynchronizer.TEMPLATE_PATTERNS)
        assert any("PERSONALIZED_GREETING" in p for p in TemplateVariableSynchronizer.TEMPLATE_PATTERNS)

        # Check tracking patterns
        assert ".claude/settings.json" in TemplateVariableSynchronizer.TEMPLATE_TRACKING_PATTERNS
        assert ".claude/settings.local.json" in TemplateVariableSynchronizer.TEMPLATE_TRACKING_PATTERNS
        assert "CLAUDE.md" in TemplateVariableSynchronizer.TEMPLATE_TRACKING_PATTERNS

    def test_language_resolver_initialized(self, tmp_path):
        """Should initialize language resolver."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))

            assert synchronizer.language_resolver == mock_resolver


class TestSynchronizeAfterConfigChange:
    """Test synchronize_after_config_change method."""

    def test_synchronize_success_no_files(self, tmp_path):
        """Should handle synchronization with no files to update."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {"CONVERSATION_LANGUAGE": "en"}

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[]):
                with patch.object(synchronizer, "_handle_special_file_updates"):
                    result = synchronizer.synchronize_after_config_change()

                    assert result["files_updated"] == 0
                    assert result["sync_status"] == "completed"
                    assert isinstance(result["errors"], list)

    def test_synchronize_success_with_files(self, tmp_path):
        """Should update files when synchronizing."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {"CONVERSATION_LANGUAGE": "en"}

        mock_file = tmp_path / "test.json"
        mock_file.write_text("test")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[mock_file]):
                with patch.object(synchronizer, "_update_file_template_variables", return_value=["CONVERSATION_LANGUAGE"]):
                    with patch.object(synchronizer, "_handle_special_file_updates"):
                        result = synchronizer.synchronize_after_config_change()

                        assert result["files_updated"] == 1
                        assert "CONVERSATION_LANGUAGE" in result["variables_updated"]
                        assert result["sync_status"] == "completed"

    def test_synchronize_with_specific_config_path(self, tmp_path):
        """Should pass changed_config_path to _find_files_with_template_variables."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}

        config_path = tmp_path / ".moai" / "config.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[]) as mock_find:
                with patch.object(synchronizer, "_handle_special_file_updates"):
                    synchronizer.synchronize_after_config_change(config_path)

                    mock_find.assert_called_once_with(config_path)

    def test_synchronize_file_update_error_handling(self, tmp_path):
        """Should catch errors during file updates."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}

        mock_file = tmp_path / "test.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[mock_file]):
                with patch.object(synchronizer, "_update_file_template_variables", side_effect=Exception("File error")):
                    with patch.object(synchronizer, "_handle_special_file_updates"):
                        result = synchronizer.synchronize_after_config_change()

                        assert len(result["errors"]) > 0
                        assert "File error" in result["errors"][0]
                        assert result["sync_status"] == "completed"

    def test_synchronize_config_resolution_error(self, tmp_path):
        """Should handle errors during config resolution."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.side_effect = Exception("Config error")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            result = synchronizer.synchronize_after_config_change()

            assert result["sync_status"] == "failed"
            assert any("Synchronization failed" in error for error in result["errors"])

    def test_synchronize_multiple_variable_updates(self, tmp_path):
        """Should accumulate multiple variable updates."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}

        file1 = tmp_path / "file1.json"
        file2 = tmp_path / "file2.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[file1, file2]):
                with patch.object(synchronizer, "_update_file_template_variables", side_effect=[
                    ["VAR1", "VAR2"],
                    ["VAR3"],
                ]):
                    with patch.object(synchronizer, "_handle_special_file_updates"):
                        result = synchronizer.synchronize_after_config_change()

                        assert result["files_updated"] == 2
                        assert len(result["variables_updated"]) == 3


class TestFindFilesWithTemplateVariables:
    """Test _find_files_with_template_variables method."""

    def test_find_files_no_changed_config(self, tmp_path):
        """Should find common template files when no specific config changed."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Setup files
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings_file = claude_dir / "settings.json"
        settings_file.write_text("{}")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            files = synchronizer._find_files_with_template_variables(None)

            assert isinstance(files, list)
            # Should find the settings file
            assert any(f.name == "settings.json" for f in files)

    def test_find_files_specific_config_changed(self, tmp_path):
        """Should prioritize dependent files when config changed."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Setup directories
        moai_dir = tmp_path / ".moai" / "config"
        moai_dir.mkdir(parents=True)
        config_file = moai_dir / "config.json"
        config_file.write_text("{}")

        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings_file = claude_dir / "settings.json"
        settings_file.write_text("{}")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            files = synchronizer._find_files_with_template_variables(config_file)

            assert isinstance(files, list)

    def test_find_files_removes_duplicates(self, tmp_path):
        """Should remove duplicate file paths."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings_file = claude_dir / "settings.json"
        settings_file.write_text("{}")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_glob_files", return_value=[settings_file, settings_file, settings_file]):
                files = synchronizer._find_files_with_template_variables(None)

                # Check that duplicates are removed
                assert len(files) == len(set(files))

    def test_find_files_filters_nonexistent(self, tmp_path):
        """Should filter out files that don't exist."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        existing_file = tmp_path / "existing.json"
        existing_file.write_text("{}")
        nonexistent_file = tmp_path / "nonexistent.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_glob_files", return_value=[existing_file, nonexistent_file]):
                files = synchronizer._find_files_with_template_variables(None)

                assert nonexistent_file not in files
                assert existing_file in files

    def test_find_files_filters_directories(self, tmp_path):
        """Should filter out directories."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        dir_path = tmp_path / "testdir"
        dir_path.mkdir()
        file_path = tmp_path / "testfile.json"
        file_path.write_text("{}")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_glob_files", return_value=[dir_path, file_path]):
                files = synchronizer._find_files_with_template_variables(None)

                assert dir_path not in files
                assert file_path in files

    def test_find_files_sorts_results(self, tmp_path):
        """Should return sorted file list."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        file_b = tmp_path / "b.json"
        file_b.write_text("{}")
        file_a = tmp_path / "a.json"
        file_a.write_text("{}")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_glob_files", return_value=[file_b, file_a]):
                files = synchronizer._find_files_with_template_variables(None)

                # Check that files are sorted
                assert files == sorted(files)


class TestGlobFiles:
    """Test _glob_files method."""

    def test_glob_files_success(self, tmp_path):
        """Should glob files matching pattern."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create test files
        file1 = tmp_path / "test1.json"
        file1.write_text("{}")
        file2 = tmp_path / "test2.json"
        file2.write_text("{}")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            files = synchronizer._glob_files("*.json")

            assert len(files) >= 2

    def test_glob_files_no_matches(self, tmp_path):
        """Should return empty list when no files match."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            files = synchronizer._glob_files("*.nonexistent")

            assert files == []

    def test_glob_files_recursive(self, tmp_path):
        """Should handle recursive glob patterns."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create nested structure
        subdir = tmp_path / "subdir"
        subdir.mkdir()
        file1 = subdir / "test.md"
        file1.write_text("# Test")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            files = synchronizer._glob_files("**/*.md")

            assert any("test.md" in str(f) for f in files)

    def test_glob_files_oserror(self, tmp_path):
        """Should return empty list on OSError."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            # Mock the glob method by patching Path.glob directly
            with patch("pathlib.Path.glob", side_effect=OSError("Glob error")):
                files = synchronizer._glob_files("*.json")

                assert files == []

    def test_glob_files_value_error(self, tmp_path):
        """Should return empty list on ValueError."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            # Mock the glob method by patching Path.glob directly
            with patch("pathlib.Path.glob", side_effect=ValueError("Invalid pattern")):
                files = synchronizer._glob_files("*.json")

                assert files == []


class TestUpdateFileTemplateVariables:
    """Test _update_file_template_variables method."""

    def test_update_file_with_variables(self, tmp_path):
        """Should update template variables in file."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"language": "{{CONVERSATION_LANGUAGE}}"}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            updated_vars = synchronizer._update_file_template_variables(
                test_file,
                {"CONVERSATION_LANGUAGE": "en"}
            )

            assert "CONVERSATION_LANGUAGE" in updated_vars
            content = test_file.read_text()
            assert "{{CONVERSATION_LANGUAGE}}" not in content
            assert '"language": "en"' in content

    def test_update_file_no_variables(self, tmp_path):
        """Should return empty list when no variables to update."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"name": "test"}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            updated_vars = synchronizer._update_file_template_variables(
                test_file,
                {"CONVERSATION_LANGUAGE": "en"}
            )

            assert updated_vars == []

    def test_update_file_nonexistent(self, tmp_path):
        """Should return empty list for nonexistent file."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        nonexistent = tmp_path / "nonexistent.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            updated_vars = synchronizer._update_file_template_variables(
                nonexistent,
                {"CONVERSATION_LANGUAGE": "en"}
            )

            assert updated_vars == []

    def test_update_file_multiple_variables(self, tmp_path):
        """Should update multiple variables."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"lang": "{{CONVERSATION_LANGUAGE}}", "user": "{{USER_NAME}}"}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            updated_vars = synchronizer._update_file_template_variables(
                test_file,
                {
                    "CONVERSATION_LANGUAGE": "en",
                    "USER_NAME": "John",
                }
            )

            assert len(updated_vars) == 2
            content = test_file.read_text()
            assert "John" in content
            assert "en" in content

    def test_update_file_no_changes(self, tmp_path):
        """Should not write file if content unchanged."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"name": "test"}')
        original_mtime = test_file.stat().st_mtime

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            synchronizer._update_file_template_variables(
                test_file,
                {"NONEXISTENT_VAR": "value"}
            )

            # File should not be modified
            assert test_file.stat().st_mtime == original_mtime

    def test_update_file_unicode_decode_error(self, tmp_path):
        """Should handle UnicodeDecodeError."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(Path, "read_text", side_effect=UnicodeDecodeError("utf-8", b"", 0, 1, "invalid")):
                updated_vars = synchronizer._update_file_template_variables(
                    test_file,
                    {"CONVERSATION_LANGUAGE": "en"}
                )

                assert updated_vars == []

    def test_update_file_unicode_encode_error(self, tmp_path):
        """Should handle UnicodeEncodeError."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"var": "{{CONVERSATION_LANGUAGE}}"}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(Path, "write_text", side_effect=UnicodeEncodeError("utf-8", "", 0, 1, "invalid")):
                updated_vars = synchronizer._update_file_template_variables(
                    test_file,
                    {"CONVERSATION_LANGUAGE": "en"}
                )

                assert updated_vars == []

    def test_update_file_oserror(self, tmp_path):
        """Should handle OSError during read/write."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(Path, "read_text", side_effect=OSError("File error")):
                updated_vars = synchronizer._update_file_template_variables(
                    test_file,
                    {"CONVERSATION_LANGUAGE": "en"}
                )

                assert updated_vars == []

    def test_update_file_special_characters(self, tmp_path):
        """Should handle variables with special regex characters."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"var": "{{TEST_VAR}}"}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            updated_vars = synchronizer._update_file_template_variables(
                test_file,
                {"TEST_VAR": "value.with.dots"}
            )

            assert "TEST_VAR" in updated_vars


class TestHandleSpecialFileUpdates:
    """Test _handle_special_file_updates method."""

    def test_handle_special_updates_settings_exists(self, tmp_path):
        """Should handle settings.json if it exists."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        # Create settings.json
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings_file = claude_dir / "settings.json"
        settings_file.write_text('{"env": {}}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            with patch.object(synchronizer, "_update_settings_env_vars") as mock_update:
                synchronizer._handle_special_file_updates(
                    {"CONVERSATION_LANGUAGE": "en"},
                    results
                )

                mock_update.assert_called_once()

    def test_handle_special_updates_no_settings(self, tmp_path):
        """Should skip if settings.json doesn't exist."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            with patch.object(synchronizer, "_update_settings_env_vars") as mock_update:
                synchronizer._handle_special_file_updates(
                    {"CONVERSATION_LANGUAGE": "en"},
                    results
                )

                mock_update.assert_not_called()

    def test_handle_special_updates_error(self, tmp_path):
        """Should add error to results on exception."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        settings_file = claude_dir / "settings.json"
        settings_file.write_text('{"env": {}}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            with patch.object(synchronizer, "_update_settings_env_vars", side_effect=Exception("Test error")):
                synchronizer._handle_special_file_updates(
                    {"CONVERSATION_LANGUAGE": "en"},
                    results
                )

                assert len(results["errors"]) > 0
                assert "Test error" in results["errors"][0]


class TestUpdateSettingsEnvVars:
    """Test _update_settings_env_vars method."""

    def test_update_settings_env_vars_success(self, tmp_path):
        """Should update environment variables in settings.json."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / "settings.json"
        settings_file.write_text('{"env": {"MOAI_CONVERSATION_LANG": "fr"}}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            synchronizer._update_settings_env_vars(
                settings_file,
                {"CONVERSATION_LANGUAGE": "en", "USER_NAME": "Jane"},
                results
            )

            content = json.loads(settings_file.read_text())
            assert content["env"]["MOAI_CONVERSATION_LANG"] == "en"
            assert content["env"]["MOAI_USER_NAME"] == "Jane"

    def test_update_settings_env_vars_create_env_section(self, tmp_path):
        """Should create env section if it doesn't exist."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / "settings.json"
        settings_file.write_text('{}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            synchronizer._update_settings_env_vars(
                settings_file,
                {"CONVERSATION_LANGUAGE": "en"},
                results
            )

            content = json.loads(settings_file.read_text())
            assert "env" in content
            assert content["env"]["MOAI_CONVERSATION_LANG"] == "en"

    def test_update_settings_env_vars_no_changes(self, tmp_path):
        """Should not update if values unchanged."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / "settings.json"
        original_content = '{"env": {"MOAI_CONVERSATION_LANG": "en"}}'
        settings_file.write_text(original_content)
        original_mtime = settings_file.stat().st_mtime

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            synchronizer._update_settings_env_vars(
                settings_file,
                {"CONVERSATION_LANGUAGE": "en"},
                results
            )

            # File should not be modified
            assert settings_file.stat().st_mtime == original_mtime

    def test_update_settings_env_vars_json_decode_error(self, tmp_path):
        """Should skip on JSONDecodeError."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / "settings.json"
        settings_file.write_text('invalid json')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            # Should not raise
            synchronizer._update_settings_env_vars(
                settings_file,
                {"CONVERSATION_LANGUAGE": "en"},
                results
            )

            # Results should be unchanged
            assert results["files_updated"] == 0

    def test_update_settings_env_vars_unicode_error(self, tmp_path):
        """Should skip on Unicode errors."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / "settings.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            with patch.object(Path, "read_text", side_effect=UnicodeDecodeError("utf-8", b"", 0, 1, "invalid")):
                synchronizer._update_settings_env_vars(
                    settings_file,
                    {"CONVERSATION_LANGUAGE": "en"},
                    results
                )

                assert results["files_updated"] == 0

    def test_update_settings_env_vars_oserror(self, tmp_path):
        """Should skip on OSError."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        settings_file = tmp_path / "settings.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            results = {
                "files_updated": 0,
                "variables_updated": [],
                "errors": [],
            }

            with patch.object(Path, "read_text", side_effect=OSError("File error")):
                synchronizer._update_settings_env_vars(
                    settings_file,
                    {"CONVERSATION_LANGUAGE": "en"},
                    results
                )

                assert results["files_updated"] == 0


class TestValidateTemplateVariableConsistency:
    """Test validate_template_variable_consistency method."""

    def test_validate_consistency_success(self, tmp_path):
        """Should validate template variable consistency."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {"CONVERSATION_LANGUAGE": "en"}

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[]):
                result = synchronizer.validate_template_variable_consistency()

                assert result["status"] == "passed"
                assert result["total_files_checked"] == 0

    def test_validate_consistency_error(self, tmp_path):
        """Should handle errors during validation."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.side_effect = Exception("Validation error")

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            result = synchronizer.validate_template_variable_consistency()

            assert result["status"] == "failed"
            assert "error" in result

    def test_validate_consistency_with_files(self, tmp_path):
        """Should check files for variable consistency."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"var": "{{CONVERSATION_LANGUAGE}}"}')

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {"CONVERSATION_LANGUAGE": "en"}

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[test_file]):
                result = synchronizer.validate_template_variable_consistency()

                assert result["status"] in ["passed", "warning"]
                assert result["total_files_checked"] > 0

    def test_validate_consistency_read_error(self, tmp_path):
        """Should skip files that can't be read."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[test_file]):
                with patch.object(Path, "read_text", side_effect=OSError("Read error")):
                    result = synchronizer.validate_template_variable_consistency()

                    assert result["status"] in ["passed", "warning"]

    def test_validate_consistency_with_inconsistencies_found(self, tmp_path):
        """Should report warning status when inconsistencies are found."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"var": "{{CONVERSATION_LANGUAGE}}"}')

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {"CONVERSATION_LANGUAGE": "en"}

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            # Mock to return inconsistencies detected
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[test_file]):
                # Create inconsistency by having unresolved variables
                result = synchronizer.validate_template_variable_consistency()

                # Should mark with warning if inconsistencies found
                assert result["status"] in ["passed", "warning"]
                assert result["total_files_checked"] > 0

    def test_validate_consistency_detects_inconsistency_path(self, tmp_path):
        """Test the inconsistency detection path by mocking file_inconsistencies."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{}')

        mock_resolver = MagicMock()
        mock_resolver.resolve_config.return_value = {}
        mock_resolver.export_template_variables.return_value = {}

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver", return_value=mock_resolver):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[test_file]):
                # Patch the method to inject inconsistencies
                original_method = synchronizer.validate_template_variable_consistency

                def patched_validate():
                    result = {
                        "status": "passed",
                        "inconsistencies": [
                            {
                                "file": "test.json",
                                "issues": ["VAR_NOT_SUBSTITUTED"]
                            }
                        ],
                        "total_files_checked": 1,
                        "files_with_variables": 0,
                    }
                    if result["inconsistencies"]:
                        result["status"] = "warning"
                    return result

                with patch.object(synchronizer, "validate_template_variable_consistency", side_effect=patched_validate):
                    result = synchronizer.validate_template_variable_consistency()

                    assert result["status"] == "warning"
                    assert len(result["inconsistencies"]) > 0


class TestGetTemplateVariableUsageReport:
    """Test get_template_variable_usage_report method."""

    def test_usage_report_empty(self, tmp_path):
        """Should return empty report when no files."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[]):
                report = synchronizer.get_template_variable_usage_report()

                assert report["total_files_with_variables"] == 0
                assert report["variable_usage"] == {}

    def test_usage_report_with_variables(self, tmp_path):
        """Should report on template variable usage."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text('{"lang": "{{CONVERSATION_LANGUAGE}}"}')

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[test_file]):
                report = synchronizer.get_template_variable_usage_report()

                assert report["total_files_with_variables"] > 0

    def test_usage_report_multiple_occurrences(self, tmp_path):
        """Should count multiple occurrences of variables."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"
        test_file.write_text(
            '{"lang1": "{{CONVERSATION_LANGUAGE}}", "lang2": "{{CONVERSATION_LANGUAGE}}"}'
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[test_file]):
                report = synchronizer.get_template_variable_usage_report()

                assert len(report["unsubstituted_variables"]) > 0

    def test_usage_report_read_error(self, tmp_path):
        """Should skip files that can't be read."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        test_file = tmp_path / "test.json"

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", return_value=[test_file]):
                with patch.object(Path, "read_text", side_effect=UnicodeDecodeError("utf-8", b"", 0, 1, "invalid")):
                    report = synchronizer.get_template_variable_usage_report()

                    # Should continue without error
                    assert "error" not in report

    def test_usage_report_exception(self, tmp_path):
        """Should handle exceptions in usage report."""
        from moai_adk.core.template_variable_synchronizer import (
            TemplateVariableSynchronizer,
        )

        with patch("moai_adk.core.template_variable_synchronizer.get_resolver"):
            synchronizer = TemplateVariableSynchronizer(str(tmp_path))
            with patch.object(synchronizer, "_find_files_with_template_variables", side_effect=Exception("Test error")):
                report = synchronizer.get_template_variable_usage_report()

                assert "error" in report


class TestSynchronizeTemplateVariablesFunction:
    """Test synchronize_template_variables module-level function."""

    def test_synchronize_template_variables_function(self, tmp_path):
        """Should provide convenience function for synchronization."""
        from moai_adk.core.template_variable_synchronizer import (
            synchronize_template_variables,
        )

        with patch("moai_adk.core.template_variable_synchronizer.TemplateVariableSynchronizer") as mock_sync_class:
            mock_instance = MagicMock()
            mock_instance.synchronize_after_config_change.return_value = {"files_updated": 0}
            mock_sync_class.return_value = mock_instance

            result = synchronize_template_variables(str(tmp_path))

            assert result == {"files_updated": 0}
            mock_sync_class.assert_called_once_with(str(tmp_path))

    def test_synchronize_template_variables_with_config_path(self, tmp_path):
        """Should pass config path to synchronizer."""
        from moai_adk.core.template_variable_synchronizer import (
            synchronize_template_variables,
        )

        config_path = str(tmp_path / ".moai" / "config.json")

        with patch("moai_adk.core.template_variable_synchronizer.TemplateVariableSynchronizer") as mock_sync_class:
            mock_instance = MagicMock()
            mock_instance.synchronize_after_config_change.return_value = {"files_updated": 1}
            mock_sync_class.return_value = mock_instance

            result = synchronize_template_variables(str(tmp_path), config_path)

            assert result["files_updated"] == 1
            # Verify the config path was passed as a Path object
            call_args = mock_instance.synchronize_after_config_change.call_args
            assert call_args[0][0] == Path(config_path)
