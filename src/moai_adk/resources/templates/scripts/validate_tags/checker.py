#!/usr/bin/env python3
"""
Tag checking logic for orphan tags and broken links
"""

import re
from typing import Dict, List, Set, Tuple
from .parser import TagReference


def find_orphan_tags(tag_index: Dict[str, List[TagReference]]) -> List[str]:
    """Find orphan tags (unreferenced tags)"""
    orphan_tags = []

    # Extract references to other tags
    referenced_tags: Set[str] = set()

    for tag_key, tag_refs in tag_index.items():
        for tag_ref in tag_refs:
            # Find other tag references in context
            context_tags = re.findall(
                r"@([A-Z]+)[-:]([A-Z0-9-_]+)", tag_ref.context
            )
            for ref_type, ref_id in context_tags:
                if f"{ref_type}:{ref_id}" != tag_key:  # Exclude self
                    referenced_tags.add(f"{ref_type}:{ref_id}")

    # Find unreferenced tags (excluding root tags)
    root_tag_types = ["REQ", "SPEC", "VISION"]

    for tag_key in tag_index:
        tag_type = tag_key.split(":")[0]
        if tag_key not in referenced_tags and tag_type not in root_tag_types:
            orphan_tags.append(tag_key)

    return orphan_tags


def find_broken_links(tag_index: Dict[str, List[TagReference]]) -> List[Tuple[str, str]]:
    """Find broken links (references to non-existent tags)"""
    broken_links = []

    for tag_key, tag_refs in tag_index.items():
        for tag_ref in tag_refs:
            # Find other tag references in context
            context_tags = re.findall(
                r"@([A-Z]+)[-:]([A-Z0-9-_]+)", tag_ref.context
            )

            for ref_type, ref_id in context_tags:
                referenced_tag = f"{ref_type}:{ref_id}"

                # Not self and not in index
                if referenced_tag != tag_key and referenced_tag not in tag_index:
                    broken_links.append((tag_key, referenced_tag))

    return broken_links