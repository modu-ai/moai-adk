"""
Comprehensive tests for command_helpers module.

Tests helper functions for command context management and project metadata extraction.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.command_helpers import (
    build_phase_result,
    detect_tech_stack,
    extract_project_metadata,
    load_previous_phase,
    save_command_context,
    validate_phase_files,
)


class TestExtractProjectMetadata:
    """Test extract_project_metadata function."""

    def test_extract_valid_config(self):
        """Test extracting metadata from valid config.json."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create config directory and file
            config_dir = Path(tmpdir) / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_file = config_dir / "config.json"

            config_data = {
                "project": {
                    "name": "test-project",
                    "mode": "team",
                    "owner": "@john",
                },
                "language": {"conversation_language": "ko"},
            }

            config_file.write_text(json.dumps(config_data))

            # Test
            metadata = extract_project_metadata(tmpdir)

            assert metadata["project_name"] == "test-project"
            assert metadata["mode"] == "team"
            assert metadata["owner"] == "@john"
            assert metadata["language"] == "ko"

    def test_extract_with_defaults(self):
        """Test extraction uses defaults for missing keys."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = Path(tmpdir) / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_file = config_dir / "config.json"

            # Minimal config
            config_data = {"project": {"name": "minimal"}, "language": {}}

            config_file.write_text(json.dumps(config_data))

            metadata = extract_project_metadata(tmpdir)

            assert metadata["project_name"] == "minimal"
            assert metadata["mode"] == "personal"  # Default value
            assert metadata["owner"] == "@user"  # Default value
            assert metadata["language"] == "en"  # Default value

    def test_extract_missing_config_file(self):
        """Test error when config file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with pytest.raises(
                FileNotFoundError, match="Config file not found"
            ):
                extract_project_metadata(tmpdir)

    def test_extract_invalid_json(self):
        """Test error handling for invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = Path(tmpdir) / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_file = config_dir / "config.json"

            config_file.write_text("{ invalid json }")

            with pytest.raises(json.JSONDecodeError):
                extract_project_metadata(tmpdir)

    def test_extract_empty_config(self):
        """Test extraction from empty config."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = Path(tmpdir) / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_file = config_dir / "config.json"

            config_file.write_text(json.dumps({}))

            metadata = extract_project_metadata(tmpdir)

            assert metadata["project_name"] == "Unknown"
            assert metadata["mode"] == "personal"
            assert metadata["tech_stack"] == []


class TestDetectTechStack:
    """Test detect_tech_stack function."""

    def test_detect_python(self):
        """Test Python project detection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir, "pyproject.toml").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "python" in tech_stack

    def test_detect_javascript(self):
        """Test JavaScript project detection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir, "package.json").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "javascript" in tech_stack

    def test_detect_go(self):
        """Test Go project detection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir, "go.mod").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "go" in tech_stack

    def test_detect_rust(self):
        """Test Rust project detection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir, "Cargo.toml").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "rust" in tech_stack

    def test_detect_java(self):
        """Test Java project detection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir, "pom.xml").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "java" in tech_stack

    def test_detect_ruby(self):
        """Test Ruby project detection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir, "Gemfile").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "ruby" in tech_stack

    def test_detect_multiple_languages(self):
        """Test detection of multiple language indicators."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir, "pyproject.toml").touch()
            Path(tmpdir, "package.json").touch()
            Path(tmpdir, "go.mod").touch()

            tech_stack = detect_tech_stack(tmpdir)

            assert "python" in tech_stack
            assert "javascript" in tech_stack
            assert "go" in tech_stack

    def test_detect_no_indicators(self):
        """Test default to Python when no indicators found."""
        with tempfile.TemporaryDirectory() as tmpdir:
            tech_stack = detect_tech_stack(tmpdir)

            # Should default to Python
            assert "python" in tech_stack

    def test_detect_case_sensitive(self):
        """Test detection is case-sensitive (correct behavior)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Wrong case - should not detect
            Path(tmpdir, "pyproject.TOML").touch()

            tech_stack = detect_tech_stack(tmpdir)

            # Should default to Python only
            assert tech_stack == ["python"]


class TestBuildPhaseResult:
    """Test build_phase_result function."""

    def test_build_basic_phase_result(self):
        """Test building basic phase result."""
        result = build_phase_result(
            phase_name="0-project",
            status="completed",
            outputs={"key": "value"},
            files_created=["/path/to/file1.txt", "/path/to/file2.txt"],
        )

        assert result["phase"] == "0-project"
        assert result["status"] == "completed"
        assert result["outputs"]["key"] == "value"
        assert len(result["files_created"]) == 2
        assert "timestamp" in result
        assert "next_phase" not in result

    def test_build_with_next_phase(self):
        """Test building phase result with next phase."""
        result = build_phase_result(
            phase_name="1-plan",
            status="completed",
            outputs={},
            files_created=[],
            next_phase="2-run",
        )

        assert result["phase"] == "1-plan"
        assert result["next_phase"] == "2-run"

    def test_timestamp_format(self):
        """Test timestamp is in ISO format."""
        result = build_phase_result(
            phase_name="test",
            status="completed",
            outputs={},
            files_created=[],
        )

        timestamp = result["timestamp"]
        # Should be ISO format and end with Z
        assert timestamp.endswith("Z")
        assert "T" in timestamp

    def test_empty_outputs(self):
        """Test with empty outputs dictionary."""
        result = build_phase_result(
            phase_name="test",
            status="completed",
            outputs={},
            files_created=[],
        )

        assert result["outputs"] == {}

    def test_empty_files_created(self):
        """Test with empty files list."""
        result = build_phase_result(
            phase_name="test",
            status="completed",
            outputs={"data": "test"},
            files_created=[],
        )

        assert result["files_created"] == []

    def test_status_values(self):
        """Test different status values."""
        for status in ["completed", "error", "interrupted"]:
            result = build_phase_result(
                phase_name="test",
                status=status,
                outputs={},
                files_created=[],
            )
            assert result["status"] == status


