#!/usr/bin/env python3
"""
v4.0 Enterprise Compliance Validator

Checks all v4.0 requirements and generates compliance report.

Usage:
    python3 scripts/validate-v4-compliance.py moai-core-agent-guide
    python3 scripts/validate-v4-compliance.py --all
    python3 scripts/validate-v4-compliance.py --all --report reports/validation.txt
"""

import re
import sys
from pathlib import Path
from typing import Dict, Tuple


class V4Validator:
    """Validate v4.0 Enterprise compliance"""
    
    REQUIRED_CHECKS = [
        "version_4_0",
        "has_primary_agent",
        "has_keywords",
        "has_tier",
        "progressive_disclosure_level1",
        "progressive_disclosure_level2",
        "min_10_examples",
        "context7_section",
        "best_practices",
        "official_references"
    ]
    
    OPTIONAL_CHECKS = [
        "progressive_disclosure_level3",
        "decision_tree",
        "related_skills",
        "has_orchestration",
        "has_secondary_agents",
        "version_history"
    ]
    
    def validate_skill(self, skill_path: str) -> Dict[str, any]:
        """Validate single skill"""
        skill_md = Path(skill_path) / "SKILL.md"
        
        if not skill_md.exists():
            return {"error": "SKILL.md not found"}
        
        content = skill_md.read_text(encoding='utf-8')
        
        checks = {
            # Required
            "version_4_0": self._check_version(content),
            "has_primary_agent": self._check_primary_agent(content),
            "has_keywords": self._check_keywords(content),
            "has_tier": self._check_tier(content),
            "progressive_disclosure_level1": self._check_level1(content),
            "progressive_disclosure_level2": self._check_level2(content),
            "min_10_examples": self._check_examples(content),
            "context7_section": self._check_context7(content),
            "best_practices": self._check_best_practices(content),
            "official_references": self._check_references(content),
            
            # Optional
            "progressive_disclosure_level3": self._check_level3(content),
            "decision_tree": self._check_decision_tree(content),
            "related_skills": self._check_related_skills(content),
            "has_orchestration": self._check_orchestration(content),
            "has_secondary_agents": self._check_secondary_agents(content),
            "version_history": self._check_version_history(content)
        }
        
        # Calculate scores
        required_pass = sum(1 for check in self.REQUIRED_CHECKS if checks[check])
        optional_pass = sum(1 for check in self.OPTIONAL_CHECKS if checks[check])
        
        checks["required_score"] = f"{required_pass}/{len(self.REQUIRED_CHECKS)}"
        checks["optional_score"] = f"{optional_pass}/{len(self.OPTIONAL_CHECKS)}"
        checks["all_required_pass"] = (required_pass == len(self.REQUIRED_CHECKS))
        
        # Count examples
        code_blocks = len(re.findall(r'```', content))
        checks["example_count"] = code_blocks // 2
        
        # File size
        checks["size_kb"] = len(content) / 1024
        
        return checks
    
    def _check_version(self, content: str) -> bool:
        """Check version is 4.0.0"""
        return bool(re.search(r'^version:\s*["\']?4\.0\.0["\']?', content, re.MULTILINE))
    
    def _check_primary_agent(self, content: str) -> bool:
        """Check primary-agent is defined"""
        return bool(re.search(r'^primary-agent:\s*["\']?\w+', content, re.MULTILINE))
    
    def _check_keywords(self, content: str) -> Tuple[bool, int]:
        """Check keywords field exists with 3+ keywords"""
        match = re.search(r'^keywords:\s*\[(.*?)\]', content, re.MULTILINE)
        if not match:
            return False
        keywords = [k.strip(' "\'') for k in match.group(1).split(',')]
        return len([k for k in keywords if k]) >= 3
    
    def _check_tier(self, content: str) -> bool:
        """Check tier is defined"""
        return bool(re.search(r'^tier:\s*\w+', content, re.MULTILINE))
    
    def _check_level1(self, content: str) -> bool:
        """Check Level 1 exists"""
        return bool(re.search(r'###\s*Level 1[:\s]', content))
    
    def _check_level2(self, content: str) -> bool:
        """Check Level 2 exists"""
        return bool(re.search(r'###\s*Level 2[:\s]', content))
    
    def _check_level3(self, content: str) -> bool:
        """Check Level 3 exists (optional)"""
        return bool(re.search(r'###\s*Level 3[:\s]', content))
    
    def _check_examples(self, content: str) -> bool:
        """Check at least 10 code examples"""
        code_blocks = len(re.findall(r'```', content))
        examples = code_blocks // 2
        return examples >= 10
    
    def _check_context7(self, content: str) -> bool:
        """Check Context7 integration section"""
        return bool(re.search(r'Context7.*Integration|MCP.*Integration', content, re.IGNORECASE))
    
    def _check_best_practices(self, content: str) -> bool:
        """Check best practices checklist"""
        return bool(re.search(r'Best Practices|Best Practice', content, re.IGNORECASE))
    
    def _check_decision_tree(self, content: str) -> bool:
        """Check decision tree"""
        return bool(re.search(r'Decision Tree', content, re.IGNORECASE))
    
    def _check_related_skills(self, content: str) -> bool:
        """Check related skills section"""
        return bool(re.search(r'Related Skills|Integration with Other Skills', content, re.IGNORECASE))
    
    def _check_references(self, content: str) -> bool:
        """Check official references"""
        return bool(re.search(r'Official References|References', content, re.IGNORECASE))
    
    def _check_orchestration(self, content: str) -> bool:
        """Check orchestration metadata"""
        return bool(re.search(r'^orchestration:', content, re.MULTILINE))
    
    def _check_secondary_agents(self, content: str) -> bool:
        """Check secondary-agents is defined"""
        return bool(re.search(r'^secondary-agents:', content, re.MULTILINE))
    
    def _check_version_history(self, content: str) -> bool:
        """Check version history section"""
        return bool(re.search(r'Version History|Changelog', content, re.IGNORECASE))
    
    def validate_all(self, skills_dir: str = ".claude/skills") -> Dict[str, Dict]:
        """Validate all skills"""
        results = {}
        skills_path = Path(skills_dir)
        
        for skill_dir in sorted(skills_path.iterdir()):
            if not skill_dir.is_dir() or skill_dir.name.startswith('.'):
                continue
            
            skill_md = skill_dir / "SKILL.md"
            if not skill_md.exists():
                continue
            
            try:
                results[skill_dir.name] = self.validate_skill(str(skill_dir))
            except Exception as e:
                results[skill_dir.name] = {
                    "error": str(e),
                    "all_required_pass": False
                }
        
        return results
    
    def generate_report(self, results: Dict[str, Dict]) -> str:
        """Generate validation report"""
        lines = []
        lines.append("=" * 80)
        lines.append("v4.0 Enterprise Compliance Report")
        lines.append("=" * 80)
        lines.append("")
        
        passed = []
        failed = []
        
        for skill, checks in results.items():
            if checks.get("all_required_pass", False):
                passed.append((skill, checks))
            else:
                failed.append((skill, checks))
        
        total = len(passed) + len(failed)
        pass_rate = (len(passed) / total * 100) if total > 0 else 0
        
        lines.append(f"Total Skills: {total}")
        lines.append(f"✅ Passed: {len(passed)} ({pass_rate:.1f}%)")
        lines.append(f"❌ Failed: {len(failed)} ({100-pass_rate:.1f}%)")
        lines.append("")
        
        if passed:
            lines.append("=" * 80)
            lines.append("✅ PASSED SKILLS")
            lines.append("=" * 80)
            
            for skill, checks in sorted(passed):
                examples = checks.get("example_count", 0)
                size = checks.get("size_kb", 0)
                lines.append(f"\n✅ {skill}")
                lines.append(f"   Examples: {examples}, Size: {size:.1f}KB")
                lines.append(f"   Required: {checks.get('required_score', 'N/A')}")
                lines.append(f"   Optional: {checks.get('optional_score', 'N/A')}")
        
        if failed:
            lines.append("")
            lines.append("=" * 80)
            lines.append("❌ FAILED SKILLS")
            lines.append("=" * 80)
            
            for skill, checks in sorted(failed):
                if "error" in checks:
                    lines.append(f"\n❌ {skill}: {checks['error']}")
                    continue
                
                examples = checks.get("example_count", 0)
                size = checks.get("size_kb", 0)
                lines.append(f"\n❌ {skill}")
                lines.append(f"   Examples: {examples}, Size: {size:.1f}KB")
                lines.append(f"   Required: {checks.get('required_score', 'N/A')}")
                lines.append(f"   Optional: {checks.get('optional_score', 'N/A')}")
                lines.append(f"   Issues:")
                
                for check in self.REQUIRED_CHECKS:
                    if not checks.get(check, False):
                        check_name = check.replace('_', ' ').title()
                        lines.append(f"      🔴 REQUIRED: {check_name}")
                
                for check in self.OPTIONAL_CHECKS:
                    if not checks.get(check, False):
                        check_name = check.replace('_', ' ').title()
                        lines.append(f"      ⚠️  optional: {check_name}")
        
        lines.append("")
        lines.append("=" * 80)
        
        return "\n".join(lines)


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Validate v4.0 Enterprise compliance")
    parser.add_argument("skill", nargs="?", help="Skill name to validate")
    parser.add_argument("--all", action="store_true", help="Validate all skills")
    parser.add_argument("--skills-dir", default=".claude/skills", help="Skills directory")
    parser.add_argument("--report", help="Save report to file")
    
    args = parser.parse_args()
    
    validator = V4Validator()
    
    if args.all:
        results = validator.validate_all(args.skills_dir)
        report = validator.generate_report(results)
        print(report)
        
        if args.report:
            Path(args.report).parent.mkdir(parents=True, exist_ok=True)
            Path(args.report).write_text(report)
            print(f"\n📄 Report saved to {args.report}")
    
    elif args.skill:
        skill_path = Path(args.skills_dir) / args.skill
        checks = validator.validate_skill(str(skill_path))
        
        if "error" in checks:
            print(f"❌ Error: {checks['error']}")
            sys.exit(1)
        
        print(f"\n📊 Validation: {args.skill}")
        print("=" * 80)
        
        examples = checks.get("example_count", 0)
        size = checks.get("size_kb", 0)
        print(f"Examples: {examples}, Size: {size:.1f}KB")
        print(f"Required: {checks.get('required_score', 'N/A')}")
        print(f"Optional: {checks.get('optional_score', 'N/A')}")
        print()
        
        print("Required Checks:")
        for check in validator.REQUIRED_CHECKS:
            result = checks.get(check, False)
            status = "✅" if result else "❌"
            check_name = check.replace('_', ' ').title()
            print(f"  {status} {check_name}")
        
        print("\nOptional Checks:")
        for check in validator.OPTIONAL_CHECKS:
            result = checks.get(check, False)
            status = "✅" if result else "⚠️ "
            check_name = check.replace('_', ' ').title()
            print(f"  {status} {check_name}")
        
        print()
        print("=" * 80)
        
        if checks.get("all_required_pass", False):
            print("✅ PASSED: All required checks passed")
            sys.exit(0)
        else:
            print("❌ FAILED: Some required checks failed")
            sys.exit(1)
    
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
