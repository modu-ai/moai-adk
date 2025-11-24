#!/usr/bin/env python3
"""
Phase 2 Task 2: Module Optimization
Analyzes and optimizes module files across all 12 refactored skills.
"""

import os
import re
from pathlib import Path
from typing import Dict, List, Tuple
from dataclasses import dataclass
from collections import Counter

@dataclass
class ModuleAnalysis:
    skill_name: str
    module_name: str
    line_count: int
    has_code_examples: bool
    example_count: int
    has_headings: bool
    heading_count: int
    issues: List[str]
    recommendations: List[str]

class ModuleOptimizer:
    """Analyze and optimize skill modules."""

    REFACTORED_SKILLS = [
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
    ]

    def __init__(self, skills_base_dir: str):
        self.skills_base_dir = Path(skills_base_dir)
        self.analyses: Dict[str, List[ModuleAnalysis]] = {}

    def analyze_all_modules(self) -> Dict[str, List[ModuleAnalysis]]:
        """Analyze modules in all refactored skills."""

        total_modules = 0
        for skill_name in self.REFACTORED_SKILLS:
            skill_path = self.skills_base_dir / skill_name
            modules_dir = skill_path / "modules"

            if not modules_dir.exists():
                continue

            skill_analyses = []
            for module_file in modules_dir.glob("*.md"):
                analysis = self.analyze_module(skill_name, module_file)
                skill_analyses.append(analysis)
                total_modules += 1

            self.analyses[skill_name] = skill_analyses

        return self.analyses

    def analyze_module(self, skill_name: str, module_path: Path) -> ModuleAnalysis:
        """Analyze individual module file."""

        with open(module_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Count lines
        lines = content.split('\n')
        line_count = len([line for line in lines if line.strip()])

        # Check for code examples
        code_block_pattern = re.compile(r'```.*?```', re.DOTALL)
        code_blocks = code_block_pattern.findall(content)
        has_code_examples = len(code_blocks) > 0
        example_count = len(code_blocks)

        # Check for headings
        heading_pattern = re.compile(r'^#{1,6}\s+', re.MULTILINE)
        headings = heading_pattern.findall(content)
        has_headings = len(headings) > 0
        heading_count = len(headings)

        # Identify issues
        issues = []
        recommendations = []

        if line_count < 100:
            issues.append(f"Too short ({line_count} lines, target: 200-300)")
            recommendations.append("Expand content with more examples and explanations")

        if line_count > 500:
            issues.append(f"Too long ({line_count} lines, target: 200-300)")
            recommendations.append("Consider splitting into multiple modules")

        if not has_code_examples:
            issues.append("No code examples")
            recommendations.append("Add practical code examples")

        if example_count < 3:
            issues.append(f"Few code examples ({example_count}, target: 5+)")
            recommendations.append("Add more diverse examples")

        if heading_count < 3:
            issues.append(f"Few headings ({heading_count}, target: 5+)")
            recommendations.append("Add more section headings for structure")

        return ModuleAnalysis(
            skill_name=skill_name,
            module_name=module_path.name,
            line_count=line_count,
            has_code_examples=has_code_examples,
            example_count=example_count,
            has_headings=has_headings,
            heading_count=heading_count,
            issues=issues,
            recommendations=recommendations
        )

    def generate_optimization_report(self) -> str:
        """Generate comprehensive optimization report."""

        report_lines = [
            "=" * 80,
            "PHASE 2 TASK 2: MODULE OPTIMIZATION REPORT",
            "=" * 80,
            ""
        ]

        # Overall statistics
        total_modules = sum(len(analyses) for analyses in self.analyses.values())
        total_issues = sum(
            len(analysis.issues)
            for analyses in self.analyses.values()
            for analysis in analyses
        )

        report_lines.append(f"Total modules analyzed: {total_modules}")
        report_lines.append(f"Total issues found: {total_issues}")
        report_lines.append("")

        # Module size distribution
        report_lines.append("MODULE SIZE DISTRIBUTION")
        report_lines.append("-" * 80)

        size_categories = {
            "Too short (<100 lines)": 0,
            "Short (100-199 lines)": 0,
            "Optimal (200-300 lines)": 0,
            "Long (301-500 lines)": 0,
            "Too long (>500 lines)": 0
        }

        for analyses in self.analyses.values():
            for analysis in analyses:
                if analysis.line_count < 100:
                    size_categories["Too short (<100 lines)"] += 1
                elif analysis.line_count < 200:
                    size_categories["Short (100-199 lines)"] += 1
                elif analysis.line_count <= 300:
                    size_categories["Optimal (200-300 lines)"] += 1
                elif analysis.line_count <= 500:
                    size_categories["Long (301-500 lines)"] += 1
                else:
                    size_categories["Too long (>500 lines)"] += 1

        for category, count in size_categories.items():
            percentage = (count / total_modules * 100) if total_modules > 0 else 0
            report_lines.append(f"{category:<30} {count:>3} ({percentage:>5.1f}%)")

        report_lines.append("")

        # Code example analysis
        report_lines.append("CODE EXAMPLE COVERAGE")
        report_lines.append("-" * 80)

        modules_with_examples = sum(
            1 for analyses in self.analyses.values()
            for analysis in analyses
            if analysis.has_code_examples
        )
        example_coverage = (modules_with_examples / total_modules * 100) if total_modules > 0 else 0

        report_lines.append(f"Modules with code examples: {modules_with_examples}/{total_modules} ({example_coverage:.1f}%)")

        avg_examples = sum(
            analysis.example_count
            for analyses in self.analyses.values()
            for analysis in analyses
        ) / total_modules if total_modules > 0 else 0

        report_lines.append(f"Average code examples per module: {avg_examples:.1f}")
        report_lines.append("")

        # Per-skill breakdown
        report_lines.append("PER-SKILL MODULE ANALYSIS")
        report_lines.append("=" * 80)

        for skill_name in sorted(self.analyses.keys()):
            analyses = self.analyses[skill_name]

            if not analyses:
                continue

            report_lines.append(f"\n{skill_name}")
            report_lines.append("-" * 40)
            report_lines.append(f"{'Module':<35} {'Lines':<8} {'Examples':<10} {'Status'}")
            report_lines.append("-" * 80)

            for analysis in sorted(analyses, key=lambda a: a.module_name):
                status = "✅ Good" if not analysis.issues else f"⚠️  {len(analysis.issues)} issues"

                report_lines.append(
                    f"{analysis.module_name:<35} {analysis.line_count:<8} "
                    f"{analysis.example_count:<10} {status}"
                )

                # Show issues
                if analysis.issues:
                    for issue in analysis.issues:
                        report_lines.append(f"    ⚠️  {issue}")

        # Recommendations summary
        report_lines.append("")
        report_lines.append("TOP RECOMMENDATIONS")
        report_lines.append("=" * 80)

        all_recommendations = []
        for analyses in self.analyses.values():
            for analysis in analyses:
                all_recommendations.extend(analysis.recommendations)

        recommendation_counts = Counter(all_recommendations)
        for recommendation, count in recommendation_counts.most_common(10):
            report_lines.append(f"{count:>3}x  {recommendation}")

        return "\n".join(report_lines)

    def save_optimization_details(self, output_path: str):
        """Save detailed optimization data as JSON."""
        import json

        optimization_data = {}

        for skill_name, analyses in self.analyses.items():
            optimization_data[skill_name] = [
                {
                    "module_name": analysis.module_name,
                    "line_count": analysis.line_count,
                    "has_code_examples": analysis.has_code_examples,
                    "example_count": analysis.example_count,
                    "heading_count": analysis.heading_count,
                    "issues": analysis.issues,
                    "recommendations": analysis.recommendations
                }
                for analysis in analyses
            ]

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(optimization_data, f, indent=2)

        print(f"✅ Detailed optimization data saved: {output_path}")

def main():
    """Main optimization execution."""

    skills_base_dir = "/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills"

    print("🔍 Starting Phase 2 Task 2: Module Optimization...")
    print()

    optimizer = ModuleOptimizer(skills_base_dir)
    analyses = optimizer.analyze_all_modules()

    # Generate and print report
    report = optimizer.generate_optimization_report()
    print(report)

    # Save detailed JSON
    json_output = "/Users/goos/MoAI/MoAI-ADK/.moai/logs/phase2-task2-optimization-report.json"
    optimizer.save_optimization_details(json_output)

    print()
    print("=" * 80)
    print("OPTIMIZATION ANALYSIS COMPLETE")
    print("=" * 80)

if __name__ == "__main__":
    main()
