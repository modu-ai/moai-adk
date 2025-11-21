# Advanced Patterns - Skill Factory

## Pattern 1: Multi-Domain Skill Generation with Context7

Automatically generate skills across multiple domains with intelligent Context7 pattern matching.

```python
from typing import Dict, List
from dataclasses import dataclass

@dataclass
class DomainContext:
    domain: str
    frameworks: List[str]
    libraries: List[str]
    patterns: List[str]
    best_practices: List[str]

class MultiDomainSkillFactory:
    """Generate skills tailored to specific domains."""

    async def create_cross_domain_skill(
        self,
        skill_name: str,
        domains: List[str],
        context7_client
    ) -> Dict[str, str]:
        """Create skill with domain-specific content from Context7."""

        skill_content = {}

        # Get Context7 patterns for each domain
        for domain in domains:
            # Fetch latest patterns
            patterns = await context7_client.get_library_docs(
                context7_library_id=f"/moai/domains/{domain}",
                topic=f"{skill_name} {domain} patterns best practices",
                tokens=5000
            )

            # Generate domain-specific SKILL section
            skill_content[domain] = self._generate_domain_section(
                skill_name, domain, patterns
            )

        # Generate unified SKILL.md
        unified_content = self._merge_domain_content(
            skill_name, skill_content
        )

        return {
            'skill_md': unified_content,
            'domain_modules': skill_content,
            'integration_points': self._identify_integration_points(domains)
        }

    def _generate_domain_section(self, skill_name: str, domain: str, patterns: Dict) -> str:
        """Generate domain-specific content section."""
        return f"""
## {domain.title()} Context

**Frameworks**: {', '.join(patterns.get('frameworks', []))}
**Key Libraries**: {', '.join(patterns.get('libraries', []))}

### Domain-Specific Patterns

{self._format_patterns(patterns.get('patterns', []))}

### Best Practices for {domain.title()}

{self._format_best_practices(patterns.get('best_practices', []))}
"""
```

## Pattern 2: Template-Based Skill Generation with Inheritance

Create reusable skill templates with inheritance chains for consistency.

```python
from abc import ABC, abstractmethod
from typing import Optional

class SkillTemplate(ABC):
    """Base template for skill generation."""

    def __init__(self, skill_name: str):
        self.skill_name = skill_name
        self.sections = {}

    @abstractmethod
    def generate_quick_reference(self) -> str:
        """Generate Quick Reference section."""
        pass

    @abstractmethod
    def generate_capabilities(self) -> List[str]:
        """List core capabilities."""
        pass

    def generate_skill_md(self) -> str:
        """Generate complete SKILL.md with template sections."""
        sections = [
            self.generate_quick_reference(),
            self._generate_what_it_does(),
            self._generate_when_to_use(),
            self._generate_best_practices(),
            self._generate_works_well_with()
        ]
        return '\n\n'.join(sections)

class LanguageSkillTemplate(SkillTemplate):
    """Template for language-specific skills."""

    def __init__(self, skill_name: str, language: str, version: str):
        super().__init__(skill_name)
        self.language = language
        self.version = version

    def generate_quick_reference(self) -> str:
        return f"""
# {self.skill_name}

**Language**: {self.language}
**Version**: {self.version}
**Best For**: [Domain]

Key features:
{chr(10).join([f'- {cap}' for cap in self.generate_capabilities()])}
"""

class BackendSkillTemplate(SkillTemplate):
    """Template for backend development skills."""

    def generate_quick_reference(self) -> str:
        return f"""
# {self.skill_name} - Backend Development

**Focus Areas**:
- API Design and Development
- Database Integration
- Performance Optimization
- Security Hardening

{chr(10).join([f'- {cap}' for cap in self.generate_capabilities()])}
"""

# Usage
python_skill = LanguageSkillTemplate('moai-lang-python-advanced', 'Python', '3.12')
skill_content = python_skill.generate_skill_md()
```

