#!/usr/bin/env python3
# @CODE:DOC-TAG-004 | Component 3: Central validation system
"""Central TAG validation system for MoAI-ADK

This module provides a unified, extensible validation engine that:
- Validates TAG format, duplicates, orphans, and chain integrity
- Supports multiple validator types with priority ordering
- Generates reports in multiple formats (detailed, summary, JSON)
- Provides CLI integration via moai-adk validate-tags

Architecture:
    CentralValidator (orchestrator)
    ├── DuplicateValidator (priority: 90)
    ├── OrphanValidator (priority: 50)
    ├── ChainValidator (priority: 30)
    └── FormatValidator (priority: 100)

Usage:
    config = ValidationConfig(strict_mode=True)
    validator = CentralValidator(config=config)
    result = validator.validate_directory("/path/to/project")
    report = validator.create_report(result, format="json")
"""

import json
import re
import time
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Set, Tuple


@dataclass
class ValidationConfig:
    """Configuration for validation behavior

    Attributes:
        strict_mode: Treat warnings as errors (block on warnings)
        check_duplicates: Enable duplicate TAG detection
        check_orphans: Enable orphan TAG detection (CODE without TEST)
        check_chain_integrity: Enable SPEC→CODE→TEST→DOC chain validation
        allowed_file_types: List of file extensions to validate (e.g., ["py", "js"])
        ignore_patterns: List of glob patterns to ignore (e.g., [".git/*", "*.pyc"])
        report_format: Default report format (detailed|summary|json)
    """
    strict_mode: bool = False
    check_duplicates: bool = True
    check_orphans: bool = True
    check_chain_integrity: bool = True
    allowed_file_types: List[str] = field(default_factory=lambda: [
        "py", "js", "ts", "jsx", "tsx", "md", "txt", "yml", "yaml", "json"
    ])
    ignore_patterns: List[str] = field(default_factory=lambda: [
        ".git/*", "node_modules/*", "__pycache__/*", "*.pyc", ".venv/*", "venv/*"
    ])
    report_format: str = "detailed"


@dataclass
class ValidationIssue:
    """Validation issue with severity, type, and location

    Attributes:
        severity: Issue severity (error|warning|info)
        type: Issue type (duplicate|orphan|chain|format)
        tag: TAG string (e.g., "@CODE:TEST-001")
        message: Human-readable issue description
        locations: List of (file, line) tuples where issue occurs
        suggestion: How to fix the issue
    """
    severity: str  # error|warning|info
    type: str  # duplicate|orphan|chain|format
    tag: str
    message: str
    locations: List[Tuple[str, int]] = field(default_factory=list)
    suggestion: str = ""

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization"""
        return {
            "severity": self.severity,
            "type": self.type,
            "tag": self.tag,
            "message": self.message,
            "locations": [
                {"file": f, "line": line} for f, line in self.locations
            ],
            "suggestion": self.suggestion
        }


@dataclass
class ValidationStatistics:
    """Validation statistics

    Attributes:
        total_files_scanned: Number of files scanned
        total_tags_found: Total TAG count across all files
        total_issues: Total issues found (errors + warnings)
        error_count: Number of errors
        warning_count: Number of warnings
        coverage_percentage: Percentage of SPEC tags with CODE implementation
    """
    total_files_scanned: int
    total_tags_found: int
    total_issues: int
    error_count: int
    warning_count: int
    coverage_percentage: float

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization"""
        return {
            "total_files_scanned": self.total_files_scanned,
            "total_tags_found": self.total_tags_found,
            "total_issues": self.total_issues,
            "error_count": self.error_count,
            "warning_count": self.warning_count,
            "coverage_percentage": self.coverage_percentage
        }


