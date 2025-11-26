"""
Integration tests for Command Context Saving functionality

Tests the integration between 0-project command and ContextManager.
Verifies that phase results are correctly saved after project-manager execution.
"""

import json
import os
import tempfile
from datetime import datetime, timezone
from pathlib import Path

import pytest

# Import ContextManager
try:
    from moai_adk.core.context_manager import (
        ContextManager,
        substitute_template_variables,
        validate_and_convert_path,
        validate_no_template_vars,
    )
except ImportError:
    pytest.skip("ContextManager not available", allow_module_level=True)


@pytest.fixture
def mock_project_dir():
    """Create a mock project directory structure for testing"""
    with tempfile.TemporaryDirectory() as tmpdir:
        project_root = Path(tmpdir)

        # Create MoAI directory structure
        (project_root / ".moai" / "memory" / "command-state").mkdir(parents=True)
        (project_root / ".moai" / "project").mkdir(parents=True)
        (project_root / ".moai" / "config").mkdir(parents=True)
        (project_root / "src").mkdir(parents=True)

        # Create config.json
        config = {
            "project": {"name": "TestProject", "mode": "personal", "owner": "@testuser"},
            "language": {"conversation_language": "en"},
        }
        config_path = project_root / ".moai" / "config" / "config.json"
        with open(config_path, "w") as f:
            json.dump(config, f, indent=2)

        yield str(project_root)


@pytest.fixture
def context_manager(mock_project_dir):
    """Initialize ContextManager for testing"""
    return ContextManager(mock_project_dir)


