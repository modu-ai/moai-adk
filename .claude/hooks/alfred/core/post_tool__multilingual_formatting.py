#!/usr/bin/env python3
"""
PostToolUse Hook: Multilingual Code Formatting

This hook automatically detects the project language and formats modified code
files using appropriate formatters after any file creation or modification.

Trigger: PostToolUse event
Events: Write, Edit, NotebookEdit, MultiEdit operations

Architecture:
1. Detect project languages using LanguageDetector
2. Map modified file to language
3. Run language-specific formatter
4. Provide feedback to user
5. Non-blocking (tool failures don't halt workflow)

Supported Formatters:
- Python: ruff format
- JavaScript/JSX: prettier
- TypeScript: prettier
- Go: gofmt
- Rust: rustfmt
- Java: spotless
- Ruby: rubocop (auto-fix)
- PHP: php-cs-fixer
"""

import json
import sys
import logging
from pathlib import Path
from typing import Dict, List, Optional

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent))

from language_detector import LanguageDetector
from formatters import FormatterRegistry

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(levelname)s: %(message)s'
)
logger = logging.getLogger(__name__)


class MultilingualFormattingHook:
    """
    Orchestrates multilingual code formatting for modified files

    This hook:
    - Detects project languages
    - Maps files to languages
    - Runs appropriate formatters
    - Handles errors gracefully
    - Provides detailed feedback
    """

    def __init__(self, project_root: Optional[Path] = None):
        """
        Initialize formatting hook

        Args:
            project_root: Root directory of project
        """
        self.project_root = project_root or Path.cwd()
        self.detector = LanguageDetector(self.project_root)
        self.formatter_registry = FormatterRegistry(self.project_root)
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

    def format_file(self, file_path: Path) -> bool:
        """
        Format a single file

        Args:
            file_path: Path to file to format

        Returns:
            True if formatting succeeded or skipped
        """
        if not file_path.exists():
            logger.warning(f"File does not exist: {file_path}")
            return True

        language = self.get_language_for_file(file_path)
        if not language:
            logger.debug(f"No formatter for file type: {file_path.name}")
            return True

        logger.info(f"Formatting {language} file: {file_path.name}...")
        success = self.formatter_registry.format_file(language, file_path)

        if success:
            logger.info(f"✅ Successfully formatted {file_path.name}")
        else:
            logger.warning(f"⚠️ Formatting warning for {file_path.name}")

        return success

    def format_files(self, file_paths: List[Path]) -> Dict[str, any]:
        """
        Format multiple files and return summary

        Args:
            file_paths: List of file paths to format

        Returns:
            Dictionary with formatting summary
        """
        if not file_paths:
            return {
                "status": "skipped",
                "reason": "No files to format",
                "files_checked": 0
            }

        summary = {
            "status": "completed",
            "total_files": len(file_paths),
            "files_formatted": 0,
            "files_by_language": {},
            "languages_detected": self.detector.detect_languages()
        }

        for file_path in file_paths:
            language = self.get_language_for_file(file_path)
            if not language:
                continue

            if language not in summary["files_by_language"]:
                summary["files_by_language"][language] = {
                    "count": 0,
                    "formatted": 0
                }

            summary["files_by_language"][language]["count"] += 1

            if self.format_file(file_path):
                summary["files_formatted"] += 1
                summary["files_by_language"][language]["formatted"] += 1

        return summary

    def should_format_file(self, file_path: Path) -> bool:
        """
        Determine if file should be formatted

        Args:
            file_path: Path to file

        Returns:
            True if file should be formatted
        """
        # Skip hidden files
        if file_path.name.startswith('.'):
            return False

        # Skip files in hidden directories
        if any(part.startswith('.') for part in file_path.parts):
            return False

        # Skip common directories
        skip_dirs = {
            'node_modules', 'venv', '.venv', '__pycache__', '.git',
            'dist', 'build', '.next', 'out', 'coverage', '.pytest_cache'
        }
        if any(part in skip_dirs for part in file_path.parts):
            return False

        # Skip certain file patterns
        skip_patterns = {'.min.js', '.min.css', '.bundle.js'}
        if any(file_path.name.endswith(pattern) for pattern in skip_patterns):
            return False

        return True

    def get_summary_message(self, summary: Dict) -> str:
        """
        Generate human-readable summary message

        Args:
            summary: Summary dictionary from format_files

        Returns:
            Formatted summary message
        """
        if summary["status"] == "skipped":
            return "✅ No files to format"

        lines = [
            f"🎨 Multilingual Code Formatting Summary",
            f"  Detected languages: {', '.join(summary['languages_detected'])}",
            f"  Files formatted: {summary['files_formatted']}/{summary['total_files']}"
        ]

        if summary["files_by_language"]:
            lines.append(f"  Language breakdown:")
            for lang, stats in summary["files_by_language"].items():
                lines.append(
                    f"    ✅ {lang}: {stats['count']} file(s) formatted"
                )

        return "\n".join(lines)


def main():
    """
    Main entry point for PostToolUse formatting hook

    Reads modified files from environment/stdin and runs formatting
    """
    try:
        # Get project root from environment or use current directory
        project_root = Path(os.environ.get('CLAUDE_PROJECT_ROOT', Path.cwd()))

        # Initialize hook
        hook = MultilingualFormattingHook(project_root)

        # For testing: accept file paths from command line
        if len(sys.argv) > 1:
            file_paths = [Path(arg) for arg in sys.argv[1:]]
        else:
            # In production, files would come from Claude Code hook context
            logger.warning("No files provided to formatting hook")
            return

        # Filter files that should be formatted
        files_to_format = [f for f in file_paths if hook.should_format_file(f)]

        if not files_to_format:
            logger.info("✅ No files to format (all skipped)")
            return

        # Run formatting
        logger.info(f"Starting multilingual code formatting for {len(files_to_format)} file(s)...")
        summary = hook.format_files(files_to_format)

        # Print summary
        print(hook.get_summary_message(summary))

    except Exception as e:
        logger.error(f"❌ Formatting hook error: {e}", exc_info=True)
        # Non-blocking - don't fail the entire operation


if __name__ == "__main__":
    import os
    main()