## Pattern 3: Automated Skill Validation with Quality Gates

Validate generated skills against enterprise standards before deployment.

```python
from enum import Enum
from typing import Tuple

class SkillQualityLevel(Enum):
    PRODUCTION = "production"
    STAGING = "staging"
    DEVELOPMENT = "development"

class SkillValidator:
    """Multi-level validation for generated skills."""

    PRODUCTION_STANDARDS = {
        'min_lines': 300,
        'max_lines': 500,
        'min_examples': 10,
        'min_sections': 7,
        'coverage_target': 90,
        'security_scan': True,
        'context7_integration': True
    }

    async def validate_skill(
        self,
        skill_content: str,
        quality_level: SkillQualityLevel = SkillQualityLevel.PRODUCTION
    ) -> Tuple[bool, List[str]]:
        """Validate skill against quality standards."""

        issues = []
        standards = self._get_standards_for_level(quality_level)

        # Structure validation
        if not self._validate_structure(skill_content):
            issues.append("Invalid SKILL.md structure")

        # Line count validation
        lines = len(skill_content.split('\n'))
        if lines < standards['min_lines']:
            issues.append(f"Too short: {lines} lines (min: {standards['min_lines']})")
        if lines > standards['max_lines']:
            issues.append(f"Too long: {lines} lines (max: {standards['max_lines']})")

        # Content validation
        examples_count = skill_content.count('```')
        if examples_count < standards['min_examples']:
            issues.append(f"Too few examples: {examples_count} (min: {standards['min_examples']})")

        # Security validation
        if standards['security_scan']:
            security_issues = await self._security_scan(skill_content)
            issues.extend(security_issues)

        # Context7 integration check
        if standards['context7_integration']:
            if 'context7' not in skill_content.lower():
                issues.append("Missing Context7 integration documentation")

        return len(issues) == 0, issues

    def _get_standards_for_level(self, level: SkillQualityLevel) -> Dict:
        """Get validation standards for quality level."""
        if level == SkillQualityLevel.PRODUCTION:
            return self.PRODUCTION_STANDARDS
        elif level == SkillQualityLevel.STAGING:
            return {k: v * 0.9 for k, v in self.PRODUCTION_STANDARDS.items()}
        else:  # DEVELOPMENT
            return {k: v * 0.7 for k, v in self.PRODUCTION_STANDARDS.items()}
```

## Pattern 4: Batch Skill Generation with Parallel Processing

Generate multiple skills in parallel for team standardization.

```python
import asyncio
from concurrent.futures import ThreadPoolExecutor
from typing import List, Dict

class BatchSkillFactory:
    """Generate multiple skills efficiently with parallelization."""

    def __init__(self, max_workers: int = 4):
        self.max_workers = max_workers
        self.executor = ThreadPoolExecutor(max_workers=max_workers)

    async def generate_skills_batch(
        self,
        skill_specs: List[Dict],
        context7_client
    ) -> Dict[str, Dict]:
        """Generate multiple skills in parallel."""

        tasks = [
            self._generate_single_skill_async(spec, context7_client)
            for spec in skill_specs
        ]

        results = await asyncio.gather(*tasks, return_exceptions=True)

        # Process results
        generated_skills = {}
        errors = {}

        for spec, result in zip(skill_specs, results):
            if isinstance(result, Exception):
                errors[spec['name']] = str(result)
            else:
                generated_skills[spec['name']] = result

        return {
            'successful': generated_skills,
            'failed': errors,
            'total': len(skill_specs),
            'success_rate': len(generated_skills) / len(skill_specs)
        }

    async def _generate_single_skill_async(self, spec: Dict, context7_client) -> Dict:
        """Generate single skill asynchronously."""

        # Fetch Context7 patterns
        patterns = await context7_client.get_library_docs(
            context7_library_id=f"/moai/skills/{spec['domain']}",
            topic=f"{spec['name']} patterns examples best practices",
            tokens=3000
        )

        # Generate skill content
        return {
            'skill_md': self._generate_skill_content(spec, patterns),
            'examples': self._generate_examples(spec, patterns),
            'reference': self._generate_reference(spec),
            'metadata': self._generate_metadata(spec)
        }

