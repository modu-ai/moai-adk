#!/usr/bin/env python3
# @CODE:TAG-AUTO-CORRECTOR-001 | @SPEC:TAG-AUTO-001
# type: ignore
"""TAG error automatic correction system

System for automatically correcting TAG policy violations and generating smart TAGs.
Integrates with Post-Tool-Use hooks for real-time TAG error correction.

Key Features:
- Automatic generation of missing TAGs
- Automatic removal of duplicate TAGs
- Automatic TAG chain connection recovery
- Smart TAG suggestion system
"""

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple

from .policy_validator import PolicyViolation, PolicyViolationType


@dataclass
class AutoCorrection:
    """Automatic correction information

    Attributes:
        file_path: Path to file to be corrected
        original_content: Original content
        corrected_content: Corrected content
        description: Correction description
        confidence: Correction confidence (0.0-1.0)
        requires_review: Whether manual review is required
    """
    file_path: str
    original_content: str
    corrected_content: str
    description: str
    confidence: float
    requires_review: bool = False


@dataclass
class AutoCorrectionConfig:
    """Automatic correction configuration

    Attributes:
        enable_auto_fix: Enable automatic correction
        confidence_threshold: Minimum confidence for automatic application
        create_missing_specs: Automatically create missing SPECs
        create_missing_tests: Automatically create missing TESTs
        remove_duplicates: Automatically remove duplicate TAGs
        backup_before_fix: Create backup before correction
    """
    enable_auto_fix: bool = False
    confidence_threshold: float = 0.8
    create_missing_specs: bool = False
    create_missing_tests: bool = False
    remove_duplicates: bool = True
    backup_before_fix: bool = True


