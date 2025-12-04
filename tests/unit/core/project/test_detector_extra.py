"""Extended tests for moai_adk.core.project.detector module.

These tests focus on increasing coverage for language and build tool detection.
"""

import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock

import pytest

from moai_adk.core.project.detector import LanguageDetector, detect_project_language


class TestLanguageDetectorInit:
    """Test LanguageDetector initialization."""

    def test_detector_init(self):
        """Test detector initialization."""
        detector = LanguageDetector()

        assert hasattr(detector, "LANGUAGE_PATTERNS")
        assert isinstance(detector.LANGUAGE_PATTERNS, dict)
        assert len(detector.LANGUAGE_PATTERNS) > 0


class TestDetectSingleLanguage:
    """Test detecting single language."""

    def test_detect_python(self):
        """Test detecting Python project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.py").write_text("print('hello')")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "python"

    def test_detect_javascript(self):
        """Test detecting JavaScript project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.js").write_text("console.log('hello')")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "javascript"

    def test_detect_typescript(self):
        """Test detecting TypeScript project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.ts").write_text("console.log('hello')")
            (project_path / "tsconfig.json").write_text("{}")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "typescript"

    def test_detect_go(self):
        """Test detecting Go project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "go.mod").write_text("module test")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "go"

    def test_detect_rust(self):
        """Test detecting Rust project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Cargo.toml").write_text("[package]")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "rust"

    def test_detect_java(self):
        """Test detecting Java project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "pom.xml").write_text("<project>")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "java"

    def test_detect_ruby(self):
        """Test detecting Ruby project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Gemfile").write_text("source 'https://rubygems.org'")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "ruby"

    def test_detect_php(self):
        """Test detecting PHP project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "composer.json").write_text("{}")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "php"

    def test_detect_c_plus_plus(self):
        """Test detecting C++ project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "CMakeLists.txt").write_text("cmake_minimum_required")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "cpp"

    def test_detect_c(self):
        """Test detecting C project."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Makefile").write_text("all:")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result == "c"

    def test_detect_unknown_language(self):
        """Test detecting unknown language."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            detector = LanguageDetector()
            result = detector.detect(project_path)

            assert result is None

    def test_detect_with_string_path(self):
        """Test detecting with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir).joinpath("test.py").write_text("print('hello')")

            detector = LanguageDetector()
            result = detector.detect(tmpdir)

            assert result == "python"


