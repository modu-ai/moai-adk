# # REMOVED_ORPHAN_TEST:LANG-003 | SPEC: SPEC-LANGUAGE-DETECTION-001.md | CODE: .claude/agents/alfred/tdd-implementer.md
"""Integration tests for tdd-implementer agent language detection.

Tests that tdd-implementer correctly detects project language and selects
appropriate workflow templates for CI/CD generation.
"""

from moai_adk.core.project.detector import LanguageDetector


def test_tdd_implementer_detects_python_project(tmp_path):
    """Test: tdd-implementer detects Python project and selects python workflow"""
    # Setup: Create mock Python project structure
    project_dir = tmp_path / "test_python_project"
    project_dir.mkdir()
    (project_dir / "pyproject.toml").touch()
    (project_dir / ".moai").mkdir()
    (project_dir / ".moai/config.json").write_text('{"project": {"language": "python"}}')

    # Action: Simulate tdd-implementer workflow generation
    detector = LanguageDetector()
    language = detector.detect(project_dir)
    template_path = detector.get_workflow_template_path(language)

    # Assert: Python workflow selected
    assert language == "python"
    assert "python-tag-validation.yml" in template_path
    # Note: Template existence checked in template tests, not here


def test_tdd_implementer_detects_javascript_project(tmp_path):
    """Test: tdd-implementer detects JavaScript project and selects javascript workflow"""
    # Setup: Create mock JavaScript project
    project_dir = tmp_path / "test_js_project"
    project_dir.mkdir()
    (project_dir / "package.json").write_text('{"name": "test"}')

    # Action: Detect and get template
    detector = LanguageDetector()
    language = detector.detect(project_dir)
    template_path = detector.get_workflow_template_path(language)

    # Assert: JavaScript workflow selected
    assert language == "javascript"
    assert "javascript-tag-validation.yml" in template_path


def test_tdd_implementer_detects_typescript_project(tmp_path):
    """Test: tdd-implementer correctly handles TypeScript (priority over JS)"""
    # Setup: Both package.json and tsconfig.json present
    project_dir = tmp_path / "test_ts_project"
    project_dir.mkdir()
    (project_dir / "package.json").write_text('{"name": "test"}')
    (project_dir / "tsconfig.json").write_text("{}")

    # Action: Detect language
    detector = LanguageDetector()
    language = detector.detect(project_dir)
    template_path = detector.get_workflow_template_path(language)

    # Assert: TypeScript has priority over JavaScript
    assert language == "typescript"
    assert "typescript-tag-validation.yml" in template_path


def test_tdd_implementer_wrong_workflow_never_applied(tmp_path):
    """Test: Python workflow never applied to JavaScript project"""
    # Setup: JavaScript project
    project_dir = tmp_path / "test_js_project"
    project_dir.mkdir()
    (project_dir / "package.json").write_text('{"name": "test"}')

    # Action: Get workflow path
    detector = LanguageDetector()
    language = detector.detect(project_dir)
    template_path = detector.get_workflow_template_path(language)

    # Assert: Python-specific workflow NOT selected
    assert "python-tag-validation.yml" not in template_path
    assert "javascript-tag-validation.yml" in template_path


def test_tdd_implementer_detects_go_project(tmp_path):
    """Test: tdd-implementer detects Go project"""
    # Setup: Go project with go.mod
    project_dir = tmp_path / "test_go_project"
    project_dir.mkdir()
    (project_dir / "go.mod").write_text("module test")

    # Action: Detect and get template
    detector = LanguageDetector()
    language = detector.detect(project_dir)
    template_path = detector.get_workflow_template_path(language)

    # Assert: Go workflow selected
    assert language == "go"
    assert "go-tag-validation.yml" in template_path
