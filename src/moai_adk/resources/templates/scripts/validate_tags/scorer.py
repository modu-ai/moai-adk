#!/usr/bin/env python3
"""
Quality scoring logic for validate_tags module
"""


def calculate_quality_score(
    total_tags: int,
    valid_tags: int,
    orphan_count: int,
    broken_count: int
) -> float:
    """Calculate overall quality score"""
    if total_tags == 0:
        return 1.0

    # Base score from valid tags
    validity_score = valid_tags / total_tags

    # Penalty for orphans and broken links
    orphan_penalty = min(orphan_count * 0.1, 0.3)
    broken_penalty = min(broken_count * 0.2, 0.4)

    final_score = validity_score - orphan_penalty - broken_penalty
    return max(0.0, min(1.0, final_score))