class TestValidatePhaseFiles:
    """Test validate_phase_files function."""

    def test_validate_simple_paths(self):
        """Test validation of simple relative paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test files
            Path(tmpdir, "file1.txt").touch()
            Path(tmpdir, "file2.txt").touch()

            result = validate_phase_files(["file1.txt", "file2.txt"], tmpdir)

            assert len(result) == 2
            assert all(Path(p).is_absolute() for p in result)

    def test_validate_nested_paths(self):
        """Test validation of nested relative paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create nested structure
            Path(tmpdir, "dir1").mkdir()
            Path(tmpdir, "dir1", "file.txt").touch()

            result = validate_phase_files(["dir1/file.txt"], tmpdir)

            assert len(result) == 1
            assert result[0].endswith("file.txt")

    def test_validate_empty_list(self):
        """Test with empty path list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            result = validate_phase_files([], tmpdir)

            assert result == []

    @patch("moai_adk.core.command_helpers.CONTEXT_MANAGER_AVAILABLE", False)
    def test_validate_fallback_mode(self):
        """Test fallback mode when ContextManager unavailable."""
        with tempfile.TemporaryDirectory() as tmpdir:
            result = validate_phase_files(["file.txt", "dir/file.txt"], tmpdir)

            # Should return absolute paths using fallback
            assert len(result) == 2
            assert all(Path(p).is_absolute() for p in result)


class TestSaveCommandContext:
    """Test save_command_context function."""

    @patch("moai_adk.core.command_helpers.ContextManager")
    def test_save_context_success(self, mock_context_manager_class):
        """Test successful context saving."""
        # Setup mock
        mock_cm = MagicMock()
        mock_cm.save_phase_result.return_value = "/path/to/saved.json"
        mock_context_manager_class.return_value = mock_cm

        with tempfile.TemporaryDirectory() as tmpdir:
            result = save_command_context(
                phase_name="0-project",
                project_root=tmpdir,
                outputs={"key": "value"},
                files_created=["file.txt"],
            )

        assert result == "/path/to/saved.json"
        mock_cm.save_phase_result.assert_called_once()

    @patch("moai_adk.core.command_helpers.ContextManager")
    def test_save_context_with_next_phase(self, mock_context_manager_class):
        """Test saving context with next phase."""
        mock_cm = MagicMock()
        mock_cm.save_phase_result.return_value = "/path/saved.json"
        mock_context_manager_class.return_value = mock_cm

        with tempfile.TemporaryDirectory() as tmpdir:
            save_command_context(
                phase_name="1-plan",
                project_root=tmpdir,
                outputs={},
                files_created=[],
                next_phase="2-run",
            )

        call_args = mock_cm.save_phase_result.call_args[0][0]
        assert call_args.get("next_phase") == "2-run"

    @patch("moai_adk.core.command_helpers.ContextManager")
    def test_save_context_error_handling(self, mock_context_manager_class):
        """Test error handling when save fails."""
        mock_cm = MagicMock()
        mock_cm.save_phase_result.side_effect = Exception("Save failed")
        mock_context_manager_class.return_value = mock_cm

        with tempfile.TemporaryDirectory() as tmpdir:
            result = save_command_context(
                phase_name="test",
                project_root=tmpdir,
                outputs={},
                files_created=[],
            )

        # Should return None on error
        assert result is None

    @patch("moai_adk.core.command_helpers.CONTEXT_MANAGER_AVAILABLE", False)
    def test_save_context_manager_unavailable(self):
        """Test when ContextManager is not available."""
        with tempfile.TemporaryDirectory() as tmpdir:
            result = save_command_context(
                phase_name="test",
                project_root=tmpdir,
                outputs={},
                files_created=[],
            )

        # Should return None when unavailable
        assert result is None

    @patch("moai_adk.core.command_helpers.ContextManager")
    def test_save_context_custom_status(self, mock_context_manager_class):
        """Test saving with custom status."""
        mock_cm = MagicMock()
        mock_cm.save_phase_result.return_value = "/path/saved.json"
        mock_context_manager_class.return_value = mock_cm

        with tempfile.TemporaryDirectory() as tmpdir:
            save_command_context(
                phase_name="test",
                project_root=tmpdir,
                outputs={},
                files_created=[],
                status="error",
            )

        call_args = mock_cm.save_phase_result.call_args[0][0]
        assert call_args.get("status") == "error"


class TestLoadPreviousPhase:
    """Test load_previous_phase function."""

    @patch("moai_adk.core.command_helpers.ContextManager")
    def test_load_phase_success(self, mock_context_manager_class):
        """Test successful phase loading."""
        mock_cm = MagicMock()
        phase_data = {"phase": "0-project", "status": "completed"}
        mock_cm.load_latest_phase.return_value = phase_data
        mock_context_manager_class.return_value = mock_cm

        with tempfile.TemporaryDirectory() as tmpdir:
            result = load_previous_phase(tmpdir)

        assert result["phase"] == "0-project"
        assert result["status"] == "completed"

    @patch("moai_adk.core.command_helpers.ContextManager")
    def test_load_phase_none(self, mock_context_manager_class):
        """Test loading when no phase exists."""
        mock_cm = MagicMock()
        mock_cm.load_latest_phase.return_value = None
        mock_context_manager_class.return_value = mock_cm

        with tempfile.TemporaryDirectory() as tmpdir:
            result = load_previous_phase(tmpdir)

        assert result is None

    @patch("moai_adk.core.command_helpers.ContextManager")
    def test_load_phase_error(self, mock_context_manager_class):
        """Test error handling during phase loading."""
        mock_cm = MagicMock()
        mock_cm.load_latest_phase.side_effect = Exception("Load failed")
        mock_context_manager_class.return_value = mock_cm

        with tempfile.TemporaryDirectory() as tmpdir:
            result = load_previous_phase(tmpdir)

        # Should return None on error
        assert result is None

    @patch("moai_adk.core.command_helpers.CONTEXT_MANAGER_AVAILABLE", False)
    def test_load_phase_manager_unavailable(self):
        """Test when ContextManager is not available."""
        with tempfile.TemporaryDirectory() as tmpdir:
            result = load_previous_phase(tmpdir)

        # Should return None when unavailable
        assert result is None


class TestIntegrationMetadataExtraction:
    """Integration tests for metadata extraction."""

    def test_extract_and_detect_full_project(self):
        """Test extracting metadata and detecting tech stack together."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create config
            config_dir = Path(tmpdir) / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_file = config_dir / "config.json"

            config_data = {
                "project": {"name": "full-project"},
                "language": {"conversation_language": "en"},
            }
            config_file.write_text(json.dumps(config_data))

            # Create tech indicators
            Path(tmpdir, "pyproject.toml").touch()
            Path(tmpdir, "package.json").touch()

            # Extract metadata
            metadata = extract_project_metadata(tmpdir)
            assert metadata["project_name"] == "full-project"

            # Detect stack
            tech_stack = detect_tech_stack(tmpdir)
            assert "python" in tech_stack
            assert "javascript" in tech_stack


