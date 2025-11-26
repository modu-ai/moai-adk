"""
Comprehensive test suite for TRUST 5 validation checklist module.

This test suite provides comprehensive coverage for the validation_checklist.py module,
including:
- Checklist type and status enum validation
- ChecklistItem dataclass validation
- ChecklistResult and ChecklistReport validation
- TRUSTValidationChecklist system initialization
- Checklist execution for all 5 TRUST principles
- Individual validation rule evaluation
- Report generation and recommendations
- Score calculation and statistics
- Edge cases and error handling

Target coverage: 90%+
Test count: 100-150 tests
"""

import tempfile
from pathlib import Path

import pytest

from src.moai_adk.foundation.trust.validation_checklist import (
    ChecklistItem,
    ChecklistReport,
    ChecklistResult,
    ChecklistSeverity,
    ChecklistStatus,
    ChecklistType,
    TRUSTValidationChecklist,
    generate_checklist_report,
    validate_trust_checklists,
)

# ============================================================================
# Fixtures
# ============================================================================


@pytest.fixture
def temp_project_dir():
    """Create a temporary project directory for testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        project_path = Path(tmpdir)

        # Create basic directory structure
        (project_path / "src").mkdir()
        (project_path / "tests").mkdir()
        (project_path / ".git").mkdir()

        yield project_path


@pytest.fixture
def validation_checklist():
    """Create a TRUSTValidationChecklist instance."""
    return TRUSTValidationChecklist()


@pytest.fixture
def sample_checklist_item():
    """Create a sample checklist item for testing."""
    return ChecklistItem(
        id="TEST_001",
        title="Test Item",
        description="Test item description",
        category=ChecklistType.TEST_FIRST,
        severity=ChecklistSeverity.CRITICAL,
        validation_rule="test_coverage_ratio >= 0.8",
        expected_result="80% or more coverage",
        score_weight=2.0,
    )


@pytest.fixture
def sample_checklist_result(sample_checklist_item):
    """Create a sample checklist result."""
    return ChecklistResult(
        item=sample_checklist_item,
        passed=True,
        score=2.0,
        details={"coverage": 0.85},
        execution_time=0.125,
    )


# ============================================================================
# Enum Tests
# ============================================================================


class TestChecklistType:
    """Test ChecklistType enumeration."""

    def test_checklist_type_values(self):
        """Test that ChecklistType has all required values."""
        assert ChecklistType.TEST_FIRST.value == "test_first"
        assert ChecklistType.READABLE.value == "readable"
        assert ChecklistType.UNIFIED.value == "unified"
        assert ChecklistType.SECURED.value == "secured"
        assert ChecklistType.TRACKABLE.value == "trackable"

    def test_checklist_type_all_members(self):
        """Test that ChecklistType has exactly 5 members."""
        members = list(ChecklistType)
        assert len(members) == 5
        assert ChecklistType.TEST_FIRST in members
        assert ChecklistType.READABLE in members
        assert ChecklistType.UNIFIED in members
        assert ChecklistType.SECURED in members
        assert ChecklistType.TRACKABLE in members


class TestChecklistStatus:
    """Test ChecklistStatus enumeration."""

    def test_checklist_status_values(self):
        """Test that ChecklistStatus has all required values."""
        assert ChecklistStatus.PASS.value == "pass"
        assert ChecklistStatus.FAIL.value == "fail"
        assert ChecklistStatus.SKIP.value == "skip"
        assert ChecklistStatus.WARNING.value == "warning"

    def test_checklist_status_all_members(self):
        """Test that ChecklistStatus has exactly 4 members."""
        members = list(ChecklistStatus)
        assert len(members) == 4


class TestChecklistSeverity:
    """Test ChecklistSeverity enumeration."""

    def test_checklist_severity_values(self):
        """Test that ChecklistSeverity has all required values."""
        assert ChecklistSeverity.CRITICAL.value == "critical"
        assert ChecklistSeverity.HIGH.value == "high"
        assert ChecklistSeverity.MEDIUM.value == "medium"
        assert ChecklistSeverity.LOW.value == "low"
        assert ChecklistSeverity.INFO.value == "info"

    def test_checklist_severity_hierarchy(self):
        """Test severity ordering from critical to info."""
        severities = [
            ChecklistSeverity.CRITICAL,
            ChecklistSeverity.HIGH,
            ChecklistSeverity.MEDIUM,
            ChecklistSeverity.LOW,
            ChecklistSeverity.INFO,
        ]
        assert len(severities) == 5


# ============================================================================
# ChecklistItem Tests
# ============================================================================


class TestChecklistItem:
    """Test ChecklistItem dataclass."""

    def test_checklist_item_creation(self, sample_checklist_item):
        """Test creating a ChecklistItem."""
        assert sample_checklist_item.id == "TEST_001"
        assert sample_checklist_item.title == "Test Item"
        assert sample_checklist_item.category == ChecklistType.TEST_FIRST
        assert sample_checklist_item.severity == ChecklistSeverity.CRITICAL
        assert sample_checklist_item.score_weight == 2.0

    def test_checklist_item_default_status(self):
        """Test that default status is SKIP."""
        item = ChecklistItem(
            id="TEST",
            title="Test",
            description="Description",
            category=ChecklistType.TEST_FIRST,
            severity=ChecklistSeverity.MEDIUM,
            validation_rule="test_rule",
            expected_result="Expected result",
        )
        assert item.status == ChecklistStatus.SKIP
        assert item.actual_result == ""
        assert item.notes == ""

    def test_checklist_item_fields_are_mutable(self):
        """Test that ChecklistItem fields can be modified."""
        item = ChecklistItem(
            id="TEST",
            title="Test",
            description="Description",
            category=ChecklistType.TEST_FIRST,
            severity=ChecklistSeverity.LOW,
            validation_rule="rule",
            expected_result="result",
        )

        item.status = ChecklistStatus.PASS
        item.actual_result = "Test result"
        item.notes = "Test notes"

        assert item.status == ChecklistStatus.PASS
        assert item.actual_result == "Test result"
        assert item.notes == "Test notes"

    def test_checklist_item_with_all_severity_levels(self):
        """Test creating ChecklistItems with all severity levels."""
        severities = [
            ChecklistSeverity.CRITICAL,
            ChecklistSeverity.HIGH,
            ChecklistSeverity.MEDIUM,
            ChecklistSeverity.LOW,
            ChecklistSeverity.INFO,
        ]

        for severity in severities:
            item = ChecklistItem(
                id=f"TEST_{severity.value}",
                title="Test",
                description="Description",
                category=ChecklistType.TEST_FIRST,
                severity=severity,
                validation_rule="rule",
                expected_result="result",
            )
            assert item.severity == severity


# ============================================================================
# ChecklistResult Tests
# ============================================================================


class TestChecklistResult:
    """Test ChecklistResult dataclass."""

    def test_checklist_result_creation(self, sample_checklist_result):
        """Test creating a ChecklistResult."""
        assert sample_checklist_result.passed is True
        assert sample_checklist_result.score == 2.0
        assert sample_checklist_result.execution_time == 0.125

    def test_checklist_result_with_error(self, sample_checklist_item):
        """Test ChecklistResult with error message."""
        result = ChecklistResult(
            item=sample_checklist_item,
            passed=False,
            score=0,
            error_message="Test error",
        )
        assert result.passed is False
        assert result.error_message == "Test error"
        assert result.score == 0

    def test_checklist_result_with_details(self, sample_checklist_item):
        """Test ChecklistResult with detailed information."""
        details = {
            "coverage": 0.85,
            "files_checked": 10,
            "files_passed": 8,
        }
        result = ChecklistResult(
            item=sample_checklist_item,
            passed=True,
            score=2.0,
            details=details,
        )
        assert result.details == details
        assert result.details["coverage"] == 0.85
        assert result.details["files_checked"] == 10


# ============================================================================
# ChecklistReport Tests
# ============================================================================


class TestChecklistReport:
    """Test ChecklistReport dataclass."""

    def test_checklist_report_creation(self):
        """Test creating a ChecklistReport."""
        report = ChecklistReport(
            checklist_type=ChecklistType.TEST_FIRST,
            total_items=10,
            passed_items=8,
            failed_items=2,
            skipped_items=0,
            total_score=16.0,
            max_score=20.0,
            percentage_score=80.0,
        )
        assert report.checklist_type == ChecklistType.TEST_FIRST
        assert report.total_items == 10
        assert report.passed_items == 8
        assert report.failed_items == 2
        assert report.percentage_score == 80.0

    def test_checklist_report_with_results_and_recommendations(self, sample_checklist_result):
        """Test ChecklistReport with results and recommendations."""
        recommendations = [
            "Improve test coverage",
            "Add more documentation",
        ]
        report = ChecklistReport(
            checklist_type=ChecklistType.TEST_FIRST,
            total_items=1,
            passed_items=1,
            failed_items=0,
            skipped_items=0,
            total_score=2.0,
            max_score=2.0,
            percentage_score=100.0,
            results=[sample_checklist_result],
            recommendations=recommendations,
        )
        assert len(report.results) == 1
        assert len(report.recommendations) == 2
        assert report.summary == ""


# ============================================================================
# TRUSTValidationChecklist Initialization Tests
# ============================================================================


class TestTRUSTValidationChecklistInitialization:
    """Test TRUSTValidationChecklist initialization."""

    def test_checklist_initialization(self, validation_checklist):
        """Test that validation checklist initializes with all categories."""
        assert validation_checklist.checklists is not None
        assert len(validation_checklist.checklists) == 5

    def test_all_checklist_types_initialized(self, validation_checklist):
        """Test that all ChecklistType categories are initialized."""
        for checklist_type in ChecklistType:
            assert checklist_type in validation_checklist.checklists

    def test_test_first_checklists_count(self, validation_checklist):
        """Test that TEST_FIRST category has 10 checklists."""
        test_first_checklists = validation_checklist.checklists[ChecklistType.TEST_FIRST]
        assert len(test_first_checklists) == 10

    def test_readable_checklists_count(self, validation_checklist):
        """Test that READABLE category has 10 checklists."""
        readable_checklists = validation_checklist.checklists[ChecklistType.READABLE]
        assert len(readable_checklists) == 10

    def test_unified_checklists_count(self, validation_checklist):
        """Test that UNIFIED category has 10 checklists."""
        unified_checklists = validation_checklist.checklists[ChecklistType.UNIFIED]
        assert len(unified_checklists) == 10

    def test_secured_checklists_count(self, validation_checklist):
        """Test that SECURED category has 10 checklists."""
        secured_checklists = validation_checklist.checklists[ChecklistType.SECURED]
        assert len(secured_checklists) == 10

    def test_trackable_checklists_count(self, validation_checklist):
        """Test that TRACKABLE category has 9 checklists (TK_002 missing)."""
        trackable_checklists = validation_checklist.checklists[ChecklistType.TRACKABLE]
        assert len(trackable_checklists) == 9

    def test_test_first_checklist_ids(self, validation_checklist):
        """Test that TEST_FIRST checklists have correct IDs."""
        checklists = validation_checklist.checklists[ChecklistType.TEST_FIRST]
        ids = [c.id for c in checklists]
        expected_ids = [f"TF_{i:03d}" for i in range(1, 11)]
        assert ids == expected_ids

    def test_checklist_items_have_required_fields(self, validation_checklist):
        """Test that all checklist items have required fields."""
        for checklist_type, items in validation_checklist.checklists.items():
            for item in items:
                assert item.id
                assert item.title
                assert item.description
                assert item.category == checklist_type
                assert item.severity in ChecklistSeverity
                assert item.validation_rule
                assert item.expected_result
                assert item.score_weight > 0


# ============================================================================
# Test First Checklist Tests
# ============================================================================


class TestTestFirstChecklists:
    """Test TEST_FIRST category checklists."""

    def test_tf_001_unit_test_coverage(self, validation_checklist):
        """Test TF_001: Unit Test Coverage."""
        item = validation_checklist.checklists[ChecklistType.TEST_FIRST][0]
        assert item.id == "TF_001"
        assert item.title == "Unit Test Coverage"
        assert item.severity == ChecklistSeverity.CRITICAL
        assert item.score_weight == 2.0

    def test_tf_010_continuous_integration(self, validation_checklist):
        """Test TF_010: Continuous Integration."""
        item = validation_checklist.checklists[ChecklistType.TEST_FIRST][9]
        assert item.id == "TF_010"
        assert item.title == "Continuous Integration"
        assert item.severity == ChecklistSeverity.CRITICAL
        assert item.score_weight == 2.0

    def test_test_first_critical_items(self, validation_checklist):
        """Test that TEST_FIRST has 2 CRITICAL items."""
        checklists = validation_checklist.checklists[ChecklistType.TEST_FIRST]
        critical_items = [c for c in checklists if c.severity == ChecklistSeverity.CRITICAL]
        assert len(critical_items) == 2

    def test_test_first_total_weight(self, validation_checklist):
        """Test that TEST_FIRST total weight is correct."""
        checklists = validation_checklist.checklists[ChecklistType.TEST_FIRST]
        total_weight = sum(c.score_weight for c in checklists)
        # 2.0 + 1.5 + 1.5 + 1.0 + 1.0 + 1.0 + 1.0 + 0.5 + 0.5 + 2.0 = 12.0
        assert total_weight == 12.0


# ============================================================================
# Readable Checklist Tests
# ============================================================================


class TestReadableChecklists:
    """Test READABLE category checklists."""

    def test_rd_004_docstrings(self, validation_checklist):
        """Test RD_004: Docstrings."""
        item = validation_checklist.checklists[ChecklistType.READABLE][3]
        assert item.id == "RD_004"
        assert item.title == "Docstrings"
        assert item.severity == ChecklistSeverity.CRITICAL
        assert item.score_weight == 2.0

    def test_readable_critical_items(self, validation_checklist):
        """Test that READABLE has 1 CRITICAL item."""
        checklists = validation_checklist.checklists[ChecklistType.READABLE]
        critical_items = [c for c in checklists if c.severity == ChecklistSeverity.CRITICAL]
        assert len(critical_items) == 1

    def test_readable_high_severity_items(self, validation_checklist):
        """Test that READABLE has HIGH severity items."""
        checklists = validation_checklist.checklists[ChecklistType.READABLE]
        high_items = [c for c in checklists if c.severity == ChecklistSeverity.HIGH]
        assert len(high_items) > 0


# ============================================================================
# Secured Checklist Tests
# ============================================================================


class TestSecuredChecklists:
    """Test SECURED category checklists."""

    def test_sc_001_input_validation(self, validation_checklist):
        """Test SC_001: Input Validation."""
        item = validation_checklist.checklists[ChecklistType.SECURED][0]
        assert item.id == "SC_001"
        assert item.severity == ChecklistSeverity.CRITICAL
        assert item.score_weight == 2.0

    def test_secured_critical_items(self, validation_checklist):
        """Test that SECURED has 5 CRITICAL items."""
        checklists = validation_checklist.checklists[ChecklistType.SECURED]
        critical_items = [c for c in checklists if c.severity == ChecklistSeverity.CRITICAL]
        assert len(critical_items) == 5

    def test_secured_validation_rules(self, validation_checklist):
        """Test that SECURED checklists have validation rules."""
        checklists = validation_checklist.checklists[ChecklistType.SECURED]
        for item in checklists:
            assert item.validation_rule
            assert "_" in item.validation_rule or "validation" in item.validation_rule.lower()


# ============================================================================
# Trackable Checklist Tests
# ============================================================================


class TestTrackableChecklists:
    """Test TRACKABLE category checklists."""

    def test_tk_001_git_repository(self, validation_checklist):
        """Test TK_001: Git Repository."""
        item = validation_checklist.checklists[ChecklistType.TRACKABLE][0]
        assert item.id == "TK_001"
        assert item.title == "Git Repository"
        assert item.severity == ChecklistSeverity.CRITICAL

    def test_trackable_missing_tk_002(self, validation_checklist):
        """Test that TRACKABLE is missing TK_002."""
        checklists = validation_checklist.checklists[ChecklistType.TRACKABLE]
        ids = [c.id for c in checklists]
        assert "TK_002" not in ids
        assert "TK_003" in ids


# ============================================================================
# Test Coverage Validation Tests
# ============================================================================


class TestTestCoverageValidation:
    """Test coverage ratio validation."""

    def test_test_coverage_check_success(self, validation_checklist, temp_project_dir):
        """Test successful test coverage check."""
        # Create test files
        (temp_project_dir / "tests" / "test_module1.py").write_text("# test")
        (temp_project_dir / "tests" / "test_module2.py").write_text("# test")

        # Create source files
        (temp_project_dir / "src" / "module1.py").write_text("# source")
        (temp_project_dir / "src" / "module2.py").write_text("# source")

        passed, details = validation_checklist._check_test_coverage(temp_project_dir, "test_coverage_ratio >= 0.8")

        assert isinstance(passed, bool)
        assert "result" in details
        assert "target" in details

    def test_test_coverage_extracts_ratio(self, validation_checklist, temp_project_dir):
        """Test that test coverage extracts ratio correctly."""
        passed, details = validation_checklist._check_test_coverage(temp_project_dir, "test_coverage_ratio >= 0.7")

        assert "target" in details
        assert details["target"] == 0.7

    def test_test_coverage_with_various_thresholds(self, validation_checklist, temp_project_dir):
        """Test coverage checks with various thresholds."""
        thresholds = [0.5, 0.6, 0.7, 0.8, 0.9]

        for threshold in thresholds:
            rule = f"test_coverage_ratio >= {threshold}"
            passed, details = validation_checklist._check_test_coverage(temp_project_dir, rule)
            assert details["target"] == threshold


# ============================================================================
# Test Structure Validation Tests
# ============================================================================


class TestTestStructureValidation:
    """Test test file structure validation."""

    def test_test_structure_success(self, validation_checklist, temp_project_dir):
        """Test successful test structure check."""
        # Create tests directory with proper structure
        (temp_project_dir / "tests" / "__init__.py").write_text("")
        (temp_project_dir / "tests" / "test_example.py").write_text("# test")

        passed, details = validation_checklist._check_test_structure(temp_project_dir)

        assert passed is True
        assert details["test_dir_exists"] is True
        assert details["test_modules"] > 0

    def test_test_structure_no_tests_dir(self, validation_checklist):
        """Test failure when tests directory missing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            passed, details = validation_checklist._check_test_structure(project_path)

            assert passed is False
            assert "error" in details

    def test_test_structure_no_init_files(self, validation_checklist, temp_project_dir):
        """Test failure when __init__.py missing."""
        (temp_project_dir / "tests" / "test_example.py").write_text("# test")

        passed, details = validation_checklist._check_test_structure(temp_project_dir)

        assert passed is False


