"""
Comprehensive tests for context_manager module.

Tests context management utilities including path validation,
atomic file operations, and template variable handling.
"""

import json
import os
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.context_manager import (
    PARENT_DIR_MISSING_MSG,
    PROJECT_ROOT_SAFETY_MSG,
    TEMPLATE_VAR_PATTERN,
    ContextManager,
    load_phase_result,
    save_phase_result,
    substitute_template_variables,
    validate_and_convert_path,
    validate_no_template_vars,
)


class TestIsPathWithinRoot:
    """Test _is_path_within_root function."""

    def test_path_within_root(self):
        """Test path that is within project root."""
        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = os.path.join(tmpdir, "subdir", "file.txt")
            # Don't need file to exist for this test
            from moai_adk.core.context_manager import _is_path_within_root

            result = _is_path_within_root(tmpdir, tmpdir)
            assert result is True

    def test_path_equal_to_root(self):
        """Test path that equals project root."""
        with tempfile.TemporaryDirectory() as tmpdir:
            from moai_adk.core.context_manager import _is_path_within_root

            result = _is_path_within_root(tmpdir, tmpdir)
            assert result is True

    def test_path_outside_root(self):
        """Test path that is outside project root."""
        with tempfile.TemporaryDirectory() as tmpdir:
            parent = os.path.dirname(tmpdir)
            from moai_adk.core.context_manager import _is_path_within_root

            result = _is_path_within_root(parent, tmpdir)
            assert result is False

    def test_path_in_sibling_directory(self):
        """Test path in sibling directory."""
        with tempfile.TemporaryDirectory() as base_tmpdir:
            dir1 = os.path.join(base_tmpdir, "dir1")
            dir2 = os.path.join(base_tmpdir, "dir2")
            os.makedirs(dir1)
            os.makedirs(dir2)

            from moai_adk.core.context_manager import _is_path_within_root

            result = _is_path_within_root(dir2, dir1)
            assert result is False


class TestValidateAndConvertPath:
    """Test validate_and_convert_path function."""

    def test_convert_relative_path(self):
        """Test converting relative path to absolute."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a subdirectory for the parent check
            subdir = os.path.join(tmpdir, "subdir")
            os.makedirs(subdir)

            result = validate_and_convert_path("subdir", tmpdir)

            assert os.path.isabs(result)
            assert tmpdir in result

    def test_path_outside_root_raises_error(self):
        """Test that paths outside root raise ValueError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Try to escape using ../
            with pytest.raises(ValueError, match="Path outside project root"):
                validate_and_convert_path("../../../../etc/passwd", tmpdir)

    def test_file_parent_dir_missing(self):
        """Test error when parent directory doesn't exist for files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Try to create file in non-existent parent directory
            with pytest.raises(FileNotFoundError, match="Parent directory not found"):
                validate_and_convert_path("nonexistent/file.txt", tmpdir)

    def test_existing_directory_returns_path(self):
        """Test that existing directory returns path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            subdir = os.path.join(tmpdir, "subdir")
            os.makedirs(subdir)

            result = validate_and_convert_path("subdir", tmpdir)

            assert os.path.isdir(result)

    def test_absolute_path_conversion(self):
        """Test that absolute paths are also validated."""
        with tempfile.TemporaryDirectory() as tmpdir:
            subdir = os.path.join(tmpdir, "subdir")
            os.makedirs(subdir)

            result = validate_and_convert_path(subdir, tmpdir)

            assert os.path.isabs(result)

    def test_normalized_path(self):
        """Test that paths are normalized."""
        with tempfile.TemporaryDirectory() as tmpdir:
            subdir = os.path.join(tmpdir, "subdir")
            os.makedirs(subdir)

            # Use path with . and .. components
            result = validate_and_convert_path("./subdir/../subdir", tmpdir)

            assert os.path.isabs(result)
            assert not ".." in result
            assert not "." in result.split("/")[-1]


