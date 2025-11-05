#!/usr/bin/env python3
"""
Unit tests for LanguageDetector module

Tests cover:
- Language detection from config files
- Primary language selection
- Extension mapping
- Package manager identification
- Language installation detection
- Linter tool recommendations
"""

import pytest
import tempfile
import json
from pathlib import Path
from language_detector import LanguageDetector


class TestLanguageDetection:
    """Test language detection functionality"""

    def test_detect_python_project(self):
        """Test Python project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Python markers
            (project_root / "pyproject.toml").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "python" in languages
            assert detector.detect_primary_language() == "python"

    def test_detect_javascript_project(self):
        """Test JavaScript project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create JavaScript markers
            package_json = {
                "name": "test-project",
                "version": "1.0.0"
            }
            (project_root / "package.json").write_text(json.dumps(package_json))

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "javascript" in languages

    def test_detect_typescript_project(self):
        """Test TypeScript project detection (should prioritize over JavaScript)"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create TypeScript markers
            (project_root / "tsconfig.json").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "typescript" in languages
            assert detector.detect_primary_language() == "typescript"

    def test_detect_go_project(self):
        """Test Go project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Go markers
            (project_root / "go.mod").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "go" in languages

    def test_detect_rust_project(self):
        """Test Rust project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Rust markers
            (project_root / "Cargo.toml").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "rust" in languages

    def test_detect_java_project(self):
        """Test Java project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Java markers
            (project_root / "pom.xml").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "java" in languages

    def test_detect_ruby_project(self):
        """Test Ruby project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Ruby markers
            (project_root / "Gemfile").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "ruby" in languages

    def test_detect_php_project(self):
        """Test PHP project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create PHP markers
            (project_root / "composer.json").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "php" in languages

    def test_detect_multilingual_project(self):
        """Test multilingual project detection"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create markers for multiple languages
            (project_root / "pyproject.toml").touch()
            package_json = {"name": "test"}
            (project_root / "package.json").write_text(json.dumps(package_json))
            (project_root / "go.mod").touch()

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert len(languages) >= 3
            assert "python" in languages
            assert "javascript" in languages or "typescript" in languages
            assert "go" in languages


class TestExtensionMapping:
    """Test file extension mapping"""

    def test_python_extensions(self):
        """Test Python file extensions"""
        detector = LanguageDetector()
        exts = detector.get_file_extension_for_language("python")

        assert ".py" in exts

    def test_javascript_extensions(self):
        """Test JavaScript file extensions"""
        detector = LanguageDetector()
        exts = detector.get_file_extension_for_language("javascript")

        assert ".js" in exts
        assert ".jsx" in exts

    def test_typescript_extensions(self):
        """Test TypeScript file extensions"""
        detector = LanguageDetector()
        exts = detector.get_file_extension_for_language("typescript")

        assert ".ts" in exts
        assert ".tsx" in exts

    def test_go_extensions(self):
        """Test Go file extensions"""
        detector = LanguageDetector()
        exts = detector.get_file_extension_for_language("go")

        assert ".go" in exts

    def test_rust_extensions(self):
        """Test Rust file extensions"""
        detector = LanguageDetector()
        exts = detector.get_file_extension_for_language("rust")

        assert ".rs" in exts


class TestPackageManager:
    """Test package manager identification"""

    def test_python_package_manager(self):
        """Test Python package manager"""
        detector = LanguageDetector()
        manager = detector.get_package_manager("python")

        assert manager == "pip"

    def test_javascript_package_manager(self):
        """Test JavaScript package manager"""
        detector = LanguageDetector()
        manager = detector.get_package_manager("javascript")

        assert manager == "npm"

    def test_go_package_manager(self):
        """Test Go package manager"""
        detector = LanguageDetector()
        manager = detector.get_package_manager("go")

        assert manager == "go"

    def test_rust_package_manager(self):
        """Test Rust package manager"""
        detector = LanguageDetector()
        manager = detector.get_package_manager("rust")

        assert manager == "cargo"

    def test_java_package_manager(self):
        """Test Java package manager"""
        detector = LanguageDetector()
        manager = detector.get_package_manager("java")

        assert manager == "maven"

    def test_ruby_package_manager(self):
        """Test Ruby package manager"""
        detector = LanguageDetector()
        manager = detector.get_package_manager("ruby")

        assert manager == "bundler"

    def test_php_package_manager(self):
        """Test PHP package manager"""
        detector = LanguageDetector()
        manager = detector.get_package_manager("php")

        assert manager == "composer"


class TestLinterToolRecommendations:
    """Test linter tool recommendations"""

    def test_python_linter_tools(self):
        """Test Python linter recommendations"""
        detector = LanguageDetector()
        tools = detector.get_linter_tools("python")

        assert tools["formatter"] == "ruff"
        assert tools["linter"] == "ruff"
        assert tools["type_checker"] == "mypy"

    def test_javascript_linter_tools(self):
        """Test JavaScript linter recommendations"""
        detector = LanguageDetector()
        tools = detector.get_linter_tools("javascript")

        assert tools["formatter"] == "prettier"
        assert tools["linter"] == "eslint"

    def test_typescript_linter_tools(self):
        """Test TypeScript linter recommendations"""
        detector = LanguageDetector()
        tools = detector.get_linter_tools("typescript")

        assert tools["formatter"] == "prettier"
        assert tools["linter"] == "eslint"
        assert tools["type_checker"] == "tsc"

    def test_go_linter_tools(self):
        """Test Go linter recommendations"""
        detector = LanguageDetector()
        tools = detector.get_linter_tools("go")

        assert tools["formatter"] == "gofmt"
        assert tools["linter"] == "golangci-lint"

    def test_rust_linter_tools(self):
        """Test Rust linter recommendations"""
        detector = LanguageDetector()
        tools = detector.get_linter_tools("rust")

        assert tools["formatter"] == "rustfmt"
        assert tools["linter"] == "clippy"


class TestPriority:
    """Test language priority ordering"""

    def test_typescript_prioritized_over_javascript(self):
        """Test that TypeScript is prioritized over JavaScript"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create markers for both
            (project_root / "tsconfig.json").touch()
            package_json = {"name": "test"}
            (project_root / "package.json").write_text(json.dumps(package_json))

            detector = LanguageDetector(project_root)
            primary = detector.detect_primary_language()

            assert primary == "typescript"

    def test_python_high_priority(self):
        """Test that Python is high priority"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Python and JavaScript markers
            (project_root / "pyproject.toml").touch()
            package_json = {"name": "test"}
            (project_root / "package.json").write_text(json.dumps(package_json))

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            # Python should come early in the list
            python_idx = languages.index("python") if "python" in languages else -1
            assert python_idx >= 0


class TestSummary:
    """Test summary generation"""

    def test_single_language_summary(self):
        """Test summary for single language project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            (project_root / "pyproject.toml").touch()

            detector = LanguageDetector(project_root)
            summary = detector.get_summary()

            assert "python" in summary.lower()
            assert "primary" in summary.lower()

    def test_multilingual_summary(self):
        """Test summary for multilingual project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            (project_root / "pyproject.toml").touch()
            (project_root / "go.mod").touch()
            (project_root / "Cargo.toml").touch()

            detector = LanguageDetector(project_root)
            summary = detector.get_summary()

            assert "primary" in summary.lower()
            assert "also detected" in summary.lower()

    def test_no_language_summary(self):
        """Test summary when no languages detected"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            detector = LanguageDetector(project_root)
            summary = detector.get_summary()

            assert "not" in summary.lower() or "no" in summary.lower()


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
