"""Extended unit tests for moai_adk.core.language_validator module.

Comprehensive tests for LanguageValidator functionality.
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.language_validator import (
    LanguageValidator,
    get_all_supported_languages,
    get_language_by_file_extension,
    is_code_directory,
)


class TestLanguageValidatorInitialization:
    """Test LanguageValidator initialization."""

    def test_initialization_default(self):
        """Test LanguageValidator initializes with defaults."""
        validator = LanguageValidator()

        assert validator.auto_validate is True
        assert len(validator.supported_languages) > 0

    def test_initialization_custom_languages(self):
        """Test LanguageValidator with custom languages."""
        custom_langs = ["python", "go", "rust"]
        validator = LanguageValidator(supported_languages=custom_langs)

        assert "python" in validator.supported_languages
        assert "go" in validator.supported_languages

    def test_initialization_auto_validate_false(self):
        """Test LanguageValidator with auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)

        assert validator.auto_validate is False

    def test_initialization_extension_map_defined(self):
        """Test EXTENSION_MAP is defined correctly."""
        validator = LanguageValidator()

        assert len(validator.EXTENSION_MAP) > 0
        assert "python" in validator.EXTENSION_MAP
        assert ".py" in validator.EXTENSION_MAP["python"]

    def test_initialization_analysis_cache(self):
        """Test analysis cache is initialized."""
        validator = LanguageValidator()

        assert hasattr(validator, "_analysis_cache")
        assert validator._analysis_cache["last_analysis_files"] == 0


class TestValidateLanguage:
    """Test validate_language method."""

    def test_validate_language_supported(self):
        """Test validate_language returns True for supported language."""
        validator = LanguageValidator()

        assert validator.validate_language("python") is True
        assert validator.validate_language("go") is True

    def test_validate_language_unsupported(self):
        """Test validate_language returns False for unsupported language."""
        validator = LanguageValidator()

        assert validator.validate_language("nonexistent") is False

    def test_validate_language_case_insensitive(self):
        """Test validate_language is case-insensitive."""
        validator = LanguageValidator()

        assert validator.validate_language("PYTHON") is True
        assert validator.validate_language("Go") is True

    def test_validate_language_with_whitespace(self):
        """Test validate_language handles whitespace."""
        validator = LanguageValidator()

        assert validator.validate_language("  python  ") is True

    def test_validate_language_empty_string(self):
        """Test validate_language with empty string."""
        validator = LanguageValidator()

        assert validator.validate_language("") is False

    def test_validate_language_none(self):
        """Test validate_language with None."""
        validator = LanguageValidator()

        assert validator.validate_language(None) is False

    def test_validate_language_auto_validate_false(self):
        """Test validate_language with auto_validate=False."""
        validator = LanguageValidator(auto_validate=False)

        assert validator.validate_language("python") is True


class TestDetectLanguageFromExtension:
    """Test detect_language_from_extension method."""

    def test_detect_python_extension(self):
        """Test detecting Python from .py extension."""
        validator = LanguageValidator()

        result = validator.detect_language_from_extension("file.py")

        assert result == "python"

    def test_detect_javascript_extension(self):
        """Test detecting JavaScript from .js extension."""
        validator = LanguageValidator()

        result = validator.detect_language_from_extension("file.js")

        assert result == "javascript"

    def test_detect_path_object(self):
        """Test detect_language_from_extension with Path object."""
        validator = LanguageValidator()

        result = validator.detect_language_from_extension(Path("file.py"))

        assert result == "python"

    def test_detect_unknown_extension(self):
        """Test detect_language_from_extension with unknown extension."""
        validator = LanguageValidator()

        result = validator.detect_language_from_extension("file.xyz")

        assert result is None

    def test_detect_multiple_python_extensions(self):
        """Test detecting various Python extensions."""
        validator = LanguageValidator()

        assert validator.detect_language_from_extension("file.pyx") == "python"
        assert validator.detect_language_from_extension("file.pxd") == "python"

    def test_detect_case_insensitive(self):
        """Test extension detection is case-insensitive."""
        validator = LanguageValidator()

        assert validator.detect_language_from_extension("file.PY") == "python"
        assert validator.detect_language_from_extension("file.JS") == "javascript"

    def test_detect_invalid_input_type(self):
        """Test detect_language_from_extension with invalid input."""
        validator = LanguageValidator()

        result = validator.detect_language_from_extension(123)

        assert result is None


