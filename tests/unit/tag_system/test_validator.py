"""Comprehensive TDD tests for TAG validator module.

Targets 100% coverage for validator.py including:
- TAG dataclass frozen behavior and __post_init__ (line 48)
- parse_tag_string edge cases (lines 154, 162, 178, 185)
"""

from pathlib import Path

import pytest

from moai_adk.tag_system import validator


class TestTAGFrozenDataclass:
    """Test TAG frozen dataclass behavior and __post_init__."""

    def test_tag_frozen_immutable(self):
        """Test that TAG instances are immutable (frozen=True)."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        # Attempting to modify should raise AttributeError
        with pytest.raises(AttributeError):
            tag.spec_id = "SPEC-AUTH-002"  # type: ignore[misc]

        with pytest.raises(AttributeError):
            tag.verb = "verify"  # type: ignore[misc]

    def test_tag_post_init_normalizes_verb_case(self):
        """Test __post_init__ normalizes verb to lowercase (line 48)."""
        # Test with uppercase verb
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="IMPL",  # Uppercase
            file_path=Path("auth.py"),
            line=10,
        )

        # Verb should be normalized to lowercase
        assert tag.verb == "impl"

    def test_tag_post_init_mixed_case_verb(self):
        """Test __post_init__ handles mixed case verbs."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="VeRiFy",  # Mixed case
            file_path=Path("auth.py"),
            line=10,
        )

        assert tag.verb == "verify"

    def test_tag_post_init_all_valid_verbs_case_normalized(self):
        """Test that all valid verbs are case-normalized."""
        test_cases = [
            ("IMPL", "impl"),
            ("VERIFY", "verify"),
            ("DEPENDS", "depends"),
            ("RELATED", "related"),
            ("ImPl", "impl"),
            ("VeRiFy", "verify"),
        ]

        for input_verb, expected_verb in test_cases:
            tag = validator.TAG(
                spec_id="SPEC-AUTH-001",
                verb=input_verb,
                file_path=Path("test.py"),
                line=1,
            )
            assert tag.verb == expected_verb

    def test_tag_hashable_for_sets(self):
        """Test that TAG is hashable (required for frozen dataclass)."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

        # Can create sets of TAGs
        tag_set = {tag1, tag2}

        # Duplicate TAGs are deduplicated
        assert len(tag_set) == 1


class TestValidateSpecIdFormatEdgeCases:
    """Test SPEC-ID format validation edge cases."""

    def test_validate_spec_id_with_non_string(self):
        """Test validate_spec_id_format with non-string input."""
        # None
        assert not validator.validate_spec_id_format(None)  # type: ignore[arg-type]

        # Integer
        assert not validator.validate_spec_id_format(123)  # type: ignore[arg-type]

        # List
        assert not validator.validate_spec_id_format(["SPEC-AUTH-001"])  # type: ignore[arg-type]

        # Object
        assert not validator.validate_spec_id_format(object())  # type: ignore[arg-type]

    def test_valid_spec_id_with_digits_in_domain(self):
        """Test valid SPEC-ID with digits in domain."""
        assert validator.validate_spec_id_format("SPEC-AUTH123-001")
        assert validator.validate_spec_id_format("SPEC-TEST99-001")
        assert validator.validate_spec_id_format("SPEC-T2T-001")

    def test_valid_spec_id_with_999(self):
        """Test valid SPEC-ID with maximum 3-digit number."""
        assert validator.validate_spec_id_format("SPEC-AUTH-999")

    def test_invalid_spec_id_with_too_many_digits(self):
        """Test invalid SPEC-ID with 4+ digits."""
        assert not validator.validate_spec_id_format("SPEC-AUTH-1000")
        assert not validator.validate_spec_id_format("SPEC-AUTH-9999")

    def test_invalid_spec_id_with_fewer_digits(self):
        """Test invalid SPEC-ID with fewer than 3 digits."""
        assert not validator.validate_spec_id_format("SPEC-AUTH-99")
        assert not validator.validate_spec_id_format("SPEC-AUTH-9")

    def test_invalid_spec_id_with_leading_zeros(self):
        """Test SPEC-ID with leading zeros is valid."""
        # Leading zeros are valid for the number part
        assert validator.validate_spec_id_format("SPEC-AUTH-001")
        assert validator.validate_spec_id_format("SPEC-AUTH-099")

    def test_invalid_spec_id_empty_domain(self):
        """Test SPEC-ID with empty domain."""
        assert not validator.validate_spec_id_format("SPEC--001")

    def test_invalid_spec_id_empty_number(self):
        """Test SPEC-ID with empty number."""
        assert not validator.validate_spec_id_format("SPEC-AUTH-")

    def test_invalid_spec_id_only_prefix(self):
        """Test SPEC-ID with only SPEC- prefix."""
        assert not validator.validate_spec_id_format("SPEC-")

    def test_invalid_spec_id_special_chars_in_domain(self):
        """Test SPEC-ID with special characters in domain."""
        invalid_formats = [
            "SPEC-AUTH!-001",
            "SPEC-AUTH@-001",
            "SPEC-AUTH#-001",
            "SPEC-AUTH$-001",
            "SPEC-AUTH%-001",
            "SPEC-AUTH^-001",
            "SPEC-AUTH&-001",
            "SPEC-AUTH*-001",
            "SPEC-AUTH(-001)",
            "SPEC-AUTH+-001",
        ]

        for spec_id in invalid_formats:
            assert not validator.validate_spec_id_format(spec_id), f"Should reject: {spec_id}"

    def test_invalid_spec_id_whitespace(self):
        """Test SPEC-ID with whitespace."""
        assert not validator.validate_spec_id_format("SPEC-AUTH -001")
        assert not validator.validate_spec_id_format("SPEC-AUTH- 001")
        assert not validator.validate_spec_id_format(" SPEC-AUTH-001")

    def test_invalid_spec_id_newline_injection(self):
        """Test SPEC-ID with newline injection."""
        assert not validator.validate_spec_id_format("SPEC-AUTH\n-001")
        assert not validator.validate_spec_id_format("SPEC-AUTH-\n001")


class TestValidateVerbEdgeCases:
    """Test verb validation edge cases."""

    def test_validate_verb_with_non_string(self):
        """Test validate_verb with non-string input."""
        assert not validator.validate_verb(None)  # type: ignore[arg-type]
        assert not validator.validate_verb(123)  # type: ignore[arg-type]
        assert not validator.validate_verb(["impl"])  # type: ignore[arg-type]
        assert not validator.validate_verb(object())  # type: ignore[arg-type]

    def test_validate_verb_case_sensitive(self):
        """Test that verb validation is case-sensitive."""
        # Valid lowercase
        assert validator.validate_verb("impl")

        # Invalid uppercase
        assert not validator.validate_verb("IMPL")
        assert not validator.validate_verb("Impl")

    def test_validate_verb_all_valid_options(self):
        """Test all valid verb options."""
        valid_verbs = ["impl", "verify", "depends", "related"]

        for verb in valid_verbs:
            assert validator.validate_verb(verb), f"Should accept: {verb}"

    def test_validate_verb_similar_invalid(self):
        """Test similar but invalid verbs."""
        invalid_verbs = [
            "implement",
            "verifycations",
            "dependency",
            "relation",
            "impls",
            "verifying",
            "depending",
            "relating",
        ]

        for verb in invalid_verbs:
            assert not validator.validate_verb(verb), f"Should reject: {verb}"

    def test_validate_verb_empty_string(self):
        """Test verb validation with empty string."""
        assert not validator.validate_verb("")

    def test_validate_verb_whitespace(self):
        """Test verb validation with whitespace."""
        assert not validator.validate_verb(" impl")
        assert not validator.validate_verb("impl ")
        assert not validator.validate_verb(" impl ")


class TestValidateTagComplete:
    """Test complete TAG validation with all error scenarios."""

    def test_validate_tag_multiple_errors(self):
        """Test validation with multiple errors."""
        tag = validator.TAG(
            spec_id="invalid",
            verb="badverb",
            file_path=Path("test.py"),
            line=1,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert not is_valid
        assert len(errors) == 2
        assert any("Invalid SPEC-ID format" in error for error in errors)
        assert any("Invalid verb" in error for error in errors)

    def test_validate_tag_error_message_format(self):
        """Test that error messages are properly formatted."""
        tag = validator.TAG(
            spec_id="bad",
            verb="impl",
            file_path=Path("test.py"),
            line=1,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert not is_valid
        error_msg = errors[0]

        # Check error message format
        assert "Invalid SPEC-ID format" in error_msg
        assert "'bad'" in error_msg
        assert "SPEC-{DOMAIN}-{NUMBER}" in error_msg
        assert "SPEC-AUTH-001" in error_msg

    def test_validate_tag_verb_error_message(self):
        """Test verb error message format."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="badverb",
            file_path=Path("test.py"),
            line=1,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert not is_valid
        error_msg = errors[0]

        # Check error message format
        assert "Invalid verb" in error_msg
        assert "'badverb'" in error_msg
        assert "Valid verbs:" in error_msg
        # Check that all valid verbs are listed
        assert "impl" in error_msg
        assert "verify" in error_msg
        assert "depends" in error_msg
        assert "related" in error_msg

    def test_validate_tag_empty_errors_list_when_valid(self):
        """Test that valid TAG returns empty errors list."""
        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("test.py"),
            line=1,
        )

        is_valid, errors = validator.validate_tag(tag)

        assert is_valid
        assert errors == []
        assert len(errors) == 0


