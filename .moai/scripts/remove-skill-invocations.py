#!/usr/bin/env python3
"""
Remove explicit Skill() invocation patterns from agent files.
Skills are now auto-loaded via YAML frontmatter, making explicit calls redundant.
"""

import re
from pathlib import Path
from typing import Tuple

def remove_skill_patterns(content: str) -> Tuple[str, int]:
    """Remove all Skill() invocation patterns from content."""

    original_content = content
    count = 0

    # Pattern 1: Remove `Skill("skill-name")` patterns
    # Matches: `Skill("moai-domain-backend")`, Skill("moai-lang-python")
    pattern1 = r'`?Skill\("([^"]+)"\)`?'
    def replace_pattern1(match):
        nonlocal count
        count += 1
        # Just return the skill name without Skill() wrapper
        return f'{match.group(1)}'

    content = re.sub(pattern1, replace_pattern1, content)

    # Pattern 2: Remove section headers mentioning "load via Skill"
    pattern2 = r'\*\*Skills\*\* \(load via `Skill\("skill-name"\)`\):'
    content = re.sub(pattern2, '**Available Skills**:', content)

    # Pattern 3: Remove instructional text about Skill() calls
    pattern3 = r'load via `Skill\("skill-name"\)`'
    content = re.sub(pattern3, 'auto-loaded from YAML frontmatter', content)

    # Pattern 4: Clean up any remaining backtick-wrapped skill names in lists
    # Convert: - `moai-domain-backend` to: - moai-domain-backend
    # But keep context after em-dash
    pattern4 = r'^(\s*-\s+)`([^`]+)`(\s+–)'
    content = re.sub(pattern4, r'\1\2\3', content, flags=re.MULTILINE)

    # Pattern 5: Update phrases like "Invoke Skill(...)" to "Uses ..."
    pattern5 = r'Invoke\s+Skill\("([^"]+)"\)'
    content = re.sub(pattern5, r'Uses \1', content)

    # Pattern 6: Update phrases like "load Skill(...)" to "uses ..."
    pattern6 = r'load\s+Skill\("([^"]+)"\)'
    content = re.sub(pattern6, r'uses \1', content)

    # Pattern 7: Remove instructional phrases about explicit Skill invocation
    pattern7 = r'\s*-\s*Always use explicit syntax:\s*`Skill\("skill-name"\)`\n'
    content = re.sub(pattern7, '', content)

    pattern8 = r'\s*-\s*Do NOT rely on keyword matching or auto-triggering\n'
    content = re.sub(pattern8, '', content)

    # Pattern 9: Update table cells with Skill()
    pattern9 = r'\|\s*Skill\("([^"]+)"\)\s*\|'
    content = re.sub(pattern9, r'| \1 |', content)

    # Pattern 10: Update phrases in markdown like "Reference: Skill(...)"
    pattern10 = r'\*\*Reference\*\*:\s*`Skill\("([^"]+)"\)`'
    content = re.sub(pattern10, r'**Reference**: \1', content)

    # Pattern 11: Update list items with Skill() prefix
    pattern11 = r'^(\s*-\s*)`Skill\("([^"]+)"\)`(\s*$)'
    content = re.sub(pattern11, r'\1\2\3', content, flags=re.MULTILINE)

    return content, count

def process_file(file_path: Path) -> Tuple[bool, int]:
    """Process a single agent file."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            original_content = f.read()

        new_content, count = remove_skill_patterns(original_content)

        if new_content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            return True, count

        return False, 0

    except Exception as e:
        print(f"❌ Error processing {file_path}: {e}")
        return False, 0

def main():
    """Process all agent files in both directories."""

    directories = [
        Path("/Users/goos/MoAI/MoAI-ADK/.claude/agents/moai"),
        Path("/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/agents/moai")
    ]

    total_files = 0
    total_modified = 0
    total_patterns = 0

    for directory in directories:
        print(f"\n{'='*60}")
        print(f"Processing: {directory}")
        print(f"{'='*60}")

        agent_files = sorted(directory.glob("*.md"))
        dir_modified = 0
        dir_patterns = 0

        for file_path in agent_files:
            total_files += 1
            modified, count = process_file(file_path)

            if modified:
                total_modified += 1
                dir_modified += 1
                total_patterns += count
                dir_patterns += count
                print(f"✅ {file_path.name}: Removed {count} Skill() patterns")
            else:
                print(f"⚪ {file_path.name}: No changes needed")

        print(f"\n📊 Directory Summary:")
        print(f"   Files modified: {dir_modified}/{len(agent_files)}")
        print(f"   Patterns removed: {dir_patterns}")

    print(f"\n{'='*60}")
    print(f"🎯 FINAL SUMMARY")
    print(f"{'='*60}")
    print(f"Total files processed: {total_files}")
    print(f"Total files modified: {total_modified}")
    print(f"Total Skill() patterns removed: {total_patterns}")
    print(f"\n✅ All explicit Skill() invocations have been removed!")
    print(f"   Skills are now auto-loaded via YAML frontmatter.")

if __name__ == "__main__":
    main()
