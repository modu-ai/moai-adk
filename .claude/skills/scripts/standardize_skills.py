#!/usr/bin/env python3
"""
Skills standardization unified script
- YAML field cleanup (remove version, author, license, tags, model)
- Add allowed-tools (only when missing)
"""

import sys
from pathlib import Path

# Alfred agent tools
ALFRED_TOOLS = ["Read", "Write", "Edit", "Bash", "TodoWrite"]
# Language skill tools
LANG_TOOLS = ["Read", "Bash"]
# Domain skill tools
DOMAIN_TOOLS = ["Read", "Bash"]

def parse_yaml_frontmatter(content):
    """Parse YAML frontmatter (simple parser)"""
    if not content.startswith('---'):
        return None, content

    parts = content.split('---', 2)
    if len(parts) < 3:
        return None, content

    yaml_str = parts[1]
    body = parts[2]

    # Parse YAML into a dictionary
    data = {}
    current_key = None
    in_list = False

    for line in yaml_str.strip().split('\n'):
        if not line.strip():
            continue

        # list items
        if line.strip().startswith('-'):
            if in_list and current_key:
                if isinstance(data[current_key], list):
                    data[current_key].append(line.strip()[1:].strip())
            continue

        # key-value pairs
        if ':' in line:
            key, value = line.split(':', 1)
            key = key.strip()
            value = value.strip()

            if not value: # start of list
                in_list = True
                current_key = key
                data[key] = []
            else:
                in_list = False
                current_key = None
                data[key] = value

    return data, body

def build_yaml_frontmatter(data):
    """Convert dictionary to YAML frontmatter"""
    lines = []
    for key, value in data.items():
        if isinstance(value, list):
            lines.append(f"{key}:")
            for item in value:
                lines.append(f"  - {item}")
        else:
            lines.append(f"{key}: {value}")

    return '\n'.join(lines)

def standardize_skill(skill_file):
    """Standardize skill file"""
    content = skill_file.read_text()

    data, body = parse_yaml_frontmatter(content)

    if data is None:
        print(f"⚠️  No YAML frontmatter: {skill_file}")
        return False

    # Extract only fields to keep
    preserved = {}

    if 'name' in data:
        preserved['name'] = data['name']
    if 'description' in data:
        preserved['description'] = data['description']

    # Handle allowed-tools
    if 'allowed-tools' in data:
        # Keep if already present
        preserved['allowed-tools'] = data['allowed-tools']
    else:
        # Add based on skill type if missing
        name = data.get('name', '')

        if 'alfred' in name:
            tools = ALFRED_TOOLS
        elif 'lang' in name:
            tools = LANG_TOOLS
        elif 'domain' in name:
            tools = DOMAIN_TOOLS
        elif 'claude-code' in name:
            # moai-claude-code already has allowed-tools (skip)
            tools = None
        else:
            # Default
            tools = ["Read"]

        if tools:
            preserved['allowed-tools'] = tools

    # Rewrite file
    new_yaml = build_yaml_frontmatter(preserved)
    new_content = f"---\n{new_yaml}\n---{body}"

    skill_file.write_text(new_content)
    print(f"✅ Standardized: {skill_file.name}")
    return True

def main():
    """Main function"""
    base_dir = Path("/Users/goos/MoAI/MoAI-ADK")

    # .claude/skills/
    skills_dir = base_dir / ".claude/skills"
    success_count = 0
    fail_count = 0

    for skill_dir in sorted(skills_dir.glob("moai-*")):
        skill_file = skill_dir / "SKILL.md"
        if skill_file.exists():
            try:
                if standardize_skill(skill_file):
                    success_count += 1
                else:
                    fail_count += 1
            except Exception as e:
                print(f"❌ Error: {skill_file.name} - {e}")
                fail_count += 1

    # src/moai_adk/templates/.claude/skills/
    templates_dir = base_dir / "src/moai_adk/templates/.claude/skills"
    for skill_dir in sorted(templates_dir.glob("moai-*")):
        skill_file = skill_dir / "SKILL.md"
        if skill_file.exists():
            try:
                if standardize_skill(skill_file):
                    success_count += 1
                else:
                    fail_count += 1
            except Exception as e:
                print(f"❌ Error: {skill_file.name} - {e}")
                fail_count += 1

    print(f"\n📊 Summary: {success_count} succeeded, {fail_count} failed")
    return 0 if fail_count == 0 else 1

if __name__ == "__main__":
    sys.exit(main())
