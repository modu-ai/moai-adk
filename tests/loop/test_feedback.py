# Loop Feedback Tests - RED Phase
"""Tests for feedback generation (FeedbackGenerator)."""

from datetime import datetime, timezone

from moai_adk.loop.feedback import FeedbackGenerator

from moai_adk.astgrep.models import ASTMatch
from moai_adk.loop.state import (
    ASTIssueSnapshot,
    DiagnosticSnapshot,
)
from moai_adk.lsp.models import Diagnostic, DiagnosticSeverity, Position, Range


class TestFeedbackGeneratorInit:
    """Tests for FeedbackGenerator initialization."""

    def test_create_feedback_generator(self):
        """FeedbackGenerator should be creatable with default settings."""
        generator = FeedbackGenerator()
        assert generator is not None

    def test_create_feedback_generator_with_options(self):
        """FeedbackGenerator should accept configuration options."""
        generator = FeedbackGenerator(
            max_priority_issues=5,
            include_suggestions=True,
        )
        assert generator.max_priority_issues == 5
        assert generator.include_suggestions is True


class TestFeedbackGeneratorGenerateFeedback:
    """Tests for FeedbackGenerator.generate_feedback method."""

    def test_generate_feedback_basic(self):
        """FeedbackGenerator should generate feedback from snapshots."""
        generator = FeedbackGenerator()
        now = datetime.now(timezone.utc)

        lsp_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=2,
            warning_count=3,
            info_count=1,
            files_affected=["main.py", "utils.py"],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=4,
            by_severity={"warning": 3, "info": 1},
            by_rule={"no-console": 2, "no-eval": 2},
            files_affected=["app.js"],
        )

        feedback = generator.generate_feedback(lsp_snapshot, ast_snapshot)

        assert isinstance(feedback, str)
        assert len(feedback) > 0
        # Should mention the counts
        assert "2" in feedback  # error count
        assert "error" in feedback.lower()

    def test_generate_feedback_no_issues(self):
        """FeedbackGenerator should handle no issues case."""
        generator = FeedbackGenerator()
        now = datetime.now(timezone.utc)

        lsp_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=0,
            warning_count=0,
            info_count=0,
            files_affected=[],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=0,
            by_severity={},
            by_rule={},
            files_affected=[],
        )

        feedback = generator.generate_feedback(lsp_snapshot, ast_snapshot)

        assert isinstance(feedback, str)
        # Should indicate no issues
        assert "no" in feedback.lower() or "0" in feedback or "clean" in feedback.lower()

    def test_generate_feedback_with_previous_snapshot(self):
        """FeedbackGenerator should compare with previous snapshot."""
        generator = FeedbackGenerator()
        now = datetime.now(timezone.utc)

        current_lsp = DiagnosticSnapshot(
            timestamp=now,
            error_count=2,
            warning_count=3,
            info_count=0,
            files_affected=["main.py"],
        )
        current_ast = ASTIssueSnapshot(
            timestamp=now,
            total_issues=3,
            by_severity={"warning": 3},
            by_rule={"no-console": 3},
            files_affected=["app.js"],
        )
        previous_lsp = DiagnosticSnapshot(
            timestamp=now,
            error_count=5,
            warning_count=5,
            info_count=2,
            files_affected=["main.py", "other.py"],
        )
        previous_ast = ASTIssueSnapshot(
            timestamp=now,
            total_issues=8,
            by_severity={"warning": 5, "error": 3},
            by_rule={"no-console": 5, "no-eval": 3},
            files_affected=["app.js", "utils.js"],
        )

        feedback = generator.generate_feedback(
            current_lsp,
            current_ast,
            previous_snapshot=(previous_lsp, previous_ast),
        )

        assert isinstance(feedback, str)
        # Should indicate improvement
        assert "improve" in feedback.lower() or "reduce" in feedback.lower() or "fix" in feedback.lower()

    def test_generate_feedback_errors_only(self):
        """FeedbackGenerator should highlight errors prominently."""
        generator = FeedbackGenerator()
        now = datetime.now(timezone.utc)

        lsp_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=5,
            warning_count=0,
            info_count=0,
            files_affected=["main.py"],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=2,
            by_severity={"error": 2},
            by_rule={"sql-injection": 2},
            files_affected=["db.py"],
        )

        feedback = generator.generate_feedback(lsp_snapshot, ast_snapshot)

        # Should emphasize errors
        assert "error" in feedback.lower()
        assert "5" in feedback  # LSP errors
        assert "2" in feedback  # AST errors


class TestFeedbackGeneratorFormatForHook:
    """Tests for FeedbackGenerator.format_for_hook method."""

    def test_format_for_hook(self):
        """FeedbackGenerator should format feedback for hook output."""
        generator = FeedbackGenerator()
        feedback_text = "Found 5 errors and 3 warnings. Focus on fixing errors first."

        result = generator.format_for_hook(feedback_text)

        assert isinstance(result, dict)
        assert "feedback" in result or "message" in result or "text" in result
        # Should contain the original feedback text
        feedback_content = result.get("feedback") or result.get("message") or result.get("text", "")
        assert "5 errors" in feedback_content

    def test_format_for_hook_includes_metadata(self):
        """FeedbackGenerator should include metadata in hook output."""
        generator = FeedbackGenerator()
        feedback_text = "No issues found. Code is clean."

        result = generator.format_for_hook(feedback_text)

        assert isinstance(result, dict)
        # Should have type or source field
        assert "type" in result or "source" in result

    def test_format_for_hook_empty_feedback(self):
        """FeedbackGenerator should handle empty feedback."""
        generator = FeedbackGenerator()

        result = generator.format_for_hook("")

        assert isinstance(result, dict)


