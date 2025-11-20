"""Unit tests for extended language detection (11 new languages).

Tests for LanguageDetector class extensions:
- 11 language detection tests
- 5 build tool detection tests
- 4 priority conflict tests
- 3 error handling tests
- 4 backward compatibility tests
- 3 integration tests

Total: 30 acceptance criteria tests
"""

from pathlib import Path

from moai_adk.core.project.detector import LanguageDetector


class TestLanguageDetectionExtended:
    """Test 11 new language detection methods."""

    def test_detect_ruby(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-001-RUBY | Ruby project detection (Gemfile)."""
        (tmp_path / "Gemfile").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "ruby"

    def test_detect_php(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-002-PHP | PHP project detection (composer.json)."""
        (tmp_path / "composer.json").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "php"

    def test_detect_java(self, tmp_path: Path):
        (tmp_path / "pom.xml").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "java"

    def test_detect_java_gradle(self, tmp_path: Path):
        (tmp_path / "build.gradle").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "java"

    def test_detect_rust(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-004-RUST | Rust project detection (Cargo.toml)."""
        (tmp_path / "Cargo.toml").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "rust"

    def test_detect_dart(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-005-DART | Dart project detection (pubspec.yaml)."""
        (tmp_path / "pubspec.yaml").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "dart"

    def test_detect_swift(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-006-SWIFT | Swift project detection (Package.swift)."""
        (tmp_path / "Package.swift").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "swift"

    def test_detect_kotlin(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-007-KOTLIN | Kotlin project detection (build.gradle.kts)."""
        (tmp_path / "build.gradle.kts").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "kotlin"

    def test_detect_csharp(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-008-CSHARP | C# project detection (*.csproj)."""
        (tmp_path / "MyApp.csproj").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "csharp"

    def test_detect_c(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-009-C | C project detection (Makefile + *.c)."""
        (tmp_path / "Makefile").touch()
        (tmp_path / "main.c").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "c"

    def test_detect_cpp(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-010-CPP | C++ project detection (CMakeLists.txt + *.cpp)."""
        (tmp_path / "CMakeLists.txt").touch()
        (tmp_path / "main.cpp").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "cpp"

    def test_detect_shell(self, tmp_path: Path):
        """# REMOVED_ORPHAN_TEST:LDE-011-SHELL | Shell project detection (*.sh files)."""
        (tmp_path / "script.sh").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "shell"


class TestBuildToolDetection:
    """Test build tool auto-detection for 5 languages."""

    def test_maven_detection(self, tmp_path: Path):
        (tmp_path / "pom.xml").touch()
        detector = LanguageDetector()
        assert detector.detect_build_tool(tmp_path, language="java") == "maven"

    def test_gradle_detection(self, tmp_path: Path):
        (tmp_path / "build.gradle").touch()
        detector = LanguageDetector()
        assert detector.detect_build_tool(tmp_path, language="java") == "gradle"

    def test_cmake_detection(self, tmp_path: Path):
        (tmp_path / "CMakeLists.txt").touch()
        detector = LanguageDetector()
        assert detector.detect_build_tool(tmp_path) == "cmake"

    def test_spm_detection(self, tmp_path: Path):
        (tmp_path / "Package.swift").touch()
        detector = LanguageDetector()
        assert detector.detect_build_tool(tmp_path) == "spm"

    def test_dotnet_detection(self, tmp_path: Path):
        (tmp_path / "MyApp.csproj").touch()
        detector = LanguageDetector()
        assert detector.detect_build_tool(tmp_path) == "dotnet"


class TestPackageManagerDetection:
    """Test package manager auto-detection."""

    def test_bundle_detection(self, tmp_path: Path):
        (tmp_path / "Gemfile").touch()
        detector = LanguageDetector()
        assert detector.detect_package_manager(tmp_path) == "bundle"

    def test_composer_detection(self, tmp_path: Path):
        (tmp_path / "composer.json").touch()
        detector = LanguageDetector()
        assert detector.detect_package_manager(tmp_path) == "composer"

    def test_cargo_detection(self, tmp_path: Path):
        (tmp_path / "Cargo.toml").touch()
        detector = LanguageDetector()
        assert detector.detect_package_manager(tmp_path) == "cargo"


class TestPriorityConflicts:
    """Test priority rules when multiple language files coexist."""

    def test_kotlin_over_java(self, tmp_path: Path):
        (tmp_path / "build.gradle.kts").touch()
        (tmp_path / "pom.xml").touch()
        detector = LanguageDetector()
        # Kotlin should be detected first due to higher priority
        assert detector.detect(tmp_path) == "kotlin"

    def test_cpp_over_c(self, tmp_path: Path):
        (tmp_path / "CMakeLists.txt").touch()
        (tmp_path / "main.cpp").touch()
        (tmp_path / "utils.c").touch()
        detector = LanguageDetector()
        # C++ should be detected first
        assert detector.detect(tmp_path) == "cpp"

    def test_rust_highest_priority(self, tmp_path: Path):
        (tmp_path / "Cargo.toml").touch()
        (tmp_path / "pyproject.toml").touch()
        detector = LanguageDetector()
        # Rust should be detected first (highest priority in SPEC)
        assert detector.detect(tmp_path) == "rust"

    def test_ruby_over_python(self, tmp_path: Path):
        (tmp_path / "Gemfile").touch()
        (tmp_path / "requirements.txt").touch()
        detector = LanguageDetector()
        # Ruby should be detected before Python
        assert detector.detect(tmp_path) == "ruby"


class TestErrorHandling:
    """Test error handling for edge cases."""

    def test_unknown_language_returns_none(self, tmp_path: Path):
        detector = LanguageDetector()
        assert detector.detect(tmp_path) is None

    def test_workflow_template_for_unsupported_language(self):
        detector = LanguageDetector()
        # No workflow mapping available - should return None
        result = detector.get_workflow_template_path("cobol")
        assert result is None

    def test_no_build_tool_detected(self, tmp_path: Path):
        (tmp_path / "script.py").touch()
        detector = LanguageDetector()
        assert detector.detect_build_tool(tmp_path) is None


class TestBackwardCompatibility:
    """Test that existing 4 language detections still work (regression tests)."""

    def test_existing_python_support(self, tmp_path: Path):
        (tmp_path / "pyproject.toml").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "python"

    def test_existing_javascript_support(self, tmp_path: Path):
        (tmp_path / "package.json").touch()
        detector = LanguageDetector()
        # Note: package.json could be detected as typescript if tsconfig exists
        # This test ensures backward compatibility
        assert detector.detect(tmp_path) in ["javascript", "typescript"]

    def test_existing_typescript_support(self, tmp_path: Path):
        (tmp_path / "tsconfig.json").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "typescript"

    def test_existing_go_support(self, tmp_path: Path):
        (tmp_path / "go.mod").touch()
        detector = LanguageDetector()
        assert detector.detect(tmp_path) == "go"


class TestIntegration:
    """Integration tests for workflow template path retrieval."""

    def test_get_workflow_template_for_ruby(self):
        detector = LanguageDetector()
        path = detector.get_workflow_template_path("ruby")
        assert path == ".github/workflows/ruby-tag-validation.yml"

    def test_get_workflow_template_for_java(self):
        detector = LanguageDetector()
        path = detector.get_workflow_template_path("java")
        assert path == ".github/workflows/java-tag-validation.yml"

    def test_supported_languages_count(self):
        detector = LanguageDetector()
        supported = detector.get_supported_languages_for_workflows()
        assert len(supported) == 15
        assert "ruby" in supported
        assert "php" in supported
        assert "shell" in supported


# Summary of test coverage:
# ✅ 11 language detection tests (Ruby, PHP, Java, Rust, Dart, Swift, Kotlin, C#, C, C++, Shell)
# ✅ 5 build tool detection tests (Maven, Gradle, CMake, SPM, dotnet)
# ✅ 3 package manager detection tests (bundle, composer, cargo)
# ✅ 4 priority conflict tests (Kotlin vs Java, C++ vs C, Rust highest, Ruby vs Python)
# ✅ 3 error handling tests (unknown language, unsupported workflow, no build tool)
# ✅ 4 backward compatibility tests (Python, JS, TS, Go)
# ✅ 3 integration tests (Ruby workflow path, Java workflow path, 15 languages count)
# Total: 33 tests (exceeds SPEC requirement of 30 tests)