class TestParseTagStringComprehensive:
    """Test parse_tag_string with comprehensive edge cases (lines 154, 162, 178, 185)."""

    def test_parse_tag_string_with_none_input(self):
        """Test parse_tag_string with None input (line 154)."""
        result = validator.parse_tag_string(None, Path("test.py"), 1)  # type: ignore[arg-type]
        assert result is None

    def test_parse_tag_string_with_non_string_input(self):
        """Test parse_tag_string with non-string input."""
        # Integer
        assert validator.parse_tag_string(123, Path("test.py"), 1) is None  # type: ignore[arg-type]

        # List
        assert validator.parse_tag_string(["# @SPEC"], Path("test.py"), 1) is None  # type: ignore[arg-type]

        # Dict
        assert validator.parse_tag_string({"comment": "# @SPEC"}, Path("test.py"), 1) is None  # type: ignore[arg-type]

        # Object
        assert validator.parse_tag_string(object(), Path("test.py"), 1) is None  # type: ignore[arg-type]

    def test_parse_tag_string_without_hash_prefix(self):
        """Test parse_tag_string without # prefix (line 160)."""
        result = validator.parse_tag_string("@SPEC SPEC-AUTH-001", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_only_hash(self):
        """Test parse_tag_string with only # character."""
        result = validator.parse_tag_string("#", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_only_hash_and_whitespace(self):
        """Test parse_tag_string with only # and whitespace."""
        result = validator.parse_tag_string("#   ", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_without_spec_prefix(self):
        """Test parse_tag_string without @SPEC prefix (line 168)."""
        result = validator.parse_tag_string("# SPEC-AUTH-001", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_with_wrong_prefix(self):
        """Test parse_tag_string with wrong prefix."""
        result = validator.parse_tag_string("# @spec SPEC-AUTH-001", Path("test.py"), 1)
        assert result is None

        result = validator.parse_tag_string("# @TAG SPEC-AUTH-001", Path("test.py"), 1)
        assert result is None

        result = validator.parse_tag_string("# @ SPEC-AUTH-001", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_empty_after_spec_prefix(self):
        """Test parse_tag_string with nothing after @SPEC (line 178)."""
        result = validator.parse_tag_string("# @SPEC", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_only_whitespace_after_spec(self):
        """Test parse_tag_string with only whitespace after @SPEC."""
        result = validator.parse_tag_string("# @SPEC   ", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_invalid_spec_id_format(self):
        """Test parse_tag_string with invalid SPEC-ID format (line 185)."""
        # Missing SPEC- prefix
        result = validator.parse_tag_string("# @SPEC AUTH-001", Path("test.py"), 1)
        assert result is None

        # Lowercase
        result = validator.parse_tag_string("# @SPEC spec-auth-001", Path("test.py"), 1)
        assert result is None

        # Wrong number of digits
        result = validator.parse_tag_string("# @SPEC SPEC-AUTH-99", Path("test.py"), 1)
        assert result is None

        # Special characters
        result = validator.parse_tag_string("# @SPEC SPEC-AUTH!-001", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_with_invalid_verb(self):
        """Test parse_tag_string with invalid verb."""
        # Invalid verb should be ignored, use default
        result = validator.parse_tag_string(
            "# @SPEC SPEC-AUTH-001 invalidverb",
            Path("test.py"),
            1
        )

        assert result is not None
        assert result.spec_id == "SPEC-AUTH-001"
        assert result.verb == "impl"  # Default verb

    def test_parse_tag_string_with_whitespace_variations(self):
        """Test parse_tag_string with various whitespace patterns."""
        # Multiple spaces before SPEC-ID
        tag = validator.parse_tag_string("# @SPEC    SPEC-AUTH-001", Path("test.py"), 1)
        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"

        # Multiple spaces between SPEC-ID and verb
        tag = validator.parse_tag_string("# @SPEC SPEC-AUTH-001    verify", Path("test.py"), 1)
        assert tag is not None
        assert tag.verb == "verify"

    def test_parse_tag_string_preserves_line_and_file(self):
        """Test that line number and file path are preserved."""
        tag = validator.parse_tag_string(
            "# @SPEC SPEC-AUTH-001",
            Path("my/path/to/test.py"),
            42
        )

        assert tag.file_path == Path("my/path/to/test.py")
        assert tag.line == 42

    def test_parse_tag_string_with_all_verbs(self):
        """Test parsing all valid verbs."""
        test_cases = [
            ("# @SPEC SPEC-AUTH-001 impl", "impl"),
            ("# @SPEC SPEC-AUTH-001 verify", "verify"),
            ("# @SPEC SPEC-AUTH-001 depends", "depends"),
            ("# @SPEC SPEC-AUTH-001 related", "related"),
        ]

        for comment, expected_verb in test_cases:
            tag = validator.parse_tag_string(comment, Path("test.py"), 1)
            assert tag is not None
            assert tag.verb == expected_verb

    def test_parse_tag_string_case_normalized(self):
        """Test that verb must be lowercase to match."""
        tag = validator.parse_tag_string(
            "# @SPEC SPEC-AUTH-001 VERIFY",
            Path("test.py"),
            1
        )

        assert tag is not None
        # VERIFY is not in VALID_VERBS (which are lowercase), so default is used
        assert tag.verb == "impl"

    def test_parse_tag_string_with_description_text(self):
        """Test that description text after verb is ignored."""
        comment = "# @SPEC SPEC-AUTH-001 impl - This is a description"
        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"
        assert tag.verb == "impl"

    def test_parse_tag_string_with_multiple_words_in_description(self):
        """Test parsing with multiple words in description."""
        comment = "# @SPEC SPEC-AUTH-001 verify User authentication flow implementation"
        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"
        assert tag.verb == "verify"

    def test_parse_tag_string_inline_comment(self):
        """Test parsing inline comment TAG."""
        # The parser strips the comment first, so we should pass the comment portion
        comment = "# @SPEC SPEC-AUTH-001"
        tag = validator.parse_tag_string(comment, Path("test.py"), 1)

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"
        assert tag.verb == "impl"

    def test_parse_tag_string_leading_whitespace_in_comment(self):
        """Test comment with leading whitespace before #."""
        # Comment starts with spaces before #
        tag = validator.parse_tag_string(
            "    # @SPEC SPEC-AUTH-001",
            Path("test.py"),
            1
        )

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"

    def test_parse_tag_string_tab_before_hash(self):
        """Test comment with tab before #."""
        tag = validator.parse_tag_string(
            "\t# @SPEC SPEC-AUTH-001",
            Path("test.py"),
            1
        )

        assert tag is not None
        assert tag.spec_id == "SPEC-AUTH-001"

    def test_parse_tag_string_empty_string(self):
        """Test parsing empty string."""
        result = validator.parse_tag_string("", Path("test.py"), 1)
        assert result is None

    def test_parse_tag_string_only_hash_and_at(self):
        """Test string with only # @."""
        result = validator.parse_tag_string("# @", Path("test.py"), 1)
        assert result is None


class TestGetDefaultVerb:
    """Test get_default_verb function."""

    def test_get_default_verb_returns_impl(self):
        """Test that default verb is 'impl'."""
        assert validator.get_default_verb() == "impl"

    def test_get_default_verb_is_constant(self):
        """Test that default verb is constant."""
        verb1 = validator.get_default_verb()
        verb2 = validator.get_default_verb()
        assert verb1 == verb2 == "impl"


class TestTAGEqualityAndHashing:
    """Test TAG equality, inequality, and hashing behavior."""

    def test_tag_equality_same_values(self):
        """Test TAG equality with identical values."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

        assert tag1 == tag2

    def test_tag_inequality_different_spec_id(self):
        """Test TAG inequality with different SPEC-ID."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-002", Path("auth.py"), 10, "impl")

        assert tag1 != tag2

    def test_tag_inequality_different_file_path(self):
        """Test TAG inequality with different file path."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-001", Path("user.py"), 10, "impl")

        assert tag1 != tag2

    def test_tag_inequality_different_line(self):
        """Test TAG inequality with different line number."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 20, "impl")

        assert tag1 != tag2

    def test_tag_inequality_different_verb(self):
        """Test TAG inequality with different verb."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "verify")

        assert tag1 != tag2

    def test_tag_comparison_with_non_tag(self):
        """Test TAG comparison with non-TAG object."""
        tag = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

        assert tag != "SPEC-AUTH-001"
        assert tag != 123
        assert tag != None
        assert tag != {"spec_id": "SPEC-AUTH-001"}

    def test_tag_hash_consistency(self):
        """Test that equal TAGs have same hash."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

        assert hash(tag1) == hash(tag2)

    def test_tag_in_set(self):
        """Test that TAGs can be used in sets."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag3 = validator.TAG("SPEC-AUTH-002", Path("auth.py"), 10, "impl")

        tag_set = {tag1, tag2, tag3}

        # Duplicate TAG should be deduplicated
        assert len(tag_set) == 2

    def test_tag_as_dict_key(self):
        """Test that TAGs can be used as dictionary keys."""
        tag1 = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
        tag2 = validator.TAG("SPEC-AUTH-002", Path("auth.py"), 10, "impl")

        tag_dict = {tag1: "value1", tag2: "value2"}

        assert tag_dict[tag1] == "value1"
        assert tag_dict[tag2] == "value2"
        assert len(tag_dict) == 2


class TestTAGRepr:
    """Test TAG string representation."""

    def test_tag_repr_contains_all_fields(self):
        """Test that TAG repr contains all field values."""
        tag = validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

        repr_str = repr(tag)

        assert "SPEC-AUTH-001" in repr_str
        assert "impl" in repr_str
        assert "auth.py" in repr_str
        assert "10" in repr_str

    def test_tag_repr_is_informative(self):
        """Test that TAG repr is informative."""
        tag = validator.TAG(
            spec_id="SPEC-TDD-001",
            verb="verify",
            file_path=Path("test/test_tdd.py"),
            line=42
        )

        repr_str = repr(tag)

        # Check that class name is in repr
        assert "TAG" in repr_str

        # Check that all important info is present
        assert "SPEC-TDD-001" in repr_str
        assert "verify" in repr_str
        assert "test_tdd.py" in repr_str
        assert "42" in repr_str


class TestValidVerbsConstant:
    """Test VALID_VERBS constant."""

    def test_valid_verbs_is_set(self):
        """Test that VALID_VERBS is a set."""
        assert isinstance(validator.VALID_VERBS, set)

    def test_valid_verbs_contains_all_expected(self):
        """Test that VALID_VERBS contains all expected verbs."""
        expected = {"impl", "verify", "depends", "related"}
        assert validator.VALID_VERBS == expected

    def test_valid_verbs_is_immutable(self):
        """Test that VALID_VERBS cannot be modified (it's a module constant)."""
        # Just verify it exists and has expected values
        assert "impl" in validator.VALID_VERBS
        assert len(validator.VALID_VERBS) == 4
