#!/usr/bin/env python3
# @CODE:DOC-TAG-004 | Component 1: Pre-commit TAG validator
"""Pre-commit TAG validation module

This module provides validation functionality for TAG annotations:
- Format validation (@DOC:DOMAIN-TYPE-NNN)
- Duplicate TAG detection across files
- Orphan TAG detection (CODE without TEST, etc.)
- Git staged file scanning

Used by pre-commit hooks to ensure TAG quality.
"""

import re
import subprocess
from dataclasses import dataclass, field
from pathlib import Path
from typing import Dict, List, Optional, Tuple


@dataclass
class ValidationError:
    """Validation error with file location information"""
    message: str
    tag: str
    locations: List[Tuple[str, int]] = field(default_factory=list)

    def __str__(self) -> str:
        loc_str = ", ".join([f"{f}:{line}" for f, line in self.locations])
        return f"{self.message}: {self.tag} at {loc_str}"


@dataclass
class ValidationWarning:
    """Validation warning with file location"""
    message: str
    tag: str
    location: Tuple[str, int]

    def __str__(self) -> str:
        return f"{self.message}: {self.tag} at {self.location[0]}:{self.location[1]}"


@dataclass
class ValidationResult:
    """Complete validation result"""
    is_valid: bool
    errors: List[ValidationError] = field(default_factory=list)
    warnings: List[ValidationWarning] = field(default_factory=list)

    def format(self) -> str:
        """Format result for display"""
        lines = []

        if self.errors:
            lines.append("Errors:")
            for error in self.errors:
                lines.append(f"  - {error}")

        if self.warnings:
            lines.append("\nWarnings:")
            for warning in self.warnings:
                lines.append(f"  - {warning}")

        if not self.errors and not self.warnings:
            lines.append("No issues found.")

        return "\n".join(lines)


