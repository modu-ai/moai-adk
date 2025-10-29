#!/usr/bin/env python3
# @CODE:DOC-TAG-004 | TAG validation core module (Components 1, 2, 3 & 4)
"""TAG validation and management for MoAI-ADK

This module provides TAG validation functionality for:
- Pre-commit hook validation (Component 1)
- CI/CD pipeline validation (Component 2)
- Central validation system (Component 3)
- Documentation & Reporting (Component 4)
- TAG format checking
- Duplicate detection
- Orphan detection
- Chain integrity validation
"""

# Component 1: Pre-commit validator
# Component 2: CI/CD validator
from .ci_validator import CIValidator
from .pre_commit_validator import (
    PreCommitValidator,
    ValidationError,
    ValidationResult,
    ValidationWarning,
)

# Component 4: Documentation & Reporting
from .reporter import (
    CoverageAnalyzer,
    CoverageMetrics,
    InventoryGenerator,
    MatrixGenerator,
    ReportFormatter,
    ReportGenerator,
    ReportResult,
    StatisticsGenerator,
    StatisticsReport,
    TagInventory,
    TagMatrix,
)

# Component 3: Central validation system
from .validator import (
    CentralValidationResult,
    CentralValidator,
    ChainValidator,
    DuplicateValidator,
    FormatValidator,
    OrphanValidator,
    TagValidator,
    ValidationConfig,
    ValidationIssue,
    ValidationStatistics,
)

__all__ = [
    # Component 1
    "PreCommitValidator",
    "ValidationResult",
    "ValidationError",
    "ValidationWarning",
    # Component 2
    "CIValidator",
    # Component 3
    "ValidationConfig",
    "TagValidator",
    "DuplicateValidator",
    "OrphanValidator",
    "ChainValidator",
    "FormatValidator",
    "CentralValidator",
    "CentralValidationResult",
    "ValidationIssue",
    "ValidationStatistics",
    # Component 4
    "TagInventory",
    "TagMatrix",
    "InventoryGenerator",
    "MatrixGenerator",
    "CoverageAnalyzer",
    "StatisticsGenerator",
    "ReportFormatter",
    "ReportGenerator",
    "CoverageMetrics",
    "StatisticsReport",
    "ReportResult",
]
