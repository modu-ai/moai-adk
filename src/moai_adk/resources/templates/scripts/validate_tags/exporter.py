#!/usr/bin/env python3
"""
Report export logic for validate_tags module
"""

from datetime import datetime
from typing import Any, Dict
from .parser import TagHealthReport


def create_report_data(report: TagHealthReport) -> Dict[str, Any]:
    """Create structured report data for export"""
    return {
        "timestamp": datetime.now().isoformat(),
        "summary": {
            "total_tags": report.total_tags,
            "valid_tags": report.valid_tags,
            "invalid_tags": report.invalid_tags,
            "orphan_tags": report.orphan_tags,
            "broken_links": report.broken_links,
            "quality_score": round(report.quality_score, 2)
        },
        "issues": {
            "format_violations": [issue for issue in report.issues if "Invalid" in issue],
            "orphan_tags": [issue for issue in report.issues if "orphan" in issue.lower()],
            "broken_links": [issue for issue in report.issues if "broken" in issue.lower()],
            "chain_violations": [issue for issue in report.issues if "Chain" in issue]
        },
        "recommendations": report.recommendations
    }