"""
Core multilingual linting architecture modules

This package provides:
- LanguageDetector: Automatic project language detection
- LinterRegistry: Language-specific linting runners
- FormatterRegistry: Language-specific formatting runners
- MultilingualLintingHook: Hook integration orchestration
"""

from .language_detector import LanguageDetector
from .linters import LinterRegistry
from .formatters import FormatterRegistry
from .post_tool__multilingual_linting import MultilingualLintingHook

__all__ = [
    "LanguageDetector",
    "LinterRegistry",
    "FormatterRegistry",
    "MultilingualLintingHook",
]

__version__ = "1.0.0"