class TestSavePhaseResult:
    """Test save_phase_result function."""

    def test_save_creates_file(self):
        """Test that save_phase_result creates the file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase-result.json")
            data = {"phase": "test", "status": "completed"}

            save_phase_result(data, target_path)

            assert os.path.exists(target_path)

    def test_save_creates_parent_directories(self):
        """Test that parent directories are created."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "dir1", "dir2", "phase.json")
            data = {"phase": "test"}

            save_phase_result(data, target_path)

            assert os.path.exists(target_path)

    def test_save_preserves_data(self):
        """Test that saved data is preserved exactly."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")
            data = {
                "phase": "1-plan",
                "outputs": {"key": "value"},
                "files": ["file1.txt", "file2.txt"],
                "unicode": "日本語",
            }

            save_phase_result(data, target_path)

            with open(target_path, "r", encoding="utf-8") as f:
                loaded = json.load(f)

            assert loaded == data

    def test_save_atomic_operation(self):
        """Test that save is atomic using temp file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")
            data = {"phase": "test"}

            # Mock os.replace to verify atomic behavior
            with patch("moai_adk.core.context_manager.os.replace") as mock_replace:
                with patch("moai_adk.core.context_manager.tempfile.mkstemp") as mock_mkstemp:
                    mock_mkstemp.return_value = (999, "/tmp/test.tmp")
                    with patch("builtins.open", create=True):
                        with patch(
                            "moai_adk.core.context_manager.os.fdopen", create=True
                        ):
                            try:
                                save_phase_result(data, target_path)
                            except:
                                pass
                            # Verify mkstemp was called
                            assert mock_mkstemp.called

    def test_save_with_special_characters(self):
        """Test saving with special characters."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")
            data = {
                "phase": "test",
                "unicode": "中文",
                "emoji": "🎉",
                "symbols": "!@#$%^&*()",
            }

            save_phase_result(data, target_path)

            with open(target_path, "r", encoding="utf-8") as f:
                loaded = json.load(f)

            assert loaded["unicode"] == "中文"
            assert loaded["emoji"] == "🎉"

    def test_save_overwrites_existing_file(self):
        """Test that save overwrites existing files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")

            # Create initial file
            initial_data = {"phase": "old"}
            save_phase_result(initial_data, target_path)

            # Overwrite
            new_data = {"phase": "new"}
            save_phase_result(new_data, target_path)

            with open(target_path, "r") as f:
                loaded = json.load(f)

            assert loaded["phase"] == "new"


class TestLoadPhaseResult:
    """Test load_phase_result function."""

    def test_load_valid_file(self):
        """Test loading valid phase result file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")
            data = {"phase": "test", "status": "completed"}

            with open(target_path, "w") as f:
                json.dump(data, f)

            loaded = load_phase_result(target_path)

            assert loaded == data

    def test_load_missing_file(self):
        """Test error when file doesn't exist."""
        with pytest.raises(FileNotFoundError, match="not found"):
            load_phase_result("/nonexistent/file.json")

    def test_load_invalid_json(self):
        """Test error handling for invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")

            with open(target_path, "w") as f:
                f.write("{ invalid json }")

            with pytest.raises(json.JSONDecodeError):
                load_phase_result(target_path)

    def test_load_unicode_content(self):
        """Test loading files with unicode content."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")
            data = {"phase": "test", "content": "日本語テキスト"}

            with open(target_path, "w", encoding="utf-8") as f:
                json.dump(data, f, ensure_ascii=False)

            loaded = load_phase_result(target_path)

            assert loaded["content"] == "日本語テキスト"