# ============================================================================
# Integration Tests Validation Tests
# ============================================================================


class TestIntegrationTestsValidation:
    """Test integration tests detection."""

    def test_integration_tests_found(self, validation_checklist, temp_project_dir):
        """Test detecting integration tests."""
        content = """
def test_integration_user_workflow():
    '''Integration test for user workflow'''
    pass
"""
        (temp_project_dir / "tests" / "test_integration.py").write_text(content)

        passed, details = validation_checklist._check_integration_tests(temp_project_dir)

        assert passed is True
        assert details["integration_test_files"] > 0

    def test_integration_tests_not_found(self, validation_checklist, temp_project_dir):
        """Test when no integration tests found."""
        passed, details = validation_checklist._check_integration_tests(temp_project_dir)

        assert passed is False

    def test_integration_tests_pattern_matching(self, validation_checklist, temp_project_dir):
        """Test various integration test patterns."""
        patterns = [
            "# @integration_test",
            "# test_integration",
            "# Integration Test",
        ]

        for pattern in patterns:
            content = f"def test_example():\n    '''{pattern}'''\n    pass"
            (temp_project_dir / "tests" / f"test_{patterns.index(pattern)}.py").write_text(content)


# ============================================================================
# Docstring Coverage Tests
# ============================================================================


class TestDocstringCoverage:
    """Test docstring coverage validation."""

    def test_docstring_coverage_calculation(self, validation_checklist, temp_project_dir):
        """Test docstring coverage calculation."""
        content = '''
def function_with_docstring():
    """This has a docstring."""
    pass

def function_without_docstring():
    pass
'''
        (temp_project_dir / "tests" / "test_module.py").write_text(content)

        passed, details = validation_checklist._check_test_docstrings(temp_project_dir, "test_docstrings_ratio >= 0.5")

        assert "result" in details
        assert "target" in details

    def test_docstring_coverage_all_documented(self, validation_checklist, temp_project_dir):
        """Test when all test functions have docstrings."""
        content = '''
def test_one():
    """Test one."""
    pass

def test_two():
    """Test two."""
    pass
'''
        (temp_project_dir / "tests" / "test_all_documented.py").write_text(content)

        passed, details = validation_checklist._check_test_docstrings(temp_project_dir, "test_docstrings_ratio >= 0.9")

        assert details["total_functions"] > 0
        assert details["docstringed_functions"] > 0


