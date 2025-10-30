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
        # JavaScript/TypeScript moved to third for priority (Issue #131 fix)
        # @CODE:LANG-DETECT-JS-001 | ISSUE: #131 Language Detection Priority
        "typescript": [
            "*.ts",
            "tsconfig.json",
            "vite.config.ts",         # Vite (TypeScript): build config
            "nest-cli.json"           # NestJS: framework config
        ],
        # JavaScript patterns must come after TypeScript due to overlap
        # (tsconfig.json implies TypeScript, but package.json could mean JS or TS)
        "javascript": [
            "*.js",
            "package.json",
            "next.config.js",         # Next.js: framework config (can be JS or TS, but indicates JS primary)
            "vite.config.js",         # Vite (JavaScript): build config
            "webpack.config.js",      # Webpack: bundler config (unique to JS projects)
            "src/index.js",           # Express/Node.js: common entry point
            "server.js"               # Express/Node.js: server file (unique identifier)
        ],
        "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
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


def detect_project_language(path: str | Path = ".") -> str | None:
    """Detect the project language (helper).

    Args:
        path: Directory to inspect (default: current directory).

    Returns:
        Detected language name (lowercase) or None.
    """
    detector = LanguageDetector()
    return detector.detect(path)
