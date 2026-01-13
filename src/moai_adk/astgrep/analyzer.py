# AST-grep analyzer
"""MoAI AST-grep analyzer for code scanning and transformation.

This module provides the main analyzer class that wraps the ast-grep (sg) CLI
for AST-based code analysis, pattern searching, and code transformation.
"""

from __future__ import annotations

import fnmatch
import json
import subprocess
import time
from pathlib import Path
from typing import Any

from moai_adk.astgrep.models import (
    ASTMatch,
    ProjectScanResult,
    ReplaceResult,
    ScanConfig,
    ScanResult,
)
from moai_adk.lsp.models import Position, Range

# Language mapping from file extensions to ast-grep language names
EXTENSION_TO_LANGUAGE: dict[str, str] = {
    ".py": "python",
    ".js": "javascript",
    ".mjs": "javascript",
    ".cjs": "javascript",
    ".jsx": "javascriptreact",
    ".ts": "typescript",
    ".mts": "typescript",
    ".cts": "typescript",
    ".tsx": "typescriptreact",
    ".go": "go",
    ".rs": "rust",
    ".java": "java",
    ".kt": "kotlin",
    ".kts": "kotlin",
    ".rb": "ruby",
    ".c": "c",
    ".h": "c",
    ".cpp": "cpp",
    ".cc": "cpp",
    ".cxx": "cpp",
    ".hpp": "cpp",
    ".cs": "csharp",
    ".swift": "swift",
    ".lua": "lua",
    ".html": "html",
    ".vue": "vue",
    ".svelte": "svelte",
}


