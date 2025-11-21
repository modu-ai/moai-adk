#!/usr/bin/env python3
"""
Phase 2: Progressive Disclosure Structure Implementation
Process Tier 1 and Tier 2 skills simultaneously
"""

import os
import re
from pathlib import Path
from typing import Tuple, List

SKILLS_DIR = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")
TEMPLATES_DIR = Path("/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills")

TIER1_STRUCTURE = """
## Quick Reference (30 seconds)

Core patterns and immediate use cases. Copy-paste ready snippets.

{quick_content}

---

## Implementation Guide

Step-by-step setup and common patterns.

{implementation_content}

---

## Advanced Patterns

Complex use cases, optimization techniques, and edge cases.

{advanced_content}
"""

def count_lines(file_path: Path) -> int:
    """Count lines in file"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return len(f.readlines())
    except Exception:
        return 0

def categorize_skill(line_count: int) -> str:
    """Categorize skill by line count"""
    if line_count < 500:
        return "tier1"
    elif line_count < 1000:
        return "tier2"
    else:
        return "tier3"

def extract_sections(content: str) -> dict:
    """Extract existing sections from skill content"""
    sections = {
        'metadata': '',
        'quick': '',
        'implementation': '',
        'advanced': '',
        'footer': ''
    }
    
    # Extract metadata (YAML frontmatter)
    yaml_match = re.match(r'^---\n(.*?)\n---\n', content, re.DOTALL)
    if yaml_match:
        sections['metadata'] = f"---\n{yaml_match.group(1)}\n---\n"
        content = content[yaml_match.end():]
    
    # Split by major headers
    parts = re.split(r'\n## ', content)
    
    if parts:
        sections['quick'] = parts[0] if len(parts) > 0 else ''
    
    # Detect existing sections
    for part in parts[1:]:
        lower_part = part.lower()
        if 'quick' in lower_part or 'reference' in lower_part:
            sections['quick'] += '\n## ' + part
        elif 'implementation' in lower_part or 'usage' in lower_part:
            sections['implementation'] += '\n## ' + part
        elif 'advanced' in lower_part or 'expert' in lower_part:
            sections['advanced'] += '\n## ' + part
        else:
            sections['implementation'] += '\n## ' + part
    
    return sections

def apply_tier1_structure(skill_path: Path) -> bool:
    """Apply 3-level progressive disclosure to Tier 1 skill"""
    try:
        with open(skill_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Check if already has progressive structure
        if '## Quick Reference' in content and '## Implementation Guide' in content:
            return False  # Already processed
        
        sections = extract_sections(content)
        
        # Reorganize with clear progressive structure
        new_content = sections['metadata'] + "\n"
        new_content += "## Quick Reference (30 seconds)\n\n"
        new_content += sections['quick'].strip() + "\n\n"
        new_content += "---\n\n"
        new_content += "## Implementation Guide\n\n"
        new_content += sections['implementation'].strip() + "\n\n"
        new_content += "---\n\n"
        new_content += "## Advanced Patterns\n\n"
        new_content += sections['advanced'].strip() + "\n\n"
        
        # Write back
        with open(skill_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        
        return True
    except Exception as e:
        print(f"Error processing {skill_path}: {e}")
        return False

def apply_tier2_structure(skill_path: Path) -> Tuple[bool, int]:
    """Apply optimization + reference.md to Tier 2 skill"""
    try:
        with open(skill_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        sections = extract_sections(content)
        original_lines = len(content.split('\n'))
        
        # Extract detailed content to reference.md
        reference_content = ""
        if 'API Reference' in content or 'Official References' in content:
            # Extract reference sections
            ref_match = re.search(r'## (API Reference|Official References).*', content, re.DOTALL)
            if ref_match:
                reference_content = ref_match.group(0)
                content = content[:ref_match.start()]
        
        # Apply progressive structure to main content
        sections = extract_sections(content)
        
        new_content = sections['metadata'] + "\n"
        new_content += "## Quick Reference (30 seconds)\n\n"
        new_content += sections['quick'].strip() + "\n\n"
        new_content += "---\n\n"
        new_content += "## Core Implementation\n\n"
        new_content += sections['implementation'].strip() + "\n\n"
        
        if reference_content:
            new_content += "\n\n---\n\n"
            new_content += "## Reference & Resources\n\n"
            new_content += "See [reference.md](reference.md) for detailed API reference and official documentation.\n"
        
        # Write main SKILL.md
        with open(skill_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        
        # Write reference.md if content extracted
        if reference_content:
            ref_path = skill_path.parent / "reference.md"
            with open(ref_path, 'w', encoding='utf-8') as f:
                f.write(reference_content)
        
        new_lines = len(new_content.split('\n'))
        reduction = int((1 - new_lines / original_lines) * 100)
        
        return True, reduction
    except Exception as e:
        print(f"Error processing {skill_path}: {e}")
        return False, 0

def process_all_skills():
    """Process all skills in parallel"""
    tier1_count = 0
    tier1_success = 0
    tier2_count = 0
    tier2_success = 0
    tier2_reduction_total = 0
    
    print("PHASE 2: Progressive Disclosure Implementation")
    print("=" * 60)
    
    # Collect all skills
    skill_files = list(SKILLS_DIR.glob("*/SKILL.md"))
    
    for skill_file in skill_files:
        line_count = count_lines(skill_file)
        tier = categorize_skill(line_count)
        
        if tier == "tier1":
            tier1_count += 1
            if apply_tier1_structure(skill_file):
                tier1_success += 1
                print(f"✓ Tier 1: {skill_file.parent.name} ({line_count} lines)")
        
        elif tier == "tier2":
            tier2_count += 1
            success, reduction = apply_tier2_structure(skill_file)
            if success:
                tier2_success += 1
                tier2_reduction_total += reduction
                print(f"✓ Tier 2: {skill_file.parent.name} ({line_count} → {reduction}% reduction)")
    
    print("\n" + "=" * 60)
    print("TIER 1 RESULTS:")
    print(f"  Processed: {tier1_success}/{tier1_count}")
    print(f"  Success Rate: {int(tier1_success/tier1_count*100) if tier1_count > 0 else 0}%")
    
    print("\nTIER 2 RESULTS:")
    print(f"  Processed: {tier2_success}/{tier2_count}")
    print(f"  Average Reduction: {int(tier2_reduction_total/tier2_success) if tier2_success > 0 else 0}%")
    print(f"  Success Rate: {int(tier2_success/tier2_count*100) if tier2_count > 0 else 0}%")
    
    return tier1_success, tier2_success

if __name__ == "__main__":
    tier1, tier2 = process_all_skills()
    print(f"\n✓ PHASE 2A COMPLETE: {tier1} Tier 1 + {tier2} Tier 2 skills processed")

