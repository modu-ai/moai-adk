# @TEST:TRUST-001 | SPEC: SPEC-TRUST-001/spec.md
"""
Integrated TRUST principle validation tests

20 Given-When-Then-style test cases:
- AC-001: Test coverage ≥85% (2 cases: pass/fail)
- AC-002: File ≤300 LOC (2 cases: pass/fail)
- AC-003: Function ≤50 LOC (2 cases: pass/fail)
- AC-004: Parameters ≤5 (2 cases: pass/fail)
- AC-005: Cyclomatic complexity ≤10 (2 cases: pass/fail)
- AC-006: @TAG chain completeness (2 cases: pass/fail)
- AC-007: Orphan TAG detection (2 cases: detected/none)
- AC-008: Report generation (2 cases: markdown/json)
- AC-009: Error messaging (2 cases: specific/generic)
- AC-010: Tool selection per language (2 cases: python/typescript)
"""

from pathlib import Path

import pytest


class TestTrustChecker:
    """@TEST:TRUST-001: Integrated TRUST principle validation"""

    @pytest.fixture
    def trust_checker(self):
        """Create a TrustChecker instance"""
        from moai_adk.core.quality.trust_checker import TrustChecker
        return TrustChecker()

    @pytest.fixture
    def sample_project_path(self, tmp_path: Path) -> Path:
        """Create a sample project directory for tests"""
        project = tmp_path / "sample_project"
        project.mkdir()
        (project / "src").mkdir()
        (project / "tests").mkdir()
        (project / ".moai").mkdir()
        return project

    # ========================================
    # AC-001: Validate test coverage ≥85%
    # ========================================

    def test_should_pass_when_coverage_above_85_percent(self, trust_checker, sample_project_path):
        """
        Given: Project test coverage equals 87%
        When: trust_checker.validate_coverage() runs
        Then: ValidationResult.passed = True
        """
        # Arrange
        coverage_data = {"total_coverage": 87.5}

        # Act
        result = trust_checker.validate_coverage(sample_project_path, coverage_data)

        # Assert
        assert result.passed is True
        assert result.message == "Test coverage: 87.5% (Target: 85%)"

    def test_should_fail_when_coverage_below_85_percent(self, trust_checker, sample_project_path):
        """
        Given: Project test coverage equals 78%
        When: trust_checker.validate_coverage() runs
        Then: ValidationResult.passed = False and includes low-coverage files
        """
        # Arrange
        coverage_data = {
            "total_coverage": 78.0,
            "low_coverage_files": [
                {"file": "src/utils/helper.py", "coverage": 72.0},
                {"file": "src/core/validator.py", "coverage": 78.0},
            ],
        }

        # Act
        result = trust_checker.validate_coverage(sample_project_path, coverage_data)

        # Assert
        assert result.passed is False
        assert "78.0%" in result.message
        assert "src/utils/helper.py" in result.details

    # ========================================
    # AC-002: Validate file size ≤300 LOC
    # ========================================

    def test_should_pass_when_all_files_within_300_loc(self, trust_checker, sample_project_path):
        """
        Given: All source files stay within 300 LOC
        When: trust_checker.validate_file_size() runs
        Then: ValidationResult.passed = True
        """
        # Arrange
        (sample_project_path / "src" / "small.py").write_text("\n".join([f"# Line {i}" for i in range(200)]))

        # Act
        result = trust_checker.validate_file_size(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All files within 300 LOC" in result.message

    def test_should_fail_when_file_exceeds_300_loc(self, trust_checker, sample_project_path):
        """
        Given: A file has 342 LOC, exceeding the 300 LOC limit
        When: trust_checker.validate_file_size() runs
        Then: ValidationResult.passed = False and cites violating files
        """
        # Arrange
        large_file = sample_project_path / "src" / "large.py"
        large_file.write_text("\n".join([f"# Line {i}" for i in range(342)]))

        # Act
        result = trust_checker.validate_file_size(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "large.py" in result.details
        assert "342 LOC" in result.details

    # ========================================
    # AC-003: Validate function size ≤50 LOC
    # ========================================

    def test_should_pass_when_all_functions_within_50_loc(self, trust_checker, sample_project_path):
        """
        Given: Every function stays within 50 LOC
        When: trust_checker.validate_function_size() runs
        Then: ValidationResult.passed = True
        """
        # Arrange
        code = """
def small_function():
    # 30 LOC function
""" + "\n".join([f"    pass  # Line {i}" for i in range(30)])
        (sample_project_path / "src" / "functions.py").write_text(code)

        # Act
        result = trust_checker.validate_function_size(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All functions within 50 LOC" in result.message

    def test_should_fail_when_function_exceeds_50_loc(self, trust_checker, sample_project_path):
        """
        Given: A function has 58 LOC, exceeding the 50 LOC limit
        When: trust_checker.validate_function_size() runs
        Then: ValidationResult.passed = False and lists violating functions
        """
        # Arrange
        code = """
def large_function():
    # 58 LOC function
""" + "\n".join([f"    pass  # Line {i}" for i in range(58)])
        (sample_project_path / "src" / "functions.py").write_text(code)

        # Act
        result = trust_checker.validate_function_size(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "large_function" in result.details
        assert " LOC" in result.details  # Actual LOC may reach 60 (header + body)

    # ========================================
    # AC-004: Validate parameters ≤5
    # ========================================

    def test_should_pass_when_all_params_within_5(self, trust_checker, sample_project_path):
        """
        Given: Every function declares at most 5 parameters
        When: trust_checker.validate_param_count() runs
        Then: ValidationResult.passed = True
        """
        # Arrange
        code = """
def function_with_4_params(a, b, c, d):
    pass
"""
        (sample_project_path / "src" / "params.py").write_text(code)

        # Act
        result = trust_checker.validate_param_count(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All functions within 5 parameters" in result.message

    def test_should_fail_when_params_exceed_5(self, trust_checker, sample_project_path):
        """
        Given: A function declares 7 parameters, exceeding the limit of 5
        When: trust_checker.validate_param_count() runs
        Then: ValidationResult.passed = False and lists violating functions
        """
        # Arrange
        code = """
def function_with_7_params(a, b, c, d, e, f, g):
    pass
"""
        (sample_project_path / "src" / "params.py").write_text(code)

        # Act
        result = trust_checker.validate_param_count(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "function_with_7_params" in result.details
        assert "7 parameters" in result.details

    # ========================================
    # AC-005: Validate cyclomatic complexity ≤10
    # ========================================

    def test_should_pass_when_complexity_within_10(self, trust_checker, sample_project_path):
        """
        Given: Every function has cyclomatic complexity ≤10
        When: trust_checker.validate_complexity() runs
        Then: ValidationResult.passed = True
        """
        # Arrange
        code = """
def simple_function(x):
    if x > 0:
        return x
    return 0
"""
        (sample_project_path / "src" / "complexity.py").write_text(code)

        # Act
        result = trust_checker.validate_complexity(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All functions within complexity 10" in result.message

    def test_should_fail_when_complexity_exceeds_10(self, trust_checker, sample_project_path):
        """
        Given: A function reports cyclomatic complexity 15, exceeding the limit of 10
        When: trust_checker.validate_complexity() runs
        Then: ValidationResult.passed = False and lists violating functions
        """
        # Arrange - create complexity 13 with 12 nested if statements
        code = """
def complex_function(x):
    if x > 0:
        if x > 10:
            if x > 20:
                if x > 30:
                    if x > 40:
                        if x > 50:
                            if x > 60:
                                if x > 70:
                                    if x > 80:
                                        if x > 90:
                                            if x > 100:
                                                if x > 110:
                                                    return x
    return 0
"""
        (sample_project_path / "src" / "complexity.py").write_text(code)

        # Act
        result = trust_checker.validate_complexity(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "complex_function" in result.details
        assert "complexity" in result.details.lower()

    # ========================================
    # AC-006: Validate @TAG chain completeness
    # ========================================

    def test_should_pass_when_tag_chain_complete(self, trust_checker, sample_project_path):
        """
        Given: @SPEC, @CODE, and @TEST TAGs are all connected
        When: trust_checker.validate_tag_chain() runs
        Then: ValidationResult.passed = True
        """
        # Arrange
        (sample_project_path / ".moai" / "specs").mkdir(parents=True)
        (sample_project_path / ".moai" / "specs" / "SPEC-AUTH-001.md").write_text("# @SPEC:AUTH-001")
        (sample_project_path / "src" / "auth.py").write_text("# @CODE:AUTH-001")
        (sample_project_path / "tests" / "test_auth.py").write_text("# @TEST:AUTH-001")

        # Act
        result = trust_checker.validate_tag_chain(sample_project_path)

        # Assert
        assert result.passed is True
        assert "TAG chain complete" in result.message

    def test_should_fail_when_tag_chain_broken(self, trust_checker, sample_project_path):
        """
        Given: @CODE:AUTH-001 exists but @SPEC:AUTH-001 is missing
        When: trust_checker.validate_tag_chain() runs
        Then: ValidationResult.passed = False and highlights broken chains
        """
        # Arrange
        (sample_project_path / "src" / "auth.py").write_text("# @CODE:AUTH-001")

        # Act
        result = trust_checker.validate_tag_chain(sample_project_path)

        # Assert
        assert result.passed is False
        assert "auth-001" in result.details.lower()  # Compare in lowercase
        assert "broken" in result.details.lower()

    # ========================================
    # AC-007: Detect orphan TAGs
    # ========================================

    def test_should_detect_orphan_tags(self, trust_checker, sample_project_path):
        """
        Given: @CODE:USER-005 exists but @SPEC:USER-005 is missing (orphan TAG)
        When: trust_checker.detect_orphan_tags() runs
        Then: Returns the orphan TAG list
        """
        # Arrange
        (sample_project_path / "src" / "user.py").write_text("# @CODE:USER-005")

        # Act
        orphans = trust_checker.detect_orphan_tags(sample_project_path)

        # Assert
        assert len(orphans) > 0
        assert any("USER-005" in tag for tag in orphans)

    def test_should_return_empty_when_no_orphan_tags(self, trust_checker, sample_project_path):
        """
        Given: All TAGs are properly connected
        When: trust_checker.detect_orphan_tags() runs
        Then: Returns an empty list
        """
        # Arrange
        (sample_project_path / ".moai" / "specs").mkdir(parents=True)
        (sample_project_path / ".moai" / "specs" / "SPEC-USER-001.md").write_text("# @SPEC:USER-001")
        (sample_project_path / "src" / "user.py").write_text("# @CODE:USER-001")

        # Act
        orphans = trust_checker.detect_orphan_tags(sample_project_path)

        # Assert
        assert len(orphans) == 0

    # ========================================
    # AC-008: Generate validation reports
    # ========================================

    def test_should_generate_markdown_report(self, trust_checker, sample_project_path):
        """
        Given: TRUST validation completed
        When: trust_checker.generate_report(format="markdown") runs
        Then: Produces a markdown report
        """
        # Arrange
        results = {"coverage": {"passed": True, "value": 87}}

        # Act
        report = trust_checker.generate_report(results, format="markdown")

        # Assert
        assert "# TRUST Validation Report" in report
        assert "87%" in report
        assert "✅" in report or "PASS" in report

    def test_should_generate_json_report(self, trust_checker, sample_project_path):
        """
        Given: TRUST validation completed
        When: trust_checker.generate_report(format="json") runs
        Then: Produces a JSON report
        """
        # Arrange
        results = {"coverage": {"passed": True, "value": 87}}

        # Act
        report_json = trust_checker.generate_report(results, format="json")

        # Assert
        import json

        report = json.loads(report_json)
        assert "coverage" in report
        assert report["coverage"]["passed"] is True
        assert report["coverage"]["value"] == 87

    # ========================================
    # AC-009: Provide detailed error messages on failure
    # ========================================

    def test_should_provide_specific_error_message(self, trust_checker, sample_project_path):
        """
        Given: Test coverage is 78% (below the threshold)
        When: trust_checker.validate_coverage() runs
        Then: Includes detailed guidance (coverage, low files, recommended action)
        """
        # Arrange
        coverage_data = {
            "total_coverage": 78.0,
            "low_coverage_files": [{"file": "src/utils/helper.py", "coverage": 72.0}],
        }

        # Act
        result = trust_checker.validate_coverage(sample_project_path, coverage_data)

        # Assert
        assert "78.0%" in result.message
        assert "helper.py" in result.details
        assert "recommended" in result.details.lower()

    def test_should_provide_generic_error_when_no_details(self, trust_checker, sample_project_path):
        """
        Given: Validation fails but no extra details are provided
        When: trust_checker.validate_coverage() runs
        Then: Returns a generic error message
        """
        # Arrange
        # Simulate a failure without detailed information

        # Act
        result = trust_checker.validate_coverage(sample_project_path, {"total_coverage": 70.0})

        # Assert
        assert result.passed is False
        assert "coverage" in result.message.lower()

    # ========================================
    # AC-010: Automatically select tools per language
    # ========================================

    def test_should_select_python_tools(self, trust_checker, sample_project_path):
        """
        Given: .moai/config.json defines project.language as \"python\"
        When: trust_checker.select_tools() runs
        Then: Select pytest, coverage.py, mypy, and ruff
        """
        # Arrange
        config = {"project": {"language": "python"}}
        import json

        (sample_project_path / ".moai" / "config.json").write_text(json.dumps(config))

        # Act
        tools = trust_checker.select_tools(sample_project_path)

        # Assert
        assert tools["test_framework"] == "pytest"
        assert tools["coverage_tool"] == "coverage.py"
        assert tools["linter"] == "ruff"
        assert tools["type_checker"] == "mypy"

    def test_should_select_typescript_tools(self, trust_checker, sample_project_path):
        """
        Given: .moai/config.json defines project.language as \"typescript\"
        When: trust_checker.select_tools() runs
        Then: Select Vitest, Biome, and tsc
        """
        # Arrange
        config = {"project": {"language": "typescript"}}
        import json

        (sample_project_path / ".moai" / "config.json").write_text(json.dumps(config))

        # Act
        tools = trust_checker.select_tools(sample_project_path)

        # Assert
        assert tools["test_framework"] == "vitest"
        assert tools["linter"] == "biome"
        assert tools["type_checker"] == "tsc"