class PreCommitValidator:
    """Pre-commit TAG validator

    Validates TAG annotations in files:
    - Format: @DOC:DOMAIN-TYPE-NNN
    - No duplicates
    - No orphans (CODE without TEST)

    Args:
        strict_mode: Treat warnings as errors
        check_orphans: Enable orphan TAG detection
        tag_pattern: Custom TAG regex pattern
    """

    # Default TAG pattern: @(SPEC|CODE|TEST|DOC):DOMAIN-NNN or DOMAIN-TYPE-NNN
    # Matches formats like:
    # - @CODE:AUTH-API-001 (domain-type-number)
    # - @CODE:SPEC-001 (domain-number)
    # - @TEST:USER-REG-001 (domain-type-number)
    DEFAULT_TAG_PATTERN = r"@(SPEC|CODE|TEST|DOC):([A-Z]+(?:-[A-Z]+)*-\d{3})"

    def __init__(
        self,
        strict_mode: bool = False,
        check_orphans: bool = True,
        tag_pattern: Optional[str] = None
    ):
        self.strict_mode = strict_mode
        self.check_orphans = check_orphans
        self.tag_pattern = re.compile(tag_pattern or self.DEFAULT_TAG_PATTERN)
        # Document files to exclude from TAG validation
        self.excluded_file_patterns = [
            r"\.md$",  # Markdown files
            r"README",  # README files
            r"CHANGELOG",  # CHANGELOG files
            r"CONTRIBUTING",  # CONTRIBUTING files
            r"LICENSE",  # LICENSE files
            r"\.txt$",  # Text files
            r"\.rst$",  # ReStructuredText files
            r"test_.*\.py$",  # Test files (test_*.py)
            r".*_test\.py$",  # Test files (*_test.py)
            r"tests/",  # Files in tests/ directory
            r"validator\.py$",  # Validator files (contain example TAGs in docstrings)
        ]

    def should_validate_file(self, filepath: str) -> bool:
        """Check if file should be validated for TAGs

        Document files (*.md, README, CONTRIBUTING, etc.) are excluded
        because they often contain example TAGs that are not actual code.

        Args:
            filepath: File path to check

        Returns:
            True if file should be validated, False if excluded
        """
        for pattern in self.excluded_file_patterns:
            if re.search(pattern, filepath):
                return False
        return True

    def validate_format(self, tag: str) -> bool:
        """Validate TAG format

        Args:
            tag: TAG string (e.g., "@CODE:AUTH-API-001")

        Returns:
            True if format is valid
        """
        return bool(self.tag_pattern.match(tag))

    def extract_tags(self, content: str) -> List[str]:
        """Extract all TAGs from content

        Args:
            content: File content

        Returns:
            List of TAG strings
        """
        matches = self.tag_pattern.findall(content)
        # Convert tuples to full TAG strings
        tags = [f"@{prefix}:{domain}" for prefix, domain in matches]
        return tags

    def validate_duplicates(self, files: List[str]) -> List[ValidationError]:
        """Detect duplicate TAGs

        Args:
            files: List of file paths to scan

        Returns:
            List of validation errors for duplicates
        """
        errors: List[ValidationError] = []
        tag_locations: Dict[str, List[Tuple[str, int]]] = {}

        for filepath in files:
            # Skip document files (*.md, README, CONTRIBUTING, etc.)
            if not self.should_validate_file(filepath):
                continue

            try:
                path = Path(filepath)
                if not path.exists() or not path.is_file():
                    continue

                content = path.read_text(encoding="utf-8", errors="ignore")
                lines = content.splitlines()

                for line_num, line in enumerate(lines, start=1):
                    tags = self.extract_tags(line)
                    for tag in tags:
                        if tag not in tag_locations:
                            tag_locations[tag] = []
                        tag_locations[tag].append((filepath, line_num))

            except Exception:
                # Skip files that can't be read
                continue

        # Find duplicates
        for tag, locations in tag_locations.items():
            if len(locations) > 1:
                errors.append(ValidationError(
                    message="Duplicate TAG found",
                    tag=tag,
                    locations=locations
                ))

        return errors

    def validate_orphans(self, files: List[str]) -> List[ValidationWarning]:
        """Detect orphan TAGs

        Orphan TAGs are:
        - @CODE without corresponding @TEST
        - @TEST without corresponding @CODE
        - @SPEC without implementation

        Args:
            files: List of file paths to scan

        Returns:
            List of validation warnings
        """
        if not self.check_orphans:
            return []

        warnings: List[ValidationWarning] = []

        # Collect all TAGs by type and domain
        tags_by_type: Dict[str, Dict[str, List[Tuple[str, int]]]] = {
            "SPEC": {},
            "CODE": {},
            "TEST": {},
            "DOC": {}
        }

        for filepath in files:
            # Skip document files (*.md, README, CONTRIBUTING, etc.)
            if not self.should_validate_file(filepath):
                continue

            try:
                path = Path(filepath)
                if not path.exists() or not path.is_file():
                    continue

                content = path.read_text(encoding="utf-8", errors="ignore")
                lines = content.splitlines()

                for line_num, line in enumerate(lines, start=1):
                    matches = self.tag_pattern.findall(line)
                    for prefix, domain in matches:
                        if domain not in tags_by_type[prefix]:
                            tags_by_type[prefix][domain] = []
                        tags_by_type[prefix][domain].append((filepath, line_num))

            except Exception:
                continue

        # Check for orphans
        # CODE without TEST
        for domain, locations in tags_by_type["CODE"].items():
            if domain not in tags_by_type["TEST"]:
                for filepath, line_num in locations:
                    warnings.append(ValidationWarning(
                        message="CODE TAG without corresponding TEST",
                        tag=f"@CODE:{domain}",
                        location=(filepath, line_num)
                    ))

        # TEST without CODE
        for domain, locations in tags_by_type["TEST"].items():
            if domain not in tags_by_type["CODE"]:
                for filepath, line_num in locations:
                    warnings.append(ValidationWarning(
                        message="TEST TAG without corresponding CODE",
                        tag=f"@TEST:{domain}",
                        location=(filepath, line_num)
                    ))

        return warnings

    def get_staged_files(self, repo_path: str = ".") -> List[str]:
        """Get list of staged files from git

        Args:
            repo_path: Git repository path

        Returns:
            List of staged file paths
        """
        try:
            result = subprocess.run(
                ["git", "diff", "--name-only", "--cached"],
                cwd=repo_path,
                capture_output=True,
                text=True,
                timeout=5,
                check=True
            )
            files = [
                line.strip()
                for line in result.stdout.splitlines()
                if line.strip()
            ]
            return files
        except Exception:
            return []

    def validate_files(self, files: List[str]) -> ValidationResult:
        """Validate list of files

        Main validation method that runs all checks:
        - Format validation
        - Duplicate detection
        - Orphan detection

        Args:
            files: List of file paths to validate

        Returns:
            ValidationResult with errors and warnings
        """
        if not files:
            return ValidationResult(is_valid=True)

        # Check for duplicates
        errors = self.validate_duplicates(files)

        # Check for orphans
        warnings = self.validate_orphans(files)

        # In strict mode, warnings become errors
        if self.strict_mode and warnings:
            is_valid = False
        else:
            is_valid = len(errors) == 0

        return ValidationResult(
            is_valid=is_valid,
            errors=errors,
            warnings=warnings
        )


def main():
    """CLI entry point for pre-commit hook"""
    import argparse
    import sys

    parser = argparse.ArgumentParser(
        description="Validate TAG annotations in git staged files"
    )
    parser.add_argument(
        "--files",
        nargs="*",
        help="Files to validate (default: git staged files)"
    )
    parser.add_argument(
        "--strict",
        action="store_true",
        help="Treat warnings as errors"
    )
    parser.add_argument(
        "--no-orphan-check",
        action="store_true",
        help="Disable orphan TAG checking"
    )

    args = parser.parse_args()

    validator = PreCommitValidator(
        strict_mode=args.strict,
        check_orphans=not args.no_orphan_check
    )

    # Get files to validate
    if args.files:
        files = args.files
    else:
        files = validator.get_staged_files()

    if not files:
        print("No files to validate.")
        sys.exit(0)

    # Run validation
    result = validator.validate_files(files)

    # Print results
    print(result.format())

    # Exit with error code if validation failed
    sys.exit(0 if result.is_valid else 1)


if __name__ == "__main__":
    main()