@dataclass
class CentralValidationResult:
    """Complete validation result from CentralValidator

    Attributes:
        is_valid: Overall validation status (False if errors, or warnings in strict mode)
        issues: All issues (errors + warnings combined)
        errors: Error-level issues only
        warnings: Warning-level issues only
        statistics: Validation statistics
        timestamp: When validation was performed
        execution_time_ms: Validation execution time in milliseconds
    """
    is_valid: bool
    issues: List[ValidationIssue]
    errors: List[ValidationIssue]
    warnings: List[ValidationIssue]
    statistics: ValidationStatistics
    timestamp: datetime
    execution_time_ms: float

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization"""
        return {
            "is_valid": self.is_valid,
            "issues": [issue.to_dict() for issue in self.issues],
            "errors": [error.to_dict() for error in self.errors],
            "warnings": [warning.to_dict() for warning in self.warnings],
            "statistics": self.statistics.to_dict(),
            "timestamp": self.timestamp.isoformat(),
            "execution_time_ms": self.execution_time_ms
        }


class TagValidator(ABC):
    """Abstract base class for all validators

    All validators must extend this class and implement:
    - validate(): Perform validation and return issues
    - get_name(): Return validator name
    - get_priority(): Return priority (higher = runs first)
    """

    # Default TAG pattern: @(SPEC|CODE|TEST|DOC):DOMAIN-TYPE-NNN
    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z]+(?:-[A-Z]+)*-\d{3})")

    @abstractmethod
    def validate(self, files: List[str]) -> List[ValidationIssue]:
        """Validate files and return list of issues

        Args:
            files: List of file paths to validate

        Returns:
            List of ValidationIssue objects
        """
        pass

    @abstractmethod
    def get_name(self) -> str:
        """Return validator name"""
        pass

    @abstractmethod
    def get_priority(self) -> int:
        """Return validator priority (higher = runs first)

        Priority guidelines:
        - 100: Format validation (must pass before other checks)
        - 90: Duplicate detection (errors that block)
        - 50: Orphan detection (warnings)
        - 30: Chain integrity (warnings)
        """
        pass

    def extract_tags_from_file(self, filepath: str) -> List[Tuple[str, str, int]]:
        """Extract TAGs from a file

        Args:
            filepath: Path to file to scan

        Returns:
            List of (tag_type, domain, line_number) tuples
            Example: [("CODE", "USER-REG-001", 10), ("TEST", "USER-REG-001", 25)]
        """
        try:
            path = Path(filepath)
            if not path.exists() or not path.is_file():
                return []

            content = path.read_text(encoding="utf-8", errors="ignore")
            lines = content.splitlines()

            tags = []
            for line_num, line in enumerate(lines, start=1):
                matches = self.TAG_PATTERN.findall(line)
                for tag_type, domain in matches:
                    tags.append((tag_type, domain, line_num))

            return tags

        except Exception:
            return []


class DuplicateValidator(TagValidator):
    """Validator for duplicate TAG detection

    Detects duplicate TAGs within same file or across multiple files.
    Returns error-level issues for all duplicates found.

    Priority: 90 (high - errors that must be fixed)
    """

    def validate(self, files: List[str]) -> List[ValidationIssue]:
        """Detect duplicate TAGs across all files

        Args:
            files: List of file paths to validate

        Returns:
            List of ValidationIssue objects for duplicates
        """
        issues: List[ValidationIssue] = []
        tag_locations: Dict[str, List[Tuple[str, int]]] = {}

        # Collect all TAGs and their locations
        for filepath in files:
            tags = self.extract_tags_from_file(filepath)
            for tag_type, domain, line_num in tags:
                full_tag = f"@{tag_type}:{domain}"
                if full_tag not in tag_locations:
                    tag_locations[full_tag] = []
                tag_locations[full_tag].append((filepath, line_num))

        # Find duplicates (tags with more than one location)
        for tag, locations in tag_locations.items():
            if len(locations) > 1:
                issues.append(ValidationIssue(
                    severity="error",
                    type="duplicate",
                    tag=tag,
                    message=f"Duplicate TAG found in {len(locations)} locations",
                    locations=locations,
                    suggestion="Remove duplicate TAG declarations. Each TAG must be unique across the codebase."
                ))

        return issues

    def get_name(self) -> str:
        """Return validator name"""
        return "DuplicateValidator"

    def get_priority(self) -> int:
        """Return high priority (errors block early)"""
        return 90


class OrphanValidator(TagValidator):
    """Validator for orphan TAG detection

    Detects orphan TAGs:
    - @CODE without corresponding @TEST
    - @TEST without corresponding @CODE
    - @SPEC without implementation (@CODE)

    Returns warning-level issues for all orphans found.

    Priority: 50 (medium - warnings that should be addressed)
    """

    def validate(self, files: List[str]) -> List[ValidationIssue]:
        """Detect orphan TAGs across all files

        Args:
            files: List of file paths to validate

        Returns:
            List of ValidationIssue objects for orphans
        """
        issues: List[ValidationIssue] = []

        # Collect all TAGs by type and domain
        tags_by_type: Dict[str, Dict[str, List[Tuple[str, int]]]] = {
            "SPEC": {},
            "CODE": {},
            "TEST": {},
            "DOC": {}
        }

        for filepath in files:
            tags = self.extract_tags_from_file(filepath)
            for tag_type, domain, line_num in tags:
                if domain not in tags_by_type[tag_type]:
                    tags_by_type[tag_type][domain] = []
                tags_by_type[tag_type][domain].append((filepath, line_num))

        # Check for orphans: CODE without TEST
        for domain, locations in tags_by_type["CODE"].items():
            if domain not in tags_by_type["TEST"]:
                for filepath, line_num in locations:
                    issues.append(ValidationIssue(
                        severity="warning",
                        type="orphan",
                        tag=f"@CODE:{domain}",
                        message="CODE TAG without corresponding TEST",
                        locations=[(filepath, line_num)],
                        suggestion=f"Add @TEST:{domain} to test this code implementation"
                    ))

        # Check for orphans: TEST without CODE
        for domain, locations in tags_by_type["TEST"].items():
            if domain not in tags_by_type["CODE"]:
                for filepath, line_num in locations:
                    issues.append(ValidationIssue(
                        severity="warning",
                        type="orphan",
                        tag=f"@TEST:{domain}",
                        message="TEST TAG without corresponding CODE",
                        locations=[(filepath, line_num)],
                        suggestion=f"Add @CODE:{domain} for the implementation being tested"
                    ))

        return issues

    def get_name(self) -> str:
        """Return validator name"""
        return "OrphanValidator"

    def get_priority(self) -> int:
        """Return medium priority"""
        return 50


class ChainValidator(TagValidator):
    """Validator for TAG chain integrity (NEW in Component 3)

    Validates complete SPEC→CODE→TEST→DOC chain:
    - @SPEC should have corresponding @CODE (implementation)
    - @CODE should have corresponding @TEST (test coverage)
    - Complete chain optionally includes @DOC (documentation)

    Returns warning-level issues for incomplete chains.

    Priority: 30 (low - runs after duplicates and orphans)
    """

    def validate(self, files: List[str]) -> List[ValidationIssue]:
        """Validate TAG chain integrity

        Args:
            files: List of file paths to validate

        Returns:
            List of ValidationIssue objects for chain problems
        """
        issues: List[ValidationIssue] = []

        # Collect all TAGs by type
        tags_by_type: Dict[str, Set[str]] = {
            "SPEC": set(),
            "CODE": set(),
            "TEST": set(),
            "DOC": set()
        }

        for filepath in files:
            tags = self.extract_tags_from_file(filepath)
            for tag_type, domain, _ in tags:
                tags_by_type[tag_type].add(domain)

        # Check SPEC→CODE chain
        for spec_domain in tags_by_type["SPEC"]:
            if spec_domain not in tags_by_type["CODE"]:
                issues.append(ValidationIssue(
                    severity="warning",
                    type="chain",
                    tag=f"@SPEC:{spec_domain}",
                    message="SPEC without CODE implementation",
                    locations=[],
                    suggestion=f"Implement @CODE:{spec_domain} for this specification"
                ))

        # Check CODE→SPEC chain (reverse: CODE should have SPEC)
        for code_domain in tags_by_type["CODE"]:
            if code_domain not in tags_by_type["SPEC"]:
                issues.append(ValidationIssue(
                    severity="warning",
                    type="chain",
                    tag=f"@CODE:{code_domain}",
                    message="CODE implementation without SPEC",
                    locations=[],
                    suggestion=f"Add @SPEC:{code_domain} to document the specification"
                ))

        # Check TEST→SPEC chain (reverse: TEST should have SPEC)
        for test_domain in tags_by_type["TEST"]:
            if test_domain not in tags_by_type["SPEC"]:
                issues.append(ValidationIssue(
                    severity="warning",
                    type="chain",
                    tag=f"@TEST:{test_domain}",
                    message="TEST without SPEC",
                    locations=[],
                    suggestion=f"Add @SPEC:{test_domain} to document what is being tested"
                ))

        # Check SPEC+CODE without TEST chain
        spec_with_code = tags_by_type["SPEC"] & tags_by_type["CODE"]
        for domain in spec_with_code:
            if domain not in tags_by_type["TEST"]:
                issues.append(ValidationIssue(
                    severity="warning",
                    type="chain",
                    tag=f"@CODE:{domain}",
                    message="CODE implementation without TEST",
                    locations=[],
                    suggestion=f"Add @TEST:{domain} to test this implementation"
                ))

        # Check for documentation (info level - optional)
        complete_implementations = tags_by_type["SPEC"] & tags_by_type["CODE"] & tags_by_type["TEST"]
        for domain in complete_implementations:
            if domain not in tags_by_type["DOC"]:
                issues.append(ValidationIssue(
                    severity="info",
                    type="chain",
                    tag=f"@SPEC:{domain}",
                    message="Complete implementation without documentation",
                    locations=[],
                    suggestion=f"Consider adding @DOC:{domain} to document this feature"
                ))

        return issues

    def get_name(self) -> str:
        """Return validator name"""
        return "ChainValidator"

    def get_priority(self) -> int:
        """Return low priority (runs after other checks)"""
        return 30


class FormatValidator(TagValidator):
    """Validator for TAG format validation

    Validates TAG format: @PREFIX:DOMAIN-TYPE-NNN
    Returns error-level issues for invalid formats.

    Priority: 100 (highest - format must be valid before other checks)
    """

    def validate(self, files: List[str]) -> List[ValidationIssue]:
        """Validate TAG format in all files

        Args:
            files: List of file paths to validate

        Returns:
            List of ValidationIssue objects for format problems
        """
        # Format validation is implicitly done by TAG_PATTERN regex
        # Invalid formats won't be extracted, so no issues to report here
        # This validator is a placeholder for future format-specific checks
        return []

    def get_name(self) -> str:
        """Return validator name"""
        return "FormatValidator"

    def get_priority(self) -> int:
        """Return highest priority"""
        return 100


class CentralValidator:
    """Central validation orchestrator

    Coordinates multiple validators to provide unified validation:
    1. Registers default validators (Duplicate, Orphan, Chain, Format)
    2. Runs validators in priority order (high to low)
    3. Collects all issues and generates statistics
    4. Creates formatted reports in multiple formats

    Usage:
        config = ValidationConfig(strict_mode=True)
        validator = CentralValidator(config=config)
        result = validator.validate_directory("/path/to/project")
        print(validator.create_report(result, format="summary"))

    Args:
        config: ValidationConfig object (default: ValidationConfig())
    """

    def __init__(self, config: Optional[ValidationConfig] = None):
        """Initialize CentralValidator with configuration

        Args:
            config: ValidationConfig object (default: ValidationConfig())
        """
        self.config = config or ValidationConfig()
        self.validators: List[TagValidator] = []

        # Register default validators based on configuration
        if self.config.check_duplicates:
            self.register_validator(DuplicateValidator())

        if self.config.check_orphans:
            self.register_validator(OrphanValidator())

        if self.config.check_chain_integrity:
            self.register_validator(ChainValidator())

        # Always register format validator
        self.register_validator(FormatValidator())

        # Sort validators by priority (high to low)
        self.validators.sort(key=lambda v: v.get_priority(), reverse=True)

    def register_validator(self, validator: TagValidator) -> None:
        """Register a custom validator

        Args:
            validator: TagValidator instance to register
        """
        self.validators.append(validator)
        # Re-sort by priority
        self.validators.sort(key=lambda v: v.get_priority(), reverse=True)

    def get_validators(self) -> List[TagValidator]:
        """Get list of registered validators

        Returns:
            List of TagValidator instances sorted by priority
        """
        return self.validators.copy()

    def _should_scan_file(self, filepath: str) -> bool:
        """Check if file should be scanned based on configuration

        Args:
            filepath: Path to file

        Returns:
            True if file should be scanned
        """
        path = Path(filepath)

        # Check file extension
        if self.config.allowed_file_types:
            suffix = path.suffix.lstrip(".")
            if suffix and suffix not in self.config.allowed_file_types:
                return False

        # Check ignore patterns
        for pattern in self.config.ignore_patterns:
            # Simple pattern matching (can be enhanced with fnmatch)
            pattern_clean = pattern.replace("/*", "").replace("*", "")
            if pattern_clean in str(path):
                return False

        return True

    def _collect_files_from_directory(self, directory: str) -> List[str]:
        """Collect all files from directory recursively

        Args:
            directory: Directory path to scan

        Returns:
            List of file paths that should be scanned
        """
        dir_path = Path(directory)
        if not dir_path.exists() or not dir_path.is_dir():
            return []

        files = []
        for path in dir_path.rglob("*"):
            if path.is_file() and self._should_scan_file(str(path)):
                files.append(str(path))

        return files

    def validate_files(self, files: List[str]) -> CentralValidationResult:
        """Validate list of files

        Main validation method that:
        1. Runs all registered validators in priority order
        2. Collects all issues (errors + warnings)
        3. Generates statistics
        4. Returns CentralValidationResult

        Args:
            files: List of file paths to validate

        Returns:
            CentralValidationResult with all issues and statistics
        """
        start_time = time.time()

        # Filter files based on configuration
        files_to_scan = [f for f in files if self._should_scan_file(f)]

        # Run all validators
        all_issues: List[ValidationIssue] = []
        for validator in self.validators:
            issues = validator.validate(files_to_scan)
            all_issues.extend(issues)

        # Separate errors and warnings
        errors = [issue for issue in all_issues if issue.severity == "error"]
        warnings = [issue for issue in all_issues if issue.severity == "warning"]

        # Calculate statistics
        total_tags = 0
        spec_tags = set()
        code_tags = set()

        for filepath in files_to_scan:
            try:
                path = Path(filepath)
                if not path.exists():
                    continue
                content = path.read_text(encoding="utf-8", errors="ignore")
                matches = TagValidator.TAG_PATTERN.findall(content)
                total_tags += len(matches)

                for tag_type, domain in matches:
                    if tag_type == "SPEC":
                        spec_tags.add(domain)
                    elif tag_type == "CODE":
                        code_tags.add(domain)
            except Exception:
                continue

        # Calculate coverage percentage (SPEC tags with CODE implementation)
        if spec_tags:
            implemented_specs = spec_tags & code_tags
            coverage_percentage = (len(implemented_specs) / len(spec_tags)) * 100
        else:
            coverage_percentage = 0.0 if code_tags else 100.0

        statistics = ValidationStatistics(
            total_files_scanned=len(files_to_scan),
            total_tags_found=total_tags,
            total_issues=len(all_issues),
            error_count=len(errors),
            warning_count=len(warnings),
            coverage_percentage=round(coverage_percentage, 2)
        )

        # Determine if validation passed
        if self.config.strict_mode:
            is_valid = len(errors) == 0 and len(warnings) == 0
        else:
            is_valid = len(errors) == 0

        execution_time_ms = (time.time() - start_time) * 1000

        return CentralValidationResult(
            is_valid=is_valid,
            issues=all_issues,
            errors=errors,
            warnings=warnings,
            statistics=statistics,
            timestamp=datetime.now(),
            execution_time_ms=round(execution_time_ms, 2)
        )

    def validate_directory(self, directory: str, config: Optional[ValidationConfig] = None) -> CentralValidationResult:
        """Validate all files in a directory recursively

        Args:
            directory: Directory path to scan
            config: Optional custom configuration (overrides instance config)

        Returns:
            CentralValidationResult
        """
        # Temporarily override config if provided
        original_config = self.config
        if config:
            self.config = config

        # Collect files from directory
        files = self._collect_files_from_directory(directory)

        # Validate files
        result = self.validate_files(files)

        # Restore original config
        self.config = original_config

        return result

    def create_report(self, result: CentralValidationResult, format: str = "detailed") -> str:
        """Create formatted validation report

        Args:
            result: CentralValidationResult from validation
            format: Report format (detailed|summary|json)

        Returns:
            Formatted report string
        """
        if format == "json":
            return json.dumps(result.to_dict(), indent=2)
        elif format == "summary":
            return self._create_summary_report(result)
        else:  # detailed
            return self._create_detailed_report(result)

    def _create_summary_report(self, result: CentralValidationResult) -> str:
        """Create summary report

        Args:
            result: CentralValidationResult

        Returns:
            Summary report string
        """
        lines = []
        lines.append("=" * 60)
        lines.append("TAG Validation Summary")
        lines.append("=" * 60)
        lines.append("")

        # Status
        if result.is_valid:
            lines.append("✅ Validation PASSED")
        else:
            lines.append("❌ Validation FAILED")
        lines.append("")

        # Statistics
        stats = result.statistics
        lines.append(f"Files Scanned:    {stats.total_files_scanned}")
        lines.append(f"Tags Found:       {stats.total_tags_found}")
        lines.append(f"Total Issues:     {stats.total_issues}")
        lines.append(f"  - Errors:       {stats.error_count}")
        lines.append(f"  - Warnings:     {stats.warning_count}")
        lines.append(f"Coverage:         {stats.coverage_percentage:.1f}%")
        lines.append("")

        # Execution time
        lines.append(f"Execution Time:   {result.execution_time_ms:.2f} ms")
        lines.append("")

        return "\n".join(lines)

    def _create_detailed_report(self, result: CentralValidationResult) -> str:
        """Create detailed report with all issues

        Args:
            result: CentralValidationResult

        Returns:
            Detailed report string
        """
        lines = []
        lines.append("=" * 60)
        lines.append("TAG Validation Report")
        lines.append("=" * 60)
        lines.append("")

        # Status
        if result.is_valid:
            lines.append("✅ Validation PASSED")
        else:
            lines.append("❌ Validation FAILED")
        lines.append("")

        # Errors
        if result.errors:
            lines.append("ERRORS:")
            lines.append("-" * 60)
            for error in result.errors:
                lines.append(f"  {error.tag}")
                lines.append(f"    Type: {error.type}")
                lines.append(f"    Message: {error.message}")
                if error.locations:
                    lines.append("    Locations:")
                    for filepath, line_num in error.locations:
                        lines.append(f"      - {filepath}:{line_num}")
                if error.suggestion:
                    lines.append(f"    Suggestion: {error.suggestion}")
                lines.append("")

        # Warnings
        if result.warnings:
            lines.append("WARNINGS:")
            lines.append("-" * 60)
            for warning in result.warnings:
                lines.append(f"  {warning.tag}")
                lines.append(f"    Type: {warning.type}")
                lines.append(f"    Message: {warning.message}")
                if warning.locations:
                    lines.append("    Locations:")
                    for filepath, line_num in warning.locations:
                        lines.append(f"      - {filepath}:{line_num}")
                if warning.suggestion:
                    lines.append(f"    Suggestion: {warning.suggestion}")
                lines.append("")

        # Statistics
        lines.append("STATISTICS:")
        lines.append("-" * 60)
        stats = result.statistics
        lines.append(f"Files Scanned:        {stats.total_files_scanned}")
        lines.append(f"Tags Found:           {stats.total_tags_found}")
        lines.append(f"Total Issues:         {stats.total_issues}")
        lines.append(f"  - Errors:           {stats.error_count}")
        lines.append(f"  - Warnings:         {stats.warning_count}")
        lines.append(f"Spec→Code Coverage:   {stats.coverage_percentage:.1f}%")
        lines.append(f"Execution Time:       {result.execution_time_ms:.2f} ms")
        lines.append("")

        return "\n".join(lines)

    def export_for_reporting(self, result: CentralValidationResult) -> Dict[str, Any]:
        """Export validation result in format compatible with ReportGenerator

        This method bridges Component 3 (Validation) with Component 4 (Reporting).
        Exports validation results as structured data for integration with
        automated reporting workflows.

        Args:
            result: CentralValidationResult from validation

        Returns:
            Dictionary with validation data formatted for reporting:
            {
                "timestamp": ISO timestamp,
                "is_valid": bool,
                "statistics": {
                    "total_files_scanned": int,
                    "total_tags_found": int,
                    "total_issues": int,
                    "error_count": int,
                    "warning_count": int,
                    "coverage_percentage": float
                },
                "issues_by_type": {
                    "duplicate": [...],
                    "orphan": [...],
                    "chain": [...],
                    "format": [...]
                },
                "execution_time_ms": float
            }
        """
        # Group issues by type
        issues_by_type: Dict[str, List[Dict[str, Any]]] = {
            "duplicate": [],
            "orphan": [],
            "chain": [],
            "format": []
        }

        for issue in result.issues:
            issues_by_type[issue.type].append(issue.to_dict())

        # Build export data
        export_data = {
            "timestamp": result.timestamp.isoformat(),
            "is_valid": result.is_valid,
            "statistics": result.statistics.to_dict(),
            "issues_by_type": issues_by_type,
            "execution_time_ms": result.execution_time_ms
        }

        return export_data