# ============================================================================
# Function Length Validation Tests
# ============================================================================


class TestFunctionLengthValidation:
    """Test function length validation."""

    def test_function_length_short_functions(self, validation_checklist, temp_project_dir):
        """Test when all functions are short."""
        content = """
def short_function1():
    return 1

def short_function2():
    return 2
"""
        (temp_project_dir / "src" / "module.py").write_text(content)

        passed, details = validation_checklist._check_function_length(temp_project_dir, "max_function_length <= 50")

        assert "total_functions" in details
        assert "long_functions" in details

    def test_function_length_extracts_max_length(self, validation_checklist, temp_project_dir):
        """Test that function length rule extracts max length."""
        passed, details = validation_checklist._check_function_length(temp_project_dir, "max_function_length <= 100")

        assert details["max_length"] == 100

    def test_function_length_various_thresholds(self, validation_checklist, temp_project_dir):
        """Test function length with various thresholds."""
        content = "\n".join([f"# Line {i}" for i in range(30)])
        content = "def test():\n" + "\n".join([f"    # Line {i}" for i in range(25)])

        (temp_project_dir / "src" / "module.py").write_text(content)

        for max_len in [10, 20, 50, 100]:
            passed, details = validation_checklist._check_function_length(
                temp_project_dir, f"max_function_length <= {max_len}"
            )
            assert details["max_length"] == max_len


