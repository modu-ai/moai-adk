#!/usr/bin/env python3
"""
Migrate MoAI-ADK Skills to Anthropic Official Standard A (Minimal Implementation)

Removes non-standard fields (tier, auto-load, version, author, license, tags, metadata)
Keeps only: name, description, allowed-tools
"""

import re
import yaml
from pathlib import Path
from typing import Dict, Any, List
import shutil


def parse_skill_file(file_path: Path) -> tuple[Dict[str, Any], str]:
    """Parse SKILL.md file and extract frontmatter + content."""
    content = file_path.read_text(encoding='utf-8')

    # Match YAML frontmatter
    match = re.match(r'^---\n(.*?)\n---\n(.*)$', content, re.DOTALL)
    if not match:
        raise ValueError(f"No YAML frontmatter found in {file_path}")

    frontmatter_str = match.group(1)
    markdown_content = match.group(2)

    # Try to parse YAML
    try:
        frontmatter = yaml.safe_load(frontmatter_str)
    except yaml.YAMLError:
        # Handle malformed YAML (e.g., allowed-tools interrupted by other fields)
        # Reconstruct manually
        frontmatter = parse_malformed_yaml(frontmatter_str)

    return frontmatter, markdown_content


def parse_malformed_yaml(yaml_str: str) -> Dict[str, Any]:
    """Parse malformed YAML frontmatter (fallback for broken allowed-tools)."""
    result = {}
    lines = yaml_str.strip().split('\n')

    i = 0
    while i < len(lines):
        line = lines[i].strip()

        if not line or line.startswith('#'):
            i += 1
            continue

        if ':' in line:
            key, value = line.split(':', 1)
            key = key.strip()
            value = value.strip()

            # Check if next lines are list items (for allowed-tools)
            if not value and i + 1 < len(lines) and lines[i + 1].strip().startswith('-'):
                # Collect list items
                list_items = []
                i += 1
                while i < len(lines) and lines[i].strip().startswith('-'):
                    item = lines[i].strip()[1:].strip()  # Remove '-' and whitespace
                    if item:
                        list_items.append(item)
                    i += 1
                result[key] = list_items
                continue
            else:
                # Simple key-value
                if value.startswith('"') and value.endswith('"'):
                    value = value[1:-1]  # Remove quotes
                result[key] = value

        i += 1

    return result


def convert_to_standard_a(frontmatter: Dict[str, Any]) -> Dict[str, Any]:
    """Convert frontmatter to Standard A (minimal implementation)."""
    standard_a = {}

    # Required fields
    if 'name' in frontmatter:
        standard_a['name'] = frontmatter['name']
    if 'description' in frontmatter:
        standard_a['description'] = frontmatter['description']

    # Optional: allowed-tools (security)
    if 'allowed-tools' in frontmatter:
        standard_a['allowed-tools'] = frontmatter['allowed-tools']

    # Remove all other fields:
    # - tier (non-standard)
    # - auto-load (non-standard)
    # - version, author, license, tags, metadata (keep minimal)

    return standard_a


def write_skill_file(file_path: Path, frontmatter: Dict[str, Any], content: str):
    """Write SKILL.md file with new frontmatter."""
    # Convert frontmatter to YAML string
    yaml_str = yaml.dump(frontmatter, allow_unicode=True, default_flow_style=False, sort_keys=False)

    # Combine frontmatter + content
    new_content = f"---\n{yaml_str}---\n{content}"

    file_path.write_text(new_content, encoding='utf-8')


def migrate_skill(file_path: Path, backup: bool = True) -> bool:
    """Migrate a single SKILL.md file to Standard A."""
    try:
        # Parse
        frontmatter, content = parse_skill_file(file_path)

        # Backup
        if backup:
            backup_path = file_path.with_suffix('.md.bak')
            shutil.copy2(file_path, backup_path)
            print(f"  ✅ Backup: {backup_path.name}")

        # Convert
        new_frontmatter = convert_to_standard_a(frontmatter)

        # Write
        write_skill_file(file_path, new_frontmatter, content)

        # Report changes
        removed_fields = set(frontmatter.keys()) - set(new_frontmatter.keys())
        if removed_fields:
            print(f"  🔧 Removed: {', '.join(sorted(removed_fields))}")

        return True
    except Exception as e:
        print(f"  ❌ Error: {e}")
        return False


def main():
    """Main migration function."""
    root = Path(__file__).parent.parent

    # Two directories to migrate
    directories = [
        root / ".claude" / "skills",
        root / "src" / "moai_adk" / "templates" / ".claude" / "skills"
    ]

    total_files = 0
    success_count = 0

    for directory in directories:
        print(f"\n📂 Migrating: {directory.relative_to(root)}")

        if not directory.exists():
            print(f"  ⚠️  Directory not found, skipping")
            continue

        # Find all SKILL.md files
        skill_files = sorted(directory.glob("*/SKILL.md"))

        for skill_file in skill_files:
            total_files += 1
            print(f"\n📄 {skill_file.parent.name}/SKILL.md")

            if migrate_skill(skill_file, backup=True):
                success_count += 1

    # Summary
    print(f"\n" + "="*60)
    print(f"✨ Migration Complete")
    print(f"="*60)
    print(f"Total files: {total_files}")
    print(f"Success: {success_count}")
    print(f"Failed: {total_files - success_count}")
    print(f"\n💡 Tip: Review changes with `git diff .claude/skills/`")


if __name__ == "__main__":
    main()