class MoAIASTGrepAnalyzer:
    """AST-grep based code analyzer.

    Provides methods for scanning files and projects, searching for patterns,
    and performing code transformations using AST-grep (sg CLI).

    Attributes:
        config: Scanning configuration.
    """

    def __init__(self, config: ScanConfig | None = None) -> None:
        """Initialize the analyzer.

        Args:
            config: Optional scanning configuration. Uses defaults if not provided.
        """
        self.config = config or ScanConfig()
        self._sg_available: bool | None = None

    def is_sg_available(self) -> bool:
        """Check if the sg (ast-grep) CLI is available.

        Returns:
            True if sg CLI is available, False otherwise.
        """
        if self._sg_available is None:
            try:
                result = subprocess.run(
                    ["sg", "--version"],
                    capture_output=True,
                    text=True,
                    timeout=5,
                )
                self._sg_available = result.returncode == 0
            except (FileNotFoundError, subprocess.TimeoutExpired):
                self._sg_available = False
        return self._sg_available

    def _detect_language(self, file_path: str) -> str:
        """Detect the programming language from file extension.

        Args:
            file_path: Path to the file.

        Returns:
            Language name for ast-grep.
        """
        ext = Path(file_path).suffix.lower()
        return EXTENSION_TO_LANGUAGE.get(ext, "text")

    def _should_include_file(self, file_path: Path, config: ScanConfig) -> bool:
        """Check if a file should be included in scanning.

        Args:
            file_path: Path to check.
            config: Scan configuration with include/exclude patterns.

        Returns:
            True if file should be included, False otherwise.
        """
        file_str = str(file_path)

        # Check exclude patterns
        for pattern in config.exclude_patterns:
            if pattern in file_str or fnmatch.fnmatch(file_path.name, pattern):
                return False

        # Check include patterns (if specified)
        if config.include_patterns:
            for pattern in config.include_patterns:
                if fnmatch.fnmatch(file_path.name, pattern):
                    return True
            return False

        # Check if file has a supported extension
        return file_path.suffix.lower() in EXTENSION_TO_LANGUAGE

    def scan_file(self, file_path: str, config: ScanConfig | None = None) -> ScanResult:
        """Scan a single file for issues.

        Args:
            file_path: Path to the file to scan.
            config: Optional config override for this scan.

        Returns:
            ScanResult with matches found in the file.

        Raises:
            FileNotFoundError: If the file does not exist.
        """
        path = Path(file_path)
        if not path.exists():
            raise FileNotFoundError(f"File not found: {file_path}")

        scan_config = config or self.config
        language = self._detect_language(file_path)
        start_time = time.time()
        matches: list[ASTMatch] = []

        if self.is_sg_available():
            try:
                matches = self._run_sg_scan(file_path, language, scan_config)
            except (subprocess.SubprocessError, FileNotFoundError):
                # Graceful degradation - return empty result
                pass

        scan_time_ms = (time.time() - start_time) * 1000

        return ScanResult(
            file_path=file_path,
            matches=matches,
            scan_time_ms=scan_time_ms,
            language=language,
        )

    def _run_sg_scan(self, file_path: str, language: str, config: ScanConfig) -> list[ASTMatch]:
        """Run sg scan command on a file.

        Args:
            file_path: Path to scan.
            language: Programming language.
            config: Scan configuration.

        Returns:
            List of ASTMatch objects.
        """
        cmd = ["sg", "scan", "--json", file_path]

        if config.rules_path:
            cmd.extend(["--config", config.rules_path])

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=60,
            )

            if result.returncode == 0 and result.stdout:
                return self._parse_sg_output(result.stdout, file_path)
        except (subprocess.SubprocessError, json.JSONDecodeError):
            pass

        return []

    def _parse_sg_output(self, output: str, file_path: str) -> list[ASTMatch]:
        """Parse sg JSON output into ASTMatch objects.

        Args:
            output: JSON output from sg command.
            file_path: Path to the scanned file.

        Returns:
            List of ASTMatch objects.
        """
        matches: list[ASTMatch] = []

        try:
            data = json.loads(output)
            if isinstance(data, list):
                for item in data:
                    # Skip null or None items
                    if item is None:
                        continue
                    match = self._parse_sg_match(item, file_path)
                    if match:
                        matches.append(match)
        except json.JSONDecodeError:
            pass

        return matches

    def _parse_sg_match(self, item: dict[str, Any], file_path: str) -> ASTMatch | None:
        """Parse a single match from sg output.

        Args:
            item: Match dictionary from sg output.
            file_path: Path to the file.

        Returns:
            ASTMatch object or None if parsing fails.
        """
        # Validate item is a dictionary
        if not isinstance(item, dict):
            return None

        try:
            rule_id = item.get("ruleId", item.get("rule_id", "custom-pattern"))
            severity = item.get("severity", "warning")
            message = item.get("message", "Pattern match found")

            # Parse range - validate it's a dict
            range_data = item.get("range")
            if not isinstance(range_data, dict):
                return None

            start = range_data.get("start", {"line": 0, "column": 0})
            end = range_data.get("end", {"line": 0, "column": 0})

            # Validate start and end are dicts
            if not isinstance(start, dict) or not isinstance(end, dict):
                return None

            match_range = Range(
                start=Position(
                    line=start.get("line", 0),
                    character=start.get("column", start.get("character", 0)),
                ),
                end=Position(
                    line=end.get("line", 0),
                    character=end.get("column", end.get("character", 0)),
                ),
            )

            suggested_fix = item.get("fix", item.get("suggested_fix"))

            return ASTMatch(
                rule_id=rule_id,
                severity=severity,
                message=message,
                file_path=file_path,
                range=match_range,
                suggested_fix=suggested_fix,
            )
        except (KeyError, TypeError, AttributeError):
            return None

    def scan_project(self, project_path: str, config: ScanConfig | None = None) -> ProjectScanResult:
        """Scan an entire project for issues.

        Args:
            project_path: Root path of the project to scan.
            config: Optional config override for this scan.

        Returns:
            ProjectScanResult with aggregated results.

        Raises:
            FileNotFoundError: If the project path does not exist.
        """
        path = Path(project_path)
        if not path.exists():
            raise FileNotFoundError(f"Project path not found: {project_path}")

        scan_config = config or self.config
        start_time = time.time()

        results_by_file: dict[str, ScanResult] = {}
        summary_by_severity: dict[str, int] = {
            "error": 0,
            "warning": 0,
            "info": 0,
            "hint": 0,
        }
        total_matches = 0

        # Find all files to scan
        for file_path in path.rglob("*"):
            if not file_path.is_file():
                continue

            if not self._should_include_file(file_path, scan_config):
                continue

            result = self.scan_file(str(file_path), scan_config)

            if result.matches:
                results_by_file[str(file_path)] = result
                total_matches += len(result.matches)

                for match in result.matches:
                    severity = match.severity.lower()
                    if severity in summary_by_severity:
                        summary_by_severity[severity] += 1

        scan_time_ms = (time.time() - start_time) * 1000

        return ProjectScanResult(
            project_path=project_path,
            files_scanned=len(list(path.rglob("*"))),
            total_matches=total_matches,
            results_by_file=results_by_file,
            summary_by_severity=summary_by_severity,
            scan_time_ms=scan_time_ms,
        )

    def pattern_search(self, pattern: str, language: str, path: str) -> list[ASTMatch]:
        """Search for a custom pattern in files.

        Args:
            pattern: AST-grep pattern to search for.
            language: Programming language for the pattern.
            path: File or directory path to search in.

        Returns:
            List of ASTMatch objects for found matches.

        Raises:
            FileNotFoundError: If the path does not exist.
        """
        search_path = Path(path)
        if not search_path.exists():
            raise FileNotFoundError(f"Path not found: {path}")

        matches: list[ASTMatch] = []

        if not self.is_sg_available():
            return matches

        try:
            cmd = [
                "sg",
                "run",
                "--pattern",
                pattern,
                "--lang",
                language,
                "--json",
                path,
            ]

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=120,
            )

            if result.returncode == 0 and result.stdout:
                matches = self._parse_pattern_search_output(result.stdout, pattern)
        except (subprocess.SubprocessError, json.JSONDecodeError):
            pass

        return matches

    def _parse_pattern_search_output(self, output: str, pattern: str) -> list[ASTMatch]:
        """Parse pattern search output from sg.

        Args:
            output: JSON output from sg run command.
            pattern: The pattern that was searched.

        Returns:
            List of ASTMatch objects.
        """
        matches: list[ASTMatch] = []

        try:
            data = json.loads(output)
            if isinstance(data, list):
                for item in data:
                    file_path = item.get("file", item.get("path", ""))
                    match = self._parse_sg_match(item, file_path)
                    if match:
                        # Override rule_id for pattern searches
                        match.rule_id = f"pattern:{pattern[:30]}"
                        matches.append(match)
        except json.JSONDecodeError:
            pass

        return matches

    def pattern_replace(
        self,
        pattern: str,
        replacement: str,
        language: str,
        path: str,
        dry_run: bool = True,
    ) -> ReplaceResult:
        """Replace pattern matches in files.

        Args:
            pattern: AST-grep pattern to search for.
            replacement: Replacement pattern.
            language: Programming language for the pattern.
            path: File or directory path to process.
            dry_run: If True, don't actually modify files.

        Returns:
            ReplaceResult with details of changes.

        Raises:
            FileNotFoundError: If the path does not exist.
        """
        search_path = Path(path)
        if not search_path.exists():
            raise FileNotFoundError(f"Path not found: {path}")

        matches_found = 0
        files_modified = 0
        changes: list[dict[str, Any]] = []

        if not self.is_sg_available():
            return ReplaceResult(
                pattern=pattern,
                replacement=replacement,
                matches_found=0,
                files_modified=0,
                changes=[],
                dry_run=dry_run,
            )

        try:
            # First, find all matches to count them
            search_matches = self.pattern_search(pattern, language, path)
            matches_found = len(search_matches)

            # Track which files have matches
            files_with_matches = set(m.file_path for m in search_matches)
            files_modified = len(files_with_matches)

            # Record changes for reporting
            for match in search_matches:
                changes.append(
                    {
                        "file_path": match.file_path,
                        "old_code": pattern,
                        "new_code": replacement,
                        "range": {
                            "start": {
                                "line": match.range.start.line,
                                "character": match.range.start.character,
                            },
                            "end": {
                                "line": match.range.end.line,
                                "character": match.range.end.character,
                            },
                        },
                    }
                )

            # Only perform actual replacement if not dry run and there are matches
            if not dry_run and matches_found > 0:
                cmd = [
                    "sg",
                    "run",
                    "--pattern",
                    pattern,
                    "--rewrite",
                    replacement,
                    "--lang",
                    language,
                    path,
                ]

                subprocess.run(
                    cmd,
                    capture_output=True,
                    text=True,
                    timeout=120,
                )

        except (subprocess.SubprocessError, json.JSONDecodeError):
            pass

        return ReplaceResult(
            pattern=pattern,
            replacement=replacement,
            matches_found=matches_found,
            files_modified=files_modified,
            changes=changes,
            dry_run=dry_run,
        )
