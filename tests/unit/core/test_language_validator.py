"""
Language Validator Tests

Test cases for language validation functionality.
"""

from pathlib import Path


class TestLanguageValidator:
    """Test suite for language validator functionality."""

    def test_language_validator_creation(self):
        """Test that language validator can be created successfully."""
        # This test should fail initially as the LanguageValidator class doesn't exist
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()
        assert validator is not None

    def test_language_validator_with_supported_languages(self):
        """Test language validator with supported languages."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator(supported_languages=["python", "javascript", "typescript"])
        assert validator.supported_languages == {"python", "javascript", "typescript"}

    def test_language_validator_supported_languages_default(self):
        """Test language validator default supported languages."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()
        # Should have default supported languages
        assert len(validator.supported_languages) > 0
        assert "python" in validator.supported_languages

    def test_validate_supported_language(self):
        """Test validation of supported languages."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator(supported_languages=["python", "javascript", "typescript"])

        # Test valid languages
        assert validator.validate_language("python") is True
        assert validator.validate_language("javascript") is True
        assert validator.validate_language("typescript") is True

        # Test invalid languages
        assert validator.validate_language("java") is False
        assert validator.validate_language("rust") is False
        assert validator.validate_language("") is False

    def test_validate_language_case_insensitive(self):
        """Test that language validation is case insensitive."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator(supported_languages=["Python", "JavaScript"])

        assert validator.validate_language("python") is True
        assert validator.validate_language("PYTHON") is True
        assert validator.validate_language("Python") is True

    def test_detect_language_from_extension(self):
        """Test language detection from file extension."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Test various file extensions
        assert validator.detect_language_from_extension("test.py") == "python"
        assert validator.detect_language_from_extension("app.js") == "javascript"
        assert validator.detect_language_from_extension("component.ts") == "typescript"
        assert validator.detect_language_from_extension("main.go") == "go"
        assert validator.detect_language_from_extension("lib.rs") == "rust"

        # Test unknown extension
        assert validator.detect_language_from_extension("unknown.xyz") is None

    def test_detect_language_from_extension_with_path_object(self):
        """Test language detection from Path object."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Test with Path objects
        assert validator.detect_language_from_extension(Path("test.py")) == "python"
        assert validator.detect_language_from_extension(Path("app.js")) == "javascript"

    def test_validate_project_structure(self):
        """Test project structure validation."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Mock project structure for Python
        project_structure = {
            "src/main.py": True,
            "tests/test_main.py": True,
            "README.md": True,
            "requirements.txt": True,
            "node_modules/package.json": False,  # Should be excluded
        }

        # Test Python project validation
        is_valid, issues = validator.validate_project_structure(project_structure, "python")
        assert isinstance(is_valid, bool)
        assert isinstance(issues, list)

    def test_get_expected_directories(self):
        """Test getting expected directories for a language."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Test Python expected directories
        python_dirs = validator.get_expected_directories("python")
        assert isinstance(python_dirs, list)
        assert len(python_dirs) > 0
        assert "src/" in python_dirs or "src/" in [d.lower() for d in python_dirs]

        # Test JavaScript expected directories
        js_dirs = validator.get_expected_directories("javascript")
        assert isinstance(js_dirs, list)
        assert len(js_dirs) > 0

    def test_get_expected_directories_unsupported_language(self):
        """Test getting expected directories for unsupported language."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Test unsupported language
        dirs = validator.get_expected_directories("unsupported_lang")
        assert isinstance(dirs, list)
        # Should return empty list or default directories
        assert len(dirs) >= 0

    def test_get_file_extensions(self):
        """Test getting file extensions for a language."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Test Python extensions
        python_extensions = validator.get_file_extensions("python")
        assert isinstance(python_extensions, list)
        assert ".py" in python_extensions

        # Test JavaScript extensions
        js_extensions = validator.get_file_extensions("javascript")
        assert isinstance(js_extensions, list)
        assert ".js" in js_extensions
        assert ".jsx" in js_extensions

    def test_validate_file_extension(self):
        """Test file extension validation."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Test valid Python files
        assert validator.validate_file_extension("test.py", "python") is True
        assert validator.validate_file_extension("main.py", "python") is True

        # Test invalid Python files
        assert validator.validate_file_extension("test.js", "python") is False
        assert validator.validate_file_extension("test.java", "python") is False

        # Test with no specific language (should accept any)
        assert validator.validate_file_extension("test.py", None) is True
        assert validator.validate_file_extension("test.js", None) is True

    def test_get_supported_languages(self):
        """Test getting list of supported languages."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()
        languages = validator.get_supported_languages()

        assert isinstance(languages, list)
        assert len(languages) > 0
        assert "python" in languages

    def test_normalize_language_code(self):
        """Test language code normalization."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Test various cases
        assert validator.normalize_language_code("PYTHON") == "python"
        assert validator.normalize_language_code("Python") == "python"
        assert validator.normalize_language_code("python") == "python"
        assert validator.normalize_language_code("  python  ") == "python"

        # Test empty string
        assert validator.normalize_language_code("") == ""

    def test_validate_project_configuration(self):
        """Test project configuration validation."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        # Valid Python configuration
        valid_config = {"project": {"language": "python", "name": "test-project"}, "directories": {"source": "src/"}}

        is_valid, issues = validator.validate_project_configuration(valid_config)
        assert isinstance(is_valid, bool)
        assert isinstance(issues, list)

        # Invalid configuration - unsupported language
        invalid_config = {"project": {"language": "cobol", "name": "test-project"}}

        is_valid, issues = validator.validate_project_configuration(invalid_config)
        assert isinstance(is_valid, bool)
        assert isinstance(issues, list)

    def test_language_statistics(self):
        """Test language statistics from file list."""
        # This test should fail initially
        from moai_adk.core.language_validator import LanguageValidator

        validator = LanguageValidator()

        files = ["src/main.py", "src/utils.py", "tests/test_main.py", "app.js", "styles.css", "README.md"]

        stats = validator.get_language_statistics(files)

        assert isinstance(stats, dict)
        assert "python" in stats
        assert "javascript" in stats
        assert stats["python"] == 3  # 3 Python files
        assert stats["javascript"] == 1  # 1 JavaScript file