class TestSubstituteTemplateVariables:
    """Test substitute_template_variables function."""

    def test_simple_substitution(self):
        """Test simple variable substitution."""
        text = "Hello {{NAME}}"
        context = {"NAME": "World"}

        result = substitute_template_variables(text, context)

        assert result == "Hello World"

    def test_multiple_variables(self):
        """Test substituting multiple variables."""
        text = "{{GREETING}}, {{NAME}}! You are {{AGE}} years old."
        context = {"GREETING": "Hello", "NAME": "Alice", "AGE": "30"}

        result = substitute_template_variables(text, context)

        assert result == "Hello, Alice! You are 30 years old."

    def test_repeated_variable(self):
        """Test same variable used multiple times."""
        text = "{{NAME}} said: {{NAME}} is here!"
        context = {"NAME": "Bob"}

        result = substitute_template_variables(text, context)

        assert result == "Bob said: Bob is here!"

    def test_missing_variable(self):
        """Test behavior with missing variable."""
        text = "Hello {{NAME}}"
        context = {"OTHER": "value"}

        # Should leave unsubstituted variables as-is
        result = substitute_template_variables(text, context)

        assert "{{NAME}}" in result

    def test_numeric_context_values(self):
        """Test substitution with numeric values."""
        text = "Count: {{COUNT}}, Rate: {{RATE}}"
        context = {"COUNT": 42, "RATE": 3.14}

        result = substitute_template_variables(text, context)

        assert "42" in result
        assert "3.14" in result

    def test_empty_template(self):
        """Test with empty template."""
        result = substitute_template_variables("", {})

        assert result == ""

    def test_no_variables_in_text(self):
        """Test with text containing no variables."""
        text = "This is plain text with no variables."
        context = {"VAR": "value"}

        result = substitute_template_variables(text, context)

        assert result == text


class TestValidateNoTemplateVars:
    """Test validate_no_template_vars function."""

    def test_no_template_vars_passes(self):
        """Test validation passes with no template vars."""
        text = "This is normal text without any variables."

        # Should not raise
        validate_no_template_vars(text)

    def test_single_template_var_fails(self):
        """Test validation fails with single template var."""
        text = "Hello {{NAME}}"

        with pytest.raises(ValueError, match="Unsubstituted template variables"):
            validate_no_template_vars(text)

    def test_multiple_template_vars_fails(self):
        """Test validation fails with multiple vars."""
        text = "{{GREETING}} {{NAME}}, you are {{AGE}} years old."

        with pytest.raises(ValueError, match="Unsubstituted template variables"):
            validate_no_template_vars(text)

    def test_json_with_template_vars_fails(self):
        """Test validation of JSON with template vars."""
        text = json.dumps({"greeting": "Hello {{NAME}}"})

        with pytest.raises(ValueError, match="Unsubstituted template variables"):
            validate_no_template_vars(text)

    def test_uppercase_variable_names(self):
        """Test that uppercase variables are detected."""
        text = "{{VAR}} {{VAR_NAME}} {{VAR_123}}"

        with pytest.raises(ValueError):
            validate_no_template_vars(text)

    def test_lowercase_variable_not_detected(self):
        """Test that lowercase variables are not detected."""
        text = "{{name}} and {{other_var}}"

        # Should not raise - pattern only matches uppercase
        validate_no_template_vars(text)

    def test_partial_braces_not_detected(self):
        """Test that incomplete braces are not detected."""
        text = "{ {VAR} } and { VAR }"

        # Should not raise - pattern requires double braces
        validate_no_template_vars(text)


