#!/usr/bin/env python3
"""
MoAI-ADK Skill Upgrader to v4.0 Enterprise

Automates skill upgrades to v4.0 Enterprise standard:
- Add/update YAML frontmatter
- Restructure to Progressive Disclosure (3 levels)
- Add Context7 integration section
- Expand code examples to 10+
- Add best practices checklist
- Add decision tree
- Add related skills
- Add official references

Usage:
    python3 scripts/upgrade-skills-to-v4.py --skill moai-alfred-agent-guide
    python3 scripts/upgrade-skills-to-v4.py --batch phase1
    python3 scripts/upgrade-skills-to-v4.py --validate-all
"""

import re
import yaml
import shutil
from pathlib import Path
from typing import List, Dict
from dataclasses import dataclass
from datetime import datetime


@dataclass
class SkillMetadata:
    """Skill metadata for v4.0"""
    name: str
    version: str
    created: str
    updated: str
    status: str
    tier: str
    description: str
    allowed_tools: str
    primary_agent: str
    secondary_agents: List[str]
    keywords: List[str]
    tags: List[str]
    orchestration: Dict[str, any]


class SkillUpgrader:
    """Automate skill upgrades to v4.0 Enterprise"""
    
    V4_TEMPLATE = """---
name: {name}
version: 4.0.0
created: {created}
updated: {updated}
status: active
tier: {tier}
description: "{description}"
allowed-tools: "{allowed_tools}"
primary-agent: "{primary_agent}"
secondary-agents: {secondary_agents}
keywords: {keywords}
tags: {tags}
orchestration:
  can_resume: true
  typical_chain_position: "{chain_position}"
  depends_on: []
---

# {name}

**{title}**

> **Primary Agent**: {primary_agent}  
> **Secondary Agents**: {secondary_agents_str}  
> **Version**: 4.0.0  
> **Keywords**: {keywords_str}

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

{level1_content}

---

### Level 2: Practical Implementation (Common Patterns)

{level2_content}

---

### Level 3: Advanced Patterns (Expert Reference)

{level3_content}

---

## 🎯 Best Practices Checklist

{best_practices}

---

## 🔗 Context7 MCP Integration

{context7_section}

---

## 📊 Decision Tree

{decision_tree}

---

## 🔄 Integration with Other Skills

{related_skills}

---

## 📚 Official References

{references}

---

## 📈 Version History

**v4.0.0** ({updated})
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references

{previous_versions}

---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: {updated}  
**Maintained by**: Primary Agent ({primary_agent})
"""
    
    def __init__(self, skills_dir: str = ".claude/skills"):
        self.skills_dir = Path(skills_dir)
        self.backup_suffix = datetime.now().strftime(".backup-%Y%m%d-%H%M%S")
        self.stats = {
            "upgraded": [],
            "failed": [],
            "skipped": []
        }
    
    def upgrade_skill(self, skill_name: str, dry_run: bool = False) -> bool:
        """
        Upgrade single skill to v4.0
        
        Args:
            skill_name: Name of skill to upgrade
            dry_run: If True, only validate without writing
            
        Returns:
            True if successful, False otherwise
        """
        skill_path = self.skills_dir / skill_name
        skill_md = skill_path / "SKILL.md"
        
        if not skill_md.exists():
            print(f"❌ {skill_name}: SKILL.md not found")
            self.stats["failed"].append(skill_name)
            return False
        
        print(f"\n📝 Upgrading {skill_name}...")
        
        try:
            # Step 1: Backup original
            if not dry_run:
                backup_path = skill_md.with_suffix(f".md{self.backup_suffix}")
                shutil.copy2(skill_md, backup_path)
                print(f"  ✅ Backup created: {backup_path.name}")
            
            # Step 2: Parse current content
            current = self._parse_skill(skill_md)
            
            # Step 3: Extract metadata
            metadata = self._extract_metadata(skill_name, current)
            
            # Step 4: Check current version
            if current.get("version", "").startswith("4."):
                print(f"  ⏭️  Already v4.0, skipping")
                self.stats["skipped"].append(skill_name)
                return True
            
            # Step 5: Restructure to Progressive Disclosure
            new_content = self._restructure_content(current, metadata)
            
            # Step 6: Add Context7 section (template)
            new_content = self._add_context7_section(new_content, metadata)
            
            # Step 7: Generate v4.0 content
            v4_content = self._generate_v4_content(metadata, new_content)
            
            # Step 8: Validate
            validation = self._validate_v4(v4_content)
            
            if not validation["valid"]:
                print(f"  ❌ Validation failed:")
                for issue in validation["issues"]:
                    print(f"     - {issue}")
                self.stats["failed"].append(skill_name)
                return False
            
            # Step 9: Write new content
            if not dry_run:
                skill_md.write_text(v4_content, encoding='utf-8')
                print(f"  ✅ Upgraded to v4.0.0")
                print(f"  📊 Stats:")
                print(f"     - Size: {len(current['raw'])} → {len(v4_content)} bytes")
                print(f"     - Examples: {current.get('examples', 0)} → {validation['examples']}")
                print(f"     - Sections: {current.get('sections', 0)} → {validation['sections']}")
            else:
                print(f"  🔍 Dry run - no changes made")
            
            self.stats["upgraded"].append(skill_name)
            return True
            
        except Exception as e:
            print(f"  ❌ Error: {e}")
            self.stats["failed"].append(skill_name)
            return False
    
    def batch_upgrade(self, skill_names: List[str], dry_run: bool = False) -> Dict[str, List[str]]:
        """Upgrade multiple skills"""
        print(f"\n🚀 Batch upgrading {len(skill_names)} skills...")
        
        for i, skill_name in enumerate(skill_names, 1):
            print(f"\n[{i}/{len(skill_names)}]", end=" ")
            self.upgrade_skill(skill_name, dry_run=dry_run)
        
        # Print summary
        print(f"\n{'='*80}")
        print(f"📊 Batch Upgrade Summary")
        print(f"{'='*80}")
        print(f"✅ Upgraded: {len(self.stats['upgraded'])}")
        print(f"❌ Failed: {len(self.stats['failed'])}")
        print(f"⏭️  Skipped: {len(self.stats['skipped'])}")
        
        if self.stats["failed"]:
            print(f"\n❌ Failed skills:")
            for skill in self.stats["failed"]:
                print(f"   - {skill}")
        
        return self.stats
    
    def _parse_skill(self, skill_md: Path) -> Dict:
        """Parse existing skill file"""
        content = skill_md.read_text(encoding='utf-8')
        
        # Extract frontmatter
        frontmatter_match = re.search(r'^---\n(.*?)\n---', content, re.DOTALL)
        frontmatter = {}
        body = content
        
        if frontmatter_match:
            frontmatter_text = frontmatter_match.group(1)
            frontmatter = yaml.safe_load(frontmatter_text)
            body = content[frontmatter_match.end():].strip()
        
        # Count code examples
        code_blocks = len(re.findall(r'```', body))
        examples = code_blocks // 2
        
        # Count sections
        sections = len(re.findall(r'^##\s+', body, re.MULTILINE))
        
        return {
            "raw": content,
            "frontmatter": frontmatter,
            "body": body,
            "version": frontmatter.get("version", "unknown"),
            "examples": examples,
            "sections": sections
        }
    
    def _extract_metadata(self, skill_name: str, current: Dict) -> SkillMetadata:
        """Extract and enhance metadata for v4.0"""
        fm = current["frontmatter"]
        
        # Determine tier
        tier = self._determine_tier(skill_name)
        
        # Determine primary agent
        primary_agent = self._determine_primary_agent(skill_name, fm)
        
        # Generate enhanced description
        description = self._enhance_description(fm.get("description", ""), skill_name)
        
        # Determine secondary agents
        secondary_agents = self._determine_secondary_agents(skill_name, primary_agent)
        
        # Extract/generate keywords
        keywords = self._extract_keywords(skill_name, current["body"])
        
        # Enhanced tags
        tags = fm.get("tags", [])
        if not tags:
            tags = self._generate_tags(skill_name)
        
        # Determine chain position
        chain_position = self._determine_chain_position(skill_name)
        
        return SkillMetadata(
            name=skill_name,
            version="4.0.0",
            created=fm.get("created", datetime.now().strftime("%Y-%m-%d")),
            updated=datetime.now().strftime("%Y-%m-%d"),
            status="active",
            tier=tier,
            description=description,
            allowed_tools=self._enhance_allowed_tools(fm.get("allowed-tools", "Read, Glob, Grep")),
            primary_agent=primary_agent,
            secondary_agents=secondary_agents,
            keywords=keywords,
            tags=tags,
            orchestration={
                "can_resume": True,
                "typical_chain_position": chain_position,
                "depends_on": []
            }
        )
    
    def _determine_tier(self, skill_name: str) -> str:
        """Determine skill tier"""
        if "foundation" in skill_name:
            return "foundation"
        elif "essentials" in skill_name:
            return "essentials"
        elif "domain" in skill_name:
            return "domain"
        elif "lang-" in skill_name:
            return "language"
        elif "baas-" in skill_name:
            return "baas"
        else:
            return "specialization"
    
    def _determine_primary_agent(self, skill_name: str, frontmatter: Dict) -> str:
        """Determine primary agent for skill"""
        # Mapping of skill patterns to agents
        agent_map = {
            "alfred-agent": "alfred",
            "alfred-workflow": "alfred",
            "alfred-personas": "alfred",
            "alfred-context": "alfred",
            "alfred-todowrite": "plan-agent",
            "alfred-spec": "spec-builder",
            "alfred-git": "git-manager",
            "alfred-practices": "qa-validator",
            "alfred-code-reviewer": "code-reviewer",
            "alfred-config": "config-manager",
            "alfred-session": "session-manager",
            "context7": "mcp-context7-integrator",
            "domain-backend": "backend-expert",
            "domain-frontend": "frontend-expert",
            "domain-database": "database-expert",
            "domain-devops": "devops-expert",
            "domain-security": "security-expert",
            "domain-ml": "ml-expert",
            "domain-data": "data-science-expert",
            "domain-mobile": "mobile-expert",
            "domain-web-api": "api-expert",
            "security-": "security-expert",
            "lang-": "language-expert",
            "docs-": "doc-syncer",
            "test": "test-engineer",
            "mcp-": "mcp-builder",
        }
        
        for pattern, agent in agent_map.items():
            if pattern in skill_name:
                return agent
        
        return "alfred"  # Default
    
    def _determine_secondary_agents(self, skill_name: str, primary: str) -> List[str]:
        """Determine secondary agents"""
        # Common collaborators
        common = ["alfred", "plan-agent"]
        
        secondary = []
        
        if primary != "alfred":
            secondary.append("alfred")
        
        if "domain-" in skill_name or "security-" in skill_name:
            secondary.extend(["qa-validator", "doc-syncer"])
        
        if "alfred-" in skill_name:
            secondary.extend(["plan-agent", "session-manager"])
        
        # Remove duplicates and primary
        secondary = list(set(secondary))
        if primary in secondary:
            secondary.remove(primary)
        
        return secondary[:3]  # Max 3 secondary agents
    
    def _extract_keywords(self, skill_name: str, body: str) -> List[str]:
        """Extract/generate keywords for auto-triggering"""
        keywords = []
        
        # Extract from skill name
        name_parts = skill_name.replace("moai-", "").split("-")
        keywords.extend(name_parts[:3])
        
        # Extract from content (common technical terms)
        technical_terms = re.findall(r'\b(api|test|debug|perf|auth|security|database|frontend|backend|git|docker|kubernetes|ci|cd|tdd|spec)\b', body.lower())
        keywords.extend(list(set(technical_terms))[:5])
        
        return keywords[:5]  # Max 5 keywords
    
    def _generate_tags(self, skill_name: str) -> List[str]:
        """Generate relevant tags"""
        tags = []
        
        if "alfred" in skill_name:
            tags.append("alfred-core")
        if "domain" in skill_name:
            tags.append("domain-expert")
        if "security" in skill_name:
            tags.extend(["security", "best-practices"])
        if "lang" in skill_name:
            tags.append("programming-language")
        if "test" in skill_name:
            tags.append("testing")
        if "docs" in skill_name:
            tags.append("documentation")
        
        return tags
    
    def _enhance_description(self, current_desc: str, skill_name: str) -> str:
        """Enhance description for v4.0"""
        if not current_desc:
            current_desc = f"Enhanced {skill_name.replace('moai-', '').replace('-', ' ')} with AI-powered features"
        
        # Add Context7 mention if not present
        if "context7" not in current_desc.lower() and "mcp" not in current_desc.lower():
            current_desc += ". Enhanced with Context7 MCP for up-to-date documentation."
        
        return current_desc
    
    def _enhance_allowed_tools(self, current_tools) -> str:
        """Add Context7 tools to allowed-tools"""
        # Handle both string and list input
        if isinstance(current_tools, list):
            tools = current_tools
        else:
            tools = current_tools.split(", ")
        
        context7_tools = [
            "WebSearch",
            "WebFetch",
            "mcp__context7__resolve-library-id",
            "mcp__context7__get-library-docs"
        ]
        
        for tool in context7_tools:
            if tool not in tools:
                tools.append(tool)
        
        return ", ".join(tools)
    
    def _determine_chain_position(self, skill_name: str) -> str:
        """Determine typical position in agent chain"""
        if "alfred-agent" in skill_name or "alfred-workflow" in skill_name:
            return "initial"
        elif "alfred-git" in skill_name or "docs-" in skill_name:
            return "terminal"
        else:
            return "middle"
    
    def _restructure_content(self, current: Dict, metadata: SkillMetadata) -> Dict:
        """Restructure content to Progressive Disclosure"""
        body = current["body"]
        
        # Split into sections
        sections = re.split(r'\n##\s+', body)
        
        # Organize into 3 levels
        level1 = self._extract_level1(sections, metadata)
        level2 = self._extract_level2(sections, metadata)
        level3 = self._extract_level3(sections, metadata)
        
        return {
            "level1": level1,
            "level2": level2,
            "level3": level3,
            "best_practices": self._generate_best_practices(sections),
            "decision_tree": self._generate_decision_tree(metadata),
            "related_skills": self._generate_related_skills(metadata),
            "references": self._extract_references(sections)
        }
    
    def _extract_level1(self, sections: List[str], metadata: SkillMetadata) -> str:
        """Extract/generate Level 1 content"""
        # Find "What It Does" or first section
        for section in sections:
            if "what" in section.lower() or "overview" in section.lower():
                return section.strip()
        
        # Generate default
        return f"""**Purpose**: {metadata.description}

**When to Use:**
- ✅ [Use case 1]
- ✅ [Use case 2]
- ✅ [Use case 3]

**Quick Start Pattern:**

```python
# Basic example
# TODO: Add practical example
```
"""
    
    def _extract_level2(self, sections: List[str], metadata: SkillMetadata) -> str:
        """Extract/generate Level 2 content"""
        patterns = []
        
        for section in sections:
            if "pattern" in section.lower() or "example" in section.lower():
                patterns.append(section.strip())
        
        if not patterns:
            # Generate template patterns
            return """#### Pattern 1: Basic Usage

**Use Case**: [When to use]

**Implementation:**

```
# Code example
```

#### Pattern 2: Advanced Usage

**Use Case**: [When to use]

**Implementation:**

```
# Code example
```
"""
        
        return "\n\n---\n\n".join(patterns)
    
    def _extract_level3(self, sections: List[str], metadata: SkillMetadata) -> str:
        """Extract/generate Level 3 content"""
        return """> **Note**: Advanced patterns for complex scenarios.

**Coming soon**: Deep dive into expert-level usage.
"""
    
    def _generate_best_practices(self, sections: List[str]) -> str:
        """Generate best practices checklist"""
        return """**Must-Have:**
- ✅ [Critical practice 1]
- ✅ [Critical practice 2]

**Recommended:**
- ✅ [Recommended practice 1]
- ✅ [Recommended practice 2]

**Security:**
- 🔒 [Security practice 1]
"""
    
    def _generate_decision_tree(self, metadata: SkillMetadata) -> str:
        """Generate decision tree"""
        return f"""**When to use {metadata.name}:**

```
Start
  ├─ Need {metadata.keywords[0] if metadata.keywords else 'feature'}?
  │   ├─ YES → Use this skill
  │   └─ NO → Consider alternatives
  └─ Complex scenario?
      ├─ YES → See Level 3
      └─ NO → Start with Level 1
```
"""
    
    def _generate_related_skills(self, metadata: SkillMetadata) -> str:
        """Generate related skills section"""
        return """**Prerequisite Skills:**
- Skill("prerequisite-1") – [Why needed]

**Complementary Skills:**
- Skill("complementary-1") – [How they work together]

**Next Steps:**
- Skill("next-step-1") – [When to use after this]
"""
    
    def _extract_references(self, sections: List[str]) -> str:
        """Extract official references"""
        for section in sections:
            if "reference" in section.lower() or "link" in section.lower():
                return section.strip()
        
        return """**Primary Documentation:**
- [Official Docs](https://...) – Complete reference

**Best Practices:**
- [Best Practices Guide](https://...) – Official recommendations
"""
    
    def _add_context7_section(self, content: Dict, metadata: SkillMetadata) -> Dict:
        """Add Context7 integration section"""
        content["context7"] = f"""**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with [{metadata.keywords[0] if metadata.keywords else 'libraries'}]
- Need latest documentation
- Verifying technical details

**Example Usage:**

```python
# Fetch latest documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/org/library",
    topic="{metadata.keywords[0] if metadata.keywords else 'topic'}",
    tokens=5000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| [Library 1] | `/org/lib1` | [When to use] |
"""
        
        return content
    
    def _generate_v4_content(self, metadata: SkillMetadata, content: Dict) -> str:
        """Generate complete v4.0 content"""
        return self.V4_TEMPLATE.format(
            name=metadata.name,
            created=metadata.created,
            updated=metadata.updated,
            tier=metadata.tier,
            description=metadata.description,
            allowed_tools=metadata.allowed_tools,
            primary_agent=metadata.primary_agent,
            secondary_agents=yaml.dump(metadata.secondary_agents, default_flow_style=True).strip(),
            keywords=yaml.dump(metadata.keywords, default_flow_style=True).strip(),
            tags=yaml.dump(metadata.tags, default_flow_style=True).strip(),
            chain_position=metadata.orchestration["typical_chain_position"],
            title=metadata.name.replace("moai-", "").replace("-", " ").title(),
            secondary_agents_str=", ".join(metadata.secondary_agents) if metadata.secondary_agents else "none",
            keywords_str=", ".join(metadata.keywords),
            level1_content=content["level1"],
            level2_content=content["level2"],
            level3_content=content["level3"],
            best_practices=content["best_practices"],
            context7_section=content.get("context7", ""),
            decision_tree=content["decision_tree"],
            related_skills=content["related_skills"],
            references=content["references"],
            previous_versions=""
        )
    
    def _validate_v4(self, content: str) -> Dict:
        """Validate v4.0 compliance"""
        issues = []
        
        # Check frontmatter
        if not re.search(r'^---\nname:', content):
            issues.append("Missing YAML frontmatter")
        
        # Check version
        if "version: 4.0.0" not in content:
            issues.append("Version not 4.0.0")
        
        # Check primary agent
        if "primary-agent:" not in content:
            issues.append("Missing primary-agent")
        
        # Check keywords
        if "keywords:" not in content:
            issues.append("Missing keywords")
        
        # Check Progressive Disclosure
        if "### Level 1:" not in content:
            issues.append("Missing Level 1")
        if "### Level 2:" not in content:
            issues.append("Missing Level 2")
        
        # Count code examples
        code_blocks = len(re.findall(r'```', content))
        examples = code_blocks // 2
        
        if examples < 10:
            issues.append(f"Only {examples} code examples (need 10+)")
        
        # Check Context7 section
        if "Context7" not in content and "MCP" not in content:
            issues.append("Missing Context7 integration section")
        
        # Check best practices
        if "Best Practices Checklist" not in content:
            issues.append("Missing best practices checklist")
        
        # Count sections
        sections = len(re.findall(r'^##\s+', content, re.MULTILINE))
        
        return {
            "valid": len(issues) == 0,
            "issues": issues,
            "examples": examples,
            "sections": sections
        }
    
    def validate_all_v4_skills(self) -> Dict[str, Dict]:
        """Validate all v4.0 skills compliance"""
        results = {}
        
        for skill_dir in sorted(self.skills_dir.iterdir()):
            if not skill_dir.is_dir() or skill_dir.name.startswith('.'):
                continue
            
            skill_md = skill_dir / "SKILL.md"
            if not skill_md.exists():
                continue
            
            try:
                content = skill_md.read_text(encoding='utf-8')
                validation = self._validate_v4(content)
                results[skill_dir.name] = validation
            except Exception as e:
                results[skill_dir.name] = {
                    "valid": False,
                    "issues": [str(e)],
                    "examples": 0,
                    "sections": 0
                }
        
        return results