class TestProjectCommandContextSaving:
    """Test suite for 0-project command context saving integration"""

    def test_0_project_saves_phase_result(self, context_manager, mock_project_dir):
        """
        Test that 0-project command saves phase result to JSON file.

        WHEN project-manager agent completes successfully
        THEN a JSON file should be created in .moai/memory/command-state/
        """
        # Arrange: Mock phase result data
        phase_data = {
            "phase": "0-project",
            "timestamp": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
            "status": "completed",
            "outputs": {"project_name": "TestProject", "mode": "personal", "language": "en", "tech_stack": ["python"]},
            "files_created": [
                os.path.join(mock_project_dir, ".moai/project/product.md"),
                os.path.join(mock_project_dir, ".moai/project/structure.md"),
            ],
            "next_phase": "1-plan",
        }

        # Act: Save phase result
        saved_path = context_manager.save_phase_result(phase_data)

        # Assert: File exists and is in correct location
        assert os.path.exists(saved_path), f"Phase result file should exist: {saved_path}"
        assert saved_path.startswith(context_manager.get_state_dir()), "File should be in command-state directory"
        assert saved_path.endswith(".json"), "File should have .json extension"

        # Verify file content
        with open(saved_path, "r") as f:
            loaded_data = json.load(f)

        assert loaded_data["phase"] == "0-project", "Phase name should match"
        assert loaded_data["status"] == "completed", "Status should be completed"
        assert loaded_data["next_phase"] == "1-plan", "Next phase should be set"

    def test_0_project_json_schema(self, context_manager, mock_project_dir):
        """
        Test that saved JSON complies with expected schema.

        WHEN phase result is saved
        THEN JSON should contain all required fields:
        - phase (string)
        - timestamp (ISO 8601 format)
        - status (string: completed/error/interrupted)
        - outputs (dict)
        - files_created (list of absolute paths)
        - next_phase (string)
        """
        # Arrange
        phase_data = {
            "phase": "0-project",
            "timestamp": "2025-11-12T10:30:00Z",
            "status": "completed",
            "outputs": {"project_name": "TestProject", "mode": "personal"},
            "files_created": [os.path.join(mock_project_dir, ".moai/project/product.md")],
            "next_phase": "1-plan",
        }

        # Act
        saved_path = context_manager.save_phase_result(phase_data)

        # Assert: Load and validate schema
        with open(saved_path, "r") as f:
            loaded_data = json.load(f)

        # Required fields
        assert "phase" in loaded_data, "Schema missing 'phase' field"
        assert "timestamp" in loaded_data, "Schema missing 'timestamp' field"
        assert "status" in loaded_data, "Schema missing 'status' field"
        assert "outputs" in loaded_data, "Schema missing 'outputs' field"
        assert "files_created" in loaded_data, "Schema missing 'files_created' field"
        assert "next_phase" in loaded_data, "Schema missing 'next_phase' field"

        # Field types
        assert isinstance(loaded_data["phase"], str), "phase should be string"
        assert isinstance(loaded_data["timestamp"], str), "timestamp should be string"
        assert isinstance(loaded_data["status"], str), "status should be string"
        assert isinstance(loaded_data["outputs"], dict), "outputs should be dict"
        assert isinstance(loaded_data["files_created"], list), "files_created should be list"
        assert isinstance(loaded_data["next_phase"], str), "next_phase should be string"

        # Timestamp format (ISO 8601)
        assert loaded_data["timestamp"].endswith("Z"), "timestamp should be UTC (end with Z)"

        # Status values
        assert loaded_data["status"] in [
            "completed",
            "error",
            "interrupted",
        ], "status should be one of: completed, error, interrupted"

    def test_0_project_paths_are_absolute(self, context_manager, mock_project_dir):
        """
        Test that all file paths in saved JSON are absolute.

        WHEN files_created contains paths
        THEN all paths must be absolute (start with /)
        AND all paths must be within project root
        """
        # Arrange: Mix of relative and absolute paths (should all be converted)
        relative_paths = [".moai/project/product.md", ".moai/project/structure.md", "src/main.py"]

        # Convert to absolute using validate_and_convert_path
        absolute_paths = [validate_and_convert_path(rel_path, mock_project_dir) for rel_path in relative_paths]

        phase_data = {
            "phase": "0-project",
            "timestamp": "2025-11-12T10:30:00Z",
            "status": "completed",
            "outputs": {"project_name": "TestProject"},
            "files_created": absolute_paths,
            "next_phase": "1-plan",
        }

        # Act
        saved_path = context_manager.save_phase_result(phase_data)

        # Assert: All paths are absolute
        with open(saved_path, "r") as f:
            loaded_data = json.load(f)

        for file_path in loaded_data["files_created"]:
            assert os.path.isabs(file_path), f"Path should be absolute: {file_path}"
            assert file_path.startswith(mock_project_dir), f"Path should be within project root: {file_path}"

    def test_0_project_no_template_vars(self, context_manager, mock_project_dir):
        """
        Test that saved JSON contains no unsubstituted template variables.

        WHEN phase result is saved
        THEN no {{VARIABLE}} patterns should exist in outputs
        """
        # Arrange: Data with template variables (should be substituted before saving)
        raw_data = {
            "phase": "0-project",
            "timestamp": "2025-11-12T10:30:00Z",
            "status": "completed",
            "outputs": {
                "project_name": "{{PROJECT_NAME}}",  # Should be substituted
                "mode": "{{MODE}}",  # Should be substituted
                "language": "en",
            },
            "files_created": ["{{PROJECT_ROOT}}/.moai/project/product.md"],  # Should be substituted
            "next_phase": "1-plan",
        }

        # Substitute template variables
        context = {"PROJECT_NAME": "TestProject", "MODE": "personal", "PROJECT_ROOT": mock_project_dir}

        # Convert outputs to JSON string, substitute, then parse back
        outputs_str = json.dumps(raw_data["outputs"])
        outputs_substituted = substitute_template_variables(outputs_str, context)
        raw_data["outputs"] = json.loads(outputs_substituted)

        # Substitute files_created
        raw_data["files_created"] = [substitute_template_variables(path, context) for path in raw_data["files_created"]]

        # Act
        saved_path = context_manager.save_phase_result(raw_data)

        # Assert: No template variables in saved JSON
        with open(saved_path, "r") as f:
            content = f.read()

        # Validate no template variables exist
        try:
            validate_no_template_vars(content)
        except ValueError as e:
            pytest.fail(f"Template variables found in saved JSON: {e}")

        # Parse and verify substituted values
        loaded_data = json.loads(content)
        assert loaded_data["outputs"]["project_name"] == "TestProject", "PROJECT_NAME should be substituted"
        assert loaded_data["outputs"]["mode"] == "personal", "MODE should be substituted"
        assert mock_project_dir in loaded_data["files_created"][0], "PROJECT_ROOT should be substituted in file paths"

    # Additional test: Error handling for graceful failure
    def test_0_project_save_failure_graceful(self, context_manager, mock_project_dir):
        """
        Test graceful error handling when context save fails.

        WHEN save_phase_result fails (e.g., permission error)
        THEN error should be raised but not block command completion
        """
        # Arrange: Invalid target directory (should fail)
        invalid_data = {
            "phase": "0-project",
            "timestamp": "2025-11-12T10:30:00Z",
            "status": "completed",
            "outputs": {},
            "files_created": [],
            "next_phase": "1-plan",
        }

        # Make state_dir read-only to trigger error
        state_dir = context_manager.get_state_dir()
        original_mode = os.stat(state_dir).st_mode

        try:
            os.chmod(state_dir, 0o444)  # Read-only

            # Act & Assert: Should raise IOError
            with pytest.raises(IOError) as exc_info:
                context_manager.save_phase_result(invalid_data)

            assert "Failed to write" in str(exc_info.value), "Error message should indicate write failure"

        finally:
            # Cleanup: Restore permissions
            os.chmod(state_dir, original_mode)

    # Integration test: Load latest phase result
    def test_load_latest_phase_result(self, context_manager, mock_project_dir):
        """
        Test loading the most recent phase result.

        WHEN multiple phase results exist
        THEN load_latest_phase should return the most recent one
        """
        # Arrange: Save multiple phase results
        phase1 = {
            "phase": "0-project",
            "timestamp": "2025-11-12T10:00:00Z",
            "status": "completed",
            "outputs": {"version": 1},
            "files_created": [],
            "next_phase": "1-plan",
        }

        phase2 = {
            "phase": "1-plan",
            "timestamp": "2025-11-12T11:00:00Z",
            "status": "completed",
            "outputs": {"version": 2},
            "files_created": [],
            "next_phase": "2-run",
        }

        # Act: Save both phases
        context_manager.save_phase_result(phase1)
        import time

        time.sleep(0.1)  # Ensure different timestamps
        context_manager.save_phase_result(phase2)

        # Load latest
        latest = context_manager.load_latest_phase()

        # Assert: Latest should be phase2
        assert latest is not None, "Should load latest phase"
        assert latest["phase"] == "1-plan", "Latest phase should be 1-plan"
        assert latest["outputs"]["version"] == 2, "Should load most recent version"
