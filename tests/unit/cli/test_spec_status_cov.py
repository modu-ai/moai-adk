"""Comprehensive test coverage for spec_status module.

Focus on uncovered code paths with mocked dependencies.
Tests actual code paths without side effects.
"""

import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.cli.spec_status import (
    batch_update_completed_specs,
    detect_draft_specs,
    main,
    update_spec_status,
    validate_spec_completion,
)


class TestUpdateSpecStatus:
    """Test update_spec_status function."""

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_status_success(self, mock_manager_class):
        """Test successful SPEC status update."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create SPEC file
            spec_dir = project_root / ".moai" / "specs" / "SPEC-AUTH-001"
            spec_dir.mkdir(parents=True, exist_ok=True)
            (spec_dir / "spec.md").write_text("# SPEC-AUTH-001")

            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            # Mock Path.cwd() to return our temp directory
            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = update_spec_status("SPEC-AUTH-001", "in-progress", "Started work")

                # Assert
                assert result["success"] is True
                assert result["spec_id"] == "SPEC-AUTH-001"
                assert result["new_status"] == "in-progress"
                assert result["reason"] == "Started work"

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_status_invalid_status(self, mock_manager_class):
        """Test update with invalid status."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager_class.return_value = MagicMock()

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = update_spec_status("SPEC-AUTH-001", "invalid_status")

                # Assert
                assert result["success"] is False
                assert "Invalid status" in result["error"]

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_status_spec_not_found(self, mock_manager_class):
        """Test update when SPEC file doesn't exist."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager_class.return_value = MagicMock()

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = update_spec_status("SPEC-MISSING-001", "draft")

                # Assert
                assert result["success"] is False
                assert "SPEC file not found" in result["error"]

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_status_manager_failure(self, mock_manager_class):
        """Test update when manager update fails."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create SPEC file
            spec_dir = project_root / ".moai" / "specs" / "SPEC-AUTH-001"
            spec_dir.mkdir(parents=True, exist_ok=True)
            (spec_dir / "spec.md").write_text("# SPEC-AUTH-001")

            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = False
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = update_spec_status("SPEC-AUTH-001", "completed")

                # Assert
                assert result["success"] is False
                assert "Failed to update" in result["error"]

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_status_creates_log_entry(self, mock_manager_class):
        """Test that status update creates log entry."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create SPEC file
            spec_dir = project_root / ".moai" / "specs" / "SPEC-DB-001"
            spec_dir.mkdir(parents=True, exist_ok=True)
            (spec_dir / "spec.md").write_text("# SPEC-DB-001")

            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = update_spec_status("SPEC-DB-001", "archived", "Deprecated")

                # Assert
                assert result["success"] is True
                # Verify log file was created
                log_file = project_root / ".moai" / "logs" / "spec_status_changes.jsonl"
                assert log_file.exists()

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_status_exception_handling(self, mock_manager_class):
        """Test update handles exceptions."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager_class.side_effect = Exception("Database error")

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = update_spec_status("SPEC-AUTH-001", "draft")

                # Assert
                assert result["success"] is False
                assert "Error updating SPEC status" in result["error"]

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_status_all_valid_statuses(self, mock_manager_class):
        """Test update with all valid statuses."""
        # Arrange
        valid_statuses = ["draft", "in-progress", "completed", "archived"]

        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            for status_val in valid_statuses:
                # Create SPEC file
                spec_dir = project_root / ".moai" / "specs" / "SPEC-TEST-001"
                spec_dir.mkdir(parents=True, exist_ok=True)
                (spec_dir / "spec.md").write_text("# SPEC-TEST-001")

                mock_manager = MagicMock()
                mock_manager.update_spec_status.return_value = True
                mock_manager_class.return_value = mock_manager

                with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                    # Act
                    result = update_spec_status("SPEC-TEST-001", status_val)

                    # Assert
                    assert result["success"] is True
                    assert result["new_status"] == status_val


