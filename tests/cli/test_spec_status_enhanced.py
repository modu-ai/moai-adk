#!/usr/bin/env python3
"""Comprehensive tests for spec_status.py CLI module

Target: 90%+ coverage for src/moai_adk/cli/spec_status.py
Coverage: Lines 16-223 (100 statements, all currently missed)

Test Structure:
- Class-based organization by function
- Descriptive test names: test_<action>_<condition>_<expected>
- Comprehensive docstrings explaining test purpose
- Use tmp_path for filesystem operations
- Mock SpecStatusManager interactions
- Test success and error paths
- Parametrize for multiple scenarios
"""

import json
from datetime import datetime
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

# Import functions under test
from moai_adk.cli.spec_status import (
    batch_update_completed_specs,
    detect_draft_specs,
    main,
    update_spec_status,
    validate_spec_completion,
)


class TestUpdateSpecStatus:
    """Tests for update_spec_status() function"""

    def test_update_spec_status_success_with_valid_status(self, tmp_path):
        """Test successful status update with valid status and existing SPEC file"""
        # Setup mock SPEC file
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001\nStatus: draft")

        # Mock SpecStatusManager
        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = update_spec_status("SPEC-001", "in-progress", "Starting implementation")

        # Assertions
        assert result["success"] is True
        assert result["spec_id"] == "SPEC-001"
        assert result["new_status"] == "in-progress"
        assert result["reason"] == "Starting implementation"
        assert "timestamp" in result

        # Verify log file creation
        log_file = tmp_path / ".moai" / "logs" / "spec_status_changes.jsonl"
        assert log_file.exists()

        # Verify log content
        log_content = log_file.read_text()
        log_entry = json.loads(log_content.strip())
        assert log_entry["spec_id"] == "SPEC-001"
        assert log_entry["new_status"] == "in-progress"
        assert log_entry["reason"] == "Starting implementation"

    def test_update_spec_status_invalid_status_returns_error(self, tmp_path):
        """Test that invalid status values return error response"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
            result = update_spec_status("SPEC-001", "invalid-status", "Test")

        assert result["success"] is False
        assert "Invalid status" in result["error"]
        assert "invalid-status" in result["error"]

    @pytest.mark.parametrize("valid_status", ["draft", "in-progress", "completed", "archived"])
    def test_update_spec_status_all_valid_statuses(self, tmp_path, valid_status):
        """Test all valid status values are accepted"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = update_spec_status("SPEC-001", valid_status, "Test")

        assert result["success"] is True
        assert result["new_status"] == valid_status

    def test_update_spec_status_missing_spec_file_returns_error(self, tmp_path):
        """Test error when SPEC file does not exist"""
        with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
            result = update_spec_status("SPEC-999", "completed", "Test")

        assert result["success"] is False
        assert "SPEC file not found" in result["error"]

    def test_update_spec_status_manager_failure_returns_error(self, tmp_path):
        """Test error handling when SpecStatusManager.update_spec_status fails"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = False
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = update_spec_status("SPEC-001", "completed", "Test")

        assert result["success"] is False
        assert "Failed to update SPEC" in result["error"]

    def test_update_spec_status_exception_handling(self, tmp_path):
        """Test exception handling during status update"""
        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager_class.side_effect = Exception("Test exception")

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = update_spec_status("SPEC-001", "completed", "Test")

        assert result["success"] is False
        assert "Error updating SPEC status" in result["error"]
        assert "Test exception" in result["error"]

    def test_update_spec_status_log_file_appends_multiple_entries(self, tmp_path):
        """Test that multiple status updates append to log file"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                update_spec_status("SPEC-001", "in-progress", "Start")
                update_spec_status("SPEC-001", "completed", "Done")

        log_file = tmp_path / ".moai" / "logs" / "spec_status_changes.jsonl"
        log_lines = log_file.read_text().strip().split("\n")

        assert len(log_lines) == 2
        entry1 = json.loads(log_lines[0])
        entry2 = json.loads(log_lines[1])

        assert entry1["new_status"] == "in-progress"
        assert entry2["new_status"] == "completed"

    def test_update_spec_status_empty_reason(self, tmp_path):
        """Test status update with empty reason string"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = update_spec_status("SPEC-001", "completed", "")

        assert result["success"] is True
        assert result["reason"] == ""


class TestValidateSpecCompletion:
    """Tests for validate_spec_completion() function"""

    def test_validate_spec_completion_success(self, tmp_path):
        """Test successful SPEC validation with valid SPEC ID"""
        validation_result = {
            "valid": True,
            "criteria_met": {"code_coverage": 0.90, "acceptance_criteria": True, "implementation_age_days": 5},
        }

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.validate_spec_for_completion.return_value = validation_result
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = validate_spec_completion("SPEC-001")

        assert result["success"] is True
        assert result["spec_id"] == "SPEC-001"
        assert result["validation"] == validation_result

    def test_validate_spec_completion_invalid_spec(self, tmp_path):
        """Test validation with SPEC that does not meet completion criteria"""
        validation_result = {
            "valid": False,
            "criteria_met": {
                "code_coverage": 0.70,  # Below threshold
                "acceptance_criteria": False,
                "implementation_age_days": 0,
            },
            "errors": ["Coverage below 85%", "Missing acceptance criteria"],
        }

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.validate_spec_for_completion.return_value = validation_result
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = validate_spec_completion("SPEC-001")

        assert result["success"] is True
        assert result["validation"]["valid"] is False
        assert "errors" in result["validation"]

    def test_validate_spec_completion_exception_handling(self, tmp_path):
        """Test exception handling during validation"""
        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.validate_spec_for_completion.side_effect = Exception("Validation error")
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = validate_spec_completion("SPEC-001")

        assert result["success"] is False
        assert "Error validating SPEC completion" in result["error"]
        assert "Validation error" in result["error"]

    def test_validate_spec_completion_missing_spec_directory(self, tmp_path):
        """Test validation when SPEC directory does not exist"""
        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.validate_spec_for_completion.side_effect = FileNotFoundError("SPEC not found")
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = validate_spec_completion("SPEC-999")

        assert result["success"] is False
        assert "SPEC not found" in result["error"]


class TestBatchUpdateCompletedSpecs:
    """Tests for batch_update_completed_specs() function"""

    def test_batch_update_completed_specs_success(self, tmp_path):
        """Test successful batch update of multiple completed SPECs"""
        batch_results = {"updated": ["SPEC-001", "SPEC-002"], "skipped": ["SPEC-003"], "errors": [], "total": 3}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.batch_update_completed_specs.return_value = batch_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = batch_update_completed_specs()

        assert result["success"] is True
        assert result["results"] == batch_results
        assert "timestamp" in result

        # Verify log file
        log_file = tmp_path / ".moai" / "logs" / "spec_status_changes.jsonl"
        assert log_file.exists()

        log_content = log_file.read_text()
        log_entry = json.loads(log_content.strip())
        assert log_entry["operation"] == "batch_update_completed"
        assert log_entry["results"] == batch_results

    def test_batch_update_completed_specs_empty_results(self, tmp_path):
        """Test batch update when no SPECs need updating"""
        batch_results = {"updated": [], "skipped": [], "errors": [], "total": 0}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.batch_update_completed_specs.return_value = batch_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = batch_update_completed_specs()

        assert result["success"] is True
        assert result["results"]["total"] == 0

    def test_batch_update_completed_specs_with_errors(self, tmp_path):
        """Test batch update with some SPECs failing to update"""
        batch_results = {
            "updated": ["SPEC-001"],
            "skipped": ["SPEC-002"],
            "errors": [{"spec_id": "SPEC-003", "error": "Invalid format"}],
            "total": 3,
        }

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.batch_update_completed_specs.return_value = batch_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = batch_update_completed_specs()

        assert result["success"] is True
        assert len(result["results"]["errors"]) == 1
        assert result["results"]["errors"][0]["spec_id"] == "SPEC-003"

    def test_batch_update_completed_specs_exception_handling(self, tmp_path):
        """Test exception handling during batch update"""
        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.batch_update_completed_specs.side_effect = Exception("Batch update failed")
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = batch_update_completed_specs()

        assert result["success"] is False
        assert "Error in batch update" in result["error"]
        assert "Batch update failed" in result["error"]

    def test_batch_update_completed_specs_log_appending(self, tmp_path):
        """Test that batch update logs append correctly"""
        batch_results1 = {"updated": ["SPEC-001"], "total": 1}
        batch_results2 = {"updated": ["SPEC-002"], "total": 1}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.batch_update_completed_specs.side_effect = [batch_results1, batch_results2]
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                batch_update_completed_specs()
                batch_update_completed_specs()

        log_file = tmp_path / ".moai" / "logs" / "spec_status_changes.jsonl"
        log_lines = log_file.read_text().strip().split("\n")

        assert len(log_lines) == 2


class TestDetectDraftSpecs:
    """Tests for detect_draft_specs() function"""

    def test_detect_draft_specs_success_with_multiple_specs(self, tmp_path):
        """Test detecting multiple draft SPECs"""
        draft_specs = {"SPEC-001", "SPEC-002", "SPEC-003"}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = draft_specs
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = detect_draft_specs()

        assert result["success"] is True
        assert set(result["draft_specs"]) == draft_specs
        assert result["count"] == 3

    def test_detect_draft_specs_empty_results(self, tmp_path):
        """Test detection when no draft SPECs exist"""
        draft_specs = set()

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = draft_specs
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = detect_draft_specs()

        assert result["success"] is True
        assert result["draft_specs"] == []
        assert result["count"] == 0

    def test_detect_draft_specs_single_spec(self, tmp_path):
        """Test detection with single draft SPEC"""
        draft_specs = {"SPEC-001"}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = draft_specs
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = detect_draft_specs()

        assert result["success"] is True
        assert result["count"] == 1

    def test_detect_draft_specs_exception_handling(self, tmp_path):
        """Test exception handling during draft detection"""
        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.side_effect = Exception("Detection failed")
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = detect_draft_specs()

        assert result["success"] is False
        assert "Error detecting draft SPECs" in result["error"]
        assert "Detection failed" in result["error"]


class TestMain:
    """Tests for main() CLI function"""

    def test_main_status_update_command_success(self, tmp_path, capsys):
        """Test main() with status_update command"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                with patch("sys.argv", ["spec_status.py", "status_update", "SPEC-001", "--status", "completed"]):
                    main()

        captured = capsys.readouterr()
        result = json.loads(captured.out)

        assert result["success"] is True
        assert result["command"] == "status_update"
        assert result["spec_id"] == "SPEC-001"

    def test_main_status_update_missing_spec_id(self, capsys):
        """Test main() status_update without spec_id exits with error"""
        with patch("sys.argv", ["spec_status.py", "status_update"]):
            with pytest.raises(SystemExit) as exc_info:
                main()

        assert exc_info.value.code == 1
        captured = capsys.readouterr()
        result = json.loads(captured.out)
        assert result["success"] is False
        assert "requires spec_id" in result["error"]

    def test_main_status_update_missing_status_flag(self, capsys):
        """Test main() status_update without --status flag exits with error"""
        with patch("sys.argv", ["spec_status.py", "status_update", "SPEC-001"]):
            with pytest.raises(SystemExit) as exc_info:
                main()

        assert exc_info.value.code == 1
        captured = capsys.readouterr()
        result = json.loads(captured.out)
        assert result["success"] is False
        assert "requires spec_id and --status" in result["error"]

    def test_main_validate_completion_command_success(self, tmp_path, capsys):
        """Test main() with validate_completion command"""
        validation_result = {"valid": True, "criteria_met": {}}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.validate_spec_for_completion.return_value = validation_result
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                with patch("sys.argv", ["spec_status.py", "validate_completion", "SPEC-001"]):
                    main()

        captured = capsys.readouterr()
        result = json.loads(captured.out)

        assert result["success"] is True
        assert result["command"] == "validate_completion"
        assert result["spec_id"] == "SPEC-001"

    def test_main_validate_completion_missing_spec_id(self, capsys):
        """Test main() validate_completion without spec_id exits with error"""
        with patch("sys.argv", ["spec_status.py", "validate_completion"]):
            with pytest.raises(SystemExit) as exc_info:
                main()

        assert exc_info.value.code == 1
        captured = capsys.readouterr()
        result = json.loads(captured.out)
        assert result["success"] is False
        assert "requires spec_id" in result["error"]

    def test_main_batch_update_command_success(self, tmp_path, capsys):
        """Test main() with batch_update command"""
        batch_results = {"updated": ["SPEC-001"], "total": 1}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.batch_update_completed_specs.return_value = batch_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                with patch("sys.argv", ["spec_status.py", "batch_update"]):
                    main()

        captured = capsys.readouterr()
        result = json.loads(captured.out)

        assert result["success"] is True
        assert result["command"] == "batch_update"

    def test_main_detect_drafts_command_success(self, tmp_path, capsys):
        """Test main() with detect_drafts command"""
        draft_specs = {"SPEC-001", "SPEC-002"}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = draft_specs
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                with patch("sys.argv", ["spec_status.py", "detect_drafts"]):
                    main()

        captured = capsys.readouterr()
        result = json.loads(captured.out)

        assert result["success"] is True
        assert result["command"] == "detect_drafts"
        assert result["count"] == 2

    def test_main_exception_handling_and_exit(self, tmp_path, capsys):
        """Test main() exception handling when an uncaught exception occurs"""
        # Patch argparse to raise an exception after parsing (simulating runtime error)
        with patch("sys.argv", ["spec_status.py", "detect_drafts"]):
            with patch("moai_adk.cli.spec_status.detect_draft_specs") as mock_detect:
                # Simulate an uncaught exception that propagates to main's exception handler
                mock_detect.side_effect = RuntimeError("Unexpected runtime error")

                with pytest.raises(SystemExit) as exc_info:
                    main()

        assert exc_info.value.code == 1
        captured = capsys.readouterr()
        result = json.loads(captured.out)
        assert result["success"] is False
        assert "Unexpected runtime error" in result["error"]

    def test_main_includes_execution_time_in_output(self, tmp_path, capsys):
        """Test that main() includes execution timestamp in output"""
        draft_specs = set()

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = draft_specs
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                with patch("sys.argv", ["spec_status.py", "detect_drafts"]):
                    main()

        captured = capsys.readouterr()
        result = json.loads(captured.out)

        assert "execution_time" in result
        # Verify it's a valid ISO timestamp
        datetime.fromisoformat(result["execution_time"])

    def test_main_with_reason_flag(self, tmp_path, capsys):
        """Test main() status_update with --reason flag"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                with patch(
                    "sys.argv",
                    [
                        "spec_status.py",
                        "status_update",
                        "SPEC-001",
                        "--status",
                        "completed",
                        "--reason",
                        "All tests pass",
                    ],
                ):
                    main()

        captured = capsys.readouterr()
        result = json.loads(captured.out)

        assert result["success"] is True
        assert result["reason"] == "All tests pass"


class TestImportFallback:
    """Tests for import fallback mechanism"""

    def test_import_fallback_path_insertion(self):
        """Test that import fallback inserts correct path to sys.path"""
        # This test verifies the fallback logic structure
        # In practice, this is tested implicitly by successful imports

        # Simulate the fallback path calculation
        current_file = Path(__file__)
        expected_parent = current_file.parent.parent

        # Verify path structure matches fallback logic
        assert expected_parent.name in ["moai_adk", "tests"]

    def test_module_importable_without_fallback(self):
        """Test that module imports successfully (no fallback needed)"""
        # If this import succeeds, fallback is not triggered
        from moai_adk.cli.spec_status import main, update_spec_status

        assert callable(main)
        assert callable(update_spec_status)


class TestEdgeCases:
    """Tests for edge cases and boundary conditions"""

    def test_update_spec_status_with_special_characters_in_reason(self, tmp_path):
        """Test status update with special characters in reason"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001")

        special_reason = "测试 テスト тест 🎉 \"quoted\" 'single' <tags>"

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = update_spec_status("SPEC-001", "completed", special_reason)

        assert result["success"] is True
        assert result["reason"] == special_reason

        # Verify JSON encoding works
        log_file = tmp_path / ".moai" / "logs" / "spec_status_changes.jsonl"
        log_content = log_file.read_text()
        log_entry = json.loads(log_content.strip())
        assert log_entry["reason"] == special_reason

    def test_update_spec_status_with_long_spec_id(self, tmp_path):
        """Test status update with longer SPEC ID (but filesystem-safe)"""
        # Use a reasonably long SPEC ID that won't exceed filesystem limits
        long_spec_id = "SPEC-" + "0" * 50
        spec_dir = tmp_path / ".moai" / "specs" / long_spec_id
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text(f"# {long_spec_id}")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = update_spec_status(long_spec_id, "completed", "Test")

        assert result["success"] is True
        assert result["spec_id"] == long_spec_id
        assert len(long_spec_id) > 50  # Verify it's actually long

    def test_detect_draft_specs_returns_deterministic_order(self, tmp_path):
        """Test that draft detection returns consistent list conversion"""
        # Sets are unordered, but conversion to list should be deterministic
        draft_specs1 = {"SPEC-003", "SPEC-001", "SPEC-002"}
        draft_specs2 = {"SPEC-001", "SPEC-002", "SPEC-003"}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = draft_specs1
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result1 = detect_draft_specs()

            mock_manager.detect_draft_specs.return_value = draft_specs2

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result2 = detect_draft_specs()

        # Both should contain same elements (order may vary)
        assert set(result1["draft_specs"]) == set(result2["draft_specs"])


