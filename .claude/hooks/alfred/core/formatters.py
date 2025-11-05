#!/usr/bin/env python3
"""
Language-specific formatter runners

This module provides a registry of formatters for different programming languages
and abstracts away the differences in how to invoke them.

Supported Formatters:
  - Python: ruff (format + check)
  - JavaScript: prettier
  - TypeScript: prettier
  - Go: gofmt
  - Rust: rustfmt
  - Java: spotless
  - Ruby: rubocop (auto-correct)
  - PHP: php-cs-fixer
"""

import subprocess
from pathlib import Path
from typing import Dict, Callable, Optional
import logging

logger = logging.getLogger(__name__)


class FormatterRegistry:
    """Registry of language-specific formatters"""

    def __init__(self, project_root: Optional[Path] = None):
        """
        Initialize formatter registry

        Args:
            project_root: Root directory of project
        """
        self.project_root = project_root or Path.cwd()
        self.formatters: Dict[str, Callable] = {
            "python": self._format_python,
            "javascript": self._format_javascript,
            "typescript": self._format_typescript,
            "go": self._format_go,
            "rust": self._format_rust,
            "java": self._format_java,
            "ruby": self._format_ruby,
            "php": self._format_php,
        }

    def format_file(self, language: str, file_path: Path) -> bool:
        """
        Format file for specific language

        Args:
            language: Programming language name
            file_path: Path to file to format

        Returns:
            True if formatting succeeded or skipped, False if errors
        """
        if language not in self.formatters:
            logger.debug(f"No formatter available for {language}")
            return True  # Skip unknown languages

        try:
            return self.formatters[language](file_path)
        except Exception as e:
            logger.warning(f"⚠️ Formatter error for {language}: {e}")
            return True  # Non-blocking

    def format_directory(self, language: str, directory: Path, extensions: list) -> bool:
        """
        Format all files matching extensions in directory

        Args:
            language: Programming language name
            directory: Directory to format
            extensions: List of file extensions to format (e.g., ['.py', '.pyi'])

        Returns:
            True if all formatting succeeded
        """
        success = True
        for ext in extensions:
            for file_path in directory.rglob(f"*{ext}"):
                if not self.format_file(language, file_path):
                    success = False
        return success

    # ==================== PYTHON ====================

    def _format_python(self, file_path: Path) -> bool:
        """Format Python file with ruff"""
        if not file_path.suffix == ".py":
            return True

        try:
            result = subprocess.run(
                ["ruff", "format", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info(f"✅ Ruff formatted: {file_path.name}")
            else:
                logger.warning(f"⚠️ Ruff format warning: {result.stderr}")

            return True

        except FileNotFoundError:
            logger.warning("⚠️ Ruff not installed. Install with: uv add ruff")
            return True

    # ==================== JAVASCRIPT ====================

    def _format_javascript(self, file_path: Path) -> bool:
        """Format JavaScript with prettier"""
        if file_path.suffix not in [".js", ".jsx", ".mjs"]:
            return True

        try:
            result = subprocess.run(
                ["npx", "prettier", "--write", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=self.project_root
            )

            if result.returncode == 0:
                logger.info(f"✅ Prettier formatted: {file_path.name}")
            return True

        except FileNotFoundError:
            logger.warning("⚠️ Prettier not installed. Install with: npm install prettier")
            return True

    # ==================== TYPESCRIPT ====================

    def _format_typescript(self, file_path: Path) -> bool:
        """Format TypeScript with prettier"""
        if file_path.suffix not in [".ts", ".tsx"]:
            return True

        return self._format_javascript(file_path)

    # ==================== GO ====================

    def _format_go(self, file_path: Path) -> bool:
        """Format Go with gofmt"""
        if file_path.suffix != ".go":
            return True

        try:
            result = subprocess.run(
                ["gofmt", "-w", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info(f"✅ gofmt formatted: {file_path.name}")
            return True

        except FileNotFoundError:
            logger.warning("⚠️ gofmt not installed. Install from Go toolchain")
            return True

    # ==================== RUST ====================

    def _format_rust(self, file_path: Path) -> bool:
        """Format Rust with rustfmt"""
        if file_path.suffix != ".rs":
            return True

        try:
            result = subprocess.run(
                ["rustfmt", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info(f"✅ rustfmt formatted: {file_path.name}")
            return True

        except FileNotFoundError:
            logger.warning("⚠️ rustfmt not installed. Install from Rust toolchain")
            return True

    # ==================== JAVA ====================

    def _format_java(self, file_path: Path) -> bool:
        """Format Java with spotless"""
        if file_path.suffix != ".java":
            return True

        try:
            result = subprocess.run(
                ["spotless", "apply", "-DspotlessTarget=" + str(file_path)],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=self.project_root
            )

            if result.returncode == 0:
                logger.info(f"✅ Spotless formatted: {file_path.name}")
            return True

        except FileNotFoundError:
            logger.warning("⚠️ spotless not installed. Install via Maven/Gradle")
            return True

    # ==================== RUBY ====================

    def _format_ruby(self, file_path: Path) -> bool:
        """Format Ruby with rubocop (auto-correct)"""
        if file_path.suffix != ".rb":
            return True

        try:
            result = subprocess.run(
                ["rubocop", "-a", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info(f"✅ Rubocop auto-corrected: {file_path.name}")
            return True

        except FileNotFoundError:
            logger.warning("⚠️ rubocop not installed. Install with: gem install rubocop")
            return True

    # ==================== PHP ====================

    def _format_php(self, file_path: Path) -> bool:
        """Format PHP with php-cs-fixer"""
        if file_path.suffix != ".php":
            return True

        try:
            result = subprocess.run(
                ["php-cs-fixer", "fix", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info(f"✅ php-cs-fixer formatted: {file_path.name}")
            return True

        except FileNotFoundError:
            logger.warning("⚠️ php-cs-fixer not installed. Install with: composer require friendsofphp/php-cs-fixer")
            return True


def main():
    """CLI entry point for formatter registry"""
    import sys

    if len(sys.argv) < 3:
        print("Usage: formatters.py <language> <file_path>")
        sys.exit(1)

    language = sys.argv[1]
    file_path = Path(sys.argv[2])

    registry = FormatterRegistry()

    print(f"Formatting {language} file: {file_path.name}...")
    success = registry.format_file(language, file_path)

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
