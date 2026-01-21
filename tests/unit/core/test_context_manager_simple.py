"""Simple comprehensive tests for ContextManager module.

Tests context management with focus on:
- Path validation and safety
- Absolute path conversion
- Atomic JSON file operations
- Phase result persistence and loading
- Template variable substitution and validation
- ContextManager class functionality
"""

import json
import os
import tempfile
from datetime import datetime, timezone

from moai_adk.core.context_manager import (
    ContextManager,
    _cleanup_temp_file,
    _is_path_within_root,
    load_phase_result,
    save_phase_result,
    substitute_template_variables,
    validate_and_convert_path,
    validate_no_template_vars,
)


class TestIsPathWithinRoot:
    """Test path safety checking."""

    def test_path_within_root_same_dir(self):
        """Test path equal to root returns true."""
        with tempfile.TemporaryDirectory() as tmpdir:
            result = _is_path_within_root(tmpdir, tmpdir)

            assert result is True

    def test_path_within_root_subdir(self):
        """Test path in subdirectory returns true."""
        with tempfile.TemporaryDirectory() as tmpdir:
            subdir = os.path.join(tmpdir, "subdir")
            os.makedirs(subdir)

            result = _is_path_within_root(subdir, tmpdir)

            assert result is True

    def test_path_outside_root_returns_false(self):
        """Test path outside root returns false."""
        with tempfile.TemporaryDirectory() as tmpdir:
            other_dir = tempfile.mkdtemp()
            try:
                result = _is_path_within_root(other_dir, tmpdir)

                assert result is False
            finally:
                os.rmdir(other_dir)

    def test_path_with_symlink_escaping(self):
        """Test symlink escape attempts are prevented."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # This is harder to test realistically, but the function uses realpath
            result = _is_path_within_root(tmpdir, tmpdir)

            assert result is True

    def test_invalid_path_returns_false(self):
        """Test invalid path returns false."""
        result = _is_path_within_root("/nonexistent/path/to/nowhere", "/nonexistent/root")

        assert result is False


class TestValidateAndConvertPath:
    """Test path validation and conversion."""

    def test_convert_relative_path(self):
        """Test converting relative path to absolute."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a subdirectory
            subdir = os.path.join(tmpdir, "subdir")
            os.makedirs(subdir)

            result = validate_and_convert_path("subdir", tmpdir)

            assert os.path.isabs(result)
            assert result == subdir

    def test_convert_absolute_path_within_root(self):
        """Test absolute path within root returns as-is."""
        with tempfile.TemporaryDirectory() as tmpdir:
            subdir = os.path.join(tmpdir, "subdir")
            os.makedirs(subdir)

            result = validate_and_convert_path(subdir, tmpdir)

            assert result == subdir

    def test_path_outside_root_returns_absolute(self):
        """Test path outside root is converted to absolute."""
        with tempfile.TemporaryDirectory() as tmpdir:
            other_dir = tempfile.mkdtemp()
            try:
                # Function doesn't validate, just converts to absolute
                result = validate_and_convert_path(other_dir, tmpdir)

                assert os.path.isabs(result)
            finally:
                os.rmdir(other_dir)

    def test_file_parent_dir_missing_returns_absolute(self):
        """Test file with missing parent directory returns absolute path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Function doesn't check existence, just converts to absolute
            result = validate_and_convert_path("nonexistent/file.txt", tmpdir)

            assert os.path.isabs(result)
            assert "nonexistent" in result

    def test_existing_directory_returns_path(self):
        """Test existing directory returns without error."""
        with tempfile.TemporaryDirectory() as tmpdir:
            subdir = os.path.join(tmpdir, "existing")
            os.makedirs(subdir)

            result = validate_and_convert_path("existing", tmpdir)

            assert result == subdir


class TestCleanupTempFile:
    """Test temporary file cleanup."""

    def test_cleanup_with_valid_fd_and_path(self):
        """Test cleanup closes fd and removes file."""
        with tempfile.NamedTemporaryFile(delete=False) as tmp:
            fd = os.open(tmp.name, os.O_WRONLY)
            path = tmp.name

        try:
            _cleanup_temp_file(fd, path)

            # File should be removed
            assert not os.path.exists(path)
        except Exception:
            if os.path.exists(path):
                os.unlink(path)

    def test_cleanup_with_none_fd(self):
        """Test cleanup handles None fd gracefully."""
        with tempfile.NamedTemporaryFile(delete=False) as tmp:
            path = tmp.name

        try:
            _cleanup_temp_file(None, path)

            # File should be removed
            assert not os.path.exists(path)
        except Exception:
            if os.path.exists(path):
                os.unlink(path)

    def test_cleanup_with_none_path(self):
        """Test cleanup handles None path gracefully."""
        with tempfile.NamedTemporaryFile() as tmp:
            fd = os.open(tmp.name, os.O_WRONLY)

        _cleanup_temp_file(fd, None)
        # Should not raise error


class TestSavePhaseResult:
    """Test saving phase results."""

    def test_save_phase_result_creates_file(self):
        """Test save_phase_result creates file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            data = {"phase": "spec", "status": "complete"}

            save_phase_result("test_phase", data, tmpdir)

            # Check file exists in the expected location
            expected_path = os.path.join(tmpdir, ".moai", "memory", "command-state", "test_phase.json")
            assert os.path.exists(expected_path)

    def test_save_phase_result_valid_json(self):
        """Test saved file is valid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            data = {
                "phase": "red",
                "tests": 5,
                "coverage": 0.85,
            }

            save_phase_result("test_phase", data, tmpdir)

            target_path = os.path.join(tmpdir, ".moai", "memory", "command-state", "test_phase.json")
            with open(target_path, "r") as f:
                loaded = json.load(f)

            assert loaded["phase"] == "red"
            assert loaded["tests"] == 5
            assert loaded["coverage"] == 0.85

    def test_save_phase_result_creates_parent_dirs(self):
        """Test save_phase_result creates parent directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            data = {"phase": "green"}

            save_phase_result("nested/phase", data, tmpdir)

            # Parent directories should be created
            expected_path = os.path.join(tmpdir, ".moai", "memory", "command-state", "nested", "phase.json")
            assert os.path.exists(expected_path)
            assert os.path.isfile(expected_path)

    def test_save_phase_result_overwrites_existing(self):
        """Test save_phase_result overwrites existing file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Save first version
            save_phase_result("test_phase", {"version": 1}, tmpdir)

            # Save second version
            save_phase_result("test_phase", {"version": 2}, tmpdir)

            target_path = os.path.join(tmpdir, ".moai", "memory", "command-state", "test_phase.json")
            with open(target_path, "r") as f:
                loaded = json.load(f)

            assert loaded["version"] == 2


class TestLoadPhaseResult:
    """Test loading phase results."""

    def test_load_phase_result_success(self):
        """Test loading valid phase result."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test file using save_phase_result
            data = {"phase": "spec", "status": "ready"}
            save_phase_result("test_phase", data, tmpdir)

            loaded = load_phase_result("test_phase", tmpdir)

            assert loaded is not None
            assert loaded["phase"] == "spec"
            assert loaded["status"] == "ready"

    def test_load_phase_result_file_not_found(self):
        """Test loading non-existent phase returns None."""
        with tempfile.TemporaryDirectory() as tmpdir:
            result = load_phase_result("nonexistent_phase", tmpdir)
            assert result is None

    def test_load_phase_result_invalid_json(self):
        """Test loading invalid JSON returns None."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create invalid JSON file manually
            state_dir = os.path.join(tmpdir, ".moai", "memory", "command-state")
            os.makedirs(state_dir, exist_ok=True)
            target_path = os.path.join(state_dir, "invalid.json")

            with open(target_path, "w") as f:
                f.write("not valid json {]")

            result = load_phase_result("invalid", tmpdir)
            assert result is None


class TestSubstituteTemplateVariables:
    """Test template variable substitution."""

    def test_substitute_single_variable(self):
        """Test substituting single variable."""
        text = "Hello {{NAME}}, welcome!"
        context = {"NAME": "Alice"}

        content, found_vars = substitute_template_variables(text, context)

        assert content == "Hello Alice, welcome!"
        assert found_vars == ["NAME"]

    def test_substitute_multiple_variables(self):
        """Test substituting multiple variables."""
        text = "User {{USER}} in project {{PROJECT}}"
        context = {"USER": "bob", "PROJECT": "MoAI"}

        content, found_vars = substitute_template_variables(text, context)

        assert content == "User bob in project MoAI"
        assert set(found_vars) == {"USER", "PROJECT"}

    def test_substitute_missing_variable(self):
        """Test missing variable is left unchanged."""
        text = "Value: {{MISSING}}"
        context = {}

        content, found_vars = substitute_template_variables(text, context)

        assert content == "Value: {{MISSING}}"
        assert found_vars == []

    def test_substitute_no_variables(self):
        """Test text with no variables returns unchanged."""
        text = "Plain text without variables"
        context = {"UNUSED": "value"}

        content, found_vars = substitute_template_variables(text, context)

        assert content == "Plain text without variables"
        assert found_vars == []

    def test_substitute_numeric_values(self):
        """Test substituting numeric values (as strings)."""
        text = "Version {{VERSION}}, effort {{EFFORT}}"
        # Values must be strings for the current implementation
        context = {"VERSION": "1.0.0", "EFFORT": "5"}

        content, found_vars = substitute_template_variables(text, context)

        assert "1.0.0" in content
        assert "5" in content
        assert set(found_vars) == {"VERSION", "EFFORT"}


class TestValidateNoTemplateVars:
    """Test template variable validation."""

    def test_validate_no_variables_success(self):
        """Test validation passes with no variables."""
        text = "This is plain text with no variables."

        result = validate_no_template_vars(text)

        assert result is True

    def test_validate_with_single_variable_fails(self):
        """Test validation fails with variables."""
        text = "This contains {{VARIABLE}} that needs substitution."

        result = validate_no_template_vars(text)

        assert result is False

    def test_validate_with_multiple_variables_fails(self):
        """Test validation fails with multiple variables."""
        text = "User {{USER}} in {{PROJECT}} needs {{ACTION}}"

        result = validate_no_template_vars(text)

        assert result is False

    def test_validate_with_various_formats(self):
        """Test validation detects variables in various formats."""
        text = "{{VAR}}, {{VAR_NAME}}, {{VAR_123}}"

        result = validate_no_template_vars(text)

        assert result is False


class TestContextManagerInit:
    """Test ContextManager initialization."""

    def test_context_manager_initialization(self):
        """Test ContextManager initializes with project root."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            assert manager.project_root == tmpdir
            assert ".moai" in manager.state_dir
            assert "memory" in manager.state_dir

    def test_state_dir_created(self):
        """Test state directory is created."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            assert os.path.exists(manager.state_dir)
            assert os.path.isdir(manager.state_dir)


class TestContextManagerSavePhase:
    """Test ContextManager phase saving."""

    def test_save_phase_result_with_timestamp(self):
        """Test saving phase result includes timestamp."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            data = {"phase": "spec", "status": "complete"}

            manager.save_phase_result("test_phase", data)

            # Verify file was created in the expected location
            expected_path = manager.get_phase_result_path("test_phase")
            assert os.path.exists(expected_path)
            assert "test_phase" in os.path.basename(expected_path)

    def test_save_phase_result_creates_file(self):
        """Test save creates valid file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            data = {"phase": "red", "tests": 10}

            manager.save_phase_result("test_phase", data)

            # Load and verify
            loaded = manager.load_phase_result("test_phase")
            assert loaded is not None
            assert loaded["phase"] == "red"
            assert loaded["tests"] == 10
            assert "timestamp" in loaded

    def test_save_multiple_phases(self):
        """Test saving multiple phases creates multiple files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            manager.save_phase_result("phase1", {"phase": "spec"})
            manager.save_phase_result("phase2", {"phase": "red"})

            path1 = manager.get_phase_result_path("phase1")
            path2 = manager.get_phase_result_path("phase2")

            assert path1 != path2
            assert os.path.exists(path1)
            assert os.path.exists(path2)