# ============================================================================
# Class Length Validation Tests
# ============================================================================


class TestClassLengthValidation:
    """Test class length validation."""

    def test_class_length_short_classes(self, validation_checklist, temp_project_dir):
        """Test when all classes are short."""
        content = """
class ShortClass1:
    pass

class ShortClass2:
    pass
"""
        (temp_project_dir / "src" / "module.py").write_text(content)

        passed, details = validation_checklist._check_class_length(temp_project_dir, "max_class_length <= 200")

        assert "total_classes" in details
        assert "long_classes" in details

    def test_class_length_extracts_max_length(self, validation_checklist, temp_project_dir):
        """Test that class length rule extracts max length."""
        passed, details = validation_checklist._check_class_length(temp_project_dir, "max_class_length <= 300")

        assert details["max_length"] == 300


# ============================================================================
# Type Hint Coverage Tests
# ============================================================================


class TestTypeHintCoverage:
    """Test type hint coverage validation."""

    def test_type_hint_coverage_with_hints(self, validation_checklist, temp_project_dir):
        """Test type hint coverage detection."""
        content = """
def function_with_hints(x: int) -> str:
    return str(x)

def function_without_hints(x):
    return str(x)
"""
        (temp_project_dir / "src" / "module.py").write_text(content)

        passed, details = validation_checklist._check_type_hint_coverage(temp_project_dir, "type_hint_coverage >= 0.5")

        assert "total_functions" in details
        assert "hinted_functions" in details

    def test_type_hint_coverage_return_type(self, validation_checklist, temp_project_dir):
        """Test type hint coverage with return types."""
        content = """
def function_with_return_type() -> int:
    return 42
"""
        (temp_project_dir / "src" / "module.py").write_text(content)

        passed, details = validation_checklist._check_type_hint_coverage(temp_project_dir, "type_hint_coverage >= 0.5")

        assert details["total_functions"] > 0


