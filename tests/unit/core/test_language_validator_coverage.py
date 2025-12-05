"""
Comprehensive Language Validator Coverage Tests

This test file provides comprehensive coverage for language_validator.py,
including edge cases, error conditions, and all code paths.
"""

import pytest
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock

from moai_adk.core.language_validator import (
    LanguageValidator,
    get_all_supported_languages,
    get_language_by_file_extension,
    is_code_directory,
    get_exclude_patterns,
    LANGUAGE_DIRECTORY_MAP,
)


class TestModuleLevelFunctions:
    """Test module-level functions in language_validator."""

    def test_get_all_supported_languages(self):
        """Test get_all_supported_languages function."""
        languages = get_all_supported_languages()
        assert isinstance(languages, set)
        assert "python" in languages
        assert "javascript" in languages
        assert "typescript" in languages
        assert "java" in languages
        assert "go" in languages
        assert "rust" in languages
        assert "cpp" in languages
        assert "c" in languages

    def test_get_language_by_file_extension_with_string_extension(self):
        """Test get_language_by_file_extension with string extensions."""
        assert get_language_by_file_extension(".py") == "python"
        assert get_language_by_file_extension(".js") == "javascript"
        assert get_language_by_file_extension(".ts") == "typescript"
        assert get_language_by_file_extension(".java") == "java"
        assert get_language_by_file_extension(".go") == "go"
        assert get_language_by_file_extension(".rs") == "rust"
        assert get_language_by_file_extension(".cpp") == "cpp"
        assert get_language_by_file_extension(".c") == "c"

    def test_get_language_by_file_extension_with_path_object(self):
        """Test get_language_by_file_extension with Path objects."""
        assert get_language_by_file_extension(Path("test.py")) == "python"
        assert get_language_by_file_extension(Path("app.js")) == "javascript"
        assert get_language_by_file_extension(Path("component.ts")) == "typescript"

    def test_get_language_by_file_extension_with_filename_string(self):
        """Test get_language_by_file_extension with filename strings."""
        assert get_language_by_file_extension("test.py") == "python"
        assert get_language_by_file_extension("app.js") == "javascript"
        assert get_language_by_file_extension("file.pyw") == "python"
        assert get_language_by_file_extension("file.pyx") == "python"

    def test_get_language_by_file_extension_no_dot(self):
        """Test get_language_by_file_extension with filename without leading dot."""
        # Without a dot prefix, the function extracts and prepends .
        assert get_language_by_file_extension("py") is None  # Returns None for bare extension
        assert get_language_by_file_extension("js") is None

    def test_get_language_by_file_extension_case_insensitive(self):
        """Test get_language_by_file_extension is case insensitive."""
        assert get_language_by_file_extension(".PY") == "python"
        assert get_language_by_file_extension(".JS") == "javascript"
        assert get_language_by_file_extension(".TS") == "typescript"

    def test_get_language_by_file_extension_unknown(self):
        """Test get_language_by_file_extension with unknown extension."""
        assert get_language_by_file_extension(".unknown") is None
        assert get_language_by_file_extension(".xyz") is None
        assert get_language_by_file_extension("file.unknown") is None

    def test_get_language_by_file_extension_multiple_dots(self):
        """Test get_language_by_file_extension with multiple dots in filename."""
        assert get_language_by_file_extension("test.spec.py") == "python"
        assert get_language_by_file_extension("app.config.js") == "javascript"

    def test_get_language_by_file_extension_empty_string(self):
        """Test get_language_by_file_extension with empty string."""
        result = get_language_by_file_extension("")
        assert result is None

    def test_is_code_directory_positive_cases(self):
        """Test is_code_directory with code directories."""
        assert is_code_directory("src/main.py") is True
        assert is_code_directory("lib/utils.js") is True
        assert is_code_directory("app/views.py") is True
        assert is_code_directory("components/Button.tsx") is True
        assert is_code_directory("modules/auth.go") is True
        assert is_code_directory("packages/core/index.ts") is True

    def test_is_code_directory_negative_cases(self):
        """Test is_code_directory with non-code directories."""
        assert is_code_directory("docs/README.md") is False
        assert is_code_directory("config/settings.yaml") is False
        assert is_code_directory("tests/test_file.py") is False
        assert is_code_directory("data/dataset.json") is False

    def test_is_code_directory_nested(self):
        """Test is_code_directory with nested paths."""
        assert is_code_directory("project/src/module/file.py") is True
        assert is_code_directory("project/lib/utils/helpers.js") is True
        assert is_code_directory("repo/app/models/user.py") is True

    def test_get_exclude_patterns(self):
        """Test get_exclude_patterns function."""
        patterns = get_exclude_patterns()
        assert isinstance(patterns, list)
        assert "*.pyc" in patterns
        assert "*.pyo" in patterns
        assert "__pycache__" in patterns
        assert ".git" in patterns
        assert "node_modules" in patterns
        assert ".venv" in patterns

    def test_language_directory_map_constant(self):
        """Test LANGUAGE_DIRECTORY_MAP constant."""
        assert "python" in LANGUAGE_DIRECTORY_MAP
        assert "javascript" in LANGUAGE_DIRECTORY_MAP
        assert "typescript" in LANGUAGE_DIRECTORY_MAP
        assert "src" in LANGUAGE_DIRECTORY_MAP["python"]
        assert "tests" in LANGUAGE_DIRECTORY_MAP["python"]
        assert "examples" in LANGUAGE_DIRECTORY_MAP["python"]


