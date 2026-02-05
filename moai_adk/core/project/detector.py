# TEST: tests/unit/test_language_detector_extended.py
"""Language detector module.

Automatically detects 20 programming languages.

Extended detection supports:
- 11 new languages: Ruby, PHP, Java, Rust, Dart, Swift, Kotlin, C#, C, C++, Shell
- 5 build tool detection: Maven, Gradle, CMake, SPM, dotnet
- Package manager detection: bundle, composer, cargo
- Priority conflict resolution for multi-language projects
"""

from pathlib import Path


class LanguageDetector:
    """Automatically detect up to 20 programming languages.

    Prioritizes framework-specific files (e.g., Laravel, Django) over
    generic language files to improve accuracy in mixed-language projects.
    """

    LANGUAGE_PATTERNS = {
        # Priority order (highest to lowest):
        # 1. Rust, 2. Dart, 3. Swift, 4. Kotlin, 5. C#, 6. Java, 7. Ruby, 8. PHP
        # 9. Go, 10. Python, 11. TypeScript, 12. JavaScript, 13. C++, 14. C, 15. Shell
        "rust": ["*.rs", "Cargo.toml"],
        "dart": ["*.dart", "pubspec.yaml"],
        "swift": ["*.swift", "Package.swift"],
        "kotlin": ["*.kt", "build.gradle.kts"],
        "csharp": ["*.cs", "*.csproj"],
        "java": ["*.java", "pom.xml", "build.gradle"],
        # Ruby moved for priority (Rails detection)
        "ruby": [
            "*.rb",
            "Gemfile",
            "Gemfile.lock",  # Bundler: lock file (unique to Ruby)
            "config/routes.rb",  # Rails: routing file (unique identifier)
            "app/controllers/",  # Rails: controller directory
            "Rakefile",  # Rails/Ruby: task file
        ],
        # PHP after Ruby (Laravel detection)
        "php": [
            "*.php",
            "composer.json",
            "artisan",  # Laravel: CLI tool (unique identifier)
            "app/",  # Laravel: application directory
            "bootstrap/laravel.php",  # Laravel: bootstrap file
        ],
        "go": ["*.go", "go.mod"],
        "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
        "typescript": ["*.ts", "tsconfig.json"],
        "javascript": ["*.js", "package.json"],
        "cpp": ["*.cpp", "CMakeLists.txt"],
        "c": ["*.c", "Makefile"],
        "shell": ["*.sh", "*.bash"],
        # Additional languages (lower priority)
        "elixir": ["*.ex", "mix.exs"],
        "scala": ["*.scala", "build.sbt"],
        "clojure": ["*.clj", "project.clj"],
        "haskell": ["*.hs", "*.cabal"],
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

    def get_workflow_template_path(self, language: str) -> str:
        """Get the GitHub Actions workflow template path for a language.


        Args:
            language: Programming language name (lowercase).

        Returns:
            Workflow template file path relative to templates directory.

        Raises:
            ValueError: If language is not supported for workflows.
        """
        # Language-specific workflow template mapping
        workflow_templates = {
            "python": "python-tag-validation.yml",
            "javascript": "javascript-tag-validation.yml",
            "typescript": "typescript-tag-validation.yml",
            "go": "go-tag-validation.yml",
        }

        if language not in workflow_templates:
            raise ValueError(f"Language '{language}' is not supported for workflows")

        return workflow_templates[language]

    def detect_package_manager(self, path: str | Path = ".") -> str | None:
        """Detect the package manager for the detected language.


        Args:
            path: Directory to inspect.

        Returns:
            Package manager name or None if not detected.
        """
        path = Path(path)

        # Ruby
        if (path / "Gemfile").exists():
            return "bundle"

        # PHP
        if (path / "composer.json").exists():
            return "composer"

        # Java/Kotlin
        if (path / "pom.xml").exists():
            return "maven"
        if (path / "build.gradle").exists() or (path / "build.gradle.kts").exists():
            return "gradle"

        # Rust
        if (path / "Cargo.toml").exists():
            return "cargo"

        # Dart/Flutter
        if (path / "pubspec.yaml").exists():
            return "dart_pub"

        # Swift
        if (path / "Package.swift").exists():
            return "spm"

        # C#
        if list(path.glob("*.csproj")) or list(path.glob("*.sln")):
            return "dotnet"

        # Python
        if (path / "pyproject.toml").exists():
            return "pip"

        # JavaScript/TypeScript (check in priority order)
        # Check for lock files and package managers
        if (path / "bun.lockb").exists():
            return "bun"
        elif (path / "pnpm-lock.yaml").exists():
            return "pnpm"
        elif (path / "yarn.lock").exists():
            return "yarn"
        elif (path / "package-lock.json").exists():
            return "npm"
        elif (path / "package.json").exists():
            # Default to npm for package.json without lock files
            return "npm"

        # Go
        if (path / "go.mod").exists():
            return "go_modules"

        return None

    def detect_build_tool(self, path: str | Path = ".", language: str | None = None) -> str | None:
        """Detect the build tool for the detected language.


        Args:
            path: Directory to inspect.
            language: Optional language hint for disambiguation.

        Returns:
            Build tool name or None if not detected.
        """
        path = Path(path)

        # C/C++
        if (path / "CMakeLists.txt").exists():
            return "cmake"
        if (path / "Makefile").exists():
            return "make"

        # Java/Kotlin
        if language in ["java", "kotlin"]:
            if (path / "pom.xml").exists():
                return "maven"
            if (path / "build.gradle").exists() or (path / "build.gradle.kts").exists():
                return "gradle"

        # Rust
        if (path / "Cargo.toml").exists():
            return "cargo"

        # Swift
        if (path / "Package.swift").exists():
            return "spm"
        if list(path.glob("*.xcodeproj")) or list(path.glob("*.xcworkspace")):
            return "xcode"

        # C#
        if list(path.glob("*.csproj")) or list(path.glob("*.sln")):
            return "dotnet"

        return None

    def get_supported_languages_for_workflows(self) -> list[str]:
        """Get the list of languages with dedicated CI/CD workflow support.


        Returns:
            List of supported language names (15 total).
        """
        return [
            "python",
            "javascript",
            "typescript",
            "go",
            "ruby",
            "php",
            "java",
            "rust",
            "dart",
            "swift",
            "kotlin",
            "csharp",
            "c",
            "cpp",
            "shell",
        ]


def detect_project_language(path: str | Path = ".") -> str | None:
    """Detect the project language (helper).

    Args:
        path: Directory to inspect (default: current directory).

    Returns:
        Detected language name (lowercase) or None.
    """
    detector = LanguageDetector()
    return detector.detect(path)
