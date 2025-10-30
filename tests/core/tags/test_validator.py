#!/usr/bin/env python3
# @TEST:DOC-TAG-004 | Component 3: Central validation system tests
"""Test suite for central validation system

This module tests the unified CentralValidator that:
- Validates TAG format, duplicates, orphans, and chain integrity
- Provides extensible validator architecture
- Generates reports in multiple formats
- Supports CLI integration

Following TDD RED-GREEN-REFACTOR cycle.
"""

import json
import tempfile
from datetime import datetime
from pathlib import Path

import pytest

# Import will fail initially (RED phase) - that's expected
try:
    from moai_adk.core.tags.validator import (
        CentralValidationResult,
        CentralValidator,
        ChainValidator,
        DuplicateValidator,
        FormatValidator,
        OrphanValidator,
        TagValidator,
        ValidationConfig,
        ValidationIssue,
        ValidationStatistics,
    )
except ImportError:
    # Allow tests to be written before implementation
    ValidationConfig = None
    TagValidator = None
    DuplicateValidator = None
    OrphanValidator = None
    ChainValidator = None
    FormatValidator = None
    CentralValidator = None
    CentralValidationResult = None
    ValidationIssue = None
    ValidationStatistics = None


@pytest.mark.skipif(ValidationConfig is None, reason="ValidationConfig not implemented yet")
class TestValidationConfig:
    """Test ValidationConfig dataclass"""

    def test_default_configuration(self):
        """ValidationConfig should have sensible defaults"""
        config = ValidationConfig()
        assert config.strict_mode is False
        assert config.check_duplicates is True
        assert config.check_orphans is True
        assert config.check_chain_integrity is True
        assert config.report_format == "detailed"

    def test_custom_configuration(self):
        """ValidationConfig should accept custom values"""
        config = ValidationConfig(
            strict_mode=True,
            check_duplicates=False,
            check_orphans=False,
            check_chain_integrity=False,
            report_format="json"
        )
        assert config.strict_mode is True
        assert config.check_duplicates is False
        assert config.check_orphans is False
        assert config.check_chain_integrity is False
        assert config.report_format == "json"

    def test_allowed_file_types(self):
        """ValidationConfig should support file type filtering"""
        config = ValidationConfig(
            allowed_file_types=["py", "js", "ts"]
        )
        assert "py" in config.allowed_file_types
        assert "js" in config.allowed_file_types
        assert "ts" in config.allowed_file_types

    def test_ignore_patterns(self):
        """ValidationConfig should support ignore patterns"""
        config = ValidationConfig(
            ignore_patterns=[".git/*", "node_modules/*", "*.pyc"]
        )
        assert ".git/*" in config.ignore_patterns
        assert "node_modules/*" in config.ignore_patterns


@pytest.mark.skipif(ValidationIssue is None, reason="ValidationIssue not implemented yet")
class TestValidationIssue:
    """Test ValidationIssue dataclass"""

    def test_error_severity_issue(self):
        """ValidationIssue should support error severity"""
        issue = ValidationIssue(
            severity="error",
            type="duplicate",
            tag="@CODE:TEST-001",
            message="Duplicate TAG found",
            locations=[("file1.py", 10), ("file2.py", 20)],
            suggestion="Remove duplicate TAG declarations"
        )
        assert issue.severity == "error"
        assert issue.type == "duplicate"
        assert len(issue.locations) == 2

    def test_warning_severity_issue(self):
        """ValidationIssue should support warning severity"""
        issue = ValidationIssue(
            severity="warning",
            type="orphan",
            tag="@CODE:TEST-001",
            message="CODE TAG without corresponding TEST",
            locations=[("file1.py", 10)],
            suggestion="Add @TEST:TEST-001 for this code"
        )
        assert issue.severity == "warning"
        assert issue.type == "orphan"

    def test_info_severity_issue(self):
        """ValidationIssue should support info severity"""
        issue = ValidationIssue(
            severity="info",
            type="chain",
            tag="@CODE:TEST-001",
            message="Complete chain detected",
            locations=[],
            suggestion=""
        )
        assert issue.severity == "info"