class TestFeedbackGeneratorPrioritizeIssues:
    """Tests for FeedbackGenerator.prioritize_issues method."""

    def test_prioritize_issues_basic(self):
        """FeedbackGenerator should prioritize issues by severity."""
        generator = FeedbackGenerator()

        diagnostics = [
            Diagnostic(
                range=Range(start=Position(line=10, character=0), end=Position(line=10, character=20)),
                severity=DiagnosticSeverity.ERROR,
                code="E001",
                source="pyright",
                message="Undefined variable 'x'",
            ),
            Diagnostic(
                range=Range(start=Position(line=5, character=0), end=Position(line=5, character=10)),
                severity=DiagnosticSeverity.WARNING,
                code="W001",
                source="pyright",
                message="Unused import",
            ),
        ]
        ast_matches = [
            ASTMatch(
                rule_id="sql-injection",
                severity="error",
                message="Potential SQL injection",
                file_path="db.py",
                range=Range(start=Position(line=20, character=0), end=Position(line=20, character=50)),
                suggested_fix="Use parameterized queries",
            ),
        ]

        priority_list = generator.prioritize_issues(diagnostics, ast_matches)

        assert isinstance(priority_list, list)
        assert len(priority_list) > 0
        # Errors should come first
        assert "error" in priority_list[0].lower() or "sql" in priority_list[0].lower()

    def test_prioritize_issues_empty(self):
        """FeedbackGenerator should handle empty issue lists."""
        generator = FeedbackGenerator()

        priority_list = generator.prioritize_issues([], [])

        assert isinstance(priority_list, list)
        assert len(priority_list) == 0

    def test_prioritize_issues_respects_max_limit(self):
        """FeedbackGenerator should respect max_priority_issues limit."""
        generator = FeedbackGenerator(max_priority_issues=3)

        diagnostics = [
            Diagnostic(
                range=Range(start=Position(line=i, character=0), end=Position(line=i, character=10)),
                severity=DiagnosticSeverity.ERROR,
                code=f"E{i:03d}",
                source="pyright",
                message=f"Error {i}",
            )
            for i in range(10)
        ]

        priority_list = generator.prioritize_issues(diagnostics, [])

        assert len(priority_list) <= 3

    def test_prioritize_issues_security_first(self):
        """FeedbackGenerator should prioritize security issues."""
        generator = FeedbackGenerator()

        diagnostics = [
            Diagnostic(
                range=Range(start=Position(line=1, character=0), end=Position(line=1, character=10)),
                severity=DiagnosticSeverity.ERROR,
                code="E001",
                source="pyright",
                message="Type error",
            ),
        ]
        ast_matches = [
            ASTMatch(
                rule_id="sql-injection",
                severity="error",
                message="SQL injection vulnerability",
                file_path="db.py",
                range=Range(start=Position(line=5, character=0), end=Position(line=5, character=30)),
                suggested_fix="Use parameterized queries",
            ),
        ]

        priority_list = generator.prioritize_issues(diagnostics, ast_matches)

        # Security (AST) issue should be first
        assert "sql" in priority_list[0].lower() or "injection" in priority_list[0].lower()

    def test_prioritize_issues_formats_nicely(self):
        """FeedbackGenerator should format issues as readable strings."""
        generator = FeedbackGenerator()

        diagnostics = [
            Diagnostic(
                range=Range(start=Position(line=10, character=5), end=Position(line=10, character=15)),
                severity=DiagnosticSeverity.ERROR,
                code="E001",
                source="pyright",
                message="Undefined variable 'foo'",
            ),
        ]

        priority_list = generator.prioritize_issues(diagnostics, [])

        assert len(priority_list) == 1
        # Should include file location or line number
        issue_text = priority_list[0]
        assert "10" in issue_text or "line" in issue_text.lower()  # line number
        assert "foo" in issue_text or "Undefined" in issue_text  # message content


class TestFeedbackGeneratorIntegration:
    """Integration tests for FeedbackGenerator."""

    def test_full_workflow(self):
        """FeedbackGenerator should work in a complete workflow."""
        generator = FeedbackGenerator(max_priority_issues=5)
        now = datetime.now(timezone.utc)

        # Create snapshots
        lsp_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=3,
            warning_count=5,
            info_count=2,
            files_affected=["main.py", "utils.py"],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=4,
            by_severity={"error": 2, "warning": 2},
            by_rule={"sql-injection": 2, "no-eval": 2},
            files_affected=["db.py"],
        )

        # Generate feedback
        feedback = generator.generate_feedback(lsp_snapshot, ast_snapshot)
        assert len(feedback) > 0

        # Format for hook
        hook_output = generator.format_for_hook(feedback)
        assert isinstance(hook_output, dict)

        # Create diagnostic and AST match objects for prioritization
        diagnostics = [
            Diagnostic(
                range=Range(start=Position(line=10, character=0), end=Position(line=10, character=20)),
                severity=DiagnosticSeverity.ERROR,
                code="E001",
                source="pyright",
                message="Type error in main.py",
            ),
        ]
        ast_matches = [
            ASTMatch(
                rule_id="sql-injection",
                severity="error",
                message="SQL injection in db.py",
                file_path="db.py",
                range=Range(start=Position(line=5, character=0), end=Position(line=5, character=30)),
                suggested_fix="Use parameterized queries",
            ),
        ]

        # Prioritize issues
        priority_list = generator.prioritize_issues(diagnostics, ast_matches)
        assert len(priority_list) > 0
