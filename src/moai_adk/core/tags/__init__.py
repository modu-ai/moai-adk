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
from .pre_commit_validator import (
    PreCommitValidator,
    ValidationResult,
    ValidationError,
    ValidationWarning,
)

# Component 2: CI/CD validator
from .ci_validator import CIValidator

# Component 3: Central validation system
from .validator import (
    ValidationConfig,
    TagValidator,
    DuplicateValidator,
    OrphanValidator,
    ChainValidator,
    FormatValidator,
    CentralValidator,
    CentralValidationResult,
    ValidationIssue,
    ValidationStatistics,
)

# Component 4: Documentation & Reporting
from .reporter import (
    TagInventory,
    TagMatrix,
    InventoryGenerator,
    MatrixGenerator,
    CoverageAnalyzer,
    StatisticsGenerator,
    ReportFormatter,
    ReportGenerator,
    CoverageMetrics,
    StatisticsReport,
    ReportResult,
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
