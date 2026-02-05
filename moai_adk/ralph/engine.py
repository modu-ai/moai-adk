# Ralph Engine
"""Ralph Engine - Unified code quality assurance system.

This module provides the main RalphEngine class that integrates LSP,
AST-grep, and the Loop Controller into a cohesive intelligent code
quality system.
"""

from __future__ import annotations

from pathlib import Path
from typing import Any

from moai_adk.astgrep import MoAIASTGrepAnalyzer, ScanConfig, ScanResult
from moai_adk.loop import LoopState, LoopStorage, MoAILoopController
from moai_adk.lsp import Diagnostic, MoAILSPClient


class DiagnosisResult:
    """Result of a file diagnosis combining LSP and AST-grep results.

    Attributes:
        file_path: Path to the diagnosed file.
        lsp_diagnostics: List of LSP diagnostics found.
        ast_result: AST-grep scan result.
        total_issues: Total number of issues found.
    """

    def __init__(
        self,
        file_path: str,
        lsp_diagnostics: list[Diagnostic],
        ast_result: ScanResult,
    ) -> None:
        """Initialize DiagnosisResult.

        Args:
            file_path: Path to the diagnosed file.
            lsp_diagnostics: List of LSP diagnostics.
            ast_result: AST-grep scan result.
        """
        self.file_path = file_path
        self.lsp_diagnostics = lsp_diagnostics
        self.ast_result = ast_result

    @property
    def total_issues(self) -> int:
        """Get total number of issues found."""
        return len(self.lsp_diagnostics) + len(self.ast_result.matches)

    @property
    def has_errors(self) -> bool:
        """Check if there are any error-level issues."""
        # Check LSP diagnostics for errors (severity 1 is Error)
        for diag in self.lsp_diagnostics:
            if diag.severity.value == 1:
                return True

        # Check AST-grep matches for errors
        for match in self.ast_result.matches:
            if match.severity.lower() == "error":
                return True

        return False

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary representation.

        Returns:
            Dictionary with diagnosis results.
        """
        return {
            "file_path": self.file_path,
            "lsp_diagnostics": [
                {
                    "message": d.message,
                    "severity": d.severity.name,
                    "range": {
                        "start": {"line": d.range.start.line, "character": d.range.start.character},
                        "end": {"line": d.range.end.line, "character": d.range.end.character},
                    },
                    "source": d.source,
                    "code": d.code,
                }
                for d in self.lsp_diagnostics
            ],
            "ast_matches": [
                {
                    "rule_id": m.rule_id,
                    "message": m.message,
                    "severity": m.severity,
                    "range": {
                        "start": {"line": m.range.start.line, "character": m.range.start.character},
                        "end": {"line": m.range.end.line, "character": m.range.end.character},
                    },
                    "suggested_fix": m.suggested_fix,
                }
                for m in self.ast_result.matches
            ],
            "total_issues": self.total_issues,
            "has_errors": self.has_errors,
        }


class RalphEngine:
    """MoAI Ralph Engine - LSP + AST-grep + Loop integrated system.

    Provides a unified interface for intelligent code quality assurance
    by combining three core technologies:

    1. LSP (Language Server Protocol): Real-time diagnostics and code intelligence
    2. AST-grep: Structural pattern matching and security scanning
    3. Ralph Loop: Autonomous feedback loops for iterative improvement

    Attributes:
        project_root: Root directory of the project.
        lsp_client: LSP client for diagnostics.
        ast_analyzer: AST-grep analyzer for pattern matching.
        loop_controller: Controller for feedback loops.
    """

    def __init__(
        self,
        project_root: str | Path | None = None,
        ast_config: ScanConfig | None = None,
        loop_storage: LoopStorage | None = None,
    ) -> None:
        """Initialize the Ralph Engine.

        Args:
            project_root: Root directory of the project. Defaults to current directory.
            ast_config: Configuration for AST-grep scanning.
            loop_storage: Storage backend for loop states.
        """
        self.project_root = Path(project_root) if project_root else Path.cwd()
        self.lsp_client = MoAILSPClient(self.project_root)
        self.ast_analyzer = MoAIASTGrepAnalyzer(config=ast_config)
        self.loop_controller = MoAILoopController(storage=loop_storage)

    async def diagnose_file(self, file_path: str) -> DiagnosisResult:
        """Diagnose a file using both LSP and AST-grep.

        Combines LSP diagnostics (type errors, syntax errors, etc.)
        with AST-grep pattern matches (security issues, code smells)
        to provide comprehensive file analysis.

        Args:
            file_path: Path to the file to diagnose.

        Returns:
            DiagnosisResult containing all found issues.

        Raises:
            FileNotFoundError: If the file does not exist.
        """
        full_path = self._resolve_path(file_path)

        if not full_path.exists():
            raise FileNotFoundError(f"File not found: {file_path}")

        # Get LSP diagnostics
        lsp_diagnostics = await self.lsp_client.get_diagnostics(str(full_path))

        # Get AST-grep results
        ast_result = self.ast_analyzer.scan_file(str(full_path))

        return DiagnosisResult(
            file_path=str(full_path),
            lsp_diagnostics=lsp_diagnostics,
            ast_result=ast_result,
        )

    async def diagnose_project(self) -> dict[str, DiagnosisResult]:
        """Diagnose all files in the project.

        Scans the entire project using both LSP and AST-grep,
        returning results organized by file path.

        Returns:
            Dictionary mapping file paths to their diagnosis results.
        """
        results: dict[str, DiagnosisResult] = {}

        # Use AST-grep to scan the entire project first
        project_scan = self.ast_analyzer.scan_project(str(self.project_root))

        # Get diagnosis for each file that had AST-grep matches
        for file_path, scan_result in project_scan.results_by_file.items():
            lsp_diagnostics = await self.lsp_client.get_diagnostics(file_path)
            results[file_path] = DiagnosisResult(
                file_path=file_path,
                lsp_diagnostics=lsp_diagnostics,
                ast_result=scan_result,
            )

        return results

    def start_feedback_loop(
        self,
        promise: str,
        max_iterations: int = 10,
    ) -> LoopState:
        """Start a new Ralph feedback loop.

        Creates an autonomous feedback loop that continuously monitors
        code quality and provides feedback until the promise is fulfilled
        or max iterations is reached.

        Args:
            promise: The completion condition for the loop (e.g., "Fix all errors").
            max_iterations: Maximum number of iterations allowed.

        Returns:
            The initial LoopState for the new loop.
        """
        return self.loop_controller.start_loop(
            promise=promise,
            max_iterations=max_iterations,
        )

    def cancel_loop(self, loop_id: str) -> bool:
        """Cancel an active feedback loop.

        Args:
            loop_id: The ID of the loop to cancel.

        Returns:
            True if the loop was cancelled, False otherwise.
        """
        return self.loop_controller.cancel_loop(loop_id)

    def get_active_loop(self) -> LoopState | None:
        """Get the currently active feedback loop.

        Returns:
            The active LoopState, or None if no loop is active.
        """
        return self.loop_controller.get_active_loop()

    def get_loop_status(self, loop_id: str) -> LoopState | None:
        """Get the status of a specific loop.

        Args:
            loop_id: The ID of the loop to query.

        Returns:
            The LoopState, or None if not found.
        """
        return self.loop_controller.get_loop_status(loop_id)

    async def run_feedback_iteration(self, state: LoopState):
        """Run a single feedback loop iteration.

        Gathers current diagnostics and generates feedback
        for the given loop state.

        Args:
            state: The current loop state.

        Returns:
            FeedbackResult containing the iteration results.
        """
        return await self.loop_controller.run_feedback_loop(state)

    def check_completion(self, state: LoopState):
        """Check if a loop's completion condition is met.

        Args:
            state: The loop state to check.

        Returns:
            CompletionResult indicating completion status and progress.
        """
        return self.loop_controller.check_completion(state)

    def is_ast_grep_available(self) -> bool:
        """Check if AST-grep (sg CLI) is available.

        Returns:
            True if sg CLI is available, False otherwise.
        """
        return self.ast_analyzer.is_sg_available()

    async def shutdown(self) -> None:
        """Shutdown the engine and cleanup resources.

        Stops all LSP servers and releases resources.
        """
        await self.lsp_client.cleanup()

    def _resolve_path(self, file_path: str) -> Path:
        """Resolve a file path relative to project root.

        Args:
            file_path: Path to resolve (absolute or relative).

        Returns:
            Resolved absolute Path.
        """
        path = Path(file_path)
        if path.is_absolute():
            return path
        return self.project_root / path

    def __repr__(self) -> str:
        """Return string representation."""
        return f"RalphEngine(project_root={self.project_root})"
