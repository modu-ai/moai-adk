#!/usr/bin/env python3
"""
Skill Format Validation Script
Validates skills against official Claude Code format
"""

import re
import yaml
from pathlib import Path
from typing import Dict, List, Optional
import sys

class SkillValidator:
    """Validate skill format compliance."""
    
    OFFICIAL_TOOLS = ['Read', 'Write', 'Edit', 'Bash', 'Grep', 'Glob', 'Task', 'AskUserQuestion']
    MAX_NAME_LENGTH = 64
    MAX_DESCRIPTION_LENGTH = 1024
    MAX_RECOMMENDED_LINES = 500
    
    def __init__(self):
        self.validated_count = 0
        self.warnings = []
        self.errors = []
    
    def validate_frontmatter_syntax(self, content: str) -> List[str]:
        """Check YAML frontmatter syntax."""
        errors = []
        
        # Check frontmatter exists and is at start
        if not content.startswith('---\n'):
            errors.append("Frontmatter must start with '---' at beginning of file")
            return errors
        
        # Extract frontmatter
        match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
        if not match:
            errors.append("Frontmatter not properly closed with '---'")
            return errors
        
        # Parse YAML
        try:
            yaml_data = yaml.safe_load(match.group(1))
        except yaml.YAMLError as e:
            errors.append(f"Invalid YAML syntax: {e}")
            return errors
        
        # Check for required fields
        if 'name' not in yaml_data:
            errors.append("Missing required field: 'name'")
        if 'description' not in yaml_data:
            errors.append("Missing required field: 'description'")
        
        # Check for prohibited fields (custom fields not in official spec)
        prohibited = ['version', 'status', 'tier', 'created', 'updated', 'keywords']
        found_prohibited = [f for f in prohibited if f in yaml_data]
        if found_prohibited:
            errors.append(f"Prohibited custom fields found: {found_prohibited} (remove these)")
        
        return errors
    
    def validate_name_format(self, content: str) -> List[str]:
        """Check name format compliance."""
        errors = []
        
        match = re.search(r'^name:\s*(.+)$', content, re.MULTILINE)
        if not match:
            return ["Cannot extract name from frontmatter"]
        
        name = match.group(1).strip('"\'')
        
        # Check format
        if not re.match(r'^[a-z0-9-]+$', name):
            errors.append(f"Name '{name}' must be lowercase letters, numbers, hyphens only")
        
        # Check length
        if len(name) > self.MAX_NAME_LENGTH:
            errors.append(f"Name too long: {len(name)} chars (max {self.MAX_NAME_LENGTH})")
        
        return errors
    
    def validate_description(self, content: str) -> List[str]:
        """Check description compliance."""
        errors = []
        
        match = re.search(r'^description:\s*(.+)$', content, re.MULTILINE)
        if not match:
            return ["Cannot extract description from frontmatter"]
        
        description = match.group(1).strip('"\'')
        
        # Check length
        if len(description) > self.MAX_DESCRIPTION_LENGTH:
            errors.append(f"Description too long: {len(description)} chars (max {self.MAX_DESCRIPTION_LENGTH})")
        
        # Check clarity (basic heuristics)
        if len(description) < 20:
            errors.append(f"Description too short: {len(description)} chars (should explain WHAT and WHEN)")
        
        return errors
    
    def validate_allowed_tools(self, content: str) -> List[str]:
        """Check allowed-tools format."""
        errors = []
        
        match = re.search(r'^allowed-tools:\s*(.+)$', content, re.MULTILINE)
        if not match:
            return []  # Optional field
        
        tools_str = match.group(1).strip()
        
        # Check if it's comma-separated (not array syntax)
        if tools_str.startswith('[') or tools_str.startswith('-'):
            errors.append("allowed-tools must be comma-separated string, not array syntax")
            return errors
        
        # Parse tools
        tools = [t.strip() for t in tools_str.split(',')]
        
        # Validate each tool
        invalid_tools = [t for t in tools if t not in self.OFFICIAL_TOOLS]
        if invalid_tools:
            errors.append(f"Invalid tools: {invalid_tools}. Allowed: {self.OFFICIAL_TOOLS}")
        
        return errors
    
    def validate_progressive_disclosure(self, content: str) -> List[str]:
        """Check for Progressive Disclosure structure."""
        warnings = []
        
        # Check for recommended sections
        if '## Quick Reference' not in content and '## What It Does' not in content:
            warnings.append("Missing Quick Reference section (30-second value)")
        
        if '## Implementation' not in content and '## Core' not in content:
            warnings.append("Missing Implementation Guide section")
        
        if '## Advanced' not in content:
            warnings.append("Missing Advanced Patterns section")
        
        return warnings
    
    def validate_max_lines(self, content: str) -> List[str]:
        """Check line count recommendation."""
        warnings = []
        
        line_count = len(content.splitlines())
        if line_count > self.MAX_RECOMMENDED_LINES:
            warnings.append(f"Skill has {line_count} lines (recommended <{self.MAX_RECOMMENDED_LINES})")
        
        return warnings
    
    def validate_skill(self, skill_path: Path) -> Dict:
        """Validate a single skill."""
        
        result = {
            'skill': skill_path.parent.name,
            'errors': [],
            'warnings': [],
            'passed': False
        }
        
        try:
            content = skill_path.read_text(encoding='utf-8')
            
            # Run all checks
            result['errors'].extend(self.validate_frontmatter_syntax(content))
            result['errors'].extend(self.validate_name_format(content))
            result['errors'].extend(self.validate_description(content))
            result['errors'].extend(self.validate_allowed_tools(content))
            result['warnings'].extend(self.validate_progressive_disclosure(content))
            result['warnings'].extend(self.validate_max_lines(content))
            
            result['passed'] = len(result['errors']) == 0
            
        except Exception as e:
            result['errors'].append(f"Validation exception: {e}")
        
        return result
    
    def validate_batch(self, skill_dirs: List[Path]) -> Dict:
        """Validate multiple skills."""
        
        print(f"\n{'='*60}")
        print(f"SKILL FORMAT VALIDATION")
        print(f"Skills: {len(skill_dirs)}")
        print(f"{'='*60}\n")
        
        results = []
        passed_count = 0
        
        for skill_dir in sorted(skill_dirs):
            skill_md = skill_dir / 'SKILL.md'
            if not skill_md.exists():
                continue
            
            result = self.validate_skill(skill_md)
            results.append(result)
            
            if result['passed']:
                passed_count += 1
                print(f"✓ {result['skill']}")
            else:
                print(f"✗ {result['skill']}")
                for error in result['errors']:
                    print(f"    ERROR: {error}")
            
            if result['warnings']:
                for warning in result['warnings']:
                    print(f"    WARNING: {warning}")
        
        print(f"\n{'='*60}")
        print(f"VALIDATION SUMMARY")
        print(f"{'='*60}")
        print(f"Passed: {passed_count}/{len(results)}")
        print(f"Failed: {len(results) - passed_count}")
        
        return {
            'total': len(results),
            'passed': passed_count,
            'failed': len(results) - passed_count,
            'results': results
        }


def main():
    """Main validation execution."""
    
    import argparse
    parser = argparse.ArgumentParser(description='Validate skill format compliance')
    parser.add_argument('--skill', type=str, help='Validate specific skill (default: all)')
    parser.add_argument('--tier1', action='store_true', help='Validate only Tier 1 skills (<300 lines)')
    args = parser.parse_args()
    
    skills_dir = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')
    
    skill_dirs = []
    for skill_dir in skills_dir.iterdir():
        if not skill_dir.is_dir():
            continue
        
        skill_md = skill_dir / 'SKILL.md'
        if not skill_md.exists():
            continue
        
        # Filter by tier if requested
        if args.tier1:
            line_count = len(skill_md.read_text(encoding='utf-8').splitlines())
            if line_count >= 300:
                continue
        
        # Filter by specific skill if provided
        if args.skill and skill_dir.name != args.skill:
            continue
        
        skill_dirs.append(skill_dir)
    
    if not skill_dirs:
        print("No skills found matching criteria")
        sys.exit(1)
    
    # Execute validation
    validator = SkillValidator()
    result = validator.validate_batch(skill_dirs)
    
    # Exit with error code if any failures
    sys.exit(0 if result['failed'] == 0 else 1)


if __name__ == '__main__':
    main()
