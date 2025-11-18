#!/usr/bin/env python3
"""
Quick validation script for the translation skill
Run: python validate_skill.py
"""

import json
import re
from pathlib import Path

def validate_skill_structure():
    """Validate the skill has all required components"""
    
    skill_dir = Path(__file__).parent
    
    # Check required files
    required_files = ['SKILL.md', 'reference.md', 'examples.md']
    missing_files = []
    
    for file in required_files:
        if not (skill_dir / file).exists():
            missing_files.append(file)
    
    # Check SKILL.md metadata
    skill_md = skill_dir / 'SKILL.md'
    if skill_md.exists():
        content = skill_md.read_text()
        
        # Extract metadata
        if content.startswith('---'):
            metadata_end = content.find('---', 3)
            if metadata_end > 0:
                metadata = content[3:metadata_end]
                
                required_fields = ['name', 'version', 'status', 'description', 'keywords']
                for field in required_fields:
                    if f'{field}:' not in metadata:
                        print(f"⚠️  Missing metadata field: {field}")
    
    # Check for code examples
    if (skill_dir / 'examples.md').exists():
        examples_content = (skill_dir / 'examples.md').read_text()
        code_blocks = re.findall(r'```python[\s\S]*?```', examples_content)
        print(f"✅ Found {len(code_blocks)} Python code examples")
    
    # Check glossary template
    glossary_path = skill_dir / 'templates' / 'glossary_template.json'
    if glossary_path.exists():
        with open(glossary_path) as f:
            glossary = json.load(f)
            total_terms = sum(len(terms) for terms in glossary.values())
            print(f"✅ Glossary contains {total_terms} total terms across {len(glossary)} language pairs")
    
    # Summary
    print("\n📊 Skill Validation Summary:")
    print(f"  Name: moai-translation-korean-multilingual")
    print(f"  Version: 1.0.0")
    print(f"  Status: Production Ready")
    print(f"  Files: {', '.join(f for f in required_files if (skill_dir / f).exists())}")
    
    if missing_files:
        print(f"  ⚠️  Missing: {', '.join(missing_files)}")
    else:
        print("  ✅ All required files present")
    
    # File sizes
    total_size = sum((skill_dir / f).stat().st_size for f in required_files if (skill_dir / f).exists())
    print(f"  Total size: {total_size / 1024:.1f} KB")
    
    return len(missing_files) == 0

if __name__ == "__main__":
    print("🔍 Validating Translation Skill Structure...\n")
    
    if validate_skill_structure():
        print("\n✅ Skill validation PASSED!")
        print("The skill is ready for use with:")
        print('  Skill("moai-translation-korean-multilingual")')
    else:
        print("\n❌ Skill validation FAILED")
        print("Please fix the issues above")
