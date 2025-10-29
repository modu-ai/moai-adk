#!/usr/bin/env python3
# @TEST:DOC-TAG-004 | Component 1: Pre-commit validator tests
"""Test suite for pre-commit TAG validation

This module tests the pre-commit validator that checks:
- TAG format validation (@DOC:DOMAIN-TYPE-NNN)
- Duplicate TAG detection
- Orphan TAG detection
- File scanning and validation

Following TDD RED-GREEN-REFACTOR cycle.
"""

import tempfile
from pathlib import Path
from typing import List
import pytest

from moai_adk.core.tags.pre_commit_validator import (
    PreCommitValidator,
    ValidationResult,
    ValidationError,
    ValidationWarning,
)


class TestTagFormatValidation:
    """Test TAG format validation"""

    def test_valid_tag_format(self):
        """Valid TAG format should pass validation"""
        validator = PreCommitValidator()

        # Valid formats
        assert validator.validate_format("@DOC:SPEC-001") is True
        assert validator.validate_format("@CODE:AUTH-API-001") is True
        assert validator.validate_format("@TEST:DB-QUERY-099") is True
        assert validator.validate_format("@SPEC:USER-REG-001") is True

    def test_invalid_tag_format(self):
        """Invalid TAG format should fail validation"""
        validator = PreCommitValidator()

        # Missing @
        assert validator.validate_format("DOC:SPEC-001") is False

        # Missing colon
        assert validator.validate_format("@DOC-SPEC-001") is False

        # Invalid prefix
        assert validator.validate_format("@INVALID:SPEC-001") is False

        # Missing domain
        assert validator.validate_format("@DOC:001") is False

        # Invalid number format
        assert validator.validate_format("@DOC:SPEC-ABC") is False

    def test_tag_pattern_extraction(self):
        """Extract TAG pattern from content"""
        validator = PreCommitValidator()

        content = """
        # @CODE:AUTH-API-001
        def login():
            pass

        # @CODE:AUTH-API-002
        def logout():
            pass
        """

        tags = validator.extract_tags(content)
        assert len(tags) == 2
        assert "@CODE:AUTH-API-001" in tags
        assert "@CODE:AUTH-API-002" in tags


class TestDuplicateDetection:
    """Test duplicate TAG detection"""

    def test_no_duplicates(self):
        """No duplicates should return empty list"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test files
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @CODE:TEST-002\n")

            errors = validator.validate_duplicates([str(file1), str(file2)])
            assert len(errors) == 0

    def test_duplicate_in_same_file(self):
        """Duplicate in same file should be detected"""
        validator = PreCommitValidator()

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

            errors = validator.validate_duplicates([str(file1)])
            assert len(errors) == 1
            assert "TEST-001" in errors[0].tag
            assert "duplicate" in errors[0].message.lower()

    def test_duplicate_across_files(self):
        """Duplicate across files should be detected"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @CODE:TEST-001\n")

            errors = validator.validate_duplicates([str(file1), str(file2)])
            assert len(errors) == 1
            assert "TEST-001" in errors[0].tag
            assert len(errors[0].locations) == 2

    def test_multiple_duplicates(self):
        """Multiple duplicates should all be detected"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# @CODE:TEST-001
# @CODE:TEST-002
# @CODE:TEST-001
# @CODE:TEST-002
""")

            errors = validator.validate_duplicates([str(file1)])
            assert len(errors) == 2  # TEST-001 and TEST-002 both duplicated


class TestOrphanDetection:
    """Test orphan TAG detection"""

    def test_no_orphans(self):
        """Valid TAG chain should have no orphans"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create complete chain: SPEC -> CODE -> TEST -> DOC
            spec_file = Path(tmpdir) / "spec.md"
            spec_file.write_text("# @SPEC:USER-REG-001\n")

            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            doc_file = Path(tmpdir) / "README.md"
            doc_file.write_text("# @DOC:USER-REG-001\n")

            warnings = validator.validate_orphans([
                str(spec_file), str(code_file), str(test_file), str(doc_file)
            ])
            assert len(warnings) == 0

    def test_orphan_code_without_test(self):
        """CODE without TEST should generate warning"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            code_file = Path(tmpdir) / "auth.py"
            code_file.write_text("# @CODE:USER-REG-001\n")

            warnings = validator.validate_orphans([str(code_file)])
            assert len(warnings) >= 1
            assert any("USER-REG-001" in w.tag for w in warnings)
            assert any("test" in w.message.lower() for w in warnings)

    def test_orphan_test_without_code(self):
        """TEST without CODE should generate warning"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-001\n")

            warnings = validator.validate_orphans([str(test_file)])
            assert len(warnings) >= 1
            assert any("USER-REG-001" in w.tag for w in warnings)

    def test_multiple_orphans(self):
        """Multiple orphans should all be detected"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # CODE without TEST
            code1 = Path(tmpdir) / "auth.py"
            code1.write_text("# @CODE:AUTH-001\n")

            # TEST without CODE
            test1 = Path(tmpdir) / "test_payment.py"
            test1.write_text("# @TEST:PAY-001\n")

            warnings = validator.validate_orphans([str(code1), str(test1)])
            assert len(warnings) >= 2


