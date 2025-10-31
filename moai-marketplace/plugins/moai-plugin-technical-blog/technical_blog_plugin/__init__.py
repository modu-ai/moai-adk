"""
Technical Blog Writing Plugin
Auto-orchestrates 7 specialist agents for blog creation based on natural language directives
"""

__version__ = "2.0.0-dev"
__author__ = "GOOS"

from .parser import DirectiveParser
from .orchestrator import BlogWriteOrchestrator

__all__ = ["DirectiveParser", "BlogWriteOrchestrator"]
