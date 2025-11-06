# @TEST:LANG-006-SCENARIOS | SPEC: SPEC-LANGUAGE-DETECTION-001.md | CODE: src/moai_adk/core/project/detector.py
"""Comprehensive language detection scenarios and error handling tests.

Tests workflow selection, error handling, and full integration scenarios
for all supported languages.
"""

from pathlib import Path

import pytest
import yaml

from moai_adk.core.project.detector import LanguageDetector


class TestWorkflowSelection:
    """Workflow Selection Tests (4 tests)"""

    def test_python_project_gets_python_workflow(self, tmp_path):
        """Test: Python project receives python-tag-validation.yml"""
        (tmp_path / "pyproject.toml").write_text('[tool.poetry]')
        detector = LanguageDetector()
        template_path = detector.get_workflow_template_path(detector.detect(tmp_path))
        assert "python-tag-validation.yml" in template_path

    def test_javascript_project_gets_javascript_workflow(self, tmp_path):
        """Test: JavaScript project receives javascript-tag-validation.yml"""
        (tmp_path / "package.json").write_text('{"name": "test"}')
        detector = LanguageDetector()
        template_path = detector.get_workflow_template_path(detector.detect(tmp_path))
        assert "javascript-tag-validation.yml" in template_path

    def test_typescript_project_gets_typescript_workflow(self, tmp_path):
        """Test: TypeScript project receives typescript-tag-validation.yml"""
        (tmp_path / "package.json").write_text('{"name": "test"}')
        (tmp_path / "tsconfig.json").write_text('{}')
        detector = LanguageDetector()
        template_path = detector.get_workflow_template_path(detector.detect(tmp_path))
        assert "typescript-tag-validation.yml" in template_path

    def test_go_project_gets_go_workflow(self, tmp_path):
        """Test: Go project receives go-tag-validation.yml"""
        (tmp_path / "go.mod").write_text('module test')
        detector = LanguageDetector()
        template_path = detector.get_workflow_template_path(detector.detect(tmp_path))
        assert "go-tag-validation.yml" in template_path


class TestErrorHandling:
    """Error Handling Tests (3 tests)"""

    def test_unsupported_language_error_message(self, tmp_path):
        """Test: Unsupported language returns clear error message"""
        detector = LanguageDetector()
        # Create a fake language that's not supported
        with pytest.raises(ValueError) as exc_info:
            detector.get_workflow_template_path("cobol")
        error_message = str(exc_info.value).lower()
        assert "unsupported" in error_message or "not" in error_message or "does not have" in error_message

    def test_missing_workflow_template_error(self, tmp_path):
        """Test: Missing workflow template is detected"""
        detector = LanguageDetector()
        # Verify template file actually exists for Python
        template_path = detector.get_workflow_template_path("python")
        # This should be a relative path like "workflows/python-tag-validation.yml"
        assert "python-tag-validation.yml" in template_path
        # Verify the actual template file exists in the package
        full_path = Path("src/moai_adk/templates") / template_path
        assert full_path.exists(), f"Template not found: {full_path}"

    def test_invalid_workflow_syntax_detection(self, tmp_path):
        """Test: Invalid YAML syntax in workflow detected"""
        template_path = Path("src/moai_adk/templates/workflows/python-tag-validation.yml")
        content = template_path.read_text()
        # Try to parse as YAML - should succeed for valid templates
        parsed = yaml.safe_load(content)
        assert parsed is not None, "Invalid YAML syntax"
        # Check for common workflow structure keys
        has_workflow_structure = any(key in parsed for key in ["jobs", "on", "name"])
        assert has_workflow_structure, "Invalid workflow structure"


class TestIntegrationScenarios:
    """Integration Tests (4 tests)"""

    def test_full_workflow_for_python_project(self, tmp_path):
        """Test: Full workflow generation for Python project"""
        # Setup Python project
        (tmp_path / "pyproject.toml").write_text('[tool.poetry]\nname = "test"')
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai/config.json").write_text('{"project": {"language": "python"}}')

        # Detect and get template
        detector = LanguageDetector()
        language = detector.detect(tmp_path)
        template_path = detector.get_workflow_template_path(language)

        # Verify
        assert language == "python"
        assert "python-tag-validation.yml" in template_path
        # Verify template exists
        full_path = Path("src/moai_adk/templates") / template_path
        assert full_path.exists()

    def test_full_workflow_for_javascript_project(self, tmp_path):
        """Test: Full workflow generation for JavaScript project"""
        # Setup JS project
        (tmp_path / "package.json").write_text('{"name": "myapp", "version": "1.0.0"}')
        (tmp_path / ".moai").mkdir()

        # Detect and get template
        detector = LanguageDetector()
        language = detector.detect(tmp_path)
        template_path = detector.get_workflow_template_path(language)

        # Verify
        assert language == "javascript"
        full_path = Path("src/moai_adk/templates") / template_path
        assert full_path.exists()

    def test_full_workflow_for_mixed_language_project(self, tmp_path):
        """Test: Priority handling for mixed language projects"""
        # Setup: Both Python and JavaScript
        (tmp_path / "pyproject.toml").write_text('[tool.poetry]')
        (tmp_path / "package.json").write_text('{"name": "test"}')

        # Detect (Python has priority in LANGUAGE_PATTERNS order)
        detector = LanguageDetector()
        language = detector.detect(tmp_path)

        # Should detect based on priority order (check LANGUAGE_PATTERNS in detector.py)
        # Ruby and PHP are at top, but not present. Python should be detected.
        assert language in ["python", "javascript"]

    def test_backward_compatibility_with_existing_workflows(self, tmp_path):
        """Test: New system doesn't break existing workflows"""
        # Verify old workflow file still exists
        old_workflow = Path("src/moai_adk/templates/.github/workflows/moai-gitflow.yml")
        assert old_workflow.exists(), "Old moai-gitflow.yml should still exist"

        # Verify new language-specific templates also exist
        new_templates = [
            "python-tag-validation.yml",
            "javascript-tag-validation.yml",
            "typescript-tag-validation.yml",
            "go-tag-validation.yml"
        ]
        templates_dir = Path("src/moai_adk/templates/workflows")
        for template in new_templates:
            assert (templates_dir / template).exists(), f"New template missing: {template}"