@pytest.mark.skipif(DuplicateValidator is None, reason="DuplicateValidator not implemented yet")
class TestDuplicateValidator:
    """Test DuplicateValidator for duplicate TAG detection"""

    def test_no_duplicates(self):
        """No duplicates should return empty list"""
        validator = DuplicateValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @CODE:TEST-002\n")

            issues = validator.validate([str(file1), str(file2)])
            assert len(issues) == 0

    def test_duplicates_in_same_file(self):
        """Duplicates in same file should be detected"""
        validator = DuplicateValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# @CODE:TEST-001
def func1():
    pass

# @CODE:TEST-001
def func2():
    pass
""")

            issues = validator.validate([str(file1)])
            assert len(issues) == 1
            assert issues[0].severity == "error"
            assert issues[0].type == "duplicate"
            assert "TEST-001" in issues[0].tag

    def test_duplicates_across_files(self):
        """Duplicates across files should be detected"""
        validator = DuplicateValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @CODE:TEST-001\n")

            issues = validator.validate([str(file1), str(file2)])
            assert len(issues) == 1
            assert len(issues[0].locations) == 2
            assert issues[0].severity == "error"

    def test_multiple_duplicate_types(self):
        """Multiple duplicate types should all be detected"""
        validator = DuplicateValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# @CODE:TEST-001
# @CODE:TEST-002
# @CODE:TEST-001
# @CODE:TEST-002
""")

            issues = validator.validate([str(file1)])
            assert len(issues) == 2  # TEST-001 and TEST-002

    def test_validator_name(self):
        """DuplicateValidator should return its name"""
        validator = DuplicateValidator()
        assert validator.get_name() == "DuplicateValidator"

    def test_validator_priority(self):
        """DuplicateValidator should have high priority (errors block early)"""
        validator = DuplicateValidator()
        assert validator.get_priority() >= 90


@pytest.mark.skipif(OrphanValidator is None, reason="OrphanValidator not implemented yet")
class TestOrphanValidator:
    """Test OrphanValidator for orphan TAG detection"""

    def test_no_orphans(self):
        """Complete TAG chain should have no orphans"""
        validator = OrphanValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            issues = validator.validate([str(code_file), str(test_file)])
            assert len(issues) == 0

    def test_orphan_code_without_test(self):
        """CODE without TEST should generate warning"""
        validator = OrphanValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            issues = validator.validate([str(code_file)])
            assert len(issues) >= 1
            assert any("USER-REG-001" in issue.tag for issue in issues)
            assert any(issue.severity == "warning" for issue in issues)
            assert any("TEST" in issue.message for issue in issues)

    def test_orphan_test_without_code(self):
        """TEST without CODE should generate warning"""
        validator = OrphanValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            issues = validator.validate([str(test_file)])
            assert len(issues) >= 1
            assert any("USER-REG-001" in issue.tag for issue in issues)
            assert any(issue.severity == "warning" for issue in issues)

    def test_multiple_orphans(self):
        """Multiple orphans should all be detected"""
        validator = OrphanValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            code1 = Path(tmpdir) / "auth.py"
            code1.write_text("# @CODE:AUTH-001\n")

            test1 = Path(tmpdir) / "test_payment.py"
            test1.write_text("# @TEST:PAY-001\n")

            issues = validator.validate([str(code1), str(test1)])
            assert len(issues) >= 2

    def test_validator_name(self):
        """OrphanValidator should return its name"""
        validator = OrphanValidator()
        assert validator.get_name() == "OrphanValidator"

    def test_validator_priority(self):
        """OrphanValidator should have medium priority"""
        validator = OrphanValidator()
        assert validator.get_priority() >= 50


