"""Unit tests for moai_adk.core.template.languages module."""

from moai_adk.core.template.languages import LANGUAGE_TEMPLATES, get_language_template


class TestLanguageTemplates:
    """Test language template functionality."""

    def test_language_templates_dict(self):
        """Test LANGUAGE_TEMPLATES dictionary."""
        assert isinstance(LANGUAGE_TEMPLATES, dict)
        assert len(LANGUAGE_TEMPLATES) > 0
        assert "python" in LANGUAGE_TEMPLATES
        assert "typescript" in LANGUAGE_TEMPLATES

    def test_get_language_template_python(self):
        """Test getting Python template."""
        template = get_language_template("python")
        assert template.endswith(".md.j2")
        assert "python" in template.lower()

    def test_get_language_template_typescript(self):
        """Test getting TypeScript template."""
        template = get_language_template("typescript")
        assert "typescript" in template.lower()

    def test_get_language_template_case_insensitive(self):
        """Test case-insensitive template lookup."""
        template1 = get_language_template("python")
        template2 = get_language_template("PYTHON")
        template3 = get_language_template("Python")

        assert template1 == template2 == template3

    def test_get_language_template_unknown(self):
        """Test getting template for unknown language."""
        template = get_language_template("unknown_language")
        assert "default" in template.lower()

    def test_get_language_template_empty(self):
        """Test getting template with empty string."""
        template = get_language_template("")
        assert "default" in template.lower()

    def test_get_language_template_none(self):
        """Test getting template with None."""
        template = get_language_template(None)
        assert "default" in template.lower()

    def test_all_language_templates_valid(self):
        """Test all language templates have valid format."""
        for lang, template in LANGUAGE_TEMPLATES.items():
            assert isinstance(lang, str)
            assert isinstance(template, str)
            assert template.endswith(".md.j2")
            assert ".moai" in template
