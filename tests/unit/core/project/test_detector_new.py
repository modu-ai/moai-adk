"""
Comprehensive tests for LanguageDetector module.

Tests cover:
- LanguageDetector class initialization
- detect method for single language detection
- detect_multiple method for multi-language projects
- detect_package_manager method
- detect_build_tool method
- detect_project_language helper function
"""

import tempfile
from pathlib import Path
from unittest import mock
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.project.detector import (
    LanguageDetector,
    detect_project_language,
)


class TestLanguageDetector:
    """Test suite for LanguageDetector class."""

    def test_initialization(self):
        """Test LanguageDetector initialization."""
        # Arrange & Act
        detector = LanguageDetector()

        # Assert
        assert detector is not None
        assert hasattr(detector, "LANGUAGE_PATTERNS")
        assert len(detector.LANGUAGE_PATTERNS) > 0

    def test_language_patterns_contains_major_languages(self):
        """Test LANGUAGE_PATTERNS contains all major languages."""
        # Arrange
        detector = LanguageDetector()

        # Act
        languages = list(detector.LANGUAGE_PATTERNS.keys())

        # Assert
        assert "python" in languages
        assert "javascript" in languages
        assert "typescript" in languages
        assert "go" in languages
        assert "rust" in languages
        assert "java" in languages
        assert "ruby" in languages
        assert "php" in languages

    def test_detect_python_project(self):
        """Test detect returns 'python' for Python project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "main.py").write_text("print('hello')")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "python"

    def test_detect_javascript_project(self):
        """Test detect returns 'javascript' for JavaScript project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "package.json").write_text("{}")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "javascript"

    def test_detect_typescript_project(self):
        """Test detect returns 'typescript' for TypeScript project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "tsconfig.json").write_text("{}")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "typescript"

    def test_detect_go_project(self):
        """Test detect returns 'go' for Go project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "go.mod").write_text("")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "go"

    def test_detect_rust_project(self):
        """Test detect returns 'rust' for Rust project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "Cargo.toml").write_text("")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "rust"

    def test_detect_java_project(self):
        """Test detect returns 'java' for Java project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "pom.xml").write_text("")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "java"

    def test_detect_ruby_project(self):
        """Test detect returns 'ruby' for Ruby project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "Gemfile").write_text("")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "ruby"

    def test_detect_php_project(self):
        """Test detect returns 'php' for PHP project."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "composer.json").write_text("")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "php"

    def test_detect_returns_none_for_empty_directory(self):
        """Test detect returns None for directory with no language files."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            result = detector.detect(path)

            # Assert
            assert result is None

    def test_detect_with_string_path(self):
        """Test detect works with string paths."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            (Path(temp_dir) / "main.py").write_text("")

            # Act
            result = detector.detect(temp_dir)

            # Assert
            assert result == "python"

    def test_detect_uses_priority_order(self):
        """Test detect respects language priority order."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            # Create both Rust and Python files
            (path / "main.rs").write_text("")
            (path / "main.py").write_text("")

            # Act
            result = detector.detect(path)

            # Assert - Rust should be detected first due to priority
            assert result == "rust"

    def test_detect_multiple_returns_list(self):
        """Test detect_multiple returns list of detected languages."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "main.py").write_text("")
            (path / "main.js").write_text("")

            # Act
            result = detector.detect_multiple(path)

            # Assert
            assert isinstance(result, list)
            assert "python" in result
            assert "javascript" in result

    def test_detect_multiple_with_no_languages(self):
        """Test detect_multiple returns empty list for no languages."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            result = detector.detect_multiple(path)

            # Assert
            assert result == []

    def test_detect_package_manager_detects_npm(self):
        """Test detect_package_manager detects npm."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "package-lock.json").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "npm"

    def test_detect_package_manager_detects_yarn(self):
        """Test detect_package_manager detects yarn."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "yarn.lock").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "yarn"

    def test_detect_package_manager_detects_pnpm(self):
        """Test detect_package_manager detects pnpm."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "pnpm-lock.yaml").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "pnpm"

    def test_detect_package_manager_detects_bun(self):
        """Test detect_package_manager detects bun."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "bun.lockb").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "bun"

    def test_detect_package_manager_detects_pip(self):
        """Test detect_package_manager detects pip."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "pyproject.toml").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "pip"

    def test_detect_package_manager_detects_cargo(self):
        """Test detect_package_manager detects cargo."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "Cargo.toml").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "cargo"

    def test_detect_package_manager_detects_maven(self):
        """Test detect_package_manager detects maven."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "pom.xml").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "maven"

    def test_detect_package_manager_detects_gradle(self):
        """Test detect_package_manager detects gradle."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "build.gradle").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "gradle"

    def test_detect_package_manager_returns_none(self):
        """Test detect_package_manager returns None for unknown."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result is None

    def test_detect_build_tool_detects_cmake(self):
        """Test detect_build_tool detects CMake."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "CMakeLists.txt").write_text("")

            # Act
            result = detector.detect_build_tool(path)

            # Assert
            assert result == "cmake"

    def test_detect_build_tool_detects_make(self):
        """Test detect_build_tool detects Make."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "Makefile").write_text("")

            # Act
            result = detector.detect_build_tool(path)

            # Assert
            assert result == "make"

    def test_detect_build_tool_with_language_hint(self):
        """Test detect_build_tool uses language hint."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "pom.xml").write_text("")

            # Act
            result = detector.detect_build_tool(path, language="java")

            # Assert
            assert result == "maven"

    def test_detect_build_tool_returns_none(self):
        """Test detect_build_tool returns None for unknown."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            result = detector.detect_build_tool(path)

            # Assert
            assert result is None

    def test_get_workflow_template_path_for_python(self):
        """Test get_workflow_template_path returns correct path for Python."""
        # Arrange
        detector = LanguageDetector()

        # Act
        result = detector.get_workflow_template_path("python")

        # Assert
        assert "python-tag-validation.yml" in result

    def test_get_workflow_template_path_for_typescript(self):
        """Test get_workflow_template_path returns correct path for TypeScript."""
        # Arrange
        detector = LanguageDetector()

        # Act
        result = detector.get_workflow_template_path("typescript")

        # Assert
        assert "typescript-tag-validation.yml" in result

    def test_get_workflow_template_path_for_go(self):
        """Test get_workflow_template_path returns correct path for Go."""
        # Arrange
        detector = LanguageDetector()

        # Act
        result = detector.get_workflow_template_path("go")

        # Assert
        assert "go-tag-validation.yml" in result

    def test_get_workflow_template_path_raises_for_unsupported(self):
        """Test get_workflow_template_path raises for unsupported language."""
        # Arrange
        detector = LanguageDetector()

        # Act & Assert
        with pytest.raises(ValueError):
            detector.get_workflow_template_path("unsupported_language")

    def test_get_supported_languages_for_workflows(self):
        """Test get_supported_languages_for_workflows returns list."""
        # Arrange
        detector = LanguageDetector()

        # Act
        result = detector.get_supported_languages_for_workflows()

        # Assert
        assert isinstance(result, list)
        assert len(result) > 0
        assert "python" in result
        assert "typescript" in result
        assert "go" in result

    def test_check_patterns_with_extension_pattern(self):
        """Test _check_patterns matches file extensions."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "main.py").write_text("")

            # Act
            result = detector._check_patterns(path, ["*.py"])

            # Assert
            assert result is True

    def test_check_patterns_with_file_pattern(self):
        """Test _check_patterns matches specific files."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "Gemfile").write_text("")

            # Act
            result = detector._check_patterns(path, ["Gemfile"])

            # Assert
            assert result is True

    def test_check_patterns_returns_false_for_no_match(self):
        """Test _check_patterns returns False when no patterns match."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            result = detector._check_patterns(path, ["*.py", "Gemfile"])

            # Assert
            assert result is False

    def test_detect_project_language_helper_function(self):
        """Test detect_project_language helper function."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "main.py").write_text("")

            # Act
            result = detect_project_language(path)

            # Assert
            assert result == "python"

    def test_detect_priority_dart_over_c_plus_plus(self):
        """Test detect respects language priority."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "main.dart").write_text("")
            (path / "main.cpp").write_text("")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "dart"

    def test_detect_with_nested_files(self):
        """Test detect finds files in nested directories."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            src_dir = path / "src"
            src_dir.mkdir()
            (src_dir / "main.py").write_text("")

            # Act
            result = detector.detect(path)

            # Assert
            assert result == "python"

    def test_detect_package_manager_priority_bun_over_npm(self):
        """Test detect_package_manager uses priority for package managers."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            (path / "bun.lockb").write_text("")
            (path / "package-lock.json").write_text("")

            # Act
            result = detector.detect_package_manager(path)

            # Assert
            assert result == "bun"

    def test_detect_package_manager_with_string_path(self):
        """Test detect_package_manager works with string paths."""
        # Arrange
        detector = LanguageDetector()

        with tempfile.TemporaryDirectory() as temp_dir:
            (Path(temp_dir) / "package.json").write_text("")

            # Act
            result = detector.detect_package_manager(temp_dir)

            # Assert
            assert result == "npm"
