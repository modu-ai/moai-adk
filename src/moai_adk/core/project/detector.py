# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_language_detector.py
# @CODE:LANG-DETECT-001 | SPEC: SPEC-LANG-DETECT-001.md | TEST: tests/unit/test_detector.py
"""Language detector module.

Automatically detects 20 programming languages.
"""

from pathlib import Path


class LanguageDetector:
    """Automatically detect up to 20 programming languages.

    Prioritizes framework-specific files (e.g., Laravel, Django) over
    generic language files to improve accuracy in mixed-language projects.
    """

    LANGUAGE_PATTERNS = {
        # Ruby moved to top for priority (Rails detection over generic frameworks)
        # @CODE:LANG-DETECT-RUBY-001 | SPEC: Issue #51 Language Detection Fix
        "ruby": [
            "*.rb",
            "Gemfile",
            "Gemfile.lock",           # Bundler: lock file (unique to Ruby)
            "config/routes.rb",       # Rails: routing file (unique identifier)
            "app/controllers/",       # Rails: controller directory
            "Rakefile"                # Rails/Ruby: task file
        ],
        # PHP moved to second for priority (Laravel detection after Rails)
        "php": [
            "*.php",
            "composer.json",
            "artisan",                # Laravel: CLI tool (unique identifier)
            "app/",                   # Laravel: application directory
            "bootstrap/laravel.php"   # Laravel: bootstrap file
        ],
        "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
        "typescript": ["*.ts", "tsconfig.json"],
        "javascript": ["*.js", "package.json"],
        "java": ["*.java", "pom.xml", "build.gradle"],
        "go": ["*.go", "go.mod"],
        "rust": ["*.rs", "Cargo.toml"],
        "dart": ["*.dart", "pubspec.yaml"],
        "swift": ["*.swift", "Package.swift"],
        "kotlin": ["*.kt", "build.gradle.kts"],
        "csharp": ["*.cs", "*.csproj"],
        "elixir": ["*.ex", "mix.exs"],
        "scala": ["*.scala", "build.sbt"],
        "clojure": ["*.clj", "project.clj"],
        "haskell": ["*.hs", "*.cabal"],
        "c": ["*.c", "Makefile"],
        "cpp": ["*.cpp", "CMakeLists.txt"],
        "shell": ["*.sh", "*.bash"],
        "lua": ["*.lua"],
    }

    def detect(self, path: str | Path = ".") -> str | None:
        """Detect a single language (in priority order).

        Args:
            path: Directory to inspect.

        Returns:
            Detected language name (lowercase) or None.
        """
        path = Path(path)

        # Inspect each language in priority order
        for language, patterns in self.LANGUAGE_PATTERNS.items():
            if self._check_patterns(path, patterns):
                return language

        return None

    def detect_multiple(self, path: str | Path = ".") -> list[str]:
        """Detect multiple languages.

        Args:
            path: Directory to inspect.

        Returns:
            List of all detected language names.
        """
        path = Path(path)
        detected = []

        for language, patterns in self.LANGUAGE_PATTERNS.items():
            if self._check_patterns(path, patterns):
                detected.append(language)

        return detected

    def _check_patterns(self, path: Path, patterns: list[str]) -> bool:
        """Check whether any pattern matches.

        Args:
            path: Directory to inspect.
            patterns: List of glob patterns.

        Returns:
            True when any pattern matches.
        """
        for pattern in patterns:
            # Extension pattern (e.g., *.py)
            if pattern.startswith("*."):
                if list(path.rglob(pattern)):
                    return True
            # Specific file name (e.g., pyproject.toml)
            else:
                if (path / pattern).exists():
                    return True

        return False

    # @CODE:LANG-002 | SPEC: SPEC-LANGUAGE-DETECTION-001.md | TEST: tests/unit/test_detector.py
    def detect_package_manager(self, path: str | Path = ".") -> str:
        """Detect JavaScript/TypeScript package manager.

        Checks for lock files in priority order (highest to lowest):
        1. bun.lockb (Bun) - fastest runtime
        2. pnpm-lock.yaml (pnpm) - efficient disk usage
        3. yarn.lock (Yarn) - Facebook's package manager
        4. package-lock.json (npm) - default Node.js package manager

        Priority order reflects modern best practices and performance characteristics.

        Args:
            path: Project root directory to inspect. Defaults to current directory.

        Returns:
            Package manager name: 'bun' | 'pnpm' | 'yarn' | 'npm'

        Example:
            >>> detector = LanguageDetector()
            >>> detector.detect_package_manager("/path/to/project")
            'pnpm'
        """
        path = Path(path)

        # Check in priority order (Bun → pnpm → Yarn → npm)
        if (path / "bun.lockb").exists():
            return "bun"
        elif (path / "pnpm-lock.yaml").exists():
            return "pnpm"
        elif (path / "yarn.lock").exists():
            return "yarn"
        else:
            # Default to npm (most common, works everywhere)
            return "npm"

    # @CODE:LANG-002 | SPEC: SPEC-LANGUAGE-DETECTION-001.md | TEST: tests/unit/test_detector.py
    def get_workflow_template_path(self, language: str) -> str:
        """Get workflow template path for detected language.

        Returns the relative path to the language-specific GitHub Actions workflow template.
        These templates are pre-configured with language-specific testing tools, linting,
        and coverage reporting.

        Args:
            language: Detected language name (lowercase).
                      Supported: 'python', 'javascript', 'typescript', 'go'

        Returns:
            Relative path to workflow template (e.g., 'workflows/python-tag-validation.yml')

        Raises:
            ValueError: If language doesn't have a dedicated workflow template

        Example:
            >>> detector = LanguageDetector()
            >>> detector.get_workflow_template_path("python")
            'workflows/python-tag-validation.yml'
        """
        # Language-to-template mapping (add new languages here)
        template_mapping = {
            "python": "python-tag-validation.yml",
            "javascript": "javascript-tag-validation.yml",
            "typescript": "typescript-tag-validation.yml",
            "go": "go-tag-validation.yml",
        }

        if language not in template_mapping:
            supported = ", ".join(sorted(template_mapping.keys()))
            raise ValueError(
                f"Language '{language}' does not have a dedicated workflow template. "
                f"Supported languages: {supported}"
            )

        template_filename = template_mapping[language]
        # Return path relative to templates directory
        return f"workflows/{template_filename}"

    # @CODE:LANG-002 | SPEC: SPEC-LANGUAGE-DETECTION-001.md | TEST: tests/unit/test_detector.py
    def get_supported_languages_for_workflows(self) -> list[str]:
        """Get list of languages with dedicated workflow templates.

        Returns languages that have pre-built GitHub Actions workflow templates
        with language-specific testing, linting, and coverage configuration.

        Returns:
            List of supported language names (lowercase)

        Example:
            >>> detector = LanguageDetector()
            >>> detector.get_supported_languages_for_workflows()
            ['python', 'javascript', 'typescript', 'go']

        Note:
            While LanguageDetector can detect 20+ languages, only these 4
            have dedicated CI/CD workflow templates. Other languages fall back
            to generic workflows.
        """
        return ["python", "javascript", "typescript", "go"]


def detect_project_language(path: str | Path = ".") -> str | None:
    """Detect the project language (helper).

    Args:
        path: Directory to inspect (default: current directory).

    Returns:
        Detected language name (lowercase) or None.
    """
    detector = LanguageDetector()
    return detector.detect(path)