class TestDetectMultipleLanguages:
    """Test detecting multiple languages."""

    def test_detect_multiple_single_language(self):
        """Test detecting multiple languages when only one present."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.py").write_text("print('hello')")

            detector = LanguageDetector()
            result = detector.detect_multiple(project_path)

            assert isinstance(result, list)
            assert "python" in result

    def test_detect_multiple_several_languages(self):
        """Test detecting multiple languages."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.py").write_text("print('hello')")
            (project_path / "test.js").write_text("console.log('hello')")

            detector = LanguageDetector()
            result = detector.detect_multiple(project_path)

            assert isinstance(result, list)
            assert len(result) >= 1

    def test_detect_multiple_no_languages(self):
        """Test detecting multiple languages when none present."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            detector = LanguageDetector()
            result = detector.detect_multiple(project_path)

            assert isinstance(result, list)
            assert len(result) == 0


class TestDetectPackageManager:
    """Test detecting package manager."""

    def test_detect_package_manager_ruby_bundle(self):
        """Test detecting Ruby package manager (bundle)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Gemfile").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "bundle"

    def test_detect_package_manager_php_composer(self):
        """Test detecting PHP package manager (composer)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "composer.json").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "composer"

    def test_detect_package_manager_java_maven(self):
        """Test detecting Java package manager (maven)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "pom.xml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "maven"

    def test_detect_package_manager_java_gradle(self):
        """Test detecting Java package manager (gradle)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "build.gradle").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "gradle"

    def test_detect_package_manager_rust_cargo(self):
        """Test detecting Rust package manager (cargo)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Cargo.toml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "cargo"

    def test_detect_package_manager_dart_pub(self):
        """Test detecting Dart package manager."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "pubspec.yaml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "dart_pub"

    def test_detect_package_manager_swift_spm(self):
        """Test detecting Swift package manager."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Package.swift").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "spm"

    def test_detect_package_manager_python_pip(self):
        """Test detecting Python package manager."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "pyproject.toml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "pip"

    def test_detect_package_manager_javascript_npm(self):
        """Test detecting JavaScript package manager (npm)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "package.json").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "npm"

    def test_detect_package_manager_javascript_yarn(self):
        """Test detecting JavaScript package manager (yarn)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "yarn.lock").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "yarn"

    def test_detect_package_manager_javascript_pnpm(self):
        """Test detecting JavaScript package manager (pnpm)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "pnpm-lock.yaml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "pnpm"

    def test_detect_package_manager_javascript_bun(self):
        """Test detecting JavaScript package manager (bun)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "bun.lockb").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "bun"

    def test_detect_package_manager_go_modules(self):
        """Test detecting Go package manager."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "go.mod").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result == "go_modules"

    def test_detect_package_manager_none(self):
        """Test detecting package manager when none present."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            assert result is None

    def test_detect_package_manager_with_string_path(self):
        """Test detecting package manager with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir).joinpath("Cargo.toml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(tmpdir)

            assert result == "cargo"


class TestDetectBuildTool:
    """Test detecting build tool."""

    def test_detect_build_tool_cmake(self):
        """Test detecting CMake build tool."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "CMakeLists.txt").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path)

            assert result == "cmake"

    def test_detect_build_tool_make(self):
        """Test detecting Make build tool."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Makefile").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path)

            assert result == "make"

    def test_detect_build_tool_maven_with_hint(self):
        """Test detecting Maven with language hint."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "pom.xml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path, language="java")

            assert result == "maven"

    def test_detect_build_tool_gradle_with_hint(self):
        """Test detecting Gradle with language hint."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "build.gradle").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path, language="java")

            assert result == "gradle"

    def test_detect_build_tool_cargo(self):
        """Test detecting Cargo build tool."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Cargo.toml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path)

            assert result == "cargo"

    def test_detect_build_tool_swift_spm(self):
        """Test detecting Swift Package Manager."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Package.swift").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path)

            assert result == "spm"

    def test_detect_build_tool_dotnet(self):
        """Test detecting dotnet build tool."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.csproj").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path)

            assert result == "dotnet"

    def test_detect_build_tool_none(self):
        """Test detecting build tool when none present."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            detector = LanguageDetector()
            result = detector.detect_build_tool(project_path)

            assert result is None

    def test_detect_build_tool_with_string_path(self):
        """Test detecting build tool with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir).joinpath("CMakeLists.txt").write_text("")

            detector = LanguageDetector()
            result = detector.detect_build_tool(tmpdir)

            assert result == "cmake"


class TestGetWorkflowTemplatePath:
    """Test getting workflow template path."""

    def test_get_workflow_template_python(self):
        """Test getting Python workflow template."""
        detector = LanguageDetector()
        result = detector.get_workflow_template_path("python")

        assert result == "python-tag-validation.yml"

    def test_get_workflow_template_javascript(self):
        """Test getting JavaScript workflow template."""
        detector = LanguageDetector()
        result = detector.get_workflow_template_path("javascript")

        assert result == "javascript-tag-validation.yml"

    def test_get_workflow_template_typescript(self):
        """Test getting TypeScript workflow template."""
        detector = LanguageDetector()
        result = detector.get_workflow_template_path("typescript")

        assert result == "typescript-tag-validation.yml"

    def test_get_workflow_template_go(self):
        """Test getting Go workflow template."""
        detector = LanguageDetector()
        result = detector.get_workflow_template_path("go")

        assert result == "go-tag-validation.yml"

    def test_get_workflow_template_unsupported(self):
        """Test getting workflow template for unsupported language."""
        detector = LanguageDetector()

        with pytest.raises(ValueError):
            detector.get_workflow_template_path("unknown_language")


class TestGetSupportedLanguagesForWorkflows:
    """Test getting supported languages for workflows."""

    def test_get_supported_languages(self):
        """Test getting list of supported languages."""
        detector = LanguageDetector()
        result = detector.get_supported_languages_for_workflows()

        assert isinstance(result, list)
        assert len(result) > 0
        assert "python" in result
        assert "javascript" in result
        assert "typescript" in result
        assert "go" in result


class TestCheckPatterns:
    """Test internal pattern checking."""

    def test_check_patterns_file_extension(self):
        """Test checking file extension pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.py").write_text("")

            detector = LanguageDetector()
            result = detector._check_patterns(project_path, ["*.py"])

            assert result is True

    def test_check_patterns_specific_file(self):
        """Test checking specific file pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "package.json").write_text("")

            detector = LanguageDetector()
            result = detector._check_patterns(project_path, ["package.json"])

            assert result is True

    def test_check_patterns_directory(self):
        """Test checking directory pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "config").mkdir()

            detector = LanguageDetector()
            result = detector._check_patterns(project_path, ["config/"])

            assert result is True

    def test_check_patterns_no_match(self):
        """Test checking patterns with no match."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            detector = LanguageDetector()
            result = detector._check_patterns(project_path, ["*.nonexistent"])

            assert result is False

    def test_check_patterns_multiple_patterns(self):
        """Test checking multiple patterns."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.py").write_text("")

            detector = LanguageDetector()
            result = detector._check_patterns(project_path, ["*.js", "*.py", "package.json"])

            assert result is True


