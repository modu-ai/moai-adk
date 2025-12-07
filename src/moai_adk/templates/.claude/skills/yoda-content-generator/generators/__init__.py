"""
moai-yoda-content-generator generators package

Provides intelligent content generation capabilities:
- context7_injector: Parallel knowledge fetching (Context7 MCP + WebSearch)
- plan_expander: Plan-to-Section expansion with template integration
- section_builder: Batch section generation (5 parallel)
- mermaid_embedder: Mermaid diagram inline SVG embedding
"""

__version__ = "1.0.0"
__author__ = "GoosLab"

from .context7_injector import Context7Injector
from .plan_expander import PlanExpander
from .section_builder import SectionBuilder
from .mermaid_embedder import MermaidEmbedder

__all__ = [
    "Context7Injector",
    "PlanExpander",
    "SectionBuilder",
    "MermaidEmbedder",
]
