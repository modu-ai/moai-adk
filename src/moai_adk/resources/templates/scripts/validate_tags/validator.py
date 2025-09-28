#!/usr/bin/env python3
"""
Tag validation logic for validate_tags module
"""

import re
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from .parser import TagReference


class TagValidator:
    """16-Core TAG system validator"""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"
        self.indexes_dir = self.moai_dir / "indexes"

        # 16-Core tag schema
        self.tag_categories = {
            "Primary": ["REQ", "SPEC", "DESIGN", "TASK", "TEST"],
            "Steering": ["VISION", "STRUCT", "TECH", "ADR"],
            "Implementation": ["FEATURE", "API", "DATA"],
            "Quality": ["PERF", "SEC", "DEBT", "TODO"],
            "Legacy": ["US", "FR", "NFR", "BUG", "REVIEW"],
        }

        self.valid_tag_types = []
        for category_tags in self.tag_categories.values():
            self.valid_tag_types.extend(category_tags)

        # Traceability chains
        self.traceability_chains = {
            "Primary": ["REQ", "DESIGN", "TASK", "TEST"],
            "Development": ["SPEC", "ADR", "TASK", "API", "TEST"],
            "Quality": ["PERF", "SEC", "DEBT", "REVIEW"],
        }

    def validate_tag_format(self, tag: TagReference) -> Tuple[bool, Optional[str]]:
        """Validate tag format"""
        # Valid tag type check
        if tag.tag_type not in self.valid_tag_types:
            return False, f"Invalid tag type '{tag.tag_type}'"

        # Tag ID format check
        if not re.match(r"^[A-Z0-9-_]+$", tag.tag_id):
            return False, f"Invalid tag ID format '{tag.tag_id}'"

        # Length limits
        if len(tag.tag_id) < 2:
            return False, f"Tag ID '{tag.tag_id}' too short"

        if len(tag.tag_id) > 50:
            return False, f"Tag ID '{tag.tag_id}' too long"

        return True, None

    def build_tag_index(self, tags: List[TagReference]) -> Dict[str, List[TagReference]]:
        """Build tag index"""
        index = {}
        for tag in tags:
            tag_key = f"{tag.tag_type}:{tag.tag_id}"
            if tag_key not in index:
                index[tag_key] = []
            index[tag_key].append(tag)
        return index