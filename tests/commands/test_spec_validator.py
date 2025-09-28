"""
@TEST:SPEC-VALIDATOR-001 SpecValidator Unit Tests
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:VALIDATOR-001 → @TEST:SPEC-VALIDATOR-001

Tests for input validation module following TRUST principles:
- T: Test-first development
- R: Readable test structure
- U: Unified single responsibility
- S: Secure input validation
- T: Trackable test coverage
"""

import pytest
from src.moai_adk.commands.spec_validator import SpecValidator


class TestSpecValidator:
    """Test class for SpecValidator following TRUST principles"""

    def setup_method(self):
        """Setup SpecValidator instance for each test"""
        self.validator = SpecValidator()

    # Happy Path Tests
    def test_should_validate_normal_spec_name(self):
        """Test normal spec name validation (happy path)"""
        # Given
        spec_name = "USER-AUTH"

        # When
        result = self.validator.validate_spec_name(spec_name)

        # Then
        assert result == "USER-AUTH"

    def test_should_validate_normal_description(self):
        """Test normal description validation (happy path)"""
        # Given
        description = "User authentication system implementation"

        # When
        result = self.validator.validate_description(description)

        # Then
        assert result == "User authentication system implementation"

    # Edge Cases
    def test_should_normalize_spec_name_with_whitespace(self):
        """Test spec name normalization with leading/trailing whitespace"""
        # Given
        spec_name = "  user-auth  "

        # When
        result = self.validator.validate_spec_name(spec_name)

        # Then
        assert result == "USER-AUTH"

    def test_should_handle_max_length_spec_name(self):
        """Test spec name at maximum allowed length (50 chars)"""
        # Given
        spec_name = "A" * 50

        # When
        result = self.validator.validate_spec_name(spec_name)

        # Then
        assert result == "A" * 50
        assert len(result) == 50

    def test_should_handle_max_length_description(self):
        """Test description at maximum allowed length (500 chars)"""
        # Given
        description = "A" * 500

        # When
        result = self.validator.validate_description(description)

        # Then
        assert result == "A" * 500
        assert len(result) == 500

    # Error Cases - Spec Name
    def test_should_reject_empty_spec_name(self):
        """Test rejection of empty spec name"""
        # Given
        spec_name = ""

        # When & Then
        with pytest.raises(ValueError, match="spec_name은 비어있지 않은 문자열이어야 합니다"):
            self.validator.validate_spec_name(spec_name)

    def test_should_reject_none_spec_name(self):
        """Test rejection of None spec name"""
        # Given
        spec_name = None

        # When & Then
        with pytest.raises(ValueError, match="spec_name은 비어있지 않은 문자열이어야 합니다"):
            self.validator.validate_spec_name(spec_name)

    def test_should_reject_too_long_spec_name(self):
        """Test rejection of spec name exceeding 50 characters"""
        # Given
        spec_name = "A" * 51

        # When & Then
        with pytest.raises(ValueError, match="명세 이름이 너무 깁니다 \\(최대 50자\\)"):
            self.validator.validate_spec_name(spec_name)

    def test_should_reject_unsafe_characters_in_spec_name(self):
        """Test rejection of unsafe characters in spec name"""
        # Given
        unsafe_names = ["test/spec", "test\\spec", "test<spec", "test>spec",
                       "test:spec", 'test"spec', "test|spec", "test?spec", "test*spec"]

        # When & Then
        for unsafe_name in unsafe_names:
            with pytest.raises(ValueError, match="명세 이름에 안전하지 않은 문자가 포함되어 있습니다"):
                self.validator.validate_spec_name(unsafe_name)

    # Error Cases - Description
    def test_should_reject_empty_description(self):
        """Test rejection of empty description"""
        # Given
        description = ""

        # When & Then
        with pytest.raises(ValueError, match="description은 비어있지 않은 문자열이어야 합니다"):
            self.validator.validate_description(description)

    def test_should_reject_none_description(self):
        """Test rejection of None description"""
        # Given
        description = None

        # When & Then
        with pytest.raises(ValueError, match="description은 비어있지 않은 문자열이어야 합니다"):
            self.validator.validate_description(description)

    def test_should_reject_too_long_description(self):
        """Test rejection of description exceeding 500 characters"""
        # Given
        description = "A" * 501

        # When & Then
        with pytest.raises(ValueError, match="설명이 너무 깁니다 \\(최대 500자\\)"):
            self.validator.validate_description(description)

    # Boundary Tests
    def test_should_handle_whitespace_only_spec_name(self):
        """Test handling of whitespace-only spec name"""
        # Given
        spec_name = "   "

        # When & Then
        with pytest.raises(ValueError, match="spec_name은 비어있지 않은 문자열이어야 합니다"):
            self.validator.validate_spec_name(spec_name)

    def test_should_handle_whitespace_only_description(self):
        """Test handling of whitespace-only description"""
        # Given
        description = "   "

        # When & Then
        with pytest.raises(ValueError, match="description은 비어있지 않은 문자열이어야 합니다"):
            self.validator.validate_description(description)

    def test_should_normalize_description_whitespace(self):
        """Test description whitespace normalization"""
        # Given
        description = "  Test description with extra spaces  "

        # When
        result = self.validator.validate_description(description)

        # Then
        assert result == "Test description with extra spaces"
        assert not result.startswith(" ")
        assert not result.endswith(" ")