def main():
    """CLI interface for skill upgrader"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Upgrade MoAI-ADK skills to v4.0 Enterprise")
    parser.add_argument("--skill", help="Single skill to upgrade")
    parser.add_argument("--batch", choices=["phase1", "phase2", "phase3", "phase4", "all"], 
                       help="Batch upgrade by phase")
    parser.add_argument("--validate-all", action="store_true", 
                       help="Validate all skills for v4.0 compliance")
    parser.add_argument("--dry-run", action="store_true", 
                       help="Dry run without writing changes")
    parser.add_argument("--skills-dir", default=".claude/skills", 
                       help="Skills directory path")
    
    args = parser.parse_args()
    
    upgrader = SkillUpgrader(args.skills_dir)
    
    if args.validate_all:
        results = upgrader.validate_all_v4_skills()
        print("\n📊 Validation Results:")
        for skill, validation in results.items():
            status = "✅" if validation["valid"] else "❌"
            print(f"{status} {skill}: {validation['examples']} examples, {validation['sections']} sections")
            if not validation["valid"]:
                for issue in validation["issues"]:
                    print(f"   - {issue}")
    
    elif args.skill:
        upgrader.upgrade_skill(args.skill, dry_run=args.dry_run)
    
    elif args.batch:
        # Define phase batches
        phases = {
            "phase1": [
                # Unknown versions (16)
                "moai-domain-backend", "moai-domain-frontend", "moai-domain-database",
                "moai-domain-security", "moai-domain-web-api", "moai-domain-data-science",
                "moai-domain-devops", "moai-domain-ml", "moai-domain-mobile-app",
                "moai-security-authentication", "moai-security-authorization",
                "moai-security-encryption", "moai-security-owasp",
                "moai-mcp-builder", "moai-project-documentation", "moai-webapp-testing",
                # Alfred Core top 5
                "moai-alfred-agent-guide", "moai-alfred-workflow", "moai-alfred-context-budget",
                "moai-alfred-personas", "moai-alfred-todowrite-pattern"
            ],
            "phase2": [
                # Alfred Core middle (16)
                "moai-alfred-spec-authoring", "moai-alfred-practices",
                "moai-alfred-proactive-suggestions", "moai-alfred-clone-pattern",
                "moai-alfred-code-reviewer", "moai-alfred-config-schema",
                "moai-alfred-dev-guide", "moai-alfred-expertise-detection",
                "moai-alfred-issue-labels", "moai-alfred-language-detection",
                "moai-alfred-rules", "moai-alfred-session-state",
                "moai-context7-integration", "moai-lang-shell",
                "moai-lang-template", "moai-project-config-manager"
            ],
            "phase3": [
                # v1.0 docs & project tools (9)
                "moai-docs-generation", "moai-docs-linting",
                "moai-docs-unified", "moai-docs-validation",
                "moai-project-batch-questions", "moai-project-language-initializer",
                "moai-project-template-optimizer", "moai-change-logger",
                "moai-tag-policy-validator"
            ],
            "phase4": [
                # v1.0 specialized (7)
                "moai-design-systems", "moai-jit-docs-enhanced",
                "moai-learning-optimizer", "moai-mermaid-diagram-expert",
                "moai-readme-expert", "moai-session-info",
                "moai-streaming-ui"
            ]
        }
        
        if args.batch == "all":
            all_skills = []
            for phase_skills in phases.values():
                all_skills.extend(phase_skills)
            upgrader.batch_upgrade(all_skills, dry_run=args.dry_run)
        else:
            upgrader.batch_upgrade(phases[args.batch], dry_run=args.dry_run)
    
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