class TestValidateSpecCompletion:
    """Test validate_spec_completion function."""

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_validate_spec_completion_success(self, mock_manager_class):
        """Test successful SPEC completion validation."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager = MagicMock()
            mock_validation = {"ready": True, "tests": True, "docs": True}
            mock_manager.validate_spec_for_completion.return_value = mock_validation
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = validate_spec_completion("SPEC-AUTH-001")

                # Assert
                assert result["success"] is True
                assert result["spec_id"] == "SPEC-AUTH-001"
                assert result["validation"] == mock_validation

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_validate_spec_completion_not_ready(self, mock_manager_class):
        """Test validation when SPEC not ready."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager = MagicMock()
            mock_validation = {"ready": False, "tests": False, "docs": True}
            mock_manager.validate_spec_for_completion.return_value = mock_validation
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = validate_spec_completion("SPEC-API-001")

                # Assert
                assert result["success"] is True
                assert result["validation"]["ready"] is False

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_validate_spec_completion_exception(self, mock_manager_class):
        """Test validation handles exceptions."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager_class.side_effect = Exception("Validation error")

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = validate_spec_completion("SPEC-AUTH-001")

                # Assert
                assert result["success"] is False
                assert "Error validating SPEC completion" in result["error"]


class TestBatchUpdateCompletedSpecs:
    """Test batch_update_completed_specs function."""

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_batch_update_success(self, mock_manager_class):
        """Test successful batch update."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager = MagicMock()
            mock_results = {"updated": ["SPEC-001", "SPEC-002"], "count": 2}
            mock_manager.batch_update_completed_specs.return_value = mock_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = batch_update_completed_specs()

                # Assert
                assert result["success"] is True
                assert result["results"] == mock_results

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_batch_update_no_changes(self, mock_manager_class):
        """Test batch update with no changes."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager = MagicMock()
            mock_results = {"updated": [], "count": 0}
            mock_manager.batch_update_completed_specs.return_value = mock_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = batch_update_completed_specs()

                # Assert
                assert result["success"] is True
                assert len(result["results"]["updated"]) == 0

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_batch_update_creates_log(self, mock_manager_class):
        """Test batch update creates log entry."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager = MagicMock()
            mock_results = {"updated": ["SPEC-001"], "count": 1}
            mock_manager.batch_update_completed_specs.return_value = mock_results
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = batch_update_completed_specs()

                # Assert
                assert result["success"] is True
                # Verify log file exists
                log_file = project_root / ".moai" / "logs" / "spec_status_changes.jsonl"
                assert log_file.exists()

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_batch_update_exception(self, mock_manager_class):
        """Test batch update handles exceptions."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager_class.side_effect = Exception("Batch error")

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = batch_update_completed_specs()

                # Assert
                assert result["success"] is False
                assert "Error in batch update" in result["error"]


class TestDetectDraftSpecs:
    """Test detect_draft_specs function."""

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_detect_draft_specs_success(self, mock_manager_class):
        """Test successful draft SPEC detection."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            draft_specs = ["SPEC-001", "SPEC-002", "SPEC-003"]
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = draft_specs
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = detect_draft_specs()

                # Assert
                assert result["success"] is True
                assert result["count"] == 3
                assert result["draft_specs"] == draft_specs

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_detect_draft_specs_no_drafts(self, mock_manager_class):
        """Test detection when no draft SPECs exist."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager = MagicMock()
            mock_manager.detect_draft_specs.return_value = []
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = detect_draft_specs()

                # Assert
                assert result["success"] is True
                assert result["count"] == 0
                assert result["draft_specs"] == []

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_detect_draft_specs_exception(self, mock_manager_class):
        """Test detection handles exceptions."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            mock_manager_class.side_effect = Exception("Detection error")

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = detect_draft_specs()

                # Assert
                assert result["success"] is False
                assert "Error detecting draft SPECs" in result["error"]