# Usage example
batch_factory = BatchSkillFactory(max_workers=4)

skill_specs = [
    {'name': 'moai-lang-go', 'domain': 'language', 'frameworks': ['gin', 'grpc']},
    {'name': 'moai-lang-rust', 'domain': 'language', 'frameworks': ['axum', 'tokio']},
    {'name': 'moai-lang-java', 'domain': 'language', 'frameworks': ['spring', 'quarkus']},
    {'name': 'moai-lang-csharp', 'domain': 'language', 'frameworks': ['asp.net', 'efcore']}
]

results = await batch_factory.generate_skills_batch(skill_specs, context7_client)
print(f"Generated {results['success_rate']*100:.1f}% skills successfully")
```

## Pattern 5: Context7-Powered Content Generation

Leverage Context7 for real-time pattern matching and content generation.

```python
from dataclasses import dataclass

@dataclass
class Context7Content:
    library_id: str
    topic: str
    frameworks: List[str]
    patterns: List[str]
    examples: List[str]
    best_practices: List[str]
    version_info: Dict[str, str]

class Context7SkillGenerator:
    """Generate skills using Context7 latest documentation."""

    async def generate_with_context7(
        self,
        skill_name: str,
        domain: str,
        context7_client
    ) -> str:
        """Generate skill content using Context7."""

        # Fetch context-aware content
        library_id = f"/moai/{domain}/{skill_name}"

        content = await context7_client.get_library_docs(
            context7_library_id=library_id,
            topic=f"{skill_name} advanced patterns best practices examples 2025",
            tokens=6000
        )

        # Build Context7Content object
        context7_content = Context7Content(
            library_id=library_id,
            topic=content.get('topic', ''),
            frameworks=self._extract_frameworks(content),
            patterns=self._extract_patterns(content),
            examples=self._extract_examples(content),
            best_practices=self._extract_best_practices(content),
            version_info=self._extract_version_info(content)
        )

        # Generate comprehensive skill
        skill_md = self._assemble_skill(skill_name, domain, context7_content)

        return skill_md

    def _assemble_skill(
        self,
        skill_name: str,
        domain: str,
        content: Context7Content
    ) -> str:
        """Assemble final SKILL.md from Context7 content."""

        sections = [
            f"# {skill_name}",
            f"**Domain**: {domain}",
            f"**Latest Versions**: {self._format_versions(content.version_info)}",
            "",
            "## Quick Reference",
            f"Focus areas: {', '.join(content.frameworks)}",
            "",
            "## Core Patterns",
            self._format_patterns_list(content.patterns),
            "",
            "## Best Practices",
            self._format_best_practices_list(content.best_practices),
            "",
            "## Context7 Integration",
            f"This skill is powered by Context7 library documentation: {content.library_id}"
        ]

        return '\n'.join(sections)
