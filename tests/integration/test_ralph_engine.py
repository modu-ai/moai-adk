"""Integration tests for Ralph Engine.

Tests the unified Ralph Engine that integrates LSP, AST-grep,
and Loop Controller for intelligent code quality assurance.
"""

from __future__ import annotations

from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from moai_adk.astgrep import ScanConfig, ScanResult
from moai_adk.loop import LoopStatus, LoopStorage
from moai_adk.lsp import Diagnostic, DiagnosticSeverity, Position, Range
from moai_adk.ralph import RalphEngine
from moai_adk.ralph.engine import DiagnosisResult


class TestRalphEngineInitialization:
    """Test RalphEngine initialization."""

    def test_default_initialization(self, tmp_path: Path):
        """Test engine initializes with default settings."""
        engine = RalphEngine(project_root=tmp_path)

        assert engine.project_root == tmp_path
        assert engine.lsp_client is not None
        assert engine.ast_analyzer is not None
        assert engine.loop_controller is not None

    def test_custom_config_initialization(self, tmp_path: Path):
        """Test engine initializes with custom configuration."""
        config = ScanConfig(
            rules_path=str(tmp_path / "rules"),
            exclude_patterns=["*.test.py"],
        )
        storage = LoopStorage(storage_dir=tmp_path / "loop")

        engine = RalphEngine(
            project_root=tmp_path,
            ast_config=config,
            loop_storage=storage,
        )

        assert engine.ast_analyzer.config == config
        assert engine.loop_controller.storage == storage

    def test_repr(self, tmp_path: Path):
        """Test string representation."""
        engine = RalphEngine(project_root=tmp_path)
        assert str(tmp_path) in repr(engine)
        assert "RalphEngine" in repr(engine)


class TestDiagnosisResult:
    """Test DiagnosisResult class."""

    def test_total_issues_calculation(self):
        """Test total issues is sum of LSP and AST-grep issues."""
        lsp_diags = [
            Diagnostic(
                range=Range(start=Position(0, 0), end=Position(0, 10)),
                severity=DiagnosticSeverity.ERROR,
                code="E001",
                source="test",
                message="Error 1",
            ),
            Diagnostic(
                range=Range(start=Position(1, 0), end=Position(1, 10)),
                severity=DiagnosticSeverity.WARNING,
                code="W001",
                source="test",
                message="Warning 1",
            ),
        ]
        ast_result = ScanResult(
            file_path="test.py",
            matches=[MagicMock(severity="warning")],
            scan_time_ms=10.0,
            language="python",
        )

        result = DiagnosisResult(
            file_path="test.py",
            lsp_diagnostics=lsp_diags,
            ast_result=ast_result,
        )

        assert result.total_issues == 3

    def test_has_errors_with_lsp_error(self):
        """Test has_errors returns True with LSP error."""
        lsp_diags = [
            Diagnostic(
                range=Range(start=Position(0, 0), end=Position(0, 10)),
                severity=DiagnosticSeverity.ERROR,
                code="E001",
                source="test",
                message="Error",
            ),
        ]
        ast_result = ScanResult(
            file_path="test.py",
            matches=[],
            scan_time_ms=10.0,
            language="python",
        )

        result = DiagnosisResult(
            file_path="test.py",
            lsp_diagnostics=lsp_diags,
            ast_result=ast_result,
        )

        assert result.has_errors is True

    def test_has_errors_with_ast_error(self):
        """Test has_errors returns True with AST-grep error."""
        ast_result = ScanResult(
            file_path="test.py",
            matches=[MagicMock(severity="error")],
            scan_time_ms=10.0,
            language="python",
        )

        result = DiagnosisResult(
            file_path="test.py",
            lsp_diagnostics=[],
            ast_result=ast_result,
        )

        assert result.has_errors is True

    def test_has_errors_false_with_warnings_only(self):
        """Test has_errors returns False with only warnings."""
        lsp_diags = [
            Diagnostic(
                range=Range(start=Position(0, 0), end=Position(0, 10)),
                severity=DiagnosticSeverity.WARNING,
                code="W001",
                source="test",
                message="Warning",
            ),
        ]
        ast_result = ScanResult(
            file_path="test.py",
            matches=[MagicMock(severity="warning")],
            scan_time_ms=10.0,
            language="python",
        )

        result = DiagnosisResult(
            file_path="test.py",
            lsp_diagnostics=lsp_diags,
            ast_result=ast_result,
        )

        assert result.has_errors is False

    def test_to_dict(self):
        """Test DiagnosisResult serialization to dict."""
        lsp_diags = [
            Diagnostic(
                range=Range(start=Position(0, 0), end=Position(0, 10)),
                severity=DiagnosticSeverity.ERROR,
                message="Test error",
                source="pylsp",
                code="E001",
            ),
        ]
        ast_result = ScanResult(
            file_path="test.py",
            matches=[],
            scan_time_ms=10.0,
            language="python",
        )

        result = DiagnosisResult(
            file_path="test.py",
            lsp_diagnostics=lsp_diags,
            ast_result=ast_result,
        )

        data = result.to_dict()

        assert data["file_path"] == "test.py"
        assert len(data["lsp_diagnostics"]) == 1
        assert data["lsp_diagnostics"][0]["message"] == "Test error"
        assert data["total_issues"] == 1
        assert data["has_errors"] is True