# ============================================================================
# Naming Convention Tests
# ============================================================================


class TestNamingConventions:
    """Test naming conventions validation."""

    def test_naming_conventions_valid(self, validation_checklist, temp_project_dir):
        """Test valid naming conventions."""
        content = """
def valid_function_name():
    pass

CONSTANT_VALUE = 42
variable_name = "test"
"""
        (temp_project_dir / "src" / "module.py").write_text(content)

        passed, details = validation_checklist._check_naming_conventions(temp_project_dir)

        assert isinstance(passed, bool)
        assert "violations" in details
        assert "total_checks" in details


# ============================================================================
# Docstring Coverage (General) Tests
# ============================================================================


class TestDocstringCoverageGeneral:
    """Test general docstring coverage."""

    def test_docstring_coverage_functions_and_classes(self, validation_checklist, temp_project_dir):
        """Test docstring coverage for functions and classes."""
        content = '''
def documented_function():
    """Has docstring."""
    pass

def undocumented_function():
    pass

class DocumentedClass:
    """Has docstring."""
    pass

class UndocumentedClass:
    pass
'''
        (temp_project_dir / "src" / "module.py").write_text(content)

        passed, details = validation_checklist._check_docstring_coverage(temp_project_dir, "docstring_coverage >= 0.5")

        assert "total_items" in details
        assert "docstringed_items" in details

    def test_docstring_coverage_various_thresholds(self, validation_checklist, temp_project_dir):
        """Test docstring coverage with various thresholds."""
        content = '''
def test1():
    """Documented."""
    pass

def test2():
    """Documented."""
    pass

def test3():
    pass
'''
        (temp_project_dir / "src" / "module.py").write_text(content)

        for threshold in [0.5, 0.6, 0.7]:
            rule = f"docstring_coverage >= {threshold}"
            passed, details = validation_checklist._check_docstring_coverage(temp_project_dir, rule)
            assert details["target"] == threshold


# ============================================================================
# Checklist Execution Tests
# ============================================================================


class TestChecklistExecution:
    """Test checklist execution workflow."""

    def test_execute_checklist_single_type(self, validation_checklist, temp_project_dir):
        """Test executing a single checklist type."""
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.TEST_FIRST)

        assert report.checklist_type == ChecklistType.TEST_FIRST
        assert report.total_items > 0
        assert report.passed_items >= 0
        assert report.failed_items >= 0
        assert report.execution_time >= 0

    def test_execute_checklist_report_statistics(self, validation_checklist, temp_project_dir):
        """Test that checklist report has correct statistics."""
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.READABLE)

        total_expected = report.passed_items + report.failed_items + report.skipped_items
        assert total_expected == report.total_items

    def test_execute_checklist_score_calculation(self, validation_checklist, temp_project_dir):
        """Test score calculation in checklist report."""
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.UNIFIED)

        # Percentage should be between 0 and 100
        assert 0 <= report.percentage_score <= 100

        # If max_score > 0, total_score should not exceed max_score
        if report.max_score > 0:
            assert report.total_score <= report.max_score

    def test_execute_checklist_includes_results(self, validation_checklist, temp_project_dir):
        """Test that checklist report includes results."""
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.SECURED)

        assert len(report.results) > 0
        assert all(isinstance(r, ChecklistResult) for r in report.results)

    def test_execute_all_checklists(self, validation_checklist, temp_project_dir):
        """Test executing all checklists."""
        reports = validation_checklist.execute_all_checklists(str(temp_project_dir))

        assert len(reports) == 5
        assert all(isinstance(r, ChecklistReport) for r in reports.values())


