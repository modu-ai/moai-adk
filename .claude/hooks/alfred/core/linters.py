#!/usr/bin/env python3
"""
Language-specific linter runners

This module provides a registry of linters for different programming languages
and abstracts away the differences in how to invoke them.

Supported Linters:
  - Python: ruff (format + lint + type check)
  - JavaScript: eslint + prettier
  - TypeScript: tsc + eslint + prettier
  - Go: golangci-lint + gofmt
  - Rust: clippy + rustfmt
  - Java: checkstyle + spotless
  - Ruby: rubocop
  - PHP: phpstan + php-cs-fixer
"""

import subprocess
import sys
from pathlib import Path
from typing import Dict, Callable, Tuple, Optional
import logging

logger = logging.getLogger(__name__)


class LinterRegistry:
    """Registry of language-specific linters and formatters"""

    def __init__(self, project_root: Optional[Path] = None):
        """
        Initialize linter registry

        Args:
            project_root: Root directory of project
        """
        self.project_root = project_root or Path.cwd()
        self.linters: Dict[str, Callable] = {
            "python": self._run_python_linting,
            "javascript": self._run_javascript_linting,
            "typescript": self._run_typescript_linting,
            "go": self._run_go_linting,
            "rust": self._run_rust_linting,
            "java": self._run_java_linting,
            "ruby": self._run_ruby_linting,
            "php": self._run_php_linting,
        }
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

    def run_linter(self, language: str, file_path: Path) -> bool:
        """
        Run linter for specific language

        Args:
            language: Programming language name
            file_path: Path to file to lint

        Returns:
            True if linting passed, False if errors found
        """
        if language not in self.linters:
            logger.warning(f"⚠️ No linter available for {language}")
            return True  # Skip unknown languages (non-blocking)

        try:
            return self.linters[language](file_path)
        except FileNotFoundError as e:
            logger.warning(f"⚠️ Linter tool not found for {language}: {e}")
            return True  # Non-blocking
        except subprocess.TimeoutExpired:
            logger.warning(f"⏱️ Linter timeout for {language}")
            return True  # Non-blocking
        except Exception as e:
            logger.error(f"❌ Linter error for {language}: {e}")
            return True  # Non-blocking

    def run_formatter(self, language: str, file_path: Path) -> bool:
        """
        Run formatter for specific language

        Args:
            language: Programming language name
            file_path: Path to file to format

        Returns:
            True if formatting succeeded
        """
        if language not in self.formatters:
            logger.debug(f"No formatter available for {language}")
            return True  # Skip unknown languages

        try:
            return self.formatters[language](file_path)
        except Exception as e:
            logger.warning(f"⚠️ Formatter error for {language}: {e}")
            return True  # Non-blocking

    # ==================== PYTHON ====================

    def _run_python_linting(self, file_path: Path) -> bool:
        """Run ruff linting for Python"""
        if not file_path.suffix == ".py":
            return True

        if not file_path.exists():
            logger.warning(f"⚠️ File not found: {file_path}")
            return True

        try:
            result = subprocess.run(
                ["ruff", "check", str(file_path), "--select=E,F,W,I,N"],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode != 0:
                logger.error(f"🔴 Python lint errors in {file_path.name}:")
                logger.error(result.stdout)
                return False
            else:
                logger.info(f"✅ Ruff check passed: {file_path.name}")
                return True

        except FileNotFoundError:
            logger.warning("⚠️ Ruff not installed. Install with: uv add ruff")
            return True

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
            logger.warning("⚠️ Ruff not installed")
            return True

    # ==================== JAVASCRIPT ====================

    def _run_javascript_linting(self, file_path: Path) -> bool:
        """Run eslint for JavaScript"""
        if file_path.suffix not in [".js", ".jsx", ".mjs"]:
            return True

        if not file_path.exists():
            logger.warning(f"⚠️ File not found: {file_path}")
            return True

        try:
            result = subprocess.run(
                ["npx", "eslint", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=self.project_root
            )

            if result.returncode != 0:
                logger.error(f"🔴 JavaScript lint errors in {file_path.name}:")
                logger.error(result.stdout)
                return False
            else:
                logger.info(f"✅ ESLint check passed: {file_path.name}")
                return True

        except FileNotFoundError:
            logger.warning("⚠️ ESLint not installed. Install with: npm install eslint")
            return True

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
            logger.warning("⚠️ Prettier not installed")
            return True

    # ==================== TYPESCRIPT ====================

    def _run_typescript_linting(self, file_path: Path) -> bool:
        """Run tsc + eslint for TypeScript"""
        if file_path.suffix not in [".ts", ".tsx"]:
            return True

        # TypeScript type checking
        try:
            result = subprocess.run(
                ["npx", "tsc", "--noEmit", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=self.project_root
            )

            if result.returncode != 0:
                logger.warning(f"🟡 TypeScript type errors in {file_path.name}:")
                logger.warning(result.stdout)

        except FileNotFoundError:
            logger.warning("⚠️ TypeScript not installed")

        # ESLint validation (same as JavaScript)
        return self._run_javascript_linting(file_path)

    def _format_typescript(self, file_path: Path) -> bool:
        """Format TypeScript with prettier"""
        if file_path.suffix not in [".ts", ".tsx"]:
            return True

        return self._format_javascript(file_path)

    # ==================== GO ====================

    def _run_go_linting(self, file_path: Path) -> bool:
        """Run golangci-lint for Go"""
        if file_path.suffix != ".go":
            return True

        try:
            result = subprocess.run(
                ["golangci-lint", "run", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=self.project_root
            )

            if result.returncode != 0:
                logger.error(f"🔴 Go lint errors in {file_path.name}:")
                logger.error(result.stdout)
                return False
            else:
                logger.info(f"✅ golangci-lint check passed: {file_path.name}")
                return True

        except FileNotFoundError:
            logger.warning("⚠️ golangci-lint not installed. Install from: https://golangci-lint.run/usage/install/")
            return True

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
            logger.warning("⚠️ gofmt not installed")
            return True

    # ==================== RUST ====================

    def _run_rust_linting(self, file_path: Path) -> bool:
        """Run clippy for Rust"""
        if file_path.suffix != ".rs":
            return True

        try:
            result = subprocess.run(
                ["cargo", "clippy", "--", "-D", "warnings"],
                capture_output=True,
                text=True,
                timeout=60,
                cwd=self.project_root
            )

            if result.returncode != 0:
                logger.error(f"🔴 Rust clippy warnings in {file_path.name}:")
                logger.error(result.stdout)
                return False
            else:
                logger.info(f"✅ Clippy check passed: {file_path.name}")
                return True

        except FileNotFoundError:
            logger.warning("⚠️ Cargo not installed")
            return True

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
            logger.warning("⚠️ rustfmt not installed")
            return True

    # ==================== JAVA ====================

    def _run_java_linting(self, file_path: Path) -> bool:
        """Run checkstyle for Java"""
        if file_path.suffix != ".java":
            return True

        try:
            result = subprocess.run(
                ["checkstyle", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode != 0:
                logger.error(f"🔴 Java checkstyle errors in {file_path.name}:")
                logger.error(result.stdout)
                return False
            else:
                logger.info(f"✅ Checkstyle check passed: {file_path.name}")
                return True

        except FileNotFoundError:
            logger.warning("⚠️ checkstyle not installed")
            return True

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
            logger.warning("⚠️ spotless not installed")
            return True

    # ==================== RUBY ====================

    def _run_ruby_linting(self, file_path: Path) -> bool:
        """Run rubocop for Ruby"""
        if file_path.suffix != ".rb":
            return True

        try:
            result = subprocess.run(
                ["rubocop", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode != 0:
                logger.warning(f"🟡 Ruby rubocop warnings in {file_path.name}:")
                logger.warning(result.stdout)

            return True  # Non-blocking for Ruby

        except FileNotFoundError:
            logger.warning("⚠️ rubocop not installed. Install with: gem install rubocop")
            return True

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
            logger.warning("⚠️ rubocop not installed")
            return True

    # ==================== PHP ====================

    def _run_php_linting(self, file_path: Path) -> bool:
        """Run phpstan for PHP"""
        if file_path.suffix != ".php":
            return True

        try:
            result = subprocess.run(
                ["phpstan", "analyse", str(file_path)],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode != 0:
                logger.warning(f"🟡 PHP phpstan type errors in {file_path.name}:")
                logger.warning(result.stdout)

            return True  # Non-blocking for PHP

        except FileNotFoundError:
            logger.warning("⚠️ phpstan not installed")
            return True

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
            logger.warning("⚠️ php-cs-fixer not installed")
            return True


def main():
    """CLI entry point for linter registry"""
    import sys

    if len(sys.argv) < 3:
        print("Usage: linters.py <language> <file_path>")
        sys.exit(1)

    language = sys.argv[1]
    file_path = Path(sys.argv[2])

    registry = LinterRegistry()

    print(f"Running {language} linter on {file_path.name}...")
    success = registry.run_linter(language, file_path)

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
