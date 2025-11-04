#!/usr/bin/env python3
"""
Language Detector: Automatically detect project language(s)

This module detects the programming language(s) used in a project by analyzing
the presence of language-specific configuration files and markers.

Supported Languages:
  - Python (pyproject.toml, setup.py, requirements.txt)
  - JavaScript (package.json, webpack.config.js)
  - TypeScript (tsconfig.json, package.json with ts)
  - Go (go.mod, go.sum)
  - Rust (Cargo.toml, Cargo.lock)
  - Java (pom.xml, build.gradle)
  - Ruby (Gemfile, Gemfile.lock)
  - PHP (composer.json, composer.lock)
  - C/C++ (CMakeLists.txt, Makefile)
  - Kotlin (build.gradle.kts, pom.xml)
  - SQL (migrations/, *.sql)
"""

import json
import subprocess
from pathlib import Path
from typing import List, Dict, Optional, Tuple
import logging

logger = logging.getLogger(__name__)


class LanguageDetector:
    """
    Detect programming language(s) from project structure

    This class analyzes project files to identify which programming languages
    are used, supporting both single-language and multi-language projects.
    """

    # Language markers: file patterns that indicate language presence
    LANGUAGE_MARKERS = {
        "python": [
            "pyproject.toml",
            "setup.py",
            "setup.cfg",
            "requirements.txt",
            "Pipfile",
            "poetry.lock",
        ],
        "typescript": [
            "tsconfig.json",
            "package.json",  # with "typescript" or ".ts" files
        ],
        "javascript": [
            "package.json",
            "webpack.config.js",
            "babel.config.js",
            ".eslintrc.js",
        ],
        "go": [
            "go.mod",
            "go.sum",
        ],
        "rust": [
            "Cargo.toml",
            "Cargo.lock",
        ],
        "java": [
            "pom.xml",
            "build.gradle",
            "settings.gradle",
            "build.gradle.kts",
        ],
        "ruby": [
            "Gemfile",
            "Gemfile.lock",
            "Rakefile",
        ],
        "php": [
            "composer.json",
            "composer.lock",
            "phpunit.xml",
        ],
        "csharp": [
            "*.csproj",
            "*.sln",
        ],
        "kotlin": [
            "build.gradle.kts",
            "pom.xml",
        ],
        "sql": [
            "migrations/",
            "*.sql",
        ],
    }

    # Priority order (main language listed first)
    PRIORITY_ORDER = [
        "typescript",
        "python",
        "go",
        "rust",
        "java",
        "ruby",
        "php",
        "javascript",
        "kotlin",
        "csharp",
        "sql",
    ]

    def __init__(self, project_root: Optional[Path] = None):
        """
        Initialize language detector

        Args:
            project_root: Root directory of project. Defaults to current working directory.
        """
        self.project_root = project_root or Path.cwd()
        self._detected_languages: Optional[List[str]] = None
        self._primary_language: Optional[str] = None

    def detect_languages(self) -> List[str]:
        """
        Detect all programming languages in project

        Returns:
            List of detected languages in priority order
        """
        if self._detected_languages is not None:
            return self._detected_languages

        detected = {}

        for language, markers in self.LANGUAGE_MARKERS.items():
            for marker in markers:
                if self._path_matches(marker):
                    detected[language] = True
                    break

        # Return in priority order
        result = [lang for lang in self.PRIORITY_ORDER if lang in detected]
        self._detected_languages = result or list(detected.keys())
        return self._detected_languages

    def detect_primary_language(self) -> str:
        """
        Detect primary/main language of project

        Returns:
            Primary programming language
        """
        if self._primary_language is not None:
            return self._primary_language

        languages = self.detect_languages()
        self._primary_language = languages[0] if languages else "unknown"
        return self._primary_language

    def get_file_extension_for_language(self, language: str) -> List[str]:
        """
        Get file extensions for a programming language

        Args:
            language: Programming language name

        Returns:
            List of file extensions (with leading dot)
        """
        extensions = {
            "python": [".py"],
            "javascript": [".js", ".jsx", ".mjs"],
            "typescript": [".ts", ".tsx"],
            "go": [".go"],
            "rust": [".rs"],
            "java": [".java"],
            "ruby": [".rb"],
            "php": [".php"],
            "csharp": [".cs"],
            "kotlin": [".kt", ".kts"],
            "sql": [".sql"],
        }
        return extensions.get(language, [])

    def get_package_manager(self, language: str) -> str:
        """
        Get primary package manager for language

        Args:
            language: Programming language name

        Returns:
            Package manager name (e.g., "pip", "npm", "cargo")
        """
        managers = {
            "python": "pip",
            "javascript": "npm",
            "typescript": "npm",
            "go": "go",
            "rust": "cargo",
            "java": "maven",
            "ruby": "bundler",
            "php": "composer",
            "csharp": "nuget",
            "kotlin": "gradle",
        }
        return managers.get(language, "unknown")

    def is_language_installed(self, language: str) -> bool:
        """
        Check if language runtime/compiler is installed

        Args:
            language: Programming language name

        Returns:
            True if language is installed, False otherwise
        """
        check_commands = {
            "python": ["python", "--version"],
            "javascript": ["node", "--version"],
            "typescript": ["npx", "tsc", "--version"],
            "go": ["go", "version"],
            "rust": ["rustc", "--version"],
            "java": ["java", "-version"],
            "ruby": ["ruby", "--version"],
            "php": ["php", "--version"],
            "csharp": ["dotnet", "--version"],
            "kotlin": ["kotlinc", "-version"],
        }

        cmd = check_commands.get(language)
        if not cmd:
            return False

        try:
            subprocess.run(
                cmd,
                capture_output=True,
                timeout=5,
                check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError, subprocess.TimeoutExpired):
            return False

    def get_linter_tools(self, language: str) -> Dict[str, str]:
        """
        Get linter/formatter tools recommended for language

        Args:
            language: Programming language name

        Returns:
            Dictionary with tool names (formatter, linter, type_checker)
        """
        tools = {
            "python": {
                "formatter": "ruff",
                "linter": "ruff",
                "type_checker": "mypy",
            },
            "javascript": {
                "formatter": "prettier",
                "linter": "eslint",
                "type_checker": None,
            },
            "typescript": {
                "formatter": "prettier",
                "linter": "eslint",
                "type_checker": "tsc",
            },
            "go": {
                "formatter": "gofmt",
                "linter": "golangci-lint",
                "type_checker": None,
            },
            "rust": {
                "formatter": "rustfmt",
                "linter": "clippy",
                "type_checker": None,
            },
            "java": {
                "formatter": "spotless",
                "linter": "checkstyle",
                "type_checker": None,
            },
            "ruby": {
                "formatter": "rubocop",
                "linter": "rubocop",
                "type_checker": "sorbet",
            },
            "php": {
                "formatter": "php-cs-fixer",
                "linter": "phpstan",
                "type_checker": "psalm",
            },
        }
        return tools.get(language, {})

    def _path_matches(self, pattern: str) -> bool:
        """
        Check if file path or glob pattern exists

        Args:
            pattern: File name or glob pattern

        Returns:
            True if pattern matches existing file(s)
        """
        if "*" in pattern or "?" in pattern:
            # Glob pattern matching
            matches = list(self.project_root.glob(pattern))
            return len(matches) > 0
        else:
            # Direct file matching
            return (self.project_root / pattern).exists()

    def analyze_package_json(self) -> Tuple[bool, bool]:
        """
        Analyze package.json to determine if project is TypeScript or JavaScript

        Returns:
            Tuple of (is_typescript, is_javascript)
        """
        package_json_path = self.project_root / "package.json"
        if not package_json_path.exists():
            return False, False

        try:
            with open(package_json_path, "r") as f:
                package_json = json.load(f)

            # Check for TypeScript
            has_typescript = (
                "typescript" in package_json.get("dependencies", {})
                or "typescript" in package_json.get("devDependencies", {})
                or (self.project_root / "tsconfig.json").exists()
            )

            # Check for TypeScript files
            if not has_typescript:
                ts_files = list(self.project_root.glob("**/*.ts")) + \
                           list(self.project_root.glob("**/*.tsx"))
                has_typescript = len(ts_files) > 0

            return has_typescript, True  # Always JavaScript if package.json exists

        except (json.JSONDecodeError, IOError):
            return False, True  # Assume JavaScript if json is invalid

    def get_summary(self) -> str:
        """
        Get human-readable summary of detected languages

        Returns:
            Summary string describing detected languages
        """
        languages = self.detect_languages()
        if not languages or languages[0] == "unknown":
            return "No recognized programming language detected"

        primary = self.detect_primary_language()
        if len(languages) == 1:
            return f"Primary language: {primary}"
        else:
            others = ", ".join(languages[1:])
            return f"Primary: {primary} | Also detected: {others}"


def main():
    """CLI entry point for language detection"""
    import sys

    project_root = Path(sys.argv[1]) if len(sys.argv) > 1 else Path.cwd()
    detector = LanguageDetector(project_root)

    print(f"📁 Project root: {project_root}")
    print(f"🔍 Detected languages: {detector.detect_languages()}")
    print(f"⭐ Primary language: {detector.detect_primary_language()}")
    print(f"\n{detector.get_summary()}")

    # Show tool recommendations
    primary = detector.detect_primary_language()
    if primary != "unknown":
        tools = detector.get_linter_tools(primary)
        print(f"\n🛠️  Recommended tools for {primary}:")
        for tool_type, tool_name in tools.items():
            status = "✅" if tool_name else "❌"
            print(f"  {status} {tool_type}: {tool_name or 'Not available'}")


if __name__ == "__main__":
    main()
