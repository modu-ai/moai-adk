# # REMOVED_ORPHAN_TEST:LANG-005 | SPEC: SPEC-LANGUAGE-DETECTION-001.md | CODE: .moai/docs/
"""Documentation tests for language detection feature.

Tests that language detection and workflow template documentation exists
and contains required information.
"""

from pathlib import Path

import pytest


class TestDocumentation:
    """Documentation Tests (3 tests)"""

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_language_detection_guide_exists(self):
        """Assert: Language detection guide exists"""
        guide = Path(".moai/docs/language-detection-guide.md")
        assert guide.exists(), "Language detection guide not found"
        content = guide.read_text()
        assert "Supported Languages" in content, "Missing 'Supported Languages' section"
        assert "detect" in content.lower(), "Missing detection method documentation"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_workflow_templates_guide_exists(self):
        """Assert: Workflow templates guide exists"""
        guide = Path(".moai/docs/workflow-templates.md")
        assert guide.exists(), "Workflow templates guide not found"
        content = guide.read_text()
        assert "python-tag-validation.yml" in content, "Missing Python workflow documentation"
        assert "javascript-tag-validation.yml" in content, "Missing JavaScript workflow documentation"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_readme_has_language_support_section(self):
        """Assert: README has Language Support section"""
        readme = Path("README.md")
        assert readme.exists(), "README.md not found"
        content = readme.read_text()
        # Check for language support mention (flexible matching)
        has_language_section = any(
            keyword in content.lower()
            for keyword in ["language support", "languages", "detect language", "workflow template"]
        )
        assert has_language_section, "README missing language support information"
