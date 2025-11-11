#!/usr/bin/env python3
# # REMOVED_ORPHAN_TEST:PRECOMMIT-001 | Component 1: Pre-commit validator tests
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

import pytest

from moai_adk.core.tags.pre_commit_validator import (
    PreCommitValidator,
    ValidationError,
    ValidationResult,
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
            # Create test files with UNIQUE TAGs
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# # REMOVED_ORPHAN_CODE:UNIQUE-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# # REMOVED_ORPHAN_CODE:UNIQUE-002\n")

            errors = validator.validate_duplicates([str(file1), str(file2)])
            assert len(errors) == 0

    def test_duplicate_in_same_file(self):
        """Duplicate in same file should be detected"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# # REMOVED_ORPHAN_CODE:DUPLIC-001
def func1():
    pass

# # REMOVED_ORPHAN_CODE:DUPLIC-001
def func2():
    pass
""")

            errors = validator.validate_duplicates([str(file1)])
            assert len(errors) == 1
            assert "DUPLIC-001" in errors[0].tag
            assert "duplicate" in errors[0].message.lower()

    def test_duplicate_across_files(self):
        """Duplicate across files should be detected"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# # REMOVED_ORPHAN_CODE:CROSS-001\n")

            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# # REMOVED_ORPHAN_CODE:CROSS-001\n")

            errors = validator.validate_duplicates([str(file1), str(file2)])
            assert len(errors) == 1
            assert "CROSS-001" in errors[0].tag
            assert len(errors[0].locations) == 2

    def test_multiple_duplicates(self):
        """Multiple duplicates should all be detected"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("""
# # REMOVED_ORPHAN_CODE:MULTI-001
# # REMOVED_ORPHAN_CODE:MULTI-001
# # REMOVED_ORPHAN_CODE:MULTI-002
# # REMOVED_ORPHAN_CODE:MULTI-002
""")

            errors = validator.validate_duplicates([str(file1)])
            assert len(errors) == 2  # MULTI-001 and MULTI-002 both duplicated
            tag_ids = {error.tag for error in errors}
            assert any("MULTI-001" in tag for tag in tag_ids)
            assert any("MULTI-002" in tag for tag in tag_ids)


class TestOrphanDetection:
    """Test orphan TAG detection"""

    @pytest.mark.skip(reason="TAG chain matching logic under review - # REMOVED_ORPHAN_CODE:SKIP-001")
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
            test_file.write_text("# @TEST:USER-REG-PRECOMMIT-001\n")

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

    @pytest.mark.skip(reason="TAG chain matching logic under review - # REMOVED_ORPHAN_CODE:SKIP-002")
    def test_orphan_test_without_code(self):
        """TEST without CODE should generate warning"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test_auth.py"
            test_file.write_text("# @TEST:USER-REG-PRECOMMIT-001\n")

            warnings = validator.validate_orphans([str(test_file)])
            assert len(warnings) >= 1
            assert any("USER-REG-001" in w.tag for w in warnings)

    @pytest.mark.skip(reason="TAG chain matching logic under review - # REMOVED_ORPHAN_CODE:SKIP-003")
    def test_multiple_orphans(self):
        """Multiple orphans should all be detected"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # CODE without TEST
            code1 = Path(tmpdir) / "auth.py"
            code1.write_text("# # REMOVED_ORPHAN_CODE:AUTH-004\n")

            # TEST without CODE
            test1 = Path(tmpdir) / "test_payment.py"
            test1.write_text("# # REMOVED_ORPHAN_TEST:PAY-001\n")

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
# # REMOVED_ORPHAN_CODE:ERROR-001
# # REMOVED_ORPHAN_CODE:ERROR-001
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
            file1.write_text("# # REMOVED_ORPHAN_CODE:WARN-001\n")

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
# # REMOVED_ORPHAN_CODE:MIXED-001
# # REMOVED_ORPHAN_CODE:MIXED-001
""")

            # Orphan (warning)
            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# # REMOVED_ORPHAN_CODE:ORPH-001\n")

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
            file1.write_text("# # REMOVED_ORPHAN_CODE:STAGE-001\n")
            subprocess.run(["git", "add", "file1.py"], cwd=tmpdir)

            # Create unstaged file
            file2 = Path(tmpdir) / "file2.py"
            file2.write_text("# # REMOVED_ORPHAN_CODE:UNSTAGE-001\n")

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


