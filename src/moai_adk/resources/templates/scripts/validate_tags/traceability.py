#!/usr/bin/env python3
"""
Traceability chain validation logic
"""

from typing import Dict, List
from .parser import TagReference


def validate_traceability_chains(
    tag_index: Dict[str, List[TagReference]],
    traceability_chains: Dict[str, List[str]]
) -> List[str]:
    """Validate traceability chains"""
    chain_violations = []

    for chain_name, chain_types in traceability_chains.items():
        # Find tags in this chain
        chain_tags = {}
        for tag_type in chain_types:
            chain_tags[tag_type] = [
                key for key in tag_index.keys()
                if key.startswith(f"{tag_type}:")
            ]

        # Check chain completeness for each starting tag
        for start_type in chain_types[:1]:  # Only check from first type
            for start_tag in chain_tags.get(start_type, []):
                missing_links = []
                for target_type in chain_types[1:]:
                    # Check if there's a reference path
                    has_link = any(
                        target_tag in str(tag_refs)
                        for target_tag in chain_tags.get(target_type, [])
                        for tag_refs in tag_index.get(start_tag, [])
                    )
                    if not has_link and chain_tags.get(target_type):
                        missing_links.append(target_type)

                if missing_links:
                    chain_violations.append(
                        f"Chain {chain_name}: {start_tag} missing links to {missing_links}"
                    )

    return chain_violations