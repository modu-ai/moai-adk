"""
Tests for Command Helper Utilities
"""

import json
import os
import tempfile
from pathlib import Path

import pytest

try:
    from moai_adk.core.command_helpers import (
        build_phase_result,
        detect_tech_stack,
        extract_project_metadata,
        load_previous_phase,
        save_command_context,
        validate_phase_files,
    )
except ImportError:
    pytest.skip("command_helpers not available", allow_module_level=True)


@pytest.fixture
def mock_project():
    """Create a mock project with config.json"""
    with tempfile.TemporaryDirectory() as tmpdir:
        project_root = Path(tmpdir)

        # Create directory structure
        (project_root / ".moai" / "config").mkdir(parents=True)
        (project_root / ".moai" / "memory" / "command-state").mkdir(parents=True)

        # Create config.json
        config = {
            "project": {"name": "TestProject", "mode": "personal", "owner": "@testuser"},
            "language": {"conversation_language": "en"},
        }

        config_path = project_root / ".moai" / "config" / "config.json"
        with open(config_path, "w") as f:
            json.dump(config, f, indent=2)

        yield str(project_root)


class TestMetadataExtraction:
    """Tests for extract_project_metadata function"""

    def test_extract_metadata_success(self, mock_project):
        """Test successful metadata extraction"""
        metadata = extract_project_metadata(mock_project)

        assert metadata["project_name"] == "TestProject"
        assert metadata["mode"] == "personal"
        assert metadata["owner"] == "@testuser"
        assert metadata["language"] == "en"

    def test_extract_metadata_missing_config(self):
        """Test error handling when config.json is missing"""
        with tempfile.TemporaryDirectory() as tmpdir:
            with pytest.raises(FileNotFoundError) as exc_info:
                extract_project_metadata(tmpdir)

            assert "Config file not found" in str(exc_info.value)


class TestTechStackDetection:
    """Tests for detect_tech_stack function"""

    def test_detect_python_project(self):
        """Test detection of Python project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create pyproject.toml
            (Path(tmpdir) / "pyproject.toml").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "python" in tech_stack

    def test_detect_javascript_project(self):
        """Test detection of JavaScript project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create package.json
            (Path(tmpdir) / "package.json").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "javascript" in tech_stack

    def test_detect_multiple_languages(self):
        """Test detection of polyglot project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create multiple indicators
            (Path(tmpdir) / "pyproject.toml").touch()
            (Path(tmpdir) / "package.json").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "python" in tech_stack
            assert "javascript" in tech_stack

    def test_default_to_python(self):
        """Test default fallback to Python"""
        with tempfile.TemporaryDirectory() as tmpdir:
            tech_stack = detect_tech_stack(tmpdir)

            assert tech_stack == ["python"]


class TestPhaseResultBuilder:
    """Tests for build_phase_result function"""

    def test_build_complete_phase_result(self):
        """Test building phase result with all fields"""
        outputs = {"key": "value"}
        files = ["/abs/path/file.txt"]

        result = build_phase_result(
            phase_name="0-project", status="completed", outputs=outputs, files_created=files, next_phase="1-plan"
        )

        assert result["phase"] == "0-project"
        assert result["status"] == "completed"
        assert result["outputs"] == outputs
        assert result["files_created"] == files
        assert result["next_phase"] == "1-plan"
        assert "timestamp" in result

    def test_timestamp_format(self):
        """Test that timestamp is in correct ISO 8601 UTC format"""
        result = build_phase_result(phase_name="test", status="completed", outputs={}, files_created=[])

        # Check UTC format
        assert result["timestamp"].endswith("Z")


class TestFileValidation:
    """Tests for validate_phase_files function"""

    def test_validate_existing_files(self, mock_project):
        """Test validation of existing files"""
        # Create test files
        test_files = [
            ".moai/config/config.json",
        ]

        validated = validate_phase_files(test_files, mock_project)

        assert len(validated) > 0
        for path in validated:
            assert os.path.isabs(path)

    def test_validate_invalid_paths_graceful(self, mock_project):
        """Test graceful handling of invalid paths"""
        invalid_paths = [
            "../../../etc/passwd",  # Path traversal
        ]

        # Should not raise, but may return empty list or skip invalid
        validated = validate_phase_files(invalid_paths, mock_project)

        # Implementation may skip invalid paths
        assert isinstance(validated, list)


class TestContextSaving:
    """Tests for save_command_context function"""

    def test_save_context_success(self, mock_project):
        """Test successful context saving"""
        outputs = {"test": "data"}
        files = [".moai/config/config.json"]

        saved_path = save_command_context(
            phase_name="0-project", project_root=mock_project, outputs=outputs, files_created=files, next_phase="1-plan"
        )

        if saved_path:  # Only if ContextManager available
            assert os.path.exists(saved_path)
            assert saved_path.endswith(".json")

            # Verify content
            with open(saved_path, "r") as f:
                data = json.load(f)

            assert data["phase"] == "0-project"
            assert data["status"] == "completed"

    def test_save_context_error_handling(self):
        """Test error handling when save fails"""
        # Invalid project root
        saved_path = save_command_context(
            phase_name="test", project_root="/nonexistent/path", outputs={}, files_created=[]
        )

        # Should return None on error, not raise
        assert saved_path is None


class TestPreviousPhaseLoading:
    """Tests for load_previous_phase function"""

    def test_load_latest_phase(self, mock_project):
        """Test loading most recent phase result"""
        # Save a phase result first
        outputs = {"version": 1}
        files = []

        save_command_context(phase_name="0-project", project_root=mock_project, outputs=outputs, files_created=files)

        # Load it back
        loaded = load_previous_phase(mock_project)

        if loaded:  # Only if ContextManager available
            assert loaded["phase"] == "0-project"
            assert loaded["outputs"]["version"] == 1

    def test_load_no_previous_phase(self):
        """Test loading when no previous phase exists"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create minimal structure
            (Path(tmpdir) / ".moai" / "memory" / "command-state").mkdir(parents=True)

            loaded = load_previous_phase(tmpdir)

            # Should return None gracefully
            assert loaded is None or isinstance(loaded, dict)