class TestDocumentFileExclusion:
    """Test document file exclusion from TAG validation"""

    def test_markdown_files_excluded(self):
        """Markdown files should be excluded from TAG validation"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create markdown file with duplicate TAGs
            md_file = Path(tmpdir) / "CONTRIBUTING.md"
            md_file.write_text("""
# Contributing Guide

Example # REMOVED_ORPHAN_CODE:AUTH-004 in markdown
More # REMOVED_ORPHAN_CODE:AUTH-004 elsewhere

Example @TEST:AUTH-004 in docs
More @TEST:AUTH-004 in example code
""")

            # Duplicate TAGs in markdown should NOT be flagged
            result = validator.validate_files([str(md_file)])
            assert result.is_valid is True
            assert len(result.errors) == 0

    def test_readme_files_excluded(self):
        """README files should be excluded from TAG validation"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            readme = Path(tmpdir) / "README.md"
            readme.write_text("Example # REMOVED_ORPHAN_CODE:README-001\nExample # REMOVED_ORPHAN_CODE:README-001\n")

            result = validator.validate_files([str(readme)])
            assert result.is_valid is True

    def test_changelog_files_excluded(self):
        """CHANGELOG files should be excluded from TAG validation"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            changelog = Path(tmpdir) / "CHANGELOG.md"
            changelog.write_text("""
## v1.0.0

- Fixed # REMOVED_ORPHAN_CODE:BUG-001
- Also # REMOVED_ORPHAN_CODE:BUG-001

## v0.9.0

- Added # REMOVED_ORPHAN_TEST:FEAT-001
- Also # REMOVED_ORPHAN_TEST:FEAT-001
""")

            result = validator.validate_files([str(changelog)])
            assert result.is_valid is True

    def test_code_files_still_validated(self):
        """Non-document files should still be validated for TAGs"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            py_file = Path(tmpdir) / "auth.py"
            py_file.write_text("""
# # REMOVED_ORPHAN_CODE:AUTH-004
def login():
    pass

# # REMOVED_ORPHAN_CODE:AUTH-004  <- Duplicate in code file
def verify():
    pass
""")

            # Duplicate TAGs in code file SHOULD be flagged
            result = validator.validate_files([str(py_file)])
            assert result.is_valid is False
            assert len(result.errors) == 1

    def test_mixed_files_validation(self):
        """Mix of document and code files - only code should be checked"""
        validator = PreCommitValidator()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Document with duplicates (should be ignored)
            doc = Path(tmpdir) / "CONTRIBUTING.md"
            doc.write_text("Example # REMOVED_ORPHAN_CODE:FEAT-001\nExample # REMOVED_ORPHAN_CODE:FEAT-001\n")

            # Code with duplicates (should be flagged)
            code = Path(tmpdir) / "feature.py"
            code.write_text("# # REMOVED_ORPHAN_CODE:BUG-001\n# # REMOVED_ORPHAN_CODE:BUG-001\n")

            result = validator.validate_files([str(doc), str(code)])
            assert result.is_valid is False
            # Only the code file duplicate should be flagged
            assert len(result.errors) == 1
            assert "BUG-001" in result.errors[0].tag


class TestConfigurableValidation:
    """Test configurable validation rules"""

    def test_strict_mode_blocks_warnings(self):
        """Strict mode should treat warnings as errors"""
        validator = PreCommitValidator(strict_mode=True)

        with tempfile.TemporaryDirectory() as tmpdir:
            # CODE without TEST (warning in normal mode)
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# # REMOVED_ORPHAN_CODE:TEST-002\n")

            result = validator.validate_files([str(file1)])
            # In strict mode, warnings should block commit
            assert result.is_valid is False

    def test_disable_orphan_check(self):
        """Should be able to disable orphan checking"""
        validator = PreCommitValidator(check_orphans=False)

        with tempfile.TemporaryDirectory() as tmpdir:
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# # REMOVED_ORPHAN_CODE:TEST-002\n")

            result = validator.validate_files([str(file1)])
            # No orphan warnings when check is disabled
            assert len(result.warnings) == 0

    def test_custom_tag_pattern(self):
        """Should support custom TAG patterns"""
        # Custom pattern: @TAG:PROJECT-NNN
        validator = PreCommitValidator(
            tag_pattern=r"@(SPEC|CODE|TEST|DOC):[A-Z]+-\d{3}"
        )

        assert validator.validate_format("# REMOVED_ORPHAN_CODE:PRJ-001") is True
        assert validator.validate_format("@CODE:PROJECT-FEAT-001") is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
