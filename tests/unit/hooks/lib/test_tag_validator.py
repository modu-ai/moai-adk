"""Tests for TAG validator module (T1: TAG Pattern Definition)."""

from pathlib import Path

import pytest

from moai_adk.tag_system import validator


class TestTAGModel:
    """Test TAG data model."""

    def test_tag_creation_with_all_fields(self):
        """Test creating a TAG with all required fields."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        assert tag.spec_id == "SPEC-AUTH-001"
        assert tag.verb == "impl"
        assert tag.file_path == Path("auth.py")
        assert tag.line == 10

    def test_tag_creation_with_default_verb(self):
        """Test creating a TAG with default verb."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            file_path=Path("auth.py"),
            line=10,
        )

        assert tag.verb == "impl"  # Default verb

    def test_tag_equality(self):
        """Test TAG equality comparison."""
        tag1 = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )
        tag2 = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        assert tag1 == tag2

    def test_tag_inequality_different_spec_id(self):
        """Test TAG inequality with different SPEC-ID."""
        tag1 = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )
        tag2 = validator.TAG(
            spec_id="SPEC-AUTH-002",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        assert tag1 != tag2

    def test_tag_repr(self):
        """Test TAG string representation."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        repr_str = repr(tag)
        assert "SPEC-AUTH-001" in repr_str
        assert "impl" in repr_str
        assert "auth.py" in repr_str
        assert "10" in repr_str


class TestTAGPatternValidation:
    """Test TAG pattern validation (T1.1: TAG Pattern Definition)."""

    def test_valid_spec_id_format(self):
        """Test valid SPEC-ID format: SPEC-{DOMAIN}-{NUMBER}."""
        # Valid formats
        assert validator.validate_spec_id_format("SPEC-AUTH-001")
        assert validator.validate_spec_id_format("SPEC-TAG-001")
        assert validator.validate_spec_id_format("SPEC-TDD-999")
        assert validator.validate_spec_id_format("SPEC-I18N-001")

    def test_invalid_spec_id_format_lowercase(self):
        """Test invalid SPEC-ID with lowercase prefix."""
        assert not validator.validate_spec_id_format("spec-auth-001")
        assert not validator.validate_spec_id_format("Spec-Auth-001")

    def test_invalid_spec_id_format_missing_prefix(self):
        """Test invalid SPEC-ID missing SPEC- prefix."""
        assert not validator.validate_spec_id_format("AUTH-001")
        assert not validator.validate_spec_id_format("TAG-001")

    def test_invalid_spec_id_format_wrong_separator(self):
        """Test invalid SPEC-ID with wrong separators."""
        assert not validator.validate_spec_id_format("SPEC_AUTH_001")
        assert not validator.validate_spec_id_format("SPEC.AUTH.001")
        assert not validator.validate_spec_id_format("SPEC-AUTH_001")

    def test_invalid_spec_id_format_missing_digits(self):
        """Test invalid SPEC-ID with missing or wrong number format."""
        assert not validator.validate_spec_id_format("SPEC-AUTH-1")
        assert not validator.validate_spec_id_format("SPEC-AUTH-01")
        assert not validator.validate_spec_id_format("SPEC-AUTH-0001")
        assert not validator.validate_spec_id_format("SPEC-AUTH-ABC")

    def test_invalid_spec_id_format_empty_domain(self):
        """Test invalid SPEC-ID with empty domain."""
        assert not validator.validate_spec_id_format("SPEC--001")

    def test_valid_verb_options(self):
        """Test valid verb options (T1.2: Optional verbs)."""
        valid_verbs = ["impl", "verify", "depends", "related"]

        for verb in valid_verbs:
            assert validator.validate_verb(verb)

    def test_invalid_verb(self):
        """Test invalid verb."""
        assert not validator.validate_verb("invalid")
        assert not validator.validate_verb("implement")
        assert not validator.validate_verb("test")

    def test_default_verb_is_impl(self):
        """Test that default verb is 'impl' (T1.1: TAG Pattern Definition)."""
        assert validator.get_default_verb() == "impl"


class TestTAGValidation:
    """Test complete TAG validation."""

    def test_validate_tag_with_valid_tag(self):
        """Test validating a valid TAG."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert is_valid
        assert len(errors) == 0

    def test_validate_tag_with_invalid_spec_id(self):
        """Test validating TAG with invalid SPEC-ID."""
        tag = validator.TAG(
            spec_id="invalid-format",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert not is_valid
        assert len(errors) > 0
        assert any("Invalid SPEC-ID format" in error for error in errors)

    def test_validate_tag_with_invalid_verb(self):
        """Test validating TAG with invalid verb."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="invalid",
            file_path=Path("auth.py"),
            line=10,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert not is_valid
        assert len(errors) > 0
        assert any("Invalid verb" in error for error in errors)


class TestTAGStringParsing:
    """Test parsing TAG from comment strings."""

    def test_parse_tag_string_valid_format(self):
        """Test parsing valid TAG string format."""
        comment = "# @SPEC SPEC-AUTH-001"

        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"
        assert tag.verb == "impl"  # Default
        assert tag.file_path == Path("test.py")
        assert tag.line == 1

    def test_parse_tag_string_with_verb(self):
        """Test parsing TAG string with explicit verb."""
        comment = "# @SPEC SPEC-AUTH-001 verify"

        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"
        assert tag.verb == "verify"

    def test_parse_tag_string_with_description(self):
        """Test parsing TAG string with descriptive text."""
        comment = "# @SPEC SPEC-AUTH-001 impl - User authentication flow"

        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"
        assert tag.verb == "impl"

    def test_parse_tag_string_invalid_format(self):
        """Test parsing invalid TAG string format."""
        comment = "# @spec auth-001"  # Wrong case and format

        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is None

    def test_parse_tag_string_missing_prefix(self):
        """Test parsing TAG string without @SPEC prefix."""
        comment = "# SPEC-AUTH-001"

        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is None

    def test_parse_tag_string_empty_comment(self):
        """Test parsing empty comment string."""
        comment = "# Just a regular comment"

        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is None


class TestValidationErrors:
    """Test validation error messages."""

    def test_invalid_spec_id_error_message(self):
        """Test error message for invalid SPEC-ID format."""
        tag = validator.TAG(
            spec_id="invalid",
            verb="impl",
            file_path=Path("test.py"),
            line=1,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert not is_valid
        assert errors
        assert "Invalid SPEC-ID format" in errors[0]

    def test_invalid_verb_error_message(self):
        """Test error message for invalid verb."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="badverb",
            file_path=Path("test.py"),
            line=1,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert not is_valid
        assert errors
        assert "Invalid verb" in errors[0]


class TestSecurityValidation:
    """Test security-related validation."""

    def test_path_traversal_prevention(self):
        """Test that path traversal attempts are rejected."""
        # Path traversal in SPEC-ID should fail format validation
        assert not validator.validate_spec_id_format("../../etc/passwd")
        assert not validator.validate_spec_id_format("SPEC-../etc-001")

    def test_code_injection_prevention(self):
        """Test that code injection attempts are rejected."""
        # Command injection attempts should fail format validation
        assert not validator.validate_spec_id_format("SPEC-001; rm -rf /")
        assert not validator.validate_spec_id_format("SPEC-001 && cat /etc/passwd")

    def test_special_characters_rejected(self):
        """Test that special characters are rejected in SPEC-ID."""
        assert not validator.validate_spec_id_format("SPEC-AUTH!001")
        assert not validator.validate_spec_id_format("SPEC-AUTH@001")
        assert not validator.validate_spec_id_format("SPEC-AUTH$001")
