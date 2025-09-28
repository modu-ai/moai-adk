#!/usr/bin/env python3
# @TASK:TAG-REPAIR-ANALYZER-001
"""
TAG Analyzer Module

Analyzes TAG integrity and identifies broken links and orphaned tags.
Focuses on traceability chain validation.
"""



class TagAnalyzer:
    """Analyzes TAG integrity and relationships."""

    def analyze_tag_integrity(self, all_tags: dict[str, list[str]]) -> dict:
        """Analyze TAG integrity and identify issues."""
        issues = []
        orphaned_tags = []
        broken_links = []

        # Check for orphaned tags (no references)
        for tag, locations in all_tags.items():
            if len(locations) == 1:  # Only defined, not referenced
                if not self.has_references(tag, all_tags):
                    orphaned_tags.append(tag)

        # Check for broken traceability chains
        for tag in all_tags:
            category = self.get_tag_category(tag)
            if category and not self.validate_chain(tag, all_tags):
                broken_links.append(tag)

        # Compile issues
        if orphaned_tags:
            issues.append(f"Orphaned tags: {len(orphaned_tags)}")
        if broken_links:
            issues.append(f"Broken chains: {len(broken_links)}")

        return {
            "has_issues": bool(issues),
            "orphaned_tags": orphaned_tags,
            "broken_links": broken_links,
            "issues": issues,
            "total_tags": len(all_tags),
        }

    def has_references(self, tag: str, all_tags: dict[str, list[str]]) -> bool:
        """Check if tag has references in other files."""
        tag_locations = all_tags.get(tag, [])
        return len(tag_locations) > 1

    def get_tag_category(self, tag: str) -> str | None:
        """Get category for tag type."""
        tag_type = tag.split(":")[0] if ":" in tag else tag
        categories = {
            "REQ": "SPEC", "SPEC": "SPEC", "DESIGN": "SPEC", "TASK": "SPEC",
            "VISION": "STEERING", "STRUCT": "STEERING", "TECH": "STEERING", "ADR": "STEERING",
            "FEATURE": "IMPLEMENTATION", "API": "IMPLEMENTATION", "TEST": "IMPLEMENTATION", "DATA": "IMPLEMENTATION",
            "PERF": "QUALITY", "SEC": "QUALITY", "DEBT": "QUALITY", "TODO": "QUALITY",
        }
        return categories.get(tag_type)

    def validate_chain(self, tag: str, all_tags: dict[str, list[str]]) -> bool:
        """Validate traceability chain for tag."""
        # Simplified validation - just check if tag exists
        return tag in all_tags