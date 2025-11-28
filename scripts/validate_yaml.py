#!/usr/bin/env python3
"""
Validate YAML frontmatter in all MDX files
"""

import yaml
from pathlib import Path


def validate_frontmatter(file_path: Path):
    """Validate YAML frontmatter"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        if not content.startswith('---'):
            return True, "No frontmatter"

        parts = content.split('---\n', 2)
        if len(parts) < 3:
            return False, "Incomplete frontmatter"

        frontmatter = parts[1]

        # Try to parse YAML
        yaml.safe_load(frontmatter)
        return True, "Valid"

    except yaml.YAMLError as e:
        return False, str(e)
    except Exception as e:
        return False, str(e)


def main():
    docs_dir = Path(__file__).parent.parent / 'docs' / 'pages'
    mdx_files = list(docs_dir.rglob('*.mdx'))

    print(f"🔍 Validating {len(mdx_files)} MDX files...\n")

    errors = []
    for mdx_file in mdx_files:
        is_valid, message = validate_frontmatter(mdx_file)
        if not is_valid:
            errors.append((mdx_file, message))
            print(f"❌ {mdx_file.relative_to(docs_dir)}")
            print(f"   Error: {message}\n")

    if not errors:
        print("\n✅ All frontmatter valid!")
    else:
        print(f"\n❌ Found {len(errors)} errors")

    return len(errors)


if __name__ == '__main__':
    import sys
    sys.exit(main())