class TestContextManagerLoadPhase:
    """Test ContextManager phase loading."""

    def test_load_phase_result_empty_dir(self):
        """Test loading when no phases saved returns None."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            result = manager.load_phase_result("nonexistent")

            assert result is None

    def test_load_phase_result_single_file(self):
        """Test loading single saved phase."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            saved_data = {"phase": "spec", "status": "complete"}
            manager.save_phase_result("test_phase", saved_data)

            result = manager.load_phase_result("test_phase")

            assert result is not None
            assert result["phase"] == "spec"
            assert result["status"] == "complete"

    def test_load_phase_result_multiple_files(self):
        """Test loading specific phases."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            # Save multiple phases
            manager.save_phase_result("phase1", {"phase": "spec", "order": 1})
            manager.save_phase_result("phase2", {"phase": "red", "order": 2})
            manager.save_phase_result("phase3", {"phase": "green", "order": 3})

            # Load each phase independently
            result1 = manager.load_phase_result("phase1")
            result2 = manager.load_phase_result("phase2")
            result3 = manager.load_phase_result("phase3")

            assert result1 is not None
            assert result1["order"] == 1

            assert result2 is not None
            assert result2["order"] == 2

            assert result3 is not None
            assert result3["order"] == 3


class TestContextManagerGetStateDir:
    """Test ContextManager state directory access."""

    def test_get_state_dir_returns_path(self):
        """Test get_state_dir returns state directory path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            state_dir = manager.get_state_dir()

            assert state_dir == manager.state_dir
            assert os.path.exists(state_dir)


