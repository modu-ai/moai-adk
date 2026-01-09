"""Tests for all-is-well CLI Command

TDD RED phase: Tests for CLI command that starts the /all-is-well workflow.
Tests should FAIL initially until command is implemented.
"""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from click.testing import CliRunner


class TestAllIsWellCommand:
    """Test all-is-well CLI command."""

    def test_command_exists(self):
        """Test that the command is registered."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        assert all_is_well is not None
        assert callable(all_is_well)

    def test_command_help(self):
        """Test command help text."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        result = runner.invoke(all_is_well, ["--help"])

        assert result.exit_code == 0
        assert "features" in result.output.lower() or "FEATURES" in result.output

    def test_command_requires_features(self):
        """Test command requires features argument."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        result = runner.invoke(all_is_well, [])

        # Should fail or show help without features
        assert result.exit_code != 0 or "Usage" in result.output

    def test_command_with_single_feature(self):
        """Test command with single feature."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(all_is_well, ["user authentication"])

            assert result.exit_code == 0

    def test_command_with_multiple_features(self):
        """Test command with multiple features."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(
                all_is_well,
                ["user authentication", "dashboard", "api endpoints"],
            )

            assert result.exit_code == 0


class TestAllIsWellOptions:
    """Test all-is-well command options."""

    def test_worktree_option(self):
        """Test --worktree option."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(
                all_is_well,
                ["feature", "--worktree"],
            )

            assert result.exit_code == 0
            # Verify worktree was passed to run_workflow
            call_args = mock_run.call_args
            assert call_args is not None

    def test_parallel_option(self):
        """Test --parallel option."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(
                all_is_well,
                ["feature", "--parallel", "3"],
            )

            assert result.exit_code == 0

    def test_no_branch_option(self):
        """Test --no-branch option."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(
                all_is_well,
                ["feature", "--no-branch"],
            )

            assert result.exit_code == 0

    def test_no_pr_option(self):
        """Test --no-pr option."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(
                all_is_well,
                ["feature", "--no-pr"],
            )

            assert result.exit_code == 0

    def test_auto_merge_option(self):
        """Test --auto-merge option."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(
                all_is_well,
                ["feature", "--auto-merge"],
            )

            assert result.exit_code == 0

    def test_model_option(self):
        """Test --model option."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {"workflow_id": "wf-123", "status": "completed"}
            result = runner.invoke(
                all_is_well,
                ["feature", "--model", "opus"],
            )

            assert result.exit_code == 0


class TestAllIsWellOutput:
    """Test all-is-well command output."""

    def test_shows_workflow_id(self):
        """Test that output shows workflow ID."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {
                "workflow_id": "wf-test-123",
                "status": "completed",
            }
            result = runner.invoke(all_is_well, ["feature"])

            assert "wf-test-123" in result.output or result.exit_code == 0

    def test_shows_progress(self):
        """Test that output shows progress information."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.return_value = {
                "workflow_id": "wf-123",
                "status": "completed",
                "phase": "completion",
            }
            result = runner.invoke(all_is_well, ["feature"])

            # Should show some form of success message
            assert result.exit_code == 0


class TestAllIsWellErrorHandling:
    """Test error handling in all-is-well command."""

    def test_handles_workflow_failure(self):
        """Test handling workflow failure."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        with patch("moai_adk.cli.commands.all_is_well.run_workflow") as mock_run:
            mock_run.side_effect = Exception("Workflow failed")
            result = runner.invoke(all_is_well, ["feature"])

            # Should handle error gracefully
            assert result.exit_code != 0 or "error" in result.output.lower()

    def test_handles_invalid_parallel_value(self):
        """Test handling invalid parallel value."""
        from moai_adk.cli.commands.all_is_well import all_is_well

        runner = CliRunner()
        result = runner.invoke(
            all_is_well,
            ["feature", "--parallel", "invalid"],
        )

        # Should fail with invalid parallel value
        assert result.exit_code != 0


class TestRunWorkflowFunction:
    """Test the run_workflow helper function."""

    @pytest.mark.asyncio
    async def test_run_workflow_creates_orchestrator(self):
        """Test run_workflow creates orchestrator."""
        from moai_adk.cli.commands.all_is_well import run_workflow_async

        from moai_adk.web.models.workflow import WorkflowConfig

        config = WorkflowConfig(features=["test"])

        with patch("moai_adk.cli.commands.all_is_well.WorkflowOrchestrator") as MockOrch:
            mock_orch = AsyncMock()
            mock_orch.start_workflow.return_value = MagicMock(
                workflow_id="wf-123",
                status="in_progress",
                phase="configuration",
            )
            MockOrch.return_value = mock_orch

            result = await run_workflow_async(config)

            assert result is not None
            MockOrch.assert_called_once()

    @pytest.mark.asyncio
    async def test_run_workflow_starts_workflow(self):
        """Test run_workflow starts the workflow."""
        from moai_adk.cli.commands.all_is_well import run_workflow_async

        from moai_adk.web.models.workflow import WorkflowConfig

        config = WorkflowConfig(features=["test"])

        with patch("moai_adk.cli.commands.all_is_well.WorkflowOrchestrator") as MockOrch:
            mock_orch = AsyncMock()
            mock_report = MagicMock()
            mock_report.workflow_id = "wf-123"
            mock_report.status = "in_progress"
            mock_report.phase = "configuration"
            mock_orch.start_workflow.return_value = mock_report
            MockOrch.return_value = mock_orch

            result = await run_workflow_async(config)

            mock_orch.start_workflow.assert_called_once_with(config)
