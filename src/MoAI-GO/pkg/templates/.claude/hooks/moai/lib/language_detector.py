#!/usr/bin/env python3
"""Language Detection Library for Multi-Language Support.

Detects programming language from file path and project configuration.
Supports Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin, Ruby, PHP, C/C++.
"""

from __future__ import annotations

from pathlib import Path

# Language configuration with test file patterns
LANGUAGE_CONFIGS: dict[str, dict] = {
    "Python": {
        "extensions": [".py", ".pyi"],
        "indicators": ["pyproject.toml", "requirements.txt", "setup.py", "setup.cfg", "poetry.lock"],
        "test_frameworks": ["pytest", "unittest"],
        "coverage_tools": ["coverage.py", "pytest-cov"],
        "test_patterns": ["tests/test_*.py", "test_*.py"],
        "source_patterns": ["*.py", "src/**/*.py"],
        "test_command": "uv run pytest -q tests/",
        "coverage_command": "uv run coverage run --branch -m pytest -q tests/ && uv run coverage report --format=json",
    },
    "TypeScript": {
        "extensions": [".ts", ".tsx"],
        "indicators": ["package.json", "tsconfig.json"],
        "test_frameworks": ["jest", "vitest", "mocha"],
        "coverage_tools": ["c8", "istanbul", "vitest"],
        "test_patterns": ["**/*.test.ts", "**/*.spec.ts", "**/*.test.tsx"],
        "source_patterns": ["*.ts", "*.tsx", "src/**/*.ts"],
        "test_command": "npm test -- --run",
        "coverage_command": "npm run test:coverage -- --reporter=json",
    },
    "JavaScript": {
        "extensions": [".js", ".jsx"],
        "indicators": ["package.json"],
        "test_frameworks": ["jest", "vitest", "mocha"],
        "coverage_tools": ["c8", "istanbul", "nyc"],
        "test_patterns": ["**/*.test.js", "**/*.spec.js"],
        "source_patterns": ["*.js", "*.jsx", "src/**/*.js"],
        "test_command": "npm test -- --run",
        "coverage_command": "npm run test:coverage -- --coverageReporters=json",
    },
    "Go": {
        "extensions": [".go"],
        "indicators": ["go.mod", "go.sum"],
        "test_frameworks": ["testing", "testify"],
        "coverage_tools": ["go test -cover"],
        "test_patterns": ["*_test.go"],
        "source_patterns": ["*.go", "cmd/**/*.go", "pkg/**/*.go"],
        "test_command": "go test ./...",
        "coverage_command": "go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out",
    },
    "Rust": {
        "extensions": [".rs"],
        "indicators": ["Cargo.toml", "Cargo.lock"],
        "test_frameworks": ["cargo test"],
        "coverage_tools": ["tarpaulin", "cargo-llvm-cov"],
        "test_patterns": ["**/*_test.rs", "tests/**/*.rs"],
        "source_patterns": ["*.rs", "src/**/*.rs"],
        "test_command": "cargo test",
        "coverage_command": "cargo tarpaulin --out Json -- --test-threads=1",
    },
    "Java": {
        "extensions": [".java"],
        "indicators": ["pom.xml", "build.gradle", "build.gradle.kts", "gradle.properties"],
        "test_frameworks": ["JUnit"],
        "coverage_tools": ["JaCoCo"],
        "test_patterns": ["**/*Test.java", "src/test/**/*.java"],
        "source_patterns": ["*.java", "src/main/**/*.java"],
        "test_command": "mvn test",
        "coverage_command": "mvn jacoco:report",
    },
    "Kotlin": {
        "extensions": [".kt", ".kts"],
        "indicators": ["build.gradle.kts", "pom.xml", "gradle.properties"],
        "test_frameworks": ["JUnit", "Kotest"],
        "coverage_tools": ["JaCoCo", "Kover"],
        "test_patterns": ["**/*Test.kt", "src/test/**/*.kt"],
        "source_patterns": ["*.kt", "src/main/**/*.kt"],
        "test_command": "gradle test",
        "coverage_command": "gradle jacocoTestReport",
    },
    "Ruby": {
        "extensions": [".rb"],
        "indicators": ["Gemfile", "Rakefile"],
        "test_frameworks": ["RSpec", "Minitest"],
        "coverage_tools": ["SimpleCov"],
        "test_patterns": ["**/*_spec.rb", "test/**/*.rb"],
        "source_patterns": ["*.rb", "lib/**/*.rb"],
        "test_command": "bundle exec rspec",
        "coverage_command": "COVERAGE=true bundle exec rspec",
    },
    "PHP": {
        "extensions": [".php"],
        "indicators": ["composer.json"],
        "test_frameworks": ["Pest", "PHPUnit"],
        "coverage_tools": ["phpunit coverage", "pest coverage"],
        "test_patterns": ["**/*Test.php", "tests/**/*.php"],
        "source_patterns": ["*.php", "src/**/*.php"],
        "test_command": "vendor/bin/pest",
        "coverage_command": "vendor/bin/pest --coverage",
    },
    "C": {
        "extensions": [".c", ".cpp", ".cc", ".h", ".hpp"],
        "indicators": ["CMakeLists.txt", "Makefile", "configure.ac"],
        "test_frameworks": ["googletest", "Catch2"],
        "coverage_tools": ["gcov", "lcov"],
        "test_patterns": ["**/*_test.cc", "**/*_test.c", "tests/**/*.c"],
        "source_patterns": ["*.c", "*.cpp", "*.cc", "*.h", "*.hpp", "src/**/*.c"],
        "test_command": "make test",
        "coverage_command": "make coverage",
    },
}


