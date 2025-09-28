#!/usr/bin/env python3
# @TASK:TAG-REPAIR-GENERATOR-001
"""
Repair Generator Module

Generates repair suggestions and creates missing TAG references.
Focuses on intelligent repair strategy generation.
"""

from datetime import datetime


class RepairGenerator:
    """Generates repair suggestions for broken TAG chains."""

    def generate_repair_preview(
        self, analysis: dict, all_tags: dict, tag_categories: dict
    ) -> dict:
        """Generate repair suggestions preview."""
        repairs = []

        # Generate repairs for orphaned tags
        for tag in analysis.get("orphaned_tags", []):
            repair = self.create_orphan_repair(tag, tag_categories)
            if repair:
                repairs.append(repair)

        # Generate repairs for broken links
        for tag in analysis.get("broken_links", []):
            repair = self.create_link_repair(tag, all_tags)
            if repair:
                repairs.append(repair)

        return {
            "repairs": repairs,
            "total_repairs": len(repairs),
            "timestamp": datetime.now().isoformat(),
        }

    def create_orphan_repair(self, tag: str, tag_categories: dict) -> dict | None:
        """Create repair suggestion for orphaned tag."""
        tag_type = tag.split(":")[0] if ":" in tag else tag

        # Find appropriate category
        category = None
        for cat, types in tag_categories.items():
            if tag_type in types:
                category = cat
                break

        if not category:
            return None

        return {
            "type": "orphan_repair",
            "tag": tag,
            "category": category,
            "action": f"Create reference for {tag}",
            "priority": "medium",
        }

    def create_link_repair(self, tag: str, all_tags: dict) -> dict | None:
        """Create repair suggestion for broken link."""
        return {
            "type": "link_repair",
            "tag": tag,
            "action": f"Fix broken chain for {tag}",
            "priority": "high",
        }

    def extract_requirements_from_tag(self, tag: str) -> dict:
        """Extract requirements info from TAG."""
        return {
            "tag": tag,
            "type": tag.split(":")[0] if ":" in tag else tag,
            "id": tag.split(":")[-1] if ":" in tag else "001",
        }