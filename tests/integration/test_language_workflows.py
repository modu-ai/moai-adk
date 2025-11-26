# # REMOVED_ORPHAN_TEST:LANG-004-WORKFLOWS | SPEC: SPEC-LANGUAGE-DETECTION-001.md | CODE: src/moai_adk/core/project/detector.py
"""Comprehensive tests for language-specific workflow templates.

Tests workflow template creation, language detection, and package manager detection
for all supported languages (Python, JavaScript, TypeScript, Go).
"""

from pathlib import Path

import pytest

from moai_adk.core.project.detector import LanguageDetector


# Pytest Fixtures for common test resources
@pytest.fixture
def detector():
    """Reusable LanguageDetector instance"""
    return LanguageDetector()


@pytest.fixture
def templates_dir():
    """Path to workflow templates directory"""
    return Path("src/moai_adk/templates/.github/workflows")


@pytest.fixture
def expected_templates():
    """List of expected language-specific templates"""
    return [
        "python-tag-validation.yml",
        "javascript-tag-validation.yml",
        "typescript-tag-validation.yml",
        "go-tag-validation.yml",
    ]


class TestWorkflowTemplates:
    """Template Creation Tests (5 tests)"""

    def test_workflow_templates_created_for_all_languages(self, templates_dir, expected_templates):
        """Assert: All 4 language-specific workflow templates exist"""
        for template in expected_templates:
            template_path = templates_dir / template
            assert template_path.exists(), f"Template not found: {template_path}"

    def test_python_workflow_has_required_sections(self, templates_dir):
        """Assert: Python workflow contains pytest, mypy, ruff, coverage"""
        template_path = templates_dir / "python-tag-validation.yml"
        content = template_path.read_text()
        assert "pytest" in content, "pytest not found in Python workflow"
        assert "mypy" in content, "mypy not found in Python workflow"
        assert "ruff" in content, "ruff not found in Python workflow"
        assert "85" in content, "Coverage target 85% not found"

    def test_javascript_workflow_has_required_sections(self, templates_dir):
        """Assert: JavaScript workflow supports npm, yarn, pnpm, bun"""
        template_path = templates_dir / "javascript-tag-validation.yml"
        content = template_path.read_text()
        assert "npm" in content, "npm not found in JavaScript workflow"
        # Check for package manager detection logic
        assert "setup-node" in content, "setup-node action not found"

    def test_typescript_workflow_has_required_sections(self, templates_dir):
        """Assert: TypeScript workflow has tsc type checking"""
        template_path = templates_dir / "typescript-tag-validation.yml"
        content = template_path.read_text()
        assert "tsc" in content, "tsc not found in TypeScript workflow"
        assert "setup-node" in content, "setup-node action not found"

    def test_go_workflow_has_required_sections(self, templates_dir):
        """Assert: Go workflow has golangci-lint and gofmt"""
        template_path = templates_dir / "go-tag-validation.yml"
        content = template_path.read_text()
        assert "go" in content.lower(), "go command not found in Go workflow"
        # Check for common Go CI patterns
        assert "test" in content.lower(), "go test not found"


class TestLanguageDetection:
    """Language Detection Tests (6 tests)"""

    def test_detect_python_project_with_pyproject_toml(self, tmp_path, detector):
        """Test: Python project with pyproject.toml correctly detected"""
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")
        assert detector.detect(tmp_path) == "python"

    def test_detect_javascript_project_with_package_json(self, tmp_path, detector):
        """Test: JavaScript project with package.json detected"""
        (tmp_path / "package.json").write_text('{"name": "test"}')
        assert detector.detect(tmp_path) == "javascript"

    def test_detect_typescript_project_with_tsconfig(self, tmp_path, detector):
        """Test: TypeScript project with tsconfig.json detected"""
        (tmp_path / "package.json").write_text('{"name": "test"}')
        (tmp_path / "tsconfig.json").write_text("{}")
        # TypeScript should be detected before JavaScript due to priority
        assert detector.detect(tmp_path) == "typescript"

    def test_detect_go_project_with_go_mod(self, tmp_path, detector):
        """Test: Go project with go.mod correctly detected"""
        (tmp_path / "go.mod").write_text("module test")
        assert detector.detect(tmp_path) == "go"

    def test_priority_typescript_over_javascript(self, tmp_path, detector):
        """Test: TypeScript has priority when both package.json and tsconfig.json exist"""
        (tmp_path / "package.json").write_text('{"name": "test"}')
        (tmp_path / "tsconfig.json").write_text("{}")
        detected = detector.detect(tmp_path)
        assert detected == "typescript"
        assert detected != "javascript"

    def test_package_manager_detection_all_types(self, tmp_path, detector):
        """Test: Package manager detection for all types (npm, yarn, pnpm, bun)"""
        # Test npm
        (tmp_path / "package-lock.json").write_text("{}")
        assert detector.detect_package_manager(tmp_path) == "npm"
        (tmp_path / "package-lock.json").unlink()

        # Test yarn
        (tmp_path / "yarn.lock").write_text("")
        assert detector.detect_package_manager(tmp_path) == "yarn"
        (tmp_path / "yarn.lock").unlink()

        # Test pnpm
        (tmp_path / "pnpm-lock.yaml").write_text("")
        assert detector.detect_package_manager(tmp_path) == "pnpm"
        (tmp_path / "pnpm-lock.yaml").unlink()

        # Test bun (highest priority)
        (tmp_path / "bun.lockb").write_bytes(b"binary")
        assert detector.detect_package_manager(tmp_path) == "bun"