class TestFileScanningAndValidation:
    """Test file scanning and complete validation"""

    def test_validate_empty_file_list(self):
        """Empty file list should return success"""
        validator = PreCommitValidator()
        result = validator.validate_files([])
        assert result.is_valid is True
        assert len(result.errors) == 0

    def test_validate_files_with_errors(self):
        """Files with errors should fail validation"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # File with duplicate TAGs
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# @CODE:TEST-001
# @CODE:TEST-001
""")

            result = validator.validate_files([str(file1)])
            assert result.is_valid is False
            assert len(result.errors) > 0

    def test_validate_files_with_warnings_only(self):
        """Files with only warnings should pass validation"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # CODE without TEST (warning, not error)
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            # Warnings don't block commit by default
            assert result.is_valid is True
            assert len(result.warnings) > 0

    def test_validate_mixed_errors_and_warnings(self):
        """Mix of errors and warnings should fail validation"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Duplicate (error)
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# @CODE:TEST-001
# @CODE:TEST-001
""")

            # Orphan (warning)
            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @CODE:TEST-002\n")

            result = validator.validate_files([str(file1), str(file2)])
            assert result.is_valid is False
            assert len(result.errors) > 0
            assert len(result.warnings) > 0

    def test_git_staged_files_scanning(self):
        """Scan only git staged files"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Initialize git repo
            import subprocess
            subprocess.run(["git", "init"], cwd=tmpdir, check=True)
            subprocess.run(["git", "config", "user.email", "test@example.com"], cwd=tmpdir)
            subprocess.run(["git", "config", "user.name", "Test"], cwd=tmpdir)

            # Create and stage file
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")
            subprocess.run(["git", "add", "file1.py"], cwd=tmpdir)

            # Create unstaged file
            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# @CODE:TEST-002\n")

            staged_files = validator.get_staged_files(tmpdir)
            assert "file1.py" in staged_files
            assert "file2.py" not in staged_files


class TestValidationResult:
    """Test ValidationResult data structure"""

    def test_validation_result_success(self):
        """Successful validation result"""
        result = ValidationResult(
            is_valid=True,
            errors=[],
            warnings=[]
        )
        assert result.is_valid is True
        assert len(result.errors) == 0
        assert len(result.warnings) == 0

    def test_validation_result_with_errors(self):
        """Validation result with errors"""
        error = ValidationError(
            message="Duplicate TAG found",
            tag="TEST-001",
            locations=[("file1.py", 1), ("file2.py", 1)]
        )

        result = ValidationResult(
            is_valid=False,
            errors=[error],
            warnings=[]
        )
        assert result.is_valid is False
        assert len(result.errors) == 1

    def test_validation_result_formatting(self):
        """Format validation result for display"""
        error = ValidationError(
            message="Duplicate TAG",
            tag="TEST-001",
            locations=[("file1.py", 1)]
        )
        warning = ValidationWarning(
            message="Orphan TAG",
            tag="TEST-002",
            location=("file2.py", 1)
        )

        result = ValidationResult(
            is_valid=False,
            errors=[error],
            warnings=[warning]
        )

        formatted = result.format()
        assert "Duplicate TAG" in formatted
        assert "Orphan TAG" in formatted
        assert "file1.py" in formatted
        assert "file2.py" in formatted


class TestConfigurableValidation:
    """Test configurable validation rules"""

    def test_strict_mode_blocks_warnings(self):
        """Strict mode should treat warnings as errors"""
        validator = PreCommitValidator(strict_mode=True)

        with tempfile.TemporaryDirectory() as tmpdir:
            # CODE without TEST (warning in normal mode)
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            # In strict mode, warnings should block commit
            assert result.is_valid is False

    def test_disable_orphan_check(self):
        """Should be able to disable orphan checking"""
        validator = PreCommitValidator(check_orphans=False)

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])
            # No orphan warnings when check is disabled
            assert len(result.warnings) == 0

    def test_custom_tag_pattern(self):
        """Should support custom TAG patterns"""
        # Custom pattern: @TAG:PROJECT-NNN
        validator = PreCommitValidator(
            tag_pattern=r"@(SPEC|CODE|TEST|DOC):[A-Z]+-\d{3}"
        )

        assert validator.validate_format("@CODE:PRJ-001") is True
        assert validator.validate_format("@CODE:PROJECT-FEAT-001") is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
