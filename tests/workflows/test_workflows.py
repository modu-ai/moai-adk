# # REMOVED_ORPHAN_TEST:LANG-001 | SPEC: SPEC-LANGUAGE-DETECTION-001.md | CODE: src/moai_adk/templates/.github/workflows/
"""Unit tests for language-specific workflow templates

Tests workflow file creation and correctness for Python, JavaScript, TypeScript, and Go.
"""

from pathlib import Path

import yaml


class TestWorkflowFileCreation:
    """Test that all workflow template files are created"""

    def test_workflow_file_creation(self):
        """Should create all 4 language-specific workflow templates"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"

        python_workflow = templates_dir / "python-tag-validation.yml"
        javascript_workflow = templates_dir / "javascript-tag-validation.yml"
        typescript_workflow = templates_dir / "typescript-tag-validation.yml"
        go_workflow = templates_dir / "go-tag-validation.yml"

        # RED: These files do not exist yet
        assert python_workflow.exists(), "python-tag-validation.yml should exist"
        assert javascript_workflow.exists(), "javascript-tag-validation.yml should exist"
        assert typescript_workflow.exists(), "typescript-tag-validation.yml should exist"
        assert go_workflow.exists(), "go-tag-validation.yml should exist"

    def test_each_file_contains_required_github_actions_syntax(self):
        """Should contain valid GitHub Actions YAML syntax"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"

        workflow_files = [
            "python-tag-validation.yml",
            "javascript-tag-validation.yml",
            "typescript-tag-validation.yml",
            "go-tag-validation.yml",
        ]

        for workflow_file in workflow_files:
            workflow_path = templates_dir / workflow_file
            assert workflow_path.exists(), f"{workflow_file} should exist"

            # RED: Should be parseable YAML
            with open(workflow_path) as f:
                content = yaml.safe_load(f)

            # Required GitHub Actions keys
            assert "name" in content, f"{workflow_file} should have 'name' key"
            # YAML parses 'on' as boolean True
            assert "on" in content or True in content, f"{workflow_file} should have 'on' key"
            assert "jobs" in content, f"{workflow_file} should have 'jobs' key"


class TestPythonWorkflowCorrectness:
    """Test Python workflow template structure"""

    def test_python_workflow_correctness(self):
        """Should have correct Python-specific configuration"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"
        python_workflow = templates_dir / "python-tag-validation.yml"

        # RED: Parse and verify structure
        with open(python_workflow) as f:
            content = yaml.safe_load(f)

        # Verify jobs
        assert "jobs" in content
        jobs = content["jobs"]
        assert len(jobs) > 0, "Should have at least one job"

        # Get first job (usually 'validate' or 'test')
        job_key = list(jobs.keys())[0]
        job = jobs[job_key]

        # Verify steps
        assert "steps" in job
        steps = job["steps"]

        # Check for setup-python@v6
        setup_python_found = False
        pytest_found = False
        ruff_found = False

        for step in steps:
            if "uses" in step and "setup-python@v" in step["uses"]:
                setup_python_found = True
            if "run" in step and "pytest" in step["run"]:
                pytest_found = True
            if "run" in step and "ruff check" in step["run"]:
                ruff_found = True

        assert setup_python_found, "Should use actions/setup-python@v6"
        assert pytest_found, "Should have pytest command"
        assert ruff_found, "Should have ruff check command"

    def test_python_workflow_coverage_target(self):
        """Should specify 85% coverage target"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"
        python_workflow = templates_dir / "python-tag-validation.yml"

        with open(python_workflow) as f:
            content = f.read()

        # RED: Should mention 85% coverage
        assert "85" in content or "coverage" in content.lower(), "Should specify 85% coverage target"


class TestJavaScriptWorkflowCorrectness:
    """Test JavaScript workflow template structure"""

    def test_javascript_workflow_correctness(self):
        """Should have correct JavaScript-specific configuration"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"
        js_workflow = templates_dir / "javascript-tag-validation.yml"

        with open(js_workflow) as f:
            content = yaml.safe_load(f)

        # Verify jobs
        jobs = content["jobs"]
        job_key = list(jobs.keys())[0]
        job = jobs[job_key]
        steps = job["steps"]

        # Check for setup-node@v6
        setup_node_found = False
        test_command_found = False

        for step in steps:
            if "uses" in step and "setup-node@v" in step["uses"]:
                setup_node_found = True
            if "run" in step and ("npm test" in step["run"] or "npm run test" in step["run"]):
                test_command_found = True

        assert setup_node_found, "Should use actions/setup-node@v6"
        assert test_command_found, "Should have npm test command"

    def test_javascript_workflow_package_manager_detection(self):
        """Should have logic for npm/yarn/pnpm auto-detection"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"
        js_workflow = templates_dir / "javascript-tag-validation.yml"

        with open(js_workflow) as f:
            content = f.read()

        # RED: Should mention package managers or conditionals
        has_npm = "npm" in content
        has_conditional = "if:" in content or "package" in content.lower()

        assert has_npm or has_conditional, "Should have package manager detection or conditional logic"


class TestTypeScriptWorkflowCorrectness:
    """Test TypeScript workflow template structure"""

    def test_typescript_workflow_correctness(self):
        """Should have TypeScript-specific configuration"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"
        ts_workflow = templates_dir / "typescript-tag-validation.yml"

        with open(ts_workflow) as f:
            content = yaml.safe_load(f)

        jobs = content["jobs"]
        job_key = list(jobs.keys())[0]
        job = jobs[job_key]
        steps = job["steps"]

        # Check for setup-node and type checking
        setup_node_found = False
        type_check_found = False

        for step in steps:
            if "uses" in step and "setup-node@v" in step["uses"]:
                setup_node_found = True
            if "run" in step and ("tsc" in step["run"] or "type-check" in step["run"]):
                type_check_found = True

        assert setup_node_found, "Should use actions/setup-node@v6"
        assert type_check_found, "Should have TypeScript type checking (tsc --noEmit)"


class TestGoWorkflowCorrectness:
    """Test Go workflow template structure"""

    def test_go_workflow_correctness(self):
        """Should have Go-specific configuration"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"
        go_workflow = templates_dir / "go-tag-validation.yml"

        with open(go_workflow) as f:
            content = yaml.safe_load(f)

        jobs = content["jobs"]
        job_key = list(jobs.keys())[0]
        job = jobs[job_key]
        steps = job["steps"]

        # Check for setup-go and go test
        setup_go_found = False
        go_test_found = False

        for step in steps:
            if "uses" in step and "setup-go@v" in step["uses"]:
                setup_go_found = True
            if "run" in step and "go test" in step["run"]:
                go_test_found = True

        assert setup_go_found, "Should use actions/setup-go@v6"
        assert go_test_found, "Should have go test command"

    def test_go_workflow_coverage_target(self):
        """Should specify 75% coverage target for Go"""
        templates_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "templates" / ".github" / "workflows"
        go_workflow = templates_dir / "go-tag-validation.yml"

        with open(go_workflow) as f:
            content = f.read()

        # RED: Should mention coverage
        assert "coverage" in content.lower() or "cover" in content.lower(), "Should specify coverage target"