```

## Pattern 6: Modular Module Generation

Automatically generate advanced-patterns.md and optimization.md modules.

```python
class ModuleGenerator:
    """Generate standardized modules for skills."""

    async def generate_modules_for_skill(
        self,
        skill_name: str,
        domain: str,
        context7_client
    ) -> Dict[str, str]:
        """Generate advanced-patterns.md and optimization.md."""

        # Fetch content for both modules
        patterns_content = await context7_client.get_library_docs(
            context7_library_id=f"/moai/{domain}/{skill_name}",
            topic=f"{skill_name} advanced patterns enterprise use cases",
            tokens=4000
        )

        optimization_content = await context7_client.get_library_docs(
            context7_library_id=f"/moai/{domain}/{skill_name}",
            topic=f"{skill_name} optimization performance scaling benchmarks",
            tokens=4000
        )

        return {
            'advanced_patterns': self._generate_advanced_patterns_module(
                skill_name, patterns_content
            ),
            'optimization': self._generate_optimization_module(
                skill_name, optimization_content
            ),
            'reference': self._generate_reference_module(skill_name)
        }

    def _generate_advanced_patterns_module(self, skill_name: str, content: Dict) -> str:
        """Generate advanced-patterns.md with 7+ enterprise patterns."""

        patterns = content.get('patterns', [])
        sections = [
            f"# Advanced Patterns - {skill_name}",
            "",
        ]

        for i, pattern in enumerate(patterns[:7], 1):
            sections.append(f"## Pattern {i}: {pattern.get('name', 'Enterprise Pattern')}")
            sections.append("")
            sections.append(f"```{pattern.get('language', 'python')}")
            sections.append(pattern.get('code', ''))
            sections.append("```")
            sections.append("")

        return '\n'.join(sections)

    def _generate_optimization_module(self, skill_name: str, content: Dict) -> str:
        """Generate optimization.md with performance patterns."""

        optimizations = content.get('optimizations', [])
        sections = [
            f"# Optimization Patterns - {skill_name}",
            "",
        ]

        for opt in optimizations:
            sections.append(f"## {opt.get('title', 'Performance Optimization')}")
            sections.append("")
            sections.append(f"**Improvement**: {opt.get('improvement', '~20% faster')}")
            sections.append("")
            sections.append(opt.get('description', ''))
            sections.append("")

        return '\n'.join(sections)
```

## Pattern 7: Skill Versioning and Evolution

Track skill versions with semantic versioning and evolution tracking.

```python
from typing import Optional
from datetime import datetime

@dataclass
class SkillVersion:
    major: int
    minor: int
    patch: int
    date: datetime
    changes: List[str]
    context7_patterns: int
    test_coverage: int

class SkillVersionManager:
    """Manage skill versions and evolution tracking."""

    def create_new_version(
        self,
        current_version: SkillVersion,
        change_type: str,  # 'major', 'minor', 'patch'
        changes: List[str],
        context7_patterns: int,
        test_coverage: int
    ) -> SkillVersion:
        """Create new version with semantic versioning."""

        if change_type == 'major':
            new_version = SkillVersion(
                major=current_version.major + 1,
                minor=0,
                patch=0,
                date=datetime.now(),
                changes=changes,
                context7_patterns=context7_patterns,
                test_coverage=test_coverage
            )
        elif change_type == 'minor':
            new_version = SkillVersion(
                major=current_version.major,
                minor=current_version.minor + 1,
                patch=0,
                date=datetime.now(),
                changes=changes,
                context7_patterns=context7_patterns,
                test_coverage=test_coverage
            )
        else:  # patch
            new_version = SkillVersion(
                major=current_version.major,
                minor=current_version.minor,
                patch=current_version.patch + 1,
                date=datetime.now(),
                changes=changes,
                context7_patterns=context7_patterns,
                test_coverage=test_coverage
            )

        return new_version

    def generate_changelog(self, versions: List[SkillVersion]) -> str:
        """Generate skill changelog."""
        sections = ["# Skill Changelog\n"]

        for version in sorted(versions, key=lambda v: (v.major, v.minor, v.patch), reverse=True):
            sections.append(f"## v{version.major}.{version.minor}.{version.patch} ({version.date.strftime('%Y-%m-%d')})")
            sections.append(f"- Context7 patterns: {version.context7_patterns}")
            sections.append(f"- Test coverage: {version.test_coverage}%")
            sections.append("")
            for change in version.changes:
                sections.append(f"- {change}")
            sections.append("")

        return '\n'.join(sections)
```

---

**Advanced Patterns Summary**: 7 enterprise-grade patterns for skill generation, validation, batching, Context7 integration, modularity, and versioning.

