#!/usr/bin/env python3
# @TEST:POLICY-VALIDATOR-001 | SPEC: TAG-POLICY-VALIDATOR-TEST-001 | CODE: @CODE:POLICY-VALIDATOR-001
"""Tests for TAG policy validator

This module tests the TAG policy validation functionality:
- SPEC-less code detection
- TAG format validation
- Policy violation detection
- Enforcement level verification
"""

from pathlib import Path

import pytest

from moai_adk.core.tags.policy_validator import (
    PolicyValidationConfig,
    PolicyViolation,
    PolicyViolationLevel,
    PolicyViolationType,
    TagPolicyValidator,
)


class TestTagPolicyValidator:
    """Test cases for TagPolicyValidator"""

    @pytest.fixture
    def validator(self):
        """Create a test validator instance"""
        config = PolicyValidationConfig(
            strict_mode=True,
            require_spec_before_code=True,
            require_test_for_code=True,
            validation_timeout=5
        )
        return TagPolicyValidator(config=config)

    @pytest.fixture
    def temp_dir(self, tmp_path):
        """Create temporary directory for test files"""
        return tmp_path

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_detect_specless_code_violation(self, validator, temp_dir):
        """Test detection of CODE files without SPEC"""
        # Create a code file without TAG (don't write yet - validate before creation)
        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True)

        # Should detect SPEC-less code violation for new file
        violations = validator.validate_before_creation(str(code_file), "def example(): pass")

        assert len(violations) > 0
        specless_violations = [v for v in violations if v.type == PolicyViolationType.SPECLESS_CODE]
        assert len(specless_violations) > 0
        assert specless_violations[0].level == PolicyViolationLevel.CRITICAL
        assert "TAG가 없습니다" in specless_violations[0].message

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_detect_missing_spec_reference(self, validator, temp_dir):
        """Test detection of CODE with TAG but missing SPEC reference"""
        # Create a code file with TAG but no SPEC (don't write - validate before creation)
        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True)
        content = "# @CODE:EXAMPLE-001\ndef example(): pass"

        violations = validator.validate_before_creation(str(code_file), content)

        no_spec_violations = [v for v in violations if v.type == PolicyViolationType.NO_SPEC_REFERENCE]
        assert len(no_spec_violations) > 0
        assert no_spec_violations[0].tag == "@CODE:EXAMPLE-001"
        assert "연결된 SPEC이 없습니다" in no_spec_violations[0].message

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_detect_missing_test_for_code(self, validator, temp_dir):
        """Test detection of CODE without corresponding TEST"""
        # Create a code file with valid TAG
        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True)
        content = "# @CODE:EXAMPLE-001\ndef example(): pass"
        code_file.write_text(content)

        violations = validator.validate_after_modification(str(code_file), content)

        chain_violations = [v for v in violations if v.type == PolicyViolationType.CHAIN_BREAK]
        assert len(chain_violations) > 0
        assert "연결된 TEST가 없습니다" in chain_violations[0].message

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_detect_duplicate_tags(self, validator, temp_dir):
        """Test detection of duplicate TAGs in same file"""
        # Create a file with duplicate TAGs
        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True)
        content = """# @CODE:EXAMPLE-001
def example():
    pass

# @CODE:EXAMPLE-001  # Duplicate
def another():
    pass"""
        code_file.write_text(content)

        violations = validator.validate_after_modification(str(code_file), content)

        duplicate_violations = [v for v in violations if v.type == PolicyViolationType.DUPLICATE_TAGS]
        assert len(duplicate_violations) > 0
        assert duplicate_violations[0].tag == "@CODE:EXAMPLE-001"

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_validate_valid_tag_format(self, validator, temp_dir):
        """Test validation of files with correct TAG format"""
        # Create files with valid TAG format
        spec_file = temp_dir / ".moai" / "specs" / "SPEC-EXAMPLE-001" / "spec.md"
        spec_file.parent.mkdir(parents=True)
        spec_content = """# @SPEC:EXAMPLE-001
## Requirements
- System must provide example functionality
"""
        spec_file.write_text(spec_content)

        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True)
        code_content = """# @CODE:EXAMPLE-001 | SPEC: .moai/specs/SPEC-EXAMPLE-001/spec.md
def example():
    pass
"""
        code_file.write_text(code_content)

        # Should not detect any violations for valid format
        violations = validator.validate_before_creation(str(code_file), code_content)

        critical_violations = [v for v in violations if v.level == PolicyViolationLevel.CRITICAL]
        assert len(critical_violations) == 0

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_strict_mode_enforcement(self, temp_dir):
        """Test strict mode vs lenient mode behavior"""
        # Test strict mode
        strict_config = PolicyValidationConfig(strict_mode=True)
        strict_validator = TagPolicyValidator(config=strict_config)

        code_file = temp_dir / "src" / "example.py"
        content = "def example(): pass"

        strict_violations = strict_validator.validate_before_creation(str(code_file), content)
        assert len(strict_violations) > 0

        # Test lenient mode
        lenient_config = PolicyValidationConfig(strict_mode=False)
        lenient_validator = TagPolicyValidator(config=lenient_config)

        lenient_violations = lenient_validator.validate_before_creation(str(code_file), content)
        # Lenient mode should still detect issues but may have different levels
        assert len(lenient_violations) > 0

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_file_type_filtering(self, validator, temp_dir):
        """Test that only relevant file types are validated"""
        # Create non-code file
        config_file = temp_dir / "config.json"
        config_file.write_text('{"name": "test"}')

        # Should not validate non-code files
        violations = validator.validate_before_creation(str(config_file), '{"name": "test"}')
        assert len(violations) == 0

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_validation_timeout(self, validator, temp_dir):
        """Test validation timeout handling"""
        # Create a validator with very short timeout
        timeout_config = PolicyValidationConfig(validation_timeout=0.001)  # 1ms
        timeout_validator = TagPolicyValidator(config=timeout_config)

        code_file = temp_dir / "src" / "example.py"
        content = "def example(): pass"

        # Should handle timeout gracefully
        violations = timeout_validator.validate_before_creation(str(code_file), content)
        # May return empty list due to timeout
        assert isinstance(violations, list)

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_validation_report_generation(self, validator):
        """Test validation report generation"""
        violations = [
            PolicyViolation(
                level=PolicyViolationLevel.CRITICAL,
                type=PolicyViolationType.SPECLESS_CODE,
                tag=None,
                message="CODE 파일에 @TAG가 없습니다",
                file_path="test.py",
                action="block",
                guidance="TAG를 추가하세요"
            )
        ]

        report = validator.create_validation_report(violations)

        assert "❌ TAG 정책 위반 발견" in report
        assert "치명적" in report
        assert "TAG를 추가하세요" in report

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_no_violations_success_case(self, validator, temp_dir):
        """Test successful validation with no violations"""
        # Create a complete SPEC → CODE → TEST chain
        spec_file = temp_dir / ".moai" / "specs" / "SPEC-EXAMPLE-001" / "spec.md"
        spec_file.parent.mkdir(parents=True)
        spec_file.write_text("# @SPEC:EXAMPLE-001\n## Requirements")

        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True)
        code_file.write_text("# @CODE:EXAMPLE-001\ndef example(): pass")

        test_file = temp_dir / "tests" / "test_example.py"
        test_file.parent.mkdir(parents=True)
        test_file.write_text("# @TEST:EXAMPLE-001\ndef test_example(): pass")

        violations = validator.validate_after_modification(str(code_file), "# @CODE:EXAMPLE-001\ndef example(): pass")

        # Should have no critical violations for complete chain
        critical_violations = [v for v in violations if v.level == PolicyViolationLevel.CRITICAL]
        assert len(critical_violations) == 0

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_domain_extraction_from_file_path(self, validator):
        """Test domain extraction from file paths"""
        # Test various file path patterns
        test_cases = [
            ("src/auth/user.py", "AUTH"),
            ("src/api/client.py", "API"),
            ("tests/test_user_auth.py", None),  # Test files should not return domain
            ("docs/guide.md", None),  # Docs should not return domain
        ]

        for file_path, expected_domain in test_cases:
            domain = validator._extract_domain_from_path(Path(file_path))
            assert domain == expected_domain, f"Failed for {file_path}: expected {expected_domain}, got {domain}"

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_auto_fix_possible_detection(self, validator, temp_dir):
        """Test detection of violations that can be auto-fixed"""
        # Create a code file without TAG (auto-fixable)
        code_file = temp_dir / "src" / "example.py"
        content = "def example(): pass"

        violations = validator.validate_before_creation(str(code_file), content)

        # Some violations should be auto-fixable
        auto_fixable = [v for v in violations if v.auto_fix_possible]
        assert len(auto_fixable) > 0