class TagAutoCorrector:
    """TAG error automatic corrector

    Called from Post-Tool-Use hooks to automatically correct TAG policy violations.
    Generates and suggests optimal TAGs using smart algorithms.

    Usage:
        config = AutoCorrectionConfig(enable_auto_fix=True)
        corrector = TagAutoCorrector(config=config)

        violations = [...]
        corrections = corrector.generate_corrections(violations)

        if corrections:
            corrector.apply_corrections(corrections)
    """

    # TAG regex pattern
    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z0-9-]+-\d{3})")
    SHEBANG_PATTERN = re.compile(r"^#!.*\n")

    def __init__(self, config: Optional[AutoCorrectionConfig] = None):
        """Initialize corrector

        Args:
            config: Automatic correction configuration (default: AutoCorrectionConfig())
        """
        self.config = config or AutoCorrectionConfig()

    def generate_corrections(self, violations: List[PolicyViolation]) -> List[AutoCorrection]:
        """Generate automatic corrections for policy violations

        Args:
            violations: List of policy violations

        Returns:
            List of AutoCorrection
        """
        corrections = []

        # Group violations by file
        violations_by_file = self._group_violations_by_file(violations)

        for file_path, file_violations in violations_by_file.items():
            try:
                # Read file content
                path = Path(file_path)
                if not path.exists():
                    continue

                original_content = path.read_text(encoding="utf-8", errors="ignore")
                corrected_content = original_content

                # Apply correction for each violation
                for violation in file_violations:
                    correction = self._generate_single_correction(
                        file_path, corrected_content, violation
                    )
                    if correction:
                        corrected_content = correction.corrected_content
                        corrections.append(correction)

            except Exception:
                # Skip on file read failure
                continue

        return corrections

    def apply_corrections(self, corrections: List[AutoCorrection]) -> bool:
        """Apply automatic corrections

        Args:
            corrections: List of corrections to apply

        Returns:
            Success status
        """
        if not self.config.enable_auto_fix:
            return False

        success_count = 0
        for correction in corrections:
            if correction.confidence >= self.config.confidence_threshold:
                try:
                    # Create backup
                    if self.config.backup_before_fix:
                        self._create_backup(correction.file_path, correction.original_content)

                    # Apply correction
                    path = Path(correction.file_path)
                    path.write_text(correction.corrected_content, encoding="utf-8")
                    success_count += 1

                except Exception:
                    # Skip on correction failure
                    continue

        return success_count == len(corrections)

    def suggest_tag_for_code_file(self, file_path: str) -> Optional[Tuple[str, float]]:
        """Suggest TAG for code file

        Args:
            file_path: Code file path

        Returns:
            (TAG, confidence) tuple or None
        """
        path = Path(file_path)

        # Extract domain from file path
        domain = self._extract_domain_from_path(path)
        if not domain:
            return None

        # Check existing TAGs
        existing_tags = self._find_existing_tags_in_project(domain)

        # Calculate next number
        next_number = self._calculate_next_number(existing_tags, domain)

        tag = f"@CODE:{domain}-{next_number:03d}"
        confidence = self._calculate_tag_confidence(domain, file_path)

        return tag, confidence

    def _group_violations_by_file(self, violations: List[PolicyViolation]) -> Dict[str, List[PolicyViolation]]:
        """Group policy violations by file

        Args:
            violations: List of policy violations

        Returns:
            Dictionary of {file_path: [violations]}
        """
        grouped = {}
        for violation in violations:
            if violation.file_path:
                if violation.file_path not in grouped:
                    grouped[violation.file_path] = []
                grouped[violation.file_path].append(violation)
        return grouped

    def _generate_single_correction(self, file_path: str, content: str,
                                  violation: PolicyViolation) -> Optional[AutoCorrection]:
        """Generate correction for single policy violation

        Args:
            file_path: File path
            content: Current file content
            violation: Policy violation

        Returns:
            AutoCorrection or None
        """
        if violation.type == PolicyViolationType.MISSING_TAGS:
            return self._fix_missing_tags(file_path, content, violation)
        elif violation.type == PolicyViolationType.DUPLICATE_TAGS:
            return self._fix_duplicate_tags(file_path, content, violation)
        elif violation.type == PolicyViolationType.NO_SPEC_REFERENCE:
            return self._fix_missing_spec_reference(file_path, content, violation)
        elif violation.type == PolicyViolationType.CHAIN_BREAK:
            return self._fix_chain_break(file_path, content, violation)

        return None

    def _fix_missing_tags(self, file_path: str, content: str,
                         violation: PolicyViolation) -> Optional[AutoCorrection]:
        """Fix missing TAGs

        Args:
            file_path: File path
            content: File content
            violation: Policy violation

        Returns:
            AutoCorrection or None
        """
        # Suggest TAG for code file
        tag_suggestion = self.suggest_tag_for_code_file(file_path)
        if not tag_suggestion:
            return None

        tag, confidence = tag_suggestion

        # Find TAG insertion position
        insert_position = self._find_tag_insert_position(content)

        # Generate TAG comment
        tag_comment = f"# {tag}\n"

        # Insert TAG into content
        lines = content.splitlines()
        if insert_position is not None:
            lines.insert(insert_position, tag_comment.strip())
        else:
            # Insert at file start (after shebang)
            if self.SHEBANG_PATTERN.match(content):
                shebang_line = lines[0]
                lines = [shebang_line, "", tag_comment.strip()] + lines[1:]
            else:
                lines = [tag_comment.strip()] + lines

        corrected_content = "\n".join(lines) + "\n"

        return AutoCorrection(
            file_path=file_path,
            original_content=content,
            corrected_content=corrected_content,
            description=f"Auto-add @TAG: {tag}",
            confidence=confidence,
            requires_review=confidence < 0.9
        )

    def _fix_duplicate_tags(self, file_path: str, content: str,
                           violation: PolicyViolation) -> Optional[AutoCorrection]:
        """Fix duplicate TAGs

        Args:
            file_path: File path
            content: File content
            violation: Policy violation

        Returns:
            AutoCorrection or None
        """
        if not violation.tag:
            return None

        # Remove duplicate TAGs (keep only first)
        tag = violation.tag
        corrected_content = content

        # Find all TAGs with regex
        pattern = re.compile(re.escape(tag))
        matches = list(pattern.finditer(content))

        if len(matches) <= 1:
            return None

        # Remove all TAGs except the first
        # Process in reverse order to prevent index shift issues
        for match in reversed(matches[1:]):
            start, end = match.span()
            line_start = content.rfind('\n', 0, start) + 1
            line_end = content.find('\n', end)
            if line_end == -1:
                line_end = len(content)

            line = content[line_start:line_end]
            if line.strip() == f"#{tag}":
                # Remove line with TAG only
                corrected_content = corrected_content[:line_start] + corrected_content[line_end:]
            else:
                # Remove TAG as part of line
                corrected_content = (corrected_content[:start] +
                                  corrected_content[end:])

        return AutoCorrection(
            file_path=file_path,
            original_content=content,
            corrected_content=corrected_content,
            description=f"Remove duplicate TAG: {tag}",
            confidence=0.95,
            requires_review=False
        )

    def _fix_missing_spec_reference(self, file_path: str, content: str,
                                   violation: PolicyViolation) -> Optional[AutoCorrection]:
        """Fix missing SPEC reference

        Args:
            file_path: File path
            content: File content
            violation: Policy violation

        Returns:
            AutoCorrection or None
        """
        if not violation.tag or not self.config.create_missing_specs:
            return None

        # Extract domain from CODE TAG
        match = self.TAG_PATTERN.search(violation.tag)
        if not match:
            return None

        domain = match.group(2)

        # Automatically create SPEC file
        spec_created = self._create_spec_file(domain)
        if not spec_created:
            return None

        # Add SPEC reference to existing content
        corrected_content = self._add_spec_reference_to_content(content, domain)

        return AutoCorrection(
            file_path=file_path,
            original_content=content,
            corrected_content=corrected_content,
            description=f"Add SPEC reference: {domain}",
            confidence=0.8,
            requires_review=True
        )

    def _fix_chain_break(self, file_path: str, content: str,
                        violation: PolicyViolation) -> Optional[AutoCorrection]:
        """Fix chain break

        Args:
            file_path: File path
            content: File content
            violation: Policy violation

        Returns:
            AutoCorrection or None
        """
        if violation.type == PolicyViolationType.CHAIN_BREAK and violation.tag:
            # Generate TEST for CODE
            if "@CODE:" in violation.tag and self.config.create_missing_tests:
                return self._create_missing_test_file(violation.tag)

        return None

    def _find_tag_insert_position(self, content: str) -> Optional[int]:
        """Find TAG insertion position

        Args:
            content: File content

        Returns:
            Line number (0-based) or None
        """
        lines = content.splitlines()

        # Find position after shebang
        for i, line in enumerate(lines):
            if line.startswith('#!'):
                return i + 1

        # Find position after docstring (Python)
        for i, line in enumerate(lines):
            if line.strip().startswith('"""') or line.strip().startswith("'''"):
                # Find docstring end
                quote_char = '"""' if line.strip().startswith('"""') else "'''"
                if quote_char in line[line.find(quote_char)+3:]:
                    return i + 1
                else:
                    # Multi-line docstring
                    for j in range(i + 1, len(lines)):
                        if quote_char in lines[j]:
                            return j + 1

        # First empty line or after comment
        for i, line in enumerate(lines):
            if not line.strip() or line.strip().startswith('#'):
                return i

        return 0

    def _extract_domain_from_path(self, path: Path) -> Optional[str]:
        """Extract domain from file path

        Extract domain from file path:
        - test files (tests/ or test_*.py) → None
        - src/domain/... → domain (uppercase)
        - lib/domain/... → domain (uppercase)
        - filename (no parent dir) → filename stem (uppercase)

        Args:
            path: File path

        Returns:
            Domain string or None
        """
        parts = path.parts
        filename = path.name

        # Test files should not have domain extraction
        if "tests" in parts or filename.startswith("test_"):
            return None

        # For src/ paths, extract first directory after src
        if "src" in parts:
            src_index = parts.index("src")
            if src_index + 1 < len(parts):
                domain_part = parts[src_index + 1]
                return domain_part.upper().replace("_", "-")

        # For lib/ paths or other paths, extract parent directory
        if len(parts) > 1:
            parent_dir = parts[-2]  # Parent directory
            return parent_dir.upper().replace("_", "-")

        # For single files with no parent directory, use filename stem
        stem = path.stem.upper()
        if stem and re.match(r'^[A-Z-]+$', stem.replace("-", "")):
            return stem.replace("_", "-")

        return None

    def _find_existing_tags_in_project(self, domain: str) -> Set[str]:
        """Find existing TAGs in project

        Args:
            domain: Domain

        Returns:
            Set of existing TAG numbers
        """
        existing_numbers = set()

        # Search for TAGs from project root
        for pattern in ["**/*.py", "**/*.js", "**/*.ts", "**/*.md"]:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    try:
                        content = path.read_text(encoding="utf-8", errors="ignore")
                        matches = self.TAG_PATTERN.findall(content)
                        for tag_type, tag_domain in matches:
                            if tag_domain.startswith(domain):
                                # Extract number from domain-number format
                                if f"{domain}-" in tag_domain:
                                    number_part = tag_domain.replace(f"{domain}-", "")
                                    if number_part.isdigit():
                                        existing_numbers.add(int(number_part))
                    except Exception:
                        continue

        return existing_numbers

    def _calculate_next_number(self, existing_numbers: Set[int], domain: str) -> int:
        """Calculate next TAG number

        Args:
            existing_numbers: Set of existing TAG numbers
            domain: Domain

        Returns:
            Next number
        """
        if not existing_numbers:
            return 1

        # Find smallest empty number from 1 to 999
        for i in range(1, 1000):
            if i not in existing_numbers:
                return i

        # If all used, return last number + 1
        return max(existing_numbers) + 1

    def _calculate_tag_confidence(self, domain: str, file_path: str) -> float:
        """Calculate TAG confidence

        Args:
            domain: Domain
            file_path: File path

        Returns:
            Confidence (0.0-1.0)
        """
        confidence = 0.5  # Base confidence

        path = Path(file_path)

        # Increase confidence based on path
        if "src" in str(path):
            confidence += 0.2

        # Check domain match
        if domain.lower() in path.stem.lower():
            confidence += 0.2

        # Check file structure accuracy
        if path.suffix in ['.py', '.js', '.ts']:
            confidence += 0.1

        return min(confidence, 1.0)

    def _create_backup(self, file_path: str, content: str) -> None:
        """Create backup file

        Args:
            file_path: Original file path
            content: Original content
        """
        try:
            backup_path = Path(f"{file_path}.backup")
            backup_path.write_text(content, encoding="utf-8")
        except Exception:
            pass

    def _create_spec_file(self, domain: str) -> bool:
        """Automatically create SPEC file

        Args:
            domain: Domain

        Returns:
            Success status
        """
        try:
            spec_dir = Path(f".moai/specs/SPEC-{domain}")
            spec_dir.mkdir(parents=True, exist_ok=True)

            spec_file = spec_dir / "spec.md"
            if not spec_file.exists():
                spec_content = f"""# SPEC: {domain}

## Requirements

- [Detailed requirements]

## Implementation Guide

### TAG Links
- @SPEC:{domain} (current document)
- @CODE:{domain} (implementation file)
- @TEST:{domain} (test file)

### Acceptance Criteria
- [ ] Feature implementation complete
- [ ] Tests passing
- [ ] Documentation complete

## History

- Created: {Path('.').absolute().name}
- Status: In progress
"""
                spec_file.write_text(spec_content, encoding="utf-8")
                return True

        except Exception:
            pass

        return False

    def _add_spec_reference_to_content(self, content: str, domain: str) -> str:
        """Add SPEC reference to content

        Args:
            content: Original content
            domain: Domain

        Returns:
            Modified content
        """
        lines = content.splitlines()

        # Find existing CODE TAG
        for i, line in enumerate(lines):
            if f"@CODE:{domain}" in line:
                # Add SPEC reference
                spec_ref = f" | SPEC: .moai/specs/SPEC-{domain}/spec.md"
                if spec_ref not in line:
                    lines[i] = line + spec_ref
                break

        return "\n".join(lines) + "\n"

    def _create_missing_test_file(self, code_tag: str) -> Optional[AutoCorrection]:
        """Create missing test file

        Args:
            code_tag: CODE TAG

        Returns:
            AutoCorrection or None
        """
        match = self.TAG_PATTERN.search(code_tag)
        if not match:
            return None

        domain = match.group(2)

        try:
            test_dir = Path("tests")
            test_dir.mkdir(exist_ok=True)

            test_file = test_dir / f"test_{domain.lower()}.py"

            if not test_file.exists():
                test_content = f'''#!/usr/bin/env python3
# @TEST:{domain} | SPEC: .moai/specs/SPEC-{domain}/spec.md | CODE: @CODE:{domain}
"""Test cases for {domain} functionality.

This module tests the implementation defined in @CODE:{domain}.
"""

import pytest


class Test{domain.replace('-', '')}:
    """Test class for {domain} functionality."""

    def test_basic_functionality(self):
        """Test basic functionality."""
        # TODO: Implement test cases
        assert True

    def test_edge_cases(self):
        """Test edge cases."""
        # TODO: Implement edge case tests
        assert True

    def test_error_conditions(self):
        """Test error conditions."""
        # TODO: Implement error condition tests
        assert True
'''
                test_file.write_text(test_content, encoding="utf-8")

                return AutoCorrection(
                    file_path=str(test_file),
                    original_content="",
                    corrected_content=test_content,
                    description=f"Create test file: {domain}",
                    confidence=0.8,
                    requires_review=True
                )

        except Exception:
            pass

        return None
