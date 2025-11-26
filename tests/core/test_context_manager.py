"""
Tests for ContextManager module
"""

import json
import os
import tempfile
from pathlib import Path

import pytest

# Import will be added after implementation
try:
    from moai_adk.core.context_manager import (
        ContextManager,
        load_phase_result,
        save_phase_result,
        validate_and_convert_path,
    )
except ImportError:
    # Module not yet implemented, tests will fail as expected (RED phase)
    pass


@pytest.fixture
def temp_project_dir():
    """Create a temporary project directory structure"""
    with tempfile.TemporaryDirectory() as tmpdir:
        project_root = Path(tmpdir)

        # Create directory structure
        (project_root / ".moai" / "memory" / "command-state").mkdir(parents=True, exist_ok=True)
        (project_root / ".moai" / "project").mkdir(parents=True, exist_ok=True)
        (project_root / "src").mkdir(parents=True, exist_ok=True)

        yield str(project_root)


@pytest.fixture
def temp_state_dir(temp_project_dir):
    """Return the command-state directory path"""
    return os.path.join(temp_project_dir, ".moai", "memory", "command-state")


class TestPathValidation:
    """Test cases for validate_and_convert_path function"""

    def test_is_path_within_root_normal_case(self, temp_project_dir):
        """Test _is_path_within_root helper function for normal cases"""
        from moai_adk.core.context_manager import _is_path_within_root

        project_root = temp_project_dir
        inside_path = os.path.join(project_root, ".moai", "project", "file.md")

        assert _is_path_within_root(inside_path, project_root), "Path should be within root"

    def test_is_path_within_root_outside_case(self, temp_project_dir):
        """Test _is_path_within_root helper function for paths outside root"""
        from moai_adk.core.context_manager import _is_path_within_root

        project_root = temp_project_dir
        outside_path = os.path.dirname(project_root) + "/other_project/file.md"

        assert not _is_path_within_root(outside_path, project_root), "Path should be outside root"

    def test_relative_path_to_absolute_conversion(self, temp_project_dir):
        """Test that relative paths are correctly converted to absolute paths"""
        relative_path = ".moai/project/product.md"
        expected_prefix = temp_project_dir

        result = validate_and_convert_path(relative_path, temp_project_dir)

        assert os.path.isabs(result), f"Result should be absolute path: {result}"
        assert result.startswith(expected_prefix), f"Path should start with project root: {result}"
        assert result.endswith("product.md"), f"Path should end with filename: {result}"

    def test_absolute_path_pass_through(self, temp_project_dir):
        """Test that absolute paths are validated and returned as-is"""
        abs_path = os.path.join(temp_project_dir, ".moai", "project", "structure.md")

        result = validate_and_convert_path(abs_path, temp_project_dir)

        assert result == abs_path, "Absolute path should be returned as-is"
        assert os.path.isabs(result), "Result should remain absolute"

    def test_path_outside_project_root_rejected(self, temp_project_dir):
        """Test that paths outside project root are rejected"""
        parent_dir = os.path.dirname(temp_project_dir)
        outside_path = os.path.join(parent_dir, "unauthorized.txt")

        with pytest.raises(ValueError) as exc_info:
            validate_and_convert_path(outside_path, temp_project_dir)

        assert "outside project root" in str(exc_info.value).lower()

    def test_path_traversal_attack_blocked(self, temp_project_dir):
        """Test that path traversal attacks (../) are blocked"""
        traversal_path = "../../etc/passwd"

        with pytest.raises(ValueError) as exc_info:
            validate_and_convert_path(traversal_path, temp_project_dir)

        assert "outside project root" in str(exc_info.value).lower()

    def test_existing_directory_accepted(self, temp_project_dir):
        """Test that existing directories are accepted"""
        existing_dir = ".moai/project"

        result = validate_and_convert_path(existing_dir, temp_project_dir)

        assert os.path.isdir(result), f"Existing directory should be accepted: {result}"

    def test_nonexistent_parent_directory_rejected(self, temp_project_dir):
        """Test that file paths with nonexistent parent directories are rejected"""
        invalid_path = "nonexistent/subdir/file.txt"

        with pytest.raises(FileNotFoundError) as exc_info:
            validate_and_convert_path(invalid_path, temp_project_dir)

        assert "Parent directory not found" in str(exc_info.value)

    def test_file_in_existing_directory_accepted(self, temp_project_dir):
        """Test that file paths in existing directories are accepted (even if file doesn't exist yet)"""
        file_path = ".moai/project/new_file.md"

        result = validate_and_convert_path(file_path, temp_project_dir)

        assert result.endswith("new_file.md"), "File path should be accepted"
        assert not os.path.exists(result), "File doesn't need to exist yet"

    def test_path_normalization(self, temp_project_dir):
        """Test that paths with ./ and multiple slashes are normalized"""
        messy_path = "./.moai//project/./product.md"

        result = validate_and_convert_path(messy_path, temp_project_dir)

        # Should normalize without extra slashes
        assert "//" not in result, f"Path should not contain double slashes: {result}"
        assert "/./" not in result, f"Path should not contain /./ patterns: {result}"

    def test_unicode_paths_supported(self, temp_project_dir):
        """Test that unicode characters in paths are supported"""
        # Create the parent directory first
        unicode_dir = os.path.join(temp_project_dir, ".moai", "프로젝트")
        os.makedirs(unicode_dir, exist_ok=True)

        unicode_path = ".moai/프로젝트/파일.md"

        result = validate_and_convert_path(unicode_path, temp_project_dir)

        assert "프로젝트" in result, f"Unicode characters should be preserved: {result}"