class TestRalphEngineDiagnosis:
    """Test RalphEngine diagnosis operations."""

    @pytest.mark.asyncio
    async def test_diagnose_file(self, tmp_path: Path):
        """Test file diagnosis combines LSP and AST-grep results."""
        # Create test file
        test_file = tmp_path / "test.py"
        test_file.write_text("x: int = 'hello'  # type error\n")

        engine = RalphEngine(project_root=tmp_path)

        # Mock LSP client to return a diagnostic
        mock_diagnostic = Diagnostic(
            range=Range(start=Position(0, 0), end=Position(0, 17)),
            severity=DiagnosticSeverity.ERROR,
            code="E001",
            source="pylsp",
            message="Type mismatch",
        )

        with patch.object(
            engine.lsp_client,
            "get_diagnostics",
            new_callable=AsyncMock,
            return_value=[mock_diagnostic],
        ):
            result = await engine.diagnose_file(str(test_file))

        assert result.file_path == str(test_file)
        assert len(result.lsp_diagnostics) == 1
        assert result.lsp_diagnostics[0].message == "Type mismatch"

    @pytest.mark.asyncio
    async def test_diagnose_file_not_found(self, tmp_path: Path):
        """Test diagnosis raises error for non-existent file."""
        engine = RalphEngine(project_root=tmp_path)

        with pytest.raises(FileNotFoundError):
            await engine.diagnose_file("nonexistent.py")

    @pytest.mark.asyncio
    async def test_diagnose_file_relative_path(self, tmp_path: Path):
        """Test diagnosis resolves relative paths."""
        # Create test file
        test_file = tmp_path / "src" / "main.py"
        test_file.parent.mkdir(parents=True)
        test_file.write_text("print('hello')\n")

        engine = RalphEngine(project_root=tmp_path)

        with patch.object(
            engine.lsp_client,
            "get_diagnostics",
            new_callable=AsyncMock,
            return_value=[],
        ):
            result = await engine.diagnose_file("src/main.py")

        assert result.file_path == str(test_file)