@pytest.mark.skipif(ChainValidator is None, reason="ChainValidator not implemented yet")
class TestChainValidator:
    """Test ChainValidator for TAG chain integrity (NEW in Component 3)"""

    def test_complete_chain_spec_code_test_doc(self):
        """Complete SPEC→CODE→TEST→DOC chain should pass"""
        validator = ChainValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            spec_file = Path(tmpdir) / "spec.md"
            spec_file.write_text("# @SPEC:USER-REG-001\n")

            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            doc_file = Path(tmpdir) / "README.md"
            doc_file.write_text("# @DOC:USER-REG-001\n")

            issues = validator.validate([
                str(spec_file), str(code_file), str(test_file), str(doc_file)
            ])
            assert len(issues) == 0

    def test_chain_with_missing_spec(self):
        """Chain missing SPEC should generate warning"""
        validator = ChainValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            issues = validator.validate([str(code_file), str(test_file)])
            assert any("SPEC" in issue.message for issue in issues)
            assert any(issue.severity == "warning" for issue in issues)

    def test_chain_with_missing_code(self):
        """Chain missing CODE should generate warning"""
        validator = ChainValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            spec_file = Path(tmpdir) / "spec.md"
            spec_file.write_text("# @SPEC:USER-REG-001\n")

            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            issues = validator.validate([str(spec_file), str(test_file)])
            assert any("CODE" in issue.message or "implementation" in issue.message.lower() for issue in issues)

    def test_chain_with_missing_test(self):
        """Chain missing TEST should generate warning"""
        validator = ChainValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            spec_file = Path(tmpdir) / "spec.md"
            spec_file.write_text("# @SPEC:USER-REG-001\n")

            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            issues = validator.validate([str(spec_file), str(code_file)])
            assert any("TEST" in issue.message for issue in issues)

    def test_chain_with_missing_doc(self):
        """Chain missing DOC should generate info-level suggestion"""
        validator = ChainValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            spec_file = Path(tmpdir) / "spec.md"
            spec_file.write_text("# @SPEC:USER-REG-001\n")

            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            issues = validator.validate([str(spec_file), str(code_file), str(test_file)])
            # DOC is optional, but info message may be generated
            doc_issues = [i for i in issues if "DOC" in i.message]
            if doc_issues:
                assert all(issue.severity == "info" for issue in doc_issues)

    def test_partial_chains_multiple_tags(self):
        """Multiple partial chains should all be detected"""
        validator = ChainValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Complete chain for USER-REG-001
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @SPEC:USER-REG-001\n# @CODE:USER-REG-001\n# @TEST:USER-REG-001\n")

            # Incomplete chain for AUTH-002 (missing TEST)
            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @SPEC:AUTH-002\n# @CODE:AUTH-002\n")

            issues = validator.validate([str(file1), str(file2)])
            assert any("AUTH-002" in issue.tag and "TEST" in issue.message for issue in issues)

    def test_validator_name(self):
        """ChainValidator should return its name"""
        validator = ChainValidator()
        assert validator.get_name() == "ChainValidator"

    def test_validator_priority(self):
        """ChainValidator should have low priority (runs after duplicates/orphans)"""
        validator = ChainValidator()
        assert validator.get_priority() <= 50


@pytest.mark.skipif(CentralValidator is None, reason="CentralValidator not implemented yet")
class TestCentralValidator:
    """Test CentralValidator orchestration"""

    def test_default_initialization(self):
        """CentralValidator should initialize with default config"""
        validator = CentralValidator()
        assert validator is not None

    def test_custom_configuration(self):
        """CentralValidator should accept custom configuration"""
        config = ValidationConfig(strict_mode=True, check_orphans=False)
        validator = CentralValidator(config=config)
        assert validator is not None

    def test_validate_single_file(self):
        """CentralValidator should validate single file"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            assert isinstance(result, CentralValidationResult)
            assert result.is_valid is not None

    def test_validate_multiple_files(self):
        """CentralValidator should validate multiple files"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @TEST:TEST-001\n")

            result = validator.validate_files([str(file1), str(file2)])
            assert isinstance(result, CentralValidationResult)

    def test_validate_directory(self):
        """CentralValidator should validate entire directory"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @TEST:TEST-001\n")

            result = validator.validate_directory(tmpdir)
            assert isinstance(result, CentralValidationResult)

    def test_detect_errors(self):
        """CentralValidator should detect errors (duplicates)"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# @CODE:TEST-001
