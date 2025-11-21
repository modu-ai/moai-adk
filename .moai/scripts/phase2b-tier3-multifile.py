#!/usr/bin/env python3
"""
Phase 2B: Tier 3 Multi-file Architecture
Create multi-file structure for 1000+ line skills
"""

import os
import re
from pathlib import Path
from typing import Tuple

SKILLS_DIR = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

def extract_examples(content: str) -> str:
    """Extract examples section from content"""
    examples = []
    
    # Find example blocks
    example_patterns = [
        r'## Examples?(.*?)(?=\n## |\Z)',
        r'## Usage Examples?(.*?)(?=\n## |\Z)',
        r'```[a-z]*\n(.*?)```',
    ]
    
    for pattern in example_patterns:
        matches = re.finditer(pattern, content, re.DOTALL | re.IGNORECASE)
        for match in matches:
            examples.append(match.group(0))
    
    return '\n\n'.join(examples) if examples else ""

def extract_reference(content: str) -> str:
    """Extract API reference and documentation links"""
    reference_sections = []
    
    # Find reference sections
    ref_patterns = [
        r'## API Reference(.*?)(?=\n## |\Z)',
        r'## Official References?(.*?)(?=\n## |\Z)',
        r'## Documentation(.*?)(?=\n## |\Z)',
        r'## Links(.*?)(?=\n## |\Z)',
    ]
    
    for pattern in ref_patterns:
        matches = re.finditer(pattern, content, re.DOTALL | re.IGNORECASE)
        for match in matches:
            reference_sections.append(match.group(0))
    
    return '\n\n'.join(reference_sections) if reference_sections else ""

def create_multifile_structure(skill_path: Path) -> Tuple[bool, dict]:
    """Create multi-file architecture for Tier 3 skill"""
    try:
        with open(skill_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_lines = len(content.split('\n'))
        
        # Extract metadata
        yaml_match = re.match(r'^---\n(.*?)\n---\n', content, re.DOTALL)
        metadata = yaml_match.group(0) if yaml_match else ''
        content_body = content[yaml_match.end():] if yaml_match else content
        
        # Extract examples
        examples_content = extract_examples(content_body)
        
        # Extract reference
        reference_content = extract_reference(content_body)
        
        # Remove extracted content from main body
        if examples_content:
            for example in examples_content.split('\n\n'):
                content_body = content_body.replace(example, '')
        
        if reference_content:
            for ref in reference_content.split('\n\n'):
                content_body = content_body.replace(ref, '')
        
        # Create main SKILL.md with quick reference and core implementation
        main_sections = content_body.split('## ')
        quick_content = main_sections[0] if main_sections else ''
        implementation_content = '\n## '.join(main_sections[1:4]) if len(main_sections) > 1 else ''
        
        # Build optimized SKILL.md
        new_main = metadata + "\n"
        new_main += "## Quick Reference (30 seconds)\n\n"
        new_main += quick_content.strip() + "\n\n"
        new_main += "---\n\n"
        new_main += "## Core Implementation\n\n"
        new_main += implementation_content.strip() + "\n\n"
        new_main += "---\n\n"
        new_main += "## Additional Resources\n\n"
        new_main += "- **Examples**: See [examples.md](examples.md) for practical code samples\n"
        new_main += "- **API Reference**: See [reference.md](reference.md) for detailed documentation\n"
        
        # Write SKILL.md
        with open(skill_path, 'w', encoding='utf-8') as f:
            f.write(new_main)
        
        # Write examples.md
        if examples_content:
            examples_path = skill_path.parent / "examples.md"
            with open(examples_path, 'w', encoding='utf-8') as f:
                f.write("# Examples\n\n")
                f.write(examples_content)
        
        # Write reference.md
        if reference_content:
            reference_path = skill_path.parent / "reference.md"
            with open(reference_path, 'w', encoding='utf-8') as f:
                f.write("# API Reference & Documentation\n\n")
                f.write(reference_content)
        
        # Calculate reduction
        new_lines = len(new_main.split('\n'))
        reduction = int((1 - new_lines / original_lines) * 100)
        
        stats = {
            'original': original_lines,
            'main': new_lines,
            'examples': len(examples_content.split('\n')) if examples_content else 0,
            'reference': len(reference_content.split('\n')) if reference_content else 0,
            'reduction': reduction
        }
        
        return True, stats
    except Exception as e:
        print(f"Error processing {skill_path}: {e}")
        return False, {}

def process_tier3_skills():
    """Process all Tier 3 skills (1000+ lines)"""
    print("\nPHASE 2B: TIER 3 MULTI-FILE ARCHITECTURE")
    print("=" * 60)
    
    tier3_skills = []
    skill_files = list(SKILLS_DIR.glob("*/SKILL.md"))
    
    for skill_file in skill_files:
        try:
            with open(skill_file, 'r', encoding='utf-8') as f:
                line_count = len(f.readlines())
            
            if line_count >= 1000:
                tier3_skills.append((skill_file, line_count))
        except Exception:
            continue
    
    print(f"Found {len(tier3_skills)} Tier 3 skills\n")
    
    total_reduction = 0
    success_count = 0
    
    for skill_file, original_lines in tier3_skills:
        success, stats = create_multifile_structure(skill_file)
        
        if success:
            success_count += 1
            total_reduction += stats.get('reduction', 0)
            
            print(f"✓ {skill_file.parent.name}")
            print(f"  Original: {stats['original']} lines")
            print(f"  SKILL.md: {stats['main']} lines")
            print(f"  examples.md: {stats['examples']} lines")
            print(f"  reference.md: {stats['reference']} lines")
            print(f"  Reduction: {stats['reduction']}%\n")
    
    print("=" * 60)
    print(f"TIER 3 RESULTS:")
    print(f"  Processed: {success_count}/{len(tier3_skills)}")
    print(f"  Average Reduction: {int(total_reduction/success_count) if success_count > 0 else 0}%")
    print(f"  Success Rate: {int(success_count/len(tier3_skills)*100) if tier3_skills else 0}%")
    
    return success_count

if __name__ == "__main__":
    count = process_tier3_skills()
    print(f"\n✓ PHASE 2B COMPLETE: {count} Tier 3 skills with multi-file architecture")