class TestLanguageValidatorInitialization:
    """Test LanguageValidator initialization."""

    def test_init_default_languages(self):
        """Test initialization with default languages."""
        validator = LanguageValidator()
        assert validator.auto_validate is True
        assert len(validator.supported_languages) > 0
        assert "python" in validator.supported_languages
        assert isinstance(validator.supported_languages, set)

    def test_init_custom_languages(self):
        """Test initialization with custom languages."""
        langs = ["python", "javascript", "typescript"]
        validator = LanguageValidator(supported_languages=langs)
        assert validator.supported_languages == {"python", "javascript", "typescript"}

    def test_init_custom_languages_case_normalization(self):
        """Test initialization normalizes languages to lowercase."""
        langs = ["Python", "JAVASCRIPT", "TypeScript"]
        validator = LanguageValidator(supported_languages=langs)
        assert "python" in validator.supported_languages
        assert "javascript" in validator.supported_languages
        assert "typescript" in validator.supported_languages

    def test_init_auto_validate_false(self):
        """Test initialization with auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)
        assert validator.auto_validate is False

    def test_init_analysis_cache_created(self):
        """Test initialization creates analysis cache."""
        validator = LanguageValidator()
        assert hasattr(validator, "_analysis_cache")
        assert "last_analysis_files" in validator._analysis_cache
        assert "detected_extensions" in validator._analysis_cache
        assert "supported_languages_found" in validator._analysis_cache

    def test_init_extension_map_available(self):
        """Test initialization provides EXTENSION_MAP."""
        validator = LanguageValidator()
        assert hasattr(validator, "EXTENSION_MAP")
        assert isinstance(validator.EXTENSION_MAP, dict)
        assert "python" in validator.EXTENSION_MAP
        assert ".py" in validator.EXTENSION_MAP["python"]


class TestValidateAndNormalizeInput:
    """Test _validate_and_normalize_input method."""

    def test_validate_language_valid(self):
        """Test validation of valid language input."""
        validator = LanguageValidator()
        result = validator._validate_and_normalize_input("python", "language")
        assert result == "python"

    def test_validate_language_case_normalization(self):
        """Test language normalization to lowercase."""
        validator = LanguageValidator()
        assert validator._validate_and_normalize_input("PYTHON", "language") == "python"
        assert validator._validate_and_normalize_input("Python", "language") == "python"

    def test_validate_language_whitespace_stripping(self):
        """Test language whitespace stripping."""
        validator = LanguageValidator()
        assert validator._validate_and_normalize_input("  python  ", "language") == "python"

    def test_validate_language_non_string(self):
        """Test validation of non-string language."""
        validator = LanguageValidator()
        assert validator._validate_and_normalize_input(123, "language") is None
        assert validator._validate_and_normalize_input([], "language") is None
        assert validator._validate_and_normalize_input({}, "language") is None

    def test_validate_language_empty_string(self):
        """Test validation of empty language."""
        validator = LanguageValidator()
        result = validator._validate_and_normalize_input("", "language")
        assert result is None

    def test_validate_file_path_string(self):
        """Test validation of file path as string."""
        validator = LanguageValidator()
        result = validator._validate_and_normalize_input("test.py", "file_path")
        assert isinstance(result, Path)

    def test_validate_file_path_path_object(self):
        """Test validation of file path as Path object."""
        validator = LanguageValidator()
        path = Path("test.py")
        result = validator._validate_and_normalize_input(path, "file_path")
        assert isinstance(result, Path)

    def test_validate_file_path_invalid(self):
        """Test validation of invalid file path."""
        validator = LanguageValidator()
        assert validator._validate_and_normalize_input(123, "file_path") is None
        assert validator._validate_and_normalize_input([], "file_path") is None

    def test_validate_list_valid(self):
        """Test validation of valid list."""
        validator = LanguageValidator()
        test_list = ["python", "javascript"]
        result = validator._validate_and_normalize_input(test_list, "list")
        assert result == test_list

    def test_validate_list_invalid(self):
        """Test validation of invalid list."""
        validator = LanguageValidator()
        assert validator._validate_and_normalize_input("python", "list") is None
        assert validator._validate_and_normalize_input(123, "list") is None

    def test_validate_unsupported_type(self):
        """Test validation of unsupported input type."""
        validator = LanguageValidator()
        result = validator._validate_and_normalize_input("test", "unsupported_type")
        assert result is None

    def test_validate_none_input_language(self):
        """Test validation of None for language."""
        validator = LanguageValidator()
        result = validator._validate_and_normalize_input(None, "language")
        assert result is None

    def test_validate_none_input_non_language(self):
        """Test validation of None for non-language types."""
        validator = LanguageValidator()
        result = validator._validate_and_normalize_input(None, "file_path")
        assert result is None


class TestValidateLanguage:
    """Test validate_language method."""

    def test_validate_supported_language(self):
        """Test validation of supported language."""
        validator = LanguageValidator(supported_languages=["python", "javascript"])
        assert validator.validate_language("python") is True
        assert validator.validate_language("javascript") is True

    def test_validate_unsupported_language(self):
        """Test validation of unsupported language."""
        validator = LanguageValidator(supported_languages=["python", "javascript"])
        assert validator.validate_language("rust") is False
        assert validator.validate_language("go") is False

    def test_validate_language_case_insensitive(self):
        """Test language validation is case insensitive."""
        validator = LanguageValidator(supported_languages=["python"])
        assert validator.validate_language("PYTHON") is True
        assert validator.validate_language("Python") is True

    def test_validate_language_with_whitespace(self):
        """Test language validation with whitespace."""
        validator = LanguageValidator(supported_languages=["python"])
        assert validator.validate_language("  python  ") is True

    def test_validate_language_empty_string(self):
        """Test validation of empty language."""
        validator = LanguageValidator(supported_languages=["python"])
        assert validator.validate_language("") is False

    def test_validate_language_none(self):
        """Test validation of None language."""
        validator = LanguageValidator(supported_languages=["python"])
        assert validator.validate_language(None) is False

    def test_validate_language_auto_validate_false(self):
        """Test validation with auto_validate=False."""
        validator = LanguageValidator(supported_languages=["python"], auto_validate=False)
        assert validator.validate_language("python") is True
        # When auto_validate is False, invalid inputs still use normalize_language_code


class TestDetectLanguageFromExtension:
    """Test detect_language_from_extension method."""

    def test_detect_python_extensions(self):
        """Test detection of Python extensions."""
        validator = LanguageValidator()
        assert validator.detect_language_from_extension("test.py") == "python"
        assert validator.detect_language_from_extension("test.pyw") == "python"
        assert validator.detect_language_from_extension("test.pyx") == "python"
        assert validator.detect_language_from_extension("test.pxd") == "python"

    def test_detect_javascript_extensions(self):
        """Test detection of JavaScript extensions."""
        validator = LanguageValidator()
        assert validator.detect_language_from_extension("app.js") == "javascript"
        assert validator.detect_language_from_extension("component.jsx") == "javascript"
        assert validator.detect_language_from_extension("module.mjs") == "javascript"

    def test_detect_typescript_extensions(self):
        """Test detection of TypeScript extensions."""
        validator = LanguageValidator()
        assert validator.detect_language_from_extension("app.ts") == "typescript"
        assert validator.detect_language_from_extension("component.tsx") == "typescript"
        assert validator.detect_language_from_extension("module.cts") == "typescript"
        assert validator.detect_language_from_extension("module.mts") == "typescript"

    def test_detect_other_extensions(self):
        """Test detection of other language extensions."""
        validator = LanguageValidator()
        assert validator.detect_language_from_extension("main.go") == "go"
        assert validator.detect_language_from_extension("lib.rs") == "rust"
        assert validator.detect_language_from_extension("Main.java") == "java"
        assert validator.detect_language_from_extension("file.cpp") == "cpp"
        assert validator.detect_language_from_extension("file.c") == "c"

    def test_detect_with_path_object(self):
        """Test detection with Path object."""
        validator = LanguageValidator()
        assert validator.detect_language_from_extension(Path("test.py")) == "python"
        assert validator.detect_language_from_extension(Path("app.js")) == "javascript"

    def test_detect_case_insensitive(self):
        """Test detection is case insensitive."""
        validator = LanguageValidator()
        assert validator.detect_language_from_extension("test.PY") == "python"
        assert validator.detect_language_from_extension("app.JS") == "javascript"

    def test_detect_unknown_extension(self):
        """Test detection of unknown extension."""
        validator = LanguageValidator()
        assert validator.detect_language_from_extension("file.unknown") is None
        assert validator.detect_language_from_extension("file.xyz") is None

    def test_detect_fallback_to_system_detection(self):
        """Test fallback to existing system detection."""
        validator = LanguageValidator()
        # Test that .py falls back to get_language_by_file_extension
        result = validator.detect_language_from_extension("test.py")
        assert result == "python"

    def test_detect_invalid_input_auto_validate_true(self):
        """Test detection with invalid input and auto_validate=True."""
        validator = LanguageValidator(auto_validate=True)
        assert validator.detect_language_from_extension(None) is None
        assert validator.detect_language_from_extension(123) is None

    def test_detect_invalid_input_auto_validate_false(self):
        """Test detection with invalid input and auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)
        assert validator.detect_language_from_extension(None) is None
        assert validator.detect_language_from_extension(123) is None


