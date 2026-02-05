# AST-grep data models
"""Data models for AST-grep analyzer.

These dataclasses represent the core types used for AST-grep operations
including scanning, pattern matching, and code transformation.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from moai_adk.lsp.models import Range


@dataclass
class ASTMatch:
    """A match found by AST-grep analysis.

    Represents a single match from an AST-grep rule or pattern search.

    Attributes:
        rule_id: Identifier of the rule that matched.
        severity: Severity level (error, warning, info, hint).
        message: Human-readable message describing the match.
        file_path: Path to the file containing the match.
        range: The range in the file where the match was found.
        suggested_fix: Optional suggested fix pattern.
    """

    rule_id: str
    severity: str  # error, warning, info, hint
    message: str
    file_path: str
    range: Range
    suggested_fix: str | None


@dataclass
class ScanConfig:
    """Configuration for AST-grep scanning operations.

    Controls how files are scanned, which rules to apply, and what
    patterns to include or exclude.

    Attributes:
        rules_path: Path to custom rules file or directory.
        security_scan: Whether to perform security scanning.
        include_patterns: Glob patterns for files to include.
        exclude_patterns: Glob patterns for files to exclude.
    """

    rules_path: str | None = None
    security_scan: bool = True
    include_patterns: list[str] = field(default_factory=list)
    exclude_patterns: list[str] = field(default_factory=lambda: ["node_modules", ".git", "__pycache__"])


@dataclass
class ScanResult:
    """Result of scanning a single file.

    Contains all matches found in a file along with metadata about
    the scan operation.

    Attributes:
        file_path: Path to the scanned file.
        matches: List of matches found in the file.
        scan_time_ms: Time taken to scan the file in milliseconds.
        language: Programming language of the file.
    """

    file_path: str
    matches: list[ASTMatch]
    scan_time_ms: float
    language: str


@dataclass
class ProjectScanResult:
    """Result of scanning an entire project.

    Aggregates results from scanning multiple files with summary
    statistics.

    Attributes:
        project_path: Root path of the scanned project.
        files_scanned: Number of files scanned.
        total_matches: Total number of matches across all files.
        results_by_file: Mapping of file paths to their scan results.
        summary_by_severity: Count of matches by severity level.
        scan_time_ms: Total time taken to scan the project in milliseconds.
    """

    project_path: str
    files_scanned: int
    total_matches: int
    results_by_file: dict[str, ScanResult]
    summary_by_severity: dict[str, int]
    scan_time_ms: float


@dataclass
class ReplaceResult:
    """Result of a pattern replacement operation.

    Contains information about what was matched and potentially
    modified during a replace operation.

    Attributes:
        pattern: The pattern that was searched for.
        replacement: The replacement pattern used.
        matches_found: Number of matches found.
        files_modified: Number of files that were modified.
        changes: List of individual changes made.
        dry_run: Whether this was a dry run (no actual changes).
    """

    pattern: str
    replacement: str
    matches_found: int
    files_modified: int
    changes: list[dict[str, Any]]  # {file_path, old_code, new_code, range}
    dry_run: bool