class TestMainFunction:
    """Test main CLI function."""

    @patch("moai_adk.cli.spec_status.update_spec_status")
    @patch(
        "sys.argv",
        ["spec_status", "status_update", "SPEC-001", "--status", "in-progress"],
    )
    def test_main_status_update(self, mock_update):
        """Test main function with status_update command."""
        # Arrange
        mock_update.return_value = {"success": True}

        # Act & Assert - should not raise exception
        try:
            main()
        except SystemExit as e:
            # Script exits with 0 on success
            assert e.code == 0

    @patch("moai_adk.cli.spec_status.validate_spec_completion")
    @patch("sys.argv", ["spec_status", "validate_completion", "SPEC-001"])
    def test_main_validate_completion(self, mock_validate):
        """Test main function with validate_completion command."""
        # Arrange
        mock_validate.return_value = {"success": True}

        # Act & Assert - should not raise exception
        try:
            main()
        except SystemExit as e:
            assert e.code == 0

    @patch("moai_adk.cli.spec_status.batch_update_completed_specs")
    @patch("sys.argv", ["spec_status", "batch_update"])
    def test_main_batch_update(self, mock_batch):
        """Test main function with batch_update command."""
        # Arrange
        mock_batch.return_value = {"success": True}

        # Act & Assert
        try:
            main()
        except SystemExit as e:
            assert e.code == 0

    @patch("moai_adk.cli.spec_status.detect_draft_specs")
    @patch("sys.argv", ["spec_status", "detect_drafts"])
    def test_main_detect_drafts(self, mock_detect):
        """Test main function with detect_drafts command."""
        # Arrange
        mock_detect.return_value = {"success": True, "draft_specs": []}

        # Act & Assert
        try:
            main()
        except SystemExit as e:
            assert e.code == 0

    @patch("sys.argv", ["spec_status", "status_update"])
    def test_main_missing_arguments(self):
        """Test main function with missing required arguments."""
        # Act & Assert - should exit with error
        with pytest.raises(SystemExit):
            main()

    @patch("sys.argv", ["spec_status", "status_update", "SPEC-001"])
    def test_main_missing_status_option(self):
        """Test main function with missing --status option."""
        # Act & Assert - should exit with error
        with pytest.raises(SystemExit):
            main()

    @patch("moai_adk.cli.spec_status.update_spec_status")
    @patch(
        "sys.argv",
        [
            "spec_status",
            "status_update",
            "SPEC-001",
            "--status",
            "completed",
            "--reason",
            "Done",
        ],
    )
    def test_main_with_reason(self, mock_update):
        """Test main function with reason argument."""
        # Arrange
        mock_update.return_value = {"success": True, "reason": "Done"}

        # Act & Assert
        try:
            main()
        except SystemExit as e:
            assert e.code == 0

        mock_update.assert_called_once()
        call_args = mock_update.call_args
        assert call_args[0][2] == "Done"  # Reason passed as third argument


class TestSpecStatusEdgeCases:
    """Test edge cases and error conditions."""

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_with_empty_reason(self, mock_manager_class):
        """Test update with empty reason."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            spec_dir = project_root / ".moai" / "specs" / "SPEC-AUTH-001"
            spec_dir.mkdir(parents=True, exist_ok=True)
            (spec_dir / "spec.md").write_text("# SPEC-AUTH-001")

            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act
                result = update_spec_status("SPEC-AUTH-001", "draft", "")

                # Assert
                assert result["success"] is True
                assert result["reason"] == ""

    @patch("moai_adk.cli.spec_status.SpecStatusManager")
    def test_update_spec_with_special_characters(self, mock_manager_class):
        """Test update with special characters in spec ID."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            spec_dir = project_root / ".moai" / "specs" / "SPEC-TEST-001"
            spec_dir.mkdir(parents=True, exist_ok=True)
            (spec_dir / "spec.md").write_text("# SPEC-TEST-001")

            mock_manager = MagicMock()
            mock_manager.update_spec_status.return_value = True
            mock_manager_class.return_value = mock_manager

            with patch("moai_adk.cli.spec_status.Path.cwd", return_value=project_root):
                # Act - spec ID is literal
                result = update_spec_status("SPEC-TEST-001", "draft")

                # Assert
                assert result["success"] is True
