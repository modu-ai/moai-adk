"""
Documentation generation for SPEC-REDESIGN-001

Generates:
- product.md: Project vision and value proposition
- structure.md: System architecture and components
- tech.md: Technology stack and trade-offs
"""

from copy import deepcopy
from pathlib import Path
from typing import Any, Dict, List, Optional


class DocumentationGenerator:
    """Generates project documentation from brainstorm responses.

    Produces three complementary markdown documents for agent context:
    - product.md: Project vision, users, value, roadmap
    - structure.md: System architecture and components
    - tech.md: Technology stack, trade-offs, performance, security

    Used by project-manager, tdd-implementer, and domain expert agents.

    Example:
        >>> generator = DocumentationGenerator()
        >>> responses = {'project_vision': 'Build fast API', ...}
        >>> docs = generator.generate_all_documents(responses)
        >>> generator.save_all_documents(docs)
    """

    def __init__(self):
        """Initialize with built-in documentation templates."""
        self.templates = self._load_templates()

    @staticmethod
    def _load_templates() -> Dict[str, str]:
        """Load documentation templates.

        Returns:
            Dictionary with 'product', 'structure', 'tech' template strings.
        """
        return {
            "product": """# Project Vision

## Vision Statement
{vision}

## Target Users
{target_users}

## Value Proposition
{value_proposition}

## Roadmap
{roadmap}

## AI Analysis Insights
{ai_insights}

**Generated**: This document was auto-generated based on project analysis.
""",
            "structure": """# System Architecture

## Architecture Overview
{system_architecture}

## Core Components
{core_components}

## Component Relationships
{relationships}

## Dependencies
{dependencies}

**Design Notes**: This architecture document provides reference for implementation teams.
""",
            "tech": """# Technology Stack

## Technology Selection
{technology_selection}

## Trade-off Analysis
{trade_offs}

## Performance Considerations
{performance}

## Security Considerations
{security}

## Setup Guide
{setup_guide}

**Version**: Technology stack selection guide, subject to updates.
""",
        }

    def generate_product_md(self, responses: Dict[str, Any]) -> str:
        """Generate product.md from brainstorm responses.

        Creates marketing and strategic documentation from brainstorm responses.
        Includes vision statement, target users, value proposition, and roadmap.

        Args:
            responses: Brainstorm response dictionary with keys:
                - project_vision: High-level project vision
                - target_users: Description of target user segment
                - value_proposition: Unique value provided
                - roadmap: Development roadmap

        Returns:
            Formatted markdown content for product.md

        Example:
            >>> responses = {
            ...     'project_vision': 'Fast API framework',
            ...     'target_users': 'Python developers',
            ...     'value_proposition': 'Speed and simplicity',
            ...     'roadmap': 'v1.0 in Q1 2025',
            ... }
            >>> gen = DocumentationGenerator()
            >>> content = gen.generate_product_md(responses)
            >>> 'Vision' in content
            True
        """
        template = self.templates["product"]

        content = template.format(
            vision=responses.get("project_vision", ""),
            target_users=responses.get("target_users", ""),
            value_proposition=responses.get("value_proposition", ""),
            roadmap=responses.get("roadmap", ""),
            ai_insights=self._generate_ai_insights(responses),
        )

        return content.strip()

    def generate_structure_md(self, responses: Dict[str, Any]) -> str:
        """Generate structure.md from brainstorm responses.

        Creates architecture documentation for system design.
        Includes architecture overview, components, relationships, dependencies.

        Args:
            responses: Brainstorm response dictionary with keys:
                - system_architecture: Architecture overview
                - core_components: List/description of core components
                - relationships: How components interact
                - dependencies: External dependencies

        Returns:
            Formatted markdown content for structure.md

        Example:
            >>> responses = {
            ...     'system_architecture': 'Microservices',
            ...     'core_components': 'API, Database, Cache',
            ...     'relationships': 'API calls Database',
            ...     'dependencies': 'PostgreSQL, Redis',
            ... }
            >>> gen = DocumentationGenerator()
            >>> content = gen.generate_structure_md(responses)
            >>> 'Architecture' in content.lower()
            True
        """
        template = self.templates["structure"]

        content = template.format(
            system_architecture=responses.get("system_architecture", ""),
            core_components=responses.get("core_components", ""),
            relationships=responses.get("relationships", ""),
            dependencies=responses.get("dependencies", ""),
        )

        return content.strip()

    def generate_tech_md(self, responses: Dict[str, Any]) -> str:
        """Generate tech.md from brainstorm responses.

        Creates technical documentation for technology decisions.
        Includes tech stack, trade-offs, performance, security, setup.

        Args:
            responses: Brainstorm response dictionary with keys:
                - technology_selection: Technologies chosen
                - trade_offs: Trade-offs made in tech selection
                - performance: Performance requirements
                - security: Security considerations
                - setup_guide: How to set up the tech stack

        Returns:
            Formatted markdown content for tech.md

        Example:
            >>> responses = {
            ...     'technology_selection': 'Python, FastAPI, PostgreSQL',
            ...     'trade_offs': 'FastAPI over Django for performance',
            ...     'performance': 'Sub-100ms latency',
            ...     'security': 'OAuth2, TLS',
            ...     'setup_guide': 'pip install -r requirements.txt',
            ... }
            >>> gen = DocumentationGenerator()
            >>> content = gen.generate_tech_md(responses)
            >>> 'Technology' in content
            True
        """
        template = self.templates["tech"]

        content = template.format(
            technology_selection=responses.get("technology_selection", ""),
            trade_offs=responses.get("trade_offs", ""),
            performance=responses.get("performance", ""),
            security=responses.get("security", ""),
            setup_guide=responses.get("setup_guide", ""),
        )

        return content.strip()

    def generate_all_documents(
        self,
        brainstorm_responses: Dict[str, Any],
    ) -> Dict[str, str]:
        """Generate all three documents from brainstorm responses.

        Produces complete documentation set: product, structure, tech.

        Args:
            brainstorm_responses: Brainstorm responses with keys for
                                  product, structure, and tech generation

        Returns:
            Dictionary with 'product', 'structure', 'tech' keys and
            markdown content as values.

        Example:
            >>> gen = DocumentationGenerator()
            >>> responses = {...}
            >>> docs = gen.generate_all_documents(responses)
            >>> len(docs)
            3
        """
        return {
            "product": self.generate_product_md(brainstorm_responses),
            "structure": self.generate_structure_md(brainstorm_responses),
            "tech": self.generate_tech_md(brainstorm_responses),
        }

    @staticmethod
    def _generate_ai_insights(responses: Dict[str, Any]) -> str:
        """Generate AI analysis insights from responses.

        Analyzes brainstorm responses and generates insights about
        completeness and quality.

        Args:
            responses: Brainstorm response dictionary

        Returns:
            Formatted insights text with bullet points or status message.
        """
        # In real implementation, would call AI analysis
        insights = []

        if "project_vision" in responses:
            insights.append("- Vision is clear and actionable")

        if "target_users" in responses:
            insights.append("- Target user segment identified")

        if "value_proposition" in responses:
            insights.append("- Value proposition articulated")

        return "\n".join(insights) if insights else "AI analysis pending"

    def save_all_documents(
        self,
        documents: Dict[str, str],
        base_path: Path = Path(".moai/project"),
    ) -> None:
        """Save all generated documents to disk.

        Writes product.md, structure.md, and tech.md to .moai/project/
        directory. Creates directory if it doesn't exist.

        Args:
            documents: Dictionary with 'product', 'structure', 'tech' keys
            base_path: Directory to save documents (default: .moai/project)

        Returns:
            None

        Example:
            >>> gen = DocumentationGenerator()
            >>> docs = {'product': '...', 'structure': '...', 'tech': '...'}
            >>> gen.save_all_documents(docs)
            >>> (Path('.moai/project') / 'product.md').exists()
            True
        """
        base_path.mkdir(parents=True, exist_ok=True)

        file_mapping = {
            "product": "product.md",
            "structure": "structure.md",
            "tech": "tech.md",
        }

        for doc_type, filename in file_mapping.items():
            if doc_type in documents:
                filepath = base_path / filename
                filepath.write_text(documents[doc_type], encoding="utf-8")

    def load_document(self, doc_name: str, base_path: Path = Path(".moai/project")) -> Optional[str]:
        """Load a generated document from disk.

        Reads markdown document content for use by agents.
        Returns None if file doesn't exist.

        Args:
            doc_name: Document filename (e.g., 'product.md', 'structure.md', 'tech.md')
            base_path: Directory containing documents (default: .moai/project)

        Returns:
            Document content as string, or None if file not found.

        Example:
            >>> gen = DocumentationGenerator()
            >>> content = gen.load_document('product.md')
            >>> 'Vision' in content if content else False
            True
        """
        filepath = base_path / doc_name
        if filepath.exists():
            return filepath.read_text(encoding="utf-8")
        return None

    def create_minimal_templates(self, base_path: Path = Path(".moai/project")) -> None:
        """Create minimal template files for Quick Start mode"""
        base_path.mkdir(parents=True, exist_ok=True)

        minimal_templates = {
            "product.md": """# Project Vision

## Vision Statement
[Add your project vision here]

## Target Users
[Describe your target users]

## Value Proposition
[Explain the unique value]

## Roadmap
[Outline your development roadmap]
""",
            "structure.md": """# System Architecture

## Architecture Overview
[Describe the system architecture]

## Core Components
[List and describe core components]

## Component Relationships
[Explain how components interact]

## Dependencies
[Document external dependencies]
""",
            "tech.md": """# Technology Stack

## Technology Selection
[List selected technologies and frameworks]

## Trade-off Analysis
[Explain trade-offs made]

## Performance Considerations
[Document performance requirements]

## Security Considerations
[Document security measures]

## Setup Guide
[Provide setup instructions]
""",
        }

        for filename, content in minimal_templates.items():
            filepath = base_path / filename
            filepath.write_text(content, encoding="utf-8")