class TestGetExpectedDirectories:
    """Test get_expected_directories method."""

    def test_get_python_directories(self):
        """Test getting expected directories for Python."""
        validator = LanguageValidator()
        dirs = validator.get_expected_directories("python")
        assert isinstance(dirs, list)
        assert "src/" in dirs
        assert "tests/" in dirs
        assert "examples/" in dirs

    def test_get_javascript_directories(self):
        """Test getting expected directories for JavaScript."""
        validator = LanguageValidator()
        dirs = validator.get_expected_directories("javascript")
        assert isinstance(dirs, list)
        assert "src/" in dirs
        assert "lib/" in dirs

    def test_get_typescript_directories(self):
        """Test getting expected directories for TypeScript."""
        validator = LanguageValidator()
        dirs = validator.get_expected_directories("typescript")
        assert isinstance(dirs, list)
        assert "src/" in dirs

    def test_get_directories_unsupported_language(self):
        """Test getting directories for unsupported language."""
        validator = LanguageValidator()
        dirs = validator.get_expected_directories("unsupported")
        assert isinstance(dirs, list)
        # Should return default Python directories as fallback
        assert "src/" in dirs

    def test_get_directories_case_insensitive(self):
        """Test directory retrieval is case insensitive."""
        validator = LanguageValidator()
        dirs1 = validator.get_expected_directories("python")
        dirs2 = validator.get_expected_directories("PYTHON")
        assert dirs1 == dirs2

    def test_get_directories_with_whitespace(self):
        """Test directory retrieval with whitespace."""
        validator = LanguageValidator()
        dirs = validator.get_expected_directories("  python  ")
        assert isinstance(dirs, list)
        assert len(dirs) > 0

    def test_get_directories_invalid_input(self):
        """Test directory retrieval with invalid input."""
        validator = LanguageValidator()
        dirs = validator.get_expected_directories(None)
        assert dirs == []

    def test_get_directories_trailing_slashes(self):
        """Test that directories have trailing slashes."""
        validator = LanguageValidator()
        dirs = validator.get_expected_directories("python")
        for directory in dirs:
            assert directory.endswith("/")

    def test_get_directories_auto_validate_false(self):
        """Test directory retrieval with auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)
        dirs = validator.get_expected_directories("python")
        assert isinstance(dirs, list)


class TestDetectLanguageFromExtensionAutoValidateFalse:
    """Test detect_language_from_extension with auto_validate=False."""

    def test_detect_with_auto_validate_false_string(self):
        """Test detection with auto_validate=False using string path."""
        validator = LanguageValidator(auto_validate=False)
        # This path tests the else branch on line 212
        assert validator.detect_language_from_extension("test.py") == "python"

    def test_detect_with_auto_validate_false_path_object(self):
        """Test detection with auto_validate=False using Path object."""
        validator = LanguageValidator(auto_validate=False)
        # This path tests the elif branch on line 214
        assert validator.detect_language_from_extension(Path("test.ts")) == "typescript"


class TestGetFileExtensionAutoValidateFalse:
    """Test get_file_extensions with auto_validate=False."""

    def test_get_extensions_auto_validate_false(self):
        """Test getting extensions with auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)
        # This path tests line 269: normalize_language_code when auto_validate is False
        exts = validator.get_file_extensions("python")
        assert ".py" in exts

    def test_get_extensions_auto_validate_false_case_handling(self):
        """Test getting extensions with auto_validate=False preserves case handling."""
        validator = LanguageValidator(auto_validate=False)
        # Ensure normalize_language_code is called properly
        exts = validator.get_file_extensions("PYTHON")
        assert ".py" in exts


