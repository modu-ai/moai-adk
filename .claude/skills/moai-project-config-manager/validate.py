#!/usr/bin/env python3
"""
Validate the moai-project-config-manager skill structure and metadata
"""

from pathlib import Path

import yaml


def validate_skill():
    """Validate skill structure and metadata"""

    print("🔍 Validating moai-project-config-manager skill...")

    skill_dir = Path(__file__).parent
    skill_file = skill_dir / "SKILL.md"

    if not skill_file.exists():
        print("❌ SKILL.md not found")
        return False

    # Read and parse frontmatter
    with open(skill_file, 'r') as f:
        content = f.read()

    # Extract frontmatter
    if content.startswith('---'):
        try:
            end_index = content.find('---', 3)
            frontmatter_str = content[3:end_index].strip()
            frontmatter = yaml.safe_load(frontmatter_str)

            print("✅ Frontmatter parsed successfully")

            # Validate required fields
            required_fields = ['name', 'version', 'description', 'freedom', 'type', 'tags']
            missing_fields = [field for field in required_fields if field not in frontmatter]

            if missing_fields:
                print(f"❌ Missing required fields: {missing_fields}")
                return False

            print("✅ All required fields present")

            # Validate field values
            if frontmatter['name'] != 'moai-project-config-manager':
                print(f"❌ Incorrect skill name: {frontmatter['name']}")
                return False

            if frontmatter['type'] != 'project':
                print(f"❌ Incorrect skill type: {frontmatter['type']}")
                return False

            if frontmatter['freedom'] not in ['low', 'medium', 'high']:
                print(f"❌ Invalid freedom level: {frontmatter['freedom']}")
                return False

            if not isinstance(frontmatter['tags'], list):
                print(f"❌ Tags must be a list: {frontmatter['tags']}")
                return False

            print("✅ Field values validated")

            # Check expected tags
            expected_tags = ['project', 'configuration', 'management']
            has_expected_tags = any(tag in frontmatter['tags'] for tag in expected_tags)

            if not has_expected_tags:
                print(f"⚠️ Missing expected tags: {expected_tags}")
            else:
                print("✅ Expected tags present")

        except yaml.YAMLError as e:
            print(f"❌ Failed to parse frontmatter: {e}")
            return False

    # Check for required files
    required_files = ['SKILL.md', 'reference.md', 'examples.md']
    missing_files = [f for f in required_files if not (skill_dir / f).exists()]

    if missing_files:
        print(f"❌ Missing required files: {missing_files}")
        return False

    print("✅ All required files present")

    # Validate file sizes
    for file_name in required_files:
        file_path = skill_dir / file_name
        size_kb = file_path.stat().st_size / 1024

        if file_name == 'SKILL.md' and size_kb < 10:
            print(f"⚠️ {file_name} seems small: {size_kb:.1f}KB")
        elif file_name == 'reference.md' and size_kb < 5:
            print(f"⚠️ {file_name} seems small: {size_kb:.1f}KB")
        elif file_name == 'examples.md' and size_kb < 5:
            print(f"⚠️ {file_name} seems small: {size_kb:.1f}KB")
        else:
            print(f"✅ {file_name}: {size_kb:.1f}KB")

    print("\n🎉 Skill validation completed successfully!")
    return True

if __name__ == "__main__":
    validate_skill()
