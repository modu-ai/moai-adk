#!/usr/bin/env python3
"""
Fix YAML frontmatter in MDX files

Removes markdown formatting and emojis from YAML frontmatter
to prevent parsing errors during Next.js build.
"""

import re
from pathlib import Path


def fix_frontmatter(file_path: Path):
    """Fix YAML frontmatter in an MDX file"""
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Check if file has frontmatter
    if not content.startswith('---\n'):
        return False

    # Split content
    parts = content.split('---\n', 2)
    if len(parts) < 3:
        return False

    frontmatter = parts[1]
    body = parts[2]

    # Fix frontmatter issues
    lines = frontmatter.strip().split('\n')
    fixed_lines = []

    for line in lines:
        # Skip empty lines
        if not line.strip():
            continue

        # Check if it's a key-value pair
        if ':' in line:
            key, value = line.split(':', 1)
            value = value.strip()

            # Remove markdown bold (**text**)
            value = re.sub(r'\*\*([^*]+)\*\*', r'\1', value)

            # Remove emojis (Unicode ranges for common emojis)
            value = re.sub(r'[\U0001F300-\U0001F9FF\U0001F600-\U0001F64F\U0001F680-\U0001F6FF\U0001F1E0-\U0001F1FF\U00002700-\U000027BF\U0001F900-\U0001F9FF\U0001FA70-\U0001FAFF\U00002600-\U000026FF\U00002B50]', '', value).strip()

            # Quote the value if it contains special characters (and not already quoted)
            if value and not (value.startswith('"') and value.endswith('"')):
                if any(char in value for char in [':', '#', '*', '!', '&', '|', '[', ']', '{', '}']):
                    value = f'"{value}"'

            fixed_lines.append(f'{key}: {value}')
        else:
            # Line doesn't have a colon, skip it
            continue

    # Reconstruct file
    fixed_frontmatter = '\n'.join(fixed_lines)
    fixed_content = f"---\n{fixed_frontmatter}\n---\n\n{body}"

    # Write back
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(fixed_content)

    return True


def main():
    """Main entry point"""
    docs_dir = Path(__file__).parent.parent / 'docs' / 'pages'

    print(f"🔍 Scanning {docs_dir} for MDX files...")

    mdx_files = list(docs_dir.rglob('*.mdx'))
    print(f"📄 Found {len(mdx_files)} MDX files")

    fixed_count = 0
    for mdx_file in mdx_files:
        try:
            if fix_frontmatter(mdx_file):
                fixed_count += 1
                print(f"✅ Fixed: {mdx_file.relative_to(docs_dir)}")
        except Exception as e:
            print(f"❌ Error fixing {mdx_file}: {e}")

    print(f"\n✅ Fixed {fixed_count} files")


if __name__ == '__main__':
    main()
