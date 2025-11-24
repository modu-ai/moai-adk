#!/usr/bin/env python3
"""
Phase 2 Task 3: Metadata Consolidation
Standardizes frontmatter, creates dependency maps, and generates search index.
"""

import os
import re
import json
from pathlib import Path
from typing import Dict, List, Set, Optional
from dataclasses import dataclass, asdict
import yaml

@dataclass
class SkillMetadata:
    name: str
    description: str
    allowed_tools: List[str]
    auto_trigger_keywords: List[str]
    dependencies: List[str]
    category: str
    status: str = "production"
    version: str = "1.0.0"

class MetadataConsolidator:
    """Consolidate and standardize skill metadata."""

    def __init__(self, skills_base_dir: str):
        self.skills_base_dir = Path(skills_base_dir)
        self.metadata_registry: Dict[str, SkillMetadata] = {}
        self.dependency_map: Dict[str, List[str]] = {}

    def extract_all_metadata(self) -> Dict[str, SkillMetadata]:
        """Extract metadata from all skills."""

        for skill_dir in self.skills_base_dir.iterdir():
            if not skill_dir.is_dir():
                continue

            skill_md_path = skill_dir / "SKILL.md"
            if not skill_md_path.exists():
                continue

            metadata = self.extract_skill_metadata(skill_md_path)
            if metadata:
                self.metadata_registry[metadata.name] = metadata

        return self.metadata_registry

    def extract_skill_metadata(self, skill_md_path: Path) -> Optional[SkillMetadata]:
        """Extract metadata from SKILL.md frontmatter."""

        with open(skill_md_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Extract YAML frontmatter
        if not content.startswith('---'):
            return None

        parts = content.split('---', 2)
        if len(parts) < 3:
            return None

        try:
            frontmatter = yaml.safe_load(parts[1])
        except yaml.YAMLError:
            return None

        # Extract metadata
        name = frontmatter.get('name', skill_md_path.parent.name)
        description = frontmatter.get('description', '')

        # Parse allowed-tools (comma-separated or list)
        allowed_tools_raw = frontmatter.get('allowed-tools', '')
        if isinstance(allowed_tools_raw, str):
            allowed_tools = [tool.strip() for tool in allowed_tools_raw.split(',') if tool.strip()]
        else:
            allowed_tools = allowed_tools_raw if allowed_tools_raw else []

        # Extract auto-trigger keywords from description
        auto_trigger_keywords = self.extract_keywords(description, content)

        # Detect dependencies from content
        dependencies = self.extract_dependencies(content)

        # Categorize skill
        category = self.categorize_skill(name)

        return SkillMetadata(
            name=name,
            description=description,
            allowed_tools=allowed_tools,
            auto_trigger_keywords=auto_trigger_keywords,
            dependencies=dependencies,
            category=category
        )

    def extract_keywords(self, description: str, content: str) -> List[str]:
        """Extract auto-trigger keywords from description and content."""

        keywords = set()

        # From description
        desc_words = re.findall(r'\b[a-z]{4,}\b', description.lower())
        keywords.update(desc_words[:5])  # Take top 5 words

        # From headings
        heading_pattern = re.compile(r'^#{1,3}\s+(.+)$', re.MULTILINE)
        headings = heading_pattern.findall(content)
        for heading in headings[:3]:  # Take top 3 headings
            heading_words = re.findall(r'\b[a-z]{4,}\b', heading.lower())
            keywords.update(heading_words[:2])

        return sorted(list(keywords))[:10]  # Max 10 keywords

    def extract_dependencies(self, content: str) -> List[str]:
        """Extract skill dependencies from content."""

        dependencies = set()

        # Find Skill() references
        skill_pattern = re.compile(r'Skill\(["\']([^"\']+)["\']\)')
        skill_refs = skill_pattern.findall(content)
        dependencies.update(skill_refs)

        # Find moai- skill references
        moai_pattern = re.compile(r'`(moai-[a-z-]+)`')
        moai_refs = moai_pattern.findall(content)
        dependencies.update(moai_refs)

        return sorted(list(dependencies))

    def categorize_skill(self, skill_name: str) -> str:
        """Categorize skill based on name prefix."""

        if skill_name.startswith('moai-cc-'):
            return 'claude-code'
        elif skill_name.startswith('moai-foundation-'):
            return 'foundation'
        elif skill_name.startswith('moai-lang-'):
            return 'language'
        elif skill_name.startswith('moai-domain-'):
            return 'domain'
        elif skill_name.startswith('moai-baas-'):
            return 'backend-as-a-service'
        elif skill_name.startswith('moai-security-'):
            return 'security'
        elif skill_name.startswith('moai-essentials-'):
            return 'essentials'
        elif skill_name.startswith('moai-core-'):
            return 'core'
        else:
            return 'other'

    def build_dependency_map(self) -> Dict[str, List[str]]:
        """Build dependency map from extracted metadata."""

        for skill_name, metadata in self.metadata_registry.items():
            self.dependency_map[skill_name] = metadata.dependencies

        return self.dependency_map

    def generate_search_index(self) -> Dict[str, Dict]:
        """Generate search index for skill discovery."""

        search_index = {}

        for skill_name, metadata in self.metadata_registry.items():
            search_index[skill_name] = {
                "keywords": metadata.auto_trigger_keywords,
                "category": metadata.category,
                "description": metadata.description,
                "tools": metadata.allowed_tools,
                "dependencies": metadata.dependencies
            }

        return search_index

    def generate_consolidation_report(self) -> str:
        """Generate metadata consolidation report."""

        report_lines = [
            "=" * 80,
            "PHASE 2 TASK 3: METADATA CONSOLIDATION REPORT",
            "=" * 80,
            ""
        ]

        total_skills = len(self.metadata_registry)
        report_lines.append(f"Total skills analyzed: {total_skills}")
        report_lines.append("")

        # Category distribution
        report_lines.append("CATEGORY DISTRIBUTION")
        report_lines.append("-" * 80)

        category_counts = {}
        for metadata in self.metadata_registry.values():
            category_counts[metadata.category] = category_counts.get(metadata.category, 0) + 1

        for category in sorted(category_counts.keys()):
            count = category_counts[category]
            percentage = (count / total_skills * 100) if total_skills > 0 else 0
            report_lines.append(f"{category:<30} {count:>3} ({percentage:>5.1f}%)")

        report_lines.append("")

        # Dependency analysis
        report_lines.append("DEPENDENCY ANALYSIS")
        report_lines.append("-" * 80)

        skills_with_deps = sum(1 for metadata in self.metadata_registry.values() if metadata.dependencies)
        dep_coverage = (skills_with_deps / total_skills * 100) if total_skills > 0 else 0

        report_lines.append(f"Skills with dependencies: {skills_with_deps}/{total_skills} ({dep_coverage:.1f}%)")

        avg_deps = sum(len(metadata.dependencies) for metadata in self.metadata_registry.values()) / total_skills if total_skills > 0 else 0
        report_lines.append(f"Average dependencies per skill: {avg_deps:.1f}")
        report_lines.append("")

        # Top dependencies
        all_deps = []
        for metadata in self.metadata_registry.values():
            all_deps.extend(metadata.dependencies)

        from collections import Counter
        dep_counts = Counter(all_deps)

        report_lines.append("TOP DEPENDENCIES (most referenced skills)")
        for dep, count in dep_counts.most_common(10):
            report_lines.append(f"  {count:>2}x  {dep}")

        report_lines.append("")

        # Tool usage analysis
        report_lines.append("ALLOWED-TOOLS ANALYSIS")
        report_lines.append("-" * 80)

        all_tools = []
        for metadata in self.metadata_registry.values():
            all_tools.extend(metadata.allowed_tools)

        tool_counts = Counter(all_tools)

        report_lines.append("TOP TOOLS (most commonly allowed)")
        for tool, count in tool_counts.most_common(10):
            percentage = (count / total_skills * 100) if total_skills > 0 else 0
            report_lines.append(f"  {count:>3}x ({percentage:>5.1f}%)  {tool}")

        return "\n".join(report_lines)

    def save_metadata_registry(self, output_path: str):
        """Save complete metadata registry as JSON."""

        registry_data = {
            skill_name: asdict(metadata)
            for skill_name, metadata in self.metadata_registry.items()
        }

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(registry_data, f, indent=2)

        print(f"✅ Metadata registry saved: {output_path}")

    def save_dependency_map(self, output_path: str):
        """Save dependency map as JSON."""

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(self.dependency_map, f, indent=2)

        print(f"✅ Dependency map saved: {output_path}")

    def save_search_index(self, output_path: str):
        """Save search index as JSON."""

        search_index = self.generate_search_index()

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(search_index, f, indent=2)

        print(f"✅ Search index saved: {output_path}")

def main():
    """Main consolidation execution."""

    skills_base_dir = "/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills"

    print("🔍 Starting Phase 2 Task 3: Metadata Consolidation...")
    print()

    consolidator = MetadataConsolidator(skills_base_dir)

    # Extract all metadata
    metadata_registry = consolidator.extract_all_metadata()

    # Build dependency map
    consolidator.build_dependency_map()

    # Generate report
    report = consolidator.generate_consolidation_report()
    print(report)
    print()

    # Save artifacts
    logs_dir = Path("/Users/goos/MoAI/MoAI-ADK/.moai/logs")
    logs_dir.mkdir(parents=True, exist_ok=True)

    consolidator.save_metadata_registry(str(logs_dir / "phase2-task3-metadata-registry.json"))
    consolidator.save_dependency_map(str(logs_dir / "phase2-task3-dependency-map.json"))
    consolidator.save_search_index(str(logs_dir / "phase2-task3-search-index.json"))

    print()
    print("=" * 80)
    print("METADATA CONSOLIDATION COMPLETE")
    print("=" * 80)

if __name__ == "__main__":
    main()