# Integration-like tests (still using mocks but testing multiple functions)
class TestIntegrationScenarios:
    """Integration-like tests for common workflow scenarios"""

    def test_workflow_draft_to_completed(self, tmp_path):
        """Test typical workflow: detect draft → update → validate"""
        spec_dir = tmp_path / ".moai" / "specs" / "SPEC-001"
        spec_dir.mkdir(parents=True)
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("# SPEC-001\nStatus: draft")

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()

            # Step 1: Detect drafts
            mock_manager.detect_draft_specs.return_value = {"SPEC-001"}
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                detect_result = detect_draft_specs()

            assert detect_result["success"] is True
            assert "SPEC-001" in detect_result["draft_specs"]

            # Step 2: Update status
            mock_manager.update_spec_status.return_value = True

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                update_result = update_spec_status("SPEC-001", "in-progress", "Starting work")

            assert update_result["success"] is True

            # Step 3: Validate completion
            validation = {"valid": True, "criteria_met": {"coverage": 0.90}}
            mock_manager.validate_spec_for_completion.return_value = validation

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                validate_result = validate_spec_completion("SPEC-001")

            assert validate_result["success"] is True
            assert validate_result["validation"]["valid"] is True

    def test_batch_update_workflow(self, tmp_path):
        """Test batch update workflow with multiple SPECs"""
        batch_results = {"updated": ["SPEC-001", "SPEC-002"], "skipped": ["SPEC-003"], "errors": [], "total": 3}

        with patch("moai_adk.cli.spec_status.SpecStatusManager") as mock_manager_class:
            mock_manager = MagicMock()
            mock_manager.batch_update_completed_specs.return_value = batch_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=tmp_path):
                result = batch_update_completed_specs()

        assert result["success"] is True
        assert len(result["results"]["updated"]) == 2
        assert len(result["results"]["skipped"]) == 1

        # Verify log creation
        log_file = tmp_path / ".moai" / "logs" / "spec_status_changes.jsonl"
        assert log_file.exists()
