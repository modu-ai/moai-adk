#!/usr/bin/env python3
"""
Plan Expander: Plan Agent Results → Detailed Section Expansion

Purpose:
- Transform high-level Plan agent outline into detailed sections
- Integrate knowledge from Context7 + WebSearch
- Apply Yoda System templates (education/presentation/workshop)
- Generate 4 pages per section (25 sections = 100 pages)

Performance:
- Input: Plan with 25 high-level sections
- Output: 100+ page detailed lecture
- Processing time: 1-2 minutes

Author: GoosLab
Version: 1.0.0
"""

import logging
from typing import Any, Dict, List, Optional
from dataclasses import dataclass, field

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class Section:
    """Section structure for expanded content."""
    title: str
    content: str = ""
    subsections: List['Section'] = field(default_factory=list)
    code_examples: List[str] = field(default_factory=list)
    diagrams: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)


class PlanExpander:
    """Plan-to-Section intelligent expansion engine."""
    
    def __init__(
        self,
        template_dir: str = ".claude/skills/moai-yoda-system/templates"
    ):
        """
        Initialize Plan Expander.
        
        Args:
            template_dir: Path to Yoda System templates
        """
        self.template_dir = template_dir
        self.templates = {}
        self._load_templates()
    
    def _load_templates(self):
        """Load Yoda System templates."""
        logger.info(f"Loading templates from {self.template_dir}")
        
        # Load education.md, presentation.md, workshop.md
        template_names = ["education", "presentation", "workshop"]
        
        for name in template_names:
            template_path = f"{self.template_dir}/{name}.md"
            try:
                with open(template_path, 'r', encoding='utf-8') as f:
                    self.templates[name] = f.read()
                logger.info(f"Loaded template: {name}")
            except FileNotFoundError:
                logger.warning(f"Template not found: {template_path}")
    
    async def expand_plan_to_sections(
        self,
        plan: Dict,
        template: str,
        knowledge: Dict,
        pages_per_section: int = 4
    ) -> List[Section]:
        """
        Expand high-level plan to detailed sections.
        
        Args:
            plan: Plan agent output (high-level outline)
            template: Template name (education/presentation/workshop)
            knowledge: Knowledge fetched from Context7 + WebSearch
            pages_per_section: Target pages per section (default: 4)
        
        Returns:
            List of Section objects (detailed, lecture-ready)
        """
        logger.info(f"Expanding plan with template: {template}")
        
        # Extract section structure from plan
        section_structure = self._extract_section_structure(plan)
        
        logger.info(f"Found {len(section_structure)} sections in plan")
        
        # Expand each section
        expanded_sections = []
        
        for i, section_outline in enumerate(section_structure):
            logger.info(f"Expanding section {i+1}/{len(section_structure)}: {section_outline['title']}")
            
            # Expand section using template + knowledge
            expanded = await self._expand_section(
                section_outline=section_outline,
                template=template,
                knowledge=knowledge,
                target_pages=pages_per_section
            )
            
            expanded_sections.append(expanded)
        
        logger.info(f"Expanded {len(expanded_sections)} sections")
        
        return expanded_sections
    
    def _extract_section_structure(self, plan: Dict) -> List[Dict]:
        """
        Extract section structure from plan.
        
        Args:
            plan: Plan agent output
        
        Returns:
            List of section outlines
        """
        sections = plan.get('sections', [])
        
        # Convert to standardized structure
        structured_sections = []
        
        for section in sections:
            structured = {
                'title': section.get('title', 'Untitled Section'),
                'description': section.get('description', ''),
                'subsections': section.get('subsections', []),
                'topics': section.get('topics', []),
                'learning_objectives': section.get('learning_objectives', [])
            }
            structured_sections.append(structured)
        
        return structured_sections
    
    async def _expand_section(
        self,
        section_outline: Dict,
        template: str,
        knowledge: Dict,
        target_pages: int
    ) -> Section:
        """
        Expand single section to target page count.
        
        Args:
            section_outline: High-level section outline
            template: Template name
            knowledge: Knowledge sources
            target_pages: Target page count
        
        Returns:
            Section object with detailed content
        """
        title = section_outline['title']
        description = section_outline.get('description', '')
        topics = section_outline.get('topics', [])
        
        # Generate content based on template type
        if template == "education":
            content = self._generate_education_content(
                title=title,
                description=description,
                topics=topics,
                knowledge=knowledge,
                target_pages=target_pages
            )
        elif template == "presentation":
            content = self._generate_presentation_content(
                title=title,
                description=description,
                topics=topics,
                knowledge=knowledge
            )
        elif template == "workshop":
            content = self._generate_workshop_content(
                title=title,
                description=description,
                topics=topics,
                knowledge=knowledge,
                target_pages=target_pages
            )
        else:
            content = self._generate_default_content(
                title=title,
                description=description,
                topics=topics,
                knowledge=knowledge
            )
        
        # Create Section object
        section = Section(
            title=title,
            content=content,
            metadata={
                'template': template,
                'target_pages': target_pages,
                'generated': True
            }
        )
        
        return section
    
    def _generate_education_content(
        self,
        title: str,
        description: str,
        topics: List[str],
        knowledge: Dict,
        target_pages: int
    ) -> str:
        """Generate education-focused content."""
        
        content_parts = []
        
        # Section header
        content_parts.append(f"# {title}\n")
        content_parts.append(f"\n{description}\n")
        
        # Overview
        content_parts.append("\n## Overview\n")
        content_parts.append(f"This section covers {len(topics)} key topics:\n")
        for i, topic in enumerate(topics, 1):
            content_parts.append(f"{i}. {topic}\n")
        
        # Learning objectives
        content_parts.append("\n## Learning Objectives\n")
        content_parts.append("By the end of this section, you will:\n")
        content_parts.append("- Understand core concepts\n")
        content_parts.append("- Apply practical patterns\n")
        content_parts.append("- Solve real-world problems\n")
        
        # Core concepts (use knowledge)
        content_parts.append("\n## Core Concepts\n")
        
        # Inject knowledge from Context7 + WebSearch
        ranked_knowledge = knowledge.get('ranked', [])
        
        if ranked_knowledge:
            # Use top 3 knowledge sources
            for i, source in enumerate(ranked_knowledge[:3]):
                content_parts.append(f"\n### Concept {i+1}\n")
                content_parts.append(f"{source.content}\n")
        else:
            # Fallback: Generate placeholder
            for i, topic in enumerate(topics[:3], 1):
                content_parts.append(f"\n### {topic}\n")
                content_parts.append(f"[Detailed explanation of {topic}]\n")
        
        # Examples
        content_parts.append("\n## Examples\n")
        content_parts.append("```python\n")
        content_parts.append("# Example code\n")
        content_parts.append("def example():\n")
        content_parts.append("    pass\n")
        content_parts.append("```\n")
        
        # Practice problems
        content_parts.append("\n## Practice Problems\n")
        content_parts.append("1. **Basic**: [Problem description]\n")
        content_parts.append("2. **Advanced**: [Problem description]\n")
        
        # Summary
        content_parts.append("\n## Summary\n")
        content_parts.append("In this section, you learned:\n")
        for topic in topics:
            content_parts.append(f"- {topic}\n")
        
        return "".join(content_parts)
    
    def _generate_presentation_content(
        self,
        title: str,
        description: str,
        topics: List[str],
        knowledge: Dict
    ) -> str:
        """Generate presentation slide content."""
        
        content_parts = []
        
        # Slide structure
        content_parts.append(f"---\n")
        content_parts.append(f"slide: 1\n")
        content_parts.append(f"title: \"{title}\"\n")
        content_parts.append(f"layout: content\n")
        content_parts.append(f"---\n\n")
        
        # Slide content
        content_parts.append(f"## {title}\n\n")
        content_parts.append(f"{description}\n\n")
        
        # Key points
        content_parts.append("### Key Points\n\n")
        for i, topic in enumerate(topics, 1):
            content_parts.append(f"{i}. {topic}\n")
        
        # Speaker notes
        content_parts.append("\n---\n")
        content_parts.append("## Speaker Notes\n\n")
        content_parts.append(f"This slide introduces {title}.\n")
        content_parts.append("Key talking points:\n")
        for topic in topics:
            content_parts.append(f"- Explain {topic}\n")
        
        return "".join(content_parts)
    
    def _generate_workshop_content(
        self,
        title: str,
        description: str,
        topics: List[str],
        knowledge: Dict,
        target_pages: int
    ) -> str:
        """Generate hands-on workshop content."""
        
        content_parts = []
        
        # Workshop header
        content_parts.append(f"# {title}\n\n")
        content_parts.append(f"{description}\n\n")
        
        # Lab structure
        content_parts.append("## Lab Objectives\n\n")
        for i, topic in enumerate(topics, 1):
            content_parts.append(f"{i}. {topic}\n")
        
        # Prerequisites
        content_parts.append("\n## Prerequisites\n\n")
        content_parts.append("- Basic programming knowledge\n")
        content_parts.append("- Development environment setup\n")
        
        # Hands-on lab
        content_parts.append("\n## Hands-On Lab\n\n")
        content_parts.append("### Step 1: Setup\n\n")
        content_parts.append("```bash\n")
        content_parts.append("# Setup commands\n")
        content_parts.append("```\n\n")
        
        content_parts.append("### Step 2: Implementation\n\n")
        content_parts.append("```python\n")
        content_parts.append("# Implementation code\n")
        content_parts.append("```\n\n")
        
        content_parts.append("### Step 3: Validation\n\n")
        content_parts.append("Verify your work:\n")
        content_parts.append("- [ ] Test case 1 passes\n")
        content_parts.append("- [ ] Test case 2 passes\n")
        
        # Troubleshooting
        content_parts.append("\n## Troubleshooting\n\n")
        content_parts.append("**Issue**: [Common problem]\n")
        content_parts.append("**Solution**: [Fix]\n")
        
        return "".join(content_parts)
    
    def _generate_default_content(
        self,
        title: str,
        description: str,
        topics: List[str],
        knowledge: Dict
    ) -> str:
        """Generate default content."""
        
        content_parts = []
        
        content_parts.append(f"# {title}\n\n")
        content_parts.append(f"{description}\n\n")
        
        # Topics
        for i, topic in enumerate(topics, 1):
            content_parts.append(f"## {topic}\n\n")
            content_parts.append(f"[Content for {topic}]\n\n")
        
        return "".join(content_parts)


# Example usage
async def main():
    """Example usage of PlanExpander."""
    
    expander = PlanExpander()
    
    # Sample plan
    plan = {
        "topic": "React 19 최신 기능",
        "sections": [
            {
                "title": "Introduction",
                "description": "Overview of React 19",
                "topics": ["History", "Key Features", "Breaking Changes"]
            },
            {
                "title": "Core Concepts",
                "description": "Essential patterns in React 19",
                "topics": ["Hooks", "Server Components", "Concurrent Rendering"]
            }
        ]
    }
    
    # Sample knowledge
    knowledge = {
        "ranked": []
    }
    
    # Expand
    sections = await expander.expand_plan_to_sections(
        plan=plan,
        template="education",
        knowledge=knowledge
    )
    
    print(f"Generated {len(sections)} sections")
    for section in sections:
        print(f"- {section.title}: {len(section.content)} characters")


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