def detect_language(file_path: Path) -> str | None:
    """Detect programming language from file path and project context.

    Args:
        file_path: Path to the source file

    Returns:
        Language name or None if unknown
    """
    # First try: File extension
    ext = file_path.suffix.lower()

    for language, config in LANGUAGE_CONFIGS.items():
        if ext in config["extensions"]:
            return language

    # Second try: Project indicators
    project_root = _find_project_root(file_path)
    if project_root:
        for language, config in LANGUAGE_CONFIGS.items():
            for indicator in config["indicators"]:
                if (project_root / indicator).exists():
                    # Verify extension matches
                    if ext in config.get("extensions", []):
                        return language

    return None


def _find_project_root(file_path: Path) -> Path | None:
    """Find project root directory by looking for indicator files.

    Args:
        file_path: Starting file path

    Returns:
        Path to project root or None
    """
    current = file_path.parent if file_path.is_file() else file_path

    # Common project root indicators
    root_indicators = [
        "pyproject.toml",
        "package.json",
        "go.mod",
        "Cargo.toml",
        "pom.xml",
        "build.gradle",
        "Gemfile",
        "composer.json",
        "CMakeLists.txt",
        ".git",
    ]

    # Search up the directory tree
    max_depth = 5
    depth = 0

    while depth < max_depth and current != current.parent:
        for indicator in root_indicators:
            if (current / indicator).exists():
                return current

        current = current.parent
        depth += 1

    return None


def get_test_file_path(source_file: Path) -> Path | None:
    """Determine the expected test file path for a source file.

    Args:
        source_file: Path to the source file

    Returns:
        Expected test file path or None
    """
    language = detect_language(source_file)
    if not language:
        return None

    stem = source_file.stem

    # Language-specific test file conventions
    if language == "Python":
        # module.py → tests/test_module.py
        # src/module.py → tests/test_module.py
        return Path("tests") / f"test_{stem}.py"

    elif language in ("TypeScript", "JavaScript"):
        # module.ts → module.test.ts
        return source_file.parent / f"{stem}.test{source_file.suffix}"

    elif language == "Go":
        # module.go → module_test.go (same directory)
        return source_file.parent / f"{stem}_test.go"

    elif language == "Rust":
        # module.rs → module_test.rs (same directory for unit tests)
        # or tests/module_test.rs (integration tests)
        return source_file.parent / f"{stem}_test.rs"

    elif language == "Java":
        # Module.java → ModuleTest.java
        # Usually in src/test/java/.../ModuleTest.java
        return source_file.parent / f"{stem}Test.java"

    elif language == "Kotlin":
        # Module.kt → ModuleTest.kt
        return source_file.parent / f"{stem}Test.kt"

    elif language == "Ruby":
        # module.rb → module_spec.rb
        return source_file.parent / f"{stem}_spec.rb"

    elif language == "PHP":
        # Module.php → ModuleTest.php
        return source_file.parent / f"{stem}Test.php"

    elif language == "C":
        # module.cc → module_test.cc
        return source_file.parent / f"{stem}_test.cc"

    return None


def get_test_command(language: str) -> str | None:
    """Get test command for a language.

    Args:
        language: Programming language name

    Returns:
        Test command string or None
    """
    config = LANGUAGE_CONFIGS.get(language, {})
    return config.get("test_command")


def get_coverage_command(language: str) -> str | None:
    """Get coverage command for a language.

    Args:
        language: Programming language name

    Returns:
        Coverage command string or None
    """
    config = LANGUAGE_CONFIGS.get(language, {})
    return config.get("coverage_command")


def get_coverage_target() -> int:
    """Get target coverage percentage from quality.yaml.

    Returns:
        Target coverage percentage (default: 100)
    """
    # Try to read from config
    config_path = Path(".moai/config/sections/quality.yaml")
    if config_path.exists():
        import yaml

        try:
            with open(config_path) as f:
                config = yaml.safe_load(f)
                return config.get("constitution", {}).get("test_coverage_target", 100)
        except Exception:
            pass

    return 100


__all__ = [
    "LANGUAGE_CONFIGS",
    "detect_language",
    "get_test_file_path",
    "get_test_command",
    "get_coverage_command",
    "get_coverage_target",
]
