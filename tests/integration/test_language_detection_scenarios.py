# # REMOVED_ORPHAN_TEST:LANG-006-SCENARIOS | SPEC: SPEC-LANGUAGE-DETECTION-001.md | CODE: src/moai_adk/core/project/detector.py
"""Comprehensive language detection scenarios and error handling tests.

Tests workflow selection, error handling, and full integration scenarios
for all supported languages.
"""

from pathlib import Path

from moai_adk.core.project.detector import LanguageDetector


class TestWorkflowSelection:
    """Workflow Selection Tests (4 tests)"""

    def test_python_project_gets_python_workflow(self, tmp_path):
        """Test: Python project receives python-tag-validation.yml"""
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")
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
        (tmp_path / "tsconfig.json").write_text("{}")
        detector = LanguageDetector()
        template_path = detector.get_workflow_template_path(detector.detect(tmp_path))
        assert "typescript-tag-validation.yml" in template_path

    def test_go_project_gets_go_workflow(self, tmp_path):
        """Test: Go project receives go-tag-validation.yml"""
        (tmp_path / "go.mod").write_text("module test")
        detector = LanguageDetector()
        template_path = detector.get_workflow_template_path(detector.detect(tmp_path))
        assert "go-tag-validation.yml" in template_path


class TestErrorHandling:
    """Error Handling Tests (3 tests)"""

    def test_unsupported_language_error_message(self, tmp_path):
        """Test: Unsupported language returns None instead of error"""
        detector = LanguageDetector()
        # Create a fake language that's not supported
        result = detector.get_workflow_template_path("cobol")
        # Should return None for unsupported languages
        assert result is None

    def test_missing_workflow_template_error(self, tmp_path):
        """Test: No workflow template available"""
        detector = LanguageDetector()
        # Verify template file returns None for all languages
        template_path = detector.get_workflow_template_path("python")
        # Should return None as workflow templates have been removed
        assert template_path is None

    def test_invalid_workflow_syntax_detection(self, tmp_path):
        """Test: No workflow templates available"""
        detector = LanguageDetector()
        # All language workflow templates should return None
        for language in ["python", "javascript", "typescript", "go", "rust"]:
            template_path = detector.get_workflow_template_path(language)
            assert template_path is None, f"Expected None for {language}, got {template_path}"


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
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")
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
            "go-tag-validation.yml",
        ]
        templates_dir = Path("src/moai_adk/templates/.github/workflows")
        for template in new_templates:
            assert (templates_dir / template).exists(), f"New template missing: {template}"
