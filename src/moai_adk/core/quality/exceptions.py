"""
Exceptions for guideline checking system.

@FEATURE:QUALITY-EXCEPTIONS Exception definitions for guideline validation
@DESIGN:SEPARATED-EXCEPTIONS-001 Extracted from oversized guideline_checker.py (924 LOC)
"""


class GuidelineError(Exception):
    """Guideline compliance violation exception."""