class TestGetFileExtensions:
    """Test get_file_extensions method."""

    def test_get_python_extensions(self):
        """Test getting Python extensions."""
        validator = LanguageValidator()
        exts = validator.get_file_extensions("python")
        assert ".py" in exts
        assert ".pyw" in exts
        assert ".pyx" in exts

    def test_get_javascript_extensions(self):
        """Test getting JavaScript extensions."""
        validator = LanguageValidator()
        exts = validator.get_file_extensions("javascript")
        assert ".js" in exts
        assert ".jsx" in exts
        assert ".mjs" in exts

    def test_get_typescript_extensions(self):
        """Test getting TypeScript extensions."""
        validator = LanguageValidator()
        exts = validator.get_file_extensions("typescript")
        assert ".ts" in exts
        assert ".tsx" in exts

    def test_get_extensions_unsupported_language(self):
        """Test getting extensions for unsupported language."""
        validator = LanguageValidator()
        exts = validator.get_file_extensions("unsupported")
        assert isinstance(exts, list)
        assert len(exts) == 0

    def test_get_extensions_case_insensitive(self):
        """Test extension retrieval is case insensitive."""
        validator = LanguageValidator()
        exts1 = validator.get_file_extensions("python")
        exts2 = validator.get_file_extensions("PYTHON")
        assert exts1 == exts2

    def test_get_extensions_invalid_input(self):
        """Test extension retrieval with invalid input."""
        validator = LanguageValidator()
        exts = validator.get_file_extensions(None)
        assert exts == []

    def test_get_extensions_returns_list(self):
        """Test that get_file_extensions returns a list."""
        validator = LanguageValidator()
        exts = validator.get_file_extensions("python")
        assert isinstance(exts, list)


class TestGetAllSupportedExtensions:
    """Test get_all_supported_extensions method."""

    def test_get_all_extensions(self):
        """Test getting all supported extensions."""
        validator = LanguageValidator()
        exts = validator.get_all_supported_extensions()
        assert isinstance(exts, set)
        assert ".py" in exts
        assert ".js" in exts
        assert ".ts" in exts

    def test_all_extensions_not_empty(self):
        """Test that all extensions set is not empty."""
        validator = LanguageValidator()
        exts = validator.get_all_supported_extensions()
        assert len(exts) > 0

    def test_all_extensions_includes_common_languages(self):
        """Test that all extensions include common languages."""
        validator = LanguageValidator()
        exts = validator.get_all_supported_extensions()
        # Check for extensions from different languages
        assert any(ext in exts for ext in [".py", ".pyw", ".pyx"])
        assert any(ext in exts for ext in [".js", ".jsx", ".mjs"])
        assert any(ext in exts for ext in [".ts", ".tsx"])