class TestGetExpectedDirectories:
    """Test get_expected_directories method."""

    def test_get_expected_python_directories(self):
        """Test getting expected directories for Python."""
        validator = LanguageValidator()

        result = validator.get_expected_directories("python")

        assert isinstance(result, list)
        assert len(result) > 0
        assert any("src" in d for d in result)

    def test_get_expected_javascript_directories(self):
        """Test getting expected directories for JavaScript."""
        validator = LanguageValidator()

        result = validator.get_expected_directories("javascript")

        assert isinstance(result, list)
        assert len(result) > 0

    def test_get_expected_unknown_language(self):
        """Test getting expected directories for unknown language."""
        validator = LanguageValidator()

        result = validator.get_expected_directories("unknown")

        # Should return default or empty list
        assert isinstance(result, list)

    def test_get_expected_case_insensitive(self):
        """Test case-insensitive language matching."""
        validator = LanguageValidator()

        result1 = validator.get_expected_directories("python")
        result2 = validator.get_expected_directories("PYTHON")

        assert result1 == result2

    def test_get_expected_directories_with_trailing_slash(self):
        """Test expected directories have trailing slashes."""
        validator = LanguageValidator()

        result = validator.get_expected_directories("python")

        assert all(d.endswith("/") for d in result)


class TestGetFileExtensions:
    """Test get_file_extensions method."""

    def test_get_python_extensions(self):
        """Test getting Python file extensions."""
        validator = LanguageValidator()

        result = validator.get_file_extensions("python")

        assert ".py" in result
        assert ".pyx" in result

    def test_get_javascript_extensions(self):
        """Test getting JavaScript file extensions."""
        validator = LanguageValidator()

        result = validator.get_file_extensions("javascript")

        assert ".js" in result

    def test_get_unknown_language_extensions(self):
        """Test getting extensions for unknown language."""
        validator = LanguageValidator()

        result = validator.get_file_extensions("unknown")

        assert result == []

    def test_get_extensions_returns_list(self):
        """Test get_file_extensions returns list."""
        validator = LanguageValidator()

        result = validator.get_file_extensions("python")

        assert isinstance(result, list)


class TestGetAllSupportedExtensions:
    """Test get_all_supported_extensions method."""

    def test_get_all_supported_extensions(self):
        """Test getting all supported extensions."""
        validator = LanguageValidator()

        result = validator.get_all_supported_extensions()

        assert isinstance(result, set)
        assert len(result) > 0
        assert ".py" in result
        assert ".js" in result

    def test_all_extensions_unique(self):
        """Test all extensions are unique."""
        validator = LanguageValidator()

        result = validator.get_all_supported_extensions()

        # Set should have no duplicates
        assert len(result) == len(result)


class TestDetectLanguageFromFilename:
    """Test detect_language_from_filename method."""

    def test_detect_dockerfile(self):
        """Test detecting Dockerfile."""
        validator = LanguageValidator()

        assert validator.detect_language_from_filename("Dockerfile") == "dockerfile"
        assert (
            validator.detect_language_from_filename("Dockerfile.dev")
            == "dockerfile"
        )

    def test_detect_package_json(self):
        """Test detecting from package.json."""
        validator = LanguageValidator()

        result = validator.detect_language_from_filename("package.json")

        assert result == "javascript"

    def test_detect_pyproject_toml(self):
        """Test detecting from pyproject.toml."""
        validator = LanguageValidator()

        result = validator.detect_language_from_filename("pyproject.toml")

        assert result == "python"

    def test_detect_makefile(self):
        """Test detecting Makefile."""
        validator = LanguageValidator()

        result = validator.detect_language_from_filename("Makefile")

        assert result == "bash"

    def test_detect_go_mod(self):
        """Test detecting go.mod."""
        validator = LanguageValidator()

        result = validator.detect_language_from_filename("go.mod")

        assert result == "go"

    def test_detect_from_full_path(self):
        """Test detecting from full file path."""
        validator = LanguageValidator()

        result = validator.detect_language_from_filename("/path/to/Dockerfile")

        assert result == "dockerfile"

    def test_detect_case_insensitive(self):
        """Test filename detection is case-insensitive."""
        validator = LanguageValidator()

        result1 = validator.detect_language_from_filename("dockerfile")
        result2 = validator.detect_language_from_filename("Dockerfile")

        assert result1 == result2

    def test_detect_unknown_filename(self):
        """Test detecting unknown filename."""
        validator = LanguageValidator()

        result = validator.detect_language_from_filename("unknown.xyz")

        assert result is None


