"""
@TASK:VALIDATOR-001 SpecValidator Module
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:VALIDATOR-001

Input validation module following TRUST principles:
- T: Test-driven design
- R: Readable validation logic
- U: Unified single responsibility (validation only)
- S: Secure input sanitization
- T: Trackable validation rules
"""

import logging

# Logging setup
logger = logging.getLogger(__name__)


class SpecValidator:
    """
    @TASK:INPUT-VALIDATION-001 Input validation for SPEC command

    TRUST principles applied:
    - T: Testable validation methods
    - R: Clear validation error messages
    - U: Single responsibility (validation only)
    - S: Security-focused input sanitization
    - T: Trackable validation rules
    """

    def __init__(self):
        """Initialize SpecValidator with security rules"""
        self.unsafe_chars = ["/", "\\", "<", ">", ":", '"', "|", "?", "*"]
        self.max_spec_name_length = 50
        self.max_description_length = 500

        logger.debug("SpecValidator initialized with security rules")

    def validate_spec_name(self, spec_name: str) -> str:
        """Validate and normalize spec name

        Args:
            spec_name: Spec name to validate

        Returns:
            Normalized spec name

        Raises:
            ValueError: If spec name is invalid
        """
        # Type and empty check
        if not spec_name or not isinstance(spec_name, str):
            raise ValueError("spec_name은 비어있지 않은 문자열이어야 합니다")

        # Normalize whitespace
        normalized = spec_name.strip()

        # Check if empty after strip
        if not normalized:
            raise ValueError("spec_name은 비어있지 않은 문자열이어야 합니다")

        # Convert to uppercase for consistency
        normalized = normalized.upper()

        # Length check
        if len(normalized) > self.max_spec_name_length:
            raise ValueError("명세 이름이 너무 깁니다 (최대 50자)")

        # Security check for unsafe characters
        if any(char in normalized for char in self.unsafe_chars):
            raise ValueError(
                f"명세 이름에 안전하지 않은 문자가 포함되어 있습니다: {spec_name}"
            )

        logger.debug(f"Spec name validated: {spec_name} -> {normalized}")
        return normalized

    def validate_description(self, description: str) -> str:
        """Validate and normalize description

        Args:
            description: Description to validate

        Returns:
            Normalized description

        Raises:
            ValueError: If description is invalid
        """
        # Type and empty check
        if not description or not isinstance(description, str):
            raise ValueError("description은 비어있지 않은 문자열이어야 합니다")

        # Normalize whitespace
        normalized = description.strip()

        # Check if empty after strip
        if not normalized:
            raise ValueError("description은 비어있지 않은 문자열이어야 합니다")

        # Length check
        if len(normalized) > self.max_description_length:
            raise ValueError("설명이 너무 깁니다 (최대 500자)")

        logger.debug(f"Description validated: {len(normalized)} chars")
        return normalized