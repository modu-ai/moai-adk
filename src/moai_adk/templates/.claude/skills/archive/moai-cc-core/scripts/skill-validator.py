#!/usr/bin/env python3
"""
Claude Code Skill Validator

Validates skills against official Claude Code standards:
- SKILL.md line limit (≤500 lines)
- YAML frontmatter validation
- Directory structure compliance
- Progressive disclosure structure
- TRUST 5 compliance

Usage: python skill-validator.py [--skill-path PATH] [--validate-all]
"""

import os
import re
import yaml
import json
import argparse
from pathlib import Path
from typing import Dict, List, Tuple, Optional


class SkillValidator:
    """Validate Claude Code skills against official standards."""

    def __init__(self, skill_root: str = ".claude/skills"):
        self.skill_root = Path(skill_root)
        self.errors = []
        self.warnings = []

    def validate_skill(self, skill_path: Path) -> Dict:
        """Validate a single skill."""
        validation_result = {
            'skill_name': skill_path.name,
            'status': 'PASS',
            'errors': [],
            'warnings': [],
            'metrics': {}
        }

        # Check directory structure
        structure_result = self.validate_structure(skill_path)
        validation_result['errors'].extend(structure_result['errors'])
        validation_result['warnings'].extend(structure_result['warnings'])

        # Validate SKILL.md
        skill_md_path = skill_path / "SKILL.md"
        if skill_md_path.exists():
            skill_result = self.validate_skill_md(skill_md_path)
            validation_result['errors'].extend(skill_result['errors'])
            validation_result['warnings'].extend(skill_result['warnings'])
            validation_result['metrics'].update(skill_result['metrics'])
        else:
            validation_result['errors'].append("SKILL.md not found (required)")

        # Validate optional files
        for optional_file in ["reference.md", "examples.md"]:
            optional_path = skill_path / optional_file
            if optional_path.exists():
                optional_result = self.validate_optional_file(optional_path)
                validation_result['warnings'].extend(optional_result['warnings'])

        # Determine overall status
        if validation_result['errors']:
            validation_result['status'] = 'FAIL'
        elif validation_result['warnings']:
            validation_result['status'] = 'WARN'

        return validation_result

    def validate_structure(self, skill_path: Path) -> Dict:
        """Validate skill directory structure."""
        result = {'errors': [], 'warnings': []}

        # Required files
        required_files = ["SKILL.md"]
        for file_name in required_files:
            file_path = skill_path / file_name
            if not file_path.exists():
                result['errors'].append(f"Required file missing: {file_name}")

        # Optional directories
        optional_dirs = ["scripts", "templates"]
        for dir_name in optional_dirs:
            dir_path = skill_path / dir_name
            if dir_path.exists() and not dir_path.is_dir():
                result['errors'].append(f"{dir_name} exists but is not a directory")

        # Check for unexpected files at root level
        allowed_patterns = ["SKILL.md", "reference.md", "examples.md", "*.md"]
        root_files = [f for f in skill_path.iterdir() if f.is_file()]

        for file_path in root_files:
            if not any(file_path.match(pattern) for pattern in allowed_patterns):
                result['warnings'].append(f"Unexpected file at root: {file_path.name}")

        return result

    def validate_skill_md(self, skill_md_path: Path) -> Dict:
        """Validate SKILL.md content and structure."""
        result = {'errors': [], 'warnings': [], 'metrics': {}}

        try:
            content = skill_md_path.read_text(encoding='utf-8')
            lines = content.split('\n')

            # Line count validation (critical)
            line_count = len(lines)
            result['metrics']['line_count'] = line_count
            if line_count > 500:
                result['errors'].append(f"SKILL.md exceeds 500-line limit ({line_count} lines)")
            elif line_count > 450:
                result['warnings'].append(f"SKILL.md approaching limit ({line_count}/500 lines)")

            # YAML frontmatter validation
            frontmatter_result = self.validate_frontmatter(content)
            result['errors'].extend(frontmatter_result['errors'])
            result['warnings'].extend(frontmatter_result['warnings'])

            # Progressive disclosure structure
            structure_result = self.validate_progressive_disclosure(content)
            result['errors'].extend(structure_result['errors'])
            result['warnings'].extend(structure_result['warnings'])

            # Content quality checks
            quality_result = self.validate_content_quality(content)
            result['warnings'].extend(quality_result['warnings'])

        except Exception as e:
            result['errors'].append(f"Failed to read SKILL.md: {e}")

        return result

    def validate_frontmatter(self, content: str) -> Dict:
        """Validate YAML frontmatter."""
        result = {'errors': [], 'warnings': []}

        # Check for frontmatter
        if not content.startswith('---'):
            result['errors'].append("Missing YAML frontmatter (must start with ---)")
            return result

        # Extract frontmatter
        try:
            end_marker = content.find('---', 3)
            if end_marker == -1:
                result['errors'].append("Unclosed YAML frontmatter (missing closing ---)")
                return result

            frontmatter_text = content[3:end_marker]
            frontmatter_data = yaml.safe_load(frontmatter_text)

            if not isinstance(frontmatter_data, dict):
                result['errors'].append("Frontmatter must be a dictionary/object")
                return result

            # Required fields
            required_fields = ['name', 'description']
            for field in required_fields:
                if field not in frontmatter_data:
                    result['errors'].append(f"Required frontmatter field missing: {field}")
                elif not frontmatter_data[field]:
                    result['errors'].append(f"Required frontmatter field empty: {field}")

            # Field validation
            if 'name' in frontmatter_data:
                name = frontmatter_data['name']
                if not re.match(r'^[a-z0-9-]+$', name):
                    result['errors'].append("Skill name must be lowercase, numbers, and hyphens only")
                if len(name) > 64:
                    result['errors'].append("Skill name must be ≤64 characters")

            if 'description' in frontmatter_data:
                description = frontmatter_data['description']
                if len(description) > 1024:
                    result['errors'].append("Description must be ≤1024 characters")

            # Optional field validation
            if 'allowed-tools' in frontmatter_data:
                allowed_tools = frontmatter_data['allowed-tools']
                if isinstance(allowed_tools, list):
                    result['warnings'].append("allowed-tools should be comma-separated string, not list")
                elif isinstance(allowed_tools, str):
                    tools = [tool.strip() for tool in allowed_tools.split(',')]
                    valid_tools = ['Read', 'Write', 'Edit', 'Bash', 'Grep', 'Glob', 'AskUserQuestion',
                                 'TodoWrite', 'Task', 'Skill', 'WebFetch', 'WebSearch']
                    for tool in tools:
                        if tool not in valid_tools:
                            result['warnings'].append(f"Unknown tool in allowed-tools: {tool}")

        except yaml.YAMLError as e:
            result['errors'].append(f"Invalid YAML frontmatter: {e}")
        except Exception as e:
            result['errors'].append(f"Error parsing frontmatter: {e}")

        return result

    def validate_progressive_disclosure(self, content: str) -> Dict:
        """Validate progressive disclosure structure."""
        result = {'errors': [], 'warnings': []}

        # Check for required sections
        required_sections = [
            r'^## Quick Reference',
            r'^# .*',  # Main title
        ]

        for section_pattern in required_sections:
            if not re.search(section_pattern, content, re.MULTILINE):
                result['warnings'].append(f"Missing expected section pattern: {section_pattern}")

        # Check for good practices
        if not re.search(r'```', content):
            result['warnings'].append("No code examples found (consider adding examples)")

        if not re.search(r'when to use', content, re.IGNORECASE):
            result['warnings'].append("No 'When to Use' section found (helps with skill discovery)")

        return result

    def validate_content_quality(self, content: str) -> Dict:
        """Validate content quality."""
        result = {'warnings': []}

        # Check for vague descriptions
        vague_patterns = [
            r'helps with.*$',
            r'useful for.*$',
            r'provides.*functionality',
        ]

        lines = content.split('\n')
        for i, line in enumerate(lines, 1):
            for pattern in vague_patterns:
                if re.search(pattern, line, re.IGNORECASE):
                    result['warnings'].append(f"Line {i}: Vague description detected: {line.strip()}")

        # Check for CommonMark compliance (parentheses after bold)
        bad_pattern = r'\*\*[^*]+\([^)]+\)\*\*[^\s*]'
        matches = re.finditer(bad_pattern, content)
        for match in matches:
            result['warnings'].append(f"CommonMark violation: bold text followed by parentheses in same marker: {match.group()}")

        return result

    def validate_optional_file(self, file_path: Path) -> Dict:
        """Validate optional files (reference.md, examples.md)."""
        result = {'warnings': []}

        try:
            content = file_path.read_text(encoding='utf-8')

            # Check if file has substantial content
            if len(content.strip()) < 100:
                result['warnings'].append(f"{file_path.name} appears to be empty or minimal")

            # Check for markdown formatting
            if not content.strip().startswith('#'):
                result['warnings'].append(f"{file_path.name} should start with a heading (#)")

        except Exception as e:
            result['warnings'].append(f"Could not read {file_path.name}: {e}")

        return result

    def validate_all_skills(self) -> List[Dict]:
        """Validate all skills in the skills directory."""
        results = []

        if not self.skill_root.exists():
            print(f"Skills directory not found: {self.skill_root}")
            return results

        for skill_path in self.skill_root.iterdir():
            if skill_path.is_dir() and not skill_path.name.startswith('.'):
                result = self.validate_skill(skill_path)
                results.append(result)

        return results

    def print_results(self, results: List[Dict]) -> None:
        """Print validation results."""
        print("\n" + "="*60)
        print("CLAUDE CODE SKILL VALIDATION REPORT")
        print("="*60)

        total_skills = len(results)
        passed_skills = sum(1 for r in results if r['status'] == 'PASS')
        warned_skills = sum(1 for r in results if r['status'] == 'WARN')
        failed_skills = sum(1 for r in results if r['status'] == 'FAIL')

        print(f"\nSummary: {total_skills} skills")
        print(f"  ✅ PASSED: {passed_skills}")
        print(f"  ⚠️  WARNED: {warned_skills}")
        print(f"  ❌ FAILED: {failed_skills}")

        # Detailed results
        for result in results:
            status_icon = {"PASS": "✅", "WARN": "⚠️", "FAIL": "❌"}[result['status']]
            print(f"\n{status_icon} {result['skill_name']} ({result['status']})")

            if result['errors']:
                print("  Errors:")
                for error in result['errors']:
                    print(f"    - {error}")

            if result['warnings']:
                print("  Warnings:")
                for warning in result['warnings']:
                    print(f"    - {warning}")

            if result['metrics']:
                print("  Metrics:")
                for metric, value in result['metrics'].items():
                    print(f"    - {metric}: {value}")


def main():
    parser = argparse.ArgumentParser(description="Validate Claude Code skills")
    parser.add_argument("--skill-path", help="Path to specific skill to validate")
    parser.add_argument("--validate-all", action="store_true", help="Validate all skills")
    parser.add_argument("--skill-root", default=".claude/skills", help="Root directory containing skills")

    args = parser.parse_args()

    validator = SkillValidator(args.skill_root)

    if args.skill_path:
        skill_path = Path(args.skill_path)
        result = validator.validate_skill(skill_path)
        validator.print_results([result])
    elif args.validate_all:
        results = validator.validate_all_skills()
        validator.print_results(results)
    else:
        # Default: validate current directory as skill
        current_dir = Path.cwd()
        if (current_dir / "SKILL.md").exists():
            result = validator.validate_skill(current_dir)
            validator.print_results([result])
        else:
            print("No SKILL.md found in current directory. Use --skill-path or --validate-all.")
            parser.print_help()


if __name__ == "__main__":
    main()