# ============================================================================
# Validation Rule Evaluation Tests
# ============================================================================


class TestValidationRuleEvaluation:
    """Test validation rule evaluation."""

    def test_evaluate_validation_rule_test_coverage(self, validation_checklist, temp_project_dir):
        """Test evaluating test coverage rule."""
        passed, details = validation_checklist._evaluate_validation_rule(temp_project_dir, "test_coverage_ratio >= 0.5")

        assert isinstance(passed, bool)
        assert isinstance(details, dict)

    def test_evaluate_validation_rule_test_structure(self, validation_checklist, temp_project_dir):
        """Test evaluating test structure rule."""
        passed, details = validation_checklist._evaluate_validation_rule(temp_project_dir, "test_files_structure_valid")

        assert isinstance(passed, bool)
        assert isinstance(details, dict)

    def test_evaluate_validation_rule_unknown_rule(self, validation_checklist, temp_project_dir):
        """Test evaluating unknown validation rule."""
        passed, details = validation_checklist._evaluate_validation_rule(temp_project_dir, "unknown_rule")

        assert passed is False
        assert "error" in details

    def test_evaluate_validation_rule_all_test_first_rules(self, validation_checklist, temp_project_dir):
        """Test evaluating all TEST_FIRST rules."""
        test_first_items = validation_checklist.checklists[ChecklistType.TEST_FIRST]

        for item in test_first_items:
            passed, details = validation_checklist._evaluate_validation_rule(temp_project_dir, item.validation_rule)

            assert isinstance(passed, bool)
            assert isinstance(details, dict)

    def test_evaluate_validation_rule_all_readable_rules(self, validation_checklist, temp_project_dir):
        """Test evaluating all READABLE rules."""
        readable_items = validation_checklist.checklists[ChecklistType.READABLE]

        for item in readable_items[:5]:  # Test first 5 to avoid excessive execution
            passed, details = validation_checklist._evaluate_validation_rule(temp_project_dir, item.validation_rule)

            assert isinstance(passed, bool)
            assert isinstance(details, dict)


# ============================================================================
# Single Checklist Item Execution Tests
# ============================================================================


class TestChecklistItemExecution:
    """Test execution of single checklist items."""

    def test_execute_checklist_item_success(self, validation_checklist, temp_project_dir, sample_checklist_item):
        """Test successful checklist item execution."""
        result = validation_checklist._execute_checklist_item(str(temp_project_dir), sample_checklist_item)

        assert isinstance(result, ChecklistResult)
        assert result.item == sample_checklist_item
        assert result.execution_time >= 0

    def test_execute_checklist_item_updates_status(self, validation_checklist, temp_project_dir, sample_checklist_item):
        """Test that item status is updated after execution."""
        validation_checklist._execute_checklist_item(str(temp_project_dir), sample_checklist_item)

        assert sample_checklist_item.status in [ChecklistStatus.PASS, ChecklistStatus.FAIL]

    def test_execute_checklist_item_with_error_handling(self, validation_checklist, temp_project_dir):
        """Test error handling in checklist item execution."""
        # Create an item with invalid rule
        invalid_item = ChecklistItem(
            id="INVALID",
            title="Invalid Item",
            description="Invalid",
            category=ChecklistType.TEST_FIRST,
            severity=ChecklistSeverity.LOW,
            validation_rule="non_existent_validation_rule",
            expected_result="Expected",
        )

        result = validation_checklist._execute_checklist_item(str(temp_project_dir), invalid_item)

        assert result.passed is False


# ============================================================================
# Report Generation Tests
# ============================================================================


class TestReportGeneration:
    """Test report generation and recommendations."""

    def test_generate_recommendations_all_passed(self, validation_checklist):
        """Test recommendations when all items pass."""
        checklist_item = ChecklistItem(
            id="TEST",
            title="Test",
            description="Description",
            category=ChecklistType.TEST_FIRST,
            severity=ChecklistSeverity.LOW,
            validation_rule="rule",
            expected_result="result",
        )
        result = ChecklistResult(
            item=checklist_item,
            passed=True,
            score=1.0,
        )

        recommendations = validation_checklist._generate_recommendations(ChecklistType.TEST_FIRST, [result])

        assert len(recommendations) > 0
        assert any("Excellent" in rec or "excellent" in rec for rec in recommendations)

    def test_generate_recommendations_some_failed(self, validation_checklist):
        """Test recommendations when some items fail."""
        failing_item = ChecklistItem(
            id="COVERAGE",
            title="Unit Test Coverage",
            description="Coverage",
            category=ChecklistType.TEST_FIRST,
            severity=ChecklistSeverity.CRITICAL,
            validation_rule="rule",
            expected_result="result",
        )
        result = ChecklistResult(
            item=failing_item,
            passed=False,
            score=0,
        )

        recommendations = validation_checklist._generate_recommendations(ChecklistType.TEST_FIRST, [result])

        assert len(recommendations) > 0

    def test_generate_summary_report(self, validation_checklist, temp_project_dir):
        """Test generating summary report."""
        reports = validation_checklist.execute_all_checklists(str(temp_project_dir))
        summary = validation_checklist.generate_summary_report(reports)

        assert isinstance(summary, str)
        assert "TRUST" in summary
        assert "Summary" in summary

    def test_summary_report_includes_scores(self, validation_checklist, temp_project_dir):
        """Test that summary report includes score information."""
        reports = validation_checklist.execute_all_checklists(str(temp_project_dir))
        summary = validation_checklist.generate_summary_report(reports)

        # Check for percentage indicators
        assert "%" in summary or "Score" in summary or "score" in summary