class TestRalphEngineFeedbackLoop:
    """Test RalphEngine feedback loop operations."""

    def test_start_feedback_loop(self, tmp_path: Path):
        """Test starting a new feedback loop."""
        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        state = engine.start_feedback_loop(
            promise="Fix all errors",
            max_iterations=5,
        )

        assert state.promise == "Fix all errors"
        assert state.max_iterations == 5
        assert state.status == LoopStatus.RUNNING
        assert state.current_iteration == 0

    def test_cancel_loop(self, tmp_path: Path):
        """Test cancelling an active loop."""
        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        # Start a loop
        state = engine.start_feedback_loop(promise="Test")

        # Cancel it
        result = engine.cancel_loop(state.loop_id)
        assert result is True

        # Verify it's cancelled
        updated_state = engine.get_loop_status(state.loop_id)
        assert updated_state is not None
        assert updated_state.status == LoopStatus.CANCELLED

    def test_cancel_nonexistent_loop(self, tmp_path: Path):
        """Test cancelling a non-existent loop returns False."""
        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        result = engine.cancel_loop("nonexistent-loop-id")
        assert result is False

    def test_get_active_loop(self, tmp_path: Path):
        """Test getting the active loop."""
        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        # No active loop initially
        assert engine.get_active_loop() is None

        # Start a loop
        state = engine.start_feedback_loop(promise="Test")

        # Now there's an active loop
        active = engine.get_active_loop()
        assert active is not None
        assert active.loop_id == state.loop_id

    def test_get_loop_status(self, tmp_path: Path):
        """Test getting status of a specific loop."""
        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        state = engine.start_feedback_loop(promise="Test")

        status = engine.get_loop_status(state.loop_id)
        assert status is not None
        assert status.loop_id == state.loop_id
        assert status.promise == "Test"

    def test_check_completion_with_no_issues(self, tmp_path: Path):
        """Test completion check with no remaining issues."""
        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        state = engine.start_feedback_loop(promise="Test")

        completion = engine.check_completion(state)

        # No history means no issues tracked yet
        assert completion.is_complete is True
        assert completion.remaining_issues == 0


class TestRalphEngineLifecycle:
    """Test RalphEngine lifecycle operations."""

    @pytest.mark.asyncio
    async def test_shutdown(self, tmp_path: Path):
        """Test engine shutdown cleans up resources."""
        engine = RalphEngine(project_root=tmp_path)

        with patch.object(
            engine.lsp_client,
            "cleanup",
            new_callable=AsyncMock,
        ) as mock_cleanup:
            await engine.shutdown()
            mock_cleanup.assert_called_once()

    def test_is_ast_grep_available(self, tmp_path: Path):
        """Test checking AST-grep availability."""
        engine = RalphEngine(project_root=tmp_path)

        # The result depends on whether sg is installed
        # We just verify the method runs without error
        result = engine.is_ast_grep_available()
        assert isinstance(result, bool)


class TestRalphEngineIntegration:
    """Integration tests for the complete Ralph workflow."""

    @pytest.mark.asyncio
    async def test_full_workflow_diagnosis_and_loop(self, tmp_path: Path):
        """Test complete workflow: diagnose file, start loop, check completion."""
        # Create test file
        test_file = tmp_path / "test.py"
        test_file.write_text("def hello():\n    pass\n")

        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        # Mock LSP to return no diagnostics
        with patch.object(
            engine.lsp_client,
            "get_diagnostics",
            new_callable=AsyncMock,
            return_value=[],
        ):
            # Step 1: Diagnose file
            diagnosis = await engine.diagnose_file(str(test_file))
            assert diagnosis.total_issues == 0

            # Step 2: Start feedback loop
            state = engine.start_feedback_loop(
                promise="Ensure no errors",
                max_iterations=3,
            )
            assert state.status == LoopStatus.RUNNING

            # Step 3: Check completion
            completion = engine.check_completion(state)
            assert completion.is_complete is True

            # Step 4: Shutdown
            await engine.shutdown()

    @pytest.mark.asyncio
    async def test_run_feedback_iteration(self, tmp_path: Path):
        """Test running a feedback iteration."""
        storage = LoopStorage(storage_dir=tmp_path / "loop")
        engine = RalphEngine(project_root=tmp_path, loop_storage=storage)

        # Start a loop
        state = engine.start_feedback_loop(promise="Test iteration")

        # Run an iteration
        result = await engine.run_feedback_iteration(state)

        assert result is not None
        assert result.feedback_text is not None
        assert result.lsp_diagnostics is not None
        assert result.ast_issues is not None

    @pytest.mark.asyncio
    async def test_diagnose_project(self, tmp_path: Path):
        """Test diagnosing an entire project."""
        # Create test files
        src_dir = tmp_path / "src"
        src_dir.mkdir()

        (src_dir / "main.py").write_text("print('hello')\n")
        (src_dir / "utils.py").write_text("def util(): pass\n")

        engine = RalphEngine(project_root=tmp_path)

        with patch.object(
            engine.lsp_client,
            "get_diagnostics",
            new_callable=AsyncMock,
            return_value=[],
        ):
            results = await engine.diagnose_project()

        # Results may be empty if no AST-grep matches
        assert isinstance(results, dict)
