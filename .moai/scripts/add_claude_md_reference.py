#!/usr/bin/env python3
"""
Add @CLAUDE.md reference section to all agent files.

This script inserts a standardized @CLAUDE.md reference section
into all agent files in .claude/agents/moai/ directory.

Usage:
    python add_claude_md_reference.py
"""

import re
from pathlib import Path
from typing import List, Tuple

# Reference section to insert
REFERENCE_SECTION = """
## 📋 Essential Reference

**IMPORTANT**: This agent follows Alfred's core execution directives defined in @CLAUDE.md:

- **Rule 1**: 8-Step User Request Analysis Process
- **Rule 3**: Behavioral Constraints (Never execute directly, always delegate)
- **Rule 5**: Agent Delegation Guide (7-Tier hierarchy, naming patterns)
- **Rule 6**: Foundation Knowledge Access (Conditional auto-loading)

For complete execution guidelines and mandatory rules, refer to @CLAUDE.md.

---
"""

def find_insertion_point(content: str) -> int:
    """
    Find the best insertion point for @CLAUDE.md reference.

    Priority:
    1. Before "## 🎭 Agent Persona" section
    2. Before "## 🌍 Language Handling" section
    3. After frontmatter (before first ## heading)

    Returns line index (0-based) where section should be inserted.
    """
    lines = content.split('\n')

    # Check if reference already exists
    if '@CLAUDE.md' in content:
        return -1  # Already has reference

    # Find frontmatter end
    frontmatter_end = 0
    in_frontmatter = False
    for i, line in enumerate(lines):
        if line.strip() == '---':
            if not in_frontmatter:
                in_frontmatter = True
            else:
                frontmatter_end = i + 1
                break

    # Priority 1: Before Agent Persona
    for i in range(frontmatter_end, len(lines)):
        if lines[i].startswith('## 🎭 Agent Persona'):
            return i

    # Priority 2: Before Language Handling
    for i in range(frontmatter_end, len(lines)):
        if lines[i].startswith('## 🌍 Language Handling'):
            return i

    # Priority 3: Before first ## heading after frontmatter
    for i in range(frontmatter_end, len(lines)):
        if lines[i].startswith('##'):
            return i

    # Fallback: After frontmatter
    return frontmatter_end


def process_agent_file(file_path: Path) -> Tuple[bool, str]:
    """
    Process a single agent file and add @CLAUDE.md reference.

    Returns:
        (success: bool, message: str)
    """
    try:
        content = file_path.read_text(encoding='utf-8')

        # Find insertion point
        insertion_line = find_insertion_point(content)

        if insertion_line == -1:
            return (False, f"⏭️  Skipped (already has @CLAUDE.md reference)")

        # Insert reference section
        lines = content.split('\n')
        lines.insert(insertion_line, REFERENCE_SECTION.strip())

        # Write back
        new_content = '\n'.join(lines)
        file_path.write_text(new_content, encoding='utf-8')

        return (True, f"✅ Added @CLAUDE.md reference at line {insertion_line + 1}")

    except Exception as e:
        return (False, f"❌ Error: {e}")


def main():
    """Main execution function."""
    agent_dir = Path('.claude/agents/moai')

    if not agent_dir.exists():
        print(f"❌ Agent directory not found: {agent_dir}")
        return

    # Get all agent markdown files
    agent_files = sorted(agent_dir.glob('*.md'))

    if not agent_files:
        print(f"❌ No agent files found in {agent_dir}")
        return

    print(f"📋 Found {len(agent_files)} agent files\n")

    success_count = 0
    skip_count = 0
    error_count = 0

    for agent_file in agent_files:
        success, message = process_agent_file(agent_file)

        print(f"{agent_file.name:30} → {message}")

        if success:
            success_count += 1
        elif 'Skipped' in message:
            skip_count += 1
        else:
            error_count += 1

    print(f"\n📊 Summary:")
    print(f"   ✅ Added: {success_count}")
    print(f"   ⏭️  Skipped: {skip_count}")
    print(f"   ❌ Errors: {error_count}")
    print(f"   📝 Total: {len(agent_files)}")


if __name__ == '__main__':
    main()
