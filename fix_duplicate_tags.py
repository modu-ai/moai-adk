#!/usr/bin/env python3
"""
Script to fix duplicate TAGs in the MoAI-ADK codebase.
"""

import re
import subprocess
from pathlib import Path
from typing import Dict, List, Tuple

def get_duplicate_tags() -> Dict[str, List[Tuple[str, int]]]:
    """Get all duplicate TAGs using the TAG validator."""
    from src.moai_adk.core.tags.pre_commit_validator import PreCommitValidator

    validator = PreCommitValidator()

    # Get all Python files
    result = subprocess.run(['find', 'src', 'tests', '-name', '*.py'],
                           capture_output=True, text=True)
    if result.returncode != 0:
        return {}

    files = result.stdout.strip().split('\n')
    files = [f for f in files if f]

    # Get duplicates
    errors = validator.validate_duplicates(files)

    duplicates = {}
    for error in errors:
        tag = error.tag
        locations = error.locations
        duplicates[tag] = locations

    return duplicates

def generate_unique_tag(original_tag: str, file_path: str, line_num: int) -> str:
    """Generate a unique TAG based on file path and original tag."""
    # Extract domain and type from original tag
    match = re.match(r'@([A-Z]+):([A-Z]+(?:-[A-Z]+)*)-(\d{3})', original_tag)
    if not match:
        return original_tag

    prefix, domain_type, number = match.groups()

    # Use file path to generate unique domain
    path_parts = Path(file_path).parts

    # Find relevant directory names
    domain_candidates = []
    for part in reversed(path_parts):
        if part in ['src', 'tests']:
            continue
        if part in ['core', 'cli', 'utils', 'tags']:
            domain_candidates.append(part.upper())
        elif len(part) > 3 and part.replace('_', '').replace('-', '').isalnum():
            domain_candidates.append(part.upper().replace('-', '_'))

    # Select domain based on file location
    if 'tags' in file_path:
        domain = f"TAG-{domain_type}"
    elif 'core' in file_path:
        domain = f"CORE-{domain_type}"
    elif 'cli' in file_path:
        domain = f"CLI-{domain_type}"
    elif 'utils' in file_path:
        domain = f"UTILS-{domain_type}"
    else:
        domain = domain_type

    # If we already have a good domain, use it
    if '-' in domain_type:
        domain = domain_type

    return f"@{prefix}:{domain}-{number}"

def fix_duplicate_tags():
    """Fix all duplicate TAGs."""
    duplicates = get_duplicate_tags()

    if not duplicates:
        print("✅ No duplicate TAGs found!")
        return

    print(f"🔧 Found {len(duplicates)} duplicate TAGs to fix:")

    # Create mapping for each tag
    tag_mappings = {}

    for tag, locations in duplicates.items():
        print(f"\n📍 Fixing duplicate tag: {tag}")
        print(f"   Found in {len(locations)} locations")

        # Generate unique tags for each location
        unique_tags = []
        for file_path, line_num in locations:
            unique_tag = generate_unique_tag(tag, file_path, line_num)

            # Ensure uniqueness
            counter = 1
            original_unique = unique_tag
            while unique_tag in unique_tags:
                # If still duplicate, add a suffix
                match = re.match(r'(@[A-Z]+):([A-Z]+(?:-[A-Z]+)*)-(\d{3})', original_unique)
                if match:
                    prefix, domain, number = match.groups()
                    # Use last two digits + counter
                    new_num = f"{int(number[-2:]) + counter:03d}"
                    unique_tag = f"@{prefix}:{domain}-{new_num}"

            unique_tags.append(unique_tag)
            print(f"   {file_path}:{line_num} → {unique_tag}")

        tag_mappings[tag] = list(zip(locations, unique_tags))

    # Apply fixes
    print(f"\n🔨 Applying fixes...")
    files_to_modify = {}

    for tag, mappings in tag_mappings.items():
        for (file_path, line_num), new_tag in mappings:
            if file_path not in files_to_modify:
                files_to_modify[file_path] = []
            files_to_modify[file_path].append((line_num, tag, new_tag))

    # Modify files
    for file_path, changes in files_to_modify.items():
        print(f"   📝 Modifying {file_path}")

        # Read file
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        # Apply changes (from bottom to top to preserve line numbers)
        changes.sort(reverse=True, key=lambda x: x[0])

        for line_num, old_tag, new_tag in changes:
            if 0 <= line_num - 1 < len(lines):
                lines[line_num - 1] = lines[line_num - 1].replace(old_tag, new_tag)

        # Write file
        with open(file_path, 'w', encoding='utf-8') as f:
            f.writelines(lines)

    print(f"\n✅ Fixed {len(tag_mappings)} duplicate TAG groups!")

    # Verify fixes
    print(f"\n🔍 Verifying fixes...")
    remaining_duplicates = get_duplicate_tags()

    if remaining_duplicates:
        print(f"⚠️  Still have {len(remaining_duplicates)} duplicate TAGs:")
        for tag, locations in remaining_duplicates.items():
            print(f"   {tag}: {len(locations)} locations")
    else:
        print("✅ All duplicate TAGs fixed!")

if __name__ == "__main__":
    fix_duplicate_tags()