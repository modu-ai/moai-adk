#!/usr/bin/env python3
"""
Report generation logic for validate_tags module
"""

from datetime import datetime
from typing import Any, Dict, List, Tuple
from .parser import TagHealthReport, TagReference
from .scorer import calculate_quality_score


def generate_health_report(
    all_tags: List[TagReference],
    tag_index: Dict[str, List[TagReference]],
    format_violations: List[str],
    orphan_tags: List[str],
    broken_links: List[Tuple[str, str]],
    chain_violations: List[str]
) -> TagHealthReport:
    """Generate comprehensive health report"""

    total_tags = len(all_tags)
    invalid_tags = len(format_violations)
    valid_tags = total_tags - invalid_tags

    quality_score = calculate_quality_score(
        total_tags, valid_tags, len(orphan_tags), len(broken_links)
    )

    issues = []
    recommendations = []

    # Add format violations
    issues.extend(format_violations)

    # Add orphan tag issues
    if orphan_tags:
        issues.append(f"Found {len(orphan_tags)} orphan tags")
        recommendations.append("Link orphan tags to parent requirements")

    # Add broken link issues
    if broken_links:
        issues.append(f"Found {len(broken_links)} broken links")
        recommendations.append("Fix broken tag references")

    # Add chain violations
    issues.extend(chain_violations)

    return TagHealthReport(
        total_tags=total_tags,
        valid_tags=valid_tags,
        invalid_tags=invalid_tags,
        orphan_tags=len(orphan_tags),
        broken_links=len(broken_links),
        quality_score=quality_score,
        issues=issues,
        recommendations=recommendations
    )


# create_report_data moved to exporter.py