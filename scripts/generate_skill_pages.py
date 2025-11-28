#!/usr/bin/env python3
"""
Generate Nextra MDX pages for all MoAI-ADK skills from SKILL.md files.

This script:
1. Reads all SKILL.md files from .claude/skills/
2. Converts them to Nextra-compatible MDX format
3. Adds necessary frontmatter and navigation
4. Generates pages/skills/*.mdx files
"""

import os
import re
from pathlib import Path
from typing import Dict, List


SKILLS_DIR = Path("/Users/goos/worktrees/MoAI-ADK/SPEC-NEXTRA-001/.claude/skills")
OUTPUT_DIR = Path("/Users/goos/worktrees/MoAI-ADK/SPEC-NEXTRA-001/docs/pages/skills")

def read_skill_file(skill_path: Path) -> Dict[str, str]:
    """Read SKILL.md and extract metadata and content."""
    skill_md = skill_path / "SKILL.md"

    if not skill_md.exists():
        return None

    with open(skill_md, 'r', encoding='utf-8') as f:
        content = f.read()

    # Extract skill name from directory
    skill_name = skill_path.name

    # Extract first heading as title
    title_match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
    title = title_match.group(1) if title_match else skill_name

    # Extract description (first paragraph after title)
    desc_match = re.search(r'^#\s+.+?\n\n(.+?)(?:\n\n|\n#)', content, re.MULTILINE | re.DOTALL)
    description = desc_match.group(1).strip() if desc_match else ""

    return {
        'name': skill_name,
        'title': title,
        'description': description,
        'content': content
    }


def convert_to_mdx(skill_data: Dict[str, str]) -> str:
    """Convert SKILL.md content to Nextra MDX format."""

    if not skill_data:
        return ""

    content = skill_data['content']

    # Add frontmatter
    mdx = f"""---
title: {skill_data['title']}
description: {skill_data['description'][:200] if skill_data['description'] else 'MoAI-ADK Skill'}
---

{content}

---

## Related Resources

### Agents
Check the [Agent Guide](/advanced/agents-guide) to see which agents use this skill.

### Commands
See [Core Commands](/reference/commands) for commands that leverage this skill.

### Patterns
Explore [Composition Patterns](/advanced/patterns) for real-world usage examples.

---

**Last Updated**: 2025-11-28
**Skill Type**: {skill_data['name'].split('-')[1].title()}
**Version**: 2.0
"""

    return mdx


def generate_all_skills():
    """Generate MDX pages for all skills."""

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    generated_count = 0
    skipped_count = 0

    for skill_dir in sorted(SKILLS_DIR.iterdir()):
        if not skill_dir.is_dir():
            continue

        if not skill_dir.name.startswith('moai-'):
            continue

        print(f"Processing {skill_dir.name}...")

        skill_data = read_skill_file(skill_dir)

        if not skill_data:
            print(f"  ⚠️  No SKILL.md found in {skill_dir.name}")
            skipped_count += 1
            continue

        mdx_content = convert_to_mdx(skill_data)

        output_file = OUTPUT_DIR / f"{skill_dir.name}.mdx"

        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(mdx_content)

        print(f"  ✅ Generated {output_file.name} ({len(mdx_content)} chars)")
        generated_count += 1

    print(f"\n{'='*60}")
    print(f"Generation complete!")
    print(f"  ✅ Generated: {generated_count} skills")
    print(f"  ⚠️  Skipped: {skipped_count} skills")
    print(f"{'='*60}")


if __name__ == "__main__":
    generate_all_skills()
