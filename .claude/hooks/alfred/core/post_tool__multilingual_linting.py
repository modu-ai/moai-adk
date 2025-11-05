#!/usr/bin/env python3
"""
PostToolUse Hook: Multilingual Linting Integration

This hook automatically detects the project language and runs appropriate
linting checks after any file creation or modification through Claude Code.

Trigger: PostToolUse event
Events: Write, Edit, NotebookEdit, MultiEdit operations

Architecture:
1. Detect project languages using LanguageDetector
2. Map modified file to language
3. Run language-specific linting
4. Provide feedback to user
5. Non-blocking (tool failures don't halt workflow)
"""

import json
import sys
import logging
from pathlib import Path
from typing import Dict, List, Optional, Set

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent))

from language_detector import LanguageDetector
from linters import LinterRegistry

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(levelname)s: %(message)s'
)
logger = logging.getLogger(__name__)


class MultilingualLintingHook:
    """
    Orchestrates multilingual linting for modified files

    This hook:
    - Detects project languages
    - Maps files to languages
    - Runs appropriate linters
    - Handles errors gracefully
    """

    def __init__(self, project_root: Optional[Path] = None):
        """
        Initialize linting hook

        Args:
            project_root: Root directory of project
        """
        self.project_root = project_root or Path.cwd()
        self.detector = LanguageDetector(self.project_root)
        self.linter_registry = LinterRegistry(self.project_root)
        self.file_to_language_map = self._build_file_language_map()

    def _build_file_language_map(self) -> Dict[str, str]:
        """
        Build mapping of file extensions to languages

        Returns:
            Dictionary mapping file extensions to language names
        """
        file_map = {}
        detected_languages = self.detector.detect_languages()

        for language in detected_languages:
            extensions = self.detector.get_file_extension_for_language(language)
            for ext in extensions:
                file_map[ext] = language

        return file_map

    def get_language_for_file(self, file_path: Path) -> Optional[str]:
        """
        Determine language for a file based on extension

        Args:
            file_path: Path to file

        Returns:
            Language name or None if unknown
        """
        suffix = file_path.suffix.lower()
        return self.file_to_language_map.get(suffix)

    def lint_file(self, file_path: Path) -> bool:
        """
        Lint a single file

        Args:
            file_path: Path to file to lint

        Returns:
            True if linting passed or skipped, False if errors found
        """
        if not file_path.exists():
            logger.warning(f"File does not exist: {file_path}")
            return True

        language = self.get_language_for_file(file_path)
        if not language:
            logger.debug(f"No linter for file type: {file_path.name}")
            return True

        logger.info(f"Running {language} linter on {file_path.name}...")
        success = self.linter_registry.run_linter(language, file_path)

        if not success:
            logger.warning(f"Linting found issues in {file_path.name} (non-blocking)")

        return success

    def lint_files(self, file_paths: List[Path]) -> Dict[str, any]:
        """
        Lint multiple files and return summary

        Args:
            file_paths: List of file paths to lint

        Returns:
            Dictionary with linting summary
        """
        if not file_paths:
            return {
                "status": "skipped",
                "reason": "No files to lint",
                "files_checked": 0
            }

        summary = {
            "status": "completed",
            "total_files": len(file_paths),
            "files_checked": 0,
            "files_with_issues": 0,
            "files_by_language": {},
            "languages_detected": self.detector.detect_languages()
        }

        for file_path in file_paths:
            language = self.get_language_for_file(file_path)
            if not language:
                continue

            summary["files_checked"] += 1
            if language not in summary["files_by_language"]:
                summary["files_by_language"][language] = {
                    "count": 0,
                    "passed": 0,
                    "failed": 0
                }

            summary["files_by_language"][language]["count"] += 1

            success = self.lint_file(file_path)
            if success:
                summary["files_by_language"][language]["passed"] += 1
            else:
                summary["files_by_language"][language]["failed"] += 1
                summary["files_with_issues"] += 1

        return summary

    def should_lint_file(self, file_path: Path) -> bool:
        """
        Determine if file should be linted

        Args:
            file_path: Path to file

        Returns:
            True if file should be linted
        """
        # Skip hidden files and certain directories
        if file_path.name.startswith('.'):
            return False

        if any(part.startswith('.') for part in file_path.parts):
            return False

        # Skip node_modules, venv, etc.
        skip_dirs = {'node_modules', 'venv', '.venv', '__pycache__', '.git', 'dist', 'build'}
        if any(part in skip_dirs for part in file_path.parts):
            return False

        return True

    def get_summary_message(self, summary: Dict) -> str:
        """
        Generate human-readable summary message

        Args:
            summary: Summary dictionary from lint_files

        Returns:
            Formatted summary message
        """
        if summary["status"] == "skipped":
            return "✅ No files to lint"

        lines = [
            f"🔍 Multilingual Linting Summary",
            f"  Detected languages: {', '.join(summary['languages_detected'])}",
            f"  Files checked: {summary['files_checked']}/{summary['total_files']}"
        ]

        if summary["files_by_language"]:
            lines.append(f"  Language breakdown:")
            for lang, stats in summary["files_by_language"].items():
                status = "✅" if stats["failed"] == 0 else "🔴"
                lines.append(
                    f"    {status} {lang}: {stats['count']} file(s) "
                    f"({stats['passed']} passed, {stats['failed']} failed)"
                )

        if summary["files_with_issues"] > 0:
            lines.append(f"  ⚠️ {summary['files_with_issues']} file(s) with linting issues")
            lines.append(f"     (Non-blocking - review warnings above)")
        else:
            lines.append(f"  ✅ All files passed linting checks")

        return "\n".join(lines)


def main():
    """
    Main entry point for PostToolUse hook

    Reads modified files from environment/stdin and runs linting
    """
    try:
        # Get project root from environment or use current directory
        project_root = Path(os.environ.get('CLAUDE_PROJECT_ROOT', Path.cwd()))

        # Initialize hook
        hook = MultilingualLintingHook(project_root)

        # For testing: accept file paths from command line
        if len(sys.argv) > 1:
            file_paths = [Path(arg) for arg in sys.argv[1:]]
        else:
            # In production, files would come from Claude Code hook context
            # This is a fallback for testing
            logger.warning("No files provided to linting hook")
            return

        # Filter files that should be linted
        files_to_lint = [f for f in file_paths if hook.should_lint_file(f)]

        if not files_to_lint:
            logger.info("✅ No files to lint (all skipped)")
            return

        # Run linting
        logger.info(f"Starting multilingual linting for {len(files_to_lint)} file(s)...")
        summary = hook.lint_files(files_to_lint)

        # Print summary
        print(hook.get_summary_message(summary))

    except Exception as e:
        logger.error(f"❌ Linting hook error: {e}", exc_info=True)
        # Non-blocking - don't fail the entire operation


if __name__ == "__main__":
    import os
    main()