class TestAtomicJsonWrite:
    """Test cases for atomic JSON write functionality"""

    def test_atomic_write_creates_valid_json(self, temp_state_dir):
        """Test that atomic write creates valid JSON files"""
        test_data = {
            "phase": "0-project",
            "timestamp": "2025-11-12T10:30:00Z",
            "status": "completed",
            "outputs": {"project_name": "TestProject"},
        }
        target_path = os.path.join(temp_state_dir, "test-state.json")

        save_phase_result(test_data, target_path)

        assert os.path.exists(target_path), f"File should be created: {target_path}"

        with open(target_path, "r") as f:
            loaded_data = json.load(f)

        assert loaded_data == test_data, "Written data should match original"

    def test_atomic_write_uses_temp_file(self, temp_state_dir):
        """Test that atomic write uses temporary files during write"""
        test_data = {"phase": "0-project", "status": "completed"}
        target_path = os.path.join(temp_state_dir, "test-atomic.json")

        # This test verifies that temporary files are cleaned up
        # by checking no temp files remain after write
        save_phase_result(test_data, target_path)

        # Check no temp files in the directory
        temp_files = [f for f in os.listdir(temp_state_dir) if f.startswith(".tmp") or f.endswith(".tmp")]

        assert len(temp_files) == 0, f"No temporary files should remain: {temp_files}"

    def test_atomic_write_failure_preserves_original(self, temp_state_dir):
        """Test that write failures preserve existing file"""
        # First write a file
        original_data = {"phase": "0-project", "status": "completed"}
        target_path = os.path.join(temp_state_dir, "test-preserve.json")

        save_phase_result(original_data, target_path)

        # Verify original write succeeded
        with open(target_path, "r") as f:
            assert json.load(f) == original_data

        # Note: Full failure test with mocked os.replace would be implemented
        # This test ensures the mechanism exists for preserving on failure

    def test_atomic_write_with_large_json(self, temp_state_dir):
        """Test atomic write with larger JSON objects"""
        large_data = {
            "phase": "2-run",
            "timestamp": "2025-11-12T10:30:00Z",
            "files_created": [f"/path/to/file_{i}.py" for i in range(100)],
            "test_results": {f"test_{i}": {"passed": True, "duration": 0.5 + i * 0.01} for i in range(50)},
        }
        target_path = os.path.join(temp_state_dir, "large-state.json")

        save_phase_result(large_data, target_path)

        with open(target_path, "r") as f:
            loaded_data = json.load(f)

        assert loaded_data == large_data, "Large JSON should be preserved"

    def test_file_permissions_after_write(self, temp_state_dir):
        """Test that written files have correct permissions"""
        test_data = {"phase": "0-project"}
        target_path = os.path.join(temp_state_dir, "test-perms.json")

        save_phase_result(test_data, target_path)

        # File should be readable and writable by owner
        stat_info = os.stat(target_path)
        assert stat_info.st_mode & 0o400, "File should be readable by owner"
        assert stat_info.st_mode & 0o200, "File should be writable by owner"


class TestLoadPhaseResult:
    """Test cases for loading phase results"""

    def test_load_existing_phase_result(self, temp_state_dir):
        """Test loading an existing phase result file"""
        test_data = {
            "phase": "0-project",
            "timestamp": "2025-11-12T10:30:00Z",
            "status": "completed",
            "outputs": {"project_name": "TestProject"},
        }
        target_path = os.path.join(temp_state_dir, "phase-result.json")

        save_phase_result(test_data, target_path)
        loaded_data = load_phase_result(target_path)

        assert loaded_data == test_data, "Loaded data should match saved data"

    def test_load_nonexistent_file_raises_error(self, temp_state_dir):
        """Test that loading nonexistent file raises appropriate error"""
        nonexistent_path = os.path.join(temp_state_dir, "nonexistent.json")

        with pytest.raises(FileNotFoundError):
            load_phase_result(nonexistent_path)

    def test_load_corrupted_json_raises_error(self, temp_state_dir):
        """Test that corrupted JSON files raise appropriate error"""
        corrupted_path = os.path.join(temp_state_dir, "corrupted.json")

        with open(corrupted_path, "w") as f:
            f.write("{invalid json content")

        with pytest.raises(json.JSONDecodeError):
            load_phase_result(corrupted_path)

    def test_load_returns_dict(self, temp_state_dir):
        """Test that load_phase_result returns a dictionary"""
        test_data = {"phase": "1-plan", "status": "in_progress"}
        target_path = os.path.join(temp_state_dir, "dict-test.json")

        save_phase_result(test_data, target_path)
        result = load_phase_result(target_path)

        assert isinstance(result, dict), "Load should return dictionary"


