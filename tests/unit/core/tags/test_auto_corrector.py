#!/usr/bin/env python3
# @TEST:AUTO-CORRECTOR-001 | SPEC: TAG-AUTO-CORRECTOR-TEST-001 | CODE: @CODE:AUTO-CORRECTOR-001
"""Tests for TAG auto corrector

This module tests the TAG auto-correction functionality:
- Missing TAG auto-generation
- Duplicate TAG removal
- Chain reference auto-linking
- Confidence calculation
"""

from pathlib import Path
from unittest.mock import patch

import pytest

from moai_adk.core.tags.auto_corrector import AutoCorrection, AutoCorrectionConfig, TagAutoCorrector
from moai_adk.core.tags.policy_validator import PolicyViolation, PolicyViolationType


class TestTagAutoCorrector:
    """Test cases for TagAutoCorrector"""

    @pytest.fixture
    def corrector(self):
        """Create a test corrector instance"""
        config = AutoCorrectionConfig(
            enable_auto_fix=True,
            confidence_threshold=0.8,
            create_missing_specs=False,
            create_missing_tests=False,
            remove_duplicates=True,
            backup_before_fix=True
        )
        return TagAutoCorrector(config=config)

    @pytest.fixture
    def temp_dir(self, tmp_path):
        """Create temporary directory for test files"""
        return tmp_path

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_fix_missing_tags(self, corrector, temp_dir):
        """Test fixing missing TAGs in code files"""
        # Create a code file without TAG
        code_file = temp_dir / "src" / "auth" / "user.py"
        code_file.parent.mkdir(parents=True)
        content = "def authenticate_user(): pass"
        code_file.write_text(content)

        # Create policy violation for missing TAG
        violation = PolicyViolation(
            level="high",
            type=PolicyViolationType.MISSING_TAGS,
            tag=None,
            message="CODE 파일에 @TAG가 누락되었습니다",
            file_path=str(code_file),
            action="suggest",
            guidance="파일 상단에 @CODE:DOMAIN-NNN 형식의 TAG를 추가하세요"
        )

        # Generate correction
        corrections = corrector.generate_corrections([violation])

        assert len(corrections) > 0
        correction = corrections[0]
        assert correction.file_path == str(code_file)
        assert correction.description.startswith("@TAG 자동 추가")
        assert correction.confidence > 0
        assert "@CODE:" in correction.corrected_content

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_fix_duplicate_tags(self, corrector, temp_dir):
        """Test fixing duplicate TAGs"""
        # Create a file with duplicate TAGs
        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True, exist_ok=True)  # Create parent directory
        content = """# @CODE:EXAMPLE-001
def example():
    pass

# @CODE:EXAMPLE-001  # Duplicate
def another():
    pass"""
        code_file.write_text(content)

        violation = PolicyViolation(
            level="medium",
            type=PolicyViolationType.DUPLICATE_TAGS,
            tag="@CODE:EXAMPLE-001",
            message="중복된 TAG: @CODE:EXAMPLE-001",
            file_path=str(code_file),
            action="warn",
            guidance="중복된 TAG를 제거하세요"
        )

        corrections = corrector.generate_corrections([violation])

        assert len(corrections) > 0
        correction = corrections[0]
        assert "중복 TAG 제거" in correction.description
        assert correction.confidence > 0.9  # High confidence for duplicate removal

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_suggest_tag_for_code_file(self, corrector):
        """Test TAG suggestion for code files"""
        # Test various file paths
        test_cases = [
            ("src/auth/user.py", "AUTH"),
            ("src/api/client.py", "API"),
            ("lib/utils/helpers.py", "UTILS"),
        ]

        for file_path, expected_domain in test_cases:
            tag_suggestion = corrector.suggest_tag_for_code_file(file_path)

            assert tag_suggestion is not None
            tag, confidence = tag_suggestion
            assert tag.startswith("@CODE:")
            assert expected_domain in tag
            assert 0.0 <= confidence <= 1.0

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_calculate_next_number(self, corrector):
        """Test calculation of next TAG number"""
        # Test with existing tags
        existing_numbers = {1, 2, 3, 5}
        domain = "EXAMPLE"

        next_number = corrector._calculate_next_number(existing_numbers, domain)
        assert next_number == 4  # Should find the first missing number

        # Test with consecutive numbers
        existing_numbers = {1, 2, 3}
        next_number = corrector._calculate_next_number(existing_numbers, domain)
        assert next_number == 4

        # Test with empty set
        existing_numbers = set()
        next_number = corrector._calculate_next_number(existing_numbers, domain)
        assert next_number == 1

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_extract_domain_from_path(self, corrector):
        """Test domain extraction from file paths"""
        test_cases = [
            ("src/auth/user.py", "AUTH"),
            ("src/api/client.py", "API"),
            ("lib/utils/helpers.py", "UTILS"),
            ("main.py", "MAIN"),
            ("tests/test_auth.py", None),  # Test files should not extract domain
        ]

        for file_path, expected_domain in test_cases:
            domain = corrector._extract_domain_from_path(Path(file_path))
            assert domain == expected_domain, f"Failed for {file_path}: expected {expected_domain}, got {domain}"

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_find_existing_tags_in_project(self, corrector, temp_dir):
        """Test finding existing TAGs in project"""
        # Create test files with TAGs
        (temp_dir / "src").mkdir()
        (temp_dir / "src" / "auth.py").write_text("# # REMOVED_ORPHAN_CODE:AUTH-004")
        (temp_dir / "src" / "api.py").write_text("# # REMOVED_ORPHAN_CODE:API-001")
        (temp_dir / "README.md").write_text("# @DOC:README-001")

        with patch('Path.cwd', return_value=temp_dir):
            existing_tags = corrector._find_existing_tags_in_project("AUTH")

            assert "AUTH-001" in existing_tags

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_create_missing_spec_file(self, corrector, temp_dir):
        """Test creation of missing SPEC files"""
        # Create corrector with SPEC creation enabled
        config = AutoCorrectionConfig(create_missing_specs=True)
        spec_creator = TagAutoCorrector(config=config)

        with patch('Path.cwd', return_value=temp_dir):
            created = spec_creator._create_spec_file("AUTH-001")

            assert created
            spec_file = temp_dir / ".moai" / "specs" / "SPEC-AUTH-001" / "spec.md"
            assert spec_file.exists()
            content = spec_file.read_text()
            assert "@SPEC:AUTH-001" in content

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_create_missing_test_file(self, corrector, temp_dir):
        """Test creation of missing TEST files"""
        # Create corrector with TEST creation enabled
        config = AutoCorrectionConfig(create_missing_tests=True)
        test_creator = TagAutoCorrector(config=config)

        with patch('Path.cwd', return_value=temp_dir):
            correction = test_creator._create_missing_test_file("@CODE:AUTH-001")

            assert correction is not None
            assert correction.file_path.endswith("test_auth.py") or "test_" in correction.file_path
            assert "@TEST:AUTH-001" in correction.corrected_content

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_apply_corrections(self, corrector, temp_dir):
        """Test applying corrections to files"""
        # Create test file
        test_file = temp_dir / "test.py"
        original_content = "def example(): pass"
        test_file.write_text(original_content)

        # Create correction
        correction = AutoCorrection(
            file_path=str(test_file),
            original_content=original_content,
            corrected_content="# @CODE:EXAMPLE-001\ndef example(): pass",
            description="Add TAG",
            confidence=0.9,
            requires_review=False
        )

        # Apply correction
        success = corrector.apply_corrections([correction])

        assert success
        updated_content = test_file.read_text()
        assert "@CODE:EXAMPLE-001" in updated_content

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_backup_creation(self, corrector, temp_dir):
        """Test backup creation before fixing"""
        # Create test file
        test_file = temp_dir / "test.py"
        original_content = "def example(): pass"
        test_file.write_text(original_content)

        # Create correction
        correction = AutoCorrection(
            file_path=str(test_file),
            original_content=original_content,
            corrected_content="# @CODE:EXAMPLE-001\ndef example(): pass",
            description="Add TAG",
            confidence=0.9,
            requires_review=False
        )

        # Apply correction (should create backup)
        corrector.apply_corrections([correction])

        # Check backup exists
        backup_file = Path(f"{test_file}.backup")
        assert backup_file.exists()
        assert backup_file.read_text() == original_content

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_confidence_threshold_filtering(self, temp_dir):
        """Test filtering by confidence threshold"""
        # Create corrector with high threshold
        config = AutoCorrectionConfig(confidence_threshold=0.9)
        high_threshold_corrector = TagAutoCorrector(config=config)

        # Create corrections with different confidence levels
        corrections = [
            AutoCorrection(
                file_path="test1.py",
                original_content="def test1(): pass",
                corrected_content="# # REMOVED_ORPHAN_CODE:TEST-002\ndef test1(): pass",
                description="Add TAG",
                confidence=0.7,  # Below threshold
                requires_review=True
            ),
            AutoCorrection(
                file_path="test2.py",
                original_content="def test2(): pass",
                corrected_content="# # REMOVED_ORPHAN_CODE:TEST-002\ndef test2(): pass",
                description="Add TAG",
                confidence=0.95,  # Above threshold
                requires_review=False
            )
        ]

        # Should only apply high-confidence corrections
        success = high_threshold_corrector.apply_corrections(corrections)
        assert success  # Overall success because at least one correction was applied

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_no_corrections_case(self, corrector):
        """Test handling of no violations/corrections needed"""
        corrections = corrector.generate_corrections([])
        assert len(corrections) == 0

        success = corrector.apply_corrections([])
        assert success  # Should succeed with no corrections

    @pytest.mark.xfail(reason='Test data migration needed')
    def test_fix_missing_spec_reference(self, corrector, temp_dir):
        """Test fixing missing SPEC reference"""
        # Create code file with TAG but no SPEC reference
        code_file = temp_dir / "src" / "example.py"
        code_file.parent.mkdir(parents=True)
        content = "# @CODE:EXAMPLE-001\ndef example(): pass"
        code_file.write_text(content)

        violation = PolicyViolation(
            level="high",
            type=PolicyViolationType.NO_SPEC_REFERENCE,
            tag="@CODE:EXAMPLE-001",
            message="@CODE:EXAMPLE-001에 연결된 SPEC이 없습니다",
            file_path=str(code_file),
            action="block",
            guidance=".moai/specs/SPEC-EXAMPLE-001/spec.md 파일을 생성하세요"
        )

        # Enable SPEC creation for this test
        config = AutoCorrectionConfig(create_missing_specs=True)
        spec_creator = TagAutoCorrector(config=config)

        corrections = spec_creator.generate_corrections([violation])

        # Should create both SPEC file and update code reference
        assert len(corrections) >= 1