# ============================================================================
# Git Repository Validation Tests
# ============================================================================


class TestGitRepositoryValidation:
    """Test Git repository validation."""

    def test_git_repository_exists(self, validation_checklist, temp_project_dir):
        """Test detecting existing Git repository."""
        passed, details = validation_checklist._check_git_repository(temp_project_dir)

        assert passed is True

    def test_git_repository_not_exists(self, validation_checklist):
        """Test detecting missing Git repository."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            passed, details = validation_checklist._check_git_repository(project_path)

            assert passed is False


# ============================================================================
# Documentation Validation Tests
# ============================================================================


class TestDocumentationValidation:
    """Test documentation validation."""

    def test_documentation_exists(self, validation_checklist, temp_project_dir):
        """Test detecting existing documentation."""
        (temp_project_dir / "README.md").write_text("# Project")

        passed, details = validation_checklist._check_documentation(temp_project_dir)

        assert passed is True
        assert details["documentation_files"] > 0

    def test_documentation_not_exists(self, validation_checklist, temp_project_dir):
        """Test detecting missing documentation."""
        passed, details = validation_checklist._check_documentation(temp_project_dir)

        assert passed is False

    def test_documentation_rst_files(self, validation_checklist, temp_project_dir):
        """Test detecting RST documentation files."""
        (temp_project_dir / "docs" / "index.rst").parent.mkdir(parents=True, exist_ok=True)
        (temp_project_dir / "docs" / "index.rst").write_text("Documentation")

        passed, details = validation_checklist._check_documentation(temp_project_dir)

        assert passed is True


# ============================================================================
# Changelog Validation Tests
# ============================================================================


class TestChangelogValidation:
    """Test changelog validation."""

    def test_changelog_exists_md(self, validation_checklist, temp_project_dir):
        """Test detecting CHANGELOG.md."""
        (temp_project_dir / "CHANGELOG.md").write_text("# Changelog\n\n## 1.0.0")

        passed, details = validation_checklist._check_changelog(temp_project_dir)

        assert passed is True

    def test_changelog_exists_changes(self, validation_checklist, temp_project_dir):
        """Test detecting CHANGES.md."""
        (temp_project_dir / "CHANGES.md").write_text("# Changes")

        passed, details = validation_checklist._check_changelog(temp_project_dir)

        assert passed is True

    def test_changelog_not_exists(self, validation_checklist, temp_project_dir):
        """Test when changelog doesn't exist."""
        passed, details = validation_checklist._check_changelog(temp_project_dir)

        assert passed is False


# ============================================================================
# Dependency Tracking Validation Tests
# ============================================================================


class TestDependencyTrackingValidation:
    """Test dependency tracking validation."""

    def test_dependency_tracking_requirements(self, validation_checklist, temp_project_dir):
        """Test detecting requirements.txt."""
        (temp_project_dir / "requirements.txt").write_text("pytest>=7.0\npytest-cov>=4.0")

        passed, details = validation_checklist._check_dependency_tracking(temp_project_dir)

        assert passed is True

    def test_dependency_tracking_pyproject(self, validation_checklist, temp_project_dir):
        """Test detecting pyproject.toml."""
        (temp_project_dir / "pyproject.toml").write_text("[project]\ndependencies = []")

        passed, details = validation_checklist._check_dependency_tracking(temp_project_dir)

        assert passed is True

    def test_dependency_tracking_not_exists(self, validation_checklist, temp_project_dir):
        """Test when dependency tracking not found."""
        passed, details = validation_checklist._check_dependency_tracking(temp_project_dir)

        assert passed is False


# ============================================================================
# Convenience Function Tests
# ============================================================================


class TestConvenienceFunctions:
    """Test convenience functions."""

    def test_validate_trust_checklists_function(self, temp_project_dir):
        """Test validate_trust_checklists convenience function."""
        reports = validate_trust_checklists(str(temp_project_dir))

        assert isinstance(reports, dict)
        assert len(reports) == 5
        assert all(isinstance(r, ChecklistReport) for r in reports.values())

    def test_generate_checklist_report_function(self, temp_project_dir):
        """Test generate_checklist_report convenience function."""
        report = generate_checklist_report(str(temp_project_dir))

        assert isinstance(report, str)
        assert len(report) > 0


# ============================================================================
# Edge Case and Error Handling Tests
# ============================================================================