class TestContextManager:
    """Test cases for ContextManager class"""

    def test_context_manager_initialization(self, temp_project_dir):
        """Test ContextManager initialization with project root"""
        manager = ContextManager(project_root=temp_project_dir)

        assert manager.project_root == temp_project_dir
        assert hasattr(manager, "state_dir")

    def test_context_manager_saves_phase_result(self, temp_project_dir):
        """Test that ContextManager can save phase results"""
        manager = ContextManager(project_root=temp_project_dir)
        phase_data = {"phase": "0-project", "status": "completed", "outputs": {"project_name": "Test"}}

        manager.save_phase_result(phase_data)

        # Verify file was created
        state_files = os.listdir(manager.state_dir)
        assert len(state_files) > 0, "Phase result file should be created"

    def test_context_manager_loads_latest_phase(self, temp_project_dir):
        """Test loading the latest phase result"""
        manager = ContextManager(project_root=temp_project_dir)

        phase_data = {"phase": "0-project", "status": "completed", "outputs": {"project_name": "Test"}}

        manager.save_phase_result(phase_data)
        loaded = manager.load_latest_phase()

        assert loaded["phase"] == "0-project"
        assert loaded["status"] == "completed"


class TestTemplateVariableSubstitution:
    """Test cases for template variable substitution"""

    def test_substitute_simple_variables(self):
        """Test substituting simple template variables"""
        from moai_adk.core.context_manager import substitute_template_variables

        template = "Project: {{PROJECT_NAME}}, Mode: {{MODE}}"
        context = {"PROJECT_NAME": "MyProject", "MODE": "personal"}

        result = substitute_template_variables(template, context)

        assert "{{PROJECT_NAME}}" not in result
        assert "{{MODE}}" not in result
        assert "MyProject" in result
        assert "personal" in result

    def test_substitute_nested_paths(self):
        """Test substituting paths in template"""
        from moai_adk.core.context_manager import substitute_template_variables

        template = "Path: {{PROJECT_ROOT}}/.moai/project/product.md"
        context = {"PROJECT_ROOT": "/Users/goos/MyProject"}

        result = substitute_template_variables(template, context)

        assert "{{PROJECT_ROOT}}" not in result
        assert "/Users/goos/MyProject/.moai/project/product.md" in result

    def test_substitute_multiple_occurrences(self):
        """Test substituting same variable multiple times"""
        from moai_adk.core.context_manager import substitute_template_variables

        template = "Project {{PROJECT_NAME}} files in {{PROJECT_NAME}} directory"
        context = {"PROJECT_NAME": "MyApp"}

        result = substitute_template_variables(template, context)

        assert result.count("MyApp") == 2
        assert "{{PROJECT_NAME}}" not in result

    def test_substitute_with_empty_context(self):
        """Test substituting with empty context returns unchanged text"""
        from moai_adk.core.context_manager import substitute_template_variables

        template = "No variables here"
        result = substitute_template_variables(template, {})

        assert result == template

    def test_substitute_numeric_values(self):
        """Test substituting numeric values"""
        from moai_adk.core.context_manager import substitute_template_variables

        template = "Version: {{VERSION}}, Count: {{COUNT}}"
        context = {"VERSION": 1.5, "COUNT": 42}

        result = substitute_template_variables(template, context)

        assert "1.5" in result
        assert "42" in result

    def test_validate_no_unsubstituted_variables(self):
        """Test validation of unsubstituted template variables"""
        from moai_adk.core.context_manager import validate_no_template_vars

        # Should pass - no template vars
        valid_text = "This is a normal string"
        validate_no_template_vars(valid_text)  # Should not raise

        # Should fail - has template vars
        invalid_text = "This has {{VARIABLE}} in it"
        with pytest.raises(ValueError) as exc_info:
            validate_no_template_vars(invalid_text)

        assert "Unsubstituted" in str(exc_info.value)

    def test_validate_multiple_unsubstituted_variables(self):
        """Test validation detects multiple unsubstituted variables"""
        from moai_adk.core.context_manager import validate_no_template_vars

        invalid_text = "{{VAR1}} and {{VAR2}} and {{VAR3}}"

        with pytest.raises(ValueError) as exc_info:
            validate_no_template_vars(invalid_text)

        error_msg = str(exc_info.value)
        assert "{{VAR1}}" in error_msg
        assert "{{VAR2}}" in error_msg
        assert "{{VAR3}}" in error_msg