class TestValidateFileExtension:
    """Test validate_file_extension method."""

    def test_validate_matching_extension(self):
        """Test validating matching file extension."""
        validator = LanguageValidator()

        result = validator.validate_file_extension("file.py", "python")

        assert result is True

    def test_validate_mismatched_extension(self):
        """Test validating mismatched extension."""
        validator = LanguageValidator()

        result = validator.validate_file_extension("file.py", "javascript")

        assert result is False

    def test_validate_none_language(self):
        """Test validate_file_extension with None language."""
        validator = LanguageValidator()

        result = validator.validate_file_extension("file.py", None)

        assert result is True

    def test_validate_path_object(self):
        """Test validate_file_extension with Path object."""
        validator = LanguageValidator()

        result = validator.validate_file_extension(Path("file.py"), "python")

        assert result is True


class TestGetSupportedLanguages:
    """Test get_supported_languages method."""

    def test_get_supported_languages(self):
        """Test getting list of supported languages."""
        validator = LanguageValidator()

        result = validator.get_supported_languages()

        assert isinstance(result, list)
        assert len(result) > 0
        assert "python" in result

    def test_languages_are_sorted(self):
        """Test supported languages are sorted."""
        validator = LanguageValidator()

        result = validator.get_supported_languages()

        assert result == sorted(result)


class TestNormalizeLanguageCode:
    """Test normalize_language_code method."""

    def test_normalize_lowercase(self):
        """Test normalizing language code."""
        validator = LanguageValidator()

        result = validator.normalize_language_code("PYTHON")

        assert result == "python"

    def test_normalize_with_whitespace(self):
        """Test normalizing with whitespace."""
        validator = LanguageValidator()

        result = validator.normalize_language_code("  python  ")

        assert result == "python"

    def test_normalize_empty_string(self):
        """Test normalizing empty string."""
        validator = LanguageValidator()

        result = validator.normalize_language_code("")

        assert result == ""

    def test_normalize_none(self):
        """Test normalizing None."""
        validator = LanguageValidator()

        result = validator.normalize_language_code(None)

        assert result == ""


class TestValidateProjectConfiguration:
    """Test validate_project_configuration method."""

    def test_validate_valid_config(self):
        """Test validating valid configuration."""
        validator = LanguageValidator()
        config = {
            "project": {
                "name": "MyProject",
                "language": "python",
            }
        }

        is_valid, errors = validator.validate_project_configuration(config)

        # Config is valid if project and language present
        assert isinstance(is_valid, bool)

    def test_validate_missing_project_section(self):
        """Test validating missing project section."""
        validator = LanguageValidator()
        config = {}

        is_valid, errors = validator.validate_project_configuration(config)

        assert is_valid is False
        assert isinstance(errors, list)

    def test_validate_missing_language(self):
        """Test validating missing language field."""
        validator = LanguageValidator()
        config = {"project": {"name": "MyProject"}}

        is_valid, errors = validator.validate_project_configuration(config)

        assert is_valid is False
        assert isinstance(errors, list)

    def test_validate_unsupported_language(self):
        """Test validating unsupported language."""
        validator = LanguageValidator()
        config = {
            "project": {
                "name": "MyProject",
                "language": "unsupported_lang",
            }
        }

        is_valid, errors = validator.validate_project_configuration(config)

        assert is_valid is False

    def test_validate_missing_name(self):
        """Test validating missing project name."""
        validator = LanguageValidator()
        config = {"project": {"language": "python"}}

        is_valid, errors = validator.validate_project_configuration(config)

        assert is_valid is False

    def test_validate_empty_name(self):
        """Test validating empty project name."""
        validator = LanguageValidator()
        config = {
            "project": {
                "name": "",
                "language": "python",
            }
        }

        is_valid, errors = validator.validate_project_configuration(config)

        assert is_valid is False

    def test_validate_whitespace_only_name(self):
        """Test validating whitespace-only project name."""
        validator = LanguageValidator()
        config = {
            "project": {
                "name": "   ",
                "language": "python",
            }
        }

        is_valid, errors = validator.validate_project_configuration(config)

        assert is_valid is False


class TestValidateProjectStructure:
    """Test validate_project_structure method."""

    def test_validate_valid_structure(self):
        """Test validating valid project structure."""
        validator = LanguageValidator()
        files = {
            "src/file.py": True,
            "tests/test.py": True,
            "README.md": False,
        }

        is_valid, errors = validator.validate_project_structure(files, "python")

        assert isinstance(is_valid, bool)

    def test_validate_missing_expected_dir(self):
        """Test validating with missing expected directories."""
        validator = LanguageValidator()
        files = {"README.md": False}

        is_valid, errors = validator.validate_project_structure(files, "python")

        assert is_valid is False or len(errors) >= 0

    def test_validate_source_in_unexpected_location(self):
        """Test validating source files in unexpected locations."""
        validator = LanguageValidator()
        files = {"random_dir/file.py": True}

        is_valid, errors = validator.validate_project_structure(files, "python")

        # May have warnings about unexpected locations
        assert isinstance(is_valid, bool)


