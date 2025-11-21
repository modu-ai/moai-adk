#!/usr/bin/env python3
"""
Tier 1 Skills Migration Script - Option 2: Implicit Tool Permissions
Transforms skills to official Claude Code format WITHOUT allowed-tools field
"""

import re
import yaml
from pathlib import Path
from typing import Dict, Optional, Tuple
import sys

class SkillMigratorV2:
    """Migrate skills to official format (implicit tool permissions)."""
    
    def __init__(self, dry_run: bool = True):
        self.dry_run = dry_run
        self.migrated_count = 0
        self.errors = []
    
    def normalize_frontmatter(self, content: str) -> Tuple[str, Dict]:
        """Extract and normalize YAML frontmatter (name + description only)."""
        
        # Extract existing frontmatter
        match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
        if not match:
            raise ValueError("No frontmatter found")
        
        yaml_str = match.group(1)
        yaml_data = yaml.safe_load(yaml_str)
        
        # Extract required fields only (NO allowed-tools)
        name = yaml_data.get('name', '').strip('"\'')
        description = yaml_data.get('description', '').strip('"\'')
        
        if not name or not description:
            raise ValueError(f"Missing required fields: name={name}, description={description}")
        
        # Validate name format
        if not re.match(r'^[a-z0-9-]+$', name):
            raise ValueError(f"Invalid name format: {name} (must be lowercase, hyphens, numbers only)")
        
        if len(name) > 64:
            raise ValueError(f"Name too long: {len(name)} chars (max 64)")
        
        if len(description) > 1024:
            description = description[:1021] + '...'
        
        # Generate official YAML (name + description ONLY)
        official_frontmatter = {
            'name': name,
            'description': description
        }
        
        new_frontmatter = f"---\nname: {name}\ndescription: {description}\n---\n"
        
        return new_frontmatter, yaml_data
    
    def remove_metadata_table(self, content: str) -> str:
        """Remove redundant 'Skill Metadata' table section."""
        
        # Pattern matches entire metadata section
        pattern = r'## Skill Metadata\n\n\| Field \| Value \|.*?\n---\n\n'
        return re.sub(pattern, '', content, flags=re.DOTALL)
    
    def restructure_content(self, content: str, skill_name: str) -> str:
        """Apply Progressive Disclosure structure (minimal transformation)."""
        
        # For Tier 1 (simple skills), keep structure mostly intact
        # Just ensure sections are properly organized
        
        # Extract title
        title_match = re.search(r'^# (.+)$', content, re.MULTILINE)
        title = title_match.group(1) if title_match else skill_name
        
        return content
    
    def migrate_skill(self, skill_path: Path) -> bool:
        """Migrate a single skill file."""
        
        try:
            print(f"\nMigrating: {skill_path.parent.name}")
            
            # Read original content
            original_content = skill_path.read_text(encoding='utf-8')
            
            # Step 1: Normalize frontmatter (name + description only)
            new_frontmatter, old_yaml = self.normalize_frontmatter(original_content)
            
            # Step 2: Remove frontmatter from content
            content_without_frontmatter = re.sub(
                r'^---\n.*?\n---\n', 
                '', 
                original_content, 
                count=1, 
                flags=re.DOTALL
            )
            
            # Step 3: Remove metadata table
            content_cleaned = self.remove_metadata_table(content_without_frontmatter)
            
            # Step 4: Restructure content
            content_restructured = self.restructure_content(
                content_cleaned, 
                old_yaml.get('name', '')
            )
            
            # Combine new frontmatter + cleaned content
            migrated_content = new_frontmatter + content_restructured
            
            # Validate result
            if not migrated_content.startswith('---\n'):
                raise ValueError("Frontmatter not at start of file")
            
            # Write or show result
            if self.dry_run:
                print(f"  DRY RUN - Would write {len(migrated_content)} chars")
                print(f"  Original: {len(original_content)} chars")
                print(f"  New frontmatter:\n{new_frontmatter}")
            else:
                skill_path.write_text(migrated_content, encoding='utf-8')
                print(f"  ✓ Migrated successfully")
            
            self.migrated_count += 1
            return True
            
        except Exception as e:
            error_msg = f"ERROR migrating {skill_path.parent.name}: {e}"
            print(f"  ✗ {error_msg}")
            self.errors.append((skill_path.parent.name, str(e)))
            return False
    
    def migrate_batch(self, skill_dirs: list[Path]) -> Dict:
        """Migrate multiple skills."""
        
        print(f"\n{'='*60}")
        print(f"TIER 1 SKILLS MIGRATION (Option 2: Implicit Permissions)")
        print(f"Mode: {'DRY RUN' if self.dry_run else 'LIVE MIGRATION'}")
        print(f"Skills: {len(skill_dirs)}")
        print(f"{'='*60}")
        
        for skill_dir in skill_dirs:
            skill_md = skill_dir / 'SKILL.md'
            if skill_md.exists():
                self.migrate_skill(skill_md)
        
        print(f"\n{'='*60}")
        print(f"MIGRATION SUMMARY")
        print(f"{'='*60}")
        print(f"Migrated: {self.migrated_count}/{len(skill_dirs)}")
        print(f"Errors: {len(self.errors)}")
        
        if self.errors:
            print(f"\nErrors:")
            for skill_name, error in self.errors:
                print(f"  - {skill_name}: {error}")
        
        return {
            'total': len(skill_dirs),
            'migrated': self.migrated_count,
            'errors': len(self.errors),
            'error_details': self.errors
        }


def main():
    """Main migration execution."""
    
    import argparse
    parser = argparse.ArgumentParser(description='Migrate Tier 1 skills (Option 2: Implicit Permissions)')
    parser.add_argument('--live', action='store_true', help='Execute migration (default is dry-run)')
    parser.add_argument('--skill', type=str, help='Migrate specific skill (default: all tier 1)')
    args = parser.parse_args()
    
    # Find Tier 1 skills (<500 lines)
    skills_dir = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')
    
    tier1_skills = []
    for skill_dir in sorted(skills_dir.iterdir()):
        if not skill_dir.is_dir():
            continue
        
        skill_md = skill_dir / 'SKILL.md'
        if not skill_md.exists():
            continue
        
        # Check line count
        line_count = len(skill_md.read_text(encoding='utf-8').splitlines())
        if line_count < 500:
            tier1_skills.append(skill_dir)
    
    print(f"Found {len(tier1_skills)} Tier 1 skills (<500 lines)")
    
    # Filter by specific skill if provided
    if args.skill:
        tier1_skills = [s for s in tier1_skills if s.name == args.skill]
        if not tier1_skills:
            print(f"ERROR: Skill '{args.skill}' not found")
            sys.exit(1)
    
    # Execute migration
    migrator = SkillMigratorV2(dry_run=not args.live)
    result = migrator.migrate_batch(tier1_skills)
    
    # Exit with error code if any failures
    sys.exit(0 if result['errors'] == 0 else 1)


if __name__ == '__main__':
    main()
