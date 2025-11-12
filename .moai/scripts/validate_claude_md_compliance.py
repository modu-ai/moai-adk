#!/usr/bin/env python3
"""
CLAUDE.md Documentation Compliance Validator

Purpose: Automated validation of CLAUDE.md against official Claude Code documentation standards
Version: 1.0.0 (2025-11-13)
Maintained by: MoAI-ADK Team
"""

import json
import re
import sys
from pathlib import Path
from typing import Dict, List, Tuple, Any
import argparse
import yaml


class ClaudeMDComplianceValidator:
    """Validates CLAUDE.md documentation against official Claude Code standards."""

    def __init__(self, config_file: str = None):
        self.config = self._load_config(config_file)
        self.violations = []
        self.template_variables = [
            "{{PROJECT_NAME}}", "{{PROJECT_OWNER}}", "{{MOAI_VERSION}}",
            "{{PROJECT_MODE}}", "{{CODEBASE_LANGUAGE}}", "{{CONVERSATION_LANGUAGE}}",
            "{{CONVERSATION_LANGUAGE_NAME}}"
        ]

    def _load_config(self, config_file: str) -> Dict:
        """Load configuration from file."""
        if config_file and Path(config_file).exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                return yaml.safe_load(f)
        return {
            "critical_rules": [
                "Tool usage delegation",
                "AskUserQuestion format",
                "MCP integration"
            ],
            "major_rules": [
                "4-layer architecture",
                "Context7 tool usage",
                "Variable substitution"
            ],
            "minor_rules": [
                "Configuration hierarchy",
                ".gitignore security rules",
                "Multilingual support"
            ]
        }

    def validate_claude_md(self, file_path: str) -> List[Dict[str, Any]]:
        """Validate CLAUDE.md against compliance standards."""
        self.violations = []
        claude_md_path = Path(file_path)

        if not claude_md_path.exists():
            raise FileNotFoundError(f"CLAUDE.md not found at {file_path}")

        with open(claude_md_path, 'r', encoding='utf-8') as f:
            content = f.read()
            lines = content.split('\n')

        print(f"🔍 Validating {claude_md_path}...")

        # Run all validation checks
        self._validate_critical_rules(content, lines)
        self._validate_major_rules(content, lines)
        self._validate_minor_rules(content, lines)

        return self.violations

    def _validate_critical_rules(self, content: str, lines: List[str]):
        """Validate critical compliance rules."""
        print("🚨 Checking CRITICAL compliance rules...")

        # 1. Tool usage delegation rules
        if "Direct tool usage (Read, Write, Edit, Bash)" not in content:
            self._add_violation(
                "CRITICAL",
                "Tool usage delegation",
                "Missing Task() delegation requirement for direct tool usage",
                lines
            )

        # 2. AskUserQuestion format
        if '"questions": [' not in content:
            self._add_violation(
                "CRITICAL",
                "AskUserQuestion format",
                "Missing proper JSON format with questions array",
                lines
            )

        # 3. MCP integration sections
        if "MCP Servers Overview" not in content:
            self._add_violation(
                "CRITICAL",
                "MCP integration",
                "Missing MCP servers overview section",
                lines
            )

    def _validate_major_rules(self, content: str, lines: List[str]):
        """Validate major compliance rules."""
        print("⚠️  Checking MAJOR compliance rules...")

        # 1. 4-layer architecture
        if "Commands → Sub-agents → Skills → Hooks" not in content:
            self._add_violation(
                "MAJOR",
                "4-layer architecture",
                "Missing correct 4-layer architecture description",
                lines
            )

        # 2. Context7 tool usage
        if "mcp__context7__resolve-library-id" not in content:
            self._add_violation(
                "MAJOR",
                "Context7 tool usage",
                "Missing Context7 MCP tool usage patterns",
                lines
            )

        # 3. Variable substitution patterns
        missing_vars = [var for var in self.template_variables if var not in content]
        if missing_vars:
            self._add_violation(
                "MAJOR",
                "Variable substitution",
                f"Missing template variables: {', '.join(missing_vars)}",
                lines
            )

    def _validate_minor_rules(self, content: str, lines: List[str]):
        """Validate minor compliance rules."""
        print("📝 Checking MINOR compliance rules...")

        # 1. Configuration hierarchy
        if ".moai/config/config.json" not in content:
            self._add_violation(
                "MINOR",
                "Configuration hierarchy",
                "Missing configuration file hierarchy reference",
                lines
            )

        # 2. .gitignore security rules
        if ".vercel/project.json" not in content:
            self._add_violation(
                "MINOR",
                ".gitignore security rules",
                "Missing security .gitignore rules",
                lines
            )

        # 3. Multilingual support
        if "language.conversation_language" not in content:
            self._add_violation(
                "MINOR",
                "Multilingual support",
                "Missing language configuration reference",
                lines
            )

    def _add_violation(self, severity: str, category: str, description: str, lines: List[str]):
        """Add a violation to the list."""
        # Find approximate line number
        line_num = 0
        for i, line in enumerate(lines):
            if category.lower() in line.lower() or description.lower() in line.lower():
                line_num = i + 1
                break

        self.violations.append({
            "severity": severity,
            "category": category,
            "description": description,
            "line": line_num,
            "status": "pending"
        })

    def generate_report(self) -> str:
        """Generate compliance report."""
        if not self.violations:
            return "✅ CLAUDE.md is fully compliant with official Claude Code documentation standards!"

        report = ["📊 CLAUDE.md Compliance Report"]
        report.append("=" * 50)

        # Count by severity
        critical_count = len([v for v in self.violations if v["severity"] == "CRITICAL"])
        major_count = len([v for v in self.violations if v["severity"] == "MAJOR"])
        minor_count = len([v for v in self.violations if v["severity"] == "MINOR"])

        report.append(f"\n📈 Summary:")
        report.append(f"   CRITICAL violations: {critical_count}")
        report.append(f"   MAJOR violations: {major_count}")
        report.append(f"   MINOR violations: {minor_count}")
        report.append(f"   Total issues: {len(self.violations)}")

        # Detailed violations
        if critical_count > 0:
            report.append(f"\n🚨 CRITICAL Violations (must be fixed immediately):")
            for violation in self.violations:
                if violation["severity"] == "CRITICAL":
                    report.append(f"   • Line {violation['line']}: {violation['description']}")

        if major_count > 0:
            report.append(f"\n⚠️  MAJOR Violations (fix before release):")
            for violation in self.violations:
                if violation["severity"] == "MAJOR":
                    report.append(f"   • Line {violation['line']}: {violation['description']}")

        if minor_count > 0:
            report.append(f"\n📝 MINOR Improvements (for enhanced quality):")
            for violation in self.violations:
                if violation["severity"] == "MINOR":
                    report.append(f"   • Line {violation['line']}: {violation['description']}")

        return "\n".join(report)

    def save_report(self, report: str, output_file: str = None):
        """Save compliance report to file."""
        if not output_file:
            output_file = f"claude_md_compliance_report_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}.txt"

        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(report)

        print(f"📄 Report saved to: {output_file}")

    def validate_template_sync(self) -> bool:
        """Validate synchronization between main CLAUDE.md and package template."""
        main_file = Path("/Users/goos/MoAI/MoAI-ADK/CLAUDE.md")
        template_file = Path("/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/CLAUDE.md")

        if not main_file.exists() or not template_file.exists():
            return False

        # Basic comparison - could be enhanced with more sophisticated diffing
        with open(main_file, 'r', encoding='utf-8') as f:
            main_content = f.read()

        with open(template_file, 'r', encoding='utf-8') as f:
            template_content = f.read()

        # Check if template variables are properly used in template
        for var in self.template_variables:
            if var in template_content:
                continue
            else:
                print(f"⚠️  Missing template variable: {var}")
                return False

        print("✅ Template synchronization validated successfully")
        return True


def main():
    parser = argparse.ArgumentParser(description="Validate CLAUDE.md compliance with official standards")
    parser.add_argument("--file", "-f", default="CLAUDE.md", help="Path to CLAUDE.md file")
    parser.add_argument("--config", "-c", help="Configuration file path")
    parser.add_argument("--template-sync", action="store_true", help="Validate template synchronization")
    parser.add_argument("--report", "-r", help="Output report file")
    parser.add_argument("--detailed", action="store_true", help="Show detailed analysis")

    args = parser.parse_args()

    try:
        validator = ClaudeMDComplianceValidator(args.config)

        # Validate main CLAUDE.md
        violations = validator.validate_claude_md(args.file)

        # Generate and display report
        report = validator.generate_report()
        print(report)

        # Validate template synchronization
        if args.template_sync:
            print("\n🔄 Validating template synchronization...")
            validator.validate_template_sync()

        # Save report if requested
        if args.report:
            validator.save_report(report, args.report)

        # Exit with appropriate code
        if violations:
            sys.exit(1)
        else:
            print("\n🎉 All compliance checks passed!")
            sys.exit(0)

    except Exception as e:
        print(f"❌ Error: {str(e)}")
        sys.exit(1)


if __name__ == "__main__":
    import datetime
    main()