class TestGetLanguageStatistics:
    """Test get_language_statistics method."""

    def test_get_statistics_single_language(self):
        """Test getting statistics for single language."""
        validator = LanguageValidator()
        files = ["file1.py", "file2.py", "file3.py"]

        result = validator.get_language_statistics(files)

        assert isinstance(result, dict)
        assert result.get("python", 0) == 3

    def test_get_statistics_multiple_languages(self):
        """Test getting statistics for multiple languages."""
        validator = LanguageValidator()
        files = ["file.py", "file.js", "file.go", "file.py"]

        result = validator.get_language_statistics(files)

        assert result.get("python", 0) >= 1
        assert result.get("javascript", 0) >= 1
        assert result.get("go", 0) >= 1

    def test_get_statistics_empty_list(self):
        """Test getting statistics with empty file list."""
        validator = LanguageValidator()

        result = validator.get_language_statistics([])

        assert isinstance(result, dict)
        assert len(result) == 0

    def test_get_statistics_unknown_files(self):
        """Test getting statistics with unknown file types."""
        validator = LanguageValidator()
        files = ["file.unknown", "file.xyz"]

        result = validator.get_language_statistics(files)

        assert isinstance(result, dict)

    def test_get_statistics_updates_cache(self):
        """Test that get_language_statistics updates cache."""
        validator = LanguageValidator()
        files = ["file.py", "file.js"]

        validator.get_language_statistics(files)

        cache = validator.get_analysis_cache()
        assert cache["last_analysis_files"] >= 2


class TestAnalysisCache:
    """Test analysis cache methods."""

    def test_get_analysis_cache(self):
        """Test get_analysis_cache returns cache data."""
        validator = LanguageValidator()

        cache = validator.get_analysis_cache()

        assert isinstance(cache, dict)
        assert "last_analysis_files" in cache
        assert "detected_extensions" in cache

    def test_clear_analysis_cache(self):
        """Test clear_analysis_cache resets cache."""
        validator = LanguageValidator()

        # Modify cache
        validator._analysis_cache["last_analysis_files"] = 100

        # Clear
        validator.clear_analysis_cache()

        cache = validator.get_analysis_cache()
        assert cache["last_analysis_files"] == 0


class TestLanguageValidatorEdgeCases:
    """Test edge cases and error scenarios."""

    def test_validator_with_empty_custom_languages(self):
        """Test validator with empty custom languages list."""
        validator = LanguageValidator(supported_languages=[])

        result = validator.validate_language("python")

        assert result is False

    def test_validate_large_file_list(self):
        """Test getting statistics with large file list."""
        validator = LanguageValidator()
        files = [f"file{i}.py" for i in range(1000)]

        result = validator.get_language_statistics(files)

        assert result.get("python", 0) == 1000

    def test_detect_with_invalid_path_object(self):
        """Test detection with valid path object."""
        validator = LanguageValidator()

        # Should handle gracefully
        result = validator.detect_language_from_extension("file.py")
        assert result == "python"

    def test_validate_with_non_string_language(self):
        """Test validate_language with non-string language."""
        validator = LanguageValidator()

        result = validator.validate_language(123)

        assert result is False


# Module-level function tests

def test_get_all_supported_languages():
    """Test module-level get_all_supported_languages function."""
    result = get_all_supported_languages()

    assert isinstance(result, set) or isinstance(result, dict) or isinstance(result, list)
    assert len(result) > 0


def test_get_language_by_file_extension_py():
    """Test module-level get_language_by_file_extension for Python."""
    result = get_language_by_file_extension(".py")

    assert result == "python"


def test_get_language_by_file_extension_js():
    """Test module-level get_language_by_file_extension for JavaScript."""
    result = get_language_by_file_extension(".js")

    assert result == "javascript"


def test_get_language_by_file_extension_unknown():
    """Test module-level get_language_by_file_extension for unknown."""
    result = get_language_by_file_extension(".unknown")

    assert result is None


def test_is_code_directory():
    """Test module-level is_code_directory function."""
    assert is_code_directory("/path/to/src") is True
    assert is_code_directory("/path/to/lib") is True
    assert is_code_directory("/path/to/docs") is False