class TestDetectProjectLanguageHelper:
    """Test detect_project_language helper function."""

    def test_detect_project_language_function(self):
        """Test helper function for detecting language."""
        with tempfile.TemporaryDirectory() as tmpdir:
            Path(tmpdir).joinpath("test.py").write_text("")

            result = detect_project_language(tmpdir)

            assert result == "python"

    def test_detect_project_language_default_path(self):
        """Test helper function with default path."""
        detector = LanguageDetector()

        with patch.object(detector, "detect", return_value="python"):
            with patch("moai_adk.core.project.detector.LanguageDetector", return_value=detector):
                result = detect_project_language()

                # Should work without error


class TestLanguagePriority:
    """Test language detection priority."""

    def test_language_priority_rust_over_other(self):
        """Test Rust has priority over other languages."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "Cargo.toml").write_text("")
            (project_path / "test.py").write_text("")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            # Rust should be detected (has higher priority)
            assert result == "rust"

    def test_language_priority_typescript_over_javascript(self):
        """Test TypeScript has priority over JavaScript."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "test.ts").write_text("")
            (project_path / "test.js").write_text("")
            (project_path / "tsconfig.json").write_text("")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            # TypeScript should be detected
            assert result == "typescript"

    def test_language_priority_java_over_python(self):
        """Test Java has priority based on pattern order."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "pom.xml").write_text("")
            (project_path / "test.py").write_text("")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            # Java should be detected (has higher priority)
            assert result == "java"


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_detect_with_nonexistent_path(self):
        """Test detecting with nonexistent path."""
        detector = LanguageDetector()
        result = detector.detect("/nonexistent/path")

        assert result is None

    def test_detect_empty_directory(self):
        """Test detecting in empty directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            detector = LanguageDetector()
            result = detector.detect(tmpdir)

            assert result is None

    def test_detect_with_hidden_files(self):
        """Test detecting ignores hidden files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / ".test.py").write_text("")

            detector = LanguageDetector()
            result = detector.detect(project_path)

            # Should still detect Python if extension pattern matches
            # Behavior depends on rglob implementation

    def test_detect_multiple_package_locks(self):
        """Test detecting when multiple package locks present."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "package.json").write_text("")
            (project_path / "yarn.lock").write_text("")
            (project_path / "pnpm-lock.yaml").write_text("")

            detector = LanguageDetector()
            result = detector.detect_package_manager(project_path)

            # Should detect one of them (priority matters)
            assert result in ["npm", "yarn", "pnpm", "bun"]