class TestDetectLanguageFromFilenameAutoValidateFalse:
    """Test detect_language_from_filename with auto_validate=False."""

    def test_detect_filename_auto_validate_false_with_invalid_input(self):
        """Test detection with auto_validate=False and invalid input (line 300-301)."""
        validator = LanguageValidator(auto_validate=False)
        # This tests the else branch starting at line 299 with invalid input
        assert validator.detect_language_from_filename(None) is None
        assert validator.detect_language_from_filename(123) is None
        assert validator.detect_language_from_filename("") is None

    def test_detect_filename_auto_validate_false_with_valid_input(self):
        """Test detection with auto_validate=False and valid input (line 302)."""
        validator = LanguageValidator(auto_validate=False)
        # This tests the Path creation at line 302
        result = validator.detect_language_from_filename("test.py")
        assert result == "python"


class TestValidateFileExtensionAutoValidateFalse:
    """Test validate_file_extension with auto_validate=False."""

    def test_validate_file_extension_auto_validate_false_invalid_language(self):
        """Test validation with auto_validate=False and invalid language (line 350-352)."""
        validator = LanguageValidator(auto_validate=False)
        # This tests both line 350 (when auto_validate is True) and line 352 (when False)
        # With auto_validate=False, it uses normalize_language_code
        result = validator.validate_file_extension("test.py", "python")
        assert result is True

    def test_validate_file_extension_auto_validate_false_whitespace(self):
        """Test validation with auto_validate=False handles whitespace in language."""
        validator = LanguageValidator(auto_validate=False)
        # normalize_language_code handles whitespace
        result = validator.validate_file_extension("test.py", "  python  ")
        assert result is True

    def test_validate_file_extension_auto_validate_true_none_language(self):
        """Test validation with auto_validate=True and None language returns True (line 343-345)."""
        validator = LanguageValidator(auto_validate=True)
        # When language is None, any file is valid (line 343-345)
        result = validator.validate_file_extension("test.py", None)
        assert result is True

    def test_validate_file_extension_auto_validate_true_non_string_language(self):
        """Test validation with auto_validate=True and non-string language returns False (line 350)."""
        validator = LanguageValidator(auto_validate=True)
        # _validate_and_normalize_input returns None for non-string languages (line 350)
        result = validator.validate_file_extension("test.py", [])
        assert result is False


class TestDetectLanguageFromFilename:
    """Test detect_language_from_filename method."""

    def test_detect_dockerfile(self):
        """Test detection of Dockerfile."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename("Dockerfile") == "dockerfile"
        assert validator.detect_language_from_filename("dockerfile") == "dockerfile"

    def test_detect_dockerfile_variants(self):
        """Test detection of Dockerfile variants."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename("Dockerfile.dev") == "dockerfile"
        assert validator.detect_language_from_filename("Dockerfile.prod") == "dockerfile"

    def test_detect_build_files(self):
        """Test detection of build configuration files."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename("Makefile") == "bash"
        assert validator.detect_language_from_filename("makefile") == "bash"
        assert validator.detect_language_from_filename("CMakeLists.txt") == "cpp"
        assert validator.detect_language_from_filename("pom.xml") == "java"

    def test_detect_language_config_files(self):
        """Test detection of language config files."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename("package.json") == "javascript"
        assert validator.detect_language_from_filename("pyproject.toml") == "python"
        assert validator.detect_language_from_filename("cargo.toml") == "rust"
        assert validator.detect_language_from_filename("go.mod") == "go"

    def test_detect_extension_based_fallback(self):
        """Test fallback to extension-based detection."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename("main.py") == "python"
        assert validator.detect_language_from_filename("app.js") == "javascript"

    def test_detect_case_insensitive(self):
        """Test filename detection is case insensitive."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename("DOCKERFILE") == "dockerfile"
        assert validator.detect_language_from_filename("PACKAGE.JSON") == "javascript"

    def test_detect_with_path(self):
        """Test detection with full path."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename("path/to/Dockerfile") == "dockerfile"

    def test_detect_invalid_input(self):
        """Test detection with invalid input."""
        validator = LanguageValidator()
        assert validator.detect_language_from_filename(None) is None
        assert validator.detect_language_from_filename(123) is None

    def test_detect_auto_validate_false(self):
        """Test detection with auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)
        assert validator.detect_language_from_filename("test.py") == "python"


class TestValidateFileExtension:
    """Test validate_file_extension method."""

    def test_validate_matching_extension(self):
        """Test validation of matching extension."""
        validator = LanguageValidator()
        assert validator.validate_file_extension("test.py", "python") is True
        assert validator.validate_file_extension("app.js", "javascript") is True

    def test_validate_mismatching_extension(self):
        """Test validation of mismatching extension."""
        validator = LanguageValidator()
        assert validator.validate_file_extension("test.js", "python") is False
        assert validator.validate_file_extension("test.py", "javascript") is False

    def test_validate_none_language(self):
        """Test validation with None language."""
        validator = LanguageValidator()
        # Any file is valid when no specific language is required
        assert validator.validate_file_extension("test.py", None) is True
        assert validator.validate_file_extension("test.js", None) is True

    def test_validate_case_insensitive(self):
        """Test extension validation is case insensitive."""
        validator = LanguageValidator()
        assert validator.validate_file_extension("test.PY", "python") is True
        assert validator.validate_file_extension("test.py", "PYTHON") is True

    def test_validate_with_path_object(self):
        """Test validation with Path object."""
        validator = LanguageValidator()
        assert validator.validate_file_extension(Path("test.py"), "python") is True

    def test_validate_invalid_language(self):
        """Test validation with invalid language."""
        validator = LanguageValidator()
        assert validator.validate_file_extension("test.py", "invalid_lang") is False


