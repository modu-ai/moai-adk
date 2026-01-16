"""TAG validator module for TAG System v2.0.

Implements T1: TAG Pattern Definition with format validation,
SPEC-ID validation, and verb validation.
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import List, Tuple

# TAG pattern: @SPEC SPEC-ID [verb]
# SPEC-ID format: SPEC-{DOMAIN}-{NUMBER}
# DOMAIN: Uppercase letters, at least 1 char
# NUMBER: 3 digits
SPEC_ID_PATTERN = re.compile(r"^SPEC-[A-Z]+-\d{3}$")

# Valid TAG verbs
VALID_VERBS = {"impl", "verify", "depends", "related"}

# Default verb when not specified
DEFAULT_VERB = "impl"


@dataclass(frozen=True)
class TAG:
    """Represents a single TAG annotation.

    Attributes:
        spec_id: SPEC identifier (e.g., "SPEC-AUTH-001")
        verb: TAG relationship verb (impl, verify, depends, related)
        file_path: Path to the file containing the TAG
        line: Line number where TAG appears
    """

    spec_id: str
    file_path: Path
    line: int
    verb: str = field(default=DEFAULT_VERB)

    def __post_init__(self) -> None:
        """Validate TAG fields after initialization."""
        # Normalize verb to lowercase
        if self.verb != self.verb.lower():
            object.__setattr__(self, "verb", self.verb.lower())


def validate_spec_id_format(spec_id: str) -> bool:
    """Validate SPEC-ID format: SPEC-{DOMAIN}-{NUMBER}.

    Args:
        spec_id: SPEC identifier to validate

    Returns:
        True if SPEC-ID matches required format, False otherwise

    Examples:
        >>> validate_spec_id_format("SPEC-AUTH-001")
        True
        >>> validate_spec_id_format("spec-auth-001")
        False
        >>> validate_spec_id_format("SPEC-AUTH-1")
        False
    """
    return isinstance(spec_id, str) and bool(SPEC_ID_PATTERN.match(spec_id))


def validate_verb(verb: str) -> bool:
    """Validate TAG verb.

    Args:
        verb: Verb to validate

    Returns:
        True if verb is valid, False otherwise

    Examples:
        >>> validate_verb("impl")
        True
        >>> validate_verb("invalid")
        False
    """
    return isinstance(verb, str) and verb in VALID_VERBS


def get_default_verb() -> str:
    """Get default TAG verb.

    Returns:
        Default verb string ("impl")
    """
    return DEFAULT_VERB


def validate_tag(tag: TAG) -> Tuple[bool, List[str]]:
    """Validate complete TAG.

    Args:
        tag: TAG to validate

    Returns:
        Tuple of (is_valid, error_messages)
        - is_valid: True if TAG is valid
        - error_messages: List of validation errors (empty if valid)

    Examples:
        >>> tag = TAG("SPEC-AUTH-001", Path("test.py"), 1, "impl")
        >>> is_valid, errors = validate_tag(tag)
        >>> is_valid
        True
    """
    errors = []

    # Validate SPEC-ID format
    if not validate_spec_id_format(tag.spec_id):
        errors.append(
            f"Invalid SPEC-ID format: '{tag.spec_id}'. "
            f"Expected format: SPEC-{{DOMAIN}}-{{NUMBER}} (e.g., SPEC-AUTH-001)"
        )

    # Validate verb
    if not validate_verb(tag.verb):
        errors.append(f"Invalid verb: '{tag.verb}'. Valid verbs: {', '.join(sorted(VALID_VERBS))}")

    return len(errors) == 0, errors


def parse_tag_string(comment: str, file_path: Path, line: int) -> TAG | None:
    """Parse TAG from comment string.

    Args:
        comment: Comment string to parse
        file_path: Path to the file containing the comment
        line: Line number of the comment

    Returns:
        TAG object if valid TAG found, None otherwise

    Examples:
        >>> comment = "# @SPEC SPEC-AUTH-001"
        >>> tag = parse_tag_string(comment, Path("test.py"), 1)
        >>> tag.spec_id
        'SPEC-AUTH-001'
        >>> tag.verb
        'impl'
    """
    if not isinstance(comment, str):
        return None

    # Remove leading whitespace and comment character
    comment = comment.strip()

    # Check if comment starts with #
    if not comment.startswith("#"):
        # Not a comment
        return None

    # Remove # and leading whitespace
    comment = comment[1:].strip()

    # Check for @SPEC prefix
    if not comment.startswith("@SPEC"):
        return None

    # Remove @SPEC prefix
    comment = comment[5:].strip()  # len("@SPEC") == 5

    # Split into parts
    parts = comment.split()

    if not parts:
        return None

    # First part is SPEC-ID
    spec_id = parts[0]

    # Validate SPEC-ID format
    if not validate_spec_id_format(spec_id):
        return None

    # Second part (if present) is verb
    verb = DEFAULT_VERB
    if len(parts) > 1 and parts[1] in VALID_VERBS:
        verb = parts[1]

    return TAG(spec_id=spec_id, file_path=file_path, line=line, verb=verb)


__all__ = [
    "TAG",
    "VALID_VERBS",
    "DEFAULT_VERB",
    "validate_spec_id_format",
    "validate_verb",
    "get_default_verb",
    "validate_tag",
    "parse_tag_string",
]