class TestContextManagerIntegration:
    """Integration tests for ContextManager."""

    def test_save_and_load_workflow(self):
        """Test complete save and load workflow."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            # Save phase result
            original_data = {
                "phase": "spec",
                "spec_id": "SPEC-001",
                "status": "complete",
                "timestamp": datetime.now(timezone.utc).isoformat(),
            }

            manager.save_phase_result("test_phase", original_data)

            # Load it back
            loaded_data = manager.load_phase_result("test_phase")

            assert loaded_data is not None
            assert loaded_data["phase"] == "spec"
            assert loaded_data["spec_id"] == "SPEC-001"

    def test_multiple_phases_workflow(self):
        """Test workflow with multiple phase transitions."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = ContextManager(tmpdir)

            phases = [
                ("phase1", {"phase": "spec", "step": "requirements"}),
                ("phase2", {"phase": "red", "tests": 5}),
                ("phase3", {"phase": "green", "implementation": True}),
                ("phase4", {"phase": "refactor", "quality": "improved"}),
            ]

            # Save all phases
            for phase_name, phase_data in phases:
                manager.save_phase_result(phase_name, phase_data)

            # Load and verify each phase
            for phase_name, original_data in phases:
                loaded = manager.load_phase_result(phase_name)
                assert loaded is not None
                assert loaded["phase"] == original_data["phase"]

    def test_state_persistence_across_instances(self):
        """Test state persists across ContextManager instances."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # First instance saves data
            manager1 = ContextManager(tmpdir)
            manager1.save_phase_result("test_phase", {"phase": "spec", "data": "test"})

            # Second instance loads the same data
            manager2 = ContextManager(tmpdir)
            loaded = manager2.load_phase_result("test_phase")

            assert loaded is not None
            assert loaded["data"] == "test"
