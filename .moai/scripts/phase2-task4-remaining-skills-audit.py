#!/usr/bin/env python3
"""
Phase 2 Task 4: Remaining Skills Audit
Identifies skills that still need Progressive Disclosure refactoring.
"""

import os
import re
from pathlib import Path
from typing import Dict, List, Tuple
from dataclasses import dataclass

@dataclass
class SkillAuditResult:
    skill_name: str
    skill_md_lines: int
    has_progressive_disclosure: bool
    module_count: int
    needs_refactoring: bool
    priority: str  # "HIGH", "MEDIUM", "LOW"
    reason: str

class RemainingSkillsAuditor:
    """Audit skills to identify refactoring candidates."""

    # Skills already refactored in Phase 1
    PHASE1_REFACTORED = {
        "moai-baas-firebase-ext",
        "moai-cc-configuration",
        "moai-cc-memory",
        "moai-core-personas",
        "moai-docs-unified",
        "moai-domain-nano-banana",
        "moai-domain-security",
        "moai-essentials-perf",
        "moai-foundation-git",
        "moai-nextra-architecture",
        "moai-project-documentation",
        "moai-security-api"
    }

    def __init__(self, skills_base_dir: str):
        self.skills_base_dir = Path(skills_base_dir)
        self.audit_results: List[SkillAuditResult] = []

    def audit_all_skills(self) -> List[SkillAuditResult]:
        """Audit all skills in directory."""

        for skill_dir in sorted(self.skills_base_dir.iterdir()):
            if not skill_dir.is_dir():
                continue

            # Skip already refactored
            if skill_dir.name in self.PHASE1_REFACTORED:
                continue

            skill_md_path = skill_dir / "SKILL.md"
            if not skill_md_path.exists():
                continue

            result = self.audit_skill(skill_dir)
            self.audit_results.append(result)

        return self.audit_results

    def audit_skill(self, skill_dir: Path) -> SkillAuditResult:
        """Audit individual skill."""

        skill_name = skill_dir.name
        skill_md_path = skill_dir / "SKILL.md"

        # Count lines
        with open(skill_md_path, 'r', encoding='utf-8') as f:
            content = f.read()

        lines = content.split('\n')
        line_count = len([line for line in lines if line.strip()])

        # Check for Progressive Disclosure structure
        has_progressive_disclosure = self.check_progressive_disclosure(content)

        # Count modules
        modules_dir = skill_dir / "modules"
        module_count = 0
        if modules_dir.exists():
            module_count = len(list(modules_dir.glob("*.md")))

        # Determine if needs refactoring
        needs_refactoring = False
        priority = "LOW"
        reason = ""

        if line_count > 500:
            needs_refactoring = True
            priority = "HIGH"
            reason = f"Exceeds 500 lines ({line_count} lines)"
        elif not has_progressive_disclosure:
            needs_refactoring = True
            priority = "MEDIUM"
            reason = "Missing Progressive Disclosure structure"
        elif module_count == 0:
            needs_refactoring = True
            priority = "MEDIUM"
            reason = "No module files (should have modular structure)"

        return SkillAuditResult(
            skill_name=skill_name,
            skill_md_lines=line_count,
            has_progressive_disclosure=has_progressive_disclosure,
            module_count=module_count,
            needs_refactoring=needs_refactoring,
            priority=priority,
            reason=reason
        )

    def check_progressive_disclosure(self, content: str) -> bool:
        """Check if skill has Progressive Disclosure structure."""

        # Look for level indicators
        level_patterns = [
            r'## Quick Reference',
            r'## Implementation Guide',
            r'## Advanced Patterns',
            r'## Level \d+',
            r'## Core Implementation',
        ]

        matches = sum(1 for pattern in level_patterns if re.search(pattern, content, re.IGNORECASE))

        # Consider it Progressive Disclosure if at least 2 level markers found
        return matches >= 2

    def generate_audit_report(self) -> str:
        """Generate audit report."""

        report_lines = [
            "=" * 80,
            "PHASE 2 TASK 4: REMAINING SKILLS AUDIT REPORT",
            "=" * 80,
            ""
        ]

        # Filter results
        high_priority = [r for r in self.audit_results if r.priority == "HIGH"]
        medium_priority = [r for r in self.audit_results if r.priority == "MEDIUM"]
        low_priority = [r for r in self.audit_results if r.priority == "LOW"]
        no_refactor_needed = [r for r in self.audit_results if not r.needs_refactoring]

        total_audited = len(self.audit_results)
        total_needs_refactor = len([r for r in self.audit_results if r.needs_refactoring])

        report_lines.append(f"Total skills audited: {total_audited}")
        report_lines.append(f"Phase 1 refactored: {len(self.PHASE1_REFACTORED)}")
        report_lines.append(f"Remaining skills needing refactoring: {total_needs_refactor}")
        report_lines.append("")

        # Priority distribution
        report_lines.append("PRIORITY DISTRIBUTION")
        report_lines.append("-" * 80)
        report_lines.append(f"HIGH Priority:   {len(high_priority):>3} skills (>500 lines)")
        report_lines.append(f"MEDIUM Priority: {len(medium_priority):>3} skills (structural issues)")
        report_lines.append(f"LOW Priority:    {len(low_priority):>3} skills (minor issues)")
        report_lines.append(f"No Refactor:     {len(no_refactor_needed):>3} skills (already good)")
        report_lines.append("")

        # High priority skills (>500 lines)
        if high_priority:
            report_lines.append("HIGH PRIORITY REFACTORING CANDIDATES (>500 lines)")
            report_lines.append("=" * 80)
            report_lines.append(f"{'Skill':<45} {'Lines':<8} {'Modules':<8} {'Reason'}")
            report_lines.append("-" * 80)

            for result in sorted(high_priority, key=lambda r: r.skill_md_lines, reverse=True):
                report_lines.append(
                    f"{result.skill_name:<45} {result.skill_md_lines:<8} {result.module_count:<8} {result.reason}"
                )
            report_lines.append("")

        # Medium priority skills
        if medium_priority:
            report_lines.append("MEDIUM PRIORITY REFACTORING CANDIDATES")
            report_lines.append("=" * 80)
            report_lines.append(f"{'Skill':<45} {'Lines':<8} {'Modules':<8} {'Reason'}")
            report_lines.append("-" * 80)

            for result in sorted(medium_priority, key=lambda r: r.skill_md_lines, reverse=True)[:20]:
                report_lines.append(
                    f"{result.skill_name:<45} {result.skill_md_lines:<8} {result.module_count:<8} {result.reason}"
                )
            report_lines.append("")

        # Phase 2 recommendations
        report_lines.append("PHASE 2 REFACTORING RECOMMENDATIONS")
        report_lines.append("=" * 80)

        phase2_batch1 = high_priority[:6]  # Top 6 high priority
        phase2_batch2 = high_priority[6:12] if len(high_priority) > 6 else []

        if phase2_batch1:
            report_lines.append("")
            report_lines.append("Batch 1 (Immediate - Top 6 largest):")
            for i, result in enumerate(phase2_batch1, 1):
                report_lines.append(f"  {i}. {result.skill_name} ({result.skill_md_lines} lines)")

        if phase2_batch2:
            report_lines.append("")
            report_lines.append("Batch 2 (Next iteration - Next 6):")
            for i, result in enumerate(phase2_batch2, 1):
                report_lines.append(f"  {i}. {result.skill_name} ({result.skill_md_lines} lines)")

        # Effort estimation
        report_lines.append("")
        report_lines.append("EFFORT ESTIMATION")
        report_lines.append("-" * 80)

        estimated_hours = 0
        for result in self.audit_results:
            if result.priority == "HIGH":
                estimated_hours += 2  # 2 hours per large skill
            elif result.priority == "MEDIUM":
                estimated_hours += 1  # 1 hour per medium skill

        report_lines.append(f"Estimated total effort: {estimated_hours} hours")
        report_lines.append(f"High priority effort: {len(high_priority) * 2} hours")
        report_lines.append(f"Medium priority effort: {len(medium_priority) * 1} hours")

        return "\n".join(report_lines)

    def save_audit_details(self, output_path: str):
        """Save detailed audit results as JSON."""
        import json

        audit_data = {
            "total_audited": len(self.audit_results),
            "phase1_refactored": len(self.PHASE1_REFACTORED),
            "needs_refactoring": len([r for r in self.audit_results if r.needs_refactoring]),
            "priority_breakdown": {
                "high": len([r for r in self.audit_results if r.priority == "HIGH"]),
                "medium": len([r for r in self.audit_results if r.priority == "MEDIUM"]),
                "low": len([r for r in self.audit_results if r.priority == "LOW"])
            },
            "skills": [
                {
                    "skill_name": result.skill_name,
                    "skill_md_lines": result.skill_md_lines,
                    "has_progressive_disclosure": result.has_progressive_disclosure,
                    "module_count": result.module_count,
                    "needs_refactoring": result.needs_refactoring,
                    "priority": result.priority,
                    "reason": result.reason
                }
                for result in sorted(self.audit_results, key=lambda r: (r.priority == "HIGH", r.skill_md_lines), reverse=True)
            ]
        }

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(audit_data, f, indent=2)

        print(f"✅ Detailed audit data saved: {output_path}")

def main():
    """Main audit execution."""

    skills_base_dir = "/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills"

    print("🔍 Starting Phase 2 Task 4: Remaining Skills Audit...")
    print()

    auditor = RemainingSkillsAuditor(skills_base_dir)
    results = auditor.audit_all_skills()

    # Generate and print report
    report = auditor.generate_audit_report()
    print(report)

    # Save detailed JSON
    json_output = "/Users/goos/MoAI/MoAI-ADK/.moai/logs/phase2-task4-audit-report.json"
    auditor.save_audit_details(json_output)

    print()
    print("=" * 80)
    print("REMAINING SKILLS AUDIT COMPLETE")
    print("=" * 80)

if __name__ == "__main__":
    main()