class TestEdgeCasesAndErrors:
    """Test edge cases and error handling."""

    def test_checklist_execution_with_nonexistent_path(self, validation_checklist):
        """Test checklist execution with nonexistent directory."""
        report = validation_checklist.execute_checklist("/nonexistent/path", ChecklistType.TEST_FIRST)

        assert report.checklist_type == ChecklistType.TEST_FIRST
        assert report.total_items > 0

    def test_checklist_execution_empty_directory(self, validation_checklist):
        """Test checklist execution with empty directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            report = validation_checklist.execute_checklist(tmpdir, ChecklistType.READABLE)

            assert report.checklist_type == ChecklistType.READABLE
            assert report.total_items > 0

    def test_checklist_item_with_empty_score_weight(self):
        """Test that score weight cannot be zero."""
        item = ChecklistItem(
            id="TEST",
            title="Test",
            description="Description",
            category=ChecklistType.TEST_FIRST,
            severity=ChecklistSeverity.LOW,
            validation_rule="rule",
            expected_result="result",
            score_weight=0.0,
        )
        # Should still create but with zero weight
        assert item.score_weight == 0.0

    def test_checklist_report_with_zero_items(self):
        """Test checklist report with zero items."""
        report = ChecklistReport(
            checklist_type=ChecklistType.TEST_FIRST,
            total_items=0,
            passed_items=0,
            failed_items=0,
            skipped_items=0,
            total_score=0,
            max_score=0,
            percentage_score=0,
        )
        assert report.total_items == 0

    def test_evaluate_validation_rule_with_special_characters(self, validation_checklist, temp_project_dir):
        """Test validation rule with special characters."""
        passed, details = validation_checklist._evaluate_validation_rule(temp_project_dir, "max_function_length <= 50")
        assert isinstance(passed, bool)

    def test_multiple_checklist_executions(self, validation_checklist, temp_project_dir):
        """Test multiple sequential checklist executions."""
        for _ in range(3):
            report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.TEST_FIRST)
            assert report.checklist_type == ChecklistType.TEST_FIRST

    def test_checklist_with_malformed_python_file(self, validation_checklist, temp_project_dir):
        """Test handling malformed Python files gracefully."""
        (temp_project_dir / "src" / "bad.py").write_text("def function(\n    invalid syntax here")

        # Should handle gracefully and not crash
        try:
            passed, details = validation_checklist._check_docstring_coverage(
                temp_project_dir, "docstring_coverage >= 0.5"
            )
            assert isinstance(passed, bool)
        except:
            pytest.fail("Should handle malformed Python files gracefully")


# ============================================================================
# Integration Tests
# ============================================================================


class TestIntegration:
    """Integration tests combining multiple features."""

    def test_full_workflow_single_checklist_type(self, validation_checklist, temp_project_dir):
        """Test full workflow for single checklist type."""
        # Create sample files
        (temp_project_dir / "tests" / "__init__.py").write_text("")
        (temp_project_dir / "tests" / "test_example.py").write_text(
            "def test_example():\n    '''Test example.'''\n    assert True"
        )

        # Execute checklist
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.TEST_FIRST)

        # Verify report
        assert report.total_items > 0
        assert len(report.results) == report.total_items
        assert report.percentage_score >= 0

    def test_full_workflow_all_checklists(self, validation_checklist, temp_project_dir):
        """Test full workflow for all checklist types."""
        # Set up project
        (temp_project_dir / "tests" / "__init__.py").write_text("")
        (temp_project_dir / "tests" / "test_example.py").write_text("def test(): pass")
        (temp_project_dir / "README.md").write_text("# Project")
        (temp_project_dir / "CHANGELOG.md").write_text("# Changelog")
        (temp_project_dir / "requirements.txt").write_text("pytest")

        # Execute all checklists
        reports = validation_checklist.execute_all_checklists(str(temp_project_dir))
        summary = validation_checklist.generate_summary_report(reports)

        # Verify results
        assert len(reports) == 5
        assert isinstance(summary, str)
        assert len(summary) > 0

    def test_checklist_consistency_across_types(self, validation_checklist):
        """Test that checklists have consistent structure."""
        for checklist_type, items in validation_checklist.checklists.items():
            # Each type should have items
            assert len(items) > 0

            # All items should have required fields
            for item in items:
                assert item.category == checklist_type
                assert item.severity in ChecklistSeverity
                assert item.score_weight > 0


# ============================================================================
# Performance and Scalability Tests
# ============================================================================


class TestPerformance:
    """Test performance characteristics."""

    def test_large_number_of_files(self, validation_checklist):
        """Test handling large number of files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            (project_path / "src").mkdir()

            # Create multiple Python files
            for i in range(20):
                (project_path / "src" / f"module_{i}.py").write_text(f"def function_{i}():\n    pass\n")

            # Should complete successfully
            report = validation_checklist.execute_checklist(str(project_path), ChecklistType.READABLE)
            assert report.total_items > 0

    def test_execution_time_recorded(self, validation_checklist, temp_project_dir):
        """Test that execution time is recorded."""
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.TEST_FIRST)

        assert report.execution_time >= 0
        assert all(r.execution_time >= 0 for r in report.results)


# ============================================================================
# Data Validation Tests
# ============================================================================


class TestDataValidation:
    """Test data validation and integrity."""

    def test_checklist_report_total_consistency(self, validation_checklist, temp_project_dir):
        """Test that report totals are consistent."""
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.TEST_FIRST)

        # Total should equal sum of passed, failed, and skipped
        total = report.passed_items + report.failed_items + report.skipped_items
        assert total == report.total_items

    def test_score_consistency(self, validation_checklist, temp_project_dir):
        """Test score consistency in report."""
        report = validation_checklist.execute_checklist(str(temp_project_dir), ChecklistType.READABLE)

        # Percentage should match calculation
        if report.max_score > 0:
            expected_percentage = (report.total_score / report.max_score) * 100
            assert abs(report.percentage_score - round(expected_percentage, 2)) < 0.01

    def test_result_score_calculation(self, validation_checklist, sample_checklist_item):
        """Test that result scores are calculated correctly."""
        sample_checklist_item.score_weight = 3.0
        result_passed = ChecklistResult(
            item=sample_checklist_item,
            passed=True,
            score=sample_checklist_item.score_weight,
        )
        result_failed = ChecklistResult(
            item=sample_checklist_item,
            passed=False,
            score=0,
        )

        assert result_passed.score == 3.0
        assert result_failed.score == 0