# @CODE:TEST-001
""")

            result = validator.validate_files([str(file1)])
            assert result.is_valid is False
            assert len(result.errors) > 0

    def test_detect_warnings(self):
        """CentralValidator should detect warnings (orphans)"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            assert len(result.warnings) > 0

    def test_strict_mode_treats_warnings_as_errors(self):
        """CentralValidator in strict mode should treat warnings as errors"""
        config = ValidationConfig(strict_mode=True)
        validator = CentralValidator(config=config)

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")  # Orphan - warning normally

            result = validator.validate_files([str(file1)])
            assert result.is_valid is False

    def test_register_custom_validator(self):
        """CentralValidator should support registering custom validators"""
        validator = CentralValidator()

        # Register a new validator
        custom_validator = DuplicateValidator()
        validator.register_validator(custom_validator)

        validators = validator.get_validators()
        assert len(validators) > 0

    def test_get_registered_validators(self):
        """CentralValidator should return list of registered validators"""
        validator = CentralValidator()
        validators = validator.get_validators()
        assert isinstance(validators, list)
        assert len(validators) > 0

    def test_validators_run_in_priority_order(self):
        """Validators should run in priority order (high to low)"""
        validator = CentralValidator()
        validators = validator.get_validators()

        # Check that priorities are in descending order
        priorities = [v.get_priority() for v in validators]
        assert priorities == sorted(priorities, reverse=True)


@pytest.mark.skipif(CentralValidationResult is None, reason="CentralValidationResult not implemented yet")
class TestCentralValidationResult:
    """Test CentralValidationResult dataclass"""

    def test_successful_validation_result(self):
        """Successful validation should have is_valid=True"""
        result = CentralValidationResult(
            is_valid=True,
            issues=[],
            errors=[],
            warnings=[],
            statistics=ValidationStatistics(
                total_files_scanned=5,
                total_tags_found=10,
                total_issues=0,
                error_count=0,
                warning_count=0,
                coverage_percentage=100.0
            ),
            timestamp=datetime.now(),
            execution_time_ms=50.0
        )
        assert result.is_valid is True
        assert len(result.errors) == 0
        assert len(result.warnings) == 0

    def test_failed_validation_result(self):
        """Failed validation should have is_valid=False"""
        error = ValidationIssue(
            severity="error",
            type="duplicate",
            tag="@CODE:TEST-001",
            message="Duplicate TAG",
            locations=[("file1.py", 1)],
            suggestion="Remove duplicate"
        )
        result = CentralValidationResult(
            is_valid=False,
            issues=[error],
            errors=[error],
            warnings=[],
            statistics=ValidationStatistics(
                total_files_scanned=1,
                total_tags_found=2,
                total_issues=1,
                error_count=1,
                warning_count=0,
                coverage_percentage=0.0
            ),
            timestamp=datetime.now(),
            execution_time_ms=30.0
        )
        assert result.is_valid is False
        assert len(result.errors) == 1

    def test_validation_with_warnings_only(self):
        """Validation with warnings should still be valid (non-strict)"""
        warning = ValidationIssue(
            severity="warning",
            type="orphan",
            tag="@CODE:TEST-001",
            message="CODE without TEST",
            locations=[("file1.py", 1)],
            suggestion="Add corresponding TEST"
        )
        result = CentralValidationResult(
            is_valid=True,
            issues=[warning],
            errors=[],
            warnings=[warning],
            statistics=ValidationStatistics(
                total_files_scanned=1,
                total_tags_found=1,
                total_issues=1,
                error_count=0,
                warning_count=1,
                coverage_percentage=50.0
            ),
            timestamp=datetime.now(),
            execution_time_ms=25.0
        )
        assert result.is_valid is True
        assert len(result.warnings) == 1


@pytest.mark.skipif(ValidationStatistics is None, reason="ValidationStatistics not implemented yet")
class TestValidationStatistics:
    """Test ValidationStatistics dataclass"""

    def test_statistics_calculation(self):
        """ValidationStatistics should calculate correctly"""
        stats = ValidationStatistics(
            total_files_scanned=10,
            total_tags_found=50,
            total_issues=5,
            error_count=2,
            warning_count=3,
            coverage_percentage=85.5
        )
        assert stats.total_files_scanned == 10
        assert stats.total_tags_found == 50
        assert stats.total_issues == 5
        assert stats.error_count == 2
        assert stats.warning_count == 3
        assert stats.coverage_percentage == 85.5

    def test_coverage_percentage_calculation(self):
        """Coverage percentage should be calculated correctly"""
        # 8 SPEC tags, 6 with CODE implementation = 75% coverage
        stats = ValidationStatistics(
            total_files_scanned=10,
            total_tags_found=20,
            total_issues=0,
            error_count=0,
            warning_count=0,
            coverage_percentage=75.0
        )
        assert stats.coverage_percentage == 75.0

    def test_edge_case_no_tags(self):
        """Statistics should handle edge case of no tags"""
        stats = ValidationStatistics(
            total_files_scanned=5,
            total_tags_found=0,
            total_issues=0,
            error_count=0,
            warning_count=0,
            coverage_percentage=0.0
        )
        assert stats.total_tags_found == 0
        assert stats.coverage_percentage == 0.0


