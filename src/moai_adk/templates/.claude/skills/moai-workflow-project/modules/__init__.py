"""
MoAI Menu Project - Modules Package

Integrated module system for project management including:
- DocumentationManager: Template-based documentation generation
- LanguageInitializer: Language detection and configuration
- TemplateOptimizer: Template analysis and optimization
"""

__version__ = "1.0.0"

# Import core modules
from .documentation_manager import DocumentationManager
from .language_initializer import LanguageInitializer
from .template_optimizer import TemplateOptimizer

__all__ = ["DocumentationManager", "LanguageInitializer", "TemplateOptimizer"]
