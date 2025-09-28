#!/usr/bin/env python3
"""
File Analyzer Module for MoAI Commit Helper
Handles file change analysis and classification

@TASK:FILE-ANALYSIS-001
@FEATURE:FILE-CHANGE-ANALYSIS-001
@API:ANALYZE-FILES
@TRUST:READABLE
"""

from typing import Any


class FileAnalyzer:
    """File change analyzer for smart commit processing"""

    # Constants for file type classification
    PYTHON_EXTENSIONS = {".py", ".pyx", ".pyi"}
    DOCUMENTATION_EXTENSIONS = {".md", ".rst", ".txt", ".doc", ".docx"}
    CONFIG_EXTENSIONS = {".json", ".yaml", ".yml", ".toml", ".ini", ".cfg"}
    TEST_PATTERNS = {"test_", "_test", "tests/", "test/"}

    def classify_file_change(self, status: str) -> str:
        """Classify file change type based on git status"""
        status_map = {
            "A": "added",
            "M": "modified",
            "D": "deleted",
            "R": "renamed",
            "C": "copied",
        }

        # Get first character of status for classification
        first_char = status[0] if status else "?"
        return status_map.get(first_char, "unknown")

    def analyze_file_changes(self, files: list[dict[str, Any]]) -> dict[str, Any]:
        """Analyze file changes and provide detailed classification"""
        if not files:
            return {
                "summary": {"added": 0, "modified": 0, "deleted": 0, "renamed": 0},
                "categories": {"python": 0, "docs": 0, "config": 0, "tests": 0, "other": 0},
                "total_files": 0,
                "is_simple_change": True,
            }

        summary = self._summarize_changes(files)
        categories = self._categorize_files(files)

        return {
            "summary": summary,
            "categories": categories,
            "total_files": len(files),
            "is_simple_change": self._is_simple_change(summary, categories),
            "has_tests": categories["tests"] > 0,
            "is_documentation_only": categories["docs"] > 0 and categories["python"] == 0,
        }

    def get_file_category(self, filename: str) -> str:
        """Get category of a single file"""
        filename_lower = filename.lower()

        # Check for test files first
        if any(pattern in filename_lower for pattern in self.TEST_PATTERNS):
            return "tests"

        # Check by extension
        extension = self._get_file_extension(filename)

        if extension in self.PYTHON_EXTENSIONS:
            return "python"
        elif extension in self.DOCUMENTATION_EXTENSIONS:
            return "docs"
        elif extension in self.CONFIG_EXTENSIONS:
            return "config"
        else:
            return "other"

    def suggest_commit_type(self, analysis: dict[str, Any]) -> str:
        """Suggest commit type based on file analysis"""
        categories = analysis["categories"]
        summary = analysis["summary"]

        # Guard clause: No changes
        if analysis["total_files"] == 0:
            return "chore"

        # Test-related changes
        if categories["tests"] > 0 and categories["python"] == 0:
            return "test"

        # Documentation-only changes
        if analysis["is_documentation_only"]:
            return "docs"

        # Configuration changes
        if categories["config"] > 0 and categories["python"] == 0:
            return "config"

        # New features (additions with Python files)
        if summary["added"] > 0 and categories["python"] > 0:
            return "feat"

        # Bug fixes or modifications
        if summary["modified"] > 0:
            return "fix" if analysis["is_simple_change"] else "refactor"

        # Deletions
        if summary["deleted"] > 0:
            return "remove"

        return "chore"

    def _summarize_changes(self, files: list[dict[str, Any]]) -> dict[str, int]:
        """Summarize changes by type"""
        summary = {"added": 0, "modified": 0, "deleted": 0, "renamed": 0}

        for file in files:
            file_type = file.get("type", "unknown")
            if file_type in summary:
                summary[file_type] += 1

        return summary

    def _categorize_files(self, files: list[dict[str, Any]]) -> dict[str, int]:
        """Categorize files by purpose"""
        categories = {"python": 0, "docs": 0, "config": 0, "tests": 0, "other": 0}

        for file in files:
            filename = file.get("filename", "")
            category = self.get_file_category(filename)
            categories[category] += 1

        return categories

    def _is_simple_change(self, summary: dict[str, int], categories: dict[str, int]) -> bool:
        """Determine if this is a simple change"""
        total_files = sum(summary.values())

        # Simple if few files changed
        if total_files <= 3:
            return True

        # Simple if only one category affected
        non_zero_categories = sum(1 for count in categories.values() if count > 0)
        if non_zero_categories == 1:
            return True

        return False

    def _get_file_extension(self, filename: str) -> str:
        """Extract file extension safely"""
        if "." not in filename:
            return ""

        return "." + filename.split(".")[-1].lower()