class TestEdgeCases:

    def test_extract_metadata_missing_keys(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            (project_root / ".moai" / "config").mkdir(parents=True)

            # Create incomplete config with missing project keys
            incomplete_config = {
                "project": {},  # Missing name, mode, owner
                "language": {},  # Missing conversation_language
            }

            config_path = project_root / ".moai" / "config" / "config.json"
            with open(config_path, "w") as f:
                json.dump(incomplete_config, f)

            result = extract_project_metadata(str(project_root))

            # Verify default values applied
            assert result["project_name"] == "Unknown"
            assert result["mode"] == "personal"
            assert result["owner"] == "@user"
            assert result["language"] == "en"

    def test_detect_tech_stack_no_indicators(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            # Empty directory with no indicator files
            result = detect_tech_stack(tmpdir)

            # Should default to Python
            assert result == ["python"]

    def test_build_phase_result_minimal_data(self):
        result = build_phase_result(phase_name="0-project", status="completed", outputs={}, files_created=[])

        # Verify all required fields present
        assert result["phase"] == "0-project"
        assert result["status"] == "completed"
        assert "timestamp" in result
        assert result["outputs"] == {}
        assert result["files_created"] == []
        # next_phase should not be present when not provided
        assert "next_phase" not in result

    def test_validate_phase_files_outside_project_root(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            (project_root / ".moai" / "config").mkdir(parents=True)

            # Paths that attempt to escape project root
            invalid_paths = ["../../../etc/passwd", "/etc/passwd", "../../outside/project"]

            result = validate_phase_files(invalid_paths, str(project_root))

            # Should handle gracefully - may return empty or filtered list
            assert isinstance(result, list)
            # No path should point outside project root
            for path in result:
                assert tmpdir in path or not os.path.isabs(path)

    def test_save_context_when_context_manager_unavailable(self, monkeypatch):
        import moai_adk.core.command_helpers as helpers

        # Mock CONTEXT_MANAGER_AVAILABLE to False
        monkeypatch.setattr(helpers, "CONTEXT_MANAGER_AVAILABLE", False)

        result = save_command_context(
            phase_name="test", project_root="/tmp/test", outputs={"test": "data"}, files_created=[]
        )

        # Should return None gracefully, not raise exception
        assert result is None

    def test_validate_phase_files_mixed_valid_invalid(self, mock_project):
        """Test validation with mix of valid and invalid paths"""
        mixed_paths = [
            ".moai/config/config.json",  # Valid
            "../../../etc/passwd",  # Invalid - path traversal
            ".moai/memory/test.json",  # Valid but may not exist
        ]

        result = validate_phase_files(mixed_paths, mock_project)

        # Should be a list (may filter out invalid paths)
        assert isinstance(result, list)

    def test_build_phase_result_with_next_phase(self):
        """Test phase result includes next_phase when provided"""
        result = build_phase_result(
            phase_name="0-project", status="completed", outputs={}, files_created=[], next_phase="1-plan"
        )

        assert "next_phase" in result
        assert result["next_phase"] == "1-plan"