class BrainstormQuestionGenerator:
    """Generates brainstorm questions for different depths"""

    @staticmethod
    def get_quick_questions() -> List[Dict[str, str]]:
        """Get Quick (5-10 min) brainstorm questions"""
        return [
            {
                "id": "q1_vision",
                "question": "What is your project vision in one sentence?",
                "category": "vision",
            },
            {
                "id": "q2_users",
                "question": "Who are your target users?",
                "category": "vision",
            },
            {
                "id": "q3_architecture",
                "question": "What is the basic system architecture?",
                "category": "architecture",
            },
            {
                "id": "q4_tech",
                "question": "What are your key technologies?",
                "category": "tech",
            },
            {
                "id": "q5_team",
                "question": "What is your team composition?",
                "category": "team",
            },
        ]

    @staticmethod
    def get_standard_questions() -> List[Dict[str, str]]:
        """Get Standard (10-15 min) brainstorm questions"""
        quick = BrainstormQuestionGenerator.get_quick_questions()
        additional = [
            {
                "id": "q6_tradeoffs_1",
                "question": "What major trade-offs did you make?",
                "category": "tradeoffs",
            },
            {
                "id": "q7_tradeoffs_2",
                "question": "What alternatives did you consider?",
                "category": "tradeoffs",
            },
            {
                "id": "q8_performance",
                "question": "What are your performance requirements?",
                "category": "performance",
            },
            {
                "id": "q9_security",
                "question": "What security considerations are important?",
                "category": "security",
            },
            {
                "id": "q10_scalability",
                "question": "How should the system scale?",
                "category": "scalability",
            },
        ]
        return quick + additional

    @staticmethod
    def get_deep_questions() -> List[Dict[str, str]]:
        """Get Deep (25-30 min) brainstorm questions"""
        standard = BrainstormQuestionGenerator.get_standard_questions()
        additional = [
            {
                "id": "q11_competitors_1",
                "question": "What competitors exist in this space?",
                "category": "market",
            },
            {
                "id": "q12_competitors_2",
                "question": "How do you differentiate from competitors?",
                "category": "market",
            },
            {
                "id": "q13_market",
                "question": "What market trends are relevant?",
                "category": "market",
            },
            {
                "id": "q14_innovation",
                "question": "What innovative approaches are you using?",
                "category": "innovation",
            },
            {
                "id": "q15_costs",
                "question": "What is your cost model?",
                "category": "business",
            },
            {
                "id": "q16_best_practices",
                "question": "What best practices are you following?",
                "category": "practices",
            },
        ]
        return standard + additional

    @staticmethod
    def get_questions_by_depth(depth: str) -> List[Dict[str, str]]:
        """Get questions for specified depth"""
        if depth == "quick":
            return BrainstormQuestionGenerator.get_quick_questions()
        elif depth == "standard":
            return BrainstormQuestionGenerator.get_standard_questions()
        elif depth == "deep":
            return BrainstormQuestionGenerator.get_deep_questions()
        else:
            return BrainstormQuestionGenerator.get_quick_questions()