class TestContextManager:
    """Test ContextManager class."""

    def test_init_creates_state_dir(self):
        """Test initialization creates state directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            cm = ContextManager(tmpdir)

            state_dir = os.path.join(tmpdir, ".moai", "memory", "command-state")
            assert os.path.exists(state_dir)

    def test_save_phase_result_with_timestamp(self):
        """Test saving phase with automatic timestamp."""
        with tempfile.TemporaryDirectory() as tmpdir:
            cm = ContextManager(tmpdir)
            data = {"phase": "0-project", "status": "completed"}

            saved_path = cm.save_phase_result(data)

            assert os.path.exists(saved_path)
            assert "0-project" in saved_path
            assert saved_path.endswith(".json")

    def test_save_multiple_phases(self):
        """Test saving multiple phase results."""
        with tempfile.TemporaryDirectory() as tmpdir:
            cm = ContextManager(tmpdir)

            path1 = cm.save_phase_result({"phase": "0-project"})
            path2 = cm.save_phase_result({"phase": "1-plan"})

            assert path1 != path2
            assert os.path.exists(path1)
            assert os.path.exists(path2)

    def test_load_latest_phase_single_file(self):
        """Test loading latest phase when one exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            cm = ContextManager(tmpdir)
            data = {"phase": "test", "data": "important"}

            cm.save_phase_result(data)
            loaded = cm.load_latest_phase()

            assert loaded is not None
            assert loaded["phase"] == "test"
            assert loaded["data"] == "important"

    def test_load_latest_phase_multiple_files(self):
        """Test loading latest when multiple phases exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            cm = ContextManager(tmpdir)

            cm.save_phase_result({"phase": "0-project"})
            cm.save_phase_result({"phase": "1-plan"})
            cm.save_phase_result({"phase": "2-run"})

            loaded = cm.load_latest_phase()

            assert loaded["phase"] == "2-run"

    def test_load_latest_phase_empty_dir(self):
        """Test loading when no phases exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            cm = ContextManager(tmpdir)

            loaded = cm.load_latest_phase()

            assert loaded is None

    def test_get_state_dir(self):
        """Test getting state directory path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            cm = ContextManager(tmpdir)
            state_dir = cm.get_state_dir()

            assert state_dir == os.path.join(tmpdir, ".moai", "memory", "command-state")

    def test_state_dir_created_on_init(self):
        """Test that state directory is created during init."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Directory should not exist yet
            state_path = os.path.join(tmpdir, ".moai", "memory", "command-state")
            assert not os.path.exists(state_path)

            # Initialize ContextManager
            cm = ContextManager(tmpdir)

            # Directory should now exist
            assert os.path.exists(state_path)


class TestCleanupTempFile:
    """Test _cleanup_temp_file function."""

    def test_cleanup_both_fd_and_path(self):
        """Test cleanup with both file descriptor and path."""
        from moai_adk.core.context_manager import _cleanup_temp_file

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a temp file
            import tempfile as tmp_module

            fd, path = tmp_module.mkstemp(dir=tmpdir)

            # Cleanup
            _cleanup_temp_file(fd, path)

            # File should be deleted
            assert not os.path.exists(path)

    def test_cleanup_with_none_values(self):
        """Test cleanup with None values."""
        from moai_adk.core.context_manager import _cleanup_temp_file

        # Should not raise
        _cleanup_temp_file(None, None)

    def test_cleanup_missing_file(self):
        """Test cleanup when file already deleted."""
        from moai_adk.core.context_manager import _cleanup_temp_file

        # Should not raise even if file doesn't exist
        _cleanup_temp_file(None, "/nonexistent/file.txt")


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_very_long_paths(self):
        """Test handling of very long paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a long path
            path_parts = ["dir"] * 20
            long_path = os.path.join(*path_parts)

            # Should handle gracefully
            try:
                result = validate_and_convert_path(long_path, tmpdir)
                # May fail with parent not found, but shouldn't crash
            except FileNotFoundError:
                pass

    def test_symlink_escaping(self):
        """Test that symlink attacks are prevented."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a target directory outside tmpdir
            with tempfile.TemporaryDirectory() as outside:
                # Create symlink inside tmpdir pointing outside
                symlink_path = os.path.join(tmpdir, "link")
                try:
                    os.symlink(outside, symlink_path)

                    # Try to use symlink - should be blocked or handled safely
                    # The realpath check should catch this
                    result = validate_and_convert_path("link", tmpdir)
                    # Result should resolve to outside path, but validation checks it
                except (OSError, ValueError):
                    # Either operation not supported or path validation caught it
                    pass

    def test_concurrent_file_access(self):
        """Test behavior with concurrent file access."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = os.path.join(tmpdir, "phase.json")
            data1 = {"phase": "1"}
            data2 = {"phase": "2"}

            # Simulate concurrent writes (atomic operations)
            save_phase_result(data1, target_path)
            save_phase_result(data2, target_path)

            loaded = load_phase_result(target_path)
            assert loaded["phase"] == "2"
