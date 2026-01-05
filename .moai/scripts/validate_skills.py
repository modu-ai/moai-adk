#!/usr/bin/env python3
"""
MoAI-ADK Skill Quality Validator
================================
Validates skill structure following Claude Code official best practices:
- SKILL.md under 500 lines (warning at 400+)
- Module files 200-500 lines each
- Proper frontmatter structure
- Context7 MCP tools in allowed-tools
- Progressive Disclosure compliance
"""

import os
import sys
import yaml
import argparse
from pathlib import Path
from dataclasses import dataclass
from typing import Optional


@dataclass
class ValidationResult:
    """Validation result for a single skill."""
    skill_name: str
    skill_path: Path
    skill_md_lines: int
    has_modules: bool
    module_count: int
    has_reference: bool
    has_examples: bool
    has_context7: bool
    has_allowed_tools: bool
    errors: list[str]
    warnings: list[str]
    
    @property
    def is_valid(self) -> bool:
        return len(self.errors) == 0
    
    @property
    def status(self) -> str:
        if self.errors:
            return "❌ FAIL"
        elif self.warnings:
            return "⚠️  WARN"
        return "✅ PASS"


class SkillValidator:
    """Validates MoAI-ADK skills against best practices."""
    
    # Thresholds
    SKILL_MD_ERROR_THRESHOLD = 500
    SKILL_MD_WARN_THRESHOLD = 400
    SKILL_MD_IDEAL_MAX = 300
    MODULE_MIN_LINES = 100
    MODULE_MAX_LINES = 500
    
    # Required Context7 tools
    CONTEXT7_TOOLS = [
        "mcp__context7__resolve-library-id",
        "mcp__context7__get-library-docs"
    ]
    
    def __init__(self, skills_path: Path):
        self.skills_path = skills_path
        self.results: list[ValidationResult] = []
    
    def validate_all(self) -> list[ValidationResult]:
        """Validate all skills in the directory."""
        self.results = []
        
        if not self.skills_path.exists():
            print(f"Error: Skills path not found: {self.skills_path}")
            return self.results
        
        for skill_dir in sorted(self.skills_path.iterdir()):
            if skill_dir.is_dir() and skill_dir.name.startswith("moai-"):
                result = self.validate_skill(skill_dir)
                self.results.append(result)
        
        return self.results
    
    def validate_skill(self, skill_path: Path) -> ValidationResult:
        """Validate a single skill directory."""
        errors = []
        warnings = []
        
        skill_md = skill_path / "SKILL.md"
        modules_dir = skill_path / "modules"
        reference_md = skill_path / "reference.md"
        examples_md = skill_path / "examples.md"
        
        # Check SKILL.md exists
        skill_md_lines = 0
        has_context7 = False
        has_allowed_tools = False
        
        if skill_md.exists():
            content = skill_md.read_text()
            lines = content.split("\n")
            skill_md_lines = len(lines)
            
            # Check line count
            if skill_md_lines > self.SKILL_MD_ERROR_THRESHOLD:
                errors.append(f"SKILL.md exceeds {self.SKILL_MD_ERROR_THRESHOLD} lines ({skill_md_lines})")
            elif skill_md_lines > self.SKILL_MD_WARN_THRESHOLD:
                warnings.append(f"SKILL.md approaching limit ({skill_md_lines} lines)")
            
            # Parse frontmatter
            if content.startswith("---"):
                try:
                    end_idx = content.find("---", 3)
                    if end_idx > 0:
                        frontmatter = yaml.safe_load(content[3:end_idx])
                        
                        # Check allowed-tools
                        allowed_tools = frontmatter.get("allowed-tools", "")
                        if allowed_tools:
                            has_allowed_tools = True
                            for tool in self.CONTEXT7_TOOLS:
                                if tool in allowed_tools:
                                    has_context7 = True
                                    break
                        
                        # Check context7-libraries
                        if frontmatter.get("context7-libraries"):
                            has_context7 = True
                except yaml.YAMLError:
                    warnings.append("Failed to parse SKILL.md frontmatter")
            
            # Check Context7 integration
            if not has_context7:
                warnings.append("Missing Context7 MCP integration")
        else:
            errors.append("SKILL.md not found")
        
        # Check modules directory
        has_modules = modules_dir.exists() and modules_dir.is_dir()
        module_count = 0
        
        if has_modules:
            for module_file in modules_dir.glob("*.md"):
                module_count += 1
                module_lines = len(module_file.read_text().split("\n"))
                
                if module_lines < self.MODULE_MIN_LINES:
                    warnings.append(f"Module {module_file.name} too short ({module_lines} lines)")
                elif module_lines > self.MODULE_MAX_LINES:
                    warnings.append(f"Module {module_file.name} too long ({module_lines} lines)")
        
        # Check Standard structure compliance
        has_reference = reference_md.exists()
        has_examples = examples_md.exists()
        
        # Determine structure type and validate
        if skill_md_lines > self.SKILL_MD_IDEAL_MAX and not has_modules:
            warnings.append(f"Consider modularizing: {skill_md_lines} lines without modules/")
        
        return ValidationResult(
            skill_name=skill_path.name,
            skill_path=skill_path,
            skill_md_lines=skill_md_lines,
            has_modules=has_modules,
            module_count=module_count,
            has_reference=has_reference,
            has_examples=has_examples,
            has_context7=has_context7,
            has_allowed_tools=has_allowed_tools,
            errors=errors,
            warnings=warnings
        )
    
    def print_report(self, verbose: bool = False):
        """Print validation report."""
        print("\n" + "=" * 70)
        print("MoAI-ADK Skill Quality Report")
        print("=" * 70 + "\n")
        
        total = len(self.results)
        passed = sum(1 for r in self.results if r.is_valid and not r.warnings)
        warned = sum(1 for r in self.results if r.is_valid and r.warnings)
        failed = sum(1 for r in self.results if not r.is_valid)
        
        # Summary table
        print(f"{'Skill Name':<40} {'Lines':>6} {'Modules':>8} {'C7':>4} {'Status':<10}")
        print("-" * 70)
        
        for result in self.results:
            c7 = "✓" if result.has_context7 else "✗"
            modules = f"{result.module_count}" if result.has_modules else "-"
            print(f"{result.skill_name:<40} {result.skill_md_lines:>6} {modules:>8} {c7:>4} {result.status:<10}")
            
            if verbose and (result.errors or result.warnings):
                for err in result.errors:
                    print(f"    ❌ {err}")
                for warn in result.warnings:
                    print(f"    ⚠️  {warn}")
        
        print("-" * 70)
        print(f"\nSummary: {passed} passed, {warned} warnings, {failed} failed (Total: {total})")
        
        # Statistics
        lines_list = [r.skill_md_lines for r in self.results]
        if lines_list:
            avg_lines = sum(lines_list) / len(lines_list)
            max_lines = max(lines_list)
            min_lines = min(lines_list)
            with_modules = sum(1 for r in self.results if r.has_modules)
            with_context7 = sum(1 for r in self.results if r.has_context7)
            
            print(f"\nStatistics:")
            print(f"  SKILL.md lines: avg={avg_lines:.0f}, min={min_lines}, max={max_lines}")
            print(f"  With modules/: {with_modules}/{total} ({100*with_modules/total:.0f}%)")
            print(f"  With Context7: {with_context7}/{total} ({100*with_context7/total:.0f}%)")
        
        return failed == 0


def main():
    parser = argparse.ArgumentParser(description="Validate MoAI-ADK skills")
    parser.add_argument(
        "--path", 
        default="/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills",
        help="Path to skills directory"
    )
    parser.add_argument(
        "-v", "--verbose", 
        action="store_true",
        help="Show detailed errors and warnings"
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Output as JSON"
    )
    
    args = parser.parse_args()
    
    validator = SkillValidator(Path(args.path))
    validator.validate_all()
    
    if args.json:
        import json
        output = {
            "results": [
                {
                    "name": r.skill_name,
                    "lines": r.skill_md_lines,
                    "has_modules": r.has_modules,
                    "module_count": r.module_count,
                    "has_context7": r.has_context7,
                    "valid": r.is_valid,
                    "errors": r.errors,
                    "warnings": r.warnings
                }
                for r in validator.results
            ]
        }
        print(json.dumps(output, indent=2))
    else:
        success = validator.print_report(verbose=args.verbose)
        sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