class AgentContextInjector:
    """Injects project documentation into agent context"""

    @staticmethod
    def inject_project_manager_context(
        agent_config: Dict[str, Any],
        base_path: Path = Path(".moai/project"),
    ) -> Dict[str, Any]:
        """Inject product.md into project-manager agent"""
        config = deepcopy(agent_config)

        doc_path = base_path / "product.md"
        if doc_path.exists():
            content = doc_path.read_text(encoding="utf-8")
            if "system_context" not in config:
                config["system_context"] = ""
            config["system_context"] += f"\n\n## Project Documentation\n{content}"

        return config

    @staticmethod
    def inject_tdd_implementer_context(
        agent_config: Dict[str, Any],
        base_path: Path = Path(".moai/project"),
    ) -> Dict[str, Any]:
        """Inject structure.md into tdd-implementer agent"""
        config = deepcopy(agent_config)

        doc_path = base_path / "structure.md"
        if doc_path.exists():
            content = doc_path.read_text(encoding="utf-8")
            if "architecture_context" not in config:
                config["architecture_context"] = ""
            config["architecture_context"] += f"\n\n## Architecture Reference\n{content}"

        return config

    @staticmethod
    def inject_domain_expert_context(
        agent_config: Dict[str, Any],
        agent_type: str,  # 'backend_expert' or 'frontend_expert'
        base_path: Path = Path(".moai/project"),
    ) -> Dict[str, Any]:
        """Inject tech.md into domain expert agents"""
        config = deepcopy(agent_config)

        doc_path = base_path / "tech.md"
        if doc_path.exists():
            content = doc_path.read_text(encoding="utf-8")
            if "tech_context" not in config:
                config["tech_context"] = ""
            config["tech_context"] += f"\n\n## Technology Stack Reference\n{content}"

        return config
