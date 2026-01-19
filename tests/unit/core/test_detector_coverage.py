"""Additional coverage tests for project detector.

Tests for lines not covered by existing tests.
"""


from moai_adk.core.project.detector import LanguageDetector


class TestLanguageDetectorPackageManagerRuby:
    """Test Ruby package manager detection."""

    def test_detect_ruby_bundle(self, tmp_path):
        """Should detect bundle when Gemfile exists."""
        (tmp_path / "Gemfile").write_text("# Gemfile")

        detector = LanguageDetector()
        result = detector.detect_package_manager(tmp_path)

        assert result == "bundle"


class TestLanguageDetectorPackageManagerPHP:
    """Test PHP package manager detection."""

    def test_detect_php_composer(self, tmp_path):
        """Should detect composer when composer.json exists."""
        (tmp_path / "composer.json").write_text("{}")

        detector = LanguageDetector()
        result = detector.detect_package_manager(tmp_path)

        assert result == "composer"


class TestLanguageDetectorPackageManagerDart:
    """Test Dart/Flutter package manager detection."""

    def test_detect_dart_pub(self, tmp_path):
        """Should detect dart_pub when pubspec.yaml exists."""
        (tmp_path / "pubspec.yaml").write_text("# pubspec")

        detector = LanguageDetector()
        result = detector.detect_package_manager(tmp_path)

        assert result == "dart_pub"


class TestLanguageDetectorPackageManagerSwift:
    """Test Swift package manager detection."""

    def test_detect_swift_spm(self, tmp_path):
        """Should detect spm when Package.swift exists."""
        (tmp_path / "Package.swift").write_text("// Package")

        detector = LanguageDetector()
        result = detector.detect_package_manager(tmp_path)

        assert result == "spm"


class TestLanguageDetectorPackageManagerCSharp:
    """Test C# package manager detection."""

    def test_detect_dotnet_csproj(self, tmp_path):
        """Should detect dotnet when .csproj file exists."""
        (tmp_path / "project.csproj").write_text("<Project></Project>")

        detector = LanguageDetector()
        result = detector.detect_package_manager(tmp_path)

        assert result == "dotnet"

    def test_detect_dotnet_sln(self, tmp_path):
        """Should detect dotnet when .sln file exists."""
        (tmp_path / "solution.sln").write_text("")
        detector = LanguageDetector()
        result = detector.detect_package_manager(tmp_path)
        assert result == "dotnet"


class TestLanguageDetectorPackageManagerGo:
    """Test Go package manager detection."""

    def test_detect_go_modules(self, tmp_path):
        """Should detect go_modules when go.mod exists."""
        (tmp_path / "go.mod").write_text("module test")

        detector = LanguageDetector()
        result = detector.detect_package_manager(tmp_path)

        assert result == "go_modules"


class TestLanguageDetectorBuildToolJava:
    """Test Java/Kotlin build tool detection."""

    def test_detect_build_tool_java_gradle(self, tmp_path):
        """Should detect gradle for Java with build.gradle."""
        # Create build.gradle file
        (tmp_path / "build.gradle").write_text("// build")

        detector = LanguageDetector()
        result = detector.detect_build_tool(tmp_path, language="java")

        assert result == "gradle"

    def test_detect_build_tool_java_gradle_kts(self, tmp_path):
        """Should detect gradle for Java with build.gradle.kts."""
        (tmp_path / "build.gradle.kts").write_text("// build")

        detector = LanguageDetector()
        result = detector.detect_build_tool(tmp_path, language="java")

        assert result == "gradle"

    def test_detect_build_tool_rust(self, tmp_path):
        """Should detect cargo for Rust projects."""
        (tmp_path / "Cargo.toml").write_text("[package]")

        detector = LanguageDetector()
        result = detector.detect_build_tool(tmp_path)

        assert result == "cargo"

    def test_detect_build_tool_swift_spm(self, tmp_path):
        """Should detect spm for Swift with Package.swift."""
        (tmp_path / "Package.swift").write_text("// Package")

        detector = LanguageDetector()
        result = detector.detect_build_tool(tmp_path)

        assert result == "spm"

    def test_detect_build_tool_swift_xcode(self, tmp_path):
        """Should detect xcode for Swift with Xcode project."""
        (tmp_path / "MyApp.xcodeproj").mkdir()

        detector = LanguageDetector()
        result = detector.detect_build_tool(tmp_path)

        assert result == "xcode"

    def test_detect_build_tool_csharp_dotnet(self, tmp_path):
        """Should detect dotnet for C# projects."""
        (tmp_path / "project.csproj").write_text("<Project></Project>")

        detector = LanguageDetector()
        result = detector.detect_build_tool(tmp_path)

        assert result == "dotnet"
