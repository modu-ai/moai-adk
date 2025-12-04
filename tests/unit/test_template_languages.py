"""Unit tests for template/languages.py module

Tests for language template mapping.
"""

from moai_adk.core.template.languages import LANGUAGE_TEMPLATES, get_language_template


class TestLanguageTemplates:
    """Test LANGUAGE_TEMPLATES constant"""

    def test_language_templates_is_dict(self):
        """LANGUAGE_TEMPLATES should be a dictionary"""
        assert isinstance(LANGUAGE_TEMPLATES, dict)

    def test_language_templates_contains_common_languages(self):
        """Should contain common programming languages"""
        expected_languages = ["python", "typescript", "java", "go", "rust"]
        for lang in expected_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_language_templates_has_20_languages(self):
        """Should have exactly 20 language mappings"""
        assert len(LANGUAGE_TEMPLATES) == 20

    def test_all_templates_have_j2_extension(self):
        """All template paths should end with .md.j2"""
        for template_path in LANGUAGE_TEMPLATES.values():
            assert template_path.endswith(".md.j2")

    def test_all_templates_are_in_tech_dir(self):
        """All template paths should be in .moai/project/tech/ directory"""
        for template_path in LANGUAGE_TEMPLATES.values():
            assert template_path.startswith(".moai/project/tech/")


class TestGetLanguageTemplate:
    """Test get_language_template function"""

    def test_get_language_template_python(self):
        """Should return Python template path"""
        result = get_language_template("python")
        assert result == ".moai/project/tech/python.md.j2"

    def test_get_language_template_typescript(self):
        """Should return TypeScript template path"""
        result = get_language_template("typescript")
        assert result == ".moai/project/tech/typescript.md.j2"

    def test_get_language_template_case_insensitive(self):
        """Should be case-insensitive"""
        assert get_language_template("Python") == get_language_template("python")
        assert get_language_template("PYTHON") == get_language_template("python")
        assert get_language_template("TypeScript") == get_language_template("typescript")

    def test_get_language_template_unknown_language(self):
        """Should return default template for unknown language"""
        result = get_language_template("unknown_lang")
        assert result == ".moai/project/tech/default.md.j2"

    def test_get_language_template_empty_string(self):
        """Should return default template for empty string"""
        result = get_language_template("")
        assert result == ".moai/project/tech/default.md.j2"

    def test_get_language_template_all_supported_languages(self):
        """Should return correct path for all 20 supported languages"""
        languages = [
            "python",
            "typescript",
            "javascript",
            "java",
            "go",
            "rust",
            "dart",
            "swift",
            "kotlin",
            "csharp",
            "php",
            "ruby",
            "elixir",
            "scala",
            "clojure",
            "haskell",
            "c",
            "cpp",
            "lua",
            "ocaml",
        ]

        for lang in languages:
            result = get_language_template(lang)
            assert result == LANGUAGE_TEMPLATES[lang]
            assert result.endswith(".md.j2")

    def test_get_language_template_special_characters(self):
        """Should handle special characters gracefully"""
        result = get_language_template("C++")
        # C++ is mapped as 'cpp' in lowercase
        assert result == ".moai/project/tech/default.md.j2"

        result = get_language_template("cpp")
        assert result == ".moai/project/tech/cpp.md.j2"

    def test_get_language_template_with_whitespace(self):
        """Should handle whitespace in language name"""
        result = get_language_template(" python ")
        # Current implementation doesn't strip, so this would be unknown
        assert result == ".moai/project/tech/default.md.j2"
