# Loop Feedback Generator
"""Feedback generation for Claude from LSP and AST-grep results.

Provides functionality to transform diagnostic and AST analysis results
into actionable feedback for Claude during feedback loop iterations.
"""

from __future__ import annotations

from typing import Any

from moai_adk.astgrep.models import ASTMatch
from moai_adk.loop.state import ASTIssueSnapshot, DiagnosticSnapshot
from moai_adk.lsp.models import Diagnostic, DiagnosticSeverity

# Security-related rule IDs that should be prioritized
SECURITY_RULES = frozenset(
    {
        "sql-injection",
        "xss",
        "command-injection",
        "path-traversal",
        "eval-usage",
        "no-eval",
        "hardcoded-secret",
        "insecure-random",
    }
)


class FeedbackGenerator:
    """Generator for Claude feedback from LSP and AST-grep results.

    Transforms diagnostic snapshots and AST issue snapshots into
    human-readable feedback text and prioritized issue lists.

    Attributes:
        max_priority_issues: Maximum number of priority issues to return.
        include_suggestions: Whether to include fix suggestions.
    """

    def __init__(
        self,
        max_priority_issues: int = 10,
        include_suggestions: bool = True,
    ):
        """Initialize the feedback generator.

        Args:
            max_priority_issues: Maximum number of priority issues to return.
            include_suggestions: Whether to include fix suggestions.
        """
        self.max_priority_issues = max_priority_issues
        self.include_suggestions = include_suggestions

    def generate_feedback(
        self,
        lsp_snapshot: DiagnosticSnapshot,
        ast_snapshot: ASTIssueSnapshot,
        previous_snapshot: tuple[DiagnosticSnapshot, ASTIssueSnapshot] | None = None,
    ) -> str:
        """Generate feedback text from LSP and AST snapshots.

        Creates a human-readable summary of the current diagnostic state,
        optionally comparing with a previous snapshot to show progress.

        Args:
            lsp_snapshot: Current LSP diagnostic snapshot.
            ast_snapshot: Current AST issue snapshot.
            previous_snapshot: Optional previous snapshots for comparison.

        Returns:
            Formatted feedback text for Claude.
        """
        lines: list[str] = []

        # Check if there are no issues
        total_lsp = lsp_snapshot.total_issues()
        total_ast = ast_snapshot.total_issues

        if total_lsp == 0 and total_ast == 0:
            lines.append("No issues found. Code is clean.")
            return "\n".join(lines)

        # LSP Diagnostics summary
        lines.append("## LSP Diagnostics Summary")
        lines.append(f"- Errors: {lsp_snapshot.error_count}")
        lines.append(f"- Warnings: {lsp_snapshot.warning_count}")
        lines.append(f"- Info: {lsp_snapshot.info_count}")
        if lsp_snapshot.files_affected:
            lines.append(f"- Files affected: {', '.join(lsp_snapshot.files_affected)}")
        lines.append("")

        # AST Issues summary
        lines.append("## AST-grep Issues Summary")
        lines.append(f"- Total issues: {ast_snapshot.total_issues}")
        if ast_snapshot.by_severity:
            severity_parts = [f"{count} {sev}" for sev, count in ast_snapshot.by_severity.items()]
            lines.append(f"- By severity: {', '.join(severity_parts)}")
        if ast_snapshot.by_rule:
            rule_parts = [f"{rule}: {count}" for rule, count in ast_snapshot.by_rule.items()]
            lines.append(f"- By rule: {', '.join(rule_parts)}")
        if ast_snapshot.files_affected:
            lines.append(f"- Files affected: {', '.join(ast_snapshot.files_affected)}")
        lines.append("")

        # Comparison with previous snapshot
        if previous_snapshot is not None:
            prev_lsp, prev_ast = previous_snapshot
            lines.append("## Progress")

            lsp_diff = prev_lsp.total_issues() - total_lsp
            ast_diff = prev_ast.total_issues - total_ast

            if lsp_diff > 0:
                lines.append(f"- LSP issues reduced by {lsp_diff}")
            elif lsp_diff < 0:
                lines.append(f"- LSP issues increased by {-lsp_diff}")
            else:
                lines.append("- LSP issues unchanged")

            if ast_diff > 0:
                lines.append(f"- AST issues fixed: {ast_diff}")
            elif ast_diff < 0:
                lines.append(f"- AST issues increased by {-ast_diff}")
            else:
                lines.append("- AST issues unchanged")

            total_improvement = lsp_diff + ast_diff
            if total_improvement > 0:
                lines.append(f"- Overall improvement: {total_improvement} issues fixed")
            lines.append("")

        # Priority guidance
        if lsp_snapshot.error_count > 0 or ast_snapshot.by_severity.get("error", 0) > 0:
            lines.append("## Priority")
            lines.append("Focus on fixing errors first, then warnings.")

        return "\n".join(lines)

    def format_for_hook(self, feedback: str) -> dict[str, Any]:
        """Format feedback text for hook output.

        Converts the feedback text into a structured dictionary format
        suitable for Claude Code hook consumption.

        Args:
            feedback: The feedback text to format.

        Returns:
            Dictionary with feedback and metadata.
        """
        return {
            "type": "loop_feedback",
            "source": "moai_loop_controller",
            "feedback": feedback,
            "text": feedback,
        }

    def prioritize_issues(
        self,
        diagnostics: list[Diagnostic],
        ast_matches: list[ASTMatch],
    ) -> list[str]:
        """Prioritize and format issues for Claude.

        Sorts issues by severity and type, with security issues first,
        then errors, then warnings. Returns formatted issue strings.

        Args:
            diagnostics: List of LSP diagnostics.
            ast_matches: List of AST-grep matches.

        Returns:
            List of formatted issue strings, sorted by priority.
        """
        if not diagnostics and not ast_matches:
            return []

        priority_issues: list[tuple[int, str]] = []

        # Process AST matches (security issues get highest priority)
        for match in ast_matches:
            is_security = match.rule_id.lower() in SECURITY_RULES
            is_error = match.severity.lower() == "error"

            # Priority: lower is higher priority
            # Security errors: 0, Security warnings: 1, Regular errors: 2, Regular warnings: 3
            if is_security and is_error:
                priority = 0
            elif is_security:
                priority = 1
            elif is_error:
                priority = 2
            else:
                priority = 3

            line_num = match.range.start.line + 1  # Convert to 1-based
            issue_text = f"[{match.severity.upper()}] {match.message} at {match.file_path}:{line_num}"
            if self.include_suggestions and match.suggested_fix:
                issue_text += f" (Fix: {match.suggested_fix})"

            priority_issues.append((priority, issue_text))

        # Process LSP diagnostics
        for diag in diagnostics:
            is_error = diag.severity == DiagnosticSeverity.ERROR
            is_warning = diag.severity == DiagnosticSeverity.WARNING

            # Priority: errors: 2, warnings: 3, info/hint: 4
            if is_error:
                priority = 2
            elif is_warning:
                priority = 3
            else:
                priority = 4

            line_num = diag.range.start.line + 1  # Convert to 1-based
            severity_name = diag.severity.name
            issue_text = f"[{severity_name}] {diag.message} at line {line_num}"
            if diag.source:
                issue_text = f"[{severity_name}] ({diag.source}) {diag.message} at line {line_num}"

            priority_issues.append((priority, issue_text))

        # Sort by priority (lower number = higher priority)
        priority_issues.sort(key=lambda x: x[0])

        # Extract just the issue texts, limited to max
        return [issue_text for _, issue_text in priority_issues[: self.max_priority_issues]]
