#!/usr/bin/env python3
"""
Phase 2 Quality Validation Script
Validates all 12 refactored skills against enterprise standards.
"""

import os
import re
import json
from pathlib import Path
from typing import Dict, List, Tuple
from dataclasses import dataclass, asdict

@dataclass
class SkillValidationResult:
    skill_name: str
    skill_md_lines: int = 0
    skill_md_target: int = 300
    skill_md_compliant: bool = False
    module_count: int = 0
    module_files: List[str] = None
    has_reference: bool = False
    has_examples: bool = False
    frontmatter_valid: bool = False
    frontmatter_errors: List[str] = None
    markdown_issues: List[str] = None
    link_issues: List[str] = None
    overall_status: str = "UNKNOWN"

    def __post_init__(self):
        if self.module_files is None:
            self.module_files = []
        if self.frontmatter_errors is None:
            self.frontmatter_errors = []
        if self.markdown_issues is None:
            self.markdown_issues = []
        if self.link_issues is None:
            self.link_issues = []

class SkillValidator:
    """Comprehensive skill validation."""

    REFACTORED_SKILLS = [
        "moai-project-documentation",
        "moai-baas-firebase-ext",
        "moai-docs-unified",
        "moai-domain-nano-banana",
        "moai-security-api",
        "moai-nextra-architecture",
        "moai-essentials-perf",
        "moai-cc-configuration",
        "moai-domain-security",
        "moai-foundation-git",
        "moai-core-personas",
        "moai-cc-memory"
    ]

    def __init__(self, skills_base_dir: str):
        self.skills_base_dir = Path(skills_base_dir)
        self.results: Dict[str, SkillValidationResult] = {}

    def validate_all_skills(self) -> Dict[str, SkillValidationResult]:
        """Validate all 12 refactored skills."""

        for skill_name in self.REFACTORED_SKILLS:
            skill_path = self.skills_base_dir / skill_name

            if not skill_path.exists():
                print(f"⚠️  Skill not found: {skill_name}")
                continue

            result = self.validate_skill(skill_path, skill_name)
            self.results[skill_name] = result

        return self.results

    def validate_skill(self, skill_path: Path, skill_name: str) -> SkillValidationResult:
        """Validate individual skill."""

        result = SkillValidationResult(skill_name=skill_name)

        # Check SKILL.md
        skill_md_path = skill_path / "SKILL.md"
        if skill_md_path.exists():
            result.skill_md_lines = self.count_lines(skill_md_path)
            result.skill_md_compliant = result.skill_md_lines <= result.skill_md_target

            # Validate frontmatter
            result.frontmatter_valid, result.frontmatter_errors = self.validate_frontmatter(skill_md_path)

            # Check markdown issues
            result.markdown_issues = self.check_markdown_issues(skill_md_path)

            # Check links
            result.link_issues = self.check_links(skill_md_path, skill_path)

        # Check modules
        modules_dir = skill_path / "modules"
        if modules_dir.exists():
            result.module_files = [f.name for f in modules_dir.glob("*.md")]
            result.module_count = len(result.module_files)

        # Check reference.md and examples.md
        result.has_reference = (skill_path / "reference.md").exists()
        result.has_examples = (skill_path / "examples.md").exists()

        # Overall status
        result.overall_status = self.calculate_status(result)

        return result

    def count_lines(self, file_path: Path) -> int:
        """Count non-empty lines in file."""
        with open(file_path, 'r', encoding='utf-8') as f:
            return sum(1 for line in f if line.strip())

    def validate_frontmatter(self, skill_md_path: Path) -> Tuple[bool, List[str]]:
        """Validate YAML frontmatter."""
        errors = []

        with open(skill_md_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Check for frontmatter
        if not content.startswith('---'):
            errors.append("Missing YAML frontmatter opening delimiter")
            return False, errors

        # Extract frontmatter
        parts = content.split('---', 2)
        if len(parts) < 3:
            errors.append("Incomplete YAML frontmatter (missing closing delimiter)")
            return False, errors

        frontmatter = parts[1].strip()

        # Check required fields
        required_fields = ['name', 'description']
        for field in required_fields:
            if not re.search(rf'^{field}:', frontmatter, re.MULTILINE):
                errors.append(f"Missing required field: {field}")

        # Check allowed-tools format (comma-separated, no brackets)
        if re.search(r'allowed-tools:\s*\[', frontmatter):
            errors.append("allowed-tools should use comma-separated format, not brackets")

        return len(errors) == 0, errors

    def check_markdown_issues(self, skill_md_path: Path) -> List[str]:
        """Check CommonMark compliance issues."""
        issues = []

        with open(skill_md_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Check for bold+parentheses anti-pattern
        # Pattern: **text(description)**next
        bad_pattern = re.compile(r'\*\*[^*]+\([^)]+\)\*\*[^\s*]')
        matches = bad_pattern.findall(content)
        if matches:
            issues.append(f"Found {len(matches)} bold+parentheses violations (CommonMark)")
            for match in matches[:3]:  # Show first 3 examples
                issues.append(f"  Example: {match[:50]}...")

        return issues

    def check_links(self, skill_md_path: Path, skill_path: Path) -> List[str]:
        """Check link validity."""
        issues = []

        with open(skill_md_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Remove code blocks to avoid false positives
        content_without_code = re.sub(r'```.*?```', '', content, flags=re.DOTALL)

        # Find all markdown links [text](path)
        link_pattern = re.compile(r'\[([^\]]+)\]\(([^)]+)\)')
        links = link_pattern.findall(content_without_code)

        for link_text, link_path in links:
            # Skip external links
            if link_path.startswith('http://') or link_path.startswith('https://'):
                continue

            # Skip anchor links
            if link_path.startswith('#'):
                continue

            # Check if file exists
            target_path = skill_path / link_path
            if not target_path.exists():
                issues.append(f"Broken link: [{link_text}]({link_path})")

        return issues

    def calculate_status(self, result: SkillValidationResult) -> str:
        """Calculate overall validation status."""

        critical_issues = []
        warnings = []

        # Critical checks
        if not result.skill_md_compliant:
            critical_issues.append(f"SKILL.md exceeds {result.skill_md_target} lines ({result.skill_md_lines} lines)")

        if not result.frontmatter_valid:
            critical_issues.append(f"Invalid frontmatter: {len(result.frontmatter_errors)} errors")

        if result.link_issues:
            critical_issues.append(f"{len(result.link_issues)} broken links")

        # Warnings
        if result.module_count < 4:
            warnings.append(f"Only {result.module_count} modules (expected 4-6)")

        if result.markdown_issues:
            warnings.append(f"{len(result.markdown_issues)} markdown issues")

        # Determine status
        if critical_issues:
            return f"❌ FAIL ({len(critical_issues)} critical)"
        elif warnings:
            return f"⚠️  WARNING ({len(warnings)} warnings)"
        else:
            return "✅ PASS"

    def generate_report(self) -> str:
        """Generate validation report."""

        report_lines = [
            "=" * 80,
            "PHASE 2 QUALITY VALIDATION REPORT",
            "=" * 80,
            "",
            f"Validated {len(self.results)} skills",
            ""
        ]

        # Summary table
        report_lines.append("SUMMARY")
        report_lines.append("-" * 80)
        report_lines.append(f"{'Skill':<35} {'Lines':<8} {'Modules':<8} {'Status'}")
        report_lines.append("-" * 80)

        pass_count = 0
        warn_count = 0
        fail_count = 0

        for skill_name, result in sorted(self.results.items()):
            status_icon = result.overall_status.split()[0]

            if "PASS" in result.overall_status:
                pass_count += 1
            elif "WARNING" in result.overall_status:
                warn_count += 1
            elif "FAIL" in result.overall_status:
                fail_count += 1

            report_lines.append(
                f"{skill_name:<35} {result.skill_md_lines:<8} {result.module_count:<8} {result.overall_status}"
            )

        report_lines.append("-" * 80)
        report_lines.append(f"✅ PASS: {pass_count}  |  ⚠️  WARNING: {warn_count}  |  ❌ FAIL: {fail_count}")
        report_lines.append("")

        # Detailed findings
        report_lines.append("")
        report_lines.append("DETAILED FINDINGS")
        report_lines.append("=" * 80)

        for skill_name, result in sorted(self.results.items()):
            if "FAIL" in result.overall_status or "WARNING" in result.overall_status:
                report_lines.append("")
                report_lines.append(f"{skill_name} ({result.overall_status})")
                report_lines.append("-" * 40)

                if not result.skill_md_compliant:
                    report_lines.append(f"  ❌ SKILL.md: {result.skill_md_lines} lines (target: ≤{result.skill_md_target})")

                if result.frontmatter_errors:
                    report_lines.append(f"  ❌ Frontmatter errors:")
                    for error in result.frontmatter_errors:
                        report_lines.append(f"     - {error}")

                if result.markdown_issues:
                    report_lines.append(f"  ⚠️  Markdown issues:")
                    for issue in result.markdown_issues[:5]:  # Show first 5
                        report_lines.append(f"     - {issue}")

                if result.link_issues:
                    report_lines.append(f"  ❌ Broken links:")
                    for issue in result.link_issues[:5]:  # Show first 5
                        report_lines.append(f"     - {issue}")

                if result.module_count < 4:
                    report_lines.append(f"  ⚠️  Only {result.module_count} modules (expected 4-6)")
                    report_lines.append(f"     Modules: {', '.join(result.module_files)}")

        return "\n".join(report_lines)

    def save_json_report(self, output_path: str):
        """Save detailed JSON report."""

        json_data = {
            skill_name: asdict(result)
            for skill_name, result in self.results.items()
        }

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(json_data, f, indent=2)

        print(f"✅ Detailed JSON report saved: {output_path}")

def main():
    """Main validation execution."""

    skills_base_dir = "/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills"

    print("🔍 Starting Phase 2 Quality Validation...")
    print()

    validator = SkillValidator(skills_base_dir)
    results = validator.validate_all_skills()

    # Generate and print report
    report = validator.generate_report()
    print(report)

    # Save JSON report
    json_output = "/Users/goos/MoAI/MoAI-ADK/.moai/logs/phase2-validation-report.json"
    validator.save_json_report(json_output)

    print()
    print("=" * 80)
    print("VALIDATION COMPLETE")
    print("=" * 80)

if __name__ == "__main__":
    main()
