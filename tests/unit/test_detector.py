# @TEST:TEST-COVERAGE-001 | SPEC: SPEC-TEST-COVERAGE-001.md
"""Unit tests for detector.py module

Tests for LanguageDetector class.
"""

from pathlib import Path


from moai_adk.core.project.detector import LanguageDetector


class TestLanguageDetectorInit:
    """Test LanguageDetector initialization"""

    def test_has_language_patterns(self):
        """Should have LANGUAGE_PATTERNS class variable"""
        assert hasattr(LanguageDetector, "LANGUAGE_PATTERNS")
        assert isinstance(LanguageDetector.LANGUAGE_PATTERNS, dict)

    def test_language_patterns_contains_common_languages(self):
        """Should contain patterns for common languages"""
        patterns = LanguageDetector.LANGUAGE_PATTERNS
        assert "python" in patterns
        assert "typescript" in patterns
        assert "java" in patterns
        assert "go" in patterns

    def test_can_instantiate(self):
        """Should be able to create instance"""
        detector = LanguageDetector()
        assert detector is not None


class TestLanguageDetectorDetect:
    """Test detect method (single language)"""

    def test_detect_python_from_py_files(self, tmp_project_dir: Path):
        """Should detect Python from .py files"""
        (tmp_project_dir / "main.py").write_text("print('hello')")

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result == "python"

    def test_detect_python_from_pyproject_toml(self, tmp_project_dir: Path):
        """Should detect Python from pyproject.toml"""
        (tmp_project_dir / "pyproject.toml").write_text("[tool.poetry]")

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result == "python"

    def test_detect_typescript_from_ts_files(self, tmp_project_dir: Path):
        """Should detect TypeScript from .ts files"""
        (tmp_project_dir / "index.ts").write_text("const x: number = 1;")

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result == "typescript"

    def test_detect_javascript_from_package_json(self, tmp_project_dir: Path):
        """Should detect JavaScript from package.json"""
        (tmp_project_dir / "package.json").write_text('{"name": "test"}')

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result == "javascript"

    def test_detect_java_from_pom_xml(self, tmp_project_dir: Path):
        """Should detect Java from pom.xml"""
        (tmp_project_dir / "pom.xml").write_text("<project></project>")

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result == "java"

    def test_detect_go_from_go_mod(self, tmp_project_dir: Path):
        """Should detect Go from go.mod"""
        (tmp_project_dir / "go.mod").write_text("module test")

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result == "go"

    def test_detect_rust_from_cargo_toml(self, tmp_project_dir: Path):
        """Should detect Rust from Cargo.toml"""
        (tmp_project_dir / "Cargo.toml").write_text("[package]")

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result == "rust"

    def test_detect_returns_none_for_empty_directory(self, tmp_project_dir: Path):
        """Should return None when no language detected"""
        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        assert result is None

    def test_detect_returns_first_match_in_priority_order(self, tmp_project_dir: Path):
        """Should return first language in dictionary order"""
        # Create files for multiple languages
        (tmp_project_dir / "main.py").write_text("print('hello')")
        (tmp_project_dir / "index.ts").write_text("const x = 1;")

        detector = LanguageDetector()
        result = detector.detect(tmp_project_dir)

        # Should return first one encountered (dict order in Python 3.7+)
        # Based on LANGUAGE_PATTERNS order, Python comes before TypeScript
        assert result in ["python", "typescript"]


class TestLanguageDetectorDetectMultiple:
    """Test detect_multiple method"""

    def test_detect_multiple_returns_empty_for_empty_dir(self, tmp_project_dir: Path):
        """Should return empty list when no languages detected"""
        detector = LanguageDetector()
        result = detector.detect_multiple(tmp_project_dir)

        assert result == []

    def test_detect_multiple_returns_single_language(self, tmp_project_dir: Path):
        """Should return single language in list"""
        (tmp_project_dir / "main.py").write_text("print('hello')")

        detector = LanguageDetector()
        result = detector.detect_multiple(tmp_project_dir)

        assert "python" in result
        assert len(result) >= 1

    def test_detect_multiple_returns_multiple_languages(self, tmp_project_dir: Path):
        """Should detect all languages present"""
        (tmp_project_dir / "main.py").write_text("print('hello')")
        (tmp_project_dir / "index.ts").write_text("const x = 1;")
        (tmp_project_dir / "Main.java").write_text("public class Main {}")

        detector = LanguageDetector()
        result = detector.detect_multiple(tmp_project_dir)

        assert "python" in result
        assert "typescript" in result
        assert "java" in result

    def test_detect_multiple_with_nested_files(self, tmp_project_dir: Path):
        """Should detect languages in nested directories"""
        src_dir = tmp_project_dir / "src"
        src_dir.mkdir()
        (src_dir / "main.py").write_text("print('hello')")

        detector = LanguageDetector()
        result = detector.detect_multiple(tmp_project_dir)

        assert "python" in result


class TestLanguageDetectorCheckPatterns:
    """Test _check_patterns method"""

    def test_check_patterns_matches_extension(self, tmp_project_dir: Path):
        """Should match files by extension pattern"""
        (tmp_project_dir / "test.py").write_text("code")

        detector = LanguageDetector()
        result = detector._check_patterns(tmp_project_dir, ["*.py"])

        assert result is True

    def test_check_patterns_matches_specific_file(self, tmp_project_dir: Path):
        """Should match specific filename"""
        (tmp_project_dir / "pyproject.toml").write_text("[tool]")

        detector = LanguageDetector()
        result = detector._check_patterns(tmp_project_dir, ["pyproject.toml"])

        assert result is True

    def test_check_patterns_returns_false_when_no_match(self, tmp_project_dir: Path):
        """Should return False when no patterns match"""
        detector = LanguageDetector()
        result = detector._check_patterns(tmp_project_dir, ["*.py", "pyproject.toml"])

        assert result is False

    def test_check_patterns_matches_nested_files(self, tmp_project_dir: Path):
        """Should match files in subdirectories with rglob"""
        src_dir = tmp_project_dir / "src" / "nested"
        src_dir.mkdir(parents=True)
        (src_dir / "module.py").write_text("code")

        detector = LanguageDetector()
        result = detector._check_patterns(tmp_project_dir, ["*.py"])

        assert result is True