class TestGetSupportedLanguages:
    """Test get_supported_languages method."""

    def test_get_supported_languages_sorted(self):
        """Test that supported languages are sorted."""
        validator = LanguageValidator()
        langs = validator.get_supported_languages()
        assert langs == sorted(langs)

    def test_get_supported_languages_returns_list(self):
        """Test that supported languages returns a list."""
        validator = LanguageValidator()
        langs = validator.get_supported_languages()
        assert isinstance(langs, list)

    def test_get_supported_languages_contains_expected(self):
        """Test that supported languages contain expected languages."""
        validator = LanguageValidator()
        langs = validator.get_supported_languages()
        assert "python" in langs
        assert "javascript" in langs

    def test_get_supported_languages_custom(self):
        """Test get_supported_languages with custom languages."""
        validator = LanguageValidator(supported_languages=["go", "rust", "python"])
        langs = validator.get_supported_languages()
        assert langs == ["go", "python", "rust"]  # Sorted


class TestNormalizeLanguageCode:
    """Test normalize_language_code method."""

    def test_normalize_uppercase(self):
        """Test normalization of uppercase language code."""
        validator = LanguageValidator()
        assert validator.normalize_language_code("PYTHON") == "python"
        assert validator.normalize_language_code("JAVASCRIPT") == "javascript"

    def test_normalize_mixed_case(self):
        """Test normalization of mixed case language code."""
        validator = LanguageValidator()
        assert validator.normalize_language_code("Python") == "python"
        assert validator.normalize_language_code("JavaScript") == "javascript"

    def test_normalize_whitespace(self):
        """Test normalization of language code with whitespace."""
        validator = LanguageValidator()
        assert validator.normalize_language_code("  python  ") == "python"
        assert validator.normalize_language_code("\tjavascript\t") == "javascript"

    def test_normalize_empty_string(self):
        """Test normalization of empty language code."""
        validator = LanguageValidator()
        assert validator.normalize_language_code("") == ""

    def test_normalize_none(self):
        """Test normalization of None language code."""
        validator = LanguageValidator()
        assert validator.normalize_language_code(None) == ""

    def test_normalize_non_string(self):
        """Test normalization of non-string language code."""
        validator = LanguageValidator()
        assert validator.normalize_language_code(123) == ""
        assert validator.normalize_language_code([]) == ""