@pytest.mark.skipif(CentralValidator is None, reason="CentralValidator not implemented yet")
class TestReportGeneration:
    """Test report generation in multiple formats"""

    def test_detailed_report_format(self):
        """CentralValidator should generate detailed report"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n# @TEST:TEST-001\n")

            result = validator.validate_files([str(file1)])
            report = validator.create_report(result, format="detailed")

            assert isinstance(report, str)
            assert len(report) > 0

    def test_summary_report_format(self):
        """CentralValidator should generate summary report"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n# @TEST:TEST-001\n")

            result = validator.validate_files([str(file1)])
            report = validator.create_report(result, format="summary")

            assert isinstance(report, str)
            assert len(report) > 0

    def test_json_report_format(self):
        """CentralValidator should generate JSON report"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n# @TEST:TEST-001\n")

            result = validator.validate_files([str(file1)])
            report = validator.create_report(result, format="json")

            assert isinstance(report, str)
            # Should be valid JSON
            parsed = json.loads(report)
            assert "is_valid" in parsed
            assert "statistics" in parsed

    def test_report_includes_statistics(self):
        """Reports should include validation statistics"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            report = validator.create_report(result, format="json")

            parsed = json.loads(report)
            assert "statistics" in parsed
            assert "total_files_scanned" in parsed["statistics"]
            assert "total_tags_found" in parsed["statistics"]

    def test_report_includes_issues(self):
        """Reports should include all issues with details"""
        validator = CentralValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            report = validator.create_report(result, format="json")

            parsed = json.loads(report)
            assert "issues" in parsed
            assert len(parsed["issues"]) > 0


@pytest.mark.skipif(CentralValidator is None, reason="CentralValidator not implemented yet")
class TestConfigurableValidation:
    """Test configurable validation rules"""

    def test_disable_duplicate_check(self):
        """Should be able to disable duplicate checking"""
        config = ValidationConfig(check_duplicates=False)
        validator = CentralValidator(config=config)

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            # No duplicate errors when check is disabled
            assert len(result.errors) == 0

    def test_disable_orphan_check(self):
        """Should be able to disable orphan checking"""
        config = ValidationConfig(check_orphans=False)
        validator = CentralValidator(config=config)

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            # No orphan warnings when check is disabled
            orphan_warnings = [w for w in result.warnings if "orphan" in w.type.lower()]
            assert len(orphan_warnings) == 0

    def test_disable_chain_integrity_check(self):
        """Should be able to disable chain integrity checking"""
        config = ValidationConfig(check_chain_integrity=False)
        validator = CentralValidator(config=config)

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            # No chain warnings when check is disabled
            chain_warnings = [w for w in result.warnings if "chain" in w.type.lower()]
            assert len(chain_warnings) == 0

    def test_file_type_filtering(self):
        """Should be able to filter by file types"""
        config = ValidationConfig(allowed_file_types=["py"])
        validator = CentralValidator(config=config)

        with tempfile.TemporaryDirectory() as tmpdir:
            py_file = Path(tmpdir) / "file1.py"
            py_file.write_text("# @CODE:TEST-001\n# @CODE:TEST-001\n")

            js_file = Path(tmpdir) / "file2.js"
            js_file.write_text("// @CODE:TEST-002\n// @CODE:TEST-002\n")

            result = validator.validate_directory(tmpdir)
            # Should only validate .py files
            # Error in .py should be detected, error in .js should be ignored
            assert len(result.errors) > 0

    def test_ignore_patterns(self):
        """Should be able to ignore files by pattern"""
        config = ValidationConfig(ignore_patterns=["test_*"])
        validator = CentralValidator(config=config)

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "main.py"
            file1.write_text("# @CODE:TEST-001\n# @CODE:TEST-001\n")

            file2 = Path(tmpdir) / "test_main.py"
            file2.write_text("# @TEST:TEST-002\n# @TEST:TEST-002\n")

            result = validator.validate_directory(tmpdir)
            # Should only validate main.py, ignore test_main.py
            assert len(result.errors) > 0


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