class TestPhaseResultBuilding:
    """Test phase result building and structure."""

    def test_phase_result_structure(self):
        """Test complete phase result structure."""
        result = build_phase_result(
            phase_name="complete-test",
            status="completed",
            outputs={"report": "generated", "count": 5},
            files_created=["src/test.py", "tests/test_test.py"],
            next_phase="3-sync",
        )

        # Verify all required fields
        assert "phase" in result
        assert "timestamp" in result
        assert "status" in result
        assert "outputs" in result
        assert "files_created" in result
        assert "next_phase" in result

        # Verify data types
        assert isinstance(result["outputs"], dict)
        assert isinstance(result["files_created"], list)
        assert isinstance(result["timestamp"], str)


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_extract_project_special_characters(self):
        """Test with special characters in config values."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_dir = Path(tmpdir) / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_file = config_dir / "config.json"

            config_data = {
                "project": {
                    "name": "프로젝트-特殊文字",
                    "owner": "@user@example.com",
                },
                "language": {"conversation_language": "en"},
            }

            config_file.write_text(json.dumps(config_data, ensure_ascii=False))

            metadata = extract_project_metadata(tmpdir)
            assert "프로젝트" in metadata["project_name"]

    def test_build_phase_with_complex_outputs(self):
        """Test building phase with nested output structures."""
        complex_outputs = {
            "nested": {"deep": {"data": [1, 2, 3]}},
            "list": ["a", "b", "c"],
            "unicode": "日本語",
        }

        result = build_phase_result(
            phase_name="complex",
            status="completed",
            outputs=complex_outputs,
            files_created=[],
        )

        assert result["outputs"]["nested"]["deep"]["data"] == [1, 2, 3]
        assert result["outputs"]["unicode"] == "日本語"

    def test_validate_files_with_special_paths(self):
        """Test validating files with special path characters."""
        with tempfile.TemporaryDirectory() as tmpdir:
            special_dir = Path(tmpdir) / "dir-with-dash"
            special_dir.mkdir()
            (special_dir / "file_with_underscore.txt").touch()

            result = validate_phase_files(
                ["dir-with-dash/file_with_underscore.txt"], tmpdir
            )

            assert len(result) == 1
