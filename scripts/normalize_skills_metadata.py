#!/usr/bin/env python3
"""
GREEN PHASE: Metadata Normalization Implementation
Adds missing metadata fields to all skills
"""

import os
import re
import yaml
from pathlib import Path
from datetime import datetime

PROJECT_ROOT = Path("/Users/goos/MoAI/MoAI-ADK")
SKILLS_DIR = PROJECT_ROOT / ".claude" / "skills"
TODAY = datetime.now().strftime("%Y-%m-%d")


def load_skill_md(skill_path: Path) -> tuple[dict, str]:
    """Load YAML frontmatter and full content."""
    skill_md = skill_path / "SKILL.md"
    if not skill_md.exists():
        return {}, ""

    content = skill_md.read_text(encoding="utf-8")

    if not content.startswith("---"):
        return {}, content

    try:
        end = content.find("---", 3)
        if end == -1:
            return {}, content

        frontmatter_str = content[3:end].strip()
        metadata = yaml.safe_load(frontmatter_str) or {}
        body = content[end + 3 :].strip()
        return metadata, body
    except Exception as e:
        print(f"Error parsing {skill_path.name}: {e}")
        return {}, content


def calculate_compliance_score(metadata: dict, description: str) -> int:
    """Calculate compliance score based on present fields."""
    required_fields = {
        "name",
        "description",
        "version",
        "modularized",
        "last_updated",
        "compliance_score",
    }
    optional_fields = {
        "allowed-tools",
        "dependencies",
        "category_tier",
        "auto_trigger_keywords",
        "agent_coverage",
    }

    present_required = sum(
        1 for field in required_fields if field in metadata and metadata[field]
    )
    present_optional = sum(
        1 for field in optional_fields if field in metadata and metadata[field]
    )

    # Check description quality
    desc_len = len(description or "")
    desc_score = 25 if 100 <= desc_len <= 200 else 15

    total_score = (
        (present_required / len(required_fields)) * 50  # 50% for required fields
        + (present_optional / len(optional_fields)) * 25  # 25% for optional fields
        + desc_score  # 25% for description
    )

    return int(total_score)


def normalize_skill_metadata(skill_path: Path) -> bool:
    """Add missing metadata to a skill."""
    metadata, body = load_skill_md(skill_path)
    if not metadata:
        return False

    modified = False
    description = metadata.get("description", "")

    # Ensure version field (required)
    if "version" not in metadata:
        metadata["version"] = "1.0.0"
        modified = True

    # Ensure modularized field (required)
    if "modularized" not in metadata:
        # Check if skill has modules directory
        has_modules = (skill_path / "modules").exists()
        metadata["modularized"] = has_modules
        modified = True

    # Ensure last_updated field (required)
    if "last_updated" not in metadata:
        metadata["last_updated"] = TODAY
        modified = True

    # Fix description length if needed
    if description:
        if len(description) < 100:
            # Try to expand from body if available
            pass  # Will be handled by developer
        elif len(description) > 200:
            # Truncate gracefully
            if len(description) > 300:
                metadata["description"] = description[:200].rsplit(" ", 1)[0] + "..."
                modified = True

    # Calculate and add compliance score (required)
    compliance_score = calculate_compliance_score(metadata, description)
    if (
        "compliance_score" not in metadata
        or metadata.get("compliance_score") != compliance_score
    ):
        metadata["compliance_score"] = compliance_score
        modified = True

    # Add auto_trigger_keywords if missing (recommended)
    if "auto_trigger_keywords" not in metadata:
        keywords = extract_keywords_from_name(skill_path.name, description)
        if keywords:
            metadata["auto_trigger_keywords"] = keywords
            modified = True

    if modified:
        # Reconstruct SKILL.md
        frontmatter_lines = ["---"]
        for key in [
            "name",
            "description",
            "version",
            "modularized",
            "allowed-tools",
            "last_updated",
            "compliance_score",
            "auto_trigger_keywords",
            "category_tier",
            "agent_coverage",
        ]:
            if key in metadata and metadata[key] is not None:
                if key == "compliance_score":
                    frontmatter_lines.append(f"{key}: {metadata[key]}")
                elif isinstance(metadata[key], list):
                    frontmatter_lines.append(f"{key}:")
                    for item in metadata[key]:
                        frontmatter_lines.append(f"  - {item}")
                elif isinstance(metadata[key], bool):
                    frontmatter_lines.append(f"{key}: {str(metadata[key]).lower()}")
                else:
                    frontmatter_lines.append(f"{key}: {metadata[key]}")

        frontmatter_lines.append("---")
        new_content = "\n".join(frontmatter_lines) + "\n\n" + body

        skill_md_path = skill_path / "SKILL.md"
        skill_md_path.write_text(new_content, encoding="utf-8")
        return True

    return False


def extract_keywords_from_name(skill_name: str, description: str) -> list:
    """Extract auto-trigger keywords from skill name and description."""
    keywords = []

    # Extract from name
    parts = skill_name.replace("moai-", "").split("-")
    keywords.extend(parts)

    # Add common keywords from description
    if description:
        desc_lower = description.lower()
        common_keywords = {
            "authentication": ["auth", "login", "token", "oauth", "jwt"],
            "api": ["rest", "graphql", "endpoint", "request"],
            "database": ["db", "sql", "query", "table"],
            "performance": ["speed", "optimize", "latency", "benchmark"],
            "testing": ["test", "unit", "integration", "e2e"],
            "security": ["encrypt", "secure", "vulnerability", "compliance"],
            "python": ["python", "django", "fastapi", "async"],
            "typescript": ["typescript", "react", "node", "frontend"],
        }

        for keyword, variants in common_keywords.items():
            if any(v in desc_lower for v in variants):
                if keyword not in keywords:
                    keywords.append(keyword)

    # Deduplicate and limit
    keywords = list(set(keywords))[:10]
    return sorted(keywords) if keywords else []


def main():
    """Main function."""
    if not SKILLS_DIR.exists():
        print(f"Error: Skills directory not found: {SKILLS_DIR}")
        return

    modified_count = 0
    total_count = 0

    for skill_dir in sorted(SKILLS_DIR.iterdir()):
        if not skill_dir.is_dir() or skill_dir.name.startswith("."):
            continue

        total_count += 1
        if normalize_skill_metadata(skill_dir):
            modified_count += 1
            print(f"✓ Normalized: {skill_dir.name}")

    print(f"\nMetadata normalization complete!")
    print(f"Modified: {modified_count}/{total_count} skills")


if __name__ == "__main__":
    main()
