#!/usr/bin/env python3
"""
Fix empty allowed-tools fields by extracting from backup files.
"""

import re
from pathlib import Path
import yaml


def extract_allowed_tools_from_backup(backup_path: Path) -> list[str]:
    """Extract allowed-tools list from backup file (malformed YAML)."""
    content = backup_path.read_text(encoding='utf-8')

    # Extract frontmatter
    match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
    if not match:
        return []

    frontmatter_str = match.group(1)
    lines = frontmatter_str.split('\n')

    # Find all list items (lines starting with '  -')
    # These are orphaned allowed-tools items in malformed YAML
    tools = []

    for line in lines:
        # Match lines like "  - Read" or "  - Bash"
        if line.strip().startswith('-') and not line.strip().startswith('---'):
            tool = line.strip()[1:].strip()  # Remove '-' and whitespace
            if tool:
                tools.append(tool)

    return tools


def fix_empty_allowed_tools(skill_file: Path):
    """Fix empty allowed-tools field in SKILL.md."""
    backup_file = skill_file.with_suffix('.md.bak')

    if not backup_file.exists():
        print(f"  ⚠️  No backup file: {backup_file.name}")
        return False

    # Extract allowed-tools from backup
    tools = extract_allowed_tools_from_backup(backup_file)

    if not tools:
        print(f"  ⚠️  No tools found in backup")
        return False

    # Read current file
    content = skill_file.read_text(encoding='utf-8')

    # Parse frontmatter
    match = re.match(r'^---\n(.*?)\n---\n(.*)$', content, re.DOTALL)
    if not match:
        print(f"  ❌ No frontmatter found")
        return False

    frontmatter_str = match.group(1)
    markdown_content = match.group(2)

    # Parse YAML
    frontmatter = yaml.safe_load(frontmatter_str)

    # Update allowed-tools
    frontmatter['allowed-tools'] = tools

    # Write back
    yaml_str = yaml.dump(frontmatter, allow_unicode=True, default_flow_style=False, sort_keys=False)
    new_content = f"---\n{yaml_str}---\n{markdown_content}"
    skill_file.write_text(new_content, encoding='utf-8')

    print(f"  ✅ Fixed: {len(tools)} tools")
    return True


def main():
    """Fix all files with empty allowed-tools."""
    root = Path(__file__).parent.parent

    directories = [
        root / ".claude" / "skills",
        root / "src" / "moai_adk" / "templates" / ".claude" / "skills"
    ]

    total_fixed = 0

    for directory in directories:
        print(f"\n📂 Checking: {directory.relative_to(root)}")

        if not directory.exists():
            continue

        # Find all SKILL.md files with empty allowed-tools
        skill_files = sorted(directory.glob("*/SKILL.md"))

        for skill_file in skill_files:
            content = skill_file.read_text(encoding='utf-8')

            # Check if allowed-tools is empty
            if "allowed-tools: ''" in content:
                print(f"\n📄 {skill_file.parent.name}/SKILL.md")
                if fix_empty_allowed_tools(skill_file):
                    total_fixed += 1

    print(f"\n✨ Fixed {total_fixed} files")


if __name__ == "__main__":
    main()
