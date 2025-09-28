#!/usr/bin/env python3
"""
Tag parsing and data structures for validate_tags module
"""

from dataclasses import dataclass, field
from typing import List


@dataclass
class TagReference:
    """Tag reference information"""
    tag_type: str
    tag_id: str
    file_path: str
    line_number: int
    context: str


@dataclass
class TagHealthReport:
    """Tag health report data structure"""
    total_tags: int = 0
    valid_tags: int = 0
    invalid_tags: int = 0
    orphan_tags: int = 0
    broken_links: int = 0
    quality_score: float = 0.0
    issues: List[str] = field(default_factory=list)
    recommendations: List[str] = field(default_factory=list)