class TestValidateProjectConfiguration:
    """Test validate_project_configuration method."""

    def test_validate_valid_config(self):
        """Test validation of valid configuration."""
        validator = LanguageValidator(auto_validate=False)
        config = {
            "project": {
                "language": "python",
                "name": "test-project",
            }
        }
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is True
        assert len(issues) == 0

    def test_validate_missing_project_section(self):
        """Test validation with missing project section."""
        validator = LanguageValidator(auto_validate=False)
        config = {}
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is False
        assert any("Missing 'project' section" in issue for issue in issues)

    def test_validate_missing_language_field(self):
        """Test validation with missing language field."""
        validator = LanguageValidator(auto_validate=False)
        config = {"project": {"name": "test-project"}}
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is False
        assert any("Missing 'language' field" in issue for issue in issues)

    def test_validate_unsupported_language(self):
        """Test validation with unsupported language."""
        validator = LanguageValidator(auto_validate=False)
        config = {
            "project": {
                "language": "cobol",
                "name": "test-project",
            }
        }
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is False
        assert any("Unsupported language" in issue for issue in issues)

    def test_validate_missing_name_field(self):
        """Test validation with missing name field."""
        validator = LanguageValidator(auto_validate=False)
        config = {"project": {"language": "python"}}
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is False
        assert any("Missing 'name' field" in issue for issue in issues)

    def test_validate_empty_name(self):
        """Test validation with empty project name."""
        validator = LanguageValidator(auto_validate=False)
        config = {
            "project": {
                "language": "python",
                "name": "",
            }
        }
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is False
        assert any("Project name" in issue for issue in issues)

    def test_validate_whitespace_only_name(self):
        """Test validation with whitespace-only project name."""
        validator = LanguageValidator(auto_validate=False)
        config = {
            "project": {
                "language": "python",
                "name": "   ",
            }
        }
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is False
        assert any("empty or contain only whitespace" in issue for issue in issues)

    def test_validate_non_string_name(self):
        """Test validation with non-string project name."""
        validator = LanguageValidator(auto_validate=False)
        config = {
            "project": {
                "language": "python",
                "name": 123,
            }
        }
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is False
        assert any("must be a non-empty string" in issue for issue in issues)

    def test_validate_invalid_config_format(self):
        """Test validation with invalid config format."""
        validator = LanguageValidator()
        is_valid, issues = validator.validate_project_configuration(None)
        assert is_valid is False
        assert any("Invalid configuration format" in issue for issue in issues)

    def test_validate_auto_validate_false(self):
        """Test validation with auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)
        config = {
            "project": {
                "language": "python",
                "name": "test-project",
            }
        }
        is_valid, issues = validator.validate_project_configuration(config)
        assert isinstance(is_valid, bool)


class TestValidateProjectStructure:
    """Test validate_project_structure method."""

    def test_validate_valid_structure(self):
        """Test validation of valid project structure."""
        validator = LanguageValidator()
        structure = {
            "src/main.py": True,
            "tests/test_main.py": True,
            "README.md": False,
        }
        is_valid, issues = validator.validate_project_structure(structure, "python")
        assert isinstance(is_valid, bool)
        assert isinstance(issues, list)

    def test_validate_missing_expected_directories(self):
        """Test validation with missing expected directories."""
        validator = LanguageValidator()
        structure = {
            "docs/readme.md": False,
            "config/config.yaml": False,
        }
        is_valid, issues = validator.validate_project_structure(structure, "python")
        assert is_valid is False
        assert len(issues) > 0

    def test_validate_files_in_unexpected_locations(self):
        """Test validation with files in unexpected locations."""
        validator = LanguageValidator()
        structure = {
            "src/main.py": True,
            "tests/test_main.py": True,
            "backup/old_main.py": True,
        }
        is_valid, issues = validator.validate_project_structure(structure, "python")
        # backup is not in code_directory list, so it will generate issues
        assert len(issues) > 0

    def test_validate_multiple_languages(self):
        """Test validation with different languages."""
        validator = LanguageValidator()

        # Python structure
        py_structure = {"src/main.py": True}
        is_valid, issues = validator.validate_project_structure(py_structure, "python")
        assert isinstance(is_valid, bool)

        # JavaScript structure
        js_structure = {"src/main.js": True}
        is_valid, issues = validator.validate_project_structure(js_structure, "javascript")
        assert isinstance(is_valid, bool)

    def test_validate_empty_structure(self):
        """Test validation with empty structure."""
        validator = LanguageValidator()
        structure = {}
        is_valid, issues = validator.validate_project_structure(structure, "python")
        assert is_valid is False

    def test_validate_invalid_input(self):
        """Test validation with invalid input."""
        validator = LanguageValidator()
        is_valid, issues = validator.validate_project_structure(None, "python")
        assert is_valid is False
        assert any("Invalid input format" in issue for issue in issues)


class TestGetLanguageStatistics:
    """Test get_language_statistics method."""

    def test_statistics_single_language(self):
        """Test statistics for single language."""
        validator = LanguageValidator()
        files = ["main.py", "utils.py", "config.py"]
        stats = validator.get_language_statistics(files)
        assert isinstance(stats, dict)
        assert stats["python"] == 3

    def test_statistics_multiple_languages(self):
        """Test statistics for multiple languages."""
        validator = LanguageValidator()
        files = [
            "main.py",
            "utils.py",
            "app.js",
            "component.ts",
            "script.sh",
        ]
        stats = validator.get_language_statistics(files)
        assert stats["python"] == 2
        assert stats["javascript"] == 1
        assert stats["typescript"] == 1
        assert stats["bash"] == 1

    def test_statistics_empty_list(self):
        """Test statistics with empty file list."""
        validator = LanguageValidator()
        stats = validator.get_language_statistics([])
        assert isinstance(stats, dict)
        assert len(stats) == 0

    def test_statistics_unknown_extensions(self):
        """Test statistics with unknown extensions."""
        validator = LanguageValidator()
        files = ["readme.txt", "data.json", "image.png"]
        stats = validator.get_language_statistics(files)
        assert len(stats) == 1
        assert "json" in stats

    def test_statistics_updates_analysis_cache(self):
        """Test that statistics updates analysis cache."""
        validator = LanguageValidator()
        files = ["main.py", "utils.py", "app.js"]
        stats = validator.get_language_statistics(files)
        cache = validator.get_analysis_cache()
        assert cache["last_analysis_files"] == 3
        assert cache["supported_languages_found"] == 2

    def test_statistics_tracks_extensions(self):
        """Test that statistics tracks detected extensions."""
        validator = LanguageValidator()
        files = ["main.py", "app.js", "script.ts"]
        stats = validator.get_language_statistics(files)
        cache = validator.get_analysis_cache()
        assert ".py" in cache["detected_extensions"]
        assert ".js" in cache["detected_extensions"]
        assert ".ts" in cache["detected_extensions"]

    def test_statistics_with_path_objects(self):
        """Test statistics with Path objects."""
        validator = LanguageValidator()
        files = [Path("main.py"), Path("app.js")]
        stats = validator.get_language_statistics(files)
        assert stats["python"] == 1
        assert stats["javascript"] == 1

    def test_statistics_invalid_input(self):
        """Test statistics with invalid input."""
        validator = LanguageValidator()
        stats = validator.get_language_statistics(None)
        assert stats == {}

    def test_statistics_with_none_elements(self):
        """Test statistics with None elements in list."""
        validator = LanguageValidator()
        files = ["main.py", None, "app.js"]
        stats = validator.get_language_statistics(files)
        assert stats["python"] == 1
        assert stats["javascript"] == 1


class TestGetAnalysisCache:
    """Test get_analysis_cache and clear_analysis_cache methods."""

    def test_get_analysis_cache_initial_state(self):
        """Test initial analysis cache state."""
        validator = LanguageValidator()
        cache = validator.get_analysis_cache()
        assert isinstance(cache, dict)
        assert "last_analysis_files" in cache
        assert "detected_extensions" in cache
        assert "supported_languages_found" in cache

    def test_analysis_cache_is_copy(self):
        """Test that get_analysis_cache returns a copy."""
        validator = LanguageValidator()
        cache1 = validator.get_analysis_cache()
        cache1["test_key"] = "test_value"
        cache2 = validator.get_analysis_cache()
        assert "test_key" not in cache2

    def test_clear_analysis_cache(self):
        """Test clearing analysis cache."""
        validator = LanguageValidator()
        # Run statistics to populate cache
        validator.get_language_statistics(["main.py", "app.js"])
        cache = validator.get_analysis_cache()
        assert cache["last_analysis_files"] > 0

        # Clear cache
        validator.clear_analysis_cache()
        cache = validator.get_analysis_cache()
        assert cache["last_analysis_files"] == 0
        assert cache["detected_extensions"] == []
        assert cache["supported_languages_found"] == 0

    def test_cache_accumulation(self):
        """Test that cache tracks multiple analysis runs."""
        validator = LanguageValidator()
        validator.get_language_statistics(["main.py"])
        cache1 = validator.get_analysis_cache()
        files_count1 = cache1["last_analysis_files"]

        validator.get_language_statistics(["main.py", "app.js", "script.ts"])
        cache2 = validator.get_analysis_cache()
        files_count2 = cache2["last_analysis_files"]

        assert files_count2 > files_count1


class TestExtensionMapCoverage:
    """Test coverage of EXTENSION_MAP in LanguageValidator."""

    def test_extension_map_completeness(self):
        """Test that EXTENSION_MAP is comprehensive."""
        validator = LanguageValidator()
        extension_map = validator.EXTENSION_MAP

        # Check for major language families
        assert "python" in extension_map
        assert "javascript" in extension_map
        assert "typescript" in extension_map
        assert "java" in extension_map
        assert "go" in extension_map
        assert "rust" in extension_map
        assert "cpp" in extension_map
        assert "c" in extension_map

    def test_extension_map_language_coverage(self):
        """Test extension coverage for various languages."""
        validator = LanguageValidator()
        extension_map = validator.EXTENSION_MAP

        # Test language-specific extensions
        assert ".py" in extension_map["python"]
        assert ".js" in extension_map["javascript"]
        assert ".ts" in extension_map["typescript"]
        assert ".java" in extension_map["java"]
        assert ".go" in extension_map["go"]
        assert ".rs" in extension_map["rust"]


class TestEdgeCasesAndBoundaries:
    """Test edge cases and boundary conditions."""

    def test_very_long_language_code(self):
        """Test validation of very long language code."""
        validator = LanguageValidator(supported_languages=["python"])
        long_code = "a" * 1000
        assert validator.validate_language(long_code) is False

    def test_special_characters_in_language(self):
        """Test validation of language with special characters."""
        validator = LanguageValidator()
        assert validator.validate_language("python!@#$") is False

    def test_unicode_language_code(self):
        """Test validation of unicode language code."""
        validator = LanguageValidator()
        result = validator.validate_language("pythön")
        assert isinstance(result, bool)

    def test_very_long_file_path(self):
        """Test extension detection with very long file path."""
        validator = LanguageValidator()
        long_path = "a" * 1000 + ".py"
        result = validator.detect_language_from_extension(long_path)
        assert result == "python"

    def test_many_files_statistics(self):
        """Test statistics with many files."""
        validator = LanguageValidator()
        files = ["main.py"] * 1000 + ["app.js"] * 500
        stats = validator.get_language_statistics(files)
        assert stats["python"] == 1000
        assert stats["javascript"] == 500

    def test_large_project_structure(self):
        """Test project structure validation with large structure."""
        validator = LanguageValidator()
        structure = {f"src/module{i}.py": True for i in range(100)}
        structure.update({f"tests/test{i}.py": True for i in range(50)})
        is_valid, issues = validator.validate_project_structure(structure, "python")
        assert isinstance(is_valid, bool)

    def test_duplicate_files_in_statistics(self):
        """Test statistics with duplicate files."""
        validator = LanguageValidator()
        files = ["main.py", "main.py", "main.py"]
        stats = validator.get_language_statistics(files)
        assert stats["python"] == 3


class TestIntegrationScenarios:
    """Test integration scenarios combining multiple methods."""

    def test_full_project_validation_workflow(self):
        """Test full project validation workflow."""
        validator = LanguageValidator(auto_validate=False)

        # Check supported languages
        langs = validator.get_supported_languages()
        assert "python" in langs

        # Validate configuration
        config = {
            "project": {
                "language": "python",
                "name": "test-project",
            }
        }
        is_valid, issues = validator.validate_project_configuration(config)
        assert is_valid is True

        # Validate structure
        structure = {"src/main.py": True, "tests/test_main.py": True}
        is_valid, issues = validator.validate_project_structure(structure, "python")
        assert isinstance(is_valid, bool)

        # Get statistics
        files = ["main.py", "utils.py", "app.js"]
        stats = validator.get_language_statistics(files)
        assert "python" in stats

    def test_multi_language_project(self):
        """Test handling of multi-language projects."""
        validator = LanguageValidator()

        files = [
            "backend/main.py",
            "backend/utils.py",
            "frontend/index.js",
            "frontend/App.tsx",
            "scripts/build.sh",
        ]

        stats = validator.get_language_statistics(files)
        assert "python" in stats
        assert "javascript" in stats
        assert "typescript" in stats

    def test_custom_validator_subset(self):
        """Test validator with custom language subset."""
        validator = LanguageValidator(supported_languages=["python", "javascript"])

        # Should validate
        assert validator.validate_language("python") is True

        # Should not validate
        assert validator.validate_language("rust") is False

        # Should detect from extension
        assert validator.detect_language_from_extension("test.py